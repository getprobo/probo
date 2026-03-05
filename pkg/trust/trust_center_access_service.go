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

package trust

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mail"
)

type (
	TrustCenterAccessService struct {
		svc    *TenantService
		iamSvc *iam.Service
		logger *log.Logger
	}

	TrustCenterAccessRequest struct {
		TrustCenterID      gid.GID
		IdentityID         gid.GID
		DocumentIDs        []gid.GID
		ReportIDs          []gid.GID
		TrustCenterFileIDs []gid.GID
	}
)

const (
	TrustCenterAccessURLFormat = "https://%s/organizations/%s/trust-center/access"
)

func (s TrustCenterAccessService) ensureAccessInTx(
	ctx context.Context,
	tx pg.Conn,
	trustCenterID gid.GID,
	identityID gid.GID,
) (*coredata.TrustCenterAccess, *coredata.TrustCenter, error) {
	now := time.Now()

	trustCenter := &coredata.TrustCenter{}
	if err := trustCenter.LoadByID(ctx, tx, s.svc.scope, trustCenterID); err != nil {
		return nil, nil, fmt.Errorf("cannot load trust center: %w", err)
	}

	identity := &coredata.Identity{}
	if err := identity.LoadByID(ctx, tx, identityID); err != nil {
		return nil, nil, fmt.Errorf("cannot load identity: %w", err)
	}

	existingAccess := &coredata.TrustCenterAccess{}
	err := existingAccess.LoadByTrustCenterIDAndIdentityID(ctx, tx, s.svc.scope, trustCenterID, identityID)
	if err == nil {
		return existingAccess, trustCenter, nil
	}

	if !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, nil, fmt.Errorf("cannot load trust center access: %w", err)
	}

	access := &coredata.TrustCenterAccess{
		ID:             gid.New(s.svc.scope.GetTenantID(), coredata.TrustCenterAccessEntityType),
		OrganizationID: trustCenter.OrganizationID,
		TenantID:       s.svc.scope.GetTenantID(),
		IdentityID:     identityID,
		TrustCenterID:  trustCenterID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if trustCenter.NonDisclosureAgreementFileID != nil && s.svc.esign != nil {
		sig, err := s.svc.esign.CreateSignature(
			ctx,
			tx,
			&esign.CreateSignatureRequest{
				OrganizationID: access.OrganizationID,
				DocumentType:   coredata.ElectronicSignatureDocumentTypeNDA,
				FileID:         *trustCenter.NonDisclosureAgreementFileID,
				SignerEmail:    identity.EmailAddress,
			},
		)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot create pending signature: %w", err)
		}
		access.ElectronicSignatureID = &sig.ID
	}

	if err := access.Insert(ctx, tx, s.svc.scope); err != nil {
		return nil, nil, fmt.Errorf("cannot insert trust center access: %w", err)
	}

	return access, trustCenter, nil
}

func (s TrustCenterAccessService) EnsureAccess(
	ctx context.Context,
	trustCenterID gid.GID,
	identityID gid.GID,
) (*coredata.TrustCenterAccess, error) {
	var access *coredata.TrustCenterAccess

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			now := time.Now()

			access, _, err := s.ensureAccessInTx(ctx, tx, trustCenterID, identityID)
			if err != nil {
				return fmt.Errorf("cannot ensure access presence: %w", err)
			}

			identity := &coredata.Identity{}
			if err := identity.LoadByID(ctx, tx, identityID); err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			profile := &coredata.MembershipProfile{}
			if err := profile.LoadByIdentityIDAndOrganizationID(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(access.ID),
				identityID,
				access.OrganizationID,
			); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load profile: %w", err)
				}

				profile = &coredata.MembershipProfile{
					ID:             gid.New(access.TenantID, coredata.MembershipProfileEntityType),
					IdentityID:     identityID,
					OrganizationID: access.OrganizationID,
					EmailAddress:   identity.EmailAddress,
					Source:         coredata.ProfileSourceManual,
					State:          coredata.ProfileStateActive,
					FullName:       identity.FullName,
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				if err := profile.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot insert profile: %w", err)
				}
			}

			return err
		},
	)

	return access, err
}

