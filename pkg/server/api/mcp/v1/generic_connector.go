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
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/probo"
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

func (r *Resolver) createAPIKeyConnector(
	ctx context.Context,
	input *types.CreateAPIKeyConnectorInput,
) (*mcp.CallToolResult, types.CreateAPIKeyConnectorOutput, error) {
	scope, err := r.Authorize(ctx, input.OrganizationID, probo.ActionConnectorCreate)
	if err != nil {
		return nil, types.CreateAPIKeyConnectorOutput{}, err
	}

	registry := r.proboSvc.Connectors.ProviderRegistry()
	if registry == nil {
		return nil, types.CreateAPIKeyConnectorOutput{}, fmt.Errorf("internal server error")
	}

	reg, err := genericAPIKeyProvider(registry, input.Provider)
	if err != nil {
		return nil, types.CreateAPIKeyConnectorOutput{}, err
	}

	apiKey := strings.TrimSpace(input.APIKey)
	if apiKey == "" {
		return nil, types.CreateAPIKeyConnectorOutput{}, fmt.Errorf("api_key is required")
	}

	if err := registry.ValidateAPIKey(reg.Provider, apiKey); err != nil {
		return nil, types.CreateAPIKeyConnectorOutput{}, err
	}

	conn := registry.NewAPIKeyConnection(reg.Provider, apiKey)

	raw, err := r.apiKeySettings(ctx, registry, reg, apiKey)
	if err != nil {
		return nil, types.CreateAPIKeyConnectorOutput{}, err
	}

	if rejected := r.accessReview.CheckNewAPIKeyConnector(ctx, reg.Provider, conn, raw); rejected != nil {
		return nil, types.CreateAPIKeyConnectorOutput{}, fmt.Errorf("%s", rejected.Message)
	}

	cnnctr, err := r.proboSvc.Connectors.Create(ctx, scope, probo.CreateConnectorRequest{
		OrganizationID: input.OrganizationID,
		Name:           input.Name,
		Provider:       reg.Provider,
		Protocol:       coredata.ConnectorProtocolAPIKey,
		Connection:     conn,
		RawSettings:    raw,
	})
	if err != nil {
		return nil, types.CreateAPIKeyConnectorOutput{}, connectorMCPWriteError(ctx, r.logger, "cannot create API key connector", err)
	}

	r.accessReview.SelectSoleOrganization(ctx, scope, cnnctr.ID)

	return nil, types.CreateAPIKeyConnectorOutput{
		Connector: types.NewConnector(cnnctr, r.connectorConnectionStatus(ctx, scope, cnnctr.ID)),
	}, nil
}

