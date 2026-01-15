// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestRightsRequest_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	query := `
		mutation CreateRightsRequest($input: CreateRightsRequestInput!) {
			createRightsRequest(input: $input) {
				rightsRequestEdge {
					node {
						id
						requestType
						requestState
						dataSubject
						contact
						details
						actionTaken
					}
				}
			}
		}
	`

	var result struct {
		CreateRightsRequest struct {
			RightsRequestEdge struct {
				Node struct {
					ID           string `json:"id"`
					RequestType  string `json:"requestType"`
					RequestState string `json:"requestState"`
					DataSubject  string `json:"dataSubject"`
					Contact      string `json:"contact"`
					Details      string `json:"details"`
					ActionTaken  string `json:"actionTaken"`
				} `json:"node"`
			} `json:"rightsRequestEdge"`
		} `json:"createRightsRequest"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"requestType":    "ACCESS",
			"requestState":   "TODO",
			"dataSubject":    "John Doe",
			"contact":        "john.doe@example.com",
			"details":        "Request access to personal data",
			"actionTaken":    "Initial review completed",
		},
	}, &result)
	require.NoError(t, err)

	rr := result.CreateRightsRequest.RightsRequestEdge.Node
	assert.NotEmpty(t, rr.ID)
	assert.Equal(t, "ACCESS", rr.RequestType)
	assert.Equal(t, "TODO", rr.RequestState)
	assert.Equal(t, "John Doe", rr.DataSubject)
	assert.Equal(t, "john.doe@example.com", rr.Contact)
	assert.Equal(t, "Request access to personal data", rr.Details)
	assert.Equal(t, "Initial review completed", rr.ActionTaken)
}

func TestRightsRequest_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	createQuery := `
		mutation CreateRightsRequest($input: CreateRightsRequestInput!) {
			createRightsRequest(input: $input) {
				rightsRequestEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateRightsRequest struct {
			RightsRequestEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"rightsRequestEdge"`
		} `json:"createRightsRequest"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"requestType":    "ACCESS",
			"requestState":   "TODO",
			"dataSubject":    "Original Subject",
			"contact":        "original@example.com",
		},
	}, &createResult)
	require.NoError(t, err)
	rrID := createResult.CreateRightsRequest.RightsRequestEdge.Node.ID

	query := `
		mutation UpdateRightsRequest($input: UpdateRightsRequestInput!) {
			updateRightsRequest(input: $input) {
				rightsRequest {
					id
					requestType
					requestState
					dataSubject
					contact
					details
					actionTaken
				}
			}
		}
	`

	var result struct {
		UpdateRightsRequest struct {
			RightsRequest struct {
				ID           string `json:"id"`
				RequestType  string `json:"requestType"`
				RequestState string `json:"requestState"`
				DataSubject  string `json:"dataSubject"`
				Contact      string `json:"contact"`
				Details      string `json:"details"`
				ActionTaken  string `json:"actionTaken"`
			} `json:"rightsRequest"`
		} `json:"updateRightsRequest"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":           rrID,
			"requestType":  "DELETION",
			"requestState": "IN_PROGRESS",
			"dataSubject":  "Updated Subject",
			"contact":      "updated@example.com",
			"details":      "Updated details",
			"actionTaken":  "Processing deletion request",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, rrID, result.UpdateRightsRequest.RightsRequest.ID)
	assert.Equal(t, "DELETION", result.UpdateRightsRequest.RightsRequest.RequestType)
	assert.Equal(t, "IN_PROGRESS", result.UpdateRightsRequest.RightsRequest.RequestState)
	assert.Equal(t, "Updated Subject", result.UpdateRightsRequest.RightsRequest.DataSubject)
	assert.Equal(t, "updated@example.com", result.UpdateRightsRequest.RightsRequest.Contact)
	assert.Equal(t, "Updated details", result.UpdateRightsRequest.RightsRequest.Details)
	assert.Equal(t, "Processing deletion request", result.UpdateRightsRequest.RightsRequest.ActionTaken)
}

