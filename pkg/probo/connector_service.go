// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package probo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

// ErrInstallStateAlreadyUsed is returned when an install callback replays a
// state another request already claimed or completed. The vendor's proof stays
// valid (unlike an OAuth code, which the vendor itself burns) and the
// (organization_id, provider, protocol) unique index was dropped in
// 20260819T142937Z, so without this a browser refresh or a link-preview
// prefetch would create a second connector.
var ErrInstallStateAlreadyUsed = errors.New("connector install state already used")

type (
	ConnectorService struct {
		svc *Service
	}

	CreateConnectorRequest struct {
		OrganizationID gid.GID
		Provider       coredata.ConnectorProvider
		Protocol       coredata.ConnectorProtocol
		Connection     connector.Connection
		// RawSettings is the provider-specific settings payload as
		// already-marshalled JSON. Callers build it from the typed
		// gqlgen input (or OAuth callback metadata); the service layer
		// never sees the typed structs.
		RawSettings json.RawMessage
	}

	ReconnectConnectorRequest struct {
		ConnectorID    gid.GID
		OrganizationID gid.GID
		Provider       coredata.ConnectorProvider
		Connection     connector.Connection
		// RawSettings, when non-empty, replaces the connector's
		// provider-specific settings on reconnect. Datadog captures its
		// per-customer API domain on every OAuth callback (the domain
		// drives the driver's API host), so a reconnect must refresh it;
		// empty leaves the existing settings intact.
		RawSettings json.RawMessage
	}

	// CompleteConnectorInstallRequest carries the verified outcome of an
	// app-install ceremony: the vendor tenant id the callback proved control
	// of, and the single-use state claim that proof was spent against.
	CompleteConnectorInstallRequest struct {
		OrganizationID gid.GID
		Provider       coredata.ConnectorProvider
		// SettingsKey is the JSON key inside RawSettings carrying the vendor
		// tenant id (provider.InstallConfig.SettingsResourceKey). It is what
		// the idempotency lookup reads back out of settings.
		SettingsKey string
		// ResourceID is the CANONICALIZED vendor tenant id. A second spelling
		// of the same tenant takes a different advisory lock and misses the
		// idempotency lookup, so it would bind twice.
		ResourceID string
		Connection connector.Connection
		// State and ProcessingToken identify the claim this completion burns,
		// in the same transaction as the insert.
		State           string
		ProcessingToken string
	}
)

func (car *CreateConnectorRequest) Validate() error {
	v := validator.New()
	v.Check(car.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(car.Provider, "provider", validator.Required(), validator.OneOfSlice(coredata.ConnectorProviders()))
	v.Check(car.Protocol, "protocol", validator.Required(), validator.OneOfSlice(coredata.ConnectorProtocols()))
	v.Check(car.Connection, "connection", validator.Required())
	v.Check(car.RawSettings, "raw_settings", validJSONRawMessage)

	return v.Error()
}

// validJSONRawMessage rejects a non-empty RawSettings that does not
// parse as JSON. Empty RawSettings is allowed (providers without
// extra settings).
func validJSONRawMessage(value any) *validator.ValidationError {
	raw, ok := value.(json.RawMessage)
	if !ok || len(raw) == 0 {
		return nil
	}

	if !json.Valid(raw) {
		return &validator.ValidationError{
			Code:    validator.ErrorCodeInvalidFormat,
			Message: "must be valid JSON",
		}
	}

	return nil
}

func (rcr *ReconnectConnectorRequest) Validate() error {
	v := validator.New()
	v.Check(rcr.ConnectorID, "connector_id", validator.Required(), validator.GID(coredata.ConnectorEntityType))
	v.Check(rcr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(rcr.Provider, "provider", validator.Required(), validator.OneOfSlice(coredata.ConnectorProviders()))
	v.Check(rcr.Connection, "connection", validator.Required())
	v.Check(rcr.RawSettings, "raw_settings", validJSONRawMessage)

	return v.Error()
}

func (s *ConnectorService) ListForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ConnectorOrderField],
	filter *coredata.ConnectorFilter,
) (*page.Page[*coredata.Connector, coredata.ConnectorOrderField], error) {
	var connectors coredata.Connectors

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return connectors.LoadByOrganizationIDWithoutDecryptedConnection(
				ctx,
				conn,
				scope,
				organizationID,
				cursor,
				filter,
			)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list connectors: %w", err)
	}

	return page.NewPage(connectors, cursor), nil
}

func (s *ConnectorService) ListAllForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (coredata.Connectors, error) {
	var connectors coredata.Connectors

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return connectors.LoadAllByOrganizationIDWithoutDecryptedConnection(
				ctx,
				conn,
				scope,
				organizationID,
			)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list all connectors: %w", err)
	}

	return connectors, nil
}

