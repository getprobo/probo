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

// This file drives randomized differential stress tests: identical random
// operation sequences are applied in lockstep to a native Go document and a
// Rust/WASM reference document, and their full observable state must match after
// every commit, after save/load round trips, and after concurrent merges. It
// exercises the incremental sequence and map caches under insertion, deletion,
// replacement, reload, and merge, which is where cache-invalidation bugs hide.
// Mark values are intentionally excluded (a known native mark-boundary defect is
// tracked separately); text content, list and map values, block structure, and
// heads are all compared.

package automerge_test

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

const stressListKey = "list"

const stressTextKey = "text"

var stressMapKeys = []string{"a", "b", "c", "d"}

// canonicalDocument renders the comparable observable state of a document: its
// sorted heads, every root map key with its scalar value, the list contents,
// and the text contents. Nested object identities are excluded because they are
// engine-specific; their materialized values are what must agree.
func canonicalDocument(t *testing.T, ctx context.Context, document *automerge.Document) string {
	t.Helper()

	return "heads:" + strings.Join(sortedHeadHex(t, ctx, document), ",") + "\n" +
		canonicalValues(t, ctx, document)
}

// canonicalValues renders the materialized state without heads. Two engines that
// independently re-encode the same concurrent edits can pick different change
// bytes (and therefore hashes) while still converging to identical values, so
// convergence tests compare values while single-actor determinism tests also
// compare heads.
func canonicalValues(t *testing.T, ctx context.Context, document *automerge.Document) string {
	t.Helper()

	var builder strings.Builder

	keys, err := document.Root().Keys(ctx)
	require.NoError(t, err)
	sort.Strings(keys)

	for _, key := range keys {
		if key == stressListKey || key == stressTextKey {
			continue
		}

		value, err := document.Root().Scalar(ctx, key)
		require.NoError(t, err)
		fmt.Fprintf(&builder, "map[%s]=%s\n", key, canonicalScalar(value))
	}

	list, err := document.Root().Object(ctx, stressListKey)
	require.NoError(t, err)

	length, err := list.Len(ctx)
	require.NoError(t, err)

	builder.WriteString("list:")

	for index := range length {
		value, err := list.ScalarAt(ctx, index)
		require.NoError(t, err)
		builder.WriteString(canonicalScalar(value))
		builder.WriteString(",")
	}

	builder.WriteString("\n")

	text, err := document.Text(ctx, stressTextKey)
	require.NoError(t, err)

	content, err := text.String(ctx)
	require.NoError(t, err)
	fmt.Fprintf(&builder, "text:%q\n", content)

	return builder.String()
}

func canonicalScalar(value automerge.Scalar) string {
	switch value.Type {
	case automerge.ScalarTypeString:
		return "s:" + value.String
	case automerge.ScalarTypeInt:
		return fmt.Sprintf("i:%d", value.Int)
	case automerge.ScalarTypeUint:
		return fmt.Sprintf("u:%d", value.Uint)
	case automerge.ScalarTypeBoolean:
		return fmt.Sprintf("b:%t", value.Bool)
	case automerge.ScalarTypeNull:
		return "null"
	default:
		return string(value.Type)
	}
}

// stressActor drives one document (native or reference) so an identical
// operation can be applied to both engines in lockstep.
type stressActor struct {
	document *automerge.Document
	list     *automerge.Object
	text     *automerge.Text
}

func newStressActor(t *testing.T, ctx context.Context, engine rustParityEngine, id byte) *stressActor {
	t.Helper()

	document, err := engine.open(ctx, actor(id))
	require.NoError(t, err)
	closeDocument(t, document)

	list, err := document.Root().CreateObject(ctx, stressListKey, automerge.ObjectTypeList)
	require.NoError(t, err)

	text, err := document.CreateText(ctx, stressTextKey)
	require.NoError(t, err)

	_, err = document.Commit(ctx, "seed", commitTime)
	require.NoError(t, err)

	return &stressActor{document: document, list: list, text: text}
}

func randomScalar(random *rand.Rand) automerge.Scalar {
	switch random.Intn(4) {
	case 0:
		return automerge.Scalar{Type: automerge.ScalarTypeString, String: randomLetters(random, 1+random.Intn(4))}
	case 1:
		return automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(random.Intn(1000)) - 500}
	case 2:
		return automerge.Scalar{Type: automerge.ScalarTypeUint, Uint: uint64(random.Intn(1000))}
	default:
		return automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: random.Intn(2) == 0}
	}
}

