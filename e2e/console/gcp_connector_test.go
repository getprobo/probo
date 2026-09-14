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
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
)

const (
	gcpFixtureProviderResource = "projects/123456789012/locations/global/workloadIdentityPools/probo-pool/providers/probo"
	gcpSecondProviderResource  = "projects/111111111111/locations/global/workloadIdentityPools/probo-pool/providers/probo"
	gcpFixtureServiceAccount   = "probo-audit@example-project.iam.gserviceaccount.com"
)

const gcpConnectorSetupQuery = `
	query($organizationId: ID!) {
		gcpConnectorSetup(organizationId: $organizationId) {
			issuer
			audience
			subject
			suggestedServiceAccountName
			terraformSnippet
			terraformBulkSnippet
		}
	}
`

const createGCPWorkloadIdentityConnectorMutation = `
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

const createGCPWorkloadIdentityConnectorWithStatusMutation = `
	mutation($input: CreateWorkloadIdentityConnectorInput!) {
		createWorkloadIdentityConnector(input: $input) {
			connector {
				id
				connectionStatus
			}
		}
	}
`

type gcpConnectorSetupResult struct {
	GCPConnectorSetup struct {
		Issuer                      string `json:"issuer"`
		Audience                    string `json:"audience"`
		Subject                     string `json:"subject"`
		SuggestedServiceAccountName string `json:"suggestedServiceAccountName"`
		TerraformSnippet            string `json:"terraformSnippet"`
		TerraformBulkSnippet        string `json:"terraformBulkSnippet"`
	} `json:"gcpConnectorSetup"`
}

func TestGCPConnectorSetup(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	var result gcpConnectorSetupResult

	err := owner.Execute(gcpConnectorSetupQuery, map[string]any{
		"organizationId": orgID,
	}, &result)
	require.NoError(t, err)

	setup := result.GCPConnectorSetup
	assert.Contains(t, setup.Issuer, orgID)
	assert.Equal(t, cloudgcp.AudienceTemplate, setup.Audience)
	assert.Equal(t, orgID, setup.Subject)
	assert.Equal(t, cloudgcp.DefaultServiceAccountName, setup.SuggestedServiceAccountName)
	assert.Contains(t, setup.TerraformSnippet, setup.Issuer)
	assert.Contains(t, setup.TerraformSnippet, setup.Subject)
	assert.Contains(t, setup.TerraformSnippet, cloudgcp.DefaultTerraformModuleSource)
	assert.Contains(t, setup.TerraformSnippet, cloudgcp.DefaultServiceAccountName)
	assert.Contains(t, setup.TerraformBulkSnippet, setup.Issuer)
	assert.Contains(t, setup.TerraformBulkSnippet, setup.Subject)
	assert.Contains(t, setup.TerraformBulkSnippet, "for_each")
	assert.Contains(t, setup.TerraformBulkSnippet, "var.project_ids")
	assert.Contains(t, setup.TerraformBulkSnippet, "output \"connectors\"")
	assert.Contains(t, setup.TerraformBulkSnippet, cloudgcp.DefaultTerraformModuleSource)
}

func TestCreateGCPWorkloadIdentityConnector(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	var result createWorkloadIdentityConnectorResult

	err := owner.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId":              orgID,
			"provider":                    "GCP",
			"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
			"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
		},
	}, &result)
	require.NoError(t, err)

	connector := result.CreateWorkloadIdentityConnector.Connector
	assert.NotEmpty(t, connector.ID)
	assert.Equal(t, "GCP", connector.Provider)
	assert.Equal(t, "WORKLOAD_IDENTITY", connector.Protocol)

	t.Run("allows a second connector for the same provider", func(t *testing.T) {
		t.Parallel()

		var second createWorkloadIdentityConnectorResult

		err := owner.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId":              orgID,
				"provider":                    "GCP",
				"gcpWorkloadIdentityProvider": gcpSecondProviderResource,
				"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
			},
		}, &second)
		require.NoError(t, err)

		secondID := second.CreateWorkloadIdentityConnector.Connector.ID
		assert.NotEmpty(t, secondID)
		assert.NotEqual(t, connector.ID, secondID)
	})
}

func TestCreateGCPWorkloadIdentityConnector_InvalidProviderResource(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bogus := "projects/alice/locations/global/workloadIdentityPools/probo-pool/providers/probo"

	err := owner.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId":              owner.GetOrganizationID().String(),
			"provider":                    "GCP",
			"gcpWorkloadIdentityProvider": bogus,
			"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
		},
	}, &createWorkloadIdentityConnectorResult{})
	testutil.RequireErrorCode(t, err, "INVALID")
	assert.NotContains(t, err.Error(), bogus)
	assert.NotContains(t, err.Error(), "alice")
}

func TestCreateGCPWorkloadIdentityConnector_InvalidServiceAccountEmail(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bogus := "alice@example.com"

	err := owner.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId":              owner.GetOrganizationID().String(),
			"provider":                    "GCP",
			"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
			"gcpServiceAccountEmail":      bogus,
		},
	}, &createWorkloadIdentityConnectorResult{})
	testutil.RequireErrorCode(t, err, "INVALID")
	assert.NotContains(t, err.Error(), bogus)
	assert.NotContains(t, err.Error(), "alice")
}

func TestGCPConnectorConnectionStatus_ImpersonationFailure(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	resp, err := owner.Do(createGCPWorkloadIdentityConnectorWithStatusMutation, map[string]any{
		"input": map[string]any{
			"organizationId":              orgID,
			"provider":                    "GCP",
			"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
			"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
		},
	})
	require.NoError(t, err)

	var created createWorkloadIdentityConnectorWithStatusResult
	require.NoError(t, json.Unmarshal(resp.Data, &created))

	connector := created.CreateWorkloadIdentityConnector.Connector
	require.NotEmpty(t, connector.ID)
	assert.Equal(t, "DISCONNECTED", connector.ConnectionStatus)

	payload := resp.DataString()
	assert.NotContains(t, payload, gcpFixtureProviderResource)
	assert.NotContains(t, payload, gcpFixtureServiceAccount)
	assert.NotContains(t, strings.ToLower(payload), "permissiondenied")
	assert.NotContains(t, strings.ToLower(payload), "accessdenied")
}

func TestGCPConnector_RBAC(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	orgID := owner.GetOrganizationID().String()

	t.Run("viewer cannot read setup", func(t *testing.T) {
		t.Parallel()

		err := viewer.Execute(gcpConnectorSetupQuery, map[string]any{
			"organizationId": orgID,
		}, &gcpConnectorSetupResult{})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to read gcp connector setup")
	})

	t.Run("viewer cannot create connector", func(t *testing.T) {
		t.Parallel()

		err := viewer.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId":              orgID,
				"provider":                    "GCP",
				"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
				"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
			},
		}, &createWorkloadIdentityConnectorResult{})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create gcp connector")
	})

	t.Run("viewer cannot read connection status", func(t *testing.T) {
		t.Parallel()

		var created createWorkloadIdentityConnectorResult

		err := owner.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId":              orgID,
				"provider":                    "GCP",
				"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
				"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
			},
		}, &created)
		require.NoError(t, err)

		err = viewer.Execute(organizationConnectorStatusQuery, map[string]any{
			"id": orgID,
		}, &organizationConnectorStatusResult{})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to read gcp connector status")
	})
}

func TestGCPConnector_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1 := testutil.NewClient(t, testutil.RoleOwner)
	org2 := testutil.NewClient(t, testutil.RoleOwner)
	org1ID := org1.GetOrganizationID().String()

	err := org1.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId":              org1ID,
			"provider":                    "GCP",
			"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
			"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
		},
	}, &createWorkloadIdentityConnectorResult{})
	require.NoError(t, err)

	t.Run("cannot read setup for another organization", func(t *testing.T) {
		t.Parallel()

		err := org2.Execute(gcpConnectorSetupQuery, map[string]any{
			"organizationId": org1ID,
		}, &gcpConnectorSetupResult{})
		testutil.RequireForbiddenError(t, err, "org B should not read org A's gcp connector setup")
	})

	t.Run("cannot create connector in another organization", func(t *testing.T) {
		t.Parallel()

		err := org2.Execute(createGCPWorkloadIdentityConnectorMutation, map[string]any{
			"input": map[string]any{
				"organizationId":              org1ID,
				"provider":                    "GCP",
				"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
				"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
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

const createGcpAccessReviewSourcesMutation = `
	mutation($input: CreateGcpAccessReviewSourcesInput!) {
		createGcpAccessReviewSources(input: $input) {
			accessReviewSourceEdges {
				node {
					id
					name
				}
			}
			failures {
				index
				projectId
				reason
			}
		}
	}
