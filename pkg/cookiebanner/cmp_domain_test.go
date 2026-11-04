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

package cookiebanner

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestCookieDomainCoversHost(t *testing.T) {
	t.Parallel()

	assert.True(t, cookieDomainCoversHost("example.com", "app.example.com"))
	assert.True(t, cookieDomainCoversHost(".example.com", "example.com"))
	assert.True(t, cookieDomainCoversHost("example.com", "www.example.com"))
	assert.False(t, cookieDomainCoversHost("www.example.com", "app.example.com"))
	assert.False(t, cookieDomainCoversHost("com", "example.com"))
	assert.False(t, cookieDomainCoversHost("vercel.app", "foo.vercel.app"))
	assert.False(t, cookieDomainCoversHost("", "example.com"))
}

func TestBestSourceKeepsSiblingProvenance(t *testing.T) {
	t.Parallel()

	script := coredata.CookieSourceScript
	httpSource := coredata.CookieSourceHTTP
	scriptBanner := gid.New(gid.NewTenantID(), coredata.CookieBannerEntityType)
	httpBanner := gid.New(gid.NewTenantID(), coredata.CookieBannerEntityType)

	source, provenance := bestSource([]*coredata.TrackerPattern{
		{Source: &httpSource, SourceCookieBannerID: &httpBanner},
		{Source: &script, SourceCookieBannerID: &scriptBanner},
	})

	require.NotNil(t, source)
	require.NotNil(t, provenance)
	assert.Equal(t, coredata.CookieSourceScript, *source)
	assert.Equal(t, scriptBanner, *provenance)
}

func TestCreateCookieBanner_AttachesCMPDomain(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)
	orgID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        orgID,
			TenantID:  tenantID,
			Name:      "CMP Domain Create Org",
			CreatedAt: now,
			UpdatedAt: now,
		}

		return org.Insert(ctx, tx)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, orgID)
			return err
		})
	})

	svc := NewService(client, false, 0)
	banner, err := svc.CreateCookieBanner(ctx, scope, CreateCookieBannerRequest{
		OrganizationID:    orgID,
		Name:              "Shop",
		Origin:            "https://www.shop.example.com",
		CookiePolicyURL:   "https://shop.example.com/cookies",
		ConsentExpiryDays: 180,
	})
	require.NoError(t, err)
	require.NotNil(t, banner.CMPDomainID)
	assert.Equal(t, "https://shop.example.com", banner.Origin)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var domain coredata.CMPDomain
		if err := domain.LoadByID(ctx, conn, scope, *banner.CMPDomainID); err != nil {
			return err
		}

		assert.Equal(t, "example.com", domain.Domain)
		assert.NotNil(t, domain.SourcePromotionRequestedAt)

		return nil
	}))
}

