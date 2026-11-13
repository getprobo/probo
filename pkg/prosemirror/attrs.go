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
	"bytes"
	"encoding/json"
	"fmt"
)

// decodeAttrs unmarshals raw into a copy of fallback. An empty raw value
// returns fallback, which is how omitted Tiptap attributes take their defaults.
func decodeAttrs[T any](raw json.RawMessage, fallback T, what string) (T, error) {
	if len(raw) == 0 {
		return fallback, nil
	}

	if err := json.Unmarshal(raw, &fallback); err != nil {
		return fallback, fmt.Errorf("cannot parse %s attrs: %w", what, err)
	}

	return fallback, nil
}

// writeAttrs re-encodes v. It leaves dst untouched when the bytes match,
// so an unchanged node keeps its original JSON.
func writeAttrs(dst *json.RawMessage, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if bytes.Equal(*dst, raw) {
		return nil
	}

	*dst = raw

	return nil
}

// SetAttrs re-encodes v onto the node's attrs. Unchanged bytes keep the
// original JSON so sanitize stays byte-stable.
func (n *Node) SetAttrs(v any) error {
	return writeAttrs(&n.Attrs, v)
}
