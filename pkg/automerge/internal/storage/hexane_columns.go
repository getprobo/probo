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

package storage

import (
	"fmt"
	"slices"
	"unicode/utf8"

	"go.probo.inc/probo/pkg/automerge/internal/hexane"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

type (
	hexaneOperationSplice struct {
		index       int
		deleteCount int
		inserted    *hexaneOperationColumns
	}

	hexaneChangeColumns struct {
		actors          *hexane.Column[uint64]
		sequences       *hexane.DeltaColumn
		maxOps          *hexane.DeltaColumn
		times           *hexane.DeltaColumn
		messages        *hexane.Column[string]
		dependencySizes *hexane.PrefixColumn
		dependencies    *hexane.DeltaColumn
		extraMetadata   *hexane.Column[uint64]
		extraLengths    *hexane.PrefixColumn
		extraData       *hexane.RawColumn
	}

	hexaneOperationColumns struct {
		objectActors      *hexane.Column[uint64]
		objectCounters    *hexane.Column[uint64]
		keyActors         *hexane.Column[uint64]
		keyCounters       *hexane.DeltaColumn
		keyStrings        *hexane.Column[string]
		idActors          *hexane.Column[uint64]
		idCounters        *hexane.DeltaColumn
		inserts           *hexane.BooleanColumn
		actions           *hexane.Column[uint64]
		valueMetadata     *hexane.Column[uint64]
		valueLengths      *hexane.PrefixColumn
		valueData         *hexane.RawColumn
		successorSizes    *hexane.PrefixColumn
		successorActors   *hexane.Column[uint64]
		successorCounters *hexane.DeltaColumn
		markExpands       *hexane.BooleanColumn
		markExpandTrue    int
		markNames         *hexane.Column[string]
	}
)

func buildHexaneDocumentChangeColumns(
	changes []*opset.Change,
	actorIndexes map[opset.ActorID]uint64,
) ([]encodedColumn, error) {
	count := len(changes)
	indexes := make(map[opset.ChangeHash]uint64, count)
	for i, change := range changes {
		indexes[*change.Hash] = uint64(i)
	}

	var (
		actors         = make([]hexane.Value[uint64], count)
		sequences      = make([]hexane.Value[int64], count)
		maxOps         = make([]hexane.Value[int64], count)
		times          = make([]hexane.Value[int64], count)
		messages       = make([]hexane.Value[string], count)
		dependencySize = make([]hexane.Value[uint64], count)
		dependencies   []hexane.Value[int64]
		extraMetadata  = make([]hexane.Value[uint64], count)
		extraData      = hexane.NewRawColumn()
	)

	for i, change := range changes {
		actorIndex, ok := actorIndexes[change.Actor]
		if !ok {
			return nil, fmt.Errorf("change %d actor is not in the actor table", i)
		}

		actors[i] = hexane.Some(actorIndex)
		sequences[i] = hexane.Some(int64(change.Sequence))
		maxOps[i] = hexane.Some(int64(change.MaxOp))
		times[i] = hexane.Some(change.Time)
		if change.Message != "" {
			messages[i] = hexane.Some(change.Message)
		}

		dependencySize[i] = hexane.Some(uint64(len(change.Dependencies)))
		for _, dependency := range change.Dependencies {
			index, ok := indexes[dependency]
			if !ok {
				return nil, fmt.Errorf("change %d depends on an absent change", i)
			}

			dependencies = append(dependencies, hexane.Some(int64(index)))
		}

		metadata, data, err := encodeScalar(changeExtra(change))
		if err != nil {
			return nil, fmt.Errorf("cannot encode change %d extra: %w", i, err)
		}

		extraMetadata[i] = hexaneValue(metadata)
		extraData.Insert(extraData.Len(), data...)
	}

	actorData, err := hexane.NewColumnFromValues(
		hexane.Uint64Codec(),
		actors...,
	).Bytes()
	if err != nil {
		return nil, fmt.Errorf("cannot encode change actors: %w", err)
	}
	sequenceData, err := hexane.NewDeltaColumnFromValues(sequences...).Bytes()
	if err != nil {
		return nil, fmt.Errorf("cannot encode change sequences: %w", err)
	}
	maxOpData, err := hexane.NewDeltaColumnFromValues(maxOps...).Bytes()
	if err != nil {
		return nil, fmt.Errorf("cannot encode change maximum operations: %w", err)
	}
	timeData, err := hexane.NewDeltaColumnFromValues(times...).Bytes()
	if err != nil {
		return nil, fmt.Errorf("cannot encode change times: %w", err)
	}
	messageData, err := hexane.NewColumnFromValues(
		hexane.StringCodec(),
		messages...,
	).Bytes()
	if err != nil {
		return nil, fmt.Errorf("cannot encode change messages: %w", err)
	}
	dependencySizeData, err := hexane.NewColumnFromValues(
		hexane.Uint64Codec(),
		dependencySize...,
	).Bytes()
	if err != nil {
		return nil, fmt.Errorf("cannot encode change dependency sizes: %w", err)
	}
	dependencyData, err := hexane.NewDeltaColumnFromValues(dependencies...).Bytes()
	if err != nil {
		return nil, fmt.Errorf("cannot encode change dependencies: %w", err)
	}
	extraMetadataData, err := hexane.NewColumnFromValues(
		hexane.Uint64Codec(),
		extraMetadata...,
	).Bytes()
	if err != nil {
		return nil, fmt.Errorf("cannot encode change extra metadata: %w", err)
	}

	return withData(
		[]encodedColumn{
			{specification: 1, data: actorData},
			{specification: 3, data: sequenceData},
			{specification: 19, data: maxOpData},
			{specification: 35, data: timeData},
			{specification: 53, data: messageData},
			{specification: 64, data: dependencySizeData},
			{specification: 67, data: dependencyData},
			{specification: 86, data: extraMetadataData},
			{specification: 87, data: extraData.Bytes()},
		},
	), nil
}

func buildHexaneDocumentOperationColumns(
	operations []opset.Operation,
	actorIndexes map[opset.ActorID]uint64,
) ([]encodedColumn, error) {
	count := len(operations)
	var (
		objectActors      = make([]hexane.Value[uint64], count)
		objectCounters    = make([]hexane.Value[uint64], count)
		keyActors         = make([]hexane.Value[uint64], count)
		keyCounters       = make([]hexane.Value[int64], count)
		keyStrings        = make([]hexane.Value[string], count)
		idActors          = make([]hexane.Value[uint64], count)
		idCounters        = make([]hexane.Value[int64], count)
		inserts           = make([]bool, count)
		actions           = make([]hexane.Value[uint64], count)
		valueMetadata     = make([]hexane.Value[uint64], count)
		valueData         = hexane.NewRawColumn()
		successorSize     = make([]hexane.Value[uint64], count)
		successorActors   []hexane.Value[uint64]
		successorCounters []hexane.Value[int64]
		markExpands       = make([]bool, count)
		hasMarkExpand     bool
		markNames         = make([]hexane.Value[string], count)
	)

	for i, operation := range operations {
		if operation.Key.Property != nil && !utf8.ValidString(*operation.Key.Property) {
			return nil, fmt.Errorf("operation %d key is not valid UTF-8", i)
		}
		if operation.MarkName != nil && !utf8.ValidString(*operation.MarkName) {
			return nil, fmt.Errorf("operation %d mark name is not valid UTF-8", i)
		}
		if operation.Value != nil &&
			operation.Value.Type == opset.ScalarString &&
			!utf8.ValidString(operation.Value.String) {
			return nil, fmt.Errorf("operation %d value is not valid UTF-8", i)
		}
		index, ok := actorIndexes[operation.ID.Actor]
		if !ok {
			return nil, fmt.Errorf("operation %d actor is not in the actor table", i)
		}

		idActors[i] = hexane.Some(index)
		idCounters[i] = hexane.Some(int64(operation.ID.Counter))
		if !operation.Object.IsRoot {
			index, ok := actorIndexes[operation.Object.OpID.Actor]
			if !ok {
				return nil, fmt.Errorf("operation %d object actor is unknown", i)
			}

			objectActors[i] = hexane.Some(index)
			objectCounters[i] = hexane.Some(operation.Object.OpID.Counter)
		}

		switch {
		case operation.Key.Property != nil:
			keyStrings[i] = hexane.Some(*operation.Key.Property)
		case operation.Key.IsHead:
			keyCounters[i] = hexane.Some(int64(0))
		case operation.Key.Element != nil:
			index, ok := actorIndexes[operation.Key.Element.Actor]
			if !ok {
				return nil, fmt.Errorf("operation %d key actor is unknown", i)
			}

			keyActors[i] = hexane.Some(index)
			keyCounters[i] = hexane.Some(int64(operation.Key.Element.Counter))
		default:
			return nil, fmt.Errorf("operation %d has no key", i)
		}

		inserts[i] = operation.Insert
		actions[i] = hexane.Some(uint64(operation.Action))

		metadata, data, err := encodeScalar(operation.Value)
		if err != nil {
			return nil, fmt.Errorf("cannot encode operation %d value: %w", i, err)
		}

		valueMetadata[i] = hexaneValue(metadata)
		valueData.Insert(valueData.Len(), data...)
		successorSize[i] = hexane.Some(uint64(len(operation.Successors)))
		for _, successor := range operation.Successors {
			index, ok := actorIndexes[successor.Actor]
			if !ok {
				return nil, fmt.Errorf("operation %d successor actor is unknown", i)
			}

			successorActors = append(successorActors, hexane.Some(index))
			successorCounters = append(
				successorCounters,
				hexane.Some(int64(successor.Counter)),
			)
		}

		if operation.MarkExpand != nil && *operation.MarkExpand {
			markExpands[i] = true
			hasMarkExpand = true
		}
		if operation.MarkName != nil {
			markNames[i] = hexane.Some(*operation.MarkName)
		}
	}

	objectActorData, err := hexaneBytes(objectActors, hexane.Uint64Codec())
	if err != nil {
		return nil, fmt.Errorf("cannot encode operation object actors: %w", err)
	}
	objectCounterData, err := hexaneBytes(objectCounters, hexane.Uint64Codec())
	if err != nil {
		return nil, fmt.Errorf("cannot encode operation object counters: %w", err)
	}
	keyActorData, err := hexaneBytes(keyActors, hexane.Uint64Codec())
	if err != nil {
		return nil, fmt.Errorf("cannot encode operation key actors: %w", err)
	}
	keyCounterData, err := hexane.NewDeltaColumnFromValues(keyCounters...).Bytes()
	if err != nil {
		return nil, fmt.Errorf("cannot encode operation key counters: %w", err)
	}
	keyStringData, err := hexaneBytes(keyStrings, hexane.StringCodec())
	if err != nil {
		return nil, fmt.Errorf("cannot encode operation key strings: %w", err)
	}
	idActorData, err := hexaneBytes(idActors, hexane.Uint64Codec())
	if err != nil {
		return nil, fmt.Errorf("cannot encode operation identifier actors: %w", err)
	}
	idCounterData, err := hexane.NewDeltaColumnFromValues(idCounters...).Bytes()
	if err != nil {
		return nil, fmt.Errorf("cannot encode operation identifier counters: %w", err)
	}
	actionData, err := hexaneBytes(actions, hexane.Uint64Codec())
	if err != nil {
		return nil, fmt.Errorf("cannot encode operation actions: %w", err)
	}
	valueMetadataData, err := hexaneBytes(valueMetadata, hexane.Uint64Codec())
	if err != nil {
		return nil, fmt.Errorf("cannot encode operation value metadata: %w", err)
	}
	successorSizeData, err := hexaneBytes(successorSize, hexane.Uint64Codec())
	if err != nil {
		return nil, fmt.Errorf("cannot encode operation successor sizes: %w", err)
	}
	successorActorData, err := hexaneBytes(successorActors, hexane.Uint64Codec())
	if err != nil {
		return nil, fmt.Errorf("cannot encode operation successor actors: %w", err)
	}
	successorCounterData, err := hexane.NewDeltaColumnFromValues(
		successorCounters...,
	).Bytes()
	if err != nil {
		return nil, fmt.Errorf("cannot encode operation successor counters: %w", err)
	}
	markNameData, err := hexaneBytes(markNames, hexane.StringCodec())
	if err != nil {
		return nil, fmt.Errorf("cannot encode operation mark names: %w", err)
	}

	var markExpandData []byte
	if hasMarkExpand {
		markExpandData = hexane.NewBooleanColumnFromValues(markExpands...).Bytes()
	}

	return withData(
		[]encodedColumn{
			{specification: 1, data: objectActorData},
			{specification: 2, data: objectCounterData},
			{specification: 17, data: keyActorData},
			{specification: 19, data: keyCounterData},
			{specification: 21, data: keyStringData},
			{specification: 33, data: idActorData},
			{specification: 35, data: idCounterData},
			{
				specification: 52,
				data:          hexane.NewBooleanColumnFromValues(inserts...).Bytes(),
			},
			{specification: 66, data: actionData},
			{specification: 86, data: valueMetadataData},
			{specification: 87, data: valueData.Bytes()},
			{specification: 128, data: successorSizeData},
			{specification: 129, data: successorActorData},
			{specification: 131, data: successorCounterData},
			{specification: 148, data: markExpandData},
			{specification: 165, data: markNameData},
		},
	), nil
}

func hexaneBytes[T any](
	values []hexane.Value[T],
	codec hexane.Codec[T],
) ([]byte, error) {
	return hexane.NewColumnFromValues(codec, values...).Bytes()
}

func hexaneValue[T any](value optional[T]) hexane.Value[T] {
	if !value.valid {
		return hexane.Null[T]()
	}

	return hexane.Some(value.value)
}

func newHexaneChangeColumns(
	changes []*opset.Change,
	actorIndexes map[opset.ActorID]uint64,
) (*hexaneChangeColumns, error) {
	dependencyIndexes := make(
		map[opset.ChangeHash]uint64,
		len(changes),
	)
	for i, change := range changes {
		if change.Hash == nil {
			return nil, fmt.Errorf("change %d has no hash", i)
		}
		dependencyIndexes[*change.Hash] = uint64(i)
	}

	return newHexaneChangeColumnsForSplice(
		changes,
		actorIndexes,
		dependencyIndexes,
	)
}

func newHexaneChangeColumnsForSplice(
	changes []*opset.Change,
	actorIndexes map[opset.ActorID]uint64,
	dependencyIndexes map[opset.ChangeHash]uint64,
) (*hexaneChangeColumns, error) {
	var (
		actors          = make([]hexane.Value[uint64], len(changes))
		sequences       = make([]hexane.Value[int64], len(changes))
		maxOps          = make([]hexane.Value[int64], len(changes))
		times           = make([]hexane.Value[int64], len(changes))
		messages        = make([]hexane.Value[string], len(changes))
		dependencySizes = make([]uint64, len(changes))
		dependencies    []hexane.Value[int64]
		extraMetadata   = make([]hexane.Value[uint64], len(changes))
		extraLengths    = make([]uint64, len(changes))
		extraData       []byte
	)

	for i, change := range changes {
		actorIndex, ok := actorIndexes[change.Actor]
		if !ok {
			return nil, fmt.Errorf("change %d actor is not in the actor table", i)
		}
		actors[i] = hexane.Some(actorIndex)
		sequences[i] = hexane.Some(int64(change.Sequence))
		maxOps[i] = hexane.Some(int64(change.MaxOp))
		times[i] = hexane.Some(change.Time)
		if change.Message != "" {
			messages[i] = hexane.Some(change.Message)
		}

		dependencySizes[i] = uint64(len(change.Dependencies))
		for _, dependency := range change.Dependencies {
			index, ok := dependencyIndexes[dependency]
			if !ok {
				return nil, fmt.Errorf("change %d depends on an absent change", i)
			}
			dependencies = append(dependencies, hexane.Some(int64(index)))
		}

		metadata, data, err := encodeScalar(changeExtra(change))
		if err != nil {
			return nil, fmt.Errorf("cannot encode change %d extra: %w", i, err)
		}
		extraMetadata[i] = hexaneValue(metadata)
		extraLengths[i] = uint64(len(data))
		extraData = append(extraData, data...)
	}

	return &hexaneChangeColumns{
		actors:          hexane.NewColumnFromValues(hexane.Uint64Codec(), actors...),
		sequences:       hexane.NewDeltaColumnFromValues(sequences...),
		maxOps:          hexane.NewDeltaColumnFromValues(maxOps...),
		times:           hexane.NewDeltaColumnFromValues(times...),
		messages:        hexane.NewColumnFromValues(hexane.StringCodec(), messages...),
		dependencySizes: hexane.NewPrefixColumnFromValues(dependencySizes...),
		dependencies:    hexane.NewDeltaColumnFromValues(dependencies...),
		extraMetadata: hexane.NewColumnFromValues(
			hexane.Uint64Codec(),
			extraMetadata...,
		),
		extraLengths: hexane.NewPrefixColumnFromValues(extraLengths...),
		extraData:    hexane.NewRawColumnFromBytes(extraData),
	}, nil
}

func decodeHexaneChangeColumns(
	columns []encodedColumn,
	count int,
) (*hexaneChangeColumns, error) {
	actors, err := decodeULEBColumn(encodedColumnData(columns, 1))
	if err != nil {
		return nil, fmt.Errorf("cannot decode change actors: %w", err)
	}
	sequences, err := decodeSignedDeltaColumn(encodedColumnData(columns, 3))
	if err != nil {
		return nil, fmt.Errorf("cannot decode change sequences: %w", err)
	}
	maxOps, err := decodeSignedDeltaColumn(encodedColumnData(columns, 19))
	if err != nil {
		return nil, fmt.Errorf("cannot decode change maximum operations: %w", err)
	}
	times, err := decodeSignedDeltaColumn(encodedColumnData(columns, 35))
	if err != nil {
		return nil, fmt.Errorf("cannot decode change times: %w", err)
	}
	messages, err := decodeStringColumn(encodedColumnData(columns, 53))
	if err != nil {
		return nil, fmt.Errorf("cannot decode change messages: %w", err)
	}
	dependencySizes, err := decodeULEBColumn(encodedColumnData(columns, 64))
	if err != nil {
		return nil, fmt.Errorf("cannot decode change dependency sizes: %w", err)
	}
	dependencies, err := decodeSignedDeltaColumn(encodedColumnData(columns, 67))
	if err != nil {
		return nil, fmt.Errorf("cannot decode change dependencies: %w", err)
	}
	extraMetadata, err := decodeULEBColumn(encodedColumnData(columns, 86))
	if err != nil {
		return nil, fmt.Errorf("cannot decode change extra metadata: %w", err)
	}

	return &hexaneChangeColumns{
		actors:          hexane.NewColumnFromValues(hexane.Uint64Codec(), hexaneValues(actors)...),
		sequences:       hexane.NewDeltaColumnFromValues(hexaneValues(sequences)...),
		maxOps:          hexane.NewDeltaColumnFromValues(hexaneValues(maxOps)...),
		times:           hexane.NewDeltaColumnFromValues(hexaneValues(times)...),
		messages:        hexane.NewColumnFromValues(hexane.StringCodec(), paddedHexaneValues(messages, count)...),
		dependencySizes: hexane.NewPrefixColumnFromValues(presentUint64s(dependencySizes)...),
		dependencies:    hexane.NewDeltaColumnFromValues(hexaneValues(dependencies)...),
		extraMetadata: hexane.NewColumnFromValues(
			hexane.Uint64Codec(),
			hexaneValues(extraMetadata)...,
		),
		extraLengths: hexane.NewPrefixColumnFromValues(metadataLengths(extraMetadata)...),
		extraData:    hexane.NewRawColumnFromBytes(encodedColumnData(columns, 87)),
	}, nil
}

func newHexaneOperationColumns(
	operations []opset.Operation,
	actorIndexes map[opset.ActorID]uint64,
) (*hexaneOperationColumns, error) {
	count := len(operations)
	var (
		objectActors      = make([]hexane.Value[uint64], count)
		objectCounters    = make([]hexane.Value[uint64], count)
		keyActors         = make([]hexane.Value[uint64], count)
		keyCounters       = make([]hexane.Value[int64], count)
		keyStrings        = make([]hexane.Value[string], count)
		idActors          = make([]hexane.Value[uint64], count)
		idCounters        = make([]hexane.Value[int64], count)
		inserts           = make([]bool, count)
		actions           = make([]hexane.Value[uint64], count)
		valueMetadata     = make([]hexane.Value[uint64], count)
		valueLengths      = make([]uint64, count)
		valueData         []byte
		successorSizes    = make([]uint64, count)
		successorActors   []hexane.Value[uint64]
		successorCounters []hexane.Value[int64]
		markExpands       = make([]bool, count)
		markExpandTrue    int
		markNames         = make([]hexane.Value[string], count)
	)

	for i, operation := range operations {
		if operation.Key.Property != nil && !utf8.ValidString(*operation.Key.Property) {
			return nil, fmt.Errorf("operation %d key is not valid UTF-8", i)
		}
		if operation.MarkName != nil && !utf8.ValidString(*operation.MarkName) {
			return nil, fmt.Errorf("operation %d mark name is not valid UTF-8", i)
		}
		if operation.Value != nil &&
			operation.Value.Type == opset.ScalarString &&
			!utf8.ValidString(operation.Value.String) {
			return nil, fmt.Errorf("operation %d value is not valid UTF-8", i)
		}
		index, ok := actorIndexes[operation.ID.Actor]
		if !ok {
			return nil, fmt.Errorf(
				"operation %d actor is not in the actor table",
				i,
			)
		}
		idActors[i] = hexane.Some(index)
		idCounters[i] = hexane.Some(int64(operation.ID.Counter))

		if !operation.Object.IsRoot {
			index, ok := actorIndexes[operation.Object.OpID.Actor]
			if !ok {
				return nil, fmt.Errorf(
					"operation %d object actor is unknown",
					i,
				)
			}
			objectActors[i] = hexane.Some(index)
			objectCounters[i] = hexane.Some(
				operation.Object.OpID.Counter,
			)
		}

		switch {
		case operation.Key.Property != nil:
			keyStrings[i] = hexane.Some(*operation.Key.Property)
		case operation.Key.IsHead:
			keyCounters[i] = hexane.Some(int64(0))
		case operation.Key.Element != nil:
			index, ok := actorIndexes[operation.Key.Element.Actor]
			if !ok {
				return nil, fmt.Errorf(
					"operation %d key actor is unknown",
					i,
				)
			}
			keyActors[i] = hexane.Some(index)
			keyCounters[i] = hexane.Some(
				int64(operation.Key.Element.Counter),
			)
		default:
			return nil, fmt.Errorf("operation %d has no key", i)
		}

		inserts[i] = operation.Insert
		actions[i] = hexane.Some(uint64(operation.Action))
		metadata, data, err := encodeScalar(operation.Value)
		if err != nil {
			return nil, fmt.Errorf(
				"cannot encode operation %d value: %w",
				i,
				err,
			)
		}
		valueMetadata[i] = hexaneValue(metadata)
		valueLengths[i] = uint64(len(data))
		valueData = append(valueData, data...)

		successorSizes[i] = uint64(len(operation.Successors))
		for _, successor := range operation.Successors {
			index, ok := actorIndexes[successor.Actor]
			if !ok {
				return nil, fmt.Errorf(
					"operation %d successor actor is unknown",
					i,
				)
			}
			successorActors = append(
				successorActors,
				hexane.Some(index),
			)
			successorCounters = append(
				successorCounters,
				hexane.Some(int64(successor.Counter)),
			)
		}
		if operation.MarkExpand != nil && *operation.MarkExpand {
			markExpands[i] = true
			markExpandTrue++
		}
		if operation.MarkName != nil {
			markNames[i] = hexane.Some(*operation.MarkName)
		}
	}

	return &hexaneOperationColumns{
		objectActors: hexane.NewColumnFromValues(
			hexane.Uint64Codec(),
			objectActors...,
		),
		objectCounters: hexane.NewColumnFromValues(
			hexane.Uint64Codec(),
			objectCounters...,
		),
		keyActors: hexane.NewColumnFromValues(
			hexane.Uint64Codec(),
			keyActors...,
		),
		keyCounters: hexane.NewDeltaColumnFromValues(keyCounters...),
		keyStrings: hexane.NewColumnFromValues(
			hexane.StringCodec(),
			keyStrings...,
		),
		idActors: hexane.NewColumnFromValues(
			hexane.Uint64Codec(),
			idActors...,
		),
		idCounters: hexane.NewDeltaColumnFromValues(idCounters...),
		inserts:    hexane.NewBooleanColumnFromValues(inserts...),
		actions: hexane.NewColumnFromValues(
			hexane.Uint64Codec(),
			actions...,
		),
		valueMetadata: hexane.NewColumnFromValues(
			hexane.Uint64Codec(),
			valueMetadata...,
		),
		valueLengths: hexane.NewPrefixColumnFromValues(valueLengths...),
		valueData:    hexane.NewRawColumnFromBytes(valueData),
		successorSizes: hexane.NewPrefixColumnFromValues(
			successorSizes...,
		),
		successorActors: hexane.NewColumnFromValues(
			hexane.Uint64Codec(),
			successorActors...,
		),
		successorCounters: hexane.NewDeltaColumnFromValues(
			successorCounters...,
		),
		markExpands:    hexane.NewBooleanColumnFromValues(markExpands...),
		markExpandTrue: markExpandTrue,
		markNames: hexane.NewColumnFromValues(
			hexane.StringCodec(),
			markNames...,
		),
	}, nil
}

func decodeHexaneOperationColumns(
	columns []encodedColumn,
	count int,
) (*hexaneOperationColumns, error) {
	decodeULEB := func(specification uint32) ([]optional[uint64], error) {
		return decodeULEBColumn(encodedColumnData(columns, specification))
	}
	decodeDelta := func(specification uint32) ([]optional[int64], error) {
		return decodeSignedDeltaColumn(encodedColumnData(columns, specification))
	}

	objectActors, err := decodeULEB(1)
	if err != nil {
		return nil, fmt.Errorf("cannot decode operation object actors: %w", err)
	}
	objectCounters, err := decodeULEB(2)
	if err != nil {
		return nil, fmt.Errorf("cannot decode operation object counters: %w", err)
	}
	keyActors, err := decodeULEB(17)
	if err != nil {
		return nil, fmt.Errorf("cannot decode operation key actors: %w", err)
	}
	keyCounters, err := decodeDelta(19)
	if err != nil {
		return nil, fmt.Errorf("cannot decode operation key counters: %w", err)
	}
	keyStrings, err := decodeStringColumn(encodedColumnData(columns, 21))
	if err != nil {
		return nil, fmt.Errorf("cannot decode operation key strings: %w", err)
	}
	idActors, err := decodeULEB(33)
	if err != nil {
		return nil, fmt.Errorf("cannot decode operation identifier actors: %w", err)
	}
	idCounters, err := decodeDelta(35)
	if err != nil {
		return nil, fmt.Errorf("cannot decode operation identifier counters: %w", err)
	}
	actions, err := decodeULEB(66)
	if err != nil {
		return nil, fmt.Errorf("cannot decode operation actions: %w", err)
	}
	valueMetadata, err := decodeULEB(86)
	if err != nil {
		return nil, fmt.Errorf("cannot decode operation value metadata: %w", err)
	}
	successorSizes, err := decodeULEB(128)
	if err != nil {
		return nil, fmt.Errorf("cannot decode operation successor sizes: %w", err)
	}
	successorActors, err := decodeULEB(129)
	if err != nil {
		return nil, fmt.Errorf("cannot decode operation successor actors: %w", err)
	}
	successorCounters, err := decodeDelta(131)
	if err != nil {
		return nil, fmt.Errorf("cannot decode operation successor counters: %w", err)
	}
	markNames, err := decodeStringColumn(encodedColumnData(columns, 165))
	if err != nil {
		return nil, fmt.Errorf("cannot decode operation mark names: %w", err)
	}

	inserts, err := decodeBooleanOrZero(encodedColumnData(columns, 52), count)
	if err != nil {
		return nil, fmt.Errorf("cannot decode operation inserts: %w", err)
	}
	markExpands, err := decodeBooleanOrZero(encodedColumnData(columns, 148), count)
	if err != nil {
		return nil, fmt.Errorf("cannot decode operation mark expansion: %w", err)
	}

	return &hexaneOperationColumns{
		objectActors: hexane.NewColumnFromValues(
			hexane.Uint64Codec(),
			paddedHexaneValues(objectActors, count)...,
		),
		objectCounters: hexane.NewColumnFromValues(
			hexane.Uint64Codec(),
			paddedHexaneValues(objectCounters, count)...,
		),
		keyActors: hexane.NewColumnFromValues(
			hexane.Uint64Codec(),
			paddedHexaneValues(keyActors, count)...,
		),
		keyCounters: hexane.NewDeltaColumnFromValues(paddedHexaneValues(keyCounters, count)...),
		keyStrings: hexane.NewColumnFromValues(
			hexane.StringCodec(),
			paddedHexaneValues(keyStrings, count)...,
		),
		idActors:   hexane.NewColumnFromValues(hexane.Uint64Codec(), hexaneValues(idActors)...),
		idCounters: hexane.NewDeltaColumnFromValues(hexaneValues(idCounters)...),
		inserts:    hexane.NewBooleanColumnFromValues(inserts...),
		actions: hexane.NewColumnFromValues(
			hexane.Uint64Codec(),
			hexaneValues(actions)...,
		),
		valueMetadata: hexane.NewColumnFromValues(
			hexane.Uint64Codec(),
			hexaneValues(valueMetadata)...,
		),
		valueLengths:   hexane.NewPrefixColumnFromValues(metadataLengths(valueMetadata)...),
		valueData:      hexane.NewRawColumnFromBytes(encodedColumnData(columns, 87)),
		successorSizes: hexane.NewPrefixColumnFromValues(presentUint64s(successorSizes)...),
		successorActors: hexane.NewColumnFromValues(
			hexane.Uint64Codec(),
			hexaneValues(successorActors)...,
		),
		successorCounters: hexane.NewDeltaColumnFromValues(hexaneValues(successorCounters)...),
		markExpands:       hexane.NewBooleanColumnFromValues(markExpands...),
		markExpandTrue:    countTrue(markExpands),
		markNames: hexane.NewColumnFromValues(
			hexane.StringCodec(),
			paddedHexaneValues(markNames, count)...,
		),
	}, nil
}

func (c *hexaneChangeColumns) Clone() *hexaneChangeColumns {
	return &hexaneChangeColumns{
		actors:          c.actors.Clone(),
		sequences:       c.sequences.Clone(),
		maxOps:          c.maxOps.Clone(),
		times:           c.times.Clone(),
		messages:        c.messages.Clone(),
		dependencySizes: c.dependencySizes.Clone(),
		dependencies:    c.dependencies.Clone(),
		extraMetadata:   c.extraMetadata.Clone(),
		extraLengths:    c.extraLengths.Clone(),
		extraData:       c.extraData.Clone(),
	}
}

func (c *hexaneOperationColumns) Clone() *hexaneOperationColumns {
	return &hexaneOperationColumns{
		objectActors:      c.objectActors.Clone(),
		objectCounters:    c.objectCounters.Clone(),
		keyActors:         c.keyActors.Clone(),
		keyCounters:       c.keyCounters.Clone(),
		keyStrings:        c.keyStrings.Clone(),
		idActors:          c.idActors.Clone(),
		idCounters:        c.idCounters.Clone(),
		inserts:           c.inserts.Clone(),
		actions:           c.actions.Clone(),
		valueMetadata:     c.valueMetadata.Clone(),
		valueLengths:      c.valueLengths.Clone(),
		valueData:         c.valueData.Clone(),
		successorSizes:    c.successorSizes.Clone(),
		successorActors:   c.successorActors.Clone(),
		successorCounters: c.successorCounters.Clone(),
		markExpands:       c.markExpands.Clone(),
		markExpandTrue:    c.markExpandTrue,
		markNames:         c.markNames.Clone(),
	}
}

func (c *hexaneChangeColumns) Splice(
	index int,
	deleteCount int,
	inserted *hexaneChangeColumns,
) error {
	if index < 0 || deleteCount < 0 || index > c.actors.Len() ||
		deleteCount > c.actors.Len()-index {
		return fmt.Errorf("change splice range is out of bounds")
	}

	dependencyStart := int(c.dependencySizes.Prefix(index))
	dependencyEnd := int(c.dependencySizes.Prefix(index + deleteCount))
	extraStart := int(c.extraLengths.Prefix(index))
	extraEnd := int(c.extraLengths.Prefix(index + deleteCount))

	if err := c.actors.BatchSplice([]hexane.ColumnSplice[uint64]{{
		Index: index, DeleteCount: deleteCount, Inserted: inserted.actors,
	}}); err != nil {
		return fmt.Errorf("cannot splice change actors: %w", err)
	}
	if err := c.sequences.BatchSplice([]hexane.DeltaSplice{{
		Index: index, DeleteCount: deleteCount, Inserted: inserted.sequences,
	}}); err != nil {
		return fmt.Errorf("cannot splice change sequences: %w", err)
	}
	if err := c.maxOps.BatchSplice([]hexane.DeltaSplice{{
		Index: index, DeleteCount: deleteCount, Inserted: inserted.maxOps,
	}}); err != nil {
		return fmt.Errorf("cannot splice change maximum operations: %w", err)
	}
	if err := c.times.BatchSplice([]hexane.DeltaSplice{{
		Index: index, DeleteCount: deleteCount, Inserted: inserted.times,
	}}); err != nil {
		return fmt.Errorf("cannot splice change times: %w", err)
	}
	if err := c.messages.BatchSplice([]hexane.ColumnSplice[string]{{
		Index: index, DeleteCount: deleteCount, Inserted: inserted.messages,
	}}); err != nil {
		return fmt.Errorf("cannot splice change messages: %w", err)
	}
	if err := c.dependencySizes.BatchSplice([]hexane.PrefixSplice{{
		Index: index, DeleteCount: deleteCount, Inserted: inserted.dependencySizes,
	}}); err != nil {
		return fmt.Errorf("cannot splice change dependency sizes: %w", err)
	}
	if err := c.dependencies.BatchSplice([]hexane.DeltaSplice{{
		Index:       dependencyStart,
		DeleteCount: dependencyEnd - dependencyStart,
		Inserted:    inserted.dependencies,
	}}); err != nil {
		return fmt.Errorf("cannot splice change dependencies: %w", err)
	}
	if err := c.extraMetadata.BatchSplice([]hexane.ColumnSplice[uint64]{{
		Index: index, DeleteCount: deleteCount, Inserted: inserted.extraMetadata,
	}}); err != nil {
		return fmt.Errorf("cannot splice change extra metadata: %w", err)
	}
	if err := c.extraLengths.BatchSplice([]hexane.PrefixSplice{{
		Index: index, DeleteCount: deleteCount, Inserted: inserted.extraLengths,
	}}); err != nil {
		return fmt.Errorf("cannot splice change extra lengths: %w", err)
	}
	if err := c.extraData.BatchSplice([]hexane.RawSplice{{
		Index:       extraStart,
		DeleteCount: extraEnd - extraStart,
		Inserted:    inserted.extraData,
	}}); err != nil {
		return fmt.Errorf("cannot splice change extra data: %w", err)
	}

	return nil
}

func (c *hexaneOperationColumns) Splice(
	index int,
	deleteCount int,
	inserted *hexaneOperationColumns,
) error {
	return c.BatchSplice([]hexaneOperationSplice{{
		index: index, deleteCount: deleteCount, inserted: inserted,
	}})
}

func (c *hexaneOperationColumns) BatchSplice(
	splices []hexaneOperationSplice,
) error {
	sorted := append([]hexaneOperationSplice(nil), splices...)
	slices.SortStableFunc(
		sorted,
		func(left, right hexaneOperationSplice) int {
			return left.index - right.index
		},
	)
	previousEnd := 0
	for i, splice := range sorted {
		if splice.inserted == nil {
			return fmt.Errorf("operation splice %d has nil inserted columns", i)
		}
		if splice.index < 0 ||
			splice.deleteCount < 0 ||
			splice.index > c.idActors.Len() ||
			splice.deleteCount > c.idActors.Len()-splice.index {
			return fmt.Errorf("operation splice %d range is out of bounds", i)
		}
		if i > 0 && splice.index < previousEnd {
			return fmt.Errorf("operation splice %d overlaps its predecessor", i)
		}
		previousEnd = splice.index + splice.deleteCount
	}

	rowColumns := func(column func(*hexaneOperationColumns) *hexane.Column[uint64]) []hexane.ColumnSplice[uint64] {
		result := make([]hexane.ColumnSplice[uint64], len(sorted))
		for i, splice := range sorted {
			result[i] = hexane.ColumnSplice[uint64]{
				Index: splice.index, DeleteCount: splice.deleteCount,
				Inserted: column(splice.inserted),
			}
		}
		return result
	}
	deltaColumns := func(column func(*hexaneOperationColumns) *hexane.DeltaColumn) []hexane.DeltaSplice {
		result := make([]hexane.DeltaSplice, len(sorted))
		for i, splice := range sorted {
			result[i] = hexane.DeltaSplice{
				Index: splice.index, DeleteCount: splice.deleteCount,
				Inserted: column(splice.inserted),
			}
		}
		return result
	}
	stringColumns := func(column func(*hexaneOperationColumns) *hexane.Column[string]) []hexane.ColumnSplice[string] {
		result := make([]hexane.ColumnSplice[string], len(sorted))
		for i, splice := range sorted {
			result[i] = hexane.ColumnSplice[string]{
				Index: splice.index, DeleteCount: splice.deleteCount,
				Inserted: column(splice.inserted),
			}
		}
		return result
	}
	prefixColumns := func(column func(*hexaneOperationColumns) *hexane.PrefixColumn) []hexane.PrefixSplice {
		result := make([]hexane.PrefixSplice, len(sorted))
		for i, splice := range sorted {
			result[i] = hexane.PrefixSplice{
				Index: splice.index, DeleteCount: splice.deleteCount,
				Inserted: column(splice.inserted),
			}
		}
		return result
	}
	booleanColumns := func(column func(*hexaneOperationColumns) *hexane.BooleanColumn) []hexane.BooleanSplice {
		result := make([]hexane.BooleanSplice, len(sorted))
		for i, splice := range sorted {
			result[i] = hexane.BooleanSplice{
				Index: splice.index, DeleteCount: splice.deleteCount,
				Inserted: column(splice.inserted),
			}
		}
		return result
	}

	successors := make([]hexane.ColumnSplice[uint64], len(sorted))
	successorCounters := make([]hexane.DeltaSplice, len(sorted))
	values := make([]hexane.RawSplice, len(sorted))
	for i, splice := range sorted {
		successorStart := int(c.successorSizes.Prefix(splice.index))
		successorEnd := int(c.successorSizes.Prefix(splice.index + splice.deleteCount))
		successors[i] = hexane.ColumnSplice[uint64]{
			Index: successorStart, DeleteCount: successorEnd - successorStart,
			Inserted: splice.inserted.successorActors,
		}
		successorCounters[i] = hexane.DeltaSplice{
			Index: successorStart, DeleteCount: successorEnd - successorStart,
			Inserted: splice.inserted.successorCounters,
		}
		valueStart := int(c.valueLengths.Prefix(splice.index))
		valueEnd := int(c.valueLengths.Prefix(splice.index + splice.deleteCount))
		values[i] = hexane.RawSplice{
			Index: valueStart, DeleteCount: valueEnd - valueStart,
			Inserted: splice.inserted.valueData,
		}
		for row := splice.index; row < splice.index+splice.deleteCount; row++ {
			if c.markExpands.Get(row) {
				c.markExpandTrue--
			}
		}
		c.markExpandTrue += splice.inserted.markExpandTrue
	}

	if err := c.objectActors.BatchSplice(rowColumns(func(item *hexaneOperationColumns) *hexane.Column[uint64] { return item.objectActors })); err != nil {
		return err
	}
	if err := c.objectCounters.BatchSplice(rowColumns(func(item *hexaneOperationColumns) *hexane.Column[uint64] { return item.objectCounters })); err != nil {
		return err
	}
	if err := c.keyActors.BatchSplice(rowColumns(func(item *hexaneOperationColumns) *hexane.Column[uint64] { return item.keyActors })); err != nil {
		return err
	}
	if err := c.keyCounters.BatchSplice(deltaColumns(func(item *hexaneOperationColumns) *hexane.DeltaColumn { return item.keyCounters })); err != nil {
		return err
	}
	if err := c.keyStrings.BatchSplice(stringColumns(func(item *hexaneOperationColumns) *hexane.Column[string] { return item.keyStrings })); err != nil {
		return err
	}
	if err := c.idActors.BatchSplice(rowColumns(func(item *hexaneOperationColumns) *hexane.Column[uint64] { return item.idActors })); err != nil {
		return err
	}
	if err := c.idCounters.BatchSplice(deltaColumns(func(item *hexaneOperationColumns) *hexane.DeltaColumn { return item.idCounters })); err != nil {
		return err
	}
	if err := c.inserts.BatchSplice(booleanColumns(func(item *hexaneOperationColumns) *hexane.BooleanColumn { return item.inserts })); err != nil {
		return err
	}
	if err := c.actions.BatchSplice(rowColumns(func(item *hexaneOperationColumns) *hexane.Column[uint64] { return item.actions })); err != nil {
		return err
	}
	if err := c.valueMetadata.BatchSplice(rowColumns(func(item *hexaneOperationColumns) *hexane.Column[uint64] { return item.valueMetadata })); err != nil {
		return err
	}
	if err := c.valueLengths.BatchSplice(prefixColumns(func(item *hexaneOperationColumns) *hexane.PrefixColumn { return item.valueLengths })); err != nil {
		return err
	}
	if err := c.valueData.BatchSplice(values); err != nil {
		return err
	}
	if err := c.successorSizes.BatchSplice(prefixColumns(func(item *hexaneOperationColumns) *hexane.PrefixColumn { return item.successorSizes })); err != nil {
		return err
	}
	if err := c.successorActors.BatchSplice(successors); err != nil {
		return err
	}
	if err := c.successorCounters.BatchSplice(successorCounters); err != nil {
		return err
	}
	if err := c.markExpands.BatchSplice(booleanColumns(func(item *hexaneOperationColumns) *hexane.BooleanColumn { return item.markExpands })); err != nil {
		return err
	}
	if err := c.markNames.BatchSplice(stringColumns(func(item *hexaneOperationColumns) *hexane.Column[string] { return item.markNames })); err != nil {
		return err
	}

	return nil
}

func (c *hexaneChangeColumns) RemapActors(
	oldActors []opset.ActorID,
	newIndexes map[opset.ActorID]uint64,
) error {
	remapped, err := remapActorColumn(c.actors, oldActors, newIndexes)
	if err != nil {
		return fmt.Errorf("cannot remap change actors: %w", err)
	}
	c.actors = remapped

	return nil
}

func (c *hexaneOperationColumns) RemapActors(
	oldActors []opset.ActorID,
	newIndexes map[opset.ActorID]uint64,
) error {
	var err error
	c.objectActors, err = remapActorColumn(c.objectActors, oldActors, newIndexes)
	if err != nil {
		return fmt.Errorf("cannot remap operation object actors: %w", err)
	}
	c.keyActors, err = remapActorColumn(c.keyActors, oldActors, newIndexes)
	if err != nil {
		return fmt.Errorf("cannot remap operation key actors: %w", err)
	}
	c.idActors, err = remapActorColumn(c.idActors, oldActors, newIndexes)
	if err != nil {
		return fmt.Errorf("cannot remap operation identifier actors: %w", err)
	}
	c.successorActors, err = remapActorColumn(
		c.successorActors,
		oldActors,
		newIndexes,
	)
	if err != nil {
		return fmt.Errorf("cannot remap operation successor actors: %w", err)
	}

	return nil
}

func (c *hexaneChangeColumns) Encoded() ([]encodedColumn, error) {
	return encodeHexaneColumns(
		[]hexaneColumnEncoder{
			columnEncoder(1, c.actors),
			deltaEncoder(3, c.sequences),
			deltaEncoder(19, c.maxOps),
			deltaEncoder(35, c.times),
			columnEncoder(53, c.messages),
			prefixEncoder(64, c.dependencySizes),
			deltaEncoder(67, c.dependencies),
			columnEncoder(86, c.extraMetadata),
			rawEncoder(87, c.extraData),
		},
	)
}

func (c *hexaneOperationColumns) Encoded() ([]encodedColumn, error) {
	markExpand := []byte(nil)
	if c.markExpandTrue > 0 {
		markExpand = c.markExpands.Bytes()
	}

	return encodeHexaneColumns(
		[]hexaneColumnEncoder{
			columnEncoder(1, c.objectActors),
			columnEncoder(2, c.objectCounters),
			columnEncoder(17, c.keyActors),
			deltaEncoder(19, c.keyCounters),
			columnEncoder(21, c.keyStrings),
			columnEncoder(33, c.idActors),
			deltaEncoder(35, c.idCounters),
			bytesEncoder(52, c.inserts.Bytes()),
			columnEncoder(66, c.actions),
			columnEncoder(86, c.valueMetadata),
			rawEncoder(87, c.valueData),
			prefixEncoder(128, c.successorSizes),
			columnEncoder(129, c.successorActors),
			deltaEncoder(131, c.successorCounters),
			bytesEncoder(148, markExpand),
			columnEncoder(165, c.markNames),
		},
	)
}

type hexaneColumnEncoder struct {
	specification uint32
	bytes         func() ([]byte, error)
}

func columnEncoder[T any](
	specification uint32,
	column *hexane.Column[T],
) hexaneColumnEncoder {
	return hexaneColumnEncoder{specification: specification, bytes: column.Bytes}
}

func deltaEncoder(
	specification uint32,
	column *hexane.DeltaColumn,
) hexaneColumnEncoder {
	return hexaneColumnEncoder{specification: specification, bytes: column.Bytes}
}

func prefixEncoder(
	specification uint32,
	column *hexane.PrefixColumn,
) hexaneColumnEncoder {
	return hexaneColumnEncoder{specification: specification, bytes: column.Bytes}
}

func rawEncoder(
	specification uint32,
	column *hexane.RawColumn,
) hexaneColumnEncoder {
	return bytesEncoder(specification, column.Bytes())
}

func bytesEncoder(specification uint32, data []byte) hexaneColumnEncoder {
	return hexaneColumnEncoder{
		specification: specification,
		bytes: func() ([]byte, error) {
			return data, nil
		},
	}
}

func encodeHexaneColumns(encoders []hexaneColumnEncoder) ([]encodedColumn, error) {
	columns := make([]encodedColumn, 0, len(encoders))
	for _, encoder := range encoders {
		data, err := encoder.bytes()
		if err != nil {
			return nil, fmt.Errorf(
				"cannot encode column %d: %w",
				encoder.specification,
				err,
			)
		}
		if len(data) > 0 {
			columns = append(
				columns,
				encodedColumn{
					specification: encoder.specification,
					data:          data,
				},
			)
		}
	}

	return columns, nil
}

func remapActorColumn(
	column *hexane.Column[uint64],
	oldActors []opset.ActorID,
	newIndexes map[opset.ActorID]uint64,
) (*hexane.Column[uint64], error) {
	values := column.Values()
	for i, value := range values {
		if !value.Valid {
			continue
		}
		if value.Value >= uint64(len(oldActors)) {
			return nil, fmt.Errorf("actor index %d is out of bounds", value.Value)
		}
		index, ok := newIndexes[oldActors[value.Value]]
		if !ok {
			return nil, fmt.Errorf("actor is absent from new actor table")
		}
		values[i] = hexane.Some(index)
	}

	return hexane.NewColumnFromValues(hexane.Uint64Codec(), values...), nil
}

func encodedColumnData(columns []encodedColumn, specification uint32) []byte {
	for _, column := range columns {
		if column.specification == specification {
			return column.data
		}
	}

	return nil
}

func hexaneValues[T any](values []optional[T]) []hexane.Value[T] {
	converted := make([]hexane.Value[T], len(values))
	for i, value := range values {
		converted[i] = hexaneValue(value)
	}

	return converted
}

func paddedHexaneValues[T any](
	values []optional[T],
	count int,
) []hexane.Value[T] {
	values = slices.Grow(values, count-len(values))
	for len(values) < count {
		values = append(values, optional[T]{})
	}

	return hexaneValues(values)
}

func presentUint64s(values []optional[uint64]) []uint64 {
	converted := make([]uint64, len(values))
	for i, value := range values {
		converted[i] = value.value
	}

	return converted
}

func metadataLengths(values []optional[uint64]) []uint64 {
	lengths := make([]uint64, len(values))
	for i, value := range values {
		if value.valid {
			lengths[i] = value.value >> 4
		}
	}

	return lengths
}

func decodeBooleanOrZero(data []byte, count int) ([]bool, error) {
	if len(data) == 0 {
		return make([]bool, count), nil
	}

	return decodeBooleanColumn(data, count)
}

func countTrue(values []bool) int {
	count := 0
	for _, value := range values {
		if value {
			count++
		}
	}

	return count
}
