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

package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.probo.inc/probo/pkg/gid"
)

type (
	SlackConnection struct {
		OAuth2Connection
		Settings SlackSettings `json:"settings"`
	}

	SlackSettings struct {
		WebhookURL string `json:"webhook_url,omitempty"` // Encrypted
		Channel    string `json:"channel,omitempty"`
		ChannelID  string `json:"channel_id,omitempty"`
	}

	IncomingWebhook struct {
		URL       string `json:"url"`
		Channel   string `json:"channel"`
		ChannelID string `json:"channel_id"`
	}

	SlackTokenResponse struct {
		Ok              bool             `json:"ok"`
		Error           string           `json:"error,omitempty"`
		IncomingWebhook *IncomingWebhook `json:"incoming_webhook,omitempty"`
	}
)

const (
	SlackProvider = "SLACK"
)

var _ Connection = (*SlackConnection)(nil)

func (c *SlackConnection) Type() ProtocolType {
	return ProtocolOAuth2
}

func (c *SlackConnection) Client(ctx context.Context) (*http.Client, error) {
	return c.OAuth2Connection.Client(ctx)
}

func (c SlackConnection) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Type         string    `json:"type"`
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token,omitempty"`
		ExpiresAt    time.Time `json:"expires_at"`
		TokenType    string    `json:"token_type"`
		Scope        string    `json:"scope,omitempty"`
		WebhookURL   string    `json:"webhook_url,omitempty"`
	}{
		Type:         string(ProtocolOAuth2),
		AccessToken:  c.AccessToken,
		RefreshToken: c.RefreshToken,
		ExpiresAt:    c.ExpiresAt,
		TokenType:    c.TokenType,
		Scope:        c.Scope,
		WebhookURL:   c.Settings.WebhookURL,
	})
}

func (c *SlackConnection) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Type         string    `json:"type"`
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token,omitempty"`
		ExpiresAt    time.Time `json:"expires_at"`
		TokenType    string    `json:"token_type"`
		Scope        string    `json:"scope,omitempty"`
		WebhookURL   string    `json:"webhook_url,omitempty"`
	}{}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	c.OAuth2Connection = OAuth2Connection{
		AccessToken:  aux.AccessToken,
		RefreshToken: aux.RefreshToken,
		ExpiresAt:    aux.ExpiresAt,
		TokenType:    aux.TokenType,
		Scope:        aux.Scope,
	}
	c.Settings.WebhookURL = aux.WebhookURL

	return nil
}

func ParseSlackTokenResponse(body []byte, oauth2Conn OAuth2Connection, organizationID gid.GID) (*SlackConnection, *gid.GID, error) {
	var slackResponse SlackTokenResponse
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&slackResponse); err != nil {
		return nil, nil, fmt.Errorf("cannot decode Slack token response: %w", err)
	}

	if slackResponse.Error != "" {
		return nil, nil, fmt.Errorf("cannot complete Slack OAuth2 flow: %s", slackResponse.Error)
	}
	if !slackResponse.Ok {
		return nil, nil, fmt.Errorf("cannot complete Slack OAuth2 flow: ok=false")
	}
	if oauth2Conn.AccessToken == "" {
		return nil, nil, fmt.Errorf("cannot complete Slack OAuth2 flow: missing access token")
	}

	settings := SlackSettings{}
	if slackResponse.IncomingWebhook != nil {
		settings.WebhookURL = slackResponse.IncomingWebhook.URL
		settings.Channel = slackResponse.IncomingWebhook.Channel
		settings.ChannelID = slackResponse.IncomingWebhook.ChannelID
	}

	return &SlackConnection{
		OAuth2Connection: oauth2Conn,
		Settings:         settings,
	}, &organizationID, nil
}
