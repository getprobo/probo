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
	"context"
	"fmt"
	"net/url"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/statelesstoken"
	"go.probo.inc/probo/pkg/validator"
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

	UpdateTrustCenterDocumentAccessRequest struct {
		ID     gid.GID
		Status coredata.TrustCenterDocumentAccessStatus
	}

	UpdateTrustCenterAccessRequest struct {
		ID                      gid.GID
		Name                    *string
		Active                  *bool
		DocumentAccesses        []UpdateTrustCenterDocumentAccessRequest
		ReportAccesses          []UpdateTrustCenterDocumentAccessRequest
		TrustCenterFileAccesses []UpdateTrustCenterDocumentAccessRequest
	}

	TrustCenterAccessData struct {
		TrustCenterID gid.GID `json:"trust_center_id"`
		Email         string  `json:"email"`
	}
)

func (ctcar *CreateTrustCenterAccessRequest) Validate() error {
	v := validator.New()

	v.Check(ctcar.TrustCenterID, "trust_center_id", validator.Required(), validator.GID(coredata.TrustCenterEntityType))
	v.Check(ctcar.Email, "email", validator.Required(), validator.Email())
	v.Check(ctcar.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (utcar *UpdateTrustCenterAccessRequest) Validate() error {
	v := validator.New()

	v.Check(utcar.ID, "id", validator.Required(), validator.GID(coredata.TrustCenterAccessEntityType))
	v.Check(utcar.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	for i, docAccess := range utcar.DocumentAccesses {
		v.Check(docAccess, fmt.Sprintf("documentAccesses[%d].ID", i), validator.Required(), validator.GID(coredata.DocumentEntityType))
	}
	for i, reportAccess := range utcar.ReportAccesses {
		v.Check(reportAccess, fmt.Sprintf("reportAccesses[%d].ID", i), validator.Required(), validator.GID(coredata.ReportEntityType))
	}
	for i, reportAccess := range utcar.TrustCenterFileAccesses {
		v.Check(reportAccess, fmt.Sprintf("trustCenterFileAccesses[%d].ID", i), validator.Required(), validator.GID(coredata.TrustCenterFileEntityType))
	}

	return v.Error()
}

func (s TrustCenterAccessService) ListForTrustCenterID(
	ctx context.Context,
	trustCenterID gid.GID,
	cursor *page.Cursor[coredata.TrustCenterAccessOrderField],
) (*page.Page[*coredata.TrustCenterAccess, coredata.TrustCenterAccessOrderField], error) {
	var accesses coredata.TrustCenterAccesses

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return accesses.LoadByTrustCenterID(ctx, conn, s.svc.scope, trustCenterID, cursor)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(accesses, cursor), nil
}

func (s TrustCenterAccessService) ListAvailableDocumentAccesses(
	ctx context.Context,
	trustCenterAccessID gid.GID,
	cursor *page.Cursor[coredata.TrustCenterDocumentAccessOrderField],
) (*page.Page[*coredata.TrustCenterDocumentAccess, coredata.TrustCenterDocumentAccessOrderField], error) {
	var documentAccesses coredata.TrustCenterDocumentAccesses

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return documentAccesses.LoadAvailableByTrustCenterAccessID(ctx, conn, s.svc.scope, trustCenterAccessID, cursor)
		},
	)

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

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return access.LoadByID(ctx, conn, s.svc.scope, accessID)
		},
	)

	if err != nil {
		return nil, err
	}

	return &access, nil
}

