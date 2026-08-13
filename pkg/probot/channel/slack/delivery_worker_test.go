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
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type recordingOperationClient struct {
	clientMsgIDs []string
	reactionErr  error
	reactions    int
}

func (c *recordingOperationClient) CreateMessage(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	clientMsgID string,
) (*MessageRef, error) {
	c.clientMsgIDs = append(c.clientMsgIDs, clientMsgID)

	return &MessageRef{Channel: "C123", TS: "123.456"}, nil
}

func (c *recordingOperationClient) AddReaction(
	context.Context,
	string,
	string,
	string,
) error {
	c.reactions++

	return c.reactionErr
}

func TestDeliveryService_DuplicateOperationKeyReturnsOriginal(t *testing.T) {
	t.Parallel()

	fixture := newNotificationWorkerFixture(t)
	service := NewDeliveryService(fixture.pg)
	payload := map[string]any{"channel": "C123", "text": "hello", "thread_ts": "123.456"}
	first, inserted, err := service.Queue(
		t.Context(),
		fixture.organizationID,
		"duplicate-key",
		coredata.SlackDeliveryOperationKindPostMessage,
		payload,
	)
	require.NoError(t, err)
	assert.True(t, inserted)

	second, inserted, err := service.Queue(
		t.Context(),
		fixture.organizationID,
		"duplicate-key",
		coredata.SlackDeliveryOperationKindPostMessage,
		map[string]any{"channel": "different"},
	)
	require.NoError(t, err)
	assert.False(t, inserted)
	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, first.ClientMsgID, second.ClientMsgID)
	assert.Equal(t, first.Payload, second.Payload)
}

func TestDeliveryOperationHandler_RetryAndDeadLetter(t *testing.T) {
	t.Parallel()

	fixture := newNotificationWorkerFixture(t)
	service := NewDeliveryService(fixture.pg)
	operation, _, err := service.Queue(
		t.Context(),
		fixture.organizationID,
		"retry-key",
		coredata.SlackDeliveryOperationKindAddReaction,
		map[string]any{"channel": "C123", "reaction": "eyes", "timestamp": "123.456"},
	)
	require.NoError(t, err)

	now := time.Now().UTC()
	handler := &deliveryOperationHandler{
		pg:         fixture.pg,
		logger:     log.NewLogger(log.WithName("test")),
		staleAfter: time.Minute,
		retryBase:  time.Minute,
		retryMax:   time.Hour,
		now:        func() time.Time { return now },
		deliver: func(context.Context, *coredata.SlackDeliveryOperation) error {
			return &APIError{
				StatusCode: http.StatusServiceUnavailable,
				Code:       "service_unavailable",
			}
		},
	}

	claimed, err := handler.Claim(t.Context())
	require.NoError(t, err)
	require.Error(t, handler.Process(t.Context(), claimed))
	retried := loadDeliveryOperation(t, fixture.pg, fixture.scope, operation.ID)
	require.NotNil(t, retried.NextAttemptAt)
	assert.WithinDuration(t, now.Add(time.Minute), *retried.NextAttemptAt, time.Microsecond)
	assert.Nil(t, retried.DeadLetteredAt)

	retried.MaxAttempts = retried.AttemptCount

	require.NoError(
		t,
		fixture.pg.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				retried.ProcessingStartedAt = new(now)
				retried.NextAttemptAt = nil

				return retried.UpdateDeliveryState(ctx, conn, fixture.scope)
			},
		),
	)
	require.Error(t, handler.Process(t.Context(), *retried))
	dead := loadDeliveryOperation(t, fixture.pg, fixture.scope, operation.ID)
	require.NotNil(t, dead.DeadLetteredAt)
	assert.WithinDuration(t, now, *dead.DeadLetteredAt, time.Microsecond)
	assert.Nil(t, dead.NextAttemptAt)
}

func TestDeliveryOperationHandler_PostUsesStableClientMsgID(t *testing.T) {
	t.Parallel()

	client := &recordingOperationClient{}
	handler := &deliveryOperationHandler{
		client: func(
			context.Context,
			*coredata.SlackDeliveryOperation,
		) (operationSlackClient, error) {
			return client, nil
		},
	}
	operation := &coredata.SlackDeliveryOperation{
		OperationKind: coredata.SlackDeliveryOperationKindPostMessage,
		Payload: map[string]any{
			"channel":   "C123",
			"text":      "hello",
			"thread_ts": "123.456",
		},
		ClientMsgID: new("04a8d30f-8f4d-4f41-9b36-6440cb9821d7"),
	}

	require.NoError(t, handler.deliverOperation(t.Context(), operation))
	require.NoError(t, handler.deliverOperation(t.Context(), operation))
	assert.Equal(
		t,
		[]string{
			"04a8d30f-8f4d-4f41-9b36-6440cb9821d7",
			"04a8d30f-8f4d-4f41-9b36-6440cb9821d7",
		},
		client.clientMsgIDs,
	)
}

func TestDeliveryOperationHandler_AlreadyReactedIsSuccess(t *testing.T) {
	t.Parallel()

	client := &recordingOperationClient{
		reactionErr: &APIError{StatusCode: http.StatusOK, Code: "already_reacted"},
	}
	handler := &deliveryOperationHandler{
		client: func(
			context.Context,
			*coredata.SlackDeliveryOperation,
		) (operationSlackClient, error) {
			return client, nil
		},
	}
	operation := &coredata.SlackDeliveryOperation{
		OperationKind: coredata.SlackDeliveryOperationKindAddReaction,
		Payload: map[string]any{
			"channel":   "C123",
			"reaction":  "eyes",
			"timestamp": "123.456",
		},
	}

	require.NoError(t, handler.deliverOperation(t.Context(), operation))
	assert.Equal(t, 1, client.reactions)
}

func loadDeliveryOperation(
	t *testing.T,
	pgClient *pg.Client,
	scope coredata.Scoper,
	id gid.GID,
) *coredata.SlackDeliveryOperation {
	t.Helper()

	var operation coredata.SlackDeliveryOperation

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return operation.LoadByID(ctx, conn, scope, id)
			},
		),
	)

	return &operation
}
