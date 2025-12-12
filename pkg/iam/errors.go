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

package iam

import (
	"fmt"

	"go.probo.inc/probo/pkg/gid"
)

type ErrInvalidTokenType struct{ message string }

func NewInvalidTokenError() error {
	return &ErrInvalidTokenType{"invalid invitation token"}
}

func (e ErrInvalidTokenType) Error() string {
	return e.message
}

type ErrInvitationAlreadyAccepted struct{ InvitationID gid.GID }

func NewInvitationAlreadyAcceptedError(invitationID gid.GID) error {
	return &ErrInvitationAlreadyAccepted{InvitationID: invitationID}
}

func (e ErrInvitationAlreadyAccepted) Error() string {
	return fmt.Sprintf("invitation %q already accepted", e.InvitationID)
}

type ErrInvitationNotFound struct{ InvitationID gid.GID }

func NewInvitationNotFoundError(invitationID gid.GID) error {
	return &ErrInvitationNotFound{InvitationID: invitationID}
}

func (e ErrInvitationNotFound) Error() string {
	return fmt.Sprintf("invitation %q not found", e.InvitationID)
}

type ErrInvitationExpired struct{ InvitationID gid.GID }

func NewInvitationExpiredError(invitationID gid.GID) error {
	return &ErrInvitationExpired{InvitationID: invitationID}
}

func (e ErrInvitationExpired) Error() string {
	return fmt.Sprintf("invitation %q expired", e.InvitationID)
}

type ErrUserAlreadyExists struct{ EmailAddress string }

func NewUserAlreadyExistsError(emailAddress string) error {
	return &ErrUserAlreadyExists{EmailAddress: emailAddress}
}

func (e ErrUserAlreadyExists) Error() string {
	return fmt.Sprintf("user %q already exists", e.EmailAddress)
}

type ErrEmailAlreadyVerified struct{ message string }

func NewEmailAlreadyVerifiedError() error {
	return &ErrEmailAlreadyVerified{"email already verified"}
}

func (e ErrEmailAlreadyVerified) Error() string {
	return e.message
}

type ErrUserNotFound struct{ UserID gid.GID }

func NewUserNotFoundError(userID gid.GID) error {
	return &ErrUserNotFound{userID}
}

func (e ErrUserNotFound) Error() string {
	return fmt.Sprintf("user %q not found", e.UserID)
}

type ErrInvalidPassword struct{ message string }

func NewInvalidPasswordError(message string) error {
	return &ErrInvalidPassword{message}
}

func (e ErrInvalidPassword) Error() string {
	return e.message
}

type ErrEmailVerificationMismatch struct{ message string }

func NewEmailVerificationMismatchError() error {
	return &ErrEmailVerificationMismatch{"email verification mismatch"}
}

func (e ErrEmailVerificationMismatch) Error() string {
	return e.message
}

type ErrMembershipNotFound struct {
	MembershipID gid.GID
}

func NewMembershipNotFoundError(membershipID gid.GID) error {
	return &ErrMembershipNotFound{MembershipID: membershipID}
}

func (e ErrMembershipNotFound) Error() string {
	return fmt.Sprintf("membership %q not found", e.MembershipID)
}

type ErrOrganizationNotFound struct{ OrganizationID gid.GID }

func NewOrganizationNotFoundError(organizationID gid.GID) error {
	return &ErrOrganizationNotFound{OrganizationID: organizationID}
}

func (e ErrOrganizationNotFound) Error() string {
	return fmt.Sprintf("organization %q not found", e.OrganizationID)
}

type ErrInsufficientPermissions struct{ message string }

func NewInsufficientPermissionsError() error {
	return &ErrInsufficientPermissions{"insufficient permissions"}
}

func (e ErrInsufficientPermissions) Error() string {
	return e.message
}

type ErrSessionNotFound struct{ SessionID gid.GID }

func NewSessionNotFoundError(sessionID gid.GID) error {
	return &ErrSessionNotFound{SessionID: sessionID}
}

func (e ErrSessionNotFound) Error() string {
	return fmt.Sprintf("session %q not found", e.SessionID)
}

type ErrSessionExpired struct{ SessionID gid.GID }

func NewSessionExpiredError(sessionID gid.GID) error {
	return &ErrSessionExpired{SessionID: sessionID}
}

func (e ErrSessionExpired) Error() string {
	return fmt.Sprintf("session %q expired", e.SessionID)
}

type ErrMembershipAlreadyExists struct {
	UserID         gid.GID
	OrganizationID gid.GID
}

func NewMembershipAlreadyExistsError(userID gid.GID, organizationID gid.GID) error {
	return &ErrMembershipAlreadyExists{UserID: userID, OrganizationID: organizationID}
}

func (e ErrMembershipAlreadyExists) Error() string {
	return fmt.Sprintf("membership already exists for user %q in organization %q", e.UserID, e.OrganizationID)
}

type ErrSAMLConfigurationNotFound struct{ ConfigID gid.GID }

func NewSAMLConfigurationNotFoundError(configID gid.GID) error {
	return &ErrSAMLConfigurationNotFound{ConfigID: configID}
}

func (e ErrSAMLConfigurationNotFound) Error() string {
	return fmt.Sprintf("SAML configuration %q not found", e.ConfigID)
}

type ErrUserAPIKeyNotFound struct{ UserAPIKeyID gid.GID }

func NewUserAPIKeyNotFoundError(userAPIKeyID gid.GID) error {
	return &ErrUserAPIKeyNotFound{UserAPIKeyID: userAPIKeyID}
}

func (e ErrUserAPIKeyNotFound) Error() string {
	return fmt.Sprintf("user API key %q not found", e.UserAPIKeyID)
}

type ErrUserAPIKeyExpired struct{ UserAPIKeyID gid.GID }

func NewUserAPIKeyExpiredError(userAPIKeyID gid.GID) error {
	return &ErrUserAPIKeyExpired{UserAPIKeyID: userAPIKeyID}
}

func (e ErrUserAPIKeyExpired) Error() string {
	return fmt.Sprintf("user API key %q expired", e.UserAPIKeyID)
}

type ErrSAMLConfigurationDomainNotVerified struct{ ConfigID gid.GID }

func NewSAMLConfigurationDomainNotVerifiedError(configID gid.GID) error {
	return &ErrSAMLConfigurationDomainNotVerified{ConfigID: configID}
}

func (e ErrSAMLConfigurationDomainNotVerified) Error() string {
	return fmt.Sprintf("SAML configuration %q domain not verified", e.ConfigID)
}
