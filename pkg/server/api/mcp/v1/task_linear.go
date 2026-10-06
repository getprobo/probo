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
	"fmt"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	tasksync "go.probo.inc/probo/pkg/task/sync"
)

func linearPublishToolError(r *Resolver, ctx context.Context, err error) error {
	switch {
	case errors.Is(err, coredata.ErrResourceNotFound):
		return fmt.Errorf("task not found")
	case errors.Is(err, tasksync.ErrLinearNotConnected):
		return fmt.Errorf("linear connector is not connected")
	case errors.Is(err, tasksync.ErrLinearReconnectRequired):
		return fmt.Errorf("linear connector must be reconnected with write scopes")
	case errors.Is(err, tasksync.ErrTaskAlreadyLinked),
		errors.Is(err, coredata.ErrResourceAlreadyExists):
		return fmt.Errorf("task is already linked to an external issue")
	case errors.Is(err, tasksync.ErrLinearTeamIDRequired):
		return fmt.Errorf("linear team id is required")
	case errors.Is(err, tasksync.ErrLinearTeamNotFound):
		return fmt.Errorf("linear team was not found")
	default:
		r.logger.ErrorCtx(ctx, "cannot publish task to Linear", log.Error(err))

		return fmt.Errorf("internal error")
	}
}
