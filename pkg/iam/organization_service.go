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
	"errors"
	"fmt"
	"io"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/statelesstoken"
	"go.probo.inc/probo/pkg/validator"
)

type (
	OrganizationService struct {
		*Service
	}

	InvitationTokenData struct {
		InvitationID gid.GID `json:"invitation_id"`
	}

	UploadedFile struct {
		Content     io.Reader
		Filename    string
		Size        int64
		ContentType string
	}

	CreateOrganizationRequest struct {
		Name               string
		LogoFile           *UploadedFile
		HorizontalLogoFile *UploadedFile
	}

	UpdateOrganizationRequest struct {
		Name               *string
		LogoFile           **UploadedFile
		HorizontalLogoFile **UploadedFile
	}

	CreateSAMLConfigurationRequest struct {
		EmailDomain        string
		IdPEntityID        string
		IdPSsoURL          string
		IdPCertificate     string
		AttributeEmail     *string
		AttributeFirstname *string
		AttributeLastname  *string
		AttributeRole      *string
		AutoSignupEnabled  bool
	}

	UpdateSAMLConfigurationRequest struct {
		ID                 gid.GID
		EnforcementPolicy  *coredata.SAMLEnforcementPolicy
		IdPEntityID        *string
		IdPSsoURL          *string
		IdPCertificate     *string
		AttributeEmail     *string
		AttributeFirstname *string
		AttributeLastname  *string
		AttributeRole      *string
		AutoSignupEnabled  *bool
	}
)

const (
	TokenTypeAPIKey = "api_key"

	DefaultAttributeEmail     = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
	DefaultAttributeFirstname = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"
	DefaultAttributeLastname  = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"
	DefaultAttributeRole      = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role"
)

func (req CreateOrganizationRequest) Validate() error {
	v := validator.New()
	fv := filevalidation.NewValidator(filevalidation.WithCategories(filevalidation.CategoryImage))

	err := fv.Validate(req.LogoFile.Filename, req.LogoFile.ContentType, req.LogoFile.Size)
	if err != nil {
		return fmt.Errorf("invalid logo file: %w", err)
	}

	err = fv.Validate(req.HorizontalLogoFile.Filename, req.HorizontalLogoFile.ContentType, req.HorizontalLogoFile.Size)
	if err != nil {
		return fmt.Errorf("invalid horizontal logo file: %w", err)
	}

	v.Check(req.Name, "name", validator.Required(), validator.SafeTextNoNewLine(255))

	return v.Error()
}

func (req UpdateOrganizationRequest) Validate() error {
	v := validator.New()
	fv := filevalidation.NewValidator(filevalidation.WithCategories(filevalidation.CategoryImage))

	if req.LogoFile != nil && *req.LogoFile != nil {
		err := fv.Validate((**req.LogoFile).Filename, (**req.LogoFile).ContentType, (**req.LogoFile).Size)
		if err != nil {
			return fmt.Errorf("invalid logo file: %w", err)
		}
	}

	if req.HorizontalLogoFile != nil && *req.HorizontalLogoFile != nil {
		err := fv.Validate((**req.HorizontalLogoFile).Filename, (**req.HorizontalLogoFile).ContentType, (**req.HorizontalLogoFile).Size)
		if err != nil {
			return fmt.Errorf("invalid horizontal logo file: %w", err)
		}
	}

	v.Check(req.Name, "name", validator.Required(), validator.SafeTextNoNewLine(255))

	return v.Error()
}

func NewOrganizationService(svc *Service) *OrganizationService {
	return &OrganizationService{Service: svc}
}

func (s *OrganizationService) RemoveMember(
	ctx context.Context,
	organizationID gid.GID,
	membershipID gid.GID,
) error {
	scope := coredata.NewScopeFromObjectID(organizationID)

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			membership := coredata.Membership{}

			if err := membership.LoadByID(ctx, tx, scope, membershipID); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewMembershipNotFoundError(membershipID)
				}

				return fmt.Errorf("cannot load membership: %w", err)
			}

			if membership.OrganizationID != organizationID {
				return NewMembershipNotFoundError(membership.ID)
			}

			err := membership.Delete(ctx, tx, scope, membershipID)
			if err != nil {
				return fmt.Errorf("cannot delete membership: %w", err)
			}

			return nil
		},
	)
}

