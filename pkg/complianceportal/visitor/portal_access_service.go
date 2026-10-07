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

package visitor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/bot"
	portal "go.probo.inc/probo/pkg/complianceportal"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type PortalAccessRequest struct {
	CompliancePortalID          gid.GID
	IdentityID                  gid.GID
	CompliancePortalDocumentIDs []gid.GID
	CompliancePortalAuditIDs    []gid.GID
	CompliancePortalFileIDs     []gid.GID
}

func accessMessageParams(
	organizationID gid.GID,
	accessID gid.GID,
	eventKey string,
	purpose coredata.BotMessagePurpose,
) bot.MessageParams {
	return bot.MessageParams{
		OrganizationID: organizationID,
		Capability:     portal.AccessCapability,
		MessageType:    portal.AccessMessageType,
		Attributes: map[string]any{
			portal.AccessIDAttribute: accessID.String(),
		},
		SubjectNamespace: portal.AccessSubjectNamespace,
		SubjectKey:       accessID.String(),
		EventKey:         eventKey,
		Purpose:          purpose,
	}
}

func accessMutationEventKey(
	action string,
	operationKey string,
	compliancePortalDocumentIDs []gid.GID,
	compliancePortalAuditIDs []gid.GID,
	compliancePortalFileIDs []gid.GID,
) string {
	if operationKey != "" {
		return action + ":" + operationKey
	}

	components := make([]string, 0, len(compliancePortalDocumentIDs)+len(compliancePortalAuditIDs)+len(compliancePortalFileIDs))
	for _, id := range compliancePortalDocumentIDs {
		components = append(components, "document:"+id.String())
	}

	for _, id := range compliancePortalAuditIDs {
		components = append(components, "report:"+id.String())
	}

	for _, id := range compliancePortalFileIDs {
		components = append(components, "file:"+id.String())
	}

	return bot.StableEventKey(action, components...)
}

func requestPortalAccessBotEnqueue(
	existingCompliancePortalDocumentIDs []gid.GID,
	existingCompliancePortalAuditIDs []gid.GID,
	existingCompliancePortalFileIDs []gid.GID,
	newCompliancePortalDocumentIDs []gid.GID,
	newCompliancePortalAuditIDs []gid.GID,
	newCompliancePortalFileIDs []gid.GID,
) (eventKey string, purpose coredata.BotMessagePurpose, enqueue bool) {
	if len(newCompliancePortalDocumentIDs) == 0 &&
		len(newCompliancePortalAuditIDs) == 0 &&
		len(newCompliancePortalFileIDs) == 0 {
		return "", "", false
	}

	purpose = coredata.BotMessagePurposePost
	if len(existingCompliancePortalDocumentIDs) > 0 ||
		len(existingCompliancePortalAuditIDs) > 0 ||
		len(existingCompliancePortalFileIDs) > 0 {
		purpose = coredata.BotMessagePurposeUpdate
	}

	return accessMutationEventKey(
		"request",
		"",
		newCompliancePortalDocumentIDs,
		newCompliancePortalAuditIDs,
		newCompliancePortalFileIDs,
	), purpose, true
}

const (
	PortalAccessURLFormat = "https://%s/organizations/%s/compliance-portals/%s/visitors/%s"
)

