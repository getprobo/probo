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

package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

// probeRoundTripFunc lets a test capture the probe request and return a
// canned response without touching the network.
type probeRoundTripFunc func(*http.Request) (*http.Response, error)

func (f probeRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestBuiltinRegistry_ProbeCoverage(t *testing.T) {
	t.Parallel()

	r := NewBuiltinRegistry()

	for _, reg := range r.All() {
		// A workload identity provider's credential is a cloud SDK
		// credential, so its probe is a cloud call rather than an HTTP one and
		// none of the three HTTP forms can express it.
		hasProbe := reg.Probe != nil || reg.Endpoints.Probe != "" || reg.BuildProbeURL != nil
		if reg.SupportsWorkloadIdentity() {
			hasProbe = reg.WorkloadIdentity.Probe != nil
		}

		assert.True(t, hasProbe, "provider %s has no connection probe configured", reg.Provider)
	}
}

func TestBuildDatadogProbeURL(t *testing.T) {
	t.Parallel()

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderDatadog}
	require.NoError(t, conn.SetSettings(&coredata.DatadogConnectorSettings{
		Domain: "us3.datadoghq.com",
		Region: "US3",
	}))

	probeURL, err := buildDatadogProbeURL(conn, Endpoints{})
	require.NoError(t, err)
	assert.Equal(
		t,
		"https://api.us3.datadoghq.com/api/v2/users?page%5Bnumber%5D=0&page%5Bsize%5D=1",
		probeURL,
	)
}

func TestBuildZendeskProbeURL(t *testing.T) {
	t.Parallel()

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderZendesk}
	require.NoError(t, conn.SetSettings(&coredata.ZendeskConnectorSettings{
		Subdomain: "acme",
	}))

	probeURL, err := buildZendeskProbeURL(conn, Endpoints{})
	require.NoError(t, err)
	assert.Contains(t, probeURL, "https://acme.zendesk.com/api/v2/users.json")
}

func TestProbeGitHub_UsesProtocolEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		protocol coredata.ConnectorProtocol
		wantPath string
	}{
		{
			name:     "oauth uses authenticated user",
			protocol: coredata.ConnectorProtocolOAuth2,
			wantPath: "/user",
		},
		{
			name:     "install protocol uses installation repositories",
			protocol: coredata.ConnectorProtocolGitHubApp,
			wantPath: "/installation/repositories",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				client := &http.Client{
					Transport: probeRoundTripFunc(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, tt.wantPath, req.URL.Path)

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader(`{}`)),
							Header:     make(http.Header),
						}, nil
					}),
				}

				err := probeGitHub(
					context.Background(),
					client,
					&coredata.Connector{Protocol: tt.protocol},
					Endpoints{
						Probe:   "https://api.github.com/user",
						APIBase: "https://api.github.com",
					},
				)
				require.NoError(t, err)
			},
		)
	}
}

func TestBuildOktaProbeURL(t *testing.T) {
	t.Parallel()

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderOkta}
	require.NoError(t, conn.SetSettings(&coredata.OktaConnectorSettings{
		Domain: "acme.okta.com",
	}))

	probeURL, err := buildOktaProbeURL(conn, Endpoints{})
	require.NoError(t, err)
	assert.Equal(t, "https://acme.okta.com/api/v1/users?limit=1", probeURL)
}

func TestBuildLangfuseProbeURL(t *testing.T) {
	t.Parallel()

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderLangfuse}
	require.NoError(t, conn.SetSettings(&coredata.LangfuseConnectorSettings{
		BaseURL: "https://us.cloud.langfuse.com",
	}))

	probeURL, err := buildLangfuseProbeURL(conn, Endpoints{})
	require.NoError(t, err)
	assert.Equal(t, "https://us.cloud.langfuse.com/api/public/organizations/memberships", probeURL)
}

