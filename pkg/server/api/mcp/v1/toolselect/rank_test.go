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

package toolselect_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/toolselect"
)

func TestRank_DocumentScenarios(t *testing.T) {
	t.Parallel()

	tools, err := toolselect.LoadSpec(filepath.Join("..", "specification.yaml"))
	require.NoError(t, err)

	guidedWrong := 0
	baselineWrong := 0
	var report strings.Builder
	fmt.Fprintf(&report, "document tool selection\nscenarios: %d\ntools: %d\n\n", len(toolselect.DocumentScenarios), len(tools))

	for _, scenario := range toolselect.DocumentScenarios {
		guided := toolselect.Rank(tools, scenario.Task, true)
		baseline := toolselect.Rank(tools, scenario.Task, false)
		require.NotEmpty(t, guided)
		require.NotEmpty(t, baseline)

		guidedHit := guided[0].Name == scenario.Want
		baselineHit := baseline[0].Name == scenario.Want
		if !guidedHit {
			guidedWrong++
		}

		if !baselineHit {
			baselineWrong++
		}

		fmt.Fprintf(
			&report,
			"task: %s\nwant: %s\nguided: %s (%.2f) runner-up: %s\nbaseline: %s (%.2f)\n\n",
			scenario.Task,
			scenario.Want,
			guided[0].Name,
			guided[0].Score,
			runnerUp(guided),
			baseline[0].Name,
			baseline[0].Score,
		)
	}

	guidedRate := float64(guidedWrong) / float64(len(toolselect.DocumentScenarios))
	baselineRate := float64(baselineWrong) / float64(len(toolselect.DocumentScenarios))
	fmt.Fprintf(
		&report,
		"wrong-tool rate\nguided (name + description): %.2f (%d/%d)\nname-only baseline: %.2f (%d/%d)\n",
		guidedRate,
		guidedWrong,
		len(toolselect.DocumentScenarios),
		baselineRate,
		baselineWrong,
		len(toolselect.DocumentScenarios),
	)

	writeReport(t, report.String())
	t.Log(report.String())

	assert.Equal(t, 0, guidedWrong)
	assert.Greater(t, baselineWrong, guidedWrong)
}

func TestSpec_PublishDocumentOmitsApproverIDs(t *testing.T) {
	t.Parallel()

	tools, err := toolselect.LoadSpec(filepath.Join("..", "specification.yaml"))
	require.NoError(t, err)

	var publish *toolselect.Tool
	for i := range tools {
		if tools[i].Name == "publishDocument" {
			publish = &tools[i]
			break
		}
	}

	require.NotNil(t, publish)
	assert.Contains(t, publish.Description, "has no approver_ids")
	assert.Contains(t, publish.Description, "default approvers")
	assert.Contains(t, publish.Description, "minor")
}

func runnerUp(hits []toolselect.Hit) string {
	if len(hits) < 2 {
		return ""
	}

	return hits[1].Name
}

func writeReport(t *testing.T, report string) {
	t.Helper()

	dir := "/opt/cursor/artifacts"
	if _, err := os.Stat(dir); err != nil {
		return
	}

	path := filepath.Join(dir, "mcp-document-tool-selection.txt")
	if err := os.WriteFile(path, []byte(report), 0o644); err != nil {
		t.Logf("cannot write selection report: %v", err)
	}
}
