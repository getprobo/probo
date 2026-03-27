// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package docgen

import (
	"html/template"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestRenderHTML(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		data            DocumentData
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "basic document with all fields",
			data: DocumentData{
				Title:          "Test Document",
				Content:        `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"Main Title"}]},{"type":"paragraph","content":[{"type":"text","text":"This is "},{"type":"text","marks":[{"type":"bold"}],"text":"bold"},{"type":"text","text":" text with "},{"type":"text","marks":[{"type":"italic"}],"text":"italic"},{"type":"text","text":" formatting."}]}]}`,
				Major:          1,
				Classification: ClassificationPublic,
				Approvers:      []string{"John Doe"},
				PublishedAt:    &now,
				Signatures: []SignatureData{
					{
						SignedBy:    "Alice Smith",
						SignedAt:    &now,
						State:       coredata.DocumentVersionSignatureStateSigned,
						RequestedAt: now,
					},
				},
			},
			wantContains: []string{
				"Test Document",
				"<h1>Main Title</h1>",
				"<strong>bold</strong>",
				"<em>italic</em>",
				"<td>1.0</td>",
				"PUBLIC",
				"John Doe",
				"Alice Smith",
			},
		},
		{
			name: "document with HTML characters that need escaping",
			data: DocumentData{
				Title:     "Test & <Script> Title",
				Content:   "Normal markdown content",
				Approvers: []string{"John <script>alert('xss')</script> Doe"},
				Signatures: []SignatureData{
					{
						SignedBy: "Alice & <Bob>",
						State:    coredata.DocumentVersionSignatureStateRequested,
					},
				},
			},
			wantContains: []string{
				"Test &amp; &lt;Script&gt; Title",
				"John &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt; Doe",
				"Alice &amp; &lt;Bob&gt;",
			},
			wantNotContains: []string{
				"<script>alert('xss')</script>",
				"Test & <Script> Title",
			},
		},
		{
			name: "document with prosemirror content",
			data: DocumentData{
				Title: "ProseMirror Test",
				Content: `{"type":"doc","content":[` +
					`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Section 1"}]},` +
					`{"type":"bulletList","content":[` +
					`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Item 1"}]}]},` +
					`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Item 2"}]}]}` +
					`]},` +
					`{"type":"paragraph","content":[` +
					`{"type":"text","marks":[{"type":"bold"}],"text":"Bold text"},` +
					`{"type":"text","text":" and "},` +
					`{"type":"text","marks":[{"type":"italic"}],"text":"italic text"}` +
					`]},` +
					`{"type":"codeBlock","content":[{"type":"text","text":"code block"}]}` +
					`]}`,
			},
			wantContains: []string{
				"<h2>Section 1</h2>",
				"<ul>",
				"<li><p>Item 1</p></li>",
				"<li><p>Item 2</p></li>",
				"</ul>",
				"<strong>Bold text</strong>",
				"<em>italic text</em>",
				"<pre><code>code block</code></pre>",
			},
		},
		{
			name: "document with all classification types",
			data: DocumentData{
				Title:          "Classification Test",
				Classification: ClassificationConfidential,
			},
			wantContains: []string{"CONFIDENTIAL"},
		},
		{
			name: "empty document",
			data: DocumentData{},
			wantContains: []string{
				"<!DOCTYPE html>",
				"<html",
				"</html>",
			},
		},
		{
			name: "document with multiple signatures in different states",
			data: DocumentData{
				Title: "Signatures Test",
				Signatures: []SignatureData{
					{
						SignedBy:    "Signer 1",
						SignedAt:    &now,
						State:       coredata.DocumentVersionSignatureStateSigned,
						RequestedAt: now,
					},
					{
						SignedBy:    "Signer 2",
						State:       coredata.DocumentVersionSignatureStateRequested,
						RequestedAt: now,
					},
				},
			},
			wantContains: []string{
				"Signer 1",
				"Signer 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderHTML(tt.data)
			require.NoError(t, err)
			require.NotEmpty(t, result)

			resultStr := string(result)

			// Check that all expected content is present
			for _, want := range tt.wantContains {
				assert.Contains(t, resultStr, want, "Expected content not found: %s", want)
			}

			// Check that unwanted content is not present
			for _, wantNot := range tt.wantNotContains {
				assert.NotContains(t, resultStr, wantNot, "Unwanted content found: %s", wantNot)
			}

			// Basic HTML structure validation
			assert.Contains(t, resultStr, "<!DOCTYPE html>")
			assert.Contains(t, resultStr, "<html")
			assert.Contains(t, resultStr, "</html>")
			assert.Contains(t, resultStr, "<head>")
			assert.Contains(t, resultStr, "</head>")
			assert.Contains(t, resultStr, "<body>")
			assert.Contains(t, resultStr, "</body>")
		})
	}
}