func (s *Service) RequestPortalAccess(
	ctx context.Context,
	scope coredata.Scoper,
	req *PortalAccessRequest,
) (*coredata.CompliancePortalAccess, error) {
	if len(req.CompliancePortalDocumentIDs) == 0 &&
		len(req.CompliancePortalAuditIDs) == 0 &&
		len(req.CompliancePortalFileIDs) == 0 {
		return nil, ErrNoAccessTargets
	}

	var (
		now    = time.Now()
		access *coredata.CompliancePortalAccess
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			compliancePage := &coredata.CompliancePortal{}
			if err := compliancePage.LoadByID(ctx, tx, scope, req.CompliancePortalID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			access = &coredata.CompliancePortalAccess{}
			if err := access.LoadByCompliancePortalIDAndIdentityID(ctx, tx, scope, req.CompliancePortalID, req.IdentityID); err != nil {
				return fmt.Errorf("cannot load compliance page membership: %w", err)
			}

			if access.State != coredata.CompliancePortalAccessStateActive {
				return ErrUserInactive
			}

			existingAccesses, err := page.LoadAll(
				ctx,
				page.OrderBy[coredata.CompliancePortalDocumentAccessOrderField]{
					Field:     coredata.CompliancePortalDocumentAccessOrderFieldCreatedAt,
					Direction: page.OrderDirectionAsc,
				},
				func(ctx context.Context, cursor *page.Cursor[coredata.CompliancePortalDocumentAccessOrderField]) ([]*coredata.CompliancePortalDocumentAccess, error) {
					var batch coredata.CompliancePortalDocumentAccesses
					if err := batch.LoadByCompliancePortalAccessID(ctx, tx, scope, access.ID, cursor); err != nil {
						return nil, fmt.Errorf("cannot load existing access records: %w", err)
					}

					return batch, nil
				},
			)
			if err != nil {
				return err
			}

			existingCompliancePortalDocumentIDs, existingCompliancePortalAuditIDs, existingCompliancePortalFileIDs := extractExistingIDs(existingAccesses)
			newCompliancePortalDocumentIDs := filterExistingIDs(
				req.CompliancePortalDocumentIDs,
				existingCompliancePortalDocumentIDs,
			)
			newCompliancePortalAuditIDs := filterExistingIDs(
				req.CompliancePortalAuditIDs,
				existingCompliancePortalAuditIDs,
			)
			newCompliancePortalFileIDs := filterExistingIDs(req.CompliancePortalFileIDs, existingCompliancePortalFileIDs)
			// IDs that already have a row: REJECTED/REVOKED retries need a status
			// reset and a requested_at stamp (BulkInsert skips them via ON CONFLICT).
			rerequestCompliancePortalDocumentIDs := filterPresentIDs(
				req.CompliancePortalDocumentIDs,
				existingCompliancePortalDocumentIDs,
			)
			rerequestCompliancePortalAuditIDs := filterPresentIDs(
				req.CompliancePortalAuditIDs,
				existingCompliancePortalAuditIDs,
			)
			rerequestCompliancePortalFileIDs := filterPresentIDs(
				req.CompliancePortalFileIDs,
				existingCompliancePortalFileIDs,
			)

			var accesses coredata.CompliancePortalDocumentAccesses

			if err := accesses.BulkInsertDocumentAccesses(
				ctx,
				tx,
				scope,
				access.ID,
				access.OrganizationID,
				newCompliancePortalDocumentIDs,
				coredata.CompliancePortalDocumentAccessStatusRequested,
				now,
			); err != nil {
				return fmt.Errorf("cannot bulk insert compliance page document accesses: %w", err)
			}

			if err := accesses.BulkInsertCompliancePortalAuditAccesses(
				ctx,
				tx,
				scope,
				access.ID,
				access.OrganizationID,
				newCompliancePortalAuditIDs,
				coredata.CompliancePortalDocumentAccessStatusRequested,
				now,
			); err != nil {
				return fmt.Errorf("cannot bulk insert compliance page report accesses: %w", err)
			}

			if err := accesses.BulkInsertCompliancePortalFileAccesses(
				ctx,
				tx,
				scope,
				access.ID,
				access.OrganizationID,
				newCompliancePortalFileIDs,
				coredata.CompliancePortalDocumentAccessStatusRequested,
				now,
			); err != nil {
				return fmt.Errorf("cannot bulk insert compliance page file accesses: %w", err)
			}

			if err := coredata.RerequestByCompliancePortalDocumentIDs(
				ctx,
				tx,
				scope,
				access.ID,
				rerequestCompliancePortalDocumentIDs,
				now,
			); err != nil {
				return fmt.Errorf("cannot rerequest compliance page document accesses: %w", err)
			}

			if err := coredata.RerequestByCompliancePortalAuditIDs(
				ctx,
				tx,
				scope,
				access.ID,
				rerequestCompliancePortalAuditIDs,
				now,
			); err != nil {
				return fmt.Errorf("cannot rerequest compliance page report accesses: %w", err)
			}

			if err := coredata.RerequestByCompliancePortalFileIDs(
				ctx,
				tx,
				scope,
				access.ID,
				rerequestCompliancePortalFileIDs,
				now,
			); err != nil {
				return fmt.Errorf("cannot rerequest compliance page file accesses: %w", err)
			}

			eventKey, purpose, enqueue := requestPortalAccessBotEnqueue(
				existingCompliancePortalDocumentIDs,
				existingCompliancePortalAuditIDs,
				existingCompliancePortalFileIDs,
				newCompliancePortalDocumentIDs,
				newCompliancePortalAuditIDs,
				newCompliancePortalFileIDs,
			)
			if enqueue {
				if _, err := s.bot.EnqueueMessage(
					ctx,
					tx,
					scope,
					accessMessageParams(
						access.OrganizationID,
						access.ID,
						eventKey,
						purpose,
					),
				); err != nil {
					return fmt.Errorf("cannot enqueue compliance portal bot message: %w", err)
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

func (s *Service) GetPortalAccess(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	identityID gid.GID,
) (coredata.CompliancePortalAccess, error) {
	var access coredata.CompliancePortalAccess

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return access.LoadByCompliancePortalIDAndIdentityID(ctx, conn, scope, compliancePageID, identityID)
		},
	)

	return access, err
}

func (s *Service) HasRequestedAccess(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	identityID gid.GID,
) (bool, error) {
	var hasRequested bool

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			access := &coredata.CompliancePortalAccess{}
			if err := access.LoadByCompliancePortalIDAndIdentityID(
				ctx,
				conn,
				scope,
				compliancePortalID,
				identityID,
			); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					hasRequested = false
					return nil
				}

				return fmt.Errorf("cannot load compliance portal access: %w", err)
			}

			var documentAccesses coredata.CompliancePortalDocumentAccesses

			count, err := documentAccesses.CountVisitorRequestedByCompliancePortalAccessID(
				ctx,
				conn,
				scope,
				access.ID,
			)
			if err != nil {
				return fmt.Errorf("cannot count visitor-requested document accesses: %w", err)
			}

			hasRequested = count > 0

			return nil
		},
	)
	if err != nil {
		return false, err
	}

	return hasRequested, nil
}

