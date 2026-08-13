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
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

func TestParseSlashCommand(t *testing.T) {
	t.Parallel()

	cmd := ParseSlashCommand(
		url.Values{
			"command":      []string{" /probot "},
			"text":         []string{" bind "},
			"user_id":      []string{"U123"},
			"user_name":    []string{" ada "},
			"team_id":      []string{"T123"},
			"team_domain":  []string{" acme "},
			"response_url": []string{" https://hooks.slack.com/commands/T123/1/abc "},
		},
	)

	assert.Equal(t, SlashCommandName, cmd.Command)
	assert.Equal(t, "bind", cmd.Text)
	assert.Equal(t, "U123", cmd.UserID)
	assert.Equal(t, "ada", cmd.UserName)
	assert.Equal(t, "T123", cmd.TeamID)
	assert.Equal(t, "acme", cmd.TeamDomain)
	assert.Equal(t, "https://hooks.slack.com/commands/T123/1/abc", cmd.ResponseURL)
}

func TestHandler_HandleSlashCommand(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	bindURL := "https://console.example.com/me/probot/bind?token=test"

	t.Run(
		"returns an ephemeral bind link",
		func(t *testing.T) {
			t.Parallel()

			bindings := &stubBindingGate{bindURL: bindURL}
			handler := &Handler{bindings: bindings, logger: log.NewLogger()}
			response := handler.HandleSlashCommand(
				ctx,
				SlashCommand{
					Command:    SlashCommandName,
					Text:       "bind",
					UserID:     "U456",
					UserName:   "ada",
					TeamID:     "T789",
					TeamDomain: "acme",
				},
			)

			assert.Equal(t, SlashResponseTypeEphemeral, response.ResponseType)
			assert.Equal(t, bindSlashFallbackText, response.Text)
			assert.Equal(t, bindRequiredBlocks(bindURL), response.Blocks)
			assert.Equal(
				t,
				identitybinding.Subject{
					Provider:           ProviderName,
					ExternalTenantID:   "T789",
					ExternalUserID:     "U456",
					ExternalTenantName: "acme",
					ExternalUserName:   "@ada",
				},
				bindings.bindSubject,
			)
			assert.NotEqual(t, "in_channel", response.ResponseType)
		},
	)

	t.Run(
		"treats empty text as bind",
		func(t *testing.T) {
			t.Parallel()

			bindings := &stubBindingGate{bindURL: bindURL}
			handler := &Handler{bindings: bindings, logger: log.NewLogger()}
			response := handler.HandleSlashCommand(
				ctx,
				SlashCommand{
					Command: SlashCommandName,
					UserID:  "U456",
					TeamID:  "T789",
				},
			)

			assert.Equal(t, SlashResponseTypeEphemeral, response.ResponseType)
			require.NotEmpty(t, response.Blocks)
			assert.Equal(t, IdentitySubject("T789", "U456"), bindings.bindSubject)
		},
	)

	t.Run(
		"does not issue a token when already linked",
		func(t *testing.T) {
			t.Parallel()

			bindings := &stubBindingGate{
				binding: &identitybinding.Binding{
					IdentityID: gid.New(gid.NilTenant, coredata.IdentityEntityType),
				},
				bindURL: bindURL,
			}
			handler := &Handler{bindings: bindings, logger: log.NewLogger()}
			response := handler.HandleSlashCommand(
				ctx,
				SlashCommand{
					Command: SlashCommandName,
					Text:    "bind",
					UserID:  "U456",
					TeamID:  "T789",
				},
			)

			assert.Equal(t, SlashResponseTypeEphemeral, response.ResponseType)
			assert.Equal(t, bindSlashAlreadyLinkedText, response.Text)
			assert.Nil(t, response.Blocks)
			assert.Empty(t, bindings.bindSubject.ExternalUserID)
		},
	)

	t.Run(
		"returns usage for unknown arguments",
		func(t *testing.T) {
			t.Parallel()

			bindings := &stubBindingGate{bindURL: bindURL}
			handler := &Handler{bindings: bindings, logger: log.NewLogger()}
			response := handler.HandleSlashCommand(
				ctx,
				SlashCommand{
					Command: SlashCommandName,
					Text:    "help",
					UserID:  "U456",
					TeamID:  "T789",
				},
			)

			assert.Equal(t, SlashResponseTypeEphemeral, response.ResponseType)
			assert.Equal(t, bindSlashUsageText, response.Text)
			assert.Nil(t, response.Blocks)
			assert.Empty(t, bindings.bindSubject.ExternalUserID)
		},
	)

	t.Run(
		"returns unavailable when the user or team is missing",
		func(t *testing.T) {
			t.Parallel()

			bindings := &stubBindingGate{bindURL: bindURL}
			handler := &Handler{bindings: bindings, logger: log.NewLogger()}
			response := handler.HandleSlashCommand(
				ctx,
				SlashCommand{
					Command: SlashCommandName,
					Text:    "bind",
					TeamID:  "T789",
				},
			)

			assert.Equal(t, SlashResponseTypeEphemeral, response.ResponseType)
			assert.Equal(t, bindSlashUnavailableText, response.Text)
			assert.Empty(t, bindings.bindSubject.ExternalUserID)
		},
	)

	t.Run(
		"returns unavailable when the workspace is not installed",
		func(t *testing.T) {
			t.Parallel()

			bindings := &stubBindingGate{bindURL: bindURL}
			handler := &Handler{
				bindings:      bindings,
				installations: stubMissingInstallations{},
				logger:        log.NewLogger(),
			}
			response := handler.HandleSlashCommand(
				ctx,
				SlashCommand{
					Command: SlashCommandName,
					Text:    "bind",
					UserID:  "U456",
					TeamID:  "T789",
				},
			)

			assert.Equal(t, SlashResponseTypeEphemeral, response.ResponseType)
			assert.Equal(t, bindSlashUnavailableText, response.Text)
			assert.Empty(t, bindings.bindSubject.ExternalUserID)
		},
	)

	t.Run(
		"remembers the slash command response url",
		func(t *testing.T) {
			t.Parallel()

			bindings := &stubBindingGate{bindURL: bindURL}
			prompts := &recordingBindPromptStore{}
			handler := &Handler{
				bindings:    bindings,
				bindPrompts: prompts,
				logger:      log.NewLogger(),
			}
			response := handler.HandleSlashCommand(
				ctx,
				SlashCommand{
					Command:     SlashCommandName,
					Text:        "bind",
					UserID:      "U456",
					TeamID:      "T789",
					ResponseURL: "https://hooks.slack.com/commands/T789/1/abc",
				},
			)

			assert.Equal(t, SlashResponseTypeEphemeral, response.ResponseType)
			assert.Equal(t, "T789", prompts.teamID)
			assert.Equal(t, "U456", prompts.userID)
			assert.Equal(t, "https://hooks.slack.com/commands/T789/1/abc", prompts.responseURL)
		},
	)
}

type recordingBindPromptStore struct {
	teamID      string
	userID      string
	responseURL string
}

func (s *recordingBindPromptStore) RememberResponseURL(
	_ context.Context,
	teamID string,
	userID string,
	responseURL string,
) error {
	s.teamID = teamID
	s.userID = userID
	s.responseURL = responseURL

	return nil
}

type stubMissingInstallations struct{}

func (stubMissingInstallations) ClientByTeamID(
	context.Context,
	string,
) (*Client, *coredata.SlackbotInstallation, error) {
	return nil, nil, ErrSlackbotNotInstalled
}

func (stubMissingInstallations) DisableByTeamID(context.Context, string) error {
	return nil
}