func TestRightsRequest_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	createQuery := `
		mutation CreateRightsRequest($input: CreateRightsRequestInput!) {
			createRightsRequest(input: $input) {
				rightsRequestEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateRightsRequest struct {
			RightsRequestEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"rightsRequestEdge"`
		} `json:"createRightsRequest"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"requestType":    "ACCESS",
			"requestState":   "TODO",
		},
	}, &createResult)
	require.NoError(t, err)
	rrID := createResult.CreateRightsRequest.RightsRequestEdge.Node.ID

	query := `
		mutation DeleteRightsRequest($input: DeleteRightsRequestInput!) {
			deleteRightsRequest(input: $input) {
				deletedRightsRequestId
			}
		}
	`

	var result struct {
		DeleteRightsRequest struct {
			DeletedRightsRequestID string `json:"deletedRightsRequestId"`
		} `json:"deleteRightsRequest"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"rightsRequestId": rrID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, rrID, result.DeleteRightsRequest.DeletedRightsRequestID)
}

func TestRightsRequest_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	createQuery := `
		mutation CreateRightsRequest($input: CreateRightsRequestInput!) {
			createRightsRequest(input: $input) {
				rightsRequestEdge {
					node {
						id
					}
				}
			}
		}
	`

	for i := 0; i < 3; i++ {
		_, err := owner.Do(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"requestType":    "ACCESS",
				"requestState":   "TODO",
				"dataSubject":    fmt.Sprintf("Subject %d", i),
			},
		})
		require.NoError(t, err)
	}

	query := `
		query GetRightsRequests($id: ID!) {
			node(id: $id) {
				... on Organization {
					rightsRequests(first: 10) {
						edges {
							node {
								id
								requestType
								requestState
								dataSubject
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
			RightsRequests struct {
				Edges []struct {
					Node struct {
						ID           string `json:"id"`
						RequestType  string `json:"requestType"`
						RequestState string `json:"requestState"`
						DataSubject  string `json:"dataSubject"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"rightsRequests"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.RightsRequests.TotalCount, 3)
}

func TestRightsRequest_TypeAndStateValues(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("request type values", func(t *testing.T) {
		types := []string{"ACCESS", "DELETION", "PORTABILITY"}

		for _, requestType := range types {
			t.Run(requestType, func(t *testing.T) {
				query := `
					mutation CreateRightsRequest($input: CreateRightsRequestInput!) {
						createRightsRequest(input: $input) {
							rightsRequestEdge {
								node {
									id
									requestType
								}
							}
						}
					}
				`

				var result struct {
					CreateRightsRequest struct {
						RightsRequestEdge struct {
							Node struct {
								ID          string `json:"id"`
								RequestType string `json:"requestType"`
							} `json:"node"`
						} `json:"rightsRequestEdge"`
					} `json:"createRightsRequest"`
				}

				err := owner.Execute(query, map[string]any{
					"input": map[string]any{
						"organizationId": owner.GetOrganizationID().String(),
						"requestType":    requestType,
						"requestState":   "TODO",
					},
				}, &result)
				require.NoError(t, err)
				assert.Equal(t, requestType, result.CreateRightsRequest.RightsRequestEdge.Node.RequestType)
			})
		}
	})

	t.Run("request state values", func(t *testing.T) {
		states := []string{"TODO", "IN_PROGRESS", "DONE"}

		for _, requestState := range states {
			t.Run(requestState, func(t *testing.T) {
				query := `
					mutation CreateRightsRequest($input: CreateRightsRequestInput!) {
						createRightsRequest(input: $input) {
							rightsRequestEdge {
								node {
									id
									requestState
								}
							}
						}
					}
				`

				var result struct {
					CreateRightsRequest struct {
						RightsRequestEdge struct {
							Node struct {
								ID           string `json:"id"`
								RequestState string `json:"requestState"`
							} `json:"node"`
						} `json:"rightsRequestEdge"`
					} `json:"createRightsRequest"`
				}

				err := owner.Execute(query, map[string]any{
					"input": map[string]any{
						"organizationId": owner.GetOrganizationID().String(),
						"requestType":    "ACCESS",
						"requestState":   requestState,
					},
				}, &result)
				require.NoError(t, err)
				assert.Equal(t, requestState, result.CreateRightsRequest.RightsRequestEdge.Node.RequestState)
			})
		}
	})
}
