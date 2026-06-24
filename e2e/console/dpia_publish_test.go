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

func TestDataProtectionImpactAssessment_PublishList(t *testing.T) {
	t.Parallel()

	t.Run(
		"publish without approvers publishes immediately",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			paID := factory.NewProcessingActivity(owner).
				WithName("DPIA Publish PA").
				WithLawfulBasis("CONSENT").
				Create()
			createDPIAForPublish(t, owner, paID, "DPIA description")

			const query = `
				mutation($input: PublishDataProtectionImpactAssessmentListInput!) {
					publishDataProtectionImpactAssessmentList(input: $input) {
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
				PublishDataProtectionImpactAssessmentList struct {
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
				} `json:"publishDataProtectionImpactAssessmentList"`
			}

			err := owner.Execute(
				query,
				map[string]any{
					"input": map[string]any{
						"minor":          false,
						"organizationId": owner.GetOrganizationID(),
					},
				},
				&result,
			)
			require.NoError(t, err)

			doc := result.PublishDataProtectionImpactAssessmentList.DocumentEdge.Node
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "GENERATED", doc.WriteMode)
			assert.Equal(t, "ACTIVE", doc.Status)

			ver := result.PublishDataProtectionImpactAssessmentList.DocumentVersionEdge.Node
			assert.NotEmpty(t, ver.ID)
			assert.Equal(t, "REGISTER", ver.DocumentType)
			assert.Equal(t, "PUBLISHED", ver.Status)
			assert.Equal(t, 1, ver.Major)
			assert.Equal(t, 0, ver.Minor)
			assert.Contains(t, ver.Content, "DPIA description")
		},
	)

	t.Run(
		"publish with approvers creates draft pending approval",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			paID := factory.NewProcessingActivity(owner).
				WithName("DPIA Approval PA").
				WithLawfulBasis("CONSENT").
				Create()
			createDPIAForPublish(t, owner, paID, "DPIA Approval")

			const query = `
				mutation($input: PublishDataProtectionImpactAssessmentListInput!) {
					publishDataProtectionImpactAssessmentList(input: $input) {
						documentVersionEdge {
							node {
								status
							}
						}
					}
				}
			`

			var result struct {
				PublishDataProtectionImpactAssessmentList struct {
					DocumentVersionEdge struct {
						Node struct {
							Status string `json:"status"`
						} `json:"node"`
					} `json:"documentVersionEdge"`
				} `json:"publishDataProtectionImpactAssessmentList"`
			}

			err := owner.Execute(
				query,
				map[string]any{
					"input": map[string]any{
						"minor":          false,
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
				result.PublishDataProtectionImpactAssessmentList.DocumentVersionEdge.Node.Status,
			)
		},
	)

	t.Run(
		"document linked back to organization via dataProtectionImpactAssessmentsDocument",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			paID := factory.NewProcessingActivity(owner).
				WithName("DPIA Link PA").
				WithLawfulBasis("CONSENT").
				Create()
			createDPIAForPublish(t, owner, paID, "DPIA Link")

			const publishQuery = `
				mutation($input: PublishDataProtectionImpactAssessmentListInput!) {
					publishDataProtectionImpactAssessmentList(input: $input) {
						documentEdge { node { id } }
					}
				}
			`

			var publishResult struct {
				PublishDataProtectionImpactAssessmentList struct {
					DocumentEdge struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"documentEdge"`
				} `json:"publishDataProtectionImpactAssessmentList"`
			}

			err := owner.Execute(
				publishQuery,
				map[string]any{
					"input": map[string]any{
						"minor":          false,
						"organizationId": owner.GetOrganizationID(),
					},
				},
				&publishResult,
			)
			require.NoError(t, err)

			docID := publishResult.PublishDataProtectionImpactAssessmentList.DocumentEdge.Node.ID

			const orgQuery = `
				query($id: ID!) {
					node(id: $id) {
						... on Organization {
							dataProtectionImpactAssessmentsDocument { id }
						}
					}
				}
			`

			var orgResult struct {
				Node struct {
					DataProtectionImpactAssessmentsDocument *struct {
						ID string `json:"id"`
					} `json:"dataProtectionImpactAssessmentsDocument"`
				} `json:"node"`
			}

			err = owner.Execute(
				orgQuery,
				map[string]any{"id": owner.GetOrganizationID()},
				&orgResult,
			)
			require.NoError(t, err)
			require.NotNil(t, orgResult.Node.DataProtectionImpactAssessmentsDocument)
			assert.Equal(t, docID, orgResult.Node.DataProtectionImpactAssessmentsDocument.ID)
		},
	)
}

func TestDataProtectionImpactAssessment_PublishList_RBAC(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	paID := factory.NewProcessingActivity(owner).
		WithName("DPIA RBAC PA").
		WithLawfulBasis("CONSENT").
		Create()
	createDPIAForPublish(t, owner, paID, "DPIA RBAC")

	const query = `
		mutation($input: PublishDataProtectionImpactAssessmentListInput!) {
			publishDataProtectionImpactAssessmentList(input: $input) {
				documentEdge { node { id } }
			}
		}
	`

	t.Run("viewer cannot publish DPIA list", func(t *testing.T) {
		t.Parallel()

		err := viewer.ExecuteShouldFail(
			query,
			map[string]any{
				"input": map[string]any{
					"minor":          false,
					"organizationId": owner.GetOrganizationID(),
				},
			},
		)
		testutil.RequireForbiddenError(t, err)
	})
}

func createDPIAForPublish(t *testing.T, client *testutil.Client, processingActivityID string, description string) string {
	t.Helper()

	const query = `
		mutation($input: CreateDataProtectionImpactAssessmentInput!) {
			createDataProtectionImpactAssessment(input: $input) {
				dataProtectionImpactAssessment { id }
			}
		}
	`

	var result struct {
		CreateDataProtectionImpactAssessment struct {
			DataProtectionImpactAssessment struct {
				ID string `json:"id"`
			} `json:"dataProtectionImpactAssessment"`
		} `json:"createDataProtectionImpactAssessment"`
	}

	err := client.Execute(query, map[string]any{
		"input": map[string]any{
			"processingActivityId": processingActivityID,
			"description":          description,
			"residualRisk":         "LOW",
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateDataProtectionImpactAssessment.DataProtectionImpactAssessment.ID
}
