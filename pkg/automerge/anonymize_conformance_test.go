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
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	productionautomerge "go.probo.inc/probo/pkg/automerge"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

func TestConformance_AnonymizedDocumentLoadsAcrossEngines(t *testing.T) {
	t.Parallel()

	source, err := productionautomerge.New(actor(203))
	require.NoError(t, err)
	closeDocument(t, source)
	require.NoError(
		t,
		source.PutScalar(
			"owner-email",
			productionautomerge.StringScalar("alice@example.com"),
		),
	)
	text, err := source.CreateText("confidential-policy")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "secret contents"))
	_, err = source.Commit("private metadata", commitTime)
	require.NoError(t, err)

	anonymized, err := source.Anonymize()
	require.NoError(t, err)
	closeDocument(t, anonymized)
	data, err := anonymized.Save()
	require.NoError(t, err)

	t.Run("rust", func(t *testing.T) {
		reference, err := automerge.LoadReference(data, actor(204))
		require.NoError(t, err)
		closeDocument(t, reference)

		heads, err := reference.Heads()
		require.NoError(t, err)
		assert.Len(t, heads, 1)
	})

	t.Run("javascript", func(t *testing.T) {
		response := runOracle(
			t,
			oracleRequest{
				Action:   "inspectDataModel",
				Document: base64.StdEncoding.EncodeToString(data),
			},
		)
		require.Len(t, response.Heads, 1)

		encoded, err := json.Marshal(response.Data)
		require.NoError(t, err)
		materialized := string(encoded)
		assert.False(t, strings.Contains(materialized, "owner-email"))
		assert.False(t, strings.Contains(materialized, "alice@example.com"))
		assert.False(t, strings.Contains(materialized, "confidential-policy"))
		assert.False(t, strings.Contains(materialized, "secret contents"))
	})
}
