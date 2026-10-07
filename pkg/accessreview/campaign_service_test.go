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

package accessreview_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

func TestCreateCampaign_SnapshotsConnectorAccount(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC()

	var key cipher.EncryptionKey

	require.NoError(t, client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			org := &coredata.Organization{
				ID:        organizationID,
				TenantID:  tenantID,
				Name:      "Campaign snapshot",
				CreatedAt: now,
				UpdatedAt: now,
			}

			return org.Insert(ctx, tx)
		},
	))

	connectorID := gid.New(tenantID, coredata.ConnectorEntityType)

	require.NoError(t, client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			cnnctr := &coredata.Connector{
				ID:             connectorID,
				OrganizationID: organizationID,
				Provider:       coredata.ConnectorProviderGitHub,
				Protocol:       coredata.ConnectorProtocolOAuth2,
				Connection: &connector.OAuth2Connection{
					AccessToken: "test-token",
					TokenType:   "Bearer",
				},
				CreatedAt: now,
				UpdatedAt: now,
			}

			if err := cnnctr.SetSettings(&coredata.GitHubConnectorSettings{Organization: "acme"}); err != nil {
				return err
			}

			return cnnctr.Insert(ctx, tx, scope, key)
		},
	))

	account := &coredata.ConnectorAccount{
		ID:                gid.New(tenantID, coredata.ConnectorAccountEntityType),
		OrganizationID:    organizationID,
		ConnectorID:       connectorID,
		ExternalAccountID: "selected-account",
		Name:              "Selected",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	require.NoError(t, client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			_, err := account.Upsert(ctx, tx, scope)

			return err
		},
	))

	connectorSource := insertAccessSource(t, ctx, client, scope, &coredata.AccessReviewSource{
		ID:                 gid.New(tenantID, coredata.AccessReviewSourceEntityType),
		OrganizationID:     organizationID,
		ConnectorAccountID: &account.ID,
		Name:               "GitHub",
		CreatedAt:          now,
		UpdatedAt:          now,
	})

	csvSource := insertAccessSource(t, ctx, client, scope, &coredata.AccessReviewSource{
		ID:             gid.New(tenantID, coredata.AccessReviewSourceEntityType),
		OrganizationID: organizationID,
		Name:           "Uploaded CSV",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	svc := accessreview.NewService(
		client,
		key,
		nil,
		provider.NewBuiltinRegistry(),
		log.NewLogger(log.WithName("test")),
	)

	campaign, err := svc.CreateCampaign(
		ctx,
		scope,
		accessreview.CreateAccessReviewCampaignRequest{
			OrganizationID:        organizationID,
			Name:                  "Q3 review",
			AccessReviewSourceIDs: []gid.GID{connectorSource.ID},
		},
	)
	require.NoError(t, err)

	created := loadCampaignSourceByLiveSource(t, ctx, client, scope, campaign.ID, connectorSource.ID)
	require.NotNil(t, created.ConnectorID)
	assert.Equal(t, connectorID, *created.ConnectorID)
	require.NotNil(t, created.ConnectorAccountID)
	assert.Equal(t, account.ID, *created.ConnectorAccountID)

	_, err = svc.AddCampaignSource(
		ctx,
		scope,
		accessreview.AddCampaignSourceRequest{
			CampaignID:           campaign.ID,
			AccessReviewSourceID: csvSource.ID,
		},
	)
	require.NoError(t, err)

	addedCSV := loadCampaignSourceByLiveSource(t, ctx, client, scope, campaign.ID, csvSource.ID)
	assert.Nil(t, addedCSV.ConnectorID)
	assert.Nil(t, addedCSV.ConnectorAccountID)

	_, err = svc.AddCampaignSource(
		ctx,
		scope,
		accessreview.AddCampaignSourceRequest{
			CampaignID:           campaign.ID,
			AccessReviewSourceID: connectorSource.ID,
		},
	)
	require.NoError(t, err)

	addedAgain := loadCampaignSourceByLiveSource(t, ctx, client, scope, campaign.ID, connectorSource.ID)
	require.NotNil(t, addedAgain.ConnectorAccountID)
	assert.Equal(t, account.ID, *addedAgain.ConnectorAccountID)
	assert.Equal(t, created.ID, addedAgain.ID)
}

func loadCampaignSourceByLiveSource(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	scope coredata.Scoper,
	campaignID gid.GID,
	sourceID gid.GID,
) *coredata.AccessReviewCampaignSource {
	t.Helper()

	var sources coredata.AccessReviewCampaignSources

	require.NoError(t, client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return sources.LoadByCampaignID(ctx, conn, scope, campaignID)
		},
	))

	for _, source := range sources {
		if source.AccessReviewSourceID != nil && *source.AccessReviewSourceID == sourceID {
			return source
		}
	}

	t.Fatalf("campaign source for %s not found", sourceID)

	return nil
}
