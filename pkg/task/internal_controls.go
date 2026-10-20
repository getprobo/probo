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
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

func uniqueInternalControlIDs(ids []gid.GID) []gid.GID {
	if len(ids) == 0 {
		return []gid.GID{}
	}

	seen := make(map[gid.GID]struct{}, len(ids))

	unique := make([]gid.GID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	return unique
}

func checkInternalControlIDs(v *validator.Validator, field string, ids []gid.GID) {
	for i, id := range ids {
		v.Check(
			id,
			fmt.Sprintf("%s[%d]", field, i),
			validator.Required(),
			validator.GID(coredata.InternalControlEntityType),
		)
	}
}

func loadTaskInternalControlIDs(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	taskID gid.GID,
) ([]gid.GID, error) {
	ids, err := (coredata.InternalControlTask{}).LoadInternalControlIDsByTaskID(ctx, conn, scope, taskID)
	if err != nil {
		return nil, fmt.Errorf("cannot load task internal controls: %w", err)
	}

	return ids, nil
}

func ensureInternalControlsBelongToOrganization(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	organizationID gid.GID,
	internalControlIDs []gid.GID,
) error {
	if len(internalControlIDs) == 0 {
		return nil
	}

	internalControls := coredata.InternalControls{}
	if err := internalControls.LoadByIDs(ctx, conn, scope, internalControlIDs); err != nil {
		return fmt.Errorf("cannot load internal controls: %w", err)
	}

	for _, internalControl := range internalControls {
		if internalControl.OrganizationID != organizationID {
			return validator.ValidationErrors{&validator.ValidationError{
				Field:   "internal_control_ids",
				Code:    validator.ErrorCodeCustom,
				Message: "must belong to the task organization",
			}}
		}
	}

	return nil
}

func replaceTaskInternalControls(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	task *coredata.Task,
	internalControlIDs []gid.GID,
) error {
	ids := uniqueInternalControlIDs(internalControlIDs)
	if err := ensureInternalControlsBelongToOrganization(ctx, conn, scope, task.OrganizationID, ids); err != nil {
		return fmt.Errorf("cannot check internal controls: %w", err)
	}

	if err := (coredata.InternalControlTasks{}).Merge(
		ctx,
		conn,
		scope,
		task.ID,
		task.OrganizationID,
		task.ReferenceID,
		ids,
	); err != nil {
		return fmt.Errorf("cannot merge task internal controls: %w", err)
	}

	return nil
}

func (s *Service) InternalControlIDsByTaskIDs(
	ctx context.Context,
	scope coredata.Scoper,
	taskIDs []gid.GID,
) (map[gid.GID][]gid.GID, error) {
	var idsByTaskID map[gid.GID][]gid.GID

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error

			idsByTaskID, err = (coredata.InternalControlTask{}).LoadInternalControlIDsByTaskIDs(ctx, conn, scope, taskIDs)
			if err != nil {
				return fmt.Errorf("cannot load task internal controls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return idsByTaskID, nil
}
