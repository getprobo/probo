// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package probo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

var (
	welcomeTemplate = template.Must(
		template.New("welcome.json.tmpl").
			Funcs(template.FuncMap{
				"jsonEscape": func(s string) string {
					b, _ := json.Marshal(s)
					return string(b[1 : len(b)-1])
				},
			}).
			ParseFS(Templates, "templates/welcome.json.tmpl"),
	)
)

type (
	ConnectorService struct {
		svc *TenantService
	}

	CreateConnectorRequest struct {
		OrganizationID              gid.GID
		Provider                    coredata.ConnectorProvider
		Protocol                    coredata.ConnectorProtocol
		Connection                  connector.Connection
		TallySettings               *coredata.TallyConnectorSettings
		OnePasswordSettings         *coredata.OnePasswordConnectorSettings
		SentrySettings              *coredata.SentryConnectorSettings
		SupabaseSettings            *coredata.SupabaseConnectorSettings
		GitHubSettings              *coredata.GitHubConnectorSettings
		OnePasswordUsersAPISettings *coredata.OnePasswordUsersAPISettings
	}

	ReconnectConnectorRequest struct {
		ConnectorID    gid.GID
		OrganizationID gid.GID
		Provider       coredata.ConnectorProvider
		Connection     connector.Connection
	}
)

func (car *CreateConnectorRequest) Validate() error {
	v := validator.New()
	v.Check(car.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(car.Provider, "provider", validator.Required(), validator.OneOfSlice(coredata.ConnectorProviders()))
	v.Check(car.Protocol, "protocol", validator.Required(), validator.OneOfSlice(coredata.ConnectorProtocols()))
	v.Check(car.Connection, "connection", validator.Required())
	return v.Error()
}

func (rcr *ReconnectConnectorRequest) Validate() error {
	v := validator.New()
	v.Check(rcr.ConnectorID, "connector_id", validator.Required(), validator.GID(coredata.ConnectorEntityType))
	v.Check(rcr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(rcr.Provider, "provider", validator.Required(), validator.OneOfSlice(coredata.ConnectorProviders()))
	v.Check(rcr.Connection, "connection", validator.Required())
	return v.Error()
}

func (s *ConnectorService) ListForOrganizationID(
	ctx context.Context,
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
				s.svc.scope,
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
	ctx context.Context,
	organizationID gid.GID,
) (coredata.Connectors, error) {
	var connectors coredata.Connectors

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return connectors.LoadAllByOrganizationIDWithoutDecryptedConnection(
				ctx,
				conn,
				s.svc.scope,
				organizationID,
			)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list all connectors: %w", err)
	}

	return connectors, nil
}

func (s *ConnectorService) GetByOrganizationIDAndProvider(
	ctx context.Context,
	organizationID gid.GID,
	provider coredata.ConnectorProvider,
) (*coredata.Connector, error) {
	cnnctr := &coredata.Connector{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return cnnctr.LoadOneByOrganizationIDAndProvider(
				ctx,
				conn,
				s.svc.scope,
				s.svc.encryptionKey,
				organizationID,
				provider,
			)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get connector: %w", err)
	}

	return cnnctr, nil
}

// GetWithConnection loads a specific connector by ID and returns the
// full *coredata.Connector with Connection populated. Used by the
// initiate handler's explicit reconnect path (?connector_id=<id>),
// which needs to read the stored scope set to compute the union.
// Contrast with Get, which uses LoadMetadataByID and returns a
// connector with Connection == nil.
func (s *ConnectorService) GetWithConnection(
	ctx context.Context,
	connectorID gid.GID,
) (*coredata.Connector, error) {
	cnnctr := &coredata.Connector{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return cnnctr.LoadByID(ctx, conn, s.svc.scope, connectorID, s.svc.encryptionKey)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get connector: %w", err)
	}

	return cnnctr, nil
}

func (s *ConnectorService) Get(
	ctx context.Context,
	connectorID gid.GID,
) (*coredata.Connector, error) {
	connector := &coredata.Connector{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return connector.LoadMetadataByID(ctx, conn, s.svc.scope, connectorID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get connector: %w", err)
	}

	return connector, nil
}

func (s *ConnectorService) Delete(
	ctx context.Context,
	connectorID gid.GID,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			cnnctr := &coredata.Connector{ID: connectorID}
			return cnnctr.Delete(ctx, tx, s.svc.scope)
		},
	)
}

