// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	TaskService struct {
		svc *Service
	}

	CreateTaskRequest struct {
		OrganizationID gid.GID
		MeasureID      *gid.GID
		Name           string
		Description    *string
		Priority       coredata.TaskPriority
		TimeEstimate   *time.Duration
		AssignedToID   *gid.GID
		Deadline       *time.Time
	}

	UpdateTaskRequest struct {
		TaskID       gid.GID
		Name         *string
		Description  **string
		State        *coredata.TaskState
		Priority     *coredata.TaskPriority
		TimeEstimate **time.Duration
		Deadline     **time.Time
		AssignedToID **gid.GID
		MeasureID    **gid.GID
		Rank         *int
	}
)

func (ctr *CreateTaskRequest) Validate() error {
	v := validator.New()

	v.Check(ctr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(ctr.MeasureID, "measure_id", validator.GID(coredata.MeasureEntityType))
	v.Check(ctr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(ctr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(ctr.Priority, "priority", validator.Required(), validator.OneOfSlice(coredata.TaskPriorities()))
	v.Check(ctr.TimeEstimate, "time_estimate", validator.RangeDuration(0, 1000*time.Hour))
	v.Check(ctr.AssignedToID, "assigned_to_id", validator.GID(coredata.MembershipProfileEntityType))

	return v.Error()
}

func (utr *UpdateTaskRequest) Validate() error {
	v := validator.New()

	v.Check(utr.TaskID, "task_id", validator.Required(), validator.GID(coredata.TaskEntityType))
	v.Check(utr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(utr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(utr.Priority, "priority", validator.OneOfSlice(coredata.TaskPriorities()))
	v.Check(utr.TimeEstimate, "time_estimate", validator.RangeDuration(0, 1000*time.Hour))
	v.Check(utr.State, "state", validator.OneOfSlice(coredata.TaskStates()))
	v.Check(utr.AssignedToID, "assigned_to_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(utr.MeasureID, "measure_id", validator.GID(coredata.MeasureEntityType))
	v.Check(utr.Rank, "rank", validator.Min(1))

	return v.Error()
}

func (s TaskService) Create(
	ctx context.Context, scope coredata.Scoper,
	req CreateTaskRequest,
) (*coredata.Task, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	taskID := gid.New(scope.GetTenantID(), coredata.TaskEntityType)

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
		Priority:       req.Priority,
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
		func(ctx context.Context, conn pg.Tx) error {
			if req.MeasureID != nil {
				measure := &coredata.Measure{}
				if err := measure.LoadByID(ctx, conn, scope, *req.MeasureID); err != nil {
					return fmt.Errorf("cannot load measure: %w", err)
				}
			}

			if req.AssignedToID != nil {
				assignee := &coredata.MembershipProfile{}
				if err := assignee.LoadByID(ctx, conn, scope, *req.AssignedToID); err != nil {
					return fmt.Errorf("cannot load assignee profile: %w", err)
				}
			}

			if err := task.Insert(ctx, conn, scope); err != nil {
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
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
) (*coredata.Task, error) {
	task := &coredata.Task{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return task.LoadByID(ctx, conn, scope, taskID)
		},
	)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s TaskService) GetByIDs(
	ctx context.Context, scope coredata.Scoper,
	taskIDs ...gid.GID,
) (coredata.Tasks, error) {
	var tasks coredata.Tasks

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := tasks.LoadByIDs(
				ctx,
				conn,
				scope,
				taskIDs,
			); err != nil {
				return fmt.Errorf("cannot load tasks by ids: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s TaskService) Assign(
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
	assignedToID gid.GID,
) (*coredata.Task, error) {
	task := &coredata.Task{ID: taskID}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := task.LoadByID(ctx, conn, scope, taskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", taskID, err)
			}

			assignee := &coredata.MembershipProfile{}
			if err := assignee.LoadByID(ctx, conn, scope, assignedToID); err != nil {
				return fmt.Errorf("cannot load assignee profile: %w", err)
			}

			task.AssignedToID = &assignedToID
			task.UpdatedAt = time.Now()

			if err := task.Update(ctx, conn, scope); err != nil {
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
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
) (*coredata.Task, error) {
	task := &coredata.Task{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := task.LoadByID(ctx, conn, scope, taskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", taskID, err)
			}

			task.AssignedToID = nil
			task.UpdatedAt = time.Now()

			if err := task.Update(ctx, conn, scope); err != nil {
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
	ctx context.Context, scope coredata.Scoper,
	req UpdateTaskRequest,
) (*coredata.Task, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	task := &coredata.Task{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := task.LoadByID(ctx, conn, scope, req.TaskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", req.TaskID, err)
			}

			oldState := task.State
			oldPriority := task.Priority

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
					if err := assignee.LoadByID(ctx, conn, scope, **req.AssignedToID); err != nil {
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
					if err := measure.LoadByID(ctx, conn, scope, **req.MeasureID); err != nil {
						return fmt.Errorf("cannot load measure: %w", err)
					}

					task.MeasureID = *req.MeasureID
				}
			}

			if req.Priority != nil {
				task.Priority = *req.Priority
			}

			task.UpdatedAt = time.Now()

			targetRank := req.Rank
			priorityChanged := task.Priority != oldPriority
			stateChanged := task.State != oldState

			if priorityChanged || stateChanged {
				if err := task.NextRankForStatePriority(ctx, conn, scope); err != nil {
					return fmt.Errorf("cannot get next rank: %w", err)
				}
			}

			if err := task.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update task: %w", err)
			}

			if targetRank != nil {
				task.Rank = *targetRank
				if err := task.UpdateRank(ctx, conn, scope); err != nil {
					return fmt.Errorf("cannot update task rank: %w", err)
				}
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
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
) error {
	task := &coredata.Task{ID: taskID}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			return task.Delete(ctx, conn, scope)
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (s TaskService) CountForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			tasks := coredata.Tasks{}

			count, err = tasks.CountByOrganizationID(ctx, conn, scope, organizationID)
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
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.TaskOrderField],
) (*page.Page[*coredata.Task, coredata.TaskOrderField], error) {
	var tasks coredata.Tasks

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return tasks.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(tasks, cursor), nil
}

func (s TaskService) CountForMeasureID(
	ctx context.Context, scope coredata.Scoper,
	measureID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			tasks := coredata.Tasks{}

			count, err = tasks.CountByMeasureID(ctx, conn, scope, measureID)
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
	ctx context.Context, scope coredata.Scoper,
	measureID gid.GID,
	cursor *page.Cursor[coredata.TaskOrderField],
) (*page.Page[*coredata.Task, coredata.TaskOrderField], error) {
	var tasks coredata.Tasks

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return tasks.LoadByMeasureID(
				ctx,
				conn,
				scope,
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
