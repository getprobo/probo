// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package management

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/slack"
	"go.probo.inc/probo/pkg/validator"
)

type (
	CreateAccessRequest struct {
		TrustCenterID gid.GID
		IdentityID    gid.GID
	}

	UpdateDocumentAccessRequest struct {
		ID     gid.GID
		Status coredata.TrustCenterDocumentAccessStatus
	}

	UpdateAccessRequest struct {
		ID                      gid.GID
		DocumentAccesses        []UpdateDocumentAccessRequest
		ReportAccesses          []UpdateDocumentAccessRequest
		TrustCenterFileAccesses []UpdateDocumentAccessRequest
	}

	AccessData struct {
		TrustCenterID gid.GID   `json:"trust_center_id"`
		Email         mail.Addr `json:"email"`
	}
)

func (utcar *UpdateAccessRequest) Validate() error {
	v := validator.New()

	v.Check(utcar.ID, "id", validator.Required(), validator.GID(coredata.TrustCenterAccessEntityType))

	for i, docAccess := range utcar.DocumentAccesses {
		v.Check(docAccess.ID, fmt.Sprintf("documentAccesses[%d].ID", i), validator.Required(), validator.GID(coredata.DocumentEntityType))
	}

	for i, reportAccess := range utcar.ReportAccesses {
		v.Check(reportAccess.ID, fmt.Sprintf("reportAccesses[%d].ID", i), validator.Required(), validator.GID(coredata.FileEntityType))
	}

	for i, reportAccess := range utcar.TrustCenterFileAccesses {
		v.Check(reportAccess.ID, fmt.Sprintf("trustCenterFileAccesses[%d].ID", i), validator.Required(), validator.GID(coredata.TrustCenterFileEntityType))
	}

	return v.Error()
}

func (s *Service) ListAccesses(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	cursor *page.Cursor[coredata.TrustCenterAccessOrderField],
) (*page.Page[*coredata.TrustCenterAccess, coredata.TrustCenterAccessOrderField], error) {
	var accesses coredata.TrustCenterAccesses

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return accesses.LoadByTrustCenterID(ctx, conn, scope, compliancePageID, cursor)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(accesses, cursor), nil
}

func (s *Service) ListAvailableDocumentAccesses(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterAccessID gid.GID,
	cursor *page.Cursor[coredata.TrustCenterDocumentAccessOrderField],
) (*page.Page[*coredata.TrustCenterDocumentAccess, coredata.TrustCenterDocumentAccessOrderField], error) {
	var documentAccesses coredata.TrustCenterDocumentAccesses

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documentAccesses.LoadAvailableByTrustCenterAccessID(ctx, conn, scope, trustCenterAccessID, cursor)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(documentAccesses, cursor), nil
}

func (s *Service) GetAccess(
	ctx context.Context,
	scope coredata.Scoper,
	accessID gid.GID,
) (*coredata.TrustCenterAccess, error) {
	var access coredata.TrustCenterAccess

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return access.LoadByID(ctx, conn, scope, accessID)
		},
	)
	if err != nil {
		return nil, err
	}

	return &access, nil
}

