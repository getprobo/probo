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

package probo

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/docgen"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/statelesstoken"
	"go.probo.inc/probo/pkg/validator"
	"go.probo.inc/probo/pkg/watermarkpdf"
)

type (
	DocumentService struct {
		svc                     *TenantService
		html2pdfConverter       *html2pdf.Converter
		invitationTokenValidity time.Duration
		tokenSecret             string
	}

	ErrSignatureNotCancellable struct {
		currentState  coredata.DocumentVersionSignatureState
		expectedState coredata.DocumentVersionSignatureState
	}

	ErrDocumentVersionNotDraft struct {
	}

	ErrDocumentVersionNotPublished struct {
	}

	ErrDocumentVersionPendingApproval struct {
	}

	ErrDocumentArchived struct {
	}

	ErrDocumentDraftNotDeletable struct {
	}

	ErrDocumentNotArchived struct {
	}

	ErrDocumentVersionSignatureAlreadySigned struct {
	}

	ErrProfileContractEnded struct {
		ProfileID gid.GID
	}

	CreateDocumentRequest struct {
		OrganizationID        gid.GID
		Title                 string
		Content               string
		Classification        coredata.DocumentClassification
		DocumentType          coredata.DocumentType
		TrustCenterVisibility *coredata.TrustCenterVisibility
		DefaultApproverIDs    []gid.GID
	}

	UpdateDocumentRequest struct {
		DocumentID            gid.GID
		Title                 *string
		Content               *string
		Classification        *coredata.DocumentClassification
		DocumentType          *coredata.DocumentType
		TrustCenterVisibility *coredata.TrustCenterVisibility
		DefaultApproverIDs    *[]gid.GID
	}

	RequestSignatureRequest struct {
		DocumentVersionID gid.GID
		Signatory         gid.GID
	}

	BulkRequestSignaturesRequest struct {
		DocumentIDs  []gid.GID
		SignatoryIDs []gid.GID
	}

	BulkPublishVersionsRequest struct {
		DocumentIDs []gid.GID
		Changelog   string
	}
)

const (
	documentContentMaxTextLength = 50_000
	documentContentMaxJSONBytes  = 500_000
)

