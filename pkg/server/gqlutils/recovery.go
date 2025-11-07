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

package gqlutils

import (
	"context"
	"errors"
	"runtime/debug"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/validator"
)

func RecoverFunc(ctx context.Context, err any) error {
	if gqlErr, ok := err.(*gqlerror.Error); ok {
		return gqlErr
	}

	var errSAMLRequired auth.ErrSAMLAuthRequired
	if errors.As(asError(err), &errSAMLRequired) {
		return AuthenticationRequired(map[string]any{
			"requiresSaml":   true,
			"redirectUrl":    errSAMLRequired.RedirectURL,
			"samlConfigId":   errSAMLRequired.ConfigID.String(),
			"organizationId": errSAMLRequired.OrganizationID.String(),
		})
	}

	var errPasswordRequired auth.ErrPasswordAuthRequired
	if errors.As(asError(err), &errPasswordRequired) {
		return AuthenticationRequired(map[string]any{
			"requiresSaml":   false,
			"redirectUrl":    errPasswordRequired.RedirectURL,
			"organizationId": errPasswordRequired.OrganizationID.String(),
		})
	}

	var errValidations validator.ValidationErrors
	if errors.As(asError(err), &errValidations) {
		gqlErrors := gqlerror.List{}

		for _, err := range errValidations {
			gqlErrors = append(
				gqlErrors,
				Invalid(
					err,
					map[string]any{
						"cause": err.Code,
						"field": err.Field,
						"value": err.Value,
					},
				),
			)
		}

		return gqlErrors
	}

	var tenantAccessErr *authz.TenantAccessError
	if errTyped, ok := err.(error); ok && errors.As(errTyped, &tenantAccessErr) {
		return Unauthorized()
	}

	var permissionDeniedErr *authz.PermissionDeniedError
	if errTyped, ok := err.(error); ok && errors.As(errTyped, &permissionDeniedErr) {
		return Forbidden(permissionDeniedErr)
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
