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
	connectorAccountsQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Connector {
					accounts(first: 100) {
						totalCount
						edges { node { id externalAccountId name } }
					}
				}
			}
		}
	`

	discoveredAccountsQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Connector {
					discoveredAccounts { externalAccountId name enabled }
				}
			}
		}
	`

	disableConnectorAccountMutation = `
		mutation($input: DisableConnectorAccountInput!) {
			disableConnectorAccount(input: $input) {
				deletedConnectorAccountId
			}
		}
	`

	sourceAccountQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessReviewSource {
					connectorId
					connectorAccountId
					connectorAccount { id externalAccountId }
				}
			}
		}
	`
)

type connectorAccountsResult struct {
	Node struct {
		Accounts struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Node struct {
					ID                string  `json:"id"`
					ExternalAccountID *string `json:"externalAccountId"`
					Name              string  `json:"name"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"accounts"`
	} `json:"node"`
}

func readConnectorAccounts(t *testing.T, c *testutil.Client, connectorID string) connectorAccountsResult {
	t.Helper()

	var result connectorAccountsResult
	require.NoError(t, c.Execute(connectorAccountsQuery, map[string]any{"id": connectorID}, &result))

	return result
}

// TestConnectorAccount_Lifecycle walks the whole account lifecycle through the
// API: a connector arrives with the account its settings imply, enabling is
// idempotent per vendor identifier, several accounts coexist on one
// credential, and an unused account can be removed.
func TestConnectorAccount_Lifecycle(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	connectorID := factory.NewConnector(owner, orgID).Create()

	// Every connector that can back a source arrives with an account, so a
	// source never has to attach to a credential with nothing behind it.
	implied := readConnectorAccounts(t, owner, connectorID)
	require.Equal(t, 1, implied.Node.Accounts.TotalCount)
	assert.Nil(
		t,
		implied.Node.Accounts.Edges[0].Node.ExternalAccountID,
		"a credential the vendor never names must not invent an identifier",
	)

	first := factory.EnableConnectorAccounts(owner, orgID, connectorID, "111111111111")
	require.Len(t, first, 1)

	// Re-running enable over a discover list must not duplicate rows, which is
	// what lets a client enable without checking what it already recorded.
	again := factory.EnableConnectorAccounts(owner, orgID, connectorID, "111111111111")
	require.Len(t, again, 1)
	assert.Equal(t, first[0], again[0], "the account keeps its id, so a source pointing at it still does")

	second := factory.EnableConnectorAccounts(owner, orgID, connectorID, "222222222222")
	require.Len(t, second, 1)

	afterEnable := readConnectorAccounts(t, owner, connectorID)
	assert.Equal(t, 3, afterEnable.Node.Accounts.TotalCount, "the implied row plus two enabled accounts")

	var disabled struct{}
	require.NoError(t, owner.Execute(disableConnectorAccountMutation, map[string]any{
		"input": map[string]any{"connectorAccountId": second[0]},
	}, &disabled))

	afterDisable := readConnectorAccounts(t, owner, connectorID)
	assert.Equal(t, 2, afterDisable.Node.Accounts.TotalCount)
}

// TestConnectorAccount_TwoSourcesOnOneConnector is the product this work
// exists for: one credential, two accounts, two independent reviews.
func TestConnectorAccount_TwoSourcesOnOneConnector(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	connectorID := factory.NewConnector(owner, orgID).Create()

	accountIDs := factory.EnableConnectorAccounts(owner, orgID, connectorID, "111111111111", "222222222222")
	require.Len(t, accountIDs, 2)

	sourceIDs := make([]string, 0, 2)

	for i, accountID := range accountIDs {
		sourceIDs = append(sourceIDs, factory.NewAccessReviewSource(owner, orgID).
			WithName(factory.SafeName("Account")).
			WithConnectorID(connectorID).
			WithConnectorAccountID(accountID).
			Create())

		var result struct {
			Node struct {
				ConnectorAccountID *string `json:"connectorAccountId"`
				ConnectorAccount   *struct {
					ExternalAccountID *string `json:"externalAccountId"`
				} `json:"connectorAccount"`
			} `json:"node"`
		}

		require.NoError(t, owner.Execute(sourceAccountQuery, map[string]any{"id": sourceIDs[i]}, &result))
		require.NotNil(t, result.Node.ConnectorAccountID)
		require.NotNil(t, result.Node.ConnectorAccount)
		require.NotNil(t, result.Node.ConnectorAccount.ExternalAccountID)
		assert.Equal(t, accountIDs[i], *result.Node.ConnectorAccountID, "each source reviews its own account")
	}

	assert.NotEqual(t, sourceIDs[0], sourceIDs[1])
}

// TestConnectorAccount_DisableRefusedWhileSourced pins the refusal as a clean
// conflict rather than a raw foreign-key error: a review must not lose the
// account it reviews.
func TestConnectorAccount_DisableRefusedWhileSourced(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	connectorID := factory.NewConnector(owner, orgID).Create()

	accountIDs := factory.EnableConnectorAccounts(owner, orgID, connectorID, "111111111111")
	require.Len(t, accountIDs, 1)

	factory.NewAccessReviewSource(owner, orgID).
		WithName(factory.SafeName("Production")).
		WithConnectorID(connectorID).
		WithConnectorAccountID(accountIDs[0]).
		Create()

	var result struct{}

	err := owner.Execute(disableConnectorAccountMutation, map[string]any{
		"input": map[string]any{"connectorAccountId": accountIDs[0]},
	}, &result)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "23503", "the foreign key must not surface raw")
}

// TestConnectorAccount_SourceCreateResolvesSoleAccount pins the compatibility
// rule this phase rests on: the existing Connections flow sends a connector
// and no account, and the server resolves it rather than refusing or storing
// a null.
func TestConnectorAccount_SourceCreateResolvesSoleAccount(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	connectorID := factory.NewConnector(owner, orgID).Create()

	sourceID := factory.NewAccessReviewSource(owner, orgID).
		WithName(factory.SafeName("Implied")).
		WithConnectorID(connectorID).
		Create()

	var result struct {
		Node struct {
			ConnectorID        *string `json:"connectorId"`
			ConnectorAccountID *string `json:"connectorAccountId"`
		} `json:"node"`
	}

	require.NoError(t, owner.Execute(sourceAccountQuery, map[string]any{"id": sourceID}, &result))
	require.NotNil(t, result.Node.ConnectorID)
	require.NotNil(t, result.Node.ConnectorAccountID, "the server must resolve the account, never store null")
}

// TestConnectorAccount_DeletingConnectorRemovesAccounts pins the cascade the
// disconnect paths depend on. Access review, SCIM and Settings all delete
// connector rows without sweeping accounts first, so without it every one of
// them would start failing on a foreign key naming neither feature.
func TestConnectorAccount_DeletingConnectorRemovesAccounts(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	connectorID := factory.NewConnector(owner, orgID).Create()
	accountIDs := factory.EnableConnectorAccounts(owner, orgID, connectorID, "111111111111")
	require.Len(t, accountIDs, 1)

	const deleteConnectorMutation = `
		mutation($input: DeleteConnectorInput!) {
			deleteConnector(input: $input) { deletedConnectorId }
		}
	`

	var deleted struct{}
	require.NoError(t, owner.Execute(deleteConnectorMutation, map[string]any{
		"input": map[string]any{"connectorId": connectorID},
	}, &deleted))

	var account struct {
		Node *struct {
			ID string `json:"id"`
		} `json:"node"`
	}

	err := owner.Execute(
		`query($id: ID!) { node(id: $id) { ... on ConnectorAccount { id } } }`,
		map[string]any{"id": accountIDs[0]},
		&account,
	)
	if err == nil {
		assert.Nil(t, account.Node, "the account must not outlive its credential")
	}
}

// TestConnectorAccount_DiscoveredEmptyForNonOrgProvider pins that discovery is
// an honest empty rather than an error for the providers — most of them —
// that cover exactly one account by construction.
func TestConnectorAccount_DiscoveredEmptyForNonOrgProvider(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	connectorID := factory.NewConnector(owner, orgID).Create()

	var result struct {
		Node struct {
			DiscoveredAccounts []struct {
				ExternalAccountID string `json:"externalAccountId"`
			} `json:"discoveredAccounts"`
		} `json:"node"`
	}

	require.NoError(t, owner.Execute(discoveredAccountsQuery, map[string]any{"id": connectorID}, &result))
	assert.Empty(t, result.Node.DiscoveredAccounts)
}
