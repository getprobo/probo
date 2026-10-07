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

const (
	connectorAccountAWSRoleARN = "arn:aws:iam::123456789012:role/ProboAudit"
	connectorAccountAWSAccount = "123456789012"
)

const connectorAccountsQuery = `
	query($id: ID!) {
		node(id: $id) {
			... on Connector {
				id
				accounts {
					totalCount
					edges {
						node {
							id
							externalAccountId
							name
						}
					}
				}
			}
		}
	}
`

const disableConnectorAccountMutation = `
	mutation($input: DisableConnectorAccountInput!) {
		disableConnectorAccount(input: $input) {
			disabledConnectorAccountId
		}
	}
`

const accessReviewSourceQuery = `
	query($id: ID!) {
		node(id: $id) {
			... on AccessReviewSource {
				id
				connectorId
				connectorAccountId
			}
		}
	}
`

type accessReviewSourceResult struct {
	Node struct {
		ID                 string  `json:"id"`
		ConnectorID        *string `json:"connectorId"`
		ConnectorAccountID *string `json:"connectorAccountId"`
	} `json:"node"`
}

type connectorAccountsResult struct {
	Node struct {
		ID       string `json:"id"`
		Accounts struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Node struct {
					ID                string `json:"id"`
					ExternalAccountID string `json:"externalAccountId"`
					Name              string `json:"name"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"accounts"`
	} `json:"node"`
}

func TestConnectorAccounts_StandaloneAWSHasInitialAccount(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	connectorID := factory.NewConnector(owner).
		WithAWSRoleARN(connectorAccountAWSRoleARN).
		Create()

	var result connectorAccountsResult

	err := owner.Execute(connectorAccountsQuery, map[string]any{"id": connectorID}, &result)
	require.NoError(t, err)
	require.Equal(t, 1, result.Node.Accounts.TotalCount)
	require.Len(t, result.Node.Accounts.Edges, 1)
	assert.Equal(t, connectorAccountAWSAccount, result.Node.Accounts.Edges[0].Node.ExternalAccountID)
}

func TestEnableConnectorAccounts_AddsRowWithoutTouchingSource(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	connectorID := factory.NewConnector(owner).
		WithAWSRoleARN(connectorAccountAWSRoleARN).
		Create()

	var accounts connectorAccountsResult

	err := owner.Execute(connectorAccountsQuery, map[string]any{"id": connectorID}, &accounts)
	require.NoError(t, err)
	require.Len(t, accounts.Node.Accounts.Edges, 1)

	initialID := accounts.Node.Accounts.Edges[0].Node.ID
	sourceID := factory.NewAccessReviewSource(owner, owner.GetOrganizationID().String()).
		WithName("AWS prod").
		WithConnectorID(connectorID).
		WithConnectorAccountID(initialID).
		Create()
	require.NotEmpty(t, sourceID)

	memberID := factory.NewConnectorAccount(owner, connectorID).
		WithExternalAccountID("111111111111").
		WithName("Member").
		Create()
	require.NotEmpty(t, memberID)

	var result connectorAccountsResult

	err = owner.Execute(connectorAccountsQuery, map[string]any{"id": connectorID}, &result)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Node.Accounts.TotalCount)

	externals := make([]string, 0, len(result.Node.Accounts.Edges))
	for _, edge := range result.Node.Accounts.Edges {
		externals = append(externals, edge.Node.ExternalAccountID)
	}

	assert.ElementsMatch(t, []string{connectorAccountAWSAccount, "111111111111"}, externals)

	var source accessReviewSourceResult

	err = owner.Execute(accessReviewSourceQuery, map[string]any{"id": sourceID}, &source)
	require.NoError(t, err)
	require.NotNil(t, source.Node.ConnectorID)
	assert.Equal(t, connectorID, *source.Node.ConnectorID)
	require.NotNil(t, source.Node.ConnectorAccountID)
	assert.Equal(t, initialID, *source.Node.ConnectorAccountID)
}

