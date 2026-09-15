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
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/bot"
	portal "go.probo.inc/probo/pkg/complianceportal"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	CreateAccessRequest struct {
		CompliancePortalID      gid.GID
		ProfileID               *gid.GID
		Email                   *mail.Addr
		DocumentIDs             []gid.GID
		ReportFileIDs           []gid.GID
		CompliancePortalFileIDs []gid.GID
	}

	UpdateDocumentAccessRequest struct {
		ID     gid.GID
		Status coredata.CompliancePortalDocumentAccessStatus
	}

	UpdateAccessRequest struct {
		ID                           gid.GID
		State                        *coredata.CompliancePortalAccessState
		DocumentAccesses             []UpdateDocumentAccessRequest
		ReportAccesses               []UpdateDocumentAccessRequest
		CompliancePortalFileAccesses []UpdateDocumentAccessRequest
	}

	AccessData struct {
		CompliancePortalID gid.GID   `json:"compliance_portal_id"`
		Email              mail.Addr `json:"email"`
	}
)

func (utcar *UpdateAccessRequest) Validate() error {
	v := validator.New()

	v.Check(utcar.ID, "id", validator.Required(), validator.GID(coredata.CompliancePortalAccessEntityType))

	for i, docAccess := range utcar.DocumentAccesses {
		v.Check(docAccess.ID, fmt.Sprintf("documentAccesses[%d].ID", i), validator.Required(), validator.GID(coredata.DocumentEntityType))
	}

	for i, reportAccess := range utcar.ReportAccesses {
		v.Check(reportAccess.ID, fmt.Sprintf("reportAccesses[%d].ID", i), validator.Required(), validator.GID(coredata.FileEntityType))
	}

	for i, reportAccess := range utcar.CompliancePortalFileAccesses {
		v.Check(reportAccess.ID, fmt.Sprintf("compliancePortalFileAccesses[%d].ID", i), validator.Required(), validator.GID(coredata.CompliancePortalFileEntityType))
	}

	return v.Error()
}

func (req *CreateAccessRequest) Validate() error {
	v := validator.New()

	v.Check(
		req.CompliancePortalID,
		"compliancePortalId",
		validator.Required(),
		validator.GID(coredata.CompliancePortalEntityType),
	)

	hasProfile := req.ProfileID != nil
	hasEmail := req.Email != nil

	switch {
	case hasProfile == hasEmail:
		v.Check("", "email", validator.Required())
	case hasProfile:
		v.Check(*req.ProfileID, "profileId", validator.GID(coredata.MembershipProfileEntityType))
	case hasEmail:
		v.Check(*req.Email, "email", validator.NotEmpty())
	}

	for i, documentID := range req.DocumentIDs {
		v.Check(
			documentID,
			fmt.Sprintf("documents[%d]", i),
			validator.Required(),
			validator.GID(coredata.DocumentEntityType),
		)
	}

	for i, reportFileID := range req.ReportFileIDs {
		v.Check(
			reportFileID,
			fmt.Sprintf("reports[%d]", i),
			validator.Required(),
			validator.GID(coredata.FileEntityType),
		)
	}

	for i, fileID := range req.CompliancePortalFileIDs {
		v.Check(
			fileID,
			fmt.Sprintf("compliancePortalFiles[%d]", i),
			validator.Required(),
			validator.GID(coredata.CompliancePortalFileEntityType),
		)
	}

	return v.Error()
}

func managementAccessEventKey(req *UpdateAccessRequest) string {
	components := make(
		[]string,
		0,
		len(req.DocumentAccesses)+
			len(req.ReportAccesses)+
			len(req.CompliancePortalFileAccesses),
	)
	for _, access := range req.DocumentAccesses {
		components = append(
			components,
			fmt.Sprintf("document:%s:%s", access.ID, access.Status),
		)
	}

	for _, access := range req.ReportAccesses {
		components = append(
			components,
			fmt.Sprintf("report:%s:%s", access.ID, access.Status),
		)
	}

	for _, access := range req.CompliancePortalFileAccesses {
		components = append(
			components,
			fmt.Sprintf("file:%s:%s", access.ID, access.Status),
		)
	}

	return bot.StableEventKey("management-update", components...)
}

