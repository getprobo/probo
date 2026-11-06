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

package console_v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/statelesstoken"
)

func TestFinishConnectorCompletion_FailedInstallCheck(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		err  error
		want string
	}{
		"rejected install": {
			err:  &drivers.InstallRejectedError{Message: "Connect as an admin."},
			want: "Connect as an admin.",
		},
		"check failure": {
			err:  errors.New("slack is down"),
			want: connectorInstallCheckFailedMessage,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			providerRegistry := provider.NewRegistry()
			require.NoError(t, providerRegistry.Register(&provider.Registration{
				Provider:    coredata.ConnectorProviderSlack,
				DisplayName: "Slack",
				ValidateInstall: func(context.Context, *http.Client, provider.Endpoints) error {
					return tc.err
				},
			}))

			logger := log.NewLogger(log.WithName("test"))
			accessReviewSvc := accessreview.NewService(nil, cipher.EncryptionKey{}, connectorprotocol.NewConnectorRegistry(), providerRegistry, logger)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			safeRedirect := saferedirect.New(func(_ context.Context, host string) bool {
				return host == "console.example"
			})

			// A nil probo service panics if the connection reaches Create.
			finishConnectorCompletion(
				recorder,
				req,
				logger,
				baseurl.MustParse("https://console.example"),
				nil,
				accessReviewSvc,
				safeRedirect,
				&connectorprotocol.CompletionState{
					Provider:       string(coredata.ConnectorProviderSlack),
					OrganizationID: gid.New(gid.NewTenantID(), coredata.OrganizationEntityType).String(),
					ContinueURL:    "https://console.example/organizations/acme/access-reviews/sources",
					Connection: &connectorprotocol.SlackConnection{
						AccessToken: "xoxp-test",
						TokenType:   "user",
					},
				},
				nil,
			)

			assert.Equal(t, http.StatusSeeOther, recorder.Code)

			location, err := url.Parse(recorder.Header().Get("Location"))
			require.NoError(t, err)
			assert.Equal(t, "/organizations/acme/access-reviews/sources", location.Path)
			assert.Equal(t, tc.want, location.Query().Get("error"))
			assert.False(t, location.Query().Has("connector_id"))
		})
	}
}

func TestRedirectConnectorCallbackError_GitHubAppPreservesContinuation(t *testing.T) {
	t.Parallel()

	c := &connectorprotocol.GitHubAppConnector{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		InstallBase:  "https://github.example/apps",
		RedirectURI:  "https://probo.test/api/console/v1/connectors/github-app/complete",
	}
	registry := connectorprotocol.NewConnectorRegistry()
	require.NoError(
		t,
		registry.RegisterProtocol(
			connectorprotocol.GitHubProvider,
			connectorprotocol.ProtocolGitHubApp,
			c,
		),
	)

	initReq := httptest.NewRequest(
		http.MethodGet,
		"/?continue=https%3A%2F%2Fconsole.example%2Forganizations%2Facme",
		nil,
	)
	authorizeURL, err := c.Initiate(
		context.Background(),
		connectorprotocol.GitHubProvider,
		gid.New(gid.NewTenantID(), 0),
		connectorprotocol.InitiateOptions{},
		initReq,
	)
	require.NoError(t, err)

	parsedAuthorizeURL, err := url.Parse(authorizeURL)
	require.NoError(t, err)

	callbackReq := httptest.NewRequest(
		http.MethodGet,
		"/?error=access_denied&state="+url.QueryEscape(parsedAuthorizeURL.Query().Get("state")),
		nil,
	)
	recorder := httptest.NewRecorder()
	safeRedirect := saferedirect.New(func(_ context.Context, host string) bool {
		return host == "console.example"
	})

	redirectConnectorCallbackError(
		recorder,
		callbackReq,
		baseurl.MustParse("https://console.example"),
		registry,
		safeRedirect,
		"access_denied",
	)

	assert.Equal(t, http.StatusSeeOther, recorder.Code)

	location, err := url.Parse(recorder.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "https://console.example/organizations/acme", location.Scheme+"://"+location.Host+location.Path)
	assert.Equal(t, "access_denied", location.Query().Get("error"))
}

func TestHandleConnectorComplete_TokenRefusalRedirects(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()

	c := &connectorprotocol.OAuth2Connector{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURI:  "https://probo.test/api/console/v1/connectors/complete",
		AuthURL:      "https://provider.example/authorize",
		TokenURL:     server.URL,
		HTTPClient:   httpclient.DefaultClient(httpclient.WithSSRFProtection(), httpclient.WithSSRFAllowLoopback()),
	}
	registry := connectorprotocol.NewConnectorRegistry()
	require.NoError(t, registry.RegisterProtocol("TEST", connectorprotocol.ProtocolOAuth2, c))

	authorizeURL, err := c.InitiateWithState(
		context.Background(),
		connectorprotocol.OAuth2State{
			OrganizationID: gid.New(gid.NewTenantID(), 0).String(),
			Provider:       "TEST",
			ContinueURL:    "https://console.example/organizations/acme/settings/integrations",
		},
		connectorprotocol.InitiateOptions{Scopes: []string{"read"}},
	)
	require.NoError(t, err)

	parsedAuthorizeURL, err := url.Parse(authorizeURL)
	require.NoError(t, err)

	callbackReq := httptest.NewRequest(
		http.MethodGet,
		"/?code=the-code&state="+url.QueryEscape(parsedAuthorizeURL.Query().Get("state")),
		nil,
	)
	recorder := httptest.NewRecorder()
	safeRedirect := saferedirect.New(func(_ context.Context, host string) bool {
		return host == "console.example"
	})

	handleConnectorComplete(
		log.NewLogger(log.WithName("test")),
		baseurl.MustParse("https://console.example"),
		nil,
		nil,
		registry,
		safeRedirect,
	)(recorder, callbackReq)

	assert.Equal(t, http.StatusSeeOther, recorder.Code)

	location, err := url.Parse(recorder.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "/organizations/acme/settings/integrations", location.Path)
	assert.Equal(t, connectorCompleteFailedMessage, location.Query().Get("error"))
}

