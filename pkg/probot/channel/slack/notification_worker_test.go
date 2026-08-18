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

package slack

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	notificationWorkerFixture struct {
		pg             *pg.Client
		scope          *coredata.Scope
		organizationID gid.GID
		service        *NotificationService
	}

	failingDeliverySuccessHook struct{}
)

func (failingDeliverySuccessHook) OnSlackbotDeliverySuccess(
	context.Context,
	pg.Tx,
	coredata.Scoper,
	*coredata.SlackbotMessage,
) error {
	return errors.New("hook database write failed")
}

func newNotificationWorkerFixture(t *testing.T) notificationWorkerFixture {
	t.Helper()

	lockSlackbotMessageQueue(t)

	ctx := context.Background()
	pgClient := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)

	err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var message coredata.SlackbotMessage

			return message.LoadByID(
				ctx,
				conn,
				scope,
				gid.New(tenantID, coredata.SlackbotMessageEntityType),
			)
		},
	)
	if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		t.Skipf("slackbot_messages is unavailable in the test database: %v", err)
	}

	require.NoError(
		t,
		pgClient.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				_, err := conn.Exec(
					ctx,
					`
UPDATE slackbot_messages
SET
	sent_at = COALESCE(sent_at, CURRENT_TIMESTAMP),
	processing_started_at = NULL
WHERE sent_at IS NULL
	AND error IS NULL
	AND tenant_id = @tenant_id
`,
					pgx.StrictNamedArgs{"tenant_id": tenantID.String()},
				)

				return err
			},
		),
	)

	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now()
	organization := &coredata.Organization{
		ID:        organizationID,
		TenantID:  tenantID,
		Name:      "Slackbot Notification Worker Test",
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(
		t,
		pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return organization.Insert(ctx, tx)
			},
		),
	)
	t.Cleanup(
		func() {
			_ = pgClient.WithTx(
				context.Background(),
				func(ctx context.Context, tx pg.Tx) error {
					return organization.Delete(ctx, tx, organizationID)
				},
			)
		},
	)

	return notificationWorkerFixture{
		pg:             pgClient,
		scope:          scope,
		organizationID: organizationID,
		service:        NewNotificationService(pgClient),
	}
}

func (f notificationWorkerFixture) queue(t *testing.T) *coredata.SlackbotMessage {
	t.Helper()

	message, err := f.service.Queue(
		context.Background(),
		f.scope,
		QueueNotificationRequest{
			OrganizationID: f.organizationID,
			ChannelID:      "C123",
			MessageType:    "TEST_NOTIFICATION",
			Body:           map[string]any{"text": "hello"},
			Metadata:       map[string]any{},
		},
	)
	require.NoError(t, err)

	return message
}

func (f notificationWorkerFixture) load(
	t *testing.T,
	messageID gid.GID,
) *coredata.SlackbotMessage {
	t.Helper()

	message, err := f.service.GetByID(context.Background(), f.scope, messageID)
	require.NoError(t, err)

	return message
}

func newTestNotificationHandler(f notificationWorkerFixture) *notificationHandler {
	return &notificationHandler{
		pg:         f.pg,
		logger:     log.NewLogger(log.WithName("test")),
		staleAfter: 5 * time.Minute,
		retryBase:  time.Minute,
		retryMax:   time.Hour,
		now:        time.Now,
	}
}

func parkSlackbotMessage(
	t *testing.T,
	handler *notificationHandler,
	message coredata.SlackbotMessage,
) {
	t.Helper()

	now := handler.now()
	message.SentAt = &now
	message.ProcessingStartedAt = nil
	message.UpdatedAt = now
	scope := coredata.NewScopeFromObjectID(message.ID)

	require.NoError(
		t,
		handler.pg.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return message.UpdateDeliveryState(ctx, tx, scope)
			},
		),
	)
}

