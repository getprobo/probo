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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// OAuth2Error represents an OAuth2 protocol error with an associated
// error code per RFC 6749 §5.2 and RFC 8628 §3.5.
type OAuth2Error struct {
	code        string
	description string
}

func (e *OAuth2Error) Error() string {
	if e.description != "" {
		return e.code + ": " + e.description
	}
	return e.code
}

func (e *OAuth2Error) ErrorCode() string   { return e.code }
func (e *OAuth2Error) Description() string { return e.description }

func (e *OAuth2Error) Is(target error) bool {
	t, ok := target.(*OAuth2Error)
	if !ok {
		return false
	}
	return e.code == t.code
}

type ErrorOption func(*OAuth2Error)

func WithDescription(description string) ErrorOption {
	return func(e *OAuth2Error) {
		e.description = description
	}
}

func WithError(err error) ErrorOption {
	return func(e *OAuth2Error) {
		e.description = err.Error()
	}
}

// NewError creates a new OAuth2Error derived from a sentinel error code.
func NewError(code *OAuth2Error, opts ...ErrorOption) *OAuth2Error {
	e := &OAuth2Error{code: code.code}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

var (
	// OAuth2 error codes per RFC 6749 §5.2 and RFC 8628 §3.5.
	ErrInvalidRequest       = &OAuth2Error{code: "invalid_request"}
	ErrInvalidClient        = &OAuth2Error{code: "invalid_client"}
	ErrInvalidGrant         = &OAuth2Error{code: "invalid_grant"}
	ErrUnauthorizedClient   = &OAuth2Error{code: "unauthorized_client"}
	ErrUnsupportedGrantType = &OAuth2Error{code: "unsupported_grant_type"}
	ErrInvalidScope         = &OAuth2Error{code: "invalid_scope"}
	ErrAccessDenied         = &OAuth2Error{code: "access_denied"}
	ErrServerError          = &OAuth2Error{code: "server_error"}
	ErrInvalidRedirectURI   = &OAuth2Error{code: "invalid_redirect_uri"}

	// RFC 8628 device flow errors.
	ErrAuthorizationPending = &OAuth2Error{code: "authorization_pending"}
	ErrSlowDown             = &OAuth2Error{code: "slow_down"}
	ErrExpiredToken         = &OAuth2Error{code: "expired_token"}
)

// ConsentRequiredError is returned by Authorize when the user must approve
// the authorization request before a code can be issued.
type ConsentRequiredError struct {
	ConsentID gid.GID
	Client    *coredata.OAuth2Client
	Scopes    coredata.OAuth2Scopes
}

func (e *ConsentRequiredError) Error() string {
	return "consent required"
}
