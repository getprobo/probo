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
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestOAuth2AccessToken_CreateListUseRevoke(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	organizationID := owner.GetOrganizationID().String()
	expiresAt := time.Now().Add(90 * 24 * time.Hour).UTC().Format(time.RFC3339)

	const createMutation = `
		mutation CreateOAuth2AccessToken($input: CreateOAuth2AccessTokenInput!) {
			createOAuth2AccessToken(input: $input) {
				token
				oauth2AccessTokenEdge {
					node {
						id
						name
						scopes
					}
				}
			}
		}
	`

	createResp, err := owner.DoConnect(createMutation, map[string]any{
		"input": map[string]any{
			"name":      "E2E manual token",
			"expiresAt": expiresAt,
			"scopes":    []string{"v1:org:read"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)

	var createData struct {
		CreateOAuth2AccessToken struct {
			Token string `json:"token"`
			Edge  struct {
				Node struct {
					ID     string   `json:"id"`
					Name   string   `json:"name"`
					Scopes []string `json:"scopes"`
				} `json:"node"`
			} `json:"oauth2AccessTokenEdge"`
		} `json:"createOAuth2AccessToken"`
	}
	require.NoError(t, json.Unmarshal(createResp.Data, &createData))

	tokenID := createData.CreateOAuth2AccessToken.Edge.Node.ID
	tokenValue := createData.CreateOAuth2AccessToken.Token

	require.NotEmpty(t, tokenID)
	require.NotEmpty(t, tokenValue)
	assert.Equal(t, "E2E manual token", createData.CreateOAuth2AccessToken.Edge.Node.Name)
	assert.Equal(t, []string{"v1:org:read"}, createData.CreateOAuth2AccessToken.Edge.Node.Scopes)

	const listQuery = `
		query ListOAuth2AccessTokens {
			viewer {
				oauth2AccessTokens(first: 10) {
					totalCount
					edges {
						node {
							id
							name
						}
					}
				}
			}
		}
	`

	listResp, err := owner.DoConnect(listQuery, map[string]any{})
	require.NoError(t, err)
	require.NotNil(t, listResp)

	var listData struct {
		Viewer struct {
			OAuth2AccessTokens struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"oauth2AccessTokens"`
		} `json:"viewer"`
	}
	require.NoError(t, json.Unmarshal(listResp.Data, &listData))
	require.GreaterOrEqual(t, listData.Viewer.OAuth2AccessTokens.TotalCount, 1)

	found := false

	for _, edge := range listData.Viewer.OAuth2AccessTokens.Edges {
		if edge.Node.ID == tokenID {
			found = true

			assert.Equal(t, "E2E manual token", edge.Node.Name)
		}
	}

	require.True(t, found, "created token should appear in identity list")

	const getOrganizationQuery = `
		query GetOrganization($id: ID!) {
			node(id: $id) {
				... on Organization {
					id
					name
				}
			}
		}
	`

	allowedResp, err := testutil.ConsoleGraphQLWithAccessToken(
		t,
		tokenValue,
		getOrganizationQuery,
		map[string]any{"id": organizationID},
	)
	require.NoError(t, err)
	require.NotNil(t, allowedResp)

	const revokeMutation = `
		mutation RevokeOAuth2AccessToken($input: RevokeOAuth2AccessTokenInput!) {
			revokeOAuth2AccessToken(input: $input) {
				oauth2AccessTokenId
			}
		}
	`

	revokeResp, err := owner.DoConnect(revokeMutation, map[string]any{
		"input": map[string]any{
			"oauth2AccessTokenId": tokenID,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, revokeResp)

	deniedResp, err := testutil.ConsoleGraphQLWithAccessToken(
		t,
		tokenValue,
		getOrganizationQuery,
		map[string]any{"id": organizationID},
	)
	require.Error(t, err)
	require.Nil(t, deniedResp)
}
