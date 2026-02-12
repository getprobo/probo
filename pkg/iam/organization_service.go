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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/scim"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/slug"
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
		LogoFile           *UploadedFile
		HorizontalLogoFile *UploadedFile
		Description        **string
		WebsiteURL         **string
		Email              **string
		HeadquarterAddress **string
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

	CreateInvitationRequest struct {
		Email    mail.Addr
		FullName string
		Role     coredata.MembershipRole
	}

	UpdateProfileRequest struct {
		ID                       gid.GID
		FullName                 string
		AdditionalEmailAddresses mail.Addrs
		Kind                     coredata.MembershipProfileKind
		Position                 **string
		ContractStartDate        **time.Time
		ContractEndDate          **time.Time
	}
)

var (
	proboVendor = struct {
		Name                 string
		Description          string
		LegalName            string
		HeadquarterAddress   string
		WebsiteURL           string
		PrivacyPolicyURL     string
		TermsOfServiceURL    string
		SubprocessorsListURL string
	}{
		Name:                 "Probo",
		Description:          "Probo is an open-source compliance platform that helps startups achieve SOC 2 and ISO 27001 certifications quickly and affordably, with expert guidance and no vendor lock-in.",
		LegalName:            "Probo Inc.",
		HeadquarterAddress:   "490 Post St, Suite 640,San Francisco, CA 94102, United States",
		WebsiteURL:           "https://www.getprobo.com/",
		PrivacyPolicyURL:     "https://www.getprobo.com/privacy",
		TermsOfServiceURL:    "https://www.getprobo.com/terms",
		SubprocessorsListURL: "https://www.getprobo.com/subprocessors",
	}
)

const (
	TokenTypeAPIKey = "api_key"

	NameMaxLength    = 100
	TitleMaxLength   = 1000
	ContentMaxLength = 5000

	DefaultAttributeEmail     = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
	DefaultAttributeFirstname = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"
	DefaultAttributeLastname  = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"
	DefaultAttributeRole      = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role"
)

func (req CreateOrganizationRequest) Validate() error {
	v := validator.New()
	fv := filevalidation.NewValidator(filevalidation.WithCategories(filevalidation.CategoryImage))

	if req.LogoFile != nil {
		err := fv.Validate(req.LogoFile.Filename, req.LogoFile.ContentType, req.LogoFile.Size)
		if err != nil {
			return fmt.Errorf("invalid logo file: %w", err)
		}
	}

	if req.HorizontalLogoFile != nil {
		err := fv.Validate(req.HorizontalLogoFile.Filename, req.HorizontalLogoFile.ContentType, req.HorizontalLogoFile.Size)
		if err != nil {
			return fmt.Errorf("invalid horizontal logo file: %w", err)
		}
	}

	v.Check(req.Name, "name", validator.Required(), validator.SafeTextNoNewLine(255))

	return v.Error()
}

