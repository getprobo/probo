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
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.gearno.de/kit/log"
)

type fakeThreadClient struct {
	replies []ThreadReply
	err     error
}

func (f *fakeThreadClient) CreateMessage(
	context.Context,
	string,
	string,
	string,
	string,
) (*MessageRef, error) {
	return nil, nil
}

func (f *fakeThreadClient) ListThreadReplies(
	context.Context,
	string,
	string,
) ([]ThreadReply, error) {
	return f.replies, f.err
}

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
			for i := 0; i < 51; i++ {
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
			for i := 0; i < threadTranscriptMaxMessages+3; i++ {
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

	handler := &Handler{
		logger: log.NewLogger(log.WithOutput(io.Discard)),
	}
	transcript := handler.collectThreadTranscript(
		t.Context(),
		&fakeThreadClient{
			replies: []ThreadReply{
				{User: "U1", Text: "root"},
				{User: "U2", Text: "reply"},
			},
			err: &APIError{Code: "ratelimited"},
		},
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

	handler := &Handler{
		logger: log.NewLogger(log.WithOutput(io.Discard)),
	}
	transcript := handler.collectThreadTranscript(
		t.Context(),
		&fakeThreadClient{err: errors.New("network")},
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
