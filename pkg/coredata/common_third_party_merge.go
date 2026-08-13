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

package coredata

// Statements backing a catalog third party merge, one per affected table.
//
// The ordering these must be applied in, and the reporting of what moved,
// belong to the caller that orchestrates them: see thirdparty.MergeCatalog.
// Each function here does one table's write and returns what it touched.

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

// CommonThirdPartyMergeCounts is one row of counters describing what a merge
// of two catalog entries would touch, per table.
type CommonThirdPartyMergeCounts struct {
	// DomainsMovable and DomainsColliding partition the loser's domains by
	// whether the winner already claims the same domain.
	DomainsMovable   int64 `db:"domains_movable"`
	DomainsColliding int64 `db:"domains_colliding"`

	// TrackerPatterns counts the catalog patterns attributed to the loser.
	TrackerPatterns int64 `db:"tracker_patterns"`

	// ThirdPartiesRepointable counts the loser's organization third parties
	// whose organization does not already link the winner.
	ThirdPartiesRepointable int64 `db:"third_parties_repointable"`

	// LogoAdoptable reports whether the winner would inherit the loser's
	// logo, which happens only when the winner has none.
	LogoAdoptable bool `db:"logo_adoptable"`
}

// CountCommonThirdPartyMergeEffects counts, in one round trip, what merging
// loserID into winnerID would touch.
//
// Every predicate below mirrors the one in the corresponding write, so a
// preview built from these counters cannot disagree with the merge it
// describes.
func CountCommonThirdPartyMergeEffects(
	ctx context.Context,
	conn pg.Querier,
	winnerID gid.GID,
	loserID gid.GID,
) (CommonThirdPartyMergeCounts, error) {
	q := `
SELECT
    (
        SELECT count(*)
        FROM common_third_party_domains AS loser
        WHERE loser.common_third_party_id = @loser_id
          AND NOT EXISTS (
              SELECT 1
              FROM common_third_party_domains AS winner
              WHERE winner.common_third_party_id = @winner_id
                AND winner.domain = loser.domain
          )
    ) AS domains_movable,
    (
        SELECT count(*)
        FROM common_third_party_domains AS loser
        WHERE loser.common_third_party_id = @loser_id
          AND EXISTS (
              SELECT 1
              FROM common_third_party_domains AS winner
              WHERE winner.common_third_party_id = @winner_id
                AND winner.domain = loser.domain
          )
    ) AS domains_colliding,
    (
        SELECT count(*)
        FROM common_tracker_patterns
        WHERE common_third_party_id = @loser_id
    ) AS tracker_patterns,
    (
        SELECT count(*)
        FROM third_parties AS loser
        WHERE loser.common_third_party_id = @loser_id
          AND NOT EXISTS (
              SELECT 1
              FROM third_parties AS winner
              WHERE winner.organization_id = loser.organization_id
                AND winner.common_third_party_id = @winner_id
          )
    ) AS third_parties_repointable,
    EXISTS (
        SELECT 1
        FROM common_third_parties AS winner, common_third_parties AS loser
        WHERE winner.id = @winner_id
          AND loser.id = @loser_id
          AND winner.logo_file_id IS NULL
          AND loser.logo_file_id IS NOT NULL
    ) AS logo_adoptable
`

	args := pgx.StrictNamedArgs{
		"winner_id": winnerID,
		"loser_id":  loserID,
	}

	var counts CommonThirdPartyMergeCounts

	if err := conn.QueryRow(ctx, q, args).Scan(
		&counts.DomainsMovable,
		&counts.DomainsColliding,
		&counts.TrackerPatterns,
		&counts.ThirdPartiesRepointable,
		&counts.LogoAdoptable,
	); err != nil {
		return counts, fmt.Errorf("cannot count common third party merge effects: %w", err)
	}

	return counts, nil
}

