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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
)

type recordingStatusSetter struct {
	calls []recordedStatus
	err   error
}

type recordedStatus struct {
	channelID string
	threadTS  string
	status    string
}

func (s *recordingStatusSetter) SetStatus(
	_ context.Context,
	channelID string,
	threadTS string,
	status string,
	_ []string,
) error {
	s.calls = append(
		s.calls,
		recordedStatus{channelID: channelID, threadTS: threadTS, status: status},
	)

	return s.err
}

func newTestStatusHook(setter assistantStatusSetter) *assistantStatusHook {
	return newAssistantStatusHook(
		log.NewLogger(),
		func(context.Context) (assistantStatusSetter, error) { return setter, nil },
		"C123",
		"123.000",
	)
}

func TestAssistantStatusHook_OnRunEnd(t *testing.T) {
	t.Parallel()

	t.Run(
		"clears the indicator when the turn ends",
		func(t *testing.T) {
			t.Parallel()

			setter := &recordingStatusSetter{}
			hook := newTestStatusHook(setter)

			// Queueing send_message is not delivery; always clear on run end so
			// delivery retries/failures cannot leave Slack spinning.
			hook.OnRunEnd(t.Context(), nil, nil, nil)

			require.Len(t, setter.calls, 1)
			assert.Equal(t, "C123", setter.calls[0].channelID)
			assert.Equal(t, "123.000", setter.calls[0].threadTS)
			assert.Empty(t, setter.calls[0].status)
		},
	)

	t.Run(
		"clears the indicator when the turn fails",
		func(t *testing.T) {
			t.Parallel()

			setter := &recordingStatusSetter{}
			hook := newTestStatusHook(setter)

			hook.OnRunEnd(t.Context(), nil, nil, errors.New("model unavailable"))

			assert.Len(t, setter.calls, 1)
		},
	)

	t.Run(
		"keeps the indicator while the turn is suspended",
		func(t *testing.T) {
			t.Parallel()

			setter := &recordingStatusSetter{}
			hook := newTestStatusHook(setter)

			hook.OnRunEnd(t.Context(), nil, nil, &agent.SuspendedError{})

			assert.Empty(t, setter.calls)
		},
	)

	t.Run(
		"tolerates Slack API failures",
		func(t *testing.T) {
			t.Parallel()

			setter := &recordingStatusSetter{err: &APIError{Code: "thread_not_found"}}
			hook := newTestStatusHook(setter)

			hook.OnRunEnd(t.Context(), nil, nil, nil)

			assert.Len(t, setter.calls, 1)
		},
	)
}
