package probo

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/docgen"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/html2pdf"
	"github.com/getprobo/probo/pkg/page"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"text/template"
)

type (
	DocumentService struct {
		svc               *TenantService
		html2pdfConverter *html2pdf.Converter
	}

	ErrSignatureNotCancellable struct {
		currentState  coredata.DocumentVersionSignatureState
		expectedState coredata.DocumentVersionSignatureState
	}

	CreateDocumentRequest struct {
		OrganizationID gid.GID
		Title          string
		Content        string
		OwnerID        gid.GID
		CreatedBy      gid.GID
		DocumentType   coredata.DocumentType
	}

	UpdateDocumentVersionRequest struct {
		ID      gid.GID
		Content string
	}

	RequestSignatureRequest struct {
		DocumentVersionID gid.GID
		RequestedBy       gid.GID
		Signatory         gid.GID
	}

	BulkRequestSignaturesRequest struct {
		DocumentIDs  []gid.GID
		SignatoryIDs []gid.GID
		RequestedBy  gid.GID
	}

	SigningRequestData struct {
		OrganizationID gid.GID `json:"organization_id"`
		PeopleID       gid.GID `json:"people_id"`
	}

	BulkPublishVersionsRequest struct {
		DocumentIDs []gid.GID
		PublishedBy gid.GID
		Changelog   string
	}
)

const (
	TokenTypeSigningRequest = "signing_request"
)

func (e ErrSignatureNotCancellable) Error() string {
	return fmt.Sprintf("cannot cancel signature request: signature is in state %v, expected %v",
		e.currentState, e.expectedState)
}

