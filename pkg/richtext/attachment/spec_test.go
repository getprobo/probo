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

package attachment

import (
	"encoding/json"
	"fmt"
	stdhtml "html"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	pmhtml "go.probo.inc/probo/pkg/prosemirror/html"
	"go.probo.inc/probo/pkg/prosemirror/markdown"
)

var (
	schemaHTML = pmhtml.NewRenderer(
		schema,
		append(
			[]pmhtml.NodeRenderer{pmhtml.Doc(), pmhtml.Paragraph(), pmhtml.Text()},
			HTMLNodes()...,
		),
		pmhtml.Marks(),
	)
	schemaMarkdown = markdown.NewRenderer(
		schema,
		schemaHTML,
		append(
			[]markdown.NodeRenderer{markdown.Doc(), markdown.Paragraph(), markdown.Text()},
			MarkdownNodes()...,
		),
		markdown.Marks(),
	)
)

var schema = prosemirror.NewSchema(
	[]prosemirror.NodeSpec{
		prosemirror.Doc(),
		prosemirror.Paragraph(),
		prosemirror.Text(),
		Image(),
		File(),
	},
	nil,
)

func TestValidateDocumentContentJSON_Schema(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.TenantID{}, 25).String()

	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{
			name:    "valid image block",
			in:      `{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/img.png"}}]}`,
			wantErr: false,
		},
		{
			name:    "image with empty file id",
			in:      `{"type":"doc","content":[{"type":"image","attrs":{"fileId":"","src":"https://example.com/img.png"}}]}`,
			wantErr: false,
		},
		{
			name:    "image with whitespace file id",
			in:      `{"type":"doc","content":[{"type":"image","attrs":{"fileId":"   ","src":"https://example.com/img.png"}}]}`,
			wantErr: false,
		},
		{
			name:    "image with invalid file id",
			in:      `{"type":"doc","content":[{"type":"image","attrs":{"fileId":"not a file"}}]}`,
			wantErr: true,
		},
		{
			name:    "image with non-gid file id",
			in:      `{"type":"doc","content":[{"type":"image","attrs":{"fileId":"abc"}}]}`,
			wantErr: true,
		},
		{
			name:    "image with padded file id",
			in:      fmt.Sprintf(`{"type":"doc","content":[{"type":"image","attrs":{"fileId":" %s "}}]}`, fileID),
			wantErr: false,
		},
		{
			name:    "image with gid file id",
			in:      fmt.Sprintf(`{"type":"doc","content":[{"type":"image","attrs":{"fileId":"%s"}}]}`, fileID),
			wantErr: false,
		},
		{
			name:    "image inside paragraph",
			in:      `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"image","attrs":{"src":"https://example.com/img.png"}}]}]}`,
			wantErr: true,
		},
		{
			name:    "valid file block",
			in:      fmt.Sprintf(`{"type":"doc","content":[{"type":"file","attrs":{"fileId":"%s","fileName":"policy.pdf","mimeType":"application/pdf","size":12}}]}`, fileID),
			wantErr: false,
		},
		{
			name:    "file block without id",
			in:      `{"type":"doc","content":[{"type":"file","attrs":{"fileId":"","fileName":"policy.pdf","mimeType":"application/pdf","size":12}}]}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := schema.ValidateJSON(tt.in)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestSanitizeDocumentJSON_ImageWidth(t *testing.T) {
	t.Parallel()

	kept, err := schema.Sanitize(`{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/img.png","width":50}}]}`)
	require.NoError(t, err)

	doc, err := prosemirror.Parse(kept)
	require.NoError(t, err)

	attrs, err := doc.Content[0].ImageAttrs()
	require.NoError(t, err)
	require.NotNil(t, attrs.Width)
	assert.Equal(t, 50, *attrs.Width)

	full, err := schema.Sanitize(`{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/img.png","width":100}}]}`)
	require.NoError(t, err)

	doc, err = prosemirror.Parse(full)
	require.NoError(t, err)

	attrs, err = doc.Content[0].ImageAttrs()
	require.NoError(t, err)
	require.NotNil(t, attrs.Width)
	assert.Equal(t, 100, *attrs.Width)

	err = schema.ValidateJSON(`{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/img.png","width":0}}]}`)
	require.Error(t, err)

	err = schema.ValidateJSON(`{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/img.png","width":101}}]}`)
	require.Error(t, err)
}

func TestSanitizeDocumentJSON_ImageSrc(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"image","attrs":{"src":"javascript:alert(1)","alt":"xss"}}]}`

	out, err := schema.Sanitize(raw)
	require.NoError(t, err)

	var doc prosemirror.Node
	require.NoError(t, json.Unmarshal([]byte(out), &doc))
	img := doc.Content[0]
	attrs, err := img.ImageAttrs()
	require.NoError(t, err)
	assert.Equal(t, "", attrs.Src)
	require.NotNil(t, attrs.Alt)
	assert.Equal(t, "xss", *attrs.Alt)
}

func TestSanitizeDocumentJSON_PreservesSafeImageSrc(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/img.png","alt":"ok"}}]}`

	out, err := schema.Sanitize(raw)
	require.NoError(t, err)

	var doc prosemirror.Node
	require.NoError(t, json.Unmarshal([]byte(out), &doc))
	img := doc.Content[0]
	attrs, err := img.ImageAttrs()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/img.png", attrs.Src)
}

