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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

// capturingRoundTripper records the URL of every request it sees and answers
// 401, so a call can be observed without reaching the real provider.
type capturingRoundTripper struct {
	urls []string
}

func (c *capturingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	c.urls = append(c.urls, r.URL.String())

	return &http.Response{
		StatusCode: http.StatusUnauthorized,
		Body:       http.NoBody,
		Header:     make(http.Header),
		Request:    r,
	}, nil
}

// DocuSign is the provider whose data host is discovered from its identity
// host, so an endpoint override that moves Identity has to move every
// capability with it. This asserts all three — accounts, instance name,
// organizations — actually reach the overridden host rather than the
// compiled-in production one, which is the bug this override machinery exists
// to prevent.
func TestDocuSignUsesOverriddenIdentity(t *testing.T) {
	t.Parallel()

	const overriddenIdentity = "https://account-d.docusign.com/oauth/userinfo"

	catalog, err := provider.NewBuiltinRegistryWith(provider.WithEndpointOverrides(provider.EndpointOverrides{
		coredata.ConnectorProviderDocuSign: {
			Auth:     "https://account-d.docusign.com/oauth/auth",
			Token:    "https://account-d.docusign.com/oauth/token",
			Probe:    overriddenIdentity,
			Identity: overriddenIdentity,
		},
	}))
	require.NoError(t, err)

	reg, ok := catalog.Get(coredata.ConnectorProviderDocuSign)
	require.True(t, ok)
	require.Equal(t, overriddenIdentity, reg.Endpoints.Identity)

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderDocuSign}
	require.NoError(t, conn.SetSettings(&coredata.DocuSignConnectorSettings{AccountID: "acct-1"}))

	logger := log.NewLogger(log.WithName("test"))
	sources := NewBuiltinRegistry()

	// open builds the driver the way a campaign fetch does, so the capability
	// assertions below exercise the same object production uses.
	open := func(t *testing.T) (Driver, *capturingRoundTripper) {
		t.Helper()

		rt := &capturingRoundTripper{}

		driver, err := sources.New(
			context.Background(),
			provider.NewHTTPHandleForTest(reg, conn, &http.Client{Transport: rt}),
			logger,
		)
		require.NoError(t, err)

		return driver, rt
	}

	t.Run("listing accounts", func(t *testing.T) {
		t.Parallel()

		driver, rt := open(t)

		// discoverBaseURI's userinfo call fails immediately (401 from the
		// capturing transport), so ListAccounts errors out — the request URL it
		// made on the way there is what this test is about.
		_, err := driver.ListAccounts(context.Background())
		require.Error(t, err)

		require.NotEmpty(t, rt.urls)
		assert.Equal(t, overriddenIdentity, rt.urls[0])
	})

	t.Run("resolving the instance name", func(t *testing.T) {
		t.Parallel()

		driver, rt := open(t)

		// A non-2xx userinfo response is the resolver's terminal case: it
		// returns ("", nil) rather than an error.
		name, err := InstanceName(context.Background(), driver)
		require.NoError(t, err)
		assert.Empty(t, name)

		require.NotEmpty(t, rt.urls)
		assert.Equal(t, overriddenIdentity, rt.urls[0])
	})

	t.Run("listing organizations", func(t *testing.T) {
		t.Parallel()

		driver, rt := open(t)

		_, err := Organizations(context.Background(), driver)
		require.Error(t, err)

		require.NotEmpty(t, rt.urls)
		assert.Equal(t, overriddenIdentity, rt.urls[0])
	})
}