func (s *Service) enqueueAccessBotUpdateIfPosted(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	access *coredata.CompliancePortalAccess,
	eventKey string,
) error {
	var subject coredata.BotThreadSubject

	err := subject.LoadBySubject(
		ctx,
		tx,
		scope,
		access.OrganizationID,
		portal.AccessSubjectNamespace,
		access.ID.String(),
	)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("cannot load compliance portal bot thread: %w", err)
	}

	if _, err := s.bot.EnqueueMessage(
		ctx,
		tx,
		scope,
		bot.MessageParams{
			OrganizationID: access.OrganizationID,
			Capability:     portal.AccessCapability,
			MessageType:    portal.AccessMessageType,
			Attributes: map[string]any{
				portal.AccessIDAttribute: access.ID.String(),
			},
			SubjectNamespace: portal.AccessSubjectNamespace,
			SubjectKey:       access.ID.String(),
			EventKey:         eventKey,
			Purpose:          coredata.BotMessagePurposeUpdate,
		},
	); err != nil {
		return fmt.Errorf("cannot enqueue compliance portal bot message: %w", err)
	}

	return nil
}

func (s *Service) ListAccesses(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	cursor *page.Cursor[coredata.CompliancePortalAccessOrderField],
	filter *coredata.CompliancePortalAccessFilter,
) (*page.Page[*coredata.CompliancePortalAccess, coredata.CompliancePortalAccessOrderField], error) {
	var accesses coredata.CompliancePortalAccesses

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return accesses.LoadByCompliancePortalID(ctx, conn, scope, compliancePageID, cursor, filter)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(accesses, cursor), nil
}

func (s *Service) ListMemberCandidates(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	query string,
) ([]*coredata.MembershipProfile, error) {
	var profiles coredata.MembershipProfiles

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			portal := &coredata.CompliancePortal{}
			if err := portal.LoadByID(ctx, conn, scope, compliancePortalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			filter := coredata.NewMembershipProfileFilter(nil).
				WithMembership().
				WithoutCompliancePortalID(compliancePortalID)

			if query != "" {
				filter = filter.WithQuery(&query)
			}

			cursor := page.NewCursor(
				MemberCandidateLimit,
				nil,
				page.Head,
				page.OrderBy[coredata.MembershipProfileOrderField]{
					Field:     coredata.MembershipProfileOrderFieldFullName,
					Direction: page.OrderDirectionAsc,
				},
			)

			return profiles.LoadByOrganizationID(
				ctx,
				conn,
				scope,
				portal.OrganizationID,
				cursor,
				filter,
			)
		},
	)
	if err != nil {
		return nil, err
	}

	return profiles, nil
}

