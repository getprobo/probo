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
