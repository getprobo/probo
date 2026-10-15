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
	"strconv"
	"strings"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func (b *Engine) Stats() ([]byte, error) {
	actors := make(map[opset.ActorID]struct{})

	b.state.eachOperation(func(operation opset.Operation) bool {
		actors[operation.ID.Actor] = struct{}{}
		return true
	})

	stats := struct {
		NumChanges uint64 `json:"numChanges"`
		NumOps     uint64 `json:"numOps"`
		NumActors  uint64 `json:"numActors"`
	}{
		NumChanges: uint64(b.state.changeCount()),
		NumOps:     uint64(b.state.operationCount()),
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

	state.eachOperation(func(operation opset.Operation) bool {
		if _, err := actionObjectType(operation.Action); err != nil {
			return true
		}

		if state.isSuperseded(operation.ID) {
			return true
		}

		// A composite object concurrently assigned to the same key or list element
		// as another object is a conflict alternative. Only the winning value is
		// materialized; the loser is surfaced through the conflict flag or action.
		if !operation.Insert && operation.Key.Property != nil {
			if winner, ok := state.visibleMapObjectValue(operation.Object, *operation.Key.Property); ok &&
				winner.ID != operation.ID {
				return true
			}
		}

		if !operation.Insert && operation.Key.Element != nil {
			winner, ok := sequenceElementWinner(state, operation.Object, *operation.Key.Element)
			if !ok || winner.ID != operation.ID {
				return true
			}
		}

		makers = append(makers, operation)

		return true
	})

	sort.Slice(
		makers,
		func(i, j int) bool {
			return compareObjectCreationIDs(makers[i].ID, makers[j].ID) < 0
		},
	)

	for _, operation := range makers {
		objects = append(objects, opset.ObjectID{OpID: operation.ID})
	}

	return objects
}

func sequenceElementWinner(
	state *State,
	object opset.ObjectID,
	element opset.OpID,
) (opset.Operation, bool) {
	for _, value := range state.sequenceValues(object.OpID) {
		if value.Element == element {
			return value.Operation, true
		}
	}

	return opset.Operation{}, false
}

func compareObjectCreationIDs(left, right opset.OpID) int {
	if left.Counter != right.Counter {
		return left.Compare(right)
	}

	leftBase, leftLevel := left.Actor.Concurrency()
	rightBase, rightLevel := right.Actor.Concurrency()
	if compared := leftBase.Compare(rightBase); compared != 0 {
		return compared
	}

	switch {
	case leftLevel < rightLevel:
		return -1
	case leftLevel > rightLevel:
		return 1
	default:
		return left.Actor.Compare(right.Actor)
	}
}

func objectTypeInState(state *State, object opset.ObjectID) (string, error) {
	if object.IsRoot {
		return "map", nil
	}

	operation, ok := state.operation(object.OpID)
	if !ok {
		return "", fmt.Errorf("object %v does not exist", object.OpID)
	}

	return actionObjectType(operation.Action)
}

func objectVisibleInState(state *State, object opset.ObjectID) bool {
	if object.IsRoot {
		return true
	}

	if _, ok := state.operation(object.OpID); !ok {
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

				continue
			}

			if err := flush(); err != nil {
				return nil, err
			}

			rendered := textElementString(operation)
			runs, err := textRunsWithMarks(state, object, position, rendered)
			if err != nil {
				return nil, err
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

			position += uint64(utf16Width(rendered))
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
	patches, err := b.diffPatches(beforeHeads, afterHeads, false)
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

	b.resetPatchLog(heads)

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

	b.resetPatchLog(heads)

	data, err := json.Marshal(patches)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native incremental diff patches: %w", err)
	}

	return data, nil
}

// incrementalDiffPatches computes the incremental patches from the diff cursor
// to the current heads. When the patch log is live and the target is the full
// document heads, Rust returns the log (including isolation transitions)
// instead of a direct comparison. A target that is only an isolation frontier
// misses that fast path and is a clock diff from the cursor.
func (b *Engine) incrementalDiffPatches(heads [][32]byte) ([]patchOut, error) {
	if b.patchLogActive && sameHeads(heads, b.fullDocumentHeads()) {
		if err := b.flushPatchLog(); err != nil {
			return nil, err
		}

		return append([]patchOut(nil), b.patchLog...), nil
	}

	return b.diffPatches(b.diffCursor, heads, false)
}

// resetPatchLog matches AutoCommit::update_diff_cursor. Empty heads leave the
// log alone; a real frontier activates it and drops everything recorded so far.
func (b *Engine) resetPatchLog(heads [][32]byte) {
	b.diffCursor = append([][32]byte(nil), heads...)

	if len(heads) == 0 {
		return
	}

	b.patchLogActive = true
	b.patchLog = nil
	b.patchLogHeads = append([][32]byte(nil), heads...)
}

// patchTo records the patches that move the current view to after, matching
// AutoCommit::patch_to. Ops committed since the log was last caught up are
// frozen first, then the frontier change itself is appended.
func (b *Engine) patchTo(after [][32]byte) error {
	if !b.patchLogActive {
		return nil
	}

	if err := b.flushPatchLog(); err != nil {
		return err
	}

	if sameHeads(b.patchLogHeads, after) {
		return nil
	}

	segment, err := b.diffPatches(b.patchLogHeads, after, true)
	if err != nil {
		return err
	}

	b.patchLog = append(b.patchLog, segment...)
	b.patchLogHeads = append([][32]byte(nil), after...)

	return nil
}

// flushPatchLog appends the state difference between the log's frontier and
// the heads of the current view. That is the commit half of the patch log:
// Rust records those operations as they happen, and the next patch_to or
// incremental diff has to see them before the frontier moves.
func (b *Engine) flushPatchLog() error {
	if !b.patchLogActive {
		return nil
	}

	heads, err := b.Heads()
	if err != nil {
		return err
	}

	if sameHeads(b.patchLogHeads, heads) {
		return nil
	}

	segment, err := b.diffPatches(b.patchLogHeads, heads, false)
	if err != nil {
		return err
	}

	b.patchLog = append(b.patchLog, segment...)
	b.patchLogHeads = append([][32]byte(nil), heads...)

	return nil
}

func (b *Engine) fullDocumentHeads() [][32]byte {
	state := b.state
	if b.isolationActive && b.fullState != nil {
		state = b.fullState
	}

	return nativeToArrayHeads(state.Heads())
}

func sameHeads(left, right [][32]byte) bool {
	if len(left) != len(right) {
		return false
	}

	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}

	return true
}

