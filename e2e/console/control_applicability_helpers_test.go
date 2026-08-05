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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const applicabilityStatementNodeSelection = `
	id
	applicability
	justification
	createdAt
	updatedAt
	canUpdate: permission(action: "core:applicability-statement:update")
	canDelete: permission(action: "core:applicability-statement:delete")
	statementOfApplicability { id name }
	control { id name sectionTitle }
`

type (
	applicabilityStatementWireNode struct {
		ID                       string    `json:"id"`
		Applicability            bool      `json:"applicability"`
		Justification            string    `json:"justification"`
		CreatedAt                time.Time `json:"createdAt"`
		UpdatedAt                time.Time `json:"updatedAt"`
		CanUpdate                bool      `json:"canUpdate"`
		CanDelete                bool      `json:"canDelete"`
		StatementOfApplicability struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"statementOfApplicability"`
		Control struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			SectionTitle string `json:"sectionTitle"`
		} `json:"control"`
	}

	statementOfApplicabilityWireNode struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}

	applicabilityStatementsConnection struct {
		TotalCount int `json:"totalCount"`
		PageInfo   testutil.PageInfo
		Edges      []struct {
			Cursor string                         `json:"cursor"`
			Node   applicabilityStatementWireNode `json:"node"`
		} `json:"edges"`
	}

	statementsOfApplicabilityConnection struct {
		TotalCount int `json:"totalCount"`
		PageInfo   testutil.PageInfo
		Edges      []struct {
			Cursor string                           `json:"cursor"`
			Node   statementOfApplicabilityWireNode `json:"node"`
		} `json:"edges"`
	}

	controlRelationsWire struct {
		ID             string `json:"id"`
		Regulatory     bool   `json:"regulatory"`
		Contractual    bool   `json:"contractual"`
		RiskAssessment bool   `json:"riskAssessment"`
		Organization   struct {
			ID string `json:"id"`
		} `json:"organization"`
		Obligations struct {
			TotalCount int `json:"totalCount"`
		} `json:"obligations"`
		Documents struct {
			TotalCount int `json:"totalCount"`
		} `json:"documents"`
		Audits struct {
			TotalCount int `json:"totalCount"`
		} `json:"audits"`
	}
)

func createStatementOfApplicabilityLeaf(
	t *testing.T,
	client *testutil.Client,
	name string,
) statementOfApplicabilityWireNode {
	t.Helper()

	const mutation = `
		mutation($input: CreateStatementOfApplicabilityInput!) {
			createStatementOfApplicability(input: $input) {
				statementOfApplicabilityEdge {
					cursor
					node {
						id
						name
						createdAt
						updatedAt
					}
				}
			}
		}
	`

	var result struct {
		CreateStatementOfApplicability struct {
			StatementOfApplicabilityEdge struct {
				Node statementOfApplicabilityWireNode `json:"node"`
			} `json:"statementOfApplicabilityEdge"`
		} `json:"createStatementOfApplicability"`
	}

	err := client.Execute(
		mutation,
		map[string]any{
			"input": map[string]any{
				"organizationId": client.GetOrganizationID().String(),
				"name":           name,
			},
		},
		&result,
	)
	require.NoError(t, err)

	return result.CreateStatementOfApplicability.StatementOfApplicabilityEdge.Node
}

func createApplicabilityStatementLeaf(
	t *testing.T,
	client *testutil.Client,
	soaID, controlID string,
	applicability bool,
	justification *string,
) applicabilityStatementWireNode {
	t.Helper()

	const mutation = `
		mutation($input: CreateApplicabilityStatementInput!) {
			createApplicabilityStatement(input: $input) {
				applicabilityStatementEdge {
					cursor
					node {
						NODE
					}
				}
			}
		}
	`

	input := map[string]any{
		"statementOfApplicabilityId": soaID,
		"controlId":                  controlID,
		"applicability":              applicability,
	}
	if justification != nil {
		input["justification"] = *justification
	}

	var result struct {
		CreateApplicabilityStatement struct {
			ApplicabilityStatementEdge struct {
				Node applicabilityStatementWireNode `json:"node"`
			} `json:"applicabilityStatementEdge"`
		} `json:"createApplicabilityStatement"`
	}

	query := replaceApplicabilityStatementNodeSelection(mutation)

	err := client.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(t, err)

	return result.CreateApplicabilityStatement.ApplicabilityStatementEdge.Node
}

func createApplicabilityStatementExpectError(
	t *testing.T,
	client *testutil.Client,
	soaID, controlID string,
	applicability bool,
) error {
	t.Helper()

	const mutation = `
		mutation($input: CreateApplicabilityStatementInput!) {
			createApplicabilityStatement(input: $input) {
				applicabilityStatementEdge { node { id } }
			}
		}
	`

	_, err := client.Do(
		mutation,
		map[string]any{
			"input": map[string]any{
				"statementOfApplicabilityId": soaID,
				"controlId":                  controlID,
				"applicability":              applicability,
			},
		},
	)

	return err
}

func updateApplicabilityStatementLeaf(
	t *testing.T,
	client *testutil.Client,
	applicabilityStatementID string,
	applicability bool,
	justification *string,
) applicabilityStatementWireNode {
	t.Helper()

	const mutation = `
		mutation($input: UpdateApplicabilityStatementInput!) {
			updateApplicabilityStatement(input: $input) {
				applicabilityStatement {
					NODE
				}
			}
		}
	`

	input := map[string]any{
		"applicabilityStatementId": applicabilityStatementID,
		"applicability":            applicability,
	}
	if justification != nil {
		input["justification"] = *justification
	}

	var result struct {
		UpdateApplicabilityStatement struct {
			ApplicabilityStatement applicabilityStatementWireNode `json:"applicabilityStatement"`
		} `json:"updateApplicabilityStatement"`
	}

	query := replaceApplicabilityStatementNodeSelection(mutation)

	err := client.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(t, err)

	return result.UpdateApplicabilityStatement.ApplicabilityStatement
}

func updateApplicabilityStatementExpectError(
	t *testing.T,
	client *testutil.Client,
	applicabilityStatementID string,
	applicability bool,
) error {
	t.Helper()

	_, err := client.Do(
		`
			mutation($input: UpdateApplicabilityStatementInput!) {
				updateApplicabilityStatement(input: $input) {
					applicabilityStatement { id }
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"applicabilityStatementId": applicabilityStatementID,
				"applicability":            applicability,
			},
		},
	)

	return err
}

