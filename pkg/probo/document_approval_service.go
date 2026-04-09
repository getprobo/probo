// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"errors"
	"fmt"
	"net/url"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/statelesstoken"
	"go.probo.inc/probo/pkg/validator"
)

type (
	DocumentApprovalService struct {
		svc                     *TenantService
		html2pdfConverter       *html2pdf.Converter
		invitationTokenValidity time.Duration
		tokenSecret             string
	}

	ErrDocumentVersionNotPendingApproval struct{}

	ErrApprovalDecisionAlreadyMade struct{}

	RequestApprovalRequest struct {
		DocumentID  gid.GID
		ApproverIDs []gid.GID
		Changelog   *string
	}

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

func (req *RequestApprovalRequest) Validate() error {
	v := validator.New()

	v.Check(req.DocumentID, "document_id", validator.Required(), validator.GID(coredata.DocumentEntityType))
	v.Check(len(req.ApproverIDs), "approver_ids", validator.Min(1), validator.Max(100))
	v.Check(req.ApproverIDs, "approver_ids", validator.NoDuplicates())
	v.CheckEach(req.ApproverIDs, "approver_ids", func(index int, item any) {
		v.Check(item, fmt.Sprintf("approver_ids[%d]", index), validator.GID(coredata.MembershipProfileEntityType))
	})
	v.Check(req.Changelog, "changelog", validator.Required(), validator.SafeText(5000))

	return v.Error()
}

func (s *DocumentApprovalService) RequestApproval(
	ctx context.Context,
	req RequestApprovalRequest,
) (*coredata.DocumentVersionApprovalQuorum, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var quorum *coredata.DocumentVersionApprovalQuorum

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			document := &coredata.Document{}
			if err := document.LoadByID(ctx, tx, s.svc.scope, req.DocumentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			documentVersion, err := s.loadLatestVersion(ctx, tx, req.DocumentID)
			if err != nil {
				return fmt.Errorf("cannot load latest version: %w", err)
			}

			if documentVersion.Status != coredata.DocumentVersionStatusDraft {
				return &ErrDocumentVersionNotDraft{}
			}

			q, err := s.requestApprovalInTx(ctx, tx, document, documentVersion, req.ApproverIDs, req.Changelog)
			if err != nil {
				return err
			}
			quorum = q

			defaultApprovers := &coredata.DocumentDefaultApprovers{}
			if err := defaultApprovers.MergeByDocumentID(ctx, tx, s.svc.scope, req.DocumentID, document.OrganizationID, req.ApproverIDs); err != nil {
				return fmt.Errorf("cannot update default approvers: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return quorum, nil
}

func (s *DocumentApprovalService) requestApprovalInTx(
	ctx context.Context,
	tx pg.Tx,
	document *coredata.Document,
	documentVersion *coredata.DocumentVersion,
	approverIDs []gid.GID,
	changelog *string,
) (*coredata.DocumentVersionApprovalQuorum, error) {
	organization := &coredata.Organization{}
	if err := organization.LoadByID(ctx, tx, s.svc.scope, document.OrganizationID); err != nil {
		return nil, fmt.Errorf("cannot load organization: %w", err)
	}

	approverProfiles := &coredata.MembershipProfiles{}
	if err := approverProfiles.LoadByIDs(ctx, tx, s.svc.scope, approverIDs); err != nil {
		return nil, fmt.Errorf("cannot load approver profiles: %w", err)
	}

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
	if err := documentVersion.Update(ctx, tx, s.svc.scope); err != nil {
		return nil, fmt.Errorf("cannot update document version: %w", err)
	}

	quorum := &coredata.DocumentVersionApprovalQuorum{
		ID:             gid.New(s.svc.scope.GetTenantID(), coredata.DocumentVersionApprovalQuorumEntityType),
		OrganizationID: document.OrganizationID,
		VersionID:      documentVersion.ID,
		Status:         coredata.DocumentVersionApprovalQuorumStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := quorum.Insert(ctx, tx, s.svc.scope); err != nil {
		return nil, fmt.Errorf("cannot insert approval quorum: %w", err)
	}

	if err := s.createDecisions(ctx, tx, quorum, document.OrganizationID, approverIDs, now); err != nil {
		return nil, fmt.Errorf("cannot create approval decisions: %w", err)
	}

	if err := s.sendApprovalEmails(ctx, tx, *approverProfiles, document, organization, documentVersion.ID); err != nil {
		return nil, fmt.Errorf("cannot send approval emails: %w", err)
	}

	return quorum, nil
}

func (s *DocumentApprovalService) BulkPublishMajorVersions(
	ctx context.Context,
	req BulkPublishVersionsRequest,
) ([]*coredata.DocumentVersion, []*coredata.Document, error) {
	var publishedVersions []*coredata.DocumentVersion
	var updatedDocuments []*coredata.Document

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			for _, documentID := range req.DocumentIDs {
				dv := &coredata.DocumentVersion{}
				if err := dv.LoadLatestVersion(ctx, tx, s.svc.scope, documentID); err != nil {
					return fmt.Errorf("cannot load latest version for %q: %w", documentID, err)
				}

				// Skip documents already pending approval.
				if dv.Status == coredata.DocumentVersionStatusPendingApproval {
					continue
				}

				document := &coredata.Document{}
				if err := document.LoadByID(ctx, tx, s.svc.scope, documentID); err != nil {
					return fmt.Errorf("cannot load document %q: %w", documentID, err)
				}

				if document.ArchivedAt != nil {
					return &ErrDocumentArchived{}
				}

				defaultApprovers := &coredata.DocumentDefaultApprovers{}
				if err := defaultApprovers.LoadByDocumentID(ctx, tx, s.svc.scope, documentID); err != nil {
					return fmt.Errorf("cannot load default approvers for %q: %w", documentID, err)
				}

				if dv.Status != coredata.DocumentVersionStatusDraft {
					continue
				}

				if len(*defaultApprovers) > 0 {
					approverIDs := make([]gid.GID, len(*defaultApprovers))
					for i, a := range *defaultApprovers {
						approverIDs[i] = a.ApproverProfileID
					}

					if _, err := s.requestApprovalInTx(ctx, tx, document, dv, approverIDs, &req.Changelog); err != nil {
						return fmt.Errorf("cannot request approval for %q: %w", documentID, err)
					}
				} else {
					var err error
					document, dv, err = s.svc.Documents.publishMajorVersionInTx(ctx, tx, documentID, &req.Changelog, true)
					if err != nil {
						return fmt.Errorf("cannot publish document %q: %w", documentID, err)
					}
				}

				publishedVersions = append(publishedVersions, dv)
				updatedDocuments = append(updatedDocuments, document)
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
	ctx context.Context,
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
			if err := documentVersion.LoadByID(ctx, conn, s.svc.scope, req.DocumentVersionID); err != nil {
				return fmt.Errorf("cannot load document version: %w", err)
			}

			document = &coredata.Document{}
			if err := document.LoadByID(ctx, conn, s.svc.scope, documentVersion.DocumentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			var profile *coredata.MembershipProfile
			var err error
			quorum, profile, err = s.loadQuorumAndProfile(ctx, conn, req.DocumentVersionID, req.IdentityID, documentVersion.OrganizationID)
			if err != nil {
				return fmt.Errorf("cannot load quorum and profile: %w", err)
			}

			if quorum.Status != coredata.DocumentVersionApprovalQuorumStatusPending {
				return &ErrDocumentVersionNotPendingApproval{}
			}

			decision = &coredata.DocumentVersionApprovalDecision{}
			if err := decision.LoadByQuorumIDAndApproverID(ctx, conn, s.svc.scope, quorum.ID, profile.ID); err != nil {
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

	pdfData, err := s.generateApprovalPDF(ctx, req.DocumentVersionID)
	if err != nil {
		return nil, fmt.Errorf("cannot export document PDF: %w", err)
	}

	fileRecord := &coredata.File{
		ID:             gid.New(s.svc.scope.GetTenantID(), coredata.FileEntityType),
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
			if err := quorum.LoadByID(ctx, tx, s.svc.scope, quorumID); err != nil {
				return fmt.Errorf("cannot load quorum: %w", err)
			}

			if quorum.Status != coredata.DocumentVersionApprovalQuorumStatusPending {
				return &ErrDocumentVersionNotPendingApproval{}
			}

			decision = &coredata.DocumentVersionApprovalDecision{}
			if err := decision.LoadByQuorumIDAndApproverID(ctx, tx, s.svc.scope, quorum.ID, approverID); err != nil {
				return fmt.Errorf("cannot load approval decision: %w", err)
			}

			if decision.State != coredata.DocumentVersionApprovalDecisionStatePending {
				return &ErrApprovalDecisionAlreadyMade{}
			}

			if err := fileRecord.Insert(ctx, tx, s.svc.scope); err != nil {
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
					ConsentText:    "By clicking Approve, I consent to approve this document electronically and agree that my electronic signature has the same legal validity as a handwritten signature.",
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

			if err := decision.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update approval decision: %w", err)
			}

			if err := s.maybeApproveQuorum(ctx, tx, quorum.ID); err != nil {
				return fmt.Errorf("cannot check quorum approval: %w", err)
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
	ctx context.Context,
	req RejectDocumentVersionRequest,
) (*coredata.DocumentVersionApprovalDecision, error) {
	var decision *coredata.DocumentVersionApprovalDecision

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			documentVersion := &coredata.DocumentVersion{}
			if err := documentVersion.LoadByID(ctx, tx, s.svc.scope, req.DocumentVersionID); err != nil {
				return fmt.Errorf("cannot load document version: %w", err)
			}

			document := &coredata.Document{}
			if err := document.LoadByID(ctx, tx, s.svc.scope, documentVersion.DocumentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			quorum, profile, err := s.loadQuorumAndProfile(ctx, tx, req.DocumentVersionID, req.IdentityID, documentVersion.OrganizationID)
			if err != nil {
				return fmt.Errorf("cannot load quorum and profile: %w", err)
			}

			decision = &coredata.DocumentVersionApprovalDecision{}
			if err := decision.LoadByQuorumIDAndApproverID(ctx, tx, s.svc.scope, quorum.ID, profile.ID); err != nil {
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

			if err := decision.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update approval decision: %w", err)
			}

			quorum.Status = coredata.DocumentVersionApprovalQuorumStatusRejected
			quorum.UpdatedAt = now

			if err := quorum.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update approval quorum: %w", err)
			}

			decisions := &coredata.DocumentVersionApprovalDecisions{}
			if err := decisions.VoidPendingByQuorumID(ctx, tx, s.svc.scope, quorum.ID, now); err != nil {
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

			if err := documentVersion.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update document version status: %w", err)
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
	ctx context.Context,
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
			if err := documentVersion.LoadByID(ctx, tx, s.svc.scope, documentVersionID); err != nil {
				return fmt.Errorf("cannot load document version: %w", err)
			}

			document := &coredata.Document{}
			if err := document.LoadByID(ctx, tx, s.svc.scope, documentVersion.DocumentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			if documentVersion.Status != coredata.DocumentVersionStatusPendingApproval {
				return &ErrDocumentVersionNotPendingApproval{}
			}

			quorum = &coredata.DocumentVersionApprovalQuorum{}
			if err := quorum.LoadLastByDocumentVersionID(ctx, tx, s.svc.scope, documentVersionID); err != nil {
				return fmt.Errorf("cannot load approval quorum: %w", err)
			}

			if quorum.Status != coredata.DocumentVersionApprovalQuorumStatusPending {
				return &ErrDocumentVersionNotPendingApproval{}
			}

			now := time.Now()

			quorum.Status = coredata.DocumentVersionApprovalQuorumStatusVoided
			quorum.UpdatedAt = now

			if err := quorum.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update approval quorum: %w", err)
			}

			decisions := &coredata.DocumentVersionApprovalDecisions{}
			if err := decisions.VoidPendingByQuorumID(ctx, tx, s.svc.scope, quorum.ID, now); err != nil {
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

			if err := documentVersion.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update document version status: %w", err)
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
	ctx context.Context,
	quorumID gid.GID,
) (*coredata.DocumentVersionApprovalQuorum, error) {
	quorum := &coredata.DocumentVersionApprovalQuorum{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := quorum.LoadByID(ctx, conn, s.svc.scope, quorumID); err != nil {
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
	ctx context.Context,
	documentVersionID gid.GID,
	cursor *page.Cursor[coredata.DocumentVersionApprovalQuorumOrderField],
) (*page.Page[*coredata.DocumentVersionApprovalQuorum, coredata.DocumentVersionApprovalQuorumOrderField], error) {
	var quorums coredata.DocumentVersionApprovalQuorums

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := quorums.LoadAllByDocumentVersionID(ctx, conn, s.svc.scope, documentVersionID, cursor); err != nil {
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
	ctx context.Context,
	documentVersionID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			quorums := &coredata.DocumentVersionApprovalQuorums{}
			count, err = quorums.CountByDocumentVersionID(ctx, conn, s.svc.scope, documentVersionID)
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
	ctx context.Context,
	quorumID gid.GID,
	cursor *page.Cursor[coredata.DocumentVersionApprovalDecisionOrderField],
	filter *coredata.DocumentVersionApprovalDecisionFilter,
) (*page.Page[*coredata.DocumentVersionApprovalDecision, coredata.DocumentVersionApprovalDecisionOrderField], error) {
	var decisions coredata.DocumentVersionApprovalDecisions

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := decisions.LoadByQuorumID(ctx, conn, s.svc.scope, quorumID, cursor, filter); err != nil {
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
	ctx context.Context,
	quorumID gid.GID,
	filter *coredata.DocumentVersionApprovalDecisionFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			decisions := &coredata.DocumentVersionApprovalDecisions{}
			count, err = decisions.CountByQuorumID(ctx, conn, s.svc.scope, quorumID, filter)
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

func (s *DocumentApprovalService) GetViewerDecision(
	ctx context.Context,
	documentVersionID gid.GID,
	identityID gid.GID,
) (*coredata.DocumentVersionApprovalDecision, error) {
	var decision *coredata.DocumentVersionApprovalDecision

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			documentVersion := &coredata.DocumentVersion{}
			if err := documentVersion.LoadByID(ctx, conn, s.svc.scope, documentVersionID); err != nil {
				return fmt.Errorf("cannot load document version: %w", err)
			}

			profile := &coredata.MembershipProfile{}
			if err := profile.LoadByIdentityIDAndOrganizationID(
				ctx,
				conn,
				s.svc.scope,
				identityID,
				documentVersion.OrganizationID,
			); err != nil {
				return fmt.Errorf("cannot load viewer profile: %w", err)
			}

			quorum := &coredata.DocumentVersionApprovalQuorum{}
			if err := quorum.LoadLastByDocumentVersionID(ctx, conn, s.svc.scope, documentVersionID); err != nil {
				return fmt.Errorf("cannot load last approval quorum: %w", err)
			}

			d := &coredata.DocumentVersionApprovalDecision{}
			if err := d.LoadByQuorumIDAndApproverID(ctx, conn, s.svc.scope, quorum.ID, profile.ID); err != nil {
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

func (s *DocumentApprovalService) loadLatestVersion(
	ctx context.Context,
	conn pg.Querier,
	documentID gid.GID,
) (*coredata.DocumentVersion, error) {
	version := &coredata.DocumentVersion{}
	if err := version.LoadLatestVersion(ctx, conn, s.svc.scope, documentID); err != nil {
		return nil, fmt.Errorf("cannot load latest version for document %q: %w", documentID, err)
	}

	return version, nil
}

func (s *DocumentApprovalService) loadQuorumAndProfile(
	ctx context.Context,
	conn pg.Querier,
	documentVersionID gid.GID,
	identityID gid.GID,
	organizationID gid.GID,
) (*coredata.DocumentVersionApprovalQuorum, *coredata.MembershipProfile, error) {
	quorum := &coredata.DocumentVersionApprovalQuorum{}
	if err := quorum.LoadLastByDocumentVersionID(ctx, conn, s.svc.scope, documentVersionID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil, &ErrDocumentVersionNotPendingApproval{}
		}
		return nil, nil, fmt.Errorf("cannot load last approval quorum: %w", err)
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByIdentityIDAndOrganizationID(ctx, conn, s.svc.scope, identityID, organizationID); err != nil {
		return nil, nil, fmt.Errorf("cannot find profile for identity: %w", err)
	}

	return quorum, profile, nil
}

func (s *DocumentApprovalService) createDecisions(
	ctx context.Context,
	tx pg.Tx,
	quorum *coredata.DocumentVersionApprovalQuorum,
	organizationID gid.GID,
	approverIDs []gid.GID,
	now time.Time,
) error {
	decisions := make(coredata.DocumentVersionApprovalDecisions, 0, len(approverIDs))
	for _, approverID := range approverIDs {
		decisions = append(decisions, &coredata.DocumentVersionApprovalDecision{
			ID:             gid.New(s.svc.scope.GetTenantID(), coredata.DocumentVersionApprovalDecisionEntityType),
			OrganizationID: organizationID,
			QuorumID:       quorum.ID,
			ApproverID:     approverID,
			State:          coredata.DocumentVersionApprovalDecisionStatePending,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}

	if err := decisions.BulkInsert(ctx, tx, s.svc.scope); err != nil {
		return fmt.Errorf("cannot insert approval decisions: %w", err)
	}

	return nil
}

func (s *DocumentApprovalService) sendApprovalEmails(
	ctx context.Context,
	tx pg.Tx,
	profiles coredata.MembershipProfiles,
	document *coredata.Document,
	organization *coredata.Organization,
	documentVersionID gid.GID,
) error {
	now := time.Now()
	approvalURLPath := "/organizations/" + document.OrganizationID.String() + "/employee/approvals/" + document.ID.String()

	approvalEmails := make(coredata.Emails, 0, len(profiles))
	for _, profile := range profiles {
		emailPresenter := emails.NewPresenter(s.svc.fileManager, s.svc.bucket, s.svc.baseURL, profile.FullName)

		var (
			emailLinkURLPath = approvalURLPath
			query            = make(url.Values)
		)
		if profile.State != coredata.ProfileStateActive {
			if profile.Source != coredata.ProfileSourceSCIM {
				invitation := &coredata.Invitation{
					ID:             gid.New(document.OrganizationID.TenantID(), coredata.InvitationEntityType),
					OrganizationID: document.OrganizationID,
					UserID:         profile.ID,
					Status:         coredata.InvitationStatusPending,
					ExpiresAt:      now.Add(s.invitationTokenValidity),
					CreatedAt:      now,
				}
				if err := invitation.Insert(ctx, tx, coredata.NewScopeFromObjectID(document.OrganizationID)); err != nil {
					return fmt.Errorf("cannot insert invitation: %w", err)
				}

				invitationToken, err := statelesstoken.NewToken(
					s.tokenSecret,
					iam.TokenTypeOrganizationInvitation,
					s.invitationTokenValidity,
					iam.InvitationTokenData{InvitationID: invitation.ID},
				)
				if err != nil {
					return fmt.Errorf("cannot generate invitation token: %w", err)
				}

				emailLinkURLPath = "/auth/activate-account"
				continueURL := baseurl.MustParse(s.svc.baseURL).AppendPath(approvalURLPath).MustString()
				query.Add("token", invitationToken)
				query.Add("continue", continueURL)
			}
		}

		subject, textBody, htmlBody, err := emailPresenter.RenderDocumentApproval(
			ctx,
			emailLinkURLPath,
			query,
			organization.Name,
			document.Title,
		)
		if err != nil {
			return fmt.Errorf("cannot render approval request email: %w", err)
		}

		approvalEmails = append(approvalEmails, coredata.NewEmail(
			profile.FullName,
			profile.EmailAddress,
			subject,
			textBody,
			htmlBody,
			&coredata.EmailOptions{
				SenderName: new(organization.Name),
			},
		))
	}

	if err := approvalEmails.BulkInsert(ctx, tx); err != nil {
		return fmt.Errorf("cannot insert approval emails: %w", err)
	}

	return nil
}

func (s *DocumentApprovalService) generateApprovalPDF(
	ctx context.Context,
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
				s.svc.scope,
				documentVersionID,
				ExportPDFOptions{},
			)
			return err
		},
	)

	return pdfData, err
}

func (s *DocumentApprovalService) countDecisions(
	ctx context.Context,
	conn pg.Querier,
	quorumID gid.GID,
) (int, error) {
	decisions := &coredata.DocumentVersionApprovalDecisions{}
	count, err := decisions.CountByQuorumID(
		ctx,
		conn,
		s.svc.scope,
		quorumID,
		coredata.NewDocumentVersionApprovalDecisionFilter(nil),
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count decisions: %w", err)
	}

	return count, nil
}

func (s *DocumentApprovalService) maybeApproveQuorum(
	ctx context.Context,
	tx pg.Tx,
	quorumID gid.GID,
) error {
	totalCount, err := s.countDecisions(ctx, tx, quorumID)
	if err != nil {
		return fmt.Errorf("cannot count total decisions: %w", err)
	}

	if totalCount == 0 {
		return nil
	}

	decisions := &coredata.DocumentVersionApprovalDecisions{}
	approvedCount, err := decisions.CountApprovedByQuorumID(ctx, tx, s.svc.scope, quorumID)
	if err != nil {
		return fmt.Errorf("cannot count approved decisions: %w", err)
	}

	if approvedCount != totalCount {
		return nil
	}

	quorum := &coredata.DocumentVersionApprovalQuorum{}
	if err := quorum.LoadByID(ctx, tx, s.svc.scope, quorumID); err != nil {
		return fmt.Errorf("cannot load quorum: %w", err)
	}

	now := time.Now()
	quorum.Status = coredata.DocumentVersionApprovalQuorumStatusApproved
	quorum.UpdatedAt = now

	if err := quorum.Update(ctx, tx, s.svc.scope); err != nil {
		return fmt.Errorf("cannot update quorum: %w", err)
	}

	if err := s.publishVersion(ctx, tx, quorum.VersionID); err != nil {
		return fmt.Errorf("cannot publish version: %w", err)
	}

	return nil
}

func (s *DocumentApprovalService) publishVersion(
	ctx context.Context,
	tx pg.Tx,
	versionID gid.GID,
) error {
	version := &coredata.DocumentVersion{}
	if err := version.LoadByID(ctx, tx, s.svc.scope, versionID); err != nil {
		return fmt.Errorf("cannot load document version: %w", err)
	}

	document := &coredata.Document{}
	if err := document.LoadByID(ctx, tx, s.svc.scope, version.DocumentID); err != nil {
		return fmt.Errorf("cannot load document: %w", err)
	}

	document.CurrentPublishedMajor = &version.Major
	document.CurrentPublishedMinor = &version.Minor

	if err := s.svc.Documents.finalizePublish(ctx, tx, document, version, nil); err != nil {
		return fmt.Errorf("cannot finalize publish: %w", err)
	}

	return nil
}
