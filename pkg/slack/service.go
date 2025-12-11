package slack

import (
	"context"
	"fmt"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

type Service struct {
	pg                 *pg.Client
	logger             *log.Logger
	slackSigningSecret string
	baseURL            string
	encryptionKey      cipher.EncryptionKey
	tokenSecret        string
}

type TenantService struct {
	pg            *pg.Client
	scope         coredata.Scoper
	logger        *log.Logger
	baseURL       string
	encryptionKey cipher.EncryptionKey
	tokenSecret   string
	SlackMessages *SlackMessageService
}

func NewService(
	pg *pg.Client,
	slackSigningSecret string,
	baseURL string,
	encryptionKey cipher.EncryptionKey,
	tokenSecret string,
	logger *log.Logger,
) *Service {
	return &Service{
		pg:                 pg,
		logger:             logger,
		slackSigningSecret: slackSigningSecret,
		baseURL:            baseURL,
		encryptionKey:      encryptionKey,
		tokenSecret:        tokenSecret,
	}
}

func (s *Service) WithTenant(tenantID gid.TenantID) *TenantService {
	tenantService := &TenantService{
		pg:            s.pg,
		scope:         coredata.NewScope(tenantID),
		logger:        s.logger,
		baseURL:       s.baseURL,
		encryptionKey: s.encryptionKey,
		tokenSecret:   s.tokenSecret,
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

	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
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
