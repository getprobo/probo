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

package probo

import (
	"bytes"
	"context"
	"fmt"
	"net/mail"
	"net/url"
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/mailer"
	"github.com/getprobo/probo/pkg/page"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"go.gearno.de/kit/pg"
)

type (
	TrustCenterAccessService struct {
		svc *TenantService
	}

	CreateTrustCenterAccessRequest struct {
		TrustCenterID gid.GID
		Email         string
		Name          string
	}

	UpdateTrustCenterAccessRequest struct {
		ID          gid.GID
		Name        *string
		Active      *bool
		DocumentIDs []gid.GID
		ReportIDs   []gid.GID
	}

	DeleteTrustCenterAccessRequest struct {
		ID gid.GID
	}

	TrustCenterAccessData struct {
		TrustCenterID gid.GID `json:"trust_center_id"`
		Email         string  `json:"email"`
	}
)

const (
	trustCenterAccessEmailSubjectFormat = "Trust Center Access Invitation - %s"
	trustCenterAccessEmailHeader        = "Trust Center Access"
	trustCenterAccessEmailBodyFormat    = "You have been granted access to %s's Trust Center! Click the button below to access it:"
	trustCenterAccessEmailButtonText    = "Access Trust Center"
	trustCenterAccessEmailFooter        = "This link will expire in 7 days."
)

func (s TrustCenterAccessService) ListForTrustCenterID(
	ctx context.Context,
	trustCenterID gid.GID,
	cursor *page.Cursor[coredata.TrustCenterAccessOrderField],
) (*page.Page[*coredata.TrustCenterAccess, coredata.TrustCenterAccessOrderField], error) {
	var accesses coredata.TrustCenterAccesses

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return accesses.LoadByTrustCenterID(ctx, conn, s.svc.scope, trustCenterID, cursor)
	})

	if err != nil {
		return nil, err
	}

	return page.NewPage(accesses, cursor), nil
}

func (s TrustCenterAccessService) ListDocumentAccesses(
	ctx context.Context,
	trustCenterAccessID gid.GID,
	cursor *page.Cursor[coredata.TrustCenterDocumentAccessOrderField],
) (*page.Page[*coredata.TrustCenterDocumentAccess, coredata.TrustCenterDocumentAccessOrderField], error) {
	var documentAccesses coredata.TrustCenterDocumentAccesses

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return documentAccesses.LoadByTrustCenterAccessID(ctx, conn, s.svc.scope, trustCenterAccessID, cursor)
	})

	if err != nil {
		return nil, err
	}

	return page.NewPage(documentAccesses, cursor), nil
}

func (s TrustCenterAccessService) Get(
	ctx context.Context,
	accessID gid.GID,
) (*coredata.TrustCenterAccess, error) {
	var access coredata.TrustCenterAccess

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return access.LoadByID(ctx, conn, s.svc.scope, accessID)
	})

	if err != nil {
		return nil, err
	}

	return &access, nil
}

func (s TrustCenterAccessService) GetDocumentAccess(
	ctx context.Context,
	documentAccessID gid.GID,
) (*coredata.TrustCenterDocumentAccess, error) {
	var documentAccess coredata.TrustCenterDocumentAccess

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return documentAccess.LoadByID(ctx, conn, s.svc.scope, documentAccessID)
	})

	if err != nil {
		return nil, err
	}

	return &documentAccess, nil
}

func (s TrustCenterAccessService) CountDocumentAccesses(
	ctx context.Context,
	trustCenterAccessID gid.GID,
) (int, error) {
	var count int
	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		var documentAccesses coredata.TrustCenterDocumentAccesses
		var err error
		count, err = documentAccesses.CountByTrustCenterAccessID(ctx, conn, s.svc.scope, trustCenterAccessID)
		return err
	})

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s TrustCenterAccessService) ValidateToken(
	ctx context.Context,
	tokenString string,
) (*TrustCenterAccessData, error) {
	token, err := statelesstoken.ValidateToken[TrustCenterAccessData](
		s.svc.trustConfig.TokenSecret,
		s.svc.trustConfig.TokenType,
		tokenString,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot validate trust center access token: %w", err)
	}

	access := &coredata.TrustCenterAccess{}
	err = s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return access.LoadByTrustCenterIDAndEmail(ctx, conn, s.svc.scope, token.Data.TrustCenterID, token.Data.Email)
	})

	if err != nil {
		return nil, fmt.Errorf("access not found or revoked: %w", err)
	}

	return &token.Data, nil
}

