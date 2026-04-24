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

// DecorateEnum injects explicit `enum` constraints on top-level
// properties of the schema. jsonschema-go only reads struct tags as
// free-form descriptions, so enums cannot be encoded on the tag itself;
// callers use this helper after NewOutputType to lock down string
// fields whose allowed values live in a package-level slice.
func (o *OutputType) DecorateEnum(enums map[string][]string) error {
	if len(enums) == 0 {
		return nil
	}

	var schema map[string]any
	if err := json.Unmarshal(o.Schema, &schema); err != nil {
		return fmt.Errorf("cannot unmarshal output type schema: %w", err)
	}

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return fmt.Errorf("output type schema has no properties")
	}

	for field, values := range enums {
		prop, ok := properties[field].(map[string]any)
		if !ok {
			return fmt.Errorf("output type schema has no %q property", field)
		}
		prop["enum"] = values
	}

	decorated, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("cannot marshal decorated output type schema: %w", err)
	}
	o.Schema = decorated
	return nil
}
