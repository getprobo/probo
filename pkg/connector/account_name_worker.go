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

package connector

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/identityfederation"
)

// accountNameHandler polls connector accounts that have no resolved name,
// resolves the provider instance name, and stores it on the account.
type accountNameHandler struct {
	pg                *pg.Client
	encryptionKey     cipher.EncryptionKey
	connectorRegistry *connectorprotocol.Registry
	providerRegistry  *provider.Registry
	federation        *identityfederation.Issuer
	logger            *log.Logger
}

func newAccountNameWorker(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	connectorRegistry *connectorprotocol.Registry,
	providerRegistry *provider.Registry,
	federation *identityfederation.Issuer,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.ConnectorAccount] {
	h := &accountNameHandler{
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
		"account-name-worker",
		h,
		logger,
		append(defaultOpts, opts...)...,
	)
}

// Name resolution is best-effort display metadata, so its retry budget is
// small: a briefly unreachable provider gets a few spaced retries, and
// anything still failing keeps the placeholder name until a reconnect resets
// the budget. See LoadNextUnsyncedNameForUpdateSkipLocked for why the budget,
// not the error taxonomy, is what bounds the claim.
const (
	maxNameSyncAttempts = 5
	nameSyncBaseBackoff = time.Minute
	nameSyncMaxBackoff  = time.Hour
)

// nameSyncBackoff returns how long to hold an account out of the queue after
// its attempt-th failure, doubling from nameSyncBaseBackoff up to the cap.
func nameSyncBackoff(attempt int) time.Duration {
	backoff := nameSyncBaseBackoff << max(attempt-1, 0)
	if backoff > nameSyncMaxBackoff || backoff <= 0 {
		return nameSyncMaxBackoff
	}

	return backoff
}

func (h *accountNameHandler) Claim(ctx context.Context) (coredata.ConnectorAccount, error) {
	var account coredata.ConnectorAccount

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := account.LoadNextUnsyncedNameForUpdateSkipLocked(ctx, tx); err != nil {
				return fmt.Errorf("cannot claim next unsynced account name: %w", err)
			}

			// Charging the attempt here, rather than after Process, is what
			// bounds the claim: nothing else transitions the predicate.
			return account.RecordNameSyncAttempt(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(account.ID),
				nameSyncBackoff(account.NameSyncAttempts+1),
			)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrNoConnectorAccountNameSyncAvailable) {
			return coredata.ConnectorAccount{}, worker.ErrNoTask
		}

		return coredata.ConnectorAccount{}, err
	}

	return account, nil
}

