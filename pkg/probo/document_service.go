package probo

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/docgen"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/statelesstoken"
	"go.probo.inc/probo/pkg/watermarkpdf"
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
		OrganizationID        gid.GID
		Title                 string
		Content               string
		OwnerID               gid.GID
		Classification        coredata.DocumentClassification
		DocumentType          coredata.DocumentType
		TrustCenterVisibility *coredata.TrustCenterVisibility
	}

	UpdateDocumentRequest struct {
		DocumentID            gid.GID
		Title                 *string
		OwnerID               *gid.GID
		Classification        *coredata.DocumentClassification
		DocumentType          *coredata.DocumentType
		TrustCenterVisibility *coredata.TrustCenterVisibility
	}

	UpdateDocumentVersionRequest struct {
		ID      gid.GID
		Content string
	}

	RequestSignatureRequest struct {
		DocumentVersionID gid.GID
		Signatory         gid.GID
	}

	BulkRequestSignaturesRequest struct {
		DocumentIDs  []gid.GID
		SignatoryIDs []gid.GID
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

	documentExportEmailExpiresIn = 24 * time.Hour

	maxFilenameLength = 200
)

var (
	invalidFilenameChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f\x7f]`)
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
			return nil, fmt.Errorf("cannot generate changelog: %w", err)
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
			for _, documentID := range req.DocumentIDs {
				document, version, err := s.publishVersionInTx(ctx, tx, documentID, req.PublishedBy, &req.Changelog, true)
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

			document, documentVersion, err = s.publishVersionInTx(ctx, tx, documentID, publishedBy, changelog, false)
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
	publishedBy gid.GID,
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
			return nil, nil, &coredata.ErrDocumentVersionNoChanges{
				Message: "no changes detected",
			}
		}
	}

	if changelog != nil {
		documentVersion.Changelog = *changelog
	}

	document.CurrentPublishedVersion = &documentVersion.VersionNumber
	document.UpdatedAt = now

	documentVersion.Status = coredata.DocumentStatusPublished
	documentVersion.PublishedAt = &now
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
		ID:                    documentID,
		Title:                 req.Title,
		DocumentType:          req.DocumentType,
		TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
		Classification:        req.Classification,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if req.TrustCenterVisibility != nil {
		document.TrustCenterVisibility = *req.TrustCenterVisibility
	}

	documentVersion := &coredata.DocumentVersion{
		ID:             documentVersionID,
		DocumentID:     documentID,
		Title:          req.Title,
		OwnerID:        req.OwnerID,
		VersionNumber:  1,
		Content:        req.Content,
		Status:         coredata.DocumentStatusDraft,
		Classification: req.Classification,
		CreatedAt:      now,
		UpdatedAt:      now,
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
  pv.id AS document_version_id,
  o.name AS organization_name
FROM
	documents p
	INNER JOIN document_versions pv ON pv.document_id = p.id
	INNER JOIN document_version_signatures pvs ON pvs.document_version_id = pv.id
	INNER JOIN organizations o ON o.id = p.organization_id
WHERE
    p.tenant_id = $1
	AND pvs.signed_by = $2
	AND pvs.signed_at IS NULL
	AND pv.status = 'PUBLISHED'
	AND pv.version_number = (
		SELECT MAX(pv2.version_number)
		FROM document_versions pv2
		WHERE pv2.document_id = pv.document_id
		AND pv2.status = 'PUBLISHED'
	)
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

			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			for _, people := range peoples {
				token, err := statelesstoken.NewToken(
					s.svc.tokenSecret,
					TokenTypeSigningRequest,
					time.Hour*24*30,
					SigningRequestData{
						OrganizationID: organizationID,
						PeopleID:       people.ID,
					},
				)
				if err != nil {
					return fmt.Errorf("cannot create signing request token: %w", err)
				}

				baseURLParsed, err := url.Parse(s.svc.baseURL)
				if err != nil {
					return fmt.Errorf("cannot parse base URL: %w", err)
				}

				signRequestURL := url.URL{
					Scheme: baseURLParsed.Scheme,
					Host:   baseURLParsed.Host,
					Path:   "/documents/signing-requests",
					RawQuery: url.Values{
						"token": []string{token},
					}.Encode(),
				}

				subject, textBody, htmlBody, err := emails.RenderDocumentSigning(
					s.svc.baseURL,
					people.FullName,
					organization.Name,
					signRequestURL.String(),
				)
				if err != nil {
					return fmt.Errorf("cannot render signing request email: %w", err)
				}

				email := coredata.NewEmail(
					people.FullName,
					people.PrimaryEmailAddress,
					subject,
					textBody,
					htmlBody,
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
			documentVersion.Classification = document.Classification
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
	tx pg.Conn,
	documentVersionID gid.GID,
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
		func(conn pg.Conn) error {
			return documentVersionSignatures.LoadByDocumentVersionID(ctx, conn, s.svc.scope, documentVersionID, cursor, filter)
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
			draftVersion.Classification = document.Classification
			draftVersion.Content = latestVersion.Content
			draftVersion.Status = coredata.DocumentStatusDraft
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

func (s *DocumentService) SoftDelete(
	ctx context.Context,
	documentID gid.GID,
) error {
	document := coredata.Document{ID: documentID}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return document.SoftDelete(ctx, conn, s.svc.scope)
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

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return documents.BulkSoftDelete(ctx, conn, s.svc.scope)
		},
	)
}

func (s *DocumentService) RequestExport(
	ctx context.Context,
	documentIDs []gid.GID,
	recipientEmail string,
	recipientName string,
	options BulkExportOptions,
) (*coredata.ExportJob, error) {
	var exportJobID gid.GID
	exportJob := &coredata.ExportJob{}

	if options.WithWatermark {
		if options.WatermarkEmail == nil {
			return nil, fmt.Errorf("watermark email is required when with watermark is true")
		}
		if _, err := mail.ParseAddress(*options.WatermarkEmail); err != nil {
			return nil, fmt.Errorf("invalid email address")
		}
	}

	err := s.svc.pg.WithTx(ctx, func(conn pg.Conn) error {
		for _, documentID := range documentIDs {
			document := &coredata.Document{}
			if err := document.LoadByID(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load document %q: %w", documentID, err)
			}
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
	req UpdateDocumentRequest,
) (*coredata.Document, error) {
	document := &coredata.Document{}
	people := &coredata.People{}
	now := time.Now()

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := document.LoadByID(ctx, tx, s.svc.scope, req.DocumentID); err != nil {
				return fmt.Errorf("cannot load document %q: %w", req.DocumentID, err)
			}

			if req.Title != nil {
				document.Title = *req.Title
			}

			if req.Classification != nil {
				document.Classification = *req.Classification
			}

			if req.DocumentType != nil {
				document.DocumentType = *req.DocumentType
			}

			if req.DocumentType != nil {
				document.DocumentType = *req.DocumentType
			}

			if req.TrustCenterVisibility != nil {
				document.TrustCenterVisibility = *req.TrustCenterVisibility
			}

			if req.OwnerID != nil {
				if err := people.LoadByID(ctx, tx, s.svc.scope, *req.OwnerID); err != nil {
					return fmt.Errorf("cannot load owner %q: %w", *req.OwnerID, err)
				}
				document.OwnerID = people.ID
			}

			document.UpdatedAt = now

			if err := document.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update document: %w", err)
			}

			// Update the draft version if it exists to keep it in sync with the document
			draftVersion := &coredata.DocumentVersion{}
			err := draftVersion.LoadLatestVersion(ctx, tx, s.svc.scope, req.DocumentID)
			if err == nil && draftVersion.Status == coredata.DocumentStatusDraft {
				draftVersion.Title = document.Title
				draftVersion.OwnerID = document.OwnerID
				draftVersion.Classification = document.Classification
				draftVersion.UpdatedAt = now

				if err := draftVersion.Update(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot update draft version: %w", err)
				}
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

type ExportPDFOptions struct {
	WithWatermark  bool
	WatermarkEmail *string
	WithSignatures bool
}

type BulkExportOptions struct {
	WithWatermark  bool
	WatermarkEmail *string
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
		func(conn pg.Conn) (err error) {
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
		func(tx pg.Conn) error {
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
			defer tempFile.Close()
			defer os.Remove(tempFile.Name())

			exportArgs, err := exportJob.GetDocumentExportArguments()
			if err != nil {
				return fmt.Errorf("cannot get export arguments: %w", err)
			}

			exportOptions := BulkExportOptions{
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
					Bucket:        ref.Ref(s.svc.bucket),
					Key:           ref.Ref(uuid.String()),
					Body:          tempFile,
					ContentLength: ref.Ref(fileInfo.Size()),
					ContentType:   ref.Ref("application/zip"),
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
	conn pg.Conn,
	scope coredata.Scoper,
	documentVersionID gid.GID,
	options ExportPDFOptions,
) ([]byte, error) {
	document := &coredata.Document{}
	version := &coredata.DocumentVersion{}
	owner := &coredata.People{}
	organization := &coredata.Organization{}

	if err := version.LoadByID(ctx, conn, scope, documentVersionID); err != nil {
		return nil, fmt.Errorf("cannot load document version: %w", err)
	}

	if err := document.LoadByID(ctx, conn, scope, version.DocumentID); err != nil {
		return nil, fmt.Errorf("cannot load document: %w", err)
	}

	if err := owner.LoadByID(ctx, conn, scope, document.OwnerID); err != nil {
		return nil, fmt.Errorf("cannot load document owner: %w", err)
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
	switch document.Classification {
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
		fileErr := svc.pg.WithConn(ctx, func(conn pg.Conn) error {
			return fileRecord.LoadByID(ctx, conn, scope, *organization.HorizontalLogoFileID)
		})
		if fileErr == nil {
			base64Data, mimeType, logoErr := svc.fileManager.GetFileBase64(ctx, fileRecord)
			if logoErr == nil {
				horizontalLogoBase64 = fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
			}
		}
	}

	docData := docgen.DocumentData{
		Title:                       version.Title,
		Content:                     version.Content,
		Version:                     version.VersionNumber,
		Classification:              classification,
		Approver:                    owner.FullName,
		PublishedAt:                 version.PublishedAt,
		Signatures:                  signatureData,
		CompanyHorizontalLogoBase64: horizontalLogoBase64,
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
		if _, err := mail.ParseAddress(*options.WatermarkEmail); err != nil {
			return nil, fmt.Errorf("invalid email address")
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
	options BulkExportOptions,
) (err error) {
	archive := zip.NewWriter(file)
	defer func() {
		if closeErr := archive.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("cannot close archive: %w", closeErr)
		}
	}()

	return s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			for i, documentID := range documentIDs {
				document := &coredata.Document{}
				if err := document.LoadByID(ctx, conn, s.svc.scope, documentID); err != nil {
					return fmt.Errorf("cannot load document %q: %w", documentID, err)
				}

				documentVersion := &coredata.DocumentVersion{}
				if err := documentVersion.LoadLatestVersion(ctx, conn, s.svc.scope, documentID); err != nil {
					return fmt.Errorf("cannot load document version for %q: %w", documentID, err)
				}

				pdfOptions := ExportPDFOptions{
					WithWatermark:  options.WithWatermark,
					WatermarkEmail: options.WatermarkEmail,
					WithSignatures: options.WithSignatures,
				}

				exportedPDF, err := exportDocumentPDF(
					ctx,
					s.svc,
					s.html2pdfConverter,
					conn,
					s.svc.scope,
					documentVersion.ID,
					pdfOptions,
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
	recipientEmail string,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			file := &coredata.File{}
			if err := file.LoadByID(ctx, tx, s.svc.scope, fileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			downloadURL, err := s.GenerateDocumentExportDownloadURL(ctx, file)
			if err != nil {
				return fmt.Errorf("cannot generate download URL: %w", err)
			}

			subject, textBody, htmlBody, err := emails.RenderDocumentExport(
				s.svc.baseURL,
				recipientName,
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
			Bucket:                     ref.Ref(s.svc.bucket),
			Key:                        ref.Ref(file.FileKey),
			ResponseCacheControl:       ref.Ref("max-age=3600, public"),
			ResponseContentType:        ref.Ref(file.MimeType),
			ResponseContentDisposition: ref.Ref(fmt.Sprintf("attachment; filename=\"%s\"", file.FileName)),
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
