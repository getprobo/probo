// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package commontrackerpattern

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
)

func TestEnrichmentState(t *testing.T) {
	t.Parallel()

	requestedAt := time.Now()

	tests := []struct {
		name     string
		pattern  coredata.CommonTrackerPattern
		expected string
	}{
		{
			name:     "no payload and not queued is unenriched",
			pattern:  coredata.CommonTrackerPattern{},
			expected: "unenriched",
		},
		{
			name:     "queued takes precedence",
			pattern:  coredata.CommonTrackerPattern{EnrichmentRequestedAt: &requestedAt},
			expected: "queued",
		},
		{
			name:     "payload without fields reads enriched",
			pattern:  coredata.CommonTrackerPattern{Enrichment: json.RawMessage(`{"status":"migrated"}`)},
			expected: "enriched",
		},
		{
			name: "all fields resolved reads enriched",
			pattern: coredata.CommonTrackerPattern{
				Enrichment: json.RawMessage(`{"status":"done","fields":{"description":{"status":"found"},"third_party":{"status":"exists_external"}}}`),
			},
			expected: "enriched",
		},
		{
			name: "some fields resolved reads partial",
			pattern: coredata.CommonTrackerPattern{
				Enrichment: json.RawMessage(`{"status":"partial","fields":{"description":{"status":"found"},"third_party":{"status":"not_found"}}}`),
			},
			expected: "partial (1/2)",
		},
		{
			name: "no fields resolved reads partial zero",
			pattern: coredata.CommonTrackerPattern{
				Enrichment: json.RawMessage(`{"status":"no_result","fields":{"description":{"status":"not_found"},"third_party":{"status":"not_found"}}}`),
			},
			expected: "partial (0/2)",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.expected, enrichmentState(&tt.pattern))
			},
		)
	}
}
