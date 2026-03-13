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
)

func TestGenerateSchema_PointerFieldsStripNull(t *testing.T) {
	t.Parallel()

	type Params struct {
		Name  *string  `json:"name"`
		Count *int     `json:"count"`
		Score *float64 `json:"score"`
		Done  *bool    `json:"done"`
	}

	raw := jsonSchemaFor[Params]()

	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))

	props := schema["properties"].(map[string]any)

	for _, field := range []struct {
		name     string
		wantType string
	}{
		{"name", "string"},
		{"count", "integer"},
		{"score", "number"},
		{"done", "boolean"},
	} {
		prop := props[field.name].(map[string]any)
		assert.Equal(t, field.wantType, prop["type"], "field %s", field.name)
		assert.Nil(t, prop["types"], "field %s should not have union types", field.name)
	}
}

func TestGenerateSchema_IntegerBoundsStripped(t *testing.T) {
	t.Parallel()

	type Params struct {
		Int8   int8   `json:"int8"`
		Int16  int16  `json:"int16"`
		Int32  int32  `json:"int32"`
		Uint8  uint8  `json:"uint8"`
		Uint16 uint16 `json:"uint16"`
	}

	raw := jsonSchemaFor[Params]()

	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))

	props := schema["properties"].(map[string]any)

	for _, name := range []string{"int8", "int16", "int32", "uint8", "uint16"} {
		prop := props[name].(map[string]any)
		assert.Equal(t, "integer", prop["type"], "field %s", name)
		assert.Nil(t, prop["minimum"], "field %s should have no minimum", name)
		assert.Nil(t, prop["maximum"], "field %s should have no maximum", name)
	}
}

func TestGenerateSchema_EmptyStruct(t *testing.T) {
	t.Parallel()

	type Params struct{}

	raw := jsonSchemaFor[Params]()

	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))

	assert.Equal(t, "object", schema["type"])

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok, "empty struct should have a properties field")
	assert.Empty(t, props)
}

func TestGenerateSchema_MapField(t *testing.T) {
	t.Parallel()

	type Params struct {
		Metadata map[string]string `json:"metadata"`
	}

	raw := jsonSchemaFor[Params]()

	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))

	props := schema["properties"].(map[string]any)
	metaProp := props["metadata"].(map[string]any)

	assert.Equal(t, "object", metaProp["type"])

	addlProps, ok := metaProp["additionalProperties"].(map[string]any)
	require.True(t, ok, "map field should produce additionalProperties")
	assert.Equal(t, "string", addlProps["type"])
}

func TestGenerateSchema_NestedPointerStruct(t *testing.T) {
	t.Parallel()

	type Inner struct {
		Value *string `json:"value"`
	}
	type Params struct {
		Inner *Inner `json:"inner"`
	}

	raw := jsonSchemaFor[Params]()

	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))

	props := schema["properties"].(map[string]any)

	innerProp := props["inner"].(map[string]any)
	assert.Equal(t, "object", innerProp["type"])
	assert.Nil(t, innerProp["types"], "pointer to struct should not have union types")

	innerProps := innerProp["properties"].(map[string]any)
	valueProp := innerProps["value"].(map[string]any)
	assert.Equal(t, "string", valueProp["type"])
	assert.Nil(t, valueProp["types"], "nested pointer field should not have union types")
}

func TestGenerateSchema_SliceOfPointers(t *testing.T) {
	t.Parallel()

	type Params struct {
		Names []*string `json:"names"`
	}

	raw := jsonSchemaFor[Params]()

	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))

	props := schema["properties"].(map[string]any)

	namesProp := props["names"].(map[string]any)
	assert.Equal(t, "array", namesProp["type"])

	items := namesProp["items"].(map[string]any)
	assert.Equal(t, "string", items["type"])
	assert.Nil(t, items["types"], "array items from pointer should not have union types")
}

