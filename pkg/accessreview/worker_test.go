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

package accessreview

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

func TestFetchFailureMessage(t *testing.T) {
	t.Parallel()

	t.Run("refused setting shows its own explanation", func(t *testing.T) {
		t.Parallel()

		rejected := &drivers.SettingRejectedError{
			Code:       drivers.BetterStackTeamNotFound,
			StatusCode: http.StatusUnprocessableEntity,
			Message:    "Better Stack has no team with this name for this API token.",
		}
		err := fmt.Errorf("cannot list accounts from source Better Stack / Acme: %w", rejected)

		assert.Equal(t, rejected.Message, fetchFailureMessage(err))
	})

	t.Run("anything else stays generic", func(t *testing.T) {
		t.Parallel()

		err := errors.New("cannot fetch supabase members: unexpected status 500")

		assert.Equal(t, sourceFetchFailureMessage, fetchFailureMessage(err))
	})
}

type fetchFixture struct {
	scope            *coredata.Scope
	organizationID   gid.GID
	campaignID       gid.GID
	campaignSourceID gid.GID
}

const fetchFixtureCSV = "email,full_name\nalice@example.com,Alice\n"

func seedFetchFixture(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	csvData string,
) fetchFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	fx := fetchFixture{
		scope:            coredata.NewScope(tenantID, gid.Nil),
		organizationID:   gid.New(tenantID, coredata.OrganizationEntityType),
		campaignID:       gid.New(tenantID, coredata.AccessReviewCampaignEntityType),
		campaignSourceID: gid.New(tenantID, coredata.AccessReviewCampaignSourceEntityType),
	}
	sourceID := gid.New(tenantID, coredata.AccessReviewSourceEntityType)
	now := time.Now().UTC()

	err := client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			org := &coredata.Organization{
				ID:        fx.organizationID,
				TenantID:  tenantID,
				Name:      "Fetch lifecycle",
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := org.Insert(ctx, tx); err != nil {
				return err
			}

			source := &coredata.AccessReviewSource{
				ID:             sourceID,
				OrganizationID: fx.organizationID,
				Name:           "CSV",
				CsvData:        &csvData,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if _, err := source.Insert(ctx, tx, fx.scope); err != nil {
				return err
			}

			campaign := &coredata.AccessReviewCampaign{
				ID:             fx.campaignID,
				OrganizationID: fx.organizationID,
				Name:           "Fetch lifecycle",
				Status:         coredata.AccessReviewCampaignStatusInProgress,
				StartedAt:      &now,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if err := campaign.Insert(ctx, tx, fx.scope); err != nil {
				return err
			}

			campaignSource := &coredata.AccessReviewCampaignSource{
				ID:                     fx.campaignSourceID,
				OrganizationID:         fx.organizationID,
				AccessReviewCampaignID: fx.campaignID,
				AccessReviewSourceID:   &sourceID,
				Name:                   "CSV",
				CreatedAt:              now,
				UpdatedAt:              now,
			}

			return campaignSource.Upsert(ctx, tx, fx.scope)
		},
	)
	require.NoError(t, err)

	return fx
}

func insertFetchAttempt(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx fetchFixture,
	status coredata.AccessReviewCampaignSourceFetchStatus,
	updatedAt time.Time,
) *coredata.AccessReviewCampaignSourceFetchAttempt {
	t.Helper()

	attempt := &coredata.AccessReviewCampaignSourceFetchAttempt{
		ID:                           gid.New(fx.scope.GetTenantID(), coredata.AccessReviewCampaignSourceFetchAttemptEntityType),
		OrganizationID:               fx.organizationID,
		TenantID:                     fx.scope.GetTenantID(),
		AccessReviewCampaignSourceID: fx.campaignSourceID,
		Status:                       status,
		CreatedAt:                    updatedAt,
		UpdatedAt:                    updatedAt,
	}

	err := client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return attempt.Insert(ctx, tx, fx.scope)
		},
	)
	require.NoError(t, err)

	return attempt
}

