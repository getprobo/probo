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
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/timespan"
)

func insertTaskCreatedActivity(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	task *coredata.Task,
	actorID *gid.GID,
	now time.Time,
) error {
	activity := &coredata.TaskActivity{
		ID:             gid.New(scope.GetTenantID(), coredata.TaskActivityEntityType),
		OrganizationID: task.OrganizationID,
		TaskID:         task.ID,
		ActorID:        actorID,
		ActivityType:   coredata.TaskActivityTypeCreated,
		CreatedAt:      now,
	}

	if err := activity.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert task created activity: %w", err)
	}

	return nil
}

func insertTaskFieldActivity(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	task *coredata.Task,
	actorID *gid.GID,
	field coredata.TaskActivityField,
	oldValue *string,
	newValue *string,
	now time.Time,
) error {
	activity := &coredata.TaskActivity{
		ID:             gid.New(scope.GetTenantID(), coredata.TaskActivityEntityType),
		OrganizationID: task.OrganizationID,
		TaskID:         task.ID,
		ActorID:        actorID,
		ActivityType:   coredata.TaskActivityTypeUpdated,
		Field:          &field,
		OldValue:       oldValue,
		NewValue:       newValue,
		CreatedAt:      now,
	}

	if err := activity.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert task field activity: %w", err)
	}

	return nil
}

func insertTaskUpdateActivities(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	oldTask *coredata.Task,
	task *coredata.Task,
	actorID *gid.GID,
	now time.Time,
) error {
	if task.Name != oldTask.Name {
		if err := insertTaskFieldActivity(
			ctx,
			tx,
			scope,
			task,
			actorID,
			coredata.TaskActivityFieldName,
			new(oldTask.Name),
			new(task.Name),
			now,
		); err != nil {
			return fmt.Errorf("cannot record name activity: %w", err)
		}
	}

	if task.Content != oldTask.Content {
		if err := insertTaskFieldActivity(
			ctx,
			tx,
			scope,
			task,
			actorID,
			coredata.TaskActivityFieldDescription,
			new(oldTask.Content),
			new(task.Content),
			now,
		); err != nil {
			return fmt.Errorf("cannot record description activity: %w", err)
		}
	}

	if task.State != oldTask.State {
		if err := insertTaskFieldActivity(
			ctx,
			tx,
			scope,
			task,
			actorID,
			coredata.TaskActivityFieldState,
			new(oldTask.State.String()),
			new(task.State.String()),
			now,
		); err != nil {
			return fmt.Errorf("cannot record state activity: %w", err)
		}
	}

	if task.Priority != oldTask.Priority {
		if err := insertTaskFieldActivity(
			ctx,
			tx,
			scope,
			task,
			actorID,
			coredata.TaskActivityFieldPriority,
			new(oldTask.Priority.String()),
			new(task.Priority.String()),
			now,
		); err != nil {
			return fmt.Errorf("cannot record priority activity: %w", err)
		}
	}

	if !timespanPtrEqual(oldTask.TimeEstimate, task.TimeEstimate) {
		if err := insertTaskFieldActivity(
			ctx,
			tx,
			scope,
			task,
			actorID,
			coredata.TaskActivityFieldTimeEstimate,
			timespanValue(oldTask.TimeEstimate),
			timespanValue(task.TimeEstimate),
			now,
		); err != nil {
			return fmt.Errorf("cannot record time estimate activity: %w", err)
		}
	}

	if !timePtrEqual(oldTask.Deadline, task.Deadline) {
		if err := insertTaskFieldActivity(
			ctx,
			tx,
			scope,
			task,
			actorID,
			coredata.TaskActivityFieldDeadline,
			timeValue(oldTask.Deadline),
			timeValue(task.Deadline),
			now,
		); err != nil {
			return fmt.Errorf("cannot record deadline activity: %w", err)
		}
	}

	if !gidPtrEqual(oldTask.AssignedToID, task.AssignedToID) {
		oldAssigneeName, err := taskActivityProfileName(ctx, tx, scope, oldTask.AssignedToID)
		if err != nil {
			return fmt.Errorf("cannot load previous assignee name: %w", err)
		}

		newAssigneeName, err := taskActivityProfileName(ctx, tx, scope, task.AssignedToID)
		if err != nil {
			return fmt.Errorf("cannot load assignee name: %w", err)
		}

		if err := insertTaskFieldActivity(
			ctx,
			tx,
			scope,
			task,
			actorID,
			coredata.TaskActivityFieldAssignedTo,
			oldAssigneeName,
			newAssigneeName,
			now,
		); err != nil {
			return fmt.Errorf("cannot record assignee activity: %w", err)
		}
	}

	if !gidPtrEqual(oldTask.MeasureID, task.MeasureID) {
		oldMeasureName, err := taskActivityMeasureName(ctx, tx, scope, oldTask.MeasureID)
		if err != nil {
			return fmt.Errorf("cannot load previous measure name: %w", err)
		}

		newMeasureName, err := taskActivityMeasureName(ctx, tx, scope, task.MeasureID)
		if err != nil {
			return fmt.Errorf("cannot load measure name: %w", err)
		}

		if err := insertTaskFieldActivity(
			ctx,
			tx,
			scope,
			task,
			actorID,
			coredata.TaskActivityFieldMeasure,
			oldMeasureName,
			newMeasureName,
			now,
		); err != nil {
			return fmt.Errorf("cannot record measure activity: %w", err)
		}
	}

	return nil
}

func resolveTaskActivityActorID(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	identityID *gid.GID,
	organizationID gid.GID,
) (*gid.GID, error) {
	if identityID == nil {
		return nil, nil
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByIdentityIDAndOrganizationID(
		ctx,
		conn,
		scope,
		*identityID,
		organizationID,
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load actor profile: %w", err)
	}

	return &profile.ID, nil
}

func taskActivityProfileName(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	profileID *gid.GID,
) (*string, error) {
	if profileID == nil {
		return nil, nil
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByID(ctx, conn, scope, *profileID); err != nil {
		return nil, fmt.Errorf("cannot load profile: %w", err)
	}

	return &profile.FullName, nil
}

func taskActivityMeasureName(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	measureID *gid.GID,
) (*string, error) {
	if measureID == nil {
		return nil, nil
	}

	measure := &coredata.Measure{}
	if err := measure.LoadByID(ctx, conn, scope, *measureID); err != nil {
		return nil, fmt.Errorf("cannot load measure: %w", err)
	}

	return &measure.Name, nil
}

func timespanValue(value *timespan.TimeSpan) *string {
	if value == nil {
		return nil
	}

	return new(value.String())
}

func timeValue(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.UTC().Format(time.RFC3339Nano)

	return &formatted
}

func gidPtrEqual(a *gid.GID, b *gid.GID) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return *a == *b
}

func timespanPtrEqual(a *timespan.TimeSpan, b *timespan.TimeSpan) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return *a == *b
}

func timePtrEqual(a *time.Time, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return a.Equal(*b)
}