func (s *Service) GetPortalDocumentAccess(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	identityID gid.GID,
	documentID gid.GID,
) (*coredata.CompliancePortalDocumentAccess, error) {
	var documentAccess *coredata.CompliancePortalDocumentAccess

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			access := &coredata.CompliancePortalAccess{}

			if err := access.LoadByCompliancePortalIDAndIdentityID(ctx, conn, scope, compliancePageID, identityID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrMembershipNotFound
				}

				return fmt.Errorf("cannot load compliance page access: %w", err)
			}

			if access.State != coredata.CompliancePortalAccessStateActive {
				return ErrUserInactive
			}

			link := &coredata.CompliancePortalDocument{}
			if err := link.LoadByCompliancePortalIDAndDocumentID(ctx, conn, scope, compliancePageID, documentID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrDocumentAccessNotFound
				}

				return fmt.Errorf("cannot load portal document: %w", err)
			}

			documentAccess = &coredata.CompliancePortalDocumentAccess{}

			if err := documentAccess.LoadByCompliancePortalAccessIDAndCompliancePortalDocumentID(ctx, conn, scope, access.ID, link.ID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrDocumentAccessNotFound
				}

				return fmt.Errorf("cannot load document access: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return documentAccess, nil
}

func (s *Service) GetPortalReportFileAccess(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	identityID gid.GID,
	reportFileID gid.GID,
) (*coredata.CompliancePortalDocumentAccess, error) {
	var reportAccess *coredata.CompliancePortalDocumentAccess

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			access := &coredata.CompliancePortalAccess{}

			if err := access.LoadByCompliancePortalIDAndIdentityID(ctx, conn, scope, compliancePageID, identityID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrMembershipNotFound
				}

				return fmt.Errorf("cannot load compliance page access: %w", err)
			}

			if access.State != coredata.CompliancePortalAccessStateActive {
				return ErrUserInactive
			}

			link := &coredata.CompliancePortalAudit{}
			if err := link.LoadByCompliancePortalIDAndReportFileID(ctx, conn, scope, compliancePageID, reportFileID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrDocumentAccessNotFound
				}

				return fmt.Errorf("cannot load portal audit: %w", err)
			}

			reportAccess = &coredata.CompliancePortalDocumentAccess{}

			if err := reportAccess.LoadByCompliancePortalAccessIDAndCompliancePortalAuditID(ctx, conn, scope, access.ID, link.ID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrDocumentAccessNotFound
				}

				return fmt.Errorf("cannot load report access: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return reportAccess, nil
}

