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
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	internalencoding "go.probo.inc/probo/pkg/automerge/internal/encoding"
)

func (b *Engine) PutText(
	object uint32,
	key string,
) (uint32, error) {
	if err := b.requireRoot(object); err != nil {
		return 0, err
	}

	property := key

	operation := Operation{
		ID:     b.nextOperationID(),
		Object: RootObject(),
		Key:    Key{Property: &property},
		Action: ActionMakeText,
	}
	for _, predecessor := range b.state.visibleMapOperations(key) {
		operation.Predecessors = append(operation.Predecessors, predecessor.ID)
	}

	if err := b.addPending(operation); err != nil {
		return 0, err
	}

	return b.pushObject(ObjectID{OpID: operation.ID}), nil
}

func (b *Engine) GetText(
	object uint32,
	key string,
) (uint32, error) {
	if err := b.requireRoot(object); err != nil {
		return 0, err
	}

	operation, ok := b.state.visibleMapOperation(key, ActionMakeText)
	if !ok {
		return 0, fmt.Errorf("text property %q does not exist", key)
	}

	return b.pushObject(ObjectID{OpID: operation.ID}), nil
}

func (b *Engine) SpliceText(
	handle uint32,
	index uint32,
	deleteCount int32,
	value string,
) error {

	if deleteCount < 0 {
		return fmt.Errorf("negative text deletion is unsupported")
	}

	object, err := b.textObject(handle)
	if err != nil {
		return err
	}

	// Splice positions share the unified rich-text index space with marks and
	// blocks, so walk the full visible element sequence (text and block markers)
	// rather than the text-only view.
	sequence := b.state.sequenceElements(object.OpID)
	offsets := b.state.sequenceOffsets(object.OpID, sequence)

	start, end, previous, err := sequenceRange(sequence, offsets, index, uint32(deleteCount))
	if err != nil {
		return err
	}

	// The reference resolves the insertion anchor and inserts before deleting,
	// so replacement text is positioned against the pre-deletion sequence. That
	// ordering decides whether text replacing a marked run sits inside or
	// outside an expanding mark, so it must be preserved here.
	targets := make([]Operation, end-start)
	copy(targets, sequence[start:end])

	for offset, character := range []rune(value) {
		key := Key{IsHead: previous == nil}
		if previous != nil {
			key.Element = new(*previous)
		}

		// Only the first character resolves its anchor against neighbouring
		// mark boundaries; the rest chain onto the character before them.
		if offset == 0 {
			key = b.state.insertAnchorKey(object.OpID, key)
		}

		operation := Operation{
			ID:     b.nextOperationID(),
			Object: object,
			Key:    key,
			Insert: true,
			Action: ActionSet,
			Value:  &Scalar{Type: ScalarString, String: string(character)},
		}
		if err := b.addPending(operation); err != nil {
			return err
		}

		previous = new(operation.ID)
	}

	for _, target := range targets {
		operation := Operation{
			ID:           b.nextOperationID(),
			Object:       object,
			Key:          Key{Element: new(target.ID)},
			Action:       ActionDelete,
			Predecessors: []OpID{target.ID},
		}
		if err := b.addPending(operation); err != nil {
			return err
		}
	}

	return nil
}

type (
	updateSpanInput struct {
		Type  string                     `json:"type"`
		Text  string                     `json:"text"`
		Marks map[string]json.RawMessage `json:"marks"`
		Block json.RawMessage            `json:"block"`
	}

	updateSpansConfigInput struct {
		DefaultExpand  string            `json:"defaultExpand"`
		PerMarkExpands map[string]string `json:"perMarkExpands"`
	}

	desiredMark struct {
		name  string
		value Scalar
		start uint32
		end   uint32
	}
)

// UpdateSpans transforms the text object so its spans equal the supplied spans,
// mirroring the Rust AutoCommit::update_spans helper. The text content is
// reconciled with a minimal grapheme diff and the marks are then set to exactly
// the marks named on the spans, honoring the per-mark and default expand config.
// Block spans are not yet supported.
func (b *Engine) UpdateSpans(
	handle uint32,
	spans []byte,
	config []byte,
) error {

	object, err := b.textObject(handle)
	if err != nil {
		return err
	}

	var inputs []updateSpanInput
	if err := json.Unmarshal(spans, &inputs); err != nil {
		return fmt.Errorf("cannot decode update spans: %w", err)
	}

	var configuration updateSpansConfigInput
	if err := json.Unmarshal(config, &configuration); err != nil {
		return fmt.Errorf("cannot decode update spans config: %w", err)
	}

	target, err := targetBlockGraphemes(inputs)
	if err != nil {
		return err
	}

	current := b.currentBlockGraphemes(object)

	hook := &blockDiffHook{
		engine: b,
		handle: handle,
		old:    current,
		new:    target,
	}
	myersDiff(hook, blockTokens(current), blockTokens(target))

	if hook.err != nil {
		return hook.err
	}

	desired, err := desiredMarks(inputs)
	if err != nil {
		return err
	}

	return b.reconcileMarks(handle, object, desired, configuration)
}

