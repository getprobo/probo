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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func TestDocument_AnonymizePreservesPendingSource(t *testing.T) {
	t.Parallel()

	document, err := automerge.New(automerge.ActorID{1})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, document.Close())
	})
	require.NoError(t, document.PutString("private-key", "private value"))

	anonymized, err := document.Anonymize()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, anonymized.Close())
	})

	sourceValue, err := document.String("private-key")
	require.NoError(t, err)
	assert.Equal(t, "private value", sourceValue)

	cancelled, err := document.Rollback()
	require.NoError(t, err)
	assert.Equal(t, uint64(1), cancelled)

	keys, err := anonymized.Root().Keys()
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.NotEqual(t, "private-key", keys[0])
	value, err := anonymized.Root().Scalar(keys[0])
	require.NoError(t, err)
	assert.NotEqual(t, "private value", value.String)
	assert.Len(t, value.String, len("private value"))
}

func TestDocument_AnonymizeClosedDocument(t *testing.T) {
	t.Parallel()

	document, err := automerge.New(automerge.ActorID{1})
	require.NoError(t, err)
	require.NoError(t, document.Close())

	anonymized, err := document.Anonymize()
	assert.Nil(t, anonymized)
	assert.True(t, errors.Is(err, automerge.ErrClosed))
}
