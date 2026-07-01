// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package provider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		hasProbe := reg.Probe != nil || reg.ProbeURL != "" || reg.BuildProbeURL != nil
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

	probeURL, err := buildDatadogProbeURL(conn)
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

	probeURL, err := buildZendeskProbeURL(conn)
	require.NoError(t, err)
	assert.Contains(t, probeURL, "https://acme.zendesk.com/api/v2/users.json")
}

func TestBuildOktaProbeURL(t *testing.T) {
	t.Parallel()

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderOkta}
	require.NoError(t, conn.SetSettings(&coredata.OktaConnectorSettings{
		Domain: "acme.okta.com",
	}))

	probeURL, err := buildOktaProbeURL(conn)
	require.NoError(t, err)
	assert.Equal(t, "https://acme.okta.com/api/v1/users?limit=1", probeURL)
}

func TestBuildLangfuseProbeURL(t *testing.T) {
	t.Parallel()

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderLangfuse}
	require.NoError(t, conn.SetSettings(&coredata.LangfuseConnectorSettings{
		BaseURL: "https://us.cloud.langfuse.com",
	}))

	probeURL, err := buildLangfuseProbeURL(conn)
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

func TestBuildScalewayProbeURL(t *testing.T) {
	t.Parallel()

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderScaleway}
	require.NoError(t, conn.SetSettings(&coredata.ScalewayConnectorSettings{
		OrganizationID: "11111111-2222-3333-4444-555555555555",
	}))

	probeURL, err := buildScalewayProbeURL(conn)
	require.NoError(t, err)
	assert.Equal(
		t,
		"https://api.scaleway.com/iam/v1alpha1/users?organization_id=11111111-2222-3333-4444-555555555555&page_size=1",
		probeURL,
	)
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

			err := probeOpenRouter(context.Background(), client, &coredata.Connector{Provider: coredata.ConnectorProviderOpenRouter})

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

			err := probeHeroku(context.Background(), client, &coredata.Connector{Provider: coredata.ConnectorProviderHeroku})

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
	}{
		{"valid token", http.StatusOK, `{"data":{"me":{"id":"u-1"}}}`, false},
		{"rejected token (200 + errors)", http.StatusOK, `{"errors":[{"message":"Not Authorized"}],"data":null}`, true},
		{"null me", http.StatusOK, `{"data":{"me":null}}`, true},
		{"unauthorized status", http.StatusUnauthorized, ``, true},
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

			err := probeRailway(context.Background(), client, &coredata.Connector{Provider: coredata.ConnectorProviderRailway})

			assert.Equal(t, "https://backboard.railway.com/graphql/v2", gotURL)
			assert.Equal(t, "application/json", gotContentType)

			if tc.wantReject {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
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

			err := probeCrisp(context.Background(), client, conn)

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
