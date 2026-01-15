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
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/validator"
)

type (
	TrustCenterAccessService struct {
		svc    *TenantService
		iamSvc *iam.Service
		logger *log.Logger
	}

	TrustCenterAccessRequest struct {
		TrustCenterID      gid.GID
		Email              mail.Addr
		FullName           string
		DocumentIDs        []gid.GID
		ReportIDs          []gid.GID
		TrustCenterFileIDs []gid.GID
	}
)

const (
	TrustCenterAccessURLFormat = "https://%s/organizations/%s/trust-center/access"
)

func (tcar *TrustCenterAccessRequest) Validate() error {
	v := validator.New()

	v.Check(tcar.Email.Domain(), "email", validator.NotBlacklisted())

	return v.Error()
}

func (s TrustCenterAccessService) Request(
	ctx context.Context,
	req *TrustCenterAccessRequest,
) (*coredata.TrustCenterAccess, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request arguments: %w", err)
	}

	var (
		now         = time.Now()
		access      *coredata.TrustCenterAccess
		trustCenter *coredata.TrustCenter
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, tx, s.svc.scope, req.TrustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			// TODO: load document to ensure they are requestable

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

			existingAccess := &coredata.TrustCenterAccess{}
			err := existingAccess.LoadByTrustCenterIDAndEmail(ctx, tx, s.svc.scope, req.TrustCenterID, req.Email)
			if err == nil {
				access = existingAccess
			} else {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load trust center access: %w", err)
				}

				access = &coredata.TrustCenterAccess{
					ID:                                gid.New(s.svc.scope.GetTenantID(), coredata.TrustCenterAccessEntityType),
					OrganizationID:                    organizationID,
					TenantID:                          s.svc.scope.GetTenantID(),
					TrustCenterID:                     req.TrustCenterID,
					Email:                             req.Email,
					Name:                              req.FullName,
					Active:                            false,
					HasAcceptedNonDisclosureAgreement: false,
					CreatedAt:                         now,
					UpdatedAt:                         now,
				}

				if err := access.Insert(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot insert trust center access: %w", err)
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

	if err := s.svc.SlackMessages.QueueSlackNotification(ctx, access.Email, req.TrustCenterID); err != nil {
		s.logger.ErrorCtx(ctx, "cannot queue slack notification", log.Error(err))
	}

	return access, nil
}

func (s TrustCenterAccessService) HasAcceptedNonDisclosureAgreement(ctx context.Context, trustCenterID gid.GID, email mail.Addr) (bool, error) {
	access := &coredata.TrustCenterAccess{}
	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		err := access.LoadByTrustCenterIDAndEmail(ctx, conn, s.svc.scope, trustCenterID, email)
		if err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		return nil
	})

	if err != nil {
		return false, nil
	}

	return access.HasAcceptedNonDisclosureAgreement, nil
}

type AcceptNDARequest struct {
	TrustCenterID gid.GID
	Email         mail.Addr
	IPAddr        string
}

func (s TrustCenterAccessService) AcceptNonDisclosureAgreement(ctx context.Context, req *AcceptNDARequest) error {
	return s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		access := &coredata.TrustCenterAccess{}
		if err := access.LoadByTrustCenterIDAndEmail(ctx, tx, s.svc.scope, req.TrustCenterID, req.Email); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		trustCenter := &coredata.TrustCenter{}
		if err := trustCenter.LoadByID(ctx, tx, s.svc.scope, req.TrustCenterID); err != nil {
			return fmt.Errorf("cannot load trust center: %w", err)
		}

		acceptationLogs, err := json.Marshal(map[string]string{
			"email":     req.Email.String(),
			"timestamp": time.Now().Format(time.RFC3339),
			"ip":        req.IPAddr,
		})
		if err != nil {
			return fmt.Errorf("cannot marshal non disclosure agreement acceptation logs: %w", err)
		}

		access.HasAcceptedNonDisclosureAgreement = true
		access.UpdatedAt = time.Now()
		access.HasAcceptedNonDisclosureAgreementMetadata = acceptationLogs
		access.NDAFileID = trustCenter.NonDisclosureAgreementFileID
		if err := access.Update(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot update trust center access: %w", err)
		}

		return nil
	})
}

