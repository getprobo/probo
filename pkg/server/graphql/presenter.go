// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package graphql

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
)

func ErrorPresenter(ctx context.Context, e error) *gqlerror.Error {
	var rescuedErr *RescuedError
	if !errors.As(e, &rescuedErr) {
		return graphql.DefaultErrorPresenter(ctx, e)
	}

	err := graphql.DefaultErrorPresenter(ctx, e)
	originalErr := rescuedErr.Original

	var errResourceNotFound *coredata.ErrResourceNotFound
	var errResourceAlreadyExists *coredata.ErrResourceAlreadyExists
	var errRestrictedOperation *coredata.ErrRestrictedOperation
	var errInvalidValue *coredata.ErrInvalidValue
	var errNoChange *coredata.ErrNoChange

	if errors.As(originalErr, &errResourceNotFound) {
		err.Message = errResourceNotFound.Error()
		err.Extensions = map[string]any{
			"code": "NOT_FOUND",
		}
	} else if errors.As(originalErr, &errResourceAlreadyExists) {
		err.Message = errResourceAlreadyExists.Error()
		err.Extensions = map[string]any{
			"code": "ALREADY_EXISTS",
		}
	} else if errors.As(originalErr, &errRestrictedOperation) {
		err.Message = errRestrictedOperation.Error()
		err.Extensions = map[string]any{
			"code": "RESTRICTED_OPERATION",
		}
	} else if errors.As(originalErr, &errInvalidValue) {
		err.Message = errInvalidValue.Error()
		err.Extensions = map[string]any{
			"code": "INVALID_VALUE",
		}
	} else if errors.As(originalErr, &errNoChange) {
		err.Message = errNoChange.Error()
		err.Extensions = map[string]any{
			"code": "NO_CHANGE",
		}
	} else if originalErr.Error() == "tenant not found" {
		err.Message = "not authorized"
		err.Extensions = map[string]any{
			"code": "UNAUTHORIZED",
		}
	} else {
		logger := httpserver.LoggerFromContext(ctx)
		logger.Error("unhandled rescued error in resolver", log.Any("error", originalErr))

		err.Message = "Internal server error"
		err.Extensions = map[string]any{
			"code": "INTERNAL_SERVER_ERROR",
		}
	}

	return err
}
