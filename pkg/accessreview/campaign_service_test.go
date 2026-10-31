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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

func TestCreateCampaign_SnapshotsConnectorAccount(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)
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
				Connection: &connectorprotocol.OAuth2Connection{
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
	assert.Equal(t, account.Name, created.Name)
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
	assert.Equal(t, csvSource.Name, addedCSV.Name)
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

	sourceIDs := []gid.GID{connectorSource.ID, csvSource.ID}
	_, err = svc.UpdateCampaign(
		ctx,
		scope,
		accessreview.UpdateAccessReviewCampaignRequest{
			CampaignID:            campaign.ID,
			AccessReviewSourceIDs: &sourceIDs,
		},
	)
	require.NoError(t, err)

	merged := loadCampaignSourceByLiveSource(t, ctx, client, scope, campaign.ID, connectorSource.ID)
	assert.Equal(t, account.Name, merged.Name)

	mergedCSV := loadCampaignSourceByLiveSource(t, ctx, client, scope, campaign.ID, csvSource.ID)
	assert.Equal(t, csvSource.Name, mergedCSV.Name)

	reloaded := &coredata.AccessReviewSource{}

	require.NoError(t, client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return reloaded.LoadByID(ctx, conn, scope, connectorSource.ID)
		},
	))
	assert.Equal(t, connectorSource.Name, reloaded.Name)
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

