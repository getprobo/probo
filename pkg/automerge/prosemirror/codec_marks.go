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

package prosemirror

import (
	"encoding/json"

	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/prosemirror"
)

// boolMarkCodec handles a formatting mark whose presence is its only payload:
// the Automerge value is always the boolean true and the ProseMirror mark
// carries no attributes.
type boolMarkCodec struct {
	mark      prosemirror.MarkType
	automerge string
}

func (c boolMarkCodec) Mark() prosemirror.MarkType { return c.mark }

func (c boolMarkCodec) Automerge() string { return c.automerge }

func (c boolMarkCodec) ToScalar(pmMark) (automerge.Scalar, bool) {
	return automerge.BoolScalar(true), true
}

func (c boolMarkCodec) FromScalar(any) (mark, bool) {
	return mark{Type: string(c.mark)}, true
}

// linkCodec encodes a link as a JSON object string carrying its href and
// optional title, matching the frontend adapter.
type linkCodec struct{}

func (linkCodec) Mark() prosemirror.MarkType { return prosemirror.MarkLink }

func (linkCodec) Automerge() string { return "link" }

func (linkCodec) ToScalar(m pmMark) (automerge.Scalar, bool) {
	payload, err := json.Marshal(
		map[string]any{
			"href":  stringAttribute(m.Attrs, "href"),
			"title": nullableStringAttribute(m.Attrs, "title"),
		},
	)
	if err != nil {
		return automerge.Scalar{}, false
	}

	return automerge.StringScalar(string(payload)), true
}

func (c linkCodec) FromScalar(value any) (mark, bool) {
	raw, ok := value.(string)
	if !ok {
		return mark{}, false
	}

	var attrs map[string]any
	if err := json.Unmarshal([]byte(raw), &attrs); err != nil {
		return mark{}, false
	}

	return mark{Type: string(c.Mark()), Attrs: attrs}, true
}
