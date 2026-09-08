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

package azure_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pkgcloud "go.probo.inc/probo/pkg/cloud"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	"go.probo.inc/probo/pkg/identityfederation"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestNewSession(t *testing.T) {
	t.Parallel()

	session, err := cloudazure.NewSession(
		testIssuer(t),
		testOrganizationID(),
		testTenantIDUpper,
		testClientIDUpper,
		testSubscriptionIDUpper,
		cloudazure.EnvironmentPublic,
	)
	require.NoError(t, err)

	assert.Equal(t, pkgcloud.Azure, session.Cloud())
	assert.Equal(t, testSubscriptionID, session.AccountID())
	assert.Equal(t, testTenantID, session.TenantID())
	assert.Equal(t, cloudazure.EnvironmentPublic, session.Environment())
	assert.Equal(t, "https://graph.microsoft.com", session.GraphBaseURL())
	assert.NotNil(t, session.GraphClient())
	assert.Equal(t, cloud.AzurePublic, session.ARMClientOptions().Cloud)
}

func TestAudienceAzure_IsCarriedOnAssertion(t *testing.T) {
	t.Parallel()

	organizationID := testOrganizationID()
	token, err := testIssuer(t).Token(
		context.Background(),
		organizationID,
		identityfederation.AudienceAzure,
	)
	require.NoError(t, err)

	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(t, err)

	var claims map[string]any
	require.NoError(t, json.Unmarshal(payload, &claims))
	assert.Equal(t, identityfederation.AudienceAzure, claims["aud"])
	assert.Equal(t, organizationID.String(), claims["sub"])
}

func TestNewSession_EmptyEnvironmentDefaultsToPublic(t *testing.T) {
	t.Parallel()

	session, err := cloudazure.NewSession(
		testIssuer(t),
		testOrganizationID(),
		testTenantID,
		testClientID,
		testSubscriptionID,
		"",
	)
	require.NoError(t, err)
	assert.Equal(t, cloudazure.EnvironmentPublic, session.Environment())
}

func TestNewSessionFromToken(t *testing.T) {
	t.Parallel()

	session := cloudazure.NewSessionFromToken(testSubscriptionID, "arm-access-token")

	assert.Equal(t, pkgcloud.Azure, session.Cloud())
	assert.Equal(t, testSubscriptionID, session.AccountID())
	assert.Equal(t, cloudazure.EnvironmentPublic, session.Environment())
	assert.Equal(t, "https://graph.microsoft.com", session.GraphBaseURL())
	assert.NotNil(t, session.GraphClient())

	err := session.CheckAccess(context.Background())
	require.NoError(t, err)
}

func TestNewSession_Validation(t *testing.T) {
	t.Parallel()

	issuer := testIssuer(t)
	organizationID := testOrganizationID()

	tests := []struct {
		name           string
		tenantID       string
		clientID       string
		subscriptionID string
		environment    cloudazure.Environment
		wantMessage    string
		forbidden      string
	}{
		{
			name:           "malformed tenant id",
			tenantID:       "not-a-guid",
			clientID:       testClientID,
			subscriptionID: testSubscriptionID,
			wantMessage:    "tenantId is not a GUID",
			forbidden:      "not-a-guid",
		},
		{
			name:           "malformed client id",
			tenantID:       testTenantID,
			clientID:       "also-not-a-guid",
			subscriptionID: testSubscriptionID,
			wantMessage:    "clientId is not a GUID",
			forbidden:      "also-not-a-guid",
		},
		{
			name:           "malformed subscription id",
			tenantID:       testTenantID,
			clientID:       testClientID,
			subscriptionID: "still-not-a-guid",
			wantMessage:    "subscriptionId is not a GUID",
			forbidden:      "still-not-a-guid",
		},
		{
			name:           "nil subscription id",
			tenantID:       testTenantID,
			clientID:       testClientID,
			subscriptionID: nilGUID,
			wantMessage:    "subscriptionId is not a GUID",
			forbidden:      nilGUID,
		},
		{
			name:           "unknown environment",
			tenantID:       testTenantID,
			clientID:       testClientID,
			subscriptionID: testSubscriptionID,
			environment:    "AZURE_GERMAN",
			wantMessage:    "environment is not a supported Azure environment",
			forbidden:      "AZURE_GERMAN",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				_, err := cloudazure.NewSession(
					issuer,
					organizationID,
					tt.tenantID,
					tt.clientID,
					tt.subscriptionID,
					tt.environment,
				)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantMessage)
				assert.NotContains(t, err.Error(), tt.forbidden)
			},
		)
	}
}

func TestNewSession_DoesNotReadAzureEnv(t *testing.T) {
	t.Setenv("AZURE_CLIENT_ID", "44444444-4444-4444-4444-444444444444")
	t.Setenv("AZURE_TENANT_ID", "55555555-5555-5555-5555-555555555555")
	t.Setenv("AZURE_CLIENT_SECRET", "definitely-not-a-real-secret")

	session, err := cloudazure.NewSession(
		testIssuer(t),
		testOrganizationID(),
		testTenantID,
		testClientID,
		testSubscriptionID,
		cloudazure.EnvironmentPublic,
	)
	require.NoError(t, err)
	assert.Equal(t, pkgcloud.Azure, session.Cloud())
	assert.Equal(t, testSubscriptionID, session.AccountID())
	assert.Equal(t, testTenantID, session.TenantID())
}

func TestNewSessionFromToken_WrongHost(t *testing.T) {
	t.Parallel()

	var gotAuth string

	session := cloudazure.NewSessionFromToken(
		testSubscriptionID,
		"arm-access-token",
		cloudazure.WithEnvironment(cloudazure.EnvironmentGovernment),
		cloudazure.WithHTTPClient(
			&http.Client{
				Transport: roundTripFunc(
					func(req *http.Request) (*http.Response, error) {
						host := req.URL.Host
						if host == "management.azure.com" || strings.HasSuffix(host, ".microsoft.com") {
							t.Fatalf("government session must not dial public-cloud host %q", host)
						}

						gotAuth = req.Header.Get("Authorization")

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader("{}")),
							Header:     make(http.Header),
							Request:    req,
						}, nil
					},
				),
			},
		),
	)

	assert.Equal(t, cloudazure.EnvironmentGovernment, session.Environment())
	assert.Equal(t, "https://graph.microsoft.us", session.GraphBaseURL())
	assert.Equal(t, cloud.AzureGovernment, session.ARMClientOptions().Cloud)

	endpoint := session.ARMClientOptions().Cloud.Services[cloud.ResourceManager].Endpoint
	assert.NotContains(t, endpoint, "management.azure.com")

	graphURL, err := url.JoinPath(session.GraphBaseURL(), "v1.0", "$metadata")
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, graphURL, nil)
	require.NoError(t, err)

	resp, err := session.GraphClient().Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Bearer arm-access-token", gotAuth)
}
