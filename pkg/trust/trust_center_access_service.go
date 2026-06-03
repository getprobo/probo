// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mail"
)

type (
	TrustCenterAccessService struct {
		svc    *Service
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

func (s TrustCenterAccessService) Request(
	ctx context.Context,
	scope coredata.Scoper,
	req *TrustCenterAccessRequest,
) (*coredata.TrustCenterAccess, error) {
	var (
		now    = time.Now()
		access *coredata.TrustCenterAccess
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, tx, scope, req.TrustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			access = &coredata.TrustCenterAccess{}
			if err := access.LoadByTrustCenterIDAndIdentityID(ctx, tx, scope, req.TrustCenterID, req.IdentityID); err != nil {
				return fmt.Errorf("cannot load compliance page membership: %w", err)
			}

			organizationID := trustCenter.OrganizationID

			documentIDs := req.DocumentIDs
			if req.DocumentIDs == nil {
				var allDocuments coredata.Documents

				filter := coredata.NewDocumentTrustCenterFilter()

				if err := allDocuments.LoadAllByOrganizationID(ctx, tx, scope, organizationID, filter); err != nil {
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

				if err := allAudits.LoadAllByOrganizationID(ctx, tx, scope, organizationID, auditFilter); err != nil {
					return fmt.Errorf("cannot list audits: %w", err)
				}

				for _, audit := range allAudits {
					if audit.ReportFileID != nil {
						reportIDs = append(reportIDs, *audit.ReportFileID)
					}
				}
			}

			trustCenterFileIDs := req.TrustCenterFileIDs
			if req.TrustCenterFileIDs == nil {
				var allTrustCenterFiles coredata.TrustCenterFiles

				filter := coredata.NewTrustCenterFileFilter(
					coredata.WithTrustCenterFileVisibilities(coredata.TrustCenterVisibilityPrivate, coredata.TrustCenterVisibilityNone),
				)

				if err := allTrustCenterFiles.LoadAllByOrganizationID(ctx, tx, scope, organizationID, filter); err != nil {
					return fmt.Errorf("cannot list trust center files: %w", err)
				}

				for _, file := range allTrustCenterFiles {
					trustCenterFileIDs = append(trustCenterFileIDs, file.ID)
				}
			}

			var existingAccesses coredata.TrustCenterDocumentAccesses
			if err := existingAccesses.LoadAllByTrustCenterAccessID(ctx, tx, scope, access.ID); err != nil {
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
				scope,
				access.ID,
				access.OrganizationID,
				newDocumentIDs,
				coredata.TrustCenterDocumentAccessStatusRequested,
				now,
			); err != nil {
				return fmt.Errorf("cannot bulk insert trust center document accesses: %w", err)
			}

			if err := accesses.BulkInsertReportFileAccesses(
				ctx,
				tx,
				scope,
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
				scope,
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

	if err := s.svc.slack.QueueSlackNotification(ctx, scope, req.IdentityID, req.TrustCenterID); err != nil {
		s.logger.ErrorCtx(ctx, "cannot queue slack notification", log.Error(err))
	}

	return access, nil
}

func (s TrustCenterAccessService) GetAccess(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
	identityID gid.GID,
) (coredata.TrustCenterAccess, error) {
	var access coredata.TrustCenterAccess

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return access.LoadByTrustCenterIDAndIdentityID(ctx, conn, scope, trustCenterID, identityID)
	})

	return access, err
}

func (s TrustCenterAccessService) GetDocumentAccess(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
	identityID gid.GID,
	documentID gid.GID,
) (*coredata.TrustCenterDocumentAccess, error) {
	var documentAccess *coredata.TrustCenterDocumentAccess

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		access := &coredata.TrustCenterAccess{}

		err := access.LoadByTrustCenterIDAndIdentityID(ctx, conn, scope, trustCenterID, identityID)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrMembershipNotFound
			}

			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		profile := &coredata.MembershipProfile{}
		if err := profile.LoadByIdentityIDAndOrganizationID(ctx, conn, scope, identityID, access.OrganizationID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrUserNotFound
			}
		}

		if profile.State != coredata.ProfileStateActive {
			return ErrUserInactive
		}

		documentAccess = &coredata.TrustCenterDocumentAccess{}

		err = documentAccess.LoadByTrustCenterAccessIDAndDocumentID(ctx, conn, scope, access.ID, documentID)
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

func (s TrustCenterAccessService) GetReportFileAccess(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
	identityID gid.GID,
	reportFileID gid.GID,
) (*coredata.TrustCenterDocumentAccess, error) {
	var reportAccess *coredata.TrustCenterDocumentAccess

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		access := &coredata.TrustCenterAccess{}

		err := access.LoadByTrustCenterIDAndIdentityID(ctx, conn, scope, trustCenterID, identityID)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrMembershipNotFound
			}

			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		profile := &coredata.MembershipProfile{}
		if err := profile.LoadByIdentityIDAndOrganizationID(ctx, conn, scope, identityID, access.OrganizationID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrUserNotFound
			}
		}

		if profile.State != coredata.ProfileStateActive {
			return ErrUserInactive
		}

		reportAccess = &coredata.TrustCenterDocumentAccess{}

		err = reportAccess.LoadByTrustCenterAccessIDAndReportFileID(ctx, conn, scope, access.ID, reportFileID)
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
	scope coredata.Scoper,
	trustCenterID gid.GID,
	identityID gid.GID,
	trustCenterFileID gid.GID,
) (*coredata.TrustCenterDocumentAccess, error) {
	var fileAccess *coredata.TrustCenterDocumentAccess

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		access := &coredata.TrustCenterAccess{}

		err := access.LoadByTrustCenterIDAndIdentityID(ctx, conn, scope, trustCenterID, identityID)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrMembershipNotFound
			}

			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		profile := &coredata.MembershipProfile{}
		if err := profile.LoadByIdentityIDAndOrganizationID(ctx, conn, scope, identityID, access.OrganizationID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrUserNotFound
			}
		}

		if profile.State != coredata.ProfileStateActive {
			return ErrUserInactive
		}

		fileAccess = &coredata.TrustCenterDocumentAccess{}

		err = fileAccess.LoadByTrustCenterAccessIDAndTrustCenterFileID(ctx, conn, scope, access.ID, trustCenterFileID)
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
	scope coredata.Scoper,
	organizationID gid.GID,
	email mail.Addr,
	documentIDs []gid.GID,
	reportIDs []gid.GID,
	fileIDs []gid.GID,
) error {
	return s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		trustCenter := &coredata.TrustCenter{}
		if err := trustCenter.LoadByOrganizationID(ctx, tx, scope, organizationID); err != nil {
			return fmt.Errorf("cannot load trust center: %w", err)
		}

		identity := &coredata.Identity{}
		if err := identity.LoadByEmail(ctx, tx, email); err != nil {
			return fmt.Errorf("cannot load identity: %w", err)
		}

		access := &coredata.TrustCenterAccess{}
		if err := access.LoadByTrustCenterIDAndIdentityID(ctx, tx, scope, trustCenter.ID, identity.ID); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		profile := &coredata.MembershipProfile{}
		if err := profile.LoadByIdentityIDAndOrganizationID(ctx, tx, scope, identity.ID, access.OrganizationID); err != nil {
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
			if err := coredata.GrantByDocumentIDs(ctx, tx, scope, access.ID, documentIDs, now); err != nil {
				return fmt.Errorf("cannot grant document accesses: %w", err)
			}
		}

		if len(reportIDs) > 0 {
			if err := coredata.GrantByReportFileIDs(ctx, tx, scope, access.ID, reportIDs, now); err != nil {
				return fmt.Errorf("cannot grant report accesses: %w", err)
			}
		}

		if len(fileIDs) > 0 {
			if err := coredata.GrantByTrustCenterFileIDs(ctx, tx, scope, access.ID, fileIDs, now); err != nil {
				return fmt.Errorf("cannot grant trust center file accesses: %w", err)
			}
		}

		if shouldSendEmail {
			profile.State = coredata.ProfileStateActive

			profile.UpdatedAt = now
			if err := profile.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update profile: %w", err)
			}

			if err := s.sendAccessEmail(ctx, tx, scope, access, profile); err != nil {
				return fmt.Errorf("cannot send access email: %w", err)
			}
		}

		return nil
	})
}