func (b *Engine) diffPatches(
	beforeHeads [][32]byte,
	afterHeads [][32]byte,
	conflictAction bool,
) ([]patchOut, error) {
	// Diffs address the full history. While isolated, the pinned view omits
	// changes outside its frontier, but a cursor or historical head can still
	// name those changes.
	history := b.state
	if b.isolationActive && b.fullState != nil {
		history = b.fullState
	}

	source, ok := history.at(nativeHashes(beforeHeads))
	if !ok {
		return nil, fmt.Errorf("before heads are unknown")
	}

	target, ok := history.at(nativeHashes(afterHeads))
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
			objectPatches, err = diffObjectPatches(source, target, object, conflictAction)
		} else {
			objectPatches, err = materializeObjectPatches(target, object)
		}

		if err != nil {
			return nil, err
		}

		patches = append(patches, objectPatches...)
	}

	return exposeIntroducedObjects(patches, target)
}

// diffObjectPatches emits patches transforming an object from the source state
// into the target state, for an object present in both.
func diffObjectPatches(
	source, target *State,
	object opset.ObjectID,
	conflictAction bool,
) ([]patchOut, error) {
	objectType, err := objectTypeInState(target, object)
	if err != nil {
		return nil, err
	}

	identifier := objectIDString(object)

	switch objectType {
	case "map", "table":
		return diffMapPatches(source, target, object, identifier, conflictAction)
	case "list":
		return diffSequencePatches(source, target, object, objectType, identifier, conflictAction)
	case "text":
		patches, err := diffSequencePatches(source, target, object, objectType, identifier, conflictAction)
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

	marks = uncoveredMarkPatches(marks, patches)
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
func mapConflict(state *State, object opset.ObjectID, key string) bool {
	return len(state.visibleMapObjectOperations(object, key)) > 1
}

// uncoveredMarkPatches drops mark operations already carried by a splice that
// covers their range. A mark applied to text inserted in the same diff rides on
// that splice and is not repeated as a mark patch.
func uncoveredMarkPatches(marks []markPatchOut, patches []patchOut) []markPatchOut {
	uncovered := make([]markPatchOut, 0, len(marks))

	for _, mark := range marks {
		if markCoveredBySplice(mark, patches) {
			continue
		}

		uncovered = append(uncovered, mark)
	}

	return uncovered
}

func markCoveredBySplice(mark markPatchOut, patches []patchOut) bool {
	for _, patch := range patches {
		if patch.Action.Type != "splice_text" {
			continue
		}

		for _, candidate := range patch.Action.Marks {
			if candidate.Name != mark.Name || string(candidate.Value) != string(mark.Value) {
				continue
			}

			start := uint32(patch.Action.Index)
			end := start + uint32(utf16Width(patch.Action.Text))
			if mark.Start >= start && mark.End <= end {
				return true
			}
		}
	}

	return false
}

func diffMarkPatches(source, target *State, object opset.ObjectID) ([]markPatchOut, error) {
	begins := make([]opset.Operation, 0)

	target.eachOperation(func(operation opset.Operation) bool {
		if operation.Action != opset.ActionMark ||
			operation.Object != object ||
			operation.MarkName == nil {
			return true
		}

		if _, ok := source.operation(operation.ID); ok {
			return true
		}

		begins = append(begins, operation)

		return true
	})

	sort.Slice(
		begins,
		func(i, j int) bool {
			return begins[i].ID.Compare(begins[j].ID) < 0
		},
	)

	out := make([]markPatchOut, 0, len(begins))

	for _, begin := range begins {
		end, ok := target.operation(
			opset.OpID{Actor: begin.ID.Actor, Counter: begin.ID.Counter + 1},
		)
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
	conflictAction bool,
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
		case conflictAction && targetOK && sourceOK && targetOp.ID == sourceOp.ID &&
			mapConflict(source, object, key) != mapConflict(target, object, key):
			patches = append(
				patches,
				patchOut{
					Obj:    identifier,
					Action: patchActionOut{Type: "conflict", Key: key},
				},
			)
		case targetOK && (!sourceOK || targetOp.ID != sourceOp.ID ||
			mapConflict(source, object, key) != mapConflict(target, object, key)):
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
						Conflict: mapConflict(target, object, key),
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
	conflictAction bool,
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
			sourceConflict := len(source.visibleSequenceElementOperations(sourceValues[i].Element)) > 1
			targetConflict := len(target.visibleSequenceElementOperations(targetValues[j].Element)) > 1
			if sourceValues[i].Operation.ID == targetValues[j].Operation.ID &&
				sourceConflict == targetConflict {
				position += width(targetValues[j])
				i++
				j++

				continue
			}

			if sourceValues[i].Operation.ID == targetValues[j].Operation.ID {
				if conflictAction {
					patches = append(
						patches,
						patchOut{
							Obj: identifier,
							Action: patchActionOut{
								Type:  "conflict",
								Index: position,
							},
						},
					)
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
								Conflict: targetConflict,
							},
						},
					)
				}

				position += width(targetValues[j])
				i++
				j++

				continue
			}

			// Same element, different winning value. Text replacements become a
			// deletion of the old UTF-16 width followed by a splice. A non-string
			// winner is the object-replacement character. Lists use put_seq.
			replacement := targetValues[j].Operation
			if objectType == "text" && replacement.Action != opset.ActionMakeMap {
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

				rendered := textElementString(replacement)
				runs, err := textRunsWithMarks(target, object, position, rendered)
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

				position += uint64(utf16Width(rendered))
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

			operation := targetValues[j].Operation
			if operation.Action != opset.ActionMakeMap {
				rendered := textElementString(operation)
				runs, err := textRunsWithMarks(target, object, position, rendered)
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

				position += uint64(utf16Width(rendered))
				j++

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

// exposeIntroducedObjects materializes an object that a put or insert just
// made the visible value when that object has no patches of its own. Rust's
// patch log exposes the object's current contents at that moment, so a view
// that replaces the value can rebuild the text or children. A conflicted put
// keeps the existing value and must not be rewritten.
func exposeIntroducedObjects(patches []patchOut, target *State) ([]patchOut, error) {
	covered := make(map[string]bool, len(patches))
	for _, patch := range patches {
		if patch.Obj != "" {
			covered[patch.Obj] = true
		}
	}

	result := make([]patchOut, 0, len(patches))

	var expose func(patch patchOut) error

	expose = func(patch patchOut) error {
		result = append(result, patch)

		for _, identifier := range introducedObjectIDs(patch) {
			if covered[identifier] {
				continue
			}

			object, ok := parseObjectID(identifier)
			if !ok {
				continue
			}

			covered[identifier] = true

			children, err := materializeObjectPatches(target, object)
			if err != nil {
				return err
			}

			for _, child := range children {
				if err := expose(child); err != nil {
					return err
				}
			}
		}

		return nil
	}

	for _, patch := range patches {
		if err := expose(patch); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func introducedObjectIDs(patch patchOut) []string {
	switch patch.Action.Type {
	case "put_map", "put_seq":
		if patch.Action.Conflict || patch.Action.Value == nil || patch.Action.Value.ID == "" {
			return nil
		}

		return []string{patch.Action.Value.ID}
	case "insert":
		identifiers := make([]string, 0, len(patch.Action.Values))

		for _, value := range patch.Action.Values {
			if value.Conflict || value.Value.ID == "" {
				continue
			}

			identifiers = append(identifiers, value.Value.ID)
		}

		return identifiers
	default:
		return nil
	}
}

func parseObjectID(identifier string) (opset.ObjectID, bool) {
	if identifier == "" || identifier == "_root" {
		return opset.ObjectID{}, false
	}

	counterText, actorHex, ok := strings.Cut(identifier, "@")
	if !ok {
		return opset.ObjectID{}, false
	}

	counter, err := strconv.ParseUint(counterText, 10, 64)
	if err != nil {
		return opset.ObjectID{}, false
	}

	raw, err := hex.DecodeString(actorHex)
	if err != nil {
		return opset.ObjectID{}, false
	}

	return opset.ObjectID{
		OpID: opset.OpID{
			Counter: counter,
			Actor:   opset.ActorID(raw),
		},
	}, true
}
