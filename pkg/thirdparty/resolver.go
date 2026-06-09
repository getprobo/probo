// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package thirdparty

import (
	"context"
	"errors"
	"fmt"
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
// exact name, then slug, before insert. Callers run inside their own
// transaction and pass the logger explicitly, so it is shared by the
// tracker mapping worker and the common pattern enrichment worker.
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

	now := time.Now()
	party = coredata.CommonThirdParty{
		ID:             gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
		Name:           name,
		Slug:           partySlug,
		Category:       category,
		Certifications: []string{},
		CreatedAt:      now,
		UpdatedAt:      now,
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
