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

package drivers

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

// fakeCloudSession stands in for a real cloud session. What the credential
// exchange returns is pkg/cloud's business; what matters here is that the
// session reaches the driver.
type fakeCloudSession struct {
	accountID string
}

func (s *fakeCloudSession) Cloud() string     { return cloud.AWS }
func (s *fakeCloudSession) AccountID() string { return s.accountID }

// cloudAccountDriver is the shape a federating provider's driver takes: it
// holds a session rather than an HTTP client, and reads the accounts it reviews
// through that cloud's SDK. Here it just reports the account it reached.
type cloudAccountDriver struct {
	session cloud.Session
}

func (d *cloudAccountDriver) ListAccounts(context.Context) ([]AccountRecord, error) {
	return []AccountRecord{{ExternalID: d.session.AccountID()}}, nil
}

// A WORKLOAD_IDENTITY connector stores no credential and mints one per use, so
// it exercises a different half of the opener than every other provider. This
// walks the whole seam — stored row, minted credential, capability factory
// written against the cloud family — to pin that a federating provider needs no
// special case anywhere above the mint.
//
// It stands in for the AWS provider until one exists: the console cannot yet
// create a WORKLOAD_IDENTITY connector (there is no mutation and no connect
// dialog for it), so a real registration would be a catalog entry nobody could
// reach.
func TestCloudCredentialReachesItsDriver(t *testing.T) {
	t.Parallel()

	const federatingProvider = coredata.ConnectorProviderSlack

	session := &fakeCloudSession{accountID: "123456789012"}

	reg := &provider.Registration{
		Provider:    federatingProvider,
		DisplayName: "Federating provider",
		WorkloadIdentity: &provider.WorkloadIdentitySpec{
			NewSession: func(
				context.Context,
				*identityfederation.Issuer,
				*coredata.Connector,
			) (cloud.Session, error) {
				return session, nil
			},
		},
	}

	catalog := provider.NewRegistry()
	require.NoError(t, catalog.Register(reg))

	sources := NewRegistry()
	require.NoError(t, sources.Register(
		federatingProvider,
		provider.Over(func(
			_ context.Context,
			credential connector.CloudCredential,
			_ *provider.Handle,
			_ *log.Logger,
		) (Driver, error) {
			return &cloudAccountDriver{session: credential.Session}, nil
		}),
	))

	handle, err := provider.
		NewOpener(nil, cipher.EncryptionKey{}, catalog, nil, &identityfederation.Issuer{}).
		Open(context.Background(), &coredata.Connector{
			Provider:   federatingProvider,
			Protocol:   coredata.ConnectorProtocolWorkloadIdentity,
			Connection: &connector.WorkloadIdentityConnection{Cloud: cloud.AWS},
		})
	require.NoError(t, err)

	driver, err := sources.New(context.Background(), handle, log.NewLogger())
	require.NoError(t, err)

	accounts, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, accounts, 1)
	assert.Equal(t, "123456789012", accounts[0].ExternalID)

	// A federating provider advertises no picker and no instance name, and both
	// must read as absent rather than as a driver that answers and panics.
	orgs, err := Organizations(context.Background(), driver)
	require.NoError(t, err)
	assert.Empty(t, orgs)

	name, err := InstanceName(context.Background(), driver)
	require.NoError(t, err)
	assert.Empty(t, name)
}
