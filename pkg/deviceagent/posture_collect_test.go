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
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/deviceagent/checks"
	"go.probo.inc/probo/pkg/gid"
)

// stubCheck runs an arbitrary function so a test can end the run mid-set.
type stubCheck struct {
	key string
	run func(ctx context.Context) checks.Result
}

func (c stubCheck) Key() string { return c.key }

func (c stubCheck) Run(ctx context.Context) checks.Result {
	r := c.run(ctx)
	if r.CheckKey == "" {
		r.CheckKey = c.key
	}

	return r
}

func passingCheck(key string) stubCheck {
	return stubCheck{
		key: key,
		run: func(context.Context) checks.Result {
			return checks.Result{Status: checks.StatusPass}
		},
	}
}

func rememberableUnknownCheck(key string, ev map[string]any) stubCheck {
	return stubCheck{
		key: key,
		run: func(context.Context) checks.Result {
			return checks.Result{
				Status:       checks.StatusUnknown,
				Evidence:     ev,
				Rememberable: true,
			}
		},
	}
}

func TestAgent_CollectOnce(t *testing.T) {
	t.Parallel()

	t.Run(
		"a complete run returns every check and no error",
		func(t *testing.T) {
			t.Parallel()

			a := New(t.TempDir(), "test", nil)
			a.checkSet = func() []checks.Check {
				return []checks.Check{passingCheck("FIRST"), passingCheck("SECOND")}
			}

			results, err := a.CollectOnce(context.Background())
			require.NoError(t, err)
			require.Len(t, results, 2)
			assert.Equal(t, "FIRST", results[0].CheckKey)
			assert.False(t, results[0].ObservedAt.IsZero())
		},
	)

	t.Run(
		"a run cut short returns the partial results and an error",
		func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			dir := t.TempDir()
			a := New(dir, "test", nil)
			a.checkSet = func() []checks.Check {
				return []checks.Check{
					passingCheck("FIRST"),
					stubCheck{
						key: "SECOND",
						run: func(context.Context) checks.Result {
							cancel()

							return checks.Result{Status: checks.StatusPass}
						},
					},
					passingCheck("THIRD"),
				}
			}

			results, err := a.CollectOnce(ctx)
			require.ErrorIs(t, err, context.Canceled)
			require.Len(t, results, 2)
			assert.Equal(t, "SECOND", results[1].CheckKey)

			_, statErr := os.Stat(checkMemoryPath(dir))
			assert.ErrorIs(t, statErr, os.ErrNotExist)
		},
	)

	t.Run(
		"a later rememberable unknown keeps the last pass",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			a := New(dir, "test", nil)
			a.checkSet = func() []checks.Check {
				return []checks.Check{
					stubCheck{
						key: "SCREEN_LOCK",
						run: func(context.Context) checks.Result {
							return checks.Result{
								Status:   checks.StatusPass,
								Evidence: map[string]any{"screen_lock_enforced": true},
							}
						},
					},
				}
			}

			first, err := a.CollectOnce(context.Background())
			require.NoError(t, err)
			require.Len(t, first, 1)
			assert.Equal(t, checks.StatusPass, first[0].Status)

			a.checkSet = func() []checks.Check {
				return []checks.Check{
					rememberableUnknownCheck(
						"SCREEN_LOCK",
						map[string]any{
							"backend": "hkey_users",
							"note":    "no interactive user hives loaded",
						},
					),
				}
			}

			second, err := a.CollectOnce(context.Background())
			require.NoError(t, err)
			require.Len(t, second, 1)
			assert.Equal(t, checks.StatusPass, second[0].Status)
			assert.Equal(t, true, second[0].Evidence["screen_lock_enforced"])
			assert.Equal(t, true, second[0].Evidence["remembered"])
		},
	)
}

func TestAgent_doPosturesDropsTruncatedRun(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	var pushes atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pushes.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deviceID := gid.New(gid.NewTenantID(), coredata.DeviceEntityType)

	a := New(dir, "test", nil)
	a.cfg = &Config{ServerURL: srv.URL, DeviceID: deviceID.String()}
	a.client = NewClient(srv.URL, "api-key", "test-agent")
	a.checkSet = func() []checks.Check {
		return []checks.Check{
			passingCheck("FIRST"),
			stubCheck{
				key: "SECOND",
				run: func(context.Context) checks.Result {
					cancel()

					return checks.Result{Status: checks.StatusPass}
				},
			},
			passingCheck("THIRD"),
		}
	}

	a.doPostures(ctx)

	assert.Equal(t, int32(0), pushes.Load(), "a truncated run must not be pushed")

	batches, err := loadPendingPostureBatches(dir)
	require.NoError(t, err)
	assert.Empty(t, batches, "a truncated run must not be queued for replay")
}

func TestAgent_doPosturesPushesCompleteRun(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	var pushes atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pushes.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	deviceID := gid.New(gid.NewTenantID(), coredata.DeviceEntityType)

	a := New(dir, "test", nil)
	a.cfg = &Config{ServerURL: srv.URL, DeviceID: deviceID.String()}
	a.client = NewClient(srv.URL, "api-key", "test-agent")
	a.checkSet = func() []checks.Check {
		return []checks.Check{passingCheck("FIRST"), passingCheck("SECOND")}
	}

	a.doPostures(context.Background())

	assert.Equal(t, int32(1), pushes.Load())

	batches, err := loadPendingPostureBatches(dir)
	require.NoError(t, err)
	assert.Empty(t, batches)
}
