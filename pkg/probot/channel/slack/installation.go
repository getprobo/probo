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

package slack

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

const (
	tokenRefreshLeeway        = time.Minute
	installStateStaleAfter    = 5 * time.Minute
	credentialPersistAttempts = 3
)

var (
	ErrSlackbotNotInstalled      = errors.New("slack app is not installed")
	ErrSlackbotChannelNotFound   = errors.New("slack notification channel not found")
	ErrSlackbotStateAlreadyUsed  = errors.New("slack install state already used")
	errStaleInstallationRevision = errors.New(
		"slack installation was updated during credential refresh",
	)
)

type (
	InstallationConfig struct {
		ClientID      string
		ClientSecret  string
		RedirectURI   string
		AuthURL       string
		TokenURL      string
		APIBaseURL    string
		StateSecret   string
		SigningSecret string
	}

	InstallationService struct {
		pg            *pg.Client
		encryptionKey cipher.EncryptionKey
		cfg           InstallationConfig
		httpClient    *http.Client
		logger        *log.Logger
	}

	InstallResult struct {
		Installation *coredata.SlackbotInstallation
		ContinueURL  string
	}

	oauthExchangeError struct {
		err       error
		transient bool
	}

	slackOAuthResponse struct {
		OK           bool   `json:"ok"`
		Error        string `json:"error,omitempty"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token,omitempty"`
		ExpiresIn    int64  `json:"expires_in,omitempty"`
		Scope        string `json:"scope,omitempty"`
		BotUserID    string `json:"bot_user_id"`
		Team         struct {
			ID string `json:"id"`
		} `json:"team"`
	}
)

func (e *oauthExchangeError) Error() string {
	return e.err.Error()
}

func (e *oauthExchangeError) Unwrap() error {
	return e.err
}

var installationScopes = []string{
	"app_mentions:read",
	"assistant:write",
	"channels:history",
	"channels:read",
	"chat:write",
	"commands",
	"groups:history",
	"groups:read",
	"im:history",
	"im:write",
	"reactions:read",
	"reactions:write",
}

func (s *InstallationService) SigningSecret() string {
	return s.cfg.SigningSecret
}

func NewInstallationService(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	cfg InstallationConfig,
	logger *log.Logger,
) *InstallationService {
	httpClient := httpclient.DefaultPooledClient(
		httpclient.WithLogger(logger),
		httpclient.WithSSRFProtection(),
	)
	httpClient.Timeout = slackHTTPTimeout

	return &InstallationService{
		pg:            pgClient,
		encryptionKey: encryptionKey,
		cfg:           cfg,
		httpClient:    httpClient,
		logger:        logger,
	}
}

func (s *InstallationService) GetByOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*coredata.SlackbotInstallation, error) {
	var installation coredata.SlackbotInstallation

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return installation.LoadByOrganizationID(
				ctx,
				conn,
				scope,
				organizationID,
			)
		},
	)
	if err != nil {
		return nil, err
	}

	return &installation, nil
}

func (s *InstallationService) GetByTeamID(
	ctx context.Context,
	teamID string,
) (*coredata.SlackbotInstallation, error) {
	var installation coredata.SlackbotInstallation

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return installation.LoadByTeamID(
				ctx,
				conn,
				coredata.NewNoScope(),
				teamID,
			)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, ErrSlackbotNotInstalled
		}

		return nil, fmt.Errorf("cannot load Slack installation by team: %w", err)
	}

	if installation.Status != coredata.SlackbotInstallationStatusActive {
		return nil, ErrSlackbotNotInstalled
	}

	return &installation, nil
}

func (s *InstallationService) ClientByOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*Client, *coredata.SlackbotInstallation, error) {
	installation, credentials, err := s.loadUsableCredentials(
		ctx,
		scope,
		organizationID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot load usable Slack credentials: %w", err)
	}

	return newAPIClient(
		credentials.AccessToken,
		s.cfg.APIBaseURL,
		s.httpClient,
	), installation, nil
}

func (s *InstallationService) ClientByTeamID(
	ctx context.Context,
	teamID string,
) (*Client, *coredata.SlackbotInstallation, error) {
	var installation coredata.SlackbotInstallation

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return installation.LoadByTeamID(
				ctx,
				conn,
				coredata.NewNoScope(),
				teamID,
			)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil, ErrSlackbotNotInstalled
		}

		return nil, nil, fmt.Errorf("cannot load Slack installation by team: %w", err)
	}

	return s.ClientByOrganizationID(
		ctx,
		coredata.NewScopeFromObjectID(installation.OrganizationID),
		installation.OrganizationID,
	)
}

