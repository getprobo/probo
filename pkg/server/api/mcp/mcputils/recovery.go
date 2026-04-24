// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/validator"
)

func RecoveryMiddleware(logger *log.Logger) func(mcp.MethodHandler) mcp.MethodHandler {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (result mcp.Result, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = convertPanicToError(ctx, logger, r)
					result = nil
				}
			}()

			result, err = next(ctx, method, req)
			if err != nil {
				err = sanitizeError(ctx, logger, err)
			}
			return
		}
	}
}

func convertPanicToError(ctx context.Context, logger *log.Logger, panicValue any) error {
	if panicValue == nil {
		logger.ErrorCtx(ctx, "nil panic in MCP method handler")
		return fmt.Errorf("internal server error")
	}

	if err, ok := panicValue.(error); ok {
		return sanitizeError(ctx, logger, err)
	}

	logger.ErrorCtx(
		ctx,
		"unexpected panic in MCP method handler",
		log.Any("panic", panicValue),
		log.String("stack", string(debug.Stack())),
	)

	return fmt.Errorf("internal server error")
}

// sanitizeError classifies known error types and returns a clear message for
// those. Unknown errors are logged and replaced with a generic internal error
// to avoid leaking implementation details to the client.
func sanitizeError(ctx context.Context, logger *log.Logger, err error) error {
	var permissionDeniedErr *iam.ErrInsufficientPermissions
	if errors.As(err, &permissionDeniedErr) {
		return fmt.Errorf("permission denied")
	}

	var assumptionRequiredErr *iam.ErrAssumptionRequired
	if errors.As(err, &assumptionRequiredErr) {
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

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		return validationErrors
	}

	var validationError *validator.ValidationError
	if errors.As(err, &validationError) {
		return validationError
	}

	logger.ErrorCtx(ctx, "internal error in MCP method handler", log.Error(err))
	return fmt.Errorf("internal server error")
}
