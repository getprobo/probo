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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func (b *Engine) Stats() ([]byte, error) {
	actors := make(map[opset.ActorID]struct{})
	for id := range b.state.operations {
		actors[id.Actor] = struct{}{}
	}

	stats := struct {
		NumChanges uint64 `json:"numChanges"`
		NumOps     uint64 `json:"numOps"`
		NumActors  uint64 `json:"numActors"`
	}{
		NumChanges: uint64(len(b.state.changes)),
		NumOps:     uint64(len(b.state.operations)),
		NumActors:  uint64(len(actors)),
	}

	data, err := json.Marshal(stats)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native stats: %w", err)
	}

	return data, nil
}

type (
	patchOut struct {
		Obj    string         `json:"obj"`
		Action patchActionOut `json:"action"`
	}

	patchActionOut struct {
		Type     string           `json:"type"`
		Key      string           `json:"key,omitempty"`
		Index    uint64           `json:"index"`
		Length   uint64           `json:"length,omitempty"`
		Value    *patchValueOut   `json:"value,omitempty"`
		Values   []patchInsertOut `json:"values,omitempty"`
		Text     string           `json:"text,omitempty"`
		Conflict bool             `json:"conflict"`
		Marks    []markPatchOut   `json:"marks,omitempty"`
	}

	markPatchOut struct {
		Start uint32          `json:"start"`
		End   uint32          `json:"end"`
		Name  string          `json:"name"`
		Value json.RawMessage `json:"value"`
	}

	patchInsertOut struct {
		Value    patchValueOut `json:"value"`
		Conflict bool          `json:"conflict"`
	}

	patchValueOut struct {
		Scalar json.RawMessage `json:"scalar,omitempty"`
		Object string          `json:"object,omitempty"`
		ID     string          `json:"id,omitempty"`
	}
)

func objectIDString(object opset.ObjectID) string {
	if object.IsRoot {
		return "_root"
	}

	return fmt.Sprintf(
		"%d@%s",
		object.OpID.Counter,
		hex.EncodeToString([]byte(object.OpID.Actor)),
	)
}

func patchValueForOperation(state *State, operation opset.Operation) (patchValueOut, error) {
	if objectType, err := actionObjectType(operation.Action); err == nil {
		return patchValueOut{
			Object: objectType,
			ID:     objectIDString(opset.ObjectID{OpID: operation.ID}),
		}, nil
	}

	value, ok := state.scalarValue(operation)
	if !ok {
		return patchValueOut{}, fmt.Errorf("operation %v has no materializable value", operation.ID)
	}

	encoded, err := encodeScalarWire(value)
	if err != nil {
		return patchValueOut{}, err
	}

	return patchValueOut{Scalar: json.RawMessage(encoded)}, nil
}

// CurrentState returns the patches that materialize the document from empty,
// ordered to match upstream Rust: the root first, then other objects by
// creation operation ID, with map keys in lexical order and sequence elements
// in index order.
func (b *Engine) CurrentState() ([]byte, error) {
	patches := make([]patchOut, 0)

	for _, object := range orderedObjectsInState(b.state) {
		objectPatches, err := materializeObjectPatches(b.state, object)
		if err != nil {
			return nil, err
		}

		patches = append(patches, objectPatches...)
	}

	data, err := json.Marshal(patches)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native patches: %w", err)
	}

	return data, nil
}

// orderedObjectsInState returns the visible objects in a state, the root first
// and then every non-deleted composite object ordered by its creation ID.
func orderedObjectsInState(state *State) []opset.ObjectID {
	objects := []opset.ObjectID{opset.RootObject()}

	makers := make([]opset.Operation, 0)

	for _, operation := range state.operations {
		if _, err := actionObjectType(operation.Action); err != nil {
			continue
		}

		if state.isSuperseded(operation.ID) {
			continue
		}

		// A composite object concurrently assigned to the same map key as another
		// object is a conflict alternative; only the winning value is materialized
		// (its content spliced), matching the reference. Losing alternatives still
		// exist but are surfaced only through the put's conflict flag.
		if !operation.Insert && operation.Key.Property != nil {
			if winner, ok := state.visibleMapObjectValue(operation.Object, *operation.Key.Property); ok &&
				winner.ID != operation.ID {
				continue
			}
		}

		makers = append(makers, operation)
	}

	sort.Slice(
		makers,
		func(i, j int) bool {
			return makers[i].ID.Compare(makers[j].ID) < 0
		},
	)

	for _, operation := range makers {
		objects = append(objects, opset.ObjectID{OpID: operation.ID})
	}

	return objects
}

