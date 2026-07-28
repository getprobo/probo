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

package evidenceassessor

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew_DecoratesConfidenceEnum guards the schema-mutation trick used
// to inject an enum into the generated JSON schema. If jsonschema-go
// ever reshapes the "properties" block or stops preserving the
// confidence field path, this test will fail loudly instead of silently
// shipping an un-constrained schema.
func TestNew_DecoratesConfidenceEnum(t *testing.T) {
	t.Parallel()

	assessor, err := New(Config{})
	require.NoError(t, err)
	require.NotNil(t, assessor.outputType)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(assessor.outputType.Schema, &schema))

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

	assert.Equal(t, confidenceEnum, actual)
}

func TestEvidenceAssessment_validate(t *testing.T) {
	t.Parallel()

	valid := EvidenceAssessment{
		Summary:    "Google Workspace admin console showing enforced 2-step verification.",
		Confidence: "HIGH",
		Readable:   true,
	}
	require.NoError(t, valid.validate())

	assert.Error(t, EvidenceAssessment{Summary: "x", Confidence: "NOPE", Readable: true}.validate())
	assert.Error(t, EvidenceAssessment{Confidence: "HIGH", Readable: true}.validate())
	assert.Error(t, EvidenceAssessment{Summary: "x", Confidence: "LOW", Readable: false}.validate())
	require.NoError(t, EvidenceAssessment{
		Summary:         "Unable to describe: blank image.",
		Confidence:      "LOW",
		Readable:        false,
		RejectionReason: "blank",
	}.validate())
}
