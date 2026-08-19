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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/identityfederation"
)

// These tests cover what Open decides, not what a credential exchange returns:
// only the session's identity is asserted.
type fakeCloudSession struct {
	accountID string
}

func (s *fakeCloudSession) Cloud() string     { return cloud.AWS }
func (s *fakeCloudSession) AccountID() string { return s.accountID }

// fakeDriver is a driver whose only job is to be identifiable.
type fakeDriver struct {
	session cloud.Session
	client  *http.Client
}

func (d *fakeDriver) ListAccounts(context.Context) ([]drivers.AccountRecord, error) {
	return nil, nil
}

// Slack stands in throughout only because it is a valid provider constant. The
// AWS provider itself arrives with its own registration.
const testProvider = coredata.ConnectorProviderSlack

func workloadIdentityConnector() *coredata.Connector {
	return &coredata.Connector{
		Provider:   testProvider,
		Protocol:   coredata.ConnectorProtocolWorkloadIdentity,
		Connection: &connector.WorkloadIdentityConnection{Cloud: cloud.AWS},
	}
}

func apiKeyConnector() *coredata.Connector {
	return &coredata.Connector{
		Provider:   testProvider,
		Protocol:   coredata.ConnectorProtocolAPIKey,
		Connection: &connector.APIKeyConnection{APIKey: "secret"},
	}
}

func oauth2Connector(conn *connector.OAuth2Connection) *coredata.Connector {
	return &coredata.Connector{
		Provider:   testProvider,
		Protocol:   coredata.ConnectorProtocolOAuth2,
		Connection: conn,
	}
}

func registryWith(t *testing.T, reg *provider.Registration) *provider.Registry {
	t.Helper()

	r := provider.NewRegistry()
	require.NoError(t, r.Register(reg))

	return r
}

// workloadIdentityRegistration wires the two fields a federating provider owns:
// the session factory Open fills its Handle through, and a driver factory
// adapted for that session.
func workloadIdentityRegistration(t *testing.T, session cloud.Session) *provider.Registration {
	t.Helper()

	return &provider.Registration{
		Provider:                 testProvider,
		DisplayName:              "Slack",
		SupportsWorkloadIdentity: true,
		NewCloudSession: func(
			_ context.Context,
			issuer *identityfederation.Issuer,
			_ *coredata.Connector,
		) (cloud.Session, error) {
			require.NotNil(t, issuer, "the issuer must reach the session factory")

			return session, nil
		},
		NewDriver: provider.Cloud(
			func(
				_ context.Context,
				session cloud.Session,
				_ *coredata.Connector,
				_ *log.Logger,
			) (drivers.Driver, error) {
				return &fakeDriver{session: session}, nil
			},
		),
	}
}

func TestRuntimeOpen_WorkloadIdentity(t *testing.T) {
	t.Parallel()

	t.Run("provider does not federate into a cloud", func(t *testing.T) {
		t.Parallel()

		r := registryWith(t, &provider.Registration{
			Provider:    testProvider,
			DisplayName: "Slack",
		})

		handle, err := provider.NewRuntime(r, nil, &identityfederation.Issuer{}).
			Open(context.Background(), workloadIdentityConnector())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not federate into a cloud")
		assert.Nil(t, handle)
	})

	// The customer's cloud has nothing to validate our tokens against, so say so
	// rather than surface the cloud's opaque AccessDenied later.
	t.Run("identity federation issuer disabled", func(t *testing.T) {
		t.Parallel()

		r := registryWith(t, workloadIdentityRegistration(t, &fakeCloudSession{}))

		handle, err := provider.NewRuntime(r, nil, nil).
			Open(context.Background(), workloadIdentityConnector())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "identity federation issuer is disabled")
		assert.Nil(t, handle)
	})

	// The session reaches the driver without the caller ever seeing it.
	t.Run("driver receives the session", func(t *testing.T) {
		t.Parallel()

		want := &fakeCloudSession{accountID: "123456789012"}
		r := registryWith(t, workloadIdentityRegistration(t, want))

		handle, err := provider.NewRuntime(r, nil, &identityfederation.Issuer{}).
			Open(context.Background(), workloadIdentityConnector())
		require.NoError(t, err)

		reg, ok := r.For(handle)
		require.True(t, ok)

		driver, err := reg.NewDriver(context.Background(), handle, log.NewLogger())
		require.NoError(t, err)

		fake, ok := driver.(*fakeDriver)
		require.Truef(t, ok, "expected *fakeDriver, got %T", driver)
		assert.Same(t, want, fake.session)
	})

	// Obtaining the session is already most of the check, so a provider needing
	// nothing further registers no Probe and the check passes.
	t.Run("probe without a registered closure", func(t *testing.T) {
		t.Parallel()

		r := registryWith(t, workloadIdentityRegistration(t, &fakeCloudSession{}))

		handle, err := provider.NewRuntime(r, nil, &identityfederation.Issuer{}).
			Open(context.Background(), workloadIdentityConnector())
		require.NoError(t, err)
		require.NoError(t, handle.Probe(context.Background()))
	})

	// Nothing is minted per use for this protocol, so there is nothing to write
	// back after an open.
	t.Run("credential is never dirty", func(t *testing.T) {
		t.Parallel()

		r := registryWith(t, workloadIdentityRegistration(t, &fakeCloudSession{}))

		handle, err := provider.NewRuntime(r, nil, &identityfederation.Issuer{}).
			Open(context.Background(), workloadIdentityConnector())
		require.NoError(t, err)
		assert.False(t, handle.CredentialRotated())
	})
}

