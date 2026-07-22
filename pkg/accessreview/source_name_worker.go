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

package accessreview

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
)

// sourceNameHandler polls for access sources that have a connector but no
// synced name, resolves the provider instance name, and updates the source.
type sourceNameHandler struct {
	pg                *pg.Client
	encryptionKey     cipher.EncryptionKey
	connectorRegistry *connector.ConnectorRegistry
	providerRegistry  *provider.Registry
	logger            *log.Logger
}

func NewSourceNameWorker(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	connectorRegistry *connector.ConnectorRegistry,
	providerRegistry *provider.Registry,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.AccessReviewSource] {
	h := &sourceNameHandler{
		pg:                pgClient,
		encryptionKey:     encryptionKey,
		connectorRegistry: connectorRegistry,
		providerRegistry:  providerRegistry,
		logger:            logger,
	}

	defaultOpts := []worker.Option{
		worker.WithInterval(10 * time.Second),
		worker.WithMaxConcurrency(1),
	}

	return worker.New(
		"source-name-worker",
		h,
		logger,
		append(defaultOpts, opts...)...,
	)
}

func (h *sourceNameHandler) Claim(ctx context.Context) (coredata.AccessReviewSource, error) {
	var source coredata.AccessReviewSource

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return source.LoadNextUnsyncedNameForUpdateSkipLocked(ctx, tx)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrNoAccessReviewSourceNameSyncAvailable) {
			return coredata.AccessReviewSource{}, worker.ErrNoTask
		}

		return coredata.AccessReviewSource{}, err
	}

	return source, nil
}

func (h *sourceNameHandler) Process(ctx context.Context, source coredata.AccessReviewSource) error {
	h.logger.InfoCtx(
		ctx,
		"syncing source name",
		log.String("source_id", source.ID.String()),
		log.String("current_name", source.Name),
	)

	var (
		dbConnector coredata.Connector
		resolver    drivers.NameResolver
	)

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(source.ID)
			if source.ConnectorID == nil {
				return fmt.Errorf("source %s has no connector", source.ID)
			}

			if err := dbConnector.LoadByID(ctx, tx, scope, *source.ConnectorID, h.encryptionKey); err != nil {
				return fmt.Errorf("cannot load connector %s: %w", *source.ConnectorID, err)
			}

			var tokenBefore string
			if oauth2Conn, ok := dbConnector.Connection.(*connector.OAuth2Connection); ok {
				tokenBefore = oauth2Conn.AccessToken
			}

			httpClient, err := h.connectorHTTPClient(ctx, &dbConnector)
			if err != nil {
				return fmt.Errorf("cannot create HTTP client for connector: %w", err)
			}

			if oauth2Conn, ok := dbConnector.Connection.(*connector.OAuth2Connection); ok {
				if oauth2Conn.AccessToken != tokenBefore {
					dbConnector.UpdatedAt = time.Now()
					if err := dbConnector.Update(ctx, tx, scope, h.encryptionKey); err != nil {
						return fmt.Errorf("cannot persist refreshed token for connector %s: %w", *source.ConnectorID, err)
					}
				}
			}

			resolver = h.buildResolver(ctx, &dbConnector, httpClient)

			return nil
		},
	)
	if err != nil {
		// Resolver setup failed (missing connector, undecryptable credential,
		// or an eager refresh on a revoked token). Mark the source synced
		// rather than returning nil: an unsynced row is re-claimed every poll
		// with no backoff and hot-loops the vendor. A reconnect clears it.
		h.logger.WarnCtx(
			ctx,
			"cannot set up name resolver, keeping generic name",
			log.String("source_id", source.ID.String()),
			log.Error(err),
		)

		return h.markNameSynced(ctx, &source)
	}

	if resolver == nil {
		h.logger.InfoCtx(
			ctx,
			"no name resolver for provider, keeping generic name",
			log.String("source_id", source.ID.String()),
			log.String("provider", dbConnector.Provider.String()),
		)

		return h.markNameSynced(ctx, &source)
	}

	resolveCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	instanceName, err := resolver.ResolveInstanceName(resolveCtx)
	if err != nil {
		// A permanent failure (auth/bad-request) cannot be fixed by
		// retrying: keep the generic name and mark the source synced so the
		// worker stops re-claiming it every poll. Returning the error here
		// would leave name_synced_at NULL and re-enqueue the source forever
		// (a single unauthorized source produced millions of error logs).
		if errors.Is(err, drivers.ErrTerminalNameResolution) {
			h.logger.WarnCtx(
				ctx,
				"permanent name resolution failure, keeping generic name",
				log.String("source_id", source.ID.String()),
				log.String("provider", dbConnector.Provider.String()),
				log.Error(err),
			)

			return h.markNameSynced(ctx, &source)
		}

		h.logger.WarnCtx(
			ctx,
			"cannot resolve instance name",
			log.String("source_id", source.ID.String()),
			log.String("provider", dbConnector.Provider.String()),
			log.Error(err),
		)

		return fmt.Errorf("cannot resolve instance name for source %s: %w", source.ID, err)
	}

	if instanceName == "" {
		h.logger.InfoCtx(
			ctx,
			"instance name is empty, keeping generic name",
			log.String("source_id", source.ID.String()),
			log.String("provider", dbConnector.Provider.String()),
		)

		return h.markNameSynced(ctx, &source)
	}

	displayName := h.providerRegistry.ProviderDisplayName(dbConnector.Provider)
	newName := displayName + " " + instanceName

	h.logger.InfoCtx(
		ctx,
		"resolved source name",
		log.String("source_id", source.ID.String()),
		log.String("old_name", source.Name),
		log.String("new_name", newName),
	)

	source.Name = newName

	return h.markNameSynced(ctx, &source)
}

func (h *sourceNameHandler) markNameSynced(
	ctx context.Context,
	source *coredata.AccessReviewSource,
) error {
	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(source.ID)
			now := time.Now()

			source.NameSyncedAt = new(now)
			source.UpdatedAt = now

			if err := source.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update access source: %w", err)
			}

			return nil
		},
	)
}

// connectorHTTPClient returns an HTTP client for the given connector.
// For OAuth2 connections it uses RefreshableClient when a refresh config
// is registered for the provider, so that short-lived tokens are
// transparently refreshed.
func (h *sourceNameHandler) connectorHTTPClient(
	ctx context.Context,
	dbConnector *coredata.Connector,
) (*http.Client, error) {
	oauth2Conn, ok := dbConnector.Connection.(*connector.OAuth2Connection)
	if !ok {
		// Inject the Probo-held key for ManagedAPIKey providers (no-op
		// otherwise) before building the client.
		if err := h.providerRegistry.ApplyManagedAPIKey(dbConnector); err != nil {
			return nil, err
		}

		return dbConnector.Connection.Client(ctx)
	}

	if h.connectorRegistry != nil {
		refreshCfg := h.connectorRegistry.GetOAuth2RefreshConfig(string(dbConnector.Provider))
		if refreshCfg != nil {
			return oauth2Conn.RefreshableClient(ctx, *refreshCfg)
		}
	}

	return oauth2Conn.Client(ctx)
}

func (h *sourceNameHandler) buildResolver(
	ctx context.Context,
	dbConnector *coredata.Connector,
	httpClient *http.Client,
) drivers.NameResolver {
	reg, ok := h.providerRegistry.Get(dbConnector.Provider)
	if !ok || reg.NewNameResolver == nil {
		return nil
	}

	return reg.NewNameResolver(ctx, httpClient, dbConnector, h.logger)
}