// MergeCommonThirdPartyDomains moves the loser's domains to the winner and
//
// This is the only step that can violate a unique constraint:
// common_third_party_domains is unique on (common_third_party_id, domain),
// so a domain both rows claim would collide. The NOT EXISTS guard moves
// only the domains the winner lacks, and the delete then clears the
// collisions. Moving rather than re-inserting preserves each domain row's
// id and creation time.
//
// domain is CITEXT, so the equality below is already case-insensitive and
// matches the unique index's own semantics. Wrapping it in lower() would
// diverge from the index and let a collision through.
func MergeCommonThirdPartyDomains(
	ctx context.Context,
	tx pg.Tx,
	winnerID gid.GID,
	loserID gid.GID,
) (moved int64, dropped int64, err error) {
	moveQuery := `
UPDATE common_third_party_domains AS loser
SET
    common_third_party_id = @winner_id,
    updated_at = NOW()
WHERE
    loser.common_third_party_id = @loser_id
    AND NOT EXISTS (
        SELECT 1
        FROM common_third_party_domains AS winner
        WHERE winner.common_third_party_id = @winner_id
          AND winner.domain = loser.domain
    )
`

	args := pgx.StrictNamedArgs{
		"winner_id": winnerID,
		"loser_id":  loserID,
	}

	moveResult, err := tx.Exec(ctx, moveQuery, args)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot move common third party domains: %w", err)
	}

	// Delete explicitly instead of relying on ON DELETE CASCADE: the count
	// of dropped duplicates is part of the operator's account of the merge,
	// and a cascade would discard it silently.
	deleteQuery := `
DELETE FROM common_third_party_domains
WHERE common_third_party_id = @loser_id
`

	deleteResult, err := tx.Exec(ctx, deleteQuery, pgx.StrictNamedArgs{"loser_id": loserID})
	if err != nil {
		return 0, 0, fmt.Errorf("cannot delete merged common third party domains: %w", err)
	}

	return moveResult.RowsAffected(), deleteResult.RowsAffected(), nil
}

// RelinkOrgTrackerPatterns points each organization's unlinked tracker
// patterns at the third party that organization manages for the winner.
//
// Run after the catalog patterns have been repointed, so the resolution
// below sees the winner. It reuses LinkThirdPartyByCommonThirdPartyID, which
// only touches patterns with no third_party_id and so never overrides an
// existing link, making it safe to run for every affected organization.
//
// Spans tenants, scoping per organization: the catalog is global, and each
// organization's patterns must resolve to that organization's own third
// party, never another tenant's.
func RelinkOrgTrackerPatterns(
	ctx context.Context,
	tx pg.Tx,
	winnerID gid.GID,
) (int64, error) {
	var parties ThirdParties

	links, err := parties.LoadOrganizationLinksByCommonThirdPartyID(ctx, tx, NewNoScope(), winnerID)
	if err != nil {
		return 0, err
	}

	var relinked int64

	for _, link := range links {
		before, err := CountUnlinkedOrgPatterns(ctx, tx, link.OrganizationID, winnerID)
		if err != nil {
			return 0, err
		}

		if before == 0 {
			continue
		}

		var patterns TrackerPatterns

		if err := patterns.LinkThirdPartyByCommonThirdPartyID(
			ctx,
			tx,
			NewScope(link.OrganizationID.TenantID()),
			link.OrganizationID,
			winnerID,
			link.ThirdPartyID,
		); err != nil {
			return 0, err
		}

		relinked += before
	}

	return relinked, nil
}

// CountUnlinkedOrgPatterns counts an organization's tracker patterns that
// resolve to the given catalog entry but carry no third party yet. It is the
// same predicate LinkThirdPartyByCommonThirdPartyID updates, read before the
// write so the merge can report how many rows it moved.
func CountUnlinkedOrgPatterns(
	ctx context.Context,
	conn pg.Querier,
	organizationID gid.GID,
	commonThirdPartyID gid.GID,
) (int64, error) {
	q := `
SELECT
    count(*)
FROM
    tracker_patterns
WHERE
    organization_id = @organization_id
    AND third_party_id IS NULL
    AND common_tracker_pattern_id IN (
        SELECT id FROM common_tracker_patterns
        WHERE common_third_party_id = @common_third_party_id
    )
`

	args := pgx.StrictNamedArgs{
		"organization_id":       organizationID,
		"common_third_party_id": commonThirdPartyID,
	}

	var count int64

	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count unlinked organization tracker patterns: %w", err)
	}

	return count, nil
}