func TestCloseCampaign_RefusesFailedSourceFetch(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			org := &coredata.Organization{
				ID:        organizationID,
				TenantID:  tenantID,
				Name:      "Failed fetch",
				CreatedAt: now,
				UpdatedAt: now,
			}

			return org.Insert(ctx, tx)
		},
	))

	source := insertAccessSource(t, ctx, client, scope, &coredata.AccessReviewSource{
		ID:             gid.New(tenantID, coredata.AccessReviewSourceEntityType),
		OrganizationID: organizationID,
		Name:           "Uploaded CSV",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	svc := accessreview.NewService(
		client,
		cipher.EncryptionKey{},
		nil,
		provider.NewBuiltinRegistry(),
		log.NewLogger(log.WithName("test")),
	)

	campaign, err := svc.CreateCampaign(
		ctx,
		scope,
		accessreview.CreateAccessReviewCampaignRequest{
			OrganizationID:        organizationID,
			Name:                  "Q4 review",
			AccessReviewSourceIDs: []gid.GID{source.ID},
		},
	)
	require.NoError(t, err)

	_, err = svc.StartCampaign(ctx, scope, campaign.ID)
	require.NoError(t, err)

	// Settle the latest attempt the way the fetch worker does.
	settle := func(campaignID gid.GID, status coredata.AccessReviewCampaignSourceFetchStatus) {
		t.Helper()

		require.NoError(t, client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				var latest coredata.AccessReviewCampaignSourceFetchAttempts
				if err := latest.LoadLatestByCampaignID(ctx, tx, scope, campaignID); err != nil {
					return err
				}

				latest[0].Status = status
				if err := latest[0].Update(ctx, tx, scope); err != nil {
					return err
				}

				loaded := &coredata.AccessReviewCampaign{}
				if err := loaded.LoadByID(ctx, tx, scope, campaignID); err != nil {
					return err
				}

				loaded.Status = coredata.AccessReviewCampaignStatusPendingActions

				return loaded.Update(ctx, tx, scope)
			},
		))
	}

	settle(campaign.ID, coredata.AccessReviewCampaignSourceFetchStatusFailed)

	_, err = svc.CloseCampaign(ctx, scope, campaign.ID)
	require.ErrorIs(t, err, accessreview.ErrCampaignSourceNotFetched)
	assert.ErrorContains(t, err, source.ID.String())

	retryReq := accessreview.RetryCampaignSourceFetchRequest{
		CampaignID:           campaign.ID,
		AccessReviewSourceID: source.ID,
	}

	retried, err := svc.RetryCampaignSourceFetch(ctx, scope, retryReq)
	require.NoError(t, err)
	assert.Equal(t, coredata.AccessReviewCampaignStatusInProgress, retried.Status)

	var latest coredata.AccessReviewCampaignSourceFetchAttempts

	require.NoError(t, client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return latest.LoadLatestByCampaignID(ctx, conn, scope, campaign.ID)
		},
	))
	require.Len(t, latest, 1)
	assert.Equal(t, coredata.AccessReviewCampaignSourceFetchStatusQueued, latest[0].Status)
	assert.Equal(t, 2, latest[0].AttemptNumber)

	_, err = svc.RetryCampaignSourceFetch(ctx, scope, retryReq)
	require.ErrorIs(t, err, accessreview.ErrCampaignSourceFetchNotFailed)

	_, err = svc.RetryCampaignSourceFetch(
		ctx,
		scope,
		accessreview.RetryCampaignSourceFetchRequest{
			CampaignID:           campaign.ID,
			AccessReviewSourceID: gid.New(tenantID, coredata.AccessReviewSourceEntityType),
		},
	)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	settle(campaign.ID, coredata.AccessReviewCampaignSourceFetchStatusSuccess)

	closed, err := svc.CloseCampaign(ctx, scope, campaign.ID)
	require.NoError(t, err)
	assert.Equal(t, coredata.AccessReviewCampaignStatusCompleted, closed.Status)

	_, err = svc.RetryCampaignSourceFetch(ctx, scope, retryReq)
	require.ErrorIs(t, err, accessreview.ErrCampaignCompleted)

	deletedSource := insertAccessSource(t, ctx, client, scope, &coredata.AccessReviewSource{
		ID:             gid.New(tenantID, coredata.AccessReviewSourceEntityType),
		OrganizationID: organizationID,
		Name:           "Deleted CSV",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	orphaned, err := svc.CreateCampaign(
		ctx,
		scope,
		accessreview.CreateAccessReviewCampaignRequest{
			OrganizationID:        organizationID,
			Name:                  "Orphaned review",
			AccessReviewSourceIDs: []gid.GID{deletedSource.ID},
		},
	)
	require.NoError(t, err)

	_, err = svc.StartCampaign(ctx, scope, orphaned.ID)
	require.NoError(t, err)
	require.NoError(t, svc.DeleteSource(ctx, scope, deletedSource.ID))
	settle(orphaned.ID, coredata.AccessReviewCampaignSourceFetchStatusFailed)

	_, err = svc.CloseCampaign(ctx, scope, orphaned.ID)
	require.ErrorIs(t, err, accessreview.ErrCampaignSourceNotFetched)
	assert.ErrorContains(t, err, `"Deleted CSV" (source deleted)`)

	orphanedReq := accessreview.RetryCampaignSourceFetchRequest{
		CampaignID:           orphaned.ID,
		AccessReviewSourceID: deletedSource.ID,
	}

	_, err = svc.RetryCampaignSourceFetch(ctx, scope, orphanedReq)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	cancelled, err := svc.CancelCampaign(ctx, scope, orphaned.ID)
	require.NoError(t, err)
	assert.Equal(t, coredata.AccessReviewCampaignStatusCancelled, cancelled.Status)

	_, err = svc.RetryCampaignSourceFetch(ctx, scope, orphanedReq)
	require.ErrorIs(t, err, accessreview.ErrCampaignCancelled)

	draft, err := svc.CreateCampaign(
		ctx,
		scope,
		accessreview.CreateAccessReviewCampaignRequest{
			OrganizationID:        organizationID,
			Name:                  "Draft review",
			AccessReviewSourceIDs: []gid.GID{source.ID},
		},
	)
	require.NoError(t, err)

	_, err = svc.RetryCampaignSourceFetch(
		ctx,
		scope,
		accessreview.RetryCampaignSourceFetchRequest{
			CampaignID:           draft.ID,
			AccessReviewSourceID: source.ID,
		},
	)
	require.ErrorIs(t, err, accessreview.ErrCampaignSourceFetchNotFailed)
}

func TestStartCampaign_RefusesSourceWithoutOrganization(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)

	githubConnectorID := env.insertConnector(t, coredata.ConnectorProviderGitHub)
	githubSource := env.insertLinkedSource(t, env.insertAccount(t, githubConnectorID, "github"), "GitHub")

	sentryConnectorID := env.insertConnector(t, coredata.ConnectorProviderSentry)
	sentrySource := env.insertLinkedSource(t, env.insertAccount(t, sentryConnectorID, "sentry"), "Sentry")

	herokuConnectorID := env.insertConnector(t, coredata.ConnectorProviderHeroku)
	herokuSource := env.insertLinkedSource(t, env.insertAccount(t, herokuConnectorID, "heroku"), "Heroku")

	campaign, err := env.svc.CreateCampaign(
		env.ctx,
		env.scope,
		accessreview.CreateAccessReviewCampaignRequest{
			OrganizationID:        env.organizationID,
			Name:                  "Q4 review",
			AccessReviewSourceIDs: []gid.GID{githubSource.ID, sentrySource.ID, herokuSource.ID},
		},
	)
	require.NoError(t, err)

	_, err = env.svc.StartCampaign(env.ctx, env.scope, campaign.ID)
	require.ErrorIs(t, err, accessreview.ErrCampaignSourceNotConfigured)
	assert.ErrorContains(t, err, fmt.Sprintf("%q (%s)", "github", githubSource.ID))
	assert.NotContains(t, err.Error(), sentrySource.ID.String())
	assert.NotContains(t, err.Error(), herokuSource.ID.String())

	err = env.client.WithTx(
		env.ctx,
		func(ctx context.Context, tx pg.Tx) error {
			cnnctr := &coredata.Connector{ID: githubConnectorID, UpdatedAt: time.Now()}
			if err := cnnctr.SetSettings(&coredata.GitHubConnectorSettings{Organization: "acme"}); err != nil {
				return err
			}

			return cnnctr.UpdateSettings(ctx, tx, env.scope)
		},
	)
	require.NoError(t, err)

	started, err := env.svc.StartCampaign(env.ctx, env.scope, campaign.ID)
	require.NoError(t, err)
	assert.Equal(t, coredata.AccessReviewCampaignStatusInProgress, started.Status)
}

