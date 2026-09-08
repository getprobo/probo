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

package probo_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/probo"
)

const (
	testAWSRoleARN          = "arn:aws:iam::123456789012:role/ProboAudit"
	testGCPProvider         = "projects/123456789012/locations/global/workloadIdentityPools/probo/providers/probo"
	testGCPServiceAccount   = "probo-audit@my-project.iam.gserviceaccount.com"
	testAzureTenantID       = "a1111111-1111-4111-8111-111111111111"
	testAzureClientID       = "b2222222-2222-4222-8222-222222222222"
	testAzureSubscriptionID = "c3333333-3333-4333-8333-333333333333"
)

func TestMarshalWorkloadIdentitySettings(t *testing.T) {
	t.Parallel()

	t.Run("aws marshals role_arn", func(t *testing.T) {
		t.Parallel()

		raw, err := probo.MarshalWorkloadIdentitySettings(
			probo.WorkloadIdentitySettingsInput{
				Provider:   coredata.ConnectorProviderAWS,
				AWSRoleARN: testAWSRoleARN,
			},
		)
		require.NoError(t, err)

		var got map[string]string
		require.NoError(t, json.Unmarshal(raw, &got))
		assert.Equal(
			t,
			map[string]string{"role_arn": testAWSRoleARN},
			got,
		)
	})

	t.Run("gcp marshals canonical fields", func(t *testing.T) {
		t.Parallel()

		raw, err := probo.MarshalWorkloadIdentitySettings(
			probo.WorkloadIdentitySettingsInput{
				Provider:                    coredata.ConnectorProviderGCP,
				GCPWorkloadIdentityProvider: "https://iam.googleapis.com/" + testGCPProvider,
				GCPServiceAccountEmail:      "  " + testGCPServiceAccount + "  ",
			},
		)
		require.NoError(t, err)

		var got map[string]string
		require.NoError(t, json.Unmarshal(raw, &got))
		assert.Equal(
			t,
			map[string]string{
				"workload_identity_provider": testGCPProvider,
				"service_account_email":      testGCPServiceAccount,
			},
			got,
		)
	})

	t.Run("refuses a missing aws role arn", func(t *testing.T) {
		t.Parallel()

		_, err := probo.MarshalWorkloadIdentitySettings(
			probo.WorkloadIdentitySettingsInput{
				Provider: coredata.ConnectorProviderAWS,
			},
		)
		require.Error(t, err)
		assert.Equal(t, "awsRoleArn is required", err.Error())
		assert.NotErrorIs(t, err, probo.ErrMarshalWorkloadIdentitySettings)
	})

	t.Run("refuses missing gcp fields", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			provider string
			email    string
		}{
			{name: "empty provider", email: testGCPServiceAccount},
			{name: "empty email", provider: testGCPProvider},
			{name: "both empty"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, err := probo.MarshalWorkloadIdentitySettings(
					probo.WorkloadIdentitySettingsInput{
						Provider:                    coredata.ConnectorProviderGCP,
						GCPWorkloadIdentityProvider: tt.provider,
						GCPServiceAccountEmail:      tt.email,
					},
				)
				require.Error(t, err)
				assert.Equal(
					t,
					"gcpWorkloadIdentityProvider and gcpServiceAccountEmail are required",
					err.Error(),
				)
				assert.NotErrorIs(t, err, probo.ErrMarshalWorkloadIdentitySettings)
			})
		}
	})

	t.Run("refuses an unsupported provider", func(t *testing.T) {
		t.Parallel()

		_, err := probo.MarshalWorkloadIdentitySettings(
			probo.WorkloadIdentitySettingsInput{
				Provider: coredata.ConnectorProviderGitHub,
			},
		)
		require.Error(t, err)
		assert.Equal(t, "provider does not support workload identity", err.Error())
		assert.NotErrorIs(t, err, probo.ErrMarshalWorkloadIdentitySettings)
	})

	t.Run("refuses an invalid aws role arn", func(t *testing.T) {
		t.Parallel()

		raw := "arn:aws:iam::123456789012:user/alice"

		_, err := probo.MarshalWorkloadIdentitySettings(
			probo.WorkloadIdentitySettingsInput{
				Provider:   coredata.ConnectorProviderAWS,
				AWSRoleARN: raw,
			},
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "awsRoleArn is not an IAM role ARN")
		assert.NotContains(t, err.Error(), raw)
		assert.NotErrorIs(t, err, probo.ErrMarshalWorkloadIdentitySettings)
	})

	t.Run("refuses an invalid gcp provider resource", func(t *testing.T) {
		t.Parallel()

		raw := "projects/not-a-number/locations/global/workloadIdentityPools/probo/providers/probo"

		_, err := probo.MarshalWorkloadIdentitySettings(
			probo.WorkloadIdentitySettingsInput{
				Provider:                    coredata.ConnectorProviderGCP,
				GCPWorkloadIdentityProvider: raw,
				GCPServiceAccountEmail:      testGCPServiceAccount,
			},
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "workloadIdentityProvider is not a workload identity provider resource")
		assert.NotContains(t, err.Error(), raw)
		assert.NotErrorIs(t, err, probo.ErrMarshalWorkloadIdentitySettings)
	})

	t.Run("azure marshals canonical fields", func(t *testing.T) {
		t.Parallel()

		raw, err := probo.MarshalWorkloadIdentitySettings(
			probo.WorkloadIdentitySettingsInput{
				Provider:            coredata.ConnectorProviderAzure,
				AzureTenantID:       "  " + testAzureTenantID + "  ",
				AzureClientID:       "  " + testAzureClientID + "  ",
				AzureSubscriptionID: "  " + testAzureSubscriptionID + "  ",
				AzureEnvironment:    cloudazure.EnvironmentGovernment,
			},
		)
		require.NoError(t, err)

		var got map[string]string
		require.NoError(t, json.Unmarshal(raw, &got))
		assert.Equal(
			t,
			map[string]string{
				"tenant_id":       testAzureTenantID,
				"client_id":       testAzureClientID,
				"subscription_id": testAzureSubscriptionID,
				"environment":     string(cloudazure.EnvironmentGovernment),
			},
			got,
		)
	})

	t.Run("azure defaults an empty environment to AZURE_PUBLIC", func(t *testing.T) {
		t.Parallel()

		raw, err := probo.MarshalWorkloadIdentitySettings(
			probo.WorkloadIdentitySettingsInput{
				Provider:            coredata.ConnectorProviderAzure,
				AzureTenantID:       testAzureTenantID,
				AzureClientID:       testAzureClientID,
				AzureSubscriptionID: testAzureSubscriptionID,
			},
		)
		require.NoError(t, err)

		var got map[string]string
		require.NoError(t, json.Unmarshal(raw, &got))
		assert.Equal(t, string(cloudazure.EnvironmentPublic), got["environment"])
	})

	t.Run("refuses missing azure fields", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			tenantID       string
			clientID       string
			subscriptionID string
		}{
			{name: "empty tenant", clientID: testAzureClientID, subscriptionID: testAzureSubscriptionID},
			{name: "empty client", tenantID: testAzureTenantID, subscriptionID: testAzureSubscriptionID},
			{name: "empty subscription", tenantID: testAzureTenantID, clientID: testAzureClientID},
			{name: "all empty"},
		}

		for _, tt := range tests {
			t.Run(
				tt.name,
				func(t *testing.T) {
					t.Parallel()

					_, err := probo.MarshalWorkloadIdentitySettings(
						probo.WorkloadIdentitySettingsInput{
							Provider:            coredata.ConnectorProviderAzure,
							AzureTenantID:       tt.tenantID,
							AzureClientID:       tt.clientID,
							AzureSubscriptionID: tt.subscriptionID,
						},
					)
					require.Error(t, err)
					assert.Equal(
						t,
						"azureTenantId, azureClientId and azureSubscriptionId are required",
						err.Error(),
					)
					assert.NotErrorIs(t, err, probo.ErrMarshalWorkloadIdentitySettings)
				},
			)
		}
	})

	t.Run("refuses an unknown azure environment", func(t *testing.T) {
		t.Parallel()

		raw := "AZURE_GERMAN"

		_, err := probo.MarshalWorkloadIdentitySettings(
			probo.WorkloadIdentitySettingsInput{
				Provider:            coredata.ConnectorProviderAzure,
				AzureTenantID:       testAzureTenantID,
				AzureClientID:       testAzureClientID,
				AzureSubscriptionID: testAzureSubscriptionID,
				AzureEnvironment:    cloudazure.Environment(raw),
			},
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "environment is not a supported Azure environment")
		assert.NotContains(t, err.Error(), raw)
		assert.NotErrorIs(t, err, probo.ErrMarshalWorkloadIdentitySettings)
	})

	t.Run("refuses an invalid azure tenant id", func(t *testing.T) {
		t.Parallel()

		raw := "not-a-guid"

		_, err := probo.MarshalWorkloadIdentitySettings(
			probo.WorkloadIdentitySettingsInput{
				Provider:            coredata.ConnectorProviderAzure,
				AzureTenantID:       raw,
				AzureClientID:       testAzureClientID,
				AzureSubscriptionID: testAzureSubscriptionID,
			},
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tenantId is not a GUID")
		assert.NotContains(t, err.Error(), raw)
		assert.NotErrorIs(t, err, probo.ErrMarshalWorkloadIdentitySettings)
	})
}