func (s TrustCenterAccessService) Request(
	ctx context.Context,
	req *TrustCenterAccessRequest,
) (*coredata.TrustCenterAccess, error) {
	var (
		now    = time.Now()
		access *coredata.TrustCenterAccess
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, tx, s.svc.scope, req.TrustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			access = &coredata.TrustCenterAccess{}
			if err := access.LoadByTrustCenterIDAndIdentityID(ctx, tx, s.svc.scope, req.TrustCenterID, req.IdentityID); err != nil {
				return fmt.Errorf("cannot load compliance page membership: %w", err)
			}

			organizationID := trustCenter.OrganizationID

			documentIDs := req.DocumentIDs
			if req.DocumentIDs == nil {
				var allDocuments coredata.Documents
				filter := coredata.NewDocumentTrustCenterFilter()

				if err := allDocuments.LoadAllByOrganizationID(ctx, tx, s.svc.scope, organizationID, filter); err != nil {
					return fmt.Errorf("cannot list documents: %w", err)
				}

				for _, doc := range allDocuments {
					documentIDs = append(documentIDs, doc.ID)
				}
			}

			reportIDs := req.ReportIDs
			if req.ReportIDs == nil {
				var allAudits coredata.Audits
				auditFilter := coredata.NewAuditTrustCenterFilter()

				if err := allAudits.LoadAllByOrganizationID(ctx, tx, s.svc.scope, organizationID, auditFilter); err != nil {
					return fmt.Errorf("cannot list audits: %w", err)
				}

				for _, audit := range allAudits {
					if audit.ReportID != nil {
						reportIDs = append(reportIDs, *audit.ReportID)
					}
				}
			}

			trustCenterFileIDs := req.TrustCenterFileIDs
			if req.TrustCenterFileIDs == nil {
				var allTrustCenterFiles coredata.TrustCenterFiles
				filter := coredata.NewTrustCenterFileFilter(
					coredata.WithTrustCenterFileVisibilities(coredata.TrustCenterVisibilityPrivate, coredata.TrustCenterVisibilityNone),
				)

				if err := allTrustCenterFiles.LoadAllByOrganizationID(ctx, tx, s.svc.scope, organizationID, filter); err != nil {
					return fmt.Errorf("cannot list trust center files: %w", err)
				}

				for _, file := range allTrustCenterFiles {
					trustCenterFileIDs = append(trustCenterFileIDs, file.ID)
				}
			}

			var existingAccesses coredata.TrustCenterDocumentAccesses
			if err := existingAccesses.LoadAllByTrustCenterAccessID(ctx, tx, s.svc.scope, access.ID); err != nil {
				return fmt.Errorf("cannot load existing access records: %w", err)
			}

			existingDocumentIDs, existingReportIDs, existingTrustCenterFileIDs := extractExistingIDs(existingAccesses)
			newDocumentIDs := filterExistingIDs(documentIDs, existingDocumentIDs)
			newReportIDs := filterExistingIDs(reportIDs, existingReportIDs)
			newTrustCenterFileIDs := filterExistingIDs(trustCenterFileIDs, existingTrustCenterFileIDs)

			var accesses coredata.TrustCenterDocumentAccesses

			if err := accesses.BulkInsertDocumentAccesses(
				ctx,
				tx,
				s.svc.scope,
				access.ID,
				access.OrganizationID,
				newDocumentIDs,
				coredata.TrustCenterDocumentAccessStatusRequested,
				now,
			); err != nil {
				return fmt.Errorf("cannot bulk insert trust center document accesses: %w", err)
			}

			if err := accesses.BulkInsertReportAccesses(
				ctx,
				tx,
				s.svc.scope,
				access.ID,
				access.OrganizationID,
				newReportIDs,
				coredata.TrustCenterDocumentAccessStatusRequested,
				now,
			); err != nil {
				return fmt.Errorf("cannot bulk insert trust center report accesses: %w", err)
			}

			if err := accesses.BulkInsertTrustCenterFileAccesses(
				ctx,
				tx,
				s.svc.scope,
				access.ID,
				access.OrganizationID,
				newTrustCenterFileIDs,
				coredata.TrustCenterDocumentAccessStatusRequested,
				now,
			); err != nil {
				return fmt.Errorf("cannot bulk insert trust center file accesses: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	if err := s.svc.SlackMessages.QueueSlackNotification(ctx, req.IdentityID, req.TrustCenterID); err != nil {
		s.logger.ErrorCtx(ctx, "cannot queue slack notification", log.Error(err))
	}

	return access, nil
}

func (s TrustCenterAccessService) GetAccess(
	ctx context.Context,
	trustCenterID gid.GID,
	identityID gid.GID,
) (coredata.TrustCenterAccess, error) {
	var access coredata.TrustCenterAccess

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return access.LoadByTrustCenterIDAndIdentityID(ctx, conn, s.svc.scope, trustCenterID, identityID)
	})

	return access, err
}

