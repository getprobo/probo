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

package coredata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvidence_SetAndGetAssessment_RoundTrip(t *testing.T) {
	t.Parallel()

	in := &EvidenceAssessment{
		Summary:    "Google Workspace admin console showing enforced 2SV for all users.",
		System:     "Google Workspace",
		Setting:    "enforced 2-step verification",
		Scope:      "organization-wide",
		Language:   "en",
		Frameworks: []string{"SOC2", "ISO27001"},
		Issues:     []string{},
		Confidence: "HIGH",
		Readable:   true,
	}

	var e Evidence
	require.NoError(t, e.SetAssessment(in))
	require.NotEmpty(t, e.Assessment, "Assessment raw bytes should be populated")

	out, err := e.GetAssessment()
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, in, out)
}

func TestEvidence_SetAssessment_Nil_ClearsField(t *testing.T) {
	t.Parallel()

	e := Evidence{}
	require.NoError(t, e.SetAssessment(&EvidenceAssessment{Summary: "stub"}))
	require.NotEmpty(t, e.Assessment)

	require.NoError(t, e.SetAssessment(nil))
	assert.Empty(t, e.Assessment, "nil input should clear the raw bytes")
}

func TestEvidence_GetAssessment_EmptyReturnsNil(t *testing.T) {
	t.Parallel()

	var e Evidence
	got, err := e.GetAssessment()
	require.NoError(t, err)
	assert.Nil(t, got)
}
