// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
		OrganizationID gid.GID
		Provider       coredata.ConnectorProvider
		Protocol       coredata.ConnectorProtocol
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

func (s *ConnectorService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ConnectorOrderField],
	filter *coredata.ConnectorFilter,
) (*page.Page[*coredata.Connector, coredata.ConnectorOrderField], error) {
	var connectors coredata.Connectors

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
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

func (s *ConnectorService) GetByOrganizationIDAndProvider(
	ctx context.Context,
	organizationID gid.GID,
	provider coredata.ConnectorProvider,
) (*coredata.Connector, error) {
	var connectors coredata.Connectors

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return connectors.LoadAllByOrganizationIDProtocolAndProvider(
				ctx,
				conn,
				s.svc.scope,
				organizationID,
				coredata.ConnectorProtocolOAuth2,
				provider,
				s.svc.encryptionKey,
			)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot get connector: %w", err)
	}

	if len(connectors) == 0 {
		return nil, coredata.ErrResourceNotFound
	}

	return connectors[0], nil
}

func (s *ConnectorService) Delete(
	ctx context.Context,
	connectorID gid.GID,
) error {
	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			cnnctr := &coredata.Connector{ID: connectorID}
			return cnnctr.Delete(ctx, conn, s.svc.scope)
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

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := newConnector.Insert(ctx, conn, s.svc.scope, s.svc.encryptionKey); err != nil {
				return fmt.Errorf("cannot create connector: %w", err)
			}

			if req.Provider == coredata.ConnectorProviderSlack {
				slackConn, ok := req.Connection.(*connector.SlackConnection)
				if ok && slackConn.Settings.Channel != "" {
					var organization coredata.Organization
					if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
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
					if err := slackMessage.Insert(ctx, conn, s.svc.scope); err != nil {
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