func TestBuildPostHogProbeURL(t *testing.T) {
	t.Parallel()

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderPostHog}
	require.NoError(t, conn.SetSettings(&coredata.PostHogConnectorSettings{
		BaseURL: "https://us.posthog.com",
	}))

	probeURL, err := buildPostHogProbeURL(conn)
	require.NoError(t, err)
	assert.Equal(t, "https://us.posthog.com/api/organizations/@current/", probeURL)
}

func TestProbePostHog(t *testing.T) {
	t.Parallel()

	// A cloud OAuth connection carries no region (empty BaseURL): the token is
	// valid on exactly one PostHog region and the other rejects it with
	// 401/403. The probe must try every region and only report the credential
	// rejected when none accept it — mirroring the access-review driver — so an
	// EU token hitting us.posthog.com (probed first) does not falsely mark the
	// source disconnected while its access reviews keep working. A transient
	// 5xx on the token's own region is inconclusive, not a rejection.
	cases := []struct {
		name         string
		baseURL      string
		hostStatus   map[string]int
		wantErr      bool
		wantRejected bool
	}{
		{
			name:       "explicit region accepts",
			baseURL:    "https://us.posthog.com",
			hostStatus: map[string]int{"us.posthog.com": http.StatusOK},
			wantErr:    false,
		},
		{
			name:       "explicit region rejects",
			baseURL:    "https://us.posthog.com",
			hostStatus: map[string]int{"us.posthog.com": http.StatusUnauthorized},
			wantErr:    true,
		},
		{
			name:       "oauth EU token: US refuses, EU accepts",
			baseURL:    "",
			hostStatus: map[string]int{"us.posthog.com": http.StatusUnauthorized, "eu.posthog.com": http.StatusOK},
			wantErr:    false,
		},
		{
			name:       "oauth transient: US refuses, EU errors",
			baseURL:    "",
			hostStatus: map[string]int{"us.posthog.com": http.StatusUnauthorized, "eu.posthog.com": http.StatusInternalServerError},
			wantErr:    false,
		},
		{
			name:         "oauth dead token: every region refuses",
			baseURL:      "",
			hostStatus:   map[string]int{"us.posthog.com": http.StatusForbidden, "eu.posthog.com": http.StatusForbidden},
			wantErr:      true,
			wantRejected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &http.Client{Transport: probeRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				status, ok := tc.hostStatus[r.URL.Host]
				if !ok {
					status = http.StatusNotFound
				}

				return &http.Response{StatusCode: status, Body: http.NoBody, Header: make(http.Header)}, nil
			})}

			conn := &coredata.Connector{Provider: coredata.ConnectorProviderPostHog}
			require.NoError(t, conn.SetSettings(&coredata.PostHogConnectorSettings{BaseURL: tc.baseURL}))

			err := probePostHog(context.Background(), client, conn, posthogRegistration().Endpoints)

			if !tc.wantErr {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)

			// A credential every region refused must surface the sentinel so the
			// probe distinguishes it from an inconclusive/transient failure.
			if tc.wantRejected {
				require.ErrorIs(t, err, drivers.ErrPostHogCredentialRejected)
			}
		})
	}
}

func TestBuildScalewayProbeURL(t *testing.T) {
	t.Parallel()

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderScaleway}
	require.NoError(t, conn.SetSettings(&coredata.ScalewayConnectorSettings{
		OrganizationID: "11111111-2222-3333-4444-555555555555",
	}))

	probeURL, err := buildScalewayProbeURL(conn, Endpoints{APIBase: "https://api.scaleway.com/iam/v1alpha1"})
	require.NoError(t, err)
	assert.Equal(
		t,
		"https://api.scaleway.com/iam/v1alpha1/users?organization_id=11111111-2222-3333-4444-555555555555&page_size=1",
		probeURL,
	)
}

