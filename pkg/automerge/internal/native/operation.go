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

package native

import (
	"encoding/binary"
	"fmt"
	"math"
)

type (
	Action uint8

	OpID struct {
		Counter    uint64
		ActorIndex uint64
	}

	ObjectID struct {
		Root bool
		OpID OpID
	}

	Key struct {
		Map      bool
		Property string
		Element  OpID
		Head     bool
	}

	Scalar struct {
		Type    uint8
		Value   any
		Unknown []byte
	}

	Operation struct {
		ID           OpID
		Object       ObjectID
		Key          Key
		Insert       bool
		Action       Action
		Value        Scalar
		Predecessors []OpID
		MarkExpand   bool
		MarkName     string
	}
)

const (
	ActionMakeMap Action = iota
	ActionSet
	ActionMakeList
	ActionDelete
	ActionMakeText
	ActionIncrement
	ActionMakeTable
	ActionMark
)

const (
	columnTypeGroup        = 0
	columnTypeActor        = 1
	columnTypeInteger      = 2
	columnTypeDeltaInteger = 3
	columnTypeBoolean      = 4
	columnTypeString       = 5
	columnTypeValueMeta    = 6
	columnTypeValue        = 7
)

func columnSpec(id uint32, columnType uint32) ColumnSpec {
	return ColumnSpec(id<<4 | columnType)
}

func (c *Change) Operations() ([]Operation, error) {
	objActor := decodeRLEUint(findColumn(c.Columns, columnSpec(0, columnTypeActor)))
	objCounter := decodeRLEUint(findColumn(c.Columns, columnSpec(0, columnTypeInteger)))
	keyActor := decodeRLEUint(findColumn(c.Columns, columnSpec(1, columnTypeActor)))
	keyCounter := newDeltaDecoder(findColumn(c.Columns, columnSpec(1, columnTypeDeltaInteger)))
	keyString := decodeRLEString(findColumn(c.Columns, columnSpec(1, columnTypeString)))
	insert := newBooleanDecoder(findColumn(c.Columns, columnSpec(3, columnTypeBoolean)))
	action := decodeRLEUint(findColumn(c.Columns, columnSpec(4, columnTypeInteger)))
	valueMetadata := decodeRLEUint(findColumn(c.Columns, columnSpec(5, columnTypeValueMeta)))
	valueData := newReader(findColumn(c.Columns, columnSpec(5, columnTypeValue)))
	predGroup := decodeRLEUint(findColumn(c.Columns, columnSpec(7, columnTypeGroup)))
	predActor := decodeRLEUint(findColumn(c.Columns, columnSpec(7, columnTypeActor)))
	predCounter := newDeltaDecoder(findColumn(c.Columns, columnSpec(7, columnTypeDeltaInteger)))
	markExpand := newBooleanDecoder(findColumn(c.Columns, columnSpec(9, columnTypeBoolean)))
	markName := decodeRLEString(findColumn(c.Columns, columnSpec(10, columnTypeString)))

	operations := make([]Operation, 0)
	for !action.done() {
		actionValue, null, err := action.next()
		if err != nil {
			return nil, fmt.Errorf("cannot decode operation action: %w", err)
		}
		if null {
			actionValue = uint64(ActionSet)
		}
		if actionValue > uint64(ActionMark) {
			return nil, fmt.Errorf("invalid operation action %d", actionValue)
		}

		object, err := decodeObjectID(objActor, objCounter)
		if err != nil {
			return nil, fmt.Errorf("cannot decode operation object: %w", err)
		}
		key, err := decodeKey(keyActor, keyCounter, keyString)
		if err != nil {
			return nil, fmt.Errorf("cannot decode operation key: %w", err)
		}
		id := OpID{
			Counter:    c.StartOp + uint64(len(operations)),
			ActorIndex: 0,
		}
		inserted, err := insert.next()
		if err != nil {
			return nil, fmt.Errorf("cannot decode operation insert flag: %w", err)
		}
		value, err := decodeScalar(valueMetadata, valueData)
		if err != nil {
			return nil, fmt.Errorf("cannot decode operation value: %w", err)
		}
		predecessors, err := decodePredecessors(predGroup, predActor, predCounter)
		if err != nil {
			return nil, fmt.Errorf("cannot decode operation predecessors: %w", err)
		}
		expanded, err := markExpand.next()
		if err != nil {
			return nil, fmt.Errorf("cannot decode operation mark expansion: %w", err)
		}
		name, _, err := markName.next()
		if err != nil {
			return nil, fmt.Errorf("cannot decode operation mark name: %w", err)
		}

		operations = append(
			operations,
			Operation{
				ID:           id,
				Object:       object,
				Key:          key,
				Insert:       inserted,
				Action:       Action(actionValue),
				Value:        value,
				Predecessors: predecessors,
				MarkExpand:   expanded,
				MarkName:     name,
			},
		)
	}

	return operations, nil
}

