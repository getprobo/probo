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

package provider_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/cloud"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

const awsTestIssuerBase = "https://proboidentity.com"

func awsTestIssuer(t *testing.T) *identityfederation.Issuer {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyRing, err := jose.NewKeyRing(
		[]jose.SigningKey{{PrivateKey: key, KID: "test", Active: true}},
	)
	require.NoError(t, err)

	base, err := baseurl.Parse(awsTestIssuerBase)
	require.NoError(t, err)

	issuer, err := identityfederation.NewIssuer(base, keyRing, identityfederation.DefaultTokenTTL)
	require.NoError(t, err)

	return issuer
}

func awsTestConnector(t *testing.T, settings coredata.AWSConnectorSettings) *coredata.Connector {
	t.Helper()

	tenantID := gid.NewTenantID()
	conn := &coredata.Connector{
		ID:             gid.New(tenantID, coredata.ConnectorEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Provider:       coredata.ConnectorProviderAWS,
		Protocol:       coredata.ConnectorProtocolWorkloadIdentity,
	}
	require.NoError(t, conn.SetSettings(&settings))

	return conn
}

// TestAWSRegistration pins the shape Register validates: a workload identity
// provider declares no credential-bearing connect path, so nothing in the
// console can ask a customer for an AWS key that the driver could not use.
func TestAWSRegistration(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderAWS)
	require.True(t, ok)

	assert.True(t, reg.SupportsWorkloadIdentity())
	assert.False(t, reg.SupportsAPIKey())
	assert.False(t, reg.IsManagedAPIKey())
	assert.False(t, reg.SupportsClientCredentials())
	assert.Nil(t, reg.OAuth2)
	assert.Nil(t, reg.NewDriver)
	assert.NotEmpty(t, reg.EndpointOverrideUnsupported)
	assert.Equal(t, provider.Endpoints{}, reg.Endpoints)

	// Isolation is the per-organization issuer, so there is no grant to read
	// back beyond a successful assume.
	assert.Nil(t, reg.WorkloadIdentity.InspectGrant)
}

func TestAWSNewSession(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderAWS)
	require.True(t, ok)

	conn := awsTestConnector(t, coredata.AWSConnectorSettings{
		RoleARN: "arn:aws:iam::123456789012:role/AuditorRole",
	})

	session, err := reg.WorkloadIdentity.NewSession(context.Background(), awsTestIssuer(t), conn)
	require.NoError(t, err)

	assert.Equal(t, cloud.AWS, session.Cloud())
	assert.Equal(t, "123456789012", session.AccountID())

	awsSession, ok := session.(*cloudaws.Session)
	require.True(t, ok)
	assert.Equal(t, cloudaws.DefaultCommercialRegion, awsSession.Config().Region)
}
