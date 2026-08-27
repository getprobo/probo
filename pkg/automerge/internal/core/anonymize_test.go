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
	"math"
	"slices"
	"testing"
	"time"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

func TestEngineAnonymize_PreservesHistoryAndSource(t *testing.T) {
	t.Parallel()

	source, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, source.SetActor(makeActor(1)))
	require.NoError(t, source.PutString(0, "private-key", "secret value"))
	text, err := source.PutText(0, "private-text")
	require.NoError(t, err)
	require.NoError(t, source.SpliceText(text, 0, 0, "Alice 👋\nTomorrow"))
	require.NoError(
		t,
		source.MarkText(
			text,
			0,
			5,
			"private-mark",
			[]byte(`{"type":"string","string":"private-value"}`),
			"both",
		),
	)
	_, err = source.Commit("private first message", time.Unix(1_700_000_000, 0))
	require.NoError(t, err)

	base, err := source.Save(true, false)
	require.NoError(t, err)
	branch, err := LoadEngine(base)
	require.NoError(t, err)
	require.NoError(t, branch.SetActor(makeActor(2)))
	require.NoError(t, branch.PutString(0, "branch-key", "branch secret"))
	_, err = branch.Commit("private branch message", time.Unix(1_700_000_001, 0))
	require.NoError(t, err)

	require.NoError(t, source.PutString(0, "private-key", "another secret"))
	_, err = source.Commit("private second message", time.Unix(1_700_000_002, 0))
	require.NoError(t, err)
	branchData, err := branch.Save(true, false)
	require.NoError(t, err)
	_, err = source.Merge(branchData)
	require.NoError(t, err)

	before, err := source.Save(true, false)
	require.NoError(t, err)
	anonymized, err := source.Anonymize()
	require.NoError(t, err)
	after, err := source.Save(true, false)
	require.NoError(t, err)
	assert.Equal(t, before, after)

	anonymizedData, err := anonymized.Save(true, false)
	require.NoError(t, err)
	reloaded, err := LoadEngine(anonymizedData)
	require.NoError(t, err)
	reloadedData, err := reloaded.Save(true, false)
	require.NoError(t, err)
	assert.NotEmpty(t, reloadedData)

	sourceDocument, err := storage.Decode(before)
	require.NoError(t, err)
	anonymizedDocument, err := storage.Decode(anonymizedData)
	require.NoError(t, err)
	require.Len(t, anonymizedDocument.Changes, len(sourceDocument.Changes))

	pairs := pairChangesByActorRank(sourceDocument.Changes, anonymizedDocument.Changes)
	sourceByHash := changesByHash(sourceDocument.Changes)
	anonymizedByHash := changesByHash(anonymizedDocument.Changes)
	for _, pair := range pairs {
		sourceChange := pair.source
		anonymizedChange := pair.anonymized

		assert.Equal(t, sourceChange.Sequence, anonymizedChange.Sequence)
		assert.Equal(t, sourceChange.StartOp, anonymizedChange.StartOp)
		assert.Equal(t, sourceChange.MaxOp, anonymizedChange.MaxOp)
		assert.NotEqual(t, sourceChange.Actor, anonymizedChange.Actor)
		assert.NotEqual(t, sourceChange.Time, anonymizedChange.Time)
		assert.NotEqual(t, sourceChange.Message, anonymizedChange.Message)
		require.Len(t, anonymizedChange.Dependencies, len(sourceChange.Dependencies))
		for i := range sourceChange.Dependencies {
			sourceDependency := sourceByHash[sourceChange.Dependencies[i]]
			anonymizedDependency := anonymizedByHash[anonymizedChange.Dependencies[i]]
			assert.Equal(t, sourceDependency.Sequence, anonymizedDependency.Sequence)
			assert.Equal(
				t,
				actorRank(sourceDocument.Changes, sourceDependency.Actor),
				actorRank(anonymizedDocument.Changes, anonymizedDependency.Actor),
			)
		}
		require.Len(t, anonymizedChange.Operations, len(sourceChange.Operations))

		for j := range sourceChange.Operations {
			sourceOperation := sourceChange.Operations[j]
			anonymizedOperation := anonymizedChange.Operations[j]
			assert.Equal(t, sourceOperation.Action, anonymizedOperation.Action)
			assert.Equal(t, sourceOperation.Insert, anonymizedOperation.Insert)
			assert.Equal(t, sourceOperation.ID.Counter, anonymizedOperation.ID.Counter)
			assert.Equal(t, len(sourceOperation.Predecessors), len(anonymizedOperation.Predecessors))
			if sourceOperation.Key.Property != nil {
				require.NotNil(t, anonymizedOperation.Key.Property)
				assert.NotEqual(t, *sourceOperation.Key.Property, *anonymizedOperation.Key.Property)
				assert.Equal(t, len(*sourceOperation.Key.Property), len(*anonymizedOperation.Key.Property))
			}
			if sourceOperation.Value != nil && sourceOperation.Value.Type == opset.ScalarString {
				require.NotNil(t, anonymizedOperation.Value)
				if containsAnonymizedCharacter(sourceOperation.Value.String) {
					assert.NotEqual(t, sourceOperation.Value.String, anonymizedOperation.Value.String)
				}
				assert.Equal(t, len(sourceOperation.Value.String), len(anonymizedOperation.Value.String))
				assert.Equal(
					t,
					len(utf16.Encode([]rune(sourceOperation.Value.String))),
					len(utf16.Encode([]rune(anonymizedOperation.Value.String))),
				)
			}
			if sourceOperation.MarkName != nil {
				require.NotNil(t, anonymizedOperation.MarkName)
				assert.NotEqual(t, *sourceOperation.MarkName, *anonymizedOperation.MarkName)
			}
		}
	}
}