func TestPromoteSiblingCookieSources(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)

	t.Run("promotes a parent-domain cookie and requeues mapping", func(t *testing.T) {
		t.Parallel()

		fx := seedSiblingBanners(t, ctx, client, "example.com")
		patternB := insertSiblingPattern(t, ctx, client, fx, fx.bannerB, "_ga", coredata.CookieSourcePreExisting)
		insertSiblingDetection(
			t, ctx, client, fx, fx.bannerA, patternIDFor(t, ctx, client, fx, fx.bannerA, "_ga", coredata.CookieSourceScript),
			"_ga", coredata.CookieSourceScript, "example.com", false,
		)
		insertSiblingDetection(
			t, ctx, client, fx, fx.bannerB, patternB,
			"_ga", coredata.CookieSourcePreExisting, "example.com", false,
		)

		promoteDomain(t, ctx, client, fx)

		pattern := loadPromotedPattern(t, ctx, client, fx.scope, patternB)
		require.NotNil(t, pattern.Source)
		assert.Equal(t, coredata.CookieSourceScript, *pattern.Source)
		require.NotNil(t, pattern.SourceCookieBannerID)
		assert.Equal(t, fx.bannerA.ID, *pattern.SourceCookieBannerID)
		assert.NotNil(t, pattern.MappingRequestedAt)
	})

	t.Run("does not promote a cookie domain that misses the host", func(t *testing.T) {
		t.Parallel()

		fx := seedSiblingBanners(t, ctx, client, "example.com")
		patternA := insertSiblingPattern(t, ctx, client, fx, fx.bannerA, "_ga", coredata.CookieSourceScript)
		patternB := insertSiblingPattern(t, ctx, client, fx, fx.bannerB, "_ga", coredata.CookieSourcePreExisting)
		insertSiblingDetection(
			t, ctx, client, fx, fx.bannerA, patternA,
			"_ga", coredata.CookieSourceScript, "www.example.com", false,
		)
		insertSiblingDetection(
			t, ctx, client, fx, fx.bannerB, patternB,
			"_ga", coredata.CookieSourcePreExisting, "www.example.com", false,
		)

		promoteDomain(t, ctx, client, fx)

		pattern := loadPromotedPattern(t, ctx, client, fx.scope, patternB)
		require.NotNil(t, pattern.Source)
		assert.Equal(t, coredata.CookieSourcePreExisting, *pattern.Source)
		assert.Nil(t, pattern.SourceCookieBannerID)
		assert.Nil(t, pattern.MappingRequestedAt)
	})

	t.Run("own script detection clears sibling provenance", func(t *testing.T) {
		t.Parallel()

		fx := seedSiblingBanners(t, ctx, client, "example.com")
		patternA := insertSiblingPattern(t, ctx, client, fx, fx.bannerA, "_ga", coredata.CookieSourceHTTP)
		patternB := insertSiblingPattern(t, ctx, client, fx, fx.bannerB, "_ga", coredata.CookieSourcePreExisting)
		insertSiblingDetection(
			t, ctx, client, fx, fx.bannerA, patternA,
			"_ga", coredata.CookieSourceHTTP, "example.com", false,
		)
		insertSiblingDetection(
			t, ctx, client, fx, fx.bannerB, patternB,
			"_ga", coredata.CookieSourcePreExisting, "example.com", false,
		)
		insertUncategorised(t, ctx, client, fx, fx.bannerB.ID)

		promoteDomain(t, ctx, client, fx)

		hostOnly := false
		domain := "example.com"
		svc := NewService(client, false, 0)
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.bannerB.ID, ReportDetectedTrackersRequest{
			Cookies: []DetectedCookie{{
				Name:         "_ga",
				Source:       coredata.CookieSourceScript,
				CookieDomain: &domain,
				HostOnly:     &hostOnly,
			}},
		}))

		pattern := loadPromotedPattern(t, ctx, client, fx.scope, patternB)
		require.NotNil(t, pattern.Source)
		assert.Equal(t, coredata.CookieSourceScript, *pattern.Source)
		assert.Nil(t, pattern.SourceCookieBannerID)
	})

	t.Run("does not promote when the banner has a host-only detection", func(t *testing.T) {
		t.Parallel()

		fx := seedSiblingBanners(t, ctx, client, "example.com")
		patternA := insertSiblingPattern(t, ctx, client, fx, fx.bannerA, "_ga", coredata.CookieSourceScript)
		patternB := insertSiblingPattern(t, ctx, client, fx, fx.bannerB, "_ga", coredata.CookieSourcePreExisting)
		insertSiblingDetection(
			t, ctx, client, fx, fx.bannerA, patternA,
			"_ga", coredata.CookieSourceScript, "example.com", false,
		)
		insertSiblingDetection(
			t, ctx, client, fx, fx.bannerB, patternB,
			"_ga", coredata.CookieSourcePreExisting, "example.com", false,
		)
		insertSiblingDetection(
			t, ctx, client, fx, fx.bannerB, patternB,
			"_ga", coredata.CookieSourcePreExisting, "app.example.com", true,
		)

		promoteDomain(t, ctx, client, fx)

		assertPatternStaysPreExisting(t, ctx, client, fx.scope, patternB)
	})

	t.Run("does not promote from an excluded sibling pattern", func(t *testing.T) {
		t.Parallel()

		fx := seedSiblingBanners(t, ctx, client, "example.com")
		patternA := insertSiblingPattern(t, ctx, client, fx, fx.bannerA, "_ga", coredata.CookieSourceScript)
		patternB := insertSiblingPattern(t, ctx, client, fx, fx.bannerB, "_ga", coredata.CookieSourcePreExisting)
		excludePattern(t, ctx, client, fx.scope, patternA)
		insertSiblingDetection(
			t, ctx, client, fx, fx.bannerA, patternA,
			"_ga", coredata.CookieSourceScript, "example.com", false,
		)
		insertSiblingDetection(
			t, ctx, client, fx, fx.bannerB, patternB,
			"_ga", coredata.CookieSourcePreExisting, "example.com", false,
		)

		promoteDomain(t, ctx, client, fx)

		assertPatternStaysPreExisting(t, ctx, client, fx.scope, patternB)
	})

	t.Run("does not promote from an extension detection", func(t *testing.T) {
		t.Parallel()

		fx := seedSiblingBanners(t, ctx, client, "example.com")
		patternA := insertSiblingPattern(t, ctx, client, fx, fx.bannerA, "_ga", coredata.CookieSourceExtension)
		patternB := insertSiblingPattern(t, ctx, client, fx, fx.bannerB, "_ga", coredata.CookieSourcePreExisting)
		insertSiblingDetection(
			t, ctx, client, fx, fx.bannerA, patternA,
			"_ga", coredata.CookieSourceExtension, "example.com", false,
		)
		insertSiblingDetection(
			t, ctx, client, fx, fx.bannerB, patternB,
			"_ga", coredata.CookieSourcePreExisting, "example.com", false,
		)

		promoteDomain(t, ctx, client, fx)

		assertPatternStaysPreExisting(t, ctx, client, fx.scope, patternB)
	})

	t.Run("does not promote a pattern from another tenant", func(t *testing.T) {
		t.Parallel()

		fx := seedSiblingBanners(t, ctx, client, "example.com")
		other := seedSiblingBanners(t, ctx, client, "example.com")
		patternA := insertSiblingPattern(t, ctx, client, other, other.bannerA, "_ga", coredata.CookieSourceScript)
		patternB := insertSiblingPattern(t, ctx, client, fx, fx.bannerB, "_ga", coredata.CookieSourcePreExisting)
		insertSiblingDetection(
			t, ctx, client, other, other.bannerA, patternA,
			"_ga", coredata.CookieSourceScript, "example.com", false,
		)
		insertSiblingDetection(
			t, ctx, client, fx, fx.bannerB, patternB,
			"_ga", coredata.CookieSourcePreExisting, "example.com", false,
		)

		promoteDomain(t, ctx, client, fx)

		assertPatternStaysPreExisting(t, ctx, client, fx.scope, patternB)
	})

	t.Run("does not promote through a corsless banner", func(t *testing.T) {
		t.Parallel()

		sibling := seedSiblingBanners(t, ctx, client, "example.com")
		patternA := insertSiblingPattern(t, ctx, client, sibling, sibling.bannerA, "_ga", coredata.CookieSourceScript)
		patternB := insertSiblingPattern(t, ctx, client, sibling, sibling.bannerB, "_ga", coredata.CookieSourcePreExisting)
		insertSiblingDetection(
			t, ctx, client, sibling, sibling.bannerA, patternA,
			"_ga", coredata.CookieSourceScript, "example.com", false,
		)
		insertSiblingDetection(
			t, ctx, client, sibling, sibling.bannerB, patternB,
			"_ga", coredata.CookieSourcePreExisting, "example.com", false,
		)
		setCorsless(t, ctx, client, sibling.scope, &sibling.bannerA)

		promoteDomain(t, ctx, client, sibling)

		assertPatternStaysPreExisting(t, ctx, client, sibling.scope, patternB)

		owner := seedSiblingBanners(t, ctx, client, "example.com")
		ownerPatternA := insertSiblingPattern(t, ctx, client, owner, owner.bannerA, "_ga", coredata.CookieSourceScript)
		ownerPatternB := insertSiblingPattern(t, ctx, client, owner, owner.bannerB, "_ga", coredata.CookieSourcePreExisting)
		insertSiblingDetection(
			t, ctx, client, owner, owner.bannerA, ownerPatternA,
			"_ga", coredata.CookieSourceScript, "example.com", false,
		)
		insertSiblingDetection(
			t, ctx, client, owner, owner.bannerB, ownerPatternB,
			"_ga", coredata.CookieSourcePreExisting, "example.com", false,
		)
		setCorsless(t, ctx, client, owner.scope, &owner.bannerB)

		promoteDomain(t, ctx, client, owner)

		assertPatternStaysPreExisting(t, ctx, client, owner.scope, ownerPatternB)
	})
}

