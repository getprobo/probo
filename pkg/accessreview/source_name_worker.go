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
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/identityfederation"
)

// sourceNameHandler polls for access sources that have a connector but no
// synced name, resolves the provider instance name, and updates the source.
type sourceNameHandler struct {
	pg                *pg.Client
	encryptionKey     cipher.EncryptionKey
	connectorRegistry *connector.Registry
	providerRegistry  *provider.Registry
	federation        *identityfederation.Issuer
	logger            *log.Logger
}

func NewSourceNameWorker(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	connectorRegistry *connector.Registry,
	providerRegistry *provider.Registry,
	federation *identityfederation.Issuer,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.AccessReviewSource] {
	h := &sourceNameHandler{
		pg:                pgClient,
		encryptionKey:     encryptionKey,
		connectorRegistry: connectorRegistry,
		providerRegistry:  providerRegistry,
		federation:        federation,
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

			// The connection decides which credential the resolver can be
			// built from, the same way ProbeConnector and resolveDriver
			// pick a path. A protocol without a factory lands in the
			// default arm rather than falling through to the other kind.
			switch conn := dbConnector.Connection.(type) {
			case *connector.WorkloadIdentityConnection:
				r, err := h.newCloudNameResolver(ctx, &dbConnector)
				if err != nil {
					return err
				}

				resolver = r

				return nil

			case connector.HTTPConnection:
				r, err := h.newHTTPNameResolver(ctx, tx, scope, &dbConnector, conn)
				if err != nil {
					return err
				}

				resolver = r

				return nil

			default:
				return fmt.Errorf(
					"cannot resolve source name: %s connector has an unsupported credential",
					dbConnector.Provider,
				)
			}
		},
	)
	if err != nil {
		// Resolver setup failed (missing connector, undecryptable credential,
		// or an eager refresh on a revoked token). Mark the source synced
		// rather than returning nil: an unsynced row is re-claimed every poll
		// with no backoff and hot-loops the vendor. A reconnect clears it.
		if ctx.Err() != nil {
			return err
		}

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
		log.String("connector_id", dbConnector.ID.String()),
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

func (h *sourceNameHandler) newCloudNameResolver(
	ctx context.Context,
	dbConnector *coredata.Connector,
) (drivers.NameResolver, error) {
	reg, ok := h.providerRegistry.Get(dbConnector.Provider)
	if !ok || reg.WorkloadIdentity == nil || reg.WorkloadIdentity.NewNameResolver == nil {
		return nil, nil
	}

	session, err := h.buildCloudSession(ctx, dbConnector)
	if err != nil {
		return nil, err
	}

	return reg.WorkloadIdentity.NewNameResolver(ctx, session, dbConnector, h.logger), nil
}

func (h *sourceNameHandler) newHTTPNameResolver(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	dbConnector *coredata.Connector,
	conn connector.HTTPConnection,
) (drivers.NameResolver, error) {
	reg, ok := h.providerRegistry.Get(dbConnector.Provider)
	if !ok || reg.NewNameResolver == nil {
		return nil, nil
	}

	var tokenBefore string
	if oauth2Conn, ok := conn.(*connector.OAuth2Connection); ok {
		tokenBefore = oauth2Conn.AccessToken
	}

	httpClient, err := buildHTTPClient(
		ctx,
		h.connectorRegistry,
		h.providerRegistry,
		dbConnector.Provider,
		conn,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create HTTP client for connector: %w", err)
	}

	if oauth2Conn, ok := conn.(*connector.OAuth2Connection); ok {
		if oauth2Conn.AccessToken != tokenBefore {
			dbConnector.UpdatedAt = time.Now()
			if err := dbConnector.Update(ctx, tx, scope, h.encryptionKey); err != nil {
				return nil, fmt.Errorf("cannot persist refreshed token for connector %s: %w", dbConnector.ID, err)
			}
		}
	}

	return reg.NewNameResolver(ctx, httpClient, dbConnector, h.logger, reg.Endpoints), nil
}

func (h *sourceNameHandler) buildCloudSession(
	ctx context.Context,
	dbConnector *coredata.Connector,
) (cloud.Session, error) {
	if h.federation == nil {
		return nil, fmt.Errorf(
			"cannot reach %s connector: identity federation is not configured in this deployment",
			dbConnector.Provider,
		)
	}

	reg, ok := h.providerRegistry.Get(dbConnector.Provider)
	if !ok || reg.WorkloadIdentity == nil {
		return nil, fmt.Errorf(
			"cannot reach %s connector: provider offers no workload identity path",
			dbConnector.Provider,
		)
	}

	session, err := reg.WorkloadIdentity.NewSession(ctx, h.federation, dbConnector)
	if err != nil {
		return nil, fmt.Errorf("cannot open cloud session for %s connector: %w", dbConnector.Provider, err)
	}

	return session, nil
}
