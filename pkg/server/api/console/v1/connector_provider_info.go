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
