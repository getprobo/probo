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

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/gqlx"
	syncerrors "go.probo.inc/probo/pkg/task/sync/errors"
	linearsvc "go.probo.inc/probo/pkg/task/sync/linear"
)

const (
	publishProviderAsana  = "ASANA"
	publishProviderLinear = "LINEAR"
)

func (r *mutationResolver) linearPublishError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, coredata.ErrResourceNotFound):
		return withPublishProvider(gqlx.NotFound(ctx, err), publishProviderLinear)
	case errors.Is(err, linearsvc.ErrLinearNotConnected),
		errors.Is(err, linearsvc.ErrLinearReconnectRequired),
		errors.Is(err, linearsvc.ErrLinearTeamIDRequired),
		errors.Is(err, linearsvc.ErrLinearTeamNotFound):
		return withPublishProvider(gqlx.Invalid(ctx, err), publishProviderLinear)
	case errors.Is(err, syncerrors.ErrTaskAlreadyLinked),
		errors.Is(err, coredata.ErrResourceAlreadyExists):
		return withPublishProvider(gqlx.Conflict(ctx, err), publishProviderLinear)
	default:
		r.logger.ErrorCtx(ctx, "cannot publish task to Linear", log.Error(err))

		return withPublishProvider(gqlx.Internal(ctx), publishProviderLinear)
	}
}

func withPublishProvider(err error, provider string) error {
	switch typed := err.(type) {
	case *gqlerror.Error:
		setPublishProvider(typed, provider)
	case gqlerror.List:
		for _, item := range typed {
			setPublishProvider(item, provider)
		}
	}

	return err
}

func setPublishProvider(err *gqlerror.Error, provider string) {
	if err == nil {
		return
	}

	if err.Extensions == nil {
		err.Extensions = map[string]any{}
	}

	err.Extensions["provider"] = provider
}