func (s TrustCenterAccessService) GetDocumentAccess(
	ctx context.Context,
	trustCenterID gid.GID,
	identityID gid.GID,
	documentID gid.GID,
) (*coredata.TrustCenterDocumentAccess, error) {
	var documentAccess *coredata.TrustCenterDocumentAccess

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		access := &coredata.TrustCenterAccess{}
		err := access.LoadByTrustCenterIDAndIdentityID(ctx, conn, s.svc.scope, trustCenterID, identityID)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrMembershipNotFound
			}

			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		profile := &coredata.MembershipProfile{}
		if err := profile.LoadByIdentityIDAndOrganizationID(ctx, conn, s.svc.scope, identityID, access.OrganizationID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrUserNotFound
			}
		}

		if profile.State != coredata.ProfileStateActive {
			return ErrUserInactive
		}

		documentAccess = &coredata.TrustCenterDocumentAccess{}
		err = documentAccess.LoadByTrustCenterAccessIDAndDocumentID(ctx, conn, s.svc.scope, access.ID, documentID)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrDocumentAccessNotFound
			}

			return fmt.Errorf("cannot load document access: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return documentAccess, nil
}

func (s TrustCenterAccessService) GetReportAccess(
	ctx context.Context,
	trustCenterID gid.GID,
	identityID gid.GID,
	reportID gid.GID,
) (*coredata.TrustCenterDocumentAccess, error) {
	var reportAccess *coredata.TrustCenterDocumentAccess

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		access := &coredata.TrustCenterAccess{}
		err := access.LoadByTrustCenterIDAndIdentityID(ctx, conn, s.svc.scope, trustCenterID, identityID)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrMembershipNotFound
			}

			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		profile := &coredata.MembershipProfile{}
		if err := profile.LoadByIdentityIDAndOrganizationID(ctx, conn, s.svc.scope, identityID, access.OrganizationID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrUserNotFound
			}
		}

		if profile.State != coredata.ProfileStateActive {
			return ErrUserInactive
		}

		reportAccess = &coredata.TrustCenterDocumentAccess{}
		err = reportAccess.LoadByTrustCenterAccessIDAndReportID(ctx, conn, s.svc.scope, access.ID, reportID)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrDocumentAccessNotFound
			}

			return fmt.Errorf("cannot load report access: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return reportAccess, nil
}

func (s TrustCenterAccessService) GetTrustCenterFileAccess(
	ctx context.Context,
	trustCenterID gid.GID,
	identityID gid.GID,
	trustCenterFileID gid.GID,
) (*coredata.TrustCenterDocumentAccess, error) {
	var fileAccess *coredata.TrustCenterDocumentAccess

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		access := &coredata.TrustCenterAccess{}
		err := access.LoadByTrustCenterIDAndIdentityID(ctx, conn, s.svc.scope, trustCenterID, identityID)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrMembershipNotFound
			}

			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		profile := &coredata.MembershipProfile{}
		if err := profile.LoadByIdentityIDAndOrganizationID(ctx, conn, s.svc.scope, identityID, access.OrganizationID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrUserNotFound
			}
		}

		if profile.State != coredata.ProfileStateActive {
			return ErrUserInactive
		}

		fileAccess = &coredata.TrustCenterDocumentAccess{}
		err = fileAccess.LoadByTrustCenterAccessIDAndTrustCenterFileID(ctx, conn, s.svc.scope, access.ID, trustCenterFileID)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrDocumentAccessNotFound
			}

			return fmt.Errorf("cannot load trust center file access: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return fileAccess, nil
}

