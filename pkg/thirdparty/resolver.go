// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package thirdparty

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/slug"
)

// ResolveOrCreateCommonThirdParty links a named vendor to the global catalog,
// creating a row when none matches. Dedup is deterministic: exact name, then
// slug, then the slug with a trailing legal form removed. Callers run inside
// their own transaction and pass the logger explicitly, so it is shared by the
// tracker mapping worker and the common pattern enrichment worker.
//
// Each rung admits only differences that carry no identity, because collapsing
// two distinct vendors is unrecoverable — every reference loses which one it
// meant — while two rows for one vendor are fixed by merging them.
//
// It never seeds common_third_party_domains: observed initiator domains are a
// co-occurrence signal, not verified ownership, and the catalog's domain set
// is owned by the curated seed.
func ResolveOrCreateCommonThirdParty(
	ctx context.Context,
	tx pg.Tx,
	logger *log.Logger,
	name string,
	category coredata.ThirdPartyCategory,
) (*gid.GID, error) {
	// Trimmed here rather than trusting callers: the mapping worker forwards
	// the agent's output verbatim.
	name = strings.TrimSpace(name)

	var party coredata.CommonThirdParty
	if err := party.LoadByName(ctx, tx, name); err == nil {
		return &party.ID, nil
	} else if !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load common third party by name: %w", err)
	}

	partySlug := slug.Make(name)
	if partySlug == "" {
		return nil, nil
	}

	if err := party.LoadBySlug(ctx, tx, partySlug); err == nil {
		return &party.ID, nil
	} else if !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load common third party by slug: %w", err)
	}

	// Third rung: the slug with the legal form removed, so "Hotjar Ltd" and
	// "Hotjar" do not become two rows. The curated catalog already follows
	// that convention, storing the brand name and the legal entity separately.
	// Only the incoming candidate is reduced, never the stored name, so a
	// brand that genuinely ends in a legal form still inserts as itself.
	//
	// Two things deliberately do NOT happen here. Country codes are not
	// stripped: "OVHcloud US" is a distinct entity with its own DPA, and
	// short forms are ambiguous ("Sky IT") in a way legal forms are not — the
	// merge tooling surfaces those for a human instead. Products are not
	// folded into their parent either: the disambiguation agent does that for
	// an organization's vendor list, but "Google Analytics" and "Google Ads"
	// are legitimately distinct catalog entries.
	if strippedSlug := slug.Make(stripCorporateSuffixes(strings.ToLower(name))); strippedSlug != "" &&
		strippedSlug != partySlug {
		if err := party.LoadBySlug(ctx, tx, strippedSlug); err == nil {
			return &party.ID, nil
		} else if !errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, fmt.Errorf("cannot load common third party by suffix-stripped slug: %w", err)
		}
	}

	now := time.Now()
	party = coredata.CommonThirdParty{
		ID:             gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
		Name:           name,
		Slug:           partySlug,
		Category:       category,
		Certifications: []string{},
		// Request enrichment at creation: a freshly resolved catalog row
		// carries only name/slug/category, so the enrichment worker fills
		// the rest (URLs, address, certifications, logo). Curated seed
		// rows are inserted via Upsert without this flag, so a full
		// re-seed does not trigger an enrichment storm.
		EnrichmentRequestedAt: &now,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	// Insert inside a savepoint so a concurrent transaction that created
	// the same slug between our lookup and write does not abort the
	// caller's transaction. On the unique-violation race, reload the
	// winning row and return it instead of failing.
	insertErr := tx.Savepoint(ctx, func(ctx context.Context, sp pg.Tx) error {
		return party.Insert(ctx, sp)
	})
	if insertErr != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](insertErr); ok &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "common_third_parties_slug_key" {
			if err := party.LoadBySlug(ctx, tx, partySlug); err != nil {
				return nil, fmt.Errorf("cannot reload common third party after insert race: %w", err)
			}

			return &party.ID, nil
		}

		return nil, fmt.Errorf("cannot create common third party: %w", insertErr)
	}

	logger.InfoCtx(
		ctx,
		"created common third party from agent identification",
		log.String("name", name),
		log.String("category", category.String()),
	)

	return &party.ID, nil
}
