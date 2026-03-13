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

package agent

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/llm"
)

func TestNewOutputType_SetsNameAndSchema(t *testing.T) {
	t.Parallel()

	type Result struct {
		Answer string `json:"answer"`
		Score  int    `json:"score"`
	}

	ot, err := NewOutputType[Result]("test_result")
	require.NoError(t, err)

	assert.Equal(t, "test_result", ot.Name)
	require.NotNil(t, ot.Schema)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(ot.Schema, &schema))
	assert.Equal(t, "object", schema["type"])

	props := schema["properties"].(map[string]any)
	assert.Contains(t, props, "answer")
	assert.Contains(t, props, "score")
}

func TestNewOutputType_EmptyStruct(t *testing.T) {
	t.Parallel()

	type Empty struct{}

	ot, err := NewOutputType[Empty]("empty")
	require.NoError(t, err)

	assert.Equal(t, "empty", ot.Name)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(ot.Schema, &schema))
	assert.Equal(t, "object", schema["type"])

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, props)
}

func TestOutputType_responseFormat(t *testing.T) {
	t.Parallel()

	type Verdict struct {
		Approved bool   `json:"approved"`
		Reason   string `json:"reason"`
	}

	ot, err := NewOutputType[Verdict]("verdict")
	require.NoError(t, err)
	rf := ot.responseFormat()

	require.NotNil(t, rf)
	assert.Equal(t, llm.ResponseFormatJSONSchema, rf.Type)
	require.NotNil(t, rf.JSONSchema)
	assert.Equal(t, "verdict", rf.JSONSchema.Name)
	assert.True(t, rf.JSONSchema.Strict)
	assert.JSONEq(t, string(ot.Schema), string(rf.JSONSchema.Schema))
}

func TestOutputType_responseFormat_SchemaMatchesOutputType(t *testing.T) {
	t.Parallel()

	type Analysis struct {
		Summary  string   `json:"summary"`
		Tags     []string `json:"tags"`
		Priority *int     `json:"priority,omitempty"`
	}

	ot, err := NewOutputType[Analysis]("analysis")
	require.NoError(t, err)
	rf := ot.responseFormat()

	var schema map[string]any
	require.NoError(t, json.Unmarshal(rf.JSONSchema.Schema, &schema))

	props := schema["properties"].(map[string]any)
	assert.Contains(t, props, "summary")
	assert.Contains(t, props, "tags")
	assert.Contains(t, props, "priority")

	tagsProp := props["tags"].(map[string]any)
	assert.Equal(t, "array", tagsProp["type"])
}
