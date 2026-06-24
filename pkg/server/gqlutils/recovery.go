// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package gqlutils

import (
	"context"
	"errors"
	"runtime/debug"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/validator"
)

func RecoverFunc(ctx context.Context, err any) error {
	if gqlErr, ok := err.(*gqlerror.Error); ok {
		return gqlErr
	}

	if errValidations, ok := errors.AsType[validator.ValidationErrors](asError(err)); ok {
		gqlErrors := gqlerror.List{}

		for _, err := range errValidations {
			gqlErrors = append(
				gqlErrors,
				Invalid(
					ctx,
					err,
				),
			)
		}

		return gqlErrors
	}

	if errTyped, ok := err.(error); ok {
		if permissionDeniedErr, ok := errors.AsType[*iam.ErrInsufficientPermissions](errTyped); ok {
			return Forbidden(ctx, permissionDeniedErr)
		}
	}

	logger := httpserver.LoggerFromContext(ctx)
	logger.Error("resolver panic", log.Any("error", err), log.String("stack", string(debug.Stack())))

	return errors.New("internal server error")
}

func asError(err any) error {
	if e, ok := err.(error); ok {
		return e
	}

	return errors.New("unknown panic")
}
