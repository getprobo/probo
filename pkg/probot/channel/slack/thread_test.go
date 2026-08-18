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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
)

func TestFormatThreadTranscript(t *testing.T) {
	t.Parallel()

	t.Run(
		"keeps humans and the installing bot",
		func(t *testing.T) {
			t.Parallel()

			transcript := formatThreadTranscript(
				[]ThreadReply{
					{User: "U1", Text: "<@BOT> we should grant the report"},
					{User: "U2", Text: "only the latest"},
					{BotID: "BOTHER", User: "UOTHER", Text: "ignore me"},
					{User: "UBOT", BotID: "BPROBOT", Text: "Access request"},
					{User: "U1", Text: "<@BOT> grant the latest"},
					{Subtype: "message_deleted", User: "U1", Text: "gone"},
				},
				"UBOT",
				"BPROBOT",
			)

			assert.Equal(
				t,
				"Thread:\n<@U1>: we should grant the report\n<@U2>: only the latest\n<@UBOT>: Access request\n<@U1>: grant the latest",
				transcript,
			)
		},
	)

	t.Run(
		"includes more than fifty messages",
		func(t *testing.T) {
			t.Parallel()

			replies := make([]ThreadReply, 0, 51)
			for range 51 {
				replies = append(replies, ThreadReply{User: "U1", Text: "msg"})
			}

			transcript := formatThreadTranscript(replies, "", "")
			assert.Equal(t, 51, strings.Count(transcript, "<@"))
		},
	)

	t.Run(
		"keeps user-less replies only from the installed bot",
		func(t *testing.T) {
			t.Parallel()

			transcript := formatThreadTranscript(
				[]ThreadReply{
					{User: "U1", Text: "hello", TS: "1.000"},
					{BotID: "BPROBOT", Text: "Access request", TS: "2.000"},
					{BotID: "BOTHER", Text: "foreign bot", TS: "2.500"},
					{BotID: "BOTHER", User: "UOTHER", Text: "ignore me", TS: "3.000"},
				},
				"UBOT",
				"BPROBOT",
			)

			assert.Equal(
				t,
				"Thread:\n<@U1>: hello\n<@UBOT>: Access request",
				transcript,
			)
		},
	)

	t.Run(
		"drops user-less bot replies without an installed bot id",
		func(t *testing.T) {
			t.Parallel()

			transcript := formatThreadTranscript(
				[]ThreadReply{
					{User: "U1", Text: "hello", TS: "1.000"},
					{BotID: "BPROBOT", Text: "Access request", TS: "2.000"},
				},
				"UBOT",
				"",
			)

			assert.Equal(t, "Thread:\n<@U1>: hello", transcript)
		},
	)

	t.Run(
		"keeps thread root and newest messages when over cap",
		func(t *testing.T) {
			t.Parallel()

			replies := make([]ThreadReply, 0, threadTranscriptMaxMessages+5)

			replies = append(replies, ThreadReply{User: "UROOT", Text: "root"})
			for range threadTranscriptMaxMessages + 3 {
				replies = append(replies, ThreadReply{User: "U1", Text: "later"})
			}

			transcript := formatThreadTranscript(replies, "", "")
			assert.Contains(t, transcript, "<@UROOT>: root")
			assert.Equal(t, threadTranscriptMaxMessages, strings.Count(transcript, "<@"))
		},
	)
}

func TestCollectThreadTranscript_KeepsUserLessOwnBotReplies(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/auth.test":
					_, err := w.Write([]byte(`{"ok":true,"bot_id":"BPROBOT"}`))
					require.NoError(t, err)
				case "/api/conversations.replies":
					_, err := w.Write(
						[]byte(
							`{"ok":true,"messages":[{"user":"U1","text":"root","ts":"111.000"},{"bot_id":"BPROBOT","text":"Access request","ts":"111.001"},{"bot_id":"BOTHER","text":"foreign bot","ts":"111.002"},{"user":"U1","text":"<@BOT> grant this","ts":"111.003"}]}`,
						),
					)
					require.NoError(t, err)
				default:
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
			},
		),
	)
	t.Cleanup(server.Close)

	handler := &Service{logger: log.NewLogger()}
	transcript := handler.collectThreadTranscript(
		t.Context(),
		newTestClient(server.URL+"/api"),
		EventBody{
			Type:        EventTypeAppMention,
			User:        "U1",
			Text:        "<@BOT> grant this",
			Channel:     "C123",
			TS:          "111.003",
			ThreadTS:    "111.000",
			ChannelType: ChannelTypeChannel,
		},
		"UBOT",
	)

	assert.Equal(
		t,
		"Thread:\n<@U1>: root\n<@UBOT>: Access request\n<@U1>: grant this",
		transcript,
	)
	assert.NotContains(t, transcript, "foreign bot")
}

