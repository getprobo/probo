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

package coredata

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"
)

type (
	DocumentSuggestionChangeKind string
	DocumentSuggestionBlockKind  string
	DocumentSuggestionNodeKind   string
	DocumentSuggestionNodeAction string

	// DocumentSuggestionChanges is what a suggestion changed, read just before
	// a reviewer's decision removed its marks. Changes holds the first few;
	// ChangeCount counts them all.
	DocumentSuggestionChanges struct {
		Changes     []DocumentSuggestionChange `json:"changes"`
		ChangeCount int                        `json:"changeCount"`
	}

	// DocumentSuggestionChange is one change. Kind tells which fields are set:
	// TEXT sets Deleted and Inserted, FORMAT sets Mark and FormatAction (and
	// the link addresses for a link), BLOCK sets Before and After, and NODE
	// sets Node and NodeAction.
	DocumentSuggestionChange struct {
		Kind         DocumentSuggestionChangeKind  `json:"kind"`
		Deleted      *string                       `json:"deleted,omitempty"`
		Inserted     *string                       `json:"inserted,omitempty"`
		Mark         string                        `json:"mark,omitempty"`
		FormatAction string                        `json:"formatAction,omitempty"`
		Href         *string                       `json:"href,omitempty"`
		PreviousHref *string                       `json:"previousHref,omitempty"`
		Before       *DocumentSuggestionBlockLabel `json:"before,omitempty"`
		After        *DocumentSuggestionBlockLabel `json:"after,omitempty"`
		Node         DocumentSuggestionNodeKind    `json:"node,omitempty"`
		NodeAction   DocumentSuggestionNodeAction  `json:"nodeAction,omitempty"`
	}

	DocumentSuggestionBlockLabel struct {
		Kind  DocumentSuggestionBlockKind `json:"kind"`
		Level *int                        `json:"level,omitempty"`
	}
)

const (
	DocumentSuggestionChangeKindText   DocumentSuggestionChangeKind = "TEXT"
	DocumentSuggestionChangeKindFormat DocumentSuggestionChangeKind = "FORMAT"
	DocumentSuggestionChangeKindBlock  DocumentSuggestionChangeKind = "BLOCK"
	DocumentSuggestionChangeKindNode   DocumentSuggestionChangeKind = "NODE"
)

const (
	DocumentSuggestionBlockKindParagraph   DocumentSuggestionBlockKind = "PARAGRAPH"
	DocumentSuggestionBlockKindHeading     DocumentSuggestionBlockKind = "HEADING"
	DocumentSuggestionBlockKindQuote       DocumentSuggestionBlockKind = "QUOTE"
	DocumentSuggestionBlockKindCode        DocumentSuggestionBlockKind = "CODE"
	DocumentSuggestionBlockKindBulletList  DocumentSuggestionBlockKind = "BULLET_LIST"
	DocumentSuggestionBlockKindOrderedList DocumentSuggestionBlockKind = "ORDERED_LIST"
	DocumentSuggestionBlockKindTaskList    DocumentSuggestionBlockKind = "TASK_LIST"
	DocumentSuggestionBlockKindOther       DocumentSuggestionBlockKind = "OTHER"
)

const (
	DocumentSuggestionNodeKindTable   DocumentSuggestionNodeKind = "TABLE"
	DocumentSuggestionNodeKindRow     DocumentSuggestionNodeKind = "ROW"
	DocumentSuggestionNodeKindColumn  DocumentSuggestionNodeKind = "COLUMN"
	DocumentSuggestionNodeKindImage   DocumentSuggestionNodeKind = "IMAGE"
	DocumentSuggestionNodeKindDivider DocumentSuggestionNodeKind = "DIVIDER"
	DocumentSuggestionNodeKindFile    DocumentSuggestionNodeKind = "FILE"
	DocumentSuggestionNodeKindLine    DocumentSuggestionNodeKind = "LINE"
	DocumentSuggestionNodeKindOther   DocumentSuggestionNodeKind = "OTHER"
)

const (
	DocumentSuggestionNodeActionInsert DocumentSuggestionNodeAction = "INSERT"
	DocumentSuggestionNodeActionDelete DocumentSuggestionNodeAction = "DELETE"
)

var (
	_ driver.Valuer = DocumentSuggestionChanges{}
	_ sql.Scanner   = (*DocumentSuggestionChanges)(nil)

	_ fmt.Stringer             = DocumentSuggestionChangeKind("")
	_ encoding.TextMarshaler   = DocumentSuggestionChangeKind("")
	_ encoding.TextUnmarshaler = (*DocumentSuggestionChangeKind)(nil)

	_ fmt.Stringer             = DocumentSuggestionBlockKind("")
	_ encoding.TextMarshaler   = DocumentSuggestionBlockKind("")
	_ encoding.TextUnmarshaler = (*DocumentSuggestionBlockKind)(nil)

	_ fmt.Stringer             = DocumentSuggestionNodeKind("")
	_ encoding.TextMarshaler   = DocumentSuggestionNodeKind("")
	_ encoding.TextUnmarshaler = (*DocumentSuggestionNodeKind)(nil)

	_ fmt.Stringer             = DocumentSuggestionNodeAction("")
	_ encoding.TextMarshaler   = DocumentSuggestionNodeAction("")
	_ encoding.TextUnmarshaler = (*DocumentSuggestionNodeAction)(nil)
)

