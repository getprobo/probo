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

package mcp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMCP_RiskAnalysis_Fork(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	sourceID := factory.CreateRiskAnalysis(owner, factory.Attrs{"name": "MCP source"})
	factory.CreateRiskAnalysisDiagram(owner, sourceID, factory.Attrs{"name": "MCP diagram"})

	var result struct {
		RiskAnalysis struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			MatrixSize struct {
				Rows int `json:"rows"`
				Cols int `json:"cols"`
			} `json:"matrix_size"`
		} `json:"risk_analysis"`
	}
	mc.CallToolInto("forkRiskAnalysis", map[string]any{
		"id":   sourceID,
		"name": "MCP forked",
	}, &result)
	require.NotEmpty(t, result.RiskAnalysis.ID)
	assert.NotEqual(t, sourceID, result.RiskAnalysis.ID)
	assert.Equal(t, "MCP forked", result.RiskAnalysis.Name)
	assert.Equal(t, 5, result.RiskAnalysis.MatrixSize.Rows)
	assert.Equal(t, 5, result.RiskAnalysis.MatrixSize.Cols)
}
