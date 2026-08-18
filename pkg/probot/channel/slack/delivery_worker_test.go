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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

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

	retryAt := *retried.NextAttemptAt
	handler.now = func() time.Time { return retryAt }

	claimedAgain, err := handler.Claim(t.Context())
	require.NoError(t, err)

	claimedAgain.MaxAttempts = claimedAgain.AttemptCount
	require.Error(t, handler.Process(t.Context(), claimedAgain))
	dead := loadDeliveryOperation(t, fixture.pg, fixture.scope, operation.ID)
	require.NotNil(t, dead.DeadLetteredAt)
	assert.WithinDuration(t, retryAt, *dead.DeadLetteredAt, time.Microsecond)
	assert.Nil(t, dead.NextAttemptAt)
}

func TestDeliveryOperationHandler_PostUsesStableClientMsgID(t *testing.T) {
	t.Parallel()

	var clientMsgIDs []string

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/chat.postMessage", r.URL.Path)

				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				clientID, _ := body["client_msg_id"].(string)
				clientMsgIDs = append(clientMsgIDs, clientID)

				_, err := w.Write([]byte(`{"ok":true,"channel":"C123","ts":"123.456"}`))
				require.NoError(t, err)
			},
		),
	)
	t.Cleanup(server.Close)

	client := newTestClient(server.URL + "/api")
	handler := &deliveryOperationHandler{
		client: func(
			context.Context,
			*coredata.SlackDeliveryOperation,
		) (*Client, error) {
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
		clientMsgIDs,
	)
}

func TestDeliveryOperationHandler_AlreadyReactedIsSuccess(t *testing.T) {
	t.Parallel()

	var reactions int

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/reactions.add", r.URL.Path)

				reactions++
				_, err := w.Write([]byte(`{"ok":false,"error":"already_reacted"}`))
				require.NoError(t, err)
			},
		),
	)
	t.Cleanup(server.Close)

	client := newTestClient(server.URL + "/api")
	handler := &deliveryOperationHandler{
		client: func(
			context.Context,
			*coredata.SlackDeliveryOperation,
		) (*Client, error) {
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
	assert.Equal(t, 1, reactions)
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