func TestBuildSegmentProbeURL(t *testing.T) {
	t.Parallel()

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderSegment}
	require.NoError(t, conn.SetSettings(&coredata.SegmentConnectorSettings{
		BaseURL: "https://eu1.api.segmentapis.com",
	}))

	probeURL, err := buildSegmentProbeURL(conn, Endpoints{})
	require.NoError(t, err)
	assert.Equal(t, "https://eu1.api.segmentapis.com/users?pagination.count=1", probeURL)
}

func TestProbeOpenRouter(t *testing.T) {
	t.Parallel()

	// probeOpenRouter must reject 401/403 (bad key) and 404 (a valid but
	// personal/non-organization key, which the members endpoint rejects with
	// 404), while letting 2xx pass.
	cases := []struct {
		name       string
		status     int
		wantReject bool
	}{
		{"valid management key", http.StatusOK, false},
		{"revoked key", http.StatusUnauthorized, true},
		{"forbidden key", http.StatusForbidden, true},
		{"personal (non-org) key", http.StatusNotFound, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotURL string

			client := &http.Client{Transport: probeRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				gotURL = r.URL.String()

				return &http.Response{StatusCode: tc.status, Body: http.NoBody, Header: make(http.Header)}, nil
			})}

			err := probeOpenRouter(context.Background(), client, &coredata.Connector{Provider: coredata.ConnectorProviderOpenRouter}, openrouterRegistration().Endpoints)

			assert.Equal(t, "https://openrouter.ai/api/v1/organization/members?limit=1", gotURL)

			if tc.wantReject {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProbeHeroku(t *testing.T) {
	t.Parallel()

	// The fix's contract: probeHeroku must send Heroku's versioned Accept
	// header — a plain "application/json" returns 400, which doProbeRequest
	// reads as connected and masks a dead token — and it must map 401/403 to
	// a rejection while letting 2xx pass.
	cases := []struct {
		name       string
		status     int
		wantReject bool
	}{
		{"valid credential", http.StatusOK, false},
		{"revoked credential", http.StatusUnauthorized, true},
		{"forbidden credential", http.StatusForbidden, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotAccept, gotURL string

			client := &http.Client{Transport: probeRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				gotAccept = r.Header.Get("Accept")
				gotURL = r.URL.String()

				return &http.Response{StatusCode: tc.status, Body: http.NoBody, Header: make(http.Header)}, nil
			})}

			err := probeHeroku(context.Background(), client, &coredata.Connector{Provider: coredata.ConnectorProviderHeroku}, herokuRegistration().Endpoints)

			assert.Equal(t, "application/vnd.heroku+json; version=3", gotAccept)
			assert.Equal(t, "https://api.heroku.com/account", gotURL)

			if tc.wantReject {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProbeRailway(t *testing.T) {
	t.Parallel()

	// Railway returns HTTP 200 with a populated errors array (data.me null) for
	// a rejected token instead of 401/403, so the probe must inspect the body —
	// the generic 401/403-only contract would falsely accept a dead token.
	cases := []struct {
		name       string
		status     int
		body       string
		wantReject bool
		// wantCredentialRejected separates a token Railway refused from an
		// outage. Both are errors, but only the first may reach the caller as
		// a credential rejection: reporting a 502 that way would tell a
		// customer to reconnect a working token.
		wantCredentialRejected bool
		// wantRefused holds Railway to the same reading of a 403 as
		// every provider probed through doProbeRequest: the credential was
		// taken and the operation refused.
		wantRefused bool
	}{
		{"valid token", http.StatusOK, `{"data":{"me":{"id":"u-1"}}}`, false, false, false},
		{"rejected token (200 + errors)", http.StatusOK, `{"errors":[{"message":"Not Authorized"}],"data":null}`, true, true, false},
		{"null me", http.StatusOK, `{"data":{"me":null}}`, true, true, false},
		{"unauthorized status", http.StatusUnauthorized, ``, true, true, false},
		{"forbidden status", http.StatusForbidden, ``, true, true, true},
		{"upstream outage", http.StatusBadGateway, `<html>502 Bad Gateway</html>`, true, false, false},
		{"rate limited", http.StatusTooManyRequests, `{"errors":[{"message":"rate limited"}]}`, true, false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotURL, gotContentType string

			client := &http.Client{Transport: probeRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				gotURL = r.URL.String()
				gotContentType = r.Header.Get("Content-Type")

				return &http.Response{
					StatusCode: tc.status,
					Body:       io.NopCloser(strings.NewReader(tc.body)),
					Header:     make(http.Header),
				}, nil
			})}

			err := probeRailway(context.Background(), client, &coredata.Connector{Provider: coredata.ConnectorProviderRailway}, railwayRegistration().Endpoints)

			assert.Equal(t, "https://backboard.railway.com/graphql/v2", gotURL)
			assert.Equal(t, "application/json", gotContentType)

			if !tc.wantReject {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)

			rejected, isCredentialRejected := errors.AsType[*CredentialRejectedError](err)
			assert.Equal(t, tc.wantCredentialRejected, isCredentialRejected)

			if isCredentialRejected {
				assert.Equal(t, tc.wantRefused, rejected.OperationRefused)
			}
		})
	}
}

func TestProbeCrisp(t *testing.T) {
	t.Parallel()

	// probeCrisp must send the non-auth X-Crisp-Tier header (the generic
	// probeGET does not) and hit the configured website's operators/list
	// endpoint; 401/403 mean a rejected credential, and 404 means a valid token
	// pointed at a wrong/unbound website_id — a permanent misconfiguration that
	// must be rejected at connect time rather than fail every later review.
	cases := []struct {
		name       string
		status     int
		wantReject bool
	}{
		{"valid token", http.StatusOK, false},
		{"revoked token", http.StatusUnauthorized, true},
		{"forbidden token", http.StatusForbidden, true},
		{"wrong or unbound website (404)", http.StatusNotFound, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			conn := &coredata.Connector{Provider: coredata.ConnectorProviderCrisp}
			require.NoError(t, conn.SetSettings(&coredata.CrispConnectorSettings{WebsiteID: "abc-123"}))

			var gotURL, gotTier string

			client := &http.Client{Transport: probeRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				gotURL = r.URL.String()
				gotTier = r.Header.Get("X-Crisp-Tier")

				return &http.Response{StatusCode: tc.status, Body: http.NoBody, Header: make(http.Header)}, nil
			})}

			err := probeCrisp(context.Background(), client, conn, crispRegistration().Endpoints)

			assert.Equal(t, "https://api.crisp.chat/v1/website/abc-123/operators/list", gotURL)
			assert.Equal(t, "plugin", gotTier)

			if tc.wantReject {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestBuildProbeURLFromAPIBase pins the URL of the three probe builders whose
// host is a compile-time constant and now comes from the registration's
// Endpoints.APIBase. Register can only check the agreement between APIBase and
// the STATIC Endpoints.Probe, so these assertions are what keeps a built probe
// URL on the same host as the driver.
func TestBuildProbeURLFromAPIBase(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		reg      *Registration
		settings any
		wantURL  string
	}{
		{
			name:     "neon",
			reg:      neonRegistration(),
			settings: &coredata.NeonConnectorSettings{OrganizationID: "org-cool-breeze-12345678"},
			wantURL:  "https://console.neon.tech/api/v2/organizations/org-cool-breeze-12345678/members?limit=1",
		},
		{
			name:     "render",
			reg:      renderRegistration(),
			settings: &coredata.RenderConnectorSettings{OwnerID: "tea-csp8nlbgbbvc73a8nn9g"},
			wantURL:  "https://api.render.com/v1/owners/tea-csp8nlbgbbvc73a8nn9g/members",
		},
		{
			name:     "qovery",
			reg:      qoveryRegistration(),
			settings: &coredata.QoveryConnectorSettings{OrganizationID: "c4f2de4d-3e50-4f98-bf00-065778f7f5b5"},
			wantURL:  "https://api.qovery.com/organization/c4f2de4d-3e50-4f98-bf00-065778f7f5b5/member",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			conn := &coredata.Connector{Provider: tc.reg.Provider}
			require.NoError(t, conn.SetSettings(tc.settings))
			require.NotNil(t, tc.reg.BuildProbeURL)

			probeURL, err := tc.reg.BuildProbeURL(conn, tc.reg.Endpoints)
			require.NoError(t, err)
			assert.Equal(t, tc.wantURL, probeURL)
		})
	}
}

// TestProbeClosureURL pins the URL each of these Probe closures emits now that
// it composes from Endpoints.APIBase instead of a hardcoded literal. The other
// closures (Crisp, Heroku, OpenRouter, Railway, PostHog) assert their own URL
// in the dedicated tests above.
func TestProbeClosureURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		reg     *Registration
		wantURL string
	}{
		{"linear", linearRegistration(), "https://api.linear.app/graphql"},
		{"monday", mondayRegistration(), "https://api.monday.com/v2"},
		{"anthropic", anthropicRegistration(), "https://api.anthropic.com/v1/organizations/users?limit=1"},
		{"square", squareRegistration(), "https://connect.squareup.com/v2/merchants/me"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotURL string

			client := &http.Client{Transport: probeRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				gotURL = r.URL.String()

				return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Header: make(http.Header)}, nil
			})}

			require.NotNil(t, tc.reg.Probe)

			conn := &coredata.Connector{Provider: tc.reg.Provider}
			require.NoError(t, tc.reg.Probe(context.Background(), client, conn, tc.reg.Endpoints))
			assert.Equal(t, tc.wantURL, gotURL)
		})
	}
}