func (s *TrustCenterAccessService) GrantByIDs(
	ctx context.Context,
	organizationID gid.GID,
	email mail.Addr,
	documentIDs []gid.GID,
	reportIDs []gid.GID,
	fileIDs []gid.GID,
) error {
	return s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		trustCenter := &coredata.TrustCenter{}
		if err := trustCenter.LoadByOrganizationID(ctx, tx, s.svc.scope, organizationID); err != nil {
			return fmt.Errorf("cannot load trust center: %w", err)
		}

		identity := &coredata.Identity{}
		if err := identity.LoadByEmail(ctx, tx, email); err != nil {
			return fmt.Errorf("cannot load identity: %w", err)
		}

		access := &coredata.TrustCenterAccess{}
		if err := access.LoadByTrustCenterIDAndIdentityID(ctx, tx, s.svc.scope, trustCenter.ID, identity.ID); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		profile := &coredata.MembershipProfile{}
		if err := profile.LoadByIdentityIDAndOrganizationID(ctx, tx, s.svc.scope, identity.ID, access.OrganizationID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrUserNotFound
			}
		}

		if profile.State != coredata.ProfileStateActive {
			return ErrUserInactive
		}

		shouldSendEmail := profile.State != coredata.ProfileStateActive
		now := time.Now()

		if len(documentIDs) > 0 {
			if err := coredata.GrantByDocumentIDs(ctx, tx, s.svc.scope, access.ID, documentIDs, now); err != nil {
				return fmt.Errorf("cannot grant document accesses: %w", err)
			}
		}
		if len(reportIDs) > 0 {
			if err := coredata.GrantByReportIDs(ctx, tx, s.svc.scope, access.ID, reportIDs, now); err != nil {
				return fmt.Errorf("cannot grant report accesses: %w", err)
			}
		}
		if len(fileIDs) > 0 {
			if err := coredata.GrantByTrustCenterFileIDs(ctx, tx, s.svc.scope, access.ID, fileIDs, now); err != nil {
				return fmt.Errorf("cannot grant trust center file accesses: %w", err)
			}
		}

		if shouldSendEmail {
			profile.State = coredata.ProfileStateActive
			profile.UpdatedAt = now
			if err := profile.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update profile: %w", err)
			}

			if err := s.sendAccessEmail(ctx, tx, access, profile); err != nil {
				return fmt.Errorf("cannot send access email: %w", err)
			}
		}

		return nil
	})
}

func (s *TrustCenterAccessService) sendAccessEmail(ctx context.Context, tx pg.Conn, access *coredata.TrustCenterAccess, profile *coredata.MembershipProfile) error {
	organization := &coredata.Organization{}
	if err := organization.LoadByID(ctx, tx, s.svc.scope, access.OrganizationID); err != nil {
		return fmt.Errorf("cannot load organization: %w", err)
	}

	now := time.Now()
	access.UpdatedAt = now

	if err := access.Update(ctx, tx, s.svc.scope); err != nil {
		return fmt.Errorf("cannot update trust center access with expiration: %w", err)
	}

	emailPresenterCfg, err := s.svc.TrustCenters.EmailPresenterConfig(ctx, access.TrustCenterID)
	if err != nil {
		return fmt.Errorf("cannot get compliance page email presenter config: %w", err)
	}

	emailPresenter := emails.NewPresenterFromConfig(s.svc.fileManager, emailPresenterCfg, profile.FullName)

	subject, textBody, htmlBody, err := emailPresenter.RenderTrustCenterAccess(ctx, organization.Name)
	if err != nil {
		return fmt.Errorf("cannot render trust center access email: %w", err)
	}

	accessEmail := coredata.NewEmail(
		profile.FullName,
		profile.EmailAddress,
		subject,
		textBody,
		htmlBody,
	)

	if err := accessEmail.Insert(ctx, tx); err != nil {
		return fmt.Errorf("cannot insert access email: %w", err)
	}
	return nil
}

func (s *TrustCenterAccessService) RejectOrRevokeByIDs(
	ctx context.Context,
	organizationID gid.GID,
	email mail.Addr,
	documentIDs []gid.GID,
	reportIDs []gid.GID,
	fileIDs []gid.GID,
) error {
	return s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		trustCenter := &coredata.TrustCenter{}
		if err := trustCenter.LoadByOrganizationID(ctx, tx, s.svc.scope, organizationID); err != nil {
			return fmt.Errorf("cannot load trust center: %w", err)
		}

		identity := &coredata.Identity{}
		if err := identity.LoadByEmail(ctx, tx, email); err != nil {
			return fmt.Errorf("cannot load identity: %w", err)
		}

		access := &coredata.TrustCenterAccess{}
		if err := access.LoadByTrustCenterIDAndIdentityID(ctx, tx, s.svc.scope, trustCenter.ID, identity.ID); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		profile := &coredata.MembershipProfile{}
		if err := profile.LoadByIdentityIDAndOrganizationID(ctx, tx, s.svc.scope, identity.ID, access.OrganizationID); err != nil {
			return fmt.Errorf("cannot load profile: %w", err)
		}

		shouldSendEmail := false
		now := time.Now()

		if len(documentIDs) > 0 {
			shouldSendEmail = true
			if err := coredata.RejectOrRevokeByDocumentIDs(ctx, tx, s.svc.scope, access.ID, documentIDs, now); err != nil {
				return fmt.Errorf("cannot reject/revoke document accesses: %w", err)
			}
		}
		if len(reportIDs) > 0 {
			shouldSendEmail = true
			if err := coredata.RejectOrRevokeByReportIDs(ctx, tx, s.svc.scope, access.ID, reportIDs, now); err != nil {
				return fmt.Errorf("cannot reject/revoke report accesses: %w", err)
			}
		}
		if len(fileIDs) > 0 {
			shouldSendEmail = true
			if err := coredata.RejectOrRevokeByTrustCenterFileIDs(ctx, tx, s.svc.scope, access.ID, fileIDs, now); err != nil {
				return fmt.Errorf("cannot reject/revoke trust center file accesses: %w", err)
			}
		}

		if shouldSendEmail {
			if err := s.sendDocumentAccessRejectedEmail(ctx, tx, access, profile, documentIDs, reportIDs, fileIDs); err != nil {
				return fmt.Errorf("cannot send access email: %w", err)
			}
		}

		return nil
	})
}

