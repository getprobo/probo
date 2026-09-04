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

package gcp_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/cloud"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

const testIssuerBase = "https://proboidentity.com"

func testIssuer(t *testing.T) *identityfederation.Issuer {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyRing, err := jose.NewKeyRing(
		[]jose.SigningKey{{PrivateKey: key, KID: "test", Active: true}},
	)
	require.NoError(t, err)

	base, err := baseurl.Parse(testIssuerBase)
	require.NoError(t, err)

	issuer, err := identityfederation.NewIssuer(base, keyRing, identityfederation.DefaultTokenTTL)
	require.NoError(t, err)

	return issuer
}

func testOrganizationID() gid.GID {
	return gid.New(gid.NewTenantID(), coredata.OrganizationEntityType)
}

func TestNewSession(t *testing.T) {
	t.Parallel()

	session, err := cloudgcp.NewSession(
		testIssuer(t),
		testOrganizationID(),
		testProviderResource,
		testServiceAccount,
	)
	require.NoError(t, err)

	assert.Equal(t, cloud.GCP, session.Cloud())
	assert.Equal(t, "123456789012", session.AccountID(), "account comes from the provider resource")
	assert.Equal(t, cloudgcp.CommercialUniverse, session.UniverseDomain())
	assert.NotNil(t, session.HTTPClient())
}

func TestNewSession_S3NSUniverse(t *testing.T) {
	t.Parallel()

	session, err := cloudgcp.NewSession(
		testIssuer(t),
		testOrganizationID(),
		testProviderResource,
		testS3NSServiceAccount,
	)
	require.NoError(t, err)

	assert.Equal(t, cloudgcp.S3NSUniverse, session.UniverseDomain())
	assert.Equal(t, "123456789012", session.AccountID())
}

func TestNewSessionFromToken(t *testing.T) {
	t.Parallel()

	session := cloudgcp.NewSessionFromToken("123456789012", "sa-access-token")

	assert.Equal(t, cloud.GCP, session.Cloud())
	assert.Equal(t, "123456789012", session.AccountID())
	assert.Equal(t, cloudgcp.CommercialUniverse, session.UniverseDomain())
	assert.NotNil(t, session.HTTPClient())

	err := session.CheckAccess(context.Background())
	require.NoError(t, err)
}

func TestNewSession_Validation(t *testing.T) {
	t.Parallel()

	issuer := testIssuer(t)
	organizationID := testOrganizationID()

	tests := []struct {
		name                string
		providerResource    string
		serviceAccountEmail string
		wantMessage         string
		forbidden           string
	}{
		{
			name:                "malformed provider resource",
			providerResource:    "not-a-provider",
			serviceAccountEmail: testServiceAccount,
			wantMessage:         "workloadIdentityProvider is not a workload identity provider resource",
			forbidden:           "not-a-provider",
		},
		{
			name:                "malformed service account email",
			providerResource:    testProviderResource,
			serviceAccountEmail: "alice@example.com",
			wantMessage:         "serviceAccountEmail is not a service account email",
			forbidden:           "alice@example.com",
		},
		{
			name:                "unsupported universe host",
			providerResource:    "https://iam.example.com/" + testProviderResource,
			serviceAccountEmail: testServiceAccount,
			wantMessage:         "not a supported GCP universe",
			forbidden:           "example.com",
		},
		{
			name:                "s3ns host with commercial email",
			providerResource:    "https://iam.s3nsapis.fr/" + testProviderResource,
			serviceAccountEmail: testServiceAccount,
			wantMessage:         "name different GCP universes",
			forbidden:           testServiceAccount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := cloudgcp.NewSession(issuer, organizationID, tt.providerResource, tt.serviceAccountEmail)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantMessage)
			assert.NotContains(t, err.Error(), tt.forbidden)
		})
	}
}

func TestNewSession_DoesNotReadADC(t *testing.T) {
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/definitely/not-a-credentials-file.json")

	session, err := cloudgcp.NewSession(
		testIssuer(t),
		testOrganizationID(),
		testProviderResource,
		testServiceAccount,
	)
	require.NoError(t, err)
	assert.Equal(t, cloud.GCP, session.Cloud())
}