func TestStartCampaign_ChecksRelinkedSourceConnector(t *testing.T) {
	t.Parallel()

	t.Run(
		"relinked to a configured connector",
		func(t *testing.T) {
			t.Parallel()

			env := newAccessSourceEnv(t)

			unsetAccountID := env.insertAccount(t, env.insertConnector(t, coredata.ConnectorProviderGitHub), "unset")
			setAccountID := env.insertAccount(
				t,
				env.insertConnector(t, coredata.ConnectorProviderGitHub, &coredata.GitHubConnectorSettings{Organization: "acme"}),
				"set",
			)
			source := env.insertLinkedSource(t, unsetAccountID, "GitHub")

			campaign, err := env.svc.CreateCampaign(
				env.ctx,
				env.scope,
				accessreview.CreateAccessReviewCampaignRequest{
					OrganizationID:        env.organizationID,
					Name:                  "Q4 review",
					AccessReviewSourceIDs: []gid.GID{source.ID},
				},
			)
			require.NoError(t, err)

			relinkSource(t, env, source, setAccountID)

			started, err := env.svc.StartCampaign(env.ctx, env.scope, campaign.ID)
			require.NoError(t, err)
			assert.Equal(t, coredata.AccessReviewCampaignStatusInProgress, started.Status)
		},
	)

	t.Run(
		"relinked to an unconfigured connector",
		func(t *testing.T) {
			t.Parallel()

			env := newAccessSourceEnv(t)

			unsetAccountID := env.insertAccount(t, env.insertConnector(t, coredata.ConnectorProviderGitHub), "unset")
			setAccountID := env.insertAccount(
				t,
				env.insertConnector(t, coredata.ConnectorProviderGitHub, &coredata.GitHubConnectorSettings{Organization: "acme"}),
				"set",
			)
			source := env.insertLinkedSource(t, setAccountID, "GitHub")

			campaign, err := env.svc.CreateCampaign(
				env.ctx,
				env.scope,
				accessreview.CreateAccessReviewCampaignRequest{
					OrganizationID:        env.organizationID,
					Name:                  "Q4 review",
					AccessReviewSourceIDs: []gid.GID{source.ID},
				},
			)
			require.NoError(t, err)

			relinkSource(t, env, source, unsetAccountID)

			_, err = env.svc.StartCampaign(env.ctx, env.scope, campaign.ID)
			require.ErrorIs(t, err, accessreview.ErrCampaignSourceNotConfigured)
			assert.ErrorContains(t, err, source.ID.String())
		},
	)
}

