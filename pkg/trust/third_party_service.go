// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package trust

import (
	"context"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type ThirdPartyService struct {
	svc *Service
}

func (s ThirdPartyService) Get(
	ctx context.Context,
	scope coredata.Scoper,
	thirdPartyID gid.GID,
) (*coredata.ThirdParty, error) {
	thirdParty := &coredata.ThirdParty{}

	err := s.svc.pg.WithConn(
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

func (s ThirdPartyService) ListForOrganizationId(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyOrderField],
	filter *coredata.ThirdPartyFilter,
) (*page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField], error) {
	var thirdParties coredata.ThirdParties

	err := s.svc.pg.WithConn(
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

func (s ThirdPartyService) ListDistinctTrustCenterCategoriesForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) ([]coredata.ThirdPartyCategory, error) {
	var categories []coredata.ThirdPartyCategory

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			thirdParties := &coredata.ThirdParties{}

			result, err := thirdParties.LoadDistinctTrustCenterCategoriesByOrganizationID(ctx, conn, scope, organizationID)
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

func (s ThirdPartyService) ListDistinctTrustCenterCountriesForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) ([]coredata.CountryCode, error) {
	var countries []coredata.CountryCode

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			thirdParties := &coredata.ThirdParties{}

			result, err := thirdParties.LoadDistinctTrustCenterCountriesByOrganizationID(ctx, conn, scope, organizationID)
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

func (s ThirdPartyService) CountForTrustCenterId(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
	filter *coredata.ThirdPartyFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			trustCenter, err := s.svc.TrustCenters.Get(ctx, scope, trustCenterID)
			if err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			thirdParties := &coredata.ThirdParties{}

			count, err = thirdParties.CountByOrganizationID(ctx, conn, scope, trustCenter.OrganizationID, filter)
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