func TestHandleConnectorGitHubAppComplete_FailureRedirects(t *testing.T) {
	t.Parallel()

	c := &connectorprotocol.GitHubAppConnector{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		InstallBase:  "https://github.example/apps",
		RedirectURI:  "https://probo.test/api/console/v1/connectors/github-app/complete",
	}
	registry := connectorprotocol.NewConnectorRegistry()
	require.NoError(
		t,
		registry.RegisterProtocol(
			connectorprotocol.GitHubProvider,
			connectorprotocol.ProtocolGitHubApp,
			c,
		),
	)

	authorizeURL, err := c.Initiate(
		context.Background(),
		connectorprotocol.GitHubProvider,
		gid.New(gid.NewTenantID(), 0),
		connectorprotocol.InitiateOptions{},
		httptest.NewRequest(
			http.MethodGet,
			"/?continue=https%3A%2F%2Fconsole.example%2Forganizations%2Facme",
			nil,
		),
	)
	require.NoError(t, err)

	parsedAuthorizeURL, err := url.Parse(authorizeURL)
	require.NoError(t, err)

	// No code, so completion fails right after the state check.
	callbackReq := httptest.NewRequest(
		http.MethodGet,
		"/?state="+url.QueryEscape(parsedAuthorizeURL.Query().Get("state")),
		nil,
	)
	recorder := httptest.NewRecorder()
	safeRedirect := saferedirect.New(func(_ context.Context, host string) bool {
		return host == "console.example"
	})

	handleConnectorGitHubAppComplete(
		log.NewLogger(log.WithName("test")),
		baseurl.MustParse("https://console.example"),
		nil,
		nil,
		registry,
		safeRedirect,
	)(recorder, callbackReq)

	assert.Equal(t, http.StatusSeeOther, recorder.Code)

	location, err := url.Parse(recorder.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "/organizations/acme", location.Path)
	assert.Equal(t, connectorCompleteFailedMessage, location.Query().Get("error"))
}

func TestRedirectConnectorCallbackError_MalformedContinueFallsBack(t *testing.T) {
	t.Parallel()

	stateToken, err := statelesstoken.NewToken(
		"state-secret",
		connectorprotocol.OAuth2TokenType,
		time.Minute,
		connectorprotocol.OAuth2State{Provider: "TEST", ContinueURL: "http://a b"},
	)
	require.NoError(t, err)

	callbackReq := httptest.NewRequest(
		http.MethodGet,
		"/?code=the-code&state="+url.QueryEscape(stateToken),
		nil,
	)
	recorder := httptest.NewRecorder()
	safeRedirect := saferedirect.New(func(_ context.Context, host string) bool {
		return host == "console.example"
	})

	redirectConnectorCallbackError(
		recorder,
		callbackReq,
		baseurl.MustParse("https://console.example"),
		connectorprotocol.NewConnectorRegistry(),
		safeRedirect,
		connectorCompleteFailedMessage,
	)

	assert.Equal(t, http.StatusSeeOther, recorder.Code)

	location, err := url.Parse(recorder.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "console.example", location.Host)
	assert.Equal(t, connectorCompleteFailedMessage, location.Query().Get("error"))
}

func TestOAuthConnectorRawSettings_PagerDutyReconnectMetadata(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	completion := &connectorprotocol.CompletionState{
		Provider: coredata.ConnectorProviderPagerDuty.String(),
		ProviderMetadata: map[string]string{
			"subdomain": "new-account",
		},
		ConnectorID: "reconnect-id",
	}

	raw, ok := oauthConnectorRawSettings(
		log.NewLogger(log.WithName("test")),
		recorder,
		req,
		completion,
		req.URL.Query(),
	)
	require.True(t, ok)

	var settings coredata.PagerDutyConnectorSettings
	require.NoError(t, json.Unmarshal(raw, &settings))
	assert.Equal(t, "new-account", settings.Subdomain)
}

func TestOAuthConnectorRawSettings_VercelReconnectTeam(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?teamId=team_x", nil)
	completion := &connectorprotocol.CompletionState{
		Provider:    coredata.ConnectorProviderVercel.String(),
		ConnectorID: "reconnect-id",
	}

	raw, ok := oauthConnectorRawSettings(
		log.NewLogger(log.WithName("test")),
		recorder,
		req,
		completion,
		req.URL.Query(),
	)
	require.True(t, ok)

	var settings coredata.VercelConnectorSettings
	require.NoError(t, json.Unmarshal(raw, &settings))
	assert.Equal(t, coredata.VercelConnectorSettings{TeamID: "team_x"}, settings)
}

func TestOAuthConnectorRawSettings_VercelWithoutTeam(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	completion := &connectorprotocol.CompletionState{
		Provider:    coredata.ConnectorProviderVercel.String(),
		ConnectorID: "reconnect-id",
	}

	_, ok := oauthConnectorRawSettings(
		log.NewLogger(log.WithName("test")),
		recorder,
		req,
		completion,
		req.URL.Query(),
	)
	require.False(t, ok)
	assert.Equal(t, http.StatusBadGateway, recorder.Code)
}