func relinkSource(
	t *testing.T,
	env *accessSourceEnv,
	source *coredata.AccessReviewSource,
	accountID gid.GID,
) {
	t.Helper()

	source.ConnectorAccountID = &accountID
	source.UpdatedAt = time.Now()

	err := env.client.WithTx(
		env.ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return source.Update(ctx, tx, env.scope)
		},
	)
	require.NoError(t, err)
}

func TestEntryWrites_WaitForCampaignRetry(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	svc, scope, organizationID, sourceID := newEntryWritesFixture(t, ctx, client)
	campaignID, entryIDs := insertPendingActionsCampaign(t, ctx, client, svc, scope, organizationID, sourceID, 1)

	results := make(chan error, 2)

	// Hold the campaign row as RetryCampaignSourceFetch does while it moves
	// the campaign back to IN_PROGRESS.
	require.NoError(t, client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			holderPID := backendPID(t, ctx, tx)

			loaded := &coredata.AccessReviewCampaign{ID: campaignID}
			if err := loaded.LockForUpdate(ctx, tx, scope); err != nil {
				return err
			}

			if err := loaded.LoadByID(ctx, tx, scope, campaignID); err != nil {
				return err
			}

			loaded.Status = coredata.AccessReviewCampaignStatusInProgress
			if err := loaded.Update(ctx, tx, scope); err != nil {
				return err
			}

			go func() {
				_, err := svc.RecordDecisions(
					ctx,
					scope,
					[]accessreview.RecordAccessReviewEntryDecisionRequest{
						{EntryID: entryIDs[0], Decision: coredata.AccessReviewEntryDecisionApproved},
					},
				)
				results <- err
			}()

			go func() {
				_, err := svc.FlagEntry(
					ctx,
					scope,
					accessreview.FlagAccessReviewEntryRequest{
						EntryID: entryIDs[0],
						Flags:   []coredata.AccessReviewEntryFlag{coredata.AccessReviewEntryFlagExcessive},
					},
				)
				results <- err
			}()

			requireBlockedBehind(t, ctx, client, []int32{holderPID}, 2)
			assert.Empty(t, results, "entry writes must wait while the retry holds the campaign")

			return nil
		},
	))

	for range 2 {
		assert.ErrorIs(t, <-results, accessreview.ErrCampaignNotPendingActions)
	}
}

func TestRecordDecisions_CrossCampaignBatchesDoNotDeadlock(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	svc, scope, organizationID, sourceID := newEntryWritesFixture(t, ctx, client)
	campaignA, entriesA := insertPendingActionsCampaign(t, ctx, client, svc, scope, organizationID, sourceID, 2)
	campaignB, entriesB := insertPendingActionsCampaign(t, ctx, client, svc, scope, organizationID, sourceID, 2)

	approve := func(entryIDs ...gid.GID) []accessreview.RecordAccessReviewEntryDecisionRequest {
		decisions := make([]accessreview.RecordAccessReviewEntryDecisionRequest, len(entryIDs))
		for i, id := range entryIDs {
			decisions[i] = accessreview.RecordAccessReviewEntryDecisionRequest{
				EntryID:  id,
				Decision: coredata.AccessReviewEntryDecisionApproved,
			}
		}

		return decisions
	}

	results := make(chan error, 2)

	// Queue each batch on its first campaign lock, then release A before B:
	// locking in caller order makes the two batches deadlock.
	require.NoError(t, client.WithTx(
		ctx,
		func(ctx context.Context, txB pg.Tx) error {
			holderB := backendPID(t, ctx, txB)
			if err := (&coredata.AccessReviewCampaign{ID: campaignB}).LockForUpdate(ctx, txB, scope); err != nil {
				return err
			}

			return client.WithTx(
				ctx,
				func(ctx context.Context, txA pg.Tx) error {
					holderA := backendPID(t, ctx, txA)
					if err := (&coredata.AccessReviewCampaign{ID: campaignA}).LockForUpdate(ctx, txA, scope); err != nil {
						return err
					}

					go func() {
						_, err := svc.RecordDecisions(ctx, scope, approve(entriesA[0], entriesB[0]))
						results <- err
					}()

					go func() {
						_, err := svc.RecordDecisions(ctx, scope, approve(entriesB[1], entriesA[1]))
						results <- err
					}()

					requireBlockedBehind(t, ctx, client, []int32{holderA, holderB}, 2)

					return nil
				},
			)
		},
	))

	for range 2 {
		require.NoError(t, <-results)
	}
}

