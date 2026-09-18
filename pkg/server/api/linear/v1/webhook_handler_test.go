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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
)

type stubWebhookQueue struct {
	deliveryID string
	body       []byte
}

func (s *stubWebhookQueue) EnqueueWebhook(
	_ context.Context,
	deliveryID string,
	envelope []byte,
) (bool, error) {
	s.deliveryID = deliveryID
	s.body = envelope

	return true, nil
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

func TestWebhookHandlerAcceptsStaleTimestamp(t *testing.T) {
	t.Parallel()

	body, err := json.Marshal(map[string]any{
		"action":           "update",
		"type":             "Issue",
		"webhookId":        "wh-1",
		"webhookTimestamp": time.Now().Add(-2 * time.Minute).UnixMilli(),
		"data":             map[string]any{"id": "issue-1"},
	})
	require.NoError(t, err)

	mac := hmac.New(sha256.New, []byte("secret"))
	_, _ = mac.Write(body)

	queue := &stubWebhookQueue{}
	handler := WebhookHandler(log.NewLogger(), queue, "secret")
	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(body))
	req.Header.Set("Linear-Signature", hex.EncodeToString(mac.Sum(nil)))
	req.Header.Set("Linear-Delivery", "delivery-1")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "delivery-1", queue.deliveryID)
	assert.Equal(t, body, queue.body)
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
