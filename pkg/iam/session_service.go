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
	"context"
	"fmt"
	"net"
	"time"

	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

type (
	SessionService struct {
		*Service
	}
)

func NewSessionService(svc *Service) *SessionService {
	return &SessionService{Service: svc}
}

type (
	RevokeAllSessionsRequest struct {
		CurrentSessionID gid.GID
	}
)

func (req RevokeAllSessionsRequest) Validate() error {
	v := validator.New()
	v.Check(req.CurrentSessionID, "current_session_id", validator.GID(coredata.SessionEntityType))
	return v.Error()
}

func (s SessionService) GetSession(ctx context.Context, sessionID gid.GID) (*coredata.Session, error) {
	var (
		session = &coredata.Session{}
		now     = time.Now()
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := session.LoadByID(ctx, tx, sessionID); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(sessionID)
				}
			}

			if session.ExpireReason != nil {
				return NewSessionExpiredError(sessionID)
			}

			if now.After(session.ExpiredAt) {
				session.ExpireReason = ref.Ref(coredata.ExpireReasonIdleTimeout)
				session.ExpiredAt = now
				session.UpdatedAt = now
				if err := session.Update(ctx, tx); err != nil {
					return fmt.Errorf("cannot update session: %w", err)
				}

				return NewSessionExpiredError(sessionID)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s SessionService) CloseSession(ctx context.Context, sessionID gid.GID) error {
	return s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			session := &coredata.Session{}
			if err := session.LoadByID(ctx, conn, sessionID); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(sessionID)
				}

				return fmt.Errorf("cannot load session: %w", err)
			}

			if session.ExpireReason != nil {
				return NewSessionExpiredError(sessionID)
			}

			session.ExpireReason = ref.Ref(coredata.ExpireReasonClosed)
			session.ExpiredAt = time.Now()
			session.UpdatedAt = time.Now()
			if err := session.Update(ctx, conn); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(sessionID)
				}

				return fmt.Errorf("cannot update session: %w", err)
			}

			return nil
		},
	)
}

func (s SessionService) RevokeSession(ctx context.Context, identityID gid.GID, sessionID gid.GID) error {
	now := time.Now()

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			identity := &coredata.Identity{}
			err := identity.LoadByID(ctx, tx, identityID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewIdentityNotFoundError(identityID)
				}

				return fmt.Errorf("cannot load identity: %w", err)
			}

			session := &coredata.Session{}
			err = session.LoadByID(ctx, tx, sessionID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(sessionID)
				}

				return fmt.Errorf("cannot load session: %w", err)
			}

			// TODO: move to dedicated query instead of LoadByID
			if session.IdentityID != identityID {
				return NewSessionNotFoundError(sessionID)
			}

			if session.ExpireReason != nil {
				return NewSessionExpiredError(sessionID)
			}

			session.ExpireReason = ref.Ref(coredata.ExpireReasonRevoked)
			session.ExpiredAt = now
			session.UpdatedAt = now
			if err := session.Update(ctx, tx); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(sessionID)
				}

				return fmt.Errorf("cannot update session: %w", err)
			}

			return nil

		},
	)
}

func (s SessionService) RevokeAllSessions(ctx context.Context, currentSessionID gid.GID) (int64, error) {
	var count int64

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			session := coredata.Session{}
			err := session.LoadByID(ctx, tx, currentSessionID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(currentSessionID)
				}

				return fmt.Errorf("cannot load session: %w", err)
			}

			sessions := coredata.Sessions{}
			count, err = sessions.ExpireAllForIdentityExceptOneSession(ctx, tx, session.IdentityID, session.ID)
			if err != nil {
				return fmt.Errorf("cannot expire all sessions: %w", err)
			}

			return nil
		},
	)

	return count, err
}

func (s SessionService) UpdateSessionInfo(ctx context.Context, sessionID gid.GID, userAgent string, ipAddress net.IP) error {
	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			session := &coredata.Session{}
			err := session.LoadByID(ctx, tx, sessionID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(sessionID)
				}

				return fmt.Errorf("cannot load session: %w", err)
			}

			session.UserAgent = userAgent
			session.IPAddress = ipAddress
			session.UpdatedAt = time.Now()

			if err := session.Update(ctx, tx); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(sessionID)
				}

				return fmt.Errorf("cannot update session: %w", err)
			}

			return nil
		},
	)
}

func (s SessionService) UpdateSessionData(ctx context.Context, sessionID gid.GID, data coredata.SessionData) error {
	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			session := &coredata.Session{}
			err := session.LoadByID(ctx, tx, sessionID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(sessionID)
				}

				return fmt.Errorf("cannot load session: %w", err)
			}

			session.Data = data
			session.UpdatedAt = time.Now()

			if err := session.Update(ctx, tx); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(sessionID)
				}

				return fmt.Errorf("cannot update session: %w", err)
			}

			return nil
		},
	)
}