func (s *DocumentService) Get(
	ctx context.Context,
	documentID gid.GID,
) (*coredata.Document, error) {
	document := &coredata.Document{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return document.LoadByID(ctx, conn, s.svc.scope, documentID)
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
		func(conn pg.Conn) error {
			if err := draftVersion.LoadLatestVersion(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load draft version: %w", err)
			}

			if draftVersion.Status != coredata.DocumentStatusDraft {
				return fmt.Errorf("latest version is not a draft")
			}

			document := &coredata.Document{}
			if err := document.LoadByID(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.CurrentPublishedVersion == nil {
				initialVersionChangelog := "Initial version"
				changelog = &initialVersionChangelog
			} else {
				if err := publishedVersion.LoadByDocumentIDAndVersionNumber(ctx, conn, s.svc.scope, documentID, *document.CurrentPublishedVersion); err != nil {
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
			return nil, fmt.Errorf("failed to generate changelog: %w", err)
		}
	}

	return changelog, nil
}

func (s *DocumentService) BulkPublishVersions(
	ctx context.Context,
	req BulkPublishVersionsRequest,
) ([]*coredata.DocumentVersion, []*coredata.Document, error) {
	var publishedVersions []*coredata.DocumentVersion
	var updatedDocuments []*coredata.Document

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			people := &coredata.People{}
			if err := people.LoadByID(ctx, tx, s.svc.scope, req.PublishedBy); err != nil {
				return fmt.Errorf("cannot load people: %w", err)
			}

			for _, documentID := range req.DocumentIDs {
				document, version, err := s.publishVersionInTx(ctx, tx, documentID, people, &req.Changelog, true)
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

func (s *DocumentService) PublishVersion(
	ctx context.Context,
	documentID gid.GID,
	publishedBy gid.GID,
	changelog *string,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	var document *coredata.Document
	var documentVersion *coredata.DocumentVersion

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			var err error

			people := &coredata.People{}
			if err := people.LoadByID(ctx, tx, s.svc.scope, publishedBy); err != nil {
				return fmt.Errorf("cannot load people: %w", err)
			}

			document, documentVersion, err = s.publishVersionInTx(ctx, tx, documentID, people, changelog, false)
			if err != nil {
				return fmt.Errorf("cannot publish version: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *DocumentService) publishVersionInTx(
	ctx context.Context,
	tx pg.Conn,
	documentID gid.GID,
	publishedBy *coredata.People,
	changelog *string,
	ignoreExisting bool,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	document := &coredata.Document{}
	documentVersion := &coredata.DocumentVersion{}
	publishedVersion := &coredata.DocumentVersion{}
	now := time.Now()

	if err := document.LoadByID(ctx, tx, s.svc.scope, documentID); err != nil {
		return nil, nil, fmt.Errorf("cannot load document %q: %w", documentID, err)
	}

	if err := documentVersion.LoadLatestVersion(ctx, tx, s.svc.scope, documentID); err != nil {
		return nil, nil, fmt.Errorf("cannot load current draft: %w", err)
	}

	if ignoreExisting && documentVersion.Status == coredata.DocumentStatusPublished {
		return document, documentVersion, nil
	}

	if documentVersion.Status != coredata.DocumentStatusDraft {
		return nil, nil, fmt.Errorf("cannot publish version")
	}

	if document.CurrentPublishedVersion != nil {
		if err := publishedVersion.LoadByDocumentIDAndVersionNumber(ctx, tx, s.svc.scope, documentID, *document.CurrentPublishedVersion); err != nil {
			return nil, nil, fmt.Errorf("cannot load published version: %w", err)
		}
		if publishedVersion.Content == documentVersion.Content &&
			publishedVersion.Title == documentVersion.Title &&
			publishedVersion.OwnerID == documentVersion.OwnerID {
			return nil, nil, fmt.Errorf("cannot publish version: no changes detected")
		}
	}

	if changelog != nil {
		documentVersion.Changelog = *changelog
	}

	document.CurrentPublishedVersion = &documentVersion.VersionNumber
	document.UpdatedAt = now

	documentVersion.Status = coredata.DocumentStatusPublished
	documentVersion.PublishedAt = &now
	documentVersion.PublishedBy = &publishedBy.ID
	documentVersion.UpdatedAt = now

	if err := document.Update(ctx, tx, s.svc.scope); err != nil {
		return nil, nil, fmt.Errorf("cannot update document: %w", err)
	}

	if err := documentVersion.Update(ctx, tx, s.svc.scope); err != nil {
		return nil, nil, fmt.Errorf("cannot update document version: %w", err)
	}

	return document, documentVersion, nil
}

func (s *DocumentService) Create(
	ctx context.Context,
	req CreateDocumentRequest,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	now := time.Now()
	documentID := gid.New(s.svc.scope.GetTenantID(), coredata.DocumentEntityType)
	documentVersionID := gid.New(s.svc.scope.GetTenantID(), coredata.DocumentVersionEntityType)

	organization := &coredata.Organization{}
	people := &coredata.People{}

	document := &coredata.Document{
		ID:                documentID,
		Title:             req.Title,
		DocumentType:      req.DocumentType,
		ShowOnTrustCenter: false,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	documentVersion := &coredata.DocumentVersion{
		ID:            documentVersionID,
		DocumentID:    documentID,
		Title:         req.Title,
		OwnerID:       req.OwnerID,
		VersionNumber: 1,
		Content:       req.Content,
		Status:        coredata.DocumentStatusDraft,
		CreatedBy:     req.CreatedBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := people.LoadByID(ctx, conn, s.svc.scope, req.OwnerID); err != nil {
				return fmt.Errorf("cannot load people: %w", err)
			}

			document.OrganizationID = organization.ID
			document.OwnerID = people.ID

			if err := document.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert document: %w", err)
			}

			if err := documentVersion.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot create document version: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return document, documentVersion, nil
}

func (s *DocumentService) ListSigningRequests(
	ctx context.Context,
	organizationID gid.GID,
	peopleID gid.GID,
) ([]map[string]any, error) {
	q := `
SELECT
  p.title,
  pv.content,
  pv.id AS document_version_id
FROM
	documents p
	INNER JOIN document_versions pv ON pv.document_id = p.id
	INNER JOIN document_version_signatures pvs ON pvs.document_version_id = pv.id
WHERE
    p.tenant_id = $1
	AND pvs.signed_by = $2
	AND pvs.signed_at IS NULL
`

	var results []map[string]any
	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			rows, err := conn.Query(ctx, q, s.svc.scope.GetTenantID(), peopleID)
			if err != nil {
				return fmt.Errorf("cannot query documents: %w", err)
			}

			results, err = pgx.CollectRows(rows, pgx.RowToMap)
			if err != nil {
				return err
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (s *DocumentService) SendSigningNotifications(
	ctx context.Context,
	organizationID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			var peoples coredata.Peoples
			if err := peoples.LoadAwaitingSigning(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot load people: %w", err)
			}

			for _, people := range peoples {
				now := time.Now()

				emailID := gid.New(s.svc.scope.GetTenantID(), coredata.EmailEntityType)

				token, err := statelesstoken.NewToken(
					s.svc.tokenSecret,
					TokenTypeSigningRequest,
					time.Hour*24*7,
					SigningRequestData{
						OrganizationID: organizationID,
						PeopleID:       people.ID,
					},
				)
				if err != nil {
					return fmt.Errorf("cannot create signing request token: %w", err)
				}

				signRequestURL := url.URL{
					Scheme: "https",
					Host:   s.svc.hostname,
					Path:   "/documents/signing-requests",
					RawQuery: url.Values{
						"token": []string{token},
					}.Encode(),
				}

				data := struct {
					SignURL string
				}{
					SignURL: signRequestURL.String(),
				}

				var body bytes.Buffer
				if err := signEmailTemplate.Execute(&body, data); err != nil {
					return fmt.Errorf("failed to render email template: %w", err)
				}

				email := &coredata.Email{
					ID:             emailID,
					RecipientEmail: people.PrimaryEmailAddress,
					RecipientName:  people.FullName,
					Subject:        "Probo - Documents Signing Request",
					TextBody:       body.String(),
					CreatedAt:      now,
					UpdatedAt:      now,
				}

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

func (s *DocumentService) SignDocumentVersion(
	ctx context.Context,
	documentVersionID gid.GID,
	signatory gid.GID,
) error {
	documentVersion := &coredata.DocumentVersion{}
	documentVersionSignature := &coredata.DocumentVersionSignature{}
	now := time.Now()

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := documentVersion.LoadByID(ctx, conn, s.svc.scope, documentVersionID); err != nil {
				return fmt.Errorf("cannot load document version %q: %w", documentVersionID, err)
			}

			if documentVersion.Status != coredata.DocumentStatusPublished {
				return fmt.Errorf("cannot sign unpublished version")
			}

			if err := documentVersionSignature.LoadByDocumentVersionIDAndSignatory(ctx, conn, s.svc.scope, documentVersionID, signatory); err != nil {
				return fmt.Errorf("cannot load document version signature: %w", err)
			}

			if documentVersionSignature.State == coredata.DocumentVersionSignatureStateSigned {
				return fmt.Errorf("document version already signed")
			}

			documentVersionSignature.State = coredata.DocumentVersionSignatureStateSigned
			documentVersionSignature.SignedAt = &now
			documentVersionSignature.UpdatedAt = now

			if err := documentVersion.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update document version: %w", err)
			}

			if err := documentVersionSignature.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update document version signature: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return fmt.Errorf("cannot sign document version: %w", err)
	}

	return nil
}

func (s *DocumentService) UpdateVersion(
	ctx context.Context,
	req UpdateDocumentVersionRequest,
) (*coredata.DocumentVersion, error) {
	documentVersion := &coredata.DocumentVersion{}
	document := &coredata.Document{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := documentVersion.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load document version %q: %w", req.ID, err)
			}

			if err := document.LoadByID(ctx, conn, s.svc.scope, documentVersion.DocumentID); err != nil {
				return fmt.Errorf("cannot load document %q: %w", documentVersion.DocumentID, err)
			}

			if documentVersion.Status != coredata.DocumentStatusDraft {
				return fmt.Errorf("cannot update published version")
			}

			documentVersion.Title = document.Title
			documentVersion.OwnerID = document.OwnerID
			documentVersion.Content = req.Content
			documentVersion.UpdatedAt = time.Now()

			if err := documentVersion.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update document version: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return documentVersion, nil
}

func (s *DocumentService) GetVersionSignature(
	ctx context.Context,
	signatureID gid.GID,
) (*coredata.DocumentVersionSignature, error) {
	documentVersionSignature := &coredata.DocumentVersionSignature{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
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
		func(tx pg.Conn) error {
			for _, documentID := range req.DocumentIDs {
				documentVersion := &coredata.DocumentVersion{}
				if err := documentVersion.LoadLatestVersion(ctx, tx, s.svc.scope, documentID); err != nil {
					return fmt.Errorf("cannot load latest version for document %q: %w", documentID, err)
				}

				if documentVersion.Status != coredata.DocumentStatusPublished {
					return fmt.Errorf("cannot request signature for unpublished document %q", documentID)
				}

				requestedBy := &coredata.People{}
				if err := requestedBy.LoadByID(ctx, tx, s.svc.scope, req.RequestedBy); err != nil {
					return fmt.Errorf("cannot load requested by: %w", err)
				}

				for _, signatoryID := range req.SignatoryIDs {
					signature, err := s.createSignatureRequestInTx(ctx, tx, documentVersion.ID, requestedBy, signatoryID, true)
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
	tx pg.Conn,
	documentVersionID gid.GID,
	requestedBy *coredata.People,
	signatoryID gid.GID,
	ignoreExisting bool,
) (*coredata.DocumentVersionSignature, error) {
	signatory := &coredata.People{}

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
		DocumentVersionID: documentVersionID,
		State:             coredata.DocumentVersionSignatureStateRequested,
		RequestedBy:       requestedBy.ID,
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
	documentVersion, err := s.GetVersion(ctx, req.DocumentVersionID)
	if err != nil {
		return nil, fmt.Errorf("cannot get document version: %w", err)
	}

	if documentVersion.Status != coredata.DocumentStatusPublished {
		return nil, fmt.Errorf("cannot request signature for unpublished version")
	}

	var signature *coredata.DocumentVersionSignature
	err = s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			requestedBy := &coredata.People{}
			if err := requestedBy.LoadByID(ctx, tx, s.svc.scope, req.RequestedBy); err != nil {
				return fmt.Errorf("cannot load requested by %q: %w", req.RequestedBy, err)
			}

			signature, err = s.createSignatureRequestInTx(ctx, tx, req.DocumentVersionID, requestedBy, req.Signatory, false)
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
) (*page.Page[*coredata.DocumentVersionSignature, coredata.DocumentVersionSignatureOrderField], error) {
	var documentVersionSignatures coredata.DocumentVersionSignatures

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return documentVersionSignatures.LoadByDocumentVersionID(ctx, conn, s.svc.scope, documentVersionID, cursor)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(documentVersionSignatures, cursor), nil
}

func (s *DocumentService) CreateDraft(
	ctx context.Context,
	documentID gid.GID,
	createdBy gid.GID,
) (*coredata.DocumentVersion, error) {
	draftVersionID := gid.New(s.svc.scope.GetTenantID(), coredata.DocumentVersionEntityType)

	latestVersion := &coredata.DocumentVersion{}
	document := &coredata.Document{}
	draftVersion := &coredata.DocumentVersion{}
	now := time.Now()

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := document.LoadByID(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if err := latestVersion.LoadLatestVersion(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load latest version: %w", err)
			}

			if latestVersion.Status != coredata.DocumentStatusPublished {
				return fmt.Errorf("cannot create draft from unpublished version")
			}

			draftVersion.ID = draftVersionID
			draftVersion.DocumentID = documentID
			draftVersion.Title = document.Title
			draftVersion.OwnerID = document.OwnerID
			draftVersion.VersionNumber = latestVersion.VersionNumber + 1
			draftVersion.Content = latestVersion.Content
			draftVersion.Status = coredata.DocumentStatusDraft
			draftVersion.CreatedBy = createdBy
			draftVersion.CreatedAt = now
			draftVersion.UpdatedAt = now

			if err := draftVersion.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot create draft: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return draftVersion, nil
}

func (s *DocumentService) DeleteDraft(
	ctx context.Context,
	documentVersionID gid.GID,
) error {
	documentVersion := &coredata.DocumentVersion{}

	return s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := documentVersion.LoadByID(ctx, conn, s.svc.scope, documentVersionID); err != nil {
				return fmt.Errorf("cannot load document version: %w", err)
			}

			if documentVersion.Status != coredata.DocumentStatusDraft {
				return fmt.Errorf("cannot delete published document version")
			}

			if documentVersion.VersionNumber == 1 {
				return fmt.Errorf("cannot delete the first version of a document")
			}

			if err := documentVersion.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete document version: %w", err)
			}

			return nil
		},
	)
}

func (s *DocumentService) Delete(
	ctx context.Context,
	documentID gid.GID,
) error {
	document := coredata.Document{ID: documentID}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return document.Delete(ctx, conn, s.svc.scope)
		},
	)
}

func (s *DocumentService) ListVersions(
	ctx context.Context,
	documentID gid.GID,
	cursor *page.Cursor[coredata.DocumentVersionOrderField],
) (*page.Page[*coredata.DocumentVersion, coredata.DocumentVersionOrderField], error) {
	var documentVersions coredata.DocumentVersions

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return documentVersions.LoadByDocumentID(ctx, conn, s.svc.scope, documentID, cursor)
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
		func(conn pg.Conn) error {
			return documentVersion.LoadByID(ctx, conn, s.svc.scope, documentVersionID)
		},
	)

	if err != nil {
		return nil, err
	}

	return documentVersion, nil
}

func (s *DocumentService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.DocumentFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
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
		func(conn pg.Conn) error {
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
		func(conn pg.Conn) (err error) {
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
		func(conn pg.Conn) error {
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
		func(conn pg.Conn) (err error) {
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
		func(conn pg.Conn) error {
			return documents.LoadByRiskID(ctx, conn, s.svc.scope, riskID, cursor, filter)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(documents, cursor), nil
}

func (s *DocumentService) Update(
	ctx context.Context,
	documentID gid.GID,
	newOwnerID *gid.GID,
	documentType *coredata.DocumentType,
	title *string,
	showOnTrustCenter *bool,
) (*coredata.Document, error) {
	document := &coredata.Document{}
	people := &coredata.People{}
	now := time.Now()

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := document.LoadByID(ctx, tx, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load document %q: %w", documentID, err)
			}

			if newOwnerID != nil {
				if err := people.LoadByID(ctx, tx, s.svc.scope, *newOwnerID); err != nil {
					return fmt.Errorf("cannot load new owner %q: %w", *newOwnerID, err)
				}
				document.OwnerID = *newOwnerID
			}

			if documentType != nil {
				document.DocumentType = *documentType
			}

			if title != nil {
				document.Title = *title
			}

			if showOnTrustCenter != nil {
				document.ShowOnTrustCenter = *showOnTrustCenter
			}

			document.UpdatedAt = now

			if err := document.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update document: %w", err)
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
		func(tx pg.Conn) error {
			if err := documentVersionSignature.LoadByID(ctx, tx, s.svc.scope, documentVersionSignatureID); err != nil {
				return fmt.Errorf("cannot load document version signature: %w", err)
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

func (s *DocumentService) ExportPDF(
	ctx context.Context,
	documentVersionID gid.GID,
) ([]byte, error) {
	document := &coredata.Document{}
	version := &coredata.DocumentVersion{}
	owner := &coredata.People{}
	publishedBy := &coredata.People{}
	signatures := coredata.DocumentVersionSignatures{}
	peopleMap := make(map[gid.GID]*coredata.People)

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := version.LoadByID(ctx, conn, s.svc.scope, documentVersionID); err != nil {
				return fmt.Errorf("cannot load document version: %w", err)
			}

			if err := document.LoadByID(ctx, conn, s.svc.scope, version.DocumentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if version.PublishedBy != nil {
				if err := publishedBy.LoadByID(ctx, conn, s.svc.scope, *version.PublishedBy); err != nil {
					return fmt.Errorf("cannot load published by person: %w", err)
				}
			}

			cursor := page.NewCursor(
				100,
				nil,
				page.Head,
				page.OrderBy[coredata.DocumentVersionSignatureOrderField]{
					Field:     coredata.DocumentVersionSignatureOrderFieldCreatedAt,
					Direction: page.OrderDirectionAsc,
				},
			)

			if err := signatures.LoadByDocumentVersionID(ctx, conn, s.svc.scope, documentVersionID, cursor); err != nil {
				return fmt.Errorf("cannot load document version signatures: %w", err)
			}

			if err := owner.LoadByID(ctx, conn, s.svc.scope, document.OwnerID); err != nil {
				return fmt.Errorf("cannot load document owner: %w", err)
			}

			// TODO: refactor this to use a single query
			for _, sig := range signatures {
				if _, ok := peopleMap[sig.SignedBy]; !ok {
					people := &coredata.People{}
					if err := people.LoadByID(ctx, conn, s.svc.scope, sig.SignedBy); err != nil {
						return fmt.Errorf("cannot load people %q: %w", sig.SignedBy, err)
					}
					peopleMap[sig.SignedBy] = people
				}

				if _, ok := peopleMap[sig.RequestedBy]; !ok {
					people := &coredata.People{}
					if err := people.LoadByID(ctx, conn, s.svc.scope, sig.RequestedBy); err != nil {
						return fmt.Errorf("cannot load people %q: %w", sig.RequestedBy, err)
					}
					peopleMap[sig.RequestedBy] = people
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	classification := docgen.ClassificationInternal
	switch document.DocumentType {
	case coredata.DocumentTypePolicy:
		classification = docgen.ClassificationConfidential
	case coredata.DocumentTypeISMS:
		classification = docgen.ClassificationSecret
	}

	docData := docgen.DocumentData{
		Title:          version.Title,
		Content:        version.Content,
		Version:        version.VersionNumber,
		Classification: classification,
		Approver:       owner.FullName,
		Description:    version.Changelog,
		PublishedAt:    version.PublishedAt,
		PublishedBy:    publishedBy.FullName,
		Signatures:     make([]docgen.SignatureData, len(signatures)),
	}

	for i, sig := range signatures {
		docData.Signatures[i] = docgen.SignatureData{
			SignedBy:    peopleMap[sig.SignedBy].FullName,
			SignedAt:    sig.SignedAt,
			State:       sig.State,
			RequestedAt: sig.RequestedAt,
			RequestedBy: peopleMap[sig.RequestedBy].FullName,
		}
	}

	htmlContent, err := docgen.RenderHTML(docData)
	if err != nil {
		return nil, fmt.Errorf("cannot generate HTML: %w", err)
	}

	cfg := html2pdf.RenderConfig{
		PageFormat:      html2pdf.PageFormatA4,
		Orientation:     html2pdf.OrientationPortrait,
		MarginTop:       html2pdf.NewMarginInches(1.0),
		MarginBottom:    html2pdf.NewMarginInches(1.0),
		MarginLeft:      html2pdf.NewMarginInches(1.0),
		MarginRight:     html2pdf.NewMarginInches(1.0),
		PrintBackground: true,
		Scale:           1.0,
	}

	pdfReader, err := s.html2pdfConverter.GeneratePDF(ctx, htmlContent, cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot generate PDF: %w", err)
	}

	pdfData, err := io.ReadAll(pdfReader)
	if err != nil {
		return nil, fmt.Errorf("cannot read PDF data: %w", err)
	}
	return pdfData, nil
}