func (s TrustCenterAccessService) CountDocumentAccesses(
	ctx context.Context,
	trustCenterAccessID gid.GID,
) (int, error) {
	var count int
	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var documentAccesses coredata.TrustCenterDocumentAccesses
			var err error
			count, err = documentAccesses.CountByTrustCenterAccessID(ctx, conn, s.svc.scope, trustCenterAccessID)
			return err
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s TrustCenterAccessService) CountPendingRequestDocumentAccesses(
	ctx context.Context,
	trustCenterAccessID gid.GID,
) (int, error) {
	var count int
	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var documentAccesses coredata.TrustCenterDocumentAccesses
			var err error
			count, err = documentAccesses.CountPendingRequestByTrustCenterAccessID(ctx, conn, s.svc.scope, trustCenterAccessID)
			return err
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s TrustCenterAccessService) CountActiveDocumentAccesses(
	ctx context.Context,
	trustCenterAccessID gid.GID,
) (int, error) {
	var count int
	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var documentAccesses coredata.TrustCenterDocumentAccesses
			var err error
			count, err = documentAccesses.CountActiveByTrustCenterAccessID(ctx, conn, s.svc.scope, trustCenterAccessID)
			return err
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s TrustCenterAccessService) Create(
	ctx context.Context,
	req *CreateTrustCenterAccessRequest,
) (*coredata.TrustCenterAccess, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	var access *coredata.TrustCenterAccess
	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, tx, s.svc.scope, req.TrustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			access = &coredata.TrustCenterAccess{
				ID:                                gid.New(s.svc.scope.GetTenantID(), coredata.TrustCenterAccessEntityType),
				OrganizationID:                    trustCenter.OrganizationID,
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

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return access, nil
}

func (s TrustCenterAccessService) Update(
	ctx context.Context,
	req *UpdateTrustCenterAccessRequest,
) (*coredata.TrustCenterAccess, error) {

	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	var access *coredata.TrustCenterAccess
	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
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

			var tcdas coredata.TrustCenterDocumentAccesses

			var documentData []coredata.MergeTrustCenterDocumentAccessesData
			for _, d := range req.DocumentAccesses {
				documentData = append(documentData, coredata.MergeTrustCenterDocumentAccessesData{
					ID:     d.ID,
					Status: d.Status,
				})
			}

			if err := tcdas.MergeDocumentAccesses(ctx, tx, s.svc.scope, access.OrganizationID, access.ID, documentData); err != nil {
				return fmt.Errorf("cannot merge document accesses: %w", err)
			}

			var reportData []coredata.MergeTrustCenterDocumentAccessesData
			for _, d := range req.ReportAccesses {
				reportData = append(reportData, coredata.MergeTrustCenterDocumentAccessesData{
					ID:     d.ID,
					Status: d.Status,
				})
			}

			if err := tcdas.MergeReportAccesses(ctx, tx, s.svc.scope, access.OrganizationID, access.ID, reportData); err != nil {
				return fmt.Errorf("cannot merge report accesses: %w", err)
			}

			var fileData []coredata.MergeTrustCenterDocumentAccessesData
			for _, d := range req.TrustCenterFileAccesses {
				fileData = append(fileData, coredata.MergeTrustCenterDocumentAccessesData{
					ID:     d.ID,
					Status: d.Status,
				})
			}

			if err := tcdas.MergeTrustCenterFileAccesses(ctx, tx, s.svc.scope, access.OrganizationID, access.ID, fileData); err != nil {
				return fmt.Errorf("cannot merge trust center file accesses: %w", err)
			}

			if shouldSendEmail {
				if err := s.sendAccessEmail(ctx, tx, access); err != nil {
					return fmt.Errorf("cannot send access email: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return access, nil
}

func (s TrustCenterAccessService) Delete(
	ctx context.Context,
	trustCenterAccessID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			access := &coredata.TrustCenterAccess{}

			if err := access.LoadByID(ctx, tx, s.svc.scope, trustCenterAccessID); err != nil {
				return fmt.Errorf("cannot load trust center access: %w", err)
			}

			if err := access.Delete(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete trust center access: %w", err)
			}

			return nil
		},
	)

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

	baseURLParsed, err := url.Parse(s.svc.baseURL)
	if err != nil {
		return fmt.Errorf("cannot parse base URL: %w", err)
	}

	hostname := baseURLParsed.Host
	scheme := baseURLParsed.Scheme
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
		scheme = "https"
		path = "/access"
	}

	accessURL := url.URL{
		Scheme: scheme,
		Host:   hostname,
		Path:   path,
		RawQuery: url.Values{
			"token": []string{accessToken},
		}.Encode(),
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

func (s TrustCenterAccessService) sendTrustCenterAccessEmail(
	ctx context.Context,
	tx pg.Conn,
	name string,
	email string,
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
