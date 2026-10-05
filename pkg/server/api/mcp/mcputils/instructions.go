// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// InstructionsMiddleware fills client-facing server instructions on initialize
// and discover. mcpgen builds the server without ServerOptions, so this is
// applied on the result instead of at construction.
func InstructionsMiddleware(instructions string) func(mcp.MethodHandler) mcp.MethodHandler {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			result, err := next(ctx, method, req)
			if err != nil || instructions == "" || result == nil {
				return result, err
			}

			switch res := result.(type) {
			case *mcp.InitializeResult:
				if res.Instructions == "" {
					res.Instructions = instructions
				}
			case *mcp.DiscoverResult:
				if res.Instructions == "" {
					res.Instructions = instructions
				}
			}

			return result, nil
		}
	}
}