func TestSanitizeDocumentJSON_DropsDataImageSrc(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"image","attrs":{"src":"data:image/png;base64,iVBOR","alt":"preview"}}]}`

	out, err := schema.Sanitize(raw)
	require.NoError(t, err)

	var doc prosemirror.Node
	require.NoError(t, json.Unmarshal([]byte(out), &doc))
	attrs, err := doc.Content[0].ImageAttrs()
	require.NoError(t, err)
	assert.Empty(t, attrs.Src)
}

func TestSanitizeDocumentJSON_BlankImageFileIDIsAbsent(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"image","attrs":{"fileId":"  ","src":"https://example.com/img.png"}}]}`

	out, err := schema.Sanitize(raw)
	require.NoError(t, err)

	var doc prosemirror.Node
	require.NoError(t, json.Unmarshal([]byte(out), &doc))
	attrs, err := doc.Content[0].ImageAttrs()
	require.NoError(t, err)
	assert.Nil(t, attrs.FileID)
	assert.Equal(t, "https://example.com/img.png", attrs.Src)
}

func TestSanitizeDocumentJSON_ImageFileIDClearsSrc(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.TenantID{}, 25).String()
	raw := fmt.Sprintf(
		`{"type":"doc","content":[{"type":"image","attrs":{"fileId":"%s","src":"data:image/png;base64,iVBOR"}}]}`,
		fileID,
	)

	out, err := schema.Sanitize(raw)
	require.NoError(t, err)

	var doc prosemirror.Node
	require.NoError(t, json.Unmarshal([]byte(out), &doc))
	attrs, err := doc.Content[0].ImageAttrs()
	require.NoError(t, err)
	require.NotNil(t, attrs.FileID)
	assert.Equal(t, fileID, *attrs.FileID)
	assert.Empty(t, attrs.Src)
}

func TestSanitizeDocumentJSON_TrimsImageFileID(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.TenantID{}, 25).String()
	raw := fmt.Sprintf(
		`{"type":"doc","content":[{"type":"image","attrs":{"fileId":" %s ","src":"https://example.com/img.png"}}]}`,
		fileID,
	)

	out, err := schema.Sanitize(raw)
	require.NoError(t, err)

	var doc prosemirror.Node
	require.NoError(t, json.Unmarshal([]byte(out), &doc))
	attrs, err := doc.Content[0].ImageAttrs()
	require.NoError(t, err)
	require.NotNil(t, attrs.FileID)
	assert.Equal(t, fileID, *attrs.FileID)
	assert.Empty(t, attrs.Src)
}

func TestRenderHTML_Image(t *testing.T) {
	t.Parallel()

	raw := `{"type":"image","attrs":{"src":"https://example.com/img.png","alt":"A photo","title":"My image"}}`

	var n prosemirror.Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := schemaHTML.Render(n)
	require.NoError(t, err)
	assert.Equal(t, `<img src="https://example.com/img.png" alt="A photo" title="My image">`, got)
}

func TestRenderHTML_File(t *testing.T) {
	t.Parallel()

	raw := `{"type":"file","attrs":{"fileId":"file-1","fileName":"policy.pdf","mimeType":"application/pdf","size":12}}`

	var n prosemirror.Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := schemaHTML.Render(n)
	require.NoError(t, err)
	assert.Equal(t, `<a href="/api/files/v1/file-1" class="document-file">policy.pdf</a>`, got)
}

