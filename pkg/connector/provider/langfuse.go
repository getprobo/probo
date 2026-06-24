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
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func langfuseRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderLangfuse,
		DisplayName:    "Langfuse",
		SupportsAPIKey: true,
		// Langfuse's organization-scoped public API authenticates with HTTP
		// Basic auth where the credential is publicKey:secretKey.
		// APIKeyBasicAuthUserPass base64s the verbatim "publicKey:secretKey" the
		// operator pastes (the empty-password APIKeyBasicAuth cannot carry
		// the secret). The org API key is bound to one organization, so
		// there is nothing to pick; only the regional/self-hosted base URL
		// is per-tenant and is surfaced as an extra setting.
		APIKeyBasicAuthUserPass: true,
		ExtraSettings: []ExtraSetting{
			{Key: "baseUrl", Label: "Base URL", Required: true},
		},
		// BuildProbeURL derives the probe endpoint from the per-connection
		// base URL (the host is regional/self-hosted, so a static ProbeURL
		// cannot express it); the transport attaches the Basic credential
		// and a dead key returns 401/403.
		BuildProbeURL: buildLangfuseProbeURL,
		//
		// No NewNameResolver: the memberships endpoint carries no
		// organization name, so the source keeps its generic name.
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			settings, err := coredata.ConnectorSettings[coredata.LangfuseConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read langfuse connector settings: %w", err)
			}

			baseURL, err := normalizeLangfuseBaseURL(settings.BaseURL)
			if err != nil {
				return nil, fmt.Errorf("cannot create langfuse driver: %w", err)
			}

			return drivers.NewLangfuseDriver(c, baseURL), nil
		},
	}
}

func normalizeLangfuseBaseURL(raw string) (string, error) {
	baseURL := strings.TrimSpace(raw)
	if baseURL == "" {
		return "", fmt.Errorf("base_url is required")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("base_url must be a valid URL: %w", err)
	}

	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("base_url must be an http(s) URL")
	}

	u.Path = strings.TrimRight(u.Path, "/")
	u.RawQuery = ""
	u.Fragment = ""

	return u.String(), nil
}