func TestDisableConnectorAccount_RefusedWhenSourceReferences(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	connectorID := factory.NewConnector(owner).
		WithAWSRoleARN(connectorAccountAWSRoleARN).
		Create()

	var accounts connectorAccountsResult

	err := owner.Execute(connectorAccountsQuery, map[string]any{"id": connectorID}, &accounts)
	require.NoError(t, err)
	require.Len(t, accounts.Node.Accounts.Edges, 1)

	impliedID := accounts.Node.Accounts.Edges[0].Node.ID
	sourceID := factory.NewAccessReviewSource(owner, owner.GetOrganizationID().String()).
		WithName("AWS prod").
		WithConnectorID(connectorID).
		WithConnectorAccountID(impliedID).
		Create()
	require.NotEmpty(t, sourceID)

	err = owner.Execute(
		disableConnectorAccountMutation,
		map[string]any{"input": map[string]any{"connectorAccountId": impliedID}},
		&struct {
			DisableConnectorAccount struct {
				DisabledConnectorAccountID string `json:"disabledConnectorAccountId"`
			} `json:"disableConnectorAccount"`
		}{},
	)
	testutil.RequireErrorCode(t, err, "CONFLICT")

	memberID := factory.NewConnectorAccount(owner, connectorID).
		WithExternalAccountID("111111111111").
		WithName("Member").
		Create()

	var disabled struct {
		DisableConnectorAccount struct {
			DisabledConnectorAccountID string `json:"disabledConnectorAccountId"`
		} `json:"disableConnectorAccount"`
	}

	err = owner.Execute(
		disableConnectorAccountMutation,
		map[string]any{"input": map[string]any{"connectorAccountId": memberID}},
		&disabled,
	)
	require.NoError(t, err)
	assert.Equal(t, memberID, disabled.DisableConnectorAccount.DisabledConnectorAccountID)
}

func TestCreateAccessReviewSource_ConnectorIdOnlyResolvesAccount(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	connectorID := factory.NewConnector(owner).
		WithAWSRoleARN(connectorAccountAWSRoleARN).
		Create()

	sourceID := factory.NewAccessReviewSource(owner, owner.GetOrganizationID().String()).
		WithName("AWS dual-write").
		WithConnectorID(connectorID).
		Create()
	require.NotEmpty(t, sourceID)

	var result accessReviewSourceResult

	err := owner.Execute(accessReviewSourceQuery, map[string]any{"id": sourceID}, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Node.ConnectorID)
	assert.Equal(t, connectorID, *result.Node.ConnectorID)
	require.NotNil(t, result.Node.ConnectorAccountID)
	assert.NotEmpty(t, *result.Node.ConnectorAccountID)
}

func TestAccessReviewDrivers_OrganizationInstallSupported(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const query = `
		query {
			accessReviewDrivers {
				provider
				organizationInstallSupported
			}
		}
	`

	var result struct {
		AccessReviewDrivers []struct {
			Provider                     string `json:"provider"`
			OrganizationInstallSupported bool   `json:"organizationInstallSupported"`
		} `json:"accessReviewDrivers"`
	}

	err := owner.Execute(query, nil, &result)
	require.NoError(t, err)
	require.NotEmpty(t, result.AccessReviewDrivers)

	supported := map[string]bool{}
	for _, driver := range result.AccessReviewDrivers {
		supported[driver.Provider] = driver.OrganizationInstallSupported
	}

	assert.True(t, supported["AWS"])
	assert.True(t, supported["GCP"])
	assert.True(t, supported["AZURE"])
}

func TestDiscoveredAccounts_ViewerForbidden(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	connectorID := factory.NewConnector(owner).
		WithAWSRoleARN(connectorAccountAWSRoleARN).
		Create()

	_, err := viewer.Do(`
		query($id: ID!) {
			node(id: $id) {
				... on Connector {
					discoveredAccounts {
						externalAccountId
					}
				}
			}
		}
	`, map[string]any{"id": connectorID})
	testutil.RequireForbiddenError(t, err, "viewer should not discover connector accounts")
}