// blockOrGrapheme is one unit of the block-aware span diff: either a block
// marker with its attributes or a single grapheme cluster of text.
type blockOrGrapheme struct {
	block    map[string]any
	isBlock  bool
	grapheme string
}

func (item blockOrGrapheme) width() int {
	if item.isBlock {
		return 1
	}

	return utf16Width(item.grapheme)
}

// blockTokens renders each unit to a comparison token so the grapheme-based
// Myers diff can compare blocks by their attributes and text by its clusters.
func blockTokens(items []blockOrGrapheme) []string {
	tokens := make([]string, len(items))

	for i, item := range items {
		if !item.isBlock {
			tokens[i] = "g" + item.grapheme

			continue
		}

		encoded, _ := json.Marshal(item.block)
		tokens[i] = "b" + string(encoded)
	}

	return tokens
}

// currentBlockGraphemes materializes the text object as the block/grapheme units
// the diff operates on: block markers become blocks and text runs are split into
// grapheme clusters.
func (b *Engine) currentBlockGraphemes(object ObjectID) []blockOrGrapheme {
	items := make([]blockOrGrapheme, 0)

	var run strings.Builder

	flush := func() {
		if run.Len() == 0 {
			return
		}

		for _, grapheme := range graphemeClusters(run.String()) {
			items = append(items, blockOrGrapheme{grapheme: grapheme})
		}

		run.Reset()
	}

	for _, value := range b.state.sequenceValues(object.OpID) {
		operation := value.Operation

		if operation.Action == ActionMakeMap {
			flush()

			attributes, err := b.state.mapValue(operation.ID, make(map[OpID]struct{}))
			if err != nil || attributes == nil {
				attributes = map[string]any{}
			}

			items = append(items, blockOrGrapheme{isBlock: true, block: attributes})

			continue
		}

		if operation.Value != nil && operation.Value.Type == ScalarString {
			run.WriteString(operation.Value.String)
		}
	}

	flush()

	return items
}

// targetBlockGraphemes converts the requested spans into block/grapheme units.
func targetBlockGraphemes(spans []updateSpanInput) ([]blockOrGrapheme, error) {
	items := make([]blockOrGrapheme, 0)

	for _, span := range spans {
		switch span.Type {
		case "text":
			for _, grapheme := range graphemeClusters(span.Text) {
				items = append(items, blockOrGrapheme{grapheme: grapheme})
			}
		case "block":
			attributes := map[string]any{}
			if len(span.Block) > 0 {
				if err := json.Unmarshal(span.Block, &attributes); err != nil {
					return nil, fmt.Errorf("cannot decode block span: %w", err)
				}
			}

			items = append(items, blockOrGrapheme{isBlock: true, block: attributes})
		default:
			return nil, fmt.Errorf("unsupported update span type %q", span.Type)
		}
	}

	return items, nil
}

// blockDiffHook applies the block-aware Myers edit script, splicing text and
// splitting, joining, or rewriting block markers as required.
type blockDiffHook struct {
	engine *Engine
	handle uint32
	old    []blockOrGrapheme
	new    []blockOrGrapheme
	idx    int
	err    error
}

func (h *blockDiffHook) failed() bool {
	return h.err != nil
}

func (h *blockDiffHook) equal(oldIndex, _ int, length int) {
	for i := range length {
		h.idx += h.old[oldIndex+i].width()
	}
}

func (h *blockDiffHook) delete(oldIndex, oldLen, _ int) {
	for i := 0; i < oldLen && h.err == nil; i++ {
		item := h.old[oldIndex+i]
		if item.isBlock {
			h.err = h.engine.JoinBlock(h.handle, uint32(h.idx))

			continue
		}

		h.err = h.engine.SpliceText(h.handle, uint32(h.idx), int32(item.width()), "")
	}
}

