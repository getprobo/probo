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

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

// ErrCannotMergeIntoSelf is returned when a merge names the same catalog
// row as both winner and loser, which would delete the row it was meant to
// keep.
var ErrCannotMergeIntoSelf = errors.New("cannot merge a common third party into itself")

// MergeCommonThirdPartyResult reports what a merge moved. A merge is
// otherwise invisible — it repoints references across tables and tenants
// and then deletes a row — so the operator needs a per-table account of
// what happened, and above all which references were left behind.
type MergeCommonThirdPartyResult struct {
	// DomainsMoved and DomainsDroppedAsDup partition the loser's domains:
	// moved to the winner, or discarded because the winner already claimed
	// the same domain.
	DomainsMoved        int64
	DomainsDroppedAsDup int64

	// TrackerPatternsRepointed counts catalog patterns now attributed to
	// the winner.
	TrackerPatternsRepointed int64

	// ThirdPartiesRepointed counts organization third parties, across all
	// tenants, now linked to the winner.
	ThirdPartiesRepointed int64

	// ThirdPartiesSkipped names organization third parties left linked to
	// the loser because their organization already had a third party
	// pointing at the winner. Repointing them would create two rows in one
	// organization referencing the same catalog entry, a state no
	// constraint forbids and that the catalog-import lookup then hides
	// permanently. They end up unlinked once the loser is deleted, so they
	// are reported for follow-up by whoever owns that tenant's data.
	ThirdPartiesSkipped []gid.GID

	// LogoAdopted reports whether the winner took over the loser's logo,
	// which happens only when the winner had none.
	LogoAdopted bool
}

// MergeCommonThirdParty folds loserID into winnerID and deletes the loser.
//
// The catalog is global, so this deliberately spans tenants: an
// organization third party in any tenant may reference a catalog row, and
// all of them must be repointed before the row disappears. That is why no
// Scoper is taken — a tenant scope would silently leave other tenants'
// rows dangling on a deleted id.
//
// Every reference is repointed explicitly rather than left to the foreign
// keys. Letting ON DELETE fire would produce two broken states: catalog
// patterns would keep attribution THIRD_PARTY while their vendor went
// NULL, and the counts reported above would be unobservable.
//
// Step order matters. Domains move first because they are the only step
// that can violate a unique constraint. Patterns and organization third
// parties are repointed before the delete so no cascade sees them. The
// logo is adopted while the loser row still exists. The delete is last.
//
// Scalar metadata (website, policy URLs, certifications) and the
// enrichment payload are NOT merged: the payload records per-field
// provenance for the row's own columns, so mixing two rows' values would
// leave it describing neither. Callers re-arm enrichment on the winner
// instead, which regenerates the whole profile coherently.
func MergeCommonThirdParty(
	ctx context.Context,
	tx pg.Tx,
	winnerID gid.GID,
	loserID gid.GID,
) (MergeCommonThirdPartyResult, error) {
	var result MergeCommonThirdPartyResult

	if winnerID == loserID {
		return result, ErrCannotMergeIntoSelf
	}

	skipped, err := collidingThirdPartyIDs(ctx, tx, winnerID, loserID)
	if err != nil {
		return result, err
	}

	result.ThirdPartiesSkipped = skipped

	moved, dropped, err := mergeCommonThirdPartyDomains(ctx, tx, winnerID, loserID)
	if err != nil {
		return result, err
	}

	result.DomainsMoved = moved
	result.DomainsDroppedAsDup = dropped

	var patterns CommonTrackerPatterns

	result.TrackerPatternsRepointed, err = patterns.RepointCommonThirdPartyID(ctx, tx, loserID, winnerID)
	if err != nil {
		return result, err
	}

	result.ThirdPartiesRepointed, err = repointThirdParties(ctx, tx, winnerID, loserID)
	if err != nil {
		return result, err
	}

	result.LogoAdopted, err = adoptCommonThirdPartyLogo(ctx, tx, winnerID, loserID)
	if err != nil {
		return result, err
	}

	if err := (CommonThirdParty{}).Delete(ctx, tx, loserID); err != nil {
		return result, fmt.Errorf("cannot delete merged common third party: %w", err)
	}

	return result, nil
}

// PreviewMergeCommonThirdParty reports what MergeCommonThirdParty would do
// without writing anything.
//
// It runs the same predicates as the apply path against a read-only
// connection, so the two cannot disagree about which domains collide or
// which organization third parties would be skipped. A merge deletes a
// globally referenced row, so the operator needs that account before
// committing, not after.
func PreviewMergeCommonThirdParty(
	ctx context.Context,
	conn pg.Querier,
	winnerID gid.GID,
	loserID gid.GID,
) (MergeCommonThirdPartyResult, error) {
	var result MergeCommonThirdPartyResult

	if winnerID == loserID {
		return result, ErrCannotMergeIntoSelf
	}

	skipped, err := collidingThirdPartyIDs(ctx, conn, winnerID, loserID)
	if err != nil {
		return result, err
	}

	result.ThirdPartiesSkipped = skipped

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
    ) AS domains_moved,
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
    ) AS domains_dropped,
    (
        SELECT count(*)
        FROM common_tracker_patterns
        WHERE common_third_party_id = @loser_id
    ) AS patterns_repointed,
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
    ) AS third_parties_repointed,
    (
        SELECT count(*)
        FROM common_third_parties AS winner, common_third_parties AS loser
        WHERE winner.id = @winner_id
          AND loser.id = @loser_id
          AND winner.logo_file_id IS NULL
          AND loser.logo_file_id IS NOT NULL
    ) AS logo_adopted
`

	args := pgx.StrictNamedArgs{
		"winner_id": winnerID,
		"loser_id":  loserID,
	}

	var logoAdopted int64

	if err := conn.QueryRow(ctx, q, args).Scan(
		&result.DomainsMoved,
		&result.DomainsDroppedAsDup,
		&result.TrackerPatternsRepointed,
		&result.ThirdPartiesRepointed,
		&logoAdopted,
	); err != nil {
		return result, fmt.Errorf("cannot preview common third party merge: %w", err)
	}

	result.LogoAdopted = logoAdopted > 0

	return result, nil
}

// mergeCommonThirdPartyDomains moves the loser's domains to the winner and
// deletes whatever remains.
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
func mergeCommonThirdPartyDomains(
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

// collidingThirdPartyIDs returns the loser's organization third parties
// whose organization already links the winner.
//
// These cannot be repointed: two rows in one organization pointing at the
// same catalog entry is a state no constraint forbids, and the
// organization-scoped catalog lookup returns only the lowest id, so the
// other row would be permanently invisible to every import and mapping
// path. They are reported instead, and end up unlinked when the loser is
// deleted.
func collidingThirdPartyIDs(
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

// repointThirdParties links the loser's organization third parties to the
// winner, across every tenant, skipping the collisions described by
// collidingThirdPartyIDs.
//
// This is raw SQL rather than a loop over ThirdParty.Update because that
// method rewrites the entity's full column set from a Go struct: it would
// round-trip unrelated state through the application, clobber concurrent
// writes, and could not express the NOT EXISTS guard. Note the deliberate
// absence of any tenant predicate — see MergeCommonThirdParty.
func repointThirdParties(
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

// adoptCommonThirdPartyLogo gives the winner the loser's logo when the
// winner has none, and reports whether it did.
//
// The logo reference has no ON DELETE action, so without this the loser's
// file row survives the merge with nothing referencing it. The guard makes
// the statement a no-op when the winner already has a logo, so an existing
// one is never overwritten.
func adoptCommonThirdPartyLogo(
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
