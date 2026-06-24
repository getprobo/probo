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

package gid_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/gid"
)

func TestNewSet(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()

	t.Run(
		"empty",
		func(t *testing.T) {
			t.Parallel()

			assert.Len(t, gid.NewSet(), 0)
		},
	)

	t.Run(
		"distinct ids",
		func(t *testing.T) {
			t.Parallel()

			a := gid.New(tenantID, 1)
			b := gid.New(tenantID, 2)

			set := gid.NewSet(a, b)

			assert.Len(t, set, 2)
			assert.True(t, set.Contains(a))
			assert.True(t, set.Contains(b))
		},
	)

	t.Run(
		"deduplicates",
		func(t *testing.T) {
			t.Parallel()

			a := gid.New(tenantID, 1)
			b := gid.New(tenantID, 2)

			set := gid.NewSet(a, a, b, a)

			assert.Len(t, set, 2)
		},
	)
}

func TestSet_Contains(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()

	t.Run(
		"member",
		func(t *testing.T) {
			t.Parallel()

			a := gid.New(tenantID, 1)

			assert.True(t, gid.NewSet(a).Contains(a))
		},
	)

	t.Run(
		"non-member",
		func(t *testing.T) {
			t.Parallel()

			a := gid.New(tenantID, 1)
			b := gid.New(tenantID, 2)

			assert.False(t, gid.NewSet(a).Contains(b))
		},
	)

	t.Run(
		"nil gid only member when added",
		func(t *testing.T) {
			t.Parallel()

			a := gid.New(tenantID, 1)

			assert.False(t, gid.NewSet(a).Contains(gid.Nil))
			assert.True(t, gid.NewSet(gid.Nil).Contains(gid.Nil))
		},
	)

	t.Run(
		"empty set",
		func(t *testing.T) {
			t.Parallel()

			assert.False(t, gid.NewSet().Contains(gid.New(tenantID, 1)))
		},
	)
}
