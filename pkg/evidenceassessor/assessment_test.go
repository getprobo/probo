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

// White-box test (package evidenceassessor, not _test) so it can reach
// the unexported assessmentOutputType helper.

package evidenceassessor

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAssessmentOutputType_DecoratesConfidenceEnum guards the
// schema-mutation trick used to inject an enum into the generated JSON
// schema. If jsonschema-go ever reshapes the "properties" block or
// stops preserving the confidence field path, this test will fail
// loudly instead of silently shipping an un-constrained schema.
func TestAssessmentOutputType_DecoratesConfidenceEnum(t *testing.T) {
	t.Parallel()

	outputType, err := assessmentOutputType()
	require.NoError(t, err)
	require.NotNil(t, outputType)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(outputType.Schema, &schema))

	properties, ok := schema["properties"].(map[string]any)
	require.True(t, ok, "schema has no properties block")

	confidence, ok := properties["confidence"].(map[string]any)
	require.True(t, ok, "schema has no confidence property")

	enumRaw, ok := confidence["enum"].([]any)
	require.True(t, ok, "confidence has no enum array")

	actual := make([]string, len(enumRaw))
	for i, v := range enumRaw {
		actual[i] = v.(string)
	}
	assert.Equal(t, assessmentConfidenceEnum, actual)
}
