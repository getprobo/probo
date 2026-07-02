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

package deviceagent

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPendingPostureQueue_EnqueueTrimsOldestBatches(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	for i := range maxPendingPostureBatches + 3 {
		results := []PostureResultPayload{
			{
				CheckKey:   fmt.Sprintf("check-%d", i),
				Status:     "pass",
				ObservedAt: time.Unix(int64(i), 0).UTC(),
			},
		}
		dropped, err := enqueuePendingPostureBatch(
			dir,
			results,
			time.Unix(int64(i), 0),
		)
		require.NoError(t, err)

		if i < maxPendingPostureBatches {
			assert.Equal(t, 0, dropped)
			continue
		}

		assert.Equal(t, 1, dropped)
	}

	batches, err := loadPendingPostureBatches(dir)
	require.NoError(t, err)
	require.Len(t, batches, maxPendingPostureBatches)
	assert.Equal(t, "check-3", batches[0].Results[0].CheckKey)
	assert.Equal(t, "check-98", batches[len(batches)-1].Results[0].CheckKey)
}

func TestAgent_flushQueuedPostures(t *testing.T) {
	t.Parallel()

	t.Run(
		"clears queue when all batches are flushed",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			_, err := enqueuePendingPostureBatch(
				dir,
				[]PostureResultPayload{{CheckKey: "first", Status: "pass", ObservedAt: time.Now().UTC()}},
				time.Now().UTC(),
			)
			require.NoError(t, err)
			_, err = enqueuePendingPostureBatch(
				dir,
				[]PostureResultPayload{{CheckKey: "second", Status: "pass", ObservedAt: time.Now().UTC()}},
				time.Now().UTC(),
			)
			require.NoError(t, err)

			var calls atomic.Int32

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/agent/v1/postures", r.URL.Path)
				calls.Add(1)
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			a := New(dir, "test", nil)
			a.client = NewClient(srv.URL, "api-key", "test-agent")
			a.flushQueuedPostures(context.Background())

			batches, err := loadPendingPostureBatches(dir)
			require.NoError(t, err)
			assert.Len(t, batches, 0)
			assert.Equal(t, int32(2), calls.Load())
		},
	)

	t.Run(
		"keeps unsent tail when a later flush request fails",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			_, err := enqueuePendingPostureBatch(
				dir,
				[]PostureResultPayload{{CheckKey: "first", Status: "pass", ObservedAt: time.Now().UTC()}},
				time.Now().UTC(),
			)
			require.NoError(t, err)
			_, err = enqueuePendingPostureBatch(
				dir,
				[]PostureResultPayload{{CheckKey: "second", Status: "pass", ObservedAt: time.Now().UTC()}},
				time.Now().UTC(),
			)
			require.NoError(t, err)

			var calls atomic.Int32

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/agent/v1/postures", r.URL.Path)

				call := calls.Add(1)
				if call == 2 {
					http.Error(w, "temporary error", http.StatusServiceUnavailable)
					return
				}

				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			a := New(dir, "test", nil)
			a.client = NewClient(srv.URL, "api-key", "test-agent")
			a.flushQueuedPostures(context.Background())

			batches, err := loadPendingPostureBatches(dir)
			require.NoError(t, err)
			require.Len(t, batches, 1)
			assert.Equal(t, "second", batches[0].Results[0].CheckKey)
			assert.Equal(t, int32(2), calls.Load())
		},
	)

	t.Run(
		"applies retry backoff with jitter gate after failures",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			_, err := enqueuePendingPostureBatch(
				dir,
				[]PostureResultPayload{{CheckKey: "first", Status: "pass", ObservedAt: time.Now().UTC()}},
				time.Now().UTC(),
			)
			require.NoError(t, err)

			var calls atomic.Int32

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/agent/v1/postures", r.URL.Path)
				calls.Add(1)
				http.Error(w, "temporary error", http.StatusServiceUnavailable)
			}))
			defer srv.Close()

			now := time.Unix(10_000, 0).UTC()
			a := New(dir, "test", nil)
			a.client = NewClient(srv.URL, "api-key", "test-agent")
			a.now = func() time.Time { return now }
			a.randInt63n = func(n int64) int64 { return n / 2 }

			a.flushQueuedPostures(context.Background())
			assert.Equal(t, int32(1), calls.Load())
			assert.Equal(t, pendingFlushBackoffMin, a.pendingFlushBackoff)
			firstRetryAt := a.pendingFlushRetryAt
			require.True(t, firstRetryAt.After(now))

			a.flushQueuedPostures(context.Background())
			assert.Equal(t, int32(1), calls.Load())

			now = firstRetryAt.Add(time.Second)

			a.flushQueuedPostures(context.Background())
			assert.Equal(t, int32(2), calls.Load())
			assert.Equal(t, pendingFlushBackoffMin*2, a.pendingFlushBackoff)
		},
	)

	t.Run(
		"resets retry backoff after successful flush",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			_, err := enqueuePendingPostureBatch(
				dir,
				[]PostureResultPayload{{CheckKey: "first", Status: "pass", ObservedAt: time.Now().UTC()}},
				time.Now().UTC(),
			)
			require.NoError(t, err)

			var calls atomic.Int32

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/agent/v1/postures", r.URL.Path)

				call := calls.Add(1)
				if call == 1 {
					http.Error(w, "temporary error", http.StatusServiceUnavailable)
					return
				}

				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			now := time.Unix(20_000, 0).UTC()
			a := New(dir, "test", nil)
			a.client = NewClient(srv.URL, "api-key", "test-agent")
			a.now = func() time.Time { return now }
			a.randInt63n = func(n int64) int64 { return n / 2 }

			a.flushQueuedPostures(context.Background())
			require.Equal(t, pendingFlushBackoffMin, a.pendingFlushBackoff)
			retryAt := a.pendingFlushRetryAt
			require.True(t, retryAt.After(now))

			now = retryAt.Add(time.Second)

			a.flushQueuedPostures(context.Background())
			assert.Equal(t, int32(2), calls.Load())
			assert.Zero(t, a.pendingFlushBackoff)
			assert.True(t, a.pendingFlushRetryAt.IsZero())

			batches, err := loadPendingPostureBatches(dir)
			require.NoError(t, err)
			assert.Len(t, batches, 0)
		},
	)
}
