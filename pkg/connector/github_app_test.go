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

package connector

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
)

func TestGitHubAppConnector_InstallationFlow(t *testing.T) {
	t.Parallel()

	privateKey := newGitHubAppTestPrivateKey(t)

	var tokenRequests atomic.Int64

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && r.URL.Path == "/login/oauth/access_token":
				require.NoError(t, r.ParseForm())
				assert.Equal(t, "github-app-client-id", r.Form.Get("client_id"))
				assert.Equal(t, "github-app-client-secret", r.Form.Get("client_secret"))
				assert.Equal(t, "user-code", r.Form.Get("code"))

				_, _ = w.Write([]byte(`{"access_token":"user-token"}`))
			case r.Method == http.MethodGet && r.URL.Path == "/user/installations":
				assert.Equal(t, "Bearer user-token", r.Header.Get("Authorization"))
				assert.Equal(t, "100", r.URL.Query().Get("per_page"))

				_, _ = w.Write([]byte(`{"total_count":1,"installations":[{"id":42,"target_type":"Organization","account":{"login":"acme"}}]}`))
			case r.Method == http.MethodPost && r.URL.Path == "/app/installations/42/access_tokens":
				tokenRequests.Add(1)
				assertGitHubAppJWT(t, r.Header.Get("Authorization"))

				_ = json.NewEncoder(w).Encode(
					gitHubAppInstallationToken{
						Token:     "installation-token",
						ExpiresAt: time.Now().Add(time.Hour),
					},
				)
			case r.Method == http.MethodGet && r.URL.Path == "/orgs/acme":
				assert.Equal(t, "Bearer installation-token", r.Header.Get("Authorization"))

				_, _ = w.Write([]byte(`{"login":"acme"}`))
			default:
				http.NotFound(w, r)
			}
		}),
	)
	t.Cleanup(server.Close)

	c := &GitHubAppConnector{
		AppID:        "123456",
		ClientID:     "github-app-client-id",
		ClientSecret: "github-app-client-secret",
		Slug:         "probo-test",
		PrivateKey:   privateKey,
		InstallBase:  "https://github.com/apps",
		TokenURL:     server.URL + "/login/oauth/access_token",
		APIBase:      server.URL,
		HTTPClient:   server.Client(),
	}
	organizationID := gid.New(gid.NewTenantID(), 0)
	initReq := httptest.NewRequest(
		http.MethodGet,
		"/?continue=%2Forganizations%2Facme%2Faccess-reviews%2Fsources",
		nil,
	)

	installURL, err := c.Initiate(
		context.Background(),
		GitHubProvider,
		organizationID,
		InitiateOptions{},
		initReq,
	)
	require.NoError(t, err)

	parsedAuthURL, err := url.Parse(installURL)
	require.NoError(t, err)
	assert.Equal(t, "github.com", parsedAuthURL.Host)
	assert.Equal(t, "/login/oauth/authorize", parsedAuthURL.Path)
	assert.Equal(t, "github-app-client-id", parsedAuthURL.Query().Get("client_id"))
	require.NotEmpty(t, parsedAuthURL.Query().Get("state"))
	assert.True(t, IsGitHubAppState(parsedAuthURL.Query().Get("state")))

	callbackURL := "/?installation_id=42&setup_action=install&code=user-code&state=" +
		url.QueryEscape(parsedAuthURL.Query().Get("state"))
	callbackReq := httptest.NewRequest(http.MethodGet, callbackURL, nil)
	connection, state, err := c.CompleteWithState(context.Background(), callbackReq)
	require.NoError(t, err)

	assert.Equal(t, int64(42), connection.InstallationID)
	assert.Equal(t, "acme", state.Organization)
	assert.Equal(t, organizationID.String(), state.OrganizationID)
	assert.Equal(
		t,
		"/organizations/acme/access-reviews/sources",
		state.ContinueURL,
	)

	client := &http.Client{
		Transport: &gitHubAppTransport{
			connection: connection,
			underlying: server.Client().Transport,
		},
	}
	for range 2 {
		resp, err := client.Get(server.URL + "/orgs/acme")
		require.NoError(t, err)

		_ = resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	assert.Equal(t, int64(1), tokenRequests.Load())
}

func TestGitHubAppConnector_RejectsInvalidState(t *testing.T) {
	t.Parallel()

	c := &GitHubAppConnector{
		AppID:        "123456",
		ClientID:     "github-app-client-id",
		ClientSecret: "github-app-client-secret",
		Slug:         "probo-test",
		PrivateKey:   newGitHubAppTestPrivateKey(t),
		InstallBase:  "https://github.com/apps",
	}
	req := httptest.NewRequest(
		http.MethodGet,
		"/?installation_id=42&setup_action=install&state=tampered",
		nil,
	)

	_, _, err := c.CompleteWithState(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot validate github app state token")
}

func TestGitHubAppConnector_BindsExistingOrganizationInstallation(t *testing.T) {
	t.Parallel()

	c, state := newGitHubAppCompleteFixture(
		t,
		`{"total_count":2,"installations":[{"id":7,"target_type":"User","account":{"login":"alice"}},{"id":42,"target_type":"Organization","account":{"login":"acme"}}]}`,
	)

	callbackReq := httptest.NewRequest(
		http.MethodGet,
		"/?code=user-code&state="+url.QueryEscape(state),
		nil,
	)
	connection, completed, err := c.CompleteWithState(context.Background(), callbackReq)
	require.NoError(t, err)

	assert.Equal(t, int64(42), connection.InstallationID)
	assert.Equal(t, "acme", completed.Organization)
}

func TestGitHubAppConnector_RequiresInstallWhenNoOrganization(t *testing.T) {
	t.Parallel()

	c, state := newGitHubAppCompleteFixture(
		t,
		`{"total_count":1,"installations":[{"id":7,"target_type":"User","account":{"login":"alice"}}]}`,
	)

	callbackReq := httptest.NewRequest(
		http.MethodGet,
		"/?code=user-code&state="+url.QueryEscape(state),
		nil,
	)
	_, _, err := c.CompleteWithState(context.Background(), callbackReq)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrGitHubAppInstallationRequired)

	installURL, err := c.InstallationURL(state)
	require.NoError(t, err)

	parsedInstallURL, err := url.Parse(installURL)
	require.NoError(t, err)
	assert.Equal(t, "/apps/probo-test/installations/new", parsedInstallURL.Path)
	assert.Equal(t, state, parsedInstallURL.Query().Get("state"))
}