// GetWithConnection loads a specific connector by ID and returns the
// full *coredata.Connector with Connection populated. Used by the
// initiate handler's explicit reconnect path (?connector_id=<id>),
// which needs to read the stored scope set to compute the union.
// Contrast with Get, which uses LoadMetadataByID and returns a
// connector with Connection == nil.
func (s *ConnectorService) GetWithConnection(
	ctx context.Context, scope coredata.Scoper,
	connectorID gid.GID,
) (*coredata.Connector, error) {
	cnnctr := &coredata.Connector{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return cnnctr.LoadByID(ctx, conn, scope, connectorID, s.svc.encryptionKey)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get connector: %w", err)
	}

	return cnnctr, nil
}

func (s *ConnectorService) Get(
	ctx context.Context, scope coredata.Scoper,
	connectorID gid.GID,
) (*coredata.Connector, error) {
	connector := &coredata.Connector{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return connector.LoadMetadataByID(ctx, conn, scope, connectorID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get connector: %w", err)
	}

	return connector, nil
}

func (s *ConnectorService) Delete(
	ctx context.Context, scope coredata.Scoper,
	connectorID gid.GID,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			cnnctr := &coredata.Connector{ID: connectorID}
			return cnnctr.Delete(ctx, tx, scope)
		},
	)
}

func (s *ConnectorService) Create(
	ctx context.Context, scope coredata.Scoper,
	req CreateConnectorRequest,
) (*coredata.Connector, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	id := gid.New(scope.GetTenantID(), coredata.ConnectorEntityType)
	now := time.Now()

	newConnector := &coredata.Connector{
		ID:             id,
		OrganizationID: req.OrganizationID,
		Provider:       req.Provider,
		Protocol:       req.Protocol,
		Connection:     req.Connection,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if len(req.RawSettings) > 0 {
		newConnector.RawSettings = []byte(req.RawSettings)
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := newConnector.Insert(ctx, tx, scope, s.svc.encryptionKey); err != nil {
				return fmt.Errorf("cannot create connector: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return newConnector, nil
}

// Reconnect updates an existing OAuth2 connector's connection (token)
// in place. It validates that the loaded connector belongs to the
// expected org and provider inside the same transaction, blocking
// cross-org and cross-provider corruption via a crafted connector_id
// in the initiate URL. Refresh tokens and Slack webhook settings are
// preserved from the existing connection when the new one omits them.
func (s *ConnectorService) Reconnect(
	ctx context.Context, scope coredata.Scoper,
	req ReconnectConnectorRequest,
) (*coredata.Connector, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("cannot reconnect connector: %w", err)
	}

	cnnctr := &coredata.Connector{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := cnnctr.LoadByID(ctx, conn, scope, req.ConnectorID, s.svc.encryptionKey); err != nil {
				return fmt.Errorf("cannot load connector: %w", err)
			}

			if cnnctr.OrganizationID != req.OrganizationID {
				return fmt.Errorf("cannot reconnect connector: organization mismatch")
			}

			if cnnctr.Provider != req.Provider {
				return fmt.Errorf("cannot reconnect connector: provider mismatch")
			}

			expectedProtocol := coredata.ConnectorProtocol(req.Connection.Type())
			if cnnctr.Protocol != expectedProtocol {
				return fmt.Errorf("cannot reconnect connector: protocol mismatch")
			}

			preserveConnectionFields(req.Connection, cnnctr.Connection)
			cnnctr.Connection = req.Connection

			if len(req.RawSettings) > 0 {
				cnnctr.RawSettings = []byte(req.RawSettings)
			}

			cnnctr.UpdatedAt = time.Now()

			return cnnctr.Update(ctx, conn, scope, s.svc.encryptionKey)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot reconnect connector: %w", err)
	}

	return cnnctr, nil
}

// ClaimInstallState takes exclusive ownership of an app-install state for this
// request and returns the processing token the later Release, Burn or Complete
// must present. It runs before any outbound call to the vendor, so a replay
// cannot race a legitimate completion. Returns ErrInstallStateAlreadyUsed when
// the state is already held or already spent.
func (s *ConnectorService) ClaimInstallState(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	state string,
) (string, error) {
	processingToken, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("cannot generate connector install processing token: %w", err)
	}

	claim := coredata.NewConnectorInstallStateClaim(organizationID, state)
	claimed := false

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var err error

			claimed, err = claim.Claim(
				ctx,
				tx,
				scope,
				processingToken.String(),
				time.Now(),
				coredata.InstallStateStaleAfter,
			)

			return err
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot persist connector install state claim: %w", err)
	}

	if !claimed {
		return "", ErrInstallStateAlreadyUsed
	}

	return processingToken.String(), nil
}

// ReleaseInstallState hands the state back unspent. It is the right answer only
// to a failure a retry could fix, so a vendor blip does not cost the customer
// the rest of their window.
func (s *ConnectorService) ReleaseInstallState(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	state string,
	processingToken string,
) error {
	claim := coredata.NewConnectorInstallStateClaim(organizationID, state)

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return claim.Release(ctx, conn, scope, processingToken)
		},
	)
	if err != nil {
		return fmt.Errorf("cannot release connector install state: %w", err)
	}

	return nil
}