func TestRenderHTML_ErrorHandling(t *testing.T) {
	data := DocumentData{
		Title:   "Valid Document",
		Content: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Valid content"}]}]}`,
	}

	result, err := RenderHTML(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestTemplateFunctions(t *testing.T) {
	t.Run("now function", func(t *testing.T) {
		nowFunc := templateFuncs["now"].(func() time.Time)
		result := nowFunc()
		assert.True(t, time.Since(result) < time.Second)
	})

	t.Run("eq function", func(t *testing.T) {
		eqFunc := templateFuncs["eq"].(func(any, any) bool)
		assert.True(t, eqFunc("test", "test"))
		assert.False(t, eqFunc("test", "other"))
		assert.True(t, eqFunc(0, 0))
		assert.False(t, eqFunc(0, 1))
	})

	t.Run("lower function", func(t *testing.T) {
		lowerFunc := templateFuncs["lower"].(func(string) string)
		assert.Equal(t, "hello world", lowerFunc("HELLO WORLD"))
		assert.Equal(t, "test", lowerFunc("Test"))
	})

	t.Run("classificationString function", func(t *testing.T) {
		classFunc := templateFuncs["classificationString"].(func(Classification) string)
		assert.Equal(t, "PUBLIC", classFunc(ClassificationPublic))
		assert.Equal(t, "CONFIDENTIAL", classFunc(ClassificationConfidential))
	})

	t.Run("ProseMirrorJSONToHTML", func(t *testing.T) {
		result := ProseMirrorJSONToHTML(
			`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"plain "},{"type":"text","marks":[{"type":"bold"}],"text":"bold"}]}]}`,
		)
		assert.Contains(t, string(result), "<strong>bold</strong>")

		result = ProseMirrorJSONToHTML(
			`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"simple text"}]}]}`,
		)
		assert.Contains(t, string(result), "<p>simple text</p>")

		result = ProseMirrorJSONToHTML("**not** json")
		assert.Contains(t, string(result), "<p>**not** json</p>")

		result = ProseMirrorJSONToHTML("")
		assert.Equal(t, template.HTML(""), result)
	})
}

func TestClassificationConstants(t *testing.T) {
	tests := []struct {
		name           string
		classification Classification
		expected       string
	}{
		{"public", ClassificationPublic, "PUBLIC"},
		{"internal", ClassificationInternal, "INTERNAL"},
		{"confidential", ClassificationConfidential, "CONFIDENTIAL"},
		{"secret", ClassificationSecret, "SECRET"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.classification))
		})
	}
}

func TestHTMLEscaping(t *testing.T) {
	dangerousData := DocumentData{
		Title:     "<script>alert('xss')</script>",
		Approvers: []string{"User & <Company>"},
		Signatures: []SignatureData{
			{
				SignedBy: "<malicious>tag",
				State:    coredata.DocumentVersionSignatureStateRequested,
			},
		},
	}

	result, err := RenderHTML(dangerousData)
	require.NoError(t, err)

	resultStr := string(result)

	// Verify dangerous content is escaped
	assert.NotContains(t, resultStr, "<script>alert('xss')</script>")
	assert.NotContains(t, resultStr, "<malicious>tag")
	assert.Contains(t, resultStr, "&lt;script&gt;")
	assert.Contains(t, resultStr, "&amp;")
	assert.Contains(t, resultStr, "&#39;")
}