// applyStressOperation performs one deterministic operation, described by the
// random source, on the given actor. The same random draws applied to two
// actors produce identical operations, keeping the engines in lockstep.
func applyStressOperation(
	t *testing.T,
	ctx context.Context,
	actor *stressActor,
	op stressOperation,
) {
	t.Helper()

	switch op.kind {
	case opMapPut:
		require.NoError(t, actor.document.Root().PutScalar(ctx, op.key, op.scalar))
	case opMapDelete:
		require.NoError(t, actor.document.Root().DeleteKey(ctx, op.key))
	case opListInsert:
		require.NoError(t, actor.list.InsertScalar(ctx, op.index, op.scalar))
	case opListPut:
		require.NoError(t, actor.list.PutScalarAt(ctx, op.index, op.scalar))
	case opListDelete:
		require.NoError(t, actor.list.DeleteIndex(ctx, op.index))
	case opTextInsert:
		require.NoError(t, actor.text.Splice(ctx, uint32(op.index), 0, op.value))
	case opTextDelete:
		require.NoError(t, actor.text.Splice(ctx, uint32(op.index), int32(op.count), ""))
	}
}

type stressOpKind int

const (
	opMapPut stressOpKind = iota
	opMapDelete
	opListInsert
	opListPut
	opListDelete
	opTextInsert
	opTextDelete
)

type stressOperation struct {
	kind   stressOpKind
	key    string
	index  uint64
	count  uint64
	value  string
	scalar automerge.Scalar
}

// nextStressOperation chooses a valid operation from the current model lengths so
// indices stay in range for both engines.
func nextStressOperation(random *rand.Rand, listLen, textLen uint64, present map[string]bool) stressOperation {
	for {
		switch random.Intn(7) {
		case 0:
			return stressOperation{kind: opMapPut, key: stressMapKeys[random.Intn(len(stressMapKeys))], scalar: randomScalar(random)}
		case 1:
			existing := make([]string, 0, len(present))
			for key := range present {
				existing = append(existing, key)
			}

			if len(existing) == 0 {
				continue
			}

			sort.Strings(existing)

			return stressOperation{kind: opMapDelete, key: existing[random.Intn(len(existing))]}
		case 2:
			return stressOperation{kind: opListInsert, index: uint64(random.Intn(int(listLen) + 1)), scalar: randomScalar(random)}
		case 3:
			if listLen == 0 {
				continue
			}

			return stressOperation{kind: opListPut, index: uint64(random.Intn(int(listLen))), scalar: randomScalar(random)}
		case 4:
			if listLen == 0 {
				continue
			}

			return stressOperation{kind: opListDelete, index: uint64(random.Intn(int(listLen)))}
		case 5:
			return stressOperation{kind: opTextInsert, index: uint64(random.Intn(int(textLen) + 1)), value: randomLetters(random, 1+random.Intn(4))}
		case 6:
			if textLen == 0 {
				continue
			}

			deleteCount := 1 + random.Intn(int(textLen))

			return stressOperation{kind: opTextDelete, index: uint64(random.Intn(int(textLen) - deleteCount + 1)), count: uint64(deleteCount)}
		}
	}
}

