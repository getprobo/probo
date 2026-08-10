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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/reference"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

func TestProbeDocumentOrdering(t *testing.T) {
	ctx := context.Background()

	first, err := reference.New(ctx)
	require.NoError(t, err)
	require.NoError(t, first.SetActor(ctx, []byte{0x20, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}))

	require.NoError(t, first.PutString(ctx, 0, "title", "one"))

	list, err := first.PutObject(ctx, 0, "items", "list")
	require.NoError(t, err)
	require.NoError(t, first.InsertScalar(ctx, list, 0, []byte(`{"type":"string","string":"a"}`)))
	require.NoError(t, first.InsertScalar(ctx, list, 1, []byte(`{"type":"string","string":"b"}`)))

	body, err := first.PutText(ctx, 0, "body")
	require.NoError(t, err)
	require.NoError(t, first.SpliceText(ctx, body, 0, 0, "hello"))
	_, err = first.Commit(ctx, "first", time.Unix(1, 0))
	require.NoError(t, err)

	shared, err := first.Save(ctx)
	require.NoError(t, err)

	second, err := reference.Load(ctx, shared)
	require.NoError(t, err)
	require.NoError(t, second.SetActor(ctx, []byte{0x10, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}))

	secondBody, _, err := second.GetObject(ctx, 0, "body")
	require.NoError(t, err)
	require.NoError(t, second.SpliceText(ctx, secondBody, 5, 0, " there"))
	require.NoError(t, second.PutString(ctx, 0, "title", "two"))
	_, err = second.Commit(ctx, "second", time.Unix(2, 0))
	require.NoError(t, err)

	secondSave, err := second.Save(ctx)
	require.NoError(t, err)

	_, err = first.Merge(ctx, secondSave)
	require.NoError(t, err)

	require.NoError(t, first.SpliceText(ctx, body, 0, 1, ""))
	_, err = first.Commit(ctx, "third", time.Unix(3, 0))
	require.NoError(t, err)

	saved, err := first.Save(ctx)
	require.NoError(t, err)

	document, err := storage.Decode(saved)
	require.NoError(t, err)

	t.Logf("actors (in table order):")

	for i, actorID := range document.Actors {
		t.Logf("  [%d] %s", i, actorID)
	}

	t.Logf("changes (in stored order):")

	for i := range document.Changes {
		change := &document.Changes[i]
		t.Logf("  [%d] actor=%.4s seq=%d startOp=%d maxOp=%d deps=%v msg=%q",
			i, change.Actor.String(), change.Sequence, change.StartOp,
			change.MaxOp, change.DependencyIndexes, change.Message)
	}

	t.Logf("operations (in stored order, reconstructed per change):")

	type row struct {
		change int
		op     storage.Operation
	}

	rows := make([]row, 0)
	for i := range document.Changes {
		for _, operation := range document.Changes[i].Operations {
			rows = append(rows, row{change: i, op: operation})
		}
	}

	for _, item := range rows {
		t.Logf("  change=%d %s", item.change, formatOperation(item.op))
	}
}

func formatOperation(operation storage.Operation) string {
	object := "root"
	if !operation.Object.IsRoot {
		object = fmt.Sprintf("%.4s@%d", operation.Object.OpID.Actor.String(), operation.Object.OpID.Counter)
	}

	key := "?"

	switch {
	case operation.Key.Property != nil:
		key = "prop:" + *operation.Key.Property
	case operation.Key.IsHead:
		key = "head"
	case operation.Key.Element != nil:
		key = fmt.Sprintf("elem:%.4s@%d", operation.Key.Element.Actor.String(), operation.Key.Element.Counter)
	}

	return fmt.Sprintf(
		"id=%.4s@%d action=%d obj=%s key=%s insert=%t succ=%v",
		operation.ID.Actor.String(), operation.ID.Counter,
		operation.Action, object, key, operation.Insert, operation.Successors,
	)
}
