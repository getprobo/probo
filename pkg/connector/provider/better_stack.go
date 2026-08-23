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

// betterStackRegistration wires the Better Stack Uptime access-review
// connector. Better Stack has no third-party OAuth app for listing team
// members (its OAuth is an end-user MCP sign-in), so the connector is
// API-key only: the operator supplies a Bearer API token plus the team
// name that scopes the team-members listing.
func betterStackRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderBetterStack,
		DisplayName:      "Better Stack",
		DocumentationURL: accessReviewDocsURL("better-stack"),
		Endpoints: Endpoints{
			APIBase: "https://betterstack.com/api/v2",
			Probe:   "https://betterstack.com/api/v2/team-members",
		},
		APIKey: &APIKeySpec{
			ExtraSettings: []ExtraSetting{
				{Key: "teamName", Label: "Team Name", Required: true},
			},
		},
	}
}
