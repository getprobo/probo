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
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func (s *Service) GetThirdParty(
	ctx context.Context,
	scope coredata.Scoper,
	thirdPartyID gid.GID,
) (*coredata.ThirdParty, error) {
	thirdParty := &coredata.ThirdParty{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := thirdParty.LoadByID(ctx, conn, scope, thirdPartyID)
			if err != nil {
				return fmt.Errorf("cannot load thirdParty: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return thirdParty, nil
}

func (s *Service) ListThirdPartiesForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyOrderField],
	filter *coredata.ThirdPartyFilter,
) (*page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField], error) {
	if filter == nil {
		showOnCompliancePortal := true
		filter = coredata.NewThirdPartyFilter(&showOnCompliancePortal, nil, nil, nil, nil)
	}

	var thirdParties coredata.ThirdParties

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := thirdParties.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load thirdParties: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(thirdParties, cursor), nil
}

func (s *Service) ListDistinctPortalCategoriesForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) ([]coredata.ThirdPartyCategory, error) {
	var categories []coredata.ThirdPartyCategory

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			thirdParties := &coredata.ThirdParties{}

			result, err := thirdParties.LoadDistinctCompliancePortalCategoriesByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot load thirdParty categories: %w", err)
			}

			categories = result

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *Service) ListDistinctPortalCountriesForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) ([]coredata.CountryCode, error) {
	var countries []coredata.CountryCode

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			thirdParties := &coredata.ThirdParties{}

			result, err := thirdParties.LoadDistinctCompliancePortalCountriesByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot load thirdParty countries: %w", err)
			}

			countries = result

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return countries, nil
}

func (s *Service) CountThirdPartiesForPortalID(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	filter *coredata.ThirdPartyFilter,
) (int, error) {
	if filter == nil {
		showOnCompliancePortal := true
		filter = coredata.NewThirdPartyFilter(&showOnCompliancePortal, nil, nil, nil, nil)
	}

	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			compliancePage, err := s.GetPortal(ctx, scope, compliancePageID)
			if err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			thirdParties := &coredata.ThirdParties{}

			count, err = thirdParties.CountByOrganizationID(ctx, conn, scope, compliancePage.OrganizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count thirdParties: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
