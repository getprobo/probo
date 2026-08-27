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
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"slices"
	"unicode"
	"unicode/utf8"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

const syntheticText = "loremipsumdolorsitametconsecteturadipiscingelit"

type (
	anonymization struct {
		actors              map[opset.ActorID]opset.ActorID
		hashes              map[opset.ChangeHash]opset.ChangeHash
		structuralRotations map[structuralAlphabet]uint32
		contentRotations    map[structuralAlphabet]uint32
		synthetic           []byte
		syntheticPosition   int
	}

	structuralAlphabet uint8
)

const (
	printableASCII structuralAlphabet = 1
	asciiControl   structuralAlphabet = 2
	twoByteUTF8    structuralAlphabet = 3
	threeByteUTF8  structuralAlphabet = 4
	fourByteUTF8   structuralAlphabet = 5
)

// Anonymize returns an independent engine whose committed history has the same
// causal graph and operation shape with identifying data replaced.
func (b *Engine) Anonymize() (*Engine, error) {
	state := b.state
	if b.isolationActive && b.fullState != nil {
		state = b.fullState
	}

	changes, ok := state.allChanges()
	if !ok {
		return nil, fmt.Errorf("cannot enumerate native history")
	}

	changes = append([]*opset.Change(nil), changes...)

	if len(b.pending) > 0 {
		sequence := b.state.sequenceForActor(b.actor) + 1
		changes = append(
			changes,
			&opset.Change{
				Actor:        b.actor,
				Sequence:     sequence,
				StartOp:      b.pending[0].ID.Counter,
				MaxOp:        b.pending[len(b.pending)-1].ID.Counter,
				Dependencies: b.changeDependencies(sequence),
				Operations:   append([]opset.Operation(nil), b.pending...),
			},
		)
	}

	anonymizer, err := newAnonymization(changes)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize native anonymization: %w", err)
	}

	anonymized, err := NewEngine()
	if err != nil {
		return nil, fmt.Errorf("cannot create anonymized native engine: %w", err)
	}

	for i, source := range changes {
		change, err := anonymizer.change(source)
		if err != nil {
			return nil, fmt.Errorf("cannot anonymize native change %d: %w", i, err)
		}

		raw, err := storage.EncodeChange(change)
		if err != nil {
			return nil, fmt.Errorf("cannot encode anonymized native change %d: %w", i, err)
		}

		if err := anonymized.ApplyChanges([][]byte{raw}); err != nil {
			return nil, fmt.Errorf("cannot apply anonymized native change %d: %w", i, err)
		}

		if source.Hash != nil {
			anonymizer.hashes[*source.Hash] = *change.Hash
		}
	}

	return anonymized, nil
}

func newAnonymization(changes []*opset.Change) (*anonymization, error) {
	synthetic := []byte(syntheticText)
	if err := shuffleBytes(synthetic); err != nil {
		return nil, err
	}

	position, err := randomIndex(uint64(len(synthetic)))
	if err != nil {
		return nil, err
	}

	a := &anonymization{
		actors:              make(map[opset.ActorID]opset.ActorID),
		hashes:              make(map[opset.ChangeHash]opset.ChangeHash),
		structuralRotations: make(map[structuralAlphabet]uint32),
		contentRotations:    make(map[structuralAlphabet]uint32),
		synthetic:           synthetic,
		syntheticPosition:   int(position),
	}

	for _, alphabet := range []structuralAlphabet{
		printableASCII,
		asciiControl,
		twoByteUTF8,
		threeByteUTF8,
		fourByteUTF8,
	} {
		_, _, count := structuralRank(representativeRune(alphabet))
		for _, rotations := range []map[structuralAlphabet]uint32{
			a.structuralRotations,
			a.contentRotations,
		} {
			random, err := randomIndex(uint64(count - 1))
			if err != nil {
				return nil, err
			}

			rotations[alphabet] = uint32(random) + 1
		}
	}

	actors := collectActors(changes)

	for {
		var prefix [8]byte
		if _, err := rand.Read(prefix[:]); err != nil {
			return nil, fmt.Errorf("cannot generate anonymized actor prefix: %w", err)
		}

		mapped := make(map[opset.ActorID]opset.ActorID, len(actors))
		collides := false

		for rank, actor := range actors {
			var replacement [16]byte
			copy(replacement[:8], prefix[:])
			binary.BigEndian.PutUint64(replacement[8:], uint64(rank))

			mapped[actor] = opset.ActorID(string(replacement[:]))
			if _, exists := slices.BinarySearchFunc(
				actors,
				mapped[actor],
				func(left, right opset.ActorID) int {
					return left.Compare(right)
				},
			); exists {
				collides = true
				break
			}
		}

		if !collides {
			a.actors = mapped
			break
		}
	}

	return a, nil
}

