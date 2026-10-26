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

package api

import (
	"encoding/json"
	"errors"
	"fmt"
)

// NodeConnection reads a Relay connection from data.node.<field>.
// notFound is returned when node is null. A different __typename returns
// "expected <expectedType> node, got <actual>".
func NodeConnection[T any](field, expectedType, notFound string) func(json.RawMessage) (*Connection[T], error) {
	return func(data json.RawMessage) (*Connection[T], error) {
		return nodeConnection[T](data, field, expectedType, notFound, false)
	}
}

// NodeConnectionOrMissing is NodeConnection, except a different __typename
// also returns notFound. Lists that choose the parent type at runtime use it.
func NodeConnectionOrMissing[T any](field, expectedType, notFound string) func(json.RawMessage) (*Connection[T], error) {
	return func(data json.RawMessage) (*Connection[T], error) {
		return nodeConnection[T](data, field, expectedType, notFound, true)
	}
}

// NodeNestedConnection reads a Relay connection from data.node.<parent>.<field>.
// notFound is returned when node is null. missingParent is returned when
// parent is null or absent.
func NodeNestedConnection[T any](parent, field, expectedType, notFound, missingParent string) func(json.RawMessage) (*Connection[T], error) {
	return func(data json.RawMessage) (*Connection[T], error) {
		node, err := nodeObject(data, expectedType, notFound, false)
		if err != nil {
			return nil, err
		}

		var fields map[string]json.RawMessage
		if err := json.Unmarshal(node, &fields); err != nil {
			return nil, err
		}

		raw, ok := fields[parent]
		if !ok || len(raw) == 0 || string(raw) == "null" {
			return nil, errors.New(missingParent)
		}

		return connectionField[T](raw, field)
	}
}

// ViewerConnection reads a Relay connection from data.viewer.<field>.
// A null viewer is an empty connection.
func ViewerConnection[T any](field string) func(json.RawMessage) (*Connection[T], error) {
	return func(data json.RawMessage) (*Connection[T], error) {
		var envelope map[string]json.RawMessage
		if err := json.Unmarshal(data, &envelope); err != nil {
			return nil, err
		}

		raw := envelope["viewer"]
		if len(raw) == 0 || string(raw) == "null" {
			return &Connection[T]{}, nil
		}

		return connectionField[T](raw, field)
	}
}

// RootConnection reads a Relay connection from a top-level data field.
func RootConnection[T any](field string) func(json.RawMessage) (*Connection[T], error) {
	return func(data json.RawMessage) (*Connection[T], error) {
		return connectionField[T](data, field)
	}
}

// NodeSlice reads a JSON array from data.node.<field>.
func NodeSlice[T any](data json.RawMessage, field, expectedType, notFound string) ([]T, error) {
	node, err := nodeObject(data, expectedType, notFound, false)
	if err != nil {
		return nil, err
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(node, &fields); err != nil {
		return nil, err
	}

	raw, ok := fields[field]
	if !ok || len(raw) == 0 || string(raw) == "null" {
		if !ok {
			return nil, fmt.Errorf("response has no %s field", field)
		}

		return []T{}, nil
	}

	var items []T
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}

	if items == nil {
		items = []T{}
	}

	return items, nil
}

func nodeConnection[T any](data json.RawMessage, field, expectedType, notFound string, mismatchIsNotFound bool) (*Connection[T], error) {
	node, err := nodeObject(data, expectedType, notFound, mismatchIsNotFound)
	if err != nil {
		return nil, err
	}

	return connectionField[T](node, field)
}

func nodeObject(data json.RawMessage, expectedType, notFound string, mismatchIsNotFound bool) (json.RawMessage, error) {
	var envelope struct {
		Node json.RawMessage `json:"node"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}

	if len(envelope.Node) == 0 || string(envelope.Node) == "null" {
		return nil, errors.New(notFound)
	}

	var header struct {
		Typename string `json:"__typename"`
	}
	if err := json.Unmarshal(envelope.Node, &header); err != nil {
		return nil, err
	}

	if header.Typename != expectedType {
		if mismatchIsNotFound {
			return nil, errors.New(notFound)
		}

		return nil, fmt.Errorf("expected %s node, got %s", expectedType, header.Typename)
	}

	return envelope.Node, nil
}

func connectionField[T any](object json.RawMessage, field string) (*Connection[T], error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(object, &fields); err != nil {
		return nil, err
	}

	raw, ok := fields[field]
	if !ok || len(raw) == 0 || string(raw) == "null" {
		if !ok {
			return nil, fmt.Errorf("response has no %s connection", field)
		}

		return &Connection[T]{}, nil
	}

	var conn Connection[T]
	if err := json.Unmarshal(raw, &conn); err != nil {
		return nil, err
	}

	return &conn, nil
}