func (s *Service) GetPortalFileAccess(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	identityID gid.GID,
	compliancePortalFileID gid.GID,
) (*coredata.CompliancePortalDocumentAccess, error) {
	var fileAccess *coredata.CompliancePortalDocumentAccess

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			access := &coredata.CompliancePortalAccess{}

			if err := access.LoadByCompliancePortalIDAndIdentityID(ctx, conn, scope, compliancePageID, identityID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrMembershipNotFound
				}

				return fmt.Errorf("cannot load compliance page access: %w", err)
			}

			if access.State != coredata.CompliancePortalAccessStateActive {
				return ErrUserInactive
			}

			fileAccess = &coredata.CompliancePortalDocumentAccess{}

			if err := fileAccess.LoadByCompliancePortalAccessIDAndCompliancePortalFileID(ctx, conn, scope, access.ID, compliancePortalFileID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrDocumentAccessNotFound
				}

				return fmt.Errorf("cannot load compliance page file access: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return fileAccess, nil
}

func (s *Service) GrantPortalAccessByIDs(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	email mail.Addr,
	compliancePortalDocumentIDs []gid.GID,
	compliancePortalAuditIDs []gid.GID,
	compliancePortalFileIDs []gid.GID,
) error {
	return s.GrantPortalAccessByIDsIdempotently(
		ctx,
		scope,
		compliancePortalID,
		email,
		compliancePortalDocumentIDs,
		compliancePortalAuditIDs,
		compliancePortalFileIDs,
		"",
	)
}

