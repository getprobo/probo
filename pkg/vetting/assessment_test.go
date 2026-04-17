// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

// This test file is white-box (package vetting, not vetting_test) so it
// can reach the unexported vendorInfoOutputType helper.

package vetting

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVendorInfoOutputType_DecoratesEnums(t *testing.T) {
	t.Parallel()

	outputType, err := vendorInfoOutputType()
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
		{"category", vendorCategoryEnum},
		{"vendor_type", vendorTypeEnum},
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
