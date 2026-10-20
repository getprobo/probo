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

package testsupport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodePatches_IncrementValue(t *testing.T) {
	t.Parallel()

	patches, err := decodePatches(
		[]byte(`[{"obj":"_root","action":{"type":"increment","value":-42}}]`),
	)
	require.NoError(t, err)
	require.Len(t, patches, 1)
	assert.Equal(t, PatchIncrement, patches[0].Action)
	assert.Equal(t, int64(-42), patches[0].Delta)
}

func TestDecodePatches_RejectsInvalidIncrementValue(t *testing.T) {
	t.Parallel()

	for name, data := range map[string]string{
		"missing":    `[{"obj":"_root","action":{"type":"increment"}}]`,
		"null":       `[{"obj":"_root","action":{"type":"increment","value":null}}]`,
		"fractional": `[{"obj":"_root","action":{"type":"increment","value":1.5}}]`,
		"overflow":   `[{"obj":"_root","action":{"type":"increment","value":9223372036854775808}}]`,
	} {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				_, err := decodePatches([]byte(data))
				require.Error(t, err)
			},
		)
	}
}
