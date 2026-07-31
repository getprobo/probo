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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	CatalogDocument struct {
		ID         gid.GID
		Document   *coredata.Document
		Visibility coredata.CompliancePortalVisibility
	}

	CatalogAudit struct {
		ID         gid.GID
		Audit      *coredata.Audit
		Visibility coredata.CompliancePortalVisibility
	}

	CatalogThirdParty struct {
		ID         gid.GID
		ThirdParty *coredata.ThirdParty
	}
)

func (c *CatalogDocument) CursorKey(field coredata.DocumentOrderField) page.CursorKey {
	return c.Document.CursorKey(field)
}

func (c *CatalogAudit) CursorKey(field coredata.AuditOrderField) page.CursorKey {
	return c.Audit.CursorKey(field)
}

func (c *CatalogThirdParty) CursorKey(field coredata.ThirdPartyOrderField) page.CursorKey {
	return c.ThirdParty.CursorKey(field)
}

func (s *Service) loadPortalOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
) (gid.GID, error) {
	portal := &coredata.CompliancePortal{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return portal.LoadByID(ctx, conn, scope, compliancePortalID)
		},
	)
	if err != nil {
		return gid.GID{}, fmt.Errorf("cannot load compliance portal: %w", err)
	}

	return portal.OrganizationID, nil
}