// TestDifferentialStress_SingleDocument applies identical random operations to a
// native and a reference document and asserts their observable state matches
// after every commit, with periodic save/load round trips.
func TestDifferentialStress_SingleDocument(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	random := rand.New(rand.NewSource(0x1f2e3d4c5b6a7988))

	const (
		scenarios = 40
		steps     = 60
	)

	for scenario := range scenarios {
		native := newStressActor(t, ctx, rustParityEngines()[0], 0x01)
		reference := newStressActor(t, ctx, rustParityEngines()[1], 0x01)

		require.Equal(t,
			canonicalDocument(t, ctx, reference.document),
			canonicalDocument(t, ctx, native.document),
			"scenario %d seed diverged", scenario,
		)

		var listLen, textLen uint64

		present := make(map[string]bool)

		for step := range steps {
			op := nextStressOperation(random, listLen, textLen, present)

			applyStressOperation(t, ctx, native, op)
			applyStressOperation(t, ctx, reference, op)

			switch op.kind {
			case opMapPut:
				present[op.key] = true
			case opMapDelete:
				delete(present, op.key)
			}

			nativeCommitted := tolerantCommit(t, ctx, native.document)
			referenceCommitted := tolerantCommit(t, ctx, reference.document)
			require.Equal(t, referenceCommitted, nativeCommitted,
				"scenario %d step %d op %+v commit divergence", scenario, step, op)

			listLen = mustLen(t, ctx, native.list)
			textLen = mustTextLen(t, ctx, native.text)

			require.Equal(t,
				canonicalDocument(t, ctx, reference.document),
				canonicalDocument(t, ctx, native.document),
				"scenario %d step %d op %+v diverged", scenario, step, op,
			)
		}

		// A save/load round trip rebuilds every cache from scratch; the reloaded
		// state must equal the pre-save state and still match the reference.
		saved, err := native.document.Save(ctx)
		require.NoError(t, err)

		reloaded, err := automerge.Load(ctx, saved, actor(0x09))
		require.NoError(t, err)
		closeDocument(t, reloaded)

		require.Equal(t,
			canonicalDocument(t, ctx, native.document),
			canonicalDocument(t, ctx, reloaded),
			"scenario %d reload diverged", scenario,
		)
	}
}

// tolerantCommit commits and reports whether a change was produced. A step whose
// operation was a no-op (for example, a put of the identical value) yields no
// change on both engines, which is expected and not an error.
func tolerantCommit(t *testing.T, ctx context.Context, document *automerge.Document) bool {
	t.Helper()

	_, err := document.Commit(ctx, "step", commitTime)
	if err != nil {
		require.ErrorContains(t, err, "no operations")

		return false
	}

	return true
}

// TestDifferentialStress_ConcurrentMerge drives two forked peers per engine
// through independent random edits and then merges them in both directions,
// asserting the merged native and reference documents converge to identical
// observable state. This stresses cache invalidation across merges and the
// determinism of conflict resolution.
func TestDifferentialStress_ConcurrentMerge(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	random := rand.New(rand.NewSource(0x6c5f4e3d2c1b0a99))

	const (
		scenarios  = 40
		roundEdits = 12
		rounds     = 3
	)

	for scenario := range scenarios {
		nativeLeft := newStressActor(t, ctx, rustParityEngines()[0], 0x01)
		referenceLeft := newStressActor(t, ctx, rustParityEngines()[1], 0x01)

		// Both peers start from the same seeded document so their concurrent
		// edits build on shared history.
		seedSaved, err := nativeLeft.document.Save(ctx)
		require.NoError(t, err)

		nativeRight := forkStressActor(t, ctx, rustParityEngines()[0], seedSaved, 0x02)

		referenceSeed, err := referenceLeft.document.Save(ctx)
		require.NoError(t, err)

		referenceRight := forkStressActor(t, ctx, rustParityEngines()[1], referenceSeed, 0x02)

		for round := range rounds {
			editStressActor(t, ctx, random, nativeLeft, referenceLeft, roundEdits)
			editStressActor(t, ctx, random, nativeRight, referenceRight, roundEdits)

			mergeDocuments(t, ctx, nativeLeft.document, nativeRight.document)
			mergeDocuments(t, ctx, referenceLeft.document, referenceRight.document)

			require.Equal(t,
				canonicalValues(t, ctx, referenceLeft.document),
				canonicalValues(t, ctx, nativeLeft.document),
				"scenario %d round %d merged state diverged", scenario, round,
			)
		}
	}
}

func forkStressActor(
	t *testing.T,
	ctx context.Context,
	engine rustParityEngine,
	saved []byte,
	id byte,
) *stressActor {
	t.Helper()

	document, err := engine.load(ctx, saved, actor(id))
	require.NoError(t, err)
	closeDocument(t, document)

	list, err := document.Root().Object(ctx, stressListKey)
	require.NoError(t, err)

	text, err := document.Text(ctx, stressTextKey)
	require.NoError(t, err)

	return &stressActor{document: document, list: list, text: text}
}