func TestAnonymization_RewritesReferencesAndScalarKinds(t *testing.T) {
	t.Parallel()

	actorOne, err := opset.NewActorID(makeActor(1))
	require.NoError(t, err)
	actorTwo, err := opset.NewActorID(makeActor(2))
	require.NoError(t, err)
	object := opset.OpID{Actor: actorOne, Counter: 1}
	element := opset.OpID{Actor: actorTwo, Counter: 2}
	predecessor := opset.OpID{Actor: actorOne, Counter: 3}
	property := "name-é-界-😀"
	markName := "private-mark"
	change := &opset.Change{
		Actor:    actorTwo,
		Sequence: 1,
		StartOp:  4,
		MaxOp:    4,
		Operations: []opset.Operation{
			{
				ID:           opset.OpID{Actor: actorTwo, Counter: 4},
				Object:       opset.ObjectID{OpID: object},
				Key:          opset.Key{Property: &property, Element: &element},
				Action:       opset.ActionMark,
				Value:        &opset.Scalar{Type: opset.ScalarString, String: "secret 👋"},
				Predecessors: []opset.OpID{predecessor},
				Successors:   []opset.OpID{element},
				MarkName:     &markName,
			},
		},
		ExtraBytes: []byte("private-extra"),
	}

	anonymizer, err := newAnonymization([]*opset.Change{change})
	require.NoError(t, err)
	anonymized, err := anonymizer.change(change)
	require.NoError(t, err)
	operation := anonymized.Operations[0]

	assert.NotEqual(t, change.Actor, anonymized.Actor)
	assert.Equal(t, object.Counter, operation.Object.OpID.Counter)
	assert.NotEqual(t, object.Actor, operation.Object.OpID.Actor)
	assert.Equal(t, element.Counter, operation.Key.Element.Counter)
	assert.NotEqual(t, element.Actor, operation.Key.Element.Actor)
	assert.Equal(t, predecessor.Counter, operation.Predecessors[0].Counter)
	assert.NotEqual(t, predecessor.Actor, operation.Predecessors[0].Actor)
	assert.NotEqual(t, property, *operation.Key.Property)
	assert.Equal(t, len(property), len(*operation.Key.Property))
	assert.Equal(t, utf8.RuneCountInString(property), utf8.RuneCountInString(*operation.Key.Property))
	assert.NotEqual(t, markName, *operation.MarkName)
	assert.NotEqual(t, change.ExtraBytes, anonymized.ExtraBytes)
	assert.Len(t, anonymized.ExtraBytes, len(change.ExtraBytes))

	scalars := []*opset.Scalar{
		{Type: opset.ScalarNull},
		{Type: opset.ScalarTrue, Bool: true},
		{Type: opset.ScalarUint, Uint: 42},
		{Type: opset.ScalarInt, Int: 42},
		{Type: opset.ScalarFloat64, Float: 42},
		{Type: opset.ScalarString, String: "private value"},
		{Type: opset.ScalarBytes, Bytes: []byte("private bytes")},
		{Type: opset.ScalarCounter, Int: 42},
		{Type: opset.ScalarTimestamp, Int: 42},
		{Type: opset.ScalarType(15), Raw: []byte("private raw")},
	}
	for _, scalar := range scalars {
		replacement, err := anonymizer.scalar(scalar)
		require.NoError(t, err)
		assertScalarShape(t, scalar, replacement)
	}
}

