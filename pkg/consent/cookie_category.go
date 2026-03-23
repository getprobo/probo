// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package consent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/cookieprovider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	CreateCookieCategoryRequest struct {
		CookieBannerID gid.GID
		Name           string
		Description    string
		Rank           int
		Cookies        coredata.CookieItems
	}

	UpdateCookieCategoryRequest struct {
		ID          gid.GID
		Name        *string
		Description *string
		Rank        *int
		Cookies     *coredata.CookieItems
	}

	AddCookiesFromProviderRequest struct {
		CookieCategoryID gid.GID
		ProviderKey      string
	}
)

var ErrCannotDeleteRequiredCategory = errors.New("cannot delete a required cookie category")

func (r *CreateCookieCategoryRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieBannerID, "cookie_banner_id", validator.Required(), validator.GID(coredata.CookieBannerEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(NameMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (r *UpdateCookieCategoryRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.CookieCategoryEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(NameMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (s *Service) GetCookieCategory(
	ctx context.Context,
	categoryID gid.GID,
) (*coredata.CookieCategory, error) {
	scope := coredata.NewScopeFromObjectID(categoryID)
	category := &coredata.CookieCategory{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := category.LoadByID(ctx, conn, scope, categoryID); err != nil {
				return fmt.Errorf("cannot load cookie category: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get cookie category: %w", err)
	}

	return category, nil
}

func (s *Service) ListCookieCategoriesForCookieBannerID(
	ctx context.Context,
	cookieBannerID gid.GID,
	cursor *page.Cursor[coredata.CookieCategoryOrderField],
) (*page.Page[*coredata.CookieCategory, coredata.CookieCategoryOrderField], error) {
	scope := coredata.NewScopeFromObjectID(cookieBannerID)
	var categories coredata.CookieCategories

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := categories.LoadByCookieBannerID(
				ctx,
				conn,
				scope,
				cookieBannerID,
				cursor,
			); err != nil {
				return fmt.Errorf("cannot load cookie categories: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list cookie categories: %w", err)
	}

	return page.NewPage(categories, cursor), nil
}

func (s *Service) CountCookieCategoriesForCookieBannerID(
	ctx context.Context,
	cookieBannerID gid.GID,
) (int, error) {
	var (
		scope = coredata.NewScopeFromObjectID(cookieBannerID)
		count int
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			categories := coredata.CookieCategories{}
			count, err = categories.CountByCookieBannerID(ctx, conn, scope, cookieBannerID)
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

func (s *Service) CreateCookieCategory(
	ctx context.Context,
	req CreateCookieCategoryRequest,
) (*coredata.CookieCategory, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var (
		scope      = coredata.NewScopeFromObjectID(req.CookieBannerID)
		now        = time.Now()
		categoryID = gid.New(req.CookieBannerID.TenantID(), coredata.CookieCategoryEntityType)
		cookies    = req.Cookies
	)

	if cookies == nil {
		cookies = make(coredata.CookieItems, 0)
	}

	category := &coredata.CookieCategory{
		ID:             categoryID,
		CookieBannerID: req.CookieBannerID,
		Name:           req.Name,
		Description:    req.Description,
		Required:       false,
		Rank:           req.Rank,
		Cookies:        cookies,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := category.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert cookie category: %w", err)
			}

			banner := &coredata.CookieBanner{}
			if err := banner.LoadByID(ctx, conn, scope, req.CookieBannerID); err != nil {
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			banner.Version++
			banner.UpdatedAt = now

			if err := banner.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update cookie banner version: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create cookie category: %w", err)
	}

	return category, nil
}

func (s *Service) UpdateCookieCategory(
	ctx context.Context,
	req UpdateCookieCategoryRequest,
) (*coredata.CookieCategory, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	scope := coredata.NewScopeFromObjectID(req.ID)
	category := &coredata.CookieCategory{ID: req.ID}

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := category.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load cookie category: %w", err)
			}

			bumpVersion := false
			now := time.Now()
			category.UpdatedAt = now

			if req.Name != nil {
				if *req.Name != category.Name {
					bumpVersion = true
				}
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

			if err := category.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update cookie category: %w", err)
			}

			if bumpVersion {
				banner := &coredata.CookieBanner{}
				if err := banner.LoadByID(ctx, conn, scope, category.CookieBannerID); err != nil {
					return fmt.Errorf("cannot load cookie banner: %w", err)
				}

				banner.Version++
				banner.UpdatedAt = now

				if err := banner.Update(ctx, conn, scope); err != nil {
					return fmt.Errorf("cannot update cookie banner version: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot update cookie category: %w", err)
	}

	return category, nil
}

func (s *Service) DeleteCookieCategory(
	ctx context.Context,
	categoryID gid.GID,
) error {
	scope := coredata.NewScopeFromObjectID(categoryID)

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			category := &coredata.CookieCategory{}
			if err := category.LoadByID(ctx, conn, scope, categoryID); err != nil {
				return fmt.Errorf("cannot load cookie category: %w", err)
			}

			if category.Required {
				return ErrCannotDeleteRequiredCategory
			}

			if err := category.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot delete cookie category: %w", err)
			}

			banner := &coredata.CookieBanner{}
			if err := banner.LoadByID(ctx, conn, scope, category.CookieBannerID); err != nil {
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			banner.Version++
			banner.UpdatedAt = time.Now()

			if err := banner.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update cookie banner version: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot delete cookie category: %w", err)
	}

	return nil
}

func (r *AddCookiesFromProviderRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieCategoryID, "cookie_category_id", validator.Required(), validator.GID(coredata.CookieCategoryEntityType))
	v.Check(r.ProviderKey, "provider_key", validator.Required())

	return v.Error()
}

func (s *Service) AddCookiesFromProvider(
	ctx context.Context,
	req AddCookiesFromProviderRequest,
) (*coredata.CookieCategory, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	provider, ok := cookieprovider.ByKey(req.ProviderKey)
	if !ok {
		return nil, fmt.Errorf("invalid request: %w", validator.ValidationErrors{
			{Field: "provider_key", Code: validator.ErrorCodeInvalidFormat, Message: "unknown cookie provider"},
		})
	}

	scope := coredata.NewScopeFromObjectID(req.CookieCategoryID)
	category := &coredata.CookieCategory{}

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := category.LoadByID(ctx, conn, scope, req.CookieCategoryID); err != nil {
				return fmt.Errorf("cannot load cookie category: %w", err)
			}

			existing := make(map[string]bool, len(category.Cookies))
			for _, c := range category.Cookies {
				existing[c.Name] = true
			}

			for _, item := range provider.CookieItems() {
				if !existing[item.Name] {
					category.Cookies = append(category.Cookies, item)
				}
			}

			now := time.Now()
			category.UpdatedAt = now

			if err := category.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update cookie category: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot add cookies from provider: %w", err)
	}

	return category, nil
}
