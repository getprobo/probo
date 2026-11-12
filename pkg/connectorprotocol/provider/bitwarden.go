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

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

const bitwardenDisplayName = "Bitwarden"

// bitwardenRegistration describes the Bitwarden Password Manager connector:
// OAuth2 client credentials from the organization API key, which Bitwarden
// offers on Teams and Enterprise plans. The key is bound to one organization,
// so there is nothing to pick, and no Public API endpoint returns the
// organization's name, so there is no name resolver.
func bitwardenRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderBitwarden,
		DisplayName:      bitwardenDisplayName,
		DocumentationURL: accessReviewDocsURL("bitwarden"),
		// Endpoints is deliberately empty: both hosts follow the connector's
		// Region setting (drivers.BitwardenAPIBaseURL, BitwardenTokenURL).
		BuildProbeURL:      buildBitwardenProbeURL,
		InitialAccountFunc: bitwardenInitialAccount,
		ClientCredentials: &ClientCredentialsConfig{
			ExtraSettings: []ExtraSetting{
				{
					Key:         "region",
					Label:       "Region (us or eu)",
					Required:    true,
					Description: "us if you sign in at vault.bitwarden.com, eu if you sign in at vault.bitwarden.eu.",
				},
			},
			TokenURL: bitwardenTokenURL,
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, _ Endpoints) (drivers.Driver, error) {
			baseURL, err := bitwardenAPIBaseURL(conn)
			if err != nil {
				return nil, fmt.Errorf("cannot create bitwarden driver: %w", err)
			}

			return drivers.NewBitwardenDriver(c, baseURL), nil
		},
	}
}

// bitwardenInitialAccount is the organization the API key belongs to, under the
// provider's name since the Public API does not return the organization's.
func bitwardenInitialAccount(c *coredata.Connector) (string, string, error) {
	s, err := coredata.ConnectorSettings[coredata.BitwardenConnectorSettings](c)
	if err != nil {
		return "", "", fmt.Errorf("cannot read bitwarden connector settings: %w", err)
	}

	return s.OrganizationID, bitwardenDisplayName, nil
}

func bitwardenAPIBaseURL(c *coredata.Connector) (string, error) {
	s, err := coredata.ConnectorSettings[coredata.BitwardenConnectorSettings](c)
	if err != nil {
		return "", fmt.Errorf("cannot read bitwarden connector settings: %w", err)
	}

	apiBase, err := drivers.BitwardenAPIBaseURL(s.Region)
	if err != nil {
		return "", fmt.Errorf("cannot resolve bitwarden api base url: %w", err)
	}

	return apiBase, nil
}

func bitwardenTokenURL(c *coredata.Connector) (string, error) {
	s, err := coredata.ConnectorSettings[coredata.BitwardenConnectorSettings](c)
	if err != nil {
		return "", fmt.Errorf("cannot read bitwarden connector settings: %w", err)
	}

	tokenURL, err := drivers.BitwardenTokenURL(s.Region)
	if err != nil {
		return "", fmt.Errorf("cannot resolve bitwarden token url: %w", err)
	}

	return tokenURL, nil
}