type siblingBanners struct {
	scope    *coredata.Scope
	tenantID gid.TenantID
	orgID    gid.GID
	domain   coredata.CMPDomain
	bannerA  coredata.CookieBanner
	bannerB  coredata.CookieBanner
	category map[gid.GID]gid.GID
}

func seedSiblingBanners(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	domainName string,
) siblingBanners {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)
	orgID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC().Truncate(time.Microsecond)

	var domain coredata.CMPDomain

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        orgID,
			TenantID:  tenantID,
			Name:      "Sibling Banner Org",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		return domain.UpsertByOrganizationAndDomain(ctx, tx, tenantID, orgID, domainName, now)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, orgID)
			return err
		})
	})

	fx := siblingBanners{
		scope:    scope,
		tenantID: tenantID,
		orgID:    orgID,
		domain:   domain,
		category: map[gid.GID]gid.GID{},
	}
	fx.bannerA = insertServiceBanner(t, ctx, client, fx, "https://example.com")
	fx.bannerB = insertServiceBanner(t, ctx, client, fx, "https://app.example.com")

	return fx
}

func insertServiceBanner(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx siblingBanners,
	origin string,
) coredata.CookieBanner {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	banner := coredata.CookieBanner{
		ID:                   gid.New(fx.tenantID, coredata.CookieBannerEntityType),
		OrganizationID:       fx.orgID,
		CMPDomainID:          &fx.domain.ID,
		Name:                 origin,
		Origin:               origin,
		State:                coredata.CookieBannerStateActive,
		CookiePolicyURL:      origin + "/cookies",
		ConsentExpiryDays:    180,
		ShowBranding:         false,
		Capabilities:         coredata.DefaultCookieBannerCapabilities(),
		DefaultLanguage:      "en",
		PublisherCountryCode: "AA",
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	categoryID := gid.New(fx.tenantID, coredata.CookieCategoryEntityType)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := banner.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		category := &coredata.CookieCategory{
			ID:              categoryID,
			OrganizationID:  fx.orgID,
			CookieBannerID:  banner.ID,
			Name:            "Analytics",
			Slug:            "analytics",
			Description:     "",
			Kind:            coredata.CookieCategoryKindNormal,
			Rank:            1,
			GCMConsentTypes: []string{},
			TCFPurposeIDs:   []int{},
			PostHogConsent:  false,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		return category.Insert(ctx, tx, fx.scope)
	}))

	fx.category[banner.ID] = categoryID

	return banner
}

