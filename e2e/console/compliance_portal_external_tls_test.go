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

package console_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

type (
	compliancePortalDomainsResult struct {
		PublicURL     string                      `json:"publicUrl"`
		DefaultDomain *compliancePortalDomainNode `json:"defaultDomain"`
		CustomDomain  *compliancePortalDomainNode `json:"customDomain"`
	}

	compliancePortalDomainNode struct {
		Domain      string `json:"domain"`
		Certificate *struct {
			Status string `json:"status"`
		} `json:"certificate"`
	}
)

// TestCompliancePortal_ExternalTLSLinkedDomains checks that a domain linked to
// a portal is served and verified without a certificate in external TLS mode.
func TestCompliancePortal_ExternalTLSLinkedDomains(t *testing.T) {
	t.Parallel()

	env := testutil.StartIsolatedEnv(
		t,
		testutil.IsolatedEnvOptions{CompliancePortalTLSMode: "external"},
	)

	owner := testutil.NewClientFor(t, env.BaseURL, testutil.RoleOwner)
	compliancePortalID := factory.CreateCompliancePortal(owner)
	testutil.ActivateCompliancePortal(t, owner, compliancePortalID)

	domains := queryCompliancePortalDomains(t, owner, compliancePortalID)
	require.NotNil(t, domains.DefaultDomain, "portal must get a managed default domain")

	managedHost := domains.DefaultDomain.Domain
	assert.Equal(t, "https://"+managedHost, domains.PublicURL)

	customDomain := strings.ToLower(strings.ReplaceAll(factory.SafeName("trust"), " ", "-")) + ".e2e.localhost"
	createCompliancePortalCustomDomain(t, owner, compliancePortalID, customDomain)

	domains = queryCompliancePortalDomains(t, owner, compliancePortalID)
	require.NotNil(t, domains.CustomDomain)
	assert.Equal(t, customDomain, domains.CustomDomain.Domain)
	assert.Equal(
		t,
		"https://"+customDomain,
		domains.PublicURL,
		"public URL must move to the custom domain without a certificate",
	)

	if domains.CustomDomain.Certificate != nil {
		assert.NotEqual(t, "ACTIVE", domains.CustomDomain.Certificate.Status)
	}

	t.Run(
		"serves the custom domain by forwarded host",
		func(t *testing.T) {
			t.Parallel()

			status, _, body := forwardedPortalGet(
				t,
				env.TrustHTTPAddr,
				customDomain,
				"/.well-known/oauth-client-metadata",
			)
			require.Equal(t, http.StatusOK, status, "body: %s", body)

			var doc struct {
				ClientID     string   `json:"client_id"`
				RedirectURIs []string `json:"redirect_uris"`
			}
			require.NoError(t, json.Unmarshal(body, &doc))

			assert.Equal(t, "https://"+customDomain+"/.well-known/oauth-client-metadata", doc.ClientID)
			assert.Equal(t, []string{"https://" + customDomain + "/callback"}, doc.RedirectURIs)
		},
	)

	t.Run(
		"redirects the managed hostname to the custom domain",
		func(t *testing.T) {
			t.Parallel()

			status, header, _ := forwardedPortalGet(t, env.TrustHTTPAddr, managedHost, "/overview")

			assert.Equal(t, http.StatusPermanentRedirect, status)
			assert.Equal(t, "https://"+customDomain+"/overview", header.Get("Location"))
		},
	)

	t.Run(
		"accepts the custom domain as a redirect host",
		func(t *testing.T) {
			t.Parallel()

			continueURL := url.QueryEscape("https://" + customDomain + "/overview")
			status, header, body := forwardedPortalGet(
				t,
				env.TrustHTTPAddr,
				customDomain,
				"/initiate?continue="+continueURL,
			)

			assert.Equal(t, http.StatusFound, status, "body: %s", body)
			assert.Contains(t, header.Get("Location"), "/api/connect/v1/oauth2/authorize")
		},
	)

	t.Run(
		"accepts the custom domain as an OAuth client id",
		func(t *testing.T) {
			t.Parallel()

			status, header, body := forwardedPortalGet(
				t,
				env.TrustHTTPAddr,
				customDomain,
				"/initiate?continue=/overview",
			)
			require.Equal(t, http.StatusFound, status, "body: %s", body)

			authorizeURL, err := url.Parse(header.Get("Location"))
			require.NoError(t, err)
			require.Contains(t, authorizeURL.Path, "/api/connect/v1/oauth2/authorize")
			assert.Equal(
				t,
				"https://"+customDomain+"/.well-known/oauth-client-metadata",
				authorizeURL.Query().Get("client_id"),
			)

			resp, err := testutil.OAuth2Authorize(owner, authorizeURL.Query())
			require.NoError(t, err)

			// The client id is accepted; nothing serves its metadata document here.
			var oauthErr struct {
				Code        string `json:"error"`
				Description string `json:"error_description"`
			}
			require.NoError(t, json.Unmarshal(resp.Body, &oauthErr), "authorize body: %s", resp.Body)

			assert.Equal(t, "invalid_client", oauthErr.Code)
			assert.Contains(
				t,
				[]string{"cannot fetch client metadata document", "client metadata document unavailable"},
				oauthErr.Description,
			)
		},
	)
}

func queryCompliancePortalDomains(
	t *testing.T,
	c *testutil.Client,
	compliancePortalID string,
) compliancePortalDomainsResult {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on CompliancePortal {
					publicUrl
					defaultDomain { domain certificate { status } }
					customDomain { domain certificate { status } }
				}
			}
		}
	`

	var result struct {
		Node compliancePortalDomainsResult `json:"node"`
	}

	err := c.Execute(query, map[string]any{"id": compliancePortalID}, &result)
	require.NoError(t, err)

	return result.Node
}

func createCompliancePortalCustomDomain(
	t *testing.T,
	c *testutil.Client,
	compliancePortalID string,
	domain string,
) {
	t.Helper()

	const mutation = `
		mutation($input: CreateCustomDomainInput!) {
			createCustomDomain(input: $input) {
				customDomain { id domain }
			}
		}
	`

	err := c.Execute(mutation, map[string]any{
		"input": map[string]any{
			"compliancePortalId": compliancePortalID,
			"domain":             domain,
		},
	}, nil)
	require.NoError(t, err)
}

func forwardedPortalGet(
	t *testing.T,
	listenAddr string,
	host string,
	path string,
) (int, http.Header, []byte) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "http://"+listenAddr+path, nil)
	require.NoError(t, err)

	req.Host = host
	req.Header.Set("X-Forwarded-Proto", "https")

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, resp.Header, body
}
