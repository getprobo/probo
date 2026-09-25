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
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

const (
	slashCommandLoginArg       = "login"
	slashCommandLoginAliasArg  = "bind"
	SlashResponseTypeEphemeral = "ephemeral"
)

type (
	SlashCommand struct {
		Command     string
		Text        string
		UserID      string
		UserName    string
		TeamID      string
		TeamDomain  string
		ResponseURL string
	}

	SlashCommandResponse struct {
		ResponseType string `json:"response_type"`
		Text         string `json:"text"`
		Blocks       []any  `json:"blocks,omitempty"`
	}
)

func ephemeralSlashResponse(text string, blocks []any) SlashCommandResponse {
	return SlashCommandResponse{
		ResponseType: SlashResponseTypeEphemeral,
		Text:         text,
		Blocks:       blocks,
	}
}

func (s *Service) HandleSlashCommand(
	ctx context.Context,
	cmd SlashCommand,
) SlashCommandResponse {
	if cmd.UserID == "" || cmd.TeamID == "" {
		return ephemeralSlashResponse(bindSlashUnavailableText, nil)
	}

	if cmd.Command != SlashCommandName {
		return ephemeralSlashResponse(bindSlashUsageText, nil)
	}

	arg := strings.ToLower(cmd.Text)
	if arg != "" && arg != slashCommandLoginArg && arg != slashCommandLoginAliasArg {
		return ephemeralSlashResponse(bindSlashUsageText, nil)
	}

	var organizationID gid.GID

	if s.installations != nil {
		var err error
		if _, organizationID, _, err = s.clientForTeam(ctx, cmd.TeamID); err != nil {
			return ephemeralSlashResponse(bindSlashUnavailableText, nil)
		}
	}

	if s.bindings == nil {
		return ephemeralSlashResponse(bindSlashUnavailableText, nil)
	}

	subject := identitySubjectFromSlashCommand(cmd)

	binding, err := s.bindings.Lookup(ctx, subject)
	if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return ephemeralSlashResponse(bindSlashFailedText, nil)
	}

	if binding != nil {
		return ephemeralSlashResponse(bindSlashAlreadyLinkedText, nil)
	}

	bindURL, err := s.bindings.BindURL(ctx, subject, organizationID)
	if err != nil {
		return ephemeralSlashResponse(bindSlashFailedText, nil)
	}

	if s.bindPrompts != nil && cmd.ResponseURL != "" {
		if err := s.bindPrompts.RememberResponseURL(
			ctx,
			cmd.TeamID,
			cmd.UserID,
			cmd.ResponseURL,
		); err != nil && s.logger != nil {
			s.logger.ErrorCtx(ctx, "cannot remember Slack bind prompt", log.Error(err))
		}
	}

	return ephemeralSlashResponse(bindSlashFallbackText, bindRequiredBlocks(bindURL))
}

func identitySubjectFromSlashCommand(cmd SlashCommand) identitybinding.Subject {
	subject := IdentitySubject(cmd.TeamID, cmd.UserID)

	subject.ExternalTenantName = cmd.TeamDomain
	if cmd.UserName != "" {
		subject.ExternalUserName = "@" + strings.TrimPrefix(cmd.UserName, "@")
	}

	return subject
}
