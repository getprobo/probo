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

package coredata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeObservedAt(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		observed time.Time
		want     time.Time
	}{
		{
			name:     "zero clamps to now",
			observed: time.Time{},
			want:     now,
		},
		{
			name:     "past is unchanged",
			observed: now.Add(-time.Hour),
			want:     now.Add(-time.Hour),
		},
		{
			name:     "slightly future within skew is unchanged",
			observed: now.Add(2 * time.Minute),
			want:     now.Add(2 * time.Minute),
		},
		{
			name:     "far future beyond skew clamps to now",
			observed: now.Add(time.Hour),
			want:     now,
		},
		{
			name:     "exactly at skew boundary is unchanged",
			observed: now.Add(devicePostureObservedAtClockSkew),
			want:     now.Add(devicePostureObservedAtClockSkew),
		},
		{
			name:     "just beyond skew boundary clamps to now",
			observed: now.Add(devicePostureObservedAtClockSkew + time.Nanosecond),
			want:     now,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeObservedAt(tt.observed, now)
			assert.Equal(t, tt.want, got)
		})
	}
}
