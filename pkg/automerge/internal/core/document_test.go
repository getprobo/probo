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

package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

func TestEngineSave_RetainsUnknownColumnsAfterMutation(t *testing.T) {
	t.Parallel()

	actor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)

	property := "payload"
	identifier := opset.OpID{Actor: actor, Counter: 1}
	change := opset.Change{
		Actor:    actor,
		Sequence: 1,
		StartOp:  1,
		MaxOp:    1,
		Operations: []opset.Operation{{
			ID:     identifier,
			Object: opset.RootObject(),
			Key:    opset.Key{Property: &property},
			Action: opset.ActionSet,
			Value:  &opset.Scalar{Type: opset.ScalarString, String: "initial"},
		}},
	}
	_, err = storage.EncodeChange(&change)
	require.NoError(t, err)

	document := &opset.Document{
		Heads:   []opset.ChangeHash{*change.Hash},
		Changes: []opset.Change{change},
		UnknownColumns: []opset.RawColumn{{
			Specification: 200,
			Data:          []byte{6, 7},
		}},
	}
	data, err := storage.EncodeDocument(document, []opset.OpID{identifier}, false)
	require.NoError(t, err)

	engine, err := LoadEngine(data)
	require.NoError(t, err)
	require.NoError(t, engine.PutString(0, "key", "value"))
	_, err = engine.Commit("mutation", time.Time{})
	require.NoError(t, err)

	saved, err := engine.Save(true, false)
	require.NoError(t, err)
	decoded, err := storage.Decode(saved)
	require.NoError(t, err)

	assert.Contains(t, decoded.UnknownColumns, opset.RawColumn{
		Specification: 192,
		Data:          []byte{6, 7},
	})
}
