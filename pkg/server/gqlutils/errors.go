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

func AssumptionRequiredf(ctx context.Context, format string, a ...any) *gqlerror.Error {
	return AssumptionRequired(ctx, fmt.Errorf(format, a...))
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

func Invalid(ctx context.Context, err error) *gqlerror.Error {
	var errValidation *validator.ValidationError

	var details map[string]any

	if errors.As(err, &errValidation) {
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

func Internal(ctx context.Context) *gqlerror.Error {
	return &gqlerror.Error{
		Message: "An internal server error occurred. Please try again later.",
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]any{
			"code": "INTERNAL",
		},
	}
}
