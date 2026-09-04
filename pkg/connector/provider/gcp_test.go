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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/cloud"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const (
	gcpTestProviderResource = "projects/123456789012/locations/global/workloadIdentityPools/probo/providers/probo"
	gcpTestServiceAccount   = "probo-audit@my-project.iam.gserviceaccount.com"
)

func gcpTestConnector(t *testing.T, settings coredata.GCPConnectorSettings) *coredata.Connector {
	t.Helper()

	tenantID := gid.NewTenantID()
	conn := &coredata.Connector{
		ID:             gid.New(tenantID, coredata.ConnectorEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Provider:       coredata.ConnectorProviderGCP,
		Protocol:       coredata.ConnectorProtocolWorkloadIdentity,
	}
	require.NoError(t, conn.SetSettings(&settings))

	return conn
}

type awsForeignSession struct{}

var _ cloud.Session = awsForeignSession{}

func (awsForeignSession) Cloud() string     { return cloud.AWS }
func (awsForeignSession) AccountID() string { return "123456789012" }

func TestGCPRegistration(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderGCP)
	require.True(t, ok)

	assert.Equal(t, "Google Cloud", reg.DisplayName)
	assert.Equal(t, "https://www.probo.com/docs/product/access-review/gcp", reg.DocumentationURL)
	assert.True(t, reg.SupportsWorkloadIdentity())
	assert.False(t, reg.SupportsAPIKey())
	assert.False(t, reg.IsManagedAPIKey())
	assert.False(t, reg.SupportsClientCredentials())
	assert.Nil(t, reg.OAuth2)
	assert.Nil(t, reg.NewDriver)
	assert.NotEmpty(t, reg.EndpointOverrideUnsupported)
	assert.Equal(t, provider.Endpoints{}, reg.Endpoints)

	require.Len(t, reg.WorkloadIdentityExtraSettings(), 2)
	assert.Equal(t, "workloadIdentityProvider", reg.WorkloadIdentityExtraSettings()[0].Key)
	assert.True(t, reg.WorkloadIdentityExtraSettings()[0].Required)
	assert.Equal(t, "serviceAccountEmail", reg.WorkloadIdentityExtraSettings()[1].Key)
	assert.True(t, reg.WorkloadIdentityExtraSettings()[1].Required)
	assert.Nil(t, reg.NewNameResolver)
	require.NotNil(t, reg.WorkloadIdentity.NewNameResolver)
}

func TestGCPNewSession(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderGCP)
	require.True(t, ok)

	conn := gcpTestConnector(
		t,
		coredata.GCPConnectorSettings{
			WorkloadIdentityProvider: gcpTestProviderResource,
			ServiceAccountEmail:      gcpTestServiceAccount,
		},
	)

	session, err := reg.WorkloadIdentity.NewSession(context.Background(), awsTestIssuer(t), conn)
	require.NoError(t, err)

	assert.Equal(t, cloud.GCP, session.Cloud())
	assert.Equal(t, "123456789012", session.AccountID())

	_, ok = session.(*cloudgcp.Session)
	assert.True(t, ok)
}

func TestGCPNewDriver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderGCP)
	require.True(t, ok)

	t.Run("refuses a session on another cloud", func(t *testing.T) {
		t.Parallel()

		conn := gcpTestConnector(
			t,
			coredata.GCPConnectorSettings{
				WorkloadIdentityProvider: gcpTestProviderResource,
				ServiceAccountEmail:      gcpTestServiceAccount,
			},
		)

		_, err := reg.WorkloadIdentity.NewDriver(
			context.Background(),
			awsForeignSession{},
			conn,
			log.NewLogger(log.WithOutput(io.Discard)),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session is for AWS")
	})

	t.Run("returns a driver for a gcp session", func(t *testing.T) {
		t.Parallel()

		conn := gcpTestConnector(
			t,
			coredata.GCPConnectorSettings{
				WorkloadIdentityProvider: gcpTestProviderResource,
				ServiceAccountEmail:      gcpTestServiceAccount,
			},
		)

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

func TestGCPNewNameResolver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderGCP)
	require.True(t, ok)
	require.NotNil(t, reg.WorkloadIdentity.NewNameResolver)

	conn := gcpTestConnector(
		t,
		coredata.GCPConnectorSettings{
			WorkloadIdentityProvider: gcpTestProviderResource,
			ServiceAccountEmail:      gcpTestServiceAccount,
		},
	)
	logger := log.NewLogger(log.WithOutput(io.Discard))

	t.Run(
		"refuses a session on another cloud",
		func(t *testing.T) {
			t.Parallel()

			assert.Nil(
				t,
				reg.WorkloadIdentity.NewNameResolver(
					context.Background(),
					awsForeignSession{},
					conn,
					logger,
				),
			)
		},
	)

	t.Run(
		"returns a resolver for a gcp session",
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

func TestGCPProbe(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderGCP)
	require.True(t, ok)
	require.NotNil(t, reg.WorkloadIdentity.Probe)

	t.Run("refuses a session on another cloud", func(t *testing.T) {
		t.Parallel()

		conn := gcpTestConnector(
			t,
			coredata.GCPConnectorSettings{
				WorkloadIdentityProvider: gcpTestProviderResource,
				ServiceAccountEmail:      gcpTestServiceAccount,
			},
		)

		err := reg.WorkloadIdentity.Probe(
			context.Background(),
			awsForeignSession{},
			conn,
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session is for AWS")
	})
}
