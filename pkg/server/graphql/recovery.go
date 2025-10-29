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

package graphql

import (
	"context"
	"errors"
	"runtime/debug"

	"github.com/getprobo/probo/pkg/auth"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
)

func RecoverFunc(ctx context.Context, err any) error {
	if gqlErr, ok := err.(*gqlerror.Error); ok {
		return gqlErr
	}

	var errSAMLRequired auth.ErrSAMLAuthRequired
	if errors.As(asError(err), &errSAMLRequired) {
		return &gqlerror.Error{
			Message: "Additional authentication required to access this organization",
			Extensions: map[string]any{
				"code":           "AUTHENTICATION_REQUIRED",
				"requiresSaml":   true,
				"redirectUrl":    errSAMLRequired.RedirectURL,
				"samlConfigId":   errSAMLRequired.ConfigID.String(),
				"organizationId": errSAMLRequired.OrganizationID.String(),
			},
		}
	}

	var errPasswordRequired auth.ErrPasswordAuthRequired
	if errors.As(asError(err), &errPasswordRequired) {
		return &gqlerror.Error{
			Message: "Additional authentication required to access this organization",
			Extensions: map[string]any{
				"code":           "AUTHENTICATION_REQUIRED",
				"requiresSaml":   false,
				"redirectUrl":    errPasswordRequired.RedirectURL,
				"organizationId": errPasswordRequired.OrganizationID.String(),
			},
		}
	}

	logger := httpserver.LoggerFromContext(ctx)
	logger.Error("resolver panic", log.Any("error", err), log.Any("stack", string(debug.Stack())))

	return errors.New("internal server error")
}

func asError(err any) error {
	if e, ok := err.(error); ok {
		return e
	}
	return errors.New("unknown panic")
}
