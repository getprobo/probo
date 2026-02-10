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
	"go.probo.inc/probo/pkg/mail"
)

type ErrInvalidToken struct{ message string }

func NewInvalidTokenError() error {
	return &ErrInvalidToken{"invalid token"}
}

func (e ErrInvalidToken) Error() string {
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

type ErrIdentityAlreadyExists struct{ EmailAddress mail.Addr }

func NewIdentityAlreadyExistsError(emailAddress mail.Addr) error {
	return &ErrIdentityAlreadyExists{EmailAddress: emailAddress}
}

func (e ErrIdentityAlreadyExists) Error() string {
	return fmt.Sprintf("identity %q already exists", e.EmailAddress.String())
}

type ErrEmailAlreadyVerified struct{ message string }

func NewEmailAlreadyVerifiedError() error {
	return &ErrEmailAlreadyVerified{"email already verified"}
}

func (e ErrEmailAlreadyVerified) Error() string {
	return e.message
}

type ErrIdentityNotFound struct{ IdentityID gid.GID }

func NewIdentityNotFoundError(identityID gid.GID) error {
	return &ErrIdentityNotFound{identityID}
}

func (e ErrIdentityNotFound) Error() string {
	return fmt.Sprintf("identity %q not found", e.IdentityID)
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

type ErrMembershipInactive struct {
	MembershipID gid.GID
}

func NewMembershipInactiveError(membershipID gid.GID) error {
	return &ErrMembershipInactive{MembershipID: membershipID}
}

func (e ErrMembershipInactive) Error() string {
	return fmt.Sprintf("membership %q is inactive", e.MembershipID)
}

type ErrMembershipManagedBySCIM struct {
	MembershipID gid.GID
}

func NewMembershipManagedBySCIMError(membershipID gid.GID) error {
	return &ErrMembershipManagedBySCIM{MembershipID: membershipID}
}

func (e ErrMembershipManagedBySCIM) Error() string {
	return fmt.Sprintf("membership %q is managed by SCIM and cannot be deleted manually", e.MembershipID)
}

type ErrLastActiveOwner struct {
	MembershipID gid.GID
}

func NewLastActiveOwnerError(membershipID gid.GID) error {
	return &ErrLastActiveOwner{MembershipID: membershipID}
}

func (e ErrLastActiveOwner) Error() string {
	return fmt.Sprintf("cannot remove membership %q: last active owner of the organization", e.MembershipID)
}

type ErrOrganizationNotFound struct{ OrganizationID gid.GID }

func NewOrganizationNotFoundError(organizationID gid.GID) error {
	return &ErrOrganizationNotFound{OrganizationID: organizationID}
}

func (e ErrOrganizationNotFound) Error() string {
	return fmt.Sprintf("organization %q not found", e.OrganizationID)
}

type ErrInsufficientPermissions struct {
	IdentityID gid.GID
	EntityID   gid.GID
	Action     Action
}

func NewInsufficientPermissionsError(identityID gid.GID, entityID gid.GID, action Action) error {
	return &ErrInsufficientPermissions{IdentityID: identityID, EntityID: entityID, Action: action}
}

func (e ErrInsufficientPermissions) Error() string {
	return fmt.Sprintf("identity %q does not have sufficient permissions to perform action %s on entity %q", e.IdentityID, e.Action, e.EntityID)
}

type ErrAssumptionRequired struct {
	IdentityID   gid.GID
	MembershipID gid.GID
}

func NewAssumptionRequiredError(identityID gid.GID, membershipID gid.GID) error {
	return &ErrAssumptionRequired{IdentityID: identityID, MembershipID: membershipID}
}

func (e ErrAssumptionRequired) Error() string {
	return fmt.Sprintf("assumption for identity %q required for membership %q", e.IdentityID, e.MembershipID)
}

type ErrSessionNotFound struct{ SessionID gid.GID }

func NewSessionNotFoundError(sessionID gid.GID) error {
	return &ErrSessionNotFound{SessionID: sessionID}
}

func (e ErrSessionNotFound) Error() string {
	if e.SessionID == gid.Nil {
		return "session not found"
	}

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
	IdentityID     gid.GID
	OrganizationID gid.GID
}

func NewMembershipAlreadyExistsError(identityID gid.GID, organizationID gid.GID) error {
	return &ErrMembershipAlreadyExists{IdentityID: identityID, OrganizationID: organizationID}
}

func (e ErrMembershipAlreadyExists) Error() string {
	return fmt.Sprintf("membership already exists for identity %q in organization %q", e.IdentityID, e.OrganizationID)
}

type ErrSAMLConfigurationNotFound struct{ ConfigID gid.GID }

func NewSAMLConfigurationNotFoundError(configID gid.GID) error {
	return &ErrSAMLConfigurationNotFound{ConfigID: configID}
}

func (e ErrSAMLConfigurationNotFound) Error() string {
	return fmt.Sprintf("SAML configuration %q not found", e.ConfigID)
}

type ErrPersonalAPIKeyNotFound struct{ PersonalAPIKeyID gid.GID }

func NewPersonalAPIKeyNotFoundError(personalAPIKeyID gid.GID) error {
	return &ErrPersonalAPIKeyNotFound{PersonalAPIKeyID: personalAPIKeyID}
}

func (e ErrPersonalAPIKeyNotFound) Error() string {
	return fmt.Sprintf("personal API key %q not found", e.PersonalAPIKeyID)
}

type ErrProfileNotFound struct{ MembershipID gid.GID }

func NewProfileNotFoundError(membershipID gid.GID) error {
	return &ErrProfileNotFound{MembershipID: membershipID}
}

func (e ErrProfileNotFound) Error() string {
	return fmt.Sprintf("profile for membership %q not found", e.MembershipID)
}

type ErrPersonalAPIKeyExpired struct{ PersonalAPIKeyID gid.GID }

func NewPersonalAPIKeyExpiredError(personalAPIKeyID gid.GID) error {
	return &ErrPersonalAPIKeyExpired{PersonalAPIKeyID: personalAPIKeyID}
}

func (e ErrPersonalAPIKeyExpired) Error() string {
	return fmt.Sprintf("personal API key %q expired", e.PersonalAPIKeyID)
}

type ErrSAMLConfigurationDomainNotVerified struct{ ConfigID gid.GID }

func NewSAMLConfigurationDomainNotVerifiedError(configID gid.GID) error {
	return &ErrSAMLConfigurationDomainNotVerified{ConfigID: configID}
}

func (e ErrSAMLConfigurationDomainNotVerified) Error() string {
	return fmt.Sprintf("SAML configuration %q domain not verified", e.ConfigID)
}

type ErrUnsupportedPrincipalType struct{ EntityType uint16 }

func NewUnsupportedPrincipalTypeError(entityType uint16) error {
	return &ErrUnsupportedPrincipalType{EntityType: entityType}
}

func (e ErrUnsupportedPrincipalType) Error() string {
	return fmt.Sprintf("unsupported principal type: %d", e.EntityType)
}

type ErrSignupDisabled struct{}

func NewErrSignupDisabled() error {
	return &ErrSignupDisabled{}
}

func (e ErrSignupDisabled) Error() string {
	return "signup is disabled"
}

type ErrInvalidCredentials struct{ message string }

func NewInvalidCredentialsError(message string) error {
	return &ErrInvalidCredentials{message}
}

func (e ErrInvalidCredentials) Error() string {
	return e.message
}

type ErrInvitationNotDeleted struct {
	InvitationID gid.GID
	Status       string
}

func NewInvitationNotDeletedError(invitationID gid.GID, status string) error {
	return &ErrInvitationNotDeleted{InvitationID: invitationID, Status: status}
}

func (e ErrInvitationNotDeleted) Error() string {
	return fmt.Sprintf("cannot delete invitation %q in %q status", e.InvitationID, e.Status)
}

type ErrPasswordRequired struct {
	Reason string
}

func NewPasswordRequiredError(reason string) *ErrPasswordRequired {
	return &ErrPasswordRequired{Reason: reason}
}

func (e *ErrPasswordRequired) Error() string {
	return fmt.Sprintf("password authentication required: %s", e.Reason)
}

type ErrSAMLAuthenticationRequired struct {
	Reason      string
	RedirectURL string
}

func NewSAMLAuthenticationRequiredError(reason string, redirectURL string) *ErrSAMLAuthenticationRequired {
	return &ErrSAMLAuthenticationRequired{Reason: reason, RedirectURL: redirectURL}
}

func (e *ErrSAMLAuthenticationRequired) Error() string {
	return fmt.Sprintf("SAML authentication required: %s", e.Reason)
}

type ErrSAMLConfigurationEmailDomainAlreadyExists struct{ EmailDomain string }

func NewSAMLConfigurationEmailDomainAlreadyExistsError(emailDomain string) error {
	return &ErrSAMLConfigurationEmailDomainAlreadyExists{EmailDomain: emailDomain}
}

func (e ErrSAMLConfigurationEmailDomainAlreadyExists) Error() string {
	return fmt.Sprintf("SAML configuration email domain %q already exists", e.EmailDomain)
}

type ErrNoSCIMConfigurationFound struct{ OrganizationID gid.GID }

func NewNoSCIMConfigurationFoundError(organizationID gid.GID) error {
	return &ErrNoSCIMConfigurationFound{OrganizationID: organizationID}
}

func (e ErrNoSCIMConfigurationFound) Error() string {
	return fmt.Sprintf("SCIM configuration not found for organization %q", e.OrganizationID)
}

type ErrSCIMBridgeNotFound struct{ BridgeID gid.GID }

func NewSCIMBridgeNotFoundError(bridgeID gid.GID) error {
	return &ErrSCIMBridgeNotFound{BridgeID: bridgeID}
}

func (e ErrSCIMBridgeNotFound) Error() string {
	return fmt.Sprintf("SCIM bridge %q not found", e.BridgeID)
}

type ErrConnectorNotFound struct{ ConnectorID gid.GID }

func NewConnectorNotFoundError(connectorID gid.GID) error {
	return &ErrConnectorNotFound{ConnectorID: connectorID}
}

func (e ErrConnectorNotFound) Error() string {
	return fmt.Sprintf("connector %q not found", e.ConnectorID)
}