func (s *TrustCenterAccessService) sendAccessEmail(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	access *coredata.TrustCenterAccess,
	profile *coredata.MembershipProfile,
) error {
	organization := &coredata.Organization{}
	if err := organization.LoadByID(ctx, tx, scope, access.OrganizationID); err != nil {
		return fmt.Errorf("cannot load organization: %w", err)
	}

	now := time.Now()
	access.UpdatedAt = now

	if err := access.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update trust center access with expiration: %w", err)
	}

	emailPresenterCfg, err := s.svc.TrustCenters.EmailPresenterConfig(ctx, scope, access.TrustCenterID)
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
		&coredata.EmailOptions{
			SenderName: new(organization.Name),
		},
	)

	if err := accessEmail.Insert(ctx, tx); err != nil {
		return fmt.Errorf("cannot insert access email: %w", err)
	}

	return nil
}

func (s *TrustCenterAccessService) RejectOrRevokeByIDs(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	email mail.Addr,
	documentIDs []gid.GID,
	reportIDs []gid.GID,
	fileIDs []gid.GID,
) error {
	return s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		trustCenter := &coredata.TrustCenter{}
		if err := trustCenter.LoadByOrganizationID(ctx, tx, scope, organizationID); err != nil {
			return fmt.Errorf("cannot load trust center: %w", err)
		}

		identity := &coredata.Identity{}
		if err := identity.LoadByEmail(ctx, tx, email); err != nil {
			return fmt.Errorf("cannot load identity: %w", err)
		}

		access := &coredata.TrustCenterAccess{}
		if err := access.LoadByTrustCenterIDAndIdentityID(ctx, tx, scope, trustCenter.ID, identity.ID); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		profile := &coredata.MembershipProfile{}
		if err := profile.LoadByIdentityIDAndOrganizationID(ctx, tx, scope, identity.ID, access.OrganizationID); err != nil {
			return fmt.Errorf("cannot load profile: %w", err)
		}

		shouldSendEmail := false
		now := time.Now()

		if len(documentIDs) > 0 {
			shouldSendEmail = true

			if err := coredata.RejectOrRevokeByDocumentIDs(ctx, tx, scope, access.ID, documentIDs, now); err != nil {
				return fmt.Errorf("cannot reject/revoke document accesses: %w", err)
			}
		}

		if len(reportIDs) > 0 {
			shouldSendEmail = true

			if err := coredata.RejectOrRevokeByReportFileIDs(ctx, tx, scope, access.ID, reportIDs, now); err != nil {
				return fmt.Errorf("cannot reject/revoke report accesses: %w", err)
			}
		}

		if len(fileIDs) > 0 {
			shouldSendEmail = true

			if err := coredata.RejectOrRevokeByTrustCenterFileIDs(ctx, tx, scope, access.ID, fileIDs, now); err != nil {
				return fmt.Errorf("cannot reject/revoke trust center file accesses: %w", err)
			}
		}

		if shouldSendEmail {
			if err := s.sendDocumentAccessRejectedEmail(ctx, tx, scope, access, profile, documentIDs, reportIDs, fileIDs); err != nil {
				return fmt.Errorf("cannot send access email: %w", err)
			}
		}

		return nil
	})
}

