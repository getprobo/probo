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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/coredata"
)

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
