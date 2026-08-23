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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/identityfederation"
)

// These tests cover what an open decides, not what a credential exchange
// returns: only the session's identity is asserted.
type fakeCloudSession struct {
	accountID string
}

func (s *fakeCloudSession) Cloud() string     { return cloud.AWS }
func (s *fakeCloudSession) AccountID() string { return s.accountID }

// Slack stands in throughout only because it is a valid provider constant. The
// AWS provider itself arrives with its own registration.
const testProvider = coredata.ConnectorProviderSlack

// openerOver is a registry-free Opener: every test here calls Open with a row
// it already holds, so the database and the encryption key never come into it.
func openerOver(
	registry *provider.Registry,
	issuer *identityfederation.Issuer,
) *provider.Opener {
	return provider.NewOpener(nil, cipher.EncryptionKey{}, registry, nil, issuer)
}

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

func registryWith(t *testing.T, reg *provider.Registration) *provider.Registry {
	t.Helper()

	r := provider.NewRegistry()
	require.NoError(t, r.Register(reg))

	return r
}

// workloadIdentityRegistration wires the one field a federating provider owns:
// the session factory an open fills its Handle through.
func workloadIdentityRegistration(t *testing.T, session cloud.Session) *provider.Registration {
	t.Helper()

	return &provider.Registration{
		Provider:    testProvider,
		DisplayName: "Slack",
		WorkloadIdentity: &provider.WorkloadIdentitySpec{
			NewSession: func(
				_ context.Context,
				issuer *identityfederation.Issuer,
				_ *coredata.Connector,
			) (cloud.Session, error) {
				require.NotNil(t, issuer, "the issuer must reach the session factory")

				return session, nil
			},
		},
	}
}

func TestOpenerOpen_WorkloadIdentity(t *testing.T) {
	t.Parallel()

	t.Run("provider does not federate into a cloud", func(t *testing.T) {
		t.Parallel()

		r := registryWith(t, &provider.Registration{
			Provider:    testProvider,
			DisplayName: "Slack",
		})

		handle, err := openerOver(r, &identityfederation.Issuer{}).
			Open(context.Background(), workloadIdentityConnector())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not federate into a cloud")
		assert.Nil(t, handle)
	})

	// The customer's cloud has nothing to validate our tokens against, so say
	// so rather than surface the cloud's opaque AccessDenied later.
	t.Run("identity federation issuer disabled", func(t *testing.T) {
		t.Parallel()

		r := registryWith(t, workloadIdentityRegistration(t, &fakeCloudSession{}))

		handle, err := openerOver(r, nil).
			Open(context.Background(), workloadIdentityConnector())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "identity federation issuer is disabled")
		assert.Nil(t, handle)
	})

	t.Run("handle carries the cloud session", func(t *testing.T) {
		t.Parallel()

		want := &fakeCloudSession{accountID: "123456789012"}
		r := registryWith(t, workloadIdentityRegistration(t, want))

		handle, err := openerOver(r, &identityfederation.Issuer{}).
			Open(context.Background(), workloadIdentityConnector())
		require.NoError(t, err)

		credential, ok := handle.Credential.(connector.CloudCredential)
		require.Truef(t, ok, "expected a cloud credential, got %T", handle.Credential)
		assert.Same(t, want, credential.Session)
	})

	// Obtaining the session is already most of the check, so a provider needing
	// nothing further registers no Probe and the check passes.
	t.Run("probe without a registered closure", func(t *testing.T) {
		t.Parallel()

		r := registryWith(t, workloadIdentityRegistration(t, &fakeCloudSession{}))

		handle, err := openerOver(r, &identityfederation.Issuer{}).
			Open(context.Background(), workloadIdentityConnector())
		require.NoError(t, err)
		require.NoError(t, handle.Probe(context.Background()))
	})
}

func TestOpenerOpen_HTTP(t *testing.T) {
	t.Parallel()

	httpRegistration := &provider.Registration{
		Provider:    testProvider,
		DisplayName: "Slack",
	}

	t.Run("handle carries an authenticated client", func(t *testing.T) {
		t.Parallel()

		r := registryWith(t, httpRegistration)

		handle, err := openerOver(r, nil).
			Open(context.Background(), apiKeyConnector())
		require.NoError(t, err)

		credential, ok := handle.Credential.(connector.HTTPCredential)
		require.Truef(t, ok, "expected an HTTP credential, got %T", handle.Credential)
		assert.NotNil(t, credential.Client)
	})

	t.Run("unregistered provider", func(t *testing.T) {
		t.Parallel()

		handle, err := openerOver(provider.NewRegistry(), nil).
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

		handle, err := openerOver(r, nil).Open(context.Background(), dbConnector)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported protocol")
		assert.Nil(t, handle)
	})
}

// A capability written against the wrong credential family is a wiring bug no
// registration check can see, since it cannot inspect a closure. It must
// surface as an error naming both families rather than as a zero value reaching
// a consumer.
func TestOver_CredentialFamilyMismatch(t *testing.T) {
	t.Parallel()

	r := registryWith(t, workloadIdentityRegistration(t, &fakeCloudSession{}))

	handle, err := openerOver(r, &identityfederation.Issuer{}).
		Open(context.Background(), workloadIdentityConnector())
	require.NoError(t, err)

	// This provider federates, so its handle carries no HTTP client.
	overHTTP := provider.Over(func(
		context.Context,
		connector.HTTPCredential,
		*provider.Handle,
		*log.Logger,
	) (string, error) {
		return "built", nil
	})

	got, err := overHTTP(context.Background(), handle, log.NewLogger())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "its WORKLOAD_IDENTITY credential is a connector.CloudCredential")
	assert.Empty(t, got)
}

// The counterpart of the mismatch above: a capability written against the
// family the handle actually carries receives the live credential.
func TestOver_MatchingCredentialFamily(t *testing.T) {
	t.Parallel()

	r := registryWith(t, &provider.Registration{
		Provider:    testProvider,
		DisplayName: "Slack",
		Endpoints:   provider.Endpoints{APIBase: "https://slack.example"},
	})

	handle, err := openerOver(r, nil).Open(context.Background(), apiKeyConnector())
	require.NoError(t, err)

	overHTTP := provider.Over(func(
		_ context.Context,
		credential connector.HTTPCredential,
		h *provider.Handle,
		_ *log.Logger,
	) (string, error) {
		require.NotNil(t, credential.Client)

		return h.Endpoints.APIBase, nil
	})

	got, err := overHTTP(context.Background(), handle, log.NewLogger())
	require.NoError(t, err)
	assert.Equal(t, "https://slack.example", got)
}
