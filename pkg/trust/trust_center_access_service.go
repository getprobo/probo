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
	"net/mail"
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/usrmgr"
	"go.gearno.de/kit/pg"
)

type (
	TrustCenterAccessService struct {
		svc    *TenantService
		usrmgr *usrmgr.Service
	}

	RequestTrustCenterAccessRequest struct {
		TrustCenterID gid.GID
		Email         string
		Name          *string
		DocumentIDs   []gid.GID
		ReportIDs     []gid.GID
	}
)

const (
	TokenTypeTrustCenterAccess = "trust_center_access"
)

func (s TrustCenterAccessService) ValidateToken(
	ctx context.Context,
	trustCenterID gid.GID,
	email string,
) error {
	return s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		access := &coredata.TrustCenterAccess{}
		err := access.LoadByTrustCenterIDAndEmail(ctx, conn, s.svc.scope, trustCenterID, email)
		if err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		if !access.Active {
			return fmt.Errorf("trust center access is not active")
		}

		return nil
	})
}

func (s TrustCenterAccessService) Request(
	ctx context.Context,
	req *RequestTrustCenterAccessRequest,
) (*coredata.TrustCenterAccess, error) {
	now := time.Now()

	var access *coredata.TrustCenterAccess

	err := s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		var trustCenter *coredata.TrustCenter
		var organizationID gid.GID
		trustCenter = &coredata.TrustCenter{}
		if err := trustCenter.LoadByID(ctx, tx, s.svc.scope, req.TrustCenterID); err != nil {
			return fmt.Errorf("cannot load trust center: %w", err)
		}
		organizationID = trustCenter.OrganizationID

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
		existingAccess := &coredata.TrustCenterAccess{}
		err := existingAccess.LoadByTrustCenterIDAndEmail(ctx, tx, s.svc.scope, req.TrustCenterID, req.Email)

		if err == nil {
			access = existingAccess
		} else {
			var notFoundErr *coredata.ErrTrustCenterAccessNotFound
			if !errors.As(err, &notFoundErr) {
				return fmt.Errorf("cannot load trust center access: %w", err)
			}

			if req.Name == nil || *req.Name == "" {
				return fmt.Errorf("name is required for new access requests")
			}

			if _, err := mail.ParseAddress(req.Email); err != nil {
				return fmt.Errorf("invalid email address")
			}

			access = &coredata.TrustCenterAccess{
				ID:                                gid.New(s.svc.scope.GetTenantID(), coredata.TrustCenterAccessEntityType),
				TenantID:                          s.svc.scope.GetTenantID(),
				TrustCenterID:                     req.TrustCenterID,
				Email:                             req.Email,
				Name:                              *req.Name,
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

		existingDocumentIDs, existingReportIDs := extractExistingIDs(existingAccesses)
		newDocumentIDs := filterExistingIDs(documentIDs, existingDocumentIDs)
		newReportIDs := filterExistingIDs(reportIDs, existingReportIDs)

		var accesses coredata.TrustCenterDocumentAccesses

		if err := accesses.BulkInsertDocumentAccesses(ctx, tx, s.svc.scope, access.ID, newDocumentIDs, now); err != nil {
			return fmt.Errorf("cannot bulk insert trust center document accesses: %w", err)
		}

		if err := accesses.BulkInsertReportAccesses(ctx, tx, s.svc.scope, access.ID, newReportIDs, now); err != nil {
			return fmt.Errorf("cannot bulk insert trust center report accesses: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return access, nil
}

func (s TrustCenterAccessService) HasAcceptedNonDisclosureAgreement(ctx context.Context, trustCenterID gid.GID, email string) (bool, error) {
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

func (s TrustCenterAccessService) AcceptNonDisclosureAgreement(ctx context.Context, trustCenterID gid.GID, email string) error {
	return s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		access := &coredata.TrustCenterAccess{}
		if err := access.LoadByTrustCenterIDAndEmail(ctx, tx, s.svc.scope, trustCenterID, email); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		acceptationLogs, err := json.Marshal(map[string]string{
			"email":     email,
			"timestamp": time.Now().Format(time.RFC3339),
			"ip":        ctx.Value(coredata.ContextKeyIPAddress).(string),
		})
		if err != nil {
			return fmt.Errorf("cannot marshal non disclosure agreement acceptation logs: %w", err)
		}

		access.HasAcceptedNonDisclosureAgreement = true
		access.UpdatedAt = time.Now()
		access.HasAcceptedNonDisclosureAgreementMetadata = acceptationLogs
		if err := access.Update(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot update trust center access: %w", err)
		}

		return nil
	})
}

func (s TrustCenterAccessService) LoadDocumentAccess(
	ctx context.Context,
	trustCenterID gid.GID,
	email string,
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
	email string,
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

func extractExistingIDs(accesses coredata.TrustCenterDocumentAccesses) ([]gid.GID, []gid.GID) {
	var documentIDs []gid.GID
	var reportIDs []gid.GID

	for _, access := range accesses {
		if access.DocumentID != nil {
			documentIDs = append(documentIDs, *access.DocumentID)
		}
		if access.ReportID != nil {
			reportIDs = append(reportIDs, *access.ReportID)
		}
	}

	return documentIDs, reportIDs
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
