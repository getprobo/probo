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

package oauth2server

import (
	"errors"

	"go.probo.inc/probo/pkg/coredata"
)

var (
	// OAuth2 error codes per RFC 6749 §5.2 and RFC 8628 §3.5.
	ErrInvalidRequest       = errors.New("invalid_request")
	ErrInvalidClient        = errors.New("invalid_client")
	ErrInvalidGrant         = errors.New("invalid_grant")
	ErrUnauthorizedClient   = errors.New("unauthorized_client")
	ErrUnsupportedGrantType = errors.New("unsupported_grant_type")
	ErrInvalidScope         = errors.New("invalid_scope")
	ErrAccessDenied         = errors.New("access_denied")
	ErrServerError          = errors.New("server_error")
	ErrInvalidRedirectURI   = errors.New("invalid redirect_uri")

	// RFC 8628 device flow errors.
	ErrAuthorizationPending = errors.New("authorization_pending")
	ErrSlowDown             = errors.New("slow_down")
	ErrExpiredToken         = errors.New("expired_token")
)

// ConsentRequiredError is returned by Authorize when the user must approve
// the authorization request before a code can be issued.
type ConsentRequiredError struct {
	Client *coredata.OAuth2Client
	Scopes coredata.OAuth2Scopes
}

func (e *ConsentRequiredError) Error() string {
	return "consent required"
}

// OAuth2ErrorCode returns the OAuth2 error code string for a sentinel error.
func OAuth2ErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrInvalidRequest):
		return "invalid_request"
	case errors.Is(err, ErrInvalidClient):
		return "invalid_client"
	case errors.Is(err, ErrInvalidGrant):
		return "invalid_grant"
	case errors.Is(err, ErrUnauthorizedClient):
		return "unauthorized_client"
	case errors.Is(err, ErrUnsupportedGrantType):
		return "unsupported_grant_type"
	case errors.Is(err, ErrInvalidScope):
		return "invalid_scope"
	case errors.Is(err, ErrAccessDenied):
		return "access_denied"
	case errors.Is(err, ErrAuthorizationPending):
		return "authorization_pending"
	case errors.Is(err, ErrSlowDown):
		return "slow_down"
	case errors.Is(err, ErrExpiredToken):
		return "expired_token"
	default:
		return "server_error"
	}
}