func (s *OrganizationService) DeleteInvitation(
	ctx context.Context,
	organizationID gid.GID,
	invitationID gid.GID,
) error {
	scope := coredata.NewScopeFromObjectID(organizationID)

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			invitation := coredata.Invitation{}
			err := invitation.LoadByID(ctx, tx, scope, invitationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewInvitationNotFoundError(invitationID)
				}

				return fmt.Errorf("cannot load invitation: %w", err)
			}

			if invitation.Status != coredata.InvitationStatusPending {
				return NewInvitationNotPendingError(invitationID)
			}

			err = invitation.Delete(ctx, tx, scope, invitationID)
			if err != nil {
				return fmt.Errorf("cannot delete invitation: %w", err)
			}

			return nil
		},
	)
}

func (s *OrganizationService) ListInvitations(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.InvitationOrderField],
	filter *coredata.InvitationFilter,
) (*page.Page[*coredata.Invitation, coredata.InvitationOrderField], error) {
	var (
		invitations coredata.Invitations
		scope       = coredata.NewScopeFromObjectID(organizationID)
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := invitations.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load invitations: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(invitations, cursor), nil
}

func (s *OrganizationService) CountInvitations(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.InvitationFilter,
) (int, error) {
	var (
		count int
		scope = coredata.NewScopeFromObjectID(organizationID)
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			invitations := coredata.Invitations{}
			count, err = invitations.CountByOrganizationID(ctx, conn, scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count invitations: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *OrganizationService) InviteMember(
	ctx context.Context,
	organizationID gid.GID,
	emailAddress mail.Addr,
	fullName string,
	role coredata.MembershipRole,
) (*coredata.Invitation, error) {
	var (
		scope      = coredata.NewScopeFromObjectID(organizationID)
		now        = time.Now()
		invitation = &coredata.Invitation{
			ID:             gid.New(organizationID.TenantID(), coredata.InvitationEntityType),
			OrganizationID: organizationID,
			Email:          emailAddress,
			FullName:       fullName,
			Role:           role,
			Status:         coredata.InvitationStatusPending,
			ExpiresAt:      now.Add(s.invitationTokenValidity),
			CreatedAt:      now,
		}
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			organization := coredata.Organization{}
			err := organization.LoadByID(ctx, tx, scope, organizationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewOrganizationNotFoundError(organizationID)
				}

				return fmt.Errorf("cannot load organization: %w", err)
			}

			user := &coredata.User{}
			err = user.LoadByEmail(ctx, tx, emailAddress)
			if err != nil && err != coredata.ErrResourceNotFound {
				return fmt.Errorf("cannot load user: %w", err)
			}

			userExists := user.ID != gid.Nil
			if userExists {
				membership := &coredata.Membership{}
				err = membership.LoadByUserAndOrg(ctx, tx, scope, user.ID, organizationID)
				if err != nil && err != coredata.ErrResourceNotFound {
					return fmt.Errorf("cannot load membership: %w", err)
				}

				if membership.ID != gid.Nil {
					return NewMembershipAlreadyExistsError(user.ID, organizationID)
				}
			}

			err = invitation.Insert(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot insert invitation: %w", err)
			}

			invitationToken, err := statelesstoken.NewToken(
				s.tokenSecret,
				TokenTypeOrganizationInvitation,
				s.invitationTokenValidity,
				InvitationTokenData{InvitationID: invitation.ID},
			)
			if err != nil {
				return fmt.Errorf("cannot generate invitation token: %w", err)
			}

			baseurl, err := baseurl.Parse(s.baseURL)
			if err != nil {
				return fmt.Errorf("cannot parse base URL: %w", err)
			}

			invitationURL := baseurl.WithPath("/auth/signup-from-invitation").WithQuery("token", invitationToken).MustString()

			subject, textBody, htmlBody, err := emails.RenderInvitation(
				s.baseURL,
				invitation.FullName,
				organization.Name,
				invitationURL,
			)
			if err != nil {
				return fmt.Errorf("cannot render invitation email: %w", err)
			}

			email := coredata.NewEmail(
				invitation.FullName,
				invitation.Email,
				subject,
				textBody,
				htmlBody,
			)

			err = email.Insert(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return invitation, nil
}

func (s *OrganizationService) CreateOrganization(ctx context.Context, identityID gid.GID, req *CreateOrganizationRequest) (*coredata.Organization, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var (
		tenantID       = gid.NewTenantID()
		organizationID = gid.New(tenantID, coredata.OrganizationEntityType)
		now            = time.Now()
		organization   = &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      req.Name,
			CreatedAt: now,
			UpdatedAt: now,
		}
		logoFile           *coredata.File
		horizontalLogoFile *coredata.File
		scope              = coredata.NewScope(tenantID)
	)

	if req.LogoFile != nil {
		var (
			fileID      = gid.New(tenantID, coredata.FileEntityType)
			objectKey   = uuid.MustNewV7()
			filename    = req.LogoFile.Filename
			contentType = req.LogoFile.ContentType
			now         = time.Now()
		)

		logoFile = &coredata.File{
			ID:         fileID,
			BucketName: s.bucket,
			MimeType:   contentType,
			FileName:   filename,
			FileKey:    objectKey.String(),
			FileSize:   req.LogoFile.Size,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		fileSize, err := s.fm.PutFile(
			ctx,
			logoFile,
			req.LogoFile.Content,
			map[string]string{
				"file-id":         fileID.String(),
				"organization-id": organization.ID.String(),
			},
		)

		if err != nil {
			return nil, fmt.Errorf("cannot upload logo file: %w", err)
		}

		logoFile.FileSize = fileSize
	}

	if req.HorizontalLogoFile != nil {
		var (
			fileID      = gid.New(tenantID, coredata.FileEntityType)
			objectKey   = uuid.MustNewV7()
			filename    = req.LogoFile.Filename
			contentType = req.LogoFile.ContentType
			now         = time.Now()
		)

		horizontalLogoFile = &coredata.File{
			ID:         fileID,
			BucketName: s.bucket,
			MimeType:   contentType,
			FileName:   filename,
			FileKey:    objectKey.String(),
			FileSize:   req.LogoFile.Size,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		fileSize, err := s.fm.PutFile(
			ctx,
			horizontalLogoFile,
			req.HorizontalLogoFile.Content,
			map[string]string{
				"file-id":         fileID.String(),
				"organization-id": organization.ID.String(),
			},
		)

		if err != nil {
			return nil, fmt.Errorf("cannot upload logo file: %w", err)
		}

		horizontalLogoFile.FileSize = fileSize
	}

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {

			err := organization.Insert(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot insert organization: %w", err)
			}

			if logoFile != nil {
				err := logoFile.Insert(ctx, tx, scope)
				if err != nil {
					return fmt.Errorf("cannot insert file: %w", err)
				}
			}

			if horizontalLogoFile != nil {
				err := horizontalLogoFile.Insert(ctx, tx, scope)
				if err != nil {
					return fmt.Errorf("cannot insert file: %w", err)
				}
			}

			membership := &coredata.Membership{
				ID:             gid.New(tenantID, coredata.MembershipEntityType),
				UserID:         identityID,
				OrganizationID: organizationID,
				Role:           coredata.MembershipRoleOwner,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			err = membership.Insert(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot create membership: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot insert organization: %w", err)
	}

	return organization, nil
}

func (s *OrganizationService) UpdateOrganization(ctx context.Context, organizationID gid.GID, req *UpdateOrganizationRequest) (*coredata.Organization, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var (
		now                = time.Now()
		logoFile           *coredata.File
		horizontalLogoFile *coredata.File
		tenantID           = organizationID.TenantID()
		scope              = coredata.NewScopeFromObjectID(organizationID)
		organization       = &coredata.Organization{}
	)

	// TODO: s3 upload happen before we validate the tenantID

	if req.LogoFile != nil && *req.LogoFile != nil {
		var (
			fileID      = gid.New(tenantID, coredata.FileEntityType)
			objectKey   = uuid.MustNewV7()
			filename    = (**req.LogoFile).Filename
			contentType = (**req.LogoFile).ContentType
		)

		logoFile = &coredata.File{
			ID:         fileID,
			BucketName: s.bucket,
			MimeType:   contentType,
			FileName:   filename,
			FileKey:    objectKey.String(),
			FileSize:   (**req.LogoFile).Size,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		fileSize, err := s.fm.PutFile(
			ctx,
			logoFile,
			(**req.LogoFile).Content,
			map[string]string{
				"file-id":         fileID.String(),
				"organization-id": organizationID.String(),
			},
		)

		if err != nil {
			return nil, fmt.Errorf("cannot upload logo file: %w", err)
		}

		logoFile.FileSize = fileSize
	}

	if req.HorizontalLogoFile != nil && *req.HorizontalLogoFile != nil {
		var (
			fileID      = gid.New(tenantID, coredata.FileEntityType)
			objectKey   = uuid.MustNewV7()
			filename    = (**req.HorizontalLogoFile).Filename
			contentType = (**req.HorizontalLogoFile).ContentType
			now         = time.Now()
		)

		horizontalLogoFile = &coredata.File{
			ID:         fileID,
			BucketName: s.bucket,
			MimeType:   contentType,
			FileName:   filename,
			FileKey:    objectKey.String(),
			FileSize:   (**req.HorizontalLogoFile).Size,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		fileSize, err := s.fm.PutFile(
			ctx,
			horizontalLogoFile,
			(**req.HorizontalLogoFile).Content,
			map[string]string{
				"file-id":         fileID.String(),
				"organization-id": organizationID.String(),
			},
		)

		if err != nil {
			return nil, fmt.Errorf("cannot upload logo file: %w", err)
		}

		horizontalLogoFile.FileSize = fileSize
	}

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			err := organization.LoadByID(ctx, tx, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			organization.UpdatedAt = now

			if req.Name != nil {
				organization.Name = *req.Name
			}

			if req.LogoFile != nil {
				if *req.LogoFile != nil {
					err := logoFile.Insert(ctx, tx, scope)
					if err != nil {
						return fmt.Errorf("cannot insert file: %w", err)
					}

					organization.LogoFileID = &logoFile.ID
				} else {
					organization.LogoFileID = nil
				}

			}

			if req.HorizontalLogoFile != nil {
				if *req.HorizontalLogoFile != nil {
					err := horizontalLogoFile.Insert(ctx, tx, scope)
					if err != nil {
						return fmt.Errorf("cannot insert file: %w", err)
					}
					organization.HorizontalLogoFileID = &horizontalLogoFile.ID
				} else {
					organization.HorizontalLogoFileID = nil
				}
			}

			err = organization.Update(ctx, scope, tx)
			if err != nil {
				return fmt.Errorf("cannot update organization: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (s *OrganizationService) DeleteOrganization(ctx context.Context, organizationID gid.GID) error {
	scope := coredata.NewScopeFromObjectID(organizationID)

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			organization := &coredata.Organization{}
			err := organization.LoadByID(ctx, tx, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			err = organization.Delete(ctx, tx, organizationID)
			if err != nil {
				return fmt.Errorf("cannot delete organization: %w", err)
			}

			return nil
		},
	)
}

func (s *OrganizationService) ListMembers(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.MembershipOrderField],
) (*page.Page[*coredata.Membership, coredata.MembershipOrderField], error) {
	var (
		scope       = coredata.NewScopeFromObjectID(organizationID)
		memberships = coredata.Memberships{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := memberships.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load memberships: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(memberships, cursor), nil
}

func (s *OrganizationService) GetOrganizationForMembership(ctx context.Context, membershipID gid.GID) (*coredata.Organization, error) {
	var (
		scope        = coredata.NewScopeFromObjectID(membershipID)
		organization = &coredata.Organization{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			membership := &coredata.Membership{}
			err := membership.LoadByID(ctx, conn, scope, membershipID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewMembershipNotFoundError(membershipID)
				}

				return fmt.Errorf("cannot load membership: %w", err)
			}

			err = organization.LoadByID(ctx, conn, scope, membership.OrganizationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewOrganizationNotFoundError(membership.OrganizationID)
				}

				return fmt.Errorf("cannot load organization: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (s *OrganizationService) GetOrganizationForInvitation(ctx context.Context, invitationID gid.GID) (*coredata.Organization, error) {
	var (
		scope        = coredata.NewScopeFromObjectID(invitationID)
		organization = &coredata.Organization{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			invitation := &coredata.Invitation{}
			err := invitation.LoadByID(ctx, conn, scope, invitationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewInvitationNotFoundError(invitationID)
				}

				return fmt.Errorf("cannot load invitation: %w", err)
			}

			err = organization.LoadByID(ctx, conn, scope, invitation.OrganizationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewOrganizationNotFoundError(invitation.OrganizationID)
				}

				return fmt.Errorf("cannot load organization: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (s OrganizationService) GenerateLogoURL(
	ctx context.Context,
	organizationID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	var (
		errNoLogoFile = errors.New("no logo file found")
		scope         = coredata.NewScopeFromObjectID(organizationID)
		file          = &coredata.File{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if organization.LogoFileID == nil {
				return errNoLogoFile
			}

			if err := file.LoadByID(ctx, conn, scope, *organization.LogoFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err == errNoLogoFile {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("cannot generate logo URL: %w", err)
	}

	presignedURL, err := s.fm.GenerateFileUrl(ctx, file, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return &presignedURL, nil
}

func (s OrganizationService) GenerateHorizontalLogoURL(
	ctx context.Context,
	organizationID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	var (
		errNoLogoFile = errors.New("no logo file found")
		scope         = coredata.NewScopeFromObjectID(organizationID)
		file          = &coredata.File{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if organization.HorizontalLogoFileID == nil {
				return errNoLogoFile
			}

			if err := file.LoadByID(ctx, conn, scope, *organization.HorizontalLogoFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err == errNoLogoFile {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	presignedURL, err := s.fm.GenerateFileUrl(ctx, file, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return &presignedURL, nil
}

func (s OrganizationService) DeleteSAMLConfiguration(
	ctx context.Context,
	organizationID gid.GID,
	configID gid.GID,
) error {
	scope := coredata.NewScopeFromObjectID(organizationID)

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			var config coredata.SAMLConfiguration
			if err := config.LoadByID(ctx, tx, scope, configID); err != nil {
				return fmt.Errorf("cannot load saml configuration: %w", err)
			}

			if config.OrganizationID != organizationID {
				return NewSAMLConfigurationNotFoundError(configID)
			}

			if err := config.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete saml configuration: %w", err)
			}

			return nil
		},
	)
}

func (s OrganizationService) ListSAMLConfigurations(
	ctx context.Context,
	organizationID gid.GID, cursor *page.Cursor[coredata.SAMLConfigurationOrderField],
) (*page.Page[*coredata.SAMLConfiguration, coredata.SAMLConfigurationOrderField], error) {
	var (
		scope              = coredata.NewScopeFromObjectID(organizationID)
		samlConfigurations = coredata.SAMLConfigurations{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := samlConfigurations.LoadByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot load saml configurations: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(samlConfigurations, cursor), nil
}

func (s OrganizationService) CountSAMLConfigurations(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var (
		scope = coredata.NewScopeFromObjectID(organizationID)
		count int
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			samlConfigurations := coredata.SAMLConfigurations{}
			count, err = samlConfigurations.CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count saml configurations: %w", err)
			}

			return nil
		},
	)

	return count, err
}

func (s OrganizationService) CreateSAMLConfiguration(
	ctx context.Context,
	organizationID gid.GID,
	req *CreateSAMLConfigurationRequest,
) (*coredata.SAMLConfiguration, error) {
	var (
		now    = time.Now()
		scope  = coredata.NewScopeFromObjectID(organizationID)
		config = &coredata.SAMLConfiguration{
			ID:                 gid.New(scope.GetTenantID(), coredata.SAMLConfigurationEntityType),
			OrganizationID:     organizationID,
			EnforcementPolicy:  coredata.SAMLEnforcementPolicyOff,
			IdPEntityID:        req.IdPEntityID,
			IdPSsoURL:          req.IdPSsoURL,
			IdPCertificate:     req.IdPCertificate,
			EmailDomain:        req.EmailDomain,
			AutoSignupEnabled:  req.AutoSignupEnabled,
			AttributeEmail:     DefaultAttributeEmail,
			AttributeFirstname: DefaultAttributeFirstname,
			AttributeLastname:  DefaultAttributeLastname,
			AttributeRole:      DefaultAttributeRole,
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		// TODO create domain verification object
	)

	if req.AttributeEmail != nil {
		config.AttributeEmail = *req.AttributeEmail
	}

	if req.AttributeFirstname != nil {
		config.AttributeFirstname = *req.AttributeFirstname
	}

	if req.AttributeLastname != nil {
		config.AttributeLastname = *req.AttributeLastname
	}

	if req.AttributeRole != nil {
		config.AttributeRole = *req.AttributeRole
	}

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			organization := &coredata.Organization{}
			err := organization.LoadByID(ctx, tx, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			err = config.Insert(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot insert saml configuration: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (s OrganizationService) UpdateSAMLConfiguration(
	ctx context.Context,
	organizationID gid.GID,
	configID gid.GID,
	req *UpdateSAMLConfigurationRequest,
) (*coredata.SAMLConfiguration, error) {
	var (
		scope  = coredata.NewScopeFromObjectID(organizationID)
		config = &coredata.SAMLConfiguration{}
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			organization := &coredata.Organization{}
			err := organization.LoadByID(ctx, tx, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			config = &coredata.SAMLConfiguration{}
			err = config.LoadByID(ctx, tx, scope, configID)
			if err != nil {
				return fmt.Errorf("cannot load saml configuration: %w", err)
			}

			if req.EnforcementPolicy != nil {
				if !config.DomainVerified {
					return NewSAMLConfigurationDomainNotVerifiedError(configID)
				}

				config.EnforcementPolicy = *req.EnforcementPolicy
			}

			if req.IdPEntityID != nil {
				config.IdPEntityID = *req.IdPEntityID
			}

			if req.IdPSsoURL != nil {
				config.IdPSsoURL = *req.IdPSsoURL
			}

			if req.IdPCertificate != nil {
				config.IdPCertificate = *req.IdPCertificate
			}

			if req.AttributeEmail != nil {
				config.AttributeEmail = *req.AttributeEmail
			}

			if req.AttributeFirstname != nil {
				config.AttributeFirstname = *req.AttributeFirstname
			}

			if req.AttributeLastname != nil {
				config.AttributeLastname = *req.AttributeLastname
			}

			if req.AttributeRole != nil {
				config.AttributeRole = *req.AttributeRole
			}

			if req.AutoSignupEnabled != nil {
				config.AutoSignupEnabled = *req.AutoSignupEnabled
			}

			config.UpdatedAt = time.Now()

			err = config.Update(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot update saml configuration: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return config, nil

}

func (s OrganizationService) GetOrganization(ctx context.Context, organizationID gid.GID) (*coredata.Organization, error) {
	var (
		scope        = coredata.NewScopeFromObjectID(organizationID)
		organization = &coredata.Organization{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := organization.LoadByID(ctx, conn, scope, organizationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewOrganizationNotFoundError(organizationID)
				}

				return fmt.Errorf("cannot load organization: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return organization, nil
}
