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

package thirdparty_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/thirdparty"
)

// seedCatalogParty inserts a global catalog third party with a collision-free
// name and slug. The catalog is not tenant-scoped and is uniquely indexed on
// slug, so parallel tests must namespace their rows.
func seedCatalogParty(t *testing.T, ctx context.Context, client *pg.Client) coredata.CommonThirdParty {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	id := gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType)
	suffix := id.String()

	party := coredata.CommonThirdParty{
		ID:             id,
		Name:           "Acme " + suffix,
		Slug:           "acme-" + suffix,
		Category:       coredata.ThirdPartyCategoryAnalytics,
		Certifications: []string{},
		Review:         coredata.CommonThirdPartyReviewUnreviewed,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return party.Insert(ctx, tx)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM common_third_parties WHERE id = $1`, id)
			return err
		})
	})

	return party
}

// insertCatalogPattern inserts a catalog pattern verbatim so a test can stage
// an exact attribution and enrichment state.
func insertCatalogPattern(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	cp coredata.CommonTrackerPattern,
) {
	t.Helper()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return cp.Insert(ctx, tx)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM common_tracker_patterns WHERE id = $1`, cp.ID)
			return err
		})
	})
}

func loadCatalogPattern(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	id gid.GID,
) coredata.CommonTrackerPattern {
	t.Helper()

	var reloaded coredata.CommonTrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, id)
	}))

	return reloaded
}

// seedMergeTestOrganization inserts a placeholder organization in its own
// tenant, which both org third parties and files require.
func seedMergeTestOrganization(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
) gid.GID {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(
			ctx,
			`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
			organizationID.String(), tenantID.String(), "merge-test-org-"+organizationID.String(), now, now,
		)

		return err
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID.String())
			return err
		})
	})

	return organizationID
}

// seedFile inserts a file row so a catalog logo reference has a valid
// target. The logo foreign key has no ON DELETE action, so the row must
// really exist.
func seedFile(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
) gid.GID {
	t.Helper()

	organizationID := seedMergeTestOrganization(t, ctx, client)
	tenantID := organizationID.TenantID()
	fileID := gid.New(tenantID, coredata.FileEntityType)
	now := time.Now().UTC().Truncate(time.Microsecond)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(
			ctx,
			`INSERT INTO files
			     (id, tenant_id, organization_id, bucket_name, mime_type, file_name, file_key, file_size, visibility, created_at, updated_at)
			 VALUES ($1, $2, $3, 'test-bucket', 'image/png', 'logo.png', $4, 128, 'PRIVATE', $5, $5)`,
			fileID.String(),
			tenantID.String(),
			organizationID.String(),
			uuid.New().String(),
			now,
		)

		return err
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM files WHERE id = $1`, fileID.String())
			return err
		})
	})

	return fileID
}

func setCommonThirdPartyLogo(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	partyID gid.GID,
	fileID gid.GID,
) {
	t.Helper()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return coredata.CommonThirdParty{
			ID:         partyID,
			LogoFileID: &fileID,
			UpdatedAt:  time.Now().UTC().Truncate(time.Microsecond),
		}.UpdateLogoFileID(ctx, tx)
	}))
}

func loadCommonThirdPartyLogo(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	partyID gid.GID,
) *gid.GID {
	t.Helper()

	var party coredata.CommonThirdParty

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return party.LoadByID(ctx, conn, partyID)
	}))

	return party.LogoFileID
}

