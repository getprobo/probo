// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package prosemirror

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderHTML_Document(t *testing.T) {
	t.Parallel()

	expected, err := os.ReadFile("testdata/document.html")
	require.NoError(t, err)

	doc := loadTestDocument(t)
	got, err := RenderHTML(doc)
	require.NoError(t, err)
	assert.Equal(t, string(expected), got)
}

func TestRenderHTML_EmptyParagraph(t *testing.T) {
	t.Parallel()

	node := Node{Type: NodeParagraph}
	got, err := RenderHTML(Node{Type: NodeDoc, Content: []Node{node}})
	require.NoError(t, err)
	assert.Equal(t, "<p></p>", got)
}

func TestRenderHTML_HeadingLevels(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		level    int
		expected string
	}{
		{1, "<h1>X</h1>"},
		{2, "<h2>X</h2>"},
		{3, "<h3>X</h3>"},
		{4, "<h4>X</h4>"},
		{5, "<h5>X</h5>"},
		{6, "<h6>X</h6>"},
	} {
		t.Run(
			"level "+string(rune('0'+tc.level)),
			func(t *testing.T) {
				t.Parallel()
				raw := `{"type":"heading","attrs":{"level":` + string(rune('0'+tc.level)) + `},"content":[{"type":"text","text":"X"}]}`
				var n Node
				require.NoError(t, json.Unmarshal([]byte(raw), &n))

				got, err := RenderHTML(n)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, got)
			},
		)
	}
}

