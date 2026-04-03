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

package slack

import (
	"context"
	"fmt"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type Service struct {
	pg                 *pg.Client
	logger             *log.Logger
	slackSigningSecret string
	baseURL            string
	tokenSecret        string
}

type TenantService struct {
	pg            *pg.Client
	scope         coredata.Scoper
	logger        *log.Logger
	baseURL       string
	tokenSecret   string
	SlackMessages *SlackMessageService
}

func NewService(
	pg *pg.Client,
	slackSigningSecret string,
	baseURL string,
	tokenSecret string,
	logger *log.Logger,
) *Service {
	return &Service{
		pg:                 pg,
		logger:             logger,
		slackSigningSecret: slackSigningSecret,
		baseURL:            baseURL,
		tokenSecret:        tokenSecret,
	}
}

func (s *Service) WithTenant(tenantID gid.TenantID) *TenantService {
	tenantService := &TenantService{
		pg:          s.pg,
		scope:       coredata.NewScope(tenantID),
		logger:      s.logger,
		baseURL:     s.baseURL,
		tokenSecret: s.tokenSecret,
	}
	tenantService.SlackMessages = &SlackMessageService{svc: tenantService}

	return tenantService
}

func (s *TenantService) GetSlackClient() *Client {
	return NewClient(s.logger)
}

func (s *TenantService) GetSlackMessageService() *SlackMessageService {
	return &SlackMessageService{svc: s}
}

func (s *Service) GetSlackSigningSecret() string {
	return s.slackSigningSecret
}

func (s *Service) GetInitialSlackMessageByChannelAndTS(
	ctx context.Context,
	channelID string,
	messageTS string,
) (*coredata.SlackMessage, error) {
	var slackMessage coredata.SlackMessage

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := slackMessage.LoadInitialByChannelAndTS(ctx, conn, coredata.NewNoScope(), channelID, messageTS); err != nil {
			return fmt.Errorf("cannot load slack message: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &slackMessage, nil
}