func (s *TrustCenterAccessService) sendDocumentAccessRejectedEmail(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	access *coredata.TrustCenterAccess,
	profile *coredata.MembershipProfile,
	documentIDs []gid.GID,
	reportIDs []gid.GID,
	fileIDs []gid.GID,
) error {
	organization := &coredata.Organization{}
	if err := organization.LoadByID(ctx, tx, scope, access.OrganizationID); err != nil {
		return fmt.Errorf("cannot load organization: %w", err)
	}

	var (
		fileNames []string
		documents coredata.Documents
	)

	if len(documentIDs) > 0 {
		if err := documents.LoadByIDs(ctx, tx, scope, documentIDs); err != nil {
			return fmt.Errorf("cannot load documents by IDs: %w", err)
		}

		for _, d := range documents {
			fileNames = append(fileNames, d.Title)
		}
	}

	if len(reportIDs) > 0 {
		reportLabels, err := reportAccessLabels(ctx, tx, scope, reportIDs)
		if err != nil {
			return fmt.Errorf("cannot build report access labels: %w", err)
		}

		fileNames = append(fileNames, reportLabels...)
	}

	var files coredata.TrustCenterFiles
	if len(fileIDs) > 0 {
		if err := files.LoadByIDs(ctx, tx, scope, fileIDs); err != nil {
			return fmt.Errorf("cannot load files by IDs: %w", err)
		}

		for _, f := range files {
			fileNames = append(fileNames, f.Name)
		}
	}

	emailPresenterCfg, err := s.svc.TrustCenters.EmailPresenterConfig(ctx, scope, access.TrustCenterID)
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
		&coredata.EmailOptions{
			SenderName: new(organization.Name),
		},
	)

	if err := accessEmail.Insert(ctx, tx); err != nil {
		return fmt.Errorf("cannot insert access email: %w", err)
	}

	return nil
}