// seedOrgTrackerPattern creates a cookie banner, a category, and one
// organization tracker pattern resolving to the given catalog pattern with no
// third party of its own. It returns the tracker pattern id.
//
// A pattern in this state surfaces the catalog entry rather than a managed
// third party, which is what the catalog merge has to fix once the entry it
// resolves through moves.
func seedOrgTrackerPattern(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	organizationID gid.GID,
	commonTrackerPatternID gid.GID,
) gid.GID {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	tenantID := organizationID.TenantID()
	bannerID := gid.New(tenantID, coredata.CookieBannerEntityType)
	categoryID := gid.New(tenantID, coredata.CookieCategoryEntityType)
	patternID := gid.New(tenantID, coredata.TrackerPatternEntityType)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO cookie_banners
			     (id, tenant_id, organization_id, name, origin, state, consent_expiry_days,
			      show_branding, default_language, cookie_policy_url, capabilities, created_at, updated_at)
			 VALUES ($1,$2,$3,'merge-test',$5,'ACTIVE',180,true,'en',
			         'https://merge.test/cookies','{}'::jsonb,$4,$4)`,
			bannerID.String(), tenantID.String(), organizationID.String(), now,
			// One active banner per origin, so namespace it: an organization
			// may need several banners across tests.
			"https://"+strings.ToLower(bannerID.String())+".merge.test",
		); err != nil {
			return err
		}

		if _, err := tx.Exec(
			ctx,
			`INSERT INTO cookie_categories
			     (id, tenant_id, organization_id, cookie_banner_id, name, description, rank,
			      kind, slug, gcm_consent_types, posthog_consent, created_at, updated_at)
			 VALUES ($1,$2,$3,$4,'Analytics','',1,'NORMAL','analytics','{}',false,$5,$5)`,
			categoryID.String(), tenantID.String(), organizationID.String(), bannerID.String(), now,
		); err != nil {
			return err
		}

		_, err := tx.Exec(
			ctx,
			`INSERT INTO tracker_patterns
			     (id, tenant_id, organization_id, cookie_banner_id, cookie_category_id,
			      tracker_type, pattern, match_type, display_name, description, excluded,
			      common_tracker_pattern_id, third_party_id, created_at, updated_at)
			 VALUES ($1,$2,$3,$4,$5,'COOKIE',$6,'EXACT','Merge test','',false,$7,NULL,$8,$8)`,
			patternID.String(), tenantID.String(), organizationID.String(), bannerID.String(),
			categoryID.String(), "_merge_"+patternID.String(), commonTrackerPatternID.String(), now,
		)

		return err
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			for _, q := range []string{
				`DELETE FROM tracker_patterns WHERE id = $1`,
			} {
				if _, err := tx.Exec(ctx, q, patternID.String()); err != nil {
					return err
				}
			}

			if _, err := tx.Exec(ctx, `DELETE FROM cookie_categories WHERE id = $1`, categoryID.String()); err != nil {
				return err
			}

			_, err := tx.Exec(ctx, `DELETE FROM cookie_banners WHERE id = $1`, bannerID.String())

			return err
		})
	})

	return patternID
}

// loadOrgTrackerPatternThirdPartyID reads an organization tracker pattern's
// third party link.
func loadOrgTrackerPatternThirdPartyID(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	patternID gid.GID,
) *gid.GID {
	t.Helper()

	var got *gid.GID

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return conn.QueryRow(
			ctx,
			`SELECT third_party_id FROM tracker_patterns WHERE id = $1`,
			patternID,
		).Scan(&got)
	}))

	return got
}

// seedCommonThirdPartyDomain attaches a domain to a catalog third party.
func seedCommonThirdPartyDomain(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	partyID gid.GID,
	domain string,
) coredata.CommonThirdPartyDomain {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	d := coredata.CommonThirdPartyDomain{
		ID:                 gid.New(gid.NilTenant, coredata.CommonThirdPartyDomainEntityType),
		CommonThirdPartyID: partyID,
		Domain:             domain,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return d.Insert(ctx, tx)
	}))

	return d
}

// loadCommonThirdPartyDomains returns a party's domains, lowercased and
// sorted, so a test can assert on the resulting set.
func loadCommonThirdPartyDomains(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	partyID gid.GID,
) []string {
	t.Helper()

	var domains []string

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		rows, err := conn.Query(
			ctx,
			`SELECT lower(domain::text) FROM common_third_party_domains WHERE common_third_party_id = $1 ORDER BY lower(domain::text)`,
			partyID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var d string
			if err := rows.Scan(&d); err != nil {
				return err
			}

			domains = append(domains, d)
		}

		return rows.Err()
	}))

	return domains
}

// seedOrgThirdParty inserts an organization (in its own tenant unless an
// existing organization is reused) and one third party inside it, linked to
// the given catalog row. It returns the organization and third party ids.
func seedOrgThirdParty(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	orgID *gid.GID,
	commonThirdPartyID gid.GID,
) (gid.GID, gid.GID) {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)

	// Only mint an organization when the caller did not supply one, so reusing
	// an organization does not leave a throwaway org and tenant behind.
	var organizationID gid.GID

	if orgID != nil {
		organizationID = *orgID
	} else {
		organizationID = seedMergeTestOrganization(t, ctx, client)
	}

	tenantID := organizationID.TenantID()
	thirdPartyID := gid.New(tenantID, coredata.ThirdPartyEntityType)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(
			ctx,
			`INSERT INTO third_parties
			     (id, tenant_id, organization_id, name, category, countries, level, common_third_party_id, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, 'ANALYTICS', '{}', 1, $5, $6, $6)`,
			thirdPartyID.String(),
			tenantID.String(),
			organizationID.String(),
			"Vendor "+thirdPartyID.String(),
			commonThirdPartyID.String(),
			now,
		)

		return err
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM third_parties WHERE id = $1`, thirdPartyID.String())
			return err
		})
	})

	return organizationID, thirdPartyID
}

