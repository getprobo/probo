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
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
)

// SourceNameWorker polls for access sources that have a connector but no
// synced name, resolves the provider instance name, and updates the source.
type SourceNameWorker struct {
	pg                *pg.Client
	encryptionKey     cipher.EncryptionKey
	connectorRegistry *connector.ConnectorRegistry
	logger            *log.Logger
	interval          time.Duration
}

func NewSourceNameWorker(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	connectorRegistry *connector.ConnectorRegistry,
	logger *log.Logger,
) *SourceNameWorker {
	return &SourceNameWorker{
		pg:                pgClient,
		encryptionKey:     encryptionKey,
		connectorRegistry: connectorRegistry,
		logger:            logger,
		interval:          10 * time.Second,
	}
}

func (w *SourceNameWorker) Run(ctx context.Context) error {
	w.logger.InfoCtx(ctx, "source name worker started",
		log.String("interval", w.interval.String()),
	)

	for {
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
		}
	}
}

func (w *SourceNameWorker) processNext(ctx context.Context) error {
	var source coredata.AccessSource

	err := w.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return source.LoadNextUnsyncedNameForUpdateSkipLocked(ctx, tx)
		},
	)
	if err != nil {
		return err
	}

	w.logger.InfoCtx(ctx, "syncing source name",
		log.String("source_id", source.ID.String()),
		log.String("current_name", source.Name),
	)

	var (
		dbConnector coredata.Connector
		resolver    drivers.NameResolver
	)

	err = w.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(source.ID)
			if source.ConnectorID == nil {
				return fmt.Errorf("source %s has no connector", source.ID)
			}

			if err := dbConnector.LoadByID(ctx, tx, scope, *source.ConnectorID, w.encryptionKey); err != nil {
				return fmt.Errorf("cannot load connector %s: %w", *source.ConnectorID, err)
			}

			var tokenBefore string
			if oauth2Conn, ok := dbConnector.Connection.(*connector.OAuth2Connection); ok {
				tokenBefore = oauth2Conn.AccessToken
			}

			httpClient, err := w.connectorHTTPClient(ctx, &dbConnector)
			if err != nil {
				return fmt.Errorf("cannot create HTTP client for connector: %w", err)
			}

			if oauth2Conn, ok := dbConnector.Connection.(*connector.OAuth2Connection); ok {
				if oauth2Conn.AccessToken != tokenBefore {
					dbConnector.UpdatedAt = time.Now()
					if err := dbConnector.Update(ctx, tx, scope, w.encryptionKey); err != nil {
						return fmt.Errorf("cannot persist refreshed token for connector %s: %w", *source.ConnectorID, err)
					}
				}
			}

			resolver = w.buildResolver(&dbConnector, httpClient)
			return nil
		},
	)
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
		return w.markNameSynced(ctx, &source)
	}

	resolveCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	instanceName, err := resolver.ResolveInstanceName(resolveCtx)
	if err != nil {
		w.logger.ErrorCtx(ctx, "cannot resolve instance name",
			log.String("source_id", source.ID.String()),
			log.String("provider", dbConnector.Provider.String()),
			log.Error(err),
		)
		return fmt.Errorf("cannot resolve instance name for source %s: %w", source.ID, err)
	}

	if instanceName == "" {
		w.logger.InfoCtx(ctx, "instance name is empty, keeping generic name",
			log.String("source_id", source.ID.String()),
			log.String("provider", dbConnector.Provider.String()),
		)
		return w.markNameSynced(ctx, &source)
	}

	displayName := drivers.ProviderDisplayName(dbConnector.Provider)
	newName := displayName + " " + instanceName

	w.logger.InfoCtx(ctx, "resolved source name",
		log.String("source_id", source.ID.String()),
		log.String("old_name", source.Name),
		log.String("new_name", newName),
	)

	source.Name = newName
	return w.markNameSynced(ctx, &source)
}

func (w *SourceNameWorker) markNameSynced(
	ctx context.Context,
	source *coredata.AccessSource,
) error {
	return w.pg.WithTx(
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
func (w *SourceNameWorker) connectorHTTPClient(
	ctx context.Context,
	dbConnector *coredata.Connector,
) (*http.Client, error) {
	oauth2Conn, ok := dbConnector.Connection.(*connector.OAuth2Connection)
	if !ok {
		return dbConnector.Connection.Client(ctx)
	}

	if w.connectorRegistry != nil {
		refreshCfg := w.connectorRegistry.GetOAuth2RefreshConfig(string(dbConnector.Provider))
		if refreshCfg != nil {
			return oauth2Conn.RefreshableClient(ctx, *refreshCfg)
		}
	}

	return oauth2Conn.Client(ctx)
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
	case coredata.ConnectorProviderOpenAI:
		return drivers.NewOpenAINameResolver(httpClient)
	case coredata.ConnectorProviderSentry:
		sentrySettings, err := dbConnector.SentrySettings()
		if err != nil {
			w.logger.Error("cannot read sentry connector settings", log.Error(err))
			return nil
		}
		return drivers.NewSentryNameResolver(httpClient, sentrySettings.OrganizationSlug)
	case coredata.ConnectorProviderGitHub:
		githubSettings, err := dbConnector.GitHubSettings()
		if err != nil {
			w.logger.Error("cannot read github connector settings", log.Error(err))
			return nil
		}
		return drivers.NewGitHubNameResolver(httpClient, githubSettings.Organization)
	case coredata.ConnectorProviderSupabase:
		supabaseSettings, err := dbConnector.SupabaseSettings()
		if err != nil {
			w.logger.Error("cannot read supabase connector settings", log.Error(err))
			return nil
		}
		return drivers.NewSupabaseNameResolver(supabaseSettings.OrganizationSlug)
	case coredata.ConnectorProviderIntercom:
		return drivers.NewIntercomNameResolver(httpClient)
	case coredata.ConnectorProviderResend:
		return drivers.NewResendNameResolver()
	default:
		return nil
	}
}