func (s SessionService) GetActiveSessionForMembership(ctx context.Context, rootSessionID gid.GID, membershipID gid.GID) (*coredata.Session, error) {
	childSession := &coredata.Session{}

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			rootSession := &coredata.Session{}
			err := rootSession.LoadByID(ctx, tx, rootSessionID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(rootSessionID)
				}

				return fmt.Errorf("cannot load root session: %w", err)
			}

			if !rootSession.IsRootSession() {
				return fmt.Errorf("session %q is not a root session", rootSessionID)
			}

			if rootSession.ExpireReason != nil || time.Now().After(rootSession.ExpiredAt) {
				return NewSessionExpiredError(rootSessionID)
			}

			membership := &coredata.Membership{}
			err = membership.LoadByID(ctx, tx, coredata.NewScopeFromObjectID(membershipID), membershipID)
			if err != nil {
				return fmt.Errorf("cannot load membership: %w", err)
			}

			err = childSession.LoadByRootSessionIDAndMembershipID(ctx, tx, rootSessionID, membership.ID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(rootSessionID)
				}

				return fmt.Errorf("cannot load child session: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return childSession, nil
}

func (s SessionService) OpenPasswordChildSessionForOrganization(
	ctx context.Context,
	rootSessionID gid.GID,
	organizationID gid.GID,
) (*coredata.Session, *coredata.Membership, error) {
	var (
		now          = time.Now()
		rootSession  = &coredata.Session{}
		identity     = &coredata.Identity{}
		membership   = &coredata.Membership{}
		childSession = &coredata.Session{}
		scope        = coredata.NewScopeFromObjectID(organizationID)
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			err := rootSession.LoadByID(ctx, tx, rootSessionID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(rootSessionID)
				}
				return fmt.Errorf("cannot load session: %w", err)
			}

			if !rootSession.IsRootSession() {
				return fmt.Errorf("session %q is not a root session", rootSessionID)
			}

			if rootSession.ExpireReason != nil || now.After(rootSession.ExpiredAt) {
				return NewSessionExpiredError(rootSessionID)
			}

			err = identity.LoadByID(ctx, tx, rootSession.IdentityID)
			if err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			err = membership.LoadByIdentityInOrganization(ctx, tx, rootSession.IdentityID, organizationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewMembershipNotFoundError(organizationID)
				}
				return fmt.Errorf("cannot load membership: %w", err)
			}

			if membership.State == coredata.MembershipStateInactive {
				return NewMembershipInactiveError(membership.ID)
			}

			tenantID := scope.GetTenantID()
			childSession = &coredata.Session{
				ID:              gid.New(tenantID, coredata.SessionEntityType),
				IdentityID:      rootSession.IdentityID,
				TenantID:        &tenantID,
				MembershipID:    &membership.ID,
				ParentSessionID: &rootSession.ID,
				AuthMethod:      coredata.AuthMethodPassword,
				AuthenticatedAt: now,
				ExpiredAt:       rootSession.ExpiredAt,
				CreatedAt:       now,
				UpdatedAt:       now,
			}

			err = childSession.Insert(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot insert child session: %w", err)
			}

			// Change root session auth method to password
			rootSession.UpdatedAt = now
			rootSession.AuthMethod = coredata.AuthMethodPassword

			if err := rootSession.Update(ctx, tx); err != nil {
				return fmt.Errorf("cannot update root session: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return childSession, membership, nil
}

// OpenSAMLChildSessionForOrganization creates a SAML-authenticated child session for the given
// organization under the provided root session.
//
// This is intended to be used after a successful SAML assertion ("step-up auth") when the user
// might have an existing PASSWORD root session, but we still want a SAML child session for a
// SAML-enabled organization.
func (s SessionService) OpenSAMLChildSessionForOrganization(
	ctx context.Context,
	rootSessionID gid.GID,
	organizationID gid.GID,
) (*coredata.Session, *coredata.Membership, error) {
	var (
		now          = time.Now()
		rootSession  = &coredata.Session{}
		identity     = &coredata.Identity{}
		membership   = &coredata.Membership{}
		childSession = &coredata.Session{}
		scope        = coredata.NewScopeFromObjectID(organizationID)
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			err := rootSession.LoadByID(ctx, tx, rootSessionID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(rootSessionID)
				}
				return fmt.Errorf("cannot load session: %w", err)
			}

			if !rootSession.IsRootSession() {
				return fmt.Errorf("session %q is not a root session", rootSessionID)
			}

			if rootSession.ExpireReason != nil || now.After(rootSession.ExpiredAt) {
				return NewSessionExpiredError(rootSessionID)
			}

			err = identity.LoadByID(ctx, tx, rootSession.IdentityID)
			if err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			err = membership.LoadByIdentityInOrganization(ctx, tx, rootSession.IdentityID, organizationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewMembershipNotFoundError(organizationID)
				}
				return fmt.Errorf("cannot load membership: %w", err)
			}

			if membership.State == coredata.MembershipStateInactive {
				return NewMembershipInactiveError(membership.ID)
			}

			tenantID := scope.GetTenantID()
			childSession = &coredata.Session{
				ID:              gid.New(tenantID, coredata.SessionEntityType),
				IdentityID:      rootSession.IdentityID,
				TenantID:        &tenantID,
				MembershipID:    &membership.ID,
				ParentSessionID: &rootSession.ID,
				AuthMethod:      coredata.AuthMethodSAML,
				AuthenticatedAt: now,
				ExpiredAt:       rootSession.ExpiredAt,
				CreatedAt:       now,
				UpdatedAt:       now,
			}

			err = childSession.Insert(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot insert child session: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return childSession, membership, nil
}

func (s SessionService) AssumeOrganizationSession(
	ctx context.Context,
	sessionID gid.GID,
	organizationID gid.GID,
	continueURL string,
) (*coredata.Session, *coredata.Membership, error) {
	var (
		now          = time.Now()
		rootSession  = &coredata.Session{}
		identity     = &coredata.Identity{}
		membership   = &coredata.Membership{}
		childSession = &coredata.Session{}
		scope        = coredata.NewScopeFromObjectID(organizationID)
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := rootSession.LoadByID(ctx, tx, sessionID); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(sessionID)
				}
				return fmt.Errorf("cannot load session: %w", err)
			}

			if !rootSession.IsRootSession() {
				return fmt.Errorf("session %q is not a root session", sessionID)
			}

			if rootSession.ExpireReason != nil || now.After(rootSession.ExpiredAt) {
				return NewSessionExpiredError(sessionID)
			}

			if err := identity.LoadByID(ctx, tx, rootSession.IdentityID); err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			if err := membership.LoadByIdentityInOrganization(ctx, tx, rootSession.IdentityID, organizationID); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewMembershipNotFoundError(organizationID)
				}
				return fmt.Errorf("cannot load membership: %w", err)
			}

			if membership.State == coredata.MembershipStateInactive {
				return NewMembershipInactiveError(membership.ID)
			}

			samlConfig := &coredata.SAMLConfiguration{}
			err := samlConfig.LoadByOrganizationIDAndEmailDomain(
				ctx,
				tx,
				scope,
				organizationID,
				identity.EmailAddress.Domain(),
			)
			if err != nil && err != coredata.ErrResourceNotFound {
				return fmt.Errorf("cannot load SAML configuration: %w", err)
			}

			if err == nil && samlConfig.EnforcementPolicy == coredata.SAMLEnforcementPolicyRequired {
				if rootSession.AuthMethod != coredata.AuthMethodSAML {
					redirectURL, err := s.SAMLService.InitiateLogin(ctx, samlConfig.ID, continueURL)
					if err != nil {
						return fmt.Errorf("cannot initiate SAML login: %w", err)
					}

					return NewSAMLAuthenticationRequiredError("policy_requirement", redirectURL.String())
				}
			} else if err == nil && samlConfig.EnforcementPolicy == coredata.SAMLEnforcementPolicyOptional {
				// SAML is optional: both PASSWORD and SAML root sessions are allowed.
			} else if rootSession.AuthMethod != coredata.AuthMethodPassword {
				// No (or non-required) SAML configuration: require a password-authenticated root session
				// (eg. when switching into a password-based org from a SAML login).
				return NewPasswordRequiredError("password_authentication_required")
			}

			tenantID := scope.GetTenantID()
			childSession = &coredata.Session{
				ID:              gid.New(tenantID, coredata.SessionEntityType),
				IdentityID:      rootSession.IdentityID,
				TenantID:        &tenantID,
				MembershipID:    &membership.ID,
				ParentSessionID: &rootSession.ID,
				AuthMethod:      rootSession.AuthMethod,
				AuthenticatedAt: now,
				ExpiredAt:       rootSession.ExpiredAt,
				CreatedAt:       now,
				UpdatedAt:       now,
			}

			err = childSession.Insert(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot insert child session: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return childSession, membership, nil
}
