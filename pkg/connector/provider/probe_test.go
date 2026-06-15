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
	"net/http"
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