func (h *blockDiffHook) insert(_ int, newIndex, newLen int) {
	var run strings.Builder

	flush := func() {
		if run.Len() == 0 || h.err != nil {
			return
		}

		chars := run.String()
		if err := h.engine.SpliceText(h.handle, uint32(h.idx), 0, chars); err != nil {
			h.err = err

			return
		}

		h.idx += utf16Width(chars)

		run.Reset()
	}

	for i := 0; i < newLen && h.err == nil; i++ {
		item := h.new[newIndex+i]
		if !item.isBlock {
			run.WriteString(item.grapheme)

			continue
		}

		flush()

		if h.err != nil {
			return
		}

		blockHandle, err := h.engine.SplitBlock(h.handle, uint32(h.idx))
		if err != nil {
			h.err = err

			return
		}

		if err := h.engine.setBlockAttributes(blockHandle, item.block); err != nil {
			h.err = err

			return
		}

		h.idx++
	}

	flush()
}

// desiredMarks flattens the marks named on text spans into absolute UTF-16
// ranges in the order they appear.
func desiredMarks(spans []updateSpanInput) ([]desiredMark, error) {
	marks := make([]desiredMark, 0)
	index := uint32(0)

	for _, span := range spans {
		if span.Type == "block" {
			index++

			continue
		}

		width := uint32(utf16Width(span.Text))

		names := make([]string, 0, len(span.Marks))
		for name := range span.Marks {
			names = append(names, name)
		}

		sort.Strings(names)

		for _, name := range names {
			value, err := decodeScalarWire(span.Marks[name])
			if err != nil {
				return nil, fmt.Errorf("cannot decode mark %q value: %w", name, err)
			}

			marks = append(
				marks,
				desiredMark{
					name:  name,
					value: value,
					start: index,
					end:   index + width,
				},
			)
		}

		index += width
	}

	return marks, nil
}

// reconcileMarks removes marks that are not desired and adds the ones that are
// missing, matching the two-phase reconciliation upstream performs.
func (b *Engine) reconcileMarks(
	handle uint32,
	object ObjectID,
	desired []desiredMark,
	config updateSpansConfigInput,
) error {
	for _, current := range b.state.Marks(object.OpID) {
		keep := false

		for _, want := range desired {
			if want.name == current.Name &&
				want.start == current.Start &&
				want.end == current.End &&
				current.Value != nil &&
				scalarValuesEqual(want.value, *current.Value) {
				keep = true

				break
			}
		}

		if keep {
			continue
		}

		if err := b.markRange(
			handle,
			current.Start,
			current.End,
			current.Name,
			Scalar{Type: ScalarNull},
			config.expandFor(current.Name),
		); err != nil {
			return err
		}
	}

	for _, want := range desired {
		exists := false

		for _, current := range b.state.Marks(object.OpID) {
			if want.name == current.Name &&
				want.start == current.Start &&
				want.end == current.End &&
				current.Value != nil &&
				scalarValuesEqual(want.value, *current.Value) {
				exists = true

				break
			}
		}

		if exists {
			continue
		}

		if err := b.markRange(
			handle,
			want.start,
			want.end,
			want.name,
			want.value,
			config.expandFor(want.name),
		); err != nil {
			return err
		}
	}

	return nil
}

// setBlockAttributes writes the attribute map onto a freshly created block
// object, recursing into nested maps and lists.
func (b *Engine) setBlockAttributes(
	handle uint32,
	attributes map[string]any,
) error {
	keys := make([]string, 0, len(attributes))
	for key := range attributes {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		if err := b.setMapValue(handle, key, attributes[key]); err != nil {
			return err
		}
	}

	return nil
}

func (b *Engine) setMapValue(
	handle uint32,
	key string,
	value any,
) error {
	switch typed := value.(type) {
	case map[string]any:
		child, err := b.PutObject(handle, key, "map")
		if err != nil {
			return err
		}

		return b.setBlockAttributes(child, typed)
	case []any:
		child, err := b.PutObject(handle, key, "list")
		if err != nil {
			return err
		}

		for index, element := range typed {
			if err := b.insertListValue(child, uint64(index), element); err != nil {
				return err
			}
		}

		return nil
	default:
		scalar, err := hydrateScalar(value)
		if err != nil {
			return err
		}

		encoded, err := encodeScalarWire(scalar)
		if err != nil {
			return err
		}

		return b.PutScalar(handle, key, encoded)
	}
}

