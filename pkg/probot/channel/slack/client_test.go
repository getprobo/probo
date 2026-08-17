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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
)

const testBotToken = "xoxb-secret-test-token"

func TestClientMessageRequests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		wantBody map[string]any
		call     func(context.Context, *Client) error
	}{
		{
			name: "update message",
			path: "/api/chat.update",
			wantBody: map[string]any{
				"channel": "C123",
				"ts":      "123.456",
				"text":    "updated",
				"blocks":  []any{map[string]any{"type": "section"}},
			},
			call: func(ctx context.Context, c *Client) error {
				return c.UpdateMessage(
					ctx,
					"C123",
					"123.456",
					"updated",
					[]any{map[string]any{"type": "section"}},
				)
			},
		},
		{
			name: "post ephemeral",
			path: "/api/chat.postEphemeral",
			wantBody: map[string]any{
				"channel": "C123",
				"user":    "U123",
				"text":    "private",
				"blocks":  []any{map[string]any{"type": "divider"}},
			},
			call: func(ctx context.Context, c *Client) error {
				return c.PostEphemeral(
					ctx,
					"C123",
					"U123",
					"private",
					[]any{map[string]any{"type": "divider"}},
				)
			},
		},
		{
			name: "set assistant status",
			path: "/api/assistant.threads.setStatus",
			wantBody: map[string]any{
				"channel_id": "C123",
				"thread_ts":  "123.456",
				"status":     "is working on your request...",
				"loading_messages": []any{
					"Looking that up…",
					"Checking your workspace…",
					"Drafting a reply…",
				},
			},
			call: func(ctx context.Context, c *Client) error {
				return c.SetStatus(
					ctx,
					"C123",
					"123.456",
					"is working on your request...",
					[]string{
						"Looking that up…",
						"Checking your workspace…",
						"Drafting a reply…",
					},
				)
			},
		},
		{
			name: "clear assistant status",
			path: "/api/assistant.threads.setStatus",
			wantBody: map[string]any{
				"channel_id": "C123",
				"thread_ts":  "123.456",
				"status":     "",
			},
			call: func(ctx context.Context, c *Client) error {
				return c.SetStatus(ctx, "C123", "123.456", "", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				server := httptest.NewServer(
					http.HandlerFunc(
						func(w http.ResponseWriter, r *http.Request) {
							assert.Equal(t, http.MethodPost, r.Method)
							assert.Equal(t, tt.path, r.URL.Path)
							assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
							assert.Equal(t, "Bearer "+testBotToken, r.Header.Get("Authorization"))

							var gotBody map[string]any
							require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
							assert.Equal(t, tt.wantBody, gotBody)

							_, err := w.Write([]byte(`{"ok":true}`))
							require.NoError(t, err)
						},
					),
				)
				t.Cleanup(server.Close)

				client := newTestClient(server.URL + "/api")
				require.NoError(t, tt.call(t.Context(), client))
			},
		)
	}
}

func TestClientListConversations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cursor     string
		wantCursor string
	}{
		{name: "first page"},
		{name: "following page", cursor: "current-cursor", wantCursor: "current-cursor"},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				server := httptest.NewServer(
					http.HandlerFunc(
						func(w http.ResponseWriter, r *http.Request) {
							assert.Equal(t, http.MethodGet, r.Method)
							assert.Equal(t, "/api/conversations.list", r.URL.Path)
							assert.Equal(t, tt.wantCursor, r.URL.Query().Get("cursor"))
							assert.Equal(t, "true", r.URL.Query().Get("exclude_archived"))
							assert.Equal(t, "200", r.URL.Query().Get("limit"))
							assert.Equal(t, "public_channel,private_channel", r.URL.Query().Get("types"))
							assert.Equal(t, "Bearer "+testBotToken, r.Header.Get("Authorization"))

							_, err := w.Write(
								[]byte(
									`{"ok":true,"channels":[{"id":"C1","name":"general","is_member":true,"is_archived":false},{"id":"C2","name":"private","is_member":false,"is_archived":true}],"response_metadata":{"next_cursor":"next-cursor"}}`,
								),
							)
							require.NoError(t, err)
						},
					),
				)
				t.Cleanup(server.Close)

				page, err := newTestClient(server.URL+"/api").ListConversations(t.Context(), tt.cursor)
				require.NoError(t, err)
				assert.Equal(
					t,
					[]Conversation{
						{ID: "C1", Name: "general", IsMember: true, IsArchived: false},
						{ID: "C2", Name: "private", IsMember: false, IsArchived: true},
					},
					page.Conversations,
				)
				assert.Equal(t, "next-cursor", page.NextCursor)
			},
		)
	}
}