func decodeObjectID(
	actor *rleDecoder[uint64],
	counter *rleDecoder[uint64],
) (ObjectID, error) {
	actorValue, actorNull, err := actor.next()
	if err != nil {
		return ObjectID{}, err
	}
	counterValue, counterNull, err := counter.next()
	if err != nil {
		return ObjectID{}, err
	}
	if actorNull != counterNull {
		return ObjectID{}, fmt.Errorf("object actor and counter nullability differ")
	}
	if actorNull {
		return ObjectID{Root: true}, nil
	}
	return ObjectID{OpID: OpID{Counter: counterValue, ActorIndex: actorValue}}, nil
}

func decodeKey(
	actor *rleDecoder[uint64],
	counter *deltaDecoder,
	property *rleDecoder[string],
) (Key, error) {
	actorValue, actorNull, err := actor.next()
	if err != nil {
		return Key{}, err
	}
	counterValue, counterNull, err := counter.next()
	if err != nil {
		return Key{}, err
	}
	propertyValue, propertyNull, err := property.next()
	if err != nil {
		return Key{}, err
	}

	if !propertyNull {
		return Key{Map: true, Property: propertyValue}, nil
	}
	if actorNull != counterNull {
		if actorNull && !counterNull && counterValue == 0 {
			return Key{Head: true}, nil
		}
		return Key{}, fmt.Errorf("element actor and counter nullability differ")
	}
	if actorNull {
		return Key{Head: true}, nil
	}
	if counterValue < 0 {
		return Key{}, fmt.Errorf("element counter cannot be negative")
	}
	return Key{
		Element: OpID{
			Counter:    uint64(counterValue),
			ActorIndex: actorValue,
		},
	}, nil
}

func decodeRequiredOpID(
	actor *rleDecoder[uint64],
	counter *deltaDecoder,
) (OpID, error) {
	actorValue, actorNull, err := actor.next()
	if err != nil {
		return OpID{}, err
	}
	counterValue, counterNull, err := counter.next()
	if err != nil {
		return OpID{}, err
	}
	if actorNull || counterNull || counterValue <= 0 {
		return OpID{}, fmt.Errorf("operation ID is null or non-positive")
	}
	return OpID{Counter: uint64(counterValue), ActorIndex: actorValue}, nil
}

func decodePredecessors(
	group *rleDecoder[uint64],
	actor *rleDecoder[uint64],
	counter *deltaDecoder,
) ([]OpID, error) {
	count, null, err := group.next()
	if err != nil {
		return nil, err
	}
	if null || count == 0 {
		return nil, nil
	}
	if count > maxRLERunLength {
		return nil, fmt.Errorf("predecessor count %d exceeds limit", count)
	}

	predecessors := make([]OpID, int(count))
	for i := range predecessors {
		predecessors[i], err = decodeRequiredOpID(actor, counter)
		if err != nil {
			return nil, fmt.Errorf("cannot decode predecessor %d: %w", i, err)
		}
	}
	return predecessors, nil
}

func decodeScalar(metadata *rleDecoder[uint64], data *reader) (Scalar, error) {
	meta, null, err := metadata.next()
	if err != nil || null {
		return Scalar{}, err
	}

	valueType := uint8(meta & 0x0f)
	length := meta >> 4
	switch valueType {
	case 0:
		return Scalar{Type: valueType}, nil
	case 1:
		return Scalar{Type: valueType, Value: false}, nil
	case 2:
		return Scalar{Type: valueType, Value: true}, nil
	case 3:
		if length == 0 {
			return Scalar{Type: valueType, Value: uint64(0)}, nil
		}
		encoded, err := data.read(int(length))
		if err != nil {
			return Scalar{}, err
		}
		value, consumed, err := readULEB128(encoded)
		if err != nil || consumed != len(encoded) {
			return Scalar{}, fmt.Errorf("invalid unsigned scalar encoding")
		}
		return Scalar{Type: valueType, Value: value}, nil
	case 4, 8, 9:
		if length == 0 {
			return Scalar{Type: valueType, Value: int64(0)}, nil
		}
		encoded, err := data.read(int(length))
		if err != nil {
			return Scalar{}, err
		}
		encodedReader := newReader(encoded)
		value, err := encodedReader.readSLEB128()
		if err != nil || !encodedReader.done() {
			return Scalar{}, fmt.Errorf("invalid signed scalar encoding")
		}
		return Scalar{Type: valueType, Value: value}, nil
	case 5:
		if length != 8 {
			return Scalar{}, fmt.Errorf("float scalar length is %d, expected 8", length)
		}
		value, err := data.read(8)
		if err != nil {
			return Scalar{}, err
		}
		return Scalar{
			Type:  valueType,
			Value: math.Float64frombits(binary.LittleEndian.Uint64(value)),
		}, nil
	case 6:
		value, err := data.read(int(length))
		if err != nil {
			return Scalar{}, err
		}
		return Scalar{Type: valueType, Value: string(value)}, nil
	case 7:
		value, err := data.read(int(length))
		if err != nil {
			return Scalar{}, err
		}
		return Scalar{Type: valueType, Value: append([]byte(nil), value...)}, nil
	default:
		value, err := data.read(int(length))
		if err != nil {
			return Scalar{}, err
		}
		return Scalar{
			Type:    valueType,
			Unknown: append([]byte(nil), value...),
		}, nil
	}
}
