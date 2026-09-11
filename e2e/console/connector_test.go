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
	"slices"
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
				documentationUrl
				oauthConfigured
				configuredProtocols
				apiKeySupported
				clientCredentialsSupported
				apiKeyExtraSettings {
					key
					label
					required
				}
				apiKeyFormat {
					pattern
					example
				}
				clientCredentialsExtraSettings {
					key
					label
					required
				}
				workloadIdentitySupported
				workloadIdentityExtraSettings {
					key
					label
					required
				}
			}
		}
	`

	type settingInfo struct {
		Key      string `json:"key"`
		Label    string `json:"label"`
		Required bool   `json:"required"`
	}

	type keyFormat struct {
		Pattern string `json:"pattern"`
		Example string `json:"example"`
	}

	var result struct {
		AccessReviewDrivers []struct {
			Provider                       string        `json:"provider"`
			DisplayName                    string        `json:"displayName"`
			DocumentationURL               *string       `json:"documentationUrl"`
			OAuthConfigured                bool          `json:"oauthConfigured"`
			ConfiguredProtocols            []string      `json:"configuredProtocols"`
			APIKeySupported                bool          `json:"apiKeySupported"`
			ClientCredentialsSupported     bool          `json:"clientCredentialsSupported"`
			APIKeyExtraSettings            []settingInfo `json:"apiKeyExtraSettings"`
			APIKeyFormat                   *keyFormat    `json:"apiKeyFormat"`
			ClientCredentialsExtraSettings []settingInfo `json:"clientCredentialsExtraSettings"`
			WorkloadIdentitySupported      bool          `json:"workloadIdentitySupported"`
			WorkloadIdentityExtraSettings  []settingInfo `json:"workloadIdentityExtraSettings"`
		} `json:"accessReviewDrivers"`
	}

	err := owner.Execute(query, nil, &result)
	require.NoError(t, err)
	assert.NotEmpty(t, result.AccessReviewDrivers)

	providerNames := make(map[string]bool)
	docURLByProvider := make(map[string]*string)
	keyFormatByProvider := make(map[string]*keyFormat)
	protocolsByProvider := make(map[string][]string)
	apiKeySettingKeys := make(map[string][]string)
	clientCredentialsSettingKeys := make(map[string][]string)
	workloadIdentitySettingKeys := make(map[string][]string)
	workloadIdentitySupported := make(map[string]bool)

	for _, info := range result.AccessReviewDrivers {
		assert.NotEmpty(t, info.Provider)
		assert.NotEmpty(t, info.DisplayName)
		assert.NotNil(t, info.APIKeyExtraSettings)
		assert.NotNil(t, info.ClientCredentialsExtraSettings)
		assert.NotNil(t, info.WorkloadIdentityExtraSettings)
		providerNames[info.Provider] = true
		docURLByProvider[info.Provider] = info.DocumentationURL
		keyFormatByProvider[info.Provider] = info.APIKeyFormat
		protocolsByProvider[info.Provider] = info.ConfiguredProtocols
		workloadIdentitySupported[info.Provider] = info.WorkloadIdentitySupported
		assert.Equal(t, slices.Contains(info.ConfiguredProtocols, "OAUTH2"), info.OAuthConfigured)

		for _, s := range info.APIKeyExtraSettings {
			apiKeySettingKeys[info.Provider] = append(apiKeySettingKeys[info.Provider], s.Key)
		}

		for _, s := range info.ClientCredentialsExtraSettings {
			clientCredentialsSettingKeys[info.Provider] = append(clientCredentialsSettingKeys[info.Provider], s.Key)
		}

		for _, s := range info.WorkloadIdentityExtraSettings {
			workloadIdentitySettingKeys[info.Provider] = append(workloadIdentitySettingKeys[info.Provider], s.Key)
		}
	}

	assert.True(t, providerNames["BREX"], "expected BREX provider to be present")
	assert.True(t, providerNames["HUBSPOT"], "expected HUBSPOT provider to be present")
	assert.True(t, providerNames["AWS"], "expected AWS provider to be present when identity federation is enabled")
	assert.True(t, providerNames["GCP"], "expected GCP provider to be present when identity federation is enabled")
	assert.Equal(t, []string{"OAUTH2"}, protocolsByProvider["GITHUB"])
	assert.True(t, workloadIdentitySupported["AWS"])
	assert.Equal(t, []string{"roleArn"}, workloadIdentitySettingKeys["AWS"])
	assert.True(t, workloadIdentitySupported["GCP"])
	assert.Equal(
		t,
		[]string{"workloadIdentityProvider", "serviceAccountEmail"},
		workloadIdentitySettingKeys["GCP"],
	)
	assert.False(t, workloadIdentitySupported["BREX"])
	assert.Empty(t, workloadIdentitySettingKeys["BREX"])

	// 1Password is the only provider offering both connect paths, and each path
	// needs different settings: the SCIM-bridge driver behind the API key, the
	// Users API driver behind client credentials. A client rendering one path's
	// settings on the other would collect fields the create resolver rejects.
	assert.Equal(t, []string{"scimBridgeUrl"}, apiKeySettingKeys["ONE_PASSWORD"])
	assert.Equal(t, []string{"accountId", "region"}, clientCredentialsSettingKeys["ONE_PASSWORD"])

	// A single-path provider declares its settings on that path only.
	assert.Equal(t, []string{"baseUrl"}, apiKeySettingKeys["LANGFUSE"])
	assert.Empty(t, clientCredentialsSettingKeys["LANGFUSE"])

	// A documented provider exposes its probo.com docs URL; an undocumented one
	// exposes null. See pkg/connector/provider/docs.go.
	require.Contains(t, docURLByProvider, "ANTHROPIC")

	if url := docURLByProvider["ANTHROPIC"]; assert.NotNil(t, url) {
		assert.Equal(t, "https://www.probo.com/docs/product/access-review/anthropic", *url)
	}

	// A slug that is not the lowercased enum is the case worth pinning: Cal.com
	// publishes at /calcom, so deriving the URL from CAL_COM would 404.
	if url := docURLByProvider["CAL_COM"]; assert.NotNil(t, url) {
		assert.Equal(t, "https://www.probo.com/docs/product/access-review/calcom", *url)
	}

	if url := docURLByProvider["AWS"]; assert.NotNil(t, url) {
		assert.Equal(t, "https://www.probo.com/docs/product/access-review/aws", *url)
	}

	// Sentry carries the null case: it is an API-key provider, so the catalog
	// never skips it, and it still has no page. A provider picked here purely
	// for being undocumented today gets documented eventually and breaks this
	// assertion, as BREX and AWS did.
	//
	// Contains first: a bare Nil on a missing key passes whether or not the
	// field is really null, which would let the null path rot unnoticed.
	require.Contains(t, docURLByProvider, "SENTRY")
	assert.Nil(t, docURLByProvider["SENTRY"], "SENTRY has no doc page, documentationUrl must be null")

	// apiKeyFormat is what the connect dialog checks the pasted key against and
	// shows as its placeholder, so it has to reach the client. Langfuse pins
	// that projection; the per-provider shapes belong to the registry's own
	// tests. Sentry carries the null case, on the same Contains-first reasoning
	// as the doc URL above.
	require.Contains(t, keyFormatByProvider, "LANGFUSE")

	if format := keyFormatByProvider["LANGFUSE"]; assert.NotNil(t, format) {
		assert.Equal(t, `^pk-lf-[^:]+:sk-lf-[^:]+$`, format.Pattern)
		assert.Equal(t, "pk-lf-…:sk-lf-…", format.Example)
	}

	require.Contains(t, keyFormatByProvider, "SENTRY")
	assert.Nil(t, keyFormatByProvider["SENTRY"], "SENTRY declares no key shape, apiKeyFormat must be null")

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

	// displayName and documentationUrl are selected here and connectionStatus
	// is not: the first two read the provider registration, while the third
	// would probe Brex live from a test.
	const query = `
		mutation($input: CreateAPIKeyConnectorInput!) {
			createAPIKeyConnector(input: $input) {
				connector {
					id
					provider
					displayName
					documentationUrl
				}
			}
		}
	`

	var result struct {
		CreateAPIKeyConnector struct {
			Connector struct {
				ID               string  `json:"id"`
				Provider         string  `json:"provider"`
				DisplayName      string  `json:"displayName"`
				DocumentationURL *string `json:"documentationUrl"`
			} `json:"connector"`
		} `json:"createAPIKeyConnector"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": orgID,
			"provider":       "BREX",
			"apiKey":         "bxt_test-key-123",
		},
	}, &result)
	require.NoError(t, err)

	connector := result.CreateAPIKeyConnector.Connector
	assert.NotEmpty(t, connector.ID)
	assert.Equal(t, "BREX", connector.Provider)
	// The name a user reads, which is the registration's and not the enum.
	assert.Equal(t, "Brex", connector.DisplayName)
	require.NotNil(t, connector.DocumentationURL)
	assert.Equal(
		t,
		"https://www.probo.com/docs/product/access-review/brex",
		*connector.DocumentationURL,
	)
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

