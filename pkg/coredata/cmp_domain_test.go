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
		if err := fx.domain.SetSourcePromotionRequested(ctx, tx); err != nil {
			return err
		}

		var again coredata.CMPDomain
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

		loaded := coredata.CMPDomain{}
		if err := loaded.LoadByID(ctx, tx, fx.scope, fx.domain.ID); err != nil {
			return err
		}

		assert.NotNil(t, loaded.SourcePromotionRequestedAt)

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

		banner := &coredata.CookieBanner{ID: bannerID}
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
