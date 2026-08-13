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

package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsWindowsServiceMissing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		out  string
		want bool
	}{
		{
			name: "error 1060",
			out:  "[SC] EnumQueryServicesStatus:OpenService FAILED 1060:\nThe specified service does not exist as an installed service.",
			want: true,
		},
		{
			name: "does not exist phrasing",
			out:  "The specified service does not exist as an installed service.",
			want: true,
		},
		{
			name: "running service query",
			out:  "SERVICE_NAME: ProboAgent\nSTATE: 4 RUNNING",
			want: false,
		},
		{
			name: "empty output",
			out:  "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, isWindowsServiceMissing(tt.out))
			},
		)
	}
}

func TestWaitUntilWindowsServiceGone_SucceedsAfterMissingQuery(t *testing.T) {
	t.Parallel()

	calls := 0
	err := waitUntilWindowsServiceGone(
		func() (string, error) {
			calls++
			if calls < 3 {
				return "SERVICE_NAME: ProboAgent\nSTATE: 3 STOP_PENDING", nil
			}

			return "[SC] OpenService FAILED 1060:\nThe specified service does not exist as an installed service.",
				errors.New("exit status 1060")
		},
		time.Second,
		func(time.Duration) {},
	)

	require.NoError(t, err)
	assert.Equal(t, 3, calls)
}

func TestWaitUntilWindowsServiceGone_TimesOutWhilePresent(t *testing.T) {
	t.Parallel()

	err := waitUntilWindowsServiceGone(
		func() (string, error) {
			return "SERVICE_NAME: ProboAgent\nSTATE: 1 STOPPED", nil
		},
		time.Millisecond,
		func(time.Duration) {},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot wait for previous windows service deletion")
}
