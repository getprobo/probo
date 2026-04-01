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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeDocumentJSON_EmptyUnchanged(t *testing.T) {
	t.Parallel()

	got, err := SanitizeDocumentJSON("")
	require.NoError(t, err)
	assert.Equal(t, "", got)

	got, err = SanitizeDocumentJSON("   ")
	require.NoError(t, err)
	assert.Equal(t, "   ", got)
}

func TestSanitizeDocumentJSON_NonJSONError(t *testing.T) {
	t.Parallel()

	_, err := SanitizeDocumentJSON("plain text is not valid document JSON")
	require.Error(t, err)
}

func TestSanitizeDocumentJSON_NonDocRootError(t *testing.T) {
	t.Parallel()

	_, err := SanitizeDocumentJSON(`{"type":"paragraph","content":[]}`)
	require.Error(t, err)
}

func TestSanitizeDocumentJSON_LinkHref(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"link","attrs":{"href":"javascript:alert(1)","target":"_blank"}}],"text":"click"}]}]}`

	out, err := SanitizeDocumentJSON(raw)
	require.NoError(t, err)

	var doc Node
	require.NoError(t, json.Unmarshal([]byte(out), &doc))
	txt := doc.Content[0].Content[0]
	require.Len(t, txt.Marks, 1)
	attrs, err := txt.Marks[0].LinkAttrs()
	require.NoError(t, err)
	assert.Equal(t, "#", attrs.Href)
	require.NotNil(t, attrs.Target)
	assert.Equal(t, "_blank", *attrs.Target)
}

func TestSanitizeDocumentJSON_PreservesSafeHref(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"link","attrs":{"href":"https://example.com"}}],"text":"ok"}]}]}`

	out, err := SanitizeDocumentJSON(raw)
	require.NoError(t, err)

	var doc Node
	require.NoError(t, json.Unmarshal([]byte(out), &doc))
	txt := doc.Content[0].Content[0]
	attrs, err := txt.Marks[0].LinkAttrs()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", attrs.Href)
}

func TestSanitizeDocumentJSON_ImageSrc(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"image","attrs":{"src":"javascript:alert(1)","alt":"xss"}}]}`

	out, err := SanitizeDocumentJSON(raw)
	require.NoError(t, err)

	var doc Node
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

	out, err := SanitizeDocumentJSON(raw)
	require.NoError(t, err)

	var doc Node
	require.NoError(t, json.Unmarshal([]byte(out), &doc))
	img := doc.Content[0]
	attrs, err := img.ImageAttrs()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/img.png", attrs.Src)
}
