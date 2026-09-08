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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	"go.probo.inc/probo/pkg/identityfederation"
)

const (
	azureFixtureTenantID       = "a1111111-1111-4111-8111-111111111111"
	azureFixtureClientID       = "b2222222-2222-4222-8222-222222222222"
	azureFixtureSubscriptionID = "c3333333-3333-4333-8333-333333333333"
	azureSecondTenantID        = "d4444444-4444-4444-8444-444444444444"
	azureSecondClientID        = "e5555555-5555-4555-8555-555555555555"
	azureSecondSubscriptionID  = "f6666666-6666-4666-8666-666666666666"
	azureGovTenantID           = "aa111111-1111-4111-8111-111111111111"
	azureGovClientID           = "bb222222-2222-4222-8222-222222222222"
	azureGovSubscriptionID     = "cc333333-3333-4333-8333-333333333333"
)

const azureConnectorSetupQuery = `
	query($organizationId: ID!) {
		azureConnectorSetup(organizationId: $organizationId) {
			issuer
			audience
			subject
			suggestedApplicationName
			terraformSnippet
		}
	}
`

const createAzureWorkloadIdentityConnectorMutation = `
	mutation($input: CreateWorkloadIdentityConnectorInput!) {
		createWorkloadIdentityConnector(input: $input) {
			connector {
				id
				provider
				protocol
			}
		}
	}
`

const createAzureWorkloadIdentityConnectorWithStatusMutation = `
	mutation($input: CreateWorkloadIdentityConnectorInput!) {
		createWorkloadIdentityConnector(input: $input) {
			connector {
				id
				connectionStatus
			}
		}
	}
`

type azureConnectorSetupResult struct {
	AzureConnectorSetup struct {
		Issuer                   string `json:"issuer"`
		Audience                 string `json:"audience"`
		Subject                  string `json:"subject"`
		SuggestedApplicationName string `json:"suggestedApplicationName"`
		TerraformSnippet         string `json:"terraformSnippet"`
	} `json:"azureConnectorSetup"`
}

func TestAzureConnectorSetup(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	var result azureConnectorSetupResult

	err := owner.Execute(azureConnectorSetupQuery, map[string]any{
		"organizationId": orgID,
	}, &result)
	require.NoError(t, err)

	setup := result.AzureConnectorSetup
	assert.Contains(t, setup.Issuer, orgID)
	assert.Equal(t, identityfederation.AudienceAzure, setup.Audience)
	assert.Equal(t, orgID, setup.Subject)
	assert.Equal(t, cloudazure.DefaultApplicationName, setup.SuggestedApplicationName)
	assert.Contains(t, setup.TerraformSnippet, setup.Issuer)
	assert.Contains(t, setup.TerraformSnippet, setup.Subject)
	assert.Contains(t, setup.TerraformSnippet, cloudazure.DefaultTerraformModuleSource)
	assert.Contains(t, setup.TerraformSnippet, cloudazure.DefaultApplicationName)
}

func TestCreateAzureWorkloadIdentityConnector(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	var result createWorkloadIdentityConnectorResult

	err := owner.Execute(createAzureWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId":      orgID,
			"provider":            "AZURE",
			"azureTenantId":       azureFixtureTenantID,
			"azureClientId":       azureFixtureClientID,
			"azureSubscriptionId": azureFixtureSubscriptionID,
		},
	}, &result)
	require.NoError(t, err)

	connector := result.CreateWorkloadIdentityConnector.Connector
	assert.NotEmpty(t, connector.ID)
	assert.Equal(t, "AZURE", connector.Provider)
	assert.Equal(t, "WORKLOAD_IDENTITY", connector.Protocol)

	t.Run("allows a second connector for the same provider", func(t *testing.T) {
		t.Parallel()

		var second createWorkloadIdentityConnectorResult

		err := owner.Execute(createAzureWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId":      orgID,
				"provider":            "AZURE",
				"azureTenantId":       azureSecondTenantID,
				"azureClientId":       azureSecondClientID,
				"azureSubscriptionId": azureSecondSubscriptionID,
			},
		}, &second)
		require.NoError(t, err)

		secondID := second.CreateWorkloadIdentityConnector.Connector.ID
		assert.NotEmpty(t, secondID)
		assert.NotEqual(t, connector.ID, secondID)
	})
}

