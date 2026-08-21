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

package connector_test

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/statelesstoken"
)

func mintGitHubAppStateToken(t *testing.T) string {
	t.Helper()

	githubApp := &connector.GitHubAppConnector{
		ClientSecret: "github-app-client-secret",
		Slug:         "probo-test",
		InstallBase:  "https://github.com/apps",
	}
	organizationID := gid.New(gid.NewTenantID(), 0)
	installURL, err := githubApp.Initiate(
		context.Background(),
		connector.GitHubProvider,
		organizationID,
		connector.InitiateOptions{},
		httptest.NewRequest("GET", "/", nil),
	)
	require.NoError(t, err)

	parsedInstallURL, err := url.Parse(installURL)
	require.NoError(t, err)

	state := parsedInstallURL.Query().Get("state")
	require.NotEmpty(t, state)

	return state
}

func TestCompleteOAuth2FromRequest_RejectsMissingState(t *testing.T) {
	t.Parallel()

	registry := connector.NewConnectorRegistry()
	req := httptest.NewRequest("GET", "/?code=abc", nil)

	_, err := registry.CompleteOAuth2FromRequest(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing state parameter")
}

func TestCompleteGitHubAppFromRequest_RejectsMissingState(t *testing.T) {
	t.Parallel()

	registry := connector.NewConnectorRegistry()
	req := httptest.NewRequest("GET", "/?installation_id=42", nil)

	_, err := registry.CompleteGitHubAppFromRequest(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing state parameter")
}

func TestCompleteFromState_RejectsMissingState(t *testing.T) {
	t.Parallel()

	registry := connector.NewConnectorRegistry()
	req := httptest.NewRequest("GET", "/?code=abc", nil)

	_, err := registry.CompleteFromState(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing state parameter")
}

func TestCompleteFromState_RoutesGitHubAppState(t *testing.T) {
	t.Parallel()

	state := mintGitHubAppStateToken(t)
	require.True(t, connector.IsGitHubAppState(state))

	registry := connector.NewConnectorRegistry()
	req := httptest.NewRequest(
		"GET",
		"/?installation_id=42&setup_action=install&code=user-code&state="+url.QueryEscape(state),
		nil,
	)

	_, err := registry.CompleteFromState(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "github app")
	assert.NotContains(t, err.Error(), "missing provider field")
}

func TestInitiateGitHubApp_BuildsAuthorizeURL(t *testing.T) {
	t.Parallel()

	registry := connector.NewConnectorRegistry()
	githubApp := &connector.GitHubAppConnector{
		ClientID:     "github-app-client-id",
		ClientSecret: "github-app-client-secret",
		Slug:         "probo-test",
		InstallBase:  "https://github.com/apps",
	}
	require.NoError(
		t,
		registry.RegisterProtocol(connector.GitHubProvider, connector.ProtocolGitHubApp, githubApp),
	)

	organizationID := gid.New(gid.NewTenantID(), 0)
	authorizeURL, err := registry.InitiateGitHubApp(
		context.Background(),
		organizationID,
		connector.InitiateOptions{},
		nil,
	)
	require.NoError(t, err)

	parsed, err := url.Parse(authorizeURL)
	require.NoError(t, err)
	assert.Equal(t, "/login/oauth/authorize", parsed.Path)
	assert.Equal(t, "github-app-client-id", parsed.Query().Get("client_id"))
	require.NotEmpty(t, parsed.Query().Get("state"))
	assert.True(t, connector.IsGitHubAppState(parsed.Query().Get("state")))
}

func TestGitHubAppInstallationURL_BuildsInstallURL(t *testing.T) {
	t.Parallel()

	registry := connector.NewConnectorRegistry()
	githubApp := &connector.GitHubAppConnector{
		ClientSecret: "github-app-client-secret",
		Slug:         "probo-test",
		InstallBase:  "https://github.com/apps",
	}
	require.NoError(
		t,
		registry.RegisterProtocol(connector.GitHubProvider, connector.ProtocolGitHubApp, githubApp),
	)

	state := mintGitHubAppStateToken(t)
	installURL, err := registry.GitHubAppInstallationURL(state)
	require.NoError(t, err)

	parsed, err := url.Parse(installURL)
	require.NoError(t, err)
	assert.Equal(t, "/apps/probo-test/installations/new", parsed.Path)
	assert.Equal(t, state, parsed.Query().Get("state"))
}

func TestExtractProviderFromState_ReturnsProvider(t *testing.T) {
	t.Parallel()

	state, err := statelesstoken.NewToken(
		"secret",
		connector.OAuth2TokenType,
		connector.OAuth2TokenTTL,
		connector.OAuth2State{OrganizationID: "oid", Provider: "SLACK"},
	)
	require.NoError(t, err)

	provider, err := connector.ExtractProviderFromState(state)
	require.NoError(t, err)
	assert.Equal(t, "SLACK", provider)
}

func TestExtractProviderFromState_RejectsMissingProvider(t *testing.T) {
	t.Parallel()

	state, err := statelesstoken.NewToken(
		"secret",
		connector.OAuth2TokenType,
		connector.OAuth2TokenTTL,
		connector.OAuth2State{OrganizationID: "oid"},
	)
	require.NoError(t, err)

	_, err = connector.ExtractProviderFromState(state)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing provider field")
}

func TestExtractProviderFromState_RejectsGitHubAppToken(t *testing.T) {
	t.Parallel()

	_, err := connector.ExtractProviderFromState(mintGitHubAppStateToken(t))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected token type")
}