func objectTypeInState(state *State, object opset.ObjectID) (string, error) {
	if object.IsRoot {
		return "map", nil
	}

	operation, ok := state.operations[object.OpID]
	if !ok {
		return "", fmt.Errorf("object %v does not exist", object.OpID)
	}

	return actionObjectType(operation.Action)
}

func objectVisibleInState(state *State, object opset.ObjectID) bool {
	if object.IsRoot {
		return true
	}

	if _, ok := state.operations[object.OpID]; !ok {
		return false
	}

	return !state.isSuperseded(object.OpID)
}

// materializeObjectPatches emits the patches that build an object from empty.
func materializeObjectPatches(state *State, object opset.ObjectID) ([]patchOut, error) {
	objectType, err := objectTypeInState(state, object)
	if err != nil {
		return nil, err
	}

	identifier := objectIDString(object)

	switch objectType {
	case "map", "table":
		patches := make([]patchOut, 0)

		for _, key := range state.mapKeys(object) {
			winner, ok := state.visibleMapObjectValue(object, key)
			if !ok {
				continue
			}

			value, err := patchValueForOperation(state, winner)
			if err != nil {
				return nil, err
			}

			patches = append(
				patches,
				patchOut{
					Obj: identifier,
					Action: patchActionOut{
						Type:     "put_map",
						Key:      key,
						Value:    &value,
						Conflict: len(state.visibleMapObjectOperations(object, key)) > 1,
					},
				},
			)
		}

		return patches, nil
	case "list":
		values := state.sequenceValues(object.OpID)
		if len(values) == 0 {
			return nil, nil
		}

		inserts := make([]patchInsertOut, 0, len(values))

		for index := range values {
			value, err := patchValueForOperation(state, values[index].Operation)
			if err != nil {
				return nil, err
			}

			inserts = append(
				inserts,
				patchInsertOut{
					Value:    value,
					Conflict: len(state.visibleSequenceElementOperations(values[index].Element)) > 1,
				},
			)
		}

		return []patchOut{{
			Obj:    identifier,
			Action: patchActionOut{Type: "insert", Index: 0, Values: inserts},
		}}, nil
	case "text":
		patches := make([]patchOut, 0)
		position := uint64(0)

		var run strings.Builder

		runStart := uint64(0)

		flush := func() error {
			if run.Len() == 0 {
				return nil
			}

			runs, err := textRunsWithMarks(state, object, runStart, run.String())
			if err != nil {
				return err
			}

			for _, textRun := range runs {
				patches = append(
					patches,
					patchOut{
						Obj: identifier,
						Action: patchActionOut{
							Type:  "splice_text",
							Index: textRun.index,
							Text:  textRun.text,
							Marks: textRun.marks,
						},
					},
				)
			}

			run.Reset()

			return nil
		}

		for _, value := range state.sequenceValues(object.OpID) {
			operation := value.Operation

			if operation.Action == opset.ActionMakeMap {
				if err := flush(); err != nil {
					return nil, err
				}

				blockValue, err := patchValueForOperation(state, operation)
				if err != nil {
					return nil, err
				}

				patches = append(
					patches,
					patchOut{
						Obj: identifier,
						Action: patchActionOut{
							Type:   "insert",
							Index:  position,
							Values: []patchInsertOut{{Value: blockValue}},
						},
					},
				)
				position++

				continue
			}

			if operation.Value != nil && operation.Value.Type == opset.ScalarString {
				if run.Len() == 0 {
					runStart = position
				}

				run.WriteString(operation.Value.String)
				position += uint64(utf16Width(operation.Value.String))
			}
		}

		if err := flush(); err != nil {
			return nil, err
		}

		return patches, nil
	default:
		return nil, fmt.Errorf("unknown object type %q", objectType)
	}
}

