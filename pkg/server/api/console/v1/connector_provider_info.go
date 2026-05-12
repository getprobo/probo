// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package console_v1

import (
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
)

var apiKeyProviders = map[coredata.ConnectorProvider]bool{
	coredata.ConnectorProviderHubSpot:     true,
	coredata.ConnectorProviderDocuSign:    true,
	coredata.ConnectorProviderNotion:      true,
	coredata.ConnectorProviderGitHub:      true,
	coredata.ConnectorProviderSentry:      true,
	coredata.ConnectorProviderIntercom:    true,
	coredata.ConnectorProviderBrex:        true,
	coredata.ConnectorProviderTally:       true,
	coredata.ConnectorProviderCloudflare:  true,
	coredata.ConnectorProviderOpenAI:      true,
	coredata.ConnectorProviderSupabase:    true,
	coredata.ConnectorProviderResend:      true,
	coredata.ConnectorProviderOnePassword: true,
}

var clientCredentialsProviders = map[coredata.ConnectorProvider]bool{
	coredata.ConnectorProviderOnePassword: true,
}

var providerExtraSettingsMap = map[coredata.ConnectorProvider][]*types.ConnectorProviderSettingInfo{
	coredata.ConnectorProviderGitHub: {
		{Key: "organization", Label: "Organization", Required: true},
	},
	coredata.ConnectorProviderSentry: {
		{Key: "organizationSlug", Label: "Organization Slug", Required: true},
	},
	coredata.ConnectorProviderTally: {
		{Key: "organizationId", Label: "Organization ID", Required: true},
	},
	coredata.ConnectorProviderSupabase: {
		{Key: "organizationSlug", Label: "Organization Slug", Required: true},
	},
	coredata.ConnectorProviderOnePassword: {
		{Key: "accountId", Label: "Account ID", Required: true},
		{Key: "region", Label: "Region", Required: true},
	},
}

func providerDisplayName(provider coredata.ConnectorProvider) string {
	return drivers.ProviderDisplayName(provider)
}

func providerSupportsAPIKey(provider coredata.ConnectorProvider) bool {
	return apiKeyProviders[provider]
}

func providerSupportsClientCredentials(provider coredata.ConnectorProvider) bool {
	return clientCredentialsProviders[provider]
}

func providerExtraSettings(provider coredata.ConnectorProvider) []*types.ConnectorProviderSettingInfo {
	if settings, ok := providerExtraSettingsMap[provider]; ok {
		return settings
	}
	return []*types.ConnectorProviderSettingInfo{}
}
