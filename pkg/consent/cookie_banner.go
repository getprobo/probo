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
	"encoding/json"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	CreateCookieBannerRequest struct {
		OrganizationID gid.GID
		Name           string
		Domain         string
	}

	UpdateCookieBannerRequest struct {
		ID                   gid.GID
		Name                 *string
		Domain               *string
		State                *coredata.CookieBannerState
		Title                *string
		Description          *string
		AcceptAllLabel       *string
		RejectAllLabel       *string
		SavePreferencesLabel *string
		PrivacyPolicyURL     *string
		ConsentExpiryDays    *int
		ConsentMode          *coredata.ConsentMode
		Theme                json.RawMessage
	}
)

func (r *CreateCookieBannerRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(NameMaxLength))

	return v.Error()
}

func (r *UpdateCookieBannerRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.CookieBannerEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(NameMaxLength))
	v.Check(r.Title, "title", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(r.AcceptAllLabel, "accept_all_label", validator.SafeTextNoNewLine(NameMaxLength))
	v.Check(r.RejectAllLabel, "reject_all_label", validator.SafeTextNoNewLine(NameMaxLength))
	v.Check(r.SavePreferencesLabel, "save_preferences_label", validator.SafeTextNoNewLine(NameMaxLength))
	v.Check(r.State, "state", validator.OneOfSlice(coredata.CookieBannerStates()))
	v.Check(r.ConsentExpiryDays, "consent_expiry_days", validator.Min(ConsentExpiryDaysMin), validator.Max(ConsentExpiryDaysMax))
	v.Check(r.ConsentMode, "consent_mode", validator.OneOfSlice(coredata.ConsentModes()))

	return v.Error()
}

func (s *Service) GetCookieBanner(
	ctx context.Context,
	bannerID gid.GID,
) (*coredata.CookieBanner, error) {
	scope := coredata.NewScopeFromObjectID(bannerID)
	banner := &coredata.CookieBanner{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return banner.LoadByID(ctx, conn, scope, bannerID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get cookie banner: %w", err)
	}

	return banner, nil
}

func (s *Service) ListCookieBannersForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.CookieBannerOrderField],
) (*page.Page[*coredata.CookieBanner, coredata.CookieBannerOrderField], error) {
	scope := coredata.NewScopeFromObjectID(organizationID)
	var banners coredata.CookieBanners

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := banners.LoadByOrganizationID(
				ctx,
				conn,
				scope,
				organizationID,
				cursor,
			); err != nil {
				return fmt.Errorf("cannot load cookie banners: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list cookie banners: %w", err)
	}

	return page.NewPage(banners, cursor), nil
}

func (s *Service) CountCookieBannersForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	scope := coredata.NewScopeFromObjectID(organizationID)
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			banners := coredata.CookieBanners{}
			count, err = banners.CountByOrganizationID(ctx, conn, scope, organizationID)
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

