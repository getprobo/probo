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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

func TestService_HandleSlashCommand(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	bindURL := "https://console.example.com/organizations/org/employee/bind?token=test"

	t.Run(
		"returns an ephemeral login link",
		func(t *testing.T) {
			t.Parallel()

			bindings := &stubBindingGate{bindURL: bindURL}
			handler := &Service{bindings: bindings, logger: log.NewLogger()}
			response := handler.HandleSlashCommand(
				ctx,
				SlashCommand{
					Command:    SlashCommandName,
					Text:       "login",
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
		"treats empty text as login",
		func(t *testing.T) {
			t.Parallel()

			bindings := &stubBindingGate{bindURL: bindURL}
			handler := &Service{bindings: bindings, logger: log.NewLogger()}
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
		"accepts bind as a login alias",
		func(t *testing.T) {
			t.Parallel()

			bindings := &stubBindingGate{bindURL: bindURL}
			handler := &Service{bindings: bindings, logger: log.NewLogger()}
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
			handler := &Service{bindings: bindings, logger: log.NewLogger()}
			response := handler.HandleSlashCommand(
				ctx,
				SlashCommand{
					Command: SlashCommandName,
					Text:    "login",
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
			handler := &Service{bindings: bindings, logger: log.NewLogger()}
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
			handler := &Service{bindings: bindings, logger: log.NewLogger()}
			response := handler.HandleSlashCommand(
				ctx,
				SlashCommand{
					Command: SlashCommandName,
					Text:    "login",
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
			handler := &Service{
				bindings:      bindings,
				installations: newTestInstallationService(t, test.PGClient(t), ""),
				logger:        log.NewLogger(),
			}
			response := handler.HandleSlashCommand(
				ctx,
				SlashCommand{
					Command: SlashCommandName,
					Text:    "login",
					UserID:  "U456",
					TeamID:  uniqueSlackTeamID(t),
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

			pgClient := test.PGClient(t)
			bindings := &stubBindingGate{bindURL: bindURL}
			prompts := NewBindPromptService(pgClient, testEncryptionKey(), log.NewLogger())
			handler := &Service{
				bindings:    bindings,
				bindPrompts: prompts,
				logger:      log.NewLogger(),
			}
			teamID := uniqueSlackTeamID(t)
			userID := "U-" + gid.New(gid.NilTenant, coredata.IdentityEntityType).String()
			responseURL := "https://hooks.slack.com/commands/" + teamID + "/1/abc"
			t.Cleanup(
				func() {
					_ = pgClient.WithConn(
						context.Background(),
						func(ctx context.Context, conn pg.Querier) error {
							_, err := conn.Exec(
								ctx,
								`DELETE FROM slackbot_bind_callbacks WHERE team_id = $1 AND user_id = $2`,
								teamID,
								userID,
							)

							return err
						},
					)
				},
			)
			response := handler.HandleSlashCommand(
				ctx,
				SlashCommand{
					Command:     SlashCommandName,
					Text:        "login",
					UserID:      userID,
					TeamID:      teamID,
					ResponseURL: responseURL,
				},
			)

			assert.Equal(t, SlashResponseTypeEphemeral, response.ResponseType)

			var callback coredata.SlackbotBindCallback
			err := pgClient.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					return callback.LoadByTeamAndUser(ctx, conn, teamID, userID)
				},
			)
			if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
				t.Skipf("slackbot_bind_callbacks is unavailable in the test database: %v", err)
			}
			require.NoError(t, err)

			storedURL, err := cipher.Decrypt(callback.EncryptedResponseURL, testEncryptionKey())
			require.NoError(t, err)
			assert.Equal(t, responseURL, string(storedURL))
		},
	)
}
