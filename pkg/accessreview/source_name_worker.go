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

package accessreview

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/accessreview/drivers"
)

// SourceNameWorker polls for access sources that have a connector but no
// synced name, resolves the provider instance name, and updates the source.
type SourceNameWorker struct {
	pg            *pg.Client
	encryptionKey cipher.EncryptionKey
	logger        *log.Logger
	interval      time.Duration
}

func NewSourceNameWorker(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	logger *log.Logger,
) *SourceNameWorker {
	return &SourceNameWorker{
		pg:            pgClient,
		encryptionKey: encryptionKey,
		logger:        logger,
		interval:      10 * time.Second,
	}
}

func (w *SourceNameWorker) Run(ctx context.Context) error {
	w.logger.InfoCtx(ctx, "source name worker started",
		log.String("interval", w.interval.String()),
	)

LOOP:
	select {
	case <-ctx.Done():
		w.logger.InfoCtx(context.WithoutCancel(ctx), "source name worker stopping")
		return ctx.Err()
	case <-time.After(w.interval):
		nonCancelableCtx := context.WithoutCancel(ctx)
		for {
			if err := w.processNext(nonCancelableCtx); err != nil {
				if !errors.Is(err, coredata.ErrNoAccessSourceNameSyncAvailable) {
					w.logger.ErrorCtx(nonCancelableCtx, "cannot sync source name", log.Error(err))
				}
				break
			}
		}
		goto LOOP
	}
}

func (w *SourceNameWorker) processNext(ctx context.Context) error {
	var (
		source   coredata.AccessSource
		tenantID gid.TenantID
	)

	// Claim the next source that needs name syncing.
	err := w.pg.WithTx(ctx, func(tx pg.Conn) error {
		var err error
		tenantID, err = source.LoadNextUnsyncedNameForUpdateSkipLocked(ctx, tx)
		if err != nil {
			return err
		}

		// Mark as synced immediately to prevent other workers from
		// picking the same row. The name will be updated in a
		// subsequent step if resolution succeeds.
		now := time.Now()
		source.NameSyncedAt = new(now)
		source.UpdatedAt = now

		scope := coredata.NewScope(tenantID)
		return source.Update(ctx, tx, scope)
	})
	if err != nil {
		return err
	}

	w.logger.InfoCtx(ctx, "syncing source name",
		log.String("source_id", source.ID.String()),
		log.String("current_name", source.Name),
	)

	// Load the connector (with decrypted connection) to resolve both the
	// provider display name and the instance name in a single pass.
	var (
		dbConnector coredata.Connector
		resolver    drivers.NameResolver
	)

	err = w.pg.WithConn(ctx, func(conn pg.Conn) error {
		scope := coredata.NewScope(tenantID)
		if source.ConnectorID == nil {
			return fmt.Errorf("source %s has no connector", source.ID)
		}

		if err := dbConnector.LoadByID(ctx, conn, scope, *source.ConnectorID, w.encryptionKey); err != nil {
			return fmt.Errorf("cannot load connector %s: %w", *source.ConnectorID, err)
		}

		httpClient, err := dbConnector.Connection.Client(ctx)
		if err != nil {
			return fmt.Errorf("cannot create HTTP client for connector: %w", err)
		}

		resolver = w.buildResolver(&dbConnector, httpClient)
		return nil
	})
	if err != nil {
		w.logger.ErrorCtx(ctx, "cannot load connector for source name sync",
			log.String("source_id", source.ID.String()),
			log.Error(err),
		)
		return nil
	}

	if resolver == nil {
		w.logger.InfoCtx(ctx, "no name resolver for provider, keeping generic name",
			log.String("source_id", source.ID.String()),
			log.String("provider", dbConnector.Provider.String()),
		)
		return nil
	}

	resolveCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	instanceName, resolveErr := resolver.ResolveInstanceName(resolveCtx)
	if resolveErr != nil {
		w.logger.ErrorCtx(ctx, "cannot resolve instance name, keeping generic name",
			log.String("source_id", source.ID.String()),
			log.String("provider", dbConnector.Provider.String()),
			log.Error(resolveErr),
		)
		return nil
	}

	if instanceName == "" {
		w.logger.InfoCtx(ctx, "instance name is empty, keeping generic name",
			log.String("source_id", source.ID.String()),
			log.String("provider", dbConnector.Provider.String()),
		)
		return nil
	}

	displayName := drivers.ProviderDisplayName(dbConnector.Provider)
	newName := displayName + " " + instanceName

	w.logger.InfoCtx(ctx, "resolved source name",
		log.String("source_id", source.ID.String()),
		log.String("old_name", source.Name),
		log.String("new_name", newName),
	)

	return w.pg.WithConn(ctx, func(conn pg.Conn) error {
		scope := coredata.NewScope(tenantID)
		now := time.Now()

		source.Name = newName
		source.UpdatedAt = now
		return source.Update(ctx, conn, scope)
	})
}

func (w *SourceNameWorker) buildResolver(
	dbConnector *coredata.Connector,
	httpClient *http.Client,
) drivers.NameResolver {
	switch dbConnector.Provider {
	case coredata.ConnectorProviderSlack:
		return drivers.NewSlackNameResolver(httpClient)
	case coredata.ConnectorProviderGoogleWorkspace:
		return drivers.NewGoogleWorkspaceNameResolver(httpClient)
	case coredata.ConnectorProviderLinear:
		return drivers.NewLinearNameResolver(httpClient)
	case coredata.ConnectorProviderCloudflare:
		return drivers.NewCloudflareNameResolver(httpClient)
	case coredata.ConnectorProviderBrex:
		return drivers.NewBrexNameResolver(httpClient)
	case coredata.ConnectorProviderTally:
		tallySettings, err := dbConnector.TallySettings()
		if err != nil {
			w.logger.Error("cannot read tally connector settings", log.Error(err))
			return nil
		}
		return drivers.NewTallyNameResolver(httpClient, tallySettings.OrganizationID)
	case coredata.ConnectorProviderHubSpot:
		return drivers.NewHubSpotNameResolver(httpClient)
	case coredata.ConnectorProviderDocuSign:
		return drivers.NewDocuSignNameResolver(httpClient)
	case coredata.ConnectorProviderFigma:
		return drivers.NewFigmaNameResolver(httpClient)
	case coredata.ConnectorProviderNotion:
		return drivers.NewNotionNameResolver(httpClient)
	case coredata.ConnectorProviderOpenAI:
		return drivers.NewOpenAINameResolver(httpClient)
	case coredata.ConnectorProviderOnePassword:
		return drivers.NewOnePasswordNameResolver()
	default:
		return nil
	}
}