func collectActors(changes []*opset.Change) []opset.ActorID {
	set := make(map[opset.ActorID]struct{})
	add := func(actor opset.ActorID) {
		if actor != "" {
			set[actor] = struct{}{}
		}
	}

	for _, change := range changes {
		add(change.Actor)

		for _, operation := range change.Operations {
			add(operation.ID.Actor)

			if !operation.Object.IsRoot {
				add(operation.Object.OpID.Actor)
			}

			if operation.Key.Element != nil {
				add(operation.Key.Element.Actor)
			}

			for _, predecessor := range operation.Predecessors {
				add(predecessor.Actor)
			}

			for _, successor := range operation.Successors {
				add(successor.Actor)
			}
		}
	}

	actors := make([]opset.ActorID, 0, len(set))
	for actor := range set {
		actors = append(actors, actor)
	}

	slices.SortFunc(
		actors,
		func(left, right opset.ActorID) int {
			return left.Compare(right)
		},
	)

	return actors
}

func (a *anonymization) change(source *opset.Change) (*opset.Change, error) {
	extraBytes := source.ExtraBytes
	if len(extraBytes) == 0 && source.Extra != nil {
		extraBytes = source.Extra.Bytes
	}

	change := &opset.Change{
		Actor:        a.actor(source.Actor),
		Sequence:     source.Sequence,
		StartOp:      source.StartOp,
		MaxOp:        source.MaxOp,
		Message:      a.contentString(source.Message),
		Dependencies: make([]opset.ChangeHash, len(source.Dependencies)),
		Operations:   make([]opset.Operation, len(source.Operations)),
		ExtraBytes:   a.bytes(extraBytes),
	}

	time, err := randomInt64OtherThan(source.Time)
	if err != nil {
		return nil, fmt.Errorf("cannot anonymize change time: %w", err)
	}

	change.Time = time

	for i, dependency := range source.Dependencies {
		mapped, ok := a.hashes[dependency]
		if !ok {
			return nil, fmt.Errorf("dependency %s has not been anonymized", dependency)
		}

		change.Dependencies[i] = mapped
	}

	for i := range source.Operations {
		operation, err := a.operation(source.Operations[i])
		if err != nil {
			return nil, fmt.Errorf("cannot anonymize operation %d: %w", i, err)
		}

		change.Operations[i] = operation
	}

	return change, nil
}

func (a *anonymization) operation(source opset.Operation) (opset.Operation, error) {
	operation := source

	operation.ID = a.opID(source.ID)
	if !source.Object.IsRoot {
		operation.Object.OpID = a.opID(source.Object.OpID)
	}

	if source.Key.Property != nil {
		property := a.structuralString(*source.Key.Property, a.structuralRotations)
		operation.Key.Property = new(property)
	}

	if source.Key.Element != nil {
		operation.Key.Element = new(a.opID(*source.Key.Element))
	}

	operation.Predecessors = a.opIDs(source.Predecessors)
	operation.Successors = a.opIDs(source.Successors)

	value, err := a.scalar(source.Value)
	if err != nil {
		return opset.Operation{}, fmt.Errorf("cannot anonymize scalar: %w", err)
	}

	operation.Value = value

	if source.MarkName != nil {
		name := a.structuralString(*source.MarkName, a.structuralRotations)
		operation.MarkName = new(name)
	}

	return operation, nil
}

