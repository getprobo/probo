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

package agent

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSchema_PointerFieldsStripNull(t *testing.T) {
	t.Parallel()

	// Pointer fields without omitempty are treated as required by the schema
	// library. null is stripped and they appear as simple non-nullable types.
	type Params struct {
		Name  *string  `json:"name"`
		Count *int     `json:"count"`
		Score *float64 `json:"score"`
		Done  *bool    `json:"done"`
	}

	raw, err := jsonSchemaFor[Params]()
	require.NoError(t, err)

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

	raw, err := jsonSchemaFor[Params]()
	require.NoError(t, err)

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

	raw, err := jsonSchemaFor[Params]()
	require.NoError(t, err)

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

	raw, err := jsonSchemaFor[Params]()
	require.NoError(t, err)

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

	raw, err := jsonSchemaFor[Params]()
	require.NoError(t, err)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))

	props := schema["properties"].(map[string]any)

	// Pointer-to-struct without omitempty is in required with a simple type.
	innerProp := props["inner"].(map[string]any)
	assert.Equal(t, "object", innerProp["type"])
	assert.Nil(t, innerProp["types"], "pointer to struct should not have union types")

	// Nested pointer field without omitempty is also required and non-nullable.
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

	raw, err := jsonSchemaFor[Params]()
	require.NoError(t, err)

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

	raw, err := jsonSchemaFor[Params]()
	require.NoError(t, err)

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

	raw, err := jsonSchemaFor[Params]()
	require.NoError(t, err)

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

	raw, err := jsonSchemaFor[Params]()
	require.NoError(t, err)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))

	// All properties must appear in required (OpenAI requirement).
	required := schema["required"].([]any)
	assert.Contains(t, required, "required")
	assert.Contains(t, required, "also_needed")
	assert.Contains(t, required, "optional")
	assert.Contains(t, required, "omit_empty")

	// Optional properties must be nullable so the model can pass null.
	props := schema["properties"].(map[string]any)

	optionalType := props["optional"].(map[string]any)["type"].([]any)
	assert.Contains(t, optionalType, "string")
	assert.Contains(t, optionalType, "null")

	omitEmptyType := props["omit_empty"].(map[string]any)["type"].([]any)
	assert.Contains(t, omitEmptyType, "string")
	assert.Contains(t, omitEmptyType, "null")

	// Truly required properties must not be nullable.
	requiredType := props["required"].(map[string]any)["type"]
	assert.Equal(t, "string", requiredType)

	alsoNeededType := props["also_needed"].(map[string]any)["type"]
	assert.Equal(t, "integer", alsoNeededType)
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

	raw, err := jsonSchemaFor[Params]()
	require.NoError(t, err)

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

	raw, err := jsonSchemaFor[Params]()
	require.NoError(t, err)

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
	countType := countProp["type"].([]any)
	assert.Contains(t, countType, "integer")
	assert.Contains(t, countType, "null")
	assert.Nil(t, countProp["minimum"])
}

func TestStripNullTypes_NilSchema(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		stripNullTypes(nil)
	})
}