func TestDoProbeRequest_RejectsHTMLBody(t *testing.T) {
	t.Parallel()

	// A base URL the customer supplies can reach a server that is not the
	// provider's API: a single-page app serving its index for any unknown
	// path, an SSO portal, a proxy error page. Those answer 200, so the
	// status alone reports a healthy connector against which no request will
	// ever work — and the failure only surfaces much later, in a worker.
	cases := []struct {
		name       string
		status     int
		body       string
		wantReject bool
	}{
		{"json object", http.StatusOK, `{"site-name":"Acme"}`, false},
		{"json array", http.StatusOK, `[{"id":1}]`, false},
		{"empty body", http.StatusOK, "", false},
		{"no content", http.StatusNoContent, "", false},
		{"leading whitespace then json", http.StatusOK, "\n  {\"ok\":true}", false},
		{"html index page", http.StatusOK, "<!doctype html><html><body>Metabase</body></html>", true},
		{"uppercase doctype", http.StatusOK, "<!DOCTYPE html>\n<html lang=\"en\">", true},
		{"bare html element", http.StatusOK, "<html>\r\n<head><title>502</title></head>", true},
		{"html with leading whitespace", http.StatusOK, "\n\t<html></html>", true},
		{"html behind a byte order mark", http.StatusOK, "\xef\xbb\xbf<!doctype html>", true},
		{"html padded past the first read", http.StatusOK, strings.Repeat(" ", 600) + "<!doctype html>", true},
		{"proxy maintenance page on 5xx stays transient", http.StatusBadGateway, "<html><h1>502 Bad Gateway</h1></html>", false},
		{"not found html page stays a pass", http.StatusNotFound, "<html>404</html>", false},
		// An XML API is a legitimate thing to reach. Rejecting it would retire
		// a working connector, which is worse than missing a misconfiguration.
		{"xml declaration", http.StatusOK, `<?xml version="1.0"?><Error/>`, false},
		{"bare xml element", http.StatusOK, "<soap:Envelope><soap:Body/></soap:Envelope>", false},
		{"whitespace only", http.StatusOK, strings.Repeat(" ", 2000), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &http.Client{Transport: probeRoundTripFunc(func(_ *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: tc.status,
					Body:       io.NopCloser(strings.NewReader(tc.body)),
					Header:     make(http.Header),
				}, nil
			})}

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.test/api/user", nil)
			require.NoError(t, err)

			err = doProbeRequest(client, req)

			if !tc.wantReject {
				require.NoError(t, err)

				return
			}

			var notAPI *NotAnAPIEndpointError

			require.ErrorAs(t, err, &notAPI)
			assert.Equal(t, tc.status, notAPI.StatusCode)
		})
	}
}