func (s *Service) GrantPortalAccessByIDsIdempotently(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	email mail.Addr,
	compliancePortalDocumentIDs []gid.GID,
	compliancePortalAuditIDs []gid.GID,
	compliancePortalFileIDs []gid.GID,
	operationKey string,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			compliancePage := &coredata.CompliancePortal{}
			if err := loadPortalByID(ctx, tx, scope, compliancePortalID, compliancePage); err != nil {
				return err
			}

			if operationKey != "" {
				receipt := coredata.NewOperationReceipt(
					scope,
					compliancePage.OrganizationID,
					operationKey,
				)

				claimed, err := receipt.Claim(ctx, tx, scope)
				if err != nil {
					return fmt.Errorf("cannot claim portal access operation: %w", err)
				}

				if !claimed {
					return nil
				}
			}

			identity := &coredata.Identity{}
			if err := identity.LoadByEmail(ctx, tx, email); err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			access := &coredata.CompliancePortalAccess{}
			if err := access.LoadByCompliancePortalIDAndIdentityIDForUpdate(
				ctx,
				tx,
				scope,
				compliancePage.ID,
				identity.ID,
			); err != nil {
				return fmt.Errorf("cannot load compliance page access: %w", err)
			}

			if access.State != coredata.CompliancePortalAccessStateActive {
				return ErrUserInactive
			}

			now := time.Now()
			shouldSendEmail := false

			if len(compliancePortalDocumentIDs) > 0 {
				shouldSendEmail = true

				if err := coredata.GrantByCompliancePortalDocumentIDs(
					ctx,
					tx,
					scope,
					access.ID,
					compliancePortalDocumentIDs,
					now,
				); err != nil {
					return fmt.Errorf("cannot grant document accesses: %w", err)
				}
			}

			if len(compliancePortalAuditIDs) > 0 {
				shouldSendEmail = true

				if err := coredata.GrantByCompliancePortalAuditIDs(
					ctx,
					tx,
					scope,
					access.ID,
					compliancePortalAuditIDs,
					now,
				); err != nil {
					return fmt.Errorf("cannot grant report accesses: %w", err)
				}
			}

			if len(compliancePortalFileIDs) > 0 {
				shouldSendEmail = true

				if err := coredata.GrantByCompliancePortalFileIDs(
					ctx,
					tx,
					scope,
					access.ID,
					compliancePortalFileIDs,
					now,
				); err != nil {
					return fmt.Errorf("cannot grant compliance page file accesses: %w", err)
				}
			}

			if shouldSendEmail {
				if err := s.sendPortalAccessEmail(ctx, tx, scope, access, identity); err != nil {
					return fmt.Errorf("cannot send access email: %w", err)
				}

				if _, err := s.bot.EnqueueMessage(
					ctx,
					tx,
					scope,
					accessMessageParams(
						access.OrganizationID,
						access.ID,
						accessMutationEventKey(
							"grant",
							operationKey,
							compliancePortalDocumentIDs,
							compliancePortalAuditIDs,
							compliancePortalFileIDs,
						),
						coredata.BotMessagePurposeUpdate,
					),
				); err != nil {
					return fmt.Errorf("cannot enqueue compliance portal bot message: %w", err)
				}
			}

			return nil
		},
	)
}