func TestRuntimeOpen_HTTP(t *testing.T) {
	t.Parallel()

	httpRegistration := &provider.Registration{
		Provider:    testProvider,
		DisplayName: "Slack",
		NewDriver: provider.HTTP(
			func(
				_ context.Context,
				c *http.Client,
				_ *coredata.Connector,
				_ *log.Logger,
				_ provider.Endpoints,
			) (drivers.Driver, error) {
				return &fakeDriver{client: c}, nil
			},
		),
	}

	t.Run("driver receives an authenticated client", func(t *testing.T) {
		t.Parallel()

		r := registryWith(t, httpRegistration)

		handle, err := provider.NewRuntime(r, nil, nil).
			Open(context.Background(), apiKeyConnector())
		require.NoError(t, err)

		reg, ok := r.For(handle)
		require.True(t, ok)

		driver, err := reg.NewDriver(context.Background(), handle, log.NewLogger())
		require.NoError(t, err)

		fake, ok := driver.(*fakeDriver)
		require.Truef(t, ok, "expected *fakeDriver, got %T", driver)
		assert.NotNil(t, fake.client)
	})

	t.Run("unregistered provider", func(t *testing.T) {
		t.Parallel()

		handle, err := provider.NewRuntime(provider.NewRegistry(), nil, nil).
			Open(context.Background(), apiKeyConnector())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported provider")
		assert.Nil(t, handle)
	})

	t.Run("unsupported protocol", func(t *testing.T) {
		t.Parallel()

		r := registryWith(t, httpRegistration)

		dbConnector := apiKeyConnector()
		dbConnector.Protocol = coredata.ConnectorProtocol("SOMETHING_ELSE")

		handle, err := provider.NewRuntime(r, nil, nil).Open(context.Background(), dbConnector)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported protocol")
		assert.Nil(t, handle)
	})
}

// Providers that rotate refresh tokens (HubSpot, DocuSign) may hand back a new
// one while the access token still stands, so both are snapshotted at open.
func TestHandle_CredentialRotated(t *testing.T) {
	t.Parallel()

	// A runtime holding no OAuth app credentials keeps the stored tokens rather
	// than refreshing them, leaving each case to rotate what it means to.
	open := func(t *testing.T, dbConnector *coredata.Connector) *provider.Handle {
		t.Helper()

		r := registryWith(t, &provider.Registration{
			Provider:    testProvider,
			DisplayName: "Slack",
		})

		handle, err := provider.NewRuntime(r, nil, nil).
			Open(context.Background(), dbConnector)
		require.NoError(t, err)

		return handle
	}

	t.Run("no token rotated", func(t *testing.T) {
		t.Parallel()

		handle := open(t, oauth2Connector(&connector.OAuth2Connection{
			AccessToken:  "at",
			RefreshToken: "rt",
		}))

		assert.False(t, handle.CredentialRotated())
	})

	t.Run("access token rotated", func(t *testing.T) {
		t.Parallel()

		conn := &connector.OAuth2Connection{AccessToken: "at", RefreshToken: "rt"}

		handle := open(t, oauth2Connector(conn))
		conn.AccessToken = "at-2"

		assert.True(t, handle.CredentialRotated())
	})

	t.Run("refresh token rotated alone", func(t *testing.T) {
		t.Parallel()

		conn := &connector.OAuth2Connection{AccessToken: "at", RefreshToken: "rt"}

		handle := open(t, oauth2Connector(conn))
		conn.RefreshToken = "rt-2"

		assert.True(t, handle.CredentialRotated())
	})

	t.Run("api key connector carries no rotatable token", func(t *testing.T) {
		t.Parallel()

		handle := open(t, apiKeyConnector())

		assert.False(t, handle.CredentialRotated())
	})
}

