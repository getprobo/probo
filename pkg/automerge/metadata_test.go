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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/automerge/internal/native"
)

// TestDocument_StatsMatchReference reproduces stats_smoke_test.
func TestDocument_StatsMatchReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	factories := map[string]func(
		context.Context,
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}

	for name, factory := range factories {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				document, err := factory(ctx, actor(1))
				require.NoError(t, err)
				closeDocument(t, document)

				require.NoError(
					t,
					document.Root().PutScalar(
						ctx,
						"a",
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 1},
					),
				)
				_, err = document.Commit(ctx, "a", commitTime)
				require.NoError(t, err)

				require.NoError(
					t,
					document.Root().PutScalar(
						ctx,
						"b",
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 2},
					),
				)
				_, err = document.Commit(ctx, "b", commitTime.Add(time.Second))
				require.NoError(t, err)

				stats, err := document.Stats(ctx)
				require.NoError(t, err)
				assert.Equal(t, uint64(2), stats.NumChanges)
				assert.Equal(t, uint64(2), stats.NumOps)
				assert.Equal(t, uint64(1), stats.NumActors)
			},
		)
	}
}

func TestDocument_CommitTimeParity(t *testing.T) {
	t.Parallel()

	factories := map[string]func(
		context.Context,
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}

	for name, factory := range factories {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				ctx := context.Background()
				document, err := factory(ctx, actor(149))
				require.NoError(t, err)
				closeDocument(t, document)

				require.NoError(t, document.PutString(ctx, "zero", "value"))
				_, err = document.Commit(ctx, "zero", time.Time{})
				require.NoError(t, err)
				require.NoError(t, document.PutString(ctx, "provided", "value"))
				_, err = document.Commit(
					ctx,
					"provided",
					time.Unix(12_345, 0),
				)
				require.NoError(t, err)
				require.NoError(t, document.PutString(ctx, "current", "value"))

				before := time.Now().Unix()
				_, err = document.CommitNow(ctx, "current")
				after := time.Now().Unix()

				require.NoError(t, err)

				data, err := document.Save(ctx)
				require.NoError(t, err)
				decoded, err := native.Decode(data)
				require.NoError(t, err)
				require.Len(t, decoded.Changes, 3)
				assert.Equal(t, int64(0), decoded.Changes[0].Time)
				assert.Equal(t, int64(12_345), decoded.Changes[1].Time)
				assert.GreaterOrEqual(t, decoded.Changes[2].Time, before)
				assert.LessOrEqual(t, decoded.Changes[2].Time, after)
			},
		)
	}
}

func TestDocument_EmptyCommitTimeParity(t *testing.T) {
	t.Parallel()

	factories := map[string]func(
		context.Context,
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}

	for name, factory := range factories {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				ctx := context.Background()
				document, err := factory(ctx, actor(150))
				require.NoError(t, err)
				closeDocument(t, document)

				_, err = document.EmptyCommit(ctx, "zero", time.Time{})
				require.NoError(t, err)
				_, err = document.EmptyCommit(
					ctx,
					"provided",
					time.Unix(12_345, 0),
				)
				require.NoError(t, err)

				before := time.Now().Unix()
				_, err = document.EmptyCommitNow(ctx, "current")
				after := time.Now().Unix()

				require.NoError(t, err)

				data, err := document.Save(ctx)
				require.NoError(t, err)
				decoded, err := native.Decode(data)
				require.NoError(t, err)
				require.Len(t, decoded.Changes, 3)

				for index, change := range decoded.Changes {
					assert.Equal(t, uint64(index+1), change.Sequence)
					assert.Equal(t, uint64(1), change.StartOp)
					assert.Equal(t, uint64(0), change.MaxOp)
					assert.Empty(t, change.Operations)
				}

				assert.Equal(t, int64(0), decoded.Changes[0].Time)
				assert.Equal(t, int64(12_345), decoded.Changes[1].Time)
				assert.GreaterOrEqual(t, decoded.Changes[2].Time, before)
				assert.LessOrEqual(t, decoded.Changes[2].Time, after)
			},
		)
	}
}

