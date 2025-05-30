package probo

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	DocumentService struct {
		svc *TenantService
	}

	CreateDocumentRequest struct {
		OrganizationID gid.GID
		Title          string
		Content        string
		OwnerID        gid.GID
		CreatedBy      gid.GID
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

	SigningRequestData struct {
		OrganizationID gid.GID `json:"organization_id"`
		PeopleID       gid.GID `json:"people_id"`
	}
)

const (
	TokenTypeSigningRequest = "signing_request"
)

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

func (s *DocumentService) PublishVersion(
	ctx context.Context,
	documentID gid.GID,
	publishedBy gid.GID,
) (*coredata.Document, *coredata.DocumentVersion, error) {
	document := &coredata.Document{}
	documentVersion := &coredata.DocumentVersion{}
	now := time.Now()

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := document.LoadByID(ctx, tx, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load document %q: %w", documentID, err)
			}

			if err := documentVersion.LoadLatestVersion(ctx, tx, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load current draft: %w", err)
			}

			if documentVersion.Status != coredata.DocumentStatusDraft {
				return fmt.Errorf("cannot publish version")
			}

			document.CurrentPublishedVersion = &documentVersion.VersionNumber
			document.UpdatedAt = now

			documentVersion.Status = coredata.DocumentStatusPublished
			documentVersion.PublishedAt = &now
			documentVersion.PublishedBy = &publishedBy
			documentVersion.UpdatedAt = now

			if err := document.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update document: %w", err)
			}

			if err := documentVersion.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update document version: %w", err)
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
	now := time.Now()
	documentID := gid.New(s.svc.scope.GetTenantID(), coredata.DocumentEntityType)
	documentVersionID := gid.New(s.svc.scope.GetTenantID(), coredata.DocumentVersionEntityType)

	organization := &coredata.Organization{}
	people := &coredata.People{}

	document := &coredata.Document{
		ID:        documentID,
		Title:     req.Title,
		CreatedAt: now,
		UpdatedAt: now,
	}

	documentVersion := &coredata.DocumentVersion{
		ID:            documentVersionID,
		DocumentID:    documentID,
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

				email := &coredata.Email{
					ID:             emailID,
					RecipientEmail: people.PrimaryEmailAddress,
					RecipientName:  people.FullName,
					Subject:        "Probo - Documents Signing Request",
					TextBody:       fmt.Sprintf("Hi,\nYou have documents awaiting your signature. Please follow this link to sign them: %s", signRequestURL.String()),
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

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := documentVersion.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load document version %q: %w", req.ID, err)
			}

			if documentVersion.Status != coredata.DocumentStatusDraft {
				return fmt.Errorf("cannot update published version")
			}

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

func (s *DocumentService) RequestSignature(
	ctx context.Context,
	req RequestSignatureRequest,
) (*coredata.DocumentVersionSignature, error) {
	documentVersionSignatureID := gid.New(s.svc.scope.GetTenantID(), coredata.DocumentVersionSignatureEntityType)

	documentVersion, err := s.GetVersion(ctx, req.DocumentVersionID)
	if err != nil {
		return nil, fmt.Errorf("cannot get document version: %w", err)
	}

	if documentVersion.Status != coredata.DocumentStatusPublished {
		return nil, fmt.Errorf("cannot request signature for unpublished version")
	}

	now := time.Now()
	documentVersionSignature := &coredata.DocumentVersionSignature{
		ID:                documentVersionSignatureID,
		DocumentVersionID: req.DocumentVersionID,
		State:             coredata.DocumentVersionSignatureStateRequested,
		RequestedBy:       req.RequestedBy,
		RequestedAt:       now,
		SignedBy:          req.Signatory,
		SignedAt:          nil,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := documentVersionSignature.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert document version signature: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return documentVersionSignature, nil
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
	draftVersion := &coredata.DocumentVersion{}
	now := time.Now()

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := latestVersion.LoadLatestVersion(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load latest version: %w", err)
			}

			if latestVersion.Status != coredata.DocumentStatusPublished {
				return fmt.Errorf("cannot create draft from unpublished version")
			}

			draftVersion.ID = draftVersionID
			draftVersion.DocumentID = documentID
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

func (s *DocumentService) ListByOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.DocumentOrderField],
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
			)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(documents, cursor), nil
}

func (s *DocumentService) ListForControlID(
	ctx context.Context,
	controlID gid.GID,
	cursor *page.Cursor[coredata.DocumentOrderField],
) (*page.Page[*coredata.Document, coredata.DocumentOrderField], error) {
	var documents coredata.Documents

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return documents.LoadByControlID(ctx, conn, s.svc.scope, controlID, cursor)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(documents, cursor), nil
}

func (s *DocumentService) ListForRiskID(
	ctx context.Context,
	riskID gid.GID,
	cursor *page.Cursor[coredata.DocumentOrderField],
) (*page.Page[*coredata.Document, coredata.DocumentOrderField], error) {
	var documents coredata.Documents

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return documents.LoadByRiskID(ctx, conn, s.svc.scope, riskID, cursor)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(documents, cursor), nil
}
