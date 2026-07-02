// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package mcputils

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"go.gearno.de/kit/log"
	mcpgenmcp "go.probo.inc/mcpgen/mcp"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/validator"
)

// NewRecoverFunc returns a RecoverFunc for the generated MCP server that
// classifies panics into safe client-facing errors and logs unknown errors.
func NewRecoverFunc(logger *log.Logger) mcpgenmcp.RecoverFunc {
	return func(ctx context.Context, r any) error {
		if r == nil {
			logger.ErrorCtx(ctx, "nil panic in MCP tool handler")
			return fmt.Errorf("internal server error")
		}

		if err, ok := r.(error); ok {
			return sanitizeError(ctx, logger, err)
		}

		logger.ErrorCtx(
			ctx,
			"unexpected panic in MCP tool handler",
			log.Any("panic", r),
			log.String("stack", string(debug.Stack())),
		)

		return fmt.Errorf("internal server error")
	}
}

// sanitizeError classifies known error types and returns a clear message for
// those. Unknown errors are logged and replaced with a generic internal error
// to avoid leaking implementation details to the client.
func sanitizeError(ctx context.Context, logger *log.Logger, err error) error {
	if _, ok := errors.AsType[*iam.ErrInsufficientPermissions](err); ok {
		return fmt.Errorf("permission denied")
	}

	if _, ok := errors.AsType[*iam.ErrAssumptionRequired](err); ok {
		return fmt.Errorf("assumption required")
	}

	if errors.Is(err, coredata.ErrResourceNotFound) {
		return fmt.Errorf("resource not found")
	}

	if errors.Is(err, coredata.ErrResourceAlreadyExists) {
		return fmt.Errorf("resource already exists")
	}

	if errors.Is(err, coredata.ErrResourceInUse) {
		return fmt.Errorf("resource is in use")
	}

	if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
		return validationErrors
	}

	if validationError, ok := errors.AsType[*validator.ValidationError](err); ok {
		return validationError
	}

	logger.ErrorCtx(ctx, "internal error in MCP tool handler", log.Error(err))

	return fmt.Errorf("internal server error")
}