func (s *Service) CountDocumentCatalog(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
) (int, error) {
	organizationID, err := s.loadPortalOrganizationID(ctx, scope, compliancePortalID)
	if err != nil {
		return 0, err
	}

	var count int

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var documents coredata.Documents

			var countErr error

			count, countErr = documents.CountPublishedByCompliancePortalID(
				ctx,
				conn,
				scope,
				compliancePortalID,
				organizationID,
			)
			if countErr != nil {
				return countErr
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) ListDocumentCatalog(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	cursor *page.Cursor[coredata.DocumentOrderField],
) (*page.Page[*CatalogDocument, coredata.DocumentOrderField], error) {
	organizationID, err := s.loadPortalOrganizationID(ctx, scope, compliancePortalID)
	if err != nil {
		return nil, err
	}

	var entries []*CatalogDocument

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var documents coredata.Documents

			if err := documents.LoadPublishedByCompliancePortalID(
				ctx,
				conn,
				scope,
				compliancePortalID,
				organizationID,
				cursor,
				coredata.NewDocumentCompliancePortalFilter(),
			); err != nil {
				return fmt.Errorf("cannot load portal documents: %w", err)
			}

			documentIDs := make([]gid.GID, len(documents))
			for i, document := range documents {
				documentIDs[i] = document.ID
			}

			linksByDocumentID := map[gid.GID]*coredata.CompliancePortalDocument{}

			if len(documentIDs) > 0 {
				var loadErr error

				linksByDocumentID, loadErr = coredata.LoadCompliancePortalDocumentsByCompliancePortalIDAndDocumentIDs(
					ctx,
					conn,
					scope,
					compliancePortalID,
					documentIDs,
				)
				if loadErr != nil {
					return fmt.Errorf("cannot load portal document links: %w", loadErr)
				}
			}

			entries = make([]*CatalogDocument, 0, len(documents))
			for _, document := range documents {
				link := linksByDocumentID[document.ID]
				if link == nil {
					continue
				}

				if link.Visibility == coredata.CompliancePortalVisibilityNone {
					continue
				}

				entries = append(entries, &CatalogDocument{
					ID:         link.ID,
					Document:   document,
					Visibility: link.Visibility,
				})
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(entries, cursor), nil
}

func (s *Service) CountAuditCatalog(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
) (int, error) {
	organizationID, err := s.loadPortalOrganizationID(ctx, scope, compliancePortalID)
	if err != nil {
		return 0, err
	}

	var count int

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var audits coredata.Audits

			var countErr error

			count, countErr = audits.CountByCompliancePortalID(
				ctx,
				conn,
				scope,
				compliancePortalID,
				organizationID,
			)
			if countErr != nil {
				return countErr
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) ListAuditCatalog(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	cursor *page.Cursor[coredata.AuditOrderField],
) (*page.Page[*CatalogAudit, coredata.AuditOrderField], error) {
	organizationID, err := s.loadPortalOrganizationID(ctx, scope, compliancePortalID)
	if err != nil {
		return nil, err
	}

	var entries []*CatalogAudit

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var audits coredata.Audits

			if err := audits.LoadByCompliancePortalID(
				ctx,
				conn,
				scope,
				compliancePortalID,
				organizationID,
				cursor,
				coredata.NewAuditCompliancePortalFilter(),
			); err != nil {
				return fmt.Errorf("cannot load portal audits: %w", err)
			}

			auditIDs := make([]gid.GID, len(audits))
			for i, audit := range audits {
				auditIDs[i] = audit.ID
			}

			linksByAuditID := map[gid.GID]*coredata.CompliancePortalAudit{}

			if len(auditIDs) > 0 {
				var loadErr error

				linksByAuditID, loadErr = coredata.LoadCompliancePortalAuditsByCompliancePortalIDAndAuditIDs(
					ctx,
					conn,
					scope,
					compliancePortalID,
					auditIDs,
				)
				if loadErr != nil {
					return fmt.Errorf("cannot load portal audit links: %w", loadErr)
				}
			}

			entries = make([]*CatalogAudit, 0, len(audits))
			for _, audit := range audits {
				link := linksByAuditID[audit.ID]
				if link == nil {
					continue
				}

				if link.Visibility == coredata.CompliancePortalVisibilityNone {
					continue
				}

				entries = append(entries, &CatalogAudit{
					ID:         link.ID,
					Audit:      audit,
					Visibility: link.Visibility,
				})
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(entries, cursor), nil
}

func (s *Service) CountThirdPartyCatalog(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
) (int, error) {
	organizationID, err := s.loadPortalOrganizationID(ctx, scope, compliancePortalID)
	if err != nil {
		return 0, err
	}

	var count int

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var thirdParties coredata.ThirdParties

			var countErr error

			count, countErr = thirdParties.CountByCompliancePortalID(
				ctx,
				conn,
				scope,
				compliancePortalID,
				organizationID,
				nil,
			)
			if countErr != nil {
				return countErr
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) ListThirdPartyCatalog(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyOrderField],
) (*page.Page[*CatalogThirdParty, coredata.ThirdPartyOrderField], error) {
	organizationID, err := s.loadPortalOrganizationID(ctx, scope, compliancePortalID)
	if err != nil {
		return nil, err
	}

	var entries []*CatalogThirdParty

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var thirdParties coredata.ThirdParties

			if err := thirdParties.LoadByCompliancePortalID(
				ctx,
				conn,
				scope,
				compliancePortalID,
				organizationID,
				cursor,
				nil,
			); err != nil {
				return fmt.Errorf("cannot load portal third parties: %w", err)
			}

			thirdPartyIDs := make([]gid.GID, len(thirdParties))
			for i, thirdParty := range thirdParties {
				thirdPartyIDs[i] = thirdParty.ID
			}

			linksByThirdPartyID := map[gid.GID]*coredata.CompliancePortalThirdParty{}

			if len(thirdPartyIDs) > 0 {
				var loadErr error

				linksByThirdPartyID, loadErr = coredata.LoadCompliancePortalThirdPartiesByCompliancePortalIDAndThirdPartyIDs(
					ctx,
					conn,
					scope,
					compliancePortalID,
					thirdPartyIDs,
				)
				if loadErr != nil {
					return fmt.Errorf("cannot load portal third party links: %w", loadErr)
				}
			}

			entries = make([]*CatalogThirdParty, 0, len(thirdParties))
			for _, thirdParty := range thirdParties {
				link := linksByThirdPartyID[thirdParty.ID]
				if link == nil {
					continue
				}

				entries = append(entries, &CatalogThirdParty{
					ID:         link.ID,
					ThirdParty: thirdParty,
				})
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(entries, cursor), nil
}

func (s *Service) GetCatalogDocument(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	documentID gid.GID,
) (*CatalogDocument, error) {
	link := &coredata.CompliancePortalDocument{}
	document := &coredata.Document{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := link.LoadByCompliancePortalIDAndDocumentID(ctx, conn, scope, compliancePortalID, documentID); err != nil {
				return err
			}

			if link.Visibility == coredata.CompliancePortalVisibilityNone {
				return coredata.ErrResourceNotFound
			}

			return document.LoadByID(ctx, conn, scope, documentID)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, coredata.ErrResourceNotFound
		}

		return nil, fmt.Errorf("cannot load portal document catalog entry: %w", err)
	}

	return &CatalogDocument{
		ID:         link.ID,
		Document:   document,
		Visibility: link.Visibility,
	}, nil
}

func (s *Service) GetCatalogAudit(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	auditID gid.GID,
) (*CatalogAudit, error) {
	link := &coredata.CompliancePortalAudit{}
	audit := &coredata.Audit{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := link.LoadByCompliancePortalIDAndAuditID(ctx, conn, scope, compliancePortalID, auditID); err != nil {
				return err
			}

			if link.Visibility == coredata.CompliancePortalVisibilityNone {
				return coredata.ErrResourceNotFound
			}

			return audit.LoadByID(ctx, conn, scope, auditID)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, coredata.ErrResourceNotFound
		}

		return nil, fmt.Errorf("cannot load portal audit catalog entry: %w", err)
	}

	return &CatalogAudit{
		ID:         link.ID,
		Audit:      audit,
		Visibility: link.Visibility,
	}, nil
}

func (s *Service) GetCatalogThirdParty(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	thirdPartyID gid.GID,
) (*CatalogThirdParty, error) {
	link := &coredata.CompliancePortalThirdParty{}
	thirdParty := &coredata.ThirdParty{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := link.LoadByCompliancePortalIDAndThirdPartyID(ctx, conn, scope, compliancePortalID, thirdPartyID); err != nil {
				return err
			}

			return thirdParty.LoadByID(ctx, conn, scope, thirdPartyID)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, coredata.ErrResourceNotFound
		}

		return nil, fmt.Errorf("cannot load portal third party catalog entry: %w", err)
	}

	return &CatalogThirdParty{
		ID:         link.ID,
		ThirdParty: thirdParty,
	}, nil
}
