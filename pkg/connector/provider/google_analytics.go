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

func googleAnalyticsRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderGoogleAnalytics,
		DisplayName: "Google Analytics",
		Endpoints: Endpoints{
			Auth:  "https://accounts.google.com/o/oauth2/v2/auth",
			Token: "https://oauth2.googleapis.com/token",
			// The Analytics Admin API, a different host from the Google OAuth
			// endpoints above. v1alpha is part of the root because it is the
			// only version exposing accessBindings. The scopes below only look
			// like URLs — they are Google scope identifiers, not endpoints, and
			// must not be derived from this.
			APIBase: "https://analyticsadmin.googleapis.com/v1alpha",
		},
		OAuth2: &OAuth2Spec{
			ExtraAuthParams: map[string]string{
				"access_type": "offline",
				"prompt":      "consent",
			},
			SupportsIncrementalAuth: true,
			Scopes: []string{
				"https://www.googleapis.com/auth/analytics.readonly",
				"https://www.googleapis.com/auth/analytics.manage.users.readonly",
			},
		},
		// analytics.readonly is required to LIST accounts and properties (the
		// picker and the probe); analytics.manage.users.readonly is required to
		// read the access bindings. The manage.users scope alone cannot list
		// accounts (it returns 403), so both are requested.
		// BuildProbeURL targets the selected account's accessBindings rather
		// than the accounts list: listing accounts only needs
		// analytics.readonly, so a non-Administrator connection (or one where
		// the user declined manage.users.readonly on Google's granular consent
		// screen) would probe green and then 403 on every fetch.
		BuildProbeURL: buildGoogleAnalyticsProbeURL,
		SetOrganizationSettings: func(c *coredata.Connector, accountID string) error {
			return c.SetSettings(&coredata.GoogleAnalyticsConnectorSettings{AccountID: accountID})
		},
	}
}
