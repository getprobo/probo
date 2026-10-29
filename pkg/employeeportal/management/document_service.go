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

package management

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	AddEmployeePortalDocumentRequest struct {
		EmployeePortalID gid.GID
		DocumentID       gid.GID
	}

	DeleteEmployeePortalDocumentRequest struct {
		ID gid.GID
	}
)

func (r *AddEmployeePortalDocumentRequest) Validate() error {
	v := validator.New()

	v.Check(
		r.EmployeePortalID,
		"employee_portal_id",
		validator.Required(),
		validator.GID(coredata.EmployeePortalEntityType),
	)
	v.Check(
		r.DocumentID,
		"document_id",
		validator.Required(),
		validator.GID(coredata.DocumentEntityType),
	)

	return v.Error()
}

func (r *DeleteEmployeePortalDocumentRequest) Validate() error {
	v := validator.New()

	v.Check(
		r.ID,
		"id",
		validator.Required(),
		validator.GID(coredata.EmployeePortalDocumentEntityType),
	)

	return v.Error()
}

func (s *Service) AddDocument(
	ctx context.Context,
	scope coredata.Scoper,
	req *AddEmployeePortalDocumentRequest,
) (*coredata.EmployeePortalDocument, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var link *coredata.EmployeePortalDocument

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			portal := &coredata.EmployeePortal{}
			if err := portal.LoadByID(ctx, tx, scope, req.EmployeePortalID); err != nil {
				return fmt.Errorf("cannot load employee portal: %w", err)
			}

			document := &coredata.Document{}
			if err := document.LoadByID(ctx, tx, scope, req.DocumentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.OrganizationID != portal.OrganizationID {
				return coredata.ErrResourceNotFound
			}

			if !documentPublished(document) {
				return validator.ValidationErrors{
					{
						Field:   "document_id",
						Code:    validator.ErrorCodeCustom,
						Message: "document must be published",
					},
				}
			}

			now := time.Now()
			link = &coredata.EmployeePortalDocument{
				ID:               gid.New(scope.GetTenantID(), coredata.EmployeePortalDocumentEntityType),
				OrganizationID:   portal.OrganizationID,
				EmployeePortalID: req.EmployeePortalID,
				DocumentID:       req.DocumentID,
				CreatedAt:        now,
				UpdatedAt:        now,
			}

			if err := link.Upsert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot upsert employee portal document: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot add employee portal document: %w", err)
	}

	return link, nil
}