func (s *Service) CreateCookieBanner(
	ctx context.Context,
	req CreateCookieBannerRequest,
) (*coredata.CookieBanner, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var (
		scope    = coredata.NewScopeFromObjectID(req.OrganizationID)
		now      = time.Now()
		bannerID = gid.New(req.OrganizationID.TenantID(), coredata.CookieBannerEntityType)
		tenantID = req.OrganizationID.TenantID()
	)

	banner := &coredata.CookieBanner{
		ID:                   bannerID,
		OrganizationID:       req.OrganizationID,
		Name:                 req.Name,
		Domain:               req.Domain,
		State:                coredata.CookieBannerStateDraft,
		Title:                "We value your privacy",
		Description:          "We use cookies to enhance your browsing experience, serve personalized content, and analyze our traffic. By clicking \"Accept All\", you consent to our use of cookies.",
		AcceptAllLabel:       "Accept all",
		RejectAllLabel:       "Reject all",
		SavePreferencesLabel: "Save preferences",
		ConsentExpiryDays:    365,
		ConsentMode:          coredata.ConsentModeOptIn,
		Version:              1,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	defaultCategories := []*coredata.CookieCategory{
		{
			ID:             gid.New(tenantID, coredata.CookieCategoryEntityType),
			CookieBannerID: bannerID,
			Name:           "Necessary",
			Description:    "Essential cookies required for the website to function properly. These cannot be disabled.",
			Required:       true,
			Rank:           0,
			Cookies:        make(coredata.CookieItems, 0),
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             gid.New(tenantID, coredata.CookieCategoryEntityType),
			CookieBannerID: bannerID,
			Name:           "Analytics",
			Description:    "Cookies that help us understand how visitors interact with our website.",
			Required:       false,
			Rank:           1,
			Cookies:        make(coredata.CookieItems, 0),
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             gid.New(tenantID, coredata.CookieCategoryEntityType),
			CookieBannerID: bannerID,
			Name:           "Marketing",
			Description:    "Cookies used to deliver personalized advertisements.",
			Required:       false,
			Rank:           2,
			Cookies:        make(coredata.CookieItems, 0),
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             gid.New(tenantID, coredata.CookieCategoryEntityType),
			CookieBannerID: bannerID,
			Name:           "Preferences",
			Description:    "Cookies that remember your settings and preferences.",
			Required:       false,
			Rank:           3,
			Cookies:        make(coredata.CookieItems, 0),
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := banner.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert cookie banner: %w", err)
			}

			for _, category := range defaultCategories {
				if err := category.Insert(ctx, conn, scope); err != nil {
					return fmt.Errorf("cannot insert default cookie category %q: %w", category.Name, err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create cookie banner: %w", err)
	}

	return banner, nil
}

func (s *Service) UpdateCookieBanner(
	ctx context.Context,
	req UpdateCookieBannerRequest,
) (*coredata.CookieBanner, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	scope := coredata.NewScopeFromObjectID(req.ID)
	banner := &coredata.CookieBanner{ID: req.ID}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := banner.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			bumpVersion := false
			now := time.Now()
			banner.UpdatedAt = now

			if req.Name != nil {
				banner.Name = *req.Name
			}
			if req.Domain != nil {
				banner.Domain = *req.Domain
			}
			if req.State != nil {
				banner.State = *req.State
			}
			if req.Title != nil {
				if *req.Title != banner.Title {
					bumpVersion = true
				}
				banner.Title = *req.Title
			}
			if req.Description != nil {
				if *req.Description != banner.Description {
					bumpVersion = true
				}
				banner.Description = *req.Description
			}
			if req.AcceptAllLabel != nil {
				banner.AcceptAllLabel = *req.AcceptAllLabel
			}
			if req.RejectAllLabel != nil {
				banner.RejectAllLabel = *req.RejectAllLabel
			}
			if req.SavePreferencesLabel != nil {
				banner.SavePreferencesLabel = *req.SavePreferencesLabel
			}
			if req.PrivacyPolicyURL != nil {
				if *req.PrivacyPolicyURL != banner.PrivacyPolicyURL {
					bumpVersion = true
				}
				banner.PrivacyPolicyURL = *req.PrivacyPolicyURL
			}
			if req.ConsentExpiryDays != nil {
				banner.ConsentExpiryDays = *req.ConsentExpiryDays
			}
			if req.ConsentMode != nil {
				if *req.ConsentMode != banner.ConsentMode {
					bumpVersion = true
				}
				banner.ConsentMode = *req.ConsentMode
			}
			if req.Theme != nil {
				banner.Theme = req.Theme
			}

			if bumpVersion {
				banner.Version++
			}

			if err := banner.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update cookie banner: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot update cookie banner: %w", err)
	}

	return banner, nil
}

func (s *Service) DeleteCookieBanner(
	ctx context.Context,
	bannerID gid.GID,
) error {
	scope := coredata.NewScopeFromObjectID(bannerID)
	banner := &coredata.CookieBanner{ID: bannerID}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return banner.Delete(ctx, conn, scope)
		},
	)
	if err != nil {
		return fmt.Errorf("cannot delete cookie banner: %w", err)
	}

	return nil
}

func (s *Service) PublishCookieBanner(
	ctx context.Context,
	bannerID gid.GID,
) (*coredata.CookieBanner, error) {
	return s.UpdateCookieBanner(
		ctx,
		UpdateCookieBannerRequest{
			ID:    bannerID,
			State: new(coredata.CookieBannerStatePublished),
		},
	)
}

func (s *Service) DisableCookieBanner(
	ctx context.Context,
	bannerID gid.GID,
) (*coredata.CookieBanner, error) {
	return s.UpdateCookieBanner(
		ctx,
		UpdateCookieBannerRequest{
			ID:    bannerID,
			State: new(coredata.CookieBannerStateDisabled),
		},
	)
}
