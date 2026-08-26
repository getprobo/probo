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
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

// TestConcurrentEncodingIsByteIdentical is the strict determinism gate: two
// peers per engine make independent edits over a shared base and merge. The two
// engines must not merely converge to the same values, they must encode each
// change to the same bytes and therefore agree on every hash.
//
// This is stronger than convergence and it is what caught the conflicted-put
// divergence: assigning the value a conflicted key already resolves to must emit
// a delete of the losing siblings, not a fresh assignment and not nothing.
func TestConcurrentEncodingIsByteIdentical(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const (
		scenarios  = 120
		roundEdits = 12
		rounds     = 3
	)

	for scenario := range scenarios {
		random := rand.New(rand.NewSource(int64(scenario)))

		nativeLeft := newStressActor(t, ctx, rustParityEngines()[0], 0x01)
		referenceLeft := newStressActor(t, ctx, rustParityEngines()[1], 0x01)

		seedSaved, err := nativeLeft.document.Save(ctx)
		require.NoError(t, err)

		referenceSeed, err := referenceLeft.document.Save(ctx)
		require.NoError(t, err)

		nativeRight := forkStressActor(t, ctx, rustParityEngines()[0], seedSaved, 0x02)
		referenceRight := forkStressActor(t, ctx, rustParityEngines()[1], referenceSeed, 0x02)

		peers := []*stressActor{nativeLeft, referenceLeft, nativeRight, referenceRight}
		drainIncremental(t, ctx, peers)

		for round := range rounds {
			editStressActor(t, ctx, random, nativeLeft, referenceLeft, roundEdits)
			editStressActor(t, ctx, random, nativeRight, referenceRight, roundEdits)

			assertIdenticalEncoding(t, ctx, scenario, round, "left", nativeLeft, referenceLeft)
			assertIdenticalEncoding(t, ctx, scenario, round, "right", nativeRight, referenceRight)

			mergeDocuments(t, ctx, nativeLeft.document, nativeRight.document)
			mergeDocuments(t, ctx, referenceLeft.document, referenceRight.document)

			assert.Equalf(
				t,
				canonicalDocument(t, ctx, referenceLeft.document),
				canonicalDocument(t, ctx, nativeLeft.document),
				"scenario %d round %d merged document diverged",
				scenario,
				round,
			)

			drainIncremental(t, ctx, peers)
		}
	}
}

// TestPutMatchesReferenceOnConflictedKey pins the exact rule the strict gate
// discovered, for a map key and a list element holding concurrent values.
func TestPutMatchesReferenceOnConflictedKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name  string
		value string
	}{
		{name: "equal to the winning value", value: "R"},
		{name: "equal to the losing value", value: "L"},
		{name: "a new value", value: "N"},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				native := conflictedMapDocument(t, ctx, rustParityEngines()[0])
				reference := conflictedMapDocument(t, ctx, rustParityEngines()[1])

				put := automerge.Scalar{Type: automerge.ScalarTypeString, String: tt.value}
				require.NoError(t, native.Root().PutScalar(ctx, "key", put))
				require.NoError(t, reference.Root().PutScalar(ctx, "key", put))

				assert.Equal(
					t,
					commitAndEncode(t, ctx, reference),
					commitAndEncode(t, ctx, native),
				)
				assert.Equal(
					t,
					mapKeySignature(t, ctx, reference, "key"),
					mapKeySignature(t, ctx, native, "key"),
				)
			},
		)
	}
}

func TestPutMatchesReferenceOnConflictedListElement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, value := range []string{"R", "L", "N"} {
		t.Run(
			"put "+value,
			func(t *testing.T) {
				t.Parallel()

				native, nativeList := conflictedListDocument(t, ctx, rustParityEngines()[0])
				reference, referenceList := conflictedListDocument(t, ctx, rustParityEngines()[1])

				put := automerge.Scalar{Type: automerge.ScalarTypeString, String: value}
				require.NoError(t, nativeList.PutScalarAt(ctx, 0, put))
				require.NoError(t, referenceList.PutScalarAt(ctx, 0, put))

				assert.Equal(
					t,
					commitAndEncode(t, ctx, reference),
					commitAndEncode(t, ctx, native),
				)
				assert.Equal(
					t,
					listElementSignature(t, ctx, reference, referenceList),
					listElementSignature(t, ctx, native, nativeList),
				)
			},
		)
	}
}

// mapKeySignature describes both the resolved value and the full conflict set at
// a key, so a put that only appears to work is still caught.
func mapKeySignature(
	t *testing.T,
	ctx context.Context,
	document *automerge.Document,
	key string,
) string {
	t.Helper()

	winner, err := document.Root().Scalar(ctx, key)
	require.NoError(t, err)

	conflicts, err := document.Root().Scalars(ctx, key)
	require.NoError(t, err)

	return fmt.Sprintf(
		"heads=%v winner=%s conflicts=%s",
		sortedHeadHex(t, ctx, document),
		canonicalScalar(winner),
		describeScalars(conflicts),
	)
}