func TestRenderHTML_HeadingInvalidLevel(t *testing.T) {
	t.Parallel()

	raw := `{"type":"heading","attrs":{"level":7},"content":[{"type":"text","text":"X"}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	_, err := RenderHTML(n)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid level")
}

func TestRenderHTML_CodeBlockWithLanguage(t *testing.T) {
	t.Parallel()

	raw := `{"type":"codeBlock","attrs":{"language":"go"},"content":[{"type":"text","text":"fmt.Println()"}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderHTML(n)
	require.NoError(t, err)
	assert.Equal(t, `<pre><code class="language-go">fmt.Println()</code></pre>`, got)
}

func TestRenderHTML_CodeBlockMermaid(t *testing.T) {
	t.Parallel()

	raw := `{"type":"codeBlock","attrs":{"language":"mermaid"},"content":[{"type":"text","text":"graph TD\n  A-->B"}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderHTML(n)
	require.NoError(t, err)
	assert.Equal(t, "<pre class=\"mermaid\">graph TD\n  A--&gt;B</pre>", got)
}

func TestRenderHTML_CodeBlockWithoutLanguage(t *testing.T) {
	t.Parallel()

	raw := `{"type":"codeBlock","attrs":{"language":null},"content":[{"type":"text","text":"hello"}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderHTML(n)
	require.NoError(t, err)
	assert.Equal(t, "<pre><code>hello</code></pre>", got)
}

func TestRenderHTML_OrderedListWithStart(t *testing.T) {
	t.Parallel()

	raw := `{"type":"orderedList","attrs":{"start":5,"type":null},"content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"item"}]}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderHTML(n)
	require.NoError(t, err)
	assert.Equal(t, `<ol start="5"><li><p>item</p></li></ol>`, got)
}

func TestRenderHTML_TableCellColspan(t *testing.T) {
	t.Parallel()

	raw := `{"type":"tableCell","attrs":{"colspan":2,"rowspan":1,"colwidth":null},"content":[{"type":"paragraph","content":[{"type":"text","text":"wide"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderHTML(n)
	require.NoError(t, err)
	assert.Equal(t, `<td colspan="2"><p>wide</p></td>`, got)
}

func TestRenderHTML_TableCellColwidth(t *testing.T) {
	t.Parallel()

	raw := `{"type":"tableCell","attrs":{"colspan":1,"rowspan":1,"colwidth":[100]},"content":[{"type":"paragraph","content":[{"type":"text","text":"X"}]}]}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderHTML(n)
	require.NoError(t, err)
	assert.Equal(t, `<td style="min-width: 100px"><p>X</p></td>`, got)
}

func TestRenderHTML_HTMLEscaping(t *testing.T) {
	t.Parallel()

	text := `<script>alert("xss")</script> & more`
	node := Node{
		Type: NodeParagraph,
		Content: []Node{
			{Type: NodeText, Text: &text},
		},
	}
	got, err := RenderHTML(Node{Type: NodeDoc, Content: []Node{node}})
	require.NoError(t, err)
	assert.Equal(t, `<p>&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt; &amp; more</p>`, got)
}

func TestRenderHTML_LinkAllAttrs(t *testing.T) {
	t.Parallel()

	raw := `{"type":"text","marks":[{"type":"link","attrs":{"href":"https://example.com","target":"_blank","rel":"noopener","class":"btn","title":"Click"}}],"text":"hi"}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderHTML(n)
	require.NoError(t, err)
	assert.Equal(t, `<a href="https://example.com" target="_blank" rel="noopener" class="btn" title="Click">hi</a>`, got)
}

func TestRenderHTML_LinkMinimalAttrs(t *testing.T) {
	t.Parallel()

	raw := `{"type":"text","marks":[{"type":"link","attrs":{"href":"https://example.com","target":null,"rel":null,"class":null,"title":null}}],"text":"hi"}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderHTML(n)
	require.NoError(t, err)
	assert.Equal(t, `<a href="https://example.com">hi</a>`, got)
}

func TestRenderHTML_LinkBlankTargetDefaultRel(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "no rel",
			raw:  `{"type":"text","marks":[{"type":"link","attrs":{"href":"https://example.com","target":"_blank","rel":null}}],"text":"hi"}`,
			want: `<a href="https://example.com" target="_blank" rel="noopener noreferrer">hi</a>`,
		},
		{
			name: "case insensitive target",
			raw:  `{"type":"text","marks":[{"type":"link","attrs":{"href":"https://example.com","target":"_BLANK"}}],"text":"hi"}`,
			want: `<a href="https://example.com" target="_BLANK" rel="noopener noreferrer">hi</a>`,
		},
		{
			name: "whitespace rel treated as absent",
			raw:  `{"type":"text","marks":[{"type":"link","attrs":{"href":"https://example.com","target":" _blank ","rel":"  "}}],"text":"hi"}`,
			want: `<a href="https://example.com" target=" _blank " rel="noopener noreferrer">hi</a>`,
		},
		{
			name: "custom rel without noopener gets noopener appended",
			raw:  `{"type":"text","marks":[{"type":"link","attrs":{"href":"https://example.com","target":"_blank","rel":"nofollow"}}],"text":"hi"}`,
			want: `<a href="https://example.com" target="_blank" rel="nofollow noopener">hi</a>`,
		},
		{
			name: "custom rel already has noopener",
			raw:  `{"type":"text","marks":[{"type":"link","attrs":{"href":"https://example.com","target":"_blank","rel":"noopener nofollow"}}],"text":"hi"}`,
			want: `<a href="https://example.com" target="_blank" rel="noopener nofollow">hi</a>`,
		},
		{
			name: "custom rel with noopener case insensitive",
			raw:  `{"type":"text","marks":[{"type":"link","attrs":{"href":"https://example.com","target":"_blank","rel":"NoOpener"}}],"text":"hi"}`,
			want: `<a href="https://example.com" target="_blank" rel="NoOpener">hi</a>`,
		},
		{
			name: "custom rel without blank target unchanged",
			raw:  `{"type":"text","marks":[{"type":"link","attrs":{"href":"https://example.com","target":"_self","rel":"nofollow"}}],"text":"hi"}`,
			want: `<a href="https://example.com" target="_self" rel="nofollow">hi</a>`,
		},
	} {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()
				var n Node
				require.NoError(t, json.Unmarshal([]byte(tc.raw), &n))

				got, err := RenderHTML(n)
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			},
		)
	}
}

func TestRenderHTML_LinkSanitizesDangerousHrefs(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		href     string
		wantHref string
	}{
		{name: "javascript scheme", href: `javascript:alert(1)`, wantHref: `#`},
		{name: "javascript scheme case insensitive", href: `javaScript:alert(1)`, wantHref: `#`},
		{name: "data html", href: `data:text/html,<script>alert(1)</script>`, wantHref: `#`},
		{name: "protocol-relative", href: `//evil.example/phish`, wantHref: `#`},
		{name: "path with leading slash-slash", href: `//not-a-path`, wantHref: `#`},
		{name: "empty href", href: ``, wantHref: `#`},
		{name: "fragment only", href: `#section`, wantHref: `#section`},
		{name: "relative path", href: `docs/page`, wantHref: `docs/page`},
		{name: "absolute path", href: `/app/foo`, wantHref: `/app/foo`},
		{name: "mailto", href: `mailto:user@example.com`, wantHref: `mailto:user@example.com`},
		{name: "tel", href: `tel:+15551212`, wantHref: `tel:+15551212`},
		{name: "https preserved", href: `https://example.com/x`, wantHref: `https://example.com/x`},
	} {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()
				hrefJSON, err := json.Marshal(tc.href)
				require.NoError(t, err)
				raw := fmt.Sprintf(
					`{"type":"text","marks":[{"type":"link","attrs":{"href":%s,"target":null,"rel":null,"class":null,"title":null}}],"text":"x"}`,
					string(hrefJSON),
				)
				var n Node
				require.NoError(t, json.Unmarshal([]byte(raw), &n))

				got, err := RenderHTML(n)
				require.NoError(t, err)
				want := fmt.Sprintf(`<a href="%s">x</a>`, html.EscapeString(tc.wantHref))
				assert.Equal(t, want, got)
			},
		)
	}
}

func TestRenderHTML_Image(t *testing.T) {
	t.Parallel()

	raw := `{"type":"image","attrs":{"src":"https://example.com/img.png","alt":"A photo","title":"My image"}}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderHTML(n)
	require.NoError(t, err)
	assert.Equal(t, `<img src="https://example.com/img.png" alt="A photo" title="My image">`, got)
}

func TestRenderHTML_ImageSanitizesDangerousSrc(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		src     string
		wantSrc string
	}{
		{name: "javascript scheme", src: `javascript:alert(1)`, wantSrc: ``},
		{name: "javascript case insensitive", src: `javaScript:alert(1)`, wantSrc: ``},
		{name: "vbscript scheme", src: `vbscript:MsgBox("xss")`, wantSrc: ``},
		{name: "protocol-relative", src: `//evil.example/img.png`, wantSrc: ``},
		{name: "empty src", src: ``, wantSrc: ``},
		{name: "https preserved", src: `https://example.com/img.png`, wantSrc: `https://example.com/img.png`},
		{name: "http preserved", src: `http://example.com/img.png`, wantSrc: `http://example.com/img.png`},
		{name: "data URI preserved", src: `data:image/png;base64,iVBOR`, wantSrc: `data:image/png;base64,iVBOR`},
		{name: "absolute path", src: `/images/photo.png`, wantSrc: `/images/photo.png`},
		{name: "relative path", src: `images/photo.png`, wantSrc: `images/photo.png`},
	} {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()
				srcJSON, err := json.Marshal(tc.src)
				require.NoError(t, err)
				raw := fmt.Sprintf(
					`{"type":"image","attrs":{"src":%s}}`,
					string(srcJSON),
				)
				var n Node
				require.NoError(t, json.Unmarshal([]byte(raw), &n))

				got, err := RenderHTML(n)
				require.NoError(t, err)
				want := fmt.Sprintf(`<img src="%s">`, html.EscapeString(tc.wantSrc))
				assert.Equal(t, want, got)
			},
		)
	}
}

func TestRenderHTML_MultipleMarks(t *testing.T) {
	t.Parallel()

	raw := `{"type":"text","marks":[{"type":"bold"},{"type":"italic"}],"text":"hello"}`
	var n Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := RenderHTML(n)
	require.NoError(t, err)
	assert.Equal(t, "<strong><em>hello</em></strong>", got)
}

func TestRenderHTML_UnknownNodeType(t *testing.T) {
	t.Parallel()

	node := Node{Type: NodeType("unknownWidget")}
	_, err := RenderHTML(node)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown type")
}

func TestRenderHTML_UnknownMarkType(t *testing.T) {
	t.Parallel()

	text := "hello"
	node := Node{
		Type: NodeText,
		Text: &text,
		Marks: []Mark{
			{Type: MarkType("superscript")},
		},
	}
	_, err := RenderHTML(node)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown type")
}