func TestCreateAzureWorkloadIdentityConnector_InvalidTenantID(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bogus := "not-a-guid"

	err := owner.Execute(createAzureWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId":      owner.GetOrganizationID().String(),
			"provider":            "AZURE",
			"azureTenantId":       bogus,
			"azureClientId":       azureFixtureClientID,
			"azureSubscriptionId": azureFixtureSubscriptionID,
		},
	}, &createWorkloadIdentityConnectorResult{})
	testutil.RequireErrorCode(t, err, "INVALID")
	assert.NotContains(t, err.Error(), bogus)
}

func TestCreateAzureWorkloadIdentityConnector_InvalidClientID(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bogus := "not-a-guid"

	err := owner.Execute(createAzureWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId":      owner.GetOrganizationID().String(),
			"provider":            "AZURE",
			"azureTenantId":       azureFixtureTenantID,
			"azureClientId":       bogus,
			"azureSubscriptionId": azureFixtureSubscriptionID,
		},
	}, &createWorkloadIdentityConnectorResult{})
	testutil.RequireErrorCode(t, err, "INVALID")
	assert.NotContains(t, err.Error(), bogus)
}

func TestCreateAzureWorkloadIdentityConnector_InvalidSubscriptionID(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bogus := "not-a-guid"

	err := owner.Execute(createAzureWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId":      owner.GetOrganizationID().String(),
			"provider":            "AZURE",
			"azureTenantId":       azureFixtureTenantID,
			"azureClientId":       azureFixtureClientID,
			"azureSubscriptionId": bogus,
		},
	}, &createWorkloadIdentityConnectorResult{})
	testutil.RequireErrorCode(t, err, "INVALID")
	assert.NotContains(t, err.Error(), bogus)
}

func TestCreateAzureWorkloadIdentityConnector_UnknownEnvironment(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bogus := "AZURE_GERMAN"

	err := owner.Execute(createAzureWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId":      owner.GetOrganizationID().String(),
			"provider":            "AZURE",
			"azureTenantId":       azureFixtureTenantID,
			"azureClientId":       azureFixtureClientID,
			"azureSubscriptionId": azureFixtureSubscriptionID,
			"azureEnvironment":    bogus,
		},
	}, &createWorkloadIdentityConnectorResult{})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), azureFixtureTenantID)
	assert.NotContains(t, err.Error(), azureFixtureClientID)
	assert.NotContains(t, err.Error(), azureFixtureSubscriptionID)
}

func TestAzureConnectorConnectionStatus_ProbeFailure(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	resp, err := owner.Do(createAzureWorkloadIdentityConnectorWithStatusMutation, map[string]any{
		"input": map[string]any{
			"organizationId":      orgID,
			"provider":            "AZURE",
			"azureTenantId":       azureFixtureTenantID,
			"azureClientId":       azureFixtureClientID,
			"azureSubscriptionId": azureFixtureSubscriptionID,
		},
	})
	require.NoError(t, err)

	var created createWorkloadIdentityConnectorWithStatusResult
	require.NoError(t, json.Unmarshal(resp.Data, &created))

	connector := created.CreateWorkloadIdentityConnector.Connector
	require.NotEmpty(t, connector.ID)
	assert.Equal(t, "DISCONNECTED", connector.ConnectionStatus)

	payload := resp.DataString()
	assert.NotContains(t, payload, azureFixtureTenantID)
	assert.NotContains(t, payload, azureFixtureClientID)
	assert.NotContains(t, payload, azureFixtureSubscriptionID)
	assert.NotContains(t, strings.ToLower(payload), "aadsts")
	assert.NotContains(t, strings.ToLower(payload), "unauthorized")
	assert.NotContains(t, strings.ToLower(payload), "forbidden")
}

