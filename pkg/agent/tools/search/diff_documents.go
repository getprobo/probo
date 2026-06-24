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

package search

import (
	"context"
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/agent"
)

type (
	diffParams struct {
		TextA  string `json:"text_a" jsonschema:"The first document text to compare"`
		TextB  string `json:"text_b" jsonschema:"The second document text to compare"`
		LabelA string `json:"label_a" jsonschema:"Label for the first document (e.g. 'current version')"`
		LabelB string `json:"label_b" jsonschema:"Label for the second document (e.g. 'archived version')"`
	}

	diffResult struct {
		HasDifferences bool   `json:"has_differences"`
		UnifiedDiff    string `json:"unified_diff,omitempty"`
		AddedLines     int    `json:"added_lines"`
		RemovedLines   int    `json:"removed_lines"`
		ErrorDetail    string `json:"error_detail,omitempty"`
	}
)

const (
	maxDiffOutput = 16000
)

func DiffDocumentsTool() agent.Tool {
	return agent.FunctionTool(
		"diff_documents",
		"Compare two document texts and return a unified diff showing the differences. Useful for comparing current vs. archived versions of privacy policies, terms of service, or other legal documents.",
		func(ctx context.Context, p diffParams) (agent.ToolResult, error) {
			labelA := p.LabelA
			if labelA == "" {
				labelA = "document_a"
			}

			labelB := p.LabelB
			if labelB == "" {
				labelB = "document_b"
			}

			linesA := strings.Split(p.TextA, "\n")
			linesB := strings.Split(p.TextB, "\n")

			diff := computeDiff(linesA, linesB, labelA, labelB)

			if diff.tooLarge {
				return agent.ResultJSON(
					diffResult{
						HasDifferences: true,
						ErrorDetail:    diff.output,
					},
				), nil
			}

			result := diffResult{
				HasDifferences: diff.added > 0 || diff.removed > 0,
				AddedLines:     diff.added,
				RemovedLines:   diff.removed,
			}

			if result.HasDifferences {
				output := diff.output
				if len(output) > maxDiffOutput {
					output = output[:maxDiffOutput] + "\n[... diff truncated]"
				}

				result.UnifiedDiff = output
			}

			return agent.ResultJSON(result), nil
		},
	)
}

type (
	diffOutput struct {
		output   string
		added    int
		removed  int
		tooLarge bool
	}
)

func computeDiff(linesA, linesB []string, labelA, labelB string) diffOutput {
	// Simple line-by-line LCS-based diff.
	m, n := len(linesA), len(linesB)

	// Build LCS table (bounded to prevent excessive memory for very large docs).
	if m > 5000 || n > 5000 {
		return diffOutput{
			output:   "documents too large for detailed diff (limit 5000 lines per side)",
			tooLarge: true,
		}
	}

	// LCS length table.
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := m - 1; i >= 0; i-- {
		for j := n - 1; j >= 0; j-- {
			if linesA[i] == linesB[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	// Walk the LCS table to produce diff hunks.
	var sb strings.Builder
	fmt.Fprintf(&sb, "--- %s\n+++ %s\n", labelA, labelB)

	var added, removed int

	i, j := 0, 0
	for i < m || j < n {
		if i < m && j < n && linesA[i] == linesB[j] {
			// Context line — only emit near changes.
			i++
			j++
		} else if j < n && (i >= m || dp[i][j+1] >= dp[i+1][j]) {
			sb.WriteString("+ " + linesB[j] + "\n")

			added++
			j++
		} else if i < m {
			sb.WriteString("- " + linesA[i] + "\n")

			removed++
			i++
		}
	}

	return diffOutput{
		output:  sb.String(),
		added:   added,
		removed: removed,
	}
}