// TestCreateAPIKeyConnectorMalformedKey asserts that a key whose shape its
// provider could never have minted is refused before anything is written.
// Langfuse pastes two keys as one colon-joined string and the transport
// base64s it verbatim, so half a credential is otherwise stored and then
// authenticates as nothing.
func TestCreateAPIKeyConnectorMalformedKey(t *testing.T) {
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
			"provider":       "LANGFUSE",
			// The public half alone: the colon and the secret key are missing.
			"apiKey":          "pk-lf-11111111-2222-3333-4444-555555555555",
			"langfuseBaseUrl": "https://cloud.langfuse.com",
		},
	})
	testutil.RequireErrorCode(t, err, "INVALID", "a half-pasted key must return INVALID, not create a connector")
}

// TestCreateAPIKeyConnectorLangfuseKeyPair asserts the counterpart: a
// well-formed pair is accepted. It says nothing about the key working —
// only Langfuse can judge that, and it is not called here.
func TestCreateAPIKeyConnectorLangfuseKeyPair(t *testing.T) {
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
			"organizationId":  orgID,
			"provider":        "LANGFUSE",
			"apiKey":          "pk-lf-11111111-2222-3333-4444-555555555555:sk-lf-66666666-7777-8888-9999-000000000000",
			"langfuseBaseUrl": "https://cloud.langfuse.com",
		},
	}, &result)
	require.NoError(t, err)
	assert.NotEmpty(t, result.CreateAPIKeyConnector.Connector.ID)
	assert.Equal(t, "LANGFUSE", result.CreateAPIKeyConnector.Connector.Provider)
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
			"apiKey":         "bxt_key-to-delete",
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

