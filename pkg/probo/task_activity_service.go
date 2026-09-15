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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	TaskActivityService struct {
		svc *Service
	}
)

func (s TaskActivityService) Get(
	ctx context.Context, scope coredata.Scoper,
	taskActivityID gid.GID,
) (*coredata.TaskActivity, error) {
	taskActivity := &coredata.TaskActivity{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := taskActivity.LoadByID(ctx, conn, scope, taskActivityID); err != nil {
				return fmt.Errorf("cannot load task activity: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return taskActivity, nil
}

func (s TaskActivityService) ListForTaskID(
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
	cursor *page.Cursor[coredata.TaskActivityOrderField],
) (*page.Page[*coredata.TaskActivity, coredata.TaskActivityOrderField], error) {
	var taskActivities coredata.TaskActivities

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := taskActivities.LoadByTaskID(ctx, conn, scope, taskID, cursor); err != nil {
				return fmt.Errorf("cannot load task activities: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(taskActivities, cursor), nil
}

func (s TaskActivityService) CountForTaskID(
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			taskActivities := coredata.TaskActivities{}

			count, err = taskActivities.CountByTaskID(ctx, conn, scope, taskID)
			if err != nil {
				return fmt.Errorf("cannot count task activities: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