`

type createGcpAccessReviewSourcesResult struct {
	CreateGcpAccessReviewSources struct {
		AccessReviewSourceEdges []struct {
			Node struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		} `json:"accessReviewSourceEdges"`
		Failures []struct {
			Index     int     `json:"index"`
			ProjectID *string `json:"projectId"`
			Reason    string  `json:"reason"`
		} `json:"failures"`
	} `json:"createGcpAccessReviewSources"`
}

func TestCreateGcpAccessReviewSources(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()
	bogusProvider := "projects/alice/locations/global/workloadIdentityPools/probo-pool/providers/probo"

	resp, err := owner.Do(createGcpAccessReviewSourcesMutation, map[string]any{
		"input": map[string]any{
			"organizationId": orgID,
			"projects": []map[string]any{
				{
					"projectId":                   "invalid-project",
					"gcpWorkloadIdentityProvider": bogusProvider,
					"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
				},
				{
					"projectId":                   "example-project",
					"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
					"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
				},
			},
		},
	})
	require.NoError(t, err)

	var result createGcpAccessReviewSourcesResult
	require.NoError(t, json.Unmarshal(resp.Data, &result))

	payload := result.CreateGcpAccessReviewSources
	assert.Empty(t, payload.AccessReviewSourceEdges)
	require.Len(t, payload.Failures, 2)
	assert.Equal(t, 0, payload.Failures[0].Index)
	assert.Equal(t, "INVALID", payload.Failures[0].Reason)
	assert.Equal(t, 1, payload.Failures[1].Index)
	assert.Equal(t, "DISCONNECTED", payload.Failures[1].Reason)

	if payload.Failures[0].ProjectID != nil {
		assert.Equal(t, "invalid-project", *payload.Failures[0].ProjectID)
	}

	if payload.Failures[1].ProjectID != nil {
		assert.Equal(t, "example-project", *payload.Failures[1].ProjectID)
	}

	body := resp.DataString()
	assert.NotContains(t, body, bogusProvider)
	assert.NotContains(t, body, "alice")
	assert.NotContains(t, body, gcpFixtureProviderResource)
	assert.NotContains(t, body, gcpFixtureServiceAccount)
	assert.NotContains(t, strings.ToLower(body), "permissiondenied")
	assert.NotContains(t, strings.ToLower(body), "accessdenied")
}

func TestCreateGcpAccessReviewSources_Empty(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	err := owner.Execute(createGcpAccessReviewSourcesMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"projects":       []map[string]any{},
		},
	}, &createGcpAccessReviewSourcesResult{})
	testutil.RequireErrorCode(t, err, "INVALID")
}

func TestCreateGcpAccessReviewSources_RBAC(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	err := viewer.Execute(createGcpAccessReviewSourcesMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"projects": []map[string]any{
				{
					"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
					"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
				},
			},
		},
	}, &createGcpAccessReviewSourcesResult{})
	testutil.RequireForbiddenError(t, err, "viewer should not be able to bulk create gcp access sources")
}

func TestCreateGcpAccessReviewSources_TenantIsolation(t *testing.T) {
	t.Parallel()
	org1 := testutil.NewClient(t, testutil.RoleOwner)
	org2 := testutil.NewClient(t, testutil.RoleOwner)

	err := org2.Execute(createGcpAccessReviewSourcesMutation, map[string]any{
		"input": map[string]any{
			"organizationId": org1.GetOrganizationID().String(),
			"projects": []map[string]any{
				{
					"gcpWorkloadIdentityProvider": gcpFixtureProviderResource,
					"gcpServiceAccountEmail":      gcpFixtureServiceAccount,
				},
			},
		},
	}, &createGcpAccessReviewSourcesResult{})
	testutil.RequireForbiddenError(t, err, "org B should not bulk create gcp access sources in org A")
}
