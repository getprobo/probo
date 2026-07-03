// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
