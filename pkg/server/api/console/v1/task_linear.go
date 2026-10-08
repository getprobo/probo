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
	"errors"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/gqlutils"
	tasksync "go.probo.inc/probo/pkg/task/sync"
)

func (r *mutationResolver) linearPublishError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, coredata.ErrResourceNotFound):
		return gqlutils.NotFound(ctx, err)
	case errors.Is(err, tasksync.ErrLinearNotConnected),
		errors.Is(err, tasksync.ErrLinearReconnectRequired),
		errors.Is(err, tasksync.ErrLinearTeamIDRequired),
		errors.Is(err, tasksync.ErrLinearTeamNotFound):
		return gqlutils.Invalid(ctx, err)
	case errors.Is(err, tasksync.ErrTaskAlreadyLinked),
		errors.Is(err, coredata.ErrResourceAlreadyExists):
		return gqlutils.Conflict(ctx, err)
	default:
		r.logger.ErrorCtx(ctx, "cannot publish task to Linear", log.Error(err))

		return gqlutils.Internal(ctx)
	}
}
