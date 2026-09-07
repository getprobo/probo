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

package probo

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/validator"
)

type (
	TaskCommentService struct {
		svc *Service
	}

	CreateTaskCommentRequest struct {
		TaskID     gid.GID
		OwnerID    *gid.GID
		IdentityID gid.GID
		Content    string
	}

	UpdateTaskCommentRequest struct {
		ID      gid.GID
		OwnerID **gid.GID
		Content **string
	}
)

func (req *CreateTaskCommentRequest) Validate() error {
	v := validator.New()

	v.Check(req.TaskID, "task_id", validator.Required(), validator.GID(coredata.TaskEntityType))
	v.Check(req.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(req.IdentityID, "identity_id", validator.Required(), validator.GID(coredata.IdentityEntityType))
	v.Check(req.Content, "content", validator.MaxLen(richTextMaxJSONBytes))

	if len(req.Content) <= richTextMaxJSONBytes {
		v.Check(
			req.Content,
			"content",
			validator.HasVisibleRichText(),
			validator.ProseMirrorDocumentContent(),
			validator.ProseMirrorDocumentMaxTextLength(ContentMaxLength),
		)
	}

	return v.Error()
}

func (req *UpdateTaskCommentRequest) Validate() error {
	v := validator.New()

	v.Check(req.ID, "id", validator.Required(), validator.GID(coredata.TaskCommentEntityType))

	if req.OwnerID != nil {
		v.Check(*req.OwnerID, "owner_id", validator.Required(), validator.GID(coredata.MembershipProfileEntityType))
	}

	if req.Content != nil {
		v.Check(
			req.Content,
			"content",
			validator.MaxLen(richTextMaxJSONBytes),
			validator.ProseMirrorDocumentContent(),
			validator.ProseMirrorDocumentMaxTextLength(ContentMaxLength),
		)
	}

	return v.Error()
}

func (s TaskCommentService) Get(
	ctx context.Context, scope coredata.Scoper,
	taskCommentID gid.GID,
) (*coredata.TaskComment, error) {
	taskComment := &coredata.TaskComment{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := taskComment.LoadByID(ctx, conn, scope, taskCommentID); err != nil {
				return fmt.Errorf("cannot load task comment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return taskComment, nil
}

func (s TaskCommentService) ListForTaskID(
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
	cursor *page.Cursor[coredata.TaskCommentOrderField],
) (*page.Page[*coredata.TaskComment, coredata.TaskCommentOrderField], error) {
	var taskComments coredata.TaskComments

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := taskComments.LoadByTaskID(ctx, conn, scope, taskID, cursor); err != nil {
				return fmt.Errorf("cannot load task comments: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(taskComments, cursor), nil
}

func (s TaskCommentService) CountForTaskID(
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			taskComments := coredata.TaskComments{}

			count, err = taskComments.CountByTaskID(ctx, conn, scope, taskID)
			if err != nil {
				return fmt.Errorf("cannot count task comments: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s TaskCommentService) Create(
	ctx context.Context, scope coredata.Scoper,
	req CreateTaskCommentRequest,
) (*coredata.TaskComment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	content, err := prosemirror.DefaultDocumentJSON(&req.Content)
	if err != nil {
		return nil, fmt.Errorf("cannot sanitize task comment content: %w", err)
	}

	now := time.Now()
	taskComment := &coredata.TaskComment{
		ID:        gid.New(scope.GetTenantID(), coredata.TaskCommentEntityType),
		TaskID:    req.TaskID,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			task := &coredata.Task{}
			if err := task.LoadByID(ctx, conn, scope, req.TaskID); err != nil {
				return fmt.Errorf("cannot load task: %w", err)
			}

			taskComment.OrganizationID = task.OrganizationID

			owner := &coredata.MembershipProfile{}
			if req.OwnerID != nil {
				if err := owner.LoadByID(ctx, conn, scope, *req.OwnerID); err != nil {
					return fmt.Errorf("cannot load owner profile: %w", err)
				}
			} else {
				if err := owner.LoadByIdentityIDAndOrganizationID(
					ctx,
					conn,
					scope,
					req.IdentityID,
					task.OrganizationID,
				); err != nil {
					return fmt.Errorf("cannot load owner profile: %w", err)
				}
			}

			if owner.OrganizationID != task.OrganizationID {
				return fmt.Errorf("cannot load owner profile: %w", coredata.ErrResourceNotFound)
			}

			taskComment.OwnerID = owner.ID

			if err := taskComment.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert task comment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return taskComment, nil
}

func (s TaskCommentService) Update(
	ctx context.Context, scope coredata.Scoper,
	req UpdateTaskCommentRequest,
) (*coredata.TaskComment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	taskComment := &coredata.TaskComment{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := taskComment.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load task comment: %w", err)
			}

			if req.OwnerID != nil {
				owner := &coredata.MembershipProfile{}
				if err := owner.LoadByID(ctx, conn, scope, **req.OwnerID); err != nil {
					return fmt.Errorf("cannot load owner profile: %w", err)
				}

				if owner.OrganizationID != taskComment.OrganizationID {
					return fmt.Errorf("cannot load owner profile: %w", coredata.ErrResourceNotFound)
				}

				taskComment.OwnerID = **req.OwnerID
			}

			if req.Content != nil {
				content, err := prosemirror.DefaultDocumentJSON(*req.Content)
				if err != nil {
					return fmt.Errorf("cannot sanitize task comment content: %w", err)
				}

				taskComment.Content = content
			}

			taskComment.UpdatedAt = time.Now()

			if err := taskComment.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update task comment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return taskComment, nil
}

func (s TaskCommentService) Delete(
	ctx context.Context, scope coredata.Scoper,
	taskCommentID gid.GID,
) error {
	taskComment := coredata.TaskComment{ID: taskCommentID}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := taskComment.LoadByID(ctx, conn, scope, taskCommentID); err != nil {
				return fmt.Errorf("cannot load task comment: %w", err)
			}

			if err := taskComment.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot delete task comment: %w", err)
			}

			return nil
		},
	)
}
