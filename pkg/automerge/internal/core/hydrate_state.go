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
	"fmt"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"sort"
	"strings"
)

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

	for _, operation := range s.operations {
		if operation.Object.IsRoot ||
			operation.Object.OpID != object ||
			operation.Key.Property == nil ||
			s.isSuperseded(operation.ID) {
			continue
		}

		property := *operation.Key.Property
		properties[property] = append(properties[property], operation)
	}

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
