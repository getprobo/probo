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

package riskmanagement

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestWrapWords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			width:    28,
			expected: "",
		},
		{
			name:     "short label stays on one line",
			input:    "Store patient records",
			width:    28,
			expected: "Store patient records",
		},
		{
			name:     "exact width boundary does not wrap",
			input:    strings.Repeat("a", 28),
			width:    28,
			expected: strings.Repeat("a", 28),
		},
		{
			name:     "packs words up to width then breaks",
			input:    "Store patient conversation records",
			width:    28,
			expected: "Store patient conversation\nrecords",
		},
		{
			name:     "single oversized word is left intact",
			input:    "Supercalifragilisticexpialidocious",
			width:    28,
			expected: "Supercalifragilisticexpialidocious",
		},
		{
			name:     "non-ascii counted in runes",
			input:    "café café café café café café",
			width:    28,
			expected: "café café café café café\ncafé",
		},
		{
			name:     "non-positive width returns input",
			input:    "Store patient conversation records",
			width:    0,
			expected: "Store patient conversation records",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.expected, wrapWords(tt.input, tt.width))
			},
		)
	}
}

func TestEscapeMermaidLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ampersand uses html entity",
			input:    "Hold secrets & signing keys",
			expected: "Hold secrets &amp; signing keys",
		},
		{
			name:     "angle brackets use html entities",
			input:    "<conversation>",
			expected: "&lt;conversation&gt;",
		},
		{
			name:     "double quote uses mermaid entity",
			input:    `say "hello"`,
			expected: "say #quot;hello#quot;",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.expected, escapeMermaidLabel(tt.input))
			},
		)
	}
}

func TestMermaidLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short label unchanged",
			input:    "Store patient records",
			expected: "Store patient records",
		},
		{
			name:     "long label uses mermaid newline breaks",
			input:    "Store patient conversation records",
			expected: `Store patient conversation\nrecords`,
		},
		{
			name:     "wraps on visible text then escapes",
			input:    "Store patient <conversation> records now",
			expected: `Store patient &lt;conversation&gt;\nrecords now`,
		},
		{
			name:     "literal underscore is preserved",
			input:    "Data_Processing pipeline stage one two",
			expected: `Data_Processing pipeline\nstage one two`,
		},
		{
			name:     "ampersand keeps the existing html entity",
			input:    "Hold secrets & signing keys",
			expected: "Hold secrets &amp; signing keys",
		},
		{
			name:     "threat-style label wraps at width",
			input:    "Secret / signing-key leakage (liability)",
			expected: `Secret / signing-key leakage\n(liability)`,
		},
		{
			name:     "data node label wraps at width",
			input:    "Credentials, secrets, JWT signing keys",
			expected: `Credentials, secrets, JWT\nsigning keys`,
		},
		{
			name:     "crlf becomes a space",
			input:    "hello\r\nworld",
			expected: "hello world",
		},
		{
			name:     "bare lf becomes a space",
			input:    "hello\nworld",
			expected: "hello world",
		},
		{
			name:     "bare cr becomes a space",
			input:    "hello\rworld",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.expected, mermaidLabel(tt.input))
			},
		)
	}
}

func TestMermaidLabelWidth_ThreatHexagon(t *testing.T) {
	t.Parallel()

	got := mermaidLabelWidth(
		"Secret / signing-key leakage (liability)",
		mermaidThreatLabelWrapWidth,
	)

	assert.Equal(t, `Secret / signing-key\nleakage (liability)`, got)
}

