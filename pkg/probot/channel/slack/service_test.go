// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

var slackEventTestMu sync.Mutex

func signRequest(secret string, body []byte, ts int64) (timestamp, signature string) {
	timestamp = strconv.FormatInt(ts, 10)
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, body)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(baseString))

	return timestamp, "v0=" + hex.EncodeToString(h.Sum(nil))
}

func testLogger(t *testing.T) *log.Logger {
	t.Helper()

	return log.NewLogger()
}

func TestService_EnqueueEvent(t *testing.T) {
	t.Parallel()

	slackEventTestMu.Lock()
	t.Cleanup(slackEventTestMu.Unlock)

	eventID := "E-handler-" + gid.New(gid.NilTenant, coredata.SlackbotEventEntityType).String()
	body := []byte(fmt.Sprintf(
		`{"type":"event_callback","event_id":%q,"event":{"type":"message","channel_type":"channel","text":"hi","channel":"C1","ts":"1.1"}}`,
		eventID,
	))

	pgClient := test.PGClient(t)
	handler := NewService(nil, nil, pgClient, testLogger(t))

	var envelope Envelope
	require.NoError(t, json.Unmarshal(body, &envelope))
	require.NoError(t, handler.EnqueueEvent(t.Context(), envelope))

	var persisted coredata.SlackbotEvent

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return persisted.LoadByEventID(ctx, conn, eventID)
			},
		),
	)
	assert.JSONEq(t, string(body), string(persisted.Envelope))
	assert.Equal(t, 0, persisted.AttemptCount)

	require.NoError(t, handler.EnqueueEvent(t.Context(), envelope))

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return persisted.Delete(ctx, conn)
			},
		),
	)
}

func TestService_EnqueueEventRejectsInvalidEnvelopes(t *testing.T) {
	t.Parallel()

	handler := NewService(nil, nil, nil, testLogger(t))

	tests := []struct {
		name     string
		envelope Envelope
		want     string
	}{
		{
			name:     "unsupported type",
			envelope: Envelope{Type: "other"},
			want:     "unsupported event envelope",
		},
		{
			name: "missing event id",
			envelope: Envelope{
				Type:  EnvelopeTypeEventCallback,
				Event: &EventBody{Type: EventTypeMessage},
			},
			want: "missing event_id",
		},
		{
			name: "missing event body",
			envelope: Envelope{
				Type:    EnvelopeTypeEventCallback,
				EventID: "E-missing-body",
			},
			want: "missing event",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				err := handler.EnqueueEvent(t.Context(), tt.envelope)
				require.Error(t, err)
				_, ok := errors.AsType[*InvalidEnvelopeError](err)
				require.True(t, ok)
				assert.Equal(t, tt.want, err.Error())
			},
		)
	}
}

func TestCleanText(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "hello", cleanText("<@U123> hello"))
	assert.Equal(t, "", cleanText("   "))
}

func TestShouldHandleConversationEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		event *EventBody
		want  bool
	}{
		{
			name: "channel mention",
			event: &EventBody{
				Type:        EventTypeAppMention,
				ChannelType: ChannelTypeChannel,
			},
			want: true,
		},
		{
			name: "direct message without mention",
			event: &EventBody{
				Type:        EventTypeMessage,
				ChannelType: ChannelTypeIM,
			},
			want: true,
		},
		{
			name: "channel message without mention",
			event: &EventBody{
				Type:        EventTypeMessage,
				ChannelType: ChannelTypeChannel,
			},
		},
		{
			name: "channel reaction",
			event: &EventBody{
				Type:        EventTypeReactionAdded,
				ChannelType: ChannelTypeChannel,
				User:        "U1",
			},
		},
		{
			name: "channel message edit",
			event: &EventBody{
				Type:        EventTypeMessage,
				Subtype:     EventSubtypeMessageChanged,
				ChannelType: ChannelTypeChannel,
				User:        "U1",
			},
		},
		{
			name: "nil event",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, shouldHandleConversationEvent(tt.event))
			},
		)
	}
}

func TestReplyTargetFor(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		replyTarget{channel: "C1", threadTS: "111.000"},
		replyTargetFor(EventBody{Channel: "C1", TS: "111.000"}),
	)
	assert.Equal(
		t,
		replyTarget{channel: "C1", threadTS: "222.000"},
		replyTargetFor(EventBody{Channel: "C1", TS: "111.000", ThreadTS: "222.000"}),
	)
	assert.Equal(
		t,
		replyTarget{channel: "D1", threadTS: ""},
		replyTargetFor(EventBody{Channel: "D1", ChannelType: ChannelTypeIM, TS: "111.000", ThreadTS: "222.000"}),
	)
}

func TestVerifySignature(t *testing.T) {
	t.Parallel()

	body := []byte(`{"type":"url_verification"}`)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	_, sig := signRequest("secret", body, time.Now().Unix())

	require.NoError(t, VerifySignature("secret", ts, sig, body))
	require.Error(t, VerifySignature("secret", ts, "v0=bad", body))
	require.NoError(t, VerifyAnySignature(ts, sig, body, "other", "secret"))
	require.Error(t, VerifyAnySignature(ts, sig, body, "other"))
}
