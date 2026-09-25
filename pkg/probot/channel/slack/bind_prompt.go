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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

const bindCallbackExpiry = 30 * time.Minute

type BindPromptService struct {
	pg            *pg.Client
	encryptionKey cipher.EncryptionKey
	httpClient    *http.Client
	logger        *log.Logger
	now           func() time.Time
}

func NewBindPromptService(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	logger *log.Logger,
) *BindPromptService {
	return &BindPromptService{
		pg:            pgClient,
		encryptionKey: encryptionKey,
		httpClient: httpclient.DefaultPooledClient(
			httpclient.WithLogger(logger),
			httpclient.WithSSRFProtection(),
		),
		logger: logger,
		now:    time.Now,
	}
}

func (s *BindPromptService) RememberResponseURL(
	ctx context.Context,
	teamID string,
	userID string,
	responseURL string,
) error {
	if s == nil {
		return nil
	}

	normalized, err := normalizeSlackResponseURL(responseURL)
	if err != nil {
		return fmt.Errorf("cannot normalize Slack response URL: %w", err)
	}

	encrypted, err := cipher.Encrypt([]byte(normalized), s.encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot encrypt Slack bind response URL: %w", err)
	}

	now := s.now()
	callback := coredata.SlackbotBindCallback{
		TeamID:               teamID,
		UserID:               userID,
		EncryptedResponseURL: encrypted,
		ExpiresAt:            now.Add(bindCallbackExpiry),
		CreatedAt:            now,
	}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return callback.Upsert(ctx, conn)
		},
	)
	if err != nil {
		return fmt.Errorf("cannot persist Slack bind callback: %w", err)
	}

	return nil
}

func (s *BindPromptService) BindingConfirmed(
	ctx context.Context,
	subject identitybinding.Subject,
) error {
	if s == nil || subject.Provider != ProviderName {
		return nil
	}

	var callback coredata.SlackbotBindCallback

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := callback.LoadByTeamAndUser(
				ctx,
				conn,
				subject.ExternalTenantID,
				subject.ExternalUserID,
			); err != nil {
				return fmt.Errorf("cannot load Slack bind callback: %w", err)
			}

			if err := coredata.DeleteSlackbotBindCallback(
				ctx,
				conn,
				subject.ExternalTenantID,
				subject.ExternalUserID,
			); err != nil {
				return fmt.Errorf("cannot delete Slack bind callback: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil
		}

		return fmt.Errorf("cannot load Slack bind callback: %w", err)
	}

	if s.now().After(callback.ExpiresAt) {
		return nil
	}

	plaintext, err := cipher.Decrypt(callback.EncryptedResponseURL, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot decrypt Slack bind response URL: %w", err)
	}

	if err := s.replaceEphemeral(ctx, string(plaintext), bindSlashLinkedText); err != nil {
		if s.logger != nil {
			s.logger.ErrorCtx(ctx, "cannot replace Slack bind prompt", log.Error(err))
		}

		return nil
	}

	return nil
}

func (s *BindPromptService) replaceEphemeral(
	ctx context.Context,
	responseURL string,
	text string,
) error {
	normalized, err := normalizeSlackResponseURL(responseURL)
	if err != nil {
		return fmt.Errorf("cannot normalize Slack response URL: %w", err)
	}

	body, err := json.Marshal(
		map[string]any{
			"replace_original": true,
			"response_type":    SlashResponseTypeEphemeral,
			"text":             text,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot marshal Slack bind prompt replacement: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		normalized,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("cannot create Slack bind prompt replacement request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot replace Slack bind prompt: %w", sanitizeHTTPError(err))
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("cannot replace Slack bind prompt: status %d", resp.StatusCode)
	}

	return nil
}

func normalizeSlackResponseURL(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("cannot parse Slack response URL: %w", err)
	}

	if parsed.Scheme != "https" || parsed.Host != "hooks.slack.com" {
		return "", fmt.Errorf("slack response URL host is not allowed")
	}

	if !strings.HasPrefix(parsed.Path, "/commands/") &&
		!strings.HasPrefix(parsed.Path, "/actions/") {
		return "", fmt.Errorf("slack response URL path is not allowed")
	}

	return parsed.String(), nil
}
