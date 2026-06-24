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
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
)

func LoggingMiddleware(logger *log.Logger) func(mcp.MethodHandler) mcp.MethodHandler {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			logger.InfoCtx(
				ctx,
				fmt.Sprintf("mcp %q method started", method),
				log.String("method", method),
				log.Bool("has_params", req.GetParams() != nil),
			)

			if ctr, ok := req.(*mcp.CallToolRequest); ok {
				logger.InfoCtx(
					ctx,
					fmt.Sprintf("calling %q tool", ctr.Params.Name),
					log.String("tool_name", ctr.Params.Name),
				)
			}

			start := time.Now()
			result, err := next(ctx, method, req)
			duration := time.Since(start)

			if err != nil {
				logger.ErrorCtx(
					ctx,
					fmt.Sprintf("mcp %q method failed", method),
					log.String("method", method),
					log.Int64("duration_ms", duration.Milliseconds()),
					log.Error(err),
				)
			} else {
				logger.InfoCtx(
					ctx,
					fmt.Sprintf("mcp %q method completed", method),
					log.String("method", method),
					log.Int64("duration_ms", duration.Milliseconds()),
					log.Bool("has_result", result != nil),
				)

				if ctr, ok := result.(*mcp.CallToolResult); ok {
					logger.InfoCtx(
						ctx,
						"tool call result",
						log.Bool("is_error", ctr.IsError),
					)
				}
			}

			return result, err
		}
	}
}