// Diff returns the patches that transform the document state at the before heads
// into the state at the after heads.
func (b *Engine) Diff(
	beforeHeads [][32]byte,
	afterHeads [][32]byte,
) ([]byte, error) {
	patches, err := b.diffPatches(beforeHeads, afterHeads)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(patches)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native diff patches: %w", err)
	}

	return data, nil
}

// UpdateDiffCursor records the current heads as the incremental diff cursor so a
// following DiffIncremental reports only the changes committed since this call.
func (b *Engine) UpdateDiffCursor() error {
	heads, err := b.Heads()
	if err != nil {
		return err
	}

	b.diffCursor = heads
	b.isolationDiffTargets = nil

	return nil
}

// DiffIncremental returns the patches for the changes committed since the diff
// cursor (or from an empty document when the cursor is unset) and advances the
// cursor to the current heads, mirroring the reference diff_incremental helper.
func (b *Engine) DiffIncremental() ([]byte, error) {
	heads, err := b.Heads()
	if err != nil {
		return nil, err
	}

	patches, err := b.incrementalDiffPatches(heads)
	if err != nil {
		return nil, err
	}

	b.diffCursor = heads
	b.isolationDiffTargets = nil

	data, err := json.Marshal(patches)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native incremental diff patches: %w", err)
	}

	return data, nil
}

// incrementalDiffPatches computes the incremental patches from the diff cursor
// to the current heads. When isolation frontiers were recorded in the window,
// the diff is chained through each of them (cursor to each isolation frontier
// and finally to the current heads) so the patch stream matches the reference's
// patch-log output across isolate/integrate.
func (b *Engine) incrementalDiffPatches(heads [][32]byte) ([]patchOut, error) {
	if len(b.isolationDiffTargets) == 0 {
		return b.incrementalPatches(b.diffCursor, heads)
	}

	frontiers := make([][][32]byte, 0, len(b.isolationDiffTargets)+2)
	frontiers = append(frontiers, b.diffCursor)
	frontiers = append(frontiers, b.isolationDiffTargets...)
	frontiers = append(frontiers, heads)

	patches := make([]patchOut, 0)

	for i := 0; i+1 < len(frontiers); i++ {
		segment, err := b.incrementalPatches(frontiers[i], frontiers[i+1])
		if err != nil {
			return nil, err
		}

		patches = append(patches, segment...)
	}

	return patches, nil
}

func (b *Engine) incrementalPatches(
	beforeHeads [][32]byte,
	afterHeads [][32]byte,
) ([]patchOut, error) {
	return b.diffPatches(beforeHeads, afterHeads)
}

func (b *Engine) diffPatches(
	beforeHeads [][32]byte,
	afterHeads [][32]byte,
) ([]patchOut, error) {
	source, ok := b.state.at(nativeHashes(beforeHeads))
	if !ok {
		return nil, fmt.Errorf("before heads are unknown")
	}

	target, ok := b.state.at(nativeHashes(afterHeads))
	if !ok {
		return nil, fmt.Errorf("after heads are unknown")
	}

	patches := make([]patchOut, 0)

	for _, object := range orderedObjectsInState(target) {
		var (
			objectPatches []patchOut
			err           error
		)

		if objectVisibleInState(source, object) {
			objectPatches, err = diffObjectPatches(source, target, object)
		} else {
			objectPatches, err = materializeObjectPatches(target, object)
		}

		if err != nil {
			return nil, err
		}

		patches = append(patches, objectPatches...)
	}

	return patches, nil
}

// diffObjectPatches emits patches transforming an object from the source state
// into the target state, for an object present in both.
func diffObjectPatches(source, target *State, object opset.ObjectID) ([]patchOut, error) {
	objectType, err := objectTypeInState(target, object)
	if err != nil {
		return nil, err
	}

	identifier := objectIDString(object)

	switch objectType {
	case "map", "table":
		return diffMapPatches(source, target, object, identifier)
	case "list":
		return diffSequencePatches(source, target, object, objectType, identifier)
	case "text":
		patches, err := diffSequencePatches(source, target, object, objectType, identifier)
		if err != nil {
			return nil, err
		}

		return mergeTextMarkPatches(source, target, object, identifier, patches)
	default:
		return nil, fmt.Errorf("unknown object type %q", objectType)
	}
}

