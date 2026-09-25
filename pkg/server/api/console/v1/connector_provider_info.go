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
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
)

// connectorProviderSettingInfos projects one connect path's settings list onto
// the GraphQL type. A Registration declares one list per connect path, so
// AccessReviewDrivers calls this once per path. The result is never nil: both
// schema fields are non-null lists, and a provider with no settings on a path
// returns an empty one.
func connectorProviderSettingInfos(settings []provider.ExtraSetting) []*types.ConnectorProviderSettingInfo {
	out := make([]*types.ConnectorProviderSettingInfo, 0, len(settings))

	for _, s := range settings {
		out = append(out, &types.ConnectorProviderSettingInfo{
			Key:      s.Key,
			Label:    s.Label,
			Required: s.Required,
		})
	}

	return out
}

// connectorAPIKeyFormat surfaces the shape a provider expects a pasted key to
// have, so the connect form can check it as the customer leaves the field and
// show it as the field's placeholder. Nil for a provider that declares none.
func connectorAPIKeyFormat(reg *provider.Registration) *types.ConnectorAPIKeyFormat {
	if reg.APIKey == nil || reg.APIKey.KeyFormat == nil {
		return nil
	}

	return &types.ConnectorAPIKeyFormat{
		Pattern: reg.APIKey.KeyFormat.Pattern.String(),
		Example: reg.APIKey.KeyFormat.Example,
	}
}

func connectorProtocols(protocols []connector.ProtocolType) []coredata.ConnectorProtocol {
	out := make([]coredata.ConnectorProtocol, 0, len(protocols))
	for _, protocol := range protocols {
		out = append(out, coredata.ConnectorProtocol(protocol))
	}

	return out
}
