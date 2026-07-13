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

package github

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
	"go.probo.inc/probo/pkg/slug"
)

type thirdPartyFixture struct {
	scope          *coredata.Scope
	organizationID gid.GID
}

func seedThirdPartyFixture(t *testing.T, ctx context.Context, client *pg.Client) thirdPartyFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return (&coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "GitHub Discovery Third Party Test Org",
			CreatedAt: now,
			UpdatedAt: now,
		}).Insert(ctx, tx)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM third_parties WHERE organization_id = $1`, organizationID)
			if err != nil {
				return err
			}

			_, err = tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID)

			return err
		})
	})

	return thirdPartyFixture{
		scope:          scope,
		organizationID: organizationID,
	}
}

func seedCommonGitHubThirdParty(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	token string,
) coredata.CommonThirdParty {
	t.Helper()

	now := time.Now().UTC()
	name := "Github " + token
	party := coredata.CommonThirdParty{
		ID:             gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
		Name:           name,
		Slug:           slug.Make(name),
		Category:       coredata.ThirdPartyCategoryVersionControl,
		WebsiteURL:     new("https://github.com"),
		Certifications: []string{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return party.Insert(ctx, tx)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM common_third_parties WHERE id = $1`, party.ID)

			return err
		})
	})

	return party
}

func TestEnsureThirdParty_CreatesWhenMissing(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fixture := seedThirdPartyFixture(t, ctx, client)

	thirdParty, err := EnsureThirdParty(ctx, client, fixture.scope, fixture.organizationID)
	require.NoError(t, err)
	require.NotNil(t, thirdParty)
	assert.Equal(t, thirdPartyName, thirdParty.Name)
	assert.Equal(t, 1, thirdParty.Level)
	assert.Nil(t, thirdParty.ParentThirdPartyID)
	assert.Equal(t, coredata.ThirdPartyCategoryVersionControl, thirdParty.Category)
}

func TestEnsureThirdParty_FindsExistingByCaseInsensitiveName(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fixture := seedThirdPartyFixture(t, ctx, client)
	now := time.Now().UTC()

	existingID := gid.New(fixture.scope.GetTenantID(), coredata.ThirdPartyEntityType)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return (&coredata.ThirdParty{
			ID:             existingID,
			OrganizationID: fixture.organizationID,
			Name:           "github",
			Category:       coredata.ThirdPartyCategoryVersionControl,
			WebsiteURL:     new("https://github.com"),
			Level:          1,
			CreatedAt:      now,
			UpdatedAt:      now,
		}).Insert(ctx, tx, fixture.scope)
	}))

	thirdParty, err := EnsureThirdParty(ctx, client, fixture.scope, fixture.organizationID)
	require.NoError(t, err)
	require.NotNil(t, thirdParty)
	assert.Equal(t, existingID, thirdParty.ID)
	assert.Equal(t, "github", thirdParty.Name)
}

func TestCreateOrgThirdPartyFromCommon_LinksCatalogEntry(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fixture := seedThirdPartyFixture(t, ctx, client)
	token := slug.Make(gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType).String())
	commonParty := seedCommonGitHubThirdParty(t, ctx, client, token)

	var thirdParty *coredata.ThirdParty

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		tp, err := createOrgThirdPartyFromCommon(
			ctx,
			tx,
			fixture.scope,
			fixture.organizationID,
			&commonParty,
		)
		if err != nil {
			return err
		}

		thirdParty = tp

		return nil
	}))

	require.NotNil(t, thirdParty)
	require.NotNil(t, thirdParty.CommonThirdPartyID)
	assert.Equal(t, commonParty.ID, *thirdParty.CommonThirdPartyID)
	assert.Equal(t, commonParty.Name, thirdParty.Name)
	assert.Equal(t, 1, thirdParty.Level)
}
