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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
)

func (r *Resolver) providerDisplayName(p coredata.ConnectorProvider) string {
	return r.providerRegistry.ProviderDisplayName(p)
}

func (r *Resolver) providerSupportsAPIKey(p coredata.ConnectorProvider) bool {
	if reg, ok := r.providerRegistry.Get(p); ok {
		return reg.SupportsAPIKey
	}

	return false
}

func (r *Resolver) providerSupportsClientCredentials(p coredata.ConnectorProvider) bool {
	if reg, ok := r.providerRegistry.Get(p); ok {
		return reg.SupportsClientCredentials
	}

	return false
}

func (r *Resolver) providerExtraSettings(p coredata.ConnectorProvider) []*types.ConnectorProviderSettingInfo {
	reg, ok := r.providerRegistry.Get(p)
	if !ok || len(reg.ExtraSettings) == 0 {
		return []*types.ConnectorProviderSettingInfo{}
	}

	out := make([]*types.ConnectorProviderSettingInfo, 0, len(reg.ExtraSettings))
	for _, s := range reg.ExtraSettings {
		out = append(out, &types.ConnectorProviderSettingInfo{
			Key:      s.Key,
			Label:    s.Label,
			Required: s.Required,
		})
	}

	return out
}