func (s *InstallationService) Uninstall(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) error {
	installation, err := s.GetByOrganizationID(ctx, scope, organizationID)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("cannot load Slack installation for uninstall: %w", err)
	}

	installationID := installation.ID
	installationUpdatedAt := installation.UpdatedAt

	if installation.Status == coredata.SlackbotInstallationStatusActive {
		client, _, err := s.ClientByOrganizationID(
			ctx,
			scope,
			organizationID,
		)
		if err != nil {
			return fmt.Errorf("cannot load Slack client for uninstall: %w", err)
		}

		if err := client.UninstallApp(
			ctx,
			s.cfg.ClientID,
			s.cfg.ClientSecret,
		); err != nil {
			return fmt.Errorf("cannot uninstall Slack app: %w", err)
		}
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := installation.LoadByOrganizationIDForUpdate(
				ctx,
				tx,
				scope,
				organizationID,
			); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot lock Slack installation for uninstall: %w", err)
			}

			// Organization upserts preserve the row ID, so a reinstall during
			// apps.uninstall is only visible via UpdatedAt (or other field) drift.
			if installation.ID != installationID ||
				!installation.UpdatedAt.Equal(installationUpdatedAt) {
				return nil
			}

			if err := coredata.DeleteBotDeliveryDestinationsByProviderAndOrganizationID(
				ctx,
				tx,
				scope,
				ProviderName,
				organizationID,
			); err != nil {
				return fmt.Errorf("cannot delete Slack delivery destinations: %w", err)
			}

			if err := deleteWorkspaceIdentityState(
				ctx,
				tx,
				installation.TeamID,
			); err != nil {
				return fmt.Errorf("cannot delete Slack identity state: %w", err)
			}

			if err := installation.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete Slack installation: %w", err)
			}

			return nil
		},
	)
}

func deleteWorkspaceIdentityState(
	ctx context.Context,
	conn pg.Querier,
	teamID string,
) error {
	if err := identitybinding.DeleteByExternalTenant(
		ctx,
		conn,
		ProviderName,
		teamID,
	); err != nil {
		return fmt.Errorf("cannot delete Slack identity bindings: %w", err)
	}

	if err := coredata.DeleteSlackbotBindCallbacksByTeamID(ctx, conn, teamID); err != nil {
		return fmt.Errorf("cannot delete Slack bind callbacks: %w", err)
	}

	return nil
}

func (s *InstallationService) ListMemberConversations(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor string,
) (*ConversationsPage, error) {
	client, _, err := s.ClientByOrganizationID(ctx, scope, organizationID)
	if err != nil {
		return nil, fmt.Errorf("cannot load Slack client: %w", err)
	}

	page, err := client.ListConversations(ctx, cursor)
	if err != nil {
		return nil, fmt.Errorf("cannot list Slack conversations: %w", err)
	}

	memberConversations := make([]Conversation, 0, len(page.Conversations))
	for _, conversation := range page.Conversations {
		if conversation.IsMember && !conversation.IsArchived {
			memberConversations = append(memberConversations, conversation)
		}
	}

	page.Conversations = memberConversations

	return page, nil
}

func (s *InstallationService) DisableByTeamID(
	ctx context.Context,
	teamID string,
	eventTime *time.Time,
	botUserIDs []string,
) error {
	var installation coredata.SlackbotInstallation

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := installation.LoadByTeamID(
				ctx,
				tx,
				coredata.NewNoScope(),
				teamID,
			); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load Slack installation by team: %w", err)
			}

			if eventTime != nil {
				if !shouldDisableInstallationForEvent(installation, *eventTime, botUserIDs) {
					return nil
				}
			} else if !installationMatchesDisableIdentifiers(installation, botUserIDs) {
				return nil
			}

			installation.Status = coredata.SlackbotInstallationStatusDisabled
			installation.UpdatedAt = time.Now()

			return installation.UpdateStatus(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(installation.OrganizationID),
			)
		},
	)
}

// shouldDisableInstallationForEvent compares install and event times at second
// precision because Slack event_time is unix seconds. Same-second token
// revocations also require a bot-user match so a reinstall in that second is
// not disabled by a stale revocation for a previous bot.
func shouldDisableInstallationForEvent(
	installation coredata.SlackbotInstallation,
	eventTime time.Time,
	botUserIDs []string,
) bool {
	installSecond := installation.UpdatedAt.UTC().Truncate(time.Second)
	eventSecond := eventTime.UTC().Truncate(time.Second)

	if installSecond.After(eventSecond) {
		return false
	}

	if installSecond.Equal(eventSecond) && len(botUserIDs) > 0 {
		return installationMatchesDisableIdentifiers(installation, botUserIDs)
	}

	return true
}

func installationMatchesDisableIdentifiers(
	installation coredata.SlackbotInstallation,
	botUserIDs []string,
) bool {
	if installation.BotUserID == "" {
		return false
	}

	for _, botUserID := range botUserIDs {
		if botUserID == installation.BotUserID {
			return true
		}
	}

	return false
}

func findMemberConversation(
	ctx context.Context,
	client *Client,
	channelID string,
) (*Conversation, error) {
	cursor := ""
	for {
		page, err := client.ListConversations(ctx, cursor)
		if err != nil {
			return nil, err
		}

		for _, conversation := range page.Conversations {
			if conversation.ID == channelID &&
				conversation.IsMember &&
				!conversation.IsArchived {
				return &conversation, nil
			}
		}

		cursor = page.NextCursor
		if cursor == "" {
			return nil, ErrSlackbotChannelNotFound
		}
	}
}

func splitScopes(raw string) []string {
	return strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' '
	})
}