func (s *Service) sendPortalAccessEmail(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	access *coredata.CompliancePortalAccess,
	identity *coredata.Identity,
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

	emailPresenterCfg, err := s.GetPortalEmailPresenterConfig(ctx, scope, access.CompliancePortalID)
	if err != nil {
		return fmt.Errorf("cannot get compliance page email presenter config: %w", err)
	}

	emailPresenter := emails.NewPresenterFromConfig(emailPresenterCfg, identity.FullName)

	subject, textBody, htmlBody, err := emailPresenter.RenderCompliancePortalAccess(ctx, organization.Name)
	if err != nil {
		return fmt.Errorf("cannot render compliance page access email: %w", err)
	}

	accessEmail := coredata.NewEmail(
		identity.FullName,
		identity.EmailAddress,
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

func (s *Service) RejectOrRevokePortalAccessByIDs(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	email mail.Addr,
	compliancePortalDocumentIDs []gid.GID,
	compliancePortalAuditIDs []gid.GID,
	compliancePortalFileIDs []gid.GID,
) error {
	return s.RejectOrRevokePortalAccessByIDsIdempotently(
		ctx,
		scope,
		compliancePortalID,
		email,
		compliancePortalDocumentIDs,
		compliancePortalAuditIDs,
		compliancePortalFileIDs,
		"",
	)
}

func (s *Service) RejectOrRevokePortalAccessByIDsIdempotently(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	email mail.Addr,
	compliancePortalDocumentIDs []gid.GID,
	compliancePortalAuditIDs []gid.GID,
	compliancePortalFileIDs []gid.GID,
	operationKey string,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			compliancePage := &coredata.CompliancePortal{}
			if err := loadPortalByID(ctx, tx, scope, compliancePortalID, compliancePage); err != nil {
				return err
			}

			if operationKey != "" {
				receipt := coredata.NewOperationReceipt(
					scope,
					compliancePage.OrganizationID,
					operationKey,
				)

				claimed, err := receipt.Claim(ctx, tx, scope)
				if err != nil {
					return fmt.Errorf("cannot claim portal access operation: %w", err)
				}

				if !claimed {
					return nil
				}
			}

			identity := &coredata.Identity{}
			if err := identity.LoadByEmail(ctx, tx, email); err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			access := &coredata.CompliancePortalAccess{}
			if err := access.LoadByCompliancePortalIDAndIdentityID(ctx, tx, scope, compliancePage.ID, identity.ID); err != nil {
				return fmt.Errorf("cannot load compliance page access: %w", err)
			}

			shouldSendEmail := false
			now := time.Now()

			if len(compliancePortalDocumentIDs) > 0 {
				shouldSendEmail = true

				if err := coredata.RejectOrRevokeByCompliancePortalDocumentIDs(
					ctx,
					tx,
					scope,
					access.ID,
					compliancePortalDocumentIDs,
					now,
				); err != nil {
					return fmt.Errorf("cannot reject/revoke document accesses: %w", err)
				}
			}

			if len(compliancePortalAuditIDs) > 0 {
				shouldSendEmail = true

				if err := coredata.RejectOrRevokeByCompliancePortalAuditIDs(
					ctx,
					tx,
					scope,
					access.ID,
					compliancePortalAuditIDs,
					now,
				); err != nil {
					return fmt.Errorf("cannot reject/revoke report accesses: %w", err)
				}
			}

			if len(compliancePortalFileIDs) > 0 {
				shouldSendEmail = true

				if err := coredata.RejectOrRevokeByCompliancePortalFileIDs(
					ctx,
					tx,
					scope,
					access.ID,
					compliancePortalFileIDs,
					now,
				); err != nil {
					return fmt.Errorf("cannot reject/revoke compliance page file accesses: %w", err)
				}
			}

			if shouldSendEmail {
				if err := s.sendPortalDocumentAccessRejectedEmail(
					ctx,
					tx,
					scope,
					access,
					identity,
					compliancePortalDocumentIDs,
					compliancePortalAuditIDs,
					compliancePortalFileIDs,
				); err != nil {
					return fmt.Errorf("cannot send access email: %w", err)
				}

				if _, err := s.bot.EnqueueMessage(
					ctx,
					tx,
					scope,
					accessMessageParams(
						access.OrganizationID,
						access.ID,
						accessMutationEventKey(
							"reject-or-revoke",
							operationKey,
							compliancePortalDocumentIDs,
							compliancePortalAuditIDs,
							compliancePortalFileIDs,
						),
						coredata.BotMessagePurposeUpdate,
					),
				); err != nil {
					return fmt.Errorf("cannot enqueue compliance portal bot message: %w", err)
				}
			}

			return nil
		},
	)
}