func (r *Resolver) createClientCredentialsConnector(
	ctx context.Context,
	input *types.CreateClientCredentialsConnectorInput,
) (*mcp.CallToolResult, types.CreateClientCredentialsConnectorOutput, error) {
	scope, err := r.Authorize(ctx, input.OrganizationID, probo.ActionConnectorCreate)
	if err != nil {
		return nil, types.CreateClientCredentialsConnectorOutput{}, err
	}

	registry := r.proboSvc.Connectors.ProviderRegistry()
	if registry == nil {
		return nil, types.CreateClientCredentialsConnectorOutput{}, fmt.Errorf("internal server error")
	}

	reg, err := genericClientCredentialsProvider(registry, input.Provider)
	if err != nil {
		return nil, types.CreateClientCredentialsConnectorOutput{}, err
	}

	tokenURL, err := clientCredentialsTokenURL(reg, input.TokenURL)
	if err != nil {
		return nil, types.CreateClientCredentialsConnectorOutput{}, err
	}

	cnnctr, err := r.proboSvc.Connectors.Create(ctx, scope, probo.CreateConnectorRequest{
		OrganizationID: input.OrganizationID,
		Name:           input.Name,
		Provider:       reg.Provider,
		Protocol:       coredata.ConnectorProtocolOAuth2,
		Connection: &connector.OAuth2Connection{
			GrantType:    connector.OAuth2GrantTypeClientCredentials,
			ClientID:     input.ClientID,
			ClientSecret: input.ClientSecret,
			TokenURL:     tokenURL,
			Scope:        clientCredentialsScope(reg, input.Scope),
		},
	})
	if err != nil {
		return nil, types.CreateClientCredentialsConnectorOutput{}, connectorMCPWriteError(
			ctx,
			r.logger,
			"cannot create client credentials connector",
			err,
		)
	}

	r.accessReview.SelectSoleOrganization(ctx, scope, cnnctr.ID)

	return nil, types.CreateClientCredentialsConnectorOutput{
		Connector: types.NewConnector(cnnctr, r.connectorConnectionStatus(ctx, scope, cnnctr.ID)),
	}, nil
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

func genericAPIKeyProvider(registry *provider.Registry, name string) (*provider.Registration, error) {
	reg, ok := registry.Get(coredata.ConnectorProvider(name))
	if !ok {
		return nil, fmt.Errorf("unknown connector provider")
	}

	switch {
	case reg.SupportsInstall():
		return nil, fmt.Errorf("%s is connected through its install flow, not with an API key", reg.DisplayName)
	case reg.IsManagedAPIKey() || !reg.SupportsAPIKey():
		return nil, fmt.Errorf("%s is connected in the console, not with an API key", reg.DisplayName)
	case len(reg.APIKeyExtraSettings()) > 0:
		return nil, fmt.Errorf("%s needs an extra setting; connect it in the console", reg.DisplayName)
	}

	return reg, nil
}

func genericClientCredentialsProvider(registry *provider.Registry, name string) (*provider.Registration, error) {
	reg, ok := registry.Get(coredata.ConnectorProvider(name))
	if !ok {
		return nil, fmt.Errorf("unknown connector provider")
	}

	switch {
	case !reg.SupportsClientCredentials():
		return nil, fmt.Errorf("%s is connected in the console, not with client credentials", reg.DisplayName)
	case len(reg.ClientCredentialsExtraSettings()) > 0:
		return nil, fmt.Errorf("%s needs an extra setting; connect it in the console", reg.DisplayName)
	}

	return reg, nil
}

func (r *Resolver) apiKeySettings(
	ctx context.Context,
	registry *provider.Registry,
	reg *provider.Registration,
	apiKey string,
) (json.RawMessage, error) {
	if reg.Provider != coredata.ConnectorProviderTally {
		return nil, nil
	}

	conn := registry.NewAPIKeyConnection(reg.Provider, apiKey)

	httpClient, err := conn.Client(ctx)
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot build tally validation client", log.Error(err))

		return nil, fmt.Errorf("internal server error")
	}

	user, err := drivers.GetTallyCurrentUser(ctx, httpClient, reg.Endpoints.APIBase)
	switch {
	case errors.Is(err, drivers.ErrTallyUnauthorized):
		return nil, fmt.Errorf("tally rejected the API key: create a key under Settings > API keys and try again")
	case err != nil:
		r.logger.ErrorCtx(ctx, "cannot fetch tally current user", log.Error(err))

		return nil, fmt.Errorf("internal server error")
	}

	if user.OrganizationID == "" {
		r.logger.ErrorCtx(ctx, "tally current user has no organization id")

		return nil, fmt.Errorf("internal server error")
	}

	raw, err := json.Marshal(&coredata.TallyConnectorSettings{OrganizationID: user.OrganizationID})
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot marshal tally connector settings", log.Error(err))

		return nil, fmt.Errorf("internal server error")
	}

	return raw, nil
}

func clientCredentialsScope(reg *provider.Registration, supplied *string) string {
	if reg.OAuth2 != nil {
		if scope := connector.FormatScopeString(reg.OAuth2.Scopes); scope != "" {
			return scope
		}
	}

	if supplied == nil {
		return ""
	}

	return strings.TrimSpace(*supplied)
}

func clientCredentialsTokenURL(reg *provider.Registration, supplied *string) (string, error) {
	if pinned := reg.Endpoints.Token; pinned != "" {
		return pinned, nil
	}

	if supplied == nil || strings.TrimSpace(*supplied) == "" {
		return "", fmt.Errorf("token_url is required for this provider")
	}

	tokenURL := strings.TrimSpace(*supplied)

	parsed, err := url.Parse(tokenURL)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
		return "", fmt.Errorf("token_url must be an absolute https URL")
	}

	return tokenURL, nil
}
