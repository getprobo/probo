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

// The tests in this file reproduce upstream Rust orphan-change tests from
// automerge 0.10 (rust/automerge/tests/test_save_load_orphans.rs) against both
// the native Go engine and the Rust/WASM reference engine.

package automerge_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestRustOrphans_LoadIncrementalChangeWithoutDepsThrows reproduces
// load_incremental_change_without_deps_throws: a bare change chunk whose
// dependencies are absent cannot be loaded as a standalone document.
func TestRustOrphans_LoadIncrementalChangeWithoutDepsThrows(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc, err := engine.open(actor(1))
				require.NoError(t, err)
				closeDocument(t, doc)
				require.NoError(t, doc.PutString("key", "value"))
				_, err = doc.Commit("value", commitTime)
				require.NoError(t, err)
				_, err = doc.SaveIncremental()
				require.NoError(t, err)

				require.NoError(t, doc.PutString("key", "value2"))
				_, err = doc.Commit("value2", commitTime.Add(time.Second))
				require.NoError(t, err)
				orphan, err := doc.SaveIncremental()
				require.NoError(t, err)

				_, err = engine.load(orphan, actor(2))
				require.Error(t, err)
			},
		)
	}
}
