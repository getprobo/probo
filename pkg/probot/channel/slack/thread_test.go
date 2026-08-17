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

			transcript := formatThreadTranscript(replies, "")
			assert.Equal(t, 51, strings.Count(transcript, "<@"))
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

			transcript := formatThreadTranscript(replies, "")
			assert.Contains(t, transcript, "<@UROOT>: root")
			assert.Equal(t, threadTranscriptMaxMessages, strings.Count(transcript, "<@"))
		},
	)
}

func TestCollectThreadTranscript_KeepsPartialRepliesOnLaterPageError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/api/conversations.replies", r.URL.Path)
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
			func(w http.ResponseWriter, _ *http.Request) {
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
