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
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const (
	azureTestTenantID       = "a1111111-1111-4111-8111-111111111111"
	azureTestClientID       = "b2222222-2222-4222-8222-222222222222"
	azureTestSubscriptionID = "c3333333-3333-4333-8333-333333333333"
)

func azureTestConnector(t *testing.T, settings coredata.AzureConnectorSettings) *coredata.Connector {
	t.Helper()

	tenantID := gid.NewTenantID()
	conn := &coredata.Connector{
		ID:             gid.New(tenantID, coredata.ConnectorEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Provider:       coredata.ConnectorProviderAzure,
		Protocol:       coredata.ConnectorProtocolWorkloadIdentity,
	}
	require.NoError(t, conn.SetSettings(&settings))

	return conn
}

func TestAzureRegistration(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderAzure)
	require.True(t, ok)

	assert.Equal(t, "Microsoft Azure", reg.DisplayName)
	assert.Equal(t, "https://www.probo.com/docs/product/access-review/azure", reg.DocumentationURL)
	assert.True(t, reg.SupportsWorkloadIdentity())
	assert.False(t, reg.SupportsAPIKey())
	assert.False(t, reg.IsManagedAPIKey())
	assert.False(t, reg.SupportsClientCredentials())
	assert.Nil(t, reg.OAuth2)
	assert.Nil(t, reg.NewDriver)
	assert.NotEmpty(t, reg.EndpointOverrideUnsupported)
	assert.Equal(t, provider.Endpoints{}, reg.Endpoints)

	require.Len(t, reg.WorkloadIdentityExtraSettings(), 4)
	assert.Equal(t, "tenantId", reg.WorkloadIdentityExtraSettings()[0].Key)
	assert.True(t, reg.WorkloadIdentityExtraSettings()[0].Required)
	assert.Equal(t, "clientId", reg.WorkloadIdentityExtraSettings()[1].Key)
	assert.True(t, reg.WorkloadIdentityExtraSettings()[1].Required)
	assert.Equal(t, "subscriptionId", reg.WorkloadIdentityExtraSettings()[2].Key)
	assert.True(t, reg.WorkloadIdentityExtraSettings()[2].Required)
	assert.Equal(t, "environment", reg.WorkloadIdentityExtraSettings()[3].Key)
	assert.True(t, reg.WorkloadIdentityExtraSettings()[3].Required)
	assert.Nil(t, reg.NewNameResolver)
	require.NotNil(t, reg.WorkloadIdentity.NewNameResolver)
}

func TestAzureNewSession(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderAzure)
	require.True(t, ok)

	conn := azureTestConnector(
		t,
		coredata.AzureConnectorSettings{
			TenantID:       azureTestTenantID,
			ClientID:       azureTestClientID,
			SubscriptionID: azureTestSubscriptionID,
			Environment:    string(cloudazure.EnvironmentPublic),
		},
	)

	session, err := reg.WorkloadIdentity.NewSession(context.Background(), awsTestIssuer(t), conn)
	require.NoError(t, err)

	assert.Equal(t, cloud.Azure, session.Cloud())
	assert.Equal(t, azureTestSubscriptionID, session.AccountID())

	_, ok = session.(*cloudazure.Session)
	assert.True(t, ok)
}

func TestAzureNewDriver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderAzure)
	require.True(t, ok)

	t.Run(
		"refuses a session on another cloud",
		func(t *testing.T) {
			t.Parallel()

			conn := azureTestConnector(
				t,
				coredata.AzureConnectorSettings{
					TenantID:       azureTestTenantID,
					ClientID:       azureTestClientID,
					SubscriptionID: azureTestSubscriptionID,
					Environment:    string(cloudazure.EnvironmentPublic),
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
		},
	)

	t.Run(
		"returns a driver for an azure session",
		func(t *testing.T) {
			t.Parallel()

			conn := azureTestConnector(
				t,
				coredata.AzureConnectorSettings{
					TenantID:       azureTestTenantID,
					ClientID:       azureTestClientID,
					SubscriptionID: azureTestSubscriptionID,
					Environment:    string(cloudazure.EnvironmentPublic),
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
		},
	)
}

func TestAzureNewNameResolver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderAzure)
	require.True(t, ok)
	require.NotNil(t, reg.WorkloadIdentity.NewNameResolver)

	conn := azureTestConnector(
		t,
		coredata.AzureConnectorSettings{
			TenantID:       azureTestTenantID,
			ClientID:       azureTestClientID,
			SubscriptionID: azureTestSubscriptionID,
			Environment:    string(cloudazure.EnvironmentPublic),
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
		"returns a resolver for an azure session",
		func(t *testing.T) {
			t.Parallel()

			session, err := reg.WorkloadIdentity.NewSession(context.Background(), awsTestIssuer(t), conn)
			require.NoError(t, err)

			assert.NotNil(
				t,
				reg.WorkloadIdentity.NewNameResolver(
					context.Background(),
					session,
					conn,
					logger,
				),
			)
		},
	)
}

func TestAzureProbe(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderAzure)
	require.True(t, ok)
	require.NotNil(t, reg.WorkloadIdentity.Probe)

	t.Run(
		"refuses a session on another cloud",
		func(t *testing.T) {
			t.Parallel()

			conn := azureTestConnector(
				t,
				coredata.AzureConnectorSettings{
					TenantID:       azureTestTenantID,
					ClientID:       azureTestClientID,
					SubscriptionID: azureTestSubscriptionID,
					Environment:    string(cloudazure.EnvironmentPublic),
				},
			)

			err := reg.WorkloadIdentity.Probe(
				context.Background(),
				awsForeignSession{},
				conn,
			)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "session is for AWS")
		},
	)
}

func TestAzureEnvironment_GraphQLRoundTrip(t *testing.T) {
	t.Parallel()

	for _, env := range cloudazure.Environments() {
		t.Run(
			env.String(),
			func(t *testing.T) {
				t.Parallel()

				raw, err := env.MarshalText()
				require.NoError(t, err)

				var got cloudazure.Environment
				require.NoError(t, got.UnmarshalText(raw))
				assert.Equal(t, env, got)
			},
		)
	}
}
