// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package cloudaccount

import "errors"

// Typed sentinels returned by provider Probe methods and by
// MapSDKError. Service-layer callers translate these into typed
// last_probe_error categories and resolver-side gqlutils errors.
var (
	// ErrCredentialsInvalid signals the supplied credentials cannot
	// authenticate at all -- expired secret, malformed key,
	// revoked role.
	ErrCredentialsInvalid = errors.New("cloud account credentials are invalid")

	// ErrInsufficientPermissions signals authentication succeeded
	// but the role/SA/SP lacks one or more permissions required by
	// the configured audit modules.
	ErrInsufficientPermissions = errors.New("cloud account credentials lack required permissions")

	// ErrScopeUnreachable signals the scope identifier (account/
	// project/org/MG/subscription) is wrong, deleted, or the
	// principal cannot see it.
	ErrScopeUnreachable = errors.New("cloud account scope is unreachable")

	// ErrInstallTemplateUnavailable signals the deployment is not
	// configured to serve install assets for the requested
	// (provider, scope) combination -- typically a missing
	// AWSTemplateURL/AWSTemplateSHA256 in probodconfig.
	ErrInstallTemplateUnavailable = errors.New("cloud account install template is unavailable")
)
