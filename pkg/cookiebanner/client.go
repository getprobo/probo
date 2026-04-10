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

type Client struct {
	pg *pg.Client
}

func NewClient(pgClient *pg.Client) *Client {
	return &Client{pg: pgClient}
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
	v.Check(r.Origin, "origin", validator.Required(), validator.NotEmpty())
	v.Check(r.PrivacyPolicyURL, "privacy_policy_url", validator.Required(), validator.URL())
	v.Check(r.ConsentExpiryDays, "consent_expiry_days", validator.Required(), validator.Min(1))
	v.Check(r.ConsentMode, "consent_mode", validator.Required(), validator.OneOfSlice(coredata.CookieConsentModes()))

	return v.Error()
}

func (r *UpdateCookieBannerRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieBannerID, "cookie_banner_id", validator.Required(), validator.GID(coredata.CookieBannerEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(255))
	v.Check(r.Origin, "origin", validator.NotEmpty())
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
	v.Check(r.VisitorID, "visitor_id", validator.Required(), validator.NotEmpty())
	v.Check(r.Action, "action", validator.Required(), validator.OneOfSlice(coredata.CookieConsentActions()))

	return v.Error()
}

func (c *Client) CreateCookieBanner(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateCookieBannerRequest,
) (*coredata.CookieBanner, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var banner *coredata.CookieBanner

	err := c.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			now := time.Now()

			banner = &coredata.CookieBanner{
				ID:                gid.New(scope.GetTenantID(), coredata.CookieBannerEntityType),
				OrganizationID:    req.OrganizationID,
				Name:              req.Name,
				Origin:            req.Origin,
				State:             coredata.CookieBannerStateDraft,
				PrivacyPolicyURL:  req.PrivacyPolicyURL,
				ConsentExpiryDays: req.ConsentExpiryDays,
				ConsentMode:       req.ConsentMode,
				CreatedAt:         now,
				UpdatedAt:         now,
			}

			if err := banner.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert cookie banner: %w", err)
			}

			for _, dc := range defaultCategories {
				category := &coredata.CookieCategory{
					ID:             gid.New(scope.GetTenantID(), coredata.CookieCategoryEntityType),
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

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return banner, nil
}

func (c *Client) GetCookieBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
) (*coredata.CookieBanner, error) {
	var banner coredata.CookieBanner

	err := c.pg.WithConn(
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

func (c *Client) ListCookieBannersForOrganization(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.CookieBannerOrderField],
	filter *coredata.CookieBannerFilter,
) (coredata.CookieBanners, error) {
	var banners coredata.CookieBanners

	err := c.pg.WithConn(
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

func (c *Client) CountCookieBannersForOrganization(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	filter *coredata.CookieBannerFilter,
) (int, error) {
	var count int

	err := c.pg.WithConn(
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

func (c *Client) UpdateCookieBanner(
	ctx context.Context,
	scope coredata.Scoper,
	req UpdateCookieBannerRequest,
) (*coredata.CookieBanner, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var banner coredata.CookieBanner

	err := c.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := banner.LoadByID(ctx, tx, scope, req.CookieBannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			// TODO: remove this guard once we add versioning.
			if banner.State != coredata.CookieBannerStateDraft {
				return ErrBannerNotDraft
			}

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

func (c *Client) PublishCookieBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
) (*coredata.CookieBanner, error) {
	var banner coredata.CookieBanner

	err := c.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := banner.LoadByID(ctx, tx, scope, bannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			if banner.State == coredata.CookieBannerStatePublished {
				return ErrBannerAlreadyPublished
			}

			banner.State = coredata.CookieBannerStatePublished
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

func (c *Client) DisableCookieBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
) (*coredata.CookieBanner, error) {
	var banner coredata.CookieBanner

	err := c.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := banner.LoadByID(ctx, tx, scope, bannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			if banner.State == coredata.CookieBannerStateDisabled {
				return ErrBannerAlreadyDisabled
			}

			banner.State = coredata.CookieBannerStateDisabled
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

func (c *Client) DeleteCookieBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
) error {
	return c.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var banner coredata.CookieBanner
			if err := banner.LoadByID(ctx, tx, scope, bannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			if banner.State != coredata.CookieBannerStateDraft {
				return ErrBannerNotDraft
			}

			if err := banner.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete cookie banner: %w", err)
			}

			return nil
		},
	)
}

func (c *Client) CreateCookieCategory(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateCookieCategoryRequest,
) (*coredata.CookieCategory, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var category *coredata.CookieCategory

	err := c.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			now := time.Now()

			cookies := req.Cookies
			if cookies == nil {
				cookies = coredata.CookieItems{}
			}

			category = &coredata.CookieCategory{
				ID:             gid.New(scope.GetTenantID(), coredata.CookieCategoryEntityType),
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

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (c *Client) GetCookieCategory(
	ctx context.Context,
	scope coredata.Scoper,
	categoryID gid.GID,
) (*coredata.CookieCategory, error) {
	var category coredata.CookieCategory

	err := c.pg.WithConn(
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

func (c *Client) ListCookieCategoriesForBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
	cursor *page.Cursor[coredata.CookieCategoryOrderField],
) (coredata.CookieCategories, error) {
	var categories coredata.CookieCategories

	err := c.pg.WithConn(
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

func (c *Client) CountCookieCategoriesForBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
) (int, error) {
	var count int

	err := c.pg.WithConn(
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

func (c *Client) UpdateCookieCategory(
	ctx context.Context,
	scope coredata.Scoper,
	req UpdateCookieCategoryRequest,
) (*coredata.CookieCategory, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var category coredata.CookieCategory

	err := c.pg.WithTx(
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

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (c *Client) DeleteCookieCategory(
	ctx context.Context,
	scope coredata.Scoper,
	categoryID gid.GID,
) error {
	return c.pg.WithTx(
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

			if err := category.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete cookie category: %w", err)
			}

			return nil
		},
	)
}

func (c *Client) CreateCookieConsentRecord(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateCookieConsentRecordRequest,
) (*coredata.CookieConsentRecord, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var record *coredata.CookieConsentRecord

	err := c.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			record = &coredata.CookieConsentRecord{
				ID:             gid.New(scope.GetTenantID(), coredata.CookieConsentRecordEntityType),
				CookieBannerID: req.CookieBannerID,
				VisitorID:      req.VisitorID,
				IPAddress:      req.IPAddress,
				UserAgent:      req.UserAgent,
				ConsentData:    req.ConsentData,
				Action:         req.Action,
				CreatedAt:      time.Now(),
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

func (c *Client) ListCookieConsentRecordsForBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
	cursor *page.Cursor[coredata.CookieConsentRecordOrderField],
	filter *coredata.CookieConsentRecordFilter,
) (coredata.CookieConsentRecords, error) {
	var records coredata.CookieConsentRecords

	err := c.pg.WithConn(
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

func (c *Client) CountCookieConsentRecordsForBanner(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
	filter *coredata.CookieConsentRecordFilter,
) (int, error) {
	var count int

	err := c.pg.WithConn(
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