func (s *Service) sendPortalDocumentAccessRejectedEmail(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	access *coredata.CompliancePortalAccess,
	identity *coredata.Identity,
	compliancePortalDocumentIDs []gid.GID,
	compliancePortalAuditIDs []gid.GID,
	compliancePortalFileIDs []gid.GID,
) error {
	organization := &coredata.Organization{}
	if err := organization.LoadByID(ctx, tx, scope, access.OrganizationID); err != nil {
		return fmt.Errorf("cannot load organization: %w", err)
	}

	var fileNames []string

	for _, documentLinkID := range compliancePortalDocumentIDs {
		var link coredata.CompliancePortalDocument
		if err := link.LoadByID(ctx, tx, scope, documentLinkID); err != nil {
			return fmt.Errorf("cannot load compliance portal document: %w", err)
		}

		var document coredata.Document
		if err := document.LoadByID(ctx, tx, scope, link.DocumentID); err != nil {
			return fmt.Errorf("cannot load document: %w", err)
		}

		fileNames = append(fileNames, document.Title)
	}

	if len(compliancePortalAuditIDs) > 0 {
		reportLabels, err := reportAccessLabels(ctx, tx, scope, compliancePortalAuditIDs)
		if err != nil {
			return fmt.Errorf("cannot build report access labels: %w", err)
		}

		fileNames = append(fileNames, reportLabels...)
	}

	var files coredata.CompliancePortalFiles
	if len(compliancePortalFileIDs) > 0 {
		if err := files.LoadByIDs(ctx, tx, scope, compliancePortalFileIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
			return fmt.Errorf("cannot load files by IDs: %w", err)
		}

		for _, f := range files {
			fileNames = append(fileNames, f.Name)
		}
	}

	emailPresenterCfg, err := s.GetPortalEmailPresenterConfig(ctx, scope, access.CompliancePortalID)
	if err != nil {
		return fmt.Errorf("cannot get compliance page email presenter config: %w", err)
	}

	emailPresenter := emails.NewPresenterFromConfig(emailPresenterCfg, identity.FullName)

	subject, textBody, htmlBody, err := emailPresenter.RenderCompliancePortalDocumentAccessRejected(
		ctx,
		fileNames,
		organization.Name,
	)
	if err != nil {
		return fmt.Errorf("cannot render compliance page documents access rejected email: %w", err)
	}

	accessEmail := coredata.NewEmail(
		identity.FullName,
		identity.EmailAddress,
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

func extractExistingIDs(accesses coredata.CompliancePortalDocumentAccesses) ([]gid.GID, []gid.GID, []gid.GID) {
	var (
		compliancePortalDocumentIDs []gid.GID
		compliancePortalAuditIDs    []gid.GID
		compliancePortalFileIDs     []gid.GID
	)

	for _, access := range accesses {
		if access.CompliancePortalDocumentID != nil {
			compliancePortalDocumentIDs = append(compliancePortalDocumentIDs, *access.CompliancePortalDocumentID)
		}

		if access.CompliancePortalAuditID != nil {
			compliancePortalAuditIDs = append(compliancePortalAuditIDs, *access.CompliancePortalAuditID)
		}

		if access.CompliancePortalFileID != nil {
			compliancePortalFileIDs = append(compliancePortalFileIDs, *access.CompliancePortalFileID)
		}
	}

	return compliancePortalDocumentIDs, compliancePortalAuditIDs, compliancePortalFileIDs
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

// filterPresentIDs returns the subset of allIDs that already exist.
func filterPresentIDs(allIDs []gid.GID, existingIDs []gid.GID) []gid.GID {
	existingMap := make(map[gid.GID]bool)
	for _, id := range existingIDs {
		existingMap[id] = true
	}

	var presentIDs []gid.GID

	for _, id := range allIDs {
		if existingMap[id] {
			presentIDs = append(presentIDs, id)
		}
	}

	return presentIDs
}

func reportAccessLabels(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	compliancePortalAuditIDs []gid.GID,
) ([]string, error) {
	labels := make([]string, 0, len(compliancePortalAuditIDs))

	for _, portalAuditID := range compliancePortalAuditIDs {
		var portalAudit coredata.CompliancePortalAudit
		if err := portalAudit.LoadByID(ctx, conn, scope, portalAuditID); err != nil {
			return nil, fmt.Errorf("cannot load compliance portal audit: %w", err)
		}

		var audit coredata.Audit
		if err := audit.LoadByID(ctx, conn, scope, portalAudit.AuditID); err != nil {
			return nil, fmt.Errorf("cannot load audit: %w", err)
		}

		var framework coredata.Framework
		if err := framework.LoadByID(ctx, conn, scope, audit.FrameworkID); err != nil {
			return nil, fmt.Errorf("cannot load framework: %w", err)
		}

		if audit.Name != nil && *audit.Name != "" {
			labels = append(labels, fmt.Sprintf("%s - %s", framework.Name, *audit.Name))
			continue
		}

		labels = append(labels, framework.Name)
	}

	return labels, nil
}
