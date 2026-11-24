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

package accessreview

import (
	"context"
	"errors"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func (s *Service) GetInventoryGrant(
	ctx context.Context,
	scope coredata.Scoper,
	grantID gid.GID,
) (*coredata.AccessInventoryGrant, error) {
	grant := &coredata.AccessInventoryGrant{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := grant.LoadByID(ctx, conn, scope, grantID); err != nil {
				return fmt.Errorf("cannot load access inventory grant: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Service) GetInventoryPrincipal(
	ctx context.Context,
	scope coredata.Scoper,
	principalID gid.GID,
) (*coredata.AccessInventoryPrincipal, error) {
	principal := &coredata.AccessInventoryPrincipal{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := principal.LoadByID(ctx, conn, scope, principalID); err != nil {
				return fmt.Errorf("cannot load access inventory principal: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return principal, nil
}

func (s *Service) GetInventoryResource(
	ctx context.Context,
	scope coredata.Scoper,
	resourceID gid.GID,
) (*coredata.AccessInventoryResource, error) {
	resource := &coredata.AccessInventoryResource{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := resource.LoadByID(ctx, conn, scope, resourceID); err != nil {
				return fmt.Errorf("cannot load access inventory resource: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (s *Service) ListInventoryGrantsForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AccessInventoryGrantOrderField],
	filter *coredata.AccessInventoryGrantFilter,
) (*page.Page[*coredata.AccessInventoryGrant, coredata.AccessInventoryGrantOrderField], error) {
	var grants coredata.AccessInventoryGrants

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			resolved, err := s.resolveInventoryGrantFilter(ctx, conn, scope, organizationID, filter)
			if err != nil {
				return err
			}

			if err := grants.LoadByOrganizationID(
				ctx,
				conn,
				scope,
				organizationID,
				cursor,
				resolved,
			); err != nil {
				return fmt.Errorf("cannot load access inventory grants: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(grants, cursor), nil
}

func (s *Service) CountInventoryGrantsForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	filter *coredata.AccessInventoryGrantFilter,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			resolved, err := s.resolveInventoryGrantFilter(ctx, conn, scope, organizationID, filter)
			if err != nil {
				return err
			}

			grants := coredata.AccessInventoryGrants{}

			count, err = grants.CountByOrganizationID(ctx, conn, scope, organizationID, resolved)
			if err != nil {
				return fmt.Errorf("cannot count access inventory grants: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) ListInventoryPrincipalsForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AccessInventoryPrincipalOrderField],
	filter *coredata.AccessInventoryPrincipalFilter,
) (*page.Page[*coredata.AccessInventoryPrincipal, coredata.AccessInventoryPrincipalOrderField], error) {
	var principals coredata.AccessInventoryPrincipals

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := principals.LoadByOrganizationID(
				ctx,
				conn,
				scope,
				organizationID,
				cursor,
				filter,
			); err != nil {
				return fmt.Errorf("cannot load access inventory principals: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(principals, cursor), nil
}

func (s *Service) CountInventoryPrincipalsForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	filter *coredata.AccessInventoryPrincipalFilter,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			principals := coredata.AccessInventoryPrincipals{}

			var err error

			count, err = principals.CountByOrganizationID(ctx, conn, scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count access inventory principals: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) ListInventoryResourcesForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AccessInventoryResourceOrderField],
	filter *coredata.AccessInventoryResourceFilter,
) (*page.Page[*coredata.AccessInventoryResource, coredata.AccessInventoryResourceOrderField], error) {
	var resources coredata.AccessInventoryResources

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := resources.LoadByOrganizationID(
				ctx,
				conn,
				scope,
				organizationID,
				cursor,
				filter,
			); err != nil {
				return fmt.Errorf("cannot load access inventory resources: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(resources, cursor), nil
}

func (s *Service) CountInventoryResourcesForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	filter *coredata.AccessInventoryResourceFilter,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			resources := coredata.AccessInventoryResources{}

			var err error

			count, err = resources.CountByOrganizationID(ctx, conn, scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count access inventory resources: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) resolveInventoryGrantFilter(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	organizationID gid.GID,
	filter *coredata.AccessInventoryGrantFilter,
) (*coredata.AccessInventoryGrantFilter, error) {
	settings := &coredata.AccessInventorySettings{}
	err := settings.LoadByOrganizationID(ctx, conn, scope, organizationID)
	if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load access inventory settings: %w", err)
	}

	if errors.Is(err, coredata.ErrResourceNotFound) {
		settings = nil
	}

	hideAfter := coredata.ResolveHideAfterMissingSyncs(settings)

	return filter.WithHideAfterMissingSyncs(hideAfter), nil
}