func TestDoProbeRequest_RefusalVerdict(t *testing.T) {
	t.Parallel()

	// The two rejections a customer fixes in different places: a credential
	// the provider will not take at all, and one it takes before refusing the
	// operation. An extra status a provider opts into is the former — it
	// opted in precisely because the status is unambiguous there.
	cases := []struct {
		name        string
		status      int
		extraReject []int
		wantRefused bool
	}{
		{name: "unauthorized status is the credential", status: http.StatusUnauthorized},
		{name: "forbidden status is the authorization", status: http.StatusForbidden, wantRefused: true},
		{
			name:        "opted-in status stays the credential",
			status:      http.StatusNotFound,
			extraReject: []int{http.StatusNotFound},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &http.Client{Transport: probeRoundTripFunc(func(_ *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: tc.status,
					Body:       io.NopCloser(strings.NewReader(`{"error":"nope"}`)),
					Header:     make(http.Header),
				}, nil
			})}

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.test/api/user", nil)
			require.NoError(t, err)

			var rejected *CredentialRejectedError

			require.ErrorAs(t, doProbeRequest(client, req, tc.extraReject...), &rejected)
			assert.Equal(t, tc.status, rejected.StatusCode)
			assert.Equal(t, tc.wantRefused, rejected.OperationRefused)

			// Only a 403 is ever reclassified, so only a 403 pays to keep the
			// provider's explanation.
			if tc.status == http.StatusForbidden {
				assert.Equal(t, `{"error":"nope"}`, string(rejected.body))
			} else {
				assert.Nil(t, rejected.body)
			}
		})
	}

	t.Run("a talkative provider is cut off at the limit", func(t *testing.T) {
		t.Parallel()

		// The hosts this reads from are customer-supplied for Langfuse,
		// Metabase and SigNoz, so the body is bounded rather than trusted.
		client := &http.Client{Transport: probeRoundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusForbidden,
				Body:       io.NopCloser(strings.NewReader(strings.Repeat("a", rejectionBodyLimit*3))),
				Header:     make(http.Header),
			}, nil
		})}

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.test/api/user", nil)
		require.NoError(t, err)

		var rejected *CredentialRejectedError

		require.ErrorAs(t, doProbeRequest(client, req), &rejected)
		assert.Len(t, rejected.body, rejectionBodyLimit)
	})
}

