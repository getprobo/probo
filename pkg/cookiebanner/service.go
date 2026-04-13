// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package cookiebanner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type Service struct {
	pg *pg.Client
}

func NewService(pgClient *pg.Client) *Service {
	return &Service{pg: pgClient}
}

var defaultCategories = []struct {
	Name        string
	Description string
	Required    bool
	Rank        int
}{
	{"Necessary", "Essential cookies required for the website to function properly.", true, 0},
	{"Analytics", "Cookies that help understand how visitors interact with the website.", false, 1},
	{"Advertising", "Cookies used to deliver relevant advertisements and track campaigns.", false, 2},
	{"Functional", "Cookies that enable enhanced functionality and personalization.", false, 3},
}

type (
	CreateCookieBannerRequest struct {
		OrganizationID    gid.GID
		Name              string
		Origin            string
		PrivacyPolicyURL  string
		ConsentExpiryDays int
		ConsentMode       coredata.CookieConsentMode
	}

	CreateCookieCategoryRequest struct {
		CookieBannerID gid.GID
		Name           string
		Description    string
		Required       bool
		Rank           int
		Cookies        coredata.CookieItems
	}

	UpdateCookieBannerRequest struct {
		CookieBannerID    gid.GID
		Name              *string
		Origin            *string
		PrivacyPolicyURL  *string
		ConsentExpiryDays *int
		ConsentMode       *coredata.CookieConsentMode
	}

	UpdateCookieCategoryRequest struct {
		CookieCategoryID gid.GID
		Name             *string
		Description      *string
		Rank             *int
		Cookies          *coredata.CookieItems
	}

	CreateCookieConsentRecordRequest struct {
		CookieBannerID gid.GID
		Version        int
		VisitorID      string
		IPAddress      *string
		UserAgent      *string
		ConsentData    json.RawMessage
		Action         coredata.CookieConsentAction
	}
)

func (r *CreateCookieBannerRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(255))
	v.Check(r.Origin, "origin", validator.Required(), validator.Origin())
	v.Check(r.PrivacyPolicyURL, "privacy_policy_url", validator.Required(), validator.URL())
	v.Check(r.ConsentExpiryDays, "consent_expiry_days", validator.Required(), validator.Min(1))
	v.Check(r.ConsentMode, "consent_mode", validator.Required(), validator.OneOfSlice(coredata.CookieConsentModes()))

	return v.Error()
}

func (r *UpdateCookieBannerRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieBannerID, "cookie_banner_id", validator.Required(), validator.GID(coredata.CookieBannerEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(255))
	v.Check(r.Origin, "origin", validator.Origin())
	v.Check(r.PrivacyPolicyURL, "privacy_policy_url", validator.URL())
	v.Check(r.ConsentExpiryDays, "consent_expiry_days", validator.Min(1))
	v.Check(r.ConsentMode, "consent_mode", validator.OneOfSlice(coredata.CookieConsentModes()))

	return v.Error()
}

func (r *CreateCookieCategoryRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieBannerID, "cookie_banner_id", validator.Required(), validator.GID(coredata.CookieBannerEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(255))
	v.Check(r.Description, "description", validator.Required(), validator.SafeText(1000))
	v.Check(r.Rank, "rank", validator.Min(0))

	return v.Error()
}

func (r *UpdateCookieCategoryRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieCategoryID, "cookie_category_id", validator.Required(), validator.GID(coredata.CookieCategoryEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(255))
	v.Check(r.Description, "description", validator.SafeText(1000))
	v.Check(r.Rank, "rank", validator.Min(0))

	return v.Error()
}

func (r *CreateCookieConsentRecordRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieBannerID, "cookie_banner_id", validator.Required(), validator.GID(coredata.CookieBannerEntityType))
	v.Check(r.Version, "version", validator.Required(), validator.Min(1))
	v.Check(r.VisitorID, "visitor_id", validator.Required(), validator.NotEmpty())
	v.Check(r.Action, "action", validator.Required(), validator.OneOfSlice(coredata.CookieConsentActions()))

	return v.Error()
}

