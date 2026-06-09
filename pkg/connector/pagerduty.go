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

package connector

import "encoding/json"

const (
	PagerDutyProvider = "PAGERDUTY"
)

// AbsorbPagerDutyTokenResponse extracts the customer subdomain that
// PagerDuty's Scoped OAuth surfaces in the token-exchange response body
// and stuffs it into state.ProviderMetadata. The OAuth callback handler
// later writes that value to PagerDutyConnectorSettings. No-op when the
// state's provider is not PagerDuty or when the body has no subdomain.
func AbsorbPagerDutyTokenResponse(state *OAuth2State, body []byte) {
	if state == nil || state.Provider != PagerDutyProvider {
		return
	}

	var pd struct {
		Subdomain string `json:"subdomain"`
	}
	if err := json.Unmarshal(body, &pd); err != nil || pd.Subdomain == "" {
		return
	}

	if state.ProviderMetadata == nil {
		state.ProviderMetadata = map[string]string{}
	}

	state.ProviderMetadata["subdomain"] = pd.Subdomain
}