func (s *TrustCenterAccessService) sendDocumentAccessRejectedEmail(
	ctx context.Context,
	tx pg.Conn,
	access *coredata.TrustCenterAccess,
	profile *coredata.MembershipProfile,
	documentIDs []gid.GID,
	reportIDs []gid.GID,
	fileIDs []gid.GID,
) error {
	organization := &coredata.Organization{}
	if err := organization.LoadByID(ctx, tx, s.svc.scope, access.OrganizationID); err != nil {
		return fmt.Errorf("cannot load organization: %w", err)
	}

	var fileNames []string
	var documents coredata.Documents
	if len(documentIDs) > 0 {
		if err := documents.LoadByIDs(ctx, tx, s.svc.scope, documentIDs); err != nil {
			return fmt.Errorf("cannot load documents by IDs: %w", err)
		}
		for _, d := range documents {
			fileNames = append(fileNames, d.Title)
		}
	}
	var reports coredata.Reports
	if len(reportIDs) > 0 {
		if err := reports.LoadByIDs(ctx, tx, s.svc.scope, reportIDs); err != nil {
			return fmt.Errorf("cannot load reports by IDs: %w", err)
		}
		for _, r := range reports {
			fileNames = append(fileNames, r.Filename)
		}
	}
	var files coredata.TrustCenterFiles
	if len(fileIDs) > 0 {
		if err := files.LoadByIDs(ctx, tx, s.svc.scope, fileIDs); err != nil {
			return fmt.Errorf("cannot load files by IDs: %w", err)
		}
		for _, f := range files {
			fileNames = append(fileNames, f.Name)
		}
	}

	emailPresenterCfg, err := s.svc.TrustCenters.EmailPresenterConfig(ctx, access.TrustCenterID)
	if err != nil {
		return fmt.Errorf("cannot get compliance page email presenter config: %w", err)
	}

	emailPresenter := emails.NewPresenterFromConfig(s.svc.fileManager, emailPresenterCfg, profile.FullName)

	subject, textBody, htmlBody, err := emailPresenter.RenderTrustCenterDocumentAccessRejected(
		ctx,
		fileNames,
		organization.Name,
	)
	if err != nil {
		return fmt.Errorf("cannot render trust center documents access rejected email: %w", err)
	}

	accessEmail := coredata.NewEmail(
		profile.FullName,
		profile.EmailAddress,
		subject,
		textBody,
		htmlBody,
	)

	if err := accessEmail.Insert(ctx, tx); err != nil {
		return fmt.Errorf("cannot insert access email: %w", err)
	}
	return nil
}

func extractExistingIDs(accesses coredata.TrustCenterDocumentAccesses) ([]gid.GID, []gid.GID, []gid.GID) {
	var documentIDs []gid.GID
	var reportIDs []gid.GID
	var trustCenterFileIDs []gid.GID

	for _, access := range accesses {
		if access.DocumentID != nil {
			documentIDs = append(documentIDs, *access.DocumentID)
		}
		if access.ReportID != nil {
			reportIDs = append(reportIDs, *access.ReportID)
		}
		if access.TrustCenterFileID != nil {
			trustCenterFileIDs = append(trustCenterFileIDs, *access.TrustCenterFileID)
		}
	}

	return documentIDs, reportIDs, trustCenterFileIDs
}

func filterExistingIDs(allIDs []gid.GID, existingIDs []gid.GID) []gid.GID {
	existingMap := make(map[gid.GID]bool)
	for _, id := range existingIDs {
		existingMap[id] = true
	}

	var newIDs []gid.GID
	for _, id := range allIDs {
		if !existingMap[id] {
			newIDs = append(newIDs, id)
		}
	}

	return newIDs
}
