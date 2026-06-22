// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"slices"

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

	// OpenAI rejects schemas where required does not list every key in
	// properties.  Promote optional properties into required and mark them
	// nullable so the model knows it may pass null.
	data, err = normalizeRequiredJSON(data)
	if err != nil {
		return nil, fmt.Errorf("cannot normalize schema for %s: %w", t, err)
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

// normalizeRequiredJSON ensures every property key in an object schema also
// appears in its required array.  Properties that were not originally required
// are made nullable (their "type" becomes ["T","null"]) so the LLM knows it
// may pass null for them.  The transformation is applied recursively so nested
// object schemas are also normalised.
func normalizeRequiredJSON(data []byte) ([]byte, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return data, nil
	}

	propsRaw, hasProps := obj["properties"]
	if !hasProps {
		if itemsRaw, ok := obj["items"]; ok {
			n, err := normalizeRequiredJSON(itemsRaw)
			if err != nil {
				return nil, err
			}

			obj["items"] = n
		}

		if addlRaw, ok := obj["additionalProperties"]; ok {
			n, err := normalizeRequiredJSON(addlRaw)
			if err != nil {
				return nil, err
			}

			obj["additionalProperties"] = n
		}

		return json.Marshal(obj)
	}

	var props map[string]json.RawMessage
	if err := json.Unmarshal(propsRaw, &props); err != nil {
		return data, nil
	}

	var required []string
	if reqRaw, ok := obj["required"]; ok {
		_ = json.Unmarshal(reqRaw, &required)
	}

	requiredSet := make(map[string]bool, len(required))
	for _, r := range required {
		requiredSet[r] = true
	}

	for name, propRaw := range props {
		n, err := normalizeRequiredJSON(propRaw)
		if err != nil {
			return nil, err
		}

		if !requiredSet[name] {
			n, err = makeNullableJSON(n)
			if err != nil {
				return nil, err
			}

			required = append(required, name)
			requiredSet[name] = true
		}

		props[name] = n
	}

	propsData, err := json.Marshal(props)
	if err != nil {
		return nil, err
	}

	obj["properties"] = propsData

	if len(required) > 0 {
		reqData, err := json.Marshal(required)
		if err != nil {
			return nil, err
		}

		obj["required"] = reqData
	}

	return json.Marshal(obj)
}

// makeNullableJSON adds "null" to the "type" field of a JSON Schema object so
// that the LLM understands it may pass null for optional properties.
func makeNullableJSON(data []byte) ([]byte, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return data, nil
	}

	typeRaw, ok := obj["type"]
	if !ok {
		return data, nil
	}

	var single string
	if err := json.Unmarshal(typeRaw, &single); err == nil {
		if single != "null" {
			arr, _ := json.Marshal([]string{single, "null"})
			obj["type"] = arr
		}

		return json.Marshal(obj)
	}

	var arr []string
	if err := json.Unmarshal(typeRaw, &arr); err == nil {
		if slices.Contains(arr, "null") {
			return data, nil
		}

		arr = append(arr, "null")
		nullable, _ := json.Marshal(arr)
		obj["type"] = nullable

		return json.Marshal(obj)
	}

	return data, nil
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
