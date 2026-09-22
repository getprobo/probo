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

package tasksync

import (
	"context"
	"errors"
	"strings"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

func inboundAssigneeEmail(data *linear.IssueWebhookData) (string, bool) {
	if data == nil || !data.Has("assignee") {
		return "", false
	}

	if data.Assignee == nil {
		return "", false
	}

	email := strings.TrimSpace(data.Assignee.Email)
	if email == "" {
		return "", false
	}

	return email, true
}

func applyInboundAssignee(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	task *coredata.Task,
	data *linear.IssueWebhookData,
) (bool, error) {
	email, ok := inboundAssigneeEmail(data)
	if !ok {
		return false, nil
	}

	addr, err := mail.ParseAddr(email)
	if err != nil {
		return false, nil
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByOrganizationIDAndEmail(
		ctx,
		tx,
		scope,
		organizationID,
		addr,
	); err != nil {
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

func (s *Service) linearAssigneeID(
	ctx context.Context,
	scope coredata.Scoper,
	client *linear.Client,
	assignedToID *gid.GID,
) *string {
	if assignedToID == nil {
		return nil
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
			s.logger.WarnCtx(
				ctx,
				"cannot load task assignee for Linear sync",
				log.Error(err),
			)
		}

		return nil
	}

	for _, email := range emails {
		id, err := client.UserIDByEmail(ctx, email)
		if err != nil {
			if s.logger != nil {
				s.logger.WarnCtx(
					ctx,
					"cannot look up Linear user by email",
					log.Error(err),
				)
			}

			return nil
		}

		if id != "" {
			return &id
		}
	}

	return nil
}
