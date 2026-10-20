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

package mcpx

import (
	"context"
	"runtime/debug"

	"go.gearno.de/kit/log"
	mcpgenmcp "go.probo.inc/mcpgen/mcp"
)

// NewRecoverFunc returns a RecoverFunc for the generated MCP server that
// classifies panics into safe client-facing errors and logs unknown errors.
func NewRecoverFunc(logger *log.Logger) mcpgenmcp.RecoverFunc {
	return func(ctx context.Context, r any) error {
		if r == nil {
			logger.ErrorCtx(ctx, "nil panic in MCP tool handler")
			return ErrInternal
		}

		if err, ok := r.(error); ok {
			return MapError(ctx, logger, "internal error in MCP tool handler", err)
		}

		logger.ErrorCtx(
			ctx,
			"unexpected panic in MCP tool handler",
			log.Any("panic", r),
			log.String("stack", string(debug.Stack())),
		)

		return ErrInternal
	}
}