func (s *ConnectorService) Create(
	ctx context.Context,
	req CreateConnectorRequest,
) (*coredata.Connector, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	id := gid.New(s.svc.scope.GetTenantID(), coredata.ConnectorEntityType)
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

	switch {
	case req.TallySettings != nil:
		if err := newConnector.SetSettings(req.TallySettings); err != nil {
			return nil, fmt.Errorf("cannot set tally settings: %w", err)
		}
	case req.OnePasswordSettings != nil:
		if err := newConnector.SetSettings(req.OnePasswordSettings); err != nil {
			return nil, fmt.Errorf("cannot set one password settings: %w", err)
		}
	case req.SentrySettings != nil:
		if err := newConnector.SetSettings(req.SentrySettings); err != nil {
			return nil, fmt.Errorf("cannot set sentry settings: %w", err)
		}
	case req.SupabaseSettings != nil:
		if err := newConnector.SetSettings(req.SupabaseSettings); err != nil {
			return nil, fmt.Errorf("cannot set supabase settings: %w", err)
		}
	case req.GitHubSettings != nil:
		if err := newConnector.SetSettings(req.GitHubSettings); err != nil {
			return nil, fmt.Errorf("cannot set github settings: %w", err)
		}
	case req.OnePasswordUsersAPISettings != nil:
		if err := newConnector.SetSettings(req.OnePasswordUsersAPISettings); err != nil {
			return nil, fmt.Errorf("cannot set one password users api settings: %w", err)
		}
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := newConnector.Insert(ctx, tx, s.svc.scope, s.svc.encryptionKey); err != nil {
				return fmt.Errorf("cannot create connector: %w", err)
			}

			if req.Provider == coredata.ConnectorProviderSlack {
				slackConn, ok := req.Connection.(*connector.SlackConnection)
				if ok && slackConn.Settings.Channel != "" {
					var organization coredata.Organization
					if err := organization.LoadByID(ctx, tx, s.svc.scope, req.OrganizationID); err != nil {
						return fmt.Errorf("cannot load organization: %w", err)
					}

					data := struct {
						OrganizationName string
						ChannelName      string
					}{
						OrganizationName: organization.Name,
						ChannelName:      slackConn.Settings.Channel,
					}

					var buf bytes.Buffer
					if err := welcomeTemplate.Execute(&buf, data); err != nil {
						return fmt.Errorf("cannot execute template: %w", err)
					}

					var body map[string]any
					if err := json.NewDecoder(&buf).Decode(&body); err != nil {
						return fmt.Errorf("cannot parse template JSON: %w", err)
					}

					slackMessage := coredata.NewSlackMessage(s.svc.scope, req.OrganizationID, coredata.SlackMessageTypeWelcome, body)
					if err := slackMessage.Insert(ctx, tx, s.svc.scope); err != nil {
						return fmt.Errorf("cannot insert slack message: %w", err)
					}
				}
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
	ctx context.Context,
	req ReconnectConnectorRequest,
) (*coredata.Connector, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("cannot reconnect connector: %w", err)
	}

	cnnctr := &coredata.Connector{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := cnnctr.LoadByID(ctx, conn, s.svc.scope, req.ConnectorID, s.svc.encryptionKey); err != nil {
				return fmt.Errorf("cannot load connector: %w", err)
			}

			if cnnctr.OrganizationID != req.OrganizationID {
				return fmt.Errorf("cannot reconnect connector: organization mismatch")
			}
			if cnnctr.Provider != req.Provider {
				return fmt.Errorf("cannot reconnect connector: provider mismatch")
			}
			if cnnctr.Protocol != coredata.ConnectorProtocolOAuth2 {
				return fmt.Errorf("cannot reconnect connector: not an OAuth2 connector")
			}

			preserveConnectionFields(req.Connection, cnnctr.Connection)
			cnnctr.Connection = req.Connection
			cnnctr.UpdatedAt = time.Now()

			return cnnctr.Update(ctx, conn, s.svc.scope, s.svc.encryptionKey)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot reconnect connector: %w", err)
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