func loadLatestFetchAttempt(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx fetchFixture,
) *coredata.AccessReviewCampaignSourceFetchAttempt {
	t.Helper()

	var latest coredata.AccessReviewCampaignSourceFetchAttempts

	err := client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return latest.LoadLatestByCampaignID(ctx, conn, fx.scope, fx.campaignID)
		},
	)
	require.NoError(t, err)
	require.Len(t, latest, 1)

	return latest[0]
}

func loadFetchCampaignStatus(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx fetchFixture,
) coredata.AccessReviewCampaignStatus {
	t.Helper()

	campaign := &coredata.AccessReviewCampaign{}

	err := client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return campaign.LoadByID(ctx, conn, fx.scope, fx.campaignID)
		},
	)
	require.NoError(t, err)

	return campaign.Status
}

func TestSourceFetchHandler_RecoverStale(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	// The sweep is cross-tenant: an old date and a threshold right after it
	// keep it away from other tests' rows, which are newer or locked.
	staleAt := time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)

	capped := seedFetchFixture(t, ctx, client, fetchFixtureCSV)
	for range maxSourceFetchAttempts - 1 {
		insertFetchAttempt(t, ctx, client, capped, coredata.AccessReviewCampaignSourceFetchStatusFailed, staleAt)
	}

	insertFetchAttempt(t, ctx, client, capped, coredata.AccessReviewCampaignSourceFetchStatusFetching, staleAt)

	retried := seedFetchFixture(t, ctx, client, fetchFixtureCSV)
	insertFetchAttempt(t, ctx, client, retried, coredata.AccessReviewCampaignSourceFetchStatusFetching, staleAt)

	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	// Crashed mid-fetch, then the campaign was cancelled. Cancel leaves a
	// FETCHING attempt alone, so inserting it after the cancel gives the same
	// rows without a window where another sweep could retry it.
	cancelled := seedFetchFixture(t, ctx, client, fetchFixtureCSV)

	_, err := svc.CancelCampaign(ctx, cancelled.scope, cancelled.campaignID)
	require.NoError(t, err)

	insertFetchAttempt(t, ctx, client, cancelled, coredata.AccessReviewCampaignSourceFetchStatusFetching, staleAt)

	h := &sourceFetchHandler{
		pg:         client,
		logger:     logger,
		staleAfter: time.Since(staleAt.Add(time.Hour)),
	}

	require.NoError(t, h.RecoverStale(ctx))

	cappedAttempt := loadLatestFetchAttempt(t, ctx, client, capped)
	assert.Equal(t, maxSourceFetchAttempts, cappedAttempt.AttemptNumber, "an attempt at the cap must not be retried")
	assert.Equal(t, coredata.AccessReviewCampaignSourceFetchStatusFailed, cappedAttempt.Status)
	assert.Equal(t, coredata.AccessReviewCampaignStatusPendingActions, loadFetchCampaignStatus(t, ctx, client, capped))

	retriedAttempt := loadLatestFetchAttempt(t, ctx, client, retried)
	assert.Equal(t, 2, retriedAttempt.AttemptNumber)
	assert.Equal(t, coredata.AccessReviewCampaignSourceFetchStatusQueued, retriedAttempt.Status)
	assert.Equal(t, coredata.AccessReviewCampaignStatusInProgress, loadFetchCampaignStatus(t, ctx, client, retried))

	cancelledAttempt := loadLatestFetchAttempt(t, ctx, client, cancelled)
	assert.Equal(t, 1, cancelledAttempt.AttemptNumber, "a cancelled campaign must not get a retry")
	assert.Equal(t, coredata.AccessReviewCampaignSourceFetchStatusFailed, cancelledAttempt.Status)
}