func (h *accountNameHandler) Process(ctx context.Context, account coredata.ConnectorAccount) error {
	logger := h.logger.With(log.String("connector_account_id", account.ID.String()))

	logger.DebugCtx(ctx, "syncing account name")

	var (
		dbConnector coredata.Connector
		resolver    drivers.NameResolver
	)

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(account.ID)

			if err := dbConnector.LoadByID(ctx, tx, scope, account.ConnectorID, h.encryptionKey); err != nil {
				return fmt.Errorf("cannot load connector %s: %w", account.ConnectorID, err)
			}

			// The connection decides which credential the resolver can be
			// built from. A protocol without a factory lands in the default
			// arm rather than falling through to the other kind.
			switch dbConnector.Connection.(type) {
			case *connectorprotocol.WorkloadIdentityConnection:
				r, err := h.newCloudNameResolver(ctx, &account, &dbConnector)
				if err != nil {
					return err
				}

				resolver = r

				return nil

			case connectorprotocol.HTTPConnection:
				r, err := h.newHTTPNameResolver(ctx, tx, scope, &dbConnector)
				if err != nil {
					return err
				}

				resolver = r

				return nil

			default:
				return fmt.Errorf(
					"cannot resolve account name: %s connector has an unsupported credential",
					dbConnector.Provider,
				)
			}
		},
	)
	if err != nil {
		// Resolver setup failed (missing connector, undecryptable credential,
		// or an eager refresh on a revoked token). Mark the account synced
		// rather than returning nil: an unsynced row is re-claimed every poll
		// with no backoff and hot-loops the vendor. A reconnect clears it.
		if ctx.Err() != nil {
			return err
		}

		logger.WarnCtx(
			ctx,
			"cannot set up name resolver, keeping placeholder name",
			log.Error(err),
		)

		return h.markNameSynced(ctx, &account, "")
	}

	if resolver == nil {
		logger.DebugCtx(
			ctx,
			"no name resolver for provider, keeping placeholder name",
			log.String("provider", dbConnector.Provider.String()),
		)

		return h.markNameSynced(ctx, &account, "")
	}

	// The budget is charged at claim time, so a crash between the charge and
	// the resolve leaves the row already over it. Retire it here rather than
	// spending another request the budget did not authorise.
	if account.NameSyncAttempts > maxNameSyncAttempts {
		logger.WarnCtx(
			ctx,
			"name resolution exhausted its attempts, keeping placeholder name",
			log.String("provider", dbConnector.Provider.String()),
			log.Int("attempts", account.NameSyncAttempts),
		)

		return h.markNameSynced(ctx, &account, "")
	}

	resolveCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	instanceName, err := resolver.ResolveInstanceName(resolveCtx)
	if err != nil {
		// Retiring a known-permanent failure on its first attempt spares the
		// budget four pointless requests. It is only that shortcut: the
		// budget below, not this branch, is what bounds the claim.
		if errors.Is(err, drivers.ErrTerminalNameResolution) {
			logger.WarnCtx(
				ctx,
				"permanent name resolution failure, keeping placeholder name",
				log.String("provider", dbConnector.Provider.String()),
				log.Error(err),
			)

			return h.markNameSynced(ctx, &account, "")
		}

		if account.NameSyncAttempts >= maxNameSyncAttempts {
			logger.WarnCtx(
				ctx,
				"name resolution exhausted its attempts, keeping placeholder name",
				log.String("provider", dbConnector.Provider.String()),
				log.Int("attempts", account.NameSyncAttempts),
				log.Error(err),
			)

			return h.markNameSynced(ctx, &account, "")
		}

		return fmt.Errorf("cannot resolve instance name for account %s: %w", account.ID, err)
	}

	if instanceName == "" {
		logger.DebugCtx(
			ctx,
			"instance name is empty, keeping placeholder name",
			log.String("provider", dbConnector.Provider.String()),
		)

		return h.markNameSynced(ctx, &account, "")
	}

	logger.InfoCtx(
		ctx,
		"resolved account name",
		log.String("connector_id", dbConnector.ID.String()),
	)

	return h.markNameSynced(ctx, &account, instanceName)
}

func (h *accountNameHandler) markNameSynced(
	ctx context.Context,
	account *coredata.ConnectorAccount,
	instanceName string,
) error {
	if instanceName != "" {
		account.Name = instanceName
	}

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(account.ID)

			_, err := account.MarkNameSynced(ctx, tx, scope, time.Now())

			return err
		},
	)
}

func (h *accountNameHandler) newCloudNameResolver(
	ctx context.Context,
	account *coredata.ConnectorAccount,
	dbConnector *coredata.Connector,
) (drivers.NameResolver, error) {
	reg, ok := h.providerRegistry.Get(dbConnector.Provider)
	if !ok || reg.WorkloadIdentity == nil || reg.WorkloadIdentity.NewNameResolver == nil {
		return nil, nil
	}

	session, err := provider.OpenSession(ctx, h.federation, h.providerRegistry, dbConnector, account.ExternalAccountID)
	if err != nil {
		return nil, err
	}

	return reg.WorkloadIdentity.NewNameResolver(ctx, session, dbConnector, h.logger), nil
}

func (h *accountNameHandler) newHTTPNameResolver(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	dbConnector *coredata.Connector,
) (drivers.NameResolver, error) {
	reg, ok := h.providerRegistry.Get(dbConnector.Provider)
	if !ok || reg.NewNameResolver == nil {
		return nil, nil
	}

	oauth2Conn := connectorprotocol.OAuth2ConnectionOf(dbConnector.Connection)

	var tokenBefore string
	if oauth2Conn != nil {
		tokenBefore = oauth2Conn.AccessToken
	}

	httpClient, err := provider.HTTPClient(
		ctx,
		h.connectorRegistry,
		h.providerRegistry,
		dbConnector,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create HTTP client for connector: %w", err)
	}

	if oauth2Conn != nil && oauth2Conn.AccessToken != tokenBefore {
		dbConnector.UpdatedAt = time.Now()
		if err := dbConnector.Update(ctx, tx, scope, h.encryptionKey); err != nil {
			return nil, fmt.Errorf("cannot persist refreshed token for connector %s: %w", dbConnector.ID, err)
		}
	}

	return reg.NewNameResolver(ctx, httpClient, dbConnector, h.logger, reg.Endpoints), nil
}