// mergeTextMarkPatches folds the mark differences between the source and target
// states into the ordered sequence patches for a text object. Added or changed
// marks carry their new value; removed marks carry a null value. The reference
// emits a single mark patch positioned by the smallest affected index, so the
// combined patch is inserted before the first sequence patch beyond that index.
func mergeTextMarkPatches(
	source *State,
	target *State,
	object opset.ObjectID,
	identifier string,
	patches []patchOut,
) ([]patchOut, error) {
	// The reference derives mark patches from the mark operations applied in the
	// window, not from state comparison, so a mark range that merely grew because
	// text was spliced into an expanding mark produces no mark patch (the marks
	// ride on the splice patch instead), and a partial unmark reports the literal
	// operation range rather than the resulting split.
	marks, err := diffMarkPatches(source, target, object)
	if err != nil {
		return nil, err
	}

	if len(marks) == 0 {
		return patches, nil
	}

	anchor := marks[0].Start
	for _, mark := range marks[1:] {
		if mark.Start < anchor {
			anchor = mark.Start
		}
	}

	markPatch := patchOut{
		Obj:    identifier,
		Action: patchActionOut{Type: "mark", Marks: marks},
	}

	insertAt := len(patches)

	for index, patch := range patches {
		if patch.Action.Index > uint64(anchor) {
			insertAt = index

			break
		}
	}

	merged := make([]patchOut, 0, len(patches)+1)
	merged = append(merged, patches[:insertAt]...)
	merged = append(merged, markPatch)
	merged = append(merged, patches[insertAt:]...)

	return merged, nil
}

// diffMarkPatches reports the mark and unmark operations applied in the window
// between the source and target states as Mark patch entries, ordered by the
// operation identifier so they appear in application order. Each entry carries
// the operation's literal UTF-16 range and value (null for an unmark), matching
// the reference's operation-based diff rather than a state comparison.
func diffMarkPatches(source, target *State, object opset.ObjectID) ([]markPatchOut, error) {
	begins := make([]opset.Operation, 0)

	for id, operation := range target.operations {
		if operation.Action != opset.ActionMark ||
			operation.Object != object ||
			operation.MarkName == nil {
			continue
		}

		if _, ok := source.operations[id]; ok {
			continue
		}

		begins = append(begins, operation)
	}

	sort.Slice(
		begins,
		func(i, j int) bool {
			return begins[i].ID.Compare(begins[j].ID) < 0
		},
	)

	out := make([]markPatchOut, 0, len(begins))

	for _, begin := range begins {
		end, ok := target.operations[opset.OpID{Actor: begin.ID.Actor, Counter: begin.ID.Counter + 1}]
		if !ok {
			continue
		}

		start, finish, ok := target.markOpUTF16Range(object.OpID, begin, end)
		if !ok {
			continue
		}

		value := opset.Scalar{Type: opset.ScalarNull}
		if begin.Value != nil {
			value = *begin.Value
		}

		encoded, err := encodeScalarWire(value)
		if err != nil {
			return nil, err
		}

		out = append(
			out,
			markPatchOut{
				Start: start,
				End:   finish,
				Name:  *begin.MarkName,
				Value: json.RawMessage(encoded),
			},
		)
	}

	return out, nil
}

type textRun struct {
	index uint64
	text  string
	marks []markPatchOut
}

