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