func (b *Engine) insertListValue(
	handle uint32,
	index uint64,
	value any,
) error {
	switch typed := value.(type) {
	case map[string]any:
		child, err := b.InsertObject(handle, index, "map")
		if err != nil {
			return err
		}

		return b.setBlockAttributes(child, typed)
	case []any:
		child, err := b.InsertObject(handle, index, "list")
		if err != nil {
			return err
		}

		for offset, element := range typed {
			if err := b.insertListValue(child, uint64(offset), element); err != nil {
				return err
			}
		}

		return nil
	default:
		scalar, err := hydrateScalar(value)
		if err != nil {
			return err
		}

		encoded, err := encodeScalarWire(scalar)
		if err != nil {
			return err
		}

		return b.InsertScalar(handle, index, encoded)
	}
}

// hydrateScalar maps a decoded JSON scalar to an Automerge scalar, treating
// integral numbers as integers to match the reference block hydration.
func hydrateScalar(value any) (Scalar, error) {
	switch typed := value.(type) {
	case nil:
		return Scalar{Type: ScalarNull}, nil
	case bool:
		if typed {
			return Scalar{Type: ScalarTrue, Bool: true}, nil
		}

		return Scalar{Type: ScalarFalse}, nil
	case string:
		return Scalar{Type: ScalarString, String: typed}, nil
	case float64:
		if typed == float64(int64(typed)) {
			return Scalar{Type: ScalarInt, Int: int64(typed)}, nil
		}

		return Scalar{Type: ScalarFloat64, Float: typed}, nil
	default:
		return Scalar{}, fmt.Errorf("unsupported block attribute value %T", value)
	}
}

func (b *Engine) markRange(
	handle uint32,
	start uint32,
	end uint32,
	name string,
	value Scalar,
	expand string,
) error {
	encoded, err := encodeScalarWire(value)
	if err != nil {
		return err
	}

	return b.MarkText(handle, start, end, name, encoded, expand)
}

func (c updateSpansConfigInput) expandFor(name string) string {
	if expand, ok := c.PerMarkExpands[name]; ok && expand != "" {
		return expand
	}

	if c.DefaultExpand != "" {
		return c.DefaultExpand
	}

	return "after"
}

func (b *Engine) MarkText(
	handle uint32,
	start uint32,
	end uint32,
	name string,
	encoded []byte,
	expand string,
) error {

	if start > end {
		return fmt.Errorf("mark range is inverted")
	}

	if start == end && expand == "none" {
		return nil
	}

	object, err := b.textObject(handle)
	if err != nil {
		return err
	}

	value, err := decodeScalarWire(encoded)
	if err != nil {
		return err
	}

	expandBefore, expandAfter, err := markExpansion(expand)
	if err != nil {
		return err
	}

	// Mark boundaries are inserted like any other element, so they resolve their
	// anchors through the same query. The end anchor is resolved after the begin
	// operation exists, matching the reference, because the begin can itself
	// offer a position that the end must take into account.
	startKey, err := b.textMarkKey(object, start)
	if err != nil {
		return err
	}

	begin := Operation{
		ID:         b.nextOperationID(),
		Object:     object,
		Key:        b.state.insertAnchorKey(object.OpID, startKey),
		Insert:     true,
		Action:     ActionMark,
		Value:      &value,
		MarkExpand: &expandBefore,
		MarkName:   &name,
	}
	if err := b.addPending(begin); err != nil {
		return err
	}

	endKey, err := b.textMarkKey(object, end)
	if err != nil {
		return err
	}

	endOperation := Operation{
		ID:         b.nextOperationID(),
		Object:     object,
		Key:        b.state.insertAnchorKey(object.OpID, endKey),
		Insert:     true,
		Action:     ActionMark,
		Value:      &Scalar{Type: ScalarNull},
		MarkExpand: &expandAfter,
	}

	return b.addPending(endOperation)
}

