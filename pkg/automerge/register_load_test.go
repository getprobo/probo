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

package automerge_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	productionautomerge "go.probo.inc/probo/pkg/automerge"
)

// registerLoadEntries sizes the fixture past a quarter of a million
// operations, the scale of a generated register for a large organization.
const registerLoadEntries = 130

func TestLoad_RegisterSizedDocument(t *testing.T) {
	t.Parallel()

	fixture := benchmarkRegister(t, registerLoadEntries)

	document, err := productionautomerge.Load(fixture.data, registerActor(4))
	require.NoError(t, err)

	defer func() { _ = document.Close() }()

	stats, err := document.Stats()
	require.NoError(t, err)
	require.Greater(t, stats.NumOps, uint64(250_000))

	text, err := document.Text("body")
	require.NoError(t, err)
	spans, err := text.Spans()
	require.NoError(t, err)

	seeded := seedRegister(t, fixture.content)

	defer func() { _ = seeded.Close() }()

	seededText, err := seeded.Text("body")
	require.NoError(t, err)
	want, err := seededText.Spans()
	require.NoError(t, err)

	assert.Equal(t, want, spans)

	saved, err := document.Save()
	require.NoError(t, err)
	assert.Equal(t, fixture.data, saved)
}
