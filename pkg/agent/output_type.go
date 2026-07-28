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
	"fmt"

	"go.probo.inc/probo/pkg/llm"
)

// OutputType describes a structured output schema that the agent should
// produce. Build one with NewOutputType[T]().
type OutputType struct {
	Name   string
	Schema json.RawMessage
}

func NewOutputType[T any](name string) (*OutputType, error) {
	schema, err := jsonSchemaFor[T]()
	if err != nil {
		return nil, fmt.Errorf("cannot create output type %q: %w", name, err)
	}

	return &OutputType{
		Name:   name,
		Schema: schema,
	}, nil
}

func (o *OutputType) responseFormat() *llm.ResponseFormat {
	return &llm.ResponseFormat{
		Type: llm.ResponseFormatJSONSchema,
		JSONSchema: &llm.JSONSchema{
			Name:   o.Name,
			Schema: o.Schema,
			Strict: true,
		},
	}
}

// DecorateEnum injects an explicit `enum` constraint on a single
// top-level property of the schema. jsonschema-go reads struct tags as
// free-form descriptions only, so enums cannot be encoded on the tag
// itself; callers chain one call per enum field after NewOutputType.
func (o *OutputType) DecorateEnum(field string, values []string) error {
	var schema map[string]any
	if err := json.Unmarshal(o.Schema, &schema); err != nil {
		return fmt.Errorf("cannot unmarshal output type schema: %w", err)
	}

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return fmt.Errorf("output type schema has no properties")
	}

	prop, ok := properties[field].(map[string]any)
	if !ok {
		return fmt.Errorf("output type schema has no %q property", field)
	}

	prop["enum"] = values

	decorated, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("cannot marshal decorated output type schema: %w", err)
	}

	o.Schema = decorated

	return nil
}