func TestDocument_EmptyCommitChangesSince(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	document, err := automerge.New(ctx, actor(156))
	require.NoError(t, err)
	closeDocument(t, document)
	hash, err := document.EmptyCommit(ctx, "empty", time.Time{})
	require.NoError(t, err)

	changes, err := document.ChangesSince(ctx, nil)
	require.NoError(t, err)
	require.Len(t, changes, 1)
	assert.Equal(t, hash, changes[0].Hash)
	changes, err = document.ChangesSince(ctx, []automerge.Hash{hash})
	require.NoError(t, err)
	assert.Empty(t, changes)

	data, err := document.Save(ctx)
	require.NoError(t, err)
	reference, err := automerge.LoadReference(ctx, data, actor(157))
	require.NoError(t, err)
	closeDocument(t, reference)
	heads, err := reference.Heads(ctx)
	require.NoError(t, err)
	assert.Equal(t, []automerge.Hash{hash}, heads)
}

func TestDocument_HistoricalReadsMatchReference(t *testing.T) {
	t.Parallel()

	factories := map[string]func(
		context.Context,
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}

	for name, factory := range factories {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				ctx := context.Background()
				document, err := factory(ctx, actor(158))
				require.NoError(t, err)
				closeDocument(t, document)
				root := document.Root()
				require.NoError(
					t,
					root.PutScalar(
						ctx,
						"value",
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 1},
					),
				)
				text, err := document.CreateText(ctx, "body")
				require.NoError(t, err)
				require.NoError(t, text.Splice(ctx, 0, 0, "A"))
				first, err := document.Commit(ctx, "first", commitTime)
				require.NoError(t, err)

				require.NoError(
					t,
					root.PutScalar(
						ctx,
						"value",
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 2},
					),
				)
				require.NoError(t, text.Splice(ctx, 1, 0, "B"))
				second, err := document.Commit(
					ctx,
					"second",
					commitTime.Add(time.Second),
				)
				require.NoError(t, err)

				historicalScalar, err := root.ScalarAtHeads(
					ctx,
					"value",
					[]automerge.Hash{first},
				)
				require.NoError(t, err)
				assert.Equal(t, int64(1), historicalScalar.Int)

				currentScalar, err := root.Scalar(ctx, "value")
				require.NoError(t, err)
				assert.Equal(t, int64(2), currentScalar.Int)

				historicalText, err := text.StringAt(
					ctx,
					[]automerge.Hash{first},
				)
				require.NoError(t, err)
				assert.Equal(t, "A", historicalText)

				currentText, err := text.String(ctx)
				require.NoError(t, err)
				assert.Equal(t, "AB", currentText)

				hasHeads, err := document.HasHeads(
					ctx,
					[]automerge.Hash{first, second},
				)
				require.NoError(t, err)
				assert.True(t, hasHeads)
				hasHeads, err = document.HasHeads(ctx, nil)
				require.NoError(t, err)
				assert.True(t, hasHeads)

				var unknown automerge.Hash

				unknown[0] = 1
				hasHeads, err = document.HasHeads(
					ctx,
					[]automerge.Hash{unknown},
				)
				require.NoError(t, err)
				assert.False(t, hasHeads)
			},
		)
	}
}

func TestDocument_MissingDependenciesMatchReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nativeDocument, err := automerge.New(ctx, actor(159))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReference(ctx, actor(159))
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	var unknown automerge.Hash

	unknown[0] = 1
	nativeMissing, err := nativeDocument.MissingDependencies(
		ctx,
		[]automerge.Hash{unknown},
	)
	require.NoError(t, err)
	referenceMissing, err := referenceDocument.MissingDependencies(
		ctx,
		[]automerge.Hash{unknown},
	)
	require.NoError(t, err)
	assert.Equal(t, referenceMissing, nativeMissing)
	assert.Equal(t, []automerge.Hash{unknown}, nativeMissing)

	for _, document := range []*automerge.Document{
		nativeDocument,
		referenceDocument,
	} {
		require.NoError(t, document.PutString(ctx, "value", "known"))
		hash, err := document.Commit(ctx, "known", commitTime)
		require.NoError(t, err)
		missing, err := document.MissingDependencies(
			ctx,
			[]automerge.Hash{hash},
		)
		require.NoError(t, err)
		assert.Empty(t, missing)
	}
}
