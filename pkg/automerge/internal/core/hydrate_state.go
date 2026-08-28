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

package core

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

type hydratedValueWire struct {
	Type   string                       `json:"type"`
	Scalar json.RawMessage              `json:"scalar,omitempty"`
	Map    map[string]hydratedValueWire `json:"map,omitempty"`
	List   []hydratedValueWire          `json:"list,omitempty"`
	Text   string                       `json:"text,omitempty"`
}

// Hydrate returns the current root value as a recursively typed value.
func (b *Engine) Hydrate() ([]byte, error) {
	b.bindColumnarState()
	return hydrateState(b.state)
}

// Rescue decodes a document and returns its current value while bypassing only
// strict mark-order validation. The returned value does not preserve history.
func Rescue(data []byte) ([]byte, error) {
	document, err := storage.DecodeRescue(data)
	if err != nil {
		return nil, fmt.Errorf("cannot decode native rescue document: %w", err)
	}

	state, err := newRescueStateFromDocument(document)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize native rescue state: %w", err)
	}

	return hydrateState(state)
}

func hydrateState(state *State) ([]byte, error) {
	value, err := state.hydratedMapValue(
		opset.RootObject(),
		make(map[opset.OpID]struct{}),
	)
	if err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("cannot encode hydrated value: %w", err)
	}

	return encoded, nil
}

func (s *State) hydratedMapValue(
	object opset.ObjectID,
	visited map[opset.OpID]struct{},
) (hydratedValueWire, error) {
	if !object.IsRoot {
		if _, ok := visited[object.OpID]; ok {
			return hydratedValueWire{}, fmt.Errorf("object cycle detected")
		}

		visited[object.OpID] = struct{}{}
		defer delete(visited, object.OpID)
	}

	properties := make(map[string][]opset.Operation)

	s.eachOperation(func(operation opset.Operation) bool {
		if operation.Object != object ||
			operation.Key.Property == nil ||
			s.isSuperseded(operation.ID) {
			return true
		}

		property := *operation.Key.Property
		properties[property] = append(properties[property], operation)
		return true
	})

	result := hydratedValueWire{
		Type: "map",
		Map:  make(map[string]hydratedValueWire, len(properties)),
	}

	for property, operations := range properties {
		sort.Slice(
			operations,
			func(i, j int) bool {
				return operations[i].ID.Compare(operations[j].ID) > 0
			},
		)

		value, err := s.hydratedOperationValue(operations[0], visited)
		if err != nil {
			return hydratedValueWire{}, err
		}

		result.Map[property] = value
	}

	return result, nil
}

func (s *State) hydratedListValue(
	object opset.OpID,
	visited map[opset.OpID]struct{},
) (hydratedValueWire, error) {
	if _, ok := visited[object]; ok {
		return hydratedValueWire{}, fmt.Errorf("object cycle detected")
	}

	visited[object] = struct{}{}
	defer delete(visited, object)

	elements := s.sequenceElements(object)
	result := hydratedValueWire{
		Type: "list",
		List: make([]hydratedValueWire, 0, len(elements)),
	}

	for _, element := range elements {
		value, err := s.hydratedOperationValue(element, visited)
		if err != nil {
			return hydratedValueWire{}, err
		}

		result.List = append(result.List, value)
	}

	return result, nil
}

func (s *State) hydratedOperationValue(
	operation opset.Operation,
	visited map[opset.OpID]struct{},
) (hydratedValueWire, error) {
	switch operation.Action {
	case opset.ActionMakeMap, opset.ActionMakeTable:
		return s.hydratedMapValue(
			opset.ObjectID{OpID: operation.ID},
			visited,
		)
	case opset.ActionMakeList:
		return s.hydratedListValue(operation.ID, visited)
	case opset.ActionMakeText:
		var value strings.Builder

		for _, element := range s.sequence(operation.ID) {
			if element.Value != nil && element.Value.Type == opset.ScalarString {
				value.WriteString(element.Value.String)
			}
		}

		return hydratedValueWire{Type: "text", Text: value.String()}, nil
	case opset.ActionSet:
		if operation.Value == nil {
			return hydratedValueWire{}, fmt.Errorf(
				"set operation %v has no value",
				operation.ID,
			)
		}

		encoded, err := encodeScalarWire(*operation.Value)
		if err != nil {
			return hydratedValueWire{}, err
		}

		return hydratedValueWire{
			Type:   "scalar",
			Scalar: json.RawMessage(encoded),
		}, nil
	default:
		return hydratedValueWire{}, fmt.Errorf(
			"operation %v does not carry a hydrated value",
			operation.ID,
		)
	}
}

