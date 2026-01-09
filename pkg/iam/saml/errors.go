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

package saml

import (
	"fmt"

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

type ErrSAMLConfigurationNotFound struct{ ConfigID gid.GID }

func NewSAMLConfigurationNotFoundError(configID gid.GID) error {
	return &ErrSAMLConfigurationNotFound{ConfigID: configID}
}

func (e ErrSAMLConfigurationNotFound) Error() string {
	return fmt.Sprintf("SAML configuration %q not found", e.ConfigID)
}

type ErrSAMLDisabled struct{}

func NewSAMLDisabledError() error {
	return &ErrSAMLDisabled{}
}

func (e ErrSAMLDisabled) Error() string {
	return "SAML is disabled for this organization"
}

type ErrInvalidAssertion struct {
	AssertionID string
	Err         error
}

func NewInvalidAssertionError(assertionID string, err error) error {
	return &ErrInvalidAssertion{AssertionID: assertionID, Err: err}
}

func (e ErrInvalidAssertion) Error() string {
	return fmt.Sprintf("invalid assertion %q: %v", e.AssertionID, e.Err)
}

type ErrReplayAttackDetected struct {
	AssertionID string
}

func NewReplayAttackDetectedError(assertionID string) error {
	return &ErrReplayAttackDetected{AssertionID: assertionID}
}

func (e ErrReplayAttackDetected) Error() string {
	return fmt.Sprintf("replay attack detected for assertion %q", e.AssertionID)
}

type ErrEmailDomainMismatch struct {
	Email          mail.Addr
	ExpectedDomain string
}

func NewEmailDomainMismatchError(email mail.Addr, expectedDomain string) error {
	return &ErrEmailDomainMismatch{Email: email, ExpectedDomain: expectedDomain}
}

func (e ErrEmailDomainMismatch) Error() string {
	return fmt.Sprintf("email domain mismatch: assertion contains email %q but SAML config is for domain %q", e.Email, e.ExpectedDomain)
}

type ErrSAMLAutoSignupDisabled struct{ ConfigID gid.GID }

func NewSAMLAutoSignupDisabledError(configID gid.GID) error {
	return &ErrSAMLAutoSignupDisabled{ConfigID: configID}
}

func (e ErrSAMLAutoSignupDisabled) Error() string {
	return fmt.Sprintf("SAML auto-signup is disabled for configuration %q", e.ConfigID)
}

type ErrMembershipInactive struct{ MembershipID gid.GID }

func NewMembershipInactiveError(membershipID gid.GID) error {
	return &ErrMembershipInactive{MembershipID: membershipID}
}

func (e ErrMembershipInactive) Error() string {
	return fmt.Sprintf("membership %q is inactive", e.MembershipID)
}
