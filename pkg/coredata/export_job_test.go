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

package coredata_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestDocumentExportArguments_MarshalWatermarkText(t *testing.T) {
	t.Parallel()

	arguments := coredata.DocumentExportArguments{
		WithWatermark:  true,
		WatermarkText:  new("Internal use only"),
		WithSignatures: true,
	}

	data, err := json.Marshal(arguments)

	require.NoError(t, err)
	assert.JSONEq(
		t,
		`{
			"document_ids": null,
			"with_watermark": true,
			"watermark_text": "Internal use only",
			"with_signatures": true
		}`,
		string(data),
	)
}

func TestDocumentExportArguments_UnmarshalLegacyWatermarkEmail(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"document_ids": [],
		"with_watermark": true,
		"watermark_email": "recipient@example.com",
		"with_signatures": false
	}`)
	var arguments coredata.DocumentExportArguments

	err := json.Unmarshal(data, &arguments)

	require.NoError(t, err)
	require.NotNil(t, arguments.WatermarkText)
	assert.Equal(t, "recipient@example.com", *arguments.WatermarkText)
}
