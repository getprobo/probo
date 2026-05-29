// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package coredata

import (
	"encoding"
	"fmt"
)

type ConnectorProvider string

const (
	ConnectorProviderSlack           ConnectorProvider = "SLACK"
	ConnectorProviderGoogleWorkspace ConnectorProvider = "GOOGLE_WORKSPACE"
	ConnectorProviderLinear          ConnectorProvider = "LINEAR"
	// _ ConnectorProvider = "FIGMA" — formerly Figma; removed (no driver, no OAuth config, no usage)
	ConnectorProviderOnePassword  ConnectorProvider = "ONE_PASSWORD"
	ConnectorProviderHubSpot      ConnectorProvider = "HUBSPOT"
	ConnectorProviderDocuSign     ConnectorProvider = "DOCUSIGN"
	ConnectorProviderNotion       ConnectorProvider = "NOTION"
	ConnectorProviderBrex         ConnectorProvider = "BREX"
	ConnectorProviderTally        ConnectorProvider = "TALLY"
	ConnectorProviderCloudflare   ConnectorProvider = "CLOUDFLARE"
	ConnectorProviderGrafana      ConnectorProvider = "GRAFANA"
	ConnectorProviderOpenAI       ConnectorProvider = "OPENAI"
	ConnectorProviderPostHog      ConnectorProvider = "POSTHOG"
	ConnectorProviderSentry       ConnectorProvider = "SENTRY"
	ConnectorProviderSupabase     ConnectorProvider = "SUPABASE"
	ConnectorProviderGitHub       ConnectorProvider = "GITHUB"
	ConnectorProviderIntercom     ConnectorProvider = "INTERCOM"
	ConnectorProviderResend       ConnectorProvider = "RESEND"
	ConnectorProviderSendGrid     ConnectorProvider = "SENDGRID"
	ConnectorProviderMicrosoft365 ConnectorProvider = "MICROSOFT_365"
	ConnectorProviderGitLab       ConnectorProvider = "GITLAB"
	ConnectorProviderBitbucket    ConnectorProvider = "BITBUCKET"
	ConnectorProviderHeroku       ConnectorProvider = "HEROKU"
	ConnectorProviderPagerDuty    ConnectorProvider = "PAGERDUTY"
	ConnectorProviderAsana        ConnectorProvider = "ASANA"
	ConnectorProviderNetlify      ConnectorProvider = "NETLIFY"
	ConnectorProviderClickUp      ConnectorProvider = "CLICKUP"
	ConnectorProviderClerk        ConnectorProvider = "CLERK"
	ConnectorProviderVercel       ConnectorProvider = "VERCEL"
	ConnectorProviderMonday       ConnectorProvider = "MONDAY"
	ConnectorProviderMetabase     ConnectorProvider = "METABASE"
	ConnectorProviderTailscale    ConnectorProvider = "TAILSCALE"
	ConnectorProviderAnthropic    ConnectorProvider = "ANTHROPIC"
	ConnectorProviderCursor       ConnectorProvider = "CURSOR"
	ConnectorProviderDatadog      ConnectorProvider = "DATADOG"
)

var (
	_ fmt.Stringer             = ConnectorProvider("")
	_ encoding.TextMarshaler   = ConnectorProvider("")
	_ encoding.TextUnmarshaler = (*ConnectorProvider)(nil)
)

func ConnectorProviders() []ConnectorProvider {
	return []ConnectorProvider{
		ConnectorProviderSlack,
		ConnectorProviderGoogleWorkspace,
		ConnectorProviderLinear,
		ConnectorProviderOnePassword,
		ConnectorProviderHubSpot,
		ConnectorProviderDocuSign,
		ConnectorProviderNotion,
		ConnectorProviderBrex,
		ConnectorProviderTally,
		ConnectorProviderCloudflare,
		ConnectorProviderGrafana,
		ConnectorProviderOpenAI,
		ConnectorProviderPostHog,
		ConnectorProviderSentry,
		ConnectorProviderSupabase,
		ConnectorProviderGitHub,
		ConnectorProviderIntercom,
		ConnectorProviderResend,
		ConnectorProviderSendGrid,
		ConnectorProviderMicrosoft365,
		ConnectorProviderGitLab,
		ConnectorProviderBitbucket,
		ConnectorProviderHeroku,
		ConnectorProviderPagerDuty,
		ConnectorProviderAsana,
		ConnectorProviderNetlify,
		ConnectorProviderClickUp,
		ConnectorProviderClerk,
		ConnectorProviderVercel,
		ConnectorProviderMonday,
		ConnectorProviderMetabase,
		ConnectorProviderTailscale,
		ConnectorProviderAnthropic,
		ConnectorProviderCursor,
		ConnectorProviderDatadog,
	}
}

func (v ConnectorProvider) IsValid() bool {
	switch v {
	case
		ConnectorProviderSlack,
		ConnectorProviderGoogleWorkspace,
		ConnectorProviderLinear,
		ConnectorProviderOnePassword,
		ConnectorProviderHubSpot,
		ConnectorProviderDocuSign,
		ConnectorProviderNotion,
		ConnectorProviderBrex,
		ConnectorProviderTally,
		ConnectorProviderCloudflare,
		ConnectorProviderGrafana,
		ConnectorProviderOpenAI,
		ConnectorProviderPostHog,
		ConnectorProviderSentry,
		ConnectorProviderSupabase,
		ConnectorProviderGitHub,
		ConnectorProviderIntercom,
		ConnectorProviderResend,
		ConnectorProviderSendGrid,
		ConnectorProviderMicrosoft365,
		ConnectorProviderGitLab,
		ConnectorProviderBitbucket,
		ConnectorProviderHeroku,
		ConnectorProviderPagerDuty,
		ConnectorProviderAsana,
		ConnectorProviderNetlify,
		ConnectorProviderClickUp,
		ConnectorProviderClerk,
		ConnectorProviderVercel,
		ConnectorProviderMonday,
		ConnectorProviderMetabase,
		ConnectorProviderTailscale,
		ConnectorProviderAnthropic,
		ConnectorProviderCursor,
		ConnectorProviderDatadog:
		return true
	}

	return false
}

func (v ConnectorProvider) String() string {
	return string(v)
}

func (v ConnectorProvider) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ConnectorProvider) UnmarshalText(text []byte) error {
	val := ConnectorProvider(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ConnectorProvider value: %q", string(text))
	}

	*v = val

	return nil
}
