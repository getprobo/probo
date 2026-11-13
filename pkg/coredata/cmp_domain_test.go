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

package coredata_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type cmpDomainFixture struct {
	scope    *coredata.Scope
	tenantID gid.TenantID
	orgID    gid.GID
	domain   coredata.CMPDomain
}

func seedCMPDomainFixture(t *testing.T, ctx context.Context, client *pg.Client, domainName string) cmpDomainFixture {
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
			Name:      "CMP Domain Test Org",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		domain.ID = gid.New(tenantID, coredata.CMPDomainEntityType)

		return domain.UpsertByOrganizationAndDomain(ctx, tx, tenantID, orgID, domainName, now)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, orgID)
			return err
		})
	})

	return cmpDomainFixture{
		scope:    scope,
		tenantID: tenantID,
		orgID:    orgID,
		domain:   domain,
	}
}

func insertCMPBanner(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx cmpDomainFixture,
	origin string,
	domainID *gid.GID,
	createdAt time.Time,
) (gid.GID, gid.GID) {
	t.Helper()

	bannerID := gid.New(fx.tenantID, coredata.CookieBannerEntityType)
	categoryID := gid.New(fx.tenantID, coredata.CookieCategoryEntityType)

	now := time.Now().UTC().Truncate(time.Microsecond)
	if createdAt.IsZero() {
		createdAt = now
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		banner := &coredata.CookieBanner{
			ID:                   bannerID,
			OrganizationID:       fx.orgID,
			CMPDomainID:          domainID,
			Name:                 "CMP Domain Test Banner",
			Origin:               origin,
			State:                coredata.CookieBannerStateActive,
			CookiePolicyURL:      origin + "/cookies",
			ConsentExpiryDays:    180,
			ShowBranding:         false,
			Capabilities:         coredata.DefaultCookieBannerCapabilities(),
			DefaultLanguage:      "en",
			PublisherCountryCode: "AA",
			CreatedAt:            createdAt,
			UpdatedAt:            now,
		}
		if err := banner.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		category := &coredata.CookieCategory{
			ID:              categoryID,
			OrganizationID:  fx.orgID,
			CookieBannerID:  bannerID,
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

	return bannerID, categoryID
}

func domainCursor() *page.Cursor[coredata.CMPDomainOrderField] {
	return page.NewCursor(
		20,
		nil,
		page.Head,
		page.OrderBy[coredata.CMPDomainOrderField]{
			Field:     coredata.CMPDomainOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
	)
}

func surfaceCursor() *page.Cursor[coredata.CMPDomainSurfaceOrderField] {
	return page.NewCursor(
		20,
		nil,
		page.Head,
		page.OrderBy[coredata.CMPDomainSurfaceOrderField]{
			Field:     coredata.CMPDomainSurfaceOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
	)
}

func TestCMPDomain_UpsertIsUniquePerOrganization(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	fx := seedCMPDomainFixture(t, ctx, client, "example.com")

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		again := coredata.CMPDomain{
			ID: gid.New(fx.tenantID, coredata.CMPDomainEntityType),
		}
		if err := again.UpsertByOrganizationAndDomain(
			ctx,
			tx,
			fx.tenantID,
			fx.orgID,
			"Example.COM",
			time.Now(),
		); err != nil {
			return err
		}

		assert.Equal(t, fx.domain.ID, again.ID)

		var matched coredata.CMPDomains
		if err := matched.LoadByOrganizationID(
			ctx,
			tx,
			fx.scope,
			fx.orgID,
			domainCursor(),
			coredata.NewCMPDomainFilter(new("EXAMPLE.com")),
		); err != nil {
			return err
		}

		require.Len(t, matched, 1)
		assert.Equal(t, fx.domain.ID, matched[0].ID)

		return nil
	}))

	err := client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		duplicate := &coredata.CMPDomain{
			ID:             gid.New(fx.tenantID, coredata.CMPDomainEntityType),
			OrganizationID: fx.orgID,
			Domain:         "example.com",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		return duplicate.Insert(ctx, tx, fx.scope)
	})
	assert.ErrorIs(t, err, coredata.ErrResourceAlreadyExists)
}