func TestCancelCampaign_StopsSourceFetches(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)
	h := &sourceFetchHandler{svc: svc, pg: client, logger: logger}

	fx := seedFetchFixture(t, ctx, client, fetchFixtureCSV)
	insertFetchAttempt(t, ctx, client, fx, coredata.AccessReviewCampaignSourceFetchStatusQueued, time.Now())

	_, err := svc.CancelCampaign(ctx, fx.scope, fx.campaignID)
	require.NoError(t, err)

	queued := loadLatestFetchAttempt(t, ctx, client, fx)
	assert.Equal(t, coredata.AccessReviewCampaignSourceFetchStatusFailed, queued.Status)
	require.NotNil(t, queued.Error)
	assert.Equal(t, campaignNotInProgressFetchMessage, *queued.Error)

	// An attempt the worker claimed just before the cancel committed. Its CSV
	// has no email column, so a driver call would fail with another message.
	unreadable := seedFetchFixture(t, ctx, client, "full_name\nAlice\n")

	_, err = svc.CancelCampaign(ctx, unreadable.scope, unreadable.campaignID)
	require.NoError(t, err)

	claimed := insertFetchAttempt(t, ctx, client, unreadable, coredata.AccessReviewCampaignSourceFetchStatusFetching, time.Now())
	require.NoError(t, h.handle(ctx, claimed))

	skipped := loadLatestFetchAttempt(t, ctx, client, unreadable)
	assert.Equal(t, coredata.AccessReviewCampaignSourceFetchStatusFailed, skipped.Status)
	require.NotNil(t, skipped.Error)
	assert.Equal(t, campaignNotInProgressFetchMessage, *skipped.Error, "the driver must not be called")

	// A fetch whose driver call was in flight while a cancel was committing.
	inFlight := seedFetchFixture(t, ctx, client, fetchFixtureCSV)

	inProgress, err := svc.GetCampaign(ctx, inFlight.scope, inFlight.campaignID)
	require.NoError(t, err)

	campaignSource := &coredata.AccessReviewCampaignSource{}
	require.NoError(t, h.loadCampaignSource(ctx, inFlight.scope, inFlight.campaignSourceID, campaignSource))

	var (
		cancelPID int
		locked    = make(chan error, 1)
		releaseCh = make(chan struct{})
		release   = sync.OnceFunc(func() { close(releaseCh) })
		committed = make(chan error, 1)
		fetched   = make(chan error, 1)
	)

	cancelling := *inProgress
	cancelling.Status = coredata.AccessReviewCampaignStatusCancelled

	// Holds an uncommitted cancel the way CancelCampaign does.
	holdCancel := func(ctx context.Context, tx pg.Tx) error {
		if err := tx.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&cancelPID); err != nil {
			return err
		}

		if err := lockCampaignForUpdate(ctx, tx, inFlight.scope, inFlight.campaignID); err != nil {
			return err
		}

		return cancelling.Update(ctx, tx, inFlight.scope)
	}

	go func() {
		committed <- client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				if err := holdCancel(ctx, tx); err != nil {
					locked <- err

					return err
				}

				locked <- nil

				<-releaseCh

				return nil
			},
		)
	}()

	require.NoError(t, <-locked)
	t.Cleanup(release)

	go func() {
		_, err := svc.FetchSource(ctx, inFlight.scope, inProgress, campaignSource)
		fetched <- err
	}()

	require.Eventually(
		t,
		func() bool {
			if len(fetched) > 0 {
				return true
			}

			var waiting bool

			err := client.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					return conn.QueryRow(
						ctx,
						"SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE $1 = ANY (pg_blocking_pids(pid)))",
						cancelPID,
					).Scan(&waiting)
				},
			)

			return err == nil && waiting
		},
		10*time.Second,
		10*time.Millisecond,
		"the fetch must either wait on the cancel or return",
	)

	release()
	require.NoError(t, <-committed)
	require.ErrorIs(t, <-fetched, errCampaignNotInProgress)

	var (
		entries coredata.AccessReviewEntries
		count   int
	)

	err = client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error

			count, err = entries.CountByCampaignID(ctx, conn, inFlight.scope, inFlight.campaignID, &coredata.AccessReviewEntryFilter{})

			return err
		},
	)
	require.NoError(t, err)
	assert.Zero(t, count, "a cancelled campaign must not receive entries")
}