func buildSnapshot(
	banner *coredata.CookieBanner,
	categories coredata.CookieCategories,
) coredata.CookieBannerVersionSnapshot {
	snapshotCategories := make([]coredata.CookieBannerVersionSnapshotCategory, len(categories))
	for i, c := range categories {
		snapshotCategories[i] = coredata.CookieBannerVersionSnapshotCategory{
			Name:        c.Name,
			Description: c.Description,
			Required:    c.Required,
			Cookies:     c.Cookies,
		}
	}

	return coredata.CookieBannerVersionSnapshot{
		PrivacyPolicyURL:  banner.PrivacyPolicyURL,
		ConsentExpiryDays: banner.ConsentExpiryDays,
		ConsentMode:       string(banner.ConsentMode),
		Categories:        snapshotCategories,
	}
}

func (s *Service) ensureDraftVersion(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	banner *coredata.CookieBanner,
	categories coredata.CookieCategories,
) (*coredata.CookieBannerVersion, error) {
	snapshot := buildSnapshot(banner, categories)

	var latest coredata.CookieBannerVersion
	err := latest.LoadLatestByCookieBannerID(ctx, tx, scope, banner.ID)

	if err == nil && latest.State == coredata.CookieBannerVersionStateDraft {
		if err := latest.SetSnapshot(snapshot); err != nil {
			return nil, fmt.Errorf("cannot set snapshot: %w", err)
		}
		latest.UpdatedAt = time.Now()
		if err := latest.Update(ctx, tx, scope); err != nil {
			return nil, fmt.Errorf("cannot update draft version: %w", err)
		}
		return &latest, nil
	}

	if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load latest version: %w", err)
	}

	now := time.Now()
	version := &coredata.CookieBannerVersion{
		ID:             gid.New(scope.GetTenantID(), coredata.CookieBannerVersionEntityType),
		OrganizationID: banner.OrganizationID,
		CookieBannerID: banner.ID,
		State:          coredata.CookieBannerVersionStateDraft,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	nextVersion, err := version.LoadNextVersion(ctx, tx, scope, banner.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot determine next version: %w", err)
	}
	version.Version = nextVersion

	if err := version.SetSnapshot(snapshot); err != nil {
		return nil, fmt.Errorf("cannot set snapshot: %w", err)
	}

	if err := version.Insert(ctx, tx, scope); err != nil {
		return nil, fmt.Errorf("cannot insert draft version: %w", err)
	}

	return version, nil
}

