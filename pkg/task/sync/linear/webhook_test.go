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

package linear

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifySignature(t *testing.T) {
	t.Parallel()

	body := []byte(`{"action":"update","type":"Issue"}`)
	secret := "whsec"
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	assert.True(t, VerifySignature(secret, signature, body))
	assert.False(t, VerifySignature(secret, "deadbeef", body))
	assert.False(t, VerifySignature("", signature, body))
}

func TestWebhookTimestamp(t *testing.T) {
	t.Parallel()

	now := time.Now()
	envelope := &WebhookEnvelope{WebhookTimestamp: now.UnixMilli()}
	require.NoError(t, envelope.Timestamp(now))

	envelope.WebhookTimestamp = now.Add(-2 * time.Minute).UnixMilli()
	require.NoError(t, envelope.Timestamp(now))

	envelope.WebhookTimestamp = 0
	require.NoError(t, envelope.Timestamp(now))

	envelope.WebhookTimestamp = now.Add(-25 * time.Hour).UnixMilli()
	require.Error(t, envelope.Timestamp(now))

	envelope.WebhookTimestamp = now.Add(2 * time.Minute).UnixMilli()
	require.Error(t, envelope.Timestamp(now))
}

func TestParseEnvelopeIssueData(t *testing.T) {
	t.Parallel()

	body := []byte(`{
		"action":"update",
		"type":"Issue",
		"data":{"id":"issue-1","title":"Hello","state":{"type":"started"},"priority":2},
		"actor":{"id":"user-1"},
		"webhookTimestamp": 1
	}`)

	envelope, err := ParseEnvelope(body)
	require.NoError(t, err)
	assert.Equal(t, "update", envelope.Action)
	assert.Equal(t, "Issue", envelope.Type)

	data, err := envelope.IssueData()
	require.NoError(t, err)
	assert.Equal(t, "issue-1", data.ID)
	assert.Equal(t, "started", data.State.Type)
	assert.Equal(t, 2, data.Priority)
	assert.True(t, data.Has("title"))
	assert.True(t, data.Has("state"))
	assert.True(t, data.Has("priority"))
	assert.False(t, data.Has("dueDate"))
	assert.False(t, data.Has("description"))
	assert.False(t, data.Has("assignee"))
}
