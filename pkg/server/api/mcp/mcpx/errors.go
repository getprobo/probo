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

package mcpx

import (
	"context"
	"errors"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/validator"
)

// Client-facing errors returned by MapError. Their text is safe to show.
var (
	ErrResourceNotFound      = errors.New("resource not found")
	ErrResourceAlreadyExists = errors.New("resource already exists")
	ErrResourceInUse         = errors.New("resource is in use")
	ErrForbidden             = errors.New("permission denied")
	ErrAssumptionRequired    = errors.New("assumption required")
	ErrInternal              = errors.New("internal server error")
)

// MapError turns a service error into the error an MCP client should see.
// Generic failures become an mcpx sentinel. A domain sentinel whose text is
// already safe to show is returned without its wrap. Validation errors are
// returned as-is. Anything else is logged and returned as ErrInternal.
func MapError(ctx context.Context, logger *log.Logger, message string, err error) error {
	if _, ok := errors.AsType[*iam.ErrInsufficientPermissions](err); ok {
		return ErrForbidden
	}

	if _, ok := errors.AsType[*iam.ErrAssumptionRequired](err); ok {
		return ErrAssumptionRequired
	}

	switch {
	case errors.Is(err, accessreview.ErrNoConnectorAccount):
		return accessreview.ErrNoConnectorAccount
	case errors.Is(err, coredata.ErrMultipleConnectorAccounts):
		return coredata.ErrMultipleConnectorAccounts
	case errors.Is(err, coredata.ErrResourceNotFound):
		return ErrResourceNotFound
	case errors.Is(err, coredata.ErrResourceAlreadyExists):
		return ErrResourceAlreadyExists
	case errors.Is(err, coredata.ErrResourceInUse):
		return ErrResourceInUse
	}

	if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
		return validationErrors
	}

	if validationError, ok := errors.AsType[*validator.ValidationError](err); ok {
		return validationError
	}

	logger.ErrorCtx(ctx, message, log.Error(err))

	return ErrInternal
}