func (a *anonymization) scalar(source *opset.Scalar) (*opset.Scalar, error) {
	if source == nil {
		return nil, nil
	}

	value := *source
	switch source.Type {
	case opset.ScalarNull:
	case opset.ScalarFalse, opset.ScalarTrue:
		replacement, err := randomBool()
		if err != nil {
			return nil, err
		}

		value.Bool = replacement
		if replacement {
			value.Type = opset.ScalarTrue
		} else {
			value.Type = opset.ScalarFalse
		}
	case opset.ScalarUint:
		replacement, err := randomUint64OtherThan(source.Uint)
		if err != nil {
			return nil, err
		}

		value.Uint = replacement
	case opset.ScalarInt, opset.ScalarCounter, opset.ScalarTimestamp:
		replacement, err := randomInt64OtherThan(source.Int)
		if err != nil {
			return nil, err
		}

		value.Int = replacement
	case opset.ScalarFloat64:
		replacement, err := randomFloat64OtherThan(source.Float)
		if err != nil {
			return nil, err
		}

		value.Float = replacement
	case opset.ScalarString:
		value.String = a.contentString(source.String)
	case opset.ScalarBytes:
		value.Bytes = a.bytes(source.Bytes)
	default:
		value.Raw = a.bytes(source.Raw)
	}

	return &value, nil
}

func (a *anonymization) actor(actor opset.ActorID) opset.ActorID {
	replacement, ok := a.actors[actor]
	if !ok {
		panic("every referenced actor must be mapped")
	}

	return replacement
}

func (a *anonymization) opID(identifier opset.OpID) opset.OpID {
	identifier.Actor = a.actor(identifier.Actor)
	return identifier
}

func (a *anonymization) opIDs(identifiers []opset.OpID) []opset.OpID {
	mapped := make([]opset.OpID, len(identifiers))
	for i, identifier := range identifiers {
		mapped[i] = a.opID(identifier)
	}

	return mapped
}

func (a *anonymization) structuralString(
	value string,
	rotations map[structuralAlphabet]uint32,
) string {
	result := make([]byte, 0, len(value))
	for _, source := range value {
		alphabet, rank, count := structuralRank(source)
		rotation := rotations[alphabet]
		replacement := structuralRune(source, (rank+rotation)%count)
		result = utf8.AppendRune(result, replacement)
	}

	return string(result)
}

func (a *anonymization) contentString(value string) string {
	result := make([]byte, 0, len(value))
	for _, source := range value {
		switch {
		case unicode.IsSpace(source), source <= unicode.MaxASCII && unicode.IsControl(source):
			result = utf8.AppendRune(result, source)
		case source <= unicode.MaxASCII:
			result = append(result, a.syntheticByte(byte(source)))
		default:
			replacement := a.structuralString(string(source), a.contentRotations)
			result = append(result, replacement...)
		}
	}

	return string(result)
}

func (a *anonymization) bytes(value []byte) []byte {
	result := make([]byte, len(value))
	for i, source := range value {
		result[i] = a.syntheticByte(source)
	}

	return result
}

func (a *anonymization) syntheticByte(original byte) byte {
	for {
		replacement := a.synthetic[a.syntheticPosition]
		a.syntheticPosition = (a.syntheticPosition + 1) % len(a.synthetic)

		if replacement != original {
			return replacement
		}
	}
}

