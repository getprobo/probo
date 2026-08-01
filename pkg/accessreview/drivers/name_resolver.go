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

package drivers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// NameResolver fetches the human-readable instance name from a provider
// (e.g. Slack workspace name, Google Workspace domain).
type NameResolver interface {
	ResolveInstanceName(ctx context.Context) (string, error)
}

// ErrTerminalNameResolution marks a permanent name-resolution failure
// (auth/bad-request) that retrying cannot fix: the source-name worker keeps
// the generic name and marks the source synced instead of re-claiming it.
// Transient failures (5xx, network) stay plain errors so they keep retrying.
var ErrTerminalNameResolution = errors.New("terminal name resolution failure")

// nameStatusError classifies a non-2xx response from a name-resolution
// request. Permanent client errors (400, 401, 403, 404) wrap
// ErrTerminalNameResolution; everything else (notably 5xx) stays retryable.
func nameStatusError(what string, statusCode int) error {
	switch statusCode {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		return fmt.Errorf("cannot fetch %s: unexpected status %d: %w", what, statusCode, ErrTerminalNameResolution)
	default:
		return fmt.Errorf("cannot fetch %s: unexpected status %d", what, statusCode)
	}
}
