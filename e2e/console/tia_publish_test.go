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

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestTransferImpactAssessment_PublishList(t *testing.T) {
	t.Parallel()

	t.Run(
		"publish without approvers publishes immediately",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			paID := factory.NewProcessingActivity(owner).
				WithName("TIA Publish PA").
				WithLawfulBasis("CONSENT").
				Create()
			createTIAForPublish(t, owner, paID, "EU to US transfer")

			const query = `
				mutation($input: PublishTransferImpactAssessmentListInput!) {
					publishTransferImpactAssessmentList(input: $input) {
						documentEdge {
							node {
								id
								writeMode
								status
							}
						}
						documentVersionEdge {
							node {
								id
								title
								documentType
								status
								major
								minor
								content
							}
						}
					}
				}
			`

			var result struct {
				PublishTransferImpactAssessmentList struct {
					DocumentEdge struct {
						Node struct {
							ID        string `json:"id"`
							WriteMode string `json:"writeMode"`
							Status    string `json:"status"`
						} `json:"node"`
					} `json:"documentEdge"`
					DocumentVersionEdge struct {
						Node struct {
							ID           string `json:"id"`
							Title        string `json:"title"`
							DocumentType string `json:"documentType"`
							Status       string `json:"status"`
							Major        int    `json:"major"`
							Minor        int    `json:"minor"`
							Content      string `json:"content"`
						} `json:"node"`
					} `json:"documentVersionEdge"`
				} `json:"publishTransferImpactAssessmentList"`
			}

			err := owner.Execute(
				query,
				map[string]any{
					"input": map[string]any{
						"organizationId": owner.GetOrganizationID(),
					},
				},
				&result,
			)
			require.NoError(t, err)

			doc := result.PublishTransferImpactAssessmentList.DocumentEdge.Node
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "GENERATED", doc.WriteMode)
			assert.Equal(t, "ACTIVE", doc.Status)

			ver := result.PublishTransferImpactAssessmentList.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "REGISTER", ver.DocumentType)
			assert.Equal(t, "PUBLISHED", ver.Status)
			assert.Equal(t, 1, ver.Major)
			assert.Equal(t, 0, ver.Minor)
			assert.Contains(t, ver.Content, "EU to US transfer")
		},
	)

	t.Run(
		"publish with approvers creates draft pending approval",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			paID := factory.NewProcessingActivity(owner).
				WithName("TIA Approval PA").
				WithLawfulBasis("CONSENT").
				Create()
			createTIAForPublish(t, owner, paID, "TIA Approval")

			const query = `
				mutation($input: PublishTransferImpactAssessmentListInput!) {
					publishTransferImpactAssessmentList(input: $input) {
						documentVersionEdge {
							node {
								status
							}
						}
					}
				}
			`

			var result struct {
				PublishTransferImpactAssessmentList struct {
					DocumentVersionEdge struct {
						Node struct {
							Status string `json:"status"`
						} `json:"node"`
					} `json:"documentVersionEdge"`
				} `json:"publishTransferImpactAssessmentList"`
			}

			err := owner.Execute(
				query,
				map[string]any{
					"input": map[string]any{
						"organizationId": owner.GetOrganizationID(),
						"approverIds":    []string{owner.GetProfileID().String()},
					},
				},
				&result,
			)
			require.NoError(t, err)
			assert.Equal(
				t,
				"PENDING_APPROVAL",
				result.PublishTransferImpactAssessmentList.DocumentVersionEdge.Node.Status,
			)
		},
	)

	t.Run(
		"document linked back to organization via transferImpactAssessmentsDocument",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			paID := factory.NewProcessingActivity(owner).
				WithName("TIA Link PA").
				WithLawfulBasis("CONSENT").
				Create()
			createTIAForPublish(t, owner, paID, "TIA Link")

			const publishQuery = `
				mutation($input: PublishTransferImpactAssessmentListInput!) {
					publishTransferImpactAssessmentList(input: $input) {
						documentEdge { node { id } }
					}
				}
			`

			var publishResult struct {
				PublishTransferImpactAssessmentList struct {
					DocumentEdge struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"documentEdge"`
				} `json:"publishTransferImpactAssessmentList"`
			}

			err := owner.Execute(
				publishQuery,
				map[string]any{
					"input": map[string]any{
						"organizationId": owner.GetOrganizationID(),
					},
				},
				&publishResult,
			)
			require.NoError(t, err)

			docID := publishResult.PublishTransferImpactAssessmentList.DocumentEdge.Node.ID

			const orgQuery = `
				query($id: ID!) {
					node(id: $id) {
						... on Organization {
							transferImpactAssessmentsDocument { id }
						}
					}
				}
			`

			var orgResult struct {
				Node struct {
					TransferImpactAssessmentsDocument *struct {
						ID string `json:"id"`
					} `json:"transferImpactAssessmentsDocument"`
				} `json:"node"`
			}

			err = owner.Execute(
				orgQuery,
				map[string]any{"id": owner.GetOrganizationID()},
				&orgResult,
			)
			require.NoError(t, err)
			require.NotNil(t, orgResult.Node.TransferImpactAssessmentsDocument)
			assert.Equal(t, docID, orgResult.Node.TransferImpactAssessmentsDocument.ID)
		},
	)
}

func TestTransferImpactAssessment_PublishList_RBAC(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	paID := factory.NewProcessingActivity(owner).
		WithName("TIA RBAC PA").
		WithLawfulBasis("CONSENT").
		Create()
	createTIAForPublish(t, owner, paID, "TIA RBAC")

	const query = `
		mutation($input: PublishTransferImpactAssessmentListInput!) {
			publishTransferImpactAssessmentList(input: $input) {
				documentEdge { node { id } }
			}
		}
	`

	t.Run("viewer cannot publish TIA list", func(t *testing.T) {
		t.Parallel()

		err := viewer.ExecuteShouldFail(
			query,
			map[string]any{
				"input": map[string]any{
					"organizationId": owner.GetOrganizationID(),
				},
			},
		)
		testutil.RequireForbiddenError(t, err)
	})
}

func createTIAForPublish(t *testing.T, client *testutil.Client, processingActivityID string, transfer string) string {
	t.Helper()

	const query = `
		mutation($input: CreateTransferImpactAssessmentInput!) {
			createTransferImpactAssessment(input: $input) {
				transferImpactAssessment { id }
			}
		}
	`

	var result struct {
		CreateTransferImpactAssessment struct {
			TransferImpactAssessment struct {
				ID string `json:"id"`
			} `json:"transferImpactAssessment"`
		} `json:"createTransferImpactAssessment"`
	}

	err := client.Execute(query, map[string]any{
		"input": map[string]any{
			"processingActivityId": processingActivityID,
			"transfer":             transfer,
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateTransferImpactAssessment.TransferImpactAssessment.ID
}
