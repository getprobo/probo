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

package linear_v1

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
)

type stubWebhookQueue struct {
	deliveryID string
	body       []byte
	err        error
}

func (s *stubWebhookQueue) EnqueueWebhook(
	_ context.Context,
	deliveryID string,
	envelope []byte,
) (bool, error) {
	if s.err != nil {
		return false, s.err
	}

	s.deliveryID = deliveryID
	s.body = envelope

	return true, nil
}

func issueWebhookBody(t *testing.T) []byte {
	t.Helper()

	return issueWebhookBodyAt(t, time.Now())
}

func issueWebhookBodyAt(t *testing.T, sentAt time.Time) []byte {
	t.Helper()

	body, err := json.Marshal(
		map[string]any{
			"action":           "update",
			"type":             "Issue",
			"webhookId":        "wh-1",
			"webhookTimestamp": sentAt.UnixMilli(),
			"data":             map[string]any{"id": "issue-1"},
		},
	)
	require.NoError(t, err)

	return body
}

func signLinearBody(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)

	return hex.EncodeToString(mac.Sum(nil))
}

func TestWebhookHandlerRejectsInvalidSignature(t *testing.T) {
	t.Parallel()

	handler := WebhookHandler(log.NewLogger(), nil, "secret")
	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Linear-Signature", "nope")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestWebhookHandlerEnqueuesWithoutCheckingTimestampAge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		sentAt time.Time
	}{
		{
			name:   "recent timestamp",
			sentAt: time.Now().Add(-2 * time.Minute),
		},
		{
			name:   "older than apply window",
			sentAt: time.Now().Add(-25 * time.Hour),
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				body := issueWebhookBodyAt(t, tt.sentAt)
				queue := &stubWebhookQueue{}
				handler := WebhookHandler(log.NewLogger(), queue, "secret")
				req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(body))
				req.Header.Set("Linear-Signature", signLinearBody("secret", body))
				req.Header.Set("Linear-Delivery", "delivery-1")

				rec := httptest.NewRecorder()

				handler.ServeHTTP(rec, req)

				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, "delivery-1", queue.deliveryID)
				assert.Equal(t, body, queue.body)
			},
		)
	}
}

func TestWebhookHandlerRejectsMissingDeliveryID(t *testing.T) {
	t.Parallel()

	body, err := json.Marshal(map[string]any{
		"action":           "update",
		"type":             "Issue",
		"webhookId":        "wh-1",
		"webhookTimestamp": time.Now().UnixMilli(),
		"data":             map[string]any{"id": "issue-1"},
	})
	require.NoError(t, err)

	mac := hmac.New(sha256.New, []byte("secret"))
	_, _ = mac.Write(body)

	handler := WebhookHandler(log.NewLogger(), &stubWebhookQueue{}, "secret")
	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(body))
	req.Header.Set("Linear-Signature", hex.EncodeToString(mac.Sum(nil)))

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestWebhookHandlerRejectsWhenSecretMissing(t *testing.T) {
	t.Parallel()

	handler := WebhookHandler(log.NewLogger(), &stubWebhookQueue{}, "")
	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestWebhookHandlerRejectsOversizedBody(t *testing.T) {
	t.Parallel()

	handler := WebhookHandler(log.NewLogger(), &stubWebhookQueue{}, "secret")
	req := httptest.NewRequest(
		http.MethodPost,
		"/webhooks",
		strings.NewReader(strings.Repeat("a", maxLinearWebhookBytes+1)),
	)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}

func TestWebhookHandlerRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	body := []byte(`{"action":`)
	handler := WebhookHandler(log.NewLogger(), &stubWebhookQueue{}, "secret")
	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(body))
	req.Header.Set("Linear-Signature", signLinearBody("secret", body))
	req.Header.Set("Linear-Delivery", "delivery-1")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestWebhookHandlerRejectsEnqueueFailure(t *testing.T) {
	t.Parallel()

	body := issueWebhookBody(t)
	queue := &stubWebhookQueue{err: errors.New("inbox unavailable")}
	handler := WebhookHandler(log.NewLogger(), queue, "secret")
	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(body))
	req.Header.Set("Linear-Signature", signLinearBody("secret", body))
	req.Header.Set("Linear-Delivery", "delivery-1")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Empty(t, queue.deliveryID)
}
