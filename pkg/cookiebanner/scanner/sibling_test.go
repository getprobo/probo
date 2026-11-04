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

package scanner

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSiblingOrigin(t *testing.T) {
	t.Parallel()

	site, ok := parseOrigin("https://example.com")
	require.True(t, ok)
	api, ok := parseOrigin("https://api.example.com")
	require.True(t, ok)

	sibling, found := siblingOrigin(site, api, "https://help.example.com/docs")
	assert.True(t, found)
	assert.Equal(t, "https://help.example.com", sibling)

	_, found = siblingOrigin(site, api, "https://www.example.com/pricing")
	assert.False(t, found)

	sibling, found = siblingOrigin(site, api, "https://WWW.SHOP.EXAMPLE.COM/cart")
	assert.True(t, found)
	assert.Equal(t, "https://shop.example.com", sibling)

	_, found = siblingOrigin(site, api, "https://api.example.com/v1")
	assert.False(t, found)

	_, found = siblingOrigin(site, api, "https://example.org/about")
	assert.False(t, found)

	_, found = siblingOrigin(site, api, "mailto:hello@example.com")
	assert.False(t, found)

	vercel, ok := parseOrigin("https://foo.vercel.app")
	require.True(t, ok)
	sibling, found = siblingOrigin(vercel, origin{}, "https://bar.vercel.app")
	assert.True(t, found)
	assert.Equal(t, "https://bar.vercel.app", sibling)
}

func TestNoteSibling_CapsAndResumesFromCheckpoint(t *testing.T) {
	t.Parallel()

	site, ok := parseOrigin("https://example.com")
	require.True(t, ok)

	known := make([]string, 0, 2)
	for i := range 2 {
		known = append(known, fmt.Sprintf("https://known-%d.example.com", i))
	}

	s := &scan{
		site: site,
		api:  origin{},
		front: newFrontier(Checkpoint{
			SiblingOrigins: known,
		}, 8),
	}

	s.noteSibling("https://known-0.example.com/again")
	assert.Empty(t, s.freshSiblings)

	for i := range MaxSiblingOrigins {
		s.noteSibling(fmt.Sprintf("https://fresh-%d.example.com/page", i))
	}

	assert.Len(t, s.freshSiblings, MaxSiblingOrigins-len(known))

	s.noteSibling("https://one-too-many.example.com")
	assert.Len(t, s.freshSiblings, MaxSiblingOrigins-len(known))

	drained := s.drainSiblings()
	assert.Len(t, drained, MaxSiblingOrigins-len(known))
	assert.Empty(t, s.freshSiblings)

	checkpoint := s.front.checkpoint(1, time.Second, false)
	assert.Len(t, checkpoint.SiblingOrigins, MaxSiblingOrigins)
	assert.Contains(t, checkpoint.SiblingOrigins, "https://known-0.example.com")
	assert.NotContains(t, checkpoint.SiblingOrigins, "https://one-too-many.example.com")

	resumed := &scan{
		site:  site,
		api:   origin{},
		front: newFrontier(checkpoint, 8),
	}
	resumed.noteSibling("https://fresh-0.example.com/other")
	assert.Empty(t, resumed.freshSiblings)
}
