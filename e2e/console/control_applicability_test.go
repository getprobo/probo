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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestControlApplicability_Lifecycle(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	frameworkID := factory.NewFramework(owner).
		WithName(factory.SafeName("SOA lifecycle framework")).
		Create()
	controlID := factory.NewControl(owner, frameworkID).
		WithName(factory.SafeName("SOA lifecycle control")).
		WithSectionTitle("A.1").
		Create()

	relationsControlID := factory.NewControl(owner, frameworkID).
		WithName(factory.SafeName("SOA relations control")).
		WithSectionTitle("A.2").
		Create()

	beforeCreate := time.Now().Add(-time.Second)
	soaName := factory.SafeName("Lifecycle SOA")
	soa := createStatementOfApplicabilityLeaf(t, owner, soaName)

	assert.NotEmpty(t, soa.ID)
	assert.Equal(t, soaName, soa.Name)
	testutil.AssertTimestampsOnCreate(t, soa.CreatedAt, soa.UpdatedAt, beforeCreate)

	initialJustification := "Required for certification scope"
	createdAS := createApplicabilityStatementLeaf(
		t,
		owner,
		soa.ID,
		controlID,
		true,
		&initialJustification,
	)

	assert.Equal(t, soa.ID, createdAS.StatementOfApplicability.ID)
	assert.Equal(t, soaName, createdAS.StatementOfApplicability.Name)
	assert.Equal(t, controlID, createdAS.Control.ID)
	assert.True(t, createdAS.Applicability)
	assert.Equal(t, initialJustification, createdAS.Justification)
	assert.True(t, createdAS.CanUpdate)
	assert.True(t, createdAS.CanDelete)
	testutil.AssertTimestampsOnCreate(t, createdAS.CreatedAt, createdAS.UpdatedAt, beforeCreate)

	soaList := querySOAApplicabilityStatements(t, owner, soa.ID)
	assert.Equal(t, 1, soaList.TotalCount)
	testutil.AssertFirstPage(t, len(soaList.Edges), soaList.PageInfo, 1, false)
	require.Len(t, soaList.Edges, 1)
	assert.Equal(t, createdAS.ID, soaList.Edges[0].Node.ID)
	assert.Equal(t, controlID, soaList.Edges[0].Node.Control.ID)
	assert.NotEmpty(t, soaList.Edges[0].Cursor)

	queriedAS := queryApplicabilityStatementNode(t, owner, createdAS.ID)
	require.NotNil(t, queriedAS)
	assert.Equal(t, createdAS.ID, queriedAS.ID)
	assert.Equal(t, soa.ID, queriedAS.StatementOfApplicability.ID)
	assert.Equal(t, controlID, queriedAS.Control.ID)
	assert.True(t, queriedAS.CanUpdate)
	assert.True(t, queriedAS.CanDelete)

	updatedJustification := "Out of scope for this audit period"
	updatedAS := updateApplicabilityStatementLeaf(
		t,
		owner,
		createdAS.ID,
		false,
		&updatedJustification,
	)

	assert.Equal(t, createdAS.ID, updatedAS.ID)
	assert.False(t, updatedAS.Applicability)
	assert.Equal(t, updatedJustification, updatedAS.Justification)
	testutil.AssertTimestampsOnUpdate(
		t,
		updatedAS.CreatedAt,
		updatedAS.UpdatedAt,
		createdAS.CreatedAt,
		createdAS.UpdatedAt,
	)

	roundtrip := queryApplicabilityStatementNode(t, owner, createdAS.ID)
	require.NotNil(t, roundtrip)
	assert.False(t, roundtrip.Applicability)
	assert.Equal(t, updatedJustification, roundtrip.Justification)

	deletedASID := deleteApplicabilityStatementLeaf(t, owner, createdAS.ID)
	assert.Equal(t, createdAS.ID, deletedASID)

	afterDeleteList := querySOAApplicabilityStatements(t, owner, soa.ID)
	assert.Equal(t, 0, afterDeleteList.TotalCount)
	assert.Empty(t, afterDeleteList.Edges)

	extraSOAName := factory.SafeName("Extra SOA")
	extraSOA := createStatementOfApplicabilityLeaf(t, owner, extraSOAName)
	renamed := factory.SafeName("Renamed SOA")
	updatedSOA := updateStatementOfApplicabilityLeaf(t, owner, extraSOA.ID, renamed)
	assert.Equal(t, extraSOA.ID, updatedSOA.ID)
	assert.Equal(t, renamed, updatedSOA.Name)
	testutil.AssertTimestampsOnUpdate(
		t,
		updatedSOA.CreatedAt,
		updatedSOA.UpdatedAt,
		extraSOA.CreatedAt,
		extraSOA.UpdatedAt,
	)

	orgSOAs := queryOrganizationStatementsOfApplicability(t, owner)
	assert.GreaterOrEqual(t, orgSOAs.TotalCount, 2)
	assert.Contains(
		t,
		soaIDsFromConnection(orgSOAs),
		soa.ID,
	)
	assert.Contains(t, soaIDsFromConnection(orgSOAs), extraSOA.ID)

	deletedSOAID := deleteStatementOfApplicabilityLeaf(t, owner, extraSOA.ID)
	assert.Equal(t, extraSOA.ID, deletedSOAID)

	var deletedSOANode struct {
		Node *struct {
			ID string `json:"id"`
		} `json:"node"`
	}

	err := owner.Execute(
		`
			query($id: ID!) {
				node(id: $id) {
					... on StatementOfApplicability { id }
				}
			}
		`,
		map[string]any{"id": extraSOA.ID},
		&deletedSOANode,
	)
	testutil.AssertNodeNotAccessible(t, err, deletedSOANode.Node == nil, "statement of applicability")

	orgSOAsAfterDelete := queryOrganizationStatementsOfApplicability(t, owner)
	assert.NotContains(t, soaIDsFromConnection(orgSOAsAfterDelete), extraSOA.ID)

	legalObligationID := createObligationLeaf(
		t,
		owner,
		factory.SafeName("Legal obligation"),
		"LEGAL",
	)
	contractualObligationID := createObligationLeaf(
		t,
		owner,
		factory.SafeName("Contractual obligation"),
		"CONTRACTUAL",
	)
	mappedObligationID := createObligationLeaf(
		t,
		owner,
		factory.SafeName("Mapped obligation"),
		"LEGAL",
	)

	createControlObligationMappingLeaf(t, owner, relationsControlID, legalObligationID)
	createControlObligationMappingLeaf(t, owner, relationsControlID, contractualObligationID)
	createControlObligationMappingLeaf(t, owner, relationsControlID, mappedObligationID)

	obligationCount, obligationIDs := queryControlObligations(t, owner, relationsControlID)
	assert.Equal(t, 3, obligationCount)
	assert.Contains(t, obligationIDs, mappedObligationID)

	deleteControlObligationMappingLeaf(t, owner, relationsControlID, mappedObligationID)
	obligationCountAfter, obligationIDsAfter := queryControlObligations(t, owner, relationsControlID)
	assert.Equal(t, 2, obligationCountAfter)
	assert.NotContains(t, obligationIDsAfter, mappedObligationID)

	setupControlRelationMappings(t, owner, frameworkID, relationsControlID)

	relations := queryControlRelations(t, owner, relationsControlID)
	assert.Equal(t, relationsControlID, relations.ID)
	assert.Equal(t, orgID, relations.Organization.ID)
	assert.True(t, relations.Regulatory, "LEGAL obligation mapping should set regulatory")
	assert.True(t, relations.Contractual, "CONTRACTUAL obligation mapping should set contractual")
	assert.True(t, relations.RiskAssessment, "control-measure-risk chain should set riskAssessment")
	assert.Equal(t, 2, relations.Obligations.TotalCount)
	assert.Equal(t, 1, relations.Documents.TotalCount)
	assert.Equal(t, 1, relations.Audits.TotalCount)
}