func claimExactSlackbotMessage(
	t *testing.T,
	handler *notificationHandler,
	id gid.GID,
) (coredata.SlackbotMessage, error) {
	t.Helper()

	for range 32 {
		claimed, err := handler.Claim(t.Context())
		if err != nil {
			return coredata.SlackbotMessage{}, err
		}

		if claimed.ID == id {
			return claimed, nil
		}

		parkSlackbotMessage(t, handler, claimed)
	}

	return coredata.SlackbotMessage{}, errors.New("did not claim expected Slackbot message")
}

func TestNotificationHandler_RetriesTransientFailureWhenDue(t *testing.T) {
	t.Parallel()

	fixture := newNotificationWorkerFixture(t)
	message := fixture.queue(t)
	now := time.Now().UTC()
	handler := newTestNotificationHandler(fixture)
	handler.now = func() time.Time { return now }
	handler.deliver = func(context.Context, *coredata.SlackbotMessage) error {
		return &APIError{
			StatusCode: http.StatusServiceUnavailable,
			Code:       "service_unavailable",
		}
	}

	claimed, err := claimExactSlackbotMessage(t, handler, message.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, claimed.AttemptCount)
	require.Error(t, handler.Process(t.Context(), claimed))

	failed := fixture.load(t, message.ID)
	assert.Nil(t, failed.Error)
	assert.NotNil(t, failed.LastError)
	require.NotNil(t, failed.NextAttemptAt)
	assert.WithinDuration(t, now.Add(time.Minute), *failed.NextAttemptAt, time.Microsecond)

	_, err = claimExactSlackbotMessage(t, handler, message.ID)
	require.ErrorIs(t, err, worker.ErrNoTask)

	now = now.Add(time.Minute)
	claimed, err = claimExactSlackbotMessage(t, handler, message.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, claimed.AttemptCount)
}

func TestNotificationHandler_PermanentFailureBecomesDeadLetter(t *testing.T) {
	t.Parallel()

	fixture := newNotificationWorkerFixture(t)
	message := fixture.queue(t)
	handler := newTestNotificationHandler(fixture)
	handler.deliver = func(context.Context, *coredata.SlackbotMessage) error {
		return &APIError{
			StatusCode: http.StatusBadRequest,
			Code:       "channel_not_found",
		}
	}

	claimed, err := claimExactSlackbotMessage(t, handler, message.ID)
	require.NoError(t, err)
	require.Error(t, handler.Process(t.Context(), claimed))

	failed := fixture.load(t, message.ID)
	assert.NotNil(t, failed.Error)
	assert.NotNil(t, failed.LastError)
	assert.Nil(t, failed.NextAttemptAt)

	_, err = claimExactSlackbotMessage(t, handler, message.ID)
	require.ErrorIs(t, err, worker.ErrNoTask)
}

func TestNotificationHandler_StaleClaimIsRecovered(t *testing.T) {
	t.Parallel()

	fixture := newNotificationWorkerFixture(t)
	message := fixture.queue(t)
	handler := newTestNotificationHandler(fixture)
	claimTime := time.Now().Add(-10 * time.Minute)
	handler.now = func() time.Time { return claimTime }

	_, err := claimExactSlackbotMessage(t, handler, message.ID)
	require.NoError(t, err)
	require.NoError(t, handler.RecoverStale(t.Context()))

	recovered := fixture.load(t, message.ID)
	assert.Nil(t, recovered.ProcessingStartedAt)
	assert.NotNil(t, recovered.NextAttemptAt)
	assert.NotNil(t, recovered.LastError)
	assert.Nil(t, recovered.Error)
}