func listElementSignature(
	t *testing.T,
	ctx context.Context,
	document *automerge.Document,
	list *automerge.Object,
) string {
	t.Helper()

	winner, err := list.ScalarAt(ctx, 0)
	require.NoError(t, err)

	conflicts, err := list.ScalarsAt(ctx, 0)
	require.NoError(t, err)

	return fmt.Sprintf(
		"heads=%v winner=%s conflicts=%s",
		sortedHeadHex(t, ctx, document),
		canonicalScalar(winner),
		describeScalars(conflicts),
	)
}

func describeScalars(values []automerge.Scalar) string {
	rendered := make([]string, 0, len(values))
	for _, value := range values {
		rendered = append(rendered, canonicalScalar(value))
	}

	sort.Strings(rendered)

	return strings.Join(rendered, "|")
}

func drainIncremental(t *testing.T, ctx context.Context, actors []*stressActor) {
	t.Helper()

	for _, peer := range actors {
		_, err := peer.document.SaveIncremental(ctx)
		require.NoError(t, err)
	}
}

func assertIdenticalEncoding(
	t *testing.T,
	ctx context.Context,
	scenario int,
	round int,
	side string,
	nativeActor *stressActor,
	referenceActor *stressActor,
) {
	t.Helper()

	nativeBytes, err := nativeActor.document.SaveIncremental(ctx)
	require.NoError(t, err)

	referenceBytes, err := referenceActor.document.SaveIncremental(ctx)
	require.NoError(t, err)

	assert.Truef(
		t,
		bytes.Equal(nativeBytes, referenceBytes),
		"scenario %d round %d %s peer encoded its change differently (%d native bytes, %d reference bytes)",
		scenario,
		round,
		side,
		len(nativeBytes),
		len(referenceBytes),
	)
}

// conflictedMapDocument returns a document whose "key" property holds two
// concurrent values, "L" from the first actor and the winning "R" from the
// second.
func conflictedMapDocument(
	t *testing.T,
	ctx context.Context,
	engine rustParityEngine,
) *automerge.Document {
	t.Helper()

	base, err := engine.open(ctx, actor(0x01))
	require.NoError(t, err)
	closeDocument(t, base)

	require.NoError(
		t,
		base.Root().PutScalar(
			ctx,
			"key",
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "base"},
		),
	)
	_, err = base.Commit(ctx, "seed", commitTime)
	require.NoError(t, err)

	left, right := forkPair(t, ctx, engine, base)

	require.NoError(
		t,
		left.Root().PutScalar(
			ctx,
			"key",
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "L"},
		),
	)
	_, err = left.Commit(ctx, "left", commitTime)
	require.NoError(t, err)

	require.NoError(
		t,
		right.Root().PutScalar(
			ctx,
			"key",
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "R"},
		),
	)
	_, err = right.Commit(ctx, "right", commitTime)
	require.NoError(t, err)

	_, err = left.Merge(ctx, right)
	require.NoError(t, err)

	_, err = left.SaveIncremental(ctx)
	require.NoError(t, err)

	return left
}

func conflictedListDocument(
	t *testing.T,
	ctx context.Context,
	engine rustParityEngine,
) (*automerge.Document, *automerge.Object) {
	t.Helper()

	base, err := engine.open(ctx, actor(0x01))
	require.NoError(t, err)
	closeDocument(t, base)

	list, err := base.Root().CreateObject(ctx, "list", automerge.ObjectTypeList)
	require.NoError(t, err)
	require.NoError(
		t,
		list.InsertScalar(
			ctx,
			0,
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "base"},
		),
	)
	_, err = base.Commit(ctx, "seed", commitTime)
	require.NoError(t, err)

	left, right := forkPair(t, ctx, engine, base)

	leftList, err := left.Root().Object(ctx, "list")
	require.NoError(t, err)
	require.NoError(
		t,
		leftList.PutScalarAt(
			ctx,
			0,
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "L"},
		),
	)
	_, err = left.Commit(ctx, "left", commitTime)
	require.NoError(t, err)

	rightList, err := right.Root().Object(ctx, "list")
	require.NoError(t, err)
	require.NoError(
		t,
		rightList.PutScalarAt(
			ctx,
			0,
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "R"},
		),
	)
	_, err = right.Commit(ctx, "right", commitTime)
	require.NoError(t, err)

	_, err = left.Merge(ctx, right)
	require.NoError(t, err)

	_, err = left.SaveIncremental(ctx)
	require.NoError(t, err)

	return left, leftList
}

func forkPair(
	t *testing.T,
	ctx context.Context,
	engine rustParityEngine,
	base *automerge.Document,
) (*automerge.Document, *automerge.Document) {
	t.Helper()

	saved, err := base.Save(ctx)
	require.NoError(t, err)

	left, err := engine.load(ctx, saved, actor(0x01))
	require.NoError(t, err)
	closeDocument(t, left)

	right, err := engine.load(ctx, saved, actor(0x02))
	require.NoError(t, err)
	closeDocument(t, right)

	return left, right
}

// commitAndEncode commits the pending operations and returns the encoded change
// bytes, or a marker when the engine had nothing to record.
func commitAndEncode(t *testing.T, ctx context.Context, document *automerge.Document) string {
	t.Helper()

	if _, err := document.Commit(ctx, "put", commitTime); err != nil {
		return "no change: " + err.Error()
	}

	encoded, err := document.SaveIncremental(ctx)
	require.NoError(t, err)

	return fmt.Sprintf("%x", encoded)
}
