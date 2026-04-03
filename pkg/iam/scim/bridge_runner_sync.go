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

package scim

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/scim/bridge"
	scimclient "go.probo.inc/probo/pkg/iam/scim/bridge/client"
	"go.probo.inc/probo/pkg/iam/scim/bridge/provider"
	"go.probo.inc/probo/pkg/iam/scim/bridge/provider/googleworkspace"
)

func (r *BridgeRunner) executeSync(
	ctx context.Context,
	scimBridge *coredata.SCIMBridge,
	scope coredata.Scoper,
	logger *log.Logger,
) (stats SyncStats, duration time.Duration, dbConnector *coredata.Connector, err error) {
	start := time.Now()

	var (
		idp   provider.Provider
		token string
	)

	err = r.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var err error
			idp, token, dbConnector, err = r.prepareSync(
				ctx,
				tx,
				scimBridge,
				scope,
				logger,
			)

			if err != nil {
				return fmt.Errorf("cannot prepare sync: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		duration = time.Since(start)
		return SyncStats{}, duration, nil, err
	}

	scimClient := r.createSCIMClient(logger, token)
	syncer := bridge.NewBridge(
		idp,
		scimClient,
		bridge.WithExcludedUserNames(scimBridge.ExcludedUserNames),
	)

	created, updated, deleted, deactivated, skipped, syncErr := syncer.Run(ctx)
	duration = time.Since(start)

	if syncErr != nil {
		return SyncStats{}, duration, dbConnector, fmt.Errorf("sync failed: %w", syncErr)
	}

	stats = SyncStats{
		Created:     created,
		Updated:     updated,
		Deleted:     deleted,
		Deactivated: deactivated,
		Skipped:     skipped,
	}

	return stats, duration, dbConnector, nil
}

func (r *BridgeRunner) prepareSync(
	ctx context.Context,
	tx pg.Tx,
	scimBridge *coredata.SCIMBridge,
	scope coredata.Scoper,
	logger *log.Logger,
) (provider.Provider, string, *coredata.Connector, error) {
	if scimBridge.ConnectorID == nil {
		return nil, "", nil, fmt.Errorf("bridge has no connector configured")
	}

	dbConnector := &coredata.Connector{}
	if err := dbConnector.LoadByID(ctx, tx, scope, *scimBridge.ConnectorID, r.encryptionKey); err != nil {
		return nil, "", nil, fmt.Errorf("cannot load connector: %w", err)
	}

	idp, err := r.createProvider(ctx, logger, scimBridge.Type, dbConnector, scimBridge.ExcludedUserNames)
	if err != nil {
		return nil, "", nil, fmt.Errorf("cannot create provider: %w", err)
	}

	var scimConfig coredata.SCIMConfiguration
	if err := scimConfig.LoadByID(ctx, tx, scope, scimBridge.ScimConfigurationID); err != nil {
		return nil, "", nil, fmt.Errorf("cannot load SCIM configuration: %w", err)
	}

	token, err := GenerateToken()
	if err != nil {
		return nil, "", nil, fmt.Errorf("cannot generate SCIM token: %w", err)
	}

	scimConfig.HashedToken = HashToken(token)
	scimConfig.UpdatedAt = time.Now()
	if err := scimConfig.Update(ctx, tx, scope); err != nil {
		return nil, "", nil, fmt.Errorf("cannot update SCIM configuration token: %w", err)
	}

	return idp, token, dbConnector, nil
}

func (r *BridgeRunner) createSCIMClient(logger *log.Logger, token string) *scimclient.Client {
	scimEndpoint := r.cfg.BaseURL.WithPath("/api/connect/v1/scim/2.0").MustString()
	httpClient := httpclient.DefaultPooledClient(
		httpclient.WithLogger(logger),
		httpclient.WithTracerProvider(r.tp),
		httpclient.WithRegisterer(r.registerer),
	)

	return scimclient.NewClient(httpClient, scimEndpoint, token)
}

func (r *BridgeRunner) createProvider(
	ctx context.Context,
	logger *log.Logger,
	bridgeType coredata.SCIMBridgeType,
	dbConnector *coredata.Connector,
	excludedUserNames []string,
) (provider.Provider, error) {
	switch bridgeType {
	case coredata.SCIMBridgeTypeGoogleWorkspace:
		return r.createGoogleWorkspaceProvider(ctx, logger, dbConnector, excludedUserNames)
	default:
		return nil, fmt.Errorf("unsupported bridge type: %s", bridgeType)
	}
}

func (r *BridgeRunner) createGoogleWorkspaceProvider(
	ctx context.Context,
	logger *log.Logger,
	dbConnector *coredata.Connector,
	excludedUserNames []string,
) (provider.Provider, error) {
	if dbConnector.Connection == nil {
		return nil, fmt.Errorf("connector has no connection configured")
	}

	oauth2Conn, ok := dbConnector.Connection.(*connector.OAuth2Connection)
	if !ok {
		return nil, fmt.Errorf("connector is not an OAuth2 connection")
	}

	httpClientOpts := []httpclient.Option{
		httpclient.WithLogger(logger),
		httpclient.WithTracerProvider(r.tp),
		httpclient.WithRegisterer(r.registerer),
	}

	providerName := dbConnector.Provider.String()
	refreshCfg := r.connectorRegistry.GetOAuth2RefreshConfig(providerName)
	if refreshCfg == nil {
		logger.WarnCtx(
			ctx,
			"no OAuth2 refresh config found, using static token",
			log.String("connector_id", dbConnector.ID.String()),
			log.String("connector_provider", providerName),
		)
		httpClient, err := oauth2Conn.ClientWithOptions(ctx, httpClientOpts...)
		if err != nil {
			return nil, fmt.Errorf("cannot create HTTP client: %w", err)
		}
		return googleworkspace.New(httpClient, excludedUserNames), nil
	}

	httpClient, err := oauth2Conn.RefreshableClient(ctx, *refreshCfg, httpClientOpts...)
	if err != nil {
		return nil, fmt.Errorf("cannot create refreshable HTTP client: %w", err)
	}

	return googleworkspace.New(httpClient, excludedUserNames), nil
}