func extractExistingIDs(accesses coredata.TrustCenterDocumentAccesses) ([]gid.GID, []gid.GID, []gid.GID) {
	var (
		documentIDs        []gid.GID
		reportIDs          []gid.GID
		trustCenterFileIDs []gid.GID
	)

	for _, access := range accesses {
		if access.DocumentID != nil {
			documentIDs = append(documentIDs, *access.DocumentID)
		}

		if access.ReportFileID != nil {
			reportIDs = append(reportIDs, *access.ReportFileID)
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

func reportAccessLabels(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	reportFileIDs []gid.GID,
) ([]string, error) {
	var reportFiles coredata.Files
	if err := reportFiles.LoadByIDs(ctx, conn, scope, reportFileIDs); err != nil {
		return nil, fmt.Errorf("cannot load report files by IDs: %w", err)
	}

	fileByID := make(map[gid.GID]*coredata.File, len(reportFiles))
	for _, f := range reportFiles {
		fileByID[f.ID] = f
	}

	var audits coredata.Audits
	if err := audits.LoadByReportFileIDs(ctx, conn, scope, reportFileIDs); err != nil {
		return nil, fmt.Errorf("cannot load audits by report file IDs: %w", err)
	}

	auditByFileID := make(map[gid.GID]*coredata.Audit, len(audits))
	frameworkIDSet := make(map[gid.GID]struct{})

	for _, audit := range audits {
		if audit.ReportFileID == nil {
			continue
		}

		if _, exists := auditByFileID[*audit.ReportFileID]; exists {
			continue
		}

		auditByFileID[*audit.ReportFileID] = audit
		frameworkIDSet[audit.FrameworkID] = struct{}{}
	}

	frameworkIDs := make([]gid.GID, 0, len(frameworkIDSet))
	for frameworkID := range frameworkIDSet {
		frameworkIDs = append(frameworkIDs, frameworkID)
	}

	frameworkByID := make(map[gid.GID]*coredata.Framework, len(frameworkIDs))

	if len(frameworkIDs) > 0 {
		var frameworks coredata.Frameworks
		if err := frameworks.LoadByIDs(ctx, conn, scope, frameworkIDs); err != nil {
			return nil, fmt.Errorf("cannot load frameworks by IDs: %w", err)
		}

		for _, framework := range frameworks {
			frameworkByID[framework.ID] = framework
		}
	}

	labels := make([]string, 0, len(reportFileIDs))

	for _, fileID := range reportFileIDs {
		file, ok := fileByID[fileID]
		if !ok {
			return nil, fmt.Errorf("cannot load report file %q: %w", fileID, coredata.ErrResourceNotFound)
		}

		audit, ok := auditByFileID[fileID]
		if !ok {
			labels = append(labels, file.FileName)
			continue
		}

		framework, ok := frameworkByID[audit.FrameworkID]
		if !ok {
			labels = append(labels, file.FileName)
			continue
		}

		if audit.Name != nil && *audit.Name != "" {
			labels = append(labels, fmt.Sprintf("%s - %s", framework.Name, *audit.Name))
			continue
		}

		labels = append(labels, framework.Name)
	}

	return labels, nil
}