func newEntryWritesFixture(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
) (*accessreview.Service, *coredata.Scope, gid.GID, gid.GID) {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			org := &coredata.Organization{
				ID:        organizationID,
				TenantID:  tenantID,
				Name:      "Retry race",
				CreatedAt: now,
				UpdatedAt: now,
			}

			return org.Insert(ctx, tx)
		},
	))

	source := insertAccessSource(t, ctx, client, scope, &coredata.AccessReviewSource{
		ID:             gid.New(tenantID, coredata.AccessReviewSourceEntityType),
		OrganizationID: organizationID,
		Name:           "Uploaded CSV",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	svc := accessreview.NewService(
		client,
		cipher.EncryptionKey{},
		nil,
		provider.NewBuiltinRegistry(),
		log.NewLogger(log.WithName("test")),
	)

	return svc, scope, organizationID, source.ID
}

func insertPendingActionsCampaign(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	svc *accessreview.Service,
	scope *coredata.Scope,
	organizationID gid.GID,
	sourceID gid.GID,
	entryCount int,
) (gid.GID, []gid.GID) {
	t.Helper()

	campaign, err := svc.CreateCampaign(
		ctx,
		scope,
		accessreview.CreateAccessReviewCampaignRequest{
			OrganizationID:        organizationID,
			Name:                  "Q4 review",
			AccessReviewSourceIDs: []gid.GID{sourceID},
		},
	)
	require.NoError(t, err)

	campaignSource := loadCampaignSourceByLiveSource(t, ctx, client, scope, campaign.ID, sourceID)
	now := time.Now().UTC()
	entryIDs := make([]gid.GID, entryCount)

	require.NoError(t, client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			for i := range entryIDs {
				entryIDs[i] = gid.New(scope.GetTenantID(), coredata.AccessReviewEntryEntityType)
				email := fmt.Sprintf("user%d@example.com", i)

				entry := &coredata.AccessReviewEntry{
					ID:                           entryIDs[i],
					OrganizationID:               organizationID,
					AccessReviewCampaignID:       campaign.ID,
					AccessReviewCampaignSourceID: campaignSource.ID,
					Email:                        email,
					FullName:                     "User",
					Roles:                        []string{"member"},
					MFAStatus:                    coredata.MFAStatusUnknown,
					AuthMethod:                   coredata.AccessReviewEntryAuthMethodUnknown,
					AccountType:                  coredata.AccessReviewEntryAccountTypeUser,
					ExternalID:                   email,
					AccountKey:                   email,
					Flags:                        []coredata.AccessReviewEntryFlag{},
					FlagReasons:                  []string{},
					Decision:                     coredata.AccessReviewEntryDecisionPending,
					CreatedAt:                    now,
					UpdatedAt:                    now,
				}
				if err := entry.Upsert(ctx, tx, scope); err != nil {
					return err
				}
			}

			loaded := &coredata.AccessReviewCampaign{}
			if err := loaded.LoadByID(ctx, tx, scope, campaign.ID); err != nil {
				return err
			}

			loaded.Status = coredata.AccessReviewCampaignStatusPendingActions

			return loaded.Update(ctx, tx, scope)
		},
	))

	return campaign.ID, entryIDs
}

func backendPID(t *testing.T, ctx context.Context, tx pg.Tx) int32 {
	t.Helper()

	var pid int32
	require.NoError(t, tx.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&pid))

	return pid
}

// requireBlockedBehind waits until want backends wait, directly or behind
// another waiter, on a lock one of the holders keeps.
func requireBlockedBehind(t *testing.T, ctx context.Context, client *pg.Client, holders []int32, want int64) {
	t.Helper()

	const q = `
WITH RECURSIVE blocked(pid) AS (
	SELECT pid FROM pg_stat_activity WHERE pg_blocking_pids(pid) && $1::int4[]
	UNION
	SELECT a.pid FROM pg_stat_activity a JOIN blocked b ON b.pid = ANY(pg_blocking_pids(a.pid))
)
SELECT count(*) FROM blocked`

	require.Eventually(
		t,
		func() bool {
			var count int64

			err := client.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					return conn.QueryRow(ctx, q, holders).Scan(&count)
				},
			)

			return err == nil && count == want
		},
		10*time.Second,
		20*time.Millisecond,
	)
}
