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

func TestProcessingActivity_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	paID := factory.NewProcessingActivity(org1Owner).WithName("Org1 PA").Create()

	t.Run("cannot read processing activity from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on ProcessingActivity {
						id
						name
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": paID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "processing activity")
	})

	t.Run("cannot update processing activity from another organization", func(t *testing.T) {
		query := `
			mutation UpdateProcessingActivity($input: UpdateProcessingActivityInput!) {
				updateProcessingActivity(input: $input) {
					processingActivity { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   paID,
				"name": "Hijacked PA",
			},
		})
		require.Error(t, err, "Should not be able to update processing activity from another org")
	})

	t.Run("cannot delete processing activity from another organization", func(t *testing.T) {
		query := `
			mutation DeleteProcessingActivity($input: DeleteProcessingActivityInput!) {
				deleteProcessingActivity(input: $input) {
					deletedProcessingActivityId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
			},
		})
		require.Error(t, err, "Should not be able to delete processing activity from another org")
	})

	t.Run("cannot list processing activities from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						processingActivities(first: 100) {
							edges {
								node {
									id
									name
								}
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ProcessingActivities struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"processingActivities"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{
			"id": org1Owner.GetOrganizationID().String(),
		}, &result)
		if err == nil {
			for _, edge := range result.Node.ProcessingActivities.Edges {
				assert.NotEqual(t, paID, edge.Node.ID, "Should not see processing activity from another org")
			}
		}
	})

	// GHSA-c74x-79w6-63jh: a processing activity must not be able to store a
	// dataProtectionOfficerId belonging to another organization, whether set
	// on create or on a later update. Before the fix,
	// ProcessingActivityService.Create/Update persisted an attacker-supplied
	// profile id with no scoped ownership check, and
	// processingActivityResolver.DataProtectionOfficer authorized the
	// processing activity (caller's org) instead of the profile, letting the
	// caller read another org's member PII through the dataloader's
	// NewScopeFromObjectID(profileID) scope.
	org2ProfileID := factory.CreateUser(org2Owner, factory.Attrs{"fullName": "Org2 Confidential DPO"})

	t.Run("cannot create processing activity referencing a data protection officer from another organization", func(t *testing.T) {
		query := `
			mutation CreateProcessingActivity($input: CreateProcessingActivityInput!) {
				createProcessingActivity(input: $input) {
					processingActivityEdge {
						node { id }
					}
				}
			}
		`

		_, err := org1Owner.Do(query, map[string]any{
			"input": map[string]any{
				"organizationId":                       org1Owner.GetOrganizationID().String(),
				"name":                                 factory.SafeName("ProcessingActivity"),
				"specialOrCriminalData":                "NO",
				"lawfulBasis":                          "CONSENT",
				"internationalTransfers":               false,
				"dataProtectionImpactAssessmentNeeded": "NOT_NEEDED",
				"transferImpactAssessmentNeeded":       "NOT_NEEDED",
				"role":                                 "CONTROLLER",
				"dataProtectionOfficerId":              org2ProfileID,
			},
		})
		require.Error(t, err, "must not accept a dataProtectionOfficerId belonging to another organization")
	})

	t.Run("cannot update processing activity to reference a data protection officer from another organization", func(t *testing.T) {
		otherPaID := factory.NewProcessingActivity(org1Owner).WithName("Org1 PA Without DPO").Create()

		query := `
			mutation UpdateProcessingActivity($input: UpdateProcessingActivityInput!) {
				updateProcessingActivity(input: $input) {
					processingActivity { id }
				}
			}
		`

		_, err := org1Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":                      otherPaID,
				"dataProtectionOfficerId": org2ProfileID,
			},
		})
		require.Error(t, err, "must not accept a dataProtectionOfficerId belonging to another organization")
	})

	t.Run("cannot create processing activity referencing a thirdParty from another organization", func(t *testing.T) {
		org2ThirdPartyID := factory.NewThirdParty(org2Owner).WithName("Org2 ThirdParty").Create()

		_, err := org1Owner.Do(`
			mutation($input: CreateProcessingActivityInput!) {
				createProcessingActivity(input: $input) {
					processingActivityEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId":                       org1Owner.GetOrganizationID().String(),
				"name":                                 factory.SafeName("ProcessingActivity"),
				"specialOrCriminalData":                "NO",
				"lawfulBasis":                          "CONSENT",
				"internationalTransfers":               false,
				"dataProtectionImpactAssessmentNeeded": "NOT_NEEDED",
				"transferImpactAssessmentNeeded":       "NOT_NEEDED",
				"role":                                 "CONTROLLER",
				"thirdPartyIds":                        []string{org2ThirdPartyID},
			},
		})
		require.Error(t, err, "must not accept a thirdPartyId belonging to another organization")
	})

	t.Run("cannot update processing activity to reference a thirdParty from another organization", func(t *testing.T) {
		org2ThirdPartyID := factory.NewThirdParty(org2Owner).WithName("Org2 ThirdParty for Update").Create()
		otherPaID := factory.NewProcessingActivity(org1Owner).WithName("Org1 PA Without ThirdParties").Create()

		_, err := org1Owner.Do(`
			mutation($input: UpdateProcessingActivityInput!) {
				updateProcessingActivity(input: $input) {
					processingActivity { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":            otherPaID,
				"thirdPartyIds": []string{org2ThirdPartyID},
			},
		})
		require.Error(t, err, "must not accept a thirdPartyId belonging to another organization")
	})
}
