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

package mcp_v1

import (
	"context"
	"errors"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

// These helpers stay outside schema.resolvers.go. mcpgen treats every
// Resolver method in that file as a specification handler and comments out
// anything the spec does not name.

func (r *Resolver) taskExternalLinksByTasks(
	ctx context.Context,
	scope coredata.Scoper,
	tasks []*coredata.Task,
) (map[gid.GID]*coredata.TaskExternalLink, error) {
	if r.task.Sync == nil || len(tasks) == 0 {
		return nil, nil
	}

	ids := make([]gid.GID, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}

	return r.task.Sync.GetLinksByTaskIDs(ctx, scope, ids)
}

func (r *Resolver) internalControlIDsByTasks(
	ctx context.Context,
	scope coredata.Scoper,
	tasks []*coredata.Task,
) (map[gid.GID][]gid.GID, error) {
	if len(tasks) == 0 {
		return map[gid.GID][]gid.GID{}, nil
	}

	ids := make([]gid.GID, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}

	return r.task.InternalControlIDsByTaskIDs(ctx, scope, ids)
}

func (r *Resolver) taskWithExternalLink(
	ctx context.Context,
	scope coredata.Scoper,
	task *coredata.Task,
) (*types.Task, error) {
	internalControlIDs, err := r.task.InternalControlIDsByTaskIDs(ctx, scope, []gid.GID{task.ID})
	if err != nil {
		return nil, err
	}

	result := types.NewTask(task, internalControlIDs[task.ID])
	if r.task.Sync == nil {
		return result, nil
	}

	link, err := r.task.Sync.GetLinkByTaskID(ctx, scope, task.ID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return result, nil
		}

		return nil, err
	}

	result.ExternalLink = types.NewTaskExternalLink(link)

	return result, nil
}