func (s *Service) CountDocumentAccesses(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterAccessID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var (
				documentAccesses coredata.TrustCenterDocumentAccesses
				err              error
			)

			count, err = documentAccesses.CountByTrustCenterAccessID(ctx, conn, scope, trustCenterAccessID)

			return err
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) CountPendingRequestDocumentAccesses(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterAccessID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var (
				documentAccesses coredata.TrustCenterDocumentAccesses
				err              error
			)

			count, err = documentAccesses.CountPendingRequestByTrustCenterAccessID(ctx, conn, scope, trustCenterAccessID)

			return err
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) CountActiveDocumentAccesses(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterAccessID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var (
				documentAccesses coredata.TrustCenterDocumentAccesses
				err              error
			)

			count, err = documentAccesses.CountActiveByTrustCenterAccessID(ctx, conn, scope, trustCenterAccessID)

			return err
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) UpdateAccess(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateAccessRequest,
) (*coredata.TrustCenterAccess, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var (
		access                    *coredata.TrustCenterAccess
		trustCenterAcessActivated bool
		shouldUpdateSlackMessage  bool
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			access = &coredata.TrustCenterAccess{}

			if err := access.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load compliance page access: %w", err)
			}

			var tcdas coredata.TrustCenterDocumentAccesses

			if len(req.DocumentAccesses) > 0 {
				var documentData []coredata.MergeTrustCenterDocumentAccessesData

				documentIDs := make([]gid.GID, 0, len(req.DocumentAccesses))
				for _, d := range req.DocumentAccesses {
					documentData = append(documentData, coredata.MergeTrustCenterDocumentAccessesData{
						ID:     d.ID,
						Status: d.Status,
					})

					documentIDs = append(documentIDs, d.ID)
				}

				documents := &coredata.Documents{}
				if err := documents.LoadByIDs(ctx, tx, scope, documentIDs); err != nil {
					return fmt.Errorf("cannot load documents: %w", err)
				}

				if err := tcdas.MergeDocumentAccesses(ctx, tx, scope, access.OrganizationID, access.ID, documentData); err != nil {
					return fmt.Errorf("cannot merge document accesses: %w", err)
				}
			}

			if len(req.ReportAccesses) > 0 {
				var reportData []coredata.MergeTrustCenterDocumentAccessesData

				reportIDs := make([]gid.GID, 0, len(req.ReportAccesses))
				for _, d := range req.ReportAccesses {
					reportData = append(reportData, coredata.MergeTrustCenterDocumentAccessesData{
						ID:     d.ID,
						Status: d.Status,
					})

					reportIDs = append(reportIDs, d.ID)
				}

				files := &coredata.Files{}
				if err := files.LoadByIDs(ctx, tx, scope, reportIDs); err != nil {
					return fmt.Errorf("cannot load report files: %w", err)
				}

				if err := tcdas.MergeReportFileAccesses(ctx, tx, scope, access.OrganizationID, access.ID, reportData); err != nil {
					return fmt.Errorf("cannot merge report accesses: %w", err)
				}
			}

			if len(req.TrustCenterFileAccesses) > 0 {
				var fileData []coredata.MergeTrustCenterDocumentAccessesData

				trustCenterFileIDs := make([]gid.GID, 0, len(req.TrustCenterFileAccesses))
				for _, d := range req.TrustCenterFileAccesses {
					fileData = append(fileData, coredata.MergeTrustCenterDocumentAccessesData{
						ID:     d.ID,
						Status: d.Status,
					})

					trustCenterFileIDs = append(trustCenterFileIDs, d.ID)
				}

				trustCenterFiles := &coredata.TrustCenterFiles{}
				if err := trustCenterFiles.LoadByIDs(ctx, tx, scope, trustCenterFileIDs); err != nil {
					return fmt.Errorf("cannot load compliance page files: %w", err)
				}

				if err := tcdas.MergeTrustCenterFileAccesses(ctx, tx, scope, access.OrganizationID, access.ID, fileData); err != nil {
					return fmt.Errorf("cannot merge compliance page file accesses: %w", err)
				}
			}

			if trustCenterAcessActivated {
				if err := s.sendAccessEmail(ctx, scope, tx, access); err != nil {
					return fmt.Errorf("cannot send access email: %w", err)
				}
			}

			shouldUpdateSlackMessage = trustCenterAcessActivated ||
				len(req.DocumentAccesses) > 0 ||
				len(req.ReportAccesses) > 0 ||
				len(req.TrustCenterFileAccesses) > 0

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if shouldUpdateSlackMessage {
		if err := s.SlackMessages.QueueSlackNotification(ctx, scope, access.IdentityID, access.TrustCenterID); err != nil {
			if !errors.Is(err, slack.ErrNoSlackConnector) {
				return nil, fmt.Errorf("cannot queue slack notification: %w", err)
			}
		}
	}

	return access, nil
}

func (s *Service) DeleteAccess(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterAccessID gid.GID,
) error {
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			access := &coredata.TrustCenterAccess{}

			if err := access.LoadByID(ctx, tx, scope, trustCenterAccessID); err != nil {
				return fmt.Errorf("cannot load compliance page access: %w", err)
			}

			if err := access.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete compliance page access: %w", err)
			}

			return nil
		},
	)

	return err
}

func (s *Service) sendAccessEmail(
	ctx context.Context,
	scope coredata.Scoper,
	tx pg.Tx,
	access *coredata.TrustCenterAccess,
) error {
	organization := &coredata.Organization{}
	if err := organization.LoadByID(ctx, tx, scope, access.OrganizationID); err != nil {
		return fmt.Errorf("cannot load organization: %w", err)
	}

	now := time.Now()
	access.UpdatedAt = now

	if err := access.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update compliance page access with expiration: %w", err)
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByIdentityIDAndOrganizationID(
		ctx,
		tx,
		scope,
		access.IdentityID,
		access.OrganizationID,
	); err != nil {
		return fmt.Errorf("cannot load profile: %w", err)
	}

	emailPresenterCfg, err := s.EmailPresenterConfig(ctx, scope, access.TrustCenterID)
	if err != nil {
		return fmt.Errorf("cannot get compliance page email presenter config: %w", err)
	}

	emailPresenter := emails.NewPresenterFromConfig(emailPresenterCfg, profile.FullName)

	subject, textBody, htmlBody, err := emailPresenter.RenderTrustCenterAccess(ctx, organization.Name)
	if err != nil {
		return fmt.Errorf("cannot render compliance page access email: %w", err)
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
