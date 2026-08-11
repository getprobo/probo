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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
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