func TestCMPDomainSurface_EffectiveStateAndUpsert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	fx := seedCMPDomainFixture(t, ctx, client, "example.com")
	origin := "https://help.example.com"
	firstSeen := time.Now().UTC().Truncate(time.Microsecond)
	domainID := fx.domain.ID

	var surface coredata.CMPDomainSurface

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		surface = coredata.CMPDomainSurface{
			ID:             gid.New(fx.tenantID, coredata.CMPDomainSurfaceEntityType),
			OrganizationID: fx.orgID,
			CMPDomainID:    domainID,
			Origin:         origin,
			Status:         coredata.CMPDomainSurfaceStatusUnconfirmed,
		}

		return surface.UpsertSeen(ctx, tx, fx.scope, firstSeen)
	}))

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		loaded := coredata.CMPDomainSurface{}
		if err := loaded.LoadByID(ctx, conn, fx.scope, surface.ID); err != nil {
			return err
		}

		assert.Equal(t, coredata.CMPDomainSurfaceStatusUnconfirmed, loaded.Status)
		assert.Equal(t, coredata.CMPDomainSurfaceStateUnconfirmed, loaded.State)
		assert.True(t, loaded.LastSeenAt.Equal(firstSeen))

		return nil
	}))

	later := firstSeen.Add(time.Minute)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		again := coredata.CMPDomainSurface{
			ID:                         gid.New(fx.tenantID, coredata.CMPDomainSurfaceEntityType),
			OrganizationID:             fx.orgID,
			CMPDomainID:                domainID,
			Origin:                     origin,
			Status:                     coredata.CMPDomainSurfaceStatusIgnored,
			DiscoveredByCookieBannerID: new(gid.New(fx.tenantID, coredata.CookieBannerEntityType)),
			LastSeenAt:                 later,
		}

		return again.UpsertSeen(ctx, tx, fx.scope, later)
	}))

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		loaded := coredata.CMPDomainSurface{}
		if err := loaded.LoadByID(ctx, conn, fx.scope, surface.ID); err != nil {
			return err
		}

		assert.Equal(t, surface.ID, loaded.ID)
		assert.Equal(t, coredata.CMPDomainSurfaceStatusUnconfirmed, loaded.Status)
		assert.Nil(t, loaded.DiscoveredByCookieBannerID)
		assert.True(t, loaded.LastSeenAt.Equal(later))

		return nil
	}))

	bannerID, _ := insertCMPBanner(t, ctx, client, fx, origin, &domainID, time.Time{})

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		loaded := coredata.CMPDomainSurface{}
		if err := loaded.LoadByID(ctx, tx, fx.scope, surface.ID); err != nil {
			return err
		}

		assert.Equal(t, coredata.CMPDomainSurfaceStateCovered, loaded.State)

		if err := loaded.UpdateStatus(ctx, tx, fx.scope, coredata.CMPDomainSurfaceStatusIgnored); err != nil {
			return err
		}

		covered := coredata.CMPDomainSurface{}
		if err := covered.LoadByID(ctx, tx, fx.scope, surface.ID); err != nil {
			return err
		}

		assert.Equal(t, coredata.CMPDomainSurfaceStatusIgnored, covered.Status)
		assert.Equal(t, coredata.CMPDomainSurfaceStateCovered, covered.State)

		var surfaces coredata.CMPDomainSurfaces
		if err := surfaces.LoadByCMPDomainID(
			ctx,
			tx,
			fx.scope,
			domainID,
			surfaceCursor(),
			coredata.NewCMPDomainSurfaceFilter([]coredata.CMPDomainSurfaceState{
				coredata.CMPDomainSurfaceStateUnconfirmed,
			}),
		); err != nil {
			return err
		}

		assert.Empty(t, surfaces)

		coveredCount, err := surfaces.CountByCMPDomainID(
			ctx,
			tx,
			fx.scope,
			domainID,
			coredata.NewCMPDomainSurfaceFilter([]coredata.CMPDomainSurfaceState{
				coredata.CMPDomainSurfaceStateCovered,
			}),
		)
		if err != nil {
			return err
		}

		assert.Equal(t, 1, coveredCount)

		banner := &coredata.CookieBanner{}
		if err := banner.LoadByID(ctx, tx, fx.scope, bannerID); err != nil {
			return err
		}

		banner.State = coredata.CookieBannerStateInactive

		banner.UpdatedAt = time.Now()
		if err := banner.Update(ctx, tx, fx.scope); err != nil {
			return err
		}

		inactive := coredata.CMPDomainSurface{}
		if err := inactive.LoadByID(ctx, tx, fx.scope, surface.ID); err != nil {
			return err
		}

		assert.Equal(t, coredata.CMPDomainSurfaceStateIgnored, inactive.State)

		inactiveCovered, err := surfaces.CountByCMPDomainID(
			ctx,
			tx,
			fx.scope,
			domainID,
			coredata.NewCMPDomainSurfaceFilter([]coredata.CMPDomainSurfaceState{
				coredata.CMPDomainSurfaceStateCovered,
			}),
		)
		if err != nil {
			return err
		}

		assert.Equal(t, 0, inactiveCovered)

		if err := banner.Delete(ctx, tx, fx.scope); err != nil {
			return err
		}

		after := coredata.CMPDomainSurface{}
		if err := after.LoadByID(ctx, tx, fx.scope, surface.ID); err != nil {
			return err
		}

		assert.Equal(t, coredata.CMPDomainSurfaceStateIgnored, after.State)

		var ignored coredata.CMPDomainSurfaces
		if err := ignored.LoadByCMPDomainID(
			ctx,
			tx,
			fx.scope,
			domainID,
			surfaceCursor(),
			coredata.NewCMPDomainSurfaceFilter([]coredata.CMPDomainSurfaceState{
				coredata.CMPDomainSurfaceStateIgnored,
			}),
		); err != nil {
			return err
		}

		require.Len(t, ignored, 1)
		assert.Equal(t, surface.ID, ignored[0].ID)

		var all coredata.CMPDomainSurfaces
		if err := all.LoadByCMPDomainID(ctx, tx, fx.scope, domainID, surfaceCursor(), nil); err != nil {
			return err
		}

		require.Len(t, all, 1)

		return nil
	}))
}

