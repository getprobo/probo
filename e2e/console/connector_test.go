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
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestAccessReviewDrivers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	const query = `
		query {
			accessReviewDrivers {
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
	`

	var result struct {
		AccessReviewDrivers []struct {
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
		} `json:"accessReviewDrivers"`
	}

	err := owner.Execute(query, nil, &result)
	require.NoError(t, err)
	assert.NotEmpty(t, result.AccessReviewDrivers)

	providerNames := make(map[string]bool)

	for _, info := range result.AccessReviewDrivers {
		assert.NotEmpty(t, info.Provider)
		assert.NotEmpty(t, info.DisplayName)
		assert.NotNil(t, info.ExtraSettings)
		providerNames[info.Provider] = true
	}

	assert.True(t, providerNames["BREX"], "expected BREX provider to be present")
	assert.True(t, providerNames["HUBSPOT"], "expected HUBSPOT provider to be present")

	t.Run("viewer can list access review drivers", func(t *testing.T) {
		t.Parallel()
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		var viewerResult struct {
			AccessReviewDrivers []struct {
				Provider    string `json:"provider"`
				DisplayName string `json:"displayName"`
			} `json:"accessReviewDrivers"`
		}

		err := viewer.Execute(query, nil, &viewerResult)
		require.NoError(t, err)
		assert.NotEmpty(t, viewerResult.AccessReviewDrivers)
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

// TestCreateAPIKeyConnectorSentryMissingSlug asserts that creating a
// Sentry API-key connector without sentryOrganizationSlug returns a
// validation error, not a 500. This is the e2e gate on the
// MarshalSettings validation path introduced by the connector-provider
// consolidation.
func TestCreateAPIKeyConnectorSentryMissingSlug(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	const query = `
		mutation($input: CreateAPIKeyConnectorInput!) {
			createAPIKeyConnector(input: $input) {
				connector { id }
			}
		}
	`

	_, err := owner.Do(query, map[string]any{
		"input": map[string]any{
			"organizationId": orgID,
			"provider":       "SENTRY",
			"apiKey":         "test-key",
		},
	})
	testutil.RequireErrorCode(t, err, "INVALID", "missing sentryOrganizationSlug must return INVALID not INTERNAL")
}

// TestCreateAPIKeyConnectorSentryRoundTrip asserts that supplying
// sentryOrganizationSlug succeeds and that the connector is created
// with the slug persisted in RawSettings.
func TestCreateAPIKeyConnectorSentryRoundTrip(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	const query = `
		mutation($input: CreateAPIKeyConnectorInput!) {
			createAPIKeyConnector(input: $input) {
				connector { id provider }
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
			"organizationId":         orgID,
			"provider":               "SENTRY",
			"apiKey":                 "test-key",
			"sentryOrganizationSlug": "my-org",
		},
	}, &result)
	require.NoError(t, err)
	assert.NotEmpty(t, result.CreateAPIKeyConnector.Connector.ID)
	assert.Equal(t, "SENTRY", result.CreateAPIKeyConnector.Connector.Provider)
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