func TestNotificationHandler_ExhaustedStaleClaimBecomesDeadLetter(t *testing.T) {
	t.Parallel()

	fixture := newNotificationWorkerFixture(t)
	message := coredata.NewSlackbotMessage(
		fixture.scope,
		fixture.organizationID,
		"TEST_NOTIFICATION",
		map[string]any{"text": "hello"},
		map[string]any{},
	)
	message.ChannelID = new("C123")
	message.MaxAttempts = 1

	require.NoError(
		t,
		fixture.pg.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return message.Insert(ctx, tx, fixture.scope)
			},
		),
	)

	handler := newTestNotificationHandler(fixture)
	handler.now = func() time.Time { return time.Now().Add(-10 * time.Minute) }

	claimed, err := claimExactSlackbotMessage(t, handler, message.ID)
	require.NoError(t, err)
	assert.Equal(t, message.ID, claimed.ID)
	require.NoError(t, handler.RecoverStale(t.Context()))

	recovered := fixture.load(t, message.ID)
	assert.Nil(t, recovered.ProcessingStartedAt)
	assert.Nil(t, recovered.NextAttemptAt)
	assert.NotNil(t, recovered.LastError)
	assert.NotNil(t, recovered.Error)
}

func TestNotificationHandler_PreservesRevisionOrderingAcrossRetry(t *testing.T) {
	t.Parallel()

	fixture := newNotificationWorkerFixture(t)
	initial := fixture.queue(t)
	firstRevision, err := fixture.service.QueueRevision(
		t.Context(),
		fixture.scope,
		initial.ID,
		map[string]any{"text": "first revision"},
		map[string]any{},
	)
	require.NoError(t, err)
	_, err = fixture.service.QueueRevision(
		t.Context(),
		fixture.scope,
		initial.ID,
		map[string]any{"text": "second revision"},
		map[string]any{},
	)
	require.NoError(t, err)

	now := time.Now().UTC()
	handler := newTestNotificationHandler(fixture)
	handler.now = func() time.Time { return now }
	handler.deliver = func(
		_ context.Context,
		message *coredata.SlackbotMessage,
	) error {
		if message.ID == initial.ID {
			message.MessageTS = new("123.456")
			return nil
		}

		return &APIError{
			StatusCode: http.StatusServiceUnavailable,
			Code:       "service_unavailable",
		}
	}

	claimed, err := claimExactSlackbotMessage(t, handler, initial.ID)
	require.NoError(t, err)
	require.NoError(t, handler.Process(t.Context(), claimed))

	claimed, err = claimExactSlackbotMessage(t, handler, firstRevision.ID)
	require.NoError(t, err)
	assert.Equal(t, firstRevision.ID, claimed.ID)
	require.Error(t, handler.Process(t.Context(), claimed))

	now = now.Add(time.Minute)
	claimed, err = claimExactSlackbotMessage(t, handler, firstRevision.ID)
	require.NoError(t, err)
	assert.Equal(t, firstRevision.ID, claimed.ID)
}

func TestNotificationHandler_SuccessHookFailureSchedulesRetry(t *testing.T) {
	t.Parallel()

	fixture := newNotificationWorkerFixture(t)
	message := fixture.queue(t)
	handler := newTestNotificationHandler(fixture)
	handler.successHook = failingDeliverySuccessHook{}
	now := time.Now().UTC()
	handler.now = func() time.Time { return now }
	deliveries := 0
	handler.deliver = func(
		_ context.Context,
		message *coredata.SlackbotMessage,
	) error {
		deliveries++
		message.MessageTS = new("123.456")

		return nil
	}

	claimed, err := claimExactSlackbotMessage(t, handler, message.ID)
	require.NoError(t, err)
	err = handler.Process(t.Context(), claimed)
	require.ErrorContains(t, err, "cannot run Slackbot delivery success hook")

	delivered := fixture.load(t, message.ID)
	assert.Nil(t, delivered.SentAt)
	assert.Nil(t, delivered.Error)
	require.NotNil(t, delivered.NextAttemptAt)
	assert.Equal(t, 1, deliveries)

	now = *delivered.NextAttemptAt
	claimed, err = claimExactSlackbotMessage(t, handler, message.ID)
	require.NoError(t, err)
	assert.Equal(t, message.ID, claimed.ID)
}