func TestAzureConnectorConnectionStatus_GovernmentProbeFailure(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	resp, err := owner.Do(createAzureWorkloadIdentityConnectorWithStatusMutation, map[string]any{
		"input": map[string]any{
			"organizationId":      orgID,
			"provider":            "AZURE",
			"azureTenantId":       azureGovTenantID,
			"azureClientId":       azureGovClientID,
			"azureSubscriptionId": azureGovSubscriptionID,
			"azureEnvironment":    "AZURE_GOVERNMENT",
		},
	})
	require.NoError(t, err)

	var created createWorkloadIdentityConnectorWithStatusResult
	require.NoError(t, json.Unmarshal(resp.Data, &created))

	connector := created.CreateWorkloadIdentityConnector.Connector
	require.NotEmpty(t, connector.ID)
	assert.Equal(t, "DISCONNECTED", connector.ConnectionStatus)

	payload := resp.DataString()
	assert.NotContains(t, payload, azureGovTenantID)
	assert.NotContains(t, payload, azureGovClientID)
	assert.NotContains(t, payload, azureGovSubscriptionID)
	assert.NotContains(t, strings.ToLower(payload), "aadsts")
	assert.NotContains(t, strings.ToLower(payload), "unauthorized")
	assert.NotContains(t, strings.ToLower(payload), "forbidden")
}

func TestAzureConnector_RBAC(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	orgID := owner.GetOrganizationID().String()

	t.Run("viewer cannot read setup", func(t *testing.T) {
		t.Parallel()

		err := viewer.Execute(azureConnectorSetupQuery, map[string]any{
			"organizationId": orgID,
		}, &azureConnectorSetupResult{})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to read azure connector setup")
	})

	t.Run("viewer cannot create connector", func(t *testing.T) {
		t.Parallel()

		err := viewer.Execute(createAzureWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId":      orgID,
				"provider":            "AZURE",
				"azureTenantId":       azureFixtureTenantID,
				"azureClientId":       azureFixtureClientID,
				"azureSubscriptionId": azureFixtureSubscriptionID,
			},
		}, &createWorkloadIdentityConnectorResult{})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create azure connector")
	})

	t.Run("viewer cannot read connection status", func(t *testing.T) {
		t.Parallel()

		var created createWorkloadIdentityConnectorResult

		err := owner.Execute(createAzureWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId":      orgID,
				"provider":            "AZURE",
				"azureTenantId":       azureFixtureTenantID,
				"azureClientId":       azureFixtureClientID,
				"azureSubscriptionId": azureFixtureSubscriptionID,
			},
		}, &created)
		require.NoError(t, err)

		err = viewer.Execute(organizationConnectorStatusQuery, map[string]any{
			"id": orgID,
		}, &organizationConnectorStatusResult{})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to read azure connector status")
	})
}

func TestAzureConnector_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1 := testutil.NewClient(t, testutil.RoleOwner)
	org2 := testutil.NewClient(t, testutil.RoleOwner)
	org1ID := org1.GetOrganizationID().String()

	err := org1.Execute(createAzureWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId":      org1ID,
			"provider":            "AZURE",
			"azureTenantId":       azureFixtureTenantID,
			"azureClientId":       azureFixtureClientID,
			"azureSubscriptionId": azureFixtureSubscriptionID,
		},
	}, &createWorkloadIdentityConnectorResult{})
	require.NoError(t, err)

	t.Run("cannot read setup for another organization", func(t *testing.T) {
		t.Parallel()

		err := org2.Execute(azureConnectorSetupQuery, map[string]any{
			"organizationId": org1ID,
		}, &azureConnectorSetupResult{})
		testutil.RequireForbiddenError(t, err, "org B should not read org A's azure connector setup")
	})

	t.Run("cannot create connector in another organization", func(t *testing.T) {
		t.Parallel()

		err := org2.Execute(createAzureWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId":      org1ID,
				"provider":            "AZURE",
				"azureTenantId":       azureFixtureTenantID,
				"azureClientId":       azureFixtureClientID,
				"azureSubscriptionId": azureFixtureSubscriptionID,
			},
		}, &createWorkloadIdentityConnectorResult{})
		testutil.RequireForbiddenError(t, err, "org B should not create a connector in org A")
	})

	t.Run("cannot read connection status from another organization", func(t *testing.T) {
		t.Parallel()

		err := org2.Execute(organizationConnectorStatusQuery, map[string]any{
			"id": org1ID,
		}, &organizationConnectorStatusResult{})
		testutil.RequireForbiddenError(t, err, "org B should not read org A's connector status")
	})
}
