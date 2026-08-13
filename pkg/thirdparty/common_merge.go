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

// ErrCannotMergeIntoSelf is returned when a merge names the same catalog
// entry as both winner and loser, which would delete the entry it was meant
// to keep.
var ErrCannotMergeIntoSelf = errors.New("cannot merge a common third party into itself")

// MergeCatalogResult reports what a merge moved. A merge is otherwise
// invisible — it repoints references across tables and tenants and then
// deletes a row — so the operator needs a per-table account of what
// happened, and above all which references were left behind.
type MergeCatalogResult struct {
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

// MergeCatalog folds loserID into winnerID and deletes the loser.
//
// The catalog is global, so this deliberately spans tenants: an organization
// third party in any tenant may reference a catalog row, and all of them
// must be repointed before the row disappears. That is why no Scoper is
// taken — a tenant scope would silently leave other tenants' rows dangling
// on a deleted id.
//
// Every reference is repointed explicitly rather than left to the foreign
// keys. Letting ON DELETE fire would produce two broken states: catalog
// patterns would keep attribution THIRD_PARTY while their vendor went NULL,
// and the counts reported above would be unobservable.
//
// Step order matters, and is the reason this orchestration exists rather
// than a single statement. Domains move first because they are the only step
// that can violate a unique constraint. Patterns and organization third
// parties are repointed before the organization relink, so the relink
// resolves against a catalog that already names the winner. The logo is
// adopted while the loser row still exists. The delete is last.
//
// Scalar metadata (website, policy URLs, certifications) and the enrichment
// payload are NOT merged: the payload records per-field provenance for the
// row's own columns, so mixing two rows' values would leave it describing
// neither. Callers re-arm enrichment on the winner instead, which
// regenerates the whole profile coherently.
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

	// Those patterns' descriptions were researched against a vendor that no
	// longer exists, so re-arm them for a vendor-informed second pass.
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

// PreviewMergeCatalog reports what MergeCatalog would do without writing
// anything.
//
// It runs the same predicates as the apply path against a read-only
// connection, so the two cannot disagree about which domains collide or
// which organization third parties would be skipped. A merge deletes a
// globally referenced row, so the operator needs that account before
// committing, not after.
func PreviewMergeCatalog(
	ctx context.Context,
	conn pg.Querier,
	winnerID gid.GID,
	loserID gid.GID,
) (MergeCatalogResult, error) {
	var result MergeCatalogResult

	if winnerID == loserID {
		return result, ErrCannotMergeIntoSelf
	}

	skipped, err := coredata.CollidingThirdPartyIDs(ctx, conn, winnerID, loserID)
	if err != nil {
		return result, err
	}

	result.ThirdPartiesSkipped = skipped

	counts, err := coredata.CountCommonThirdPartyMergeEffects(ctx, conn, winnerID, loserID)
	if err != nil {
		return result, err
	}

	result.DomainsMoved = counts.DomainsMovable
	result.DomainsDroppedAsDup = counts.DomainsColliding
	result.TrackerPatternsRepointed = counts.TrackerPatterns
	result.ThirdPartiesRepointed = counts.ThirdPartiesRepointable
	result.LogoAdopted = counts.LogoAdoptable

	// Every repointed pattern is re-armed for enrichment.
	result.TrackerPatternsRequeued = result.TrackerPatternsRepointed

	// The organization relink runs after the catalog patterns move, so
	// predicting it means counting the patterns that will resolve to the
	// winner once they have: the loser's patterns plus any the winner
	// already carries, for each organization that manages the winner.
	var parties coredata.ThirdParties

	links, err := parties.LoadOrganizationLinksByCommonThirdPartyID(ctx, conn, coredata.NewNoScope(), winnerID)
	if err != nil {
		return result, err
	}

	for _, link := range links {
		for _, catalogID := range []gid.GID{winnerID, loserID} {
			count, err := coredata.CountUnlinkedOrgPatterns(ctx, conn, link.OrganizationID, catalogID)
			if err != nil {
				return result, err
			}

			result.OrgTrackerPatternsRelinked += count
		}
	}

	return result, nil
}
