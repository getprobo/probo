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

package coredata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A representative payload shape. The coredata layer is schema-agnostic
// so any JSON-serialisable struct should round-trip.
type testAssessmentPayload struct {
	Summary    string   `json:"summary"`
	Confidence string   `json:"confidence"`
	Frameworks []string `json:"frameworks"`
}

func TestEvidence_SetAssessment_ReadBack(t *testing.T) {
	t.Parallel()

	in := &testAssessmentPayload{
		Summary:    "Google Workspace admin console showing enforced 2SV for all users.",
		Confidence: "HIGH",
		Frameworks: []string{"SOC2", "ISO27001"},
	}

	var e Evidence
	require.NoError(t, e.SetAssessment(in))
	require.NotEmpty(t, e.Assessment, "Assessment raw bytes should be populated")

	var out testAssessmentPayload
	require.NoError(t, e.GetAssessment(&out))
	assert.Equal(t, *in, out)
}

func TestEvidence_SetAssessment_Nil_ClearsField(t *testing.T) {
	t.Parallel()

	e := Evidence{}
	require.NoError(t, e.SetAssessment(&testAssessmentPayload{Summary: "stub"}))
	require.NotEmpty(t, e.Assessment)

	require.NoError(t, e.SetAssessment(nil))
	assert.Empty(t, e.Assessment, "untyped nil should clear the raw bytes")
}

func TestEvidence_SetAssessment_TypedNil_ClearsField(t *testing.T) {
	t.Parallel()

	e := Evidence{}
	require.NoError(t, e.SetAssessment(&testAssessmentPayload{Summary: "stub"}))
	require.NotEmpty(t, e.Assessment)

	var typedNil *testAssessmentPayload
	require.NoError(t, e.SetAssessment(typedNil))
	assert.Empty(t, e.Assessment, "typed-nil pointer should clear, not persist JSON null")
}

func TestEvidence_GetAssessment_EmptyIsNoOp(t *testing.T) {
	t.Parallel()

	var e Evidence

	out := testAssessmentPayload{Summary: "unchanged"}
	require.NoError(t, e.GetAssessment(&out))
	assert.Equal(t, "unchanged", out.Summary, "empty column should leave dst untouched")
}