func (b *Engine) SplitBlock(
	handle uint32,
	index uint32,
) (uint32, error) {

	object, err := b.textObject(handle)
	if err != nil {
		return 0, err
	}

	sequence := b.state.sequenceElements(object.OpID)

	_, previous, err := richTextPosition(sequence, index)
	if err != nil {
		return 0, err
	}

	key := Key{IsHead: previous == nil}
	if previous != nil {
		key.Element = new(*previous)
	}

	// A block marker shares the unified rich-text sequence with text, so it must
	// resolve its anchor against neighbouring mark boundaries the same way a text
	// insertion does. Without this, a block inserted next to a mark boundary lands
	// on the wrong side of it, which then misplaces later insertions and makes the
	// marks they should or should not carry diverge from the reference.
	key = b.state.insertAnchorKey(object.OpID, key)

	operation := Operation{
		ID:     b.nextOperationID(),
		Object: object,
		Key:    key,
		Insert: true,
		Action: ActionMakeMap,
	}
	if err := b.addPending(operation); err != nil {
		return 0, err
	}

	return b.pushObject(ObjectID{OpID: operation.ID}), nil
}

func (b *Engine) JoinBlock(
	handle uint32,
	index uint32,
) error {

	object, err := b.textObject(handle)
	if err != nil {
		return err
	}

	sequence := b.state.sequenceElements(object.OpID)

	target, _, err := richTextPosition(sequence, index)
	if err != nil {
		return err
	}

	if target == nil || target.Action != ActionMakeMap {
		return fmt.Errorf("text position %d is not a block", index)
	}

	return b.addPending(
		Operation{
			ID:           b.nextOperationID(),
			Object:       object,
			Key:          Key{Element: new(target.ID)},
			Action:       ActionDelete,
			Predecessors: []OpID{target.ID},
		},
	)
}

func (b *Engine) ReplaceBlock(
	handle uint32,
	index uint32,
) (uint32, error) {
	if err := b.JoinBlock(handle, index); err != nil {
		return 0, err
	}

	return b.SplitBlock(handle, index)
}

func (b *Engine) Text(handle uint32) (string, error) {

	object, err := b.textObject(handle)
	if err != nil {
		return "", err
	}

	var output strings.Builder

	// Materialize the winning value of every visible element so that a put over
	// a text position replaces the original character, matching the reference.
	for _, value := range b.state.sequenceValues(object.OpID) {
		operation := value.Operation
		if operation.Value != nil && operation.Value.Type == ScalarString {
			output.WriteString(operation.Value.String)
		}
	}

	return output.String(), nil
}

func (b *Engine) TextAt(
	handle uint32,
	heads [][32]byte,
) (string, error) {

	object, err := b.textObject(handle)
	if err != nil {
		return "", err
	}

	historical, ok := b.state.at(nativeHashes(heads))
	if !ok {
		return "", fmt.Errorf("historical heads are unknown")
	}

	var output strings.Builder

	for _, operation := range historical.sequence(object.OpID) {
		if operation.Value != nil && operation.Value.Type == ScalarString {
			output.WriteString(operation.Value.String)
		}
	}

	return output.String(), nil
}

func (b *Engine) TextSpans(
	handle uint32,
) ([]byte, error) {

	object, err := b.textObject(handle)
	if err != nil {
		return nil, err
	}

	spans, err := b.state.RichTextSpans(object.OpID)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(spans)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native rich-text spans: %w", err)
	}

	return data, nil
}

func (b *Engine) TextSpansAt(
	handle uint32,
	heads [][32]byte,
) ([]byte, error) {

	object, err := b.textObject(handle)
	if err != nil {
		return nil, err
	}

	historical, ok := b.state.at(nativeHashes(heads))
	if !ok {
		return nil, fmt.Errorf("historical heads are unknown")
	}

	spans, err := historical.RichTextSpans(object.OpID)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(spans)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native historical rich-text spans: %w", err)
	}

	return data, nil
}

func encodeMarks(marks []MarkRange) ([]byte, error) {
	type markWire struct {
		Start uint32          `json:"start"`
		End   uint32          `json:"end"`
		Name  string          `json:"name"`
		Value json.RawMessage `json:"value"`
	}

	wire := make([]markWire, 0, len(marks))

	for _, mark := range marks {
		value := &Scalar{Type: ScalarNull}
		if mark.Value != nil {
			value = mark.Value
		}

		encoded, err := encodeScalarWire(*value)
		if err != nil {
			return nil, err
		}

		wire = append(
			wire,
			markWire{
				Start: mark.Start,
				End:   mark.End,
				Name:  mark.Name,
				Value: json.RawMessage(encoded),
			},
		)
	}

	data, err := json.Marshal(wire)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native marks: %w", err)
	}

	return data, nil
}

func (b *Engine) Marks(handle uint32) ([]byte, error) {

	object, err := b.textObject(handle)
	if err != nil {
		return nil, err
	}

	return encodeMarks(b.state.Marks(object.OpID))
}

