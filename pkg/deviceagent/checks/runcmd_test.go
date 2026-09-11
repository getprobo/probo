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

package checks

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSleepBinary = "/bin/sleep"

func TestRunCommandTermination(t *testing.T) {
	t.Parallel()

	if !isExecutableFile(testSleepBinary) {
		t.Skipf("%s is not available", testSleepBinary)
	}

	t.Run(
		"an expired deadline reports a timeout",
		func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			out := RunCommand(ctx, testSleepBinary, "30")
			require.Error(t, out.Err)
			assert.True(t, out.TimedOut)
			assert.False(t, out.Canceled)
		},
	)

	t.Run(
		"a cancelled context reports a cancellation",
		func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				time.Sleep(50 * time.Millisecond)
				cancel()
			}()

			out := RunCommand(ctx, testSleepBinary, "30")
			require.Error(t, out.Err)
			assert.True(t, out.Canceled)
			assert.False(t, out.TimedOut)
		},
	)

	t.Run(
		"a command that completes reports neither",
		func(t *testing.T) {
			t.Parallel()

			out := RunCommand(context.Background(), testSleepBinary, "0")
			require.NoError(t, out.Err)
			assert.False(t, out.TimedOut)
			assert.False(t, out.Canceled)
		},
	)
}