// textRunsWithMarks splits a run of text starting at the given UTF-16 position
// into maximal sub-runs that share the same active mark set, attaching those
// marks (sorted by name) to each sub-run. It mirrors the reference, which emits
// one splice_text patch per mark run.
func textRunsWithMarks(
	state *State,
	object opset.ObjectID,
	startPosition uint64,
	text string,
) ([]textRun, error) {
	ranges := state.Marks(object.OpID)

	activeMarks := func(position uint64) ([]markPatchOut, string, error) {
		marks := make([]markPatchOut, 0)

		for _, candidate := range ranges {
			if uint64(candidate.Start) > position || position >= uint64(candidate.End) {
				continue
			}

			value := opset.Scalar{Type: opset.ScalarNull}
			if candidate.Value != nil {
				value = *candidate.Value
			}

			encoded, err := encodeScalarWire(value)
			if err != nil {
				return nil, "", err
			}

			marks = append(
				marks,
				markPatchOut{
					Name:  candidate.Name,
					Value: json.RawMessage(encoded),
				},
			)
		}

		sort.Slice(
			marks,
			func(i, j int) bool {
				return marks[i].Name < marks[j].Name
			},
		)

		var key strings.Builder
		for _, mark := range marks {
			key.WriteString(mark.Name)
			key.WriteByte('=')
			key.Write(mark.Value)
			key.WriteByte(';')
		}

		if len(marks) == 0 {
			marks = nil
		}

		return marks, key.String(), nil
	}

	runs := make([]textRun, 0)

	var (
		builder  strings.Builder
		runMarks []markPatchOut
		runKey   string
		runStart = startPosition
		position = startPosition
		haveRun  bool
	)

	for _, character := range text {
		marks, key, err := activeMarks(position)
		if err != nil {
			return nil, err
		}

		if !haveRun {
			runStart = position
			runMarks = marks
			runKey = key
			haveRun = true
		} else if key != runKey {
			runs = append(runs, textRun{index: runStart, text: builder.String(), marks: runMarks})
			builder.Reset()

			runStart = position
			runMarks = marks
			runKey = key
		}

		builder.WriteRune(character)

		if character > 0xFFFF {
			position += 2
		} else {
			position++
		}
	}

	if haveRun {
		runs = append(runs, textRun{index: runStart, text: builder.String(), marks: runMarks})
	}

	return runs, nil
}

func diffMapPatches(
	source *State,
	target *State,
	object opset.ObjectID,
	identifier string,
) ([]patchOut, error) {
	keys := make(map[string]struct{})
	for _, key := range source.mapKeys(object) {
		keys[key] = struct{}{}
	}

	for _, key := range target.mapKeys(object) {
		keys[key] = struct{}{}
	}

	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}

	sort.Strings(ordered)

	patches := make([]patchOut, 0)

	for _, key := range ordered {
		targetOp, targetOK := target.visibleMapObjectValue(object, key)
		sourceOp, sourceOK := source.visibleMapObjectValue(object, key)

		switch {
		case targetOK && (!sourceOK || targetOp.ID != sourceOp.ID):
			value, err := patchValueForOperation(target, targetOp)
			if err != nil {
				return nil, err
			}

			patches = append(
				patches,
				patchOut{
					Obj: identifier,
					Action: patchActionOut{
						Type:     "put_map",
						Key:      key,
						Value:    &value,
						Conflict: len(target.visibleMapObjectOperations(object, key)) > 1,
					},
				},
			)
		case !targetOK && sourceOK:
			patches = append(
				patches,
				patchOut{
					Obj:    identifier,
					Action: patchActionOut{Type: "delete_map", Key: key},
				},
			)
		}
	}

	return patches, nil
}

