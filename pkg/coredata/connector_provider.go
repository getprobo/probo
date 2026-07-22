// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	ConnectorProviderSigNoz       ConnectorProvider = "SIGNOZ"
	ConnectorProviderSupabase     ConnectorProvider = "SUPABASE"
	ConnectorProviderBetterStack  ConnectorProvider = "BETTER_STACK"
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
	// ConnectorProviderClerk is retained for existing connectors but is
	// no longer a registerable access-review provider: Clerk's Backend API
	// (secret key) only exposes the customer's application end-users, not
	// the Clerk workspace/dashboard team who administer the platform, so a
	// campaign reviews the wrong population. Kept in IsValid and the
	// GraphQL enum so stored CLERK rows still validate and serialize;
	// dropped from ConnectorProviders and unregistered from the builtin
	// registry so it cannot be added or fetched.
	ConnectorProviderClerk      ConnectorProvider = "CLERK"
	ConnectorProviderVercel     ConnectorProvider = "VERCEL"
	ConnectorProviderMonday     ConnectorProvider = "MONDAY"
	ConnectorProviderMetabase   ConnectorProvider = "METABASE"
	ConnectorProviderTailscale  ConnectorProvider = "TAILSCALE"
	ConnectorProviderAnthropic  ConnectorProvider = "ANTHROPIC"
	ConnectorProviderCursor     ConnectorProvider = "CURSOR"
	ConnectorProviderDatadog    ConnectorProvider = "DATADOG"
	ConnectorProviderOkta       ConnectorProvider = "OKTA"
	ConnectorProviderZendesk    ConnectorProvider = "ZENDESK"
	ConnectorProviderQovery     ConnectorProvider = "QOVERY"
	ConnectorProviderRender     ConnectorProvider = "RENDER"
	ConnectorProviderNeon       ConnectorProvider = "NEON"
	ConnectorProviderMercury    ConnectorProvider = "MERCURY"
	ConnectorProviderApollo     ConnectorProvider = "APOLLO"
	ConnectorProviderDeepgram   ConnectorProvider = "DEEPGRAM"
	ConnectorProviderClickHouse ConnectorProvider = "CLICKHOUSE"
	ConnectorProviderLangfuse   ConnectorProvider = "LANGFUSE"
	ConnectorProviderPylon      ConnectorProvider = "PYLON"
	ConnectorProviderOpenRouter ConnectorProvider = "OPENROUTER"
	ConnectorProviderIncidentIO ConnectorProvider = "INCIDENT_IO"
	ConnectorProviderBrevo      ConnectorProvider = "BREVO"
	ConnectorProviderScaleway   ConnectorProvider = "SCALEWAY"
	ConnectorProviderYousign    ConnectorProvider = "YOUSIGN"
	ConnectorProviderRailway    ConnectorProvider = "RAILWAY"
	ConnectorProviderCrisp      ConnectorProvider = "CRISP"
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
		ConnectorProviderSigNoz,
		ConnectorProviderSupabase,
		ConnectorProviderBetterStack,
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
		ConnectorProviderVercel,
		ConnectorProviderMonday,
		ConnectorProviderMetabase,
		ConnectorProviderTailscale,
		ConnectorProviderAnthropic,
		ConnectorProviderCursor,
		ConnectorProviderDatadog,
		ConnectorProviderOkta,
		ConnectorProviderZendesk,
		ConnectorProviderQovery,
		ConnectorProviderRender,
		ConnectorProviderNeon,
		ConnectorProviderMercury,
		ConnectorProviderApollo,
		ConnectorProviderDeepgram,
		ConnectorProviderClickHouse,
		ConnectorProviderLangfuse,
		ConnectorProviderPylon,
		ConnectorProviderOpenRouter,
		ConnectorProviderIncidentIO,
		ConnectorProviderBrevo,
		ConnectorProviderScaleway,
		ConnectorProviderYousign,
		ConnectorProviderRailway,
		ConnectorProviderCrisp,
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
		ConnectorProviderSigNoz,
		ConnectorProviderSupabase,
		ConnectorProviderBetterStack,
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
		ConnectorProviderOkta,
		ConnectorProviderZendesk,
		ConnectorProviderQovery,
		ConnectorProviderRender,
		ConnectorProviderNeon,
		ConnectorProviderMercury,
		ConnectorProviderApollo,
		ConnectorProviderDeepgram,
		ConnectorProviderClickHouse,
		ConnectorProviderLangfuse,
		ConnectorProviderPylon,
		ConnectorProviderOpenRouter,
		ConnectorProviderIncidentIO,
		ConnectorProviderBrevo,
		ConnectorProviderScaleway,
		ConnectorProviderYousign,
		ConnectorProviderRailway,
		ConnectorProviderCrisp:
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
