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

package accessreview

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

func TestProviderSupportsOrganizationPicker_ConnectionCapabilities(t *testing.T) {
	t.Parallel()

	assert.True(
		t,
		ProviderSupportsOrganizationPicker(
			coredata.ConnectorProviderCloudflare,
			coredata.ConnectorProtocolAPIKey,
		),
	)
	assert.True(
		t,
		ProviderSupportsOrganizationPicker(
			coredata.ConnectorProviderDaytona,
			coredata.ConnectorProtocolAPIKey,
		),
	)
	assert.True(
		t,
		ProviderSupportsOrganizationPicker(
			coredata.ConnectorProviderGitHub,
			coredata.ConnectorProtocolOAuth2,
		),
	)
	assert.False(
		t,
		ProviderSupportsOrganizationPicker(
			coredata.ConnectorProviderGitHub,
			coredata.ConnectorProtocolGitHubApp,
		),
	)
}

// TestDaytonaOrganizationPicker walks the whole picker path for a provider
// whose organization is discovered from the pasted key rather than typed by
// the customer: the id ListOrgs reports is the one SetOrganizationSettings
// persists and SelectedSlug reads back, so the driver ends up scoped to the
// organization the customer was shown.
func TestDaytonaOrganizationPicker(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/organizations", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"org-1","name":"Acme"}]`))
	}))
	defer server.Close()

	cfg, ok := providerOrgConfigs[coredata.ConnectorProviderDaytona]
	require.True(t, ok, "daytona must declare a picker config")
	require.NotNil(t, cfg.ListOrgs)
	assert.True(t, cfg.NeedsPicker)

	orgs, err := cfg.ListOrgs(context.Background(), server.Client(), server.URL)
	require.NoError(t, err)
	require.Len(t, orgs, 1)
	assert.Equal(t, drivers.Organization{Slug: "org-1", DisplayName: "Acme"}, orgs[0])

	reg, ok := provider.NewBuiltinRegistry().Get(coredata.ConnectorProviderDaytona)
	require.True(t, ok)
	require.NotNil(t, reg.SetOrganizationSettings)

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderDaytona}
	require.NoError(t, reg.SetOrganizationSettings(conn, orgs[0].Slug))
	assert.Equal(t, "org-1", cfg.SelectedSlug(conn))
}