func soaIDsFromConnection(conn statementsOfApplicabilityConnection) []string {
	ids := make([]string, 0, len(conn.Edges))
	for _, edge := range conn.Edges {
		ids = append(ids, edge.Node.ID)
	}

	return ids
}

func setupControlRelationMappings(
	t *testing.T,
	owner *testutil.Client,
	frameworkID, controlID string,
) {
	t.Helper()

	measureID := factory.NewMeasure(owner).
		WithName(factory.SafeName("Relations measure")).
		Create()
	riskID := factory.CreateRisk(owner, factory.Attrs{"name": factory.SafeName("Relations risk")})
	documentID := factory.NewDocument(owner).
		WithTitle(factory.SafeName("Relations document")).
		Create()
	auditID := factory.CreateAudit(owner, frameworkID, factory.Attrs{
		"name": factory.SafeName("Relations audit"),
	})

	require.NoError(t, owner.Execute(
		`
			mutation($input: CreateControlMeasureMappingInput!) {
				createControlMeasureMapping(input: $input) {
					controlEdge { node { id } }
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"controlId": controlID,
				"measureId": measureID,
			},
		},
		nil,
	))

	require.NoError(t, owner.Execute(
		`
			mutation($input: CreateRiskMeasureMappingInput!) {
				createRiskMeasureMapping(input: $input) {
					riskEdge { node { id } }
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"riskId":    riskID,
				"measureId": measureID,
			},
		},
		nil,
	))

	require.NoError(t, owner.Execute(
		`
			mutation($input: CreateControlDocumentMappingInput!) {
				createControlDocumentMapping(input: $input) {
					controlEdge { node { id } }
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"controlId":  controlID,
				"documentId": documentID,
			},
		},
		nil,
	))

	require.NoError(t, owner.Execute(
		`
			mutation($input: CreateControlAuditMappingInput!) {
				createControlAuditMapping(input: $input) {
					controlEdge { node { id } }
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"controlId": controlID,
				"auditId":   auditID,
			},
		},
		nil,
	))
}