// loadThirdPartyCommonID reads an organization third party's catalog link.
func loadThirdPartyCommonID(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	thirdPartyID gid.GID,
) *gid.GID {
	t.Helper()

	var got *gid.GID

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return conn.QueryRow(
			ctx,
			`SELECT common_third_party_id FROM third_parties WHERE id = $1`,
			thirdPartyID,
		).Scan(&got)
	}))

	return got
}

func mergeInTx(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	winnerID, loserID gid.GID,
) thirdparty.MergeCatalogResult {
	t.Helper()

	var result thirdparty.MergeCatalogResult

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error

		result, err = thirdparty.MergeCatalog(ctx, tx, winnerID, loserID)

		return err
	}))

	return result
}

// TestMergeCatalog_MovesDomainsAndDropsCollisions pins the only
// step that can violate a constraint. common_third_party_domains is unique
// on (common_third_party_id, domain), so a domain both rows claim would
// collide on the move. The loser's mixed-case duplicate also proves the
// comparison follows the index's case-insensitive semantics.
func TestMergeCatalog_MovesDomainsAndDropsCollisions(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	winner := seedCatalogParty(t, ctx, client)
	loser := seedCatalogParty(t, ctx, client)

	shared := "shared-" + winner.Slug + ".com"
	unique := "unique-" + loser.Slug + ".fr"

	seedCommonThirdPartyDomain(t, ctx, client, winner.ID, shared)
	// Mixed case on purpose: domain is CITEXT, so this is the same domain
	// as the winner's and must be dropped rather than collide.
	seedCommonThirdPartyDomain(t, ctx, client, loser.ID, strings.ToUpper(shared))
	seedCommonThirdPartyDomain(t, ctx, client, loser.ID, unique)

	result := mergeInTx(t, ctx, client, winner.ID, loser.ID)

	assert.Equal(t, int64(1), result.DomainsMoved)
	assert.Equal(t, int64(1), result.DomainsDroppedAsDup)

	// Compared lowercased: the loader folds case because domain is CITEXT,
	// and these fixtures embed a mixed-case GID.
	assert.Equal(
		t,
		[]string{strings.ToLower(shared), strings.ToLower(unique)},
		loadCommonThirdPartyDomains(t, ctx, client, winner.ID),
	)

	assert.Empty(t, loadCommonThirdPartyDomains(t, ctx, client, loser.ID))
}