func TestGitHubAppConnector_RejectsAmbiguousOrganizationInstallations(t *testing.T) {
	t.Parallel()

	c, state := newGitHubAppCompleteFixture(
		t,
		`{"total_count":2,"installations":[{"id":42,"target_type":"Organization","account":{"login":"acme"}},{"id":43,"target_type":"Organization","account":{"login":"other"}}]}`,
	)

	callbackReq := httptest.NewRequest(
		http.MethodGet,
		"/?code=user-code&state="+url.QueryEscape(state),
		nil,
	)
	_, _, err := c.CompleteWithState(context.Background(), callbackReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple organization installations")
	assert.NotErrorIs(t, err, ErrGitHubAppInstallationRequired)
}

func TestGitHubAppConnector_RejectsUnauthorizedInstallation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && r.URL.Path == "/login/oauth/access_token":
				_, _ = w.Write([]byte(`{"access_token":"user-token"}`))
			case r.Method == http.MethodGet && r.URL.Path == "/user/installations":
				_, _ = w.Write([]byte(`{"total_count":0,"installations":[]}`))
			default:
				http.NotFound(w, r)
			}
		}),
	)
	t.Cleanup(server.Close)

	c := &GitHubAppConnector{
		AppID:        "123456",
		ClientID:     "github-app-client-id",
		ClientSecret: "github-app-client-secret",
		Slug:         "probo-test",
		PrivateKey:   newGitHubAppTestPrivateKey(t),
		InstallBase:  "https://github.com/apps",
		TokenURL:     server.URL + "/login/oauth/access_token",
		APIBase:      server.URL,
		HTTPClient:   server.Client(),
	}

	organizationID := gid.New(gid.NewTenantID(), 0)
	installURL, err := c.Initiate(
		context.Background(),
		GitHubProvider,
		organizationID,
		InitiateOptions{},
		nil,
	)
	require.NoError(t, err)

	parsedInstallURL, err := url.Parse(installURL)
	require.NoError(t, err)

	callbackURL := "/?installation_id=99&setup_action=install&code=user-code&state=" +
		url.QueryEscape(parsedInstallURL.Query().Get("state"))
	callbackReq := httptest.NewRequest(http.MethodGet, callbackURL, nil)

	_, _, err = c.CompleteWithState(context.Background(), callbackReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot authorize github app installation")
}

func TestGitHubAppConnection_RuntimeCredentialsAreNotPersisted(t *testing.T) {
	t.Parallel()

	conn := &GitHubAppConnection{
		AppID:          "app-id",
		PrivateKey:     "private-key",
		InstallationID: 42,
		APIBase:        "https://api.github.com",
	}

	data, err := json.Marshal(conn)
	require.NoError(t, err)

	assert.NotContains(t, string(data), "app-id")
	assert.NotContains(t, string(data), "private-key")

	var restored GitHubAppConnection
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, int64(42), restored.InstallationID)
	assert.Empty(t, restored.AppID)
	assert.Empty(t, restored.PrivateKey)
}

func newGitHubAppCompleteFixture(t *testing.T, installationsJSON string) (*GitHubAppConnector, string) {
	t.Helper()

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && r.URL.Path == "/login/oauth/access_token":
				_, _ = w.Write([]byte(`{"access_token":"user-token"}`))
			case r.Method == http.MethodGet && r.URL.Path == "/user/installations":
				_, _ = w.Write([]byte(installationsJSON))
			default:
				http.NotFound(w, r)
			}
		}),
	)
	t.Cleanup(server.Close)

	c := &GitHubAppConnector{
		AppID:        "123456",
		ClientID:     "github-app-client-id",
		ClientSecret: "github-app-client-secret",
		Slug:         "probo-test",
		PrivateKey:   newGitHubAppTestPrivateKey(t),
		InstallBase:  "https://github.com/apps",
		TokenURL:     server.URL + "/login/oauth/access_token",
		APIBase:      server.URL,
		HTTPClient:   server.Client(),
	}

	installURL, err := c.Initiate(
		context.Background(),
		GitHubProvider,
		gid.New(gid.NewTenantID(), 0),
		InitiateOptions{},
		nil,
	)
	require.NoError(t, err)

	parsedAuthURL, err := url.Parse(installURL)
	require.NoError(t, err)

	state := parsedAuthURL.Query().Get("state")
	require.NotEmpty(t, state)

	return c, state
}

func newGitHubAppTestPrivateKey(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return string(
		pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(key),
			},
		),
	)
}

func assertGitHubAppJWT(t *testing.T, authorization string) {
	t.Helper()

	token := strings.TrimPrefix(authorization, "Bearer ")
	assert.Len(t, strings.Split(token, "."), 3)
}