func (c DocumentSuggestionChanges) Value() (driver.Value, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal document suggestion changes: %w", err)
	}

	return data, nil
}

func (c *DocumentSuggestionChanges) Scan(value any) error {
	*c = DocumentSuggestionChanges{}

	if value == nil {
		return nil
	}

	var data []byte

	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("cannot scan document suggestion changes: unsupported type %T", value)
	}

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("cannot unmarshal document suggestion changes: %w", err)
	}

	return nil
}

func DocumentSuggestionChangeKinds() []DocumentSuggestionChangeKind {
	return []DocumentSuggestionChangeKind{
		DocumentSuggestionChangeKindText,
		DocumentSuggestionChangeKindFormat,
		DocumentSuggestionChangeKindBlock,
		DocumentSuggestionChangeKindNode,
	}
}

func (v DocumentSuggestionChangeKind) IsValid() bool {
	switch v {
	case
		DocumentSuggestionChangeKindText,
		DocumentSuggestionChangeKindFormat,
		DocumentSuggestionChangeKindBlock,
		DocumentSuggestionChangeKindNode:
		return true
	}

	return false
}

func (v DocumentSuggestionChangeKind) String() string {
	return string(v)
}

func (v DocumentSuggestionChangeKind) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentSuggestionChangeKind) UnmarshalText(text []byte) error {
	val := DocumentSuggestionChangeKind(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DocumentSuggestionChangeKind value: %q", string(text))
	}

	*v = val

	return nil
}

func DocumentSuggestionBlockKinds() []DocumentSuggestionBlockKind {
	return []DocumentSuggestionBlockKind{
		DocumentSuggestionBlockKindParagraph,
		DocumentSuggestionBlockKindHeading,
		DocumentSuggestionBlockKindQuote,
		DocumentSuggestionBlockKindCode,
		DocumentSuggestionBlockKindBulletList,
		DocumentSuggestionBlockKindOrderedList,
		DocumentSuggestionBlockKindTaskList,
		DocumentSuggestionBlockKindOther,
	}
}

func (v DocumentSuggestionBlockKind) IsValid() bool {
	switch v {
	case
		DocumentSuggestionBlockKindParagraph,
		DocumentSuggestionBlockKindHeading,
		DocumentSuggestionBlockKindQuote,
		DocumentSuggestionBlockKindCode,
		DocumentSuggestionBlockKindBulletList,
		DocumentSuggestionBlockKindOrderedList,
		DocumentSuggestionBlockKindTaskList,
		DocumentSuggestionBlockKindOther:
		return true
	}

	return false
}

func (v DocumentSuggestionBlockKind) String() string {
	return string(v)
}

func (v DocumentSuggestionBlockKind) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentSuggestionBlockKind) UnmarshalText(text []byte) error {
	val := DocumentSuggestionBlockKind(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DocumentSuggestionBlockKind value: %q", string(text))
	}

	*v = val

	return nil
}

func DocumentSuggestionNodeKinds() []DocumentSuggestionNodeKind {
	return []DocumentSuggestionNodeKind{
		DocumentSuggestionNodeKindTable,
		DocumentSuggestionNodeKindRow,
		DocumentSuggestionNodeKindColumn,
		DocumentSuggestionNodeKindImage,
		DocumentSuggestionNodeKindDivider,
		DocumentSuggestionNodeKindFile,
		DocumentSuggestionNodeKindLine,
		DocumentSuggestionNodeKindOther,
	}
}

func (v DocumentSuggestionNodeKind) IsValid() bool {
	switch v {
	case
		DocumentSuggestionNodeKindTable,
		DocumentSuggestionNodeKindRow,
		DocumentSuggestionNodeKindColumn,
		DocumentSuggestionNodeKindImage,
		DocumentSuggestionNodeKindDivider,
		DocumentSuggestionNodeKindFile,
		DocumentSuggestionNodeKindLine,
		DocumentSuggestionNodeKindOther:
		return true
	}

	return false
}

func (v DocumentSuggestionNodeKind) String() string {
	return string(v)
}

func (v DocumentSuggestionNodeKind) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentSuggestionNodeKind) UnmarshalText(text []byte) error {
	val := DocumentSuggestionNodeKind(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DocumentSuggestionNodeKind value: %q", string(text))
	}

	*v = val

	return nil
}

func DocumentSuggestionNodeActions() []DocumentSuggestionNodeAction {
	return []DocumentSuggestionNodeAction{
		DocumentSuggestionNodeActionInsert,
		DocumentSuggestionNodeActionDelete,
	}
}

func (v DocumentSuggestionNodeAction) IsValid() bool {
	switch v {
	case
		DocumentSuggestionNodeActionInsert,
		DocumentSuggestionNodeActionDelete:
		return true
	}

	return false
}

func (v DocumentSuggestionNodeAction) String() string {
	return string(v)
}

func (v DocumentSuggestionNodeAction) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentSuggestionNodeAction) UnmarshalText(text []byte) error {
	val := DocumentSuggestionNodeAction(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DocumentSuggestionNodeAction value: %q", string(text))
	}

	*v = val

	return nil
}
