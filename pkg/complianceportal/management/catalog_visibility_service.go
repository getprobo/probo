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
)

func (s *Service) GetDocumentLinks(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	documentIDs []gid.GID,
) (coredata.CompliancePortalDocuments, error) {
	var rows coredata.CompliancePortalDocuments

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return rows.LoadByCompliancePortalIDAndDocumentIDs(
				ctx,
				conn,
				scope,
				compliancePortalID,
				documentIDs,
			)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load portal document links: %w", err)
	}

	return rows, nil
}

// GetDocumentVisibility reports how a document is published on one compliance
// portal. Publication lives on the portal/document join row, so a document with
// no row for that portal is simply not published there.
func (s *Service) GetDocumentVisibility(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	documentID gid.GID,
) (coredata.CompliancePortalVisibility, error) {
	row := &coredata.CompliancePortalDocument{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return row.LoadByCompliancePortalIDAndDocumentID(ctx, conn, scope, compliancePortalID, documentID)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.CompliancePortalVisibilityNone, nil
		}

		return "", fmt.Errorf("cannot load portal document: %w", err)
	}

	return row.Visibility, nil
}

// GetAuditVisibility reports how an audit is published on one compliance
// portal.
func (s *Service) GetAuditVisibility(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	auditID gid.GID,
) (coredata.CompliancePortalVisibility, error) {
	row := &coredata.CompliancePortalAudit{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return row.LoadByCompliancePortalIDAndAuditID(ctx, conn, scope, compliancePortalID, auditID)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.CompliancePortalVisibilityNone, nil
		}

		return "", fmt.Errorf("cannot load portal audit: %w", err)
	}

	return row.Visibility, nil
}

// IsThirdPartyPublished reports whether a third party is listed as a
// subprocessor on one compliance portal.
func (s *Service) IsThirdPartyPublished(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	thirdPartyID gid.GID,
) (bool, error) {
	row := &coredata.CompliancePortalThirdParty{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return row.LoadByCompliancePortalIDAndThirdPartyID(ctx, conn, scope, compliancePortalID, thirdPartyID)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("cannot load portal third party: %w", err)
	}

	return true, nil
}
