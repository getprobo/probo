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

	// TrackerPatternsRequeued counts catalog patterns re-armed for
	// enrichment because their description was researched against the
	// vendor that no longer exists.
	TrackerPatternsRequeued int64

	// OrgTrackerPatternsRelinked counts organization tracker patterns now
	// pointing at the third party their organization manages for the
	// winner, across all tenants.
	OrgTrackerPatternsRelinked int64

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

	movedPatterns, err := patterns.RepointCommonThirdPartyID(ctx, tx, loserID, winnerID)
	if err != nil {
		return result, err
	}

	result.TrackerPatternsRepointed = int64(len(movedPatterns))

	// Those patterns' descriptions were researched against a vendor that no
	// longer exists, so re-arm them for a vendor-informed second pass.
	if len(movedPatterns) > 0 {
		result.TrackerPatternsRequeued, err = patterns.RequestEnrichmentByIDs(ctx, tx, movedPatterns)
		if err != nil {
			return result, err
		}
	}

	result.ThirdPartiesRepointed, err = repointThirdParties(ctx, tx, winnerID, loserID)
	if err != nil {
		return result, err
	}

	// Re-point the organization-scoped tracker patterns last, once every
	// catalog reference names the winner.
	//
	// An organization that already managed a third party for the winner has
	// patterns that, before the merge, resolved through the loser and so
	// carried no organization link. Their catalog row now names the winner,
	// which that organization does manage, so they must surface the managed
	// third party rather than the catalog entry. Nothing else would do it:
	// the import path is idempotent on (organization, catalog entry) and
	// skips an organization that already holds the row.
	result.OrgTrackerPatternsRelinked, err = relinkOrgTrackerPatterns(ctx, tx, winnerID)
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

	// Every repointed pattern is re-armed for enrichment.
	result.TrackerPatternsRequeued = result.TrackerPatternsRepointed

	// The organization relink runs after the catalog patterns move, so
	// predicting it means counting the patterns that will resolve to the
	// winner once they have: the loser's patterns plus any the winner
	// already carries, for each organization that manages the winner.
	var parties ThirdParties

	links, err := parties.LoadOrganizationLinksByCommonThirdPartyID(ctx, conn, NewNoScope(), winnerID)
	if err != nil {
		return result, err
	}

	for _, link := range links {
		for _, catalogID := range []gid.GID{winnerID, loserID} {
			count, err := countUnlinkedOrgPatterns(ctx, conn, link.OrganizationID, catalogID)
			if err != nil {
				return result, err
			}

			result.OrgTrackerPatternsRelinked += count
		}
	}

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

// relinkOrgTrackerPatterns points each organization's unlinked tracker
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
func relinkOrgTrackerPatterns(
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
		before, err := countUnlinkedOrgPatterns(ctx, tx, link.OrganizationID, winnerID)
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

// countUnlinkedOrgPatterns counts an organization's tracker patterns that
// resolve to the given catalog entry but carry no third party yet. It is the
// same predicate LinkThirdPartyByCommonThirdPartyID updates, read before the
// write so the merge can report how many rows it moved.
func countUnlinkedOrgPatterns(
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