func TestClientListThreadReplies(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/conversations.replies", r.URL.Path)
				assert.Equal(t, "C123", r.URL.Query().Get("channel"))
				assert.Equal(t, "111.000", r.URL.Query().Get("ts"))
				assert.Equal(t, "200", r.URL.Query().Get("limit"))
				assert.Equal(t, "Bearer "+testBotToken, r.Header.Get("Authorization"))

				_, err := w.Write(
					[]byte(
						`{"ok":true,"messages":[{"user":"U1","text":"root","ts":"111.000"},{"user":"U2","text":"reply","ts":"111.001"}]}`,
					),
				)
				require.NoError(t, err)
			},
		),
	)
	t.Cleanup(server.Close)

	replies, err := newTestClient(server.URL+"/api").ListThreadReplies(
		t.Context(),
		"C123",
		"111.000",
	)
	require.NoError(t, err)
	assert.Equal(
		t,
		[]ThreadReply{
			{User: "U1", Text: "root", TS: "111.000"},
			{User: "U2", Text: "reply", TS: "111.001"},
		},
		replies,
	)
}

func TestClientListThreadReplies_PaginatesAcrossPages(t *testing.T) {
	t.Parallel()

	var requests int

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/conversations.replies", r.URL.Path)
				assert.Equal(t, "200", r.URL.Query().Get("limit"))

				requests++

				switch r.URL.Query().Get("cursor") {
				case "":
					_, err := w.Write(
						[]byte(
							`{"ok":true,"messages":[{"user":"U1","text":"root","ts":"111.000"}],"response_metadata":{"next_cursor":"page-2"}}`,
						),
					)
					require.NoError(t, err)
				case "page-2":
					_, err := w.Write(
						[]byte(
							`{"ok":true,"messages":[{"user":"U2","text":"reply","ts":"111.001"}],"response_metadata":{"next_cursor":""}}`,
						),
					)
					require.NoError(t, err)
				default:
					t.Fatalf("unexpected cursor %q", r.URL.Query().Get("cursor"))
				}
			},
		),
	)
	t.Cleanup(server.Close)

	replies, err := newTestClient(server.URL+"/api").ListThreadReplies(
		t.Context(),
		"C123",
		"111.000",
	)
	require.NoError(t, err)
	assert.Equal(t, 2, requests)
	assert.Equal(
		t,
		[]ThreadReply{
			{User: "U1", Text: "root", TS: "111.000"},
			{User: "U2", Text: "reply", TS: "111.001"},
		},
		replies,
	)
}

func TestClientListThreadReplies_KeepsFirstPageWhenLaterPageFails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("cursor") == "" {
					_, err := w.Write(
						[]byte(
							`{"ok":true,"messages":[{"user":"U1","text":"root","ts":"111.000"}],"response_metadata":{"next_cursor":"page-2"}}`,
						),
					)
					require.NoError(t, err)

					return
				}

				_, err := w.Write([]byte(`{"ok":false,"error":"ratelimited"}`))
				require.NoError(t, err)
			},
		),
	)
	t.Cleanup(server.Close)

	replies, err := newTestClient(server.URL+"/api").ListThreadReplies(
		t.Context(),
		"C123",
		"111.000",
	)
	require.Error(t, err)
	assert.ErrorContains(t, err, "ratelimited")
	assert.Equal(
		t,
		[]ThreadReply{
			{User: "U1", Text: "root", TS: "111.000"},
		},
		replies,
	)
}

func TestClientUninstallApp(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/apps.uninstall", r.URL.Path)
				assert.Equal(t, "Bearer "+testBotToken, r.Header.Get("Authorization"))
				require.NoError(t, r.ParseForm())
				assert.Equal(t, "client-id", r.Form.Get("client_id"))
				assert.Equal(t, "client-secret", r.Form.Get("client_secret"))

				_, err := w.Write([]byte(`{"ok":true}`))
				require.NoError(t, err)
			},
		),
	)
	t.Cleanup(server.Close)

	require.NoError(
		t,
		newTestClient(server.URL+"/api").UninstallApp(
			t.Context(),
			"client-id",
			"client-secret",
		),
	)
}

func TestClientUninstallAppAlreadyGone(t *testing.T) {
	t.Parallel()

	for _, errorCode := range []string{"account_inactive", "invalid_auth", "token_revoked"} {
		t.Run(
			errorCode,
			func(t *testing.T) {
				t.Parallel()

				server := httptest.NewServer(
					http.HandlerFunc(
						func(w http.ResponseWriter, _ *http.Request) {
							_, err := fmt.Fprintf(
								w,
								`{"ok":false,"error":%q}`,
								errorCode,
							)
							require.NoError(t, err)
						},
					),
				)
				t.Cleanup(server.Close)

				err := newTestClient(server.URL+"/api").UninstallApp(
					t.Context(),
					"client-id",
					"client-secret",
				)
				require.NoError(t, err)
			},
		)
	}
}

func TestClientSlackAPIErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{
			name: "update message",
			call: func(ctx context.Context, c *Client) error {
				return c.UpdateMessage(ctx, "C123", "123.456", "updated", nil)
			},
		},
		{
			name: "post ephemeral",
			call: func(ctx context.Context, c *Client) error {
				return c.PostEphemeral(ctx, "C123", "U123", "private", nil)
			},
		},
		{
			name: "list conversations",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.ListConversations(ctx, "")
				return err
			},
		},
		{
			name: "list thread replies",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.ListThreadReplies(ctx, "C123", "111.000")
				return err
			},
		},
		{
			name: "uninstall app",
			call: func(ctx context.Context, c *Client) error {
				return c.UninstallApp(ctx, "client-id", "client-secret")
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				server := httptest.NewServer(
					http.HandlerFunc(
						func(w http.ResponseWriter, _ *http.Request) {
							_, err := w.Write([]byte(`{"ok":false,"error":"missing_scope"}`))
							require.NoError(t, err)
						},
					),
				)
				t.Cleanup(server.Close)

				err := tt.call(t.Context(), newTestClient(server.URL+"/api"))
				require.Error(t, err)
				assert.ErrorContains(t, err, "slack api error: missing_scope")
				assert.NotContains(t, err.Error(), testBotToken)
			},
		)
	}
}

func TestClientResponseErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		response string
		want     string
	}{
		{name: "invalid JSON", response: "not-json", want: "cannot decode slack response"},
		{name: "Slack API error", response: `{"ok":false,"error":"channel_not_found"}`, want: "slack api error: channel_not_found"},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				server := httptest.NewServer(
					http.HandlerFunc(
						func(w http.ResponseWriter, _ *http.Request) {
							_, err := w.Write([]byte(tt.response))
							require.NoError(t, err)
						},
					),
				)
				t.Cleanup(server.Close)

				err := newTestClient(server.URL+"/api").UpdateMessage(
					t.Context(),
					"C123",
					"123.456",
					"updated",
					nil,
				)
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.want)
				assert.NotContains(t, err.Error(), testBotToken)
			},
		)
	}
}

func TestClientAPIBaseOverride(t *testing.T) {
	t.Parallel()

	client := NewClient(
		testBotToken,
		"https://slack.example.test/custom/api/",
		log.NewLogger(log.WithName("test")),
	)

	got, err := client.methodURL(slackMethodUpdateMessage)
	require.NoError(t, err)
	assert.Equal(t, "https://slack.example.test/custom/api/chat.update", got)
}

func TestClientTransportErrorDoesNotExposeToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.NotFoundHandler())
	client := newTestClient(server.URL + "/api")
	server.Close()

	err := client.UpdateMessage(t.Context(), "C123", "123.456", "updated", nil)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), testBotToken)
}

func TestClientCreateMessageWithBlocks(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/chat.postMessage", r.URL.Path)
			assert.Equal(t, "Bearer "+testBotToken, r.Header.Get("Authorization"))

			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "C123", body["channel"])
			assert.Equal(t, "Access requested", body["text"])
			assert.Equal(t, "04a8d30f-8f4d-4f41-9b36-6440cb9821d7", body["client_msg_id"])

			_, _ = io.WriteString(
				w,
				`{"ok":true,"channel":"C123","ts":"123.456"}`,
			)
		},
	))
	defer server.Close()

	ref, err := newTestClient(server.URL+"/api").CreateMessageWithBlocks(
		t.Context(),
		"C123",
		"Access requested",
		[]any{map[string]any{"type": "section"}},
		"04a8d30f-8f4d-4f41-9b36-6440cb9821d7",
	)
	require.NoError(t, err)
	assert.Equal(t, "C123", ref.Channel)
	assert.Equal(t, "123.456", ref.TS)
}

func TestClientSlackAPIErrorIncludesRetryMetadata(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Retry-After", "17")
			w.WriteHeader(http.StatusTooManyRequests)
			_, err := io.WriteString(w, `{"ok":false,"error":"ratelimited"}`)
			require.NoError(t, err)
		},
	))
	t.Cleanup(server.Close)

	err := newTestClient(server.URL+"/api").UpdateMessage(
		t.Context(),
		"C123",
		"123.456",
		"updated",
		nil,
	)
	require.Error(t, err)

	apiErr, ok := errors.AsType[*APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusTooManyRequests, apiErr.StatusCode)
	assert.Equal(t, "ratelimited", apiErr.Code)
	assert.Equal(t, 17*time.Second, apiErr.RetryAfter)
}

func TestClientCreateMessageHTTPErrorIsTyped(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, err := io.WriteString(w, "service unavailable")
			require.NoError(t, err)
		},
	))
	t.Cleanup(server.Close)

	_, err := newTestClient(server.URL+"/api").CreateMessageWithBlocks(
		t.Context(),
		"C123",
		"Access requested",
		nil,
		"04a8d30f-8f4d-4f41-9b36-6440cb9821d7",
	)
	require.Error(t, err)

	apiErr, ok := errors.AsType[*APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusServiceUnavailable, apiErr.StatusCode)
}

func TestNewClientBoundsSlackRequests(t *testing.T) {
	t.Parallel()

	client := NewClient(testBotToken, "https://slack.example.com/api", log.NewLogger())
	assert.Equal(t, slackHTTPTimeout, client.httpClient.Timeout)
}

func newTestClient(apiBaseURL string) *Client {
	client := NewClient(
		testBotToken,
		apiBaseURL,
		log.NewLogger(log.WithName("test")),
	)
	client.httpClient = httpclient.DefaultPooledClient(
		httpclient.WithSSRFProtection(),
		httpclient.WithSSRFAllowLoopback(),
	)

	return client
}
