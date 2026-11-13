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

package client

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeDescriptionFiles(t *testing.T) {
	t.Parallel()

	const (
		fileURL  = "https://uploads.linear.app/org/file/evidence.pdf"
		imageURL = "https://uploads.linear.app/org/file/shot.png"
	)

	document := `{
		"type":"doc",
		"content":[
			{"type":"paragraph","content":[
				{"type":"text","text":"See the file","marks":[
					{"type":"link","attrs":{"href":"https://example.com/notes"}}
				]}
			]},
			{"type":"file","attrs":{"href":"` + fileURL + `","name":"evidence].pdf","mimetype":"application/pdf","uploadState":"finished"}},
			{"type":"image","attrs":{"src":"` + imageURL + `","alt":"screenshot"}}
		]
	}`

	encoded, err := json.Marshal(document)
	require.NoError(t, err)

	tests := []struct {
		name        string
		description string
		data        json.RawMessage
		want        string
	}{
		{
			name:        "prosemirror document",
			description: "Evidence is attached.",
			data:        json.RawMessage(document),
			want: "Evidence is attached.\n" +
				"[evidence.pdf](" + fileURL + ")\n" +
				"![screenshot](" + imageURL + ")\n",
		},
		{
			name:        "json string document",
			description: "Evidence is attached.",
			data:        encoded,
			want: "Evidence is attached.\n" +
				"[evidence.pdf](" + fileURL + ")\n" +
				"![screenshot](" + imageURL + ")\n",
		},
		{
			name:        "url already in the description",
			description: "See [evidence.pdf](" + fileURL + ")",
			data:        json.RawMessage(document),
			want: "See [evidence.pdf](" + fileURL + ")\n" +
				"![screenshot](" + imageURL + ")\n",
		},
		{
			name:        "invalid document",
			description: "Evidence is attached.",
			data:        json.RawMessage("not-json"),
			want:        "Evidence is attached.",
		},
		{
			name:        "null document",
			description: "Evidence is attached.",
			data:        json.RawMessage("null"),
			want:        "Evidence is attached.",
		},
		{
			name:        "empty document",
			description: "Evidence is attached.",
			data:        nil,
			want:        "Evidence is attached.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, MergeDescriptionFiles(tt.description, tt.data))
		})
	}
}

func TestMergeDescriptionFiles_LinkMark(t *testing.T) {
	t.Parallel()

	const fileURL = "https://uploads.linear.app/org/file/evidence.pdf"

	document := `{
		"type":"doc",
		"content":[
			{"type":"paragraph","content":[
				{"type":"text","text":"evidence.pdf","marks":[
					{"type":"link","attrs":{"href":"` + fileURL + `"}}
				]}
			]}
		]
	}`

	got := MergeDescriptionFiles("See the attachment.", json.RawMessage(document))
	assert.Equal(t, "See the attachment.\n[evidence.pdf]("+fileURL+")\n", got)
}

func TestAssetURLs_TrimsTrailingPunctuation(t *testing.T) {
	t.Parallel()

	const (
		fileURL  = "https://uploads.linear.app/org/file/evidence.pdf"
		imageURL = "https://uploads.linear.app/org/file/shot.png"
	)

	got := AssetURLs("See " + fileURL + ". And *" + imageURL + "* too.")
	assert.Equal(t, []string{fileURL, imageURL}, got)
	assert.Equal(t, []string{fileURL + "."}, AssetURLs("[notes]("+fileURL+".)"))
	assert.True(t, UploadURL(fileURL+"."))

	appended := appendMissingUploadLinks("", "See "+fileURL+".")
	assert.Equal(t, "[evidence.pdf]("+fileURL+")\n", appended)

	already := appendMissingUploadLinks("See "+fileURL+".", "See "+fileURL+".")
	assert.Equal(t, "See "+fileURL+".", already)
}

func TestMergeDescriptionFiles_KeepsDistinctPrefixURL(t *testing.T) {
	t.Parallel()

	const (
		longer  = "https://uploads.linear.app/org/file/file.v2"
		shorter = "https://uploads.linear.app/org/file/file"
	)

	document := `{"type":"doc","content":[{"type":"file","attrs":{"href":"` + shorter + `","name":"file"}}]}`
	got := MergeDescriptionFiles("See "+longer, json.RawMessage(document))
	assert.Contains(t, got, longer)
	assert.Contains(t, got, "[file]("+shorter+")")
}

func TestAppendMissingUploadLinks_SanitizesLabel(t *testing.T) {
	t.Parallel()

	const fileURL = "https://uploads.linear.app/org/file/uuid"

	got := appendMissingUploadLinks("", "[notes].pdf]("+fileURL+")")
	assert.Equal(t, "[notes.pdf]("+fileURL+")\n", got)

	wrapped := appendMissingUploadLinks("", "[line\nname]("+fileURL+")")
	assert.Equal(t, "[line name]("+fileURL+")\n", wrapped)
}

func TestAppendMissingUploadLinks(t *testing.T) {
	t.Parallel()

	const fileURL = "https://uploads.linear.app/org/file/evidence.pdf"

	got := appendMissingUploadLinks(
		"Evidence is attached.",
		"Evidence is attached.\n\n[evidence.pdf]("+fileURL+")",
	)
	assert.Equal(t, "Evidence is attached.\n[evidence.pdf]("+fileURL+")\n", got)

	unchanged := appendMissingUploadLinks("Evidence is attached.", "")
	assert.Equal(t, "Evidence is attached.", unchanged)
}

func TestParseEnvelopeIssueData_DescriptionFiles(t *testing.T) {
	t.Parallel()

	const fileURL = "https://uploads.linear.app/org/file/evidence.pdf"

	body := []byte(`{
		"action":"update",
		"type":"Issue",
		"data":{
			"id":"issue-1",
			"title":"Hello",
			"description":"Evidence is attached.",
			"descriptionData":"{\"type\":\"doc\",\"content\":[{\"type\":\"file\",\"attrs\":{\"href\":\"` + fileURL + `\",\"name\":\"evidence.pdf\",\"mimetype\":\"application/pdf\"}}]}"
		}
	}`)

	envelope, err := ParseEnvelope(body)
	require.NoError(t, err)

	data, err := envelope.IssueData()
	require.NoError(t, err)
	assert.Equal(t, "Evidence is attached.\n[evidence.pdf]("+fileURL+")\n", data.Description)
	assert.True(t, data.Has("description"))
	assert.Nil(t, data.DescriptionData)
}

func TestParseEnvelopeIssueData_SkipsDescriptionDataWithoutDescription(t *testing.T) {
	t.Parallel()

	body := []byte(`{
		"action":"update",
		"type":"Issue",
		"data":{
			"id":"issue-1",
			"title":"Hello",
			"descriptionData":"{\"type\":\"doc\",\"content\":[{\"type\":\"file\",\"attrs\":{\"href\":\"https://uploads.linear.app/org/file/evidence.pdf\",\"name\":\"evidence.pdf\"}}]}"
		}
	}`)

	envelope, err := ParseEnvelope(body)
	require.NoError(t, err)

	data, err := envelope.IssueData()
	require.NoError(t, err)
	assert.Empty(t, data.Description)
	assert.False(t, data.Has("description"))
}
