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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func (s *InstallationService) InitiateURL(
	organizationID gid.GID,
	identityID gid.GID,
	continueURL string,
) (string, error) {
	state, err := newInstallState(
		s.cfg.StateSecret,
		organizationID,
		identityID,
		continueURL,
	)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(s.cfg.AuthURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse Slack authorization URL: %w", err)
	}

	query := u.Query()
	query.Set("client_id", s.cfg.ClientID)
	query.Set("redirect_uri", s.cfg.RedirectURI)
	query.Set("scope", strings.Join(installationScopes, ","))
	query.Set("state", state)
	u.RawQuery = query.Encode()

	return u.String(), nil
}

func (s *InstallationService) Complete(
	ctx context.Context,
	state string,
	code string,
) (*InstallResult, error) {
	payload, err := validateInstallState(s.cfg.StateSecret, state)
	if err != nil {
		return nil, err
	}

	scope := coredata.NewScopeFromObjectID(payload.Data.OrganizationID)

	processingToken, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("cannot generate Slack install processing token: %w", err)
	}

	claim := coredata.NewSlackbotInstallStateClaim(
		payload.Data.OrganizationID,
		state,
	)
	claimed := false

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var err error

			claimed, err = claim.Claim(
				ctx,
				tx,
				scope,
				processingToken.String(),
				time.Now(),
				installStateStaleAfter,
			)
			if err != nil {
				return fmt.Errorf("cannot claim Slack install state: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot persist Slack install state claim: %w", err)
	}

	if !claimed {
		return nil, ErrSlackbotStateAlreadyUsed
	}

	oauthResponse, err := s.exchangeCode(ctx, code)
	if err != nil {
		persistErr := s.pg.WithConn(
			context.WithoutCancel(ctx),
			func(ctx context.Context, conn pg.Querier) error {
				if isTransientOAuthExchangeError(err) {
					return claim.Release(ctx, conn, scope, processingToken.String())
				}

				return claim.Complete(
					ctx,
					conn,
					scope,
					processingToken.String(),
					time.Now(),
				)
			},
		)
		if persistErr != nil {
			return nil, fmt.Errorf(
				"cannot finalize Slack install state after exchange failure: %w: %w",
				err,
				persistErr,
			)
		}

		return nil, fmt.Errorf("cannot exchange Slack OAuth code: %w", err)
	}

	installation := coredata.NewSlackbotInstallation(
		scope,
		payload.Data.OrganizationID,
	)
	installation.TeamID = oauthResponse.Team.ID
	installation.BotUserID = oauthResponse.BotUserID

	installation.Scopes = splitScopes(oauthResponse.Scope)
	if oauthResponse.ExpiresIn > 0 {
		installation.AccessTokenExpiresAt = new(
			time.Now().Add(time.Duration(oauthResponse.ExpiresIn) * time.Second),
		)
	}

	credentials := coredata.SlackbotInstallationCredentials{
		AccessToken: oauthResponse.AccessToken,
	}
	if oauthResponse.RefreshToken != "" {
		credentials.RefreshToken = new(oauthResponse.RefreshToken)
	}

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var existing coredata.SlackbotInstallation

			loadErr := existing.LoadByOrganizationIDForUpdate(
				ctx,
				tx,
				scope,
				payload.Data.OrganizationID,
			)
			if loadErr != nil && !errors.Is(loadErr, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load existing Slack installation: %w", loadErr)
			}

			if loadErr == nil && existing.TeamID != oauthResponse.Team.ID {
				if err := coredata.DeleteBotDeliveryDestinationsByProviderAndOrganizationID(
					ctx,
					tx,
					scope,
					ProviderName,
					payload.Data.OrganizationID,
				); err != nil {
					return fmt.Errorf("cannot clear destinations for changed Slack workspace: %w", err)
				}

				if err := deleteWorkspaceIdentityState(
					ctx,
					tx,
					existing.TeamID,
				); err != nil {
					return fmt.Errorf("cannot clear identity state for changed Slack workspace: %w", err)
				}
			}

			if _, err := installation.Upsert(
				ctx,
				tx,
				scope,
				s.encryptionKey,
				credentials,
			); err != nil {
				return fmt.Errorf("cannot save Slack installation: %w", err)
			}

			if err := claim.Complete(
				ctx,
				tx,
				scope,
				processingToken.String(),
				time.Now(),
			); err != nil {
				return fmt.Errorf("cannot complete Slack install state: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot save Slack installation: %w", err)
	}

	return &InstallResult{
		Installation: installation,
		ContinueURL:  payload.Data.ContinueURL,
	}, nil
}

func (s *InstallationService) exchangeCode(
	ctx context.Context,
	code string,
) (*slackOAuthResponse, error) {
	form := url.Values{}
	form.Set("client_id", s.cfg.ClientID)
	form.Set("client_secret", s.cfg.ClientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", s.cfg.RedirectURI)

	return s.postOAuth(ctx, form)
}

func (s *InstallationService) refreshToken(
	ctx context.Context,
	refreshToken string,
) (*slackOAuthResponse, error) {
	form := url.Values{}
	form.Set("client_id", s.cfg.ClientID)
	form.Set("client_secret", s.cfg.ClientSecret)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	return s.postOAuth(ctx, form)
}

func (s *InstallationService) postOAuth(
	ctx context.Context,
	form url.Values,
) (*slackOAuthResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.cfg.TokenURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create Slack OAuth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, &oauthExchangeError{
			err:       fmt.Errorf("cannot exchange Slack OAuth token: %w", err),
			transient: true,
		}
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &oauthExchangeError{
			err:       fmt.Errorf("cannot read Slack OAuth response: %w", err),
			transient: true,
		}
	}

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError {
		return nil, &oauthExchangeError{
			err:       fmt.Errorf("slack OAuth endpoint returned HTTP %d", resp.StatusCode),
			transient: true,
		}
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &oauthExchangeError{
			err: fmt.Errorf("slack OAuth endpoint returned HTTP %d", resp.StatusCode),
		}
	}

	var result slackOAuthResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, &oauthExchangeError{
			err:       fmt.Errorf("cannot decode Slack OAuth response: %w", err),
			transient: true,
		}
	}

	if !result.OK {
		return nil, &oauthExchangeError{
			err:       fmt.Errorf("slack OAuth error: %s", result.Error),
			transient: isTransientSlackOAuthError(result.Error),
		}
	}

	if result.AccessToken == "" {
		return nil, &oauthExchangeError{
			err: fmt.Errorf("slack OAuth response has no access token"),
		}
	}

	if form.Get("grant_type") != "refresh_token" &&
		(result.Team.ID == "" || result.BotUserID == "") {
		return nil, &oauthExchangeError{
			err: fmt.Errorf("slack OAuth response has no team or bot user"),
		}
	}

	return &result, nil
}

func isTransientOAuthExchangeError(err error) bool {
	exchangeErr, ok := errors.AsType[*oauthExchangeError](err)

	return ok && exchangeErr.transient
}

func isTransientSlackOAuthError(code string) bool {
	switch code {
	case "fatal_error", "request_timeout", "service_unavailable", "temporarily_unavailable":
		return true
	}

	return false
}