func TestCollectThreadTranscript_KeepsPartialRepliesOnLaterPageError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/auth.test":
					_, err := w.Write([]byte(`{"ok":true,"bot_id":"BPROBOT"}`))
					require.NoError(t, err)
				case "/api/conversations.replies":
					if r.URL.Query().Get("cursor") == "" {
						_, err := w.Write(
							[]byte(
								`{"ok":true,"messages":[{"user":"U1","text":"root","ts":"111.000"},{"user":"U2","text":"reply","ts":"111.001"}],"response_metadata":{"next_cursor":"page-2"}}`,
							),
						)
						require.NoError(t, err)

						return
					}

					_, err := w.Write([]byte(`{"ok":false,"error":"ratelimited"}`))
					require.NoError(t, err)
				default:
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
			},
		),
	)
	t.Cleanup(server.Close)

	handler := &Service{logger: log.NewLogger()}
	transcript := handler.collectThreadTranscript(
		t.Context(),
		newTestClient(server.URL+"/api"),
		EventBody{
			Type:        EventTypeAppMention,
			Text:        "<@BOT> ping",
			Channel:     "C123",
			TS:          "111.001",
			ThreadTS:    "111.000",
			ChannelType: ChannelTypeChannel,
		},
		"UBOT",
	)

	assert.Equal(t, "Thread:\n<@U1>: root\n<@U2>: reply", transcript)
	assert.NotContains(t, transcript, "ping")
}

func TestCollectThreadTranscript_FallsBackWhenNoRepliesCollected(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/auth.test" {
					_, err := w.Write([]byte(`{"ok":true,"bot_id":"BPROBOT"}`))
					require.NoError(t, err)

					return
				}

				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte(`{"ok":false,"error":"internal_error"}`))
				require.NoError(t, err)
			},
		),
	)
	t.Cleanup(server.Close)

	handler := &Service{logger: log.NewLogger()}
	transcript := handler.collectThreadTranscript(
		t.Context(),
		newTestClient(server.URL+"/api"),
		EventBody{
			Type:        EventTypeAppMention,
			Text:        "<@BOT> ping",
			Channel:     "C123",
			TS:          "111.001",
			ThreadTS:    "111.000",
			ChannelType: ChannelTypeChannel,
		},
		"UBOT",
	)

	assert.Equal(t, "ping", transcript)
}

func TestCollectThreadTranscript_AppendsTriggeringEventWhenMissing(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/auth.test":
					_, err := w.Write([]byte(`{"ok":true,"bot_id":"BPROBOT"}`))
					require.NoError(t, err)
				case "/api/conversations.replies":
					_, err := w.Write(
						[]byte(
							`{"ok":true,"messages":[{"user":"U1","text":"root","ts":"111.000"},{"user":"U2","text":"reply","ts":"111.001"}]}`,
						),
					)
					require.NoError(t, err)
				default:
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
			},
		),
	)
	t.Cleanup(server.Close)

	handler := &Service{logger: log.NewLogger()}
	transcript := handler.collectThreadTranscript(
		t.Context(),
		newTestClient(server.URL+"/api"),
		EventBody{
			Type:        EventTypeAppMention,
			User:        "U1",
			Text:        "<@BOT> grant this",
			Channel:     "C123",
			TS:          "111.999",
			ThreadTS:    "111.000",
			ChannelType: ChannelTypeChannel,
		},
		"UBOT",
	)

	assert.Equal(
		t,
		"Thread:\n<@U1>: root\n<@U2>: reply\n<@U1>: grant this",
		transcript,
	)
}

func TestInstalledBotIDFromReplies(t *testing.T) {
	t.Parallel()

	t.Run(
		"matches bot user id on replies",
		func(t *testing.T) {
			t.Parallel()

			botID := installedBotIDFromReplies(
				[]ThreadReply{
					{BotID: "BPROBOT", Text: "no user"},
					{User: "UBOT", BotID: "BPROBOT", Text: "with user"},
				},
				"UBOT",
			)

			assert.Equal(t, "BPROBOT", botID)
		},
	)

	t.Run(
		"returns empty when only user-less bot replies exist",
		func(t *testing.T) {
			t.Parallel()

			botID := installedBotIDFromReplies(
				[]ThreadReply{
					{BotID: "BPROBOT", Text: "no user"},
				},
				"UBOT",
			)

			assert.Empty(t, botID)
		},
	)
}

func TestAppendTriggeringEventIfMissing_AppendsTruncatedLongThreadEvent(t *testing.T) {
	t.Parallel()

	replies := make([]ThreadReply, 0, threadTranscriptMaxMessages+5)
	replies = append(replies, ThreadReply{User: "UROOT", Text: "root", TS: "0.000"})

	for i := range threadTranscriptMaxMessages + 3 {
		replies = append(
			replies,
			ThreadReply{
				User: "U1",
				Text: "later",
				TS:   fmt.Sprintf("%d.000", i+1),
			},
		)
	}

	eventTS := "2.000"
	kept := keptThreadReplies(replies, "UBOT", "BPROBOT")
	assert.False(t, threadReplyHasTS(kept, eventTS))

	transcript := appendTriggeringEventIfMissing(
		formatKeptThreadTranscript(kept, "UBOT"),
		kept,
		EventBody{
			User: "U1",
			Text: "<@BOT> grant this",
			TS:   eventTS,
		},
		"UBOT",
	)

	assert.Contains(t, transcript, "<@UROOT>: root")
	assert.Contains(t, transcript, "<@U1>: grant this")
	assert.Equal(t, threadTranscriptMaxMessages+1, strings.Count(transcript, "<@"))
}
