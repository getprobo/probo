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

package testsupport

import (
	"fmt"
	"reflect"
	"slices"
	"time"
)

type (
	// ValueType identifies one hydrated Automerge value.
	ValueType string

	// Value is a recursively hydrated Automerge value.
	Value struct {
		Type   ValueType
		Scalar Scalar
		Map    map[string]Value
		List   []Value
		Text   string
	}

	hydratedContainerID struct {
		Type     ValueType
		Pointer  uintptr
		Length   int
		Capacity int
	}

	hydratedValidationFrame struct {
		Value   Value
		Exiting bool
	}

	hydratedOperation uint8

	hydratedTask struct {
		Object    *Object
		Operation hydratedOperation
		Key       string
		Index     uint64
		Value     Value
	}
)

const (
	ValueTypeScalar ValueType = "scalar"
	ValueTypeMap    ValueType = "map"
	ValueTypeList   ValueType = "list"
	ValueTypeText   ValueType = "text"

	hydratedOperationPut    hydratedOperation = 1
	hydratedOperationInsert hydratedOperation = 2
	hydratedOperationPutAt  hydratedOperation = 3
)

// NewFrom creates and commits a document from a hydrated root map.
func NewFrom(
	actorID ActorID,
	value map[string]Value,
	message string,
	timestamp time.Time,
) (*Document, error) {
	return newFrom(actorID, value, message, timestamp, New)
}

// NewReferenceFrom creates a hydrated document using the Rust/WASM oracle.
func NewReferenceFrom(
	actorID ActorID,
	value map[string]Value,
	message string,
	timestamp time.Time,
) (*Document, error) {
	return newFrom(
		actorID,
		value,
		message,
		timestamp,
		NewReference,
	)
}

func newFrom(
	actorID ActorID,
	value map[string]Value,
	message string,
	timestamp time.Time,
	factory func(ActorID) (*Document, error),
) (*Document, error) {
	document, err := factory(actorID)
	if err != nil {
		return nil, err
	}

	if err := document.Root().PutMap(value); err != nil {
		_ = document.Close()
		return nil, err
	}

	commit := document.Commit
	if len(value) == 0 {
		commit = document.EmptyCommit
	}

	if _, err := commit(message, timestamp); err != nil {
		_ = document.Close()
		return nil, err
	}

	return document, nil
}

// PutMap assigns a batch of recursively hydrated map properties.
func (o *Object) PutMap(values map[string]Value) error {
	if err := validateHydratedValue(Value{Type: ValueTypeMap, Map: values}); err != nil {
		return err
	}

	return applyHydratedTasks(mapHydratedTasks(o, values))
}

// PutValue assigns one recursively hydrated value to a map property.
func (o *Object) PutValue(key string, value Value) error {
	if err := validateHydratedValue(value); err != nil {
		return err
	}

	return applyHydratedTasks(
		[]hydratedTask{{
			Object:    o,
			Operation: hydratedOperationPut,
			Key:       key,
			Value:     value,
		}},
	)
}

// InsertValues inserts recursively hydrated values into a list.
func (o *Object) InsertValues(
	index uint64,
	values []Value,
) error {
	if err := validateHydratedValue(Value{Type: ValueTypeList, List: values}); err != nil {
		return err
	}

	return applyHydratedTasks(listHydratedTasks(o, index, values))
}

// InsertValue inserts one recursively hydrated value into a list.
func (o *Object) InsertValue(
	index uint64,
	value Value,
) error {
	if err := validateHydratedValue(value); err != nil {
		return err
	}

	return applyHydratedTasks(
		[]hydratedTask{{
			Object:    o,
			Operation: hydratedOperationInsert,
			Index:     index,
			Value:     value,
		}},
	)
}

// PutValueAt replaces a list element with one recursively hydrated value.
func (o *Object) PutValueAt(
	index uint64,
	value Value,
) error {
	if err := validateHydratedValue(value); err != nil {
		return err
	}

	return applyHydratedTasks(
		[]hydratedTask{{
			Object:    o,
			Operation: hydratedOperationPutAt,
			Index:     index,
			Value:     value,
		}},
	)
}

// SpliceValues deletes and inserts recursively hydrated list values.
func (o *Object) SpliceValues(
	index uint64,
	deleteCount uint64,
	values []Value,
) error {
	if err := validateHydratedValue(Value{Type: ValueTypeList, List: values}); err != nil {
		return err
	}

	for range deleteCount {
		if err := o.DeleteIndex(index); err != nil {
			return err
		}
	}

	return applyHydratedTasks(listHydratedTasks(o, index, values))
}

