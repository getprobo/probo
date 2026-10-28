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

package asana

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/task/sync/internal"
)

func TestSplitAsanaProfileText_KeepsTextBeforeABarePrefix(t *testing.T) {
	t.Parallel()

	origin := "https://app.asana.com"
	text := "keep-before " + origin + "/0/profile/ not-a-gid " + origin + "/0/profile/42 tail"
	node := internal.TextNode(text, nil)

	out, changed := splitAsanaProfileText(node, origin, resolveAda)
	require.True(t, changed)

	mention, err := internal.NewMentionNode(prosemirror.MentionAttrs{ID: "profile", Label: "Ada"})
	require.NoError(t, err)
	assert.Equal(
		t,
		[]prosemirror.Node{
			internal.TextNode("keep-before "+origin+"/0/profile/ not-a-gid ", nil),
			mention,
			internal.TextNode(" tail", nil),
		},
		out,
	)
}

func TestSplitAsanaProfileText_KeepsUnresolvedNumericGID(t *testing.T) {
	t.Parallel()

	origin := "https://app.asana.com"
	unresolved := origin + "/0/profile/7"
	node := internal.TextNode(unresolved+" and "+origin+"/0/profile/42 tail", nil)

	out, changed := splitAsanaProfileText(node, origin, resolveAda)
	require.True(t, changed)

	mention, err := internal.NewMentionNode(prosemirror.MentionAttrs{ID: "profile", Label: "Ada"})
	require.NoError(t, err)
	assert.Equal(
		t,
		[]prosemirror.Node{
			internal.TextNode(unresolved, nil),
			internal.TextNode(" and ", nil),
			mention,
			internal.TextNode(" tail", nil),
		},
		out,
	)
}

func resolveAda(userGID string) (prosemirror.MentionAttrs, bool) {
	if userGID != "42" {
		return prosemirror.MentionAttrs{}, false
	}

	return prosemirror.MentionAttrs{ID: "profile", Label: "Ada"}, true
}