func TestBuildDiagramMermaidChart_WrapsLabels(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	sourceID := gid.New(tenantID, coredata.RiskAnalysisNodeEntityType)
	targetID := gid.New(tenantID, coredata.RiskAnalysisNodeEntityType)
	processID := gid.New(tenantID, coredata.RiskAnalysisProcessEntityType)

	chart := buildDiagramMermaidChart(
		coredata.RiskAnalysisNodes{
			{
				ID:       sourceID,
				NodeType: coredata.RiskAnalysisNodeTypeAsset,
				Name:     "DynamoDB (stage-scoped table)",
			},
			{
				ID:       targetID,
				NodeType: coredata.RiskAnalysisNodeTypeData,
				Name:     "Credentials, secrets, JWT signing keys",
			},
		},
		nil,
		coredata.RiskAnalysisProcesses{
			{
				ID:           processID,
				SourceNodeID: sourceID,
				TargetNodeID: targetID,
				Name:         "Hold secrets & signing keys",
			},
		},
		coredata.RiskAnalysisThreats{
			{
				ProcessID: processID,
				Name:      "Secret / signing-key leakage",
				Category:  "liability",
			},
		},
	)

	require.NotEmpty(t, chart)
	assert.Contains(t, chart, "flowchart LR")
	assert.NotContains(t, chart, "htmlLabels:")
	assert.NotContains(t, chart, "wrappingWidth:")
	assert.Contains(t, chart, mermaidLabel("Credentials, secrets, JWT signing keys"))
	assert.Contains(t, chart, mermaidLabel("Hold secrets & signing keys"))
	assert.Contains(
		t,
		chart,
		mermaidLabelWidth(
			"Secret / signing-key leakage (liability)",
			mermaidThreatLabelWrapWidth,
		),
	)
	assert.Contains(t, chart, `\n`)
	assert.NotContains(t, chart, "<br>")
	assert.Contains(t, chart, "&amp;")
	assert.Contains(t, chart, ":::nodeAsset")
	assert.Contains(t, chart, ":::nodeData")
	assert.Contains(t, chart, ":::nodeThreat")
	assert.NotRegexp(t, `(?m)^\s*class `, chart)
}

func TestBuildDiagramMermaidChart_BoundaryDoesNotDuplicateInnerNode(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	boundaryID := gid.New(tenantID, coredata.RiskAnalysisBoundaryEntityType)
	nodeID := gid.New(tenantID, coredata.RiskAnalysisNodeEntityType)
	processID := gid.New(tenantID, coredata.RiskAnalysisProcessEntityType)
	sourceID := gid.New(tenantID, coredata.RiskAnalysisNodeEntityType)

	const label = "EU representative and supervisory authority"

	chart := buildDiagramMermaidChart(
		coredata.RiskAnalysisNodes{
			{
				ID:       sourceID,
				NodeType: coredata.RiskAnalysisNodeTypeAsset,
				Name:     "Probo",
			},
			{
				ID:         nodeID,
				BoundaryID: &boundaryID,
				NodeType:   coredata.RiskAnalysisNodeTypeEntity,
				Name:       label,
			},
		},
		coredata.RiskAnalysisBoundaries{
			{
				ID:   boundaryID,
				Name: label,
			},
		},
		coredata.RiskAnalysisProcesses{
			{
				ID:           processID,
				SourceNodeID: sourceID,
				TargetNodeID: nodeID,
				Name:         "Appoint representative",
			},
		},
		coredata.RiskAnalysisThreats{
			{
				ProcessID: processID,
				Name:      "No EU representative",
				Category:  "regulatory",
			},
		},
	)

	require.NotEmpty(t, chart)
	assert.Contains(t, chart, `subgraph b0["`+mermaidLabel(label)+`"]`)
	assert.Contains(t, chart, "direction TB")
	assert.Contains(t, chart, mermaidNodeShape(coredata.RiskAnalysisNodeTypeEntity, "n1", label))
	assert.Contains(t, chart, "  style b0 "+mermaidBoundaryStyle)
	assert.NotContains(t, chart, "class b0")
	assert.NotRegexp(t, `(?m)^\s*class `, chart)
}

func TestBuildDiagramMermaidChart_NestedBoundaryKeepsInnerDirection(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	parentID := gid.New(tenantID, coredata.RiskAnalysisBoundaryEntityType)
	childID := gid.New(tenantID, coredata.RiskAnalysisBoundaryEntityType)
	nodeID := gid.New(tenantID, coredata.RiskAnalysisNodeEntityType)

	chart := buildDiagramMermaidChart(
		coredata.RiskAnalysisNodes{
			{
				ID:         nodeID,
				BoundaryID: &childID,
				NodeType:   coredata.RiskAnalysisNodeTypeData,
				Name:       "Customer records",
			},
		},
		coredata.RiskAnalysisBoundaries{
			{
				ID:   parentID,
				Name: "EU",
			},
			{
				ID:               childID,
				ParentBoundaryID: &parentID,
				Name:             "Processors",
			},
		},
		nil,
		nil,
	)

	require.NotEmpty(t, chart)
	assert.Regexp(t, `(?m)^\s+subgraph b0\["EU"\]$`, chart)
	assert.Regexp(t, `(?m)^\s+subgraph b1\["Processors"\]$`, chart)
	assert.Equal(t, 2, strings.Count(chart, "direction TB"))
	assert.Contains(t, chart, "  style b0 "+mermaidBoundaryStyle)
	assert.Contains(t, chart, "  style b1 "+mermaidBoundaryStyle)
	assert.NotRegexp(t, `(?m)^\s*class `, chart)
}
