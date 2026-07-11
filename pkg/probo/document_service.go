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

package probo

import (
	"archive/zip"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/docgen"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/llm"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/pdfutils"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/validator"
	"go.probo.inc/probo/pkg/webhook"
	webhooktypes "go.probo.inc/probo/pkg/webhook/types"
)

const DocumentSignatureConsentText = "By clicking \"Review and sign\", I consent to sign this document electronically and agree that my electronic signature has the same legal validity as a handwritten signature."

type (
	DocumentService struct {
		svc                     *Service
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

	ErrDocumentVersionNotCurrent struct {
	}

	ErrDocumentVersionPendingApproval struct {
	}

	ErrDocumentArchived struct {
	}

	ErrDocumentDraftNotDeletable struct {
	}

	ErrDocumentNotArchived struct {
	}

	ErrDocumentGenerated struct {
	}

	ErrDocumentVersionGenerated struct {
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

	SignDocumentVersionRequest struct {
		DocumentVersionID gid.GID
		IdentityID        gid.GID
		SignerFullName    string
		SignerEmail       mail.Addr
		SignerIPAddr      string
		SignerUA          string
	}

	BulkRequestSignaturesRequest struct {
		DocumentIDs  []gid.GID
		SignatoryIDs []gid.GID
	}

	BulkPublishVersionsRequest struct {
		DocumentIDs []gid.GID
		Minor       bool
		Changelog   string
	}

	PublishDocumentRequest struct {
		DocumentID  gid.GID
		Minor       bool
		ApproverIDs []gid.GID
		Changelog   string
	}

	PublishDocumentResult struct {
		Document *coredata.Document
		Version  *coredata.DocumentVersion
		Quorum   *coredata.DocumentVersionApprovalQuorum
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

func (req *PublishDocumentRequest) Validate() error {
	v := validator.New()

	v.Check(req.DocumentID, "document_id", validator.Required(), validator.GID(coredata.DocumentEntityType))
	v.Check(len(req.ApproverIDs), "approver_ids", validator.Max(100))
	v.Check(req.ApproverIDs, "approver_ids", validator.NoDuplicates())
	v.CheckEach(req.ApproverIDs, "approver_ids", func(index int, item any) {
		v.Check(item, fmt.Sprintf("approver_ids[%d]", index), validator.GID(coredata.MembershipProfileEntityType))
	})
	v.Check(req.Changelog, "changelog", validator.Required(), validator.SafeText(5000))

	// approver_ids must be an explicit choice for a major publish (an empty list
	// publishes directly without approval, a non-empty list requests approval)
	// and must be omitted for a minor publish, which ignores approvers.
	if req.Minor && req.ApproverIDs != nil {
		v.Check(req.ApproverIDs, "approver_ids", func(any) *validator.ValidationError {
			return &validator.ValidationError{
				Code:    validator.ErrorCodeCustom,
				Message: "must not be set when publishing a minor version",
			}
		})
	}

	if !req.Minor && req.ApproverIDs == nil {
		v.Check(req.ApproverIDs, "approver_ids", func(any) *validator.ValidationError {
			return &validator.ValidationError{
				Code:    validator.ErrorCodeCustom,
				Message: "must be set when publishing a major version: provide approver profile IDs to request approval, or an empty list to publish directly without approval",
			}
		})
	}

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

func (e ErrDocumentVersionNotCurrent) Error() string {
	return "document version is not the current published version"
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

func (e ErrDocumentGenerated) Error() string {
	return "cannot create draft for a generated document"
}

func (e ErrDocumentVersionGenerated) Error() string {
	return "cannot edit content of a generated document version"
}

func (e ErrDocumentVersionSignatureAlreadySigned) Error() string {
	return "document version signature already signed"
}

func (e ErrProfileContractEnded) Error() string {
	return fmt.Sprintf("cannot use profile %q: contract has ended", e.ProfileID)
}

func (s *DocumentService) Get(
	ctx context.Context, scope coredata.Scoper,
	documentID gid.GID,
) (*coredata.Document, error) {
	document := &coredata.Document{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return document.LoadByID(ctx, conn, scope, documentID)
		},
	)
	if err != nil {
		return nil, err
	}

	return document, nil
}

func (s *DocumentService) GetDefaultApprovers(
	ctx context.Context, scope coredata.Scoper,
	documentID gid.GID,
) (coredata.MembershipProfiles, error) {
	var approvers coredata.DocumentDefaultApprovers

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return approvers.LoadByDocumentID(ctx, conn, scope, documentID)
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
			if err := profiles.LoadByIDs(ctx, conn, scope, profileIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load approver profiles: %w", err)
	}

	return profiles, nil
}

func (s *DocumentService) GetByIDs(
	ctx context.Context, scope coredata.Scoper,
	documentIDs ...gid.GID,
) (coredata.Documents, error) {
	var documents coredata.Documents

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := documents.LoadByIDs(
				ctx,
				conn,
				scope,
				documentIDs,
			); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
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
	ctx context.Context, scope coredata.Scoper,
	documentVersionID gid.GID,
	cursor *page.Cursor[coredata.MembershipProfileOrderField],
) (*page.Page[*coredata.MembershipProfile, coredata.MembershipProfileOrderField], error) {
	var profiles coredata.MembershipProfiles

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := profiles.LoadByDocumentVersionID(ctx, conn, scope, documentVersionID, cursor); err != nil {
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
	ctx context.Context, scope coredata.Scoper,
	documentVersionID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			profiles := coredata.MembershipProfiles{}

			count, err = profiles.CountByDocumentVersionID(ctx, conn, scope, documentVersionID)
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
	ctx context.Context, scope coredata.Scoper,
	documentID gid.GID,
	filter *coredata.DocumentFilter,
) (*coredata.Document, error) {
	document := &coredata.Document{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := document.LoadByIDWithFilter(ctx, conn, scope, documentID, filter)
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
	ctx context.Context, scope coredata.Scoper,
	documentID gid.GID,
) (*string, error) {
	var changelog *string

	draftVersion := &coredata.DocumentVersion{}
	publishedVersion := &coredata.DocumentVersion{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := draftVersion.LoadLatestVersion(ctx, conn, scope, documentID); err != nil {
				return fmt.Errorf("cannot load draft version: %w", err)
			}

			if draftVersion.Status != coredata.DocumentVersionStatusDraft {
				return fmt.Errorf("latest version is not a draft")
			}

			document := &coredata.Document{}
			if err := document.LoadByID(ctx, conn, scope, documentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			if document.CurrentPublishedMajor == nil {
				initialVersionChangelog := "Initial version"
				changelog = &initialVersionChangelog
			} else {
				if err := publishedVersion.LoadByDocumentIDAndVersion(ctx, conn, scope, documentID, *document.CurrentPublishedMajor, *document.CurrentPublishedMinor); err != nil {
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
		changelog, err = s.generateChangelog(ctx, scope, publishedVersion.Content, draftVersion.Content)
		if err != nil {
			return nil, fmt.Errorf("cannot generate changelog: %w", err)
		}
	}

	return changelog, nil
}

//go:embed prompts/changelog_generator.txt
var changelogGeneratorSystemPrompt string

func (s DocumentService) generateChangelog(
	ctx context.Context, scope coredata.Scoper,
	oldContent, newContent string,
) (*string, error) {
	ag := agent.New(
		"changelog_generator",
		s.svc.llmClient,
		agent.WithInstructions(changelogGeneratorSystemPrompt),
		agent.WithModel(s.svc.llmConfig.Model),
		agent.WithTemperature(s.svc.llmConfig.Temperature),
		agent.WithMaxTokens(s.svc.llmConfig.MaxTokens),
	)

	result, err := ag.Run(
		ctx,
		[]llm.Message{
			{
				Role: llm.RoleUser,
				Parts: []llm.Part{
					llm.TextPart{Text: fmt.Sprintf("Old content: %s", oldContent)},
					llm.TextPart{Text: fmt.Sprintf("New content: %s", newContent)},
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot generate changelog: %w", err)
	}

	text := result.FinalMessage().Text()

	return &text, nil
}

func (s *DocumentService) PublishVersionWithDefaultApprovers(
	ctx context.Context,
	scope coredata.Scoper,
	req PublishDocumentRequest,
) (*PublishDocumentResult, error) {
	var result *PublishDocumentResult

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if !req.Minor {
				defaultApprovers := &coredata.DocumentDefaultApprovers{}
				if err := defaultApprovers.LoadByDocumentID(ctx, tx, scope, req.DocumentID); err != nil {
					return fmt.Errorf("cannot load default approvers: %w", err)
				}

				approverIDs := make([]gid.GID, len(*defaultApprovers))
				for i, a := range *defaultApprovers {
					approverIDs[i] = a.ApproverProfileID
				}

				req.ApproverIDs = approverIDs
			}

			publishResult, err := s.publish(ctx, scope, tx, req)
			if err != nil {
				return fmt.Errorf("cannot publish version: %w", err)
			}

			result = publishResult

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// PublishVersion publishes a document version. The behaviour depends on
// req.Minor and req.ApproverIDs:
//   - Minor=true: publish the existing draft as a minor version. ApproverIDs
//     are ignored.
//   - Minor=false with ApproverIDs: open an approval quorum on the draft as
//     a pending major bump (currentMajor+1.0). Result.Quorum is set.
//   - Minor=false without ApproverIDs: publish the draft immediately as a
//     major bump (currentMajor+1.0).
//
// A major publish also replaces the document's default approvers with
// req.ApproverIDs, even when the list is empty.
func (s *DocumentService) PublishVersion(
	ctx context.Context,
	scope coredata.Scoper,
	req PublishDocumentRequest,
) (*PublishDocumentResult, error) {
	var result *PublishDocumentResult

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			publishResult, err := s.publish(ctx, scope, tx, req)
			if err != nil {
				return fmt.Errorf("cannot publish version: %w", err)
			}

			result = publishResult

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *DocumentService) publish(
	ctx context.Context,
	scope coredata.Scoper,
	tx pg.Tx,
	req PublishDocumentRequest,
) (*PublishDocumentResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	result := &PublishDocumentResult{}

	dv := &coredata.DocumentVersion{}
	if err := dv.LoadLatestVersion(ctx, tx, scope, req.DocumentID); err != nil {
		return nil, fmt.Errorf("cannot load latest version: %w", err)
	}

	if dv.Status == coredata.DocumentVersionStatusPendingApproval {
		return nil, &ErrDocumentVersionPendingApproval{}
	}

	if req.Minor {
		document, version, err := s.publishMinor(ctx, scope, tx, req.DocumentID, &req.Changelog, false)
		if err != nil {
			return nil, fmt.Errorf("cannot publish minor version: %w", err)
		}

		result.Document = document
		result.Version = version

		return result, nil
	}

	document := &coredata.Document{}
	if err := document.LoadByID(ctx, tx, scope, req.DocumentID); err != nil {
		return nil, fmt.Errorf("cannot load document: %w", err)
	}

	if document.ArchivedAt != nil {
		return nil, &ErrDocumentArchived{}
	}

	defaultApprovers := &coredata.DocumentDefaultApprovers{}
	if err := defaultApprovers.MergeByDocumentID(ctx, tx, scope, req.DocumentID, document.OrganizationID, req.ApproverIDs); err != nil {
		return nil, fmt.Errorf("cannot update default approvers: %w", err)
	}

	if len(req.ApproverIDs) == 0 {
		publishedDocument, version, err := s.publishMajor(ctx, scope, tx, req.DocumentID, &req.Changelog, false)
		if err != nil {
			return nil, fmt.Errorf("cannot publish major version: %w", err)
		}

		result.Document = publishedDocument
		result.Version = version

		return result, nil
	}

	if err := s.requestMajorApproval(ctx, scope, tx, req, document, dv, result); err != nil {
		return nil, fmt.Errorf("cannot request major approval: %w", err)
	}

	return result, nil
}

func (s *DocumentService) requestMajorApproval(
	ctx context.Context,
	scope coredata.Scoper,
	tx pg.Tx,
	req PublishDocumentRequest,
	document *coredata.Document,
	dv *coredata.DocumentVersion,
	result *PublishDocumentResult,
) error {
	profiles := &coredata.MembershipProfiles{}
	if err := profiles.LoadByIDs(ctx, tx, scope, req.ApproverIDs); err != nil {
		return fmt.Errorf("cannot load approver profiles: %w", err)
	}

	now := time.Now()
	for _, p := range *profiles {
		if p.ContractEndDate != nil && p.ContractEndDate.Before(now) {
			return &ErrProfileContractEnded{ProfileID: p.ID}
		}
	}

	if dv.Status != coredata.DocumentVersionStatusDraft {
		return &ErrDocumentVersionNotDraft{}
	}

	quorum, err := s.svc.DocumentApprovals.RequestApproval(ctx, scope, tx, document, dv, req.ApproverIDs, &req.Changelog)
	if err != nil {
		return fmt.Errorf("cannot request approval: %w", err)
	}

	result.Document = document
	result.Version = dv
	result.Quorum = quorum

	if err := s.emitDocumentEvent(
		ctx,
		scope,
		tx,
		dv.DocumentID,
		coredata.WebhookEventTypeDocumentVersionApprovalQuorumRequested,
		dv,
		nil,
		&quorum.ID,
		nil,
	); err != nil {
		return fmt.Errorf("cannot emit approval quorum requested webhook: %w", err)
	}

	return nil
}

func (s *DocumentService) Create(
	ctx context.Context, scope coredata.Scoper,
	req CreateDocumentRequest,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	now := time.Now()
	documentID := gid.New(scope.GetTenantID(), coredata.DocumentEntityType)
	documentVersionID := gid.New(scope.GetTenantID(), coredata.DocumentVersionEntityType)

	organization := &coredata.Organization{}

	document := &coredata.Document{
		ID:                    documentID,
		WriteMode:             coredata.DocumentWriteModeAuthored,
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
		Orientation:    coredata.DocumentVersionOrientationPortrait,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := organization.LoadByID(ctx, conn, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			document.OrganizationID = organization.ID

			if err := document.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert document: %w", err)
			}

			documentVersion.OrganizationID = organization.ID

			if err := documentVersion.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot create document version: %w", err)
			}

			if len(req.DefaultApproverIDs) > 0 {
				profiles := &coredata.MembershipProfiles{}
				if err := profiles.LoadByIDs(ctx, conn, scope, req.DefaultApproverIDs); err != nil {
					return fmt.Errorf("cannot load approver profiles: %w", err)
				}

				approvers := &coredata.DocumentDefaultApprovers{}
				if err := approvers.MergeByDocumentID(ctx, conn, scope, documentID, organization.ID, req.DefaultApproverIDs); err != nil {
					return fmt.Errorf("cannot set default approvers: %w", err)
				}
			}

			if err := s.emitDocumentEvent(ctx, scope, conn, documentID, coredata.WebhookEventTypeDocumentCreated, nil, nil, nil, nil); err != nil {
				return fmt.Errorf("cannot emit document created webhook: %w", err)
			}

			if err := s.emitDocumentEvent(
				ctx,
				scope,
				conn,
				documentVersion.DocumentID,
				coredata.WebhookEventTypeDocumentVersionCreated,
				documentVersion,
				nil,
				nil,
				nil,
			); err != nil {
				return fmt.Errorf("cannot emit document version created webhook: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *DocumentService) SignDocumentVersionByIdentity(
	ctx context.Context, scope coredata.Scoper,
	req SignDocumentVersionRequest,
) (*coredata.DocumentVersionSignature, error) {
	var (
		documentVersion          *coredata.DocumentVersion
		document                 *coredata.Document
		documentVersionSignature *coredata.DocumentVersionSignature
	)

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			documentVersion = &coredata.DocumentVersion{}
			if err := documentVersion.LoadByID(ctx, conn, scope, req.DocumentVersionID); err != nil {
				return fmt.Errorf("cannot get document version: %w", err)
			}

			if documentVersion.Status != coredata.DocumentVersionStatusPublished {
				return &ErrDocumentVersionNotPublished{}
			}

			document = &coredata.Document{}
			if err := document.LoadByID(ctx, conn, scope, documentVersion.DocumentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			profile := &coredata.MembershipProfile{}
			// FIXME: will be done differently
			if err := profile.LoadByIdentityIDAndOrganizationID(ctx, conn, scope, req.IdentityID, documentVersion.OrganizationID); err != nil {
				return fmt.Errorf("cannot find profile record for user email in organization %q: %w", documentVersion.OrganizationID, err)
			}

			documentVersionSignature = &coredata.DocumentVersionSignature{}
			if err := documentVersionSignature.LoadByDocumentVersionIDAndSignatory(ctx, conn, scope, req.DocumentVersionID, profile.ID); err != nil {
				return fmt.Errorf("cannot load document version signature: %w", err)
			}

			if documentVersionSignature.State == coredata.DocumentVersionSignatureStateSigned {
				return &ErrDocumentVersionSignatureAlreadySigned{}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	pdfData, err := s.ExportPDF(ctx, scope, req.DocumentVersionID, ExportPDFOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot export document PDF: %w", err)
	}

	fileRecord := &coredata.File{
		ID:             gid.New(scope.GetTenantID(), coredata.FileEntityType),
		OrganizationID: documentVersion.OrganizationID,
		BucketName:     s.svc.bucket,
		MimeType:       "application/pdf",
		FileName:       fmt.Sprintf("signature-%s.pdf", documentVersionSignature.ID),
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
			"type":         "signature-document",
			"signature-id": documentVersionSignature.ID.String(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot upload signature PDF: %w", err)
	}

	fileRecord.FileSize = fileSize

	signatureID := documentVersionSignature.ID

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			documentVersion = &coredata.DocumentVersion{}
			if err := documentVersion.LoadByID(ctx, tx, scope, req.DocumentVersionID); err != nil {
				return fmt.Errorf("cannot load document version: %w", err)
			}

			if documentVersion.Status != coredata.DocumentVersionStatusPublished {
				return &ErrDocumentVersionNotPublished{}
			}

			document = &coredata.Document{}
			if err := document.LoadByID(ctx, tx, scope, documentVersion.DocumentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			documentVersionSignature = &coredata.DocumentVersionSignature{}
			if err := documentVersionSignature.LoadByID(ctx, tx, scope, signatureID); err != nil {
				return fmt.Errorf("cannot load document version signature: %w", err)
			}

			if documentVersionSignature.State == coredata.DocumentVersionSignatureStateSigned {
				return &ErrDocumentVersionSignatureAlreadySigned{}
			}

			if err := fileRecord.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert signature file record: %w", err)
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
					ConsentText:    DocumentSignatureConsentText,
					EmailSubject:   fmt.Sprintf("Your signed %s - Certificate of Completion", document.Title),
				},
			)
			if err != nil {
				return fmt.Errorf("cannot create electronic signature: %w", err)
			}

			documentVersionSignature.State = coredata.DocumentVersionSignatureStateSigned
			documentVersionSignature.SignedAt = &now
			documentVersionSignature.ElectronicSignatureID = &esig.ID
			documentVersionSignature.UpdatedAt = now

			if err := documentVersionSignature.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update document version signature: %w", err)
			}

			if err := s.emitDocumentEvent(
				ctx,
				scope,
				tx,
				documentVersion.DocumentID,
				coredata.WebhookEventTypeDocumentVersionSignatureSigned,
				documentVersion,
				documentVersionSignature,
				nil,
				nil,
			); err != nil {
				return fmt.Errorf("cannot emit document version signature signed webhook: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return documentVersionSignature, nil
}

func (s *DocumentService) updateVersionInTx(
	ctx context.Context, scope coredata.Scoper,
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

	if err := draftVersion.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update document version: %w", err)
	}

	return nil
}

func (s *DocumentService) GetVersionSignature(
	ctx context.Context, scope coredata.Scoper,
	signatureID gid.GID,
) (*coredata.DocumentVersionSignature, error) {
	documentVersionSignature := &coredata.DocumentVersionSignature{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documentVersionSignature.LoadByID(ctx, conn, scope, signatureID)
		},
	)
	if err != nil {
		return nil, err
	}

	return documentVersionSignature, nil
}

func (s *DocumentService) BulkRequestSignatures(
	ctx context.Context, scope coredata.Scoper,
	req BulkRequestSignaturesRequest,
) ([]*coredata.DocumentVersionSignature, error) {
	var signatures []*coredata.DocumentVersionSignature

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			profiles := &coredata.MembershipProfiles{}
			if err := profiles.LoadByIDs(ctx, tx, scope, req.SignatoryIDs); err != nil {
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
				if err := documentVersion.LoadLatestVersion(ctx, tx, scope, documentID); err != nil {
					return fmt.Errorf("cannot load latest version for document %q: %w", documentID, err)
				}

				if documentVersion.Status != coredata.DocumentVersionStatusPublished {
					return &ErrDocumentVersionNotPublished{}
				}

				document := &coredata.Document{}
				if err := document.LoadByID(ctx, tx, scope, documentVersion.DocumentID); err != nil {
					return fmt.Errorf("cannot load document %q: %w", documentVersion.DocumentID, err)
				}

				for _, signatoryID := range req.SignatoryIDs {
					signature, created, err := s.createSignatureRequestInTx(ctx, scope, tx, documentVersion.ID, signatoryID)
					if err != nil {
						return fmt.Errorf("cannot create signature request for document %q and signatory %q: %w", documentID, signatoryID, err)
					}

					signatures = append(signatures, signature)

					if !created {
						continue
					}

					if err := s.emitLoadedDocumentEventInTx(
						ctx,
						scope,
						tx,
						document,
						coredata.WebhookEventTypeDocumentVersionSignatureRequested,
						documentVersion,
						signature,
						nil,
						nil,
					); err != nil {
						return fmt.Errorf("cannot emit signature requested webhook: %w", err)
					}
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
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	documentVersionID gid.GID,
	signatoryID gid.GID,
) (*coredata.DocumentVersionSignature, bool, error) {
	signatory := &coredata.MembershipProfile{}
	documentVersion := &coredata.DocumentVersion{}

	if err := documentVersion.LoadByID(ctx, tx, scope, documentVersionID); err != nil {
		return nil, false, fmt.Errorf("cannot load document version: %w", err)
	}

	if err := signatory.LoadByID(ctx, tx, scope, signatoryID); err != nil {
		return nil, false, fmt.Errorf("cannot load signatory: %w", err)
	}

	// A signature applies to the whole major version: minor publishes keep it
	// and the export unions signatures across every minor of the major, so a
	// signatory must have at most one signature per major. If one already
	// exists anywhere in this major (requested or signed), reuse it instead of
	// inserting a duplicate.
	existingSignature := &coredata.DocumentVersionSignature{}

	err := existingSignature.LoadByDocumentMajorAndSignatory(ctx, tx, scope, documentVersionID, signatoryID)
	if err == nil {
		return existingSignature, false, nil
	}

	if !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, false, fmt.Errorf("cannot load existing signature for signatory: %w", err)
	}

	documentVersionSignatureID := gid.New(scope.GetTenantID(), coredata.DocumentVersionSignatureEntityType)
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

	if err := documentVersionSignature.Insert(ctx, tx, scope); err != nil {
		return nil, false, fmt.Errorf("cannot insert document version signature: %w", err)
	}

	return documentVersionSignature, true, nil
}

func (s *DocumentService) RequestSignature(
	ctx context.Context, scope coredata.Scoper,
	req RequestSignatureRequest,
) (*coredata.DocumentVersionSignature, error) {
	var signature *coredata.DocumentVersionSignature

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

			if documentVersion.Status != coredata.DocumentVersionStatusPublished {
				return &ErrDocumentVersionNotPublished{}
			}

			if document.CurrentPublishedMajor == nil ||
				document.CurrentPublishedMinor == nil ||
				documentVersion.Major != *document.CurrentPublishedMajor ||
				documentVersion.Minor != *document.CurrentPublishedMinor {
				return &ErrDocumentVersionNotCurrent{}
			}

			profile := &coredata.MembershipProfile{}
			if err := profile.LoadByID(ctx, tx, scope, req.Signatory); err != nil {
				return fmt.Errorf("cannot load signatory profile: %w", err)
			}

			if profile.ContractEndDate != nil && profile.ContractEndDate.Before(time.Now()) {
				return &ErrProfileContractEnded{ProfileID: profile.ID}
			}

			var (
				err     error
				created bool
			)

			signature, created, err = s.createSignatureRequestInTx(ctx, scope, tx, req.DocumentVersionID, req.Signatory)
			if err != nil {
				return fmt.Errorf("cannot create signature request: %w", err)
			}

			if !created {
				return nil
			}

			if err := s.emitDocumentEvent(
				ctx,
				scope,
				tx,
				documentVersion.DocumentID,
				coredata.WebhookEventTypeDocumentVersionSignatureRequested,
				documentVersion,
				signature,
				nil,
				nil,
			); err != nil {
				return fmt.Errorf("cannot emit document version signature requested webhook: %w", err)
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
	ctx context.Context, scope coredata.Scoper,
	documentVersionID gid.GID,
	cursor *page.Cursor[coredata.DocumentVersionSignatureOrderField],
	filter *coredata.DocumentVersionSignatureFilter,
) (*page.Page[*coredata.DocumentVersionSignature, coredata.DocumentVersionSignatureOrderField], error) {
	var documentVersionSignatures coredata.DocumentVersionSignatures

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documentVersionSignatures.LoadByDocumentVersionID(ctx, conn, scope, documentVersionID, cursor, filter)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(documentVersionSignatures, cursor), nil
}

func (s *DocumentService) IsVersionSignedByUserEmail(
	ctx context.Context, scope coredata.Scoper,
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
				scope,
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
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	document *coredata.Document,
	latestVersion *coredata.DocumentVersion,
) (*coredata.DocumentVersion, error) {
	now := time.Now()

	draftVersion := &coredata.DocumentVersion{
		ID:             gid.New(scope.GetTenantID(), coredata.DocumentVersionEntityType),
		OrganizationID: document.OrganizationID,
		DocumentID:     document.ID,
		Title:          latestVersion.Title,
		Major:          latestVersion.Major,
		Minor:          latestVersion.Minor + 1,
		Classification: latestVersion.Classification,
		DocumentType:   latestVersion.DocumentType,
		Content:        latestVersion.Content,
		Orientation:    latestVersion.Orientation,
		Status:         coredata.DocumentVersionStatusDraft,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := draftVersion.Insert(ctx, tx, scope); err != nil {
		return nil, fmt.Errorf("cannot create draft: %w", err)
	}

	return draftVersion, nil
}

func (s *DocumentService) deleteDraftInTx(
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	draftVersion *coredata.DocumentVersion,
) error {
	if err := draftVersion.Delete(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot delete document version: %w", err)
	}

	return nil
}

// For deletion events this must be called before the document is soft-deleted,
// since Document.LoadByID filters out soft-deleted rows.
func (s *DocumentService) emitDocumentEvent(
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	documentID gid.GID,
	eventType coredata.WebhookEventType,
	version *coredata.DocumentVersion,
	signature *coredata.DocumentVersionSignature,
	quorumID *gid.GID,
	updatedFrom any,
) error {
	document := &coredata.Document{}
	if err := document.LoadByID(ctx, tx, scope, documentID); err != nil {
		return fmt.Errorf("cannot load document for %q webhook: %w", eventType, err)
	}

	return s.emitLoadedDocumentEventInTx(ctx, scope, tx, document, eventType, version, signature, quorumID, updatedFrom)
}

func (s *DocumentService) emitLoadedDocumentEventInTx(
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	document *coredata.Document,
	eventType coredata.WebhookEventType,
	version *coredata.DocumentVersion,
	signature *coredata.DocumentVersionSignature,
	quorumID *gid.GID,
	updatedFrom any,
) error {
	subscriptions := coredata.WebhookSubscriptions{}

	exists, err := subscriptions.ExistsByOrganizationIDAndEventType(ctx, tx, scope, document.OrganizationID, eventType)
	if err != nil {
		return fmt.Errorf("cannot check webhook subscriptions for %q: %w", eventType, err)
	}

	if !exists {
		return nil
	}

	var payload any

	switch {
	case signature != nil:
		payload = webhooktypes.NewDocumentVersionSignature(signature, version, document)
	case quorumID != nil:
		payload, err = s.loadDocumentApprovalQuorumForWebhook(ctx, scope, tx, *quorumID, version, document)
		if err != nil {
			return fmt.Errorf("cannot build approval quorum payload for %q webhook: %w", eventType, err)
		}
	case version != nil:
		payload = webhooktypes.NewDocumentVersion(version, document)
	default:
		payload = webhooktypes.NewDocument(document)
	}

	if err := webhook.InsertUpdateData(
		ctx,
		tx,
		scope,
		document.OrganizationID,
		eventType,
		payload,
		updatedFrom,
	); err != nil {
		return fmt.Errorf("cannot insert %q webhook event: %w", eventType, err)
	}

	return nil
}

func (s *DocumentService) emitDocumentLifecycleEventsInTx(
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	documentIDs []gid.GID,
	eventType coredata.WebhookEventType,
) error {
	if len(documentIDs) == 0 {
		return nil
	}

	documents := coredata.Documents{}
	if err := documents.LoadByIDs(ctx, tx, scope, documentIDs); err != nil {
		return fmt.Errorf("cannot load documents for %q webhook: %w", eventType, err)
	}

	subscribed := make(map[gid.GID]bool)

	for _, document := range documents {
		hasSubscription, cached := subscribed[document.OrganizationID]
		if !cached {
			subscriptions := coredata.WebhookSubscriptions{}

			exists, err := subscriptions.ExistsByOrganizationIDAndEventType(ctx, tx, scope, document.OrganizationID, eventType)
			if err != nil {
				return fmt.Errorf("cannot check webhook subscriptions for %q: %w", eventType, err)
			}

			hasSubscription = exists
			subscribed[document.OrganizationID] = exists
		}

		if !hasSubscription {
			continue
		}

		if err := webhook.InsertData(
			ctx,
			tx,
			scope,
			document.OrganizationID,
			eventType,
			webhooktypes.NewDocument(document),
		); err != nil {
			return fmt.Errorf("cannot insert %q webhook event: %w", eventType, err)
		}
	}

	return nil
}

func (s *DocumentService) loadDocumentApprovalQuorumForWebhook(
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	quorumID gid.GID,
	version *coredata.DocumentVersion,
	document *coredata.Document,
) (*webhooktypes.DocumentApprovalQuorum, error) {
	quorum := &coredata.DocumentVersionApprovalQuorum{}
	if err := quorum.LoadByID(ctx, tx, scope, quorumID); err != nil {
		return nil, fmt.Errorf("cannot load approval quorum for webhook: %w", err)
	}

	decisions, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.DocumentVersionApprovalDecisionOrderField]{
			Field:     coredata.DocumentVersionApprovalDecisionOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(
			ctx context.Context,
			cursor *page.Cursor[coredata.DocumentVersionApprovalDecisionOrderField],
		) ([]*coredata.DocumentVersionApprovalDecision, error) {
			var batch coredata.DocumentVersionApprovalDecisions
			if err := batch.LoadByQuorumID(ctx, tx, scope, quorumID, cursor, coredata.NewDocumentVersionApprovalDecisionFilter(nil)); err != nil {
				return nil, fmt.Errorf("cannot load approval decisions: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load approval decisions for webhook: %w", err)
	}

	return webhooktypes.NewDocumentApprovalQuorum(quorum, decisions, version, document), nil
}

func (s *DocumentService) SoftDelete(
	ctx context.Context, scope coredata.Scoper,
	documentID gid.GID,
) error {
	document := coredata.Document{ID: documentID}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := s.emitDocumentEvent(ctx, scope, tx, documentID, coredata.WebhookEventTypeDocumentDeleted, nil, nil, nil, nil); err != nil {
				return fmt.Errorf("cannot emit document deleted webhook: %w", err)
			}

			if err := s.clearDocumentReferences(ctx, scope, tx, []gid.GID{documentID}); err != nil {
				return err
			}

			return document.SoftDelete(ctx, tx, scope)
		},
	)
}

func (s *DocumentService) BulkSoftDelete(
	ctx context.Context, scope coredata.Scoper,
	documentIDs []gid.GID,
) error {
	documents := coredata.Documents{}

	for _, documentID := range documentIDs {
		documents = append(documents, &coredata.Document{ID: documentID})
	}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := s.emitDocumentLifecycleEventsInTx(ctx, scope, tx, documentIDs, coredata.WebhookEventTypeDocumentDeleted); err != nil {
				return fmt.Errorf("cannot emit document deleted webhooks: %w", err)
			}

			if err := s.clearDocumentReferences(ctx, scope, tx, documentIDs); err != nil {
				return err
			}

			return documents.BulkSoftDelete(ctx, tx, scope)
		},
	)
}

func (s *DocumentService) BulkArchive(
	ctx context.Context, scope coredata.Scoper,
	documentIDs []gid.GID,
) error {
	documents := coredata.Documents{}

	for _, documentID := range documentIDs {
		documents = append(documents, &coredata.Document{ID: documentID})
	}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			for _, documentID := range documentIDs {
				if err := s.teardownDocumentWorkflowsInTx(ctx, scope, tx, documentID); err != nil {
					return err
				}
			}

			controlDocument := coredata.ControlDocument{}
			if err := controlDocument.DeleteByDocumentIDs(ctx, tx, scope, documentIDs); err != nil {
				return fmt.Errorf("cannot delete control mappings: %w", err)
			}

			riskDocument := coredata.RiskDocument{}
			if err := riskDocument.DeleteByDocumentIDs(ctx, tx, scope, documentIDs); err != nil {
				return fmt.Errorf("cannot delete risk mappings: %w", err)
			}

			measureDocument := coredata.MeasureDocument{}
			if err := measureDocument.DeleteByDocumentIDs(ctx, tx, scope, documentIDs); err != nil {
				return fmt.Errorf("cannot delete measure mappings: %w", err)
			}

			if err := s.clearDocumentReferences(ctx, scope, tx, documentIDs); err != nil {
				return err
			}

			if err := documents.BulkArchive(ctx, tx, scope); err != nil {
				return err
			}

			if err := s.emitDocumentLifecycleEventsInTx(ctx, scope, tx, documentIDs, coredata.WebhookEventTypeDocumentArchived); err != nil {
				return fmt.Errorf("cannot emit document archived webhooks: %w", err)
			}

			return nil
		},
	)
}

func (s *DocumentService) BulkUnarchive(
	ctx context.Context, scope coredata.Scoper,
	documentIDs []gid.GID,
) error {
	documents := coredata.Documents{}

	for _, documentID := range documentIDs {
		documents = append(documents, &coredata.Document{ID: documentID})
	}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			for _, documentID := range documentIDs {
				if err := s.ensureDraftReadyOnUnarchiveInTx(ctx, scope, tx, documentID); err != nil {
					return err
				}
			}

			if err := documents.BulkUnarchive(ctx, tx, scope); err != nil {
				return err
			}

			if err := s.emitDocumentLifecycleEventsInTx(ctx, scope, tx, documentIDs, coredata.WebhookEventTypeDocumentUnarchived); err != nil {
				return fmt.Errorf("cannot emit document unarchived webhooks: %w", err)
			}

			return nil
		},
	)
}

// clearDocumentReferences nullifies references to the given document IDs in
// generated_documents and statements_of_applicability. This must be called
// inside a transaction before soft-deleting or archiving documents, because
// those operations are UPDATEs and do not trigger ON DELETE SET NULL.
func (s *DocumentService) clearDocumentReferences(
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	documentIDs []gid.GID,
) error {
	datum := coredata.Datum{}
	if err := datum.ClearGeneratedDocumentID(ctx, tx, documentIDs); err != nil {
		return err
	}

	asset := coredata.Asset{}
	if err := asset.ClearGeneratedDocumentID(ctx, tx, documentIDs); err != nil {
		return err
	}

	finding := coredata.Finding{}
	if err := finding.ClearGeneratedDocumentID(ctx, tx, documentIDs); err != nil {
		return err
	}

	obligation := coredata.Obligation{}
	if err := obligation.ClearGeneratedDocumentID(ctx, tx, documentIDs); err != nil {
		return err
	}

	soa := coredata.StatementOfApplicability{}
	if err := soa.ClearDocumentIDByDocumentIDs(ctx, tx, documentIDs); err != nil {
		return err
	}

	return nil
}

func (s *DocumentService) teardownDocumentWorkflowsInTx(
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	documentID gid.GID,
) error {
	document := &coredata.Document{}
	if err := document.LoadByID(ctx, tx, scope, documentID); err != nil {
		return fmt.Errorf("cannot load document %q: %w", documentID, err)
	}

	documentVersion := &coredata.DocumentVersion{}
	if err := documentVersion.LoadLatestVersion(ctx, tx, scope, documentID); err != nil {
		return fmt.Errorf("cannot load latest document version: %w", err)
	}

	if documentVersion.Status == coredata.DocumentVersionStatusPendingApproval {
		quorum := &coredata.DocumentVersionApprovalQuorum{}
		if err := quorum.LoadLastByDocumentVersionID(ctx, tx, scope, documentVersion.ID); err != nil {
			if !errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load approval quorum: %w", err)
			}
		} else if quorum.Status == coredata.DocumentVersionApprovalQuorumStatusPending {
			if err := s.svc.DocumentApprovals.voidApprovalInTx(ctx, scope, tx, document, documentVersion, quorum); err != nil {
				return err
			}
		}
	}

	signatures := &coredata.DocumentVersionSignatures{}
	if err := signatures.LoadRequestedByDocumentID(ctx, tx, scope, documentID); err != nil {
		return fmt.Errorf("cannot load requested document version signatures: %w", err)
	}

	for _, signature := range *signatures {
		version := &coredata.DocumentVersion{}
		if err := version.LoadByID(ctx, tx, scope, signature.DocumentVersionID); err != nil {
			return fmt.Errorf("cannot load document version: %w", err)
		}

		if err := signature.Delete(ctx, tx, scope, signature.ID); err != nil {
			return fmt.Errorf("cannot delete document version signature: %w", err)
		}

		if err := s.emitDocumentEvent(
			ctx,
			scope,
			tx,
			documentID,
			coredata.WebhookEventTypeDocumentVersionSignatureCancelled,
			version,
			signature,
			nil,
			nil,
		); err != nil {
			return fmt.Errorf("cannot emit document version signature cancelled webhook: %w", err)
		}
	}

	return nil
}

func (s *DocumentService) ensureDraftReadyOnUnarchiveInTx(
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	documentID gid.GID,
) error {
	document := &coredata.Document{}
	if err := document.LoadByID(ctx, tx, scope, documentID); err != nil {
		return fmt.Errorf("cannot load document %q: %w", documentID, err)
	}

	documentVersion := &coredata.DocumentVersion{}
	if err := documentVersion.LoadLatestVersion(ctx, tx, scope, documentID); err != nil {
		return fmt.Errorf("cannot load latest document version: %w", err)
	}

	if documentVersion.Status != coredata.DocumentVersionStatusPendingApproval {
		return nil
	}

	quorum := &coredata.DocumentVersionApprovalQuorum{}
	if err := quorum.LoadLastByDocumentVersionID(ctx, tx, scope, documentVersion.ID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil
		}

		return fmt.Errorf("cannot load approval quorum: %w", err)
	}

	if quorum.Status != coredata.DocumentVersionApprovalQuorumStatusPending {
		return nil
	}

	return s.svc.DocumentApprovals.voidApprovalInTx(ctx, scope, tx, document, documentVersion, quorum)
}

func (s *DocumentService) RequestExport(
	ctx context.Context, scope coredata.Scoper,
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
			if err := document.LoadByID(ctx, conn, scope, documentID); err != nil {
				return fmt.Errorf("cannot load document %q: %w", documentID, err)
			}

			organizationID = document.OrganizationID
		}

		now := time.Now()
		exportJobID = gid.New(scope.GetTenantID(), coredata.ExportJobEntityType)

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

		if err := exportJob.Insert(ctx, conn, scope); err != nil {
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
	ctx context.Context, scope coredata.Scoper,
	documentID gid.GID,
	filter *coredata.DocumentVersionFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			documentVersions := &coredata.DocumentVersions{}
			count, err = documentVersions.CountByDocumentID(ctx, conn, scope, documentID, filter)

			return err
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *DocumentService) CountSignaturesForVersionID(
	ctx context.Context, scope coredata.Scoper,
	documentVersionID gid.GID,
	filter *coredata.DocumentVersionSignatureFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			documentVersionSignatures := &coredata.DocumentVersionSignatures{}
			count, err = documentVersionSignatures.CountByDocumentVersionID(ctx, conn, scope, documentVersionID, filter)

			return err
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *DocumentService) ListVersions(
	ctx context.Context, scope coredata.Scoper,
	documentID gid.GID,
	cursor *page.Cursor[coredata.DocumentVersionOrderField],
	filter *coredata.DocumentVersionFilter,
) (*page.Page[*coredata.DocumentVersion, coredata.DocumentVersionOrderField], error) {
	var documentVersions coredata.DocumentVersions

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := documentVersions.LoadByDocumentID(ctx, conn, scope, documentID, cursor, filter)
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
	ctx context.Context, scope coredata.Scoper,
	documentVersionID gid.GID,
) (*coredata.DocumentVersion, error) {
	documentVersion := &coredata.DocumentVersion{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documentVersion.LoadByID(ctx, conn, scope, documentVersionID)
		},
	)
	if err != nil {
		return nil, err
	}

	return documentVersion, nil
}

func (s *DocumentService) IsSigned(
	ctx context.Context, scope coredata.Scoper,
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
				scope,
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
	ctx context.Context, scope coredata.Scoper,
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
				scope,
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
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	filter *coredata.DocumentFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			documents := &coredata.Documents{}

			count, err = documents.CountByOrganizationID(ctx, conn, scope, organizationID, filter)
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
	ctx context.Context, scope coredata.Scoper,
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
				scope,
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
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
	filter *coredata.DocumentFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			documents := &coredata.Documents{}

			count, err = documents.CountByControlID(ctx, conn, scope, controlID, filter)
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
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
	cursor *page.Cursor[coredata.DocumentOrderField],
	filter *coredata.DocumentFilter,
) (*page.Page[*coredata.Document, coredata.DocumentOrderField], error) {
	var documents coredata.Documents

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documents.LoadByControlID(ctx, conn, scope, controlID, cursor, filter)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(documents, cursor), nil
}

func (s *DocumentService) CountForRiskID(
	ctx context.Context, scope coredata.Scoper,
	riskID gid.GID,
	filter *coredata.DocumentFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			documents := &coredata.Documents{}

			count, err = documents.CountByRiskID(ctx, conn, scope, riskID, filter)
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
	ctx context.Context, scope coredata.Scoper,
	riskID gid.GID,
	cursor *page.Cursor[coredata.DocumentOrderField],
	filter *coredata.DocumentFilter,
) (*page.Page[*coredata.Document, coredata.DocumentOrderField], error) {
	var documents coredata.Documents

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return documents.LoadByRiskID(ctx, conn, scope, riskID, cursor, filter)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(documents, cursor), nil
}

func (s *DocumentService) CountForMeasureID(
	ctx context.Context, scope coredata.Scoper,
	measureID gid.GID,
	filter *coredata.DocumentFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			documents := &coredata.Documents{}

			count, err = documents.CountByMeasureID(ctx, conn, scope, measureID, filter)
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
	ctx context.Context, scope coredata.Scoper,
	measureID gid.GID,
	cursor *page.Cursor[coredata.DocumentOrderField],
	filter *coredata.DocumentFilter,
) (*page.Page[*coredata.Document, coredata.DocumentOrderField], error) {
	var documents coredata.Documents

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := documents.LoadByMeasureID(ctx, conn, scope, measureID, cursor, filter); err != nil {
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
	ctx context.Context, scope coredata.Scoper,
	req UpdateDocumentRequest,
) (*coredata.Document, *coredata.DocumentVersion, bool, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, false, err
	}

	document := &coredata.Document{}

	var (
		resultVersion *coredata.DocumentVersion
		draftCreated  bool
	)

	now := time.Now()

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := document.LoadByID(ctx, tx, scope, req.DocumentID); err != nil {
				return fmt.Errorf("cannot load document %q: %w", req.DocumentID, err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			previousDocument := *document

			if req.TrustCenterVisibility != nil {
				document.TrustCenterVisibility = *req.TrustCenterVisibility
			}

			document.UpdatedAt = now

			if err := document.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update document: %w", err)
			}

			latestVersion := &coredata.DocumentVersion{}
			if err := latestVersion.LoadLatestVersion(ctx, tx, scope, req.DocumentID); err != nil {
				return fmt.Errorf("cannot load latest version: %w", err)
			}

			hasVersionChanges := req.Title != nil || req.Content != nil || req.Classification != nil || req.DocumentType != nil
			docLevelChanged := req.TrustCenterVisibility != nil || req.DefaultApproverIDs != nil

			if req.Content != nil && document.WriteMode == coredata.DocumentWriteModeGenerated {
				return &ErrDocumentVersionGenerated{}
			}

			if req.DefaultApproverIDs != nil {
				if len(*req.DefaultApproverIDs) > 0 {
					profiles := &coredata.MembershipProfiles{}
					if err := profiles.LoadByIDs(ctx, tx, scope, *req.DefaultApproverIDs); err != nil {
						return fmt.Errorf("cannot load approver profiles: %w", err)
					}
				}

				defaultApprovers := &coredata.DocumentDefaultApprovers{}
				if err := defaultApprovers.MergeByDocumentID(ctx, tx, scope, req.DocumentID, document.OrganizationID, *req.DefaultApproverIDs); err != nil {
					return fmt.Errorf("cannot update default approvers: %w", err)
				}
			}

			versionDeleted := false

			var versionPrevious any

			if hasVersionChanges {
				if latestVersion.Status == coredata.DocumentVersionStatusDraft {
					previousVersion := *latestVersion
					versionPrevious = webhooktypes.NewDocumentVersion(&previousVersion, &previousDocument)

					if err := s.updateVersionInTx(ctx, scope, tx, latestVersion, req.Content, req.Classification, req.DocumentType, req.Title); err != nil {
						return err
					}

					if document.CurrentPublishedMajor != nil && (latestVersion.Major != 0 || latestVersion.Minor != 1) {
						publishedVersion := &coredata.DocumentVersion{}
						if err := publishedVersion.LoadByDocumentIDAndVersion(
							ctx,
							tx,
							scope,
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
							if err := s.deleteDraftInTx(ctx, scope, tx, latestVersion); err != nil {
								return err
							}

							resultVersion = nil
							versionDeleted = true
						}
					}

					if !versionDeleted {
						resultVersion = latestVersion
					}
				} else {
					draftVersion, err := s.createDraftInTx(ctx, scope, tx, document, latestVersion)
					if err != nil {
						return err
					}

					if err := s.updateVersionInTx(ctx, scope, tx, draftVersion, req.Content, req.Classification, req.DocumentType, req.Title); err != nil {
						return err
					}

					resultVersion = draftVersion
					draftCreated = true
				}

				if versionDeleted {
					if err := s.emitDocumentEvent(
						ctx,
						scope,
						tx,
						latestVersion.DocumentID,
						coredata.WebhookEventTypeDocumentVersionDeleted,
						latestVersion,
						nil,
						nil,
						nil,
					); err != nil {
						return fmt.Errorf("cannot emit document version deleted webhook: %w", err)
					}
				} else {
					versionEvent := coredata.WebhookEventTypeDocumentVersionUpdated
					if draftCreated {
						versionEvent = coredata.WebhookEventTypeDocumentVersionCreated
						versionPrevious = nil
					}

					if err := s.emitDocumentEvent(ctx, scope, tx, resultVersion.DocumentID, versionEvent, resultVersion, nil, nil, versionPrevious); err != nil {
						return fmt.Errorf("cannot emit document version webhook: %w", err)
					}
				}
			}

			if docLevelChanged {
				if err := s.emitDocumentEvent(ctx, scope, tx, req.DocumentID, coredata.WebhookEventTypeDocumentUpdated, nil, nil, nil, webhooktypes.NewDocument(&previousDocument)); err != nil {
					return fmt.Errorf("cannot emit document updated webhook: %w", err)
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
	ctx context.Context, scope coredata.Scoper,
	documentID gid.GID,
) (*coredata.Document, error) {
	document := &coredata.Document{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := document.LoadByID(ctx, tx, scope, documentID); err != nil {
				return fmt.Errorf("cannot load document %q: %w", documentID, err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			latestVersion := &coredata.DocumentVersion{}
			if err := latestVersion.LoadLatestVersion(ctx, tx, scope, documentID); err != nil {
				return fmt.Errorf("cannot load latest version: %w", err)
			}

			if latestVersion.Status != coredata.DocumentVersionStatusDraft {
				return &ErrDocumentDraftNotDeletable{}
			}

			if latestVersion.Major == 0 && latestVersion.Minor == 1 {
				return &ErrDocumentDraftNotDeletable{}
			}

			if err := s.deleteDraftInTx(ctx, scope, tx, latestVersion); err != nil {
				return err
			}

			if err := s.emitDocumentEvent(
				ctx,
				scope,
				tx,
				latestVersion.DocumentID,
				coredata.WebhookEventTypeDocumentVersionDeleted,
				latestVersion,
				nil,
				nil,
				nil,
			); err != nil {
				return fmt.Errorf("cannot emit document version deleted webhook: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return document, nil
}

func (s *DocumentService) Archive(
	ctx context.Context, scope coredata.Scoper,
	documentID gid.GID,
) (*coredata.Document, error) {
	document := &coredata.Document{}
	now := time.Now()

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := document.LoadByID(ctx, tx, scope, documentID); err != nil {
				return fmt.Errorf("cannot load document %q: %w", documentID, err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			if err := s.teardownDocumentWorkflowsInTx(ctx, scope, tx, documentID); err != nil {
				return err
			}

			controlDocument := coredata.ControlDocument{}
			if err := controlDocument.DeleteByDocumentIDs(ctx, tx, scope, []gid.GID{documentID}); err != nil {
				return fmt.Errorf("cannot delete control mappings: %w", err)
			}

			riskDocument := coredata.RiskDocument{}
			if err := riskDocument.DeleteByDocumentIDs(ctx, tx, scope, []gid.GID{documentID}); err != nil {
				return fmt.Errorf("cannot delete risk mappings: %w", err)
			}

			measureDocument := coredata.MeasureDocument{}
			if err := measureDocument.DeleteByDocumentIDs(ctx, tx, scope, []gid.GID{documentID}); err != nil {
				return fmt.Errorf("cannot delete measure mappings: %w", err)
			}

			if err := s.clearDocumentReferences(ctx, scope, tx, []gid.GID{documentID}); err != nil {
				return err
			}

			document.Status = coredata.DocumentStatusArchived
			document.ArchivedAt = &now
			document.UpdatedAt = now
			document.TrustCenterVisibility = coredata.TrustCenterVisibilityNone

			if err := document.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot archive document: %w", err)
			}

			if err := s.emitDocumentEvent(ctx, scope, tx, documentID, coredata.WebhookEventTypeDocumentArchived, nil, nil, nil, nil); err != nil {
				return fmt.Errorf("cannot emit document archived webhook: %w", err)
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
	ctx context.Context, scope coredata.Scoper,
	documentID gid.GID,
) (*coredata.Document, error) {
	document := &coredata.Document{}
	now := time.Now()

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := document.LoadByID(ctx, tx, scope, documentID); err != nil {
				return fmt.Errorf("cannot load document %q: %w", documentID, err)
			}

			if document.ArchivedAt == nil {
				return &ErrDocumentNotArchived{}
			}

			if err := s.ensureDraftReadyOnUnarchiveInTx(ctx, scope, tx, documentID); err != nil {
				return err
			}

			document.Status = coredata.DocumentStatusActive
			document.ArchivedAt = nil
			document.UpdatedAt = now

			if err := document.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot unarchive document: %w", err)
			}

			if err := s.emitDocumentEvent(ctx, scope, tx, documentID, coredata.WebhookEventTypeDocumentUnarchived, nil, nil, nil, nil); err != nil {
				return fmt.Errorf("cannot emit document unarchived webhook: %w", err)
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
	ctx context.Context, scope coredata.Scoper,
	documentVersionSignatureID gid.GID,
) error {
	documentVersionSignature := &coredata.DocumentVersionSignature{}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := documentVersionSignature.LoadByID(ctx, tx, scope, documentVersionSignatureID); err != nil {
				return fmt.Errorf("cannot load document version signature: %w", err)
			}

			documentVersion := &coredata.DocumentVersion{}
			if err := documentVersion.LoadByID(ctx, tx, scope, documentVersionSignature.DocumentVersionID); err != nil {
				return fmt.Errorf("cannot load document version: %w", err)
			}

			document := &coredata.Document{}
			if err := document.LoadByID(ctx, tx, scope, documentVersion.DocumentID); err != nil {
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

			if err := documentVersionSignature.Delete(ctx, tx, scope, documentVersionSignatureID); err != nil {
				return fmt.Errorf("cannot delete document version signature: %w", err)
			}

			if err := s.emitDocumentEvent(
				ctx,
				scope,
				tx,
				documentVersion.DocumentID,
				coredata.WebhookEventTypeDocumentVersionSignatureCancelled,
				documentVersion,
				documentVersionSignature,
				nil,
				nil,
			); err != nil {
				return fmt.Errorf("cannot emit document version signature cancelled webhook: %w", err)
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
	ctx context.Context, scope coredata.Scoper,
	documentVersionID gid.GID,
	options ExportPDFOptions,
) ([]byte, error) {
	var data []byte

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) (err error) {
			data, err = exportDocumentPDF(ctx, s.svc, s.html2pdfConverter, conn, scope, documentVersionID, options)
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

func (s *DocumentService) BuildAndUploadExport(ctx context.Context, scope coredata.Scoper, exportJobID gid.GID) (*coredata.ExportJob, error) {
	exportJob := &coredata.ExportJob{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := exportJob.LoadByID(ctx, tx, scope, exportJobID); err != nil {
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
			if err := firstDocument.LoadByID(ctx, tx, scope, documentIDs[0]); err != nil {
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

			err = s.Export(ctx, scope, documentIDs, tempFile, exportOptions)
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
				ID:             gid.New(exportJob.ID.TenantID(), coredata.FileEntityType),
				OrganizationID: organizationID,
				BucketName:     s.svc.bucket,
				MimeType:       "application/zip",
				FileName:       fmt.Sprintf("Documents Export %s.zip", now.Format("2006-01-02")),
				FileKey:        uuid.String(),
				FileSize:       fileInfo.Size(),
				Visibility:     coredata.FileVisibilityPrivate,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := file.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert file: %w", err)
			}

			exportJob.FileID = &file.ID
			if err := exportJob.Update(ctx, tx, scope); err != nil {
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
	svc *Service,
	html2pdfConverter *html2pdf.Converter,
	conn pg.Querier,
	scope coredata.Scoper,
	documentVersionID gid.GID,
	options ExportPDFOptions,
) ([]byte, error) {
	version := &coredata.DocumentVersion{}
	if err := version.LoadByID(ctx, conn, scope, documentVersionID); err != nil {
		return nil, fmt.Errorf("cannot load document version: %w", err)
	}

	// Published versions with a stored PDF: use the stored file,
	// append signature page and watermark as needed.
	if version.FileID != nil {
		return exportStoredPDF(ctx, svc, html2pdfConverter, conn, scope, version, options)
	}

	// No stored PDF: generate on the fly without watermark — watermark is
	// applied after merging the signature page so all pages are watermarked.
	generateOptions := options
	generateOptions.WithWatermark = false
	generateOptions.WatermarkEmail = nil

	pdfData, err := generateDocumentPDF(ctx, svc, html2pdfConverter, conn, scope, version, generateOptions)
	if err != nil {
		return nil, err
	}

	if options.WithSignatures {
		signaturePagePDF, err := generateSignaturePagePDF(ctx, svc, html2pdfConverter, conn, scope, version)
		if err != nil {
			return nil, fmt.Errorf("cannot generate signature page: %w", err)
		}

		if signaturePagePDF != nil {
			pdfData, err = pdfutils.MergePDFs(pdfData, signaturePagePDF)
			if err != nil {
				return nil, fmt.Errorf("cannot merge signature page: %w", err)
			}
		}
	}

	if options.WithWatermark {
		if options.WatermarkEmail == nil {
			return nil, fmt.Errorf("watermark email is required with watermark enabled")
		}

		pdfData, err = pdfutils.AddConfidentialWithTimestamp(pdfData, *options.WatermarkEmail)
		if err != nil {
			return nil, fmt.Errorf("cannot add watermark to PDF: %w", err)
		}
	}

	return pdfData, nil
}

func exportStoredPDF(
	ctx context.Context,
	svc *Service,
	html2pdfConverter *html2pdf.Converter,
	conn pg.Querier,
	scope coredata.Scoper,
	version *coredata.DocumentVersion,
	options ExportPDFOptions,
) ([]byte, error) {
	fileRecord := &coredata.File{}
	if err := fileRecord.LoadByID(ctx, conn, scope, *version.FileID); err != nil {
		return nil, fmt.Errorf("cannot load document version file: %w", err)
	}

	pdfData, err := svc.fileManager.GetFileBytes(ctx, fileRecord)
	if err != nil {
		return nil, fmt.Errorf("cannot download document version PDF: %w", err)
	}

	if options.WithSignatures {
		signaturePagePDF, err := generateSignaturePagePDF(ctx, svc, html2pdfConverter, conn, scope, version)
		if err != nil {
			return nil, fmt.Errorf("cannot generate signature page: %w", err)
		}

		if signaturePagePDF != nil {
			pdfData, err = pdfutils.MergePDFs(pdfData, signaturePagePDF)
			if err != nil {
				return nil, fmt.Errorf("cannot merge signature page: %w", err)
			}
		}
	}

	if options.WithWatermark {
		if options.WatermarkEmail == nil {
			return nil, fmt.Errorf("watermark email is required with watermark enabled")
		}

		pdfData, err = pdfutils.AddConfidentialWithTimestamp(pdfData, *options.WatermarkEmail)
		if err != nil {
			return nil, fmt.Errorf("cannot add watermark to PDF: %w", err)
		}
	}

	return pdfData, nil
}

func generateSignaturePagePDF(
	ctx context.Context,
	svc *Service,
	html2pdfConverter *html2pdf.Converter,
	conn pg.Querier,
	scope coredata.Scoper,
	version *coredata.DocumentVersion,
) ([]byte, error) {
	signaturesWithPeople := &coredata.DocumentVersionSignaturesWithPeople{}
	if err := signaturesWithPeople.LoadByDocumentVersionIDWithPeople(ctx, conn, scope, version.ID, 1_000); err != nil {
		return nil, fmt.Errorf("cannot load document version signatures: %w", err)
	}

	if len(*signaturesWithPeople) == 0 {
		return nil, nil
	}

	signatureData := make([]docgen.SignatureData, len(*signaturesWithPeople))
	for i, sig := range *signaturesWithPeople {
		signatureData[i] = docgen.SignatureData{
			SignedBy:    sig.SignedByFullName,
			SignedAt:    sig.SignedAt,
			State:       sig.State,
			RequestedAt: sig.RequestedAt,
		}
	}

	isLandscape := version.Orientation == coredata.DocumentVersionOrientationLandscape

	htmlContent, err := docgen.RenderSignaturePageHTML(docgen.SignaturePageData{
		Signatures: signatureData,
		Landscape:  isLandscape,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot render signature page HTML: %w", err)
	}

	orientation := html2pdf.OrientationPortrait
	if isLandscape {
		orientation = html2pdf.OrientationLandscape
	}

	cfg := html2pdf.RenderConfig{
		PageFormat:      html2pdf.PageFormatA4,
		Orientation:     orientation,
		MarginTop:       html2pdf.NewMarginInches(1.0),
		MarginBottom:    html2pdf.NewMarginInches(1.0),
		MarginLeft:      html2pdf.NewMarginInches(1.0),
		MarginRight:     html2pdf.NewMarginInches(1.0),
		PrintBackground: true,
		Scale:           1.0,
	}

	pdfReader, err := html2pdfConverter.GeneratePDF(ctx, htmlContent, cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot generate signature page PDF: %w", err)
	}

	pdfData, err := io.ReadAll(pdfReader)
	if err != nil {
		return nil, fmt.Errorf("cannot read signature page PDF: %w", err)
	}

	return pdfData, nil
}

func generateDocumentPDF(
	ctx context.Context,
	svc *Service,
	html2pdfConverter *html2pdf.Converter,
	conn pg.Querier,
	scope coredata.Scoper,
	version *coredata.DocumentVersion,
	options ExportPDFOptions,
) ([]byte, error) {
	document := &coredata.Document{}
	organization := &coredata.Organization{}

	if err := document.LoadByID(ctx, conn, scope, version.DocumentID); err != nil {
		return nil, fmt.Errorf("cannot load document: %w", err)
	}

	// Only show approvers from the last approved quorum in the export.
	var approverNames []string

	lastQuorum := &coredata.DocumentVersionApprovalQuorum{}
	if err := lastQuorum.LoadLastByDocumentVersionID(ctx, conn, scope, version.ID); err != nil {
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
			if err := approverProfiles.LoadByIDs(ctx, conn, scope, approverProfileIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
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

	isLandscape := version.Orientation == coredata.DocumentVersionOrientationLandscape

	docData := docgen.DocumentData{
		Title:                       version.Title,
		Content:                     json.RawMessage([]byte(version.Content)),
		Major:                       version.Major,
		Minor:                       version.Minor,
		Classification:              classification,
		Approvers:                   approverNames,
		PublishedAt:                 version.PublishedAt,
		CompanyHorizontalLogoBase64: horizontalLogoBase64,
		Landscape:                   isLandscape,
	}

	htmlContent, err := docgen.RenderHTML(docData)
	if err != nil {
		return nil, fmt.Errorf("cannot generate HTML: %w", err)
	}

	orientation := html2pdf.OrientationPortrait
	if isLandscape {
		orientation = html2pdf.OrientationLandscape
	}

	cfg := html2pdf.RenderConfig{
		PageFormat:        html2pdf.PageFormatA4,
		Orientation:       orientation,
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

		watermarkedPDF, err := pdfutils.AddConfidentialWithTimestamp(pdfData, *options.WatermarkEmail)
		if err != nil {
			return nil, fmt.Errorf("cannot add watermark to PDF: %w", err)
		}

		return watermarkedPDF, nil
	}

	return pdfData, nil
}

func (s *DocumentService) Export(
	ctx context.Context, scope coredata.Scoper,
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
				if err := document.LoadByID(ctx, conn, scope, documentID); err != nil {
					return fmt.Errorf("cannot load document %q: %w", documentID, err)
				}

				documentVersion := &coredata.DocumentVersion{}
				if err := documentVersion.LoadLatestVersion(ctx, conn, scope, documentID); err != nil {
					return fmt.Errorf("cannot load document version for %q: %w", documentID, err)
				}

				exportedPDF, err := exportDocumentPDF(
					ctx,
					s.svc,
					s.html2pdfConverter,
					conn,
					scope,
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
	ctx context.Context, scope coredata.Scoper,
	fileID gid.GID,
	recipientName string,
	recipientEmail mail.Addr,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			file := &coredata.File{}
			if err := file.LoadByID(ctx, tx, scope, fileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			downloadURL, err := s.GenerateDocumentExportDownloadURL(ctx, scope, file)
			if err != nil {
				return fmt.Errorf("cannot generate download URL: %w", err)
			}

			emailPresenter := emails.NewPresenter(s.svc.baseURL, recipientName)

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
	ctx context.Context, scope coredata.Scoper,
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
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	documentID gid.GID,
	ignoreExisting bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	document := &coredata.Document{}
	documentVersion := &coredata.DocumentVersion{}

	if err := document.LoadByID(ctx, tx, scope, documentID); err != nil {
		return nil, nil, fmt.Errorf("cannot load document %q: %w", documentID, err)
	}

	if document.ArchivedAt != nil {
		return nil, nil, &ErrDocumentArchived{}
	}

	if err := documentVersion.LoadLatestVersion(ctx, tx, scope, documentID); err != nil {
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
	ctx context.Context, scope coredata.Scoper,
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

	if err := document.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update document: %w", err)
	}

	if err := documentVersion.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update document version: %w", err)
	}

	return nil
}

// publishMajor publishes the document's draft as a new major version
// (currentMajor+1.0) and emits the version-published webhook. ignoreExisting is
// set by bulk publish to treat an already-published version as an idempotent
// no-op.
func (s *DocumentService) publishMajor(
	ctx context.Context,
	scope coredata.Scoper,
	tx pg.Tx,
	documentID gid.GID,
	changelog *string,
	ignoreExisting bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	document, documentVersion, err := s.loadDraftForPublish(ctx, scope, tx, documentID, ignoreExisting)
	if err != nil {
		return nil, nil, err
	}

	if ignoreExisting && documentVersion.Status == coredata.DocumentVersionStatusPublished {
		return document, documentVersion, nil
	}

	if document.CurrentPublishedMajor != nil {
		publishedVersion := &coredata.DocumentVersion{}
		if err := publishedVersion.LoadByDocumentIDAndVersion(ctx, tx, scope, documentID, *document.CurrentPublishedMajor, *document.CurrentPublishedMinor); err != nil {
			return nil, nil, fmt.Errorf("cannot load published version: %w", err)
		}

		documentVersion.Major = *document.CurrentPublishedMajor + 1
	} else {
		documentVersion.Major = 1
	}

	documentVersion.Minor = 0
	document.CurrentPublishedMajor = &documentVersion.Major
	document.CurrentPublishedMinor = &documentVersion.Minor

	if err := s.finalizePublish(ctx, scope, tx, document, documentVersion, changelog); err != nil {
		return nil, nil, err
	}

	if err := s.cancelPreviousMajorSignatureRequestsInTx(ctx, scope, tx, documentID, documentVersion.Major); err != nil {
		return nil, nil, err
	}

	if err := s.emitDocumentEvent(
		ctx,
		scope,
		tx,
		documentVersion.DocumentID,
		coredata.WebhookEventTypeDocumentVersionPublished,
		documentVersion,
		nil,
		nil,
		nil,
	); err != nil {
		return nil, nil, fmt.Errorf("cannot emit document version published webhook: %w", err)
	}

	return document, documentVersion, nil
}

// cancelPreviousMajorSignatureRequestsInTx cancels every still-pending
// signature request attached to a prior major version of the document. A new
// major supersedes the signing obligations of older majors, so their
// REQUESTED signatures must not linger. SIGNED signatures are left untouched
// to preserve the audit trail.
func (s *DocumentService) cancelPreviousMajorSignatureRequestsInTx(
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	documentID gid.GID,
	major int,
) error {
	signatures := &coredata.DocumentVersionSignatures{}
	if err := signatures.DeleteRequestedByDocumentIDBelowMajor(ctx, tx, scope, documentID, major); err != nil {
		return fmt.Errorf("cannot cancel signature requests from previous major versions: %w", err)
	}

	return nil
}

// publishMinor publishes the document's draft as a minor version and emits the
// version-published webhook. ignoreExisting is set by bulk publish to treat an
// already-published version as an idempotent no-op.
func (s *DocumentService) publishMinor(
	ctx context.Context,
	scope coredata.Scoper,
	tx pg.Tx,
	documentID gid.GID,
	changelog *string,
	ignoreExisting bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	document, documentVersion, err := s.loadDraftForPublish(ctx, scope, tx, documentID, ignoreExisting)
	if err != nil {
		return nil, nil, err
	}

	if ignoreExisting && documentVersion.Status == coredata.DocumentVersionStatusPublished {
		return document, documentVersion, nil
	}

	document.CurrentPublishedMajor = &documentVersion.Major
	document.CurrentPublishedMinor = &documentVersion.Minor

	if err := s.finalizePublish(ctx, scope, tx, document, documentVersion, changelog); err != nil {
		return nil, nil, err
	}

	if err := s.moveRequestedSignaturesToVersionInTx(ctx, scope, tx, documentVersion.ID); err != nil {
		return nil, nil, err
	}

	if err := s.emitDocumentEvent(
		ctx,
		scope,
		tx,
		documentVersion.DocumentID,
		coredata.WebhookEventTypeDocumentVersionPublished,
		documentVersion,
		nil,
		nil,
		nil,
	); err != nil {
		return nil, nil, fmt.Errorf("cannot emit document version published webhook: %w", err)
	}

	return document, documentVersion, nil
}

// moveRequestedSignaturesToVersionInTx carries every still-pending signature
// request from a prior minor of the same major onto the newly published minor
// version. The new minor supersedes the previous one while keeping the same
// signing obligations, so REQUESTED signatures follow along with their
// notification schedule (time and count) intact. SIGNED signatures stay put.
func (s *DocumentService) moveRequestedSignaturesToVersionInTx(
	ctx context.Context, scope coredata.Scoper,
	tx pg.Tx,
	documentVersionID gid.GID,
) error {
	signatures := &coredata.DocumentVersionSignatures{}
	if err := signatures.MoveRequestedToVersionWithinMajor(ctx, tx, scope, documentVersionID); err != nil {
		return fmt.Errorf("cannot move signature requests to the newly published minor version: %w", err)
	}

	return nil
}

func (s *DocumentService) generateAndUploadPublicationPDF(
	ctx context.Context, scope coredata.Scoper,
	documentVersion *coredata.DocumentVersion,
) error {
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
				documentVersion.ID,
				ExportPDFOptions{},
			)

			return err
		},
	)
	if err != nil {
		return fmt.Errorf("cannot generate publication PDF: %w", err)
	}

	now := time.Now()

	fileRecord := &coredata.File{
		ID:             gid.New(scope.GetTenantID(), coredata.FileEntityType),
		OrganizationID: documentVersion.OrganizationID,
		BucketName:     s.svc.bucket,
		MimeType:       "application/pdf",
		FileName:       fmt.Sprintf("%s v%d.%d.pdf", documentVersion.Title, documentVersion.Major, documentVersion.Minor),
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
			"type":                "document-version-pdf",
			"document-version-id": documentVersion.ID.String(),
		},
	)
	if err != nil {
		return fmt.Errorf("cannot upload publication PDF: %w", err)
	}

	fileRecord.FileSize = fileSize

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := fileRecord.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert file record: %w", err)
			}

			documentVersion.FileID = &fileRecord.ID

			documentVersion.UpdatedAt = now
			if err := documentVersion.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update document version with file ID: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot save publication PDF file record: %w", err)
	}

	return nil
}