func diffSequencePatches(
	source *State,
	target *State,
	object opset.ObjectID,
	objectType string,
	identifier string,
) ([]patchOut, error) {
	sourceValues := source.sequenceValues(object.OpID)
	targetValues := target.sequenceValues(object.OpID)

	sourceElements := make(map[opset.OpID]struct{}, len(sourceValues))
	for _, value := range sourceValues {
		sourceElements[value.Element] = struct{}{}
	}

	targetElements := make(map[opset.OpID]struct{}, len(targetValues))
	for _, value := range targetValues {
		targetElements[value.Element] = struct{}{}
	}

	// Text objects report positions in UTF-16 code units; other sequences use
	// one unit per element.
	width := func(value sequenceValue) uint64 {
		if objectType == "text" {
			return sequenceValueUTF16Width(value)
		}

		return 1
	}

	patches := make([]patchOut, 0)
	position := uint64(0)
	i, j := 0, 0

	for i < len(sourceValues) || j < len(targetValues) {
		if i < len(sourceValues) && j < len(targetValues) &&
			sourceValues[i].Element == targetValues[j].Element {
			if sourceValues[i].Operation.ID == targetValues[j].Operation.ID {
				position += width(targetValues[j])
				i++
				j++

				continue
			}

			// Same element, different winning value. Text string replacements
			// become a delete followed by a splice in both state-comparison and
			// incremental diffs; lists and non-string text values use put_seq.
			replacement := targetValues[j].Operation
			if objectType == "text" &&
				replacement.Value != nil &&
				replacement.Value.Type == opset.ScalarString {
				patches = append(
					patches,
					patchOut{
						Obj: identifier,
						Action: patchActionOut{
							Type:   "delete_seq",
							Index:  position,
							Length: width(sourceValues[i]),
						},
					},
				)

				runs, err := textRunsWithMarks(target, object, position, replacement.Value.String)
				if err != nil {
					return nil, err
				}

				for _, run := range runs {
					patches = append(
						patches,
						patchOut{
							Obj: identifier,
							Action: patchActionOut{
								Type:  "splice_text",
								Index: run.index,
								Text:  run.text,
								Marks: run.marks,
							},
						},
					)
				}

				position += width(targetValues[j])
			} else {
				value, err := patchValueForOperation(target, targetValues[j].Operation)
				if err != nil {
					return nil, err
				}

				patches = append(
					patches,
					patchOut{
						Obj: identifier,
						Action: patchActionOut{
							Type:     "put_seq",
							Index:    position,
							Value:    &value,
							Conflict: len(target.visibleSequenceElementOperations(targetValues[j].Element)) > 1,
						},
					},
				)
				position += width(targetValues[j])
			}

			i++
			j++

			continue
		}

		if i < len(sourceValues) {
			if _, ok := targetElements[sourceValues[i].Element]; !ok {
				length := uint64(0)

				for i < len(sourceValues) {
					if _, ok := targetElements[sourceValues[i].Element]; ok {
						break
					}

					length += width(sourceValues[i])
					i++
				}

				patches = append(
					patches,
					patchOut{
						Obj: identifier,
						Action: patchActionOut{
							Type:   "delete_seq",
							Index:  position,
							Length: length,
						},
					},
				)

				continue
			}
		}

		if objectType == "text" {
			var text strings.Builder

			start := position

			for j < len(targetValues) {
				if _, ok := sourceElements[targetValues[j].Element]; ok {
					break
				}

				operation := targetValues[j].Operation
				if operation.Value == nil || operation.Value.Type != opset.ScalarString {
					break
				}

				text.WriteString(operation.Value.String)

				position += width(targetValues[j])
				j++
			}

			if text.Len() > 0 {
				runs, err := textRunsWithMarks(target, object, start, text.String())
				if err != nil {
					return nil, err
				}

				for _, run := range runs {
					patches = append(
						patches,
						patchOut{
							Obj: identifier,
							Action: patchActionOut{
								Type:  "splice_text",
								Index: run.index,
								Text:  run.text,
								Marks: run.marks,
							},
						},
					)
				}

				continue
			}
		}

		inserts := make([]patchInsertOut, 0)
		start := position

		for j < len(targetValues) {
			if _, ok := sourceElements[targetValues[j].Element]; ok {
				break
			}

			value, err := patchValueForOperation(target, targetValues[j].Operation)
			if err != nil {
				return nil, err
			}

			inserts = append(
				inserts,
				patchInsertOut{
					Value:    value,
					Conflict: len(target.visibleSequenceElementOperations(targetValues[j].Element)) > 1,
				},
			)
			position++
			j++
		}

		if len(inserts) == 0 {
			break
		}

		patches = append(
			patches,
			patchOut{
				Obj: identifier,
				Action: patchActionOut{
					Type:   "insert",
					Index:  start,
					Values: inserts,
				},
			},
		)
	}

	return patches, nil
}