func TestGenerateSchema_MapWithPointerValues(t *testing.T) {
	t.Parallel()

	type Params struct {
		Scores map[string]*int `json:"scores"`
	}

	raw := jsonSchemaFor[Params]()

	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))

	props := schema["properties"].(map[string]any)

	scoresProp := props["scores"].(map[string]any)
	assert.Equal(t, "object", scoresProp["type"])

	addlProps := scoresProp["additionalProperties"].(map[string]any)
	assert.Equal(t, "integer", addlProps["type"])
	assert.Nil(t, addlProps["types"], "map pointer values should not have union types")
	assert.Nil(t, addlProps["minimum"])
	assert.Nil(t, addlProps["maximum"])
}

func TestGenerateSchema_DescriptionFromJsonschemaTag(t *testing.T) {
	t.Parallel()

	type Params struct {
		Query string `json:"query" jsonschema:"The search query to execute"`
		Limit int    `json:"limit" jsonschema:"Maximum number of results to return"`
	}

	raw := jsonSchemaFor[Params]()

	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))

	props := schema["properties"].(map[string]any)

	queryProp := props["query"].(map[string]any)
	assert.Equal(t, "The search query to execute", queryProp["description"])

	limitProp := props["limit"].(map[string]any)
	assert.Equal(t, "Maximum number of results to return", limitProp["description"])
}

func TestGenerateSchema_RequiredVsOptional(t *testing.T) {
	t.Parallel()

	type Params struct {
		Required   string  `json:"required"`
		Optional   *string `json:"optional,omitempty"`
		OmitEmpty  string  `json:"omit_empty,omitempty"`
		AlsoNeeded int     `json:"also_needed"`
	}

	raw := jsonSchemaFor[Params]()

	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))

	required := schema["required"].([]any)
	assert.Contains(t, required, "required")
	assert.Contains(t, required, "also_needed")
	assert.NotContains(t, required, "optional")
	assert.NotContains(t, required, "omit_empty")
}

func TestGenerateSchema_DeeplyNestedStructure(t *testing.T) {
	t.Parallel()

	type Level3 struct {
		Value *int `json:"value"`
	}
	type Level2 struct {
		Items []Level3 `json:"items"`
	}
	type Level1 struct {
		Child *Level2 `json:"child"`
	}
	type Params struct {
		Root Level1 `json:"root"`
	}

	raw := jsonSchemaFor[Params]()

	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))

	props := schema["properties"].(map[string]any)

	rootProp := props["root"].(map[string]any)
	assert.Equal(t, "object", rootProp["type"])

	rootProps := rootProp["properties"].(map[string]any)
	childProp := rootProps["child"].(map[string]any)
	assert.Equal(t, "object", childProp["type"])
	assert.Nil(t, childProp["types"])

	childProps := childProp["properties"].(map[string]any)
	itemsProp := childProps["items"].(map[string]any)
	assert.Equal(t, "array", itemsProp["type"])

	itemsItems := itemsProp["items"].(map[string]any)
	assert.Equal(t, "object", itemsItems["type"])

	level3Props := itemsItems["properties"].(map[string]any)
	valueProp := level3Props["value"].(map[string]any)
	assert.Equal(t, "integer", valueProp["type"])
	assert.Nil(t, valueProp["types"])
	assert.Nil(t, valueProp["minimum"])
	assert.Nil(t, valueProp["maximum"])
}

func TestGenerateSchema_SliceOfStructs(t *testing.T) {
	t.Parallel()

	type Item struct {
		Name  string `json:"name"`
		Count *int   `json:"count,omitempty"`
	}
	type Params struct {
		Items []Item `json:"items"`
	}

	raw := jsonSchemaFor[Params]()

	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))

	props := schema["properties"].(map[string]any)

	itemsProp := props["items"].(map[string]any)
	assert.Equal(t, "array", itemsProp["type"])

	items := itemsProp["items"].(map[string]any)
	assert.Equal(t, "object", items["type"])

	itemProps := items["properties"].(map[string]any)
	assert.Contains(t, itemProps, "name")
	assert.Contains(t, itemProps, "count")

	countProp := itemProps["count"].(map[string]any)
	assert.Equal(t, "integer", countProp["type"])
	assert.Nil(t, countProp["types"])
	assert.Nil(t, countProp["minimum"])
}

func TestStripNullTypes_NilSchema(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		stripNullTypes(nil)
	})
}
