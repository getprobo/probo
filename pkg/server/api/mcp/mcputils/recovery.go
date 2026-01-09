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

package mcputils

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/iam"
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
			return
		}
	}
}

func convertPanicToError(ctx context.Context, logger *log.Logger, panicValue any) error {
	if panicValue == nil {
		logger.ErrorCtx(ctx, "nil panic in MCP method handler")
		return fmt.Errorf("internal server error")
	}

	var permissionDeniedErr *iam.ErrInsufficientPermissions
	if errTyped, ok := panicValue.(error); ok && errors.As(errTyped, &permissionDeniedErr) {
		return fmt.Errorf("permission denied: %s", permissionDeniedErr.Error())
	}

	if err, ok := panicValue.(error); ok {
		return err
	}

	// Log unexpected panics with stack trace
	logger.ErrorCtx(ctx, "unexpected panic in MCP method handler",
		log.Any("panic", panicValue),
		log.String("stack", string(debug.Stack())),
	)

	return fmt.Errorf("internal server error")
}
