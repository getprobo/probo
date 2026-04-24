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
	"encoding/json"
	"fmt"
)

// EvidenceAssessment is the structured output produced by the evidence
// describer agent. It is persisted as JSONB in evidences.assessment and
// is the canonical form; Evidence.Description is a one-sentence summary
// derived from Assessment.Summary.
type EvidenceAssessment struct {
	Summary         string   `json:"summary"          jsonschema:"One plain-text sentence (two at most) summarising what the evidence shows; no markdown, no preamble. When readable is false this field must restate the rejection reason so downstream systems displaying only the summary still inform the user."`
	System          string   `json:"system"           jsonschema:"Tool or platform shown (e.g. 'Google Workspace', 'GitHub', 'AWS IAM'); empty string if not identifiable."`
	Setting         string   `json:"setting"          jsonschema:"Specific configuration, feature, or state demonstrated; empty string if not identifiable."`
	Scope           string   `json:"scope"            jsonschema:"Who or what the evidence applies to (e.g. 'organization-wide', 'all users', 'repository acme/foo'); empty string if not stated."`
	CapturedAt      string   `json:"captured_at"      jsonschema:"ISO-8601 date or datetime visible on the file; empty string if no date is shown."`
	Language        string   `json:"language"         jsonschema:"BCP-47 language tag of the visible text (e.g. 'en', 'fr'); empty string if unclear."`
	Frameworks      []string `json:"frameworks"       jsonschema:"Compliance frameworks the evidence is plausibly relevant to (e.g. 'SOC2', 'ISO27001'); empty when unclear."`
	Issues          []string `json:"issues"           jsonschema:"Quality problems observed: redacted fields, crop, stale date, low resolution, etc."`
	Confidence      string   `json:"confidence"       jsonschema:"Overall confidence in the System / Setting / Scope identification; one of HIGH, MEDIUM, LOW."`
	Readable        bool     `json:"readable"         jsonschema:"True if the file is readable compliance evidence; false if unreadable or off-topic."`
	RejectionReason string   `json:"rejection_reason" jsonschema:"Set only when readable is false; explains why the file is not usable evidence."`
}

// EvidenceAssessmentConfidenceEnum is the canonical set of allowed
// EvidenceAssessment.Confidence values. Kept in a package-level slice
// because jsonschema struct tags are free-form descriptions and cannot
// encode enum constraints directly; consumers building an output schema
// inject it post-marshal (see pkg/evidencedescriber).
var EvidenceAssessmentConfidenceEnum = []string{"HIGH", "MEDIUM", "LOW"}

// SetAssessment marshals a typed assessment into Evidence.Assessment.
// Passing nil clears the field. This is the only supported way to write
// the assessment column from outside the coredata package; the backing
// field type is unexported.
func (e *Evidence) SetAssessment(a *EvidenceAssessment) error {
	if a == nil {
		e.Assessment = nil
		return nil
	}
	data, err := json.Marshal(a)
	if err != nil {
		return fmt.Errorf("cannot marshal evidence assessment: %w", err)
	}
	e.Assessment = data
	return nil
}

// GetAssessment unmarshals Evidence.Assessment into a typed
// EvidenceAssessment. Returns (nil, nil) when the column is NULL.
func (e *Evidence) GetAssessment() (*EvidenceAssessment, error) {
	if len(e.Assessment) == 0 {
		return nil, nil
	}
	var a EvidenceAssessment
	if err := json.Unmarshal(e.Assessment, &a); err != nil {
		return nil, fmt.Errorf("cannot unmarshal evidence assessment: %w", err)
	}
	return &a, nil
}
