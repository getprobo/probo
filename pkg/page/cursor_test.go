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

package page

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testOrderField string

func (f testOrderField) Column() string { return string(f) }
func (f testOrderField) String() string { return string(f) }

func TestNewCursor(t *testing.T) {
	t.Parallel()

	orderBy := OrderBy[testOrderField]{
		Field:     testOrderField("created_at"),
		Direction: OrderDirectionAsc,
	}

	tests := []struct {
		name         string
		size         int
		expectedSize int
	}{
		{
			name:         "negative size defaults to DefaultCursorSize",
			size:         -10,
			expectedSize: DefaultCursorSize,
		},
		{
			name:         "zero size defaults to DefaultCursorSize",
			size:         0,
			expectedSize: DefaultCursorSize,
		},
		{
			name:         "valid size is kept as-is",
			size:         50,
			expectedSize: 50,
		},
		{
			name:         "oversized value is clamped to MaxCursorSize",
			size:         500,
			expectedSize: MaxCursorSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cursor := NewCursor[testOrderField](tt.size, nil, Head, orderBy)
			assert.Equal(t, tt.expectedSize, cursor.Size)
		})
	}
}