func (s *Service) GetAccess(
	ctx context.Context,
	scope coredata.Scoper,
	accessID gid.GID,
) (*coredata.CompliancePortalAccess, error) {
	var access coredata.CompliancePortalAccess

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

func (s *Service) CreateAccess(
	ctx context.Context,
	scope coredata.Scoper,
	req *CreateAccessRequest,
) (*coredata.CompliancePortalAccess, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var access *coredata.CompliancePortalAccess

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			now := time.Now()
			compliancePortal := &coredata.CompliancePortal{}

			if err := compliancePortal.LoadByID(ctx, tx, scope, req.CompliancePortalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			identity, _, err := s.resolveAccessIdentity(
				ctx,
				tx,
				scope,
				compliancePortal,
				req,
				now,
			)
			if err != nil {
				return err
			}

			existing := &coredata.CompliancePortalAccess{}

			err = existing.LoadByCompliancePortalIDAndIdentityID(
				ctx,
				tx,
				scope,
				compliancePortal.ID,
				identity.ID,
			)
			if err == nil {
				return coredata.ErrResourceAlreadyExists
			}

			if !errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load compliance portal access: %w", err)
			}

			access = &coredata.CompliancePortalAccess{
				ID:                 gid.New(scope.GetTenantID(), coredata.CompliancePortalAccessEntityType),
				OrganizationID:     compliancePortal.OrganizationID,
				TenantID:           scope.GetTenantID(),
				IdentityID:         identity.ID,
				CompliancePortalID: compliancePortal.ID,
				State:              coredata.CompliancePortalAccessStateActive,
				CreatedAt:          now,
				UpdatedAt:          now,
			}

			if compliancePortal.NonDisclosureAgreementFileID != nil && s.esign != nil {
				sig, err := s.esign.CreateSignature(
					ctx,
					tx,
					&esign.CreateSignatureRequest{
						OrganizationID: access.OrganizationID,
						DocumentType:   coredata.ElectronicSignatureDocumentTypeNDA,
						FileID:         *compliancePortal.NonDisclosureAgreementFileID,
						SignerEmail:    identity.EmailAddress,
						ConsentText:    NDAConsentText(ref.UnrefOrZero(compliancePortal.Email)),
					},
				)
				if err != nil {
					return fmt.Errorf("cannot create pending signature: %w", err)
				}

				access.ElectronicSignatureID = &sig.ID
			}

			if err := access.Insert(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return err
				}

				return fmt.Errorf("cannot insert compliance portal access: %w", err)
			}

			if err := s.grantCreatedAccessTargets(ctx, tx, scope, access, req); err != nil {
				return err
			}

			if createAccessGrantsTargets(req) {
				if err := s.sendAccessEmail(ctx, scope, tx, access); err != nil {
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

func (s *Service) ListAccessResources(
	ctx context.Context,
	scope coredata.Scoper,
	accessID gid.GID,
	cursor *page.Cursor[coredata.CompliancePortalAccessResourceOrderField],
	filter *coredata.CompliancePortalAccessResourceFilter,
) (*page.Page[*coredata.CompliancePortalAccessResource, coredata.CompliancePortalAccessResourceOrderField], error) {
	var resources coredata.CompliancePortalAccessResources

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var access coredata.CompliancePortalAccess
			if err := access.LoadByID(ctx, conn, scope, accessID); err != nil {
				return fmt.Errorf("cannot load compliance portal access: %w", err)
			}

			return resources.LoadByCompliancePortalAccessID(
				ctx,
				conn,
				scope,
				access.ID,
				access.OrganizationID,
				access.CompliancePortalID,
				cursor,
				filter,
			)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(resources, cursor), nil
}

func (s *Service) CountAccessResources(
	ctx context.Context,
	scope coredata.Scoper,
	accessID gid.GID,
	filter *coredata.CompliancePortalAccessResourceFilter,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var access coredata.CompliancePortalAccess
			if err := access.LoadByID(ctx, conn, scope, accessID); err != nil {
				return fmt.Errorf("cannot load compliance portal access: %w", err)
			}

			var (
				resources coredata.CompliancePortalAccessResources
				err       error
			)

			count, err = resources.CountByCompliancePortalAccessID(
				ctx,
				conn,
				scope,
				access.ID,
				access.OrganizationID,
				access.CompliancePortalID,
				filter,
			)

			return err
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) GetDocumentAccess(
	ctx context.Context,
	scope coredata.Scoper,
	documentAccessID gid.GID,
) (*coredata.CompliancePortalDocumentAccess, error) {
	var documentAccess coredata.CompliancePortalDocumentAccess

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documentAccess.LoadByID(ctx, conn, scope, documentAccessID)
		},
	)
	if err != nil {
		return nil, err
	}

	return &documentAccess, nil
}

func (s *Service) GetDocumentAccessesByDocumentIDs(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalAccessID gid.GID,
	documentIDs []gid.GID,
) (coredata.CompliancePortalDocumentAccesses, error) {
	var documentAccesses coredata.CompliancePortalDocumentAccesses

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documentAccesses.LoadByCompliancePortalAccessIDAndDocumentIDs(
				ctx,
				conn,
				scope,
				compliancePortalAccessID,
				documentIDs,
			)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load document accesses: %w", err)
	}

	return documentAccesses, nil
}

func (s *Service) GetDocumentAccessesByReportFileIDs(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalAccessID gid.GID,
	reportFileIDs []gid.GID,
) (coredata.CompliancePortalDocumentAccesses, error) {
	var documentAccesses coredata.CompliancePortalDocumentAccesses

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documentAccesses.LoadByCompliancePortalAccessIDAndReportFileIDs(
				ctx,
				conn,
				scope,
				compliancePortalAccessID,
				reportFileIDs,
			)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load report file accesses: %w", err)
	}

	return documentAccesses, nil
}

