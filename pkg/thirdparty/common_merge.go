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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

var ErrCannotMergeIntoSelf = errors.New("cannot merge a common third party into itself")

// MergeCatalogResult is the per-table account of a merge, which is otherwise
// invisible: it repoints references across tables and tenants, then deletes a
// row.
type MergeCatalogResult struct {
	// Partition the loser's domains: moved, or dropped because the winner
	// already claimed the same domain.
	DomainsMoved        int64
	DomainsDroppedAsDup int64

	TrackerPatternsRepointed int64

	// Re-armed for enrichment: their descriptions named the folded vendor.
	TrackerPatternsRequeued int64

	OrgTrackerPatternsRelinked int64
	ThirdPartiesRepointed      int64

	// ThirdPartiesSkipped were left linked to the loser because their
	// organization already had a third party pointing at the winner.
	// Repointing them would put two rows in one organization on the same
	// catalog entry, which no constraint forbids and the catalog-import
	// lookup then hides permanently. They end up unlinked once the loser is
	// deleted, so they are reported for follow-up.
	ThirdPartiesSkipped []gid.GID

	LogoAdopted bool
}

// MergeCatalog folds loserID into winnerID and deletes the loser.
//
// Deliberately spans tenants, hence no Scoper: the catalog is global, so an
// organization third party in any tenant may reference the loser and all of
// them must move before it disappears. A tenant scope would leave other
// tenants dangling on a deleted id.
//
// References are repointed explicitly rather than left to ON DELETE, which
// would leave catalog patterns claiming attribution THIRD_PARTY with a NULL
// vendor and make the counts unobservable.
//
// Step order is the substance here. Domains move first, the only step that
// can violate a unique constraint. The organization relink runs after the
// catalog patterns move so it resolves against a catalog naming the winner.
// The logo is adopted while the loser still exists. The delete is last.
//
// Scalar metadata and the enrichment payload are NOT merged: the payload
// records per-field provenance for its own row's columns, so mixing two rows
// would leave it describing neither. Callers re-arm enrichment on the winner
// instead.
func MergeCatalog(
	ctx context.Context,
	tx pg.Tx,
	winnerID gid.GID,
	loserID gid.GID,
) (MergeCatalogResult, error) {
	var result MergeCatalogResult

	if winnerID == loserID {
		return result, ErrCannotMergeIntoSelf
	}

	skipped, err := coredata.CollidingThirdPartyIDs(ctx, tx, winnerID, loserID)
	if err != nil {
		return result, err
	}

	result.ThirdPartiesSkipped = skipped

	moved, dropped, err := coredata.MergeCommonThirdPartyDomains(ctx, tx, winnerID, loserID)
	if err != nil {
		return result, err
	}

	result.DomainsMoved = moved
	result.DomainsDroppedAsDup = dropped

	var patterns coredata.CommonTrackerPatterns

	movedPatterns, err := patterns.RepointCommonThirdPartyID(ctx, tx, loserID, winnerID)
	if err != nil {
		return result, err
	}

	result.TrackerPatternsRepointed = int64(len(movedPatterns))

	if len(movedPatterns) > 0 {
		result.TrackerPatternsRequeued, err = patterns.RequestEnrichmentByIDs(ctx, tx, movedPatterns)
		if err != nil {
			return result, err
		}
	}

	result.ThirdPartiesRepointed, err = coredata.RepointThirdPartiesToCommonThirdParty(ctx, tx, winnerID, loserID)
	if err != nil {
		return result, err
	}

	// Patterns that resolved through the loser carried no organization link,
	// so they surfaced the catalog entry. Their catalog row now names the
	// winner, which the organization does manage, and nothing else would
	// relink them: the import path skips an organization that already holds
	// the row.
	result.OrgTrackerPatternsRelinked, err = coredata.RelinkOrgTrackerPatterns(ctx, tx, winnerID)
	if err != nil {
		return result, err
	}

	result.LogoAdopted, err = coredata.AdoptCommonThirdPartyLogo(ctx, tx, winnerID, loserID)
	if err != nil {
		return result, err
	}

	if err := (coredata.CommonThirdParty{}).Delete(ctx, tx, loserID); err != nil {
		return result, fmt.Errorf("cannot delete merged common third party: %w", err)
	}

	return result, nil
}
