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

package probo_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/docgen"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/prosemirror"
)

func TestBuildThirdPartyListDocument_RendersMarkdownNotes(t *testing.T) {
	t.Parallel()

	notes := `# Third Party Assessment: Nango

## Executive Summary
Approve with conditions.

## Overall Risk Score
| Category | Weight | Score (0-100) | Weighted |
|----------|--------|---------------|----------|
| Security | 40% | 72 | 28.8 |

## Market Presence
- **Notable Customers**: Acme
- **Company Size Signals**: Series B
- **Market Position**: Niche leader
`

	raw, err := probo.BuildThirdPartyListDocument(
		docgen.ThirdPartyListData{
			Title:             "ThirdParties",
			OrganizationName:  "Probo",
			TotalThirdParties: 1,
			Rows: []docgen.ThirdPartyListRow{
				{
					Name: "Nango",
					RiskAssessments: []docgen.ThirdPartyListRiskAssessment{
						{
							AssessedAt:      "2026-07-31",
							ExpiresAt:       "2027-07-31",
							DataSensitivity: "High",
							BusinessImpact:  "Medium",
							Notes:           notes,
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)

	doc, err := prosemirror.Parse(raw)
	require.NoError(t, err)

	// Markdown syntax must not survive as literal text, regardless of JSON
	// whitespace around "text" keys.
	assert.NotContains(t, raw, "# Third Party Assessment")
	assert.NotContains(t, raw, "## Executive Summary")
	assert.NotContains(t, raw, "| Category")
	assert.NotContains(t, raw, "- **Notable Customers**")
	assert.False(t, anyTextNodeContains(doc, "**"))
	assert.False(t, anyTextNodeHasPrefix(doc, "#"))
	assert.False(t, anyTextNodeHasPrefix(doc, "|"))

	assert.True(t, containsHeading(doc, 4, "Third Party Assessment: Nango"))
	assert.True(t, containsHeading(doc, 5, "Executive Summary"))
	assert.True(t, containsHeading(doc, 5, "Overall Risk Score"))
	assert.True(t, containsHeading(doc, 5, "Market Presence"))
	assert.True(t, containsNodeType(doc, prosemirror.NodeTable))
	assert.True(t, containsBoldText(doc, "Notable Customers"))
	assert.Contains(t, collectText(doc), "Approve with conditions.")
}

func TestBuildThirdPartyListDocument_PlainNotesFallback(t *testing.T) {
	t.Parallel()

	raw, err := probo.BuildThirdPartyListDocument(
		docgen.ThirdPartyListData{
			Title:             "ThirdParties",
			OrganizationName:  "Probo",
			TotalThirdParties: 1,
			Rows: []docgen.ThirdPartyListRow{
				{
					Name: "Acme",
					RiskAssessments: []docgen.ThirdPartyListRiskAssessment{
						{
							AssessedAt:      "2026-07-31",
							ExpiresAt:       "2027-07-31",
							DataSensitivity: "Low",
							BusinessImpact:  "Low",
							Notes:           "Not specified",
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)

	doc, err := prosemirror.Parse(raw)
	require.NoError(t, err)
	assert.Contains(t, collectText(doc), "Notes: Not specified")
}

func TestBuildThirdPartyListDocument_EmptyNotesPlaceholder(t *testing.T) {
	t.Parallel()

	raw, err := probo.BuildThirdPartyListDocument(
		docgen.ThirdPartyListData{
			Title:             "ThirdParties",
			OrganizationName:  "Probo",
			TotalThirdParties: 1,
			Rows: []docgen.ThirdPartyListRow{
				{
					Name: "Acme",
					RiskAssessments: []docgen.ThirdPartyListRiskAssessment{
						{
							AssessedAt:      "2026-07-31",
							ExpiresAt:       "2027-07-31",
							DataSensitivity: "Low",
							BusinessImpact:  "Low",
							Notes:           "",
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)

	doc, err := prosemirror.Parse(raw)
	require.NoError(t, err)
	assert.Contains(t, collectText(doc), "Notes: —")
}

func TestBuildThirdPartyListDocument_EscapesNotesJSON(t *testing.T) {
	t.Parallel()

	raw, err := probo.BuildThirdPartyListDocument(
		docgen.ThirdPartyListData{
			Title:             "ThirdParties",
			OrganizationName:  "Probo",
			TotalThirdParties: 1,
			Rows: []docgen.ThirdPartyListRow{
				{
					Name: `Vendor "Quotes"`,
					RiskAssessments: []docgen.ThirdPartyListRiskAssessment{
						{
							AssessedAt:      "2026-07-31",
							ExpiresAt:       "2027-07-31",
							DataSensitivity: "Low",
							BusinessImpact:  "Low",
							Notes:           "Line with \"quotes\" and\nnewlines",
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)
	require.True(t, json.Valid([]byte(raw)))

	doc, err := prosemirror.Parse(raw)
	require.NoError(t, err)
	assert.Contains(t, collectText(doc), `Line with "quotes" and newlines`)
}

func containsHeading(n prosemirror.Node, level int, text string) bool {
	if n.Type == prosemirror.NodeHeading {
		attrs, err := n.HeadingAttrs()
		if err == nil && attrs.Level == level && collectText(n) == text {
			return true
		}
	}

	for _, child := range n.Content {
		if containsHeading(child, level, text) {
			return true
		}
	}

	return false
}

func containsNodeType(n prosemirror.Node, nodeType prosemirror.NodeType) bool {
	if n.Type == nodeType {
		return true
	}

	for _, child := range n.Content {
		if containsNodeType(child, nodeType) {
			return true
		}
	}

	return false
}

func containsBoldText(n prosemirror.Node, text string) bool {
	if n.Type == prosemirror.NodeText && n.Text != nil && *n.Text == text {
		for _, mark := range n.Marks {
			if mark.Type == prosemirror.MarkStrong {
				return true
			}
		}
	}

	for _, child := range n.Content {
		if containsBoldText(child, text) {
			return true
		}
	}

	return false
}

func anyTextNodeContains(n prosemirror.Node, substr string) bool {
	if n.Type == prosemirror.NodeText && n.Text != nil && strings.Contains(*n.Text, substr) {
		return true
	}

	for _, child := range n.Content {
		if anyTextNodeContains(child, substr) {
			return true
		}
	}

	return false
}

func anyTextNodeHasPrefix(n prosemirror.Node, prefix string) bool {
	if n.Type == prosemirror.NodeText && n.Text != nil && strings.HasPrefix(strings.TrimSpace(*n.Text), prefix) {
		return true
	}

	for _, child := range n.Content {
		if anyTextNodeHasPrefix(child, prefix) {
			return true
		}
	}

	return false
}

func collectText(n prosemirror.Node) string {
	if n.Text != nil {
		return *n.Text
	}

	var out strings.Builder
	for _, child := range n.Content {
		out.WriteString(collectText(child))
	}

	return out.String()
}
