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

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestControlMeasureMapping_TenantIsolation(t *testing.T) {
	t.Parallel()
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.CreateFramework(org1Owner)
	controlID := factory.CreateControl(org1Owner, frameworkID)
	org2MeasureID := factory.NewMeasure(org2Owner).Create()

	_, err := org1Owner.Do(`
		mutation($input: CreateControlMeasureMappingInput!) {
			createControlMeasureMapping(input: $input) {
				controlEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"controlId": controlID,
			"measureId": org2MeasureID,
		},
	})
	require.Error(t, err, "must not link a control to a measure from another organization")
}
func TestRiskMeasureMapping_TenantIsolation(t *testing.T) {
	t.Parallel()
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	riskID := factory.CreateRisk(org1Owner)
	org2MeasureID := factory.NewMeasure(org2Owner).Create()

	_, err := org1Owner.Do(`
		mutation($input: CreateRiskMeasureMappingInput!) {
			createRiskMeasureMapping(input: $input) {
				riskEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskId":    riskID,
			"measureId": org2MeasureID,
		},
	})
	require.Error(t, err, "must not link a risk to a measure from another organization")
}
func TestControlDocumentMapping_TenantIsolation(t *testing.T) {
	t.Parallel()
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.CreateFramework(org1Owner)
	controlID := factory.CreateControl(org1Owner, frameworkID)
	org2DocumentID := factory.NewDocument(org2Owner).Create()

	_, err := org1Owner.Do(`
		mutation($input: CreateControlDocumentMappingInput!) {
			createControlDocumentMapping(input: $input) {
				controlEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"controlId":  controlID,
			"documentId": org2DocumentID,
		},
	})
	require.Error(t, err, "must not link a control to a document from another organization")
}
func TestControlAuditMapping_TenantIsolation(t *testing.T) {
	t.Parallel()
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.CreateFramework(org1Owner)
	controlID := factory.CreateControl(org1Owner, frameworkID)
	org2FrameworkID := factory.CreateFramework(org2Owner)
	org2AuditID := factory.CreateAudit(org2Owner, org2FrameworkID)

	_, err := org1Owner.Do(`
		mutation($input: CreateControlAuditMappingInput!) {
			createControlAuditMapping(input: $input) {
				controlEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"controlId": controlID,
			"auditId":   org2AuditID,
		},
	})
	require.Error(t, err, "must not link a control to an audit from another organization")
}
func TestRiskDocumentMapping_TenantIsolation(t *testing.T) {
	t.Parallel()
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	riskID := factory.CreateRisk(org1Owner)
	org2DocumentID := factory.NewDocument(org2Owner).Create()

	_, err := org1Owner.Do(`
		mutation($input: CreateRiskDocumentMappingInput!) {
			createRiskDocumentMapping(input: $input) {
				riskEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskId":     riskID,
			"documentId": org2DocumentID,
		},
	})
	require.Error(t, err, "must not link a risk to a document from another organization")
}
func TestRiskObligationMapping_TenantIsolation(t *testing.T) {
	t.Parallel()
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	riskID := factory.CreateRisk(org1Owner)

	org2ProfileID := factory.CreateUser(org2Owner)

	var createObligationResult struct {
		CreateObligation struct {
			ObligationEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"obligationEdge"`
		} `json:"createObligation"`
	}

	err := org2Owner.Execute(`
		mutation($input: CreateObligationInput!) {
			createObligation(input: $input) {
				obligationEdge {
					node { id }
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": org2Owner.GetOrganizationID().String(),
			"area":           "Risk Management",
			"requirement":    "Org2 Obligation for Mapping Isolation",
			"ownerId":        org2ProfileID,
			"status":         "NON_COMPLIANT",
			"type":           "LEGAL",
		},
	}, &createObligationResult)
	require.NoError(t, err)

	org2ObligationID := createObligationResult.CreateObligation.ObligationEdge.Node.ID

	_, err = org1Owner.Do(`
		mutation($input: CreateRiskObligationMappingInput!) {
			createRiskObligationMapping(input: $input) {
				riskEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"riskId":       riskID,
			"obligationId": org2ObligationID,
		},
	})
	require.Error(t, err, "must not link a risk to an obligation from another organization")
}
func TestMeasureDocumentMapping_TenantIsolation(t *testing.T) {
	t.Parallel()
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	measureID := factory.NewMeasure(org1Owner).Create()
	org2DocumentID := factory.NewDocument(org2Owner).Create()

	_, err := org1Owner.Do(`
		mutation($input: CreateMeasureDocumentMappingInput!) {
			createMeasureDocumentMapping(input: $input) {
				measureEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"measureId":  measureID,
			"documentId": org2DocumentID,
		},
	})
	require.Error(t, err, "must not link a measure to a document from another organization")
}