func (s *Service) CreateCookieBanner(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateCookieBannerRequest,
) (*coredata.CookieBanner, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var banner *coredata.CookieBanner

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			now := time.Now()

			banner = &coredata.CookieBanner{
				ID:                gid.New(scope.GetTenantID(), coredata.CookieBannerEntityType),
				OrganizationID:    req.OrganizationID,
				Name:              req.Name,
				Origin:            req.Origin,
				State:             coredata.CookieBannerStateActive,
				PrivacyPolicyURL:  req.PrivacyPolicyURL,
				ConsentExpiryDays: req.ConsentExpiryDays,
				ConsentMode:       req.ConsentMode,
				CreatedAt:         now,
				UpdatedAt:         now,
			}

			if err := banner.Insert(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrOriginAlreadyInUse
				}
				return fmt.Errorf("cannot insert cookie banner: %w", err)
			}

			for _, dc := range defaultCategories {
				category := &coredata.CookieCategory{
					ID:             gid.New(scope.GetTenantID(), coredata.CookieCategoryEntityType),
					OrganizationID: banner.OrganizationID,
					CookieBannerID: banner.ID,
					Name:           dc.Name,
					Description:    dc.Description,
					Required:       dc.Required,
					Rank:           dc.Rank,
					Cookies:        coredata.CookieItems{},
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				if err := category.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert default cookie category %q: %w", dc.Name, err)
				}
			}

			var categories coredata.CookieCategories
			if err := categories.LoadAllByCookieBannerID(ctx, tx, scope, banner.ID); err != nil {
				return fmt.Errorf("cannot load cookie categories: %w", err)
			}

			if _, err := s.ensureDraftVersion(ctx, tx, scope, banner, categories); err != nil {
				return fmt.Errorf("cannot ensure draft version: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return banner, nil
}

func (s *Service) GetCookieBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
) (*coredata.CookieBanner, error) {
	var banner coredata.CookieBanner

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := banner.LoadByID(ctx, conn, scope, bannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &banner, nil
}

func (s *Service) ListCookieBannersForOrganization(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.CookieBannerOrderField],
	filter *coredata.CookieBannerFilter,
) (coredata.CookieBanners, error) {
	var banners coredata.CookieBanners

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := banners.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor, filter); err != nil {
				return fmt.Errorf("cannot list cookie banners: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return banners, nil
}

func (s *Service) CountCookieBannersForOrganization(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	filter *coredata.CookieBannerFilter,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var banners coredata.CookieBanners
			var err error

			count, err = banners.CountByOrganizationID(ctx, conn, scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count cookie banners: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) UpdateCookieBanner(
	ctx context.Context,
	scope coredata.Scoper,
	req UpdateCookieBannerRequest,
) (*coredata.CookieBanner, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var banner coredata.CookieBanner

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := banner.LoadByID(ctx, tx, scope, req.CookieBannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			consentChanged := req.PrivacyPolicyURL != nil || req.ConsentExpiryDays != nil || req.ConsentMode != nil

			if req.Name != nil {
				banner.Name = *req.Name
			}
			if req.Origin != nil {
				banner.Origin = *req.Origin
			}
			if req.PrivacyPolicyURL != nil {
				banner.PrivacyPolicyURL = *req.PrivacyPolicyURL
			}
			if req.ConsentExpiryDays != nil {
				banner.ConsentExpiryDays = *req.ConsentExpiryDays
			}
			if req.ConsentMode != nil {
				banner.ConsentMode = *req.ConsentMode
			}

			banner.UpdatedAt = time.Now()

			if err := banner.Update(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrOriginAlreadyInUse
				}
				return fmt.Errorf("cannot update cookie banner: %w", err)
			}

			if consentChanged {
				var categories coredata.CookieCategories
				if err := categories.LoadAllByCookieBannerID(ctx, tx, scope, banner.ID); err != nil {
					return fmt.Errorf("cannot load cookie categories: %w", err)
				}

				if _, err := s.ensureDraftVersion(ctx, tx, scope, &banner, categories); err != nil {
					return fmt.Errorf("cannot ensure draft version: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &banner, nil
}

func (s *Service) PublishCookieBannerVersion(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
) (*coredata.CookieBannerVersion, error) {
	var version coredata.CookieBannerVersion

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := version.LoadLatestByCookieBannerID(ctx, tx, scope, bannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrNoDraftVersion
				}
				return fmt.Errorf("cannot load latest version: %w", err)
			}

			if version.State != coredata.CookieBannerVersionStateDraft {
				return ErrNoDraftVersion
			}

			version.State = coredata.CookieBannerVersionStatePublished
			version.UpdatedAt = time.Now()

			if err := version.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot publish version: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &version, nil
}

func (s *Service) ActivateCookieBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
) (*coredata.CookieBanner, error) {
	var banner coredata.CookieBanner

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := banner.LoadByID(ctx, tx, scope, bannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			if banner.State == coredata.CookieBannerStateActive {
				return ErrBannerAlreadyActive
			}

			banner.State = coredata.CookieBannerStateActive
			banner.UpdatedAt = time.Now()

			if err := banner.Update(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrOriginAlreadyInUse
				}
				return fmt.Errorf("cannot update cookie banner: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &banner, nil
}

func (s *Service) DeactivateCookieBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
) (*coredata.CookieBanner, error) {
	var banner coredata.CookieBanner

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := banner.LoadByID(ctx, tx, scope, bannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			if banner.State == coredata.CookieBannerStateInactive {
				return ErrBannerAlreadyInactive
			}

			banner.State = coredata.CookieBannerStateInactive
			banner.UpdatedAt = time.Now()

			if err := banner.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update cookie banner: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &banner, nil
}

func (s *Service) DeleteCookieBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var banner coredata.CookieBanner
			if err := banner.LoadByID(ctx, tx, scope, bannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			if err := banner.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete cookie banner: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) CreateCookieCategory(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateCookieCategoryRequest,
) (*coredata.CookieCategory, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var category *coredata.CookieCategory

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var banner coredata.CookieBanner
			if err := banner.LoadByID(ctx, tx, scope, req.CookieBannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			now := time.Now()

			cookies := req.Cookies
			if cookies == nil {
				cookies = coredata.CookieItems{}
			}

			category = &coredata.CookieCategory{
				ID:             gid.New(scope.GetTenantID(), coredata.CookieCategoryEntityType),
				OrganizationID: banner.OrganizationID,
				CookieBannerID: req.CookieBannerID,
				Name:           req.Name,
				Description:    req.Description,
				Required:       req.Required,
				Rank:           req.Rank,
				Cookies:        cookies,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := category.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert cookie category: %w", err)
			}

			var categories coredata.CookieCategories
			if err := categories.LoadAllByCookieBannerID(ctx, tx, scope, req.CookieBannerID); err != nil {
				return fmt.Errorf("cannot load cookie categories: %w", err)
			}

			if _, err := s.ensureDraftVersion(ctx, tx, scope, &banner, categories); err != nil {
				return fmt.Errorf("cannot ensure draft version: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *Service) GetCookieCategory(
	ctx context.Context,
	scope coredata.Scoper,
	categoryID gid.GID,
) (*coredata.CookieCategory, error) {
	var category coredata.CookieCategory

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := category.LoadByID(ctx, conn, scope, categoryID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrCategoryNotFound
				}
				return fmt.Errorf("cannot load cookie category: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (s *Service) ListCookieCategoriesForBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
	cursor *page.Cursor[coredata.CookieCategoryOrderField],
) (coredata.CookieCategories, error) {
	var categories coredata.CookieCategories

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := categories.LoadByCookieBannerID(ctx, conn, scope, bannerID, cursor); err != nil {
				return fmt.Errorf("cannot list cookie categories: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *Service) CountCookieCategoriesForBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var categories coredata.CookieCategories
			var err error

			count, err = categories.CountByCookieBannerID(ctx, conn, scope, bannerID)
			if err != nil {
				return fmt.Errorf("cannot count cookie categories: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) UpdateCookieCategory(
	ctx context.Context,
	scope coredata.Scoper,
	req UpdateCookieCategoryRequest,
) (*coredata.CookieCategory, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var category coredata.CookieCategory

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := category.LoadByID(ctx, tx, scope, req.CookieCategoryID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrCategoryNotFound
				}
				return fmt.Errorf("cannot load cookie category: %w", err)
			}

			if req.Name != nil {
				category.Name = *req.Name
			}
			if req.Description != nil {
				category.Description = *req.Description
			}
			if req.Rank != nil {
				category.Rank = *req.Rank
			}
			if req.Cookies != nil {
				category.Cookies = *req.Cookies
			}

			category.UpdatedAt = time.Now()

			if err := category.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update cookie category: %w", err)
			}

			var banner coredata.CookieBanner
			if err := banner.LoadByID(ctx, tx, scope, category.CookieBannerID); err != nil {
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			var categories coredata.CookieCategories
			if err := categories.LoadAllByCookieBannerID(ctx, tx, scope, category.CookieBannerID); err != nil {
				return fmt.Errorf("cannot load cookie categories: %w", err)
			}

			if _, err := s.ensureDraftVersion(ctx, tx, scope, &banner, categories); err != nil {
				return fmt.Errorf("cannot ensure draft version: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (s *Service) DeleteCookieCategory(
	ctx context.Context,
	scope coredata.Scoper,
	categoryID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var category coredata.CookieCategory
			if err := category.LoadByID(ctx, tx, scope, categoryID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrCategoryNotFound
				}
				return fmt.Errorf("cannot load cookie category: %w", err)
			}

			if category.Required {
				return ErrCannotDeleteRequiredCategory
			}

			bannerID := category.CookieBannerID

			if err := category.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete cookie category: %w", err)
			}

			var banner coredata.CookieBanner
			if err := banner.LoadByID(ctx, tx, scope, bannerID); err != nil {
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			var categories coredata.CookieCategories
			if err := categories.LoadAllByCookieBannerID(ctx, tx, scope, bannerID); err != nil {
				return fmt.Errorf("cannot load cookie categories: %w", err)
			}

			if _, err := s.ensureDraftVersion(ctx, tx, scope, &banner, categories); err != nil {
				return fmt.Errorf("cannot ensure draft version: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) GetCookieBannerVersion(
	ctx context.Context,
	scope coredata.Scoper,
	versionID gid.GID,
) (*coredata.CookieBannerVersion, error) {
	var version coredata.CookieBannerVersion

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := version.LoadByID(ctx, conn, scope, versionID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrVersionNotFound
				}
				return fmt.Errorf("cannot load cookie banner version: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &version, nil
}

func (s *Service) ListCookieBannerVersionsForBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
	cursor *page.Cursor[coredata.CookieBannerVersionOrderField],
) (coredata.CookieBannerVersions, error) {
	var versions coredata.CookieBannerVersions

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := versions.LoadByCookieBannerID(ctx, conn, scope, bannerID, cursor); err != nil {
				return fmt.Errorf("cannot list cookie banner versions: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return versions, nil
}

func (s *Service) CountCookieBannerVersionsForBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var versions coredata.CookieBannerVersions
			var err error

			count, err = versions.CountByCookieBannerID(ctx, conn, scope, bannerID)
			if err != nil {
				return fmt.Errorf("cannot count cookie banner versions: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) CreateCookieConsentRecord(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateCookieConsentRecordRequest,
) (*coredata.CookieConsentRecord, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var record *coredata.CookieConsentRecord

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var publishedVersion coredata.CookieBannerVersion
			if err := publishedVersion.LoadByCookieBannerIDAndVersion(ctx, tx, scope, req.CookieBannerID, req.Version); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrVersionNotFound
				}
				return fmt.Errorf("cannot load cookie banner version: %w", err)
			}

			if publishedVersion.State != coredata.CookieBannerVersionStatePublished {
				return ErrVersionNotPublished
			}

			record = &coredata.CookieConsentRecord{
				ID:                    gid.New(scope.GetTenantID(), coredata.CookieConsentRecordEntityType),
				OrganizationID:        publishedVersion.OrganizationID,
				CookieBannerID:        req.CookieBannerID,
				CookieBannerVersionID: publishedVersion.ID,
				VisitorID:             req.VisitorID,
				IPAddress:             req.IPAddress,
				UserAgent:             req.UserAgent,
				ConsentData:           req.ConsentData,
				Action:                req.Action,
				CreatedAt:             time.Now(),
			}

			if err := record.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert consent record: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (s *Service) ListCookieConsentRecordsForBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
	cursor *page.Cursor[coredata.CookieConsentRecordOrderField],
	filter *coredata.CookieConsentRecordFilter,
) (coredata.CookieConsentRecords, error) {
	var records coredata.CookieConsentRecords

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := records.LoadByCookieBannerID(ctx, conn, scope, bannerID, cursor, filter); err != nil {
				return fmt.Errorf("cannot list consent records: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (s *Service) CountCookieConsentRecordsForBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
	filter *coredata.CookieConsentRecordFilter,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var records coredata.CookieConsentRecords
			var err error

			count, err = records.CountByCookieBannerID(ctx, conn, scope, bannerID, filter)
			if err != nil {
				return fmt.Errorf("cannot count consent records: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
