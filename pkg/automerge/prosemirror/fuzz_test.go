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

package prosemirror_test

import (
	"encoding/json"
	"testing"

	"go.probo.inc/probo/pkg/automerge"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
)

func FuzzRender(f *testing.F) {
	f.Add([]byte(`[
		{"Type":"block","Block":{"type":"paragraph","parents":[]}},
		{"Type":"text","Text":"Hello"}
	]`))
	f.Add([]byte(`[
		{"Type":"block","Block":{"type":"horizontal-rule","parents":[]}},
		{"Type":"text","Text":"Preserved"}
	]`))
	f.Add([]byte(`[
		{"Type":"block","Block":{"type":"table","parents":[]}},
		{"Type":"block","Block":{"type":"table-row","parents":["table"]}},
		{"Type":"block","Block":{"type":"table-cell","parents":["table","table-row"]}},
		{"Type":"text","Text":"Cell"}
	]`))
	f.Add([]byte(`[
		{"Type":"block","Block":{"type":"callout","parents":[]}},
		{"Type":"text","Text":"Unknown block"},
		{"Type":"block","Block":{"type":"paragraph","parents":["callout"]}},
		{"Type":"text","Text":"Hoisted child"}
	]`))
	f.Add([]byte(`[
		{"Type":"block","Block":{"type":"paragraph","parents":[]}},
		{"Type":"text","Text":"Marked","Marks":{"highlight":true,"strong":true}}
	]`))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1024*1024 {
			t.Skip()
		}

		var spans []automerge.Span
		if err := json.Unmarshal(data, &spans); err != nil {
			return
		}

		if len(spans) > 100_000 {
			t.Skip()
		}

		if _, err := automergeprosemirror.Render(spans); err != nil {
			t.Fatalf("Render returned an error for JSON-decoded spans: %v", err)
		}
	})
}