func insertSiblingPattern(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx siblingBanners,
	banner coredata.CookieBanner,
	name string,
	source coredata.CookieSource,
) gid.GID {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	patternID := gid.New(fx.tenantID, coredata.TrackerPatternEntityType)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		pattern := &coredata.TrackerPattern{
			ID:               patternID,
			OrganizationID:   fx.orgID,
			CookieBannerID:   banner.ID,
			CookieCategoryID: fx.category[banner.ID],
			TrackerType:      coredata.TrackerTypeCookie,
			Pattern:          name,
			MatchType:        coredata.TrackerPatternMatchTypeExact,
			Description:      "",
			Source:           &source,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		return pattern.Insert(ctx, tx, fx.scope)
	}))

	return patternID
}

func insertSiblingDetection(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx siblingBanners,
	banner coredata.CookieBanner,
	patternID gid.GID,
	name string,
	source coredata.CookieSource,
	cookieDomain string,
	hostOnly bool,
) {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		detection := &coredata.DetectedTracker{
			ID:               gid.New(fx.tenantID, coredata.DetectedTrackerEntityType),
			CookieBannerID:   banner.ID,
			TrackerPatternID: &patternID,
			TrackerType:      coredata.TrackerTypeCookie,
			Identifier:       name,
			Source:           &source,
			ObservedBy:       coredata.DetectedTrackerObserverScanner,
			CookieDomain:     &cookieDomain,
			HostOnly:         &hostOnly,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		_, err := detection.InsertIfNotExists(ctx, tx, fx.scope)

		return err
	}))
}