// TestCrispConnectsByAppInstall pins the connect path Crisp actually offers,
// through the live schema. The provider is a managed API key AND an app
// install, and those two must not both surface: the redirect is the only way in
// (its API-key dialog would have no fields and would create a connector with no
// website id), so installSupported is true exactly where apiKeyManaged and
// apiKeySupported are false.
//
// It is in the catalog at all only because the e2e probod configures the plugin
// token and plugin id; without both, Crisp ships deactivated and is absent.
func TestCrispConnectsByAppInstall(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	const query = `
		query {
			accessReviewDrivers {
				provider
				apiKeySupported
				apiKeyManaged
				installSupported
				apiKeyExtraSettings {
					key
				}
			}
		}
	`

	var result struct {
		AccessReviewDrivers []struct {
			Provider            string `json:"provider"`
			APIKeySupported     bool   `json:"apiKeySupported"`
			APIKeyManaged       bool   `json:"apiKeyManaged"`
			InstallSupported    bool   `json:"installSupported"`
			APIKeyExtraSettings []struct {
				Key string `json:"key"`
			} `json:"apiKeyExtraSettings"`
		} `json:"accessReviewDrivers"`
	}

	require.NoError(t, owner.Execute(query, nil, &result))

	crispFound := false

	for _, driver := range result.AccessReviewDrivers {
		if driver.Provider != "CRISP" {
			assert.Falsef(
				t,
				driver.InstallSupported,
				"provider %q reports an install path; crisp is the only one",
				driver.Provider,
			)

			continue
		}

		crispFound = true

		assert.True(t, driver.InstallSupported, "crisp connects by app install")
		assert.False(t, driver.APIKeyManaged, "an install provider must not also offer the API-key dialog")
		assert.False(t, driver.APIKeySupported, "the customer pastes no crisp key")
		assert.Empty(t, driver.APIKeyExtraSettings, "the ceremony supplies the website id, not a form")
	}

	assert.True(t, crispFound, "crisp is configured in the e2e probod and must be in the catalog")
}