// TestMergeCatalog_RepointsPatternsPreservingConfidence pins two
// things: catalog patterns follow the merge, and the merge does not inflate
// how well they were attributed. A merge says two rows are one vendor; it
// says nothing new about any pattern's attribution, so confidence must
// survive untouched. It also proves no pattern is left with the THIRD_PARTY
// verdict and a NULL vendor, which is what the foreign key's ON DELETE
// would have produced.
func TestMergeCatalog_RepointsPatternsPreservingConfidence(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	winner := seedCatalogParty(t, ctx, client)
	loser := seedCatalogParty(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)

	weak := coredata.CommonTrackerPattern{
		ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		CommonThirdPartyID: &loser.ID,
		TrackerType:        coredata.TrackerTypeCookie,
		Pattern:            "weak-" + loser.Slug,
		MatchType:          coredata.TrackerPatternMatchTypeExact,
		Confidence:         0.5,
		Attribution:        coredata.CommonTrackerPatternAttributionThirdParty,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	strong := weak
	strong.ID = gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType)
	strong.Pattern = "strong-" + loser.Slug
	strong.Confidence = 1

	insertCatalogPattern(t, ctx, client, weak)
	insertCatalogPattern(t, ctx, client, strong)

	result := mergeInTx(t, ctx, client, winner.ID, loser.ID)

	assert.Equal(t, int64(2), result.TrackerPatternsRepointed)

	for _, tc := range []struct {
		id         gid.GID
		confidence float32
	}{
		{weak.ID, 0.5},
		{strong.ID, 1},
	} {
		stored := loadCatalogPattern(t, ctx, client, tc.id)

		require.NotNil(t, stored.CommonThirdPartyID, "pattern must not be left without a vendor")
		assert.Equal(t, winner.ID, *stored.CommonThirdPartyID)
		assert.Equal(t, coredata.CommonTrackerPatternAttributionThirdParty, stored.Attribution)
		assert.InDelta(t, tc.confidence, stored.Confidence, 0.0001, "merge must not change confidence")
	}
}