// editStressActor applies the same random edits to a native and a reference peer
// so both engines diverge identically before a merge.
func editStressActor(
	t *testing.T,
	ctx context.Context,
	random *rand.Rand,
	native, reference *stressActor,
	edits int,
) {
	t.Helper()

	present := make(map[string]bool)

	for range edits {
		op := nextStressOperation(
			random,
			mustLen(t, ctx, native.list),
			mustTextLen(t, ctx, native.text),
			present,
		)

		applyStressOperation(t, ctx, native, op)
		applyStressOperation(t, ctx, reference, op)

		switch op.kind {
		case opMapPut:
			present[op.key] = true
		case opMapDelete:
			delete(present, op.key)
		}
	}

	tolerantCommit(t, ctx, native.document)
	tolerantCommit(t, ctx, reference.document)
}

func mergeDocuments(t *testing.T, ctx context.Context, left, right *automerge.Document) {
	t.Helper()

	_, err := left.Merge(ctx, right)
	require.NoError(t, err)

	_, err = right.Merge(ctx, left)
	require.NoError(t, err)
}

// FuzzDifferentialOperations drives the lockstep native/reference comparison
// from fuzzer-provided bytes so continuous fuzzing can keep exploring the
// operation space. Each byte selects and parameterizes one operation; after
// every commit the two engines must agree on materialized values.
func FuzzDifferentialOperations(f *testing.F) {
	f.Add([]byte{0x05, 0x41, 0x05, 0x42, 0x02, 0x10, 0x00, 0x03, 0x00, 0x20})
	f.Add([]byte{0x02, 0x00, 0x11, 0x02, 0x01, 0x22, 0x06, 0x00, 0x33, 0x05, 0x00})

	f.Fuzz(func(t *testing.T, script []byte) {
		ctx := context.Background()

		native := newStressActor(t, ctx, rustParityEngines()[0], 0x01)
		reference := newStressActor(t, ctx, rustParityEngines()[1], 0x01)

		present := make(map[string]bool)

		for cursor := 0; cursor+1 < len(script); cursor += 2 {
			op := scriptOperation(
				script[cursor],
				script[cursor+1],
				mustLen(t, ctx, native.list),
				mustTextLen(t, ctx, native.text),
				present,
			)
			if op == nil {
				continue
			}

			applyStressOperation(t, ctx, native, *op)
			applyStressOperation(t, ctx, reference, *op)

			switch op.kind {
			case opMapPut:
				present[op.key] = true
			case opMapDelete:
				delete(present, op.key)
			}

			nativeCommitted := tolerantCommit(t, ctx, native.document)
			referenceCommitted := tolerantCommit(t, ctx, reference.document)
			require.Equal(t, referenceCommitted, nativeCommitted, "commit divergence for %+v", *op)

			require.Equal(t,
				canonicalValues(t, ctx, reference.document),
				canonicalValues(t, ctx, native.document),
				"value divergence after %+v", *op,
			)
		}
	})
}

// scriptOperation decodes a two-byte instruction into a valid operation for the
// current model sizes, or nil when the instruction cannot form a valid one.
func scriptOperation(selector, param byte, listLen, textLen uint64, present map[string]bool) *stressOperation {
	key := stressMapKeys[int(param)%len(stressMapKeys)]
	scalar := automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(param)}

	switch selector % 7 {
	case 0:
		return &stressOperation{kind: opMapPut, key: key, scalar: scalar}
	case 1:
		if !present[key] {
			return nil
		}

		return &stressOperation{kind: opMapDelete, key: key}
	case 2:
		return &stressOperation{kind: opListInsert, index: uint64(param) % (listLen + 1), scalar: scalar}
	case 3:
		if listLen == 0 {
			return nil
		}

		return &stressOperation{kind: opListPut, index: uint64(param) % listLen, scalar: scalar}
	case 4:
		if listLen == 0 {
			return nil
		}

		return &stressOperation{kind: opListDelete, index: uint64(param) % listLen}
	case 5:
		return &stressOperation{kind: opTextInsert, index: uint64(param) % (textLen + 1), value: string(rune('a' + int(param)%26))}
	case 6:
		if textLen == 0 {
			return nil
		}

		return &stressOperation{kind: opTextDelete, index: uint64(param) % textLen, count: 1}
	}

	return nil
}

func mustLen(t *testing.T, ctx context.Context, object *automerge.Object) uint64 {
	t.Helper()

	length, err := object.Len(ctx)
	require.NoError(t, err)

	return length
}

func mustTextLen(t *testing.T, ctx context.Context, text *automerge.Text) uint64 {
	t.Helper()

	content, err := text.String(ctx)
	require.NoError(t, err)

	return uint64(len([]rune(content)))
}
