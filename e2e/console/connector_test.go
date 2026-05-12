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
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestConnectorProviderInfos(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	t.Run("returns provider infos", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						connectorProviderInfos {
							provider
							displayName
							oauthConfigured
							apiKeySupported
							clientCredentialsSupported
							extraSettings {
								key
								label
								required
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ConnectorProviderInfos []struct {
					Provider                   string `json:"provider"`
					DisplayName                string `json:"displayName"`
					OauthConfigured            bool   `json:"oauthConfigured"`
					APIKeySupported            bool   `json:"apiKeySupported"`
					ClientCredentialsSupported bool   `json:"clientCredentialsSupported"`
					ExtraSettings              []struct {
						Key      string `json:"key"`
						Label    string `json:"label"`
						Required bool   `json:"required"`
					} `json:"extraSettings"`
				} `json:"connectorProviderInfos"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": orgID}, &result)
		require.NoError(t, err)

		infos := result.Node.ConnectorProviderInfos
		assert.NotEmpty(t, infos)

		providerNames := make(map[string]bool)
		for _, info := range infos {
			assert.NotEmpty(t, info.Provider)
			assert.NotEmpty(t, info.DisplayName)
			assert.NotNil(t, info.ExtraSettings)
			providerNames[info.Provider] = true
		}

		assert.True(t, providerNames["SLACK"], "expected SLACK provider to be present")
		assert.True(t, providerNames["HUBSPOT"], "expected HUBSPOT provider to be present")
	})

	t.Run("viewer can list provider infos", func(t *testing.T) {
		t.Parallel()
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						connectorProviderInfos {
							provider
							displayName
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ConnectorProviderInfos []struct {
					Provider    string `json:"provider"`
					DisplayName string `json:"displayName"`
				} `json:"connectorProviderInfos"`
			} `json:"node"`
		}

		err := viewer.Execute(query, map[string]any{
			"id": viewer.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Node.ConnectorProviderInfos)
	})
}

func TestCreateAPIKeyConnector(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	const query = `
		mutation($input: CreateAPIKeyConnectorInput!) {
			createAPIKeyConnector(input: $input) {
				connector {
					id
					provider
				}
			}
		}
	`

	var result struct {
		CreateAPIKeyConnector struct {
			Connector struct {
				ID       string `json:"id"`
				Provider string `json:"provider"`
			} `json:"connector"`
		} `json:"createAPIKeyConnector"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": orgID,
			"provider":       "BREX",
			"apiKey":         "test-key-123",
		},
	}, &result)
	require.NoError(t, err)

	connector := result.CreateAPIKeyConnector.Connector
	assert.NotEmpty(t, connector.ID)
	assert.Equal(t, "BREX", connector.Provider)
}

func TestCreateAPIKeyConnectorWithSettings(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	const query = `
		mutation($input: CreateAPIKeyConnectorInput!) {
			createAPIKeyConnector(input: $input) {
				connector {
					id
					provider
				}
			}
		}
	`

	var result struct {
		CreateAPIKeyConnector struct {
			Connector struct {
				ID       string `json:"id"`
				Provider string `json:"provider"`
			} `json:"connector"`
		} `json:"createAPIKeyConnector"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId":      orgID,
			"provider":            "TALLY",
			"apiKey":              "test-key",
			"tallyOrganizationId": "org-123",
		},
	}, &result)
	require.NoError(t, err)

	connector := result.CreateAPIKeyConnector.Connector
	assert.NotEmpty(t, connector.ID)
	assert.Equal(t, "TALLY", connector.Provider)
}

func TestCreateClientCredentialsConnector(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	const query = `
		mutation($input: CreateClientCredentialsConnectorInput!) {
			createClientCredentialsConnector(input: $input) {
				connector {
					id
					provider
				}
			}
		}
	`

	var result struct {
		CreateClientCredentialsConnector struct {
			Connector struct {
				ID       string `json:"id"`
				Provider string `json:"provider"`
			} `json:"connector"`
		} `json:"createClientCredentialsConnector"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId":       orgID,
			"provider":             "ONE_PASSWORD",
			"clientId":             "test-client",
			"clientSecret":         "test-secret",
			"tokenUrl":             "https://api.1password.com/v1beta1/users/oauth2/token",
			"onePasswordAccountId": "ACC123",
			"onePasswordRegion":    "US",
		},
	}, &result)
	require.NoError(t, err)

	connector := result.CreateClientCredentialsConnector.Connector
	assert.NotEmpty(t, connector.ID)
	assert.Equal(t, "ONE_PASSWORD", connector.Provider)
}

func TestDeleteConnector(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	// First, create a connector to delete.
	const createQuery = `
		mutation($input: CreateAPIKeyConnectorInput!) {
			createAPIKeyConnector(input: $input) {
				connector {
					id
					provider
				}
			}
		}
	`

	var createResult struct {
		CreateAPIKeyConnector struct {
			Connector struct {
				ID       string `json:"id"`
				Provider string `json:"provider"`
			} `json:"connector"`
		} `json:"createAPIKeyConnector"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId": orgID,
			"provider":       "BREX",
			"apiKey":         "key-to-delete",
		},
	}, &createResult)
	require.NoError(t, err)

	connectorID := createResult.CreateAPIKeyConnector.Connector.ID
	require.NotEmpty(t, connectorID)

	// Now delete the connector.
	const deleteQuery = `
		mutation($input: DeleteConnectorInput!) {
			deleteConnector(input: $input) {
				deletedConnectorId
			}
		}
	`

	var deleteResult struct {
		DeleteConnector struct {
			DeletedConnectorID string `json:"deletedConnectorId"`
		} `json:"deleteConnector"`
	}

	err = owner.Execute(deleteQuery, map[string]any{
		"input": map[string]any{
			"connectorId": connectorID,
		},
	}, &deleteResult)
	require.NoError(t, err)
	assert.Equal(t, connectorID, deleteResult.DeleteConnector.DeletedConnectorID)
}

func TestCreateAPIKeyConnector_RBAC(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	t.Run("viewer cannot create connector", func(t *testing.T) {
		t.Parallel()

		_, err := viewer.Do(`
			mutation($input: CreateAPIKeyConnectorInput!) {
				createAPIKeyConnector(input: $input) {
					connector { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId": viewer.GetOrganizationID().String(),
				"provider":       "BREX",
				"apiKey":         "test-key",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create connector")
	})
}
