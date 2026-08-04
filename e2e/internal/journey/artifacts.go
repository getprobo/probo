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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
)

const artifactDirectoryEnvironmentVariable = "PROBO_E2E_ARTIFACT_DIR"

var sensitiveValuePattern = regexp.MustCompile(
	`(?i)(password|token|secret|authorization|cookie|api[_-]?key)([[:space:]]*["':=]+[[:space:]]*)([^,}\]\s"']+)`,
)

type failureManifest struct {
	Test      string        `json:"test"`
	StartedAt time.Time     `json:"startedAt"`
	FailedAt  time.Time     `json:"failedAt"`
	Actors    []actorRecord `json:"actors"`
	Steps     []StepRecord  `json:"steps"`
}

func (w *World) writeFailureArtifacts() (string, error) {
	root := os.Getenv(artifactDirectoryEnvironmentVariable)
	if root == "" {
		root = filepath.Join(os.TempDir(), "probo-e2e-artifacts")
	}

	return w.writeFailureArtifactsTo(root)
}

func (w *World) writeFailureArtifactsTo(root string) (string, error) {
	directory := filepath.Join(
		root,
		sanitizePathSegment(w.t.Name()),
		fmt.Sprintf("%d", time.Now().UnixNano()),
	)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", fmt.Errorf("cannot create artifact directory: %w", err)
	}

	w.mu.Lock()
	actors := append([]actorRecord(nil), w.actors...)
	steps := append([]StepRecord(nil), w.steps...)
	w.mu.Unlock()

	for i := range steps {
		steps[i].Failure = redactSensitiveText(steps[i].Failure)
	}

	manifest := failureManifest{
		Test:      w.t.Name(),
		StartedAt: w.startedAt,
		FailedAt:  time.Now(),
		Actors:    actors,
		Steps:     steps,
	}

	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("cannot encode journey manifest: %w", err)
	}

	if err := os.WriteFile(
		filepath.Join(directory, "manifest.json"),
		append(manifestJSON, '\n'),
		0o600,
	); err != nil {
		return "", fmt.Errorf("cannot write journey manifest: %w", err)
	}

	if err := os.WriteFile(
		filepath.Join(directory, "failure.txt"),
		[]byte(formatFailureSummary(steps)),
		0o600,
	); err != nil {
		return "", fmt.Errorf("cannot write journey failure summary: %w", err)
	}

	return directory, nil
}

func formatFailureSummary(steps []StepRecord) string {
	var b strings.Builder

	for _, step := range steps {
		status := "PASS"
		if step.Failure != "" {
			status = "FAIL"
		}

		fmt.Fprintf(
			&b,
			"%s step %02d [%s] %s (%s)\n",
			status,
			step.Number,
			step.Actor,
			step.Name,
			step.Duration.Round(time.Millisecond),
		)

		if step.Failure != "" {
			fmt.Fprintf(&b, "  %s\n", redactSensitiveText(step.Failure))
		}
	}

	return b.String()
}

func sanitizePathSegment(value string) string {
	value = strings.Map(
		func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
				return r
			}

			return '_'
		},
		value,
	)

	value = strings.Trim(value, "_")
	if value == "" {
		return "journey"
	}

	return value
}

func redactSensitiveText(value string) string {
	return sensitiveValuePattern.ReplaceAllString(value, "$1$2[REDACTED]")
}
