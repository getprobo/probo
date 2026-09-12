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

package mcp_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const (
	createServiceAccountMutation = `
		mutation CreateServiceAccount($input: CreateServiceAccountInput!) {
			createServiceAccount(input: $input) {
				serviceAccountEdge {
					node {
						id
						organizationId
						name
						scopes
					}
				}
			}
		}
	`
	createServiceAccountCredentialMutation = `
		mutation CreateServiceAccountCredential($input: CreateServiceAccountCredentialInput!) {
			createServiceAccountCredential(input: $input) {
				serviceAccountCredentialEdge {
					node {
						id
						serviceAccountId
						name
						scopes
						expiresAt
						lastUsedAt
						revokedAt
						createdAt
					}
				}
				token
			}
		}
	`
	listServiceAccountCredentialsQuery = `
		query ListServiceAccountCredentials($id: ID!) {
			node(id: $id) {
				... on ServiceAccount {
					credentials(first: 10) {
						edges {
							node {
								id
								serviceAccountId
								name
								scopes
								expiresAt
								lastUsedAt
								revokedAt
								createdAt
							}
						}
					}
				}
			}
		}
	`
	listServiceAccountsQuery = `
		query ListServiceAccounts($id: ID!) {
			node(id: $id) {
				... on Organization {
					serviceAccounts(first: 50) {
						edges {
							node {
								id
							}
						}
					}
				}
			}
		}
	`
	getOrganizationQuery = `
		query GetOrganization($id: ID!) {
			node(id: $id) {
				... on Organization {
					id
					name
				}
			}
		}
	`
	listThirdPartiesQuery = `
		query ListThirdParties($id: ID!) {
			node(id: $id) {
				... on Organization {
					thirdParties(first: 1) {
						totalCount
					}
				}
			}
		}
	`
	revokeServiceAccountCredentialMutation = `
		mutation RevokeServiceAccountCredential($input: RevokeServiceAccountCredentialInput!) {
			revokeServiceAccountCredential(input: $input) {
				serviceAccountCredential {
					id
					revokedAt
				}
			}
		}
	`
	disableServiceAccountMutation = `
		mutation DisableServiceAccount($input: DisableServiceAccountInput!) {
			disableServiceAccount(input: $input) {
				serviceAccount {
					id
					disabledAt
				}
			}
		}
	`
	deleteServiceAccountMutation = `
		mutation DeleteServiceAccount($input: DeleteServiceAccountInput!) {
			deleteServiceAccount(input: $input) {
				deletedServiceAccountId
			}
		}
	`
)

type serviceAccountCredential struct {
	accountID    string
	credentialID string
	token        string
}

func TestServiceAccount_ScopedCredentials(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	otherOwner := testutil.NewClient(t, testutil.RoleOwner)
	organizationID := owner.GetOrganizationID().String()

	scoped := createServiceAccountCredential(
		t,
		owner,
		[]string{"v1:org:read"},
	)

	assertCredentialMetadataDoesNotExposeSecrets(t, owner, scoped)
	requireConsoleOrganizationAccess(t, scoped.token, organizationID)

	resp, err := testutil.ConsoleGraphQLWithAccessToken(
		t,
		scoped.token,
		listThirdPartiesQuery,
		map[string]any{"id": organizationID},
	)
	testutil.RequireForbiddenError(t, err, "scope must deny third-party reads")
	require.NotNil(t, resp)
	assert.Empty(t, resp.DataString())

	resp, err = testutil.ConsoleGraphQLWithAccessToken(
		t,
		scoped.token,
		getOrganizationQuery,
		map[string]any{"id": otherOwner.GetOrganizationID().String()},
	)
	testutil.RequireForbiddenError(t, err, "service account must not cross organizations")
	require.NotNil(t, resp)
	assert.Empty(t, resp.DataString())

	revokeServiceAccountCredential(t, owner, scoped)
	requireServiceAccountAuthenticationFailure(t, scoped.token, organizationID)

	managedTarget := createServiceAccountCredential(
		t,
		owner,
		[]string{"v1:org:read"},
	)

	iamScoped := createServiceAccountCredential(
		t,
		owner,
		[]string{"v1:iam"},
	)
	requireConsoleOrganizationAccess(t, iamScoped.token, organizationID)

	mc := testutil.NewMCPClientWithAccessToken(t, owner, iamScoped.token)
	msg := mc.CallToolExpectToolError("createServiceAccount", map[string]any{
		"organization_id": organizationID,
		"name":            factory.SafeName("Nested service account"),
		"scopes":          []string{"v1:iam"},
	})
	assert.Contains(t, strings.ToLower(msg), "insufficient permissions")

	msg = mc.CallToolExpectToolError("disableServiceAccount", map[string]any{
		"id": managedTarget.accountID,
	})
	assert.Contains(t, strings.ToLower(msg), "insufficient permissions")
	requireConsoleOrganizationAccess(t, managedTarget.token, organizationID)

	disableServiceAccount(t, owner, iamScoped.accountID)
	requireServiceAccountAuthenticationFailure(t, iamScoped.token, organizationID)

	deleteServiceAccount(t, owner, managedTarget.accountID)
	requireServiceAccountAbsent(t, owner, managedTarget.accountID)
	requireServiceAccountAuthenticationFailure(t, managedTarget.token, organizationID)
}

