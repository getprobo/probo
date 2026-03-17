// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

const testCsvData = "email,full_name,role,job_title,is_admin,mfa_status,auth_method,last_login,account_created_at,external_id\njane@example.com,Jane Smith,admin,CTO,true,ENABLED,SSO,2026-01-15T00:00:00Z,2024-06-01T00:00:00Z,ext-jane"

func TestAccessReview_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	const query = `
		mutation($input: CreateAccessReviewInput!) {
			createAccessReview(input: $input) {
				accessReview {
					id
					createdAt
					updatedAt
				}
			}
		}
	`

	var result struct {
		CreateAccessReview struct {
			AccessReview struct {
				ID        string `json:"id"`
				CreatedAt string `json:"createdAt"`
				UpdatedAt string `json:"updatedAt"`
			} `json:"accessReview"`
		} `json:"createAccessReview"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
		},
	}, &result)
	require.NoError(t, err)

	assert.NotEmpty(t, result.CreateAccessReview.AccessReview.ID)
	assert.NotEmpty(t, result.CreateAccessReview.AccessReview.CreatedAt)
	assert.NotEmpty(t, result.CreateAccessReview.AccessReview.UpdatedAt)
}

func TestAccessReview_NodeQuery(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	accessReviewID := factory.CreateAccessReview(owner)

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessReview {
					id
					organization {
						id
					}
					createdAt
					updatedAt
				}
			}
		}
	`

	var result struct {
		Node struct {
			ID           string `json:"id"`
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
			CreatedAt string `json:"createdAt"`
			UpdatedAt string `json:"updatedAt"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": accessReviewID}, &result)
	require.NoError(t, err)

	assert.Equal(t, accessReviewID, result.Node.ID)
	assert.Equal(t, owner.GetOrganizationID().String(), result.Node.Organization.ID)
}

func TestAccessReview_UpdateIdentitySource(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	accessReviewID := factory.CreateAccessReview(owner)
	sourceID := factory.NewAccessSource(owner, accessReviewID).
		WithName("Identity Source").
		Create()

	const query = `
		mutation($input: UpdateAccessReviewInput!) {
			updateAccessReview(input: $input) {
				accessReview {
					id
					identitySource {
						id
						name
					}
				}
			}
		}
	`

	var result struct {
		UpdateAccessReview struct {
			AccessReview struct {
				ID             string `json:"id"`
				IdentitySource *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"identitySource"`
			} `json:"accessReview"`
		} `json:"updateAccessReview"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"accessReviewId":   accessReviewID,
			"identitySourceId": sourceID,
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, accessReviewID, result.UpdateAccessReview.AccessReview.ID)
	require.NotNil(t, result.UpdateAccessReview.AccessReview.IdentitySource)
	assert.Equal(t, sourceID, result.UpdateAccessReview.AccessReview.IdentitySource.ID)
	assert.Equal(t, "Identity Source", result.UpdateAccessReview.AccessReview.IdentitySource.Name)
}

func TestAccessSource_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	accessReviewID := factory.CreateAccessReview(owner)

	t.Run("with name only", func(t *testing.T) {
		const query = `
			mutation($input: CreateAccessSourceInput!) {
				createAccessSource(input: $input) {
					accessSourceEdge {
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
			CreateAccessSource struct {
				AccessSourceEdge struct {
					Node struct {
						ID        string `json:"id"`
						Name      string `json:"name"`
						CreatedAt string `json:"createdAt"`
						UpdatedAt string `json:"updatedAt"`
					} `json:"node"`
				} `json:"accessSourceEdge"`
			} `json:"createAccessSource"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"accessReviewId": accessReviewID,
				"name":           "Slack",
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateAccessSource.AccessSourceEdge.Node
		assert.NotEmpty(t, node.ID)
		assert.Equal(t, "Slack", node.Name)
		assert.NotEmpty(t, node.CreatedAt)
	})

	t.Run("with csv data", func(t *testing.T) {
		const query = `
			mutation($input: CreateAccessSourceInput!) {
				createAccessSource(input: $input) {
					accessSourceEdge {
						node {
							id
							name
							csvData
						}
					}
				}
			}
		`

		var result struct {
			CreateAccessSource struct {
				AccessSourceEdge struct {
					Node struct {
						ID      string  `json:"id"`
						Name    string  `json:"name"`
						CsvData *string `json:"csvData"`
					} `json:"node"`
				} `json:"accessSourceEdge"`
			} `json:"createAccessSource"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"accessReviewId": accessReviewID,
				"name":           "CSV Import",
				"csvData":        testCsvData,
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateAccessSource.AccessSourceEdge.Node
		assert.NotEmpty(t, node.ID)
		assert.Equal(t, "CSV Import", node.Name)
		require.NotNil(t, node.CsvData)
		assert.Contains(t, *node.CsvData, "jane@example.com")
	})
}

func TestAccessSource_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	accessReviewID := factory.CreateAccessReview(owner)
	sourceID := factory.NewAccessSource(owner, accessReviewID).
		WithName("Original Source").
		Create()

	const query = `
		mutation($input: UpdateAccessSourceInput!) {
			updateAccessSource(input: $input) {
				accessSource {
					id
					name
				}
			}
		}
	`

	var result struct {
		UpdateAccessSource struct {
			AccessSource struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"accessSource"`
		} `json:"updateAccessSource"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"accessSourceId": sourceID,
			"name":           "Updated Source",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, sourceID, result.UpdateAccessSource.AccessSource.ID)
	assert.Equal(t, "Updated Source", result.UpdateAccessSource.AccessSource.Name)
}

func TestAccessSource_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	accessReviewID := factory.CreateAccessReview(owner)
	sourceID := factory.NewAccessSource(owner, accessReviewID).
		WithName("Source to Delete").
		Create()

	const query = `
		mutation($input: DeleteAccessSourceInput!) {
			deleteAccessSource(input: $input) {
				deletedAccessSourceId
			}
		}
	`

	var result struct {
		DeleteAccessSource struct {
			DeletedAccessSourceID string `json:"deletedAccessSourceId"`
		} `json:"deleteAccessSource"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"accessSourceId": sourceID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, sourceID, result.DeleteAccessSource.DeletedAccessSourceID)
}

func TestAccessSource_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	accessReviewID := factory.CreateAccessReview(owner)

	for _, name := range []string{"Slack", "GitHub", "Google Workspace"} {
		factory.NewAccessSource(owner, accessReviewID).WithName(name).Create()
	}

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessReview {
					accessSources(first: 10) {
						edges {
							node {
								id
								name
							}
						}
						totalCount
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			AccessSources struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"accessSources"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": accessReviewID}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.AccessSources.TotalCount, 3)
}

func TestAccessReviewCampaign_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	accessReviewID := factory.CreateAccessReview(owner)

	t.Run("with name only", func(t *testing.T) {
		const query = `
			mutation($input: CreateAccessReviewCampaignInput!) {
				createAccessReviewCampaign(input: $input) {
					accessReviewCampaignEdge {
						node {
							id
							name
							status
							createdAt
							updatedAt
						}
					}
				}
			}
		`

		var result struct {
			CreateAccessReviewCampaign struct {
				AccessReviewCampaignEdge struct {
					Node struct {
						ID        string `json:"id"`
						Name      string `json:"name"`
						Status    string `json:"status"`
						CreatedAt string `json:"createdAt"`
						UpdatedAt string `json:"updatedAt"`
					} `json:"node"`
				} `json:"accessReviewCampaignEdge"`
			} `json:"createAccessReviewCampaign"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"accessReviewId": accessReviewID,
				"name":           "Q1 2026 Review",
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateAccessReviewCampaign.AccessReviewCampaignEdge.Node
		assert.NotEmpty(t, node.ID)
		assert.Equal(t, "Q1 2026 Review", node.Name)
		assert.Equal(t, "DRAFT", node.Status)
		assert.NotEmpty(t, node.CreatedAt)
	})

	t.Run("with framework controls", func(t *testing.T) {
		const query = `
			mutation($input: CreateAccessReviewCampaignInput!) {
				createAccessReviewCampaign(input: $input) {
					accessReviewCampaignEdge {
						node {
							id
							name
							frameworkControls
						}
					}
				}
			}
		`

		var result struct {
			CreateAccessReviewCampaign struct {
				AccessReviewCampaignEdge struct {
					Node struct {
						ID                string   `json:"id"`
						Name              string   `json:"name"`
						FrameworkControls []string `json:"frameworkControls"`
					} `json:"node"`
				} `json:"accessReviewCampaignEdge"`
			} `json:"createAccessReviewCampaign"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"accessReviewId":    accessReviewID,
				"name":              "SOC2 Campaign",
				"frameworkControls": []string{"CC6.1", "CC6.2"},
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateAccessReviewCampaign.AccessReviewCampaignEdge.Node
		assert.NotEmpty(t, node.ID)
		assert.Equal(t, "SOC2 Campaign", node.Name)
		assert.Contains(t, node.FrameworkControls, "CC6.1")
		assert.Contains(t, node.FrameworkControls, "CC6.2")
	})
}

func TestAccessReviewCampaign_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	accessReviewID := factory.CreateAccessReview(owner)
	campaignID := factory.NewAccessReviewCampaign(owner, accessReviewID).
		WithName("Original Campaign").
		Create()

	const query = `
		mutation($input: UpdateAccessReviewCampaignInput!) {
			updateAccessReviewCampaign(input: $input) {
				accessReviewCampaign {
					id
					name
				}
			}
		}
	`

	var result struct {
		UpdateAccessReviewCampaign struct {
			AccessReviewCampaign struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"accessReviewCampaign"`
		} `json:"updateAccessReviewCampaign"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"accessReviewCampaignId": campaignID,
			"name":                   "Renamed Campaign",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, campaignID, result.UpdateAccessReviewCampaign.AccessReviewCampaign.ID)
	assert.Equal(t, "Renamed Campaign", result.UpdateAccessReviewCampaign.AccessReviewCampaign.Name)
}

func TestAccessReviewCampaign_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	accessReviewID := factory.CreateAccessReview(owner)
	campaignID := factory.NewAccessReviewCampaign(owner, accessReviewID).
		WithName("Campaign to Delete").
		Create()

	const query = `
		mutation($input: DeleteAccessReviewCampaignInput!) {
			deleteAccessReviewCampaign(input: $input) {
				deletedAccessReviewCampaignId
			}
		}
	`

	var result struct {
		DeleteAccessReviewCampaign struct {
			DeletedAccessReviewCampaignID string `json:"deletedAccessReviewCampaignId"`
		} `json:"deleteAccessReviewCampaign"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"accessReviewCampaignId": campaignID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, campaignID, result.DeleteAccessReviewCampaign.DeletedAccessReviewCampaignID)
}

