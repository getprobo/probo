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

package mcp_v1

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

func (r *Resolver) listConnectorProviders(
	ctx context.Context,
) (*mcp.CallToolResult, types.ListConnectorProvidersOutput, error) {
	identity := authn.IdentityFromContext(ctx)
	if _, err := r.Authorize(ctx, identity.ID, accessreview.ActionDriverCatalogList); err != nil {
		return nil, types.ListConnectorProvidersOutput{}, err
	}

	registry := r.proboSvc.Connectors.ProviderRegistry()
	if registry == nil {
		return nil, types.ListConnectorProvidersOutput{}, fmt.Errorf("internal server error")
	}

	entries := catalogEntries(registry)
	slices.SortFunc(entries, func(a, b *types.ConnectorProviderCatalogEntry) int {
		return strings.Compare(a.DisplayName, b.DisplayName)
	})

	return nil, types.ListConnectorProvidersOutput{Providers: entries}, nil
}

func catalogEntries(registry *provider.Registry) []*types.ConnectorProviderCatalogEntry {
	registrations := registry.All()
	entries := make([]*types.ConnectorProviderCatalogEntry, 0, len(registrations))

	for _, reg := range registrations {
		if reg == nil || (reg.NewDriver == nil && !reg.SupportsWorkloadIdentity()) {
			continue
		}

		var tokenURL *string
		if pinned := reg.Endpoints.Token; pinned != "" && reg.SupportsClientCredentials() {
			tokenURL = new(pinned)
		}

		entries = append(entries, &types.ConnectorProviderCatalogEntry{
			Provider:                       string(reg.Provider),
			DisplayName:                    reg.DisplayName,
			APIKeySupported:                reg.SupportsAPIKey(),
			APIKeyManaged:                  reg.IsManagedAPIKey() && registry.ManagedConnectorReady(reg.Provider),
			InstallSupported:               reg.SupportsInstall() && registry.ManagedConnectorReady(reg.Provider),
			APIKeyExtraSettings:            settingInfos(reg.APIKeyExtraSettings()),
			ClientCredentialsSupported:     reg.SupportsClientCredentials(),
			ClientCredentialsTokenURL:      tokenURL,
			ClientCredentialsExtraSettings: settingInfos(reg.ClientCredentialsExtraSettings()),
		})
	}

	return entries
}

func settingInfos(settings []provider.ExtraSetting) []*types.ConnectorProviderSetting {
	infos := make([]*types.ConnectorProviderSetting, 0, len(settings))
	for _, setting := range settings {
		infos = append(infos, &types.ConnectorProviderSetting{
			Key:      setting.Key,
			Label:    setting.Label,
			Required: setting.Required,
		})
	}

	return infos
}
