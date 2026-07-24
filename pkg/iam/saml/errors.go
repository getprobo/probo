// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

type ErrUserInactive struct{ ProfileID gid.GID }

func NewUserInactiveError(profileID gid.GID) error {
	return &ErrUserInactive{ProfileID: profileID}
}

func (e ErrUserInactive) Error() string {
	return fmt.Sprintf("user %q is inactive", e.ProfileID)
}

type ErrSAMLSubjectAlreadyInUse struct {
	AssertionID string
}

func NewSAMLSubjectAlreadyInUseError(assertionID string) error {
	return &ErrSAMLSubjectAlreadyInUse{AssertionID: assertionID}
}

func (e ErrSAMLSubjectAlreadyInUse) Error() string {
	return fmt.Sprintf("SAML NameID is already linked to another account (assertion %q)", e.AssertionID)
}

type ErrSAMLSubjectRequired struct{}

func NewSAMLSubjectRequiredError() error {
	return &ErrSAMLSubjectRequired{}
}

func (e ErrSAMLSubjectRequired) Error() string {
	return "NameID value is required"
}
