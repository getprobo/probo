// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package probo

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	TaskService struct {
		svc *TenantService
	}

	CreateTaskRequest struct {
		OrganizationID gid.GID
		MeasureID      *gid.GID
		Name           string
		Description    *string
		TimeEstimate   *time.Duration
		AssignedToID   *gid.GID
		Deadline       *time.Time
	}

	UpdateTaskRequest struct {
		TaskID       gid.GID
		Name         *string
		Description  **string
		State        *coredata.TaskState
		TimeEstimate **time.Duration
		Deadline     **time.Time
		AssignedToID **gid.GID
		MeasureID    **gid.GID
	}
)

func (ctr *CreateTaskRequest) Validate() error {
	v := validator.New()

	v.Check(ctr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(ctr.MeasureID, "measure_id", validator.GID(coredata.MeasureEntityType))
	v.Check(ctr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(ctr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(ctr.TimeEstimate, "time_estimate", validator.RangeDuration(0, 1000*time.Hour))
	v.Check(ctr.AssignedToID, "assigned_to_id", validator.GID(coredata.MembershipProfileEntityType))

	return v.Error()
}

func (utr *UpdateTaskRequest) Validate() error {
	v := validator.New()

	v.Check(utr.TaskID, "task_id", validator.Required(), validator.GID(coredata.TaskEntityType))
	v.Check(utr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(utr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(utr.TimeEstimate, "time_estimate", validator.RangeDuration(0, 1000*time.Hour))
	v.Check(utr.State, "state", validator.OneOfSlice(coredata.TaskStates()))
	v.Check(utr.AssignedToID, "assigned_to_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(utr.MeasureID, "measure_id", validator.GID(coredata.MeasureEntityType))

	return v.Error()
}

func (s TaskService) Create(
	ctx context.Context,
	req CreateTaskRequest,
) (*coredata.Task, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	taskID := gid.New(s.svc.scope.GetTenantID(), coredata.TaskEntityType)

	referenceID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("cannot generate reference id: %w", err)
	}

	task := &coredata.Task{
		ID:             taskID,
		OrganizationID: req.OrganizationID,
		MeasureID:      req.MeasureID,
		Name:           req.Name,
		Description:    req.Description,
		TimeEstimate:   req.TimeEstimate,
		AssignedToID:   req.AssignedToID,
		Deadline:       req.Deadline,
		State:          coredata.TaskStateTodo,
		ReferenceID:    "custom-task-" + referenceID.String(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if req.MeasureID != nil {
				measure := &coredata.Measure{}
				if err := measure.LoadByID(ctx, conn, s.svc.scope, *req.MeasureID); err != nil {
					return fmt.Errorf("cannot load measure: %w", err)
				}
			}

			if req.AssignedToID != nil {
				assignee := &coredata.MembershipProfile{}
				if err := assignee.LoadByID(ctx, conn, s.svc.scope, *req.AssignedToID); err != nil {
					return fmt.Errorf("cannot load assignee profile: %w", err)
				}
			}

			if err := task.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert task: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create task: %w", err)
	}

	return task, nil
}

func (s TaskService) Get(
	ctx context.Context,
	taskID gid.GID,
) (*coredata.Task, error) {
	task := &coredata.Task{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return task.LoadByID(ctx, conn, s.svc.scope, taskID)
		},
	)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s TaskService) Assign(
	ctx context.Context,
	taskID gid.GID,
	assignedToID gid.GID,
) (*coredata.Task, error) {
	task := &coredata.Task{ID: taskID}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := task.LoadByID(ctx, conn, s.svc.scope, taskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", taskID, err)
			}

			assignee := &coredata.MembershipProfile{}
			if err := assignee.LoadByID(ctx, conn, s.svc.scope, assignedToID); err != nil {
				return fmt.Errorf("cannot load assignee profile: %w", err)
			}

			task.AssignedToID = &assignedToID
			task.UpdatedAt = time.Now()

			if err := task.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot assign task %q to %q: %w", taskID, assignedToID, err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s TaskService) Unassign(
	ctx context.Context,
	taskID gid.GID,
) (*coredata.Task, error) {
	task := &coredata.Task{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := task.LoadByID(ctx, conn, s.svc.scope, taskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", taskID, err)
			}

			task.AssignedToID = nil
			task.UpdatedAt = time.Now()

			if err := task.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot unassign task %q: %w", taskID, err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s TaskService) Update(
	ctx context.Context,
	req UpdateTaskRequest,
) (*coredata.Task, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	task := &coredata.Task{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := task.LoadByID(ctx, conn, s.svc.scope, req.TaskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", req.TaskID, err)
			}

			if req.Name != nil {
				task.Name = *req.Name
			}

			if req.Description != nil {
				task.Description = *req.Description
			}

			if req.State != nil {
				task.State = *req.State
			}

			if req.TimeEstimate != nil {
				task.TimeEstimate = *req.TimeEstimate
			}

			if req.Deadline != nil {
				task.Deadline = *req.Deadline
			}

			if req.AssignedToID != nil {
				if *req.AssignedToID == nil {
					task.AssignedToID = nil
				} else {
					assignee := &coredata.MembershipProfile{}
					if err := assignee.LoadByID(ctx, conn, s.svc.scope, **req.AssignedToID); err != nil {
						return fmt.Errorf("cannot load assignee profile: %w", err)
					}
					task.AssignedToID = *req.AssignedToID
				}
			}

			if req.MeasureID != nil {
				if *req.MeasureID == nil {
					task.MeasureID = nil
				} else {
					measure := &coredata.Measure{}
					if err := measure.LoadByID(ctx, conn, s.svc.scope, **req.MeasureID); err != nil {
						return fmt.Errorf("cannot load measure: %w", err)
					}
					task.MeasureID = *req.MeasureID
				}
			}

			task.UpdatedAt = time.Now()

			if err := task.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update task: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s TaskService) Delete(
	ctx context.Context,
	taskID gid.GID,
) error {
	task := &coredata.Task{ID: taskID}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			return task.Delete(ctx, conn, s.svc.scope)
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (s TaskService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			tasks := coredata.Tasks{}
			count, err = tasks.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count tasks: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s TaskService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.TaskOrderField],
) (*page.Page[*coredata.Task, coredata.TaskOrderField], error) {
	var tasks coredata.Tasks

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return tasks.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(tasks, cursor), nil
}

func (s TaskService) CountForMeasureID(
	ctx context.Context,
	measureID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			tasks := coredata.Tasks{}
			count, err = tasks.CountByMeasureID(ctx, conn, s.svc.scope, measureID)
			if err != nil {
				return fmt.Errorf("cannot count tasks: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s TaskService) ListForMeasureID(
	ctx context.Context,
	measureID gid.GID,
	cursor *page.Cursor[coredata.TaskOrderField],
) (*page.Page[*coredata.Task, coredata.TaskOrderField], error) {
	var tasks coredata.Tasks

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return tasks.LoadByMeasureID(
				ctx,
				conn,
				s.svc.scope,
				measureID,
				cursor,
			)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(tasks, cursor), nil
}