func (s *Service) GetDocumentAccessesByCompliancePortalFileIDs(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalAccessID gid.GID,
	compliancePortalFileIDs []gid.GID,
) (coredata.CompliancePortalDocumentAccesses, error) {
	var documentAccesses coredata.CompliancePortalDocumentAccesses

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documentAccesses.LoadByCompliancePortalAccessIDAndCompliancePortalFileIDs(
				ctx,
				conn,
				scope,
				compliancePortalAccessID,
				compliancePortalFileIDs,
			)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load compliance portal file accesses: %w", err)
	}

	return documentAccesses, nil
}

func (s *Service) CountDocumentAccesses(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalAccessID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var (
				documentAccesses coredata.CompliancePortalDocumentAccesses
				err              error
			)

			count, err = documentAccesses.CountByCompliancePortalAccessID(ctx, conn, scope, compliancePortalAccessID)

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
	compliancePortalAccessID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var (
				documentAccesses coredata.CompliancePortalDocumentAccesses
				err              error
			)

			count, err = documentAccesses.CountPendingRequestByCompliancePortalAccessID(ctx, conn, scope, compliancePortalAccessID)

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
	compliancePortalAccessID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var (
				documentAccesses coredata.CompliancePortalDocumentAccesses
				err              error
			)

			count, err = documentAccesses.CountActiveByCompliancePortalAccessID(ctx, conn, scope, compliancePortalAccessID)

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
) (*coredata.CompliancePortalAccess, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var (
		access                         *coredata.CompliancePortalAccess
		compliancePortalAcessActivated bool
		shouldRefreshBotMessage        bool
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			access = &coredata.CompliancePortalAccess{}

			if err := access.LoadByIDForUpdate(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load compliance page access: %w", err)
			}

			if req.State != nil && *req.State != access.State {
				access.State = *req.State

				access.UpdatedAt = time.Now()
				if err := access.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update compliance portal access state: %w", err)
				}

				compliancePortalAcessActivated = *req.State == coredata.CompliancePortalAccessStateActive
			}

			var tcdas coredata.CompliancePortalDocumentAccesses

			var (
				grantedDocumentIDs = map[gid.GID]struct{}{}
				grantedReportIDs   = map[gid.GID]struct{}{}
				grantedFileIDs     = map[gid.GID]struct{}{}
			)

			if len(req.DocumentAccesses) > 0 {
				var documentData []coredata.UpsertCompliancePortalDocumentAccessesData

				documentIDs := make([]gid.GID, 0, len(req.DocumentAccesses))
				for _, d := range req.DocumentAccesses {
					documentData = append(documentData, coredata.UpsertCompliancePortalDocumentAccessesData{
						ID:     d.ID,
						Status: d.Status,
					})

					documentIDs = append(documentIDs, d.ID)
				}

				var existing coredata.CompliancePortalDocumentAccesses
				if err := existing.LoadByCompliancePortalAccessIDAndDocumentIDs(
					ctx,
					tx,
					scope,
					access.ID,
					documentIDs,
				); err != nil {
					return fmt.Errorf("cannot load existing document accesses: %w", err)
				}

				grantedDocumentIDs = grantedTargetIDs(
					existing,
					func(row *coredata.CompliancePortalDocumentAccess) *gid.GID {
						return row.DocumentID
					},
				)

				documents := &coredata.Documents{}
				if err := documents.LoadByIDs(ctx, tx, scope, documentIDs); err != nil {
					return fmt.Errorf("cannot load documents: %w", err)
				}

				if err := validatePortalAccessTargets(
					ctx,
					tx,
					scope,
					access.CompliancePortalID,
					documentIDs,
					nil,
					nil,
				); err != nil {
					return err
				}

				if err := tcdas.UpsertDocumentAccesses(ctx, tx, scope, access.OrganizationID, access.ID, documentData); err != nil {
					return fmt.Errorf("cannot upsert document accesses: %w", err)
				}
			}

			if len(req.ReportAccesses) > 0 {
				var reportData []coredata.UpsertCompliancePortalDocumentAccessesData

				reportIDs := make([]gid.GID, 0, len(req.ReportAccesses))
				for _, d := range req.ReportAccesses {
					reportData = append(reportData, coredata.UpsertCompliancePortalDocumentAccessesData{
						ID:     d.ID,
						Status: d.Status,
					})

					reportIDs = append(reportIDs, d.ID)
				}

				var existing coredata.CompliancePortalDocumentAccesses
				if err := existing.LoadByCompliancePortalAccessIDAndReportFileIDs(
					ctx,
					tx,
					scope,
					access.ID,
					reportIDs,
				); err != nil {
					return fmt.Errorf("cannot load existing report accesses: %w", err)
				}

				grantedReportIDs = grantedTargetIDs(
					existing,
					func(row *coredata.CompliancePortalDocumentAccess) *gid.GID {
						return row.ReportFileID
					},
				)

				files := &coredata.Files{}
				if err := files.LoadByIDs(ctx, tx, scope, reportIDs); err != nil {
					return fmt.Errorf("cannot load report files: %w", err)
				}

				if err := validatePortalAccessTargets(
					ctx,
					tx,
					scope,
					access.CompliancePortalID,
					nil,
					reportIDs,
					nil,
				); err != nil {
					return err
				}

				if err := tcdas.UpsertReportFileAccesses(ctx, tx, scope, access.OrganizationID, access.ID, reportData); err != nil {
					return fmt.Errorf("cannot upsert report accesses: %w", err)
				}
			}

			if len(req.CompliancePortalFileAccesses) > 0 {
				var fileData []coredata.UpsertCompliancePortalDocumentAccessesData

				compliancePortalFileIDs := make([]gid.GID, 0, len(req.CompliancePortalFileAccesses))
				for _, d := range req.CompliancePortalFileAccesses {
					fileData = append(fileData, coredata.UpsertCompliancePortalDocumentAccessesData{
						ID:     d.ID,
						Status: d.Status,
					})

					compliancePortalFileIDs = append(compliancePortalFileIDs, d.ID)
				}

				var existing coredata.CompliancePortalDocumentAccesses
				if err := existing.LoadByCompliancePortalAccessIDAndCompliancePortalFileIDs(
					ctx,
					tx,
					scope,
					access.ID,
					compliancePortalFileIDs,
				); err != nil {
					return fmt.Errorf("cannot load existing compliance portal file accesses: %w", err)
				}

				grantedFileIDs = grantedTargetIDs(
					existing,
					func(row *coredata.CompliancePortalDocumentAccess) *gid.GID {
						return row.CompliancePortalFileID
					},
				)

				compliancePortalFiles := &coredata.CompliancePortalFiles{}
				if err := compliancePortalFiles.LoadByIDs(ctx, tx, scope, compliancePortalFileIDs); err != nil {
					return fmt.Errorf("cannot load compliance page files: %w", err)
				}

				if err := validatePortalAccessTargets(
					ctx,
					tx,
					scope,
					access.CompliancePortalID,
					nil,
					nil,
					compliancePortalFileIDs,
				); err != nil {
					return err
				}

				if err := tcdas.UpsertCompliancePortalFileAccesses(ctx, tx, scope, access.OrganizationID, access.ID, fileData); err != nil {
					return fmt.Errorf("cannot upsert compliance page file accesses: %w", err)
				}
			}

			newlyGranted := updateAccessNewlyGrantsTargets(
				req,
				grantedDocumentIDs,
				grantedReportIDs,
				grantedFileIDs,
			)
			if access.State == coredata.CompliancePortalAccessStateActive && newlyGranted {
				if err := s.sendAccessEmail(ctx, scope, tx, access); err != nil {
					return fmt.Errorf("cannot send access email: %w", err)
				}
			}

			shouldRefreshBotMessage = compliancePortalAcessActivated ||
				len(req.DocumentAccesses) > 0 ||
				len(req.ReportAccesses) > 0 ||
				len(req.CompliancePortalFileAccesses) > 0

			if shouldRefreshBotMessage {
				if err := s.enqueueAccessBotUpdateIfPosted(
					ctx,
					tx,
					scope,
					access,
					managementAccessEventKey(req),
				); err != nil {
					return err
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

func (s *Service) DeactivateAccess(
	ctx context.Context,
	scope coredata.Scoper,
	id gid.GID,
) (*coredata.CompliancePortalAccess, error) {
	state := coredata.CompliancePortalAccessStateDeactivated

	return s.UpdateAccess(
		ctx,
		scope,
		&UpdateAccessRequest{
			ID:    id,
			State: &state,
		},
	)
}

func (s *Service) ActivateAccess(
	ctx context.Context,
	scope coredata.Scoper,
	id gid.GID,
) (*coredata.CompliancePortalAccess, error) {
	state := coredata.CompliancePortalAccessStateActive

	return s.UpdateAccess(
		ctx,
		scope,
		&UpdateAccessRequest{
			ID:    id,
			State: &state,
		},
	)
}

func (s *Service) DeleteAccess(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalAccessID gid.GID,
) error {
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			access := &coredata.CompliancePortalAccess{}

			if err := access.LoadByID(ctx, tx, scope, compliancePortalAccessID); err != nil {
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

func validatePortalAccessTargets(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	documentIDs []gid.GID,
	reportFileIDs []gid.GID,
	compliancePortalFileIDs []gid.GID,
) error {
	for _, documentID := range documentIDs {
		link := &coredata.CompliancePortalDocument{}

		err := link.LoadByCompliancePortalIDAndDocumentID(
			ctx,
			tx,
			scope,
			compliancePortalID,
			documentID,
		)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return coredata.ErrResourceNotFound
			}

			return fmt.Errorf("cannot load portal document link: %w", err)
		}
	}

	for _, reportFileID := range reportFileIDs {
		audit := &coredata.Audit{}

		if err := audit.LoadByReportFileID(ctx, tx, scope, reportFileID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return coredata.ErrResourceNotFound
			}

			return fmt.Errorf("cannot load audit for report file: %w", err)
		}

		link := &coredata.CompliancePortalAudit{}

		err := link.LoadByCompliancePortalIDAndAuditID(
			ctx,
			tx,
			scope,
			compliancePortalID,
			audit.ID,
		)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return coredata.ErrResourceNotFound
			}

			return fmt.Errorf("cannot load portal audit link: %w", err)
		}
	}

	for _, compliancePortalFileID := range compliancePortalFileIDs {
		file := &coredata.CompliancePortalFile{}

		if err := file.LoadByCompliancePortalIDAndID(
			ctx,
			tx,
			scope,
			compliancePortalID,
			compliancePortalFileID,
		); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return coredata.ErrResourceNotFound
			}

			return fmt.Errorf("cannot load compliance portal file: %w", err)
		}
	}

	return nil
}

