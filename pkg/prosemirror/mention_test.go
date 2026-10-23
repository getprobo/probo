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

package prosemirror

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
)

func mentionDoc(mentions ...string) string {
	var b strings.Builder

	b.WriteString(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hi "}`)

	for _, m := range mentions {
		b.WriteString(`,`)
		b.WriteString(m)
	}

	b.WriteString(`]}]}`)

	return b.String()
}

func mentionNode(id string, label string) string {
	return `{"type":"mention","attrs":{"id":"` + id + `","label":"` + label + `"}}`
}

func TestSanitizeDocumentJSON_Mention(t *testing.T) {
	t.Parallel()

	profileID := gid.New(gid.TenantID{}, 3).String()

	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "valid mention keeps id and trimmed label",
			in:   mentionDoc(`{"type":"mention","attrs":{"id":"` + profileID + `","label":" Jane Doe ","mentionSuggestionChar":"@"}}`),
			want: mentionDoc(mentionNode(profileID, "Jane Doe")),
		},
		{
			name:    "invalid id",
			in:      mentionDoc(mentionNode("not-a-gid", "Jane")),
			wantErr: true,
		},
		{
			name:    "empty label",
			in:      mentionDoc(mentionNode(profileID, "  ")),
			wantErr: true,
		},
		{
			name:    "mention with children",
			in:      mentionDoc(`{"type":"mention","attrs":{"id":"` + profileID + `","label":"Jane"},"content":[{"type":"text","text":"x"}]}`),
			wantErr: true,
		},
		{
			name:    "mention as block",
			in:      `{"type":"doc","content":[` + mentionNode(profileID, "Jane") + `]}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := SanitizeDocumentJSON(tt.in)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.JSONEq(t, tt.want, got)
		})
	}
}

func TestRenderMention(t *testing.T) {
	t.Parallel()

	profileID := gid.New(gid.TenantID{}, 3).String()

	node, err := Parse(mentionDoc(mentionNode(profileID, "Jane <Doe>")))
	require.NoError(t, err)

	t.Run("html", func(t *testing.T) {
		t.Parallel()

		got, err := RenderHTML(node)
		require.NoError(t, err)
		assert.Equal(
			t,
			`<p>Hi <span class="mention" data-type="mention" data-id="`+profileID+`" data-label="Jane &lt;Doe&gt;">@Jane &lt;Doe&gt;</span></p>`,
			got,
		)
	})

	t.Run("markdown", func(t *testing.T) {
		t.Parallel()

		got, err := RenderMarkdown(node)
		require.NoError(t, err)
		assert.Contains(t, got, "Hi @Jane")
	})
}

func TestMentionedProfileIDs(t *testing.T) {
	t.Parallel()

	first := gid.New(gid.TenantID{}, 3)
	second := gid.New(gid.TenantID{}, 3)

	t.Run("empty content", func(t *testing.T) {
		t.Parallel()

		ids, err := MentionedProfileIDs("")
		require.NoError(t, err)
		assert.Empty(t, ids)
	})

	t.Run("deduplicates in document order", func(t *testing.T) {
		t.Parallel()

		content := mentionDoc(
			mentionNode(second.String(), "B"),
			mentionNode(first.String(), "A"),
			mentionNode(second.String(), "B"),
		)

		ids, err := MentionedProfileIDs(content)
		require.NoError(t, err)
		assert.Equal(t, []gid.GID{second, first}, ids)
	})

	t.Run("newly mentioned", func(t *testing.T) {
		t.Parallel()

		oldContent := mentionDoc(mentionNode(first.String(), "A"))
		newContent := mentionDoc(
			mentionNode(first.String(), "A"),
			mentionNode(second.String(), "B"),
		)

		ids, err := NewlyMentionedProfileIDs(oldContent, newContent)
		require.NoError(t, err)
		assert.Equal(t, []gid.GID{second}, ids)
	})
}
