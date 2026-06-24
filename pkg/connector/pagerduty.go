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