func createServiceAccountCredential(
	t *testing.T,
	owner *testutil.Client,
	scopes []string,
) serviceAccountCredential {
	t.Helper()

	var accountResult struct {
		CreateServiceAccount struct {
			ServiceAccountEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"serviceAccountEdge"`
		} `json:"createServiceAccount"`
	}
	err := owner.ExecuteConnect(
		createServiceAccountMutation,
		map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           factory.SafeName("Service account"),
				"scopes":         scopes,
			},
		},
		&accountResult,
	)
	require.NoError(t, err)

	accountID := accountResult.CreateServiceAccount.ServiceAccountEdge.Node.ID
	require.NotEmpty(t, accountID)
	t.Cleanup(
		func() {
			_, _ = owner.DoConnect(
				deleteServiceAccountMutation,
				map[string]any{
					"input": map[string]any{
						"serviceAccountId": accountID,
					},
				},
			)
		},
	)

	var credentialResult struct {
		CreateServiceAccountCredential struct {
			ServiceAccountCredentialEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"serviceAccountCredentialEdge"`
			Token string `json:"token"`
		} `json:"createServiceAccountCredential"`
	}
	err = owner.ExecuteConnect(
		createServiceAccountCredentialMutation,
		map[string]any{
			"input": map[string]any{
				"serviceAccountId": accountID,
				"name":             factory.SafeName("Credential"),
				"scopes":           scopes,
				"expiresAt":        time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
			},
		},
		&credentialResult,
	)
	require.NoError(t, err)

	credentialID := credentialResult.CreateServiceAccountCredential.ServiceAccountCredentialEdge.Node.ID
	require.NotEmpty(t, credentialID)
	require.NotEmpty(t, credentialResult.CreateServiceAccountCredential.Token)

	return serviceAccountCredential{
		accountID:    accountID,
		credentialID: credentialID,
		token:        credentialResult.CreateServiceAccountCredential.Token,
	}
}

func assertCredentialMetadataDoesNotExposeSecrets(
	t *testing.T,
	owner *testutil.Client,
	credential serviceAccountCredential,
) {
	t.Helper()

	resp, err := owner.DoConnect(
		listServiceAccountCredentialsQuery,
		map[string]any{"id": credential.accountID},
	)
	require.NoError(t, err)
	require.NotNil(t, resp)

	metadata := resp.DataString()
	require.Contains(t, metadata, credential.credentialID)
	assert.False(
		t,
		strings.Contains(metadata, credential.token),
		"credential metadata exposed the raw token",
	)
	assert.False(
		t,
		strings.Contains(strings.ToLower(metadata), "hash"),
		"credential metadata exposed a token hash",
	)
}

func requireConsoleOrganizationAccess(t *testing.T, token string, organizationID string) {
	t.Helper()

	resp, err := testutil.ConsoleGraphQLWithAccessToken(
		t,
		token,
		getOrganizationQuery,
		map[string]any{"id": organizationID},
	)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.DataString())
}

func revokeServiceAccountCredential(
	t *testing.T,
	owner *testutil.Client,
	credential serviceAccountCredential,
) {
	t.Helper()

	var result struct {
		RevokeServiceAccountCredential struct {
			ServiceAccountCredential struct {
				ID        string  `json:"id"`
				RevokedAt *string `json:"revokedAt"`
			} `json:"serviceAccountCredential"`
		} `json:"revokeServiceAccountCredential"`
	}
	err := owner.ExecuteConnect(
		revokeServiceAccountCredentialMutation,
		map[string]any{
			"input": map[string]any{
				"serviceAccountId":           credential.accountID,
				"serviceAccountCredentialId": credential.credentialID,
			},
		},
		&result,
	)
	require.NoError(t, err)
	assert.Equal(
		t,
		credential.credentialID,
		result.RevokeServiceAccountCredential.ServiceAccountCredential.ID,
	)
	require.NotNil(t, result.RevokeServiceAccountCredential.ServiceAccountCredential.RevokedAt)
}

func disableServiceAccount(t *testing.T, owner *testutil.Client, accountID string) {
	t.Helper()

	var result struct {
		DisableServiceAccount struct {
			ServiceAccount struct {
				ID         string  `json:"id"`
				DisabledAt *string `json:"disabledAt"`
			} `json:"serviceAccount"`
		} `json:"disableServiceAccount"`
	}
	err := owner.ExecuteConnect(
		disableServiceAccountMutation,
		map[string]any{
			"input": map[string]any{
				"serviceAccountId": accountID,
			},
		},
		&result,
	)
	require.NoError(t, err)
	assert.Equal(t, accountID, result.DisableServiceAccount.ServiceAccount.ID)
	require.NotNil(t, result.DisableServiceAccount.ServiceAccount.DisabledAt)
}

func deleteServiceAccount(t *testing.T, owner *testutil.Client, accountID string) {
	t.Helper()

	var result struct {
		DeleteServiceAccount struct {
			DeletedServiceAccountID string `json:"deletedServiceAccountId"`
		} `json:"deleteServiceAccount"`
	}
	err := owner.ExecuteConnect(
		deleteServiceAccountMutation,
		map[string]any{
			"input": map[string]any{
				"serviceAccountId": accountID,
			},
		},
		&result,
	)
	require.NoError(t, err)
	assert.Equal(t, accountID, result.DeleteServiceAccount.DeletedServiceAccountID)
}

func requireServiceAccountAbsent(t *testing.T, owner *testutil.Client, accountID string) {
	t.Helper()

	resp, err := owner.DoConnect(
		listServiceAccountsQuery,
		map[string]any{"id": owner.GetOrganizationID().String()},
	)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, strings.Contains(resp.DataString(), accountID), "deleted service account remained listed")
}

func requireServiceAccountAuthenticationFailure(t *testing.T, token string, organizationID string) {
	t.Helper()

	resp, err := testutil.ConsoleGraphQLWithAccessToken(
		t,
		token,
		getOrganizationQuery,
		map[string]any{"id": organizationID},
	)
	testutil.RequireErrorCode(t, err, "UNAUTHENTICATED")
	require.NotNil(t, resp)
	assert.Empty(t, resp.DataString())
}
