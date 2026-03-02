// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
				Content:        "# Main Title\n\nThis is **bold** text with *italic* formatting.",
				Version:        1,
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
				"<td>1</td>",
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
			name: "document with markdown content",
			data: DocumentData{
				Title:   "Markdown Test",
				Content: "## Section 1\n\n- Item 1\n- Item 2\n\n**Bold text** and *italic text*\n\n```code block```",
			},
			wantContains: []string{
				"<h2>Section 1</h2>",
				"<ul>",
				"<li>Item 1</li>",
				"<li>Item 2</li>",
				"</ul>",
				"<strong>Bold text</strong>",
				"<em>italic text</em>",
				"<code>code block</code>",
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
	// Test with data that should not cause errors
	data := DocumentData{
		Title:   "Valid Document",
		Content: "Valid content",
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

	t.Run("formatContent function", func(t *testing.T) {
		formatFunc := templateFuncs["formatContent"].(func(string) template.HTML)

		// Test markdown conversion
		result := formatFunc("**bold** text")
		assert.Contains(t, string(result), "<strong>bold</strong>")

		// Test basic text
		result = formatFunc("simple text")
		assert.Contains(t, string(result), "<p>simple text</p>")

		// Test empty content - goldmark produces empty output for empty input
		result = formatFunc("")
		// Empty content should produce empty result from goldmark
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

func TestMarkdownRendering(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     []string
	}{
		{
			name:     "headers",
			markdown: "# H1\n## H2\n### H3",
			want:     []string{"<h1>H1</h1>", "<h2>H2</h2>", "<h3>H3</h3>"},
		},
		{
			name:     "emphasis",
			markdown: "**bold** and *italic*",
			want:     []string{"<strong>bold</strong>", "<em>italic</em>"},
		},
		{
			name:     "lists",
			markdown: "- Item 1\n- Item 2",
			want:     []string{"<ul>", "<li>Item 1</li>", "<li>Item 2</li>", "</ul>"},
		},
		{
			name:     "paragraphs",
			markdown: "Paragraph 1\n\nParagraph 2",
			want:     []string{"<p>Paragraph 1</p>", "<p>Paragraph 2</p>"},
		},
		{
			name:     "code",
			markdown: "`inline code` and\n```\ncode block\n```",
			want:     []string{"<code>inline code</code>", "<pre><code>code block"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := DocumentData{
				Title:   "Markdown Test",
				Content: tt.markdown,
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
	// Create a large markdown content
	var largeContent strings.Builder
	for i := 0; i < 1000; i++ {
		largeContent.WriteString("# Section ")
		largeContent.WriteString(string(rune('A' + i%26)))
		largeContent.WriteString("\n\nThis is a paragraph with **bold** and *italic* text.\n\n")
		largeContent.WriteString("- List item 1\n- List item 2\n- List item 3\n\n")
	}

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
		Title:          "Benchmark Document",
		Content:        "# Title\n\nThis is **bold** text with *italic* formatting.\n\n- Item 1\n- Item 2",
		Version:        1,
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