func (s TrustCenterAccessService) LoadDocumentAccess(
	ctx context.Context,
	trustCenterID gid.GID,
	email mail.Addr,
	documentID gid.GID,
) (*coredata.TrustCenterDocumentAccess, error) {
	var documentAccess *coredata.TrustCenterDocumentAccess

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		access := &coredata.TrustCenterAccess{}
		err := access.LoadByTrustCenterIDAndEmail(ctx, conn, s.svc.scope, trustCenterID, email)
		if err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		if !access.Active {
			return fmt.Errorf("trust center access is not active")
		}

		documentAccess = &coredata.TrustCenterDocumentAccess{}
		err = documentAccess.LoadByTrustCenterAccessIDAndDocumentID(ctx, conn, s.svc.scope, access.ID, documentID)
		if err != nil {
			return fmt.Errorf("cannot load document access: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return documentAccess, nil
}

func (s TrustCenterAccessService) LoadReportAccess(
	ctx context.Context,
	trustCenterID gid.GID,
	email mail.Addr,
	reportID gid.GID,
) (*coredata.TrustCenterDocumentAccess, error) {
	var reportAccess *coredata.TrustCenterDocumentAccess

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		access := &coredata.TrustCenterAccess{}
		err := access.LoadByTrustCenterIDAndEmail(ctx, conn, s.svc.scope, trustCenterID, email)
		if err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		if !access.Active {
			return fmt.Errorf("trust center access is not active")
		}

		reportAccess = &coredata.TrustCenterDocumentAccess{}
		err = reportAccess.LoadByTrustCenterAccessIDAndReportID(ctx, conn, s.svc.scope, access.ID, reportID)
		if err != nil {
			return fmt.Errorf("cannot load report access: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return reportAccess, nil
}

func (s TrustCenterAccessService) LoadTrustCenterFileAccess(
	ctx context.Context,
	trustCenterID gid.GID,
	email mail.Addr,
	trustCenterFileID gid.GID,
) (*coredata.TrustCenterDocumentAccess, error) {
	var fileAccess *coredata.TrustCenterDocumentAccess

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		access := &coredata.TrustCenterAccess{}
		err := access.LoadByTrustCenterIDAndEmail(ctx, conn, s.svc.scope, trustCenterID, email)
		if err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		if !access.Active {
			return fmt.Errorf("trust center access is not active")
		}

		fileAccess = &coredata.TrustCenterDocumentAccess{}
		err = fileAccess.LoadByTrustCenterAccessIDAndTrustCenterFileID(ctx, conn, s.svc.scope, access.ID, trustCenterFileID)
		if err != nil {
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

		access := &coredata.TrustCenterAccess{}
		if err := access.LoadByTrustCenterIDAndEmail(ctx, tx, s.svc.scope, trustCenter.ID, email); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		shouldSendEmail := !access.Active
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
			access.Active = true
			access.UpdatedAt = now
			if err := access.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update trust center access: %w", err)
			}

			if err := s.sendAccessEmail(ctx, tx, access); err != nil {
				return fmt.Errorf("cannot send access email: %w", err)
			}
		}

		return nil
	})
}

func (s *TrustCenterAccessService) sendAccessEmail(ctx context.Context, tx pg.Conn, access *coredata.TrustCenterAccess) error {
	trustCenter := &coredata.TrustCenter{}

	if err := trustCenter.LoadByID(ctx, tx, s.svc.scope, access.TrustCenterID); err != nil {
		return fmt.Errorf("cannot load trust center: %w", err)
	}

	organization := &coredata.Organization{}
	if err := organization.LoadByID(ctx, tx, s.svc.scope, trustCenter.OrganizationID); err != nil {
		return fmt.Errorf("cannot load organization: %w", err)
	}

	baseURLParsed, err := url.Parse(s.svc.baseURL)
	if err != nil {
		return fmt.Errorf("cannot parse base URL: %w", err)
	}

	hostname := baseURLParsed.Host
	scheme := baseURLParsed.Scheme
	path := "/trust/" + trustCenter.Slug

	if organization.CustomDomainID != nil {
		customDomain, err := s.svc.Organizations.GetOrganizationCustomDomain(ctx, organization.ID)
		if err != nil {
			return fmt.Errorf("cannot load custom domain: %w", err)
		}

		if customDomain == nil || customDomain.SSLStatus != coredata.CustomDomainSSLStatusActive {
			return fmt.Errorf("custom domain is not active")
		}

		hostname = customDomain.Domain
		scheme = "https"
		path = ""
	}

	accessURL := url.URL{
		Scheme: scheme,
		Host:   hostname,
		Path:   path,
	}

	now := time.Now()
	expiresAt := now.Add(s.svc.trustConfig.TokenDuration)
	access.LastTokenExpiresAt = &expiresAt
	access.UpdatedAt = now

	if err := access.Update(ctx, tx, s.svc.scope); err != nil {
		return fmt.Errorf("cannot update trust center access with expiration: %w", err)
	}

	return s.sendTrustCenterAccessEmail(ctx, tx, access.Name, access.Email, organization.Name, accessURL.String())
}

func (s *TrustCenterAccessService) sendTrustCenterAccessEmail(
	ctx context.Context,
	tx pg.Conn,
	name string,
	email mail.Addr,
	companyName string,
	accessURL string,
) error {
	subject, textBody, htmlBody, err := emails.RenderTrustCenterAccess(
		s.svc.baseURL,
		name,
		companyName,
		accessURL,
		s.svc.trustConfig.TokenDuration,
	)
	if err != nil {
		return fmt.Errorf("cannot render trust center access email: %w", err)
	}

	accessEmail := coredata.NewEmail(
		name,
		email,
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

		access := &coredata.TrustCenterAccess{}
		if err := access.LoadByTrustCenterIDAndEmail(ctx, tx, s.svc.scope, trustCenter.ID, email); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
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
			if err := s.sendDocumentAccessRejectedEmail(ctx, tx, access, documentIDs, reportIDs, fileIDs); err != nil {
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
	documentIDs []gid.GID,
	reportIDs []gid.GID,
	fileIDs []gid.GID,
) error {
	trustCenter := &coredata.TrustCenter{}
	if err := trustCenter.LoadByID(ctx, tx, s.svc.scope, access.TrustCenterID); err != nil {
		return fmt.Errorf("cannot load trust center: %w", err)
	}

	organization := &coredata.Organization{}
	if err := organization.LoadByID(ctx, tx, s.svc.scope, trustCenter.OrganizationID); err != nil {
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

	subject, textBody, htmlBody, err := emails.RenderTrustCenterDocumentAccessRejected(
		s.svc.baseURL,
		access.Name,
		organization.Name,
		fileNames,
	)
	if err != nil {
		return fmt.Errorf("cannot render trust center documents access rejected email: %w", err)
	}

	accessEmail := coredata.NewEmail(
		access.Name,
		access.Email,
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
