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

func TestTreatmentPlan_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("viewer cannot create", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		riskID := factory.CreateRisk(owner)
		analysisID := factory.CreateRiskAnalysis(owner)
		factory.LinkRiskToAnalysis(owner, riskID, analysisID)

		_, err := viewer.Do(`
			mutation($input: CreateTreatmentPlanInput!) {
				createTreatmentPlan(input: $input) {
					treatmentPlanEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"riskId":             riskID,
				"riskAnalysisId":     analysisID,
				"treatment":          "MITIGATED",
				"ownerId":            owner.GetProfileID().String(),
				"inherentLikelihood": 2,
				"inherentImpact":     2,
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot create treatment plan")
	})

	t.Run("viewer can read", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		riskID := factory.CreateRisk(owner)
		analysisID := factory.CreateRiskAnalysis(owner)
		factory.LinkRiskToAnalysis(owner, riskID, analysisID)
		tpID := factory.CreateTreatmentPlan(owner, riskID, analysisID)

		var result struct {
			Node struct {
				ID        string `json:"id"`
				Treatment string `json:"treatment"`
			} `json:"node"`
		}

		err := viewer.Execute(`
			query($id: ID!) { node(id: $id) { ... on TreatmentPlan { id treatment } } }
		`, map[string]any{"id": tpID}, &result)
		require.NoError(t, err)
		assert.Equal(t, tpID, result.Node.ID)
		assert.Equal(t, "MITIGATED", result.Node.Treatment)
	})
}
