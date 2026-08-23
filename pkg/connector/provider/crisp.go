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
	"go.probo.inc/probo/pkg/coredata"
)

func crispRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderCrisp,
		DisplayName:      "Crisp",
		DocumentationURL: accessReviewDocsURL("crisp"),
		Endpoints: Endpoints{
			// Every endpoint the driver calls shares the /v1 prefix, so the
			// version segment stays in APIBase.
			APIBase: "https://api.crisp.chat/v1",
		},
		// Model B: the plugin token is Probo's own Crisp Marketplace plugin
		// credential, held server-side in bootstrap config, not pasted by
		// the customer. ManagedAPIKey injects it at connect time; the
		// customer supplies only the Website ID. SupportsAPIKey stays false
		// so the provider is hidden from the driver catalog until the
		// operator configures PROBOD_CONNECTOR_CRISP_PLUGIN_TOKEN — it ships
		// deactivated until Crisp validates the production plugin and
		// activates with no code change once the token is set.
		// The per-website plugin API also needs the plugin ID (a distinct value
		// from the token identifier), supplied via bootstrap alongside the
		// token. Require it so Crisp stays hidden until both are configured
		// rather than surfacing as connectable and failing at connect time.
		// Crisp authenticates with the plugin token presented as HTTP Basic,
		// the credential being the verbatim "identifier:key" pair.
		// APIKeyBasicUserPass base64-encodes it (the empty-password
		// APIKeyBasic cannot carry the key). A plugin token can serve
		// several websites, so the reviewed website is captured via
		// APIKey.ExtraSettings. Every request also needs the non-auth X-Crisp-Tier
		// header (set by the driver/probe/name resolver), so the probe is a
		// custom closure.
		// See Registration.EndpointOverrideUnsupported: GetCrispSubscriptionSettings,
		// called at connect time from connector_settings.go to verify plugin
		// ownership, hits crispDefaultBaseURL directly instead of APIBase.
		EndpointOverrideUnsupported: "its connect-time plugin-subscription check calls a host pinned in GetCrispSubscriptionSettings instead of APIBase",
		APIKey: &APIKeySpec{
			Managed:            true,
			RequiresResourceID: true,
			Presentation:       APIKeyBasicUserPass,
			ExtraSettings: []ExtraSetting{
				{Key: "websiteId", Label: "Website ID", Required: true},
			},
		},
		Probe: ProbeOver(probeCrisp),
	}
}
