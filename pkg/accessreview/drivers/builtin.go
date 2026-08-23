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

package drivers

import (
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
)

// NewBuiltinRegistry returns a *Registry holding every provider an access
// review can pull accounts from. It panics on a duplicate or nil registration,
// both programmer errors caught at process start rather than at fetch time.
//
// A provider present in the connector catalog but absent here can be connected
// and probed but not reviewed, which is what keeps it out of the access-source
// picker.
func NewBuiltinRegistry() *Registry {
	r := NewRegistry()

	for _, entry := range []struct {
		provider coredata.ConnectorProvider
		factory  Factory
	}{
		{coredata.ConnectorProviderAnthropic, anthropicSource()},
		{coredata.ConnectorProviderApollo, apolloSource()},
		{coredata.ConnectorProviderAsana, asanaSource()},
		{coredata.ConnectorProviderAuthentik, authentikSource()},
		{coredata.ConnectorProviderBetterStack, betterStackSource()},
		{coredata.ConnectorProviderBitbucket, bitbucketSource()},
		{coredata.ConnectorProviderBrevo, brevoSource()},
		{coredata.ConnectorProviderBrex, brexSource()},
		{coredata.ConnectorProviderCalCom, calcomSource()},
		{coredata.ConnectorProviderCalendly, calendlySource()},
		{coredata.ConnectorProviderClickHouse, clickhouseSource()},
		{coredata.ConnectorProviderClickUp, clickupSource()},
		{coredata.ConnectorProviderCloudflare, cloudflareSource()},
		{coredata.ConnectorProviderCrisp, crispSource()},
		{coredata.ConnectorProviderCursor, cursorSource()},
		{coredata.ConnectorProviderDatadog, datadogSource()},
		{coredata.ConnectorProviderDeepgram, deepgramSource()},
		{coredata.ConnectorProviderDocuSign, docusignSource()},
		{coredata.ConnectorProviderDotfile, dotfileSource()},
		{coredata.ConnectorProviderGitHub, githubSource()},
		{coredata.ConnectorProviderGitLab, gitlabSource()},
		{coredata.ConnectorProviderGoogleAnalytics, googleAnalyticsSource()},
		{coredata.ConnectorProviderGoogleWorkspace, googleWorkspaceSource()},
		{coredata.ConnectorProviderGrafana, grafanaSource()},
		{coredata.ConnectorProviderHeroku, herokuSource()},
		{coredata.ConnectorProviderHubSpot, hubspotSource()},
		{coredata.ConnectorProviderIncidentIO, incidentioSource()},
		{coredata.ConnectorProviderIntercom, intercomSource()},
		{coredata.ConnectorProviderLangfuse, langfuseSource()},
		{coredata.ConnectorProviderLinear, linearSource()},
		{coredata.ConnectorProviderMercury, mercurySource()},
		{coredata.ConnectorProviderMetabase, metabaseSource()},
		{coredata.ConnectorProviderMicrosoft365, microsoft365Source()},
		{coredata.ConnectorProviderMonday, mondaySource()},
		{coredata.ConnectorProviderNeon, neonSource()},
		{coredata.ConnectorProviderNetlify, netlifySource()},
		{coredata.ConnectorProviderNotion, notionSource()},
		{coredata.ConnectorProviderNuki, nukiSource()},
		{coredata.ConnectorProviderOkta, oktaSource()},
		{coredata.ConnectorProviderOnePassword, onePasswordSource()},
		{coredata.ConnectorProviderOpenAI, openaiSource()},
		{coredata.ConnectorProviderOpenRouter, openrouterSource()},
		{coredata.ConnectorProviderPagerDuty, pagerdutySource()},
		{coredata.ConnectorProviderPostHog, posthogSource()},
		{coredata.ConnectorProviderPylon, pylonSource()},
		{coredata.ConnectorProviderQovery, qoverySource()},
		{coredata.ConnectorProviderRailway, railwaySource()},
		{coredata.ConnectorProviderRender, renderSource()},
		{coredata.ConnectorProviderResend, resendSource()},
		{coredata.ConnectorProviderScaleway, scalewaySource()},
		{coredata.ConnectorProviderSegment, segmentSource()},
		{coredata.ConnectorProviderSendGrid, sendgridSource()},
		{coredata.ConnectorProviderSentry, sentrySource()},
		{coredata.ConnectorProviderSigNoz, signozSource()},
		{coredata.ConnectorProviderSlack, slackSource()},
		{coredata.ConnectorProviderSquare, squareSource()},
		{coredata.ConnectorProviderSupabase, supabaseSource()},
		{coredata.ConnectorProviderTailscale, tailscaleSource()},
		{coredata.ConnectorProviderTally, tallySource()},
		{coredata.ConnectorProviderUpCloud, upcloudSource()},
		{coredata.ConnectorProviderVercel, vercelSource()},
		{coredata.ConnectorProviderYousign, yousignSource()},
		{coredata.ConnectorProviderZendesk, zendeskSource()},
	} {
		if err := r.Register(entry.provider, entry.factory); err != nil {
			panic(fmt.Sprintf("cannot build builtin access review driver registry: %v", err))
		}
	}

	return r
}
