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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
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

type foreignSession struct{}

var _ cloud.Session = foreignSession{}

func (foreignSession) Cloud() string     { return cloud.GCP }
func (foreignSession) AccountID() string { return "foreign-account" }

// TestAWSRegistration pins the shape Register validates: a workload identity
// provider declares no credential-bearing connect path, so nothing in the
// console can ask a customer for an AWS key that the driver could not use.
func TestAWSRegistration(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderAWS)
	require.True(t, ok)

	assert.Equal(t, "Amazon Web Services", reg.DisplayName)
	assert.True(t, reg.SupportsWorkloadIdentity())
	assert.False(t, reg.SupportsAPIKey())
	assert.False(t, reg.IsManagedAPIKey())
	assert.False(t, reg.SupportsClientCredentials())
	assert.Nil(t, reg.OAuth2)
	assert.Nil(t, reg.NewDriver)
	assert.NotEmpty(t, reg.EndpointOverrideUnsupported)
	assert.Equal(t, provider.Endpoints{}, reg.Endpoints)

	require.Len(t, reg.WorkloadIdentityExtraSettings(), 1)
	assert.Equal(t, "roleArn", reg.WorkloadIdentityExtraSettings()[0].Key)
	assert.True(t, reg.WorkloadIdentityExtraSettings()[0].Required)
	assert.Nil(t, reg.NewNameResolver)
	require.NotNil(t, reg.WorkloadIdentity.NewNameResolver)
}

func TestAWSNewNameResolver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderAWS)
	require.True(t, ok)
	require.NotNil(t, reg.WorkloadIdentity.NewNameResolver)

	conn := awsTestConnector(t, coredata.AWSConnectorSettings{
		RoleARN: "arn:aws:iam::123456789012:role/ProboAudit",
	})
	logger := log.NewLogger(log.WithOutput(io.Discard))

	t.Run(
		"refuses a session on another cloud",
		func(t *testing.T) {
			t.Parallel()

			assert.Nil(
				t,
				reg.WorkloadIdentity.NewNameResolver(
					context.Background(),
					foreignSession{},
					conn,
					logger,
				),
			)
		},
	)

	t.Run(
		"returns a resolver for an aws session",
		func(t *testing.T) {
			t.Parallel()

			session, err := reg.WorkloadIdentity.NewSession(context.Background(), awsTestIssuer(t), conn)
			require.NoError(t, err)

			resolver := reg.WorkloadIdentity.NewNameResolver(
				context.Background(),
				session,
				conn,
				logger,
			)
			require.NotNil(t, resolver)
		},
	)
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

func TestAWSNewDriver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderAWS)
	require.True(t, ok)

	t.Run("refuses a session on another cloud", func(t *testing.T) {
		t.Parallel()

		conn := awsTestConnector(t, coredata.AWSConnectorSettings{
			RoleARN: "arn:aws:iam::123456789012:role/ProboAudit",
		})

		_, err := reg.WorkloadIdentity.NewDriver(
			context.Background(),
			foreignSession{},
			conn,
			log.NewLogger(log.WithOutput(io.Discard)),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session is for GCP")
	})

	t.Run("returns a driver for an aws session", func(t *testing.T) {
		t.Parallel()

		conn := awsTestConnector(t, coredata.AWSConnectorSettings{
			RoleARN: "arn:aws:iam::123456789012:role/ProboAudit",
		})

		session, err := reg.WorkloadIdentity.NewSession(context.Background(), awsTestIssuer(t), conn)
		require.NoError(t, err)

		driver, err := reg.WorkloadIdentity.NewDriver(
			context.Background(),
			session,
			conn,
			log.NewLogger(log.WithOutput(io.Discard)),
		)
		require.NoError(t, err)
		require.NotNil(t, driver)
	})
}