func TestCookieBanner_SourcePromotionWindow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	fx := seedCMPDomainFixture(t, ctx, client, "window.example")
	bannerID, _ := insertCMPBanner(t, ctx, client, fx, "https://shop.window.example", &fx.domain.ID, time.Time{})

	earlier := time.Now().UTC().Add(-time.Hour).Truncate(time.Microsecond)
	later := time.Now().UTC().Truncate(time.Microsecond)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		banner := coredata.CookieBanner{ID: bannerID}
		if err := banner.SetSourcePromotionRequested(ctx, tx, later); err != nil {
			return err
		}
		if err := banner.SetSourcePromotionRequested(ctx, tx, earlier); err != nil {
			return err
		}

		loaded := coredata.CookieBanner{}
		if err := loaded.LoadByID(ctx, tx, fx.scope, bannerID); err != nil {
			return err
		}

		require.NotNil(t, loaded.SourcePromotionRequestedAt)
		assert.True(t, loaded.SourcePromotionRequestedAt.Equal(earlier))
		assert.Equal(t, 0, loaded.SourcePromotionAttempts)

		return nil
	}))
}

func TestCookieBanner_SourcePromotionBudget(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	fx := seedCMPDomainFixture(t, ctx, client, "budget.example")
	bannerID, _ := insertCMPBanner(t, ctx, client, fx, "https://shop.budget.example", &fx.domain.ID, time.Time{})
	requestedAt := time.Now().UTC().Add(-time.Minute).Truncate(time.Microsecond)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		banner := coredata.CookieBanner{}
		if err := banner.LoadByID(ctx, tx, fx.scope, bannerID); err != nil {
			return err
		}

		if err := banner.SetSourcePromotionRequested(ctx, tx, requestedAt); err != nil {
			return err
		}

		if err := banner.ClearSourcePromotionRequested(ctx, tx); err != nil {
			return err
		}

		assert.Equal(t, 1, banner.SourcePromotionAttempts)
		assert.Nil(t, banner.SourcePromotionRequestedAt)
		require.NotNil(t, banner.SourcePromotionSince)
		assert.True(t, banner.SourcePromotionSince.Equal(requestedAt))

		if err := banner.RequeueSourcePromotionRequested(ctx, tx); err != nil {
			return err
		}

		if err := banner.SetSourcePromotionRequested(ctx, tx, time.Now().UTC()); err != nil {
			return err
		}

		waiting := coredata.CookieBanner{}
		if err := waiting.LoadByID(ctx, tx, fx.scope, bannerID); err != nil {
			return err
		}

		assert.Equal(t, 1, waiting.SourcePromotionAttempts)
		require.NotNil(t, waiting.SourcePromotionRequestedAt)
		assert.True(t, waiting.SourcePromotionRequestedAt.Equal(requestedAt))
		require.NotNil(t, waiting.SourcePromotionSince)
		assert.True(t, waiting.SourcePromotionSince.Equal(requestedAt))

		if err := banner.ClearSourcePromotionRequested(ctx, tx); err != nil {
			return err
		}

		if err := banner.ResetSourcePromotionAttempts(ctx, tx); err != nil {
			return err
		}

		finished := coredata.CookieBanner{}
		if err := finished.LoadByID(ctx, tx, fx.scope, bannerID); err != nil {
			return err
		}

		assert.Equal(t, 0, finished.SourcePromotionAttempts)
		assert.Nil(t, finished.SourcePromotionAttemptedAt)
		require.NotNil(t, finished.SourcePromotionSince)
		assert.True(t, finished.SourcePromotionSince.Equal(requestedAt))

		if err := banner.SetSourcePromotionRequested(ctx, tx, requestedAt); err != nil {
			return err
		}

		if err := banner.ClearSourcePromotionRequested(ctx, tx); err != nil {
			return err
		}

		if err := coredata.ResetStaleSourcePromotions(ctx, tx, -time.Minute, 3); err != nil {
			return err
		}

		retried := coredata.CookieBanner{}
		if err := retried.LoadByID(ctx, tx, fx.scope, bannerID); err != nil {
			return err
		}

		require.NotNil(t, retried.SourcePromotionRequestedAt)
		assert.True(t, retried.SourcePromotionRequestedAt.Equal(requestedAt))
		require.NotNil(t, retried.SourcePromotionSince)
		assert.True(t, retried.SourcePromotionSince.Equal(requestedAt))
		assert.Equal(t, 1, retried.SourcePromotionAttempts)

		for banner.SourcePromotionAttempts < 3 {
			if err := banner.ClearSourcePromotionRequested(ctx, tx); err != nil {
				return err
			}
		}

		if err := coredata.ResetStaleSourcePromotions(ctx, tx, -time.Minute, 3); err != nil {
			return err
		}

		capped := coredata.CookieBanner{}
		if err := capped.LoadByID(ctx, tx, fx.scope, bannerID); err != nil {
			return err
		}

		assert.Nil(t, capped.SourcePromotionRequestedAt)
		assert.Equal(t, 3, capped.SourcePromotionAttempts)

		if err := banner.SetSourcePromotionRequested(ctx, tx, time.Now().UTC()); err != nil {
			return err
		}

		fresh := coredata.CookieBanner{}
		if err := fresh.LoadByID(ctx, tx, fx.scope, bannerID); err != nil {
			return err
		}

		assert.Equal(t, 0, fresh.SourcePromotionAttempts)
		assert.NotNil(t, fresh.SourcePromotionRequestedAt)

		return nil
	}))
}

