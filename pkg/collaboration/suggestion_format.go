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

package collaboration

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"

	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/prosemirror"
)

const (
	formatActionAdd    = "add"
	formatActionRemove = "remove"
)

// FormatSuggestionMark names a style the editor can suggest.
type FormatSuggestionMark string

const (
	FormatSuggestionBold      FormatSuggestionMark = "BOLD"
	FormatSuggestionItalic    FormatSuggestionMark = "ITALIC"
	FormatSuggestionUnderline FormatSuggestionMark = "UNDERLINE"
	FormatSuggestionStrike    FormatSuggestionMark = "STRIKE"
	FormatSuggestionCode      FormatSuggestionMark = "CODE"
	FormatSuggestionLink      FormatSuggestionMark = "LINK"
)

// FormatSuggestionAction adds a style or takes one away.
type FormatSuggestionAction string

const (
	FormatSuggestionAdd    FormatSuggestionAction = "ADD"
	FormatSuggestionRemove FormatSuggestionAction = "REMOVE"
)

// FormatSuggestion paints one style onto quoted text. Href is the link
// address when Mark is LINK and Action is ADD.
type FormatSuggestion struct {
	Mark       FormatSuggestionMark
	Action     FormatSuggestionAction
	Quote      string
	Occurrence int
	Href       string
}

var (
	ErrSuggestionFormat = errors.New("unknown format suggestion")
	ErrSuggestionHref   = errors.New("link address is not allowed")
)

type formatStyle struct {
	mark       FormatSuggestionMark
	automerge  string
	suggestion string
	link       bool
}

var formatStyles = []formatStyle{
	{mark: FormatSuggestionBold, automerge: "strong", suggestion: "suggestion-format-strong"},
	{mark: FormatSuggestionItalic, automerge: "em", suggestion: "suggestion-format-em"},
	{mark: FormatSuggestionUnderline, automerge: "underline", suggestion: "suggestion-format-underline"},
	{mark: FormatSuggestionStrike, automerge: "strike", suggestion: "suggestion-format-strike"},
	{mark: FormatSuggestionCode, automerge: "code", suggestion: "suggestion-format-code"},
	{mark: FormatSuggestionLink, automerge: "link", suggestion: "suggestion-format-link", link: true},
}

type formatMarkPayload struct {
	ID     string `json:"id"`
	Action string `json:"action"`
	Href   string `json:"href,omitempty"`
}

type linkMarkPayload struct {
	Href  string  `json:"href"`
	Title *string `json:"title"`
}

type formatTarget struct {
	start uint32
	end   uint32
	marks map[string]any
}

type formatOp struct {
	name   string
	unmark bool
	value  automerge.Scalar
}

// ApplySuggestionFormat marks quote with a suggested style. Accepting keeps
// the style and rejecting restores the text. Text this suggestion inserted
// takes the style directly, because rejecting the insertion removes it.
func ApplySuggestionFormat(
	document *automerge.Document,
	suggestionID string,
	change FormatSuggestion,
) error {
	if suggestionID == "" || change.Quote == "" || change.Occurrence < 0 {
		return ErrQuoteNotFound
	}

	spec, ok := formatStyleFor(change.Mark)
	if !ok || (change.Action != FormatSuggestionAdd && change.Action != FormatSuggestionRemove) {
		return ErrSuggestionFormat
	}

	href := ""

	if spec.link && change.Action == FormatSuggestionAdd {
		safe, err := suggestionLinkHref(change.Href)
		if err != nil {
			return err
		}

		href = safe
	}

	spans, err := BodySpans(document)
	if err != nil {
		return err
	}

	visible, _, _ := suggestionUnits(spans)

	start, end, err := quoteRange(visible, change.Quote, change.Occurrence)
	if err != nil {
		return err
	}

	targets := formatTargets(spans, start, end)
	plans := make([]struct {
		start uint32
		end   uint32
		ops   []formatOp
	}, 0)
	changed := false

	for _, target := range targets {
		ops, targetChanged, err := planFormat(target.marks, spec, change.Action, suggestionID, href)
		if err != nil {
			return err
		}

		if !targetChanged {
			continue
		}

		changed = true

		plans = append(plans, struct {
			start uint32
			end   uint32
			ops   []formatOp
		}{start: target.start, end: target.end, ops: ops})
	}

	if !changed {
		return ErrSuggestionUnchanged
	}

	text, err := document.Text("body")
	if err != nil {
		return fmt.Errorf("cannot get collaboration body: %w", err)
	}

	for _, plan := range plans {
		for _, op := range plan.ops {
			if op.unmark {
				if err := text.Unmark(plan.start, plan.end, op.name, automerge.MarkExpandNone); err != nil {
					return fmt.Errorf("cannot clear suggested formatting: %w", err)
				}

				continue
			}

			if err := text.Mark(plan.start, plan.end, op.name, op.value, automerge.MarkExpandNone); err != nil {
				return fmt.Errorf("cannot mark suggested formatting: %w", err)
			}
		}
	}

	return nil
}