// CollidingThirdPartyIDs returns the loser's organization third parties
// whose organization already links the winner.
//
// These cannot be repointed: two rows in one organization pointing at the
// same catalog entry is a state no constraint forbids, and the
// organization-scoped catalog lookup returns only the lowest id, so the
// other row would be permanently invisible to every import and mapping
// path. They are reported instead, and end up unlinked when the loser is
// deleted.
func CollidingThirdPartyIDs(
	ctx context.Context,
	conn pg.Querier,
	winnerID gid.GID,
	loserID gid.GID,
) ([]gid.GID, error) {
	q := `
SELECT
    loser.id
FROM
    third_parties AS loser
WHERE
    loser.common_third_party_id = @loser_id
    AND EXISTS (
        SELECT 1
        FROM third_parties AS winner
        WHERE winner.organization_id = loser.organization_id
          AND winner.common_third_party_id = @winner_id
    )
ORDER BY
    loser.id ASC
`

	args := pgx.StrictNamedArgs{
		"winner_id": winnerID,
		"loser_id":  loserID,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query colliding third parties: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect colliding third parties: %w", err)
	}

	return ids, nil
}

// RepointThirdPartiesToCommonThirdParty links the loser's organization third parties to the
// winner, across every tenant, skipping the collisions described by
// CollidingThirdPartyIDs.
//
// This is raw SQL rather than a loop over ThirdParty.Update because that
// method rewrites the entity's full column set from a Go struct: it would
// round-trip unrelated state through the application, clobber concurrent
// writes, and could not express the NOT EXISTS guard. Note the deliberate
// absence of any tenant predicate — see MergeCommonThirdParty.
func RepointThirdPartiesToCommonThirdParty(
	ctx context.Context,
	tx pg.Tx,
	winnerID gid.GID,
	loserID gid.GID,
) (int64, error) {
	q := `
UPDATE third_parties AS loser
SET
    common_third_party_id = @winner_id,
    updated_at = NOW()
WHERE
    loser.common_third_party_id = @loser_id
    AND NOT EXISTS (
        SELECT 1
        FROM third_parties AS winner
        WHERE winner.organization_id = loser.organization_id
          AND winner.common_third_party_id = @winner_id
    )
`

	args := pgx.StrictNamedArgs{
		"winner_id": winnerID,
		"loser_id":  loserID,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot repoint third parties: %w", err)
	}

	return result.RowsAffected(), nil
}

// AdoptCommonThirdPartyLogo gives the winner the loser's logo when the
// winner has none, and reports whether it did.
//
// The logo reference has no ON DELETE action, so without this the loser's
// file row survives the merge with nothing referencing it. The guard makes
// the statement a no-op when the winner already has a logo, so an existing
// one is never overwritten.
func AdoptCommonThirdPartyLogo(
	ctx context.Context,
	tx pg.Tx,
	winnerID gid.GID,
	loserID gid.GID,
) (bool, error) {
	q := `
UPDATE common_third_parties AS winner
SET
    logo_file_id = loser.logo_file_id,
    updated_at = NOW()
FROM
    common_third_parties AS loser
WHERE
    winner.id = @winner_id
    AND loser.id = @loser_id
    AND winner.logo_file_id IS NULL
    AND loser.logo_file_id IS NOT NULL
`

	args := pgx.StrictNamedArgs{
		"winner_id": winnerID,
		"loser_id":  loserID,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot adopt merged common third party logo: %w", err)
	}

	return result.RowsAffected() > 0, nil
}