func TestRenderHTML_ImageWidth(t *testing.T) {
	t.Parallel()

	raw := `{"type":"image","attrs":{"src":"https://example.com/img.png","alt":"A photo","width":50}}`

	var n prosemirror.Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := schemaHTML.Render(n)
	require.NoError(t, err)
	assert.Equal(t, `<img src="https://example.com/img.png" alt="A photo" width="50%">`, got)
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

				var n prosemirror.Node
				require.NoError(t, json.Unmarshal([]byte(raw), &n))

				got, err := schemaHTML.Render(n)
				require.NoError(t, err)

				want := fmt.Sprintf(`<img src="%s">`, stdhtml.EscapeString(tc.wantSrc))
				assert.Equal(t, want, got)
			},
		)
	}
}

func TestRenderHTML_NodesWithoutAttrs(t *testing.T) {
	t.Parallel()

	var n prosemirror.Node
	require.NoError(t, json.Unmarshal([]byte(`{"type":"image"}`), &n))

	got, err := schemaHTML.Render(n)
	require.NoError(t, err)
	assert.Equal(t, `<img src="">`, got)
}

func TestRenderMarkdown_Image(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"image","attrs":{"src":"https://example.com/img.png","alt":"A photo","title":"My image"}}]}]}`

	var n prosemirror.Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := schemaMarkdown.Render(n)
	require.NoError(t, err)
	assert.Equal(t, "![A photo](https://example.com/img.png \"My image\")\n", got)
}

func TestRenderMarkdown_ImageWithoutTitle(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"image","attrs":{"src":"https://example.com/img.png","alt":"A photo","title":null}}]}]}`

	var n prosemirror.Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := schemaMarkdown.Render(n)
	require.NoError(t, err)
	assert.Equal(t, "![A photo](https://example.com/img.png)\n", got)
}

func TestRenderMarkdown_ImageWidth(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/img.png","alt":"A photo","width":50}}]}`

	var n prosemirror.Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := schemaMarkdown.Render(n)
	require.NoError(t, err)
	assert.Equal(t, "<img src=\"https://example.com/img.png\" alt=\"A photo\" width=\"50%\">\n", got)

	full := `{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/img.png","alt":"A photo","width":100}}]}`
	require.NoError(t, json.Unmarshal([]byte(full), &n))

	got, err = schemaMarkdown.Render(n)
	require.NoError(t, err)
	assert.Equal(t, "<img src=\"https://example.com/img.png\" alt=\"A photo\" width=\"100%\">\n", got)

	titled := `{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/img.png","alt":"A photo","title":"My image","width":50}}]}`
	require.NoError(t, json.Unmarshal([]byte(titled), &n))

	got, err = schemaMarkdown.Render(n)
	require.NoError(t, err)
	assert.Equal(t, "<img src=\"https://example.com/img.png\" alt=\"A photo\" title=\"My image\" width=\"50%\">\n", got)

	dangerous := `{"type":"doc","content":[{"type":"image","attrs":{"src":"javascript:alert(1)","alt":null,"title":null,"width":50}}]}`
	require.NoError(t, json.Unmarshal([]byte(dangerous), &n))

	got, err = schemaMarkdown.Render(n)
	require.NoError(t, err)
	assert.Equal(t, "<img src=\"\" width=\"50%\">\n", got)
}

func TestRenderMarkdown_ImageSanitizesDangerousSrc(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"image","attrs":{"src":"javascript:alert(1)","alt":null,"title":null}}]}]}`

	var n prosemirror.Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := schemaMarkdown.Render(n)
	require.NoError(t, err)
	assert.Equal(t, "![]()\n", got)
}

func TestRenderMarkdown_FileEscapesAmpersand(t *testing.T) {
	t.Parallel()

	raw := `{"type":"file","attrs":{"fileId":"file-1","fileName":"report &amp;.pdf","mimeType":"application/pdf","size":12}}`

	var n prosemirror.Node
	require.NoError(t, json.Unmarshal([]byte(raw), &n))

	got, err := schemaMarkdown.Render(n)
	require.NoError(t, err)
	assert.Equal(t, "[report \\&amp;.pdf](/api/files/v1/file-1)\n", got)
}