func planFormat(
	marks map[string]any,
	spec formatStyle,
	action FormatSuggestionAction,
	suggestionID string,
	href string,
) ([]formatOp, bool, error) {
	if markID(marks, SuggestionDeleteMark) != "" {
		return nil, false, ErrSuggestionOverlaps
	}

	insertID := markID(marks, SuggestionInsertMark)
	if insertID != "" && insertID != suggestionID {
		return nil, false, ErrSuggestionOverlaps
	}

	if _, ok := readFormat(marks, spec.suggestion); ok {
		return nil, false, ErrSuggestionOverlaps
	}

	ownInsert := insertID == suggestionID

	if action == FormatSuggestionAdd {
		return planFormatAdd(marks, spec, suggestionID, href, ownInsert)
	}

	if !markPresent(marks, spec.automerge) {
		return nil, false, nil
	}

	ops := []formatOp{{name: spec.automerge, unmark: true}}

	if !ownInsert {
		restore := ""
		if spec.link {
			restore = linkHrefOf(marks)
		}

		mark, err := formatMarkOp(spec.suggestion, suggestionID, formatActionRemove, restore)
		if err != nil {
			return nil, false, err
		}

		ops = append(ops, mark)
	}

	return ops, true, nil
}

func planFormatAdd(
	marks map[string]any,
	spec formatStyle,
	suggestionID string,
	href string,
	ownInsert bool,
) ([]formatOp, bool, error) {
	if spec.link {
		if linkHrefOf(marks) == href {
			return nil, false, nil
		}
	} else if markPresent(marks, spec.automerge) {
		return nil, false, nil
	}

	ops := make([]formatOp, 0, 4)

	if spec.mark == FormatSuggestionCode {
		for _, other := range formatStyles {
			if other.mark == FormatSuggestionCode || !markPresent(marks, other.automerge) {
				continue
			}

			if _, ok := readFormat(marks, other.suggestion); ok {
				return nil, false, ErrSuggestionOverlaps
			}

			ops = append(ops, formatOp{name: other.automerge, unmark: true})

			if ownInsert {
				continue
			}

			restore := ""
			if other.link {
				restore = linkHrefOf(marks)
			}

			mark, err := formatMarkOp(other.suggestion, suggestionID, formatActionRemove, restore)
			if err != nil {
				return nil, false, err
			}

			ops = append(ops, mark)
		}
	}

	if spec.link {
		value, err := linkScalar(href)
		if err != nil {
			return nil, false, err
		}

		ops = append(ops, formatOp{name: spec.automerge, value: value})
	} else {
		ops = append(ops, formatOp{name: spec.automerge, value: automerge.BoolScalar(true)})
	}

	if !ownInsert {
		restore := ""
		if spec.link {
			restore = linkHrefOf(marks)
		}

		mark, err := formatMarkOp(spec.suggestion, suggestionID, formatActionAdd, restore)
		if err != nil {
			return nil, false, err
		}

		ops = append(ops, mark)
	}

	return ops, true, nil
}

func formatTargets(spans []automerge.Span, from uint32, to uint32) []formatTarget {
	targets := make([]formatTarget, 0)
	engine := uint32(0)
	visible := uint32(0)
	wrote := false

	for _, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			if wrote {
				visible++
			}

			engine++

			continue
		}

		if span.Type != automerge.SpanTypeText || span.Text == "" {
			continue
		}

		width := utf16Len(span.Text)
		overlapFrom := max(visible, from)
		overlapTo := min(visible+width, to)

		if overlapFrom < overlapTo {
			delta := overlapFrom - visible
			targets = append(targets, formatTarget{
				start: engine + delta,
				end:   engine + (overlapTo - visible),
				marks: span.Marks,
			})
		}

		visible += width
		engine += width
		wrote = true
	}

	return targets
}

func formatStyleFor(mark FormatSuggestionMark) (formatStyle, bool) {
	for _, spec := range formatStyles {
		if spec.mark == mark {
			return spec, true
		}
	}

	return formatStyle{}, false
}

