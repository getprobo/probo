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

package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestDoHTTPCall_SendsStableDeliveryIdentityAndSignature(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 20, 9, 30, 0, 0, time.UTC)
	tenantID := gid.NewTenantID()
	eventID := gid.New(tenantID, coredata.WebhookEventEntityType)
	subscriptionID := gid.New(tenantID, coredata.WebhookSubscriptionEntityType)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	signingSecret := "whsec_test"
	expectedTimestamp := strconv.FormatInt(now.Unix(), 10)

	var (
		receivedBody      []byte
		receivedSignature string
	)
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				var err error
				receivedBody, err = io.ReadAll(r.Body)
				require.NoError(t, err)

				assert.Equal(t, eventID.String(), r.Header.Get("Idempotency-Key"))
				assert.Equal(t, eventID.String(), r.Header.Get("X-Probo-Webhook-Delivery-Id"))
				assert.Equal(t, "user:created", r.Header.Get("X-Probo-Webhook-Event"))
				assert.Equal(t, expectedTimestamp, r.Header.Get("X-Probo-Webhook-Timestamp"))
				receivedSignature = r.Header.Get("X-Probo-Webhook-Signature")

				w.WriteHeader(http.StatusAlreadyReported)
			},
		),
	)
	t.Cleanup(server.Close)

	h := webhookHandler{
		httpClient: server.Client(),
		timeout:    time.Second,
		now:        func() time.Time { return now },
	}
	webhookData := coredata.WebhookData{
		OrganizationID: organizationID,
		EventType:      coredata.WebhookEventTypeUserCreated,
		Data:           []byte(`{"role":"ADMIN"}`),
		CreatedAt:      now,
	}

	_, deliveryErr := h.doHTTPCall(
		context.Background(),
		eventID,
		server.URL,
		&webhookData,
		subscriptionID,
		signingSecret,
	)

	require.Nil(t, deliveryErr)
	mac := hmac.New(sha256.New, []byte(signingSecret))
	_, err := mac.Write([]byte(expectedTimestamp + ":"))
	require.NoError(t, err)
	_, err = mac.Write(receivedBody)
	require.NoError(t, err)
	assert.NotEmpty(t, receivedBody)
	assert.Equal(t, hex.EncodeToString(mac.Sum(nil)), receivedSignature)
}

func TestIsRetryableStatus_ClassifiesEndpointResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		retryable  bool
	}{
		{name: "request timeout", statusCode: http.StatusRequestTimeout, retryable: true},
		{name: "too early", statusCode: http.StatusTooEarly, retryable: true},
		{name: "rate limited", statusCode: http.StatusTooManyRequests, retryable: true},
		{name: "server error", statusCode: http.StatusServiceUnavailable, retryable: true},
		{name: "bad request", statusCode: http.StatusBadRequest, retryable: false},
		{name: "not found", statusCode: http.StatusNotFound, retryable: false},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.retryable, isRetryableStatus(tt.statusCode))
			},
		)
	}
}

func TestParseRetryAfter_ParsesSecondsAndHTTPDate(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 20, 9, 30, 0, 0, time.UTC)

	delay, ok := parseRetryAfter("120", now)
	assert.True(t, ok)
	assert.Equal(t, 2*time.Minute, delay)

	delay, ok = parseRetryAfter(now.Add(5*time.Minute).Format(http.TimeFormat), now)
	assert.True(t, ok)
	assert.Equal(t, 5*time.Minute, delay)

	_, ok = parseRetryAfter("invalid", now)
	assert.False(t, ok)

	delay, ok = parseRetryAfter("9223372036854775807", now)
	assert.True(t, ok)
	assert.Equal(t, time.Duration(1<<63-1), delay)
}