func (b *Engine) MarksAt(
	handle uint32,
	heads [][32]byte,
) ([]byte, error) {

	object, err := b.textObject(handle)
	if err != nil {
		return nil, err
	}

	historical, ok := b.state.at(nativeHashes(heads))
	if !ok {
		return nil, fmt.Errorf("historical heads are unknown")
	}

	return encodeMarks(historical.Marks(object.OpID))
}

func (b *Engine) TextCursor(
	handle uint32,
	index uint32,
) ([]byte, error) {
	return b.TextCursorMoving(handle, index, false)
}

func (b *Engine) TextCursorMoving(
	handle uint32,
	index uint32,
	moveBefore bool,
) ([]byte, error) {

	object, err := b.textObject(handle)
	if err != nil {
		return nil, err
	}

	sequence := b.state.sequence(object.OpID)

	position := uint32(0)

	for _, operation := range sequence {
		length := uint32(utf16Length(operation))
		if index >= position && index < position+length {
			data := []byte{1, 3}
			data = internalencoding.AppendLengthPrefixed(data, operation.ID.Actor.Bytes())

			data = internalencoding.AppendULEB(data, operation.ID.Counter)
			if moveBefore {
				data = append(data, 1)
			} else {
				data = append(data, 2)
			}

			return data, nil
		}

		position += length
	}

	return nil, fmt.Errorf("text cursor index %d is out of bounds", index)
}

func (b *Engine) TextCursorMovingAt(
	handle uint32,
	index uint32,
	moveBefore bool,
	heads [][32]byte,
) ([]byte, error) {

	object, err := b.textObject(handle)
	if err != nil {
		return nil, err
	}

	historical, ok := b.state.at(nativeHashes(heads))
	if !ok {
		return nil, fmt.Errorf("historical heads are unknown")
	}

	position := uint32(0)

	for _, operation := range historical.sequence(object.OpID) {
		length := uint32(utf16Length(operation))
		if index >= position && index < position+length {
			data := []byte{1, 3}
			data = internalencoding.AppendLengthPrefixed(data, operation.ID.Actor.Bytes())
			data = internalencoding.AppendULEB(data, operation.ID.Counter)

			if moveBefore {
				data = append(data, 1)
			} else {
				data = append(data, 2)
			}

			return data, nil
		}

		position += length
	}

	return nil, fmt.Errorf("text cursor index %d is out of bounds", index)
}

func (b *Engine) TextCursorPosition(
	handle uint32,
	cursor []byte,
) (uint32, error) {

	object, err := b.textObject(handle)
	if err != nil {
		return 0, err
	}

	if bytes.Equal(cursor, []byte{1, 1}) {
		return 0, nil
	}

	if bytes.Equal(cursor, []byte{1, 2}) {
		var length uint32
		for _, operation := range b.state.sequence(object.OpID) {
			length += uint32(utf16Length(operation))
		}

		return length, nil
	}

	target, move, err := decodeCursor(cursor)
	if err != nil {
		return 0, err
	}

	position := uint32(0)

	for _, operation := range b.state.sequenceAll(object.OpID) {
		if operation.ID == target {
			if b.state.isSuperseded(operation.ID) && move == 1 {
				return b.cursorMoveBeforePosition(object.OpID, operation)
			}

			return position, nil
		}

		if !b.state.isSuperseded(operation.ID) {
			position += uint32(utf16Length(operation))
		}
	}

	return 0, fmt.Errorf("text cursor target does not exist")
}

func (b *Engine) cursorMoveBeforePosition(
	object OpID,
	target Operation,
) (uint32, error) {
	visited := make(map[OpID]struct{})

	for {
		if target.Key.IsHead {
			return 0, nil
		}

		if target.Key.Element == nil {
			return 0, fmt.Errorf("text cursor target has no predecessor")
		}

		if _, ok := visited[*target.Key.Element]; ok {
			return 0, fmt.Errorf("text cursor predecessor cycle")
		}

		visited[*target.Key.Element] = struct{}{}

		var position uint32

		for _, operation := range b.state.sequence(object) {
			if operation.ID == *target.Key.Element {
				return position, nil
			}

			position += uint32(utf16Length(operation))
		}

		predecessor, ok := b.state.operations[*target.Key.Element]
		if !ok {
			return 0, fmt.Errorf("text cursor predecessor does not exist")
		}

		target = predecessor
	}
}
