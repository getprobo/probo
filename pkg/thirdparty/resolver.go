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

// ResolveOrCreateCommonThirdParty links a named vendor to the global
// catalog, creating a row when none matches. Dedup is deterministic:
// exact name, then slug, then the slug with a trailing legal form removed,
// before insert. Callers run inside their own transaction and pass the
// logger explicitly, so it is shared by the tracker mapping worker and the
// common pattern enrichment worker.
//
// The dedup ladder is deliberately conservative. Collapsing two distinct
// vendors onto one row is unrecoverable — every reference downstream loses
// which of them it meant — while leaving two rows for one vendor is fixed
// by merging them, which preserves that information. So each rung admits
// only differences that carry no identity.
//
// It never seeds common_third_party_domains: observed initiator domains
// are a co-occurrence signal, not verified vendor ownership, and the
// global catalog's domain set (used for cross-tenant domain matching) is
// owned by the curated seed instead.
func ResolveOrCreateCommonThirdParty(
	ctx context.Context,
	tx pg.Tx,
	logger *log.Logger,
	name string,
	category coredata.ThirdPartyCategory,
) (*gid.GID, error) {
	// Trim here rather than trusting callers: the tracker mapping worker
	// passes the agent's output through verbatim, so an untrimmed name
	// would be stored with its padding and miss the name lookup below.
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

	// Third rung: the slug of the name with its legal form removed. The
	// agent names vendors freely, so "Hotjar Ltd", "Hotjar Limited", and
	// "Hotjar" all arrive as distinct names with distinct slugs and would
	// each create a row. A legal form carries no identity, and the curated
	// catalog already follows that convention: it stores the brand name and
	// keeps the legal entity in a separate legal_name column.
	//
	// Only the incoming candidate is reduced — the name we store is never
	// rewritten — so a vendor whose brand genuinely ends in a legal form
	// still inserts under its own name when nothing matches.
	//
	// Two things deliberately do NOT happen here:
	//
	// Trailing country and region codes are not stripped. Unlike a legal
	// form, a country code often marks a distinct legal entity with its own
	// data residency and DPA ("OVHcloud US" is not "OVHcloud"), so
	// collapsing it would erase the jurisdictional distinction a privacy
	// register exists to record. Two-letter trailing tokens are also
	// ambiguous in a way closed-class legal forms are not ("Sky IT").
	// A trailing country code is a duplicate-candidate signal for the
	// operator merge tooling, where a human confirms before the
	// irreversible step, not a write-path rule.
	//
	// Products are not folded into their parent. The disambiguation agent
	// does fold them, which is right when matching one organization's
	// vendor list, and wrong for the catalog: "Google Analytics" and
	// "Google Ads" are legitimately distinct entries with different
	// categories and different cookie families.
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
