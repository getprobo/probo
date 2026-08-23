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
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

// nameResolutionTimeout bounds the call that asks a provider what its instance
// is called.
const nameResolutionTimeout = 10 * time.Second

// sourceNameHandler polls for access sources that have a connector but no
// synced name, resolves the provider instance name, and updates the source.
type sourceNameHandler struct {
	pg               *pg.Client
	connectors       *provider.Opener
	providerRegistry *provider.Registry
	sources          *drivers.Registry
	logger           *log.Logger
}

func NewSourceNameWorker(
	pgClient *pg.Client,
	connectors *provider.Opener,
	sources *drivers.Registry,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.AccessReviewSource] {
	h := &sourceNameHandler{
		pg:               pgClient,
		connectors:       connectors,
		providerRegistry: connectors.Providers(),
		sources:          sources,
		logger:           logger,
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

	if source.ConnectorID == nil {
		return fmt.Errorf("source %s has no connector", source.ID)
	}

	var (
		connectorProvider coredata.ConnectorProvider
		instanceName      string
		// resolutionFailed separates a resolver that ran and failed from a
		// resolver that could never be built, because only the former is worth
		// retrying — and then only when the failure is transient.
		resolutionFailed bool
	)

	err := h.connectors.Use(
		ctx,
		coredata.NewScopeFromObjectID(source.ID),
		*source.ConnectorID,
		func(ctx context.Context, handle *provider.Handle) error {
			connectorProvider = handle.Connector.Provider

			driver, err := h.sources.New(ctx, handle, h.logger)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(ctx, nameResolutionTimeout)
			defer cancel()

			resolved, err := drivers.InstanceName(ctx, driver)
			if err != nil {
				resolutionFailed = true

				return err
			}

			instanceName = resolved

			return nil
		},
	)
	if err != nil {
		// A resolver that could not be built (missing connector, undecryptable
		// credential, an eager refresh on a revoked token) and a permanent
		// resolution failure (auth, bad request) share one remedy: keep the
		// generic name and mark the source synced. Returning the error would
		// leave name_synced_at NULL, and an unsynced row is re-claimed every
		// poll with no backoff — a single unauthorized source once produced
		// millions of error logs. A reconnect clears it.
		if !resolutionFailed || errors.Is(err, drivers.ErrTerminalNameResolution) {
			h.logger.WarnCtx(
				ctx,
				"cannot resolve source name, keeping generic name",
				log.String("source_id", source.ID.String()),
				log.String("provider", connectorProvider.String()),
				log.Error(err),
			)

			return h.markNameSynced(ctx, &source)
		}

		h.logger.WarnCtx(
			ctx,
			"cannot resolve instance name",
			log.String("source_id", source.ID.String()),
			log.String("provider", connectorProvider.String()),
			log.Error(err),
		)

		return fmt.Errorf("cannot resolve instance name for source %s: %w", source.ID, err)
	}

	if instanceName == "" {
		h.logger.InfoCtx(
			ctx,
			"no instance name for provider, keeping generic name",
			log.String("source_id", source.ID.String()),
			log.String("provider", connectorProvider.String()),
		)

		return h.markNameSynced(ctx, &source)
	}

	displayName := h.providerRegistry.ProviderDisplayName(connectorProvider)
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