func (cdr *CreateDocumentRequest) Validate() error {
	v := validator.New()

	v.Check(cdr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(cdr.Title, "title", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(
		cdr.Content,
		"content",
		validator.MaxLen(documentContentMaxJSONBytes),
		validator.ProseMirrorDocumentContent(),
		validator.ProseMirrorDocumentMaxTextLength(documentContentMaxTextLength),
	)
	v.Check(cdr.Classification, "classification", validator.Required(), validator.OneOfSlice(coredata.DocumentClassifications()))
	v.Check(cdr.DocumentType, "document_type", validator.Required(), validator.OneOfSlice(coredata.DocumentTypes()))
	v.Check(cdr.TrustCenterVisibility, "trust_center_visibility", validator.OneOfSlice(coredata.TrustCenterVisibilities()))
	v.Check(len(cdr.DefaultApproverIDs), "default_approver_ids", validator.Max(100))
	v.Check(cdr.DefaultApproverIDs, "default_approver_ids", validator.NoDuplicates())
	v.CheckEach(cdr.DefaultApproverIDs, "default_approver_ids", func(_ int, item any) {
		v.Check(item, "default_approver_ids", validator.GID(coredata.MembershipProfileEntityType))
	})

	return v.Error()
}

func (udr *UpdateDocumentRequest) Validate() error {
	v := validator.New()

	v.Check(udr.DocumentID, "document_id", validator.Required(), validator.GID(coredata.DocumentEntityType))
	v.Check(udr.TrustCenterVisibility, "trust_center_visibility", validator.OneOfSlice(coredata.TrustCenterVisibilities()))
	if udr.DefaultApproverIDs != nil {
		v.Check(len(*udr.DefaultApproverIDs), "default_approver_ids", validator.Max(100))
		v.Check(*udr.DefaultApproverIDs, "default_approver_ids", validator.NoDuplicates())
		v.CheckEach(*udr.DefaultApproverIDs, "default_approver_ids", func(_ int, item any) {
			v.Check(item, "default_approver_ids", validator.GID(coredata.MembershipProfileEntityType))
		})
	}
	v.Check(udr.Title, "title", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(udr.Classification, "classification", validator.OneOfSlice(coredata.DocumentClassifications()))
	v.Check(
		udr.Content,
		"content",
		validator.MaxLen(documentContentMaxJSONBytes),
		validator.ProseMirrorDocumentContent(),
		validator.ProseMirrorDocumentMaxTextLength(documentContentMaxTextLength),
	)
	v.Check(udr.DocumentType, "document_type", validator.OneOfSlice(coredata.DocumentTypes()))

	return v.Error()
}

const (
	documentExportEmailExpiresIn = 24 * time.Hour

	maxFilenameLength = 200
)

var (
	invalidFilenameChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f\x7f]`)
)

func (e ErrSignatureNotCancellable) Error() string {
	return fmt.Sprintf(
		"cannot cancel signature request: signature is in state %v, expected %v",
		e.currentState,
		e.expectedState,
	)
}

func (e ErrDocumentVersionNotDraft) Error() string {
	return "document version is not a draft"
}

func (e ErrDocumentVersionNotPublished) Error() string {
	return "document version is not published"
}

func (e ErrDocumentVersionPendingApproval) Error() string {
	return "cannot publish a document version that is pending approval"
}

func (e ErrDocumentArchived) Error() string {
	return "cannot modify an archived document"
}

func (e ErrDocumentDraftNotDeletable) Error() string {
	return "latest version is not a deletable draft"
}

func (e ErrDocumentNotArchived) Error() string {
	return "cannot unarchive a document that is not archived"
}

func (e ErrDocumentVersionSignatureAlreadySigned) Error() string {
	return "document version signature already signed"
}

func (e ErrProfileContractEnded) Error() string {
	return fmt.Sprintf("cannot use profile %q: contract has ended", e.ProfileID)
}

func (s *DocumentService) Get(
	ctx context.Context,
	documentID gid.GID,
) (*coredata.Document, error) {
	document := &coredata.Document{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return document.LoadByID(ctx, conn, s.svc.scope, documentID)
		},
	)

	if err != nil {
		return nil, err
	}

	return document, nil
}

func (s *DocumentService) GetDefaultApprovers(
	ctx context.Context,
	documentID gid.GID,
) (coredata.MembershipProfiles, error) {
	var approvers coredata.DocumentDefaultApprovers

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return approvers.LoadByDocumentID(ctx, conn, s.svc.scope, documentID)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot load default approvers: %w", err)
	}

	if len(approvers) == 0 {
		return nil, nil
	}

	profileIDs := make([]gid.GID, len(approvers))
	for i, a := range approvers {
		profileIDs[i] = a.ApproverProfileID
	}

	var profiles coredata.MembershipProfiles

	err = s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return profiles.LoadByIDs(ctx, conn, s.svc.scope, profileIDs)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot load approver profiles: %w", err)
	}

	return profiles, nil
}

func (s *DocumentService) GetByIDs(
	ctx context.Context,
	documentIDs ...gid.GID,
) (coredata.Documents, error) {
	var documents coredata.Documents

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := documents.LoadByIDs(
				ctx,
				conn,
				s.svc.scope,
				documentIDs,
			); err != nil {
				return fmt.Errorf("cannot load documents by ids: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return documents, nil
}

func (s *DocumentService) ListVersionApprovers(
	ctx context.Context,
	documentVersionID gid.GID,
	cursor *page.Cursor[coredata.MembershipProfileOrderField],
) (*page.Page[*coredata.MembershipProfile, coredata.MembershipProfileOrderField], error) {
	var profiles coredata.MembershipProfiles

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := profiles.LoadByDocumentVersionID(ctx, conn, s.svc.scope, documentVersionID, cursor); err != nil {
				return fmt.Errorf("cannot load document version approvers: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(profiles, cursor), nil
}

func (s *DocumentService) CountVersionApprovers(
	ctx context.Context,
	documentVersionID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			profiles := coredata.MembershipProfiles{}
			count, err = profiles.CountByDocumentVersionID(ctx, conn, s.svc.scope, documentVersionID)
			if err != nil {
				return fmt.Errorf("cannot count document version approvers: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *DocumentService) GetWithFilter(
	ctx context.Context,
	documentID gid.GID,
	filter *coredata.DocumentFilter,
) (*coredata.Document, error) {
	document := &coredata.Document{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := document.LoadByIDWithFilter(ctx, conn, s.svc.scope, documentID, filter)
			if err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return document, nil
}

func (s DocumentService) GenerateChangelog(
	ctx context.Context,
	documentID gid.GID,
) (*string, error) {
	var changelog *string
	draftVersion := &coredata.DocumentVersion{}
	publishedVersion := &coredata.DocumentVersion{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := draftVersion.LoadLatestVersion(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load draft version: %w", err)
			}

			if draftVersion.Status != coredata.DocumentVersionStatusDraft {
				return fmt.Errorf("latest version is not a draft")
			}

			document := &coredata.Document{}
			if err := document.LoadByID(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			if document.CurrentPublishedMajor == nil {
				initialVersionChangelog := "Initial version"
				changelog = &initialVersionChangelog
			} else {
				if err := publishedVersion.LoadByDocumentIDAndVersion(ctx, conn, s.svc.scope, documentID, *document.CurrentPublishedMajor, *document.CurrentPublishedMinor); err != nil {
					return fmt.Errorf("cannot load published version: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	if publishedVersion.Content == draftVersion.Content {
		noDiffChangelog := "No changes detected"
		changelog = &noDiffChangelog
	}

	if changelog == nil {
		changelog, err = s.svc.agent.GenerateChangelog(ctx, publishedVersion.Content, draftVersion.Content)
		if err != nil {
			return nil, fmt.Errorf("cannot generate changelog: %w", err)
		}
	}

	return changelog, nil
}

func (s *DocumentService) BulkPublishMinorVersions(
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

				document, version, err := s.publishMinorVersionInTx(ctx, tx, documentID, &req.Changelog, true)
				if err != nil {
					return fmt.Errorf("cannot publish document %q: %w", documentID, err)
				}

				publishedVersions = append(publishedVersions, version)
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

func (s *DocumentService) PublishMajorVersion(
	ctx context.Context,
	documentID gid.GID,
	publishedBy gid.GID,
	changelog *string,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var document *coredata.Document
	var documentVersion *coredata.DocumentVersion

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			dv := &coredata.DocumentVersion{}
			if err := dv.LoadLatestVersion(ctx, tx, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load latest version: %w", err)
			}

			if dv.Status == coredata.DocumentVersionStatusPendingApproval {
				return &ErrDocumentVersionPendingApproval{}
			}

			var err error

			document, documentVersion, err = s.publishMajorVersionInTx(ctx, tx, documentID, changelog, false)
			if err != nil {
				return fmt.Errorf("cannot publish major version: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *DocumentService) PublishMinorVersion(
	ctx context.Context,
	documentID gid.GID,
	publishedBy gid.GID,
	changelog *string,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var document *coredata.Document
	var documentVersion *coredata.DocumentVersion

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			dv := &coredata.DocumentVersion{}
			if err := dv.LoadLatestVersion(ctx, tx, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load latest version: %w", err)
			}

			if dv.Status == coredata.DocumentVersionStatusPendingApproval {
				return &ErrDocumentVersionPendingApproval{}
			}

			var err error

			document, documentVersion, err = s.publishMinorVersionInTx(ctx, tx, documentID, changelog, false)
			if err != nil {
				return fmt.Errorf("cannot publish minor version: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *DocumentService) Create(
	ctx context.Context,
	req CreateDocumentRequest,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	now := time.Now()
	documentID := gid.New(s.svc.scope.GetTenantID(), coredata.DocumentEntityType)
	documentVersionID := gid.New(s.svc.scope.GetTenantID(), coredata.DocumentVersionEntityType)

	organization := &coredata.Organization{}

	document := &coredata.Document{
		ID:                    documentID,
		TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
		Status:                coredata.DocumentStatusActive,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if req.TrustCenterVisibility != nil {
		document.TrustCenterVisibility = *req.TrustCenterVisibility
	}

	content := req.Content
	if strings.TrimSpace(content) != "" {
		var sanitizeErr error
		content, sanitizeErr = prosemirror.SanitizeDocumentJSON(content)
		if sanitizeErr != nil {
			return nil, nil, fmt.Errorf("cannot sanitize document content: %w", sanitizeErr)
		}
	}

	documentVersion := &coredata.DocumentVersion{
		ID:             documentVersionID,
		DocumentID:     documentID,
		Title:          req.Title,
		Major:          0,
		Minor:          1,
		Content:        content,
		Status:         coredata.DocumentVersionStatusDraft,
		Classification: req.Classification,
		DocumentType:   req.DocumentType,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			document.OrganizationID = organization.ID

			if err := document.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert document: %w", err)
			}

			documentVersion.OrganizationID = organization.ID

			if err := documentVersion.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot create document version: %w", err)
			}

			if len(req.DefaultApproverIDs) > 0 {
				approvers := &coredata.DocumentDefaultApprovers{}
				if err := approvers.MergeByDocumentID(ctx, conn, s.svc.scope, documentID, organization.ID, req.DefaultApproverIDs); err != nil {
					return fmt.Errorf("cannot set default approvers: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *DocumentService) SendSigningNotifications(
	ctx context.Context,
	organizationID gid.GID,
) error {
	now := time.Now()

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var signatories coredata.MembershipProfiles
			if err := signatories.LoadAwaitingSigning(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot load signatories: %w", err)
			}

			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			for _, signatory := range signatories {
				emailPresenter := emails.NewPresenter(s.svc.fileManager, s.svc.bucket, s.svc.baseURL, signatory.FullName)

				var (
					employeeDocumentsURLPath = "/organizations/" + organizationID.String() + "/employee"
					emailLinkURLPath         = employeeDocumentsURLPath
					query                    = make(url.Values)
				)
				if signatory.State != coredata.ProfileStateActive {
					if signatory.Source != coredata.ProfileSourceSCIM {
						invitation := &coredata.Invitation{
							ID:             gid.New(organizationID.TenantID(), coredata.InvitationEntityType),
							OrganizationID: organizationID,
							UserID:         signatory.ID,
							Status:         coredata.InvitationStatusPending,
							ExpiresAt:      now.Add(s.invitationTokenValidity),
							CreatedAt:      now,
						}
						if err := invitation.Insert(ctx, tx, coredata.NewScopeFromObjectID(organizationID)); err != nil {
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
						continueURL := baseurl.MustParse(s.svc.baseURL).AppendPath(employeeDocumentsURLPath).MustString()
						query.Add("token", invitationToken)
						query.Add("continue", continueURL)
					}
				}

				subject, textBody, htmlBody, err := emailPresenter.RenderDocumentSigning(
					ctx,
					emailLinkURLPath,
					query,
					organization.Name,
				)
				if err != nil {
					return fmt.Errorf("cannot render signing request email: %w", err)
				}

				email := coredata.NewEmail(
					signatory.FullName,
					signatory.EmailAddress,
					subject,
					textBody,
					htmlBody,
					&coredata.EmailOptions{
						SenderName: new(organization.Name),
					},
				)

				if err := email.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot insert email: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return fmt.Errorf("cannot send signing notifications: %w", err)
	}

	return nil
}

func (s *DocumentService) SignDocumentVersionByIdentity(
	ctx context.Context,
	documentVersionID gid.GID,
	identityID gid.GID,
) (*coredata.DocumentVersionSignature, error) {
	var documentVersionSignature *coredata.DocumentVersionSignature

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			documentVersion := &coredata.DocumentVersion{}
			if err := documentVersion.LoadByID(ctx, conn, s.svc.scope, documentVersionID); err != nil {
				return fmt.Errorf("cannot get document version: %w", err)
			}

			profile := &coredata.MembershipProfile{}
			// FIXME: will be done differently
			if err := profile.LoadByIdentityIDAndOrganizationID(ctx, conn, s.svc.scope, identityID, documentVersion.OrganizationID); err != nil {
				return fmt.Errorf("cannot find profile record for user email in organization %q: %w", documentVersion.OrganizationID, err)
			}

			var signErr error
			documentVersionSignature, signErr = s.signDocumentVersionInTx(ctx, conn, documentVersionID, profile.ID)
			return signErr
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot sign document version: %w", err)
	}

	return documentVersionSignature, nil
}

func (s *DocumentService) signDocumentVersionInTx(
	ctx context.Context,
	conn pg.Tx,
	documentVersionID gid.GID,
	signatory gid.GID,
) (*coredata.DocumentVersionSignature, error) {
	documentVersion := &coredata.DocumentVersion{}
	documentVersionSignature := &coredata.DocumentVersionSignature{}
	now := time.Now()

	if err := documentVersion.LoadByID(ctx, conn, s.svc.scope, documentVersionID); err != nil {
		return nil, fmt.Errorf("cannot load document version %q: %w", documentVersionID, err)
	}

	if documentVersion.Status != coredata.DocumentVersionStatusPublished {
		return nil, fmt.Errorf("cannot sign unpublished version")
	}

	if err := documentVersionSignature.LoadByDocumentVersionIDAndSignatory(ctx, conn, s.svc.scope, documentVersionID, signatory); err != nil {
		return nil, fmt.Errorf("cannot load document version signature: %w", err)
	}

	if documentVersionSignature.State == coredata.DocumentVersionSignatureStateSigned {
		return nil, &ErrDocumentVersionSignatureAlreadySigned{}
	}

	documentVersionSignature.State = coredata.DocumentVersionSignatureStateSigned
	documentVersionSignature.SignedAt = &now
	documentVersionSignature.UpdatedAt = now

	if err := documentVersion.Update(ctx, conn, s.svc.scope); err != nil {
		return nil, fmt.Errorf("cannot update document version: %w", err)
	}

	if err := documentVersionSignature.Update(ctx, conn, s.svc.scope); err != nil {
		return nil, fmt.Errorf("cannot update document version signature: %w", err)
	}

	return documentVersionSignature, nil
}

func (s *DocumentService) updateVersionInTx(
	ctx context.Context,
	tx pg.Tx,
	draftVersion *coredata.DocumentVersion,
	content *string,
	classification *coredata.DocumentClassification,
	documentType *coredata.DocumentType,
	title *string,
) error {
	if content != nil {
		sanitized, err := prosemirror.SanitizeDocumentJSON(*content)
		if err != nil {
			return fmt.Errorf("cannot sanitize document content: %w", err)
		}
		draftVersion.Content = sanitized
	}

	if title != nil {
		draftVersion.Title = *title
	}
	if classification != nil {
		draftVersion.Classification = *classification
	}
	if documentType != nil {
		draftVersion.DocumentType = *documentType
	}
	draftVersion.UpdatedAt = time.Now()

	if err := draftVersion.Update(ctx, tx, s.svc.scope); err != nil {
		return fmt.Errorf("cannot update document version: %w", err)
	}

	return nil
}

func (s *DocumentService) GetVersionSignature(
	ctx context.Context,
	signatureID gid.GID,
) (*coredata.DocumentVersionSignature, error) {
	documentVersionSignature := &coredata.DocumentVersionSignature{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documentVersionSignature.LoadByID(ctx, conn, s.svc.scope, signatureID)
		},
	)

	if err != nil {
		return nil, err
	}

	return documentVersionSignature, nil
}

func (s *DocumentService) BulkRequestSignatures(
	ctx context.Context,
	req BulkRequestSignaturesRequest,
) ([]*coredata.DocumentVersionSignature, error) {
	var signatures []*coredata.DocumentVersionSignature

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			profiles := &coredata.MembershipProfiles{}
			if err := profiles.LoadByIDs(ctx, tx, s.svc.scope, req.SignatoryIDs); err != nil {
				return fmt.Errorf("cannot load signatory profiles: %w", err)
			}

			now := time.Now()
			for _, p := range *profiles {
				if p.ContractEndDate != nil && p.ContractEndDate.Before(now) {
					return &ErrProfileContractEnded{ProfileID: p.ID}
				}
			}

			for _, documentID := range req.DocumentIDs {
				documentVersion := &coredata.DocumentVersion{}
				if err := documentVersion.LoadLatestVersion(ctx, tx, s.svc.scope, documentID); err != nil {
					return fmt.Errorf("cannot load latest version for document %q: %w", documentID, err)
				}

				if documentVersion.Status != coredata.DocumentVersionStatusPublished {
					return fmt.Errorf("cannot request signature for unpublished document %q", documentID)
				}

				for _, signatoryID := range req.SignatoryIDs {
					signature, err := s.createSignatureRequestInTx(ctx, tx, documentVersion.ID, signatoryID, true)
					if err != nil {
						return fmt.Errorf("cannot create signature request for document %q and signatory %q: %w", documentID, signatoryID, err)
					}
					signatures = append(signatures, signature)
				}
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return signatures, nil
}

func (s *DocumentService) createSignatureRequestInTx(
	ctx context.Context,
	tx pg.Tx,
	documentVersionID gid.GID,
	signatoryID gid.GID,
	ignoreExisting bool,
) (*coredata.DocumentVersionSignature, error) {
	signatory := &coredata.MembershipProfile{}
	documentVersion := &coredata.DocumentVersion{}

	if err := documentVersion.LoadByID(ctx, tx, s.svc.scope, documentVersionID); err != nil {
		return nil, fmt.Errorf("cannot load document version: %w", err)
	}

	if err := signatory.LoadByID(ctx, tx, s.svc.scope, signatoryID); err != nil {
		return nil, fmt.Errorf("cannot load signatory: %w", err)
	}

	existingSignature := &coredata.DocumentVersionSignature{}
	err := existingSignature.LoadByDocumentVersionIDAndSignatory(ctx, tx, s.svc.scope, documentVersionID, signatoryID)
	if err == nil && ignoreExisting {
		return existingSignature, nil
	}

	documentVersionSignatureID := gid.New(s.svc.scope.GetTenantID(), coredata.DocumentVersionSignatureEntityType)
	now := time.Now()
	documentVersionSignature := &coredata.DocumentVersionSignature{
		ID:                documentVersionSignatureID,
		OrganizationID:    documentVersion.OrganizationID,
		DocumentVersionID: documentVersionID,
		State:             coredata.DocumentVersionSignatureStateRequested,
		RequestedAt:       now,
		SignedBy:          signatory.ID,
		SignedAt:          nil,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := documentVersionSignature.Insert(ctx, tx, s.svc.scope); err != nil {
		return nil, fmt.Errorf("cannot insert document version signature: %w", err)
	}

	return documentVersionSignature, nil
}

func (s *DocumentService) RequestSignature(
	ctx context.Context,
	req RequestSignatureRequest,
) (*coredata.DocumentVersionSignature, error) {
	var signature *coredata.DocumentVersionSignature

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

			if documentVersion.Status != coredata.DocumentVersionStatusPublished {
				return fmt.Errorf("cannot request signature for unpublished version")
			}

			profile := &coredata.MembershipProfile{}
			if err := profile.LoadByID(ctx, tx, s.svc.scope, req.Signatory); err != nil {
				return fmt.Errorf("cannot load signatory profile: %w", err)
			}

			if profile.ContractEndDate != nil && profile.ContractEndDate.Before(time.Now()) {
				return &ErrProfileContractEnded{ProfileID: profile.ID}
			}

			var err error
			signature, err = s.createSignatureRequestInTx(ctx, tx, req.DocumentVersionID, req.Signatory, false)
			if err != nil {
				return fmt.Errorf("cannot create signature request: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (s *DocumentService) ListSignatures(
	ctx context.Context,
	documentVersionID gid.GID,
	cursor *page.Cursor[coredata.DocumentVersionSignatureOrderField],
	filter *coredata.DocumentVersionSignatureFilter,
) (*page.Page[*coredata.DocumentVersionSignature, coredata.DocumentVersionSignatureOrderField], error) {
	var documentVersionSignatures coredata.DocumentVersionSignatures

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documentVersionSignatures.LoadByDocumentVersionID(ctx, conn, s.svc.scope, documentVersionID, cursor, filter)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(documentVersionSignatures, cursor), nil
}

func (s *DocumentService) IsVersionSignedByUserEmail(
	ctx context.Context,
	documentVersionID gid.GID,
	userEmail mail.Addr,
) (bool, error) {
	documentVersionSignature := &coredata.DocumentVersionSignature{}

	var signed bool
	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error
			signed, err = documentVersionSignature.IsSignedByUserEmail(
				ctx,
				conn,
				s.svc.scope,
				documentVersionID,
				userEmail,
			)
			return err
		},
	)

	if err != nil {
		return false, fmt.Errorf("cannot check if document version is signed: %w", err)
	}

	return signed, nil
}

func (s *DocumentService) createDraftInTx(
	ctx context.Context,
	tx pg.Tx,
	document *coredata.Document,
	latestVersion *coredata.DocumentVersion,
) (*coredata.DocumentVersion, error) {
	now := time.Now()

	draftVersion := &coredata.DocumentVersion{
		ID:             gid.New(s.svc.scope.GetTenantID(), coredata.DocumentVersionEntityType),
		OrganizationID: document.OrganizationID,
		DocumentID:     document.ID,
		Title:          latestVersion.Title,
		Major:          latestVersion.Major,
		Minor:          latestVersion.Minor + 1,
		Classification: latestVersion.Classification,
		DocumentType:   latestVersion.DocumentType,
		Content:        latestVersion.Content,
		Status:         coredata.DocumentVersionStatusDraft,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := draftVersion.Insert(ctx, tx, s.svc.scope); err != nil {
		return nil, fmt.Errorf("cannot create draft: %w", err)
	}

	return draftVersion, nil
}

func (s *DocumentService) deleteDraftInTx(
	ctx context.Context,
	tx pg.Tx,
	draftVersion *coredata.DocumentVersion,
) error {
	if err := draftVersion.Delete(ctx, tx, s.svc.scope); err != nil {
		return fmt.Errorf("cannot delete document version: %w", err)
	}

	return nil
}

func (s *DocumentService) SoftDelete(
	ctx context.Context,
	documentID gid.GID,
) error {
	document := coredata.Document{ID: documentID}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return document.SoftDelete(ctx, tx, s.svc.scope)
		},
	)
}

func (s *DocumentService) BulkSoftDelete(
	ctx context.Context,
	documentIDs []gid.GID,
) error {
	documents := coredata.Documents{}

	for _, documentID := range documentIDs {
		documents = append(documents, &coredata.Document{ID: documentID})
	}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return documents.BulkSoftDelete(ctx, tx, s.svc.scope)
		},
	)
}

func (s *DocumentService) BulkArchive(
	ctx context.Context,
	documentIDs []gid.GID,
) error {
	documents := coredata.Documents{}

	for _, documentID := range documentIDs {
		documents = append(documents, &coredata.Document{ID: documentID})
	}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			controlDocument := coredata.ControlDocument{}
			if err := controlDocument.DeleteByDocumentIDs(ctx, tx, s.svc.scope, documentIDs); err != nil {
				return fmt.Errorf("cannot delete control mappings: %w", err)
			}

			riskDocument := coredata.RiskDocument{}
			if err := riskDocument.DeleteByDocumentIDs(ctx, tx, s.svc.scope, documentIDs); err != nil {
				return fmt.Errorf("cannot delete risk mappings: %w", err)
			}

			measureDocument := coredata.MeasureDocument{}
			if err := measureDocument.DeleteByDocumentIDs(ctx, tx, s.svc.scope, documentIDs); err != nil {
				return fmt.Errorf("cannot delete measure mappings: %w", err)
			}

			return documents.BulkArchive(ctx, tx, s.svc.scope)
		},
	)
}

func (s *DocumentService) BulkUnarchive(
	ctx context.Context,
	documentIDs []gid.GID,
) error {
	documents := coredata.Documents{}

	for _, documentID := range documentIDs {
		documents = append(documents, &coredata.Document{ID: documentID})
	}

	return s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documents.BulkUnarchive(ctx, conn, s.svc.scope)
		},
	)
}

func (s *DocumentService) RequestExport(
	ctx context.Context,
	documentIDs []gid.GID,
	recipientEmail mail.Addr,
	recipientName string,
	options ExportPDFOptions,
) (*coredata.ExportJob, error) {
	var exportJobID gid.GID
	exportJob := &coredata.ExportJob{}

	if options.WithWatermark {
		if options.WatermarkEmail == nil {
			return nil, fmt.Errorf("watermark email is required when with watermark is true")
		}
	}

	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, conn pg.Tx) error {
		var organizationID gid.GID
		for _, documentID := range documentIDs {
			document := &coredata.Document{}
			if err := document.LoadByID(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load document %q: %w", documentID, err)
			}

			organizationID = document.OrganizationID
		}

		now := time.Now()
		exportJobID = gid.New(s.svc.scope.GetTenantID(), coredata.ExportJobEntityType)

		args := coredata.DocumentExportArguments{
			DocumentIDs:    documentIDs,
			WithWatermark:  options.WithWatermark,
			WatermarkEmail: options.WatermarkEmail,
			WithSignatures: options.WithSignatures,
		}
		argsJSON, err := json.Marshal(args)
		if err != nil {
			return fmt.Errorf("cannot marshal document export arguments: %w", err)
		}

		exportJob = &coredata.ExportJob{
			ID:             exportJobID,
			OrganizationID: organizationID,
			Type:           coredata.ExportJobTypeDocument,
			Arguments:      argsJSON,
			Status:         coredata.ExportJobStatusPending,
			RecipientEmail: recipientEmail,
			RecipientName:  recipientName,
			CreatedAt:      now,
		}

		if err := exportJob.Insert(ctx, conn, s.svc.scope); err != nil {
			return fmt.Errorf("cannot insert export job: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return exportJob, nil
}

func (s *DocumentService) CountVersionsForDocumentID(
	ctx context.Context,
	documentID gid.GID,
	filter *coredata.DocumentVersionFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			documentVersions := &coredata.DocumentVersions{}
			count, err = documentVersions.CountByDocumentID(ctx, conn, s.svc.scope, documentID, filter)

			return err
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *DocumentService) CountSignaturesForVersionID(
	ctx context.Context,
	documentVersionID gid.GID,
	filter *coredata.DocumentVersionSignatureFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			documentVersionSignatures := &coredata.DocumentVersionSignatures{}
			count, err = documentVersionSignatures.CountByDocumentVersionID(ctx, conn, s.svc.scope, documentVersionID, filter)

			return err
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *DocumentService) ListVersions(
	ctx context.Context,
	documentID gid.GID,
	cursor *page.Cursor[coredata.DocumentVersionOrderField],
	filter *coredata.DocumentVersionFilter,
) (*page.Page[*coredata.DocumentVersion, coredata.DocumentVersionOrderField], error) {
	var documentVersions coredata.DocumentVersions

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {

			err := documentVersions.LoadByDocumentID(ctx, conn, s.svc.scope, documentID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load document versions: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(documentVersions, cursor), nil
}

func (s *DocumentService) GetVersion(
	ctx context.Context,
	documentVersionID gid.GID,
) (*coredata.DocumentVersion, error) {
	documentVersion := &coredata.DocumentVersion{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documentVersion.LoadByID(ctx, conn, s.svc.scope, documentVersionID)
		},
	)

	if err != nil {
		return nil, err
	}

	return documentVersion, nil
}

func (s *DocumentService) IsSigned(
	ctx context.Context,
	documentID gid.GID,
	userEmail mail.Addr,
) (bool, error) {
	document := &coredata.Document{}

	var signed bool
	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error
			signed, err = document.IsLastSignableVersionSignedByUserEmail(
				ctx,
				conn,
				s.svc.scope,
				documentID,
				userEmail,
			)
			return err
		},
	)

	if err != nil {
		return false, fmt.Errorf("cannot check if document is signed: %w", err)
	}

	return signed, nil
}

func (s *DocumentService) GetViewerApprovalState(
	ctx context.Context,
	documentID gid.GID,
	identityID gid.GID,
) (coredata.DocumentVersionApprovalDecisionState, error) {
	document := &coredata.Document{}

	var state coredata.DocumentVersionApprovalDecisionState
	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error
			state, err = document.GetViewerApprovalStateForLastVersion(
				ctx,
				conn,
				s.svc.scope,
				documentID,
				identityID,
			)
			return err
		},
	)

	if err != nil {
		return "", fmt.Errorf("cannot get viewer approval state: %w", err)
	}

	return state, nil
}

func (s *DocumentService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.DocumentFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			documents := &coredata.Documents{}
			count, err = documents.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count documents: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, fmt.Errorf("cannot count documents: %w", err)
	}

	return count, nil
}

func (s *DocumentService) ListByOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.DocumentOrderField],
	filter *coredata.DocumentFilter,
) (*page.Page[*coredata.Document, coredata.DocumentOrderField], error) {
	var documents coredata.Documents

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documents.LoadByOrganizationID(
				ctx,
				conn,
				s.svc.scope,
				organizationID,
				cursor,
				filter,
			)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(documents, cursor), nil
}

func (s *DocumentService) CountForControlID(
	ctx context.Context,
	controlID gid.GID,
	filter *coredata.DocumentFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			documents := &coredata.Documents{}
			count, err = documents.CountByControlID(ctx, conn, s.svc.scope, controlID, filter)
			if err != nil {
				return fmt.Errorf("cannot count documents: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, fmt.Errorf("cannot count documents: %w", err)
	}

	return count, nil
}

func (s *DocumentService) ListForControlID(
	ctx context.Context,
	controlID gid.GID,
	cursor *page.Cursor[coredata.DocumentOrderField],
	filter *coredata.DocumentFilter,
) (*page.Page[*coredata.Document, coredata.DocumentOrderField], error) {
	var documents coredata.Documents

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documents.LoadByControlID(ctx, conn, s.svc.scope, controlID, cursor, filter)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(documents, cursor), nil
}

func (s *DocumentService) CountForRiskID(
	ctx context.Context,
	riskID gid.GID,
	filter *coredata.DocumentFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			documents := &coredata.Documents{}
			count, err = documents.CountByRiskID(ctx, conn, s.svc.scope, riskID, filter)
			if err != nil {
				return fmt.Errorf("cannot count documents: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, fmt.Errorf("cannot count documents: %w", err)
	}

	return count, nil
}

func (s *DocumentService) ListForRiskID(
	ctx context.Context,
	riskID gid.GID,
	cursor *page.Cursor[coredata.DocumentOrderField],
	filter *coredata.DocumentFilter,
) (*page.Page[*coredata.Document, coredata.DocumentOrderField], error) {
	var documents coredata.Documents

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documents.LoadByRiskID(ctx, conn, s.svc.scope, riskID, cursor, filter)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(documents, cursor), nil
}

func (s *DocumentService) CountForMeasureID(
	ctx context.Context,
	measureID gid.GID,
	filter *coredata.DocumentFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			documents := &coredata.Documents{}
			count, err = documents.CountByMeasureID(ctx, conn, s.svc.scope, measureID, filter)
			if err != nil {
				return fmt.Errorf("cannot count documents: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *DocumentService) ListForMeasureID(
	ctx context.Context,
	measureID gid.GID,
	cursor *page.Cursor[coredata.DocumentOrderField],
	filter *coredata.DocumentFilter,
) (*page.Page[*coredata.Document, coredata.DocumentOrderField], error) {
	var documents coredata.Documents

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := documents.LoadByMeasureID(ctx, conn, s.svc.scope, measureID, cursor, filter); err != nil {
				return fmt.Errorf("cannot list documents for measure: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(documents, cursor), nil
}

func (s *DocumentService) Update(
	ctx context.Context,
	req UpdateDocumentRequest,
) (*coredata.Document, *coredata.DocumentVersion, bool, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, false, err
	}

	document := &coredata.Document{}
	var resultVersion *coredata.DocumentVersion
	var draftCreated bool
	now := time.Now()

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := document.LoadByID(ctx, tx, s.svc.scope, req.DocumentID); err != nil {
				return fmt.Errorf("cannot load document %q: %w", req.DocumentID, err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			if req.TrustCenterVisibility != nil {
				document.TrustCenterVisibility = *req.TrustCenterVisibility
			}

			document.UpdatedAt = now

			if err := document.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update document: %w", err)
			}

			// Handle draft version logic for title/content/classification/type changes.
			latestVersion := &coredata.DocumentVersion{}
			if err := latestVersion.LoadLatestVersion(ctx, tx, s.svc.scope, req.DocumentID); err != nil {
				return fmt.Errorf("cannot load latest version: %w", err)
			}

			hasVersionChanges := req.Title != nil || req.Content != nil || req.Classification != nil || req.DocumentType != nil

			if !hasVersionChanges {
				if req.DefaultApproverIDs != nil {
					defaultApprovers := &coredata.DocumentDefaultApprovers{}
					if err := defaultApprovers.MergeByDocumentID(ctx, tx, s.svc.scope, req.DocumentID, document.OrganizationID, *req.DefaultApproverIDs); err != nil {
						return fmt.Errorf("cannot update default approvers: %w", err)
					}
				}
				return nil
			}

			if latestVersion.Status == coredata.DocumentVersionStatusDraft {
				// Draft exists: update it with any new values.
				if err := s.updateVersionInTx(ctx, tx, latestVersion, req.Content, req.Classification, req.DocumentType, req.Title); err != nil {
					return err
				}

				// If there is a published version and the draft matches it, delete the draft.
				// Never delete the initial draft (v0.1) since there's nothing to fall back to.
				if document.CurrentPublishedMajor != nil && (latestVersion.Major != 0 || latestVersion.Minor != 1) {
					publishedVersion := &coredata.DocumentVersion{}
					if err := publishedVersion.LoadByDocumentIDAndVersion(
						ctx,
						tx,
						s.svc.scope,
						req.DocumentID,
						*document.CurrentPublishedMajor,
						*document.CurrentPublishedMinor,
					); err != nil {
						return fmt.Errorf("cannot load published version: %w", err)
					}

					if latestVersion.Title == publishedVersion.Title &&
						latestVersion.Content == publishedVersion.Content &&
						latestVersion.Classification == publishedVersion.Classification &&
						latestVersion.DocumentType == publishedVersion.DocumentType {
						if err := s.deleteDraftInTx(ctx, tx, latestVersion); err != nil {
							return err
						}
						resultVersion = nil
						return nil
					}
				}

				resultVersion = latestVersion
			} else {
				// No draft exists: create one.
				draftVersion, err := s.createDraftInTx(ctx, tx, document, latestVersion)
				if err != nil {
					return err
				}

				if err := s.updateVersionInTx(ctx, tx, draftVersion, req.Content, req.Classification, req.DocumentType, req.Title); err != nil {
					return err
				}

				resultVersion = draftVersion
				draftCreated = true
			}

			if req.DefaultApproverIDs != nil {
				defaultApprovers := &coredata.DocumentDefaultApprovers{}
				if err := defaultApprovers.MergeByDocumentID(ctx, tx, s.svc.scope, req.DocumentID, document.OrganizationID, *req.DefaultApproverIDs); err != nil {
					return fmt.Errorf("cannot update default approvers: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, false, err
	}

	return document, resultVersion, draftCreated, nil
}

func (s *DocumentService) DeleteDraft(
	ctx context.Context,
	documentID gid.GID,
) (*coredata.Document, error) {
	document := &coredata.Document{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := document.LoadByID(ctx, tx, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load document %q: %w", documentID, err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			latestVersion := &coredata.DocumentVersion{}
			if err := latestVersion.LoadLatestVersion(ctx, tx, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load latest version: %w", err)
			}

			if latestVersion.Status != coredata.DocumentVersionStatusDraft {
				return &ErrDocumentDraftNotDeletable{}
			}

			if latestVersion.Major == 0 && latestVersion.Minor == 1 {
				return &ErrDocumentDraftNotDeletable{}
			}

			return s.deleteDraftInTx(ctx, tx, latestVersion)
		},
	)

	if err != nil {
		return nil, err
	}

	return document, nil
}

func (s *DocumentService) Archive(
	ctx context.Context,
	documentID gid.GID,
) (*coredata.Document, error) {
	document := &coredata.Document{}
	now := time.Now()

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := document.LoadByID(ctx, tx, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load document %q: %w", documentID, err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			controlDocument := coredata.ControlDocument{}
			if err := controlDocument.DeleteByDocumentIDs(ctx, tx, s.svc.scope, []gid.GID{documentID}); err != nil {
				return fmt.Errorf("cannot delete control mappings: %w", err)
			}

			riskDocument := coredata.RiskDocument{}
			if err := riskDocument.DeleteByDocumentIDs(ctx, tx, s.svc.scope, []gid.GID{documentID}); err != nil {
				return fmt.Errorf("cannot delete risk mappings: %w", err)
			}

			measureDocument := coredata.MeasureDocument{}
			if err := measureDocument.DeleteByDocumentIDs(ctx, tx, s.svc.scope, []gid.GID{documentID}); err != nil {
				return fmt.Errorf("cannot delete measure mappings: %w", err)
			}

			document.Status = coredata.DocumentStatusArchived
			document.ArchivedAt = &now
			document.UpdatedAt = now
			document.TrustCenterVisibility = coredata.TrustCenterVisibilityNone

			if err := document.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot archive document: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return document, nil
}

func (s *DocumentService) Unarchive(
	ctx context.Context,
	documentID gid.GID,
) (*coredata.Document, error) {
	document := &coredata.Document{}
	now := time.Now()

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := document.LoadByID(ctx, tx, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load document %q: %w", documentID, err)
			}

			if document.ArchivedAt == nil {
				return &ErrDocumentNotArchived{}
			}

			document.Status = coredata.DocumentStatusActive
			document.ArchivedAt = nil
			document.UpdatedAt = now

			if err := document.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot unarchive document: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return document, nil
}

func (s *DocumentService) CancelSignatureRequest(
	ctx context.Context,
	documentVersionSignatureID gid.GID,
) error {
	documentVersionSignature := &coredata.DocumentVersionSignature{}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := documentVersionSignature.LoadByID(ctx, tx, s.svc.scope, documentVersionSignatureID); err != nil {
				return fmt.Errorf("cannot load document version signature: %w", err)
			}

			documentVersion := &coredata.DocumentVersion{}
			if err := documentVersion.LoadByID(ctx, tx, s.svc.scope, documentVersionSignature.DocumentVersionID); err != nil {
				return fmt.Errorf("cannot load document version: %w", err)
			}

			document := &coredata.Document{}
			if err := document.LoadByID(ctx, tx, s.svc.scope, documentVersion.DocumentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			if documentVersionSignature.State != coredata.DocumentVersionSignatureStateRequested {
				return ErrSignatureNotCancellable{
					currentState:  documentVersionSignature.State,
					expectedState: coredata.DocumentVersionSignatureStateRequested,
				}
			}

			if err := documentVersionSignature.Delete(ctx, tx, s.svc.scope, documentVersionSignatureID); err != nil {
				return fmt.Errorf("cannot delete document version signature: %w", err)
			}

			return nil
		},
	)
}

type ExportPDFOptions struct {
	WithWatermark  bool
	WatermarkEmail *mail.Addr
	WithSignatures bool
}

func (s *DocumentService) ExportPDF(
	ctx context.Context,
	documentVersionID gid.GID,
	options ExportPDFOptions,
) ([]byte, error) {
	var data []byte

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) (err error) {
			data, err = exportDocumentPDF(ctx, s.svc, s.html2pdfConverter, conn, s.svc.scope, documentVersionID, options)
			if err != nil {
				return fmt.Errorf("cannot export document PDF: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *DocumentService) BuildAndUploadExport(ctx context.Context, exportJobID gid.GID) (*coredata.ExportJob, error) {
	exportJob := &coredata.ExportJob{}
	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := exportJob.LoadByID(ctx, tx, s.svc.scope, exportJobID); err != nil {
				return fmt.Errorf("cannot load export job: %w", err)
			}

			documentIDs, err := exportJob.GetDocumentIDs()
			if err != nil {
				return fmt.Errorf("cannot get document IDs: %w", err)
			}

			if len(documentIDs) == 0 {
				return fmt.Errorf("no document IDs found")
			}

			var organizationID gid.GID
			firstDocument := &coredata.Document{}
			if err := firstDocument.LoadByID(ctx, tx, s.svc.scope, documentIDs[0]); err != nil {
				return fmt.Errorf("cannot load document for organization ID: %w", err)
			}
			organizationID = firstDocument.OrganizationID

			tempDir := os.TempDir()
			tempFile, err := os.CreateTemp(tempDir, "probo-document-export-*.zip")
			if err != nil {
				return fmt.Errorf("cannot create temp file: %w", err)
			}
			defer func() { _ = tempFile.Close() }()
			defer func() { _ = os.Remove(tempFile.Name()) }()

			exportArgs, err := exportJob.GetDocumentExportArguments()
			if err != nil {
				return fmt.Errorf("cannot get export arguments: %w", err)
			}

			exportOptions := ExportPDFOptions{
				WithWatermark:  exportArgs.WithWatermark,
				WatermarkEmail: exportArgs.WatermarkEmail,
				WithSignatures: exportArgs.WithSignatures,
			}

			err = s.Export(ctx, documentIDs, tempFile, exportOptions)
			if err != nil {
				return fmt.Errorf("cannot export documents: %w", err)
			}

			uuid, err := uuid.NewV4()
			if err != nil {
				return fmt.Errorf("cannot generate uuid: %w", err)
			}

			if _, err := tempFile.Seek(0, 0); err != nil {
				return fmt.Errorf("cannot seek temp file: %w", err)
			}

			fileInfo, err := tempFile.Stat()
			if err != nil {
				return fmt.Errorf("cannot stat temp file: %w", err)
			}

			_, err = s.svc.s3.PutObject(
				ctx,
				&s3.PutObjectInput{
					Bucket:        new(s.svc.bucket),
					Key:           new(uuid.String()),
					Body:          tempFile,
					ContentLength: new(fileInfo.Size()),
					ContentType:   new("application/zip"),
					CacheControl:  new("private, max-age=3600"),
					Metadata: map[string]string{
						"type":            "document-export",
						"export-job-id":   exportJob.ID.String(),
						"organization-id": organizationID.String(),
					},
				},
			)
			if err != nil {
				return fmt.Errorf("cannot upload file to S3: %w", err)
			}

			now := time.Now()

			file := coredata.File{
				ID:         gid.New(exportJob.ID.TenantID(), coredata.FileEntityType),
				BucketName: s.svc.bucket,
				MimeType:   "application/zip",
				FileName:   fmt.Sprintf("Documents Export %s.zip", now.Format("2006-01-02")),
				FileKey:    uuid.String(),
				FileSize:   fileInfo.Size(),
				Visibility: coredata.FileVisibilityPrivate,
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			if err := file.Insert(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert file: %w", err)
			}

			exportJob.FileID = &file.ID
			if err := exportJob.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update export job: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return exportJob, nil
}

func exportDocumentPDF(
	ctx context.Context,
	svc *TenantService,
	html2pdfConverter *html2pdf.Converter,
	conn pg.Querier,
	scope coredata.Scoper,
	documentVersionID gid.GID,
	options ExportPDFOptions,
) ([]byte, error) {
	document := &coredata.Document{}
	version := &coredata.DocumentVersion{}
	organization := &coredata.Organization{}

	if err := version.LoadByID(ctx, conn, scope, documentVersionID); err != nil {
		return nil, fmt.Errorf("cannot load document version: %w", err)
	}

	if err := document.LoadByID(ctx, conn, scope, version.DocumentID); err != nil {
		return nil, fmt.Errorf("cannot load document: %w", err)
	}

	// Only show approvers from the last approved quorum in the export.
	var approverNames []string

	lastQuorum := &coredata.DocumentVersionApprovalQuorum{}
	if err := lastQuorum.LoadLastByDocumentVersionID(ctx, conn, scope, documentVersionID); err != nil {
		if !errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, fmt.Errorf("cannot load last approval quorum: %w", err)
		}
	} else if lastQuorum.Status == coredata.DocumentVersionApprovalQuorumStatusApproved {
		approvedDecisions := &coredata.DocumentVersionApprovalDecisions{}
		approvedFilter := coredata.NewDocumentVersionApprovalDecisionFilter(
			coredata.DocumentVersionApprovalDecisionStates{coredata.DocumentVersionApprovalDecisionStateApproved},
		)
		if err := approvedDecisions.LoadByQuorumID(
			ctx,
			conn,
			scope,
			lastQuorum.ID,
			page.NewCursor(
				100,
				nil,
				page.Head,
				page.OrderBy[coredata.DocumentVersionApprovalDecisionOrderField]{
					Field:     coredata.DocumentVersionApprovalDecisionOrderFieldCreatedAt,
					Direction: page.OrderDirectionAsc,
				},
			),
			approvedFilter,
		); err != nil {
			return nil, fmt.Errorf("cannot load approved decisions: %w", err)
		}

		approverProfileIDs := make([]gid.GID, 0, len(*approvedDecisions))
		for _, d := range *approvedDecisions {
			approverProfileIDs = append(approverProfileIDs, d.ApproverID)
		}

		if len(approverProfileIDs) > 0 {
			approverProfiles := coredata.MembershipProfiles{}
			if err := approverProfiles.LoadByIDs(ctx, conn, scope, approverProfileIDs); err != nil {
				return nil, fmt.Errorf("cannot load approver profiles: %w", err)
			}

			approverNames = make([]string, 0, len(approverProfiles))
			for _, p := range approverProfiles {
				approverNames = append(approverNames, p.FullName)
			}
		}
	}

	if err := organization.LoadByID(ctx, conn, scope, document.OrganizationID); err != nil {
		return nil, fmt.Errorf("cannot load organization: %w", err)
	}

	var signatureData []docgen.SignatureData
	if options.WithSignatures {
		signaturesWithPeople := &coredata.DocumentVersionSignaturesWithPeople{}
		if err := signaturesWithPeople.LoadByDocumentVersionIDWithPeople(ctx, conn, scope, documentVersionID, 1_000); err != nil {
			return nil, fmt.Errorf("cannot load document version signatures: %w", err)
		}

		signatureData = make([]docgen.SignatureData, len(*signaturesWithPeople))
		for i, sig := range *signaturesWithPeople {
			signatureData[i] = docgen.SignatureData{
				SignedBy:    sig.SignedByFullName,
				SignedAt:    sig.SignedAt,
				State:       sig.State,
				RequestedAt: sig.RequestedAt,
			}
		}
	}

	classification := docgen.ClassificationSecret
	switch version.Classification {
	case coredata.DocumentClassificationPublic:
		classification = docgen.ClassificationPublic
	case coredata.DocumentClassificationInternal:
		classification = docgen.ClassificationInternal
	case coredata.DocumentClassificationConfidential:
		classification = docgen.ClassificationConfidential
	}

	horizontalLogoBase64 := ""
	if organization.HorizontalLogoFileID != nil {
		fileRecord := &coredata.File{}
		fileErr := fileRecord.LoadByID(ctx, conn, scope, *organization.HorizontalLogoFileID)
		if fileErr == nil {
			base64Data, mimeType, logoErr := svc.fileManager.GetFileBase64(ctx, fileRecord)
			if logoErr == nil {
				horizontalLogoBase64 = fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
			}
		}
	}

	docData := docgen.DocumentData{
		Title:                       version.Title,
		Content:                     json.RawMessage([]byte(version.Content)),
		Major:                       version.Major,
		Minor:                       version.Minor,
		Classification:              classification,
		Approvers:                   approverNames,
		PublishedAt:                 version.PublishedAt,
		Signatures:                  signatureData,
		CompanyHorizontalLogoBase64: horizontalLogoBase64,
	}

	htmlContent, err := docgen.RenderHTML(docData)
	if err != nil {
		return nil, fmt.Errorf("cannot generate HTML: %w", err)
	}

	cfg := html2pdf.RenderConfig{
		PageFormat:        html2pdf.PageFormatA4,
		Orientation:       html2pdf.OrientationPortrait,
		MarginTop:         html2pdf.NewMarginInches(1.0),
		MarginBottom:      html2pdf.NewMarginInches(1.0),
		MarginLeft:        html2pdf.NewMarginInches(1.0),
		MarginRight:       html2pdf.NewMarginInches(1.0),
		PrintBackground:   true,
		Scale:             1.0,
		WaitForExpression: "window.__mermaidReady === true",
	}

	pdfReader, err := html2pdfConverter.GeneratePDF(ctx, htmlContent, cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot generate PDF: %w", err)
	}

	pdfData, err := io.ReadAll(pdfReader)
	if err != nil {
		return nil, fmt.Errorf("cannot read PDF data: %w", err)
	}

	if options.WithWatermark {
		if options.WatermarkEmail == nil {
			return nil, fmt.Errorf("watermark email is required with watermark enabled")
		}

		watermarkedPDF, err := watermarkpdf.AddConfidentialWithTimestamp(pdfData, *options.WatermarkEmail)
		if err != nil {
			return nil, fmt.Errorf("cannot add watermark to PDF: %w", err)
		}
		return watermarkedPDF, nil
	}

	return pdfData, nil
}

func (s *DocumentService) Export(
	ctx context.Context,
	documentIDs []gid.GID,
	file io.Writer,
	options ExportPDFOptions,
) (err error) {
	archive := zip.NewWriter(file)
	defer func() {
		if closeErr := archive.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("cannot close archive: %w", closeErr)
		}
	}()

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			for i, documentID := range documentIDs {
				document := &coredata.Document{}
				if err := document.LoadByID(ctx, conn, s.svc.scope, documentID); err != nil {
					return fmt.Errorf("cannot load document %q: %w", documentID, err)
				}

				documentVersion := &coredata.DocumentVersion{}
				if err := documentVersion.LoadLatestVersion(ctx, conn, s.svc.scope, documentID); err != nil {
					return fmt.Errorf("cannot load document version for %q: %w", documentID, err)
				}

				exportedPDF, err := exportDocumentPDF(
					ctx,
					s.svc,
					s.html2pdfConverter,
					conn,
					s.svc.scope,
					documentVersion.ID,
					options,
				)
				if err != nil {
					return fmt.Errorf("cannot export document PDF for %q: %w", documentID, err)
				}

				filename := fmt.Sprintf("%d_%s.pdf", i+1, sanitizeFilename(document.Title))
				w, err := archive.Create(filename)
				if err != nil {
					return fmt.Errorf("cannot create document in archive: %w", err)
				}

				_, err = w.Write(exportedPDF)
				if err != nil {
					return fmt.Errorf("cannot write document to archive: %w", err)
				}
			}

			return nil
		},
	)
}

func (s *DocumentService) SendExportEmail(
	ctx context.Context,
	fileID gid.GID,
	recipientName string,
	recipientEmail mail.Addr,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			file := &coredata.File{}
			if err := file.LoadByID(ctx, tx, s.svc.scope, fileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			downloadURL, err := s.GenerateDocumentExportDownloadURL(ctx, file)
			if err != nil {
				return fmt.Errorf("cannot generate download URL: %w", err)
			}

			emailPresenter := emails.NewPresenter(s.svc.fileManager, s.svc.bucket, s.svc.baseURL, recipientName)

			subject, textBody, htmlBody, err := emailPresenter.RenderDocumentExport(
				ctx,
				downloadURL,
			)
			if err != nil {
				return fmt.Errorf("cannot render document export email: %w", err)
			}

			email := coredata.NewEmail(
				recipientName,
				recipientEmail,
				subject,
				textBody,
				htmlBody,
				nil,
			)

			if err := email.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
			}

			return nil
		},
	)
}

func (s *DocumentService) GenerateDocumentExportDownloadURL(
	ctx context.Context,
	file *coredata.File,
) (string, error) {
	presignClient := s3.NewPresignClient(s.svc.s3)

	presignedReq, err := presignClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket:                     new(s.svc.bucket),
			Key:                        new(file.FileKey),
			ResponseCacheControl:       new("max-age=3600, public"),
			ResponseContentType:        new(file.MimeType),
			ResponseContentDisposition: new(fmt.Sprintf("attachment; filename=\"%s\"", file.FileName)),
		},
		func(opts *s3.PresignOptions) {
			opts.Expires = documentExportEmailExpiresIn
		},
	)

	if err != nil {
		return "", fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return presignedReq.URL, nil
}

func sanitizeFilename(title string) string {
	if title == "" {
		return "Untitled"
	}

	sanitized := invalidFilenameChars.ReplaceAllString(title, "_")

	sanitized = strings.TrimFunc(sanitized, func(r rune) bool {
		return unicode.IsSpace(r) || r == '.'
	})

	sanitized = regexp.MustCompile(`[\s_]+`).ReplaceAllString(sanitized, "_")

	if sanitized == "" || sanitized == "_" {
		sanitized = "Untitled"
	}

	if len(sanitized) > maxFilenameLength-20 {
		sanitized = sanitized[:maxFilenameLength-20]
		sanitized = strings.TrimFunc(sanitized, func(r rune) bool {
			return r == unicode.ReplacementChar
		})
	}

	return sanitized
}

func (s *DocumentService) loadDraftForPublish(
	ctx context.Context,
	tx pg.Tx,
	documentID gid.GID,
	ignoreExisting bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	document := &coredata.Document{}
	documentVersion := &coredata.DocumentVersion{}

	if err := document.LoadByID(ctx, tx, s.svc.scope, documentID); err != nil {
		return nil, nil, fmt.Errorf("cannot load document %q: %w", documentID, err)
	}

	if document.ArchivedAt != nil {
		return nil, nil, &ErrDocumentArchived{}
	}

	if err := documentVersion.LoadLatestVersion(ctx, tx, s.svc.scope, documentID); err != nil {
		return nil, nil, fmt.Errorf("cannot load current draft: %w", err)
	}

	if ignoreExisting && documentVersion.Status == coredata.DocumentVersionStatusPublished {
		return document, documentVersion, nil
	}

	if documentVersion.Status != coredata.DocumentVersionStatusDraft && documentVersion.Status != coredata.DocumentVersionStatusPendingApproval {
		return nil, nil, &ErrDocumentVersionNotDraft{}
	}

	return document, documentVersion, nil
}

func (s *DocumentService) finalizePublish(
	ctx context.Context,
	tx pg.Tx,
	document *coredata.Document,
	documentVersion *coredata.DocumentVersion,
	changelog *string,
) error {
	now := time.Now()

	if changelog != nil {
		documentVersion.Changelog = *changelog
	}

	document.UpdatedAt = now
	documentVersion.Status = coredata.DocumentVersionStatusPublished
	documentVersion.PublishedAt = &now
	documentVersion.UpdatedAt = now

	if err := document.Update(ctx, tx, s.svc.scope); err != nil {
		return fmt.Errorf("cannot update document: %w", err)
	}

	if err := documentVersion.Update(ctx, tx, s.svc.scope); err != nil {
		return fmt.Errorf("cannot update document version: %w", err)
	}

	return nil
}

func (s *DocumentService) publishMajorVersionInTx(
	ctx context.Context,
	tx pg.Tx,
	documentID gid.GID,
	changelog *string,
	ignoreExisting bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	document, documentVersion, err := s.loadDraftForPublish(ctx, tx, documentID, ignoreExisting)
	if err != nil {
		return nil, nil, err
	}

	if ignoreExisting && documentVersion.Status == coredata.DocumentVersionStatusPublished {
		return document, documentVersion, nil
	}

	if document.CurrentPublishedMajor != nil {
		publishedVersion := &coredata.DocumentVersion{}
		if err := publishedVersion.LoadByDocumentIDAndVersion(ctx, tx, s.svc.scope, documentID, *document.CurrentPublishedMajor, *document.CurrentPublishedMinor); err != nil {
			return nil, nil, fmt.Errorf("cannot load published version: %w", err)
		}

		documentVersion.Major = *document.CurrentPublishedMajor + 1
	} else {
		documentVersion.Major = 1
	}

	documentVersion.Minor = 0
	document.CurrentPublishedMajor = &documentVersion.Major
	document.CurrentPublishedMinor = &documentVersion.Minor

	if err := s.finalizePublish(ctx, tx, document, documentVersion, changelog); err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *DocumentService) publishMinorVersionInTx(
	ctx context.Context,
	tx pg.Tx,
	documentID gid.GID,
	changelog *string,
	ignoreExisting bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	document, documentVersion, err := s.loadDraftForPublish(ctx, tx, documentID, ignoreExisting)
	if err != nil {
		return nil, nil, err
	}

	if ignoreExisting && documentVersion.Status == coredata.DocumentVersionStatusPublished {
		return document, documentVersion, nil
	}

	document.CurrentPublishedMajor = &documentVersion.Major
	document.CurrentPublishedMinor = &documentVersion.Minor

	if err := s.finalizePublish(ctx, tx, document, documentVersion, changelog); err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}
