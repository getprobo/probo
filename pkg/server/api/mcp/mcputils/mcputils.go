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
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
)

func LoggingMiddleware(logger *log.Logger) func(mcp.MethodHandler) mcp.MethodHandler {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			logger.InfoCtx(ctx, fmt.Sprintf("mcp %q method started", method),
				log.String("method", method),
				log.Bool("has_params", req.GetParams() != nil),
			)

			if ctr, ok := req.(*mcp.CallToolRequest); ok {
				logger.InfoCtx(ctx, fmt.Sprintf("calling %q tool", ctr.Params.Name),
					log.String("tool_name", ctr.Params.Name),
				)
			}

			start := time.Now()
			result, err := next(ctx, method, req)
			duration := time.Since(start)

			if err != nil {
				logger.ErrorCtx(ctx, fmt.Sprintf("mcp %q method failed", method),
					log.String("method", method),
					log.Int64("duration_ms", duration.Milliseconds()),
					log.Error(err),
				)
			} else {

				logger.InfoCtx(ctx, fmt.Sprintf("mcp %q method completed", method),
					log.String("method", method),
					log.Int64("duration_ms", duration.Milliseconds()),
					log.Bool("has_result", result != nil),
				)

				if ctr, ok := result.(*mcp.CallToolResult); ok {
					logger.InfoCtx(ctx, "tool call result",
						log.Bool("is_error", ctr.IsError),
					)
				}
			}

			return result, err
		}
	}
}
