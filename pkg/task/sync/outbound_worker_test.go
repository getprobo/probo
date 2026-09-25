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

package tasksync

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
)

func TestOutboundRetryDelay(t *testing.T) {
	t.Parallel()

	assert.Equal(t, time.Minute, outboundRetryDelay(1))
	assert.Equal(t, 2*time.Minute, outboundRetryDelay(2))
	assert.Equal(t, 4*time.Minute, outboundRetryDelay(3))
	assert.Equal(t, 32*time.Minute, outboundRetryDelay(6))
	assert.Equal(t, 32*time.Minute, outboundRetryDelay(99))
}

func TestOutboundLinkMatchesJob(t *testing.T) {
	t.Parallel()

	payload := JobPayload{ExternalID: "issue-a"}

	t.Run(
		"missing link",
		func(t *testing.T) {
			t.Parallel()

			assert.False(t, outboundLinkMatchesJob(nil, payload))
		},
	)

	t.Run(
		"mismatched external id",
		func(t *testing.T) {
			t.Parallel()

			assert.False(
				t,
				outboundLinkMatchesJob(
					&coredata.TaskExternalLink{ExternalID: "issue-b"},
					payload,
				),
			)
		},
	)

	t.Run(
		"matching link",
		func(t *testing.T) {
			t.Parallel()

			assert.True(
				t,
				outboundLinkMatchesJob(
					&coredata.TaskExternalLink{ExternalID: "issue-a"},
					payload,
				),
			)
		},
	)
}

func TestOutboundLeaseHeartbeatInterval(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 30*time.Second, outboundLeaseHeartbeatInterval(5*time.Minute, 30*time.Second))
	assert.Equal(t, time.Minute, outboundLeaseHeartbeatInterval(2*time.Minute, 2*time.Minute))
	assert.Equal(t, time.Millisecond, outboundLeaseHeartbeatInterval(time.Millisecond, time.Second))
}

func TestProcessingLeaseLost(t *testing.T) {
	t.Parallel()

	t.Run(
		"direct lease error",
		func(t *testing.T) {
			t.Parallel()

			assert.True(
				t,
				processingLeaseLost(t.Context(), coredata.ErrProcessingLeaseLost),
			)
		},
	)

	t.Run(
		"cancelled cause",
		func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancelCause(t.Context())
			cancel(coredata.ErrProcessingLeaseLost)

			assert.True(t, processingLeaseLost(ctx, context.Canceled))
		},
	)

	t.Run(
		"unrelated error",
		func(t *testing.T) {
			t.Parallel()

			assert.False(t, processingLeaseLost(t.Context(), assert.AnError))
		},
	)
}
