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
