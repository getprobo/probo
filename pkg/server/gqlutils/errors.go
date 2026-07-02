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
	"fmt"
	"maps"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.probo.inc/probo/pkg/validator"
)

func AlreadyAuthenticated(ctx context.Context, err error) *gqlerror.Error {
	return &gqlerror.Error{
		Message: "Authentication not allowed for this resource/action",
		Extensions: map[string]any{
			"code": "ALREADY_AUTHENTICATED",
		},
	}
}

func AlreadyAuthenticatedf(ctx context.Context, format string, a ...any) *gqlerror.Error {
	return AlreadyAuthenticated(ctx, fmt.Errorf(format, a...))
}

func Unauthenticated(ctx context.Context, err error) *gqlerror.Error {
	return &gqlerror.Error{
		Message: err.Error(),
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]any{
			"code": "UNAUTHENTICATED",
		},
	}
}

func Unauthenticatedf(ctx context.Context, format string, a ...any) *gqlerror.Error {
	return Unauthenticated(ctx, fmt.Errorf(format, a...))
}

func AssumptionRequired(ctx context.Context, err error) *gqlerror.Error {
	return &gqlerror.Error{
		Message: err.Error(),
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]any{
			"code": "ASSUMPTION_REQUIRED",
		},
	}
}

func FullNameRequired(ctx context.Context, err error) *gqlerror.Error {
	return &gqlerror.Error{
		Message: err.Error(),
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]any{
			"code": "FULL_NAME_REQUIRED",
		},
	}
}

func FullNameRequiredf(ctx context.Context, format string, a ...any) *gqlerror.Error {
	return FullNameRequired(ctx, fmt.Errorf(format, a...))
}

func NDASignatureRequired(ctx context.Context, err error) *gqlerror.Error {
	return &gqlerror.Error{
		Message: err.Error(),
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]any{
			"code": "NDA_SIGNATURE_REQUIRED",
		},
	}
}

func NDASignatureRequiredf(ctx context.Context, format string, a ...any) *gqlerror.Error {
	return NDASignatureRequired(ctx, fmt.Errorf(format, a...))
}

func AccountAlreadyActivated(ctx context.Context, err error) *gqlerror.Error {
	return &gqlerror.Error{
		Message: err.Error(),
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]any{
			"code": "ACCOUNT_ALREADY_ACTIVATED",
		},
	}
}

func Forbidden(ctx context.Context, err error) *gqlerror.Error {
	return &gqlerror.Error{
		Message: err.Error(),
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]any{
			"code": "FORBIDDEN",
		},
	}
}

func Forbiddenf(ctx context.Context, format string, a ...any) *gqlerror.Error {
	return Forbidden(ctx, fmt.Errorf(format, a...))
}

func NotFound(ctx context.Context, err error) *gqlerror.Error {
	return &gqlerror.Error{
		Message: err.Error(),
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]any{
			"code": "NOT_FOUND",
		},
	}
}

func NotFoundf(ctx context.Context, format string, a ...any) *gqlerror.Error {
	return NotFound(ctx, fmt.Errorf(format, a...))
}

func Conflict(ctx context.Context, err error) *gqlerror.Error {
	return &gqlerror.Error{
		Message: err.Error(),
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]any{
			"code": "CONFLICT",
		},
	}
}

func Conflictf(ctx context.Context, format string, a ...any) *gqlerror.Error {
	return Conflict(ctx, fmt.Errorf(format, a...))
}

func TokenExpired(ctx context.Context, err error) *gqlerror.Error {
	return &gqlerror.Error{
		Message:    err.Error(),
		Path:       graphql.GetPath(ctx),
		Extensions: map[string]any{"code": "TOKEN_EXPIRED"},
	}
}

func TokenAlreadyUsed(ctx context.Context, err error) *gqlerror.Error {
	return &gqlerror.Error{
		Message:    err.Error(),
		Path:       graphql.GetPath(ctx),
		Extensions: map[string]any{"code": "TOKEN_ALREADY_USED"},
	}
}

func Invalid(ctx context.Context, err error) *gqlerror.Error {
	var details map[string]any

	if errValidation, ok := errors.AsType[*validator.ValidationError](err); ok {
		details = map[string]any{
			"cause": errValidation.Code,
			"field": errValidation.Field,
			"value": errValidation.Value,
		}
	}

	extensions := map[string]any{"code": "INVALID"}
	if details != nil {
		maps.Copy(extensions, details)
	}

	return &gqlerror.Error{
		Message:    err.Error(),
		Path:       graphql.GetPath(ctx),
		Extensions: extensions,
	}
}

func Invalidf(ctx context.Context, format string, a ...any) *gqlerror.Error {
	return Invalid(ctx, fmt.Errorf(format, a...))
}

func InvalidValidationErrors(ctx context.Context, errs validator.ValidationErrors) gqlerror.List {
	gqlErrors := make(gqlerror.List, 0, len(errs))
	for _, ve := range errs {
		gqlErrors = append(gqlErrors, Invalid(ctx, ve))
	}

	return gqlErrors
}

func Internal(ctx context.Context) *gqlerror.Error {
	return &gqlerror.Error{
		Message: "An internal server error occurred. Please try again later.",
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]any{
			"code": "INTERNAL",
		},
	}
}

func Unavailable(ctx context.Context, err error) *gqlerror.Error {
	return &gqlerror.Error{
		Message: err.Error(),
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]any{
			"code": "UNAVAILABLE",
		},
	}
}

func Unavailablef(ctx context.Context, format string, a ...any) *gqlerror.Error {
	return Unavailable(ctx, fmt.Errorf(format, a...))
}
