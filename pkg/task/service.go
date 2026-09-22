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

package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/prosemirror"
	tasksync "go.probo.inc/probo/pkg/task/sync"
	"go.probo.inc/probo/pkg/timespan"
	"go.probo.inc/probo/pkg/validator"
)

const (
	TitleMaxLength   = 1000
	ContentMaxLength = 5000

	richTextMaxJSONBytes  = 64 << 10
	maxRecurrenceInterval = 10 * 365 * 24 * time.Hour
)

type Service struct {
	pg     *pg.Client
	logger *log.Logger
	Sync   *tasksync.Service
}

func NewService(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	connectorRegistry *connector.Registry,
	baseURL string,
	linearAPIBaseURL string,
	logger *log.Logger,
) *Service {
	return &Service{
		pg:     pgClient,
		logger: logger,
		Sync: tasksync.NewService(
			pgClient,
			encryptionKey,
			connectorRegistry,
			baseURL,
			linearAPIBaseURL,
			logger,
			InsertUpdateActivities,
		),
	}
}

type (
	CreateTaskRequest struct {
		OrganizationID     gid.GID
		MeasureID          *gid.GID
		Name               string
		Content            *string
		State              *coredata.TaskState
		Priority           coredata.TaskPriority
		TimeEstimate       *timespan.TimeSpan
		AssignedToID       *gid.GID
		Deadline           *time.Time
		IdentityID         *gid.GID
		RecurrenceInterval *timespan.TimeSpan
	}

	UpdateTaskRequest struct {
		TaskID             gid.GID
		Name               *string
		Content            **string
		State              *coredata.TaskState
		Priority           *coredata.TaskPriority
		TimeEstimate       **timespan.TimeSpan
		Deadline           **time.Time
		AssignedToID       **gid.GID
		MeasureID          **gid.GID
		Rank               *int
		IdentityID         *gid.GID
		RecurrenceInterval **timespan.TimeSpan
	}

	UpdateTaskResult struct {
		Task     *coredata.Task
		NextTask *coredata.Task
	}
)