func insertUncategorised(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx siblingBanners,
	bannerID gid.GID,
) {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		category := &coredata.CookieCategory{
			ID:              gid.New(fx.tenantID, coredata.CookieCategoryEntityType),
			OrganizationID:  fx.orgID,
			CookieBannerID:  bannerID,
			Name:            "Uncategorised",
			Slug:            "uncategorised",
			Description:     "",
			Kind:            coredata.CookieCategoryKindUncategorised,
			Rank:            0,
			GCMConsentTypes: []string{},
			TCFPurposeIDs:   []int{},
			PostHogConsent:  false,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		return category.Insert(ctx, tx, fx.scope)
	}))
}

func patternIDFor(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx siblingBanners,
	banner coredata.CookieBanner,
	name string,
	source coredata.CookieSource,
) gid.GID {
	t.Helper()

	return insertSiblingPattern(t, ctx, client, fx, banner, name, source)
}

func setCorsless(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	scope *coredata.Scope,
	banner *coredata.CookieBanner,
) {
	t.Helper()

	banner.Capabilities.Corsless = true
	banner.UpdatedAt = time.Now()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return banner.Update(ctx, tx, scope)
	}))
}

func promoteDomain(t *testing.T, ctx context.Context, client *pg.Client, fx siblingBanners) {
	t.Helper()

	svc := NewService(client, false, 0)
	handler := &cmpDomainHandler{
		svc:    svc,
		pg:     client,
		logger: log.NewLogger(log.WithOutput(io.Discard)),
	}

	require.NoError(t, handler.Process(ctx, cmpDomainTask{domain: &fx.domain}))
}

func excludePattern(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	scope *coredata.Scope,
	patternID gid.GID,
) {
	t.Helper()

	pattern := loadPromotedPattern(t, ctx, client, scope, patternID)
	pattern.Excluded = true

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return pattern.Update(ctx, tx, scope)
	}))
}

func assertPatternStaysPreExisting(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	scope *coredata.Scope,
	patternID gid.GID,
) {
	t.Helper()

	pattern := loadPromotedPattern(t, ctx, client, scope, patternID)
	require.NotNil(t, pattern.Source)
	assert.Equal(t, coredata.CookieSourcePreExisting, *pattern.Source)
	assert.Nil(t, pattern.SourceCookieBannerID)
	assert.Nil(t, pattern.MappingRequestedAt)
}

func loadPromotedPattern(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	scope *coredata.Scope,
	patternID gid.GID,
) coredata.TrackerPattern {
	t.Helper()

	var pattern coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return pattern.LoadByID(ctx, conn, scope, patternID)
	}))

	return pattern
}