func (s *Service) sendAccessEmail(
	ctx context.Context,
	scope coredata.Scoper,
	tx pg.Tx,
	access *coredata.CompliancePortalAccess,
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

	emailPresenterCfg, err := s.EmailPresenterConfig(ctx, scope, access.CompliancePortalID)
	if err != nil {
		return fmt.Errorf("cannot get compliance page email presenter config: %w", err)
	}

	emailPresenter := emails.NewPresenterFromConfig(emailPresenterCfg, profile.FullName)

	subject, textBody, htmlBody, err := emailPresenter.RenderCompliancePortalAccess(ctx, organization.Name)
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

func NDAConsentText(contactEmail string) string {
	if contactEmail == "" {
		contactEmail = DefaultNDAContactEmail
	}

	return fmt.Sprintf(
		"By clicking \"Review and sign\", I consent to sign this document electronically and agree that my electronic signature has the same legal validity as a handwritten signature. If you have questions about the NDA, please contact %s.",
		contactEmail,
	)
}

func createAccessGrantsTargets(req *CreateAccessRequest) bool {
	return len(req.DocumentIDs) > 0 ||
		len(req.ReportFileIDs) > 0 ||
		len(req.CompliancePortalFileIDs) > 0
}

func grantedTargetIDs(
	accesses coredata.CompliancePortalDocumentAccesses,
	idOf func(*coredata.CompliancePortalDocumentAccess) *gid.GID,
) map[gid.GID]struct{} {
	ids := make(map[gid.GID]struct{}, len(accesses))
	for _, access := range accesses {
		if access.Status != coredata.CompliancePortalDocumentAccessStatusGranted {
			continue
		}

		id := idOf(access)
		if id == nil {
			continue
		}

		ids[*id] = struct{}{}
	}

	return ids
}

func requestsNewGrant(
	updates []UpdateDocumentAccessRequest,
	alreadyGranted map[gid.GID]struct{},
) bool {
	for _, update := range updates {
		if update.Status != coredata.CompliancePortalDocumentAccessStatusGranted {
			continue
		}

		if _, exists := alreadyGranted[update.ID]; !exists {
			return true
		}
	}

	return false
}

func updateAccessNewlyGrantsTargets(
	req *UpdateAccessRequest,
	grantedDocumentIDs map[gid.GID]struct{},
	grantedReportIDs map[gid.GID]struct{},
	grantedFileIDs map[gid.GID]struct{},
) bool {
	return requestsNewGrant(req.DocumentAccesses, grantedDocumentIDs) ||
		requestsNewGrant(req.ReportAccesses, grantedReportIDs) ||
		requestsNewGrant(req.CompliancePortalFileAccesses, grantedFileIDs)
}

func (s *Service) resolveAccessIdentity(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	compliancePortal *coredata.CompliancePortal,
	req *CreateAccessRequest,
	now time.Time,
) (*coredata.Identity, *coredata.MembershipProfile, error) {
	if req.ProfileID != nil {
		profile := &coredata.MembershipProfile{}
		if err := profile.LoadByID(ctx, tx, scope, *req.ProfileID); err != nil {
			return nil, nil, fmt.Errorf("cannot load profile: %w", err)
		}

		if profile.OrganizationID != compliancePortal.OrganizationID {
			return nil, nil, coredata.ErrResourceNotFound
		}

		identity := &coredata.Identity{}
		if err := identity.LoadByID(ctx, tx, profile.IdentityID); err != nil {
			return nil, nil, fmt.Errorf("cannot load identity: %w", err)
		}

		return identity, profile, nil
	}

	identity, err := findOrCreateIdentity(ctx, tx, *req.Email, now)
	if err != nil {
		return nil, nil, err
	}

	profile, err := findOrCreateVisitorProfile(
		ctx,
		tx,
		scope,
		identity,
		compliancePortal.OrganizationID,
		now,
	)
	if err != nil {
		return nil, nil, err
	}

	return identity, profile, nil
}

func findOrCreateIdentity(
	ctx context.Context,
	tx pg.Tx,
	email mail.Addr,
	now time.Time,
) (*coredata.Identity, error) {
	identity := &coredata.Identity{}

	err := identity.LoadByEmail(ctx, tx, email)
	if err == nil {
		return identity, nil
	}

	if !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load identity: %w", err)
	}

	identity = &coredata.Identity{
		ID:           gid.New(gid.NilTenant, coredata.IdentityEntityType),
		EmailAddress: email,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	insertErr := tx.Savepoint(
		ctx,
		func(ctx context.Context, sp pg.Tx) error {
			return identity.Insert(ctx, sp)
		},
	)
	if insertErr != nil {
		if errors.Is(insertErr, coredata.ErrResourceAlreadyExists) {
			if err := identity.LoadByEmail(ctx, tx, email); err != nil {
				return nil, fmt.Errorf("cannot load identity after conflict: %w", err)
			}

			return identity, nil
		}

		return nil, fmt.Errorf("cannot insert identity: %w", insertErr)
	}

	return identity, nil
}

func findOrCreateVisitorProfile(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	identity *coredata.Identity,
	organizationID gid.GID,
	now time.Time,
) (*coredata.MembershipProfile, error) {
	profile := &coredata.MembershipProfile{}

	err := profile.LoadByIdentityIDAndOrganizationID(
		ctx,
		tx,
		scope,
		identity.ID,
		organizationID,
	)
	if err == nil {
		return profile, nil
	}

	if !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load profile: %w", err)
	}

	profile = &coredata.MembershipProfile{
		ID:             gid.New(scope.GetTenantID(), coredata.MembershipProfileEntityType),
		IdentityID:     identity.ID,
		OrganizationID: organizationID,
		EmailAddress:   identity.EmailAddress,
		Source:         coredata.ProfileSourceManual,
		State:          coredata.ProfileStateActive,
		ActivatedAt:    &now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	insertErr := tx.Savepoint(
		ctx,
		func(ctx context.Context, sp pg.Tx) error {
			return profile.Insert(ctx, sp)
		},
	)
	if insertErr != nil {
		if errors.Is(insertErr, coredata.ErrResourceAlreadyExists) {
			if err := profile.LoadByIdentityIDAndOrganizationID(
				ctx,
				tx,
				scope,
				identity.ID,
				organizationID,
			); err != nil {
				return nil, fmt.Errorf("cannot load profile after conflict: %w", err)
			}

			return profile, nil
		}

		return nil, fmt.Errorf("cannot insert profile: %w", insertErr)
	}

	return profile, nil
}

func (s *Service) grantCreatedAccessTargets(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	access *coredata.CompliancePortalAccess,
	req *CreateAccessRequest,
) error {
	var tcdas coredata.CompliancePortalDocumentAccesses

	if len(req.DocumentIDs) > 0 {
		documentData := make([]coredata.UpsertCompliancePortalDocumentAccessesData, 0, len(req.DocumentIDs))
		for _, documentID := range req.DocumentIDs {
			documentData = append(documentData, coredata.UpsertCompliancePortalDocumentAccessesData{
				ID:     documentID,
				Status: coredata.CompliancePortalDocumentAccessStatusGranted,
			})
		}

		if err := validatePortalAccessTargets(
			ctx,
			tx,
			scope,
			access.CompliancePortalID,
			req.DocumentIDs,
			nil,
			nil,
		); err != nil {
			return err
		}

		if err := tcdas.UpsertDocumentAccesses(
			ctx,
			tx,
			scope,
			access.OrganizationID,
			access.ID,
			documentData,
		); err != nil {
			return fmt.Errorf("cannot upsert document accesses: %w", err)
		}
	}

	if len(req.ReportFileIDs) > 0 {
		reportData := make([]coredata.UpsertCompliancePortalDocumentAccessesData, 0, len(req.ReportFileIDs))
		for _, reportFileID := range req.ReportFileIDs {
			reportData = append(reportData, coredata.UpsertCompliancePortalDocumentAccessesData{
				ID:     reportFileID,
				Status: coredata.CompliancePortalDocumentAccessStatusGranted,
			})
		}

		if err := validatePortalAccessTargets(
			ctx,
			tx,
			scope,
			access.CompliancePortalID,
			nil,
			req.ReportFileIDs,
			nil,
		); err != nil {
			return err
		}

		if err := tcdas.UpsertReportFileAccesses(
			ctx,
			tx,
			scope,
			access.OrganizationID,
			access.ID,
			reportData,
		); err != nil {
			return fmt.Errorf("cannot upsert report accesses: %w", err)
		}
	}

	if len(req.CompliancePortalFileIDs) > 0 {
		fileData := make([]coredata.UpsertCompliancePortalDocumentAccessesData, 0, len(req.CompliancePortalFileIDs))
		for _, fileID := range req.CompliancePortalFileIDs {
			fileData = append(fileData, coredata.UpsertCompliancePortalDocumentAccessesData{
				ID:     fileID,
				Status: coredata.CompliancePortalDocumentAccessStatusGranted,
			})
		}

		if err := validatePortalAccessTargets(
			ctx,
			tx,
			scope,
			access.CompliancePortalID,
			nil,
			nil,
			req.CompliancePortalFileIDs,
		); err != nil {
			return err
		}

		if err := tcdas.UpsertCompliancePortalFileAccesses(
			ctx,
			tx,
			scope,
			access.OrganizationID,
			access.ID,
			fileData,
		); err != nil {
			return fmt.Errorf("cannot upsert compliance page file accesses: %w", err)
		}
	}

	return nil
}