func deleteApplicabilityStatementLeaf(
	t *testing.T,
	client *testutil.Client,
	applicabilityStatementID string,
) string {
	t.Helper()

	const mutation = `
		mutation($input: DeleteApplicabilityStatementInput!) {
			deleteApplicabilityStatement(input: $input) {
				deletedApplicabilityStatementId
			}
		}
	`

	var result struct {
		DeleteApplicabilityStatement struct {
			DeletedApplicabilityStatementID string `json:"deletedApplicabilityStatementId"`
		} `json:"deleteApplicabilityStatement"`
	}

	err := client.Execute(
		mutation,
		map[string]any{
			"input": map[string]any{
				"applicabilityStatementId": applicabilityStatementID,
			},
		},
		&result,
	)
	require.NoError(t, err)

	return result.DeleteApplicabilityStatement.DeletedApplicabilityStatementID
}

func deleteApplicabilityStatementExpectError(
	t *testing.T,
	client *testutil.Client,
	applicabilityStatementID string,
) error {
	t.Helper()

	_, err := client.Do(
		`
			mutation($input: DeleteApplicabilityStatementInput!) {
				deleteApplicabilityStatement(input: $input) {
					deletedApplicabilityStatementId
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"applicabilityStatementId": applicabilityStatementID,
			},
		},
	)

	return err
}

func queryApplicabilityStatementNode(
	t *testing.T,
	client *testutil.Client,
	id string,
) *applicabilityStatementWireNode {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on ApplicabilityStatement {
					NODE
				}
			}
		}
	`

	var result struct {
		Node *applicabilityStatementWireNode `json:"node"`
	}

	err := client.Execute(
		replaceApplicabilityStatementNodeSelection(query),
		map[string]any{"id": id},
		&result,
	)
	require.NoError(t, err)

	return result.Node
}

func querySOAApplicabilityStatements(
	t *testing.T,
	client *testutil.Client,
	soaID string,
) applicabilityStatementsConnection {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on StatementOfApplicability {
					applicabilityStatements(first: 10) {
						totalCount
						pageInfo {
							hasNextPage
							hasPreviousPage
							startCursor
							endCursor
						}
						edges {
							cursor
							node {
								id
								applicability
								justification
								control { id }
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			ApplicabilityStatements applicabilityStatementsConnection `json:"applicabilityStatements"`
		} `json:"node"`
	}

	err := client.Execute(query, map[string]any{"id": soaID}, &result)
	require.NoError(t, err)

	return result.Node.ApplicabilityStatements
}

func updateStatementOfApplicabilityLeaf(
	t *testing.T,
	client *testutil.Client,
	id, name string,
) statementOfApplicabilityWireNode {
	t.Helper()

	const mutation = `
		mutation($input: UpdateStatementOfApplicabilityInput!) {
			updateStatementOfApplicability(input: $input) {
				statementOfApplicability {
					id
					name
					createdAt
					updatedAt
				}
			}
		}
	`

	var result struct {
		UpdateStatementOfApplicability struct {
			StatementOfApplicability statementOfApplicabilityWireNode `json:"statementOfApplicability"`
		} `json:"updateStatementOfApplicability"`
	}

	err := client.Execute(
		mutation,
		map[string]any{
			"input": map[string]any{
				"id":   id,
				"name": name,
			},
		},
		&result,
	)
	require.NoError(t, err)

	return result.UpdateStatementOfApplicability.StatementOfApplicability
}

func deleteStatementOfApplicabilityLeaf(
	t *testing.T,
	client *testutil.Client,
	soaID string,
) string {
	t.Helper()

	const mutation = `
		mutation($input: DeleteStatementOfApplicabilityInput!) {
			deleteStatementOfApplicability(input: $input) {
				deletedStatementOfApplicabilityId
			}
		}
	`

	var result struct {
		DeleteStatementOfApplicability struct {
			DeletedStatementOfApplicabilityID string `json:"deletedStatementOfApplicabilityId"`
		} `json:"deleteStatementOfApplicability"`
	}

	err := client.Execute(
		mutation,
		map[string]any{
			"input": map[string]any{
				"statementOfApplicabilityId": soaID,
			},
		},
		&result,
	)
	require.NoError(t, err)

	return result.DeleteStatementOfApplicability.DeletedStatementOfApplicabilityID
}

func queryOrganizationStatementsOfApplicability(
	t *testing.T,
	client *testutil.Client,
) statementsOfApplicabilityConnection {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					statementsOfApplicability(first: 50) {
						totalCount
						pageInfo {
							hasNextPage
							hasPreviousPage
						}
						edges {
							cursor
							node { id name }
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			StatementsOfApplicability statementsOfApplicabilityConnection `json:"statementsOfApplicability"`
		} `json:"node"`
	}

	err := client.Execute(
		query,
		map[string]any{"id": client.GetOrganizationID().String()},
		&result,
	)
	require.NoError(t, err)

	return result.Node.StatementsOfApplicability
}

func createObligationLeaf(
	t *testing.T,
	client *testutil.Client,
	requirement, obligationType string,
) string {
	t.Helper()

	ownerID := factory.CreateUser(client)

	const mutation = `
		mutation($input: CreateObligationInput!) {
			createObligation(input: $input) {
				obligationEdge { node { id } }
			}
		}
	`

	var result struct {
		CreateObligation struct {
			ObligationEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"obligationEdge"`
		} `json:"createObligation"`
	}

	err := client.Execute(
		mutation,
		map[string]any{
			"input": map[string]any{
				"organizationId": client.GetOrganizationID().String(),
				"area":           "Compliance",
				"requirement":    requirement,
				"ownerId":        ownerID,
				"status":         "NON_COMPLIANT",
				"type":           obligationType,
			},
		},
		&result,
	)
	require.NoError(t, err)

	return result.CreateObligation.ObligationEdge.Node.ID
}

func createControlObligationMappingLeaf(
	t *testing.T,
	client *testutil.Client,
	controlID, obligationID string,
) {
	t.Helper()

	const mutation = `
		mutation($input: CreateControlObligationMappingInput!) {
			createControlObligationMapping(input: $input) {
				controlEdge { node { id } }
				obligationEdge { node { id } }
			}
		}
	`

	err := client.Execute(
		mutation,
		map[string]any{
			"input": map[string]any{
				"controlId":    controlID,
				"obligationId": obligationID,
			},
		},
		nil,
	)
	require.NoError(t, err)
}

func createControlObligationMappingExpectError(
	t *testing.T,
	client *testutil.Client,
	controlID, obligationID string,
) error {
	t.Helper()

	_, err := client.Do(
		`
			mutation($input: CreateControlObligationMappingInput!) {
				createControlObligationMapping(input: $input) {
					controlEdge { node { id } }
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"controlId":    controlID,
				"obligationId": obligationID,
			},
		},
	)

	return err
}

func deleteControlObligationMappingLeaf(
	t *testing.T,
	client *testutil.Client,
	controlID, obligationID string,
) {
	t.Helper()

	_, err := client.Do(
		`
			mutation($input: DeleteControlObligationMappingInput!) {
				deleteControlObligationMapping(input: $input) {
					deletedControlId
					deletedObligationId
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"controlId":    controlID,
				"obligationId": obligationID,
			},
		},
	)
	require.NoError(t, err)
}

func deleteControlObligationMappingExpectError(
	t *testing.T,
	client *testutil.Client,
	controlID, obligationID string,
) error {
	t.Helper()

	_, err := client.Do(
		`
			mutation($input: DeleteControlObligationMappingInput!) {
				deleteControlObligationMapping(input: $input) {
					deletedControlId
					deletedObligationId
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"controlId":    controlID,
				"obligationId": obligationID,
			},
		},
	)

	return err
}

func queryControlRelations(
	t *testing.T,
	client *testutil.Client,
	controlID string,
) controlRelationsWire {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Control {
					id
					regulatory
					contractual
					riskAssessment
					organization { id }
					obligations(first: 10) { totalCount }
					documents(first: 10) { totalCount }
					audits(first: 10) { totalCount }
				}
			}
		}
	`

	var result struct {
		Node controlRelationsWire `json:"node"`
	}

	err := client.Execute(query, map[string]any{"id": controlID}, &result)
	require.NoError(t, err)

	return result.Node
}

func queryControlObligations(
	t *testing.T,
	client *testutil.Client,
	controlID string,
) (totalCount int, obligationIDs []string) {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Control {
					obligations(first: 10) {
						totalCount
						edges { node { id requirement type } }
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Obligations struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID          string `json:"id"`
						Requirement string `json:"requirement"`
						Type        string `json:"type"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"obligations"`
		} `json:"node"`
	}

	err := client.Execute(query, map[string]any{"id": controlID}, &result)
	require.NoError(t, err)

	ids := make([]string, 0, len(result.Node.Obligations.Edges))
	for _, edge := range result.Node.Obligations.Edges {
		ids = append(ids, edge.Node.ID)
	}

	return result.Node.Obligations.TotalCount, ids
}

func replaceApplicabilityStatementNodeSelection(query string) string {
	return strings.ReplaceAll(query, "NODE", applicabilityStatementNodeSelection)
}
