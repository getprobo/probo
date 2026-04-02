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

package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/agent"
)

type diffParams struct {
	TextA  string `json:"text_a" jsonschema:"The first document text to compare"`
	TextB  string `json:"text_b" jsonschema:"The second document text to compare"`
	LabelA string `json:"label_a" jsonschema:"Label for the first document (e.g. 'current version')"`
	LabelB string `json:"label_b" jsonschema:"Label for the second document (e.g. 'archived version')"`
}

type diffResult struct {
	HasDifferences bool   `json:"has_differences"`
	UnifiedDiff    string `json:"unified_diff,omitempty"`
	AddedLines     int    `json:"added_lines"`
	RemovedLines   int    `json:"removed_lines"`
	ErrorDetail    string `json:"error_detail,omitempty"`
}

const maxDiffOutput = 16000

func DiffDocumentsTool() (agent.Tool, error) {
	return agent.FunctionTool[diffParams](
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

			data, _ := json.Marshal(result)
			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}

type diffOutput struct {
	output  string
	added   int
	removed int
}

func computeDiff(linesA, linesB []string, labelA, labelB string) diffOutput {
	// Simple line-by-line LCS-based diff.
	m, n := len(linesA), len(linesB)

	// Build LCS table (bounded to prevent excessive memory for very large docs).
	if m > 5000 || n > 5000 {
		return diffOutput{
			output:  "[documents too large for detailed diff]",
			added:   0,
			removed: 0,
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
	sb.WriteString(fmt.Sprintf("--- %s\n+++ %s\n", labelA, labelB))

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
