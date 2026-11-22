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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

// remoteTextRun describes a long text typed in one go, then edited
// concurrently by a local and a remote writer before the local writer applies
// the remote change.
type remoteTextRun struct {
	baseSize    int
	position    uint32
	localValue  string
	remoteValue string
}

// remoteTextFixture is a local writer about to apply a remote writer's
// concurrent change.
type remoteTextFixture struct {
	local     *automerge.Document
	localBody *automerge.Text
	changes   []byte
}

// newRemoteTextFixture builds the local writer and the remote change for run.
// Typing the base text in one splice anchors every character on the previous
// one, so the text is a single insertion chain as deep as it is long. Placing
// each character of a remote run deep in that chain used to walk the chain
// back to the head for the row after it. The caller closes the local writer.
func newRemoteTextFixture(
	tb testing.TB,
	engine rustParityEngine,
	run remoteTextRun,
) remoteTextFixture {
	tb.Helper()

	base, err := engine.open(actor(0x51))
	require.NoError(tb, err)

	defer func() { _ = base.Close() }()

	text, err := base.CreateText("body")
	require.NoError(tb, err)
	require.NoError(tb, text.Splice(0, 0, benchmarkText(run.baseSize)))

	_, err = base.Commit("base text", commitTime)
	require.NoError(tb, err)

	saved, err := base.Save()
	require.NoError(tb, err)

	local, err := engine.load(saved, actor(0x52))
	require.NoError(tb, err)

	remote, err := engine.load(saved, actor(0x53))
	require.NoError(tb, err)

	defer func() { _ = remote.Close() }()

	// Start the remote writer's incremental save after the shared base, so it
	// carries only the concurrent edit.
	_, err = remote.SaveIncremental()
	require.NoError(tb, err)

	// A live document has been read before remote changes arrive, which builds
	// the text's sequence index that placing a remote insertion goes through.
	localBody, err := local.Text("body")
	require.NoError(tb, err)

	_, err = localBody.String()
	require.NoError(tb, err)

	for _, edit := range []struct {
		document *automerge.Document
		value    string
	}{
		{document: local, value: run.localValue},
		{document: remote, value: run.remoteValue},
	} {
		if edit.value == "" {
			continue
		}

		body, err := edit.document.Text("body")
		require.NoError(tb, err)
		require.NoError(tb, body.Splice(run.position, 0, edit.value))

		_, err = edit.document.Commit("concurrent edit", commitTime)
		require.NoError(tb, err)
	}

	changes, err := remote.SaveIncremental()
	require.NoError(tb, err)

	return remoteTextFixture{
		local:     local,
		localBody: localBody,
		changes:   changes,
	}
}

// applyRemoteTextRun returns the local writer's text after it applies the
// remote change.
func applyRemoteTextRun(
	tb testing.TB,
	engine rustParityEngine,
	run remoteTextRun,
) string {
	tb.Helper()

	fixture := newRemoteTextFixture(tb, engine, run)

	defer func() { _ = fixture.local.Close() }()

	_, err := fixture.local.LoadIncremental(fixture.changes)
	require.NoError(tb, err)

	value, err := fixture.localBody.String()
	require.NoError(tb, err)

	return value
}

func TestApplyChanges_RemoteRunIntoLongText(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name string
		run  remoteTextRun
	}{
		{
			name: "remote run deep in the text",
			run: remoteTextRun{
				baseSize:    4_000,
				position:    3_900,
				remoteValue: strings.Repeat("r", 400),
			},
		},
		{
			name: "remote run at the head",
			run: remoteTextRun{
				baseSize:    4_000,
				position:    0,
				remoteValue: strings.Repeat("r", 400),
			},
		},
		{
			name: "remote run at the end",
			run: remoteTextRun{
				baseSize:    4_000,
				position:    4_000,
				remoteValue: strings.Repeat("r", 400),
			},
		},
		{
			name: "concurrent runs at one anchor",
			run: remoteTextRun{
				baseSize:    4_000,
				position:    3_900,
				localValue:  strings.Repeat("l", 300),
				remoteValue: strings.Repeat("r", 400),
			},
		},
		{
			name: "concurrent runs at the head",
			run: remoteTextRun{
				baseSize:    4_000,
				position:    0,
				localValue:  strings.Repeat("l", 300),
				remoteValue: strings.Repeat("r", 400),
			},
		},
	} {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				engines := rustParityEngines()
				native := applyRemoteTextRun(t, engines[0], tt.run)
				reference := applyRemoteTextRun(t, engines[1], tt.run)

				assert.Equal(t, reference, native)
				assert.Len(
					t,
					native,
					tt.run.baseSize+len(tt.run.localValue)+len(tt.run.remoteValue),
				)
			},
		)
	}
}

func BenchmarkApplyChangesRemoteRunIntoLongText(b *testing.B) {
	engine := rustParityEngines()[0]
	run := remoteTextRun{
		baseSize:    20_000,
		position:    19_900,
		remoteValue: strings.Repeat("r", 2_000),
	}

	b.ReportAllocs()
	b.StopTimer()

	for range b.N {
		fixture := newRemoteTextFixture(b, engine, run)

		b.StartTimer()

		if _, err := fixture.local.LoadIncremental(fixture.changes); err != nil {
			b.Fatal(err)
		}

		b.StopTimer()

		if err := fixture.local.Close(); err != nil {
			b.Fatal(err)
		}
	}
}