func TestAnonymization_RepeatedNumbersReceiveVariedReplacements(t *testing.T) {
	t.Parallel()

	anonymizer, err := newAnonymization(nil)
	require.NoError(t, err)

	signed := make(map[int64]struct{})
	unsigned := make(map[uint64]struct{})
	floats := make(map[uint64]struct{})

	for range 16 {
		signedValue, err := anonymizer.scalar(
			&opset.Scalar{Type: opset.ScalarInt, Int: 42},
		)
		require.NoError(t, err)
		signed[signedValue.Int] = struct{}{}

		unsignedValue, err := anonymizer.scalar(
			&opset.Scalar{Type: opset.ScalarUint, Uint: 42},
		)
		require.NoError(t, err)
		unsigned[unsignedValue.Uint] = struct{}{}

		floatValue, err := anonymizer.scalar(
			&opset.Scalar{Type: opset.ScalarFloat64, Float: 42},
		)
		require.NoError(t, err)
		floats[math.Float64bits(floatValue.Float)] = struct{}{}
	}

	assert.Greater(t, len(signed), 1)
	assert.Greater(t, len(unsigned), 1)
	assert.Greater(t, len(floats), 1)
	assert.NotContains(t, signed, int64(42))
	assert.NotContains(t, unsigned, uint64(42))
	assert.NotContains(t, floats, math.Float64bits(42))
}

func TestAnonymization_StructuralStringsPreserveShape(t *testing.T) {
	t.Parallel()

	anonymizer, err := newAnonymization(nil)
	require.NoError(t, err)

	source := []string{"name", "email", "a", "b", "é", "界", "😀"}
	replacements := make(map[string]struct{}, len(source))

	for _, value := range source {
		replacement := anonymizer.structuralString(
			value,
			anonymizer.structuralRotations,
		)
		assert.NotEqual(t, value, replacement)
		assert.Equal(t, len(value), len(replacement))
		assert.Equal(
			t,
			len(utf16.Encode([]rune(value))),
			len(utf16.Encode([]rune(replacement))),
		)
		replacements[replacement] = struct{}{}
	}

	assert.Len(t, replacements, len(source))
}

func makeActor(value byte) []byte {
	return []byte{
		value, value, value, value,
		value, value, value, value,
		value, value, value, value,
		value, value, value, value,
	}
}

type changePair struct {
	source     opset.Change
	anonymized opset.Change
}

func pairChangesByActorRank(source, anonymized []opset.Change) []changePair {
	anonymizedByRankSequence := make(map[[2]uint64]opset.Change, len(anonymized))
	for _, change := range anonymized {
		key := [2]uint64{uint64(actorRank(anonymized, change.Actor)), change.Sequence}
		anonymizedByRankSequence[key] = change
	}

	pairs := make([]changePair, 0, len(source))
	for _, change := range source {
		key := [2]uint64{uint64(actorRank(source, change.Actor)), change.Sequence}
		pairs = append(
			pairs,
			changePair{source: change, anonymized: anonymizedByRankSequence[key]},
		)
	}
	return pairs
}

func actorRank(changes []opset.Change, actor opset.ActorID) int {
	actors := make([]opset.ActorID, 0)
	for i := range changes {
		if !slices.Contains(actors, changes[i].Actor) {
			actors = append(actors, changes[i].Actor)
		}
	}
	slices.SortFunc(
		actors,
		func(left, right opset.ActorID) int {
			return left.Compare(right)
		},
	)
	return slices.Index(actors, actor)
}

func changesByHash(changes []opset.Change) map[opset.ChangeHash]opset.Change {
	result := make(map[opset.ChangeHash]opset.Change, len(changes))
	for _, change := range changes {
		result[*change.Hash] = change
	}
	return result
}

func containsAnonymizedCharacter(value string) bool {
	for _, character := range value {
		if !unicode.IsSpace(character) &&
			!(character <= unicode.MaxASCII && unicode.IsControl(character)) {
			return true
		}
	}
	return false
}

func assertScalarShape(t *testing.T, source, replacement *opset.Scalar) {
	t.Helper()

	if source.Type == opset.ScalarFalse || source.Type == opset.ScalarTrue {
		assert.Contains(t, []opset.ScalarType{opset.ScalarFalse, opset.ScalarTrue}, replacement.Type)
		return
	}

	assert.Equal(t, source.Type, replacement.Type)
	switch source.Type {
	case opset.ScalarNull:
	case opset.ScalarUint:
		assert.NotEqual(t, source.Uint, replacement.Uint)
	case opset.ScalarInt, opset.ScalarCounter, opset.ScalarTimestamp:
		assert.NotEqual(t, source.Int, replacement.Int)
	case opset.ScalarFloat64:
		assert.NotEqual(t, math.Float64bits(source.Float), math.Float64bits(replacement.Float))
		assert.True(t, replacement.Float >= 0 && replacement.Float < 1)
	case opset.ScalarString:
		assert.NotEqual(t, source.String, replacement.String)
		assert.Equal(t, len(source.String), len(replacement.String))
	case opset.ScalarBytes:
		assert.NotEqual(t, source.Bytes, replacement.Bytes)
		assert.Len(t, replacement.Bytes, len(source.Bytes))
	default:
		assert.NotEqual(t, source.Raw, replacement.Raw)
		assert.Len(t, replacement.Raw, len(source.Raw))
	}
}
