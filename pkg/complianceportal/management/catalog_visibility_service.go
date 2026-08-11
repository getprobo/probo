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

func (s *Service) GetAuditLinks(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	auditIDs []gid.GID,
) (coredata.CompliancePortalAudits, error) {
	var rows coredata.CompliancePortalAudits

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			byAuditID, err := coredata.LoadCompliancePortalAuditsByCompliancePortalIDAndAuditIDs(
				ctx,
				conn,
				scope,
				compliancePortalID,
				auditIDs,
			)
			if err != nil {
				return err
			}

			rows = make(coredata.CompliancePortalAudits, 0, len(byAuditID))
			for _, link := range byAuditID {
				rows = append(rows, link)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load portal audit links: %w", err)
	}

	return rows, nil
}

func (s *Service) GetThirdPartyLinks(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	thirdPartyIDs []gid.GID,
) (coredata.CompliancePortalThirdParties, error) {
	var rows coredata.CompliancePortalThirdParties

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			byThirdPartyID, err := coredata.LoadCompliancePortalThirdPartiesByCompliancePortalIDAndThirdPartyIDs(
				ctx,
				conn,
				scope,
				compliancePortalID,
				thirdPartyIDs,
			)
			if err != nil {
				return err
			}

			rows = make(coredata.CompliancePortalThirdParties, 0, len(byThirdPartyID))
			for _, link := range byThirdPartyID {
				rows = append(rows, link)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load portal third party links: %w", err)
	}

	return rows, nil
}