func (req UpdateOrganizationRequest) Validate() error {
	v := validator.New()
	fv := filevalidation.NewValidator(filevalidation.WithCategories(filevalidation.CategoryImage))

	v.Check(req.Name, "name", validator.SafeTextNoNewLine(255))
	v.Check(req.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(req.WebsiteURL, "website_url", validator.SafeText(2048))
	v.Check(req.Email, "email", validator.SafeText(255))
	v.Check(req.HeadquarterAddress, "headquarter_address", validator.SafeText(2048))
	v.Check(req.LogoFile, "logo_file", validator.NotEmpty())
	if req.LogoFile != nil {
		if err := fv.Validate(req.LogoFile.Filename, req.LogoFile.ContentType, req.LogoFile.Size); err != nil {
			return fmt.Errorf("invalid logo file: %w", err)
		}
	}
	v.Check(req.HorizontalLogoFile, "horizontal_logo_file", validator.NotEmpty())
	if req.HorizontalLogoFile != nil {
		if err := fv.Validate(req.HorizontalLogoFile.Filename, req.HorizontalLogoFile.ContentType, req.HorizontalLogoFile.Size); err != nil {
			return fmt.Errorf("invalid horizontal logo file: %w", err)
		}
	}

	return v.Error()
}

func (upr *UpdateProfileRequest) Validate() error {
	v := validator.New()

	v.Check(upr.ID, "id", validator.Required(), validator.GID(coredata.MembershipProfileEntityType))
	v.Check(upr.Kind, "kind", validator.OneOfSlice(coredata.MembershipProfileKinds()))
	v.Check(upr.FullName, "full_name", validator.SafeTextNoNewLine(NameMaxLength))
	v.CheckEach(upr.AdditionalEmailAddresses, "additional_email_addresses", func(index int, item any) {
		v.Check(item, fmt.Sprintf("additional_email_addresses[%d]", index), validator.Required(), validator.NotEmpty())
	})
	v.Check(upr.Position, "position", validator.SafeText(TitleMaxLength))
	v.Check(upr.ContractStartDate, "contract_start_date", validator.Before(upr.ContractEndDate))
	v.Check(upr.ContractEndDate, "contract_end_date", validator.After(upr.ContractStartDate))

	return v.Error()
}

func NewOrganizationService(svc *Service) *OrganizationService {
	return &OrganizationService{Service: svc}
}

func (s *OrganizationService) CountMemberships(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int
	scope := coredata.NewScopeFromObjectID(organizationID)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			memberships := coredata.Memberships{}
			count, err = memberships.CountByOrganizationID(ctx, conn, scope, organizationID, coredata.NewMembershipFilter())
			if err != nil {
				return fmt.Errorf("cannot count memberships: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *OrganizationService) UpdateMempership(
	ctx context.Context,
	organizationID gid.GID,
	membershipID gid.GID,
	role coredata.MembershipRole,
) (*coredata.Membership, error) {
	scope := coredata.NewScopeFromObjectID(organizationID)

	membership := coredata.Membership{}
	if err := s.pg.WithTx(ctx,
		func(tx pg.Conn) error {

			if err := membership.LoadByID(ctx, tx, scope, membershipID); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewMembershipNotFoundError(membershipID)
				}

				return fmt.Errorf("cannot load membership: %w", err)
			}

			if membership.OrganizationID != organizationID {
				return NewMembershipNotFoundError(membership.ID)
			}

			membership.Role = role
			membership.UpdatedAt = time.Now()

			if err := membership.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update membership: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &membership, nil
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

			if membership.Source == coredata.MembershipSourceSCIM {
				return NewMembershipManagedBySCIMError(membershipID)
			}

			if membership.Role == coredata.MembershipRoleOwner && membership.State == coredata.MembershipStateActive {
				memberships := coredata.Memberships{}
				filter := coredata.NewMembershipFilter().
					WithRole(coredata.MembershipRoleOwner).
					WithState(coredata.MembershipStateActive)
				count, err := memberships.CountByOrganizationID(ctx, tx, scope, organizationID, filter)
				if err != nil {
					return fmt.Errorf("cannot count active owners: %w", err)
				}

				if count <= 1 {
					return NewLastActiveOwnerError(membershipID)
				}
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

			switch invitation.Status {
			case coredata.InvitationStatusAccepted:
				return NewInvitationNotDeletedError(invitationID, invitation.Status.String())
			case coredata.InvitationStatusPending, coredata.InvitationStatusExpired:
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
	req *CreateInvitationRequest,
) (*coredata.Invitation, error) {
	var (
		scope      = coredata.NewScopeFromObjectID(organizationID)
		now        = time.Now()
		invitation = &coredata.Invitation{
			ID:             gid.New(organizationID.TenantID(), coredata.InvitationEntityType),
			OrganizationID: organizationID,
			Email:          req.Email,
			FullName:       req.FullName,
			Role:           req.Role,
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

			identity := &coredata.Identity{}
			err = identity.LoadByEmail(ctx, tx, invitation.Email)
			if err != nil && err != coredata.ErrResourceNotFound {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			identityExists := identity.ID != gid.Nil
			if identityExists {
				membership := &coredata.Membership{}
				err = membership.LoadByIdentityAndOrg(ctx, tx, scope, identity.ID, organizationID)
				if err != nil && err != coredata.ErrResourceNotFound {
					return fmt.Errorf("cannot load membership: %w", err)
				}

				if membership.ID != gid.Nil && membership.State == coredata.MembershipStateActive {
					return NewMembershipAlreadyExistsError(identity.ID, organizationID)
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

			emailPresenter := emails.NewPresenter(s.fm, s.bucket, s.baseURL, identity.FullName)

			subject, textBody, htmlBody, err := emailPresenter.RenderInvitation(
				ctx,
				"/auth/signup-from-invitation",
				invitationToken,
				organization.Name,
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

func (s *OrganizationService) CreateOrganization(
	ctx context.Context,
	identityID gid.GID,
	req *CreateOrganizationRequest,
) (*coredata.Organization, *coredata.Membership, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid request: %w", err)
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

		membership = &coredata.Membership{
			ID:             gid.New(tenantID, coredata.MembershipEntityType),
			IdentityID:     identityID,
			OrganizationID: organizationID,
			Role:           coredata.MembershipRoleOwner,
			Source:         coredata.MembershipSourceManual,
			State:          coredata.MembershipStateActive,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		organizationContext = &coredata.OrganizationContext{
			OrganizationID: organizationID,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		trustCenter = &coredata.TrustCenter{
			ID:             gid.New(tenantID, coredata.TrustCenterEntityType),
			OrganizationID: organization.ID,
			TenantID:       organization.TenantID,
			Active:         false,
			Slug:           slug.Make(organization.Name),
			CreatedAt:      now,
			UpdatedAt:      now,
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
			return nil, nil, fmt.Errorf("cannot upload logo file: %w", err)
		}

		logoFile.FileSize = fileSize
	}

	if req.HorizontalLogoFile != nil {
		var (
			fileID      = gid.New(tenantID, coredata.FileEntityType)
			objectKey   = uuid.MustNewV7()
			filename    = req.HorizontalLogoFile.Filename
			contentType = req.HorizontalLogoFile.ContentType
			now         = time.Now()
		)

		horizontalLogoFile = &coredata.File{
			ID:         fileID,
			BucketName: s.bucket,
			MimeType:   contentType,
			FileName:   filename,
			FileKey:    objectKey.String(),
			FileSize:   req.HorizontalLogoFile.Size,
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
			return nil, nil, fmt.Errorf("cannot upload logo file: %w", err)
		}

		horizontalLogoFile.FileSize = fileSize
	}

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			identity := &coredata.Identity{}
			err := identity.LoadByID(ctx, tx, identityID)
			if err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			err = organization.Insert(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot insert organization: %w", err)
			}

			if logoFile != nil {
				err := logoFile.Insert(ctx, tx, scope)
				if err != nil {
					return fmt.Errorf("cannot insert file: %w", err)
				}

				organization.LogoFileID = &logoFile.ID
				trustCenter.LogoFileID = &logoFile.ID
			}

			if horizontalLogoFile != nil {
				err := horizontalLogoFile.Insert(ctx, tx, scope)
				if err != nil {
					return fmt.Errorf("cannot insert file: %w", err)
				}

				organization.HorizontalLogoFileID = &horizontalLogoFile.ID
			}

			err = membership.Insert(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot create membership: %w", err)
			}

			profile := &coredata.MembershipProfile{
				ID:             gid.New(tenantID, coredata.MembershipProfileEntityType),
				IdentityID:     identity.ID,
				OrganizationID: organization.ID,
				MembershipID:   membership.ID,
				FullName:       identity.FullName,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			err = profile.Insert(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot insert profile: %w", err)
			}

			if err := organizationContext.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert organization context: %w", err)
			}

			if err := trustCenter.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert trust center: %w", err)
			}

			proboData := &coredata.Vendor{
				ID:                   gid.New(scope.GetTenantID(), coredata.VendorEntityType),
				TenantID:             organization.TenantID,
				OrganizationID:       organization.ID,
				Name:                 proboVendor.Name,
				Description:          &proboVendor.Description,
				Category:             coredata.VendorCategorySecurity,
				HeadquarterAddress:   &proboVendor.HeadquarterAddress,
				LegalName:            &proboVendor.LegalName,
				WebsiteURL:           &proboVendor.WebsiteURL,
				PrivacyPolicyURL:     &proboVendor.PrivacyPolicyURL,
				TermsOfServiceURL:    &proboVendor.TermsOfServiceURL,
				SubprocessorsListURL: &proboVendor.SubprocessorsListURL,
				ShowOnTrustCenter:    false,
				CreatedAt:            now,
				UpdatedAt:            now,
			}

			if err := proboData.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert vendor: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot insert organization: %w", err)
	}

	return organization, membership, nil
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
		compliancePage     = &coredata.TrustCenter{}
	)

	// TODO: s3 upload happen before we validate the tenantID

	if req.LogoFile != nil {
		var (
			fileID      = gid.New(tenantID, coredata.FileEntityType)
			objectKey   = uuid.MustNewV7()
			filename    = (*req.LogoFile).Filename
			contentType = (*req.LogoFile).ContentType
		)

		logoFile = &coredata.File{
			ID:         fileID,
			BucketName: s.bucket,
			MimeType:   contentType,
			FileName:   filename,
			FileKey:    objectKey.String(),
			FileSize:   (*req.LogoFile).Size,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		fileSize, err := s.fm.PutFile(
			ctx,
			logoFile,
			(*req.LogoFile).Content,
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

	if req.HorizontalLogoFile != nil {
		var (
			fileID      = gid.New(tenantID, coredata.FileEntityType)
			objectKey   = uuid.MustNewV7()
			filename    = (*req.HorizontalLogoFile).Filename
			contentType = (*req.HorizontalLogoFile).ContentType
			now         = time.Now()
		)

		horizontalLogoFile = &coredata.File{
			ID:         fileID,
			BucketName: s.bucket,
			MimeType:   contentType,
			FileName:   filename,
			FileKey:    objectKey.String(),
			FileSize:   (*req.HorizontalLogoFile).Size,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		fileSize, err := s.fm.PutFile(
			ctx,
			horizontalLogoFile,
			(*req.HorizontalLogoFile).Content,
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

			if req.Name != nil {
				organization.Name = *req.Name
			}

			if req.Description != nil {
				organization.Description = *req.Description
			}

			if req.WebsiteURL != nil {
				organization.WebsiteURL = *req.WebsiteURL
			}

			if req.Email != nil {
				if *req.Email != nil {
					if _, err := mail.ParseAddr(**req.Email); err != nil {
						return fmt.Errorf("invalid email address: %w", err)
					}
				}
				organization.Email = *req.Email
			}

			if req.HeadquarterAddress != nil {
				organization.HeadquarterAddress = *req.HeadquarterAddress
			}

			if logoFile != nil {
				if err := logoFile.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert file: %w", err)
				}

				organization.LogoFileID = &logoFile.ID

				// Auto set the compliance page org logo in case it wasn't already specified
				if err := compliancePage.LoadByOrganizationID(ctx, tx, scope, organizationID); err != nil {
					return fmt.Errorf("cannot load compliance page: %w", err)
				}

				if compliancePage.LogoFileID == nil {
					compliancePage.LogoFileID = &logoFile.ID
					compliancePage.UpdatedAt = now

					if err := compliancePage.Update(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot update compliance page: %w", err)
					}
				}
			}

			if horizontalLogoFile != nil {
				err := horizontalLogoFile.Insert(ctx, tx, scope)
				if err != nil {
					return fmt.Errorf("cannot insert file: %w", err)
				}

				organization.HorizontalLogoFileID = &horizontalLogoFile.ID
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
			err := memberships.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor, coredata.NewMembershipFilter())
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

func (s *OrganizationService) UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (*coredata.MembershipProfile, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var (
		scope   = coredata.NewScopeFromObjectID(req.ID)
		profile = &coredata.MembershipProfile{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := profile.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load profile: %w", err)
			}

			profile.FullName = req.FullName
			profile.Kind = req.Kind

			profile.AdditionalEmailAddresses = req.AdditionalEmailAddresses

			if req.Position != nil {
				profile.Position = *req.Position
			}

			if req.ContractStartDate != nil {
				profile.ContractStartDate = *req.ContractStartDate
			}

			if req.ContractEndDate != nil {
				profile.ContractEndDate = *req.ContractEndDate
			}

			profile.UpdatedAt = time.Now()

			if err := profile.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update profile: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *OrganizationService) GetProfile(ctx context.Context, profileID gid.GID) (*coredata.MembershipProfile, error) {
	profile := &coredata.MembershipProfile{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := profile.LoadByID(ctx, conn, coredata.NewScopeFromObjectID(profileID), profileID); err != nil {
				return fmt.Errorf("cannot load profile: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *OrganizationService) ListProfiles(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.MembershipProfileOrderField],
	filter *coredata.MembershipProfileFilter,
) (*page.Page[*coredata.MembershipProfile, coredata.MembershipProfileOrderField], error) {
	var (
		scope    = coredata.NewScopeFromObjectID(organizationID)
		profiles = coredata.MembershipProfiles{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := profiles.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor, filter); err != nil {
				return fmt.Errorf("cannot load profiles: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(profiles, cursor), nil
}

func (s OrganizationService) CountProfiles(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.MembershipProfileFilter,
) (int, error) {
	var (
		scope = coredata.NewScopeFromObjectID(organizationID)
		count int
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			profiles := coredata.MembershipProfiles{}
			count, err = profiles.CountByOrganizationID(ctx, conn, scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count profiles: %w", err)
			}

			return nil
		},
	)

	return count, err
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

func (s OrganizationService) ListSCIMEvents(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.SCIMEventOrderField],
) (*page.Page[*coredata.SCIMEvent, coredata.SCIMEventOrderField], error) {
	var (
		scope      = coredata.NewScopeFromObjectID(organizationID)
		scimEvents = coredata.SCIMEvents{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := scimEvents.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load scim events: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(scimEvents, cursor), nil
}

func (s OrganizationService) CountSCIMEvents(
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
			scimEvents := coredata.SCIMEvents{}
			count, err = scimEvents.CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count scim events: %w", err)
			}

			return nil
		},
	)

	return count, err
}

func (s OrganizationService) GetSCIMConfiguration(
	ctx context.Context,
	organizationID gid.GID,
) (*coredata.SCIMConfiguration, error) {
	var (
		scope  = coredata.NewScopeFromObjectID(organizationID)
		config = &coredata.SCIMConfiguration{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := config.LoadByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewNoSCIMConfigurationFoundError(organizationID)
				}

				return fmt.Errorf("cannot load SCIM configuration: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return config, nil
}

func (s OrganizationService) CreateSCIMConfiguration(
	ctx context.Context,
	organizationID gid.GID,
) (*coredata.SCIMConfiguration, string, error) {
	token, err := scim.GenerateToken()
	if err != nil {
		return nil, "", err
	}

	hashedToken := scim.HashToken(token)
	now := time.Now()

	config := &coredata.SCIMConfiguration{
		ID:             gid.New(organizationID.TenantID(), coredata.SCIMConfigurationEntityType),
		OrganizationID: organizationID,
		HashedToken:    hashedToken,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	scope := coredata.NewScopeFromObjectID(organizationID)

	err = s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			err := config.Insert(ctx, tx, scope)
			if err != nil {
				if err == coredata.ErrResourceAlreadyExists {
					return scim.NewSCIMConfigurationAlreadyExistsError(organizationID)
				}
				return fmt.Errorf("cannot insert SCIM configuration: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return nil, "", err
	}

	return config, token, nil
}

func (s OrganizationService) DeleteSCIMConfiguration(
	ctx context.Context,
	organizationID gid.GID,
	configID gid.GID,
) error {
	scope := coredata.NewScopeFromObjectID(configID)

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			config := &coredata.SCIMConfiguration{}
			err := config.LoadByID(ctx, tx, scope, configID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return scim.NewSCIMConfigurationNotFoundError(configID)
				}

				return fmt.Errorf("cannot load SCIM configuration: %w", err)
			}

			if config.OrganizationID != organizationID {
				return scim.NewSCIMConfigurationNotFoundError(configID)
			}

			memberships := &coredata.Memberships{}
			err = memberships.ResetSCIMSources(ctx, tx, scope, config.OrganizationID)
			if err != nil {
				return fmt.Errorf("cannot reset membership sources: %w", err)
			}

			// Delete SCIM bridge and its connector if they exist
			bridge := &coredata.SCIMBridge{}
			err = bridge.LoadBySCIMConfigurationID(ctx, tx, scope, configID)
			if err != nil && err != coredata.ErrResourceNotFound {
				return fmt.Errorf("cannot load SCIM bridge: %w", err)
			}

			if err == nil {
				// Bridge exists, delete connector if it has one
				if bridge.ConnectorID != nil {
					connector := &coredata.Connector{ID: *bridge.ConnectorID}
					err = connector.Delete(ctx, tx, scope)
					if err != nil && err != coredata.ErrResourceNotFound {
						return fmt.Errorf("cannot delete connector: %w", err)
					}
				}

				// Delete the bridge
				err = bridge.Delete(ctx, tx, scope)
				if err != nil {
					return fmt.Errorf("cannot delete SCIM bridge: %w", err)
				}
			}

			err = config.Delete(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot delete SCIM configuration: %w", err)
			}

			return nil
		},
	)
}

func (s OrganizationService) RegenerateSCIMToken(
	ctx context.Context,
	organizationID gid.GID,
	configID gid.GID,
) (*coredata.SCIMConfiguration, string, error) {
	token, err := scim.GenerateToken()
	if err != nil {
		return nil, "", err
	}

	hashedToken := scim.HashToken(token)
	config := &coredata.SCIMConfiguration{}
	scope := coredata.NewScopeFromObjectID(configID)

	err = s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			err := config.LoadByID(ctx, tx, scope, configID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return scim.NewSCIMConfigurationNotFoundError(configID)
				}

				return fmt.Errorf("cannot load SCIM configuration: %w", err)
			}

			if config.OrganizationID != organizationID {
				return scim.NewSCIMConfigurationNotFoundError(configID)
			}

			config.HashedToken = hashedToken
			config.UpdatedAt = time.Now()

			err = config.Update(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot update SCIM configuration: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, "", err
	}

	return config, token, nil
}

func (s OrganizationService) UpdateSCIMBridge(
	ctx context.Context,
	organizationID gid.GID,
	bridgeID gid.GID,
	excludedUserNames []string,
) (*coredata.SCIMBridge, error) {
	bridge := &coredata.SCIMBridge{}
	scope := coredata.NewScopeFromObjectID(bridgeID)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			err := bridge.LoadByID(ctx, tx, scope, bridgeID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return fmt.Errorf("SCIM bridge not found")
				}

				return fmt.Errorf("cannot load SCIM bridge: %w", err)
			}

			if bridge.OrganizationID != organizationID {
				return fmt.Errorf("SCIM bridge not found")
			}

			bridge.ExcludedUserNames = excludedUserNames
			bridge.UpdatedAt = time.Now()

			err = bridge.Update(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot update SCIM bridge: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return bridge, nil
}

func (s OrganizationService) ListSCIMEventsByConfigID(
	ctx context.Context,
	scimConfigurationID gid.GID,
	cursor *page.Cursor[coredata.SCIMEventOrderField],
) (*page.Page[*coredata.SCIMEvent, coredata.SCIMEventOrderField], error) {
	var (
		scope      = coredata.NewScopeFromObjectID(scimConfigurationID)
		scimEvents = coredata.SCIMEvents{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := scimEvents.LoadBySCIMConfigurationID(ctx, conn, scope, scimConfigurationID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load scim events: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(scimEvents, cursor), nil
}

func (s OrganizationService) CountSCIMEventsByConfigID(
	ctx context.Context,
	scimConfigurationID gid.GID,
) (int, error) {
	var (
		scope = coredata.NewScopeFromObjectID(scimConfigurationID)
		count int
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			scimEvents := coredata.SCIMEvents{}
			count, err = scimEvents.CountBySCIMConfigurationID(ctx, conn, scope, scimConfigurationID)
			if err != nil {
				return fmt.Errorf("cannot count scim events: %w", err)
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
		now                     = time.Now()
		scope                   = coredata.NewScopeFromObjectID(organizationID)
		domainVerificationToken = uuid.MustNewV4().String()
		config                  = &coredata.SAMLConfiguration{
			ID:                      gid.New(scope.GetTenantID(), coredata.SAMLConfigurationEntityType),
			OrganizationID:          organizationID,
			EnforcementPolicy:       coredata.SAMLEnforcementPolicyOff,
			IdPEntityID:             req.IdPEntityID,
			IdPSsoURL:               req.IdPSsoURL,
			IdPCertificate:          req.IdPCertificate,
			DomainVerificationToken: &domainVerificationToken,
			EmailDomain:             req.EmailDomain,
			AutoSignupEnabled:       req.AutoSignupEnabled,
			AttributeEmail:          DefaultAttributeEmail,
			AttributeFirstname:      DefaultAttributeFirstname,
			AttributeLastname:       DefaultAttributeLastname,
			AttributeRole:           DefaultAttributeRole,
			CreatedAt:               now,
			UpdatedAt:               now,
		}
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
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return NewSAMLConfigurationEmailDomainAlreadyExistsError(req.EmailDomain)
				}

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
				if config.DomainVerifiedAt == nil {
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

func (s OrganizationService) GetSCIMBridgeByID(ctx context.Context, bridgeID gid.GID) (*coredata.SCIMBridge, error) {
	var (
		scope  = coredata.NewScopeFromObjectID(bridgeID)
		bridge = &coredata.SCIMBridge{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := bridge.LoadByID(ctx, conn, scope, bridgeID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSCIMBridgeNotFoundError(bridgeID)
				}

				return fmt.Errorf("cannot load SCIM bridge: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return bridge, nil
}

func (s OrganizationService) GetConnectorByID(ctx context.Context, connectorID gid.GID) (*coredata.Connector, error) {
	var (
		scope     = coredata.NewScopeFromObjectID(connectorID)
		connector = &coredata.Connector{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := connector.LoadByID(ctx, conn, scope, connectorID, s.encryptionKey)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewConnectorNotFoundError(connectorID)
				}

				return fmt.Errorf("cannot load connector: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return connector, nil
}

// GetConnectorMetadataByID returns connector metadata without decrypting the connection.
// Use this when you only need provider, organization, or other metadata fields.
func (s OrganizationService) GetConnectorMetadataByID(ctx context.Context, connectorID gid.GID) (*coredata.Connector, error) {
	var (
		scope     = coredata.NewScopeFromObjectID(connectorID)
		connector = &coredata.Connector{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := connector.LoadMetadataByID(ctx, conn, scope, connectorID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewConnectorNotFoundError(connectorID)
				}

				return fmt.Errorf("cannot load connector: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return connector, nil
}

func (s OrganizationService) GetSCIMBridgeByOrganizationID(ctx context.Context, organizationID gid.GID) (*coredata.SCIMBridge, error) {
	var (
		scope  = coredata.NewScopeFromObjectID(organizationID)
		bridge = &coredata.SCIMBridge{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := bridge.LoadByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return nil // No bridge found, not an error
				}

				return fmt.Errorf("cannot load SCIM bridge: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	// If bridge ID is empty, no bridge was found
	if bridge.ID == (gid.GID{}) {
		return nil, nil
	}

	return bridge, nil
}

func (s OrganizationService) CreateSCIMBridge(
	ctx context.Context,
	organizationID gid.GID,
	scimConfigurationID gid.GID,
	connectorID gid.GID,
) (*coredata.SCIMBridge, error) {
	var (
		scope  = coredata.NewScopeFromObjectID(organizationID)
		now    = time.Now()
		bridge *coredata.SCIMBridge
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			organization := &coredata.Organization{}
			err := organization.LoadByID(ctx, tx, scope, organizationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewOrganizationNotFoundError(organizationID)
				}
				return fmt.Errorf("cannot load organization: %w", err)
			}

			config := &coredata.SCIMConfiguration{}
			err = config.LoadByID(ctx, tx, scope, scimConfigurationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return scim.NewSCIMConfigurationNotFoundError(scimConfigurationID)
				}
				return fmt.Errorf("cannot load SCIM configuration: %w", err)
			}

			if config.OrganizationID != organizationID {
				return scim.NewSCIMConfigurationNotFoundError(scimConfigurationID)
			}

			// Load and validate the connector (metadata only, no decryption needed)
			existingConnector := &coredata.Connector{}
			err = existingConnector.LoadMetadataByID(ctx, tx, scope, connectorID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewConnectorNotFoundError(connectorID)
				}

				return fmt.Errorf("cannot load connector: %w", err)
			}

			// Verify connector belongs to the same organization
			if existingConnector.OrganizationID != organizationID {
				return NewConnectorNotFoundError(connectorID)
			}

			// Map connector provider to bridge type
			var bridgeType coredata.SCIMBridgeType
			switch existingConnector.Provider {
			case coredata.ConnectorProviderGoogleWorkspace:
				bridgeType = coredata.SCIMBridgeTypeGoogleWorkspace
			default:
				return fmt.Errorf("connector provider %s is not supported for SCIM bridge", existingConnector.Provider)
			}

			bridge = &coredata.SCIMBridge{
				ID:                  gid.New(organizationID.TenantID(), coredata.SCIMBridgeEntityType),
				OrganizationID:      organizationID,
				ScimConfigurationID: scimConfigurationID,
				ConnectorID:         &connectorID,
				Type:                bridgeType,
				State:               coredata.SCIMBridgeStateActive, // Active immediately since connector already exists
				ExcludedUserNames:   []string{},
				CreatedAt:           now,
				UpdatedAt:           now,
			}

			if err := bridge.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert SCIM bridge: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return bridge, nil
}

func (s OrganizationService) DeleteSCIMBridge(ctx context.Context, organizationID gid.GID, bridgeID gid.GID) error {
	var (
		scope  = coredata.NewScopeFromObjectID(organizationID)
		bridge = &coredata.SCIMBridge{}
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			organization := &coredata.Organization{}
			err := organization.LoadByID(ctx, tx, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := bridge.LoadByID(ctx, tx, scope, bridgeID); err != nil {
				return fmt.Errorf("cannot load SCIM bridge: %w", err)
			}

			if bridge.OrganizationID != organizationID {
				return NewSCIMBridgeNotFoundError(bridgeID)
			}

			if err := bridge.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete SCIM bridge: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return err
	}

	return nil
}