func suggestionLinkHref(href string) (string, error) {
	trimmed := strings.TrimSpace(href)
	if trimmed == "" {
		return "", ErrSuggestionHref
	}

	safe := prosemirror.SafeLinkHref(trimmed)
	if safe == "#" && trimmed != "#" {
		return "", ErrSuggestionHref
	}

	return safe, nil
}

func formatMarkOp(name string, suggestionID string, action string, href string) (formatOp, error) {
	value, err := formatScalar(suggestionID, action, href)
	if err != nil {
		return formatOp{}, err
	}

	return formatOp{name: name, value: value}, nil
}

func formatScalar(suggestionID string, action string, href string) (automerge.Scalar, error) {
	encoded, err := json.Marshal(formatMarkPayload{
		ID:     suggestionID,
		Action: action,
		Href:   href,
	})
	if err != nil {
		return automerge.Scalar{}, fmt.Errorf("cannot encode format suggestion: %w", err)
	}

	return automerge.StringScalar(string(encoded)), nil
}

func linkScalar(href string) (automerge.Scalar, error) {
	encoded, err := json.Marshal(linkMarkPayload{Href: href})
	if err != nil {
		return automerge.Scalar{}, fmt.Errorf("cannot encode suggested link: %w", err)
	}

	return automerge.StringScalar(string(encoded)), nil
}

func linkJSON(href string) string {
	encoded, err := json.Marshal(linkMarkPayload{Href: href})
	if err != nil {
		return ""
	}

	return string(encoded)
}

func readFormat(marks map[string]any, name string) (formatMarkPayload, bool) {
	raw, ok := marks[name].(string)
	if !ok || raw == "" {
		return formatMarkPayload{}, false
	}

	var payload formatMarkPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil || payload.ID == "" {
		return formatMarkPayload{}, false
	}

	if payload.Action != formatActionAdd && payload.Action != formatActionRemove {
		return formatMarkPayload{}, false
	}

	return payload, true
}

func markPresent(marks map[string]any, name string) bool {
	value, ok := marks[name]
	if !ok || value == nil || value == false || value == "" {
		return false
	}

	return true
}

func linkHrefOf(marks map[string]any) string {
	raw, ok := marks["link"].(string)
	if !ok || raw == "" {
		return ""
	}

	var payload linkMarkPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return ""
	}

	return payload.Href
}

func withMarkValue(marks map[string]any, name string, value any) map[string]any {
	next := make(map[string]any, len(marks)+1)
	maps.Copy(next, marks)
	next[name] = value

	return next
}

// decideFormatMarks keeps a suggested style when accept is set and restores
// the previous style otherwise.
func decideFormatMarks(marks map[string]any, suggestionID string, accept bool) (map[string]any, bool) {
	changed := false

	for _, spec := range formatStyles {
		payload, ok := readFormat(marks, spec.suggestion)
		if !ok || payload.ID != suggestionID {
			continue
		}

		changed = true
		marks = withoutMark(marks, spec.suggestion)

		if accept {
			continue
		}

		marks = restoreFormat(marks, spec, payload)
	}

	return marks, changed
}

// revertFormatMarks drops every open format suggestion, which is the document
// readers see and the baseline a comment-only edit must preserve.
func revertFormatMarks(marks map[string]any) map[string]any {
	for _, spec := range formatStyles {
		payload, ok := readFormat(marks, spec.suggestion)
		if !ok {
			continue
		}

		marks = withoutMark(marks, spec.suggestion)
		marks = restoreFormat(marks, spec, payload)
	}

	return marks
}

func restoreFormat(marks map[string]any, spec formatStyle, payload formatMarkPayload) map[string]any {
	if payload.Action == formatActionAdd {
		if spec.link && payload.Href != "" {
			return withMarkValue(marks, spec.automerge, linkJSON(payload.Href))
		}

		return withoutMark(marks, spec.automerge)
	}

	if spec.link {
		return withMarkValue(marks, spec.automerge, linkJSON(payload.Href))
	}

	return withMarkValue(marks, spec.automerge, true)
}

func formatMarkMatches(marks map[string]any, suggestionID string) bool {
	for _, spec := range formatStyles {
		payload, ok := readFormat(marks, spec.suggestion)
		if ok && payload.ID == suggestionID {
			return true
		}
	}

	return false
}

func formatMarkIDs(marks map[string]any) []string {
	ids := make([]string, 0)

	for _, spec := range formatStyles {
		payload, ok := readFormat(marks, spec.suggestion)
		if !ok {
			continue
		}

		ids = append(ids, payload.ID)
	}

	return ids
}

func registerFormatExpand(expands map[string]automerge.MarkExpand) {
	for _, spec := range formatStyles {
		expands[spec.suggestion] = automerge.MarkExpandNone
	}
}