func TestCookieBanner_RequeueKeepsANewerClaim(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	fx := seedCMPDomainFixture(t, ctx, client, "requeue.example")
	bannerID, _ := insertCMPBanner(t, ctx, client, fx, "https://shop.requeue.example", &fx.domain.ID, time.Time{})

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		banner := coredata.CookieBanner{}
		if err := banner.LoadByID(ctx, tx, fx.scope, bannerID); err != nil {
			return err
		}

		if err := banner.SetSourcePromotionRequested(ctx, tx, time.Now().UTC()); err != nil {
			return err
		}

		if err := banner.ClearSourcePromotionRequested(ctx, tx); err != nil {
			return err
		}

		older := banner

		if err := banner.RequeueSourcePromotionRequested(ctx, tx); err != nil {
			return err
		}

		if err := banner.ClearSourcePromotionRequested(ctx, tx); err != nil {
			return err
		}

		if err := older.RequeueSourcePromotionRequested(ctx, tx); err != nil {
			return err
		}

		if err := older.ResetSourcePromotionAttempts(ctx, tx); err != nil {
			return err
		}

		loaded := coredata.CookieBanner{}
		if err := loaded.LoadByID(ctx, tx, fx.scope, bannerID); err != nil {
			return err
		}

		assert.Nil(t, loaded.SourcePromotionRequestedAt)
		assert.Equal(t, 2, loaded.SourcePromotionAttempts)
		require.NotNil(t, loaded.SourcePromotionAttemptedAt)
		assert.True(t, loaded.SourcePromotionAttemptedAt.Equal(*banner.SourcePromotionAttemptedAt))
		require.NotNil(t, loaded.SourcePromotionSince)
		assert.True(t, loaded.SourcePromotionSince.Equal(*older.SourcePromotionSince))

		return nil
	}))
}