func (s *State) mapValue(
	object opset.OpID,
	visited map[opset.OpID]struct{},
) (map[string]any, error) {
	if _, ok := visited[object]; ok {
		return nil, fmt.Errorf("object cycle detected")
	}

	visited[object] = struct{}{}
	defer delete(visited, object)

	properties := make(map[string][]opset.Operation)

	s.eachOperation(func(operation opset.Operation) bool {
		if operation.Object.IsRoot ||
			operation.Object.OpID != object ||
			operation.Key.Property == nil ||
			s.isSuperseded(operation.ID) {
			return true
		}

		property := *operation.Key.Property
		properties[property] = append(properties[property], operation)
		return true
	})

	result := make(map[string]any, len(properties))
	for property, operations := range properties {
		sort.Slice(
			operations,
			func(i, j int) bool {
				return operations[i].ID.Compare(operations[j].ID) > 0
			},
		)

		operation := operations[0]
		switch operation.Action {
		case opset.ActionMakeMap:
			value, err := s.mapValue(operation.ID, visited)
			if err != nil {
				return nil, err
			}

			result[property] = value
		case opset.ActionMakeList:
			value, err := s.listValue(operation.ID, visited)
			if err != nil {
				return nil, err
			}

			result[property] = value
		case opset.ActionMakeText:
			var value strings.Builder

			for _, element := range s.sequence(operation.ID) {
				if element.Value != nil && element.Value.Type == opset.ScalarString {
					value.WriteString(element.Value.String)
				}
			}

			result[property] = value.String()
		case opset.ActionSet:
			result[property] = scalarMaterializedValue(operation.Value)
		}
	}

	return result, nil
}

func (s *State) listValue(
	object opset.OpID,
	visited map[opset.OpID]struct{},
) ([]any, error) {
	if _, ok := visited[object]; ok {
		return nil, fmt.Errorf("object cycle detected")
	}

	visited[object] = struct{}{}
	defer delete(visited, object)

	elements := s.sequenceElements(object)
	result := make([]any, 0, len(elements))

	for _, element := range elements {
		switch element.Action {
		case opset.ActionMakeMap:
			value, err := s.mapValue(element.ID, visited)
			if err != nil {
				return nil, err
			}

			result = append(result, value)
		case opset.ActionMakeList:
			value, err := s.listValue(element.ID, visited)
			if err != nil {
				return nil, err
			}

			result = append(result, value)
		case opset.ActionMakeText:
			var value strings.Builder

			for _, textElement := range s.sequence(element.ID) {
				if textElement.Value != nil &&
					textElement.Value.Type == opset.ScalarString {
					value.WriteString(textElement.Value.String)
				}
			}

			result = append(result, value.String())
		case opset.ActionSet:
			result = append(result, scalarMaterializedValue(element.Value))
		}
	}

	return result, nil
}

func scalarMaterializedValue(value *opset.Scalar) any {
	if value == nil {
		return nil
	}

	switch value.Type {
	case opset.ScalarNull:
		return nil
	case opset.ScalarFalse:
		return false
	case opset.ScalarTrue:
		return true
	case opset.ScalarUint:
		return value.Uint
	case opset.ScalarInt, opset.ScalarCounter, opset.ScalarTimestamp:
		return value.Int
	case opset.ScalarFloat64:
		return value.Float
	case opset.ScalarString:
		return value.String
	case opset.ScalarBytes:
		return append([]byte(nil), value.Bytes...)
	default:
		return append([]byte(nil), value.Raw...)
	}
}
