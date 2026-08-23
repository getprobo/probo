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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/saferedirect"
)

func TestHandleConnectorCallbackError_GitHubAppPreservesContinuation(t *testing.T) {
	t.Parallel()

	c := &connector.GitHubAppConnector{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		InstallBase:  "https://github.example/apps",
	}
	registry := connector.NewConnectorRegistry()
	require.NoError(
		t,
		registry.RegisterProtocol(
			connector.GitHubProvider,
			connector.ProtocolGitHubApp,
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
		connector.GitHubProvider,
		gid.New(gid.NewTenantID(), 0),
		connector.InitiateOptions{},
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

	handleConnectorCallbackError(
		recorder,
		callbackReq,
		log.NewLogger(log.WithName("test")),
		baseurl.MustParse("https://console.example"),
		registry,
		safeRedirect,
		callbackReq.URL.Query(),
	)

	assert.Equal(t, http.StatusSeeOther, recorder.Code)

	location, err := url.Parse(recorder.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "https://console.example/organizations/acme", location.Scheme+"://"+location.Host+location.Path)
	assert.Equal(t, "access_denied", location.Query().Get("error"))
}

func TestOAuthConnectorRawSettings_PagerDutyReconnectMetadata(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	completion := &connector.CompletionState{
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