func (ctr *CreateTaskRequest) Validate() error {
	v := validator.New()

	v.Check(ctr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(ctr.MeasureID, "measure_id", validator.GID(coredata.MeasureEntityType))
	v.Check(ctr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(
		ctr.Content,
		"content",
		validator.MaxLen(richTextMaxJSONBytes),
		validator.ProseMirrorDocumentContent(),
		validator.ProseMirrorDocumentMaxTextLength(ContentMaxLength),
	)
	v.Check(ctr.State, "state", validator.OneOfSlice(coredata.TaskStates()))
	v.Check(ctr.Priority, "priority", validator.Required(), validator.OneOfSlice(coredata.TaskPriorities()))
	v.Check(ctr.TimeEstimate, "time_estimate", validator.RangeDuration(0, 1000*time.Hour))
	v.Check(ctr.AssignedToID, "assigned_to_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(ctr.IdentityID, "identity_id", validator.GID(coredata.IdentityEntityType))
	v.Check(ctr.RecurrenceInterval, "recurrence_interval", validator.RangeDuration(time.Nanosecond, maxRecurrenceInterval))

	if ctr.RecurrenceInterval != nil && ctr.Deadline == nil {
		v.Check(ctr.Deadline, "deadline", func(any) *validator.ValidationError {
			return &validator.ValidationError{
				Code:    validator.ErrorCodeCustom,
				Message: "deadline is required when the task is recurring",
			}
		})
	}

	if ctr.RecurrenceInterval != nil && ctr.Deadline != nil {
		v.Check(ctr.RecurrenceInterval, "recurrence_interval", func(any) *validator.ValidationError {
			if recurrenceAdvancesDeadline(*ctr.Deadline, *ctr.RecurrenceInterval) {
				return nil
			}

			return &validator.ValidationError{
				Code:    validator.ErrorCodeCustom,
				Message: "must advance the deadline",
			}
		})
	}

	if ctr.RecurrenceInterval != nil && ctr.State != nil && *ctr.State == coredata.TaskStateDone {
		v.Check(ctr.State, "state", func(any) *validator.ValidationError {
			return &validator.ValidationError{
				Code:    validator.ErrorCodeCustom,
				Message: "a recurring task cannot be created as done",
			}
		})
	}

	return v.Error()
}

func (utr *UpdateTaskRequest) Validate() error {
	v := validator.New()

	v.Check(utr.TaskID, "task_id", validator.Required(), validator.GID(coredata.TaskEntityType))
	v.Check(utr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(
		utr.Content,
		"content",
		validator.MaxLen(richTextMaxJSONBytes),
		validator.ProseMirrorDocumentContent(),
		validator.ProseMirrorDocumentMaxTextLength(ContentMaxLength),
	)
	v.Check(utr.Priority, "priority", validator.OneOfSlice(coredata.TaskPriorities()))
	v.Check(utr.TimeEstimate, "time_estimate", validator.RangeDuration(0, 1000*time.Hour))
	v.Check(utr.State, "state", validator.OneOfSlice(coredata.TaskStates()))
	v.Check(utr.AssignedToID, "assigned_to_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(utr.MeasureID, "measure_id", validator.GID(coredata.MeasureEntityType))
	v.Check(utr.Rank, "rank", validator.Min(1))
	v.Check(utr.IdentityID, "identity_id", validator.GID(coredata.IdentityEntityType))
	v.Check(utr.RecurrenceInterval, "recurrence_interval", validator.RangeDuration(time.Nanosecond, maxRecurrenceInterval))

	return v.Error()
}

func (s *Service) Create(
	ctx context.Context, scope coredata.Scoper,
	req CreateTaskRequest,
) (*coredata.Task, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	taskID := gid.New(scope.GetTenantID(), coredata.TaskEntityType)

	content, err := prosemirror.DefaultDocumentJSON(req.Content)
	if err != nil {
		return nil, fmt.Errorf("cannot sanitize task content: %w", err)
	}

	referenceID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("cannot generate reference id: %w", err)
	}

	state := coredata.TaskStateTodo
	if req.State != nil {
		state = *req.State
	}

	task := &coredata.Task{
		ID:             taskID,
		OrganizationID: req.OrganizationID,
		MeasureID:      req.MeasureID,
		Name:           req.Name,
		Content:        content,
		Priority:       req.Priority,
		TimeEstimate:   req.TimeEstimate,
		AssignedToID:   req.AssignedToID,
		Deadline:       req.Deadline,
		Recurrence:     req.RecurrenceInterval,
		State:          state,
		ReferenceID:    "custom-task-" + referenceID.String(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = s.pg.WithTx(
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

			actorID, err := ResolveActivityActorID(
				ctx,
				conn,
				scope,
				req.IdentityID,
				task.OrganizationID,
			)
			if err != nil {
				return fmt.Errorf("cannot resolve task activity actor: %w", err)
			}

			if err := InsertCreatedActivity(ctx, conn, scope, task, actorID, now); err != nil {
				return fmt.Errorf("cannot record task created event: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create task: %w", err)
	}

	return task, nil
}

func (s *Service) Get(
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
) (*coredata.Task, error) {
	task := &coredata.Task{}

	err := s.pg.WithConn(
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

func (s *Service) GetByIDs(
	ctx context.Context, scope coredata.Scoper,
	taskIDs ...gid.GID,
) (coredata.Tasks, error) {
	var tasks coredata.Tasks

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := tasks.LoadByIDs(
				ctx,
				conn,
				scope,
				taskIDs,
			); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
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

func (s *Service) Assign(
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
	assignedToID gid.GID,
	identityID *gid.GID,
) (*coredata.Task, error) {
	task := &coredata.Task{ID: taskID}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := task.LoadByID(ctx, conn, scope, taskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", taskID, err)
			}

			assignee := &coredata.MembershipProfile{}
			if err := assignee.LoadByID(ctx, conn, scope, assignedToID); err != nil {
				return fmt.Errorf("cannot load assignee profile: %w", err)
			}

			oldAssignedToID := task.AssignedToID
			if gidPtrEqual(oldAssignedToID, &assignedToID) {
				return nil
			}

			oldName, err := taskActivityProfileName(ctx, conn, scope, oldAssignedToID)
			if err != nil {
				return fmt.Errorf("cannot load previous assignee name: %w", err)
			}

			task.AssignedToID = &assignedToID
			now := time.Now()
			task.UpdatedAt = now

			if err := task.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot assign task %q to %q: %w", taskID, assignedToID, err)
			}

			actorID, err := ResolveActivityActorID(
				ctx,
				conn,
				scope,
				identityID,
				task.OrganizationID,
			)
			if err != nil {
				return fmt.Errorf("cannot resolve task activity actor: %w", err)
			}

			if err := insertFieldActivity(
				ctx,
				conn,
				scope,
				task,
				actorID,
				coredata.TaskActivityFieldAssignedTo,
				oldName,
				&assignee.FullName,
				now,
			); err != nil {
				return fmt.Errorf("cannot record task assignee event: %w", err)
			}

			if s.Sync != nil {
				if err := s.Sync.EnqueueOutbound(
					ctx,
					conn,
					scope,
					task.ID,
					tasksync.SyncActionUpdate,
				); err != nil {
					return fmt.Errorf("cannot enqueue task sync: %w", err)
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

func (s *Service) Unassign(
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
	identityID *gid.GID,
) (*coredata.Task, error) {
	task := &coredata.Task{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := task.LoadByID(ctx, conn, scope, taskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", taskID, err)
			}

			if task.AssignedToID == nil {
				return nil
			}

			oldName, err := taskActivityProfileName(ctx, conn, scope, task.AssignedToID)
			if err != nil {
				return fmt.Errorf("cannot load previous assignee name: %w", err)
			}

			task.AssignedToID = nil
			now := time.Now()
			task.UpdatedAt = now

			if err := task.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot unassign task %q: %w", taskID, err)
			}

			actorID, err := ResolveActivityActorID(
				ctx,
				conn,
				scope,
				identityID,
				task.OrganizationID,
			)
			if err != nil {
				return fmt.Errorf("cannot resolve task activity actor: %w", err)
			}

			if err := insertFieldActivity(
				ctx,
				conn,
				scope,
				task,
				actorID,
				coredata.TaskActivityFieldAssignedTo,
				oldName,
				nil,
				now,
			); err != nil {
				return fmt.Errorf("cannot record task unassign event: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *Service) Update(
	ctx context.Context, scope coredata.Scoper,
	req UpdateTaskRequest,
) (*UpdateTaskResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	task := &coredata.Task{}

	var nextTask *coredata.Task

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := task.LoadByIDForUpdate(ctx, conn, scope, req.TaskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", req.TaskID, err)
			}

			oldTask := *task

			if req.Name != nil {
				task.Name = *req.Name
			}

			if req.Content != nil {
				content, err := prosemirror.DefaultDocumentJSON(*req.Content)
				if err != nil {
					return fmt.Errorf("cannot sanitize task content: %w", err)
				}

				task.Content = content
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

			if req.RecurrenceInterval != nil {
				task.Recurrence = *req.RecurrenceInterval
			}

			settingRecurrence := req.RecurrenceInterval != nil && *req.RecurrenceInterval != nil
			if settingRecurrence && task.Deadline == nil {
				return validator.ValidationErrors{&validator.ValidationError{
					Field:   "deadline",
					Code:    validator.ErrorCodeCustom,
					Message: "deadline is required when the task is recurring",
				}}
			}

			if task.Deadline == nil {
				task.Recurrence = nil
			}

			changingRecurrence := req.RecurrenceInterval != nil || req.Deadline != nil
			if task.Recurrence != nil && task.Deadline != nil && changingRecurrence {
				if !recurrenceAdvancesDeadline(*task.Deadline, *task.Recurrence) {
					return validator.ValidationErrors{&validator.ValidationError{
						Field:   "recurrence_interval",
						Code:    validator.ErrorCodeCustom,
						Message: "must advance the deadline",
					}}
				}
			}

			now := time.Now()
			if shouldCloneRecurringTask(oldTask.State, task.State, task) {
				next, err := insertNextRecurringTask(ctx, conn, scope, task, now)
				if err != nil {
					return err
				}

				task.Recurrence = nil
				nextTask = next
			}

			task.UpdatedAt = now

			targetRank := req.Rank
			priorityChanged := task.Priority != oldTask.Priority
			stateChanged := task.State != oldTask.State

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

			actorID, err := ResolveActivityActorID(
				ctx,
				conn,
				scope,
				req.IdentityID,
				task.OrganizationID,
			)
			if err != nil {
				return fmt.Errorf("cannot resolve task activity actor: %w", err)
			}

			if err := InsertUpdateActivities(
				ctx,
				conn,
				scope,
				&oldTask,
				task,
				actorID,
				now,
			); err != nil {
				return fmt.Errorf("cannot record task update events: %w", err)
			}

			if s.Sync != nil && syncedTaskFieldsChanged(
				oldTask.Name,
				oldTask.Content,
				oldTask.State,
				oldTask.Priority,
				oldTask.Deadline,
				oldTask.AssignedToID,
				task,
			) {
				if err := s.Sync.EnqueueOutbound(
					ctx,
					conn,
					scope,
					task.ID,
					tasksync.SyncActionUpdate,
				); err != nil {
					return fmt.Errorf("cannot enqueue task sync: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &UpdateTaskResult{
		Task:     task,
		NextTask: nextTask,
	}, nil
}

func (s *Service) Delete(
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
) error {
	task := &coredata.Task{ID: taskID}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if s.Sync != nil {
				if err := s.Sync.EnqueueOutbound(
					ctx,
					conn,
					scope,
					taskID,
					tasksync.SyncActionCancel,
				); err != nil {
					return fmt.Errorf("cannot enqueue task sync: %w", err)
				}
			}

			return task.Delete(ctx, conn, scope)
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) CountForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
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

func (s *Service) ListForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.TaskOrderField],
) (*page.Page[*coredata.Task, coredata.TaskOrderField], error) {
	var tasks coredata.Tasks

	err := s.pg.WithConn(
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

func (s *Service) CountForMeasureID(
	ctx context.Context, scope coredata.Scoper,
	measureID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
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

func (s *Service) ListForMeasureID(
	ctx context.Context, scope coredata.Scoper,
	measureID gid.GID,
	cursor *page.Cursor[coredata.TaskOrderField],
) (*page.Page[*coredata.Task, coredata.TaskOrderField], error) {
	var tasks coredata.Tasks

	err := s.pg.WithConn(
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

func syncedTaskFieldsChanged(
	oldName string,
	oldContent string,
	oldState coredata.TaskState,
	oldPriority coredata.TaskPriority,
	oldDeadline *time.Time,
	oldAssignedToID *gid.GID,
	task *coredata.Task,
) bool {
	if task.Name != oldName || task.Content != oldContent || task.State != oldState || task.Priority != oldPriority {
		return true
	}

	if !gidPtrEqual(oldAssignedToID, task.AssignedToID) {
		return true
	}

	if oldDeadline == nil || task.Deadline == nil {
		return oldDeadline != task.Deadline
	}

	return !oldDeadline.Equal(*task.Deadline)
}