func validateHydratedValue(root Value) error {
	const (
		visiting = 1
		visited  = 2
	)

	states := make(map[hydratedContainerID]uint8)
	stack := []hydratedValidationFrame{{Value: root}}

	for len(stack) > 0 {
		frame := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		value := frame.Value

		switch value.Type {
		case ValueTypeScalar:
			if !validScalarType(value.Scalar.Type) {
				return fmt.Errorf("unknown scalar type %q", value.Scalar.Type)
			}
		case ValueTypeText:
		case ValueTypeMap, ValueTypeList:
			id := hydratedValueContainerID(value)
			if id.Pointer == 0 {
				continue
			}

			if frame.Exiting {
				states[id] = visited
				continue
			}

			switch states[id] {
			case visiting:
				return fmt.Errorf("hydrated value contains a container cycle")
			case visited:
				continue
			}

			states[id] = visiting

			stack = append(stack, hydratedValidationFrame{Value: value, Exiting: true})

			if value.Type == ValueTypeMap {
				keys := sortedHydratedKeys(value.Map)
				for index := len(keys) - 1; index >= 0; index-- {
					stack = append(
						stack,
						hydratedValidationFrame{Value: value.Map[keys[index]]},
					)
				}
			} else {
				for index := len(value.List) - 1; index >= 0; index-- {
					stack = append(
						stack,
						hydratedValidationFrame{Value: value.List[index]},
					)
				}
			}
		default:
			return fmt.Errorf("unknown hydrated value type %q", value.Type)
		}
	}

	return nil
}

func hydratedValueContainerID(value Value) hydratedContainerID {
	if value.Type == ValueTypeMap {
		return hydratedContainerID{
			Type:    value.Type,
			Pointer: reflect.ValueOf(value.Map).Pointer(),
		}
	}

	return hydratedContainerID{
		Type:     value.Type,
		Pointer:  reflect.ValueOf(value.List).Pointer(),
		Length:   len(value.List),
		Capacity: cap(value.List),
	}
}

func sortedHydratedKeys(values map[string]Value) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	return keys
}

func mapHydratedTasks(object *Object, values map[string]Value) []hydratedTask {
	keys := sortedHydratedKeys(values)

	tasks := make([]hydratedTask, 0, len(keys))
	for _, key := range keys {
		tasks = append(
			tasks,
			hydratedTask{
				Object:    object,
				Operation: hydratedOperationPut,
				Key:       key,
				Value:     values[key],
			},
		)
	}

	return tasks
}

func listHydratedTasks(object *Object, index uint64, values []Value) []hydratedTask {
	tasks := make([]hydratedTask, 0, len(values))
	for offset, value := range values {
		tasks = append(
			tasks,
			hydratedTask{
				Object:    object,
				Operation: hydratedOperationInsert,
				Index:     index + uint64(offset),
				Value:     value,
			},
		)
	}

	return tasks
}

func applyHydratedTasks(tasks []hydratedTask) error {
	for current := 0; current < len(tasks); current++ {
		task := tasks[current]

		child, err := applyHydratedTask(task)
		if err != nil {
			return err
		}

		if child == nil {
			continue
		}

		switch task.Value.Type {
		case ValueTypeMap:
			tasks = append(tasks, mapHydratedTasks(child, task.Value.Map)...)
		case ValueTypeList:
			tasks = append(tasks, listHydratedTasks(child, 0, task.Value.List)...)
		}
	}

	return nil
}

func applyHydratedTask(task hydratedTask) (*Object, error) {
	switch task.Value.Type {
	case ValueTypeScalar:
		switch task.Operation {
		case hydratedOperationPut:
			return nil, task.Object.PutScalar(task.Key, task.Value.Scalar)
		case hydratedOperationInsert:
			return nil, task.Object.InsertScalar(task.Index, task.Value.Scalar)
		case hydratedOperationPutAt:
			return nil, task.Object.PutScalarAt(task.Index, task.Value.Scalar)
		}
	case ValueTypeMap, ValueTypeList, ValueTypeText:
		objectType := ObjectType(task.Value.Type)

		var (
			child *Object
			err   error
		)

		switch task.Operation {
		case hydratedOperationPut:
			child, err = task.Object.CreateObject(task.Key, objectType)
		case hydratedOperationInsert:
			child, err = task.Object.InsertObject(task.Index, objectType)
		case hydratedOperationPutAt:
			child, err = task.Object.putObjectAt(task.Index, objectType)
		}

		if err != nil {
			return nil, err
		}

		if task.Value.Type == ValueTypeText {
			text := &Text{document: child.document, handle: child.handle}
			if err := text.Splice(0, 0, task.Value.Text); err != nil {
				return nil, err
			}
		}

		return child, nil
	}

	panic("validated hydrated task has invalid type or operation")
}

func (o *Object) putObjectAt(
	index uint64,
	objectType ObjectType,
) (*Object, error) {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return nil, ErrClosed
	}

	if !validObjectType(objectType) {
		return nil, fmt.Errorf("unknown Automerge object type %q", objectType)
	}

	handle, err := o.document.engine.PutObjectAt(
		o.handle,
		index,
		string(objectType),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot replace Automerge object: %w", err)
	}

	return &Object{
		document: o.document,
		handle:   handle,
		Type:     objectType,
	}, nil
}
