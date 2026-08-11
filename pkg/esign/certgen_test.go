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

package esign

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatSealSignedAt(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		signedAt time.Time
		expected string
	}{
		"microsecond precision": {
			signedAt: time.Date(2026, time.January, 2, 3, 4, 5, 961506789, time.UTC),
			expected: "2026-01-02T03:04:05.961506Z",
		},
		"UTC conversion": {
			signedAt: time.Date(
				2026,
				time.January,
				2,
				3,
				4,
				5,
				123456789,
				time.FixedZone("UTC+2", 2*60*60),
			),
			expected: "2026-01-02T01:04:05.123456Z",
		},
		"whole second": {
			signedAt: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
			expected: "2026-01-02T03:04:05Z",
		},
	}

	for name, tt := range tests {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.expected, formatSealSignedAt(tt.signedAt))
			},
		)
	}
}
