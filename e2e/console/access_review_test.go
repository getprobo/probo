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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const testCsvData = "email,full_name,role,job_title,is_admin,mfa_status,auth_method,last_login,account_created_at,external_id\njane@example.com,Jane Smith,admin,CTO,true,ENABLED,SSO,2026-01-15T00:00:00Z,2024-06-01T00:00:00Z,ext-jane"

func TestAccessSource_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	t.Run("with name only", func(t *testing.T) {
		t.Parallel()

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
				"organizationId": orgID,
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
		t.Parallel()

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
				"organizationId": orgID,
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
	orgID := owner.GetOrganizationID().String()
	sourceID := factory.NewAccessSource(owner, orgID).
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
	orgID := owner.GetOrganizationID().String()
	sourceID := factory.NewAccessSource(owner, orgID).
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
	orgID := owner.GetOrganizationID().String()

	for _, name := range []string{"Slack", "GitHub", "Google Workspace"} {
		factory.NewAccessSource(owner, orgID).WithName(name).Create()
	}

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
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

	err := owner.Execute(query, map[string]any{"id": orgID}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.AccessSources.TotalCount, 3)
}

func TestAccessReviewCampaign_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	t.Run("with name only", func(t *testing.T) {
		t.Parallel()

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
				"organizationId": orgID,
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

	t.Run("with access sources", func(t *testing.T) {
		t.Parallel()

		source1ID := factory.NewAccessSource(owner, orgID).
			WithName("Slack Source").
			Create()
		source2ID := factory.NewAccessSource(owner, orgID).
			WithName("GitHub Source").
			Create()

		const query = `
			mutation($input: CreateAccessReviewCampaignInput!) {
				createAccessReviewCampaign(input: $input) {
					accessReviewCampaignEdge {
						node {
							id
							name
							scopeSources {
								id
								name
							}
						}
					}
				}
			}
		`

		var result struct {
			CreateAccessReviewCampaign struct {
				AccessReviewCampaignEdge struct {
					Node struct {
						ID           string `json:"id"`
						Name         string `json:"name"`
						ScopeSources []struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"scopeSources"`
					} `json:"node"`
				} `json:"accessReviewCampaignEdge"`
			} `json:"createAccessReviewCampaign"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId":  orgID,
				"name":            "Campaign with Sources",
				"accessSourceIds": []string{source1ID, source2ID},
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateAccessReviewCampaign.AccessReviewCampaignEdge.Node
		assert.NotEmpty(t, node.ID)
		assert.Equal(t, "Campaign with Sources", node.Name)
		assert.Len(t, node.ScopeSources, 2)
	})

	t.Run("with framework controls", func(t *testing.T) {
		t.Parallel()

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
				"organizationId":    orgID,
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
	orgID := owner.GetOrganizationID().String()
	campaignID := factory.NewAccessReviewCampaign(owner, orgID).
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
	orgID := owner.GetOrganizationID().String()
	campaignID := factory.NewAccessReviewCampaign(owner, orgID).
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
	orgID := owner.GetOrganizationID().String()

	for _, name := range []string{"Q1 Review", "Q2 Review", "Q3 Review"} {
		factory.NewAccessReviewCampaign(owner, orgID).WithName(name).Create()
	}

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					accessReviewCampaigns(first: 10) {
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
			AccessReviewCampaigns struct {
				Edges []struct {
					Node struct {
						ID     string `json:"id"`
						Name   string `json:"name"`
						Status string `json:"status"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"accessReviewCampaigns"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": orgID}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.AccessReviewCampaigns.TotalCount, 3)

	for _, edge := range result.Node.AccessReviewCampaigns.Edges {
		assert.Equal(t, "DRAFT", edge.Node.Status)
	}
}

func TestAccessReviewCampaign_NodeQuery(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()
	campaignID := factory.NewAccessReviewCampaign(owner, orgID).
		WithName("Node Query Campaign").
		Create()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessReviewCampaign {
					id
					name
					status
					organization {
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
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
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
	assert.Equal(t, orgID, result.Node.Organization.ID)
	assert.Equal(t, 0, result.Node.Statistics.TotalCount)
}

func TestAccessReviewCampaign_StartWithCsvSource(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	sourceID := factory.NewAccessSource(owner, orgID).
		WithName("CSV Test Source").
		WithCsvData(testCsvData).
		Create()

	campaignID := factory.NewAccessReviewCampaign(owner, orgID).
		WithName("CSV Campaign").
		WithAccessSourceIDs([]string{sourceID}).
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
		},
	}, &result)
	require.NoError(t, err)

	campaign := result.StartAccessReviewCampaign.AccessReviewCampaign
	assert.Equal(t, campaignID, campaign.ID)
	assert.Equal(t, "IN_PROGRESS", campaign.Status)
	assert.NotNil(t, campaign.StartedAt)
}

func TestAccessReviewCampaign_AddAndRemoveScopeSource(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	sourceID := factory.NewAccessSource(owner, orgID).
		WithName("Scope Source").
		Create()

	campaignID := factory.NewAccessReviewCampaign(owner, orgID).
		WithName("Scope Management Campaign").
		Create()

	t.Run("add scope source", func(t *testing.T) {
		const query = `
			mutation($input: AddAccessReviewCampaignScopeSourceInput!) {
				addAccessReviewCampaignScopeSource(input: $input) {
					accessReviewCampaign {
						id
						scopeSources {
							id
							name
						}
					}
				}
			}
		`

		var result struct {
			AddAccessReviewCampaignScopeSource struct {
				AccessReviewCampaign struct {
					ID           string `json:"id"`
					ScopeSources []struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"scopeSources"`
				} `json:"accessReviewCampaign"`
			} `json:"addAccessReviewCampaignScopeSource"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"accessReviewCampaignId": campaignID,
				"accessSourceId":         sourceID,
			},
		}, &result)
		require.NoError(t, err)

		campaign := result.AddAccessReviewCampaignScopeSource.AccessReviewCampaign
		assert.Equal(t, campaignID, campaign.ID)
		assert.Len(t, campaign.ScopeSources, 1)
		assert.Equal(t, sourceID, campaign.ScopeSources[0].ID)
	})

	t.Run("remove scope source", func(t *testing.T) {
		const query = `
			mutation($input: RemoveAccessReviewCampaignScopeSourceInput!) {
				removeAccessReviewCampaignScopeSource(input: $input) {
					accessReviewCampaign {
						id
						scopeSources {
							id
						}
					}
				}
			}
		`

		var result struct {
			RemoveAccessReviewCampaignScopeSource struct {
				AccessReviewCampaign struct {
					ID           string `json:"id"`
					ScopeSources []struct {
						ID string `json:"id"`
					} `json:"scopeSources"`
				} `json:"accessReviewCampaign"`
			} `json:"removeAccessReviewCampaignScopeSource"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"accessReviewCampaignId": campaignID,
				"accessSourceId":         sourceID,
			},
		}, &result)
		require.NoError(t, err)

		campaign := result.RemoveAccessReviewCampaignScopeSource.AccessReviewCampaign
		assert.Equal(t, campaignID, campaign.ID)
		assert.Empty(t, campaign.ScopeSources)
	})
}

func TestAccessReviewCampaign_Cancel(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	sourceID := factory.NewAccessSource(owner, orgID).
		WithName("Cancel Test Source").
		WithCsvData(testCsvData).
		Create()

	campaignID := factory.NewAccessReviewCampaign(owner, orgID).
		WithName("Campaign to Cancel").
		WithAccessSourceIDs([]string{sourceID}).
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

func TestAccessReviewCampaign_Description(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	t.Run("create with description", func(t *testing.T) {
		t.Parallel()

		const query = `
			mutation($input: CreateAccessReviewCampaignInput!) {
				createAccessReviewCampaign(input: $input) {
					accessReviewCampaignEdge {
						node {
							id
							name
							description
						}
					}
				}
			}
		`

		var result struct {
			CreateAccessReviewCampaign struct {
				AccessReviewCampaignEdge struct {
					Node struct {
						ID          string `json:"id"`
						Name        string `json:"name"`
						Description string `json:"description"`
					} `json:"node"`
				} `json:"accessReviewCampaignEdge"`
			} `json:"createAccessReviewCampaign"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": orgID,
				"name":           "Q1 Review with Desc",
				"description":    "Quarterly review of all SaaS access",
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateAccessReviewCampaign.AccessReviewCampaignEdge.Node
		assert.Equal(t, "Q1 Review with Desc", node.Name)
		assert.Equal(t, "Quarterly review of all SaaS access", node.Description)
	})

	t.Run("update description", func(t *testing.T) {
		t.Parallel()

		campaignID := factory.NewAccessReviewCampaign(owner, orgID).
			WithName("Description Update Test").
			Create()

		const query = `
			mutation($input: UpdateAccessReviewCampaignInput!) {
				updateAccessReviewCampaign(input: $input) {
					accessReviewCampaign {
						id
						description
					}
				}
			}
		`

		var result struct {
			UpdateAccessReviewCampaign struct {
				AccessReviewCampaign struct {
					ID          string `json:"id"`
					Description string `json:"description"`
				} `json:"accessReviewCampaign"`
			} `json:"updateAccessReviewCampaign"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"accessReviewCampaignId": campaignID,
				"description":            "Updated description",
			},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, campaignID, result.UpdateAccessReviewCampaign.AccessReviewCampaign.ID)
		assert.Equal(t, "Updated description", result.UpdateAccessReviewCampaign.AccessReviewCampaign.Description)
	})
}

func TestAccessReviewCampaign_FullLifecycle(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	// Step 1: Create a CSV source with test data
	sourceID := factory.NewAccessSource(owner, orgID).
		WithName("Lifecycle Test Source").
		WithCsvData(testCsvData).
		Create()

	// Step 2: Create a campaign with a description and the source
	const createQuery = `
		mutation($input: CreateAccessReviewCampaignInput!) {
			createAccessReviewCampaign(input: $input) {
				accessReviewCampaignEdge {
					node {
						id
						name
						description
						status
						scopeSources {
							id
						}
					}
				}
			}
		}
	`

	var createResult struct {
		CreateAccessReviewCampaign struct {
			AccessReviewCampaignEdge struct {
				Node struct {
					ID           string `json:"id"`
					Name         string `json:"name"`
					Description  string `json:"description"`
					Status       string `json:"status"`
					ScopeSources []struct {
						ID string `json:"id"`
					} `json:"scopeSources"`
				} `json:"node"`
			} `json:"accessReviewCampaignEdge"`
		} `json:"createAccessReviewCampaign"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId":  orgID,
			"name":            "Full Lifecycle Campaign",
			"description":     "Testing the full lifecycle",
			"accessSourceIds": []string{sourceID},
		},
	}, &createResult)
	require.NoError(t, err)

	campaignNode := createResult.CreateAccessReviewCampaign.AccessReviewCampaignEdge.Node
	campaignID := campaignNode.ID
	assert.Equal(t, "DRAFT", campaignNode.Status)
	assert.Equal(t, "Testing the full lifecycle", campaignNode.Description)
	assert.Len(t, campaignNode.ScopeSources, 1)

	// Step 3: Start the campaign (triggers worker to fetch CSV data)
	const startQuery = `
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

	var startResult struct {
		StartAccessReviewCampaign struct {
			AccessReviewCampaign struct {
				ID        string  `json:"id"`
				Status    string  `json:"status"`
				StartedAt *string `json:"startedAt"`
			} `json:"accessReviewCampaign"`
		} `json:"startAccessReviewCampaign"`
	}

	err = owner.Execute(startQuery, map[string]any{
		"input": map[string]any{
			"accessReviewCampaignId": campaignID,
		},
	}, &startResult)
	require.NoError(t, err)
	assert.Equal(t, "IN_PROGRESS", startResult.StartAccessReviewCampaign.AccessReviewCampaign.Status)
	assert.NotNil(t, startResult.StartAccessReviewCampaign.AccessReviewCampaign.StartedAt)

	// Step 4: Wait for the worker to process entries and move campaign to PENDING_ACTIONS.
	// Poll the campaign status until it transitions.
	const nodeQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessReviewCampaign {
					id
					status
					entries(first: 100) {
						edges {
							node {
								id
								email
								fullName
								decision
							}
						}
						totalCount
					}
					statistics {
						totalCount
						decisionCounts {
							decision
							count
						}
					}
				}
			}
		}
	`

	type campaignQueryResult struct {
		Node struct {
			ID      string `json:"id"`
			Status  string `json:"status"`
			Entries struct {
				Edges []struct {
					Node struct {
						ID       string `json:"id"`
						Email    string `json:"email"`
						FullName string `json:"fullName"`
						Decision string `json:"decision"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"entries"`
			Statistics struct {
				TotalCount     int `json:"totalCount"`
				DecisionCounts []struct {
					Decision string `json:"decision"`
					Count    int    `json:"count"`
				} `json:"decisionCounts"`
			} `json:"statistics"`
		} `json:"node"`
	}

	var campaignResult campaignQueryResult
	require.Eventually(t, func() bool {
		err := owner.Execute(nodeQuery, map[string]any{"id": campaignID}, &campaignResult)
		if err != nil {
			return false
		}
		return campaignResult.Node.Status == "PENDING_ACTIONS"
	}, 60*time.Second, 1*time.Second, "campaign should transition to PENDING_ACTIONS")

	// Verify entries were created from CSV data
	assert.GreaterOrEqual(t, campaignResult.Node.Entries.TotalCount, 1)
	assert.Equal(t, campaignResult.Node.Entries.TotalCount, campaignResult.Node.Statistics.TotalCount)

	// All entries should be PENDING
	for _, edge := range campaignResult.Node.Entries.Edges {
		assert.Equal(t, "PENDING", edge.Node.Decision)
	}

	// Step 5: Record decisions on all entries
	const recordDecisionQuery = `
		mutation($input: RecordAccessEntryDecisionInput!) {
			recordAccessEntryDecision(input: $input) {
				accessEntry {
					id
					decision
					decidedAt
					decisionHistory {
						id
						decision
						decidedAt
					}
				}
			}
		}
	`

	for _, edge := range campaignResult.Node.Entries.Edges {
		var decisionResult struct {
			RecordAccessEntryDecision struct {
				AccessEntry struct {
					ID              string  `json:"id"`
					Decision        string  `json:"decision"`
					DecidedAt       *string `json:"decidedAt"`
					DecisionHistory []struct {
						ID       string `json:"id"`
						Decision string `json:"decision"`
					} `json:"decisionHistory"`
				} `json:"accessEntry"`
			} `json:"recordAccessEntryDecision"`
		}

		err = owner.Execute(recordDecisionQuery, map[string]any{
			"input": map[string]any{
				"accessEntryId": edge.Node.ID,
				"decision":      "APPROVED",
			},
		}, &decisionResult)
		require.NoError(t, err)

		entry := decisionResult.RecordAccessEntryDecision.AccessEntry
		assert.Equal(t, "APPROVED", entry.Decision)
		assert.NotNil(t, entry.DecidedAt)

		// Verify decision history was recorded
		assert.Len(t, entry.DecisionHistory, 1)
		assert.Equal(t, "APPROVED", entry.DecisionHistory[0].Decision)
	}

	// Step 6: Close the campaign
	const closeQuery = `
		mutation($input: CloseAccessReviewCampaignInput!) {
			closeAccessReviewCampaign(input: $input) {
				accessReviewCampaign {
					id
					status
					completedAt
				}
			}
		}
	`

	var closeResult struct {
		CloseAccessReviewCampaign struct {
			AccessReviewCampaign struct {
				ID          string  `json:"id"`
				Status      string  `json:"status"`
				CompletedAt *string `json:"completedAt"`
			} `json:"accessReviewCampaign"`
		} `json:"closeAccessReviewCampaign"`
	}

	err = owner.Execute(closeQuery, map[string]any{
		"input": map[string]any{
			"accessReviewCampaignId": campaignID,
		},
	}, &closeResult)
	require.NoError(t, err)

	closedCampaign := closeResult.CloseAccessReviewCampaign.AccessReviewCampaign
	assert.Equal(t, "COMPLETED", closedCampaign.Status)
	assert.NotNil(t, closedCampaign.CompletedAt)
}

func TestAccessReviewCampaign_CloseRequiresAllDecisions(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	sourceID := factory.NewAccessSource(owner, orgID).
		WithName("Close Guard Source").
		WithCsvData(testCsvData).
		Create()

	campaignID := factory.NewAccessReviewCampaign(owner, orgID).
		WithName("Close Guard Campaign").
		WithAccessSourceIDs([]string{sourceID}).
		Create()

	// Start the campaign
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
		},
	}, nil)
	require.NoError(t, err)

	// Wait for PENDING_ACTIONS
	const nodeQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessReviewCampaign { status }
			}
		}
	`

	require.Eventually(t, func() bool {
		var r struct {
			Node struct {
				Status string `json:"status"`
			} `json:"node"`
		}
		if err := owner.Execute(nodeQuery, map[string]any{"id": campaignID}, &r); err != nil {
			return false
		}
		return r.Node.Status == "PENDING_ACTIONS"
	}, 60*time.Second, 1*time.Second)

	// Try to close without deciding — should fail
	const closeQuery = `
		mutation($input: CloseAccessReviewCampaignInput!) {
			closeAccessReviewCampaign(input: $input) {
				accessReviewCampaign { id status }
			}
		}
	`

	_, err = owner.Do(closeQuery, map[string]any{
		"input": map[string]any{
			"accessReviewCampaignId": campaignID,
		},
	})
	require.Error(t, err, "closing a campaign with undecided entries should fail")
}

func TestAccessReviewCampaign_StartWithoutSourcesFails(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	campaignID := factory.NewAccessReviewCampaign(owner, orgID).
		WithName("No Sources Campaign").
		Create()

	const startQuery = `
		mutation($input: StartAccessReviewCampaignInput!) {
			startAccessReviewCampaign(input: $input) {
				accessReviewCampaign { id status }
			}
		}
	`

	_, err := owner.Do(startQuery, map[string]any{
		"input": map[string]any{
			"accessReviewCampaignId": campaignID,
		},
	})
	require.Error(t, err, "starting a campaign without sources should fail")
}

func TestAccessReview_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1ID := org1Owner.GetOrganizationID().String()

	t.Run("cannot create access source in another organization", func(t *testing.T) {
		t.Parallel()

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
				"organizationId": org1ID,
				"name":           "Unauthorized Source",
			},
		})
		require.Error(t, err, "Should not be able to create access source in another organization")
	})

	t.Run("cannot create campaign in another organization", func(t *testing.T) {
		t.Parallel()

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
				"organizationId": org1ID,
				"name":           "Unauthorized Campaign",
			},
		})
		require.Error(t, err, "Should not be able to create campaign in another organization")
	})
}