// A provider whose factory was adapted for the wrong credential family is a
// wiring bug Register cannot see, since it cannot inspect a closure. It must
// surface as a named error rather than as a nil client reaching a driver.
func TestHandle_AdapterMismatch(t *testing.T) {
	t.Parallel()

	r := registryWith(t, &provider.Registration{
		Provider:                 testProvider,
		DisplayName:              "Slack",
		SupportsWorkloadIdentity: true,
		NewCloudSession: func(
			context.Context,
			*identityfederation.Issuer,
			*coredata.Connector,
		) (cloud.Session, error) {
			return &fakeCloudSession{}, nil
		},
		// Wrong adapter: this provider federates, so its Handle carries no
		// HTTP client.
		NewDriver: provider.HTTP(
			func(
				context.Context,
				*http.Client,
				*coredata.Connector,
				*log.Logger,
				provider.Endpoints,
			) (drivers.Driver, error) {
				return &fakeDriver{}, nil
			},
		),
	})

	handle, err := provider.NewRuntime(r, nil, &identityfederation.Issuer{}).
		Open(context.Background(), workloadIdentityConnector())
	require.NoError(t, err)

	reg, ok := r.For(handle)
	require.True(t, ok)

	driver, err := reg.NewDriver(context.Background(), handle, log.NewLogger())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not an HTTP one")
	assert.Nil(t, driver)
}

// A provider that registers no name resolver keeps the generic source name.
func TestHandle_NameResolverAbsent(t *testing.T) {
	t.Parallel()

	r := registryWith(t, &provider.Registration{
		Provider:    testProvider,
		DisplayName: "Slack",
		NewDriver: provider.HTTP(
			func(
				context.Context,
				*http.Client,
				*coredata.Connector,
				*log.Logger,
				provider.Endpoints,
			) (drivers.Driver, error) {
				return &fakeDriver{}, nil
			},
		),
	})

	handle, err := provider.NewRuntime(r, nil, nil).
		Open(context.Background(), apiKeyConnector())
	require.NoError(t, err)

	var resolver drivers.NameResolver
	if reg, ok := r.For(handle); ok && reg.NewNameResolver != nil {
		resolver = reg.NewNameResolver(context.Background(), handle, log.NewLogger())
	}

	assert.Nil(t, resolver)
}

// A provider whose scope is captured during the OAuth callback registers no
// lister, and the picker must read that as "no organizations to choose from"
// rather than as a failure.
func TestHandle_OrganizationsAbsent(t *testing.T) {
	t.Parallel()

	r := registryWith(t, &provider.Registration{
		Provider:    testProvider,
		DisplayName: "Slack",
		NewDriver: provider.HTTP(
			func(
				context.Context,
				*http.Client,
				*coredata.Connector,
				*log.Logger,
				provider.Endpoints,
			) (drivers.Driver, error) {
				return &fakeDriver{}, nil
			},
		),
	})

	handle, err := provider.NewRuntime(r, nil, nil).
		Open(context.Background(), apiKeyConnector())
	require.NoError(t, err)

	reg, ok := r.For(handle)
	require.True(t, ok)

	var orgs []drivers.Organization
	if reg.ListOrganizations != nil {
		orgs, err = reg.ListOrganizations(context.Background(), handle)
		require.NoError(t, err)
	}

	assert.Empty(t, orgs)
}
