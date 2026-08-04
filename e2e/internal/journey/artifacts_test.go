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

package journey

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactSensitiveText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "token", input: `token="private-value"`},
		{name: "password", input: `password: hunter2`},
		{name: "authorization", input: `Authorization=Bearer-private-value`},
		{name: "API key", input: `api_key:private-value`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual := redactSensitiveText(tt.input)

			assert.Contains(t, actual, "[REDACTED]")
			assert.NotContains(t, actual, "private-value")
			assert.NotContains(t, actual, "hunter2")
		})
	}
}

func TestFormatFailureSummary(t *testing.T) {
	t.Parallel()

	summary := formatFailureSummary(
		[]StepRecord{
			{
				Number:   1,
				Actor:    "Alice",
				Name:     "creates a document",
				Duration: 12 * time.Millisecond,
			},
			{
				Number:   2,
				Actor:    "Bob",
				Name:     "approves the document",
				Duration: 8 * time.Millisecond,
				Failure:  "token=private-value expired",
			},
		},
	)

	assert.Equal(
		t,
		"PASS step 01 [Alice] creates a document (12ms)\n"+
			"FAIL step 02 [Bob] approves the document (8ms)\n"+
			"  token=[REDACTED] expired\n",
		summary,
	)
	assert.False(t, strings.Contains(summary, "private-value"))
}

func TestSanitizePathSegment(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "TestJourney_step_name", sanitizePathSegment("TestJourney/step name"))
	assert.Equal(t, "journey", sanitizePathSegment("///"))
}

func TestWorld_WriteFailureArtifacts(t *testing.T) {
	t.Parallel()

	artifactRoot := t.TempDir()

	world := &World{
		t:         t,
		startedAt: time.Now(),
		actors: []actorRecord{
			{Name: "Alice", Role: "OWNER", UserID: "user-id"},
		},
		steps: []StepRecord{
			{
				Number:   1,
				Actor:    "Alice",
				Name:     "creates a document",
				Duration: time.Millisecond,
				Failure:  "token=private-value expired",
			},
		},
	}

	directory, err := world.writeFailureArtifactsTo(artifactRoot)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(directory, artifactRoot))

	manifest, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	require.NoError(t, err)
	assert.Contains(t, string(manifest), `"test": "TestWorld_WriteFailureArtifacts"`)
	assert.Contains(t, string(manifest), `"failure": "token=[REDACTED] expired"`)
	assert.NotContains(t, string(manifest), "private-value")

	summary, err := os.ReadFile(filepath.Join(directory, "failure.txt"))
	require.NoError(t, err)
	assert.Contains(t, string(summary), "FAIL step 01 [Alice] creates a document")
	assert.NotContains(t, string(summary), "private-value")
}