func TestAccessReviewCampaign_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	accessReviewID := factory.CreateAccessReview(owner)

	for _, name := range []string{"Q1 Review", "Q2 Review", "Q3 Review"} {
		factory.NewAccessReviewCampaign(owner, accessReviewID).WithName(name).Create()
	}

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessReview {
					campaigns(first: 10) {
						edges {
							node {
								id
								name
								status
							}
						}
						totalCount
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Campaigns struct {
				Edges []struct {
					Node struct {
						ID     string `json:"id"`
						Name   string `json:"name"`
						Status string `json:"status"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"campaigns"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": accessReviewID}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Campaigns.TotalCount, 3)

	for _, edge := range result.Node.Campaigns.Edges {
		assert.Equal(t, "DRAFT", edge.Node.Status)
	}
}

func TestAccessReviewCampaign_NodeQuery(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	accessReviewID := factory.CreateAccessReview(owner)
	campaignID := factory.NewAccessReviewCampaign(owner, accessReviewID).
		WithName("Node Query Campaign").
		Create()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessReviewCampaign {
					id
					name
					status
					accessReview {
						id
					}
					statistics {
						totalCount
					}
					createdAt
					updatedAt
				}
			}
		}
	`

	var result struct {
		Node struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Status       string `json:"status"`
			AccessReview struct {
				ID string `json:"id"`
			} `json:"accessReview"`
			Statistics struct {
				TotalCount int `json:"totalCount"`
			} `json:"statistics"`
			CreatedAt string `json:"createdAt"`
			UpdatedAt string `json:"updatedAt"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": campaignID}, &result)
	require.NoError(t, err)

	assert.Equal(t, campaignID, result.Node.ID)
	assert.Equal(t, "Node Query Campaign", result.Node.Name)
	assert.Equal(t, "DRAFT", result.Node.Status)
	assert.Equal(t, accessReviewID, result.Node.AccessReview.ID)
	assert.Equal(t, 0, result.Node.Statistics.TotalCount)
}

func TestAccessReviewCampaign_StartWithCsvSource(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	accessReviewID := factory.CreateAccessReview(owner)

	sourceID := factory.NewAccessSource(owner, accessReviewID).
		WithName("CSV Test Source").
		WithCsvData(testCsvData).
		Create()

	campaignID := factory.NewAccessReviewCampaign(owner, accessReviewID).
		WithName("CSV Campaign").
		Create()

	const query = `
		mutation($input: StartAccessReviewCampaignInput!) {
			startAccessReviewCampaign(input: $input) {
				accessReviewCampaign {
					id
					status
					startedAt
				}
			}
		}
	`

	var result struct {
		StartAccessReviewCampaign struct {
			AccessReviewCampaign struct {
				ID        string  `json:"id"`
				Status    string  `json:"status"`
				StartedAt *string `json:"startedAt"`
			} `json:"accessReviewCampaign"`
		} `json:"startAccessReviewCampaign"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"accessReviewCampaignId": campaignID,
			"accessSourceIds":        []string{sourceID},
		},
	}, &result)
	require.NoError(t, err)

	campaign := result.StartAccessReviewCampaign.AccessReviewCampaign
	assert.Equal(t, campaignID, campaign.ID)
	assert.Equal(t, "IN_PROGRESS", campaign.Status)
	assert.NotNil(t, campaign.StartedAt)
}

func TestAccessReviewCampaign_Cancel(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	accessReviewID := factory.CreateAccessReview(owner)

	sourceID := factory.NewAccessSource(owner, accessReviewID).
		WithName("Cancel Test Source").
		WithCsvData(testCsvData).
		Create()

	campaignID := factory.NewAccessReviewCampaign(owner, accessReviewID).
		WithName("Campaign to Cancel").
		Create()

	// Start the campaign first
	const startQuery = `
		mutation($input: StartAccessReviewCampaignInput!) {
			startAccessReviewCampaign(input: $input) {
				accessReviewCampaign { id status }
			}
		}
	`

	err := owner.Execute(startQuery, map[string]any{
		"input": map[string]any{
			"accessReviewCampaignId": campaignID,
			"accessSourceIds":        []string{sourceID},
		},
	}, nil)
	require.NoError(t, err)

	// Cancel it
	const cancelQuery = `
		mutation($input: CancelAccessReviewCampaignInput!) {
			cancelAccessReviewCampaign(input: $input) {
				accessReviewCampaign {
					id
					status
				}
			}
		}
	`

	var result struct {
		CancelAccessReviewCampaign struct {
			AccessReviewCampaign struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"accessReviewCampaign"`
		} `json:"cancelAccessReviewCampaign"`
	}

	err = owner.Execute(cancelQuery, map[string]any{
		"input": map[string]any{
			"accessReviewCampaignId": campaignID,
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, campaignID, result.CancelAccessReviewCampaign.AccessReviewCampaign.ID)
	assert.Equal(t, "CANCELLED", result.CancelAccessReviewCampaign.AccessReviewCampaign.Status)
}

func TestAccessReview_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	accessReviewID := factory.CreateAccessReview(org1Owner)

	t.Run("cannot read access review from another organization", func(t *testing.T) {
		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on AccessReview {
						id
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID string `json:"id"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": accessReviewID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "access review")
	})

	t.Run("cannot create access source in another org's access review", func(t *testing.T) {
		const query = `
			mutation($input: CreateAccessSourceInput!) {
				createAccessSource(input: $input) {
					accessSourceEdge {
						node { id }
					}
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"accessReviewId": accessReviewID,
				"name":           "Unauthorized Source",
			},
		})
		require.Error(t, err, "Should not be able to create access source in another org's review")
	})

	t.Run("cannot create campaign in another org's access review", func(t *testing.T) {
		const query = `
			mutation($input: CreateAccessReviewCampaignInput!) {
				createAccessReviewCampaign(input: $input) {
					accessReviewCampaignEdge {
						node { id }
					}
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"accessReviewId": accessReviewID,
				"name":           "Unauthorized Campaign",
			},
		})
		require.Error(t, err, "Should not be able to create campaign in another org's review")
	})
}