func structuralRank(value rune) (structuralAlphabet, uint32, uint32) {
	codepoint := uint32(value)
	switch utf8.RuneLen(value) {
	case 1:
		if codepoint >= 0x20 && codepoint <= 0x7e {
			return printableASCII, codepoint - 0x20, 0x7e - 0x20 + 1
		}

		if codepoint < 0x20 {
			return asciiControl, codepoint, 0x21
		}

		return asciiControl, 0x20, 0x21
	case 2:
		return twoByteUTF8, codepoint - 0x80, 0x800 - 0x80
	case 3:
		if codepoint < 0xd800 {
			return threeByteUTF8, codepoint - 0x800, 0xd800 - 0x800 + 0x2000
		}

		return threeByteUTF8, 0xd800 - 0x800 + codepoint - 0xe000, 0xd800 - 0x800 + 0x2000
	case 4:
		return fourByteUTF8, codepoint - 0x10000, 0x110000 - 0x10000
	default:
		panic("Go runes use one to four UTF-8 bytes")
	}
}

func structuralRune(original rune, rank uint32) rune {
	switch utf8.RuneLen(original) {
	case 1:
		switch {
		case original >= ' ' && original <= '~':
			return rune(0x20 + rank)
		case original < ' ':
			return rune(rank)
		case rank < 0x20:
			return rune(rank)
		default:
			return 0x7f
		}
	case 2:
		return rune(0x80 + rank)
	case 3:
		if rank < 0xd800-0x800 {
			return rune(0x800 + rank)
		}

		return rune(0xe000 + rank - (0xd800 - 0x800))
	case 4:
		return rune(0x10000 + rank)
	default:
		panic("Go runes use one to four UTF-8 bytes")
	}
}

func representativeRune(alphabet structuralAlphabet) rune {
	switch alphabet {
	case printableASCII:
		return 'a'
	case asciiControl:
		return '\x00'
	case twoByteUTF8:
		return '\u0080'
	case threeByteUTF8:
		return '\u0800'
	case fourByteUTF8:
		return '\U00010000'
	default:
		panic("unknown structural alphabet")
	}
}

func shuffleBytes(value []byte) error {
	for i := len(value) - 1; i > 0; i-- {
		index, err := randomIndex(uint64(i + 1))
		if err != nil {
			return err
		}

		value[i], value[index] = value[index], value[i]
	}

	return nil
}

func randomIndex(limit uint64) (uint64, error) {
	value, err := rand.Int(rand.Reader, new(big.Int).SetUint64(limit))
	if err != nil {
		return 0, fmt.Errorf("cannot read cryptographic randomness: %w", err)
	}

	return value.Uint64(), nil
}

func randomInt64OtherThan(original int64) (int64, error) {
	for {
		var raw [4]byte
		if _, err := rand.Read(raw[:]); err != nil {
			return 0, fmt.Errorf("cannot read cryptographic randomness: %w", err)
		}

		replacement := int64(int32(binary.LittleEndian.Uint32(raw[:])))
		if replacement != original {
			return replacement, nil
		}
	}
}

func randomUint64OtherThan(original uint64) (uint64, error) {
	for {
		var raw [8]byte
		if _, err := rand.Read(raw[:]); err != nil {
			return 0, fmt.Errorf("cannot read cryptographic randomness: %w", err)
		}

		replacement := binary.LittleEndian.Uint64(raw[:])
		if replacement != original {
			return replacement, nil
		}
	}
}

func randomFloat64OtherThan(original float64) (float64, error) {
	for {
		value, err := randomUint64OtherThan(math.MaxUint64)
		if err != nil {
			return 0, err
		}

		replacement := float64(value>>11) / (1 << 53)
		if math.Float64bits(replacement) != math.Float64bits(original) {
			return replacement, nil
		}
	}
}

func randomBool() (bool, error) {
	var value [1]byte
	if _, err := rand.Read(value[:]); err != nil {
		return false, fmt.Errorf("cannot read cryptographic randomness: %w", err)
	}

	return value[0]&1 == 1, nil
}
