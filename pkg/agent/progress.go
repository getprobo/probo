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

package agent

import "context"

type (
	ProgressEventType string

	ProgressEvent struct {
		Type       ProgressEventType `json:"type"`
		Step       string            `json:"step"`
		ParentStep string            `json:"parent_step,omitempty"`
		Message    string            `json:"message"`
	}

	ProgressReporter func(ctx context.Context, event ProgressEvent)
)

const (
	ProgressEventStepStarted   ProgressEventType = "step_started"
	ProgressEventStepCompleted ProgressEventType = "step_completed"
	ProgressEventStepFailed    ProgressEventType = "step_failed"
)