func TestProseMirrorContentRendering(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name: "headers",
			content: `{"type":"doc","content":[` +
				`{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"H1"}]},` +
				`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"H2"}]},` +
				`{"type":"heading","attrs":{"level":3},"content":[{"type":"text","text":"H3"}]}` +
				`]}`,
			want: []string{"<h1>H1</h1>", "<h2>H2</h2>", "<h3>H3</h3>"},
		},
		{
			name: "emphasis",
			content: `{"type":"doc","content":[{"type":"paragraph","content":[` +
				`{"type":"text","marks":[{"type":"bold"}],"text":"bold"},` +
				`{"type":"text","text":" and "},` +
				`{"type":"text","marks":[{"type":"italic"}],"text":"italic"}` +
				`]}]}`,
			want: []string{"<strong>bold</strong>", "<em>italic</em>"},
		},
		{
			name: "lists",
			content: `{"type":"doc","content":[{"type":"bulletList","content":[` +
				`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Item 1"}]}]},` +
				`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Item 2"}]}]}` +
				`]}]}`,
			want: []string{"<ul>", "<li><p>Item 1</p></li>", "<li><p>Item 2</p></li>", "</ul>"},
		},
		{
			name: "paragraphs",
			content: `{"type":"doc","content":[` +
				`{"type":"paragraph","content":[{"type":"text","text":"Paragraph 1"}]},` +
				`{"type":"paragraph","content":[{"type":"text","text":"Paragraph 2"}]}` +
				`]}`,
			want: []string{"<p>Paragraph 1</p>", "<p>Paragraph 2</p>"},
		},
		{
			name: "code",
			content: `{"type":"doc","content":[` +
				`{"type":"paragraph","content":[{"type":"text","marks":[{"type":"code"}],"text":"inline code"}]},` +
				`{"type":"paragraph","content":[{"type":"text","text":" and "}]},` +
				`{"type":"codeBlock","content":[{"type":"text","text":"code block"}]}` +
				`]}`,
			want: []string{"<code>inline code</code>", "<pre><code>code block</code></pre>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := DocumentData{
				Title:   "ProseMirror Test",
				Content: tt.content,
			}

			result, err := RenderHTML(data)
			require.NoError(t, err)

			resultStr := string(result)
			for _, want := range tt.want {
				assert.Contains(t, resultStr, want)
			}
		})
	}
}

func TestDocumentVersionSignatureStates(t *testing.T) {
	now := time.Now()

	states := []coredata.DocumentVersionSignatureState{
		coredata.DocumentVersionSignatureStateRequested,
		coredata.DocumentVersionSignatureStateSigned,
		// Add other states if they exist
	}

	for _, state := range states {
		t.Run(string(state), func(t *testing.T) {
			data := DocumentData{
				Title: "State Test",
				Signatures: []SignatureData{
					{
						SignedBy:    "Test User",
						State:       state,
						RequestedAt: now,
					},
				},
			}

			result, err := RenderHTML(data)
			assert.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}

func TestLargeContent(t *testing.T) {
	var largeContent strings.Builder
	largeContent.WriteString(`{"type":"doc","content":[`)
	for i := range 1000 {
		if i > 0 {
			largeContent.WriteByte(',')
		}
		largeContent.WriteString(`{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"Section `)
		largeContent.WriteByte(byte('A' + i%26))
		largeContent.WriteString(`"}]},`)
		largeContent.WriteString(`{"type":"paragraph","content":[` +
			`{"type":"text","text":"This is a paragraph with "},` +
			`{"type":"text","marks":[{"type":"bold"}],"text":"bold"},` +
			`{"type":"text","text":" and "},` +
			`{"type":"text","marks":[{"type":"italic"}],"text":"italic"},` +
			`{"type":"text","text":" text."}` +
			`]},`)
		largeContent.WriteString(`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"List item 1"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"List item 2"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"List item 3"}]}]}` +
			`]}`)
	}
	largeContent.WriteString(`]}`)

	data := DocumentData{
		Title:   "Large Document",
		Content: largeContent.String(),
	}

	result, err := RenderHTML(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.True(t, len(result) > 10000) // Should be reasonably large
}

func BenchmarkGenerateHTML(b *testing.B) {
	now := time.Now()

	data := DocumentData{
		Title: "Benchmark Document",
		Content: `{"type":"doc","content":[` +
			`{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"Title"}]},` +
			`{"type":"paragraph","content":[` +
			`{"type":"text","text":"This is "},` +
			`{"type":"text","marks":[{"type":"bold"}],"text":"bold"},` +
			`{"type":"text","text":" text with "},` +
			`{"type":"text","marks":[{"type":"italic"}],"text":"italic"},` +
			`{"type":"text","text":" formatting."}` +
			`]},` +
			`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Item 1"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Item 2"}]}]}` +
			`]}` +
			`]}`,
		Major:          1,
		Classification: ClassificationPublic,
		Approvers:      []string{"John Doe"},
		PublishedAt:    &now,
		Signatures: []SignatureData{
			{
				SignedBy:    "Alice Smith",
				SignedAt:    &now,
				State:       coredata.DocumentVersionSignatureStateSigned,
				RequestedAt: now,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := RenderHTML(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
