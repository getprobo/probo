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

package objectstorage

import (
	"math"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		fileSize  int64
		wantRange blob.HTTPRange
		wantError bool
	}{
		{
			name:      "closed range",
			value:     "bytes=3-7",
			fileSize:  12,
			wantRange: blob.HTTPRange{Offset: 3, Count: 5},
		},
		{
			name:      "open-ended range",
			value:     "bytes=3-",
			fileSize:  12,
			wantRange: blob.HTTPRange{Offset: 3},
		},
		{
			name:      "end at the integer limit",
			value:     "bytes=0-9223372036854775807",
			fileSize:  12,
			wantRange: blob.HTTPRange{Count: math.MaxInt64},
		},
		{
			name:      "suffix range",
			value:     "bytes=-4",
			fileSize:  12,
			wantRange: blob.HTTPRange{Offset: 8, Count: 4},
		},
		{
			name:      "suffix longer than object",
			value:     "bytes=-20",
			fileSize:  12,
			wantRange: blob.HTTPRange{Offset: 0, Count: 12},
		},
		{
			name:      "multiple ranges",
			value:     "bytes=1-2,4-5",
			fileSize:  12,
			wantError: true,
		},
		{
			name:      "invalid start",
			value:     "bytes=-1",
			fileSize:  0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseRange(tt.value, tt.fileSize)
			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantRange, got)
		})
	}
}

func TestAzureUploadOptions_MetadataNames(t *testing.T) {
	t.Parallel()

	got := azureUploadOptions(
		PutOptions{
			Metadata: map[string]string{
				"type":            "evidence",
				"organization-id": "org-1",
				"thirdParty-id":   "third-party-1",
			},
		},
	)

	assert.Equal(
		t,
		map[string]*string{
			"type":            new("evidence"),
			"organization_id": new("org-1"),
			"thirdParty_id":   new("third-party-1"),
		},
		got.Metadata,
	)
}
