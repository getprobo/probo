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

package visitor

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func (s *Service) GetFramework(
	ctx context.Context,
	scope coredata.Scoper,
	frameworkID gid.GID,
) (*coredata.Framework, error) {
	framework := &coredata.Framework{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := framework.LoadByID(ctx, conn, scope, frameworkID)
			if err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return framework, nil
}

// GetFrameworkForCompliancePortalID loads a framework as exposed by the given
// compliance portal. A portal exposes a framework either by listing it as one
// of its compliance frameworks or by publishing an audit for it — the two paths
// through which the visitor API hands out frameworks. Any other framework,
// including one belonging to another organization, is reported as not found.
func (s *Service) GetFrameworkForCompliancePortalID(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	frameworkID gid.GID,
) (*coredata.Framework, error) {
	framework := &coredata.Framework{}
	compliancePortal := &coredata.CompliancePortal{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := loadPortalByID(ctx, conn, scope, compliancePortalID, compliancePortal); err != nil {
				return err
			}

			complianceFramework := &coredata.ComplianceFramework{}

			err := complianceFramework.LoadByCompliancePortalIDAndFrameworkID(ctx, conn, scope, compliancePortalID, frameworkID)
			if err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load compliance framework: %w", err)
				}

				portalAudit := &coredata.CompliancePortalAudit{}

				err = portalAudit.LoadByCompliancePortalIDAndFrameworkID(ctx, conn, scope, compliancePortalID, frameworkID)
				if err != nil {
					if errors.Is(err, coredata.ErrResourceNotFound) {
						return ErrFrameworkNotFound
					}

					return fmt.Errorf("cannot load compliance portal audit for framework: %w", err)
				}
			}

			if err := framework.LoadByID(ctx, conn, scope, frameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return framework, nil
}

// frameworkNamesByAuditID loads every framework linked to the given audits in
// two queries. Names stay in framework creation order, matching
// ListFrameworksForAuditID.
func (s *Service) frameworkNamesByAuditID(
	ctx context.Context,
	scope coredata.Scoper,
	auditIDs []gid.GID,
) (map[gid.GID][]string, error) {
	namesByAudit := make(map[gid.GID][]string, len(auditIDs))

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var links coredata.AuditFrameworks

			if err := links.LoadByAuditIDs(ctx, conn, scope, auditIDs); err != nil {
				return fmt.Errorf("cannot load audit frameworks: %w", err)
			}

			frameworkIDs := make([]gid.GID, 0, len(links))
			seen := make(map[gid.GID]struct{}, len(links))

			for _, link := range links {
				if _, ok := seen[link.FrameworkID]; ok {
					continue
				}

				seen[link.FrameworkID] = struct{}{}
				frameworkIDs = append(frameworkIDs, link.FrameworkID)
			}

			byID := make(map[gid.GID]*coredata.Framework, len(frameworkIDs))

			if len(frameworkIDs) > 0 {
				var frameworks coredata.Frameworks

				if err := frameworks.LoadByIDs(ctx, conn, scope, frameworkIDs); err != nil {
					return fmt.Errorf("cannot load frameworks: %w", err)
				}

				for _, framework := range frameworks {
					byID[framework.ID] = framework
				}
			}

			grouped := make(map[gid.GID][]*coredata.Framework, len(auditIDs))

			for _, link := range links {
				framework, ok := byID[link.FrameworkID]
				if !ok {
					continue
				}

				grouped[link.AuditID] = append(grouped[link.AuditID], framework)
			}

			for auditID, frameworks := range grouped {
				sort.Slice(frameworks, func(i, j int) bool {
					if frameworks[i].CreatedAt.Equal(frameworks[j].CreatedAt) {
						return frameworks[i].ID.String() < frameworks[j].ID.String()
					}

					return frameworks[i].CreatedAt.Before(frameworks[j].CreatedAt)
				})

				names := make([]string, len(frameworks))

				for i, framework := range frameworks {
					names[i] = framework.Name
				}

				namesByAudit[auditID] = names
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return namesByAudit, nil
}

func (s *Service) ListFrameworksForAuditID(
	ctx context.Context,
	scope coredata.Scoper,
	auditID gid.GID,
) (coredata.Frameworks, error) {
	return page.LoadAll(
		ctx,
		page.OrderBy[coredata.FrameworkOrderField]{
			Field:     coredata.FrameworkOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.FrameworkOrderField]) ([]*coredata.Framework, error) {
			var batch coredata.Frameworks

			err := s.pg.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					if err := batch.LoadByAuditID(ctx, conn, scope, auditID, cursor); err != nil {
						return fmt.Errorf("cannot load frameworks: %w", err)
					}

					return nil
				},
			)
			if err != nil {
				return nil, err
			}

			return batch, nil
		},
	)
}