func (s *Service) DeleteDocument(
	ctx context.Context,
	scope coredata.Scoper,
	req *DeleteEmployeePortalDocumentRequest,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := coredata.DeleteEmployeePortalDocumentByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot delete employee portal document: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) GetDocument(
	ctx context.Context,
	scope coredata.Scoper,
	linkID gid.GID,
) (*coredata.EmployeePortalDocument, error) {
	var link *coredata.EmployeePortalDocument

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			link = &coredata.EmployeePortalDocument{}
			if err := link.LoadByID(ctx, conn, scope, linkID); err != nil {
				return fmt.Errorf("cannot load employee portal document: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (s *Service) GetDocumentLinks(
	ctx context.Context,
	scope coredata.Scoper,
	employeePortalID gid.GID,
	documentIDs []gid.GID,
) (coredata.EmployeePortalDocuments, error) {
	var links coredata.EmployeePortalDocuments

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return links.LoadByEmployeePortalIDAndDocumentIDs(
				ctx,
				conn,
				scope,
				employeePortalID,
				documentIDs,
			)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load employee portal documents: %w", err)
	}

	return links, nil
}

func (s *Service) ListDocuments(
	ctx context.Context,
	scope coredata.Scoper,
	employeePortalID gid.GID,
	cursor *page.Cursor[coredata.EmployeePortalDocumentOrderField],
) (*page.Page[*coredata.EmployeePortalDocument, coredata.EmployeePortalDocumentOrderField], error) {
	var links coredata.EmployeePortalDocuments

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := (&coredata.EmployeePortal{}).LoadByID(ctx, conn, scope, employeePortalID); err != nil {
				return fmt.Errorf("cannot load employee portal: %w", err)
			}

			publishedIDs, err := publishedLinkedDocumentIDs(ctx, conn, scope, employeePortalID)
			if err != nil {
				return err
			}

			if err := links.LoadByEmployeePortalID(
				ctx,
				conn,
				scope,
				employeePortalID,
				publishedIDs,
				cursor,
			); err != nil {
				return fmt.Errorf("cannot list employee portal documents: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(links, cursor), nil
}

func (s *Service) CountDocuments(
	ctx context.Context,
	scope coredata.Scoper,
	employeePortalID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := (&coredata.EmployeePortal{}).LoadByID(ctx, conn, scope, employeePortalID); err != nil {
				return fmt.Errorf("cannot load employee portal: %w", err)
			}

			publishedIDs, err := publishedLinkedDocumentIDs(ctx, conn, scope, employeePortalID)
			if err != nil {
				return err
			}

			count = len(publishedIDs)

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func publishedLinkedDocumentIDs(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	employeePortalID gid.GID,
) ([]gid.GID, error) {
	links, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.EmployeePortalDocumentOrderField]{
			Field:     coredata.EmployeePortalDocumentOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.EmployeePortalDocumentOrderField]) ([]*coredata.EmployeePortalDocument, error) {
			var batch coredata.EmployeePortalDocuments
			if err := batch.LoadByEmployeePortalID(
				ctx,
				conn,
				scope,
				employeePortalID,
				nil,
				cursor,
			); err != nil {
				return nil, fmt.Errorf("cannot list employee portal documents: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if len(links) == 0 {
		return nil, nil
	}

	documentIDs := make([]gid.GID, len(links))
	for i, link := range links {
		documentIDs[i] = link.DocumentID
	}

	var documents coredata.Documents
	if err := documents.LoadByIDs(ctx, conn, scope, documentIDs); err != nil {
		return nil, fmt.Errorf("cannot load documents: %w", err)
	}

	publishedIDs := make([]gid.GID, 0, len(documents))
	for _, document := range documents {
		if documentPublished(document) {
			publishedIDs = append(publishedIDs, document.ID)
		}
	}

	return publishedIDs, nil
}

func (s *Service) CurrentPublishedVersions(
	ctx context.Context,
	scope coredata.Scoper,
	documentIDs []gid.GID,
) ([]*coredata.DocumentVersion, error) {
	var versions []*coredata.DocumentVersion

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if len(documentIDs) == 0 {
				return nil
			}

			var documents coredata.Documents
			if err := documents.LoadByIDs(ctx, conn, scope, documentIDs); err != nil {
				return fmt.Errorf("cannot load documents: %w", err)
			}

			for _, document := range documents {
				if !documentPublished(document) {
					continue
				}

				version := &coredata.DocumentVersion{}
				if err := version.LoadByDocumentIDAndVersion(
					ctx,
					conn,
					scope,
					document.ID,
					*document.CurrentPublishedMajor,
					*document.CurrentPublishedMinor,
				); err != nil {
					return fmt.Errorf("cannot load published document version: %w", err)
				}

				versions = append(versions, version)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return versions, nil
}

func (s *Service) IsSharedPublishedVersion(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	major int,
	minor int,
) (bool, error) {
	shared := false

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			document := &coredata.Document{}
			if err := document.LoadByID(ctx, conn, scope, documentID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load document: %w", err)
			}

			if !documentPublished(document) {
				return nil
			}

			if document.CurrentPublishedMajor == nil || document.CurrentPublishedMinor == nil {
				return nil
			}

			if *document.CurrentPublishedMajor != major || *document.CurrentPublishedMinor != minor {
				return nil
			}

			link := &coredata.EmployeePortalDocument{}
			err := link.LoadByOrganizationIDAndDocumentID(
				ctx,
				conn,
				scope,
				document.OrganizationID,
				document.ID,
			)
			if err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load employee portal document: %w", err)
			}

			shared = true

			return nil
		},
	)
	if err != nil {
		return false, fmt.Errorf("cannot check shared published document: %w", err)
	}

	return shared, nil
}

func documentPublished(document *coredata.Document) bool {
	return document.Status == coredata.DocumentStatusActive &&
		document.CurrentPublishedMajor != nil &&
		document.CurrentPublishedMinor != nil
}
