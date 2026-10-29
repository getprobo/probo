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

package tasksync

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
)

func TestContentToAsanaHTMLWrapsTheBody(t *testing.T) {
	t.Parallel()

	content := prosemirror.FromPlainText("Ship the sync")

	html, err := contentToAsanaHTML(content)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(html, "<body>"))
	assert.True(t, strings.HasSuffix(html, "</body>"))
	assert.Contains(t, html, "Ship the sync")
}

func TestContentToAsanaHTMLWritesMentionsAsNames(t *testing.T) {
	t.Parallel()

	profileID := gid.New(gid.TenantID{}, 3).String()
	content := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hi "},{"type":"mention","attrs":{"id":"` + profileID + `","label":"Jane <Doe>"}}]}]}`

	html, err := contentToAsanaHTML(content)
	require.NoError(t, err)
	assert.Contains(t, html, "Hi @Jane &lt;Doe&gt;")
	assert.NotContains(t, html, "data-type")
}

func TestContentToAsanaHTMLWritesResolvedMentions(t *testing.T) {
	t.Parallel()

	profileID := gid.New(gid.TenantID{}, 3).String()
	content := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hi "},{"type":"mention","attrs":{"id":"` + profileID + `","label":"Jane"}}]}]}`

	html, err := contentToAsanaHTMLWithMentions(content, map[string]string{profileID: "99"}, "https://asana.example.test")
	require.NoError(t, err)
	assert.Contains(t, html, `data-asana-gid="99"`)
	assert.Contains(t, html, `href="https://asana.example.test/0/profile/99"`)
	assert.Contains(t, html, "@Jane")
	assert.NotContains(t, html, "data-type")
}

func TestAsanaMentionAnchorBecomesAProfileURL(t *testing.T) {
	t.Parallel()

	html := asanaMentionAnchorsToURLs(`<body>Hi <a data-asana-gid="99">Jane</a></body>`, "https://asana.example.test")
	assert.Contains(t, html, "https://asana.example.test/0/profile/99")
	assert.NotContains(t, html, "data-asana-gid")
}

func TestRewriteAsanaProfileURLBecomesAMention(t *testing.T) {
	t.Parallel()

	content, err := asanaHTMLToContent(asanaMentionAnchorsToURLs(
		`<body>Hi <a data-asana-gid="99">Jane</a></body>`,
		"https://asana.example.test",
	))
	require.NoError(t, err)

	node, err := prosemirror.Parse(content)
	require.NoError(t, err)

	node, changed := rewriteAsanaMentionNodes(node, "https://asana.example.test", func(userGID string) (prosemirror.MentionAttrs, bool) {
		if userGID != "99" {
			return prosemirror.MentionAttrs{}, false
		}

		return prosemirror.MentionAttrs{ID: "profile", Label: "Jane"}, true
	})
	require.True(t, changed)

	encoded, err := json.Marshal(node)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), `"type":"mention"`)
	assert.Contains(t, string(encoded), `"label":"Jane"`)
}

func TestAsanaTaskURLUsesTheAPIHost(t *testing.T) {
	t.Parallel()

	taskURL, err := asanaTaskURL("https://asana.example.test/api/1.0", "1200", "3400")
	require.NoError(t, err)
	assert.Equal(t, "https://asana.example.test/0/1200/3400", taskURL)
}

func TestCommentHTMLPrefixesTheAuthor(t *testing.T) {
	t.Parallel()

	content := prosemirror.FromPlainText("Looks good")

	html, err := commentHTML("Ada <Admin>", content)
	require.NoError(t, err)
	assert.Contains(t, html, "Ada &lt;Admin&gt;")
	assert.Contains(t, html, "Looks good")
}