func (s TrustCenterAccessService) Create(
	ctx context.Context,
	req *CreateTrustCenterAccessRequest,
) (*coredata.TrustCenterAccess, error) {
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, fmt.Errorf("invalid email address")
	}

	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	now := time.Now()

	var access *coredata.TrustCenterAccess

	err := s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		trustCenter := &coredata.TrustCenter{}
		if err := trustCenter.LoadByID(ctx, tx, s.svc.scope, req.TrustCenterID); err != nil {
			return fmt.Errorf("cannot load trust center: %w", err)
		}
		organizationID := trustCenter.OrganizationID

		documentIDs := []gid.GID{}
		reportIDs := []gid.GID{}

		var allDocuments coredata.Documents
		filter := coredata.NewDocumentTrustCenterFilter()

		if err := allDocuments.LoadAllByOrganizationID(ctx, tx, s.svc.scope, organizationID, filter); err != nil {
			return fmt.Errorf("cannot list documents: %w", err)
		}

		for _, doc := range allDocuments {
			documentIDs = append(documentIDs, doc.ID)
		}

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

		access = &coredata.TrustCenterAccess{
			ID:                                gid.New(s.svc.scope.GetTenantID(), coredata.TrustCenterAccessEntityType),
			TenantID:                          s.svc.scope.GetTenantID(),
			TrustCenterID:                     req.TrustCenterID,
			Email:                             req.Email,
			Name:                              req.Name,
			Active:                            false,
			HasAcceptedNonDisclosureAgreement: false,
			CreatedAt:                         now,
			UpdatedAt:                         now,
		}

		if err := access.Insert(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot insert trust center access: %w", err)
		}

		var documentAccesses coredata.TrustCenterDocumentAccesses
		if err := documentAccesses.BulkInsertDocumentAccesses(ctx, tx, s.svc.scope, access.ID, documentIDs, now); err != nil {
			return fmt.Errorf("cannot bulk insert trust center document accesses: %w", err)
		}

		if err := documentAccesses.BulkInsertReportAccesses(ctx, tx, s.svc.scope, access.ID, reportIDs, now); err != nil {
			return fmt.Errorf("cannot bulk insert trust center report accesses: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return access, nil
}

func (s TrustCenterAccessService) Update(
	ctx context.Context,
	req *UpdateTrustCenterAccessRequest,
) (*coredata.TrustCenterAccess, error) {
	now := time.Now()

	var access *coredata.TrustCenterAccess

	if req.Name != nil && *req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	err := s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		access = &coredata.TrustCenterAccess{}

		if err := access.LoadByID(ctx, tx, s.svc.scope, req.ID); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		shouldSendEmail := req.Active != nil && *req.Active && !access.Active
		if req.Name != nil {
			access.Name = *req.Name
		}
		if req.Active != nil {
			access.Active = *req.Active
		}
		access.UpdatedAt = now

		if err := access.Update(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot update trust center access: %w", err)
		}

		if req.DocumentIDs != nil || req.ReportIDs != nil {
			if err := coredata.DeactivateByTrustCenterAccessID(ctx, tx, s.svc.scope, access.ID, now); err != nil {
				return fmt.Errorf("cannot deactivate existing document accesses: %w", err)
			}

			if req.DocumentIDs != nil {
				if err := coredata.ActivateByDocumentIDs(ctx, tx, s.svc.scope, access.ID, req.DocumentIDs, now); err != nil {
					return fmt.Errorf("cannot activate document accesses: %w", err)
				}
			}

			if req.ReportIDs != nil {
				if err := coredata.ActivateByReportIDs(ctx, tx, s.svc.scope, access.ID, req.ReportIDs, now); err != nil {
					return fmt.Errorf("cannot activate report accesses: %w", err)
				}
			}
		}

		if shouldSendEmail {
			if err := s.sendAccessEmail(ctx, tx, access); err != nil {
				return fmt.Errorf("failed to send access email: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return access, nil
}

func (s TrustCenterAccessService) Delete(
	ctx context.Context,
	req *DeleteTrustCenterAccessRequest,
) error {
	err := s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		access := &coredata.TrustCenterAccess{}

		if err := access.LoadByID(ctx, tx, s.svc.scope, req.ID); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		if err := access.Delete(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot delete trust center access: %w", err)
		}

		return nil
	})

	return err
}

func (s TrustCenterAccessService) sendAccessEmail(ctx context.Context, tx pg.Conn, access *coredata.TrustCenterAccess) error {
	accessToken, err := statelesstoken.NewToken(
		s.svc.trustConfig.TokenSecret,
		s.svc.trustConfig.TokenType,
		s.svc.trustConfig.TokenDuration,
		TrustCenterAccessData{
			TrustCenterID: access.TrustCenterID,
			Email:         access.Email,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot generate access token: %w", err)
	}

	trustCenter := &coredata.TrustCenter{}
	err = trustCenter.LoadByID(ctx, tx, s.svc.scope, access.TrustCenterID)
	if err != nil {
		return fmt.Errorf("cannot load trust center: %w", err)
	}

	organization := &coredata.Organization{}
	err = organization.LoadByID(ctx, tx, s.svc.scope, trustCenter.OrganizationID)
	if err != nil {
		return fmt.Errorf("cannot load organization: %w", err)
	}

	hostname := s.svc.hostname
	path := "/trust/" + trustCenter.Slug + "/access"

	if organization.CustomDomainID != nil {
		customDomain, err := s.svc.CustomDomains.GetOrganizationCustomDomain(ctx, organization.ID)
		if err != nil {
			return fmt.Errorf("cannot load custom domain: %w", err)
		}

		if customDomain == nil || customDomain.SSLStatus != coredata.CustomDomainSSLStatusActive {
			return fmt.Errorf("custom domain is not active")
		}

		hostname = customDomain.Domain
		path = "/access"
	}

	accessURL := url.URL{
		Scheme: "https",
		Host:   hostname,
		Path:   path,
		RawQuery: url.Values{
			"token": []string{accessToken},
		}.Encode(),
	}

	return s.sendTrustCenterAccessEmail(ctx, tx, access.Name, access.Email, organization.Name, accessURL.String())
}

func (s TrustCenterAccessService) sendTrustCenterAccessEmail(
	ctx context.Context,
	tx pg.Conn,
	name string,
	email string,
	companyName string,
	accessURL string,
) error {
	emailData := mailer.EmailData{
		Subject:    fmt.Sprintf(trustCenterAccessEmailSubjectFormat, companyName),
		Header:     trustCenterAccessEmailHeader,
		FullName:   name,
		Body:       fmt.Sprintf(trustCenterAccessEmailBodyFormat, companyName),
		ButtonText: trustCenterAccessEmailButtonText,
		ButtonURL:  accessURL,
		Footer:     trustCenterAccessEmailFooter,
	}

	textBody := bytes.NewBuffer(nil)
	err := mailer.Text().Execute(textBody, emailData)
	if err != nil {
		return fmt.Errorf("cannot execute trust center access text template: %w", err)
	}

	htmlBody := bytes.NewBuffer(nil)
	err = mailer.HTML().Execute(htmlBody, emailData)
	if err != nil {
		return fmt.Errorf("cannot execute trust center access html template: %w", err)
	}

	htmlBodyStr := htmlBody.String()
	accessEmail := coredata.NewEmail(
		name,
		email,
		fmt.Sprintf("Trust Center Access Invitation - %s", companyName),
		textBody.String(),
		&htmlBodyStr,
	)

	if err := accessEmail.Insert(ctx, tx); err != nil {
		return fmt.Errorf("cannot insert access email: %w", err)
	}
	return nil
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
