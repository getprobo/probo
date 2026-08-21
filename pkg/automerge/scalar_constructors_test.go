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
	"go.probo.inc/probo/pkg/automerge"
)

// TestScalarConstructors verifies each constructor pairs its type with the
// matching field, which is the misuse the plain struct literal invites.
func TestScalarConstructors(t *testing.T) {
	t.Parallel()

	assert.Equal(t, automerge.Scalar{Type: automerge.ScalarTypeNull}, automerge.NullScalar())
	assert.Equal(t, automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true}, automerge.BoolScalar(true))
	assert.Equal(t, automerge.Scalar{Type: automerge.ScalarTypeUint, Uint: 7}, automerge.UintScalar(7))
	assert.Equal(t, automerge.Scalar{Type: automerge.ScalarTypeInt, Int: -7}, automerge.IntScalar(-7))
	assert.Equal(t, automerge.Scalar{Type: automerge.ScalarTypeFloat64, Float: 1.5}, automerge.FloatScalar(1.5))
	assert.Equal(t, automerge.Scalar{Type: automerge.ScalarTypeString, String: "x"}, automerge.StringScalar("x"))
	assert.Equal(t, automerge.Scalar{Type: automerge.ScalarTypeBytes, Bytes: []byte{1, 2}}, automerge.BytesScalar([]byte{1, 2}))
	assert.Equal(t, automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 5}, automerge.CounterScalar(5))
	assert.Equal(t, automerge.Scalar{Type: automerge.ScalarTypeTimestamp, Int: 1000}, automerge.TimestampScalar(1000))
}

// TestActorIDString checks the actor ID renders as lowercase hex like Hash.
func TestActorIDString(t *testing.T) {
	t.Parallel()

	var actorID automerge.ActorID
	actorID[0] = 0xab
	actorID[15] = 0x01

	assert.Equal(t, "ab000000000000000000000000000001", actorID.String())
}
