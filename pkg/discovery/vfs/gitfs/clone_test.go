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

package gitfs

import (
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
)

func TestMinimumCloneOptions(t *testing.T) {
	t.Parallel()

	opts := minimumCloneOptions("https://github.com/acme/api.git", nil, "main")
	assert.Equal(t, "https://github.com/acme/api.git", opts.URL)
	assert.Equal(t, ShallowCloneDepth, opts.Depth)
	assert.True(t, opts.SingleBranch)
	assert.Equal(t, git.NoTags, opts.Tags)
	assert.Equal(t, git.NoRecurseSubmodules, opts.RecurseSubmodules)
	assert.Equal(t, plumbing.NewBranchReferenceName("main"), opts.ReferenceName)

	emptyBranch := minimumCloneOptions("https://github.com/acme/api.git", nil, "")
	assert.Equal(t, plumbing.ReferenceName(""), emptyBranch.ReferenceName)
}
