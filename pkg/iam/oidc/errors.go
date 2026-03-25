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

package oidc

import (
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
)

type ErrProviderNotEnabled struct {
	Provider coredata.OIDCProvider
}

func NewProviderNotEnabledError(provider coredata.OIDCProvider) error {
	return &ErrProviderNotEnabled{Provider: provider}
}

func (e ErrProviderNotEnabled) Error() string {
	return fmt.Sprintf("cannot authenticate: OIDC provider %q is not enabled", e.Provider)
}

type ErrInvalidState struct{}

func NewInvalidStateError() error {
	return &ErrInvalidState{}
}

func (e ErrInvalidState) Error() string {
	return "cannot validate OIDC state: invalid or expired"
}

type ErrCodeExchange struct {
	Err error
}

func NewCodeExchangeError(err error) error {
	return &ErrCodeExchange{Err: err}
}

func (e ErrCodeExchange) Error() string {
	return fmt.Sprintf("cannot exchange authorization code: %v", e.Err)
}

func (e ErrCodeExchange) Unwrap() error {
	return e.Err
}

type ErrIDTokenMissing struct{}

func NewIDTokenMissingError() error {
	return &ErrIDTokenMissing{}
}

func (e ErrIDTokenMissing) Error() string {
	return "cannot extract id_token: not present in token response"
}

type ErrMissingEmailClaim struct{}

func NewMissingEmailClaimError() error {
	return &ErrMissingEmailClaim{}
}

func (e ErrMissingEmailClaim) Error() string {
	return "cannot extract email: claim missing from id token"
}

type ErrEmailNotVerified struct{}

func NewEmailNotVerifiedError() error {
	return &ErrEmailNotVerified{}
}

func (e ErrEmailNotVerified) Error() string {
	return "cannot authenticate: email address is not verified by the OIDC provider"
}

type ErrPersonalAccountNotAllowed struct{}

func NewPersonalAccountNotAllowedError() error {
	return &ErrPersonalAccountNotAllowed{}
}

func (e ErrPersonalAccountNotAllowed) Error() string {
	return "cannot authenticate: personal accounts are not allowed, use an enterprise account"
}