func TestClassifyRejection(t *testing.T) {
	t.Parallel()

	// The classifier only ever sees a 403, and its verdict replaces the one
	// the status gave. Anything else — a 401, a failure that is not a
	// rejection, a provider that registers no classifier — is left alone.
	// Every case carries a body, so the assertion that classification drops it
	// has something to drop, and a classifier can prove it was handed one.
	const explanation = `{"error":"this feature is not available on your plan"}`

	forbidden := func() *CredentialRejectedError {
		return &CredentialRejectedError{
			StatusCode:       http.StatusForbidden,
			OperationRefused: true,
			body:             []byte(explanation),
		}
	}

	cases := []struct {
		name           string
		reg            *Registration
		err            error
		wantRefused    bool
		wantClassified bool
	}{
		{
			name:        "no classifier leaves the status verdict",
			reg:         &Registration{},
			err:         forbidden(),
			wantRefused: true,
		},
		{
			name:           "classifier can demote a forbidden to a bad credential",
			reg:            &Registration{ClassifyRejection: func([]byte) bool { return false }},
			err:            forbidden(),
			wantRefused:    false,
			wantClassified: true,
		},
		{
			name: "classifier is not consulted on a 401",
			reg:  &Registration{ClassifyRejection: func([]byte) bool { return true }},
			err: &CredentialRejectedError{
				StatusCode: http.StatusUnauthorized,
				body:       []byte(explanation),
			},
			wantRefused: false,
		},
		{
			name:           "a wrapped rejection is still refined",
			reg:            &Registration{ClassifyRejection: func([]byte) bool { return false }},
			err:            fmt.Errorf("probe: %w", forbidden()),
			wantRefused:    false,
			wantClassified: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var classified []byte

			if inner := tc.reg.ClassifyRejection; inner != nil {
				tc.reg.ClassifyRejection = func(body []byte) bool {
					classified = body

					return inner(body)
				}
			}

			var rejected *CredentialRejectedError

			require.ErrorAs(t, classifyRejection(tc.reg, tc.err), &rejected)
			assert.Equal(t, tc.wantRefused, rejected.OperationRefused)
			assert.Nil(t, rejected.body, "provider text must not outlive classification")

			if tc.wantClassified {
				assert.Equal(t, explanation, string(classified), "the classifier reads the provider's explanation")
			} else {
				assert.Nil(t, classified)
			}
		})
	}

	// Seven providers answer through a custom Probe closure, and that branch of
	// ProbeConnection is the only other place a rejection carrying provider
	// text can escape. Driven through the registry, not classifyRejection, so
	// removing the call at either branch fails here.
	t.Run("a custom Probe closure is classified too", func(t *testing.T) {
		t.Parallel()

		r := NewRegistry()
		require.NoError(t, r.Register(&Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
			Probe: func(_ context.Context, httpClient *http.Client, _ *coredata.Connector, _ Endpoints) error {
				req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.test/api/user", nil)
				if err != nil {
					return err
				}

				return fmt.Errorf("slack probe: %w", doProbeRequest(httpClient, req))
			},
			ClassifyRejection: func([]byte) bool { return false },
		}))

		client := &http.Client{Transport: probeRoundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusForbidden,
				Body:       io.NopCloser(strings.NewReader(`{"error":"go away"}`)),
				Header:     make(http.Header),
			}, nil
		})}

		err := r.ProbeConnection(
			context.Background(),
			client,
			&coredata.Connector{Provider: coredata.ConnectorProviderSlack},
		)

		var rejected *CredentialRejectedError

		require.ErrorAs(t, err, &rejected)
		assert.False(t, rejected.OperationRefused, "the registration's classifier must reach a custom Probe's rejection")
		assert.Nil(t, rejected.body, "provider text must not outlive classification")
	})

	t.Run("passes a non-rejection through untouched", func(t *testing.T) {
		t.Parallel()

		notAPI := &NotAnAPIEndpointError{StatusCode: http.StatusOK}
		assert.Same(t, notAPI, classifyRejection(&Registration{}, notAPI))
		assert.NoError(t, classifyRejection(&Registration{}, nil))
	})
}

func TestDoProbeRequest_CredentialRejectionWinsOverMarkup(t *testing.T) {
	t.Parallel()

	// A 401 that renders a login page is still a rejected credential, which
	// is the more actionable verdict of the two.
	client := &http.Client{Transport: probeRoundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(strings.NewReader("<html>sign in</html>")),
			Header:     make(http.Header),
		}, nil
	})}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.test/api/user", nil)
	require.NoError(t, err)

	var rejected *CredentialRejectedError

	require.ErrorAs(t, doProbeRequest(client, req), &rejected)
	assert.Equal(t, http.StatusUnauthorized, rejected.StatusCode)
}