// BurnInstallState spends the state without creating anything. It is what a
// refused vendor proof costs: the same state cannot be presented twice, so a
// forged callback gets exactly one attempt.
func (s *ConnectorService) BurnInstallState(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	state string,
	processingToken string,
) error {
	claim := coredata.NewConnectorInstallStateClaim(organizationID, state)

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return claim.Complete(ctx, conn, scope, processingToken, time.Now())
		},
	)
	if err != nil {
		return fmt.Errorf("cannot burn connector install state: %w", err)
	}

	return nil
}

// CompleteInstall persists the connector an app-install ceremony yielded and
// burns its state in one transaction: a crash between the two would otherwise
// spend the customer's state and leave nothing to show for it.
//
// It is find-or-create on the vendor tenant id, so a customer re-installing the
// same tenant lands on their existing connector instead of a second one.
// Nothing is updated on the reuse path: the tenant id is the ceremony's only
// durable output and it already matches.
func (s *ConnectorService) CompleteInstall(
	ctx context.Context, scope coredata.Scoper,
	req CompleteConnectorInstallRequest,
) (*coredata.Connector, error) {
	claim := coredata.NewConnectorInstallStateClaim(req.OrganizationID, req.State)
	cnnctr := &coredata.Connector{}

	// Built here, never by the provider: LockConnectorInstallResource keys on
	// ResourceID and the idempotency lookup reads settings ->> SettingsKey back
	// out, so the two must be the same value. A provider returning its own
	// marshalled settings could put something else under that key and every
	// re-install would insert another row.
	rawSettings, err := json.Marshal(map[string]string{req.SettingsKey: req.ResourceID})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal connector install settings: %w", err)
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			// Before the load, not after: FOR UPDATE on a row that does not
			// exist yet locks nothing, so two concurrent first installs of one
			// tenant would both miss and both insert.
			if err := coredata.LockConnectorInstallResource(
				ctx,
				tx,
				req.OrganizationID,
				req.Provider,
				req.ResourceID,
			); err != nil {
				return err
			}

			err := cnnctr.LoadByOrganizationIDProviderAndSettingForUpdate(
				ctx,
				tx,
				scope,
				req.OrganizationID,
				req.Provider,
				req.SettingsKey,
				req.ResourceID,
			)

			switch {
			case err == nil:
				// Reuse the existing row as-is.
			case errors.Is(err, coredata.ErrResourceNotFound):
				now := time.Now()

				*cnnctr = coredata.Connector{
					ID:             gid.New(scope.GetTenantID(), coredata.ConnectorEntityType),
					OrganizationID: req.OrganizationID,
					Provider:       req.Provider,
					Protocol:       coredata.ConnectorProtocolAPIKey,
					Connection:     req.Connection,
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				cnnctr.RawSettings = rawSettings

				if err := cnnctr.Insert(ctx, tx, scope, s.svc.encryptionKey); err != nil {
					return fmt.Errorf("cannot create connector: %w", err)
				}
			default:
				return fmt.Errorf("cannot load connector: %w", err)
			}

			if err := claim.Complete(ctx, tx, scope, req.ProcessingToken, time.Now()); err != nil {
				return fmt.Errorf("cannot complete connector install state: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return cnnctr, nil
}

// preserveConnectionFields copies refresh token and Slack webhook
// settings from oldConn into newConn when newConn omits them. Google
// omits refresh_token on incremental-auth reuse; a Slack access-review
// reconnect with no incoming-webhook scope omits the webhook settings.
func preserveConnectionFields(newConn, oldConn connector.Connection) {
	switch n := newConn.(type) {
	case *connector.OAuth2Connection:
		if o, ok := oldConn.(*connector.OAuth2Connection); ok {
			if n.RefreshToken == "" {
				n.RefreshToken = o.RefreshToken
			}
		}
	case *connector.SlackConnection:
		if o, ok := oldConn.(*connector.SlackConnection); ok {
			if n.RefreshToken == "" {
				n.RefreshToken = o.RefreshToken
			}

			if n.Settings.WebhookURL == "" {
				n.Settings.WebhookURL = o.Settings.WebhookURL
				n.Settings.Channel = o.Settings.Channel
				n.Settings.ChannelID = o.Settings.ChannelID
			}
		}
	}
}
