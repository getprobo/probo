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

package task

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

func InsertCreatedActivity(
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

func insertFieldActivity(
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

func InsertUpdateActivities(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	oldTask *coredata.Task,
	task *coredata.Task,
	actorID *gid.GID,
	now time.Time,
) error {
	if task.Name != oldTask.Name {
		if err := insertFieldActivity(
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
		if err := insertFieldActivity(
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
		if err := insertFieldActivity(
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
		if err := insertFieldActivity(
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
		if err := insertFieldActivity(
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
		if err := insertFieldActivity(
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

		if err := insertFieldActivity(
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

	return nil
}

func InsertInternalControlActivities(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	task *coredata.Task,
	actorID *gid.GID,
	now time.Time,
	oldInternalControlIDs []gid.GID,
	newInternalControlIDs []gid.GID,
) error {
	removed, added := diffInternalControlIDs(oldInternalControlIDs, newInternalControlIDs)
	if len(removed) == 0 && len(added) == 0 {
		return nil
	}

	names, err := taskActivityInternalControlNames(ctx, tx, scope, removed, added)
	if err != nil {
		return err
	}

	if len(removed) == 1 && len(added) == 1 {
		oldInternalControlName, err := internalControlActivityName(names, removed[0])
		if err != nil {
			return fmt.Errorf("cannot load previous internal control name: %w", err)
		}

		newInternalControlName, err := internalControlActivityName(names, added[0])
		if err != nil {
			return fmt.Errorf("cannot load internal control name: %w", err)
		}

		if err := insertFieldActivity(
			ctx,
			tx,
			scope,
			task,
			actorID,
			coredata.TaskActivityFieldInternalControl,
			oldInternalControlName,
			newInternalControlName,
			now,
		); err != nil {
			return fmt.Errorf("cannot record internal control activity: %w", err)
		}

		return nil
	}

	for _, internalControlID := range removed {
		oldInternalControlName, err := internalControlActivityName(names, internalControlID)
		if err != nil {
			return fmt.Errorf("cannot load previous internal control name: %w", err)
		}

		if err := insertFieldActivity(
			ctx,
			tx,
			scope,
			task,
			actorID,
			coredata.TaskActivityFieldInternalControl,
			oldInternalControlName,
			nil,
			now,
		); err != nil {
			return fmt.Errorf("cannot record internal control activity: %w", err)
		}
	}

	for _, internalControlID := range added {
		newInternalControlName, err := internalControlActivityName(names, internalControlID)
		if err != nil {
			return fmt.Errorf("cannot load internal control name: %w", err)
		}

		if err := insertFieldActivity(
			ctx,
			tx,
			scope,
			task,
			actorID,
			coredata.TaskActivityFieldInternalControl,
			nil,
			newInternalControlName,
			now,
		); err != nil {
			return fmt.Errorf("cannot record internal control activity: %w", err)
		}
	}

	return nil
}

func diffInternalControlIDs(oldInternalControlIDs []gid.GID, newInternalControlIDs []gid.GID) (removed []gid.GID, added []gid.GID) {
	oldSet := make(map[gid.GID]struct{}, len(oldInternalControlIDs))
	newSet := make(map[gid.GID]struct{}, len(newInternalControlIDs))

	for _, id := range oldInternalControlIDs {
		oldSet[id] = struct{}{}
	}

	for _, id := range newInternalControlIDs {
		newSet[id] = struct{}{}
	}

	for _, id := range oldInternalControlIDs {
		if _, ok := newSet[id]; !ok {
			removed = append(removed, id)
		}
	}

	for _, id := range newInternalControlIDs {
		if _, ok := oldSet[id]; !ok {
			added = append(added, id)
		}
	}

	return removed, added
}

func ResolveActivityActorID(
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

func taskActivityInternalControlNames(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	removed []gid.GID,
	added []gid.GID,
) (map[gid.GID]string, error) {
	ids := make([]gid.GID, 0, len(removed)+len(added))
	ids = append(ids, removed...)
	ids = append(ids, added...)

	internalControls := coredata.InternalControls{}
	if err := internalControls.LoadByIDs(ctx, conn, scope, ids); err != nil {
		return nil, fmt.Errorf("cannot load internal controls: %w", err)
	}

	names := make(map[gid.GID]string, len(internalControls))
	for _, internalControl := range internalControls {
		names[internalControl.ID] = internalControl.Name
	}

	return names, nil
}

func internalControlActivityName(names map[gid.GID]string, id gid.GID) (*string, error) {
	name, ok := names[id]
	if !ok {
		return nil, fmt.Errorf("cannot load internal control: %w", coredata.ErrResourceNotFound)
	}

	return new(name), nil
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
