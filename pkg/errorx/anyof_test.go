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

package errorx_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/errorx"
)

func TestAnyOf(t *testing.T) {
	t.Parallel()

	sentinelA := errors.New("a")
	sentinelB := errors.New("b")
	sentinelC := errors.New("c")

	tests := []struct {
		name    string
		err     error
		targets []error
		want    bool
	}{
		{
			name:    "nil error",
			err:     nil,
			targets: []error{sentinelA},
			want:    false,
		},
		{
			name:    "no targets",
			err:     sentinelA,
			targets: nil,
			want:    false,
		},
		{
			name:    "direct match",
			err:     sentinelA,
			targets: []error{sentinelA},
			want:    true,
		},
		{
			name:    "wrapped match",
			err:     fmt.Errorf("outer: %w", sentinelB),
			targets: []error{sentinelA, sentinelB},
			want:    true,
		},
		{
			name:    "no match",
			err:     fmt.Errorf("outer: %w", sentinelC),
			targets: []error{sentinelA, sentinelB},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, errorx.AnyOf(tt.err, tt.targets...))
		})
	}
}
