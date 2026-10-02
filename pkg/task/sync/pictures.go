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
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

// UploadPictureToLinkedIssue uploads a picture onto the task's Linear issue.
// A nil result means the task is not linked to Linear.
func (s *Service) UploadPictureToLinkedIssue(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
	filename string,
	contentType string,
	body []byte,
) (*linear.UploadedPicture, error) {
	client, issueID, err := s.linearClientForLinkedTask(ctx, scope, taskID)
	if err != nil {
		return nil, err
	}

	if client == nil {
		return nil, nil
	}

	uploaded, err := client.UploadPicture(ctx, issueID, filename, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("cannot upload picture to Linear issue: %w", err)
	}

	return uploaded, nil
}

// RemoveLinkedPicture deletes a picture attachment from the task's Linear issue.
// Missing issues and attachments are ignored so a local delete can finish.
func (s *Service) RemoveLinkedPicture(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
	attachmentID string,
) error {
	if attachmentID == "" {
		return nil
	}

	client, _, err := s.linearClientForLinkedTask(ctx, scope, taskID)
	if err != nil {
		return err
	}

	if client == nil {
		return nil
	}

	if err := client.DeleteAttachment(ctx, attachmentID); err != nil && !linear.IsEntityMissing(err) {
		return fmt.Errorf("cannot delete Linear picture attachment: %w", err)
	}

	return nil
}

func (s *Service) linearClientForLinkedTask(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
) (*linear.Client, string, error) {
	var link coredata.TaskExternalLink

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := link.LoadByTaskID(ctx, conn, scope, taskID); err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, "", nil
		}

		return nil, "", err
	}

	if link.Provider != coredata.ConnectorProviderLinearSync {
		return nil, "", nil
	}

	var dbConnector *coredata.Connector

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			dbConnector = &coredata.Connector{}
			if err := dbConnector.LoadByID(ctx, conn, scope, link.ConnectorID, s.encryptionKey); err != nil {
				return fmt.Errorf("cannot load Linear connector: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, "", err
	}

	client, _, err := s.linearClientForConnector(ctx, scope, dbConnector)
	if err != nil {
		return nil, "", fmt.Errorf("cannot create Linear client: %w", err)
	}

	return client, link.ExternalID, nil
}
