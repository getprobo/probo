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

package console_v1

import (
	"context"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlx"
	"go.probo.inc/probo/pkg/tag"
)

// taggableTagsParent marks a TagConnection whose ParentID is a tagged resource.
type taggableTagsParent struct{}

func (r *Resolver) tagsForResourceID(
	ctx context.Context,
	resourceID gid.GID,
	first *int,
	after *page.CursorKey,
	last *int,
	before *page.CursorKey,
	orderBy *types.TagOrderBy,
) (*types.TagConnection, error) {
	scope, err := r.authorize(ctx, resourceID, tag.ActionAssignmentGet)
	if err != nil {
		return nil, err
	}

	pageOrderBy := page.OrderBy[coredata.TagOrderField]{
		Field:     coredata.TagOrderFieldCreatedAt,
		Direction: page.OrderDirectionDesc,
	}

	if orderBy != nil {
		pageOrderBy = page.OrderBy[coredata.TagOrderField]{
			Field:     orderBy.Field,
			Direction: orderBy.Direction,
		}
	}

	cursor := types.NewCursor(first, after, last, before, pageOrderBy)

	pageResult, err := r.tag.ListForResourceID(ctx, scope, resourceID, cursor)
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot list tags for resource", log.Error(err))

		return nil, gqlx.Internal(ctx)
	}

	return types.NewTagConnection(pageResult, taggableTagsParent{}, resourceID), nil
}
