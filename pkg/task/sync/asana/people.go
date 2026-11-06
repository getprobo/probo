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

package asana

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	client "go.probo.inc/probo/pkg/task/sync/asana/client"
	"go.probo.inc/probo/pkg/task/sync/shared"
)

func (s *Service) asanaAssigneeGID(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	assignedToID *gid.GID,
) (*string, bool) {
	if assignedToID == nil {
		return nil, true
	}

	var emails []string

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			profile := &coredata.MembershipProfile{}
			if err := profile.LoadByID(ctx, conn, scope, *assignedToID); err != nil {
				return err
			}

			if profile.EmailAddress != mail.Nil {
				emails = append(emails, profile.EmailAddress.String())
			}

			for _, extra := range profile.AdditionalEmailAddresses {
				if extra != mail.Nil {
					emails = append(emails, extra.String())
				}
			}

			return nil
		},
	)
	if err != nil {
		if s.logger != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
			s.logger.WarnCtx(ctx, "cannot load task assignee for Asana sync", log.Error(err))
		}

		return nil, false
	}

	for _, email := range emails {
		id, err := account.client.UserGIDByEmail(ctx, account.settings.WorkspaceGID, email)
		if err != nil {
			if s.logger != nil {
				s.logger.WarnCtx(ctx, asanaUserLookupMessage(err))
			}

			return nil, false
		}

		if id != "" {
			return &id, true
		}
	}

	return nil, true
}

func asanaUserLookupMessage(err error) string {
	apiErr, ok := errors.AsType[*client.APIError](err)
	if !ok || apiErr == nil {
		return "asana user lookup failed"
	}

	return fmt.Sprintf("asana user lookup failed: status %d", apiErr.Status)
}

func (s *Service) applyAsanaAssignee(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	task *coredata.Task,
	email string,
) (bool, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		if task.AssignedToID == nil {
			return false, nil
		}

		task.AssignedToID = nil

		return true, nil
	}

	addr, err := mail.ParseAddr(email)
	if err != nil {
		return false, nil
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByOrganizationIDAndEmail(ctx, tx, scope, task.OrganizationID, addr); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return false, nil
		}

		return false, err
	}

	if task.AssignedToID != nil && *task.AssignedToID == profile.ID {
		return false, nil
	}

	task.AssignedToID = &profile.ID

	return true, nil
}

func (s *Service) asanaAuthorName(ctx context.Context, scope coredata.Scoper, ownerID *gid.GID) string {
	var name string

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			loaded, err := shared.CommentAuthorName(ctx, conn, scope, ownerID)
			if err != nil {
				return err
			}

			name = loaded

			return nil
		},
	)
	if err != nil {
		if s.logger != nil {
			s.logger.WarnCtx(ctx, "cannot load Asana comment author", log.Error(err))
		}

		return ""
	}

	return name
}
