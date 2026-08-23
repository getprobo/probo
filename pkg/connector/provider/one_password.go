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

func onePasswordRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderOnePassword,
		DisplayName:      "1Password",
		DocumentationURL: accessReviewDocsURL("one-password"),
		// APIBase is deliberately empty: 1Password has no single data host to
		// name. Four host families live under this one registration —
		// the per-connection SCIM bridge URL (customer-hosted, from
		// APIKey.ExtraSettings), the three regional Users API hosts resolved
		// from the Region setting (drivers.onePasswordBaseURL), and the
		// events.1password.com probe below, which is yet another host and is
		// therefore not the root any driver joins onto. A single APIBase would
		// have to be wrong for at least two of them, and Register's
		// Probe-host-matches-APIBase check would reject the honest pairing.
		Endpoints: Endpoints{
			Probe: "https://events.1password.com/api/v1/auditevents",
		},
		// Two settings shapes, one per connect path, because a different
		// driver sits behind each:
		//  - API key:            SCIMBridgeURL      (SCIM-bridge driver).
		//  - Client credentials: AccountID + Region (Users API driver).
		APIKey: &APIKeySpec{
			ExtraSettings: []ExtraSetting{
				{Key: "scimBridgeUrl", Label: "SCIM Bridge URL", Required: true},
			},
		},
		ClientCredentials: &ClientCredentialsSpec{
			ExtraSettings: []ExtraSetting{
				{Key: "accountId", Label: "Account ID", Required: true},
				{Key: "region", Label: "Region", Required: true},
			},
		},
	}
}
