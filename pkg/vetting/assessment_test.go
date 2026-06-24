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

// This test file is white-box (package vetting, not vetting_test) so it
// can reach the unexported thirdPartyInfoOutputType helper.

package vetting

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThirdPartyInfoOutputType_DecoratesEnums(t *testing.T) {
	t.Parallel()

	outputType, err := thirdPartyInfoOutputType()
	require.NoError(t, err)
	require.NotNil(t, outputType)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(outputType.Schema, &schema))

	properties, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	tests := []struct {
		field    string
		expected []string
	}{
		{"category", thirdPartyCategoryEnum},
		{"third_party_type", thirdPartyTypeEnum},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			t.Parallel()

			prop, ok := properties[tt.field].(map[string]any)
			require.True(t, ok, "schema has no %q property", tt.field)

			enumRaw, ok := prop["enum"].([]any)
			require.True(t, ok, "%q has no enum array", tt.field)

			actual := make([]string, len(enumRaw))
			for i, v := range enumRaw {
				actual[i] = v.(string)
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}
