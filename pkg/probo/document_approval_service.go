// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package probo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

const DocumentApprovalConsentText = "By clicking \"Review and approve\", I consent to approve this document electronically and agree that my electronic signature has the same legal validity as a handwritten signature."

type (
	DocumentApprovalService struct {
		svc                     *Service
		html2pdfConverter       *html2pdf.Converter
		invitationTokenValidity time.Duration
		tokenSecret             string
	}

	ErrDocumentVersionNotPendingApproval struct{}

	ErrApprovalDecisionAlreadyMade struct{}

	ApproveDocumentVersionRequest struct {
		DocumentVersionID gid.GID
		IdentityID        gid.GID
		Comment           *string
		SignerFullName    string
		SignerEmail       mail.Addr
		SignerIPAddr      string
		SignerUA          string
	}

	RejectDocumentVersionRequest struct {
		DocumentVersionID gid.GID
		IdentityID        gid.GID
		Comment           *string
	}
)

func (e ErrDocumentVersionNotPendingApproval) Error() string {
	return "document version is not pending approval"
}
func (e ErrApprovalDecisionAlreadyMade) Error() string {
	return "approval decision has already been made"
}

func (s *DocumentApprovalService) RequestApprovalInTx(
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	document *coredata.Document,
	documentVersion *coredata.DocumentVersion,
	approverIDs []gid.GID,
	changelog *string,
) (*coredata.DocumentVersionApprovalQuorum, error) {
	now := time.Now()

	documentVersion.Status = coredata.DocumentVersionStatusPendingApproval
	if changelog != nil {
		documentVersion.Changelog = *changelog
	}

	if document.CurrentPublishedMajor != nil {
		documentVersion.Major = *document.CurrentPublishedMajor + 1
	} else {
		documentVersion.Major = 1
	}

	documentVersion.Minor = 0

	documentVersion.UpdatedAt = now
	if err := documentVersion.Update(ctx, tx, scope); err != nil {
		return nil, fmt.Errorf("cannot update document version: %w", err)
	}

	quorum := &coredata.DocumentVersionApprovalQuorum{
		ID:             gid.New(scope.GetTenantID(), coredata.DocumentVersionApprovalQuorumEntityType),
		OrganizationID: document.OrganizationID,
		VersionID:      documentVersion.ID,
		Status:         coredata.DocumentVersionApprovalQuorumStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := quorum.Insert(ctx, tx, scope); err != nil {
		return nil, fmt.Errorf("cannot insert approval quorum: %w", err)
	}

	if err := s.createDecisions(ctx, scope, tx, quorum, document.OrganizationID, approverIDs, now); err != nil {
		return nil, fmt.Errorf("cannot create approval decisions: %w", err)
	}

	// Approval notifications are sent asynchronously and debounced by the
	// document notification worker, which batches all pending approvals per
	// recipient into a single email.

	return quorum, nil
}

// BulkPublishVersions publishes (or requests approval for) the latest draft of
// each document. When req.Minor is true each draft is published as a minor
// bump and approvers are not consulted. When req.Minor is false, each
// document's saved default approvers are honoured: if the document has any,
// an approval is requested for it; otherwise it is published as a major
// bump. Documents with no draft (or already pending approval) are skipped.
func (s *DocumentApprovalService) BulkPublishVersions(
	ctx context.Context, scope coredata.Scoper,
	req BulkPublishVersionsRequest,
) ([]*coredata.DocumentVersion, []*coredata.Document, error) {
	var (
		publishedVersions []*coredata.DocumentVersion
		updatedDocuments  []*coredata.Document
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			for _, documentID := range req.DocumentIDs {
				dv := &coredata.DocumentVersion{}
				if err := dv.LoadLatestVersion(ctx, tx, scope, documentID); err != nil {
					return fmt.Errorf("cannot load latest version for %q: %w", documentID, err)
				}

				if dv.Status == coredata.DocumentVersionStatusPendingApproval {
					continue
				}

				document := &coredata.Document{}
				if err := document.LoadByID(ctx, tx, scope, documentID); err != nil {
					return fmt.Errorf("cannot load document %q: %w", documentID, err)
				}

				if document.ArchivedAt != nil {
					return &ErrDocumentArchived{}
				}

				// Treat minor on an already-published version as a no-op so the
				// operation is idempotent: the doc is included in the result
				// without modification.
				if req.Minor && dv.Status == coredata.DocumentVersionStatusPublished {
					publishedVersions = append(publishedVersions, dv)
					updatedDocuments = append(updatedDocuments, document)

					continue
				}

				if dv.Status != coredata.DocumentVersionStatusDraft {
					continue
				}

				var requestedQuorum *coredata.DocumentVersionApprovalQuorum

				if req.Minor {
					var err error

					document, dv, err = s.svc.Documents.publishMinorVersionInTx(ctx, scope, tx, documentID, &req.Changelog, true)
					if err != nil {
						return fmt.Errorf("cannot publish document %q: %w", documentID, err)
					}
				} else {
					defaultApprovers := &coredata.DocumentDefaultApprovers{}
					if err := defaultApprovers.LoadByDocumentID(ctx, tx, scope, documentID); err != nil {
						return fmt.Errorf("cannot load default approvers for %q: %w", documentID, err)
					}

					if len(*defaultApprovers) > 0 {
						approverIDs := make([]gid.GID, len(*defaultApprovers))
						for i, a := range *defaultApprovers {
							approverIDs[i] = a.ApproverProfileID
						}

						quorum, err := s.RequestApprovalInTx(ctx, scope, tx, document, dv, approverIDs, &req.Changelog)
						if err != nil {
							return fmt.Errorf("cannot request approval for %q: %w", documentID, err)
						}

						requestedQuorum = quorum
					} else {
						var err error

						document, dv, err = s.svc.Documents.publishMajorVersionInTx(ctx, scope, tx, documentID, &req.Changelog, true)
						if err != nil {
							return fmt.Errorf("cannot publish document %q: %w", documentID, err)
						}
					}
				}

				publishedVersions = append(publishedVersions, dv)
				updatedDocuments = append(updatedDocuments, document)

				if requestedQuorum != nil {
					if err := s.svc.Documents.emitDocumentEventInTx(
						ctx,
						scope,
						tx,
						dv.DocumentID,
						coredata.WebhookEventTypeDocumentVersionApprovalQuorumRequested,
						dv,
						nil,
						&requestedQuorum.ID,
					); err != nil {
						return fmt.Errorf("cannot emit approval quorum requested webhook: %w", err)
					}
				} else {
					if err := s.svc.Documents.emitDocumentEventInTx(
						ctx,
						scope,
						tx,
						dv.DocumentID,
						coredata.WebhookEventTypeDocumentVersionPublished,
						dv,
						nil,
						nil,
					); err != nil {
						return fmt.Errorf("cannot emit version published webhook: %w", err)
					}
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return publishedVersions, updatedDocuments, nil
}

func (s *DocumentApprovalService) Approve(
	ctx context.Context, scope coredata.Scoper,
	req ApproveDocumentVersionRequest,
) (*coredata.DocumentVersionApprovalDecision, error) {
	var (
		documentVersion *coredata.DocumentVersion
		document        *coredata.Document
		quorum          *coredata.DocumentVersionApprovalQuorum
		decision        *coredata.DocumentVersionApprovalDecision
	)

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			documentVersion = &coredata.DocumentVersion{}
			if err := documentVersion.LoadByID(ctx, conn, scope, req.DocumentVersionID); err != nil {
				return fmt.Errorf("cannot load document version: %w", err)
			}

			document = &coredata.Document{}
			if err := document.LoadByID(ctx, conn, scope, documentVersion.DocumentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			var (
				profile *coredata.MembershipProfile
				err     error
			)

			quorum, profile, err = s.loadQuorumAndProfile(ctx, scope, conn, req.DocumentVersionID, req.IdentityID, documentVersion.OrganizationID)
			if err != nil {
				return fmt.Errorf("cannot load quorum and profile: %w", err)
			}

			if quorum.Status != coredata.DocumentVersionApprovalQuorumStatusPending {
				return &ErrDocumentVersionNotPendingApproval{}
			}

			decision = &coredata.DocumentVersionApprovalDecision{}
			if err := decision.LoadByQuorumIDAndApproverID(ctx, conn, scope, quorum.ID, profile.ID); err != nil {
				return fmt.Errorf("cannot load approval decision: %w", err)
			}

			if decision.State != coredata.DocumentVersionApprovalDecisionStatePending {
				return &ErrApprovalDecisionAlreadyMade{}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	pdfData, err := s.generateApprovalPDF(ctx, scope, req.DocumentVersionID)
	if err != nil {
		return nil, fmt.Errorf("cannot export document PDF: %w", err)
	}

	fileRecord := &coredata.File{
		ID:             gid.New(scope.GetTenantID(), coredata.FileEntityType),
		OrganizationID: documentVersion.OrganizationID,
		BucketName:     s.svc.bucket,
		MimeType:       "application/pdf",
		FileName:       fmt.Sprintf("approval-%s.pdf", decision.ID),
		FileKey:        uuid.MustNewV4().String(),
		Visibility:     coredata.FileVisibilityPrivate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	fileSize, err := s.svc.fileManager.PutFile(
		ctx,
		fileRecord,
		bytes.NewReader(pdfData),
		map[string]string{
			"type":        "approval-document",
			"decision-id": decision.ID.String(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot upload approval PDF: %w", err)
	}

	fileRecord.FileSize = fileSize

	approverID := decision.ApproverID

	quorumID := quorum.ID

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			quorum = &coredata.DocumentVersionApprovalQuorum{}
			if err := quorum.LoadByID(ctx, tx, scope, quorumID); err != nil {
				return fmt.Errorf("cannot load quorum: %w", err)
			}

			if quorum.Status != coredata.DocumentVersionApprovalQuorumStatusPending {
				return &ErrDocumentVersionNotPendingApproval{}
			}

			decision = &coredata.DocumentVersionApprovalDecision{}
			if err := decision.LoadByQuorumIDAndApproverID(ctx, tx, scope, quorum.ID, approverID); err != nil {
				return fmt.Errorf("cannot load approval decision: %w", err)
			}

			if decision.State != coredata.DocumentVersionApprovalDecisionStatePending {
				return &ErrApprovalDecisionAlreadyMade{}
			}

			if err := fileRecord.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert approval file record: %w", err)
			}

			esig, err := s.svc.esign.CreateAndAcceptSignature(
				ctx,
				tx,
				&esign.CreateAndAcceptSignatureRequest{
					OrganizationID: documentVersion.OrganizationID,
					DocumentType:   coredata.ElectronicSignatureDocumentTypeFromDocumentType(documentVersion.DocumentType),
					DocumentName:   &document.Title,
					FileID:         fileRecord.ID,
					SignerEmail:    req.SignerEmail,
					SignerFullName: req.SignerFullName,
					SignerIPAddr:   req.SignerIPAddr,
					SignerUA:       req.SignerUA,
					ConsentText:    DocumentApprovalConsentText,
					EmailSubject:   fmt.Sprintf("Your approved %s - Certificate of Completion", document.Title),
				},
			)
			if err != nil {
				return fmt.Errorf("cannot create electronic signature: %w", err)
			}

			decision.State = coredata.DocumentVersionApprovalDecisionStateApproved
			decision.Comment = req.Comment
			decision.ElectronicSignatureID = &esig.ID
			decision.DecidedAt = &now
			decision.UpdatedAt = now

			if err := decision.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update approval decision: %w", err)
			}

			if err := s.maybeApproveQuorum(ctx, scope, tx, quorum.ID); err != nil {
				return fmt.Errorf("cannot check quorum approval: %w", err)
			}

			if err := documentVersion.LoadByID(ctx, tx, scope, req.DocumentVersionID); err != nil {
				return fmt.Errorf("cannot reload document version: %w", err)
			}

			if err := quorum.LoadByID(ctx, tx, scope, quorum.ID); err != nil {
				return fmt.Errorf("cannot reload approval quorum: %w", err)
			}

			if quorum.Status == coredata.DocumentVersionApprovalQuorumStatusApproved {
				return nil
			}

			if err := s.svc.Documents.emitDocumentEventInTx(
				ctx,
				scope,
				tx,
				documentVersion.DocumentID,
				coredata.WebhookEventTypeDocumentVersionApprovalQuorumUpdated,
				documentVersion,
				nil,
				&quorum.ID,
			); err != nil {
				return fmt.Errorf("cannot emit approval quorum updated webhook: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return decision, nil
}

func (s *DocumentApprovalService) Reject(
	ctx context.Context, scope coredata.Scoper,
	req RejectDocumentVersionRequest,
) (*coredata.DocumentVersionApprovalDecision, error) {
	var decision *coredata.DocumentVersionApprovalDecision

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			documentVersion := &coredata.DocumentVersion{}
			if err := documentVersion.LoadByID(ctx, tx, scope, req.DocumentVersionID); err != nil {
				return fmt.Errorf("cannot load document version: %w", err)
			}

			document := &coredata.Document{}
			if err := document.LoadByID(ctx, tx, scope, documentVersion.DocumentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			quorum, profile, err := s.loadQuorumAndProfile(ctx, scope, tx, req.DocumentVersionID, req.IdentityID, documentVersion.OrganizationID)
			if err != nil {
				return fmt.Errorf("cannot load quorum and profile: %w", err)
			}

			decision = &coredata.DocumentVersionApprovalDecision{}
			if err := decision.LoadByQuorumIDAndApproverID(ctx, tx, scope, quorum.ID, profile.ID); err != nil {
				return fmt.Errorf("cannot load approval decision: %w", err)
			}

			if decision.State != coredata.DocumentVersionApprovalDecisionStatePending {
				return &ErrApprovalDecisionAlreadyMade{}
			}

			now := time.Now()

			decision.State = coredata.DocumentVersionApprovalDecisionStateRejected
			decision.Comment = req.Comment
			decision.DecidedAt = &now
			decision.UpdatedAt = now

			if err := decision.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update approval decision: %w", err)
			}

			quorum.Status = coredata.DocumentVersionApprovalQuorumStatusRejected
			quorum.UpdatedAt = now

			if err := quorum.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update approval quorum: %w", err)
			}

			decisions := &coredata.DocumentVersionApprovalDecisions{}
			if err := decisions.VoidPendingByQuorumID(ctx, tx, scope, quorum.ID, now); err != nil {
				return fmt.Errorf("cannot void pending decisions: %w", err)
			}

			documentVersion.Status = coredata.DocumentVersionStatusDraft
			if document.CurrentPublishedMajor != nil {
				documentVersion.Major = *document.CurrentPublishedMajor
				documentVersion.Minor = *document.CurrentPublishedMinor + 1
			} else {
				documentVersion.Major = 0
				documentVersion.Minor = 1
			}

			documentVersion.UpdatedAt = now

			if err := documentVersion.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update document version status: %w", err)
			}

			if err := s.svc.Documents.emitDocumentEventInTx(
				ctx,
				scope,
				tx,
				documentVersion.DocumentID,
				coredata.WebhookEventTypeDocumentVersionApprovalQuorumRejected,
				documentVersion,
				nil,
				&quorum.ID,
			); err != nil {
				return fmt.Errorf("cannot emit approval quorum rejected webhook: %w", err)
			}

			if err := s.svc.Documents.emitDocumentEventInTx(
				ctx,
				scope,
				tx,
				documentVersion.DocumentID,
				coredata.WebhookEventTypeDocumentVersionRejected,
				documentVersion,
				nil,
				nil,
			); err != nil {
				return fmt.Errorf("cannot emit document version rejected webhook: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return decision, nil
}

func (s *DocumentApprovalService) VoidApproval(
	ctx context.Context, scope coredata.Scoper,
	documentVersionID gid.GID,
) (*coredata.DocumentVersionApprovalQuorum, *coredata.DocumentVersion, error) {
	var (
		quorum          *coredata.DocumentVersionApprovalQuorum
		documentVersion *coredata.DocumentVersion
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			documentVersion = &coredata.DocumentVersion{}
			if err := documentVersion.LoadByID(ctx, tx, scope, documentVersionID); err != nil {
				return fmt.Errorf("cannot load document version: %w", err)
			}

			document := &coredata.Document{}
			if err := document.LoadByID(ctx, tx, scope, documentVersion.DocumentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			if documentVersion.Status != coredata.DocumentVersionStatusPendingApproval {
				return &ErrDocumentVersionNotPendingApproval{}
			}

			quorum = &coredata.DocumentVersionApprovalQuorum{}
			if err := quorum.LoadLastByDocumentVersionID(ctx, tx, scope, documentVersionID); err != nil {
				return fmt.Errorf("cannot load approval quorum: %w", err)
			}

			if quorum.Status != coredata.DocumentVersionApprovalQuorumStatusPending {
				return &ErrDocumentVersionNotPendingApproval{}
			}

			now := time.Now()

			quorum.Status = coredata.DocumentVersionApprovalQuorumStatusVoided
			quorum.UpdatedAt = now

			if err := quorum.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update approval quorum: %w", err)
			}

			decisions := &coredata.DocumentVersionApprovalDecisions{}
			if err := decisions.VoidPendingByQuorumID(ctx, tx, scope, quorum.ID, now); err != nil {
				return fmt.Errorf("cannot void pending decisions: %w", err)
			}

			documentVersion.Status = coredata.DocumentVersionStatusDraft
			if document.CurrentPublishedMajor != nil {
				documentVersion.Major = *document.CurrentPublishedMajor
				documentVersion.Minor = *document.CurrentPublishedMinor + 1
			} else {
				documentVersion.Major = 0
				documentVersion.Minor = 1
			}

			documentVersion.UpdatedAt = now

			if err := documentVersion.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update document version status: %w", err)
			}

			if err := s.svc.Documents.emitDocumentEventInTx(
				ctx,
				scope,
				tx,
				documentVersion.DocumentID,
				coredata.WebhookEventTypeDocumentVersionApprovalQuorumVoided,
				documentVersion,
				nil,
				&quorum.ID,
			); err != nil {
				return fmt.Errorf("cannot emit approval quorum voided webhook: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return quorum, documentVersion, nil
}

func (s *DocumentApprovalService) GetQuorum(
	ctx context.Context, scope coredata.Scoper,
	quorumID gid.GID,
) (*coredata.DocumentVersionApprovalQuorum, error) {
	quorum := &coredata.DocumentVersionApprovalQuorum{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := quorum.LoadByID(ctx, conn, scope, quorumID); err != nil {
				return fmt.Errorf("cannot load approval quorum: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return quorum, nil
}

func (s *DocumentApprovalService) ListQuorums(
	ctx context.Context, scope coredata.Scoper,
	documentVersionID gid.GID,
	cursor *page.Cursor[coredata.DocumentVersionApprovalQuorumOrderField],
) (*page.Page[*coredata.DocumentVersionApprovalQuorum, coredata.DocumentVersionApprovalQuorumOrderField], error) {
	var quorums coredata.DocumentVersionApprovalQuorums

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := quorums.LoadAllByDocumentVersionID(ctx, conn, scope, documentVersionID, cursor); err != nil {
				return fmt.Errorf("cannot list approval quorums: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(quorums, cursor), nil
}

func (s *DocumentApprovalService) CountQuorums(
	ctx context.Context, scope coredata.Scoper,
	documentVersionID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			quorums := &coredata.DocumentVersionApprovalQuorums{}

			count, err = quorums.CountByDocumentVersionID(ctx, conn, scope, documentVersionID)
			if err != nil {
				return fmt.Errorf("cannot count approval quorums: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *DocumentApprovalService) ListDecisions(
	ctx context.Context, scope coredata.Scoper,
	quorumID gid.GID,
	cursor *page.Cursor[coredata.DocumentVersionApprovalDecisionOrderField],
	filter *coredata.DocumentVersionApprovalDecisionFilter,
) (*page.Page[*coredata.DocumentVersionApprovalDecision, coredata.DocumentVersionApprovalDecisionOrderField], error) {
	var decisions coredata.DocumentVersionApprovalDecisions

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := decisions.LoadByQuorumID(ctx, conn, scope, quorumID, cursor, filter); err != nil {
				return fmt.Errorf("cannot list approval decisions: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(decisions, cursor), nil
}

func (s *DocumentApprovalService) CountDecisions(
	ctx context.Context, scope coredata.Scoper,
	quorumID gid.GID,
	filter *coredata.DocumentVersionApprovalDecisionFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			decisions := &coredata.DocumentVersionApprovalDecisions{}

			count, err = decisions.CountByQuorumID(ctx, conn, scope, quorumID, filter)
			if err != nil {
				return fmt.Errorf("cannot count approval decisions: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *DocumentApprovalService) GetDecision(
	ctx context.Context, scope coredata.Scoper,
	decisionID gid.GID,
) (*coredata.DocumentVersionApprovalDecision, error) {
	decision := &coredata.DocumentVersionApprovalDecision{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := decision.LoadByID(ctx, conn, scope, decisionID); err != nil {
				return fmt.Errorf("cannot load approval decision: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return decision, nil
}

func (s *DocumentApprovalService) GetViewerDecision(
	ctx context.Context, scope coredata.Scoper,
	documentVersionID gid.GID,
	identityID gid.GID,
) (*coredata.DocumentVersionApprovalDecision, error) {
	var decision *coredata.DocumentVersionApprovalDecision

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			documentVersion := &coredata.DocumentVersion{}
			if err := documentVersion.LoadByID(ctx, conn, scope, documentVersionID); err != nil {
				return fmt.Errorf("cannot load document version: %w", err)
			}

			profile := &coredata.MembershipProfile{}
			if err := profile.LoadByIdentityIDAndOrganizationID(
				ctx,
				conn,
				scope,
				identityID,
				documentVersion.OrganizationID,
			); err != nil {
				return fmt.Errorf("cannot load viewer profile: %w", err)
			}

			quorum := &coredata.DocumentVersionApprovalQuorum{}
			if err := quorum.LoadLastByDocumentVersionID(ctx, conn, scope, documentVersionID); err != nil {
				return fmt.Errorf("cannot load last approval quorum: %w", err)
			}

			d := &coredata.DocumentVersionApprovalDecision{}
			if err := d.LoadByQuorumIDAndApproverID(ctx, conn, scope, quorum.ID, profile.ID); err != nil {
				return fmt.Errorf("cannot load viewer approval decision: %w", err)
			}

			decision = d

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return decision, nil
}

func (s *DocumentApprovalService) loadQuorumAndProfile(
	ctx context.Context, scope coredata.Scoper,
	conn pg.Querier,
	documentVersionID gid.GID,
	identityID gid.GID,
	organizationID gid.GID,
) (*coredata.DocumentVersionApprovalQuorum, *coredata.MembershipProfile, error) {
	quorum := &coredata.DocumentVersionApprovalQuorum{}
	if err := quorum.LoadLastByDocumentVersionID(ctx, conn, scope, documentVersionID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil, &ErrDocumentVersionNotPendingApproval{}
		}

		return nil, nil, fmt.Errorf("cannot load last approval quorum: %w", err)
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByIdentityIDAndOrganizationID(ctx, conn, scope, identityID, organizationID); err != nil {
		return nil, nil, fmt.Errorf("cannot find profile for identity: %w", err)
	}

	return quorum, profile, nil
}

func (s *DocumentApprovalService) createDecisions(
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	quorum *coredata.DocumentVersionApprovalQuorum,
	organizationID gid.GID,
	approverIDs []gid.GID,
	now time.Time,
) error {
	decisions := make(coredata.DocumentVersionApprovalDecisions, 0, len(approverIDs))
	for _, approverID := range approverIDs {
		decisions = append(decisions, &coredata.DocumentVersionApprovalDecision{
			ID:             gid.New(scope.GetTenantID(), coredata.DocumentVersionApprovalDecisionEntityType),
			OrganizationID: organizationID,
			QuorumID:       quorum.ID,
			ApproverID:     approverID,
			State:          coredata.DocumentVersionApprovalDecisionStatePending,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}

	if err := decisions.BulkInsert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert approval decisions: %w", err)
	}

	return nil
}

func (s *DocumentApprovalService) generateApprovalPDF(
	ctx context.Context, scope coredata.Scoper,
	documentVersionID gid.GID,
) ([]byte, error) {
	var pdfData []byte

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error

			pdfData, err = exportDocumentPDF(
				ctx,
				s.svc,
				s.html2pdfConverter,
				conn,
				scope,
				documentVersionID,
				ExportPDFOptions{},
			)

			return err
		},
	)

	return pdfData, err
}

func (s *DocumentApprovalService) countDecisions(
	ctx context.Context, scope coredata.Scoper,
	conn pg.Querier,
	quorumID gid.GID,
) (int, error) {
	decisions := &coredata.DocumentVersionApprovalDecisions{}

	count, err := decisions.CountByQuorumID(
		ctx,
		conn,
		scope,
		quorumID,
		coredata.NewDocumentVersionApprovalDecisionFilter(nil),
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count decisions: %w", err)
	}

	return count, nil
}

func (s *DocumentApprovalService) maybeApproveQuorum(
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	quorumID gid.GID,
) error {
	totalCount, err := s.countDecisions(ctx, scope, tx, quorumID)
	if err != nil {
		return fmt.Errorf("cannot count total decisions: %w", err)
	}

	if totalCount == 0 {
		return nil
	}

	decisions := &coredata.DocumentVersionApprovalDecisions{}

	approvedCount, err := decisions.CountApprovedByQuorumID(ctx, tx, scope, quorumID)
	if err != nil {
		return fmt.Errorf("cannot count approved decisions: %w", err)
	}

	if approvedCount != totalCount {
		return nil
	}

	quorum := &coredata.DocumentVersionApprovalQuorum{}
	if err := quorum.LoadByID(ctx, tx, scope, quorumID); err != nil {
		return fmt.Errorf("cannot load quorum: %w", err)
	}

	now := time.Now()
	quorum.Status = coredata.DocumentVersionApprovalQuorumStatusApproved
	quorum.UpdatedAt = now

	if err := quorum.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update quorum: %w", err)
	}

	version, err := s.publishVersion(ctx, scope, tx, quorum.VersionID)
	if err != nil {
		return fmt.Errorf("cannot publish version: %w", err)
	}

	if err := s.svc.Documents.emitDocumentEventInTx(
		ctx,
		scope,
		tx,
		version.DocumentID,
		coredata.WebhookEventTypeDocumentVersionApprovalQuorumApproved,
		version,
		nil,
		&quorum.ID,
	); err != nil {
		return fmt.Errorf("cannot emit approval quorum approved webhook: %w", err)
	}

	if err := s.svc.Documents.emitDocumentEventInTx(
		ctx,
		scope,
		tx,
		version.DocumentID,
		coredata.WebhookEventTypeDocumentVersionPublished,
		version,
		nil,
		nil,
	); err != nil {
		return fmt.Errorf("cannot emit document version published webhook: %w", err)
	}

	return nil
}

func (s *DocumentApprovalService) publishVersion(
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	versionID gid.GID,
) (*coredata.DocumentVersion, error) {
	version := &coredata.DocumentVersion{}
	if err := version.LoadByID(ctx, tx, scope, versionID); err != nil {
		return nil, fmt.Errorf("cannot load document version: %w", err)
	}

	document := &coredata.Document{}
	if err := document.LoadByID(ctx, tx, scope, version.DocumentID); err != nil {
		return nil, fmt.Errorf("cannot load document: %w", err)
	}

	document.CurrentPublishedMajor = &version.Major
	document.CurrentPublishedMinor = &version.Minor

	if err := s.svc.Documents.finalizePublish(ctx, scope, tx, document, version, nil); err != nil {
		return nil, fmt.Errorf("cannot finalize publish: %w", err)
	}

	if err := s.svc.Documents.cancelPreviousMajorSignatureRequestsInTx(ctx, scope, tx, version.DocumentID, version.Major); err != nil {
		return nil, fmt.Errorf("cannot cancel signature requests from previous major versions: %w", err)
	}

	return version, nil
}