// TestMergeCatalog_RelinksOrgTrackerPatterns pins the follow-up the
// merge owes an organization that already managed a third party for the
// winner.
//
// Such an organization's patterns resolved through the loser, so they carried
// no third party of their own and surfaced the catalog entry. Once their
// catalog row names the winner they must surface the managed third party
// instead, and nothing else would do it: the import path is idempotent on
// (organization, catalog entry) and skips an organization that already holds
// the row, so the pattern would surface the catalog entry indefinitely.
func TestMergeCatalog_RelinksOrgTrackerPatterns(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	winner := seedCatalogParty(t, ctx, client)
	loser := seedCatalogParty(t, ctx, client)

	// One organization that already manages a third party for the winner.
	orgID, managed := seedOrgThirdParty(t, ctx, client, nil, winner.ID)

	now := time.Now().UTC().Truncate(time.Microsecond)
	catalogPattern := coredata.CommonTrackerPattern{
		ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		CommonThirdPartyID: &loser.ID,
		TrackerType:        coredata.TrackerTypeCookie,
		Pattern:            "relink-" + loser.Slug,
		MatchType:          coredata.TrackerPatternMatchTypeExact,
		Confidence:         0.8,
		Attribution:        coredata.CommonTrackerPatternAttributionThirdParty,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	insertCatalogPattern(t, ctx, client, catalogPattern)

	orgPattern := seedOrgTrackerPattern(t, ctx, client, orgID, catalogPattern.ID)

	require.Nil(
		t,
		loadOrgTrackerPatternThirdPartyID(t, ctx, client, orgPattern),
		"the pattern must start unlinked, surfacing the catalog entry",
	)

	result := mergeInTx(t, ctx, client, winner.ID, loser.ID)

	assert.Equal(t, int64(1), result.OrgTrackerPatternsRelinked)

	got := loadOrgTrackerPatternThirdPartyID(t, ctx, client, orgPattern)
	require.NotNil(t, got, "the pattern must now surface the managed third party")
	assert.Equal(t, managed, *got)
}

// TestMergeCatalog_RequeuesMovedPatternsForEnrichment pins that the
// patterns the merge moved are re-armed. Their descriptions were researched
// against a vendor that no longer exists, so they would otherwise keep naming
// it indefinitely.
func TestMergeCatalog_RequeuesMovedPatternsForEnrichment(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	winner := seedCatalogParty(t, ctx, client)
	loser := seedCatalogParty(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)
	cp := coredata.CommonTrackerPattern{
		ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		CommonThirdPartyID: &loser.ID,
		TrackerType:        coredata.TrackerTypeCookie,
		Pattern:            "requeue-" + loser.Slug,
		MatchType:          coredata.TrackerPatternMatchTypeExact,
		Description:        "Set by the vendor about to be folded away.",
		Confidence:         0.8,
		Attribution:        coredata.CommonTrackerPatternAttributionThirdParty,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	insertCatalogPattern(t, ctx, client, cp)

	require.Nil(
		t,
		loadCatalogPattern(t, ctx, client, cp.ID).EnrichmentRequestedAt,
		"the pattern must start off the enrichment queue",
	)

	result := mergeInTx(t, ctx, client, winner.ID, loser.ID)

	assert.Equal(t, int64(1), result.TrackerPatternsRequeued)

	stored := loadCatalogPattern(t, ctx, client, cp.ID)
	assert.NotNil(t, stored.EnrichmentRequestedAt, "a moved pattern must be re-armed for enrichment")
}

// TestMergeCatalog_RepointsThirdPartiesCrossTenant pins that the
// merge spans tenants. The catalog is global, so a row can be referenced
// from any tenant and every reference must move before the row is deleted.
func TestMergeCatalog_RepointsThirdPartiesCrossTenant(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	winner := seedCatalogParty(t, ctx, client)
	loser := seedCatalogParty(t, ctx, client)

	_, tpA := seedOrgThirdParty(t, ctx, client, nil, loser.ID)
	_, tpB := seedOrgThirdParty(t, ctx, client, nil, loser.ID)

	result := mergeInTx(t, ctx, client, winner.ID, loser.ID)

	assert.Equal(t, int64(2), result.ThirdPartiesRepointed)
	assert.Empty(t, result.ThirdPartiesSkipped)

	for _, id := range []gid.GID{tpA, tpB} {
		got := loadThirdPartyCommonID(t, ctx, client, id)
		require.NotNil(t, got)
		assert.Equal(t, winner.ID, *got)
	}
}

// TestMergeCatalog_SkipsSameOrgCollision pins the silent hazard.
//
// When one organization already links the winner, repointing its loser-linked
// third party would leave two rows in that organization pointing at the same
// catalog entry — a state no constraint forbids and that the
// organization-scoped lookup then hides permanently. The merge leaves such a
// row alone and names it, so it ends up unlinked rather than invisibly
// duplicated.
func TestMergeCatalog_SkipsSameOrgCollision(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	winner := seedCatalogParty(t, ctx, client)
	loser := seedCatalogParty(t, ctx, client)

	orgID, tpWinner := seedOrgThirdParty(t, ctx, client, nil, winner.ID)
	_, tpLoser := seedOrgThirdParty(t, ctx, client, &orgID, loser.ID)

	result := mergeInTx(t, ctx, client, winner.ID, loser.ID)

	assert.Zero(t, result.ThirdPartiesRepointed)
	assert.Equal(t, []gid.GID{tpLoser}, result.ThirdPartiesSkipped)

	stillWinner := loadThirdPartyCommonID(t, ctx, client, tpWinner)
	require.NotNil(t, stillWinner)
	assert.Equal(t, winner.ID, *stillWinner, "the pre-existing winner link must be untouched")

	// The skipped row loses its link when the loser row is deleted, which
	// is the state the operator is told to follow up on.
	assert.Nil(t, loadThirdPartyCommonID(t, ctx, client, tpLoser))
}

// TestMergeCatalog_AdoptsLoserLogoOnlyWhenWinnerHasNone pins that
// a logo is inherited but never overwritten.
func TestMergeCatalog_AdoptsLoserLogoOnlyWhenWinnerHasNone(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	t.Run("adopts when the winner has none", func(t *testing.T) {
		t.Parallel()

		winner := seedCatalogParty(t, ctx, client)
		loser := seedCatalogParty(t, ctx, client)

		loserLogo := seedFile(t, ctx, client)
		setCommonThirdPartyLogo(t, ctx, client, loser.ID, loserLogo)

		result := mergeInTx(t, ctx, client, winner.ID, loser.ID)

		assert.True(t, result.LogoAdopted)
		assert.Equal(t, &loserLogo, loadCommonThirdPartyLogo(t, ctx, client, winner.ID))
	})

	t.Run("keeps the winner's own logo", func(t *testing.T) {
		t.Parallel()

		winner := seedCatalogParty(t, ctx, client)
		loser := seedCatalogParty(t, ctx, client)

		winnerLogo := seedFile(t, ctx, client)
		loserLogo := seedFile(t, ctx, client)
		setCommonThirdPartyLogo(t, ctx, client, winner.ID, winnerLogo)
		setCommonThirdPartyLogo(t, ctx, client, loser.ID, loserLogo)

		result := mergeInTx(t, ctx, client, winner.ID, loser.ID)

		assert.False(t, result.LogoAdopted)
		assert.Equal(t, &winnerLogo, loadCommonThirdPartyLogo(t, ctx, client, winner.ID))
	})
}

// TestMergeCatalog_DeletesLoser pins that the folded row is gone.
func TestMergeCatalog_DeletesLoser(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	winner := seedCatalogParty(t, ctx, client)
	loser := seedCatalogParty(t, ctx, client)

	mergeInTx(t, ctx, client, winner.ID, loser.ID)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var gone coredata.CommonThirdParty

		err := gone.LoadByID(ctx, conn, loser.ID)
		assert.ErrorIs(t, err, coredata.ErrResourceNotFound)

		var kept coredata.CommonThirdParty

		return kept.LoadByID(ctx, conn, winner.ID)
	}))
}

