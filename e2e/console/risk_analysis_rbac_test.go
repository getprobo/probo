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

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestRiskAnalysis_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("viewer cannot create", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		_, err := viewer.Do(`
			mutation($input: CreateRiskAnalysisInput!) {
				createRiskAnalysis(input: $input) { riskAnalysisEdge { node { id } } }
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId": viewer.GetOrganizationID().String(),
				"name":           "test",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot create risk analysis")
	})

	t.Run("viewer can read", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		raID := factory.CreateRiskAnalysis(owner, factory.Attrs{"name": "Visible"})

		var result struct {
			Node struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		}

		err := viewer.Execute(`
			query($id: ID!) { node(id: $id) { ... on RiskAnalysis { id name } } }
		`, map[string]any{"id": raID}, &result)
		require.NoError(t, err)
		assert.Equal(t, "Visible", result.Node.Name)
	})
}
func TestRiskAnalysisBoundary_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("viewer cannot create", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		raID := factory.CreateRiskAnalysis(owner)
		scopeID := factory.CreateRiskAnalysisScope(owner, raID)

		_, err := viewer.Do(`
			mutation($input: CreateRiskAnalysisBoundaryInput!) {
				createRiskAnalysisBoundary(input: $input) {
					riskAnalysisBoundaryEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskAnalysisScopeId": scopeID,
				"name":                "Nope",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot create risk analysis boundary")
	})

	t.Run("viewer can read", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		raID := factory.CreateRiskAnalysis(owner)
		scopeID := factory.CreateRiskAnalysisScope(owner, raID)
		boundaryID := factory.CreateRiskAnalysisBoundary(owner, scopeID, factory.Attrs{"name": "Visible"})

		var result struct {
			Node struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		}

		err := viewer.Execute(`
			query($id: ID!) { node(id: $id) { ... on RiskAnalysisBoundary { id name } } }
		`, map[string]any{"id": boundaryID}, &result)
		require.NoError(t, err)
		assert.Equal(t, "Visible", result.Node.Name)
	})
}
