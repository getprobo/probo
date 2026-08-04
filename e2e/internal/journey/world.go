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

// Package journey provides readable, diagnostic steps for full end-to-end
// scenarios. It orchestrates the existing black-box test clients; it does not
// bypass HTTP, authentication, email delivery, or persistence.
package journey

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type (
	// World records one isolated end-to-end scenario.
	World struct {
		t         *testing.T
		startedAt time.Time

		mu     sync.Mutex
		actors []actorRecord
		steps  []StepRecord
	}

	// StepRecord describes one user-visible action in a journey.
	StepRecord struct {
		Number    int           `json:"number"`
		Actor     string        `json:"actor,omitempty"`
		Name      string        `json:"name"`
		StartedAt time.Time     `json:"startedAt"`
		Duration  time.Duration `json:"duration"`
		Failure   string        `json:"failure,omitempty"`
	}

	actorRecord struct {
		Name           string `json:"name"`
		Role           string `json:"role"`
		UserID         string `json:"userId,omitempty"`
		OrganizationID string `json:"organizationId,omitempty"`
	}
)

// New creates a scenario world and registers failure artifact collection.
func New(t *testing.T) *World {
	t.Helper()

	w := &World{
		t:         t,
		startedAt: time.Now(),
	}

	t.Cleanup(func() {
		if !t.Failed() {
			return
		}

		artifactDir, err := w.writeFailureArtifacts()
		if err != nil {
			t.Logf("journey: cannot write failure artifacts: %v", err)
			return
		}

		t.Logf("journey: failure artifacts: %s", artifactDir)
	})

	return w
}

// Step executes a named action and records enough context to identify where a
// journey stopped. Steps within one world are intentionally sequential.
func (w *World) Step(actor string, name string, fn func() error) {
	w.t.Helper()

	number := w.nextStepNumber()
	startedAt := time.Now()
	failedBefore := w.t.Failed()
	record := StepRecord{
		Number:    number,
		Actor:     actor,
		Name:      name,
		StartedAt: startedAt,
	}

	w.t.Logf("journey: step %02d started: %s", number, describeStep(actor, name))

	var stepErr error
	defer func() {
		record.Duration = time.Since(startedAt)
		if stepErr != nil {
			record.Failure = stepErr.Error()
		} else if !failedBefore && w.t.Failed() {
			record.Failure = "an assertion failed or the test stopped during this step"
		}

		w.appendStep(record)

		if record.Failure == "" {
			w.t.Logf(
				"journey: step %02d passed in %s: %s",
				number,
				record.Duration.Round(time.Millisecond),
				describeStep(actor, name),
			)
		} else {
			w.t.Logf(
				"journey: step %02d failed in %s: %s: %s",
				number,
				record.Duration.Round(time.Millisecond),
				describeStep(actor, name),
				record.Failure,
			)
		}
	}()

	stepErr = fn()
	if stepErr != nil {
		w.t.Fatalf(
			"journey failed at step %02d (%s): %v",
			number,
			describeStep(actor, name),
			stepErr,
		)
	}
}

func (w *World) nextStepNumber() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return len(w.steps) + 1
}

func (w *World) appendStep(record StepRecord) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.steps = append(w.steps, record)
}

func (w *World) registerActor(record actorRecord) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.actors = append(w.actors, record)
}

func describeStep(actor string, name string) string {
	if actor == "" {
		return name
	}

	return fmt.Sprintf("%s %s", actor, name)
}