// TestMergeCatalog_RejectsSelfMerge pins the guard: merging a row
// into itself would delete the row it was meant to keep.
func TestMergeCatalog_RejectsSelfMerge(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	party := seedCatalogParty(t, ctx, client)

	err := client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := thirdparty.MergeCatalog(ctx, tx, party.ID, party.ID)
		return err
	})

	require.True(t, errors.Is(err, thirdparty.ErrCannotMergeIntoSelf), "got %v", err)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var kept coredata.CommonThirdParty
		return kept.LoadByID(ctx, conn, party.ID)
	}))
}

// TestCountByCommonThirdPartyID_AggregatesReferences pins the histograms
// duplicate detection ranks merge winners on. The third-party count spans
// tenants because the catalog is global: a tenant-scoped count would
// understate how widely an entry is used.
func TestCountByCommonThirdPartyID_AggregatesReferences(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	busy := seedCatalogParty(t, ctx, client)
	quiet := seedCatalogParty(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)

	for i, pattern := range []string{"count-a", "count-b"} {
		insertCatalogPattern(t, ctx, client, coredata.CommonTrackerPattern{
			ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
			CommonThirdPartyID: &busy.ID,
			TrackerType:        coredata.TrackerTypeCookie,
			Pattern:            pattern + "-" + busy.Slug,
			MatchType:          coredata.TrackerPatternMatchTypeExact,
			MaxAgeSeconds:      &[]int{600 + i}[0],
			Confidence:         0.9,
			Attribution:        coredata.CommonTrackerPatternAttributionThirdParty,
			CreatedAt:          now,
			UpdatedAt:          now,
		})
	}

	// Two organizations in different tenants, so the count must cross them.
	seedOrgThirdParty(t, ctx, client, nil, busy.ID)
	seedOrgThirdParty(t, ctx, client, nil, busy.ID)
	seedOrgThirdParty(t, ctx, client, nil, quiet.ID)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var patterns coredata.CommonTrackerPatterns

		patternCounts, err := patterns.CountByCommonThirdPartyID(ctx, conn)
		if err != nil {
			return err
		}

		assert.Equal(t, 2, patternCounts[busy.ID])
		assert.Zero(t, patternCounts[quiet.ID])

		var parties coredata.ThirdParties

		orgCounts, err := parties.CountByCommonThirdPartyID(ctx, conn, coredata.NewNoScope())
		if err != nil {
			return err
		}

		assert.Equal(t, 2, orgCounts[busy.ID])
		assert.Equal(t, 1, orgCounts[quiet.ID])

		return nil
	}))
}

// TestLoadAllGroupedByCommonThirdPartyID_GroupsDomains pins the domain
// loader duplicate detection compares sets with.
func TestLoadAllGroupedByCommonThirdPartyID_GroupsDomains(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	party := seedCatalogParty(t, ctx, client)
	bare := seedCatalogParty(t, ctx, client)

	first := "alpha-" + party.Slug + ".com"
	second := "beta-" + party.Slug + ".com"
	seedCommonThirdPartyDomain(t, ctx, client, party.ID, first)
	seedCommonThirdPartyDomain(t, ctx, client, party.ID, second)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var domains coredata.CommonThirdPartyDomains

		byParty, err := domains.LoadAllGroupedByCommonThirdPartyID(ctx, conn)
		if err != nil {
			return err
		}

		assert.Equal(t, []string{first, second}, byParty[party.ID])
		assert.NotContains(t, byParty, bare.ID, "a party with no domains must be absent")

		return nil
	}))
}
