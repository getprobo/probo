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

package vetting

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONSchemaForTool_EnforcesOpenAIStrictMode(t *testing.T) {
	t.Parallel()

	raw, err := jsonSchemaForTool[saveThirdPartyInfoToolParams]()
	require.NoError(t, err)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))

	required := schema["required"].([]any)
	assert.Contains(t, required, "name")
	assert.Contains(t, required, "description")
	assert.Equal(t, false, schema["additionalProperties"])
}

func TestNewVettingOutputType_EnforcesOpenAIStrictMode(t *testing.T) {
	t.Parallel()

	outputType, err := newVettingOutputType[CrawlerOutput]("crawler")
	require.NoError(t, err)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(outputType.Schema, &schema))

	assert.Equal(t, false, schema["additionalProperties"])
	assert.NotEmpty(t, schema["required"])
}
