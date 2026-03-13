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
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
)

func jsonSchemaFor[T any]() (json.RawMessage, error) {
	t := reflect.TypeFor[T]()

	schema, err := jsonschema.ForType(t, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot generate schema for %s: %w", t, err)
	}

	stripNullTypes(schema)

	data, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal schema for %s: %w", t, err)
	}

	return json.RawMessage(data), nil
}

func mustJSONSchemaFor[T any]() json.RawMessage {
	schema, err := jsonSchemaFor[T]()
	if err != nil {
		panic(err)
	}
	return schema
}

// stripNullTypes removes "null" from union types produced by pointer fields
// (e.g. ["null","string"] becomes "string") and clears integer bounds so that
// LLM providers receive a clean schema without Go-specific type constraints.
func stripNullTypes(s *jsonschema.Schema) {
	if s == nil {
		return
	}

	if len(s.Types) > 0 {
		filtered := make([]string, 0, len(s.Types))
		for _, t := range s.Types {
			if t != "null" {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) == 1 {
			s.Type = filtered[0]
			s.Types = nil
		} else if len(filtered) > 1 {
			s.Types = filtered
		}
	}

	s.Minimum = nil
	s.Maximum = nil

	if s.Type == "object" && s.Properties == nil {
		s.Properties = make(map[string]*jsonschema.Schema)
	}

	for _, prop := range s.Properties {
		stripNullTypes(prop)
	}

	if s.Items != nil {
		stripNullTypes(s.Items)
	}

	if s.AdditionalProperties != nil {
		stripNullTypes(s.AdditionalProperties)
	}
}
