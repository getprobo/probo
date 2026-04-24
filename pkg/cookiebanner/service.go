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
	"maps"
	"net/url"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type Service struct {
	pg           *pg.Client
	showBranding bool
}

func NewService(pgClient *pg.Client, showBranding bool) *Service {
	return &Service{pg: pgClient, showBranding: showBranding}
}

type (
	CreateCookieBannerRequest struct {
		OrganizationID    gid.GID
		Name              string
		Origin            string
		PrivacyPolicyURL  *string
		CookiePolicyURL   string
		ConsentExpiryDays int
		ConsentMode       coredata.CookieConsentMode
	}

	CreateCookieCategoryRequest struct {
		CookieBannerID gid.GID
		Name           string
		Slug           string
		Description    string
		Rank           int
	}

	UpdateCookieBannerRequest struct {
		CookieBannerID    gid.GID
		Name              *string
		PrivacyPolicyURL  *string
		CookiePolicyURL   *string
		ConsentExpiryDays *int
		ConsentMode       *coredata.CookieConsentMode
		DefaultLanguage   *string
	}

	UpdateCookieCategoryRequest struct {
		CookieCategoryID gid.GID
		Name             *string
		Slug             *string
		Description      *string
		GCMConsentTypes  *[]string
		PostHogConsent   *bool
	}

	CreateCookieRequest struct {
		CookieCategoryID gid.GID
		Name             string
		Duration         string
		Description      string
	}

	UpdateCookieRequest struct {
		CookieID    gid.GID
		Name        *string
		Duration    *string
		Description *string
	}

	ReorderCookieCategoryRequest struct {
		CookieCategoryID gid.GID
		Rank             int
	}

	MoveCookieToCategoryRequest struct {
		CookieID               gid.GID
		TargetCookieCategoryID gid.GID
	}

	CreateCookieConsentRecordRequest struct {
		CookieBannerID gid.GID
		Version        int
		VisitorID      string
		IPAddress      *string
		UserAgent      *string
		ConsentData    json.RawMessage
		Action         coredata.CookieConsentAction
		SdkVersion     string
	}

	RecordConsentRequest struct {
		Version     int
		VisitorID   string
		IPAddress   *string
		UserAgent   *string
		ConsentData json.RawMessage
		Action      coredata.CookieConsentAction
		SdkVersion  string
	}

	DetectedCookie struct {
		Name     string
		Duration string
	}

	ReportDetectedCookiesRequest struct {
		Cookies []DetectedCookie
	}

	BannerConfig struct {
		BannerID          gid.GID                                        `json:"banner_id"`
		Version           int                                            `json:"version"`
		Language          string                                         `json:"language"`
		DefaultLanguage   string                                         `json:"default_language"`
		PrivacyPolicyURL  string                                         `json:"privacy_policy_url,omitempty"`
		CookiePolicyURL   string                                         `json:"cookie_policy_url"`
		ConsentExpiryDays int                                            `json:"consent_expiry_days"`
		ConsentMode       string                                         `json:"consent_mode"`
		ShowBranding      bool                                           `json:"show_branding"`
		Categories        []coredata.CookieBannerVersionSnapshotCategory `json:"categories"`
		Texts             map[string]string                              `json:"texts"`
	}

	UpsertCookieBannerTranslationRequest struct {
		CookieBannerID gid.GID
		Language       string
		Translations   json.RawMessage
	}

	VisitorConsent struct {
		VisitorID   string                       `json:"visitor_id"`
		Version     int                          `json:"version"`
		Action      coredata.CookieConsentAction `json:"action"`
		ConsentData json.RawMessage              `json:"consent_data"`
		CreatedAt   time.Time                    `json:"created_at"`
	}
)

func (r *CreateCookieBannerRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(255))
	v.Check(r.Origin, "origin", validator.Required(), validator.Origin())
	v.Check(r.PrivacyPolicyURL, "privacy_policy_url", validator.URL())
	v.Check(r.CookiePolicyURL, "cookie_policy_url", validator.Required(), validator.URL())
	v.Check(r.ConsentExpiryDays, "consent_expiry_days", validator.Required(), validator.Min(1))
	v.Check(r.ConsentMode, "consent_mode", validator.Required(), validator.OneOfSlice(coredata.CookieConsentModes()))

	return v.Error()
}

func (r *UpdateCookieBannerRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieBannerID, "cookie_banner_id", validator.Required(), validator.GID(coredata.CookieBannerEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(255))
	v.Check(r.PrivacyPolicyURL, "privacy_policy_url", validator.URL())
	v.Check(r.CookiePolicyURL, "cookie_policy_url", validator.URL())
	v.Check(r.ConsentExpiryDays, "consent_expiry_days", validator.Min(1))
	v.Check(r.ConsentMode, "consent_mode", validator.OneOfSlice(coredata.CookieConsentModes()))
	v.Check(r.DefaultLanguage, "default_language", validator.OneOfSlice(SupportedLanguages))

	return v.Error()
}

func (r *CreateCookieCategoryRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieBannerID, "cookie_banner_id", validator.Required(), validator.GID(coredata.CookieBannerEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(255))
	v.Check(r.Slug, "slug", validator.Required(), validator.Slug(100))
	v.Check(r.Description, "description", validator.Required(), validator.SafeText(1000))
	v.Check(r.Rank, "rank", validator.Min(0))

	return v.Error()
}

func (r *UpdateCookieCategoryRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieCategoryID, "cookie_category_id", validator.Required(), validator.GID(coredata.CookieCategoryEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(255))
	v.Check(r.Slug, "slug", validator.Slug(100))
	v.Check(r.Description, "description", validator.SafeText(1000))

	return v.Error()
}

func (r *CreateCookieRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieCategoryID, "cookie_category_id", validator.Required(), validator.GID(coredata.CookieCategoryEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(255))
	v.Check(r.Duration, "duration", validator.Required(), validator.SafeTextNoNewLine(255))
	v.Check(r.Description, "description", validator.SafeText(1000))

	return v.Error()
}

func (r *UpdateCookieRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieID, "cookie_id", validator.Required(), validator.GID(coredata.CookieEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(255))
	v.Check(r.Duration, "duration", validator.SafeTextNoNewLine(255))
	v.Check(r.Description, "description", validator.SafeText(1000))

	return v.Error()
}

func (r *ReorderCookieCategoryRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieCategoryID, "cookie_category_id", validator.Required(), validator.GID(coredata.CookieCategoryEntityType))
	v.Check(r.Rank, "rank", validator.Min(0))

	return v.Error()
}

func (r *MoveCookieToCategoryRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieID, "cookie_id", validator.Required(), validator.GID(coredata.CookieEntityType))
	v.Check(r.TargetCookieCategoryID, "target_cookie_category_id", validator.Required(), validator.GID(coredata.CookieCategoryEntityType))

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

func (r *RecordConsentRequest) Validate() error {
	v := validator.New()

	v.Check(r.Version, "version", validator.Required(), validator.Min(1))
	v.Check(r.VisitorID, "visitor_id", validator.Required(), validator.NotEmpty())
	v.Check(r.Action, "action", validator.Required(), validator.OneOfSlice(coredata.CookieConsentActions()))

	return v.Error()
}

func (r *UpsertCookieBannerTranslationRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieBannerID, "cookie_banner_id", validator.Required(), validator.GID(coredata.CookieBannerEntityType))
	v.Check(r.Language, "language", validator.Required(), validator.SafeTextNoNewLine(10))

	var flat map[string]json.RawMessage
	if err := json.Unmarshal(r.Translations, &flat); err != nil {
		v.Check("", "translations", validator.Required())
		return v.Error()
	}

	for key, raw := range flat {
		if key == "categories" {
			var cats map[string]json.RawMessage
			if json.Unmarshal(raw, &cats) == nil {
				for catID, catRaw := range cats {
					var catFields map[string]json.RawMessage
					if json.Unmarshal(catRaw, &catFields) == nil {
						for field, fieldRaw := range catFields {
							var s string
							if json.Unmarshal(fieldRaw, &s) == nil {
								v.Check(s, fmt.Sprintf("translations.categories.%s.%s", catID, field), validator.NoHTML(), validator.MaxLen(2000))
							}
						}
					}
				}
			}
			continue
		}

		var s string
		if json.Unmarshal(raw, &s) != nil {
			continue
		}
		v.Check(s, "translations."+key, validator.NoHTML(), validator.MaxLen(2000))
	}

	return v.Error()
}

func CanonicalizeOrigin(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	host := u.Hostname()
	host = strings.TrimPrefix(host, "www.")

	port := u.Port()
	if port != "" {
		return u.Scheme + "://" + host + ":" + port
	}

	return u.Scheme + "://" + host
}

func buildSnapshot(
	banner *coredata.CookieBanner,
	categories coredata.CookieCategories,
	allCookies coredata.Cookies,
	translations coredata.CookieBannerTranslations,
) coredata.CookieBannerVersionSnapshot {
	cookiesByCategory := make(map[gid.GID]coredata.CookieItems)
	for _, c := range allCookies {
		cookiesByCategory[c.CookieCategoryID] = append(
			cookiesByCategory[c.CookieCategoryID],
			coredata.CookieItem{
				Name:        c.Name,
				Duration:    c.Duration,
				Description: c.Description,
			},
		)
	}

	snapshotCategories := make([]coredata.CookieBannerVersionSnapshotCategory, len(categories))
	for i, c := range categories {
		cookies := cookiesByCategory[c.ID]
		if cookies == nil {
			cookies = coredata.CookieItems{}
		}
		gcmConsentTypes := c.GCMConsentTypes
		if gcmConsentTypes == nil {
			gcmConsentTypes = []string{}
		}
		snapshotCategories[i] = coredata.CookieBannerVersionSnapshotCategory{
			Name:            c.Name,
			Slug:            c.Slug,
			Description:     c.Description,
			Kind:            c.Kind,
			Cookies:         cookies,
			GCMConsentTypes: gcmConsentTypes,
			PostHogConsent:  c.PostHogConsent,
		}
	}

	snapshotTranslations := buildSnapshotTranslations(translations, categories)

	return coredata.CookieBannerVersionSnapshot{
		PrivacyPolicyURL:  banner.PrivacyPolicyURL,
		CookiePolicyURL:   banner.CookiePolicyURL,
		ConsentExpiryDays: banner.ConsentExpiryDays,
		ConsentMode:       string(banner.ConsentMode),
		DefaultLanguage:   banner.DefaultLanguage,
		Categories:        snapshotCategories,
		Translations:      snapshotTranslations,
	}
}

func buildSnapshotTranslations(
	translations coredata.CookieBannerTranslations,
	categories coredata.CookieCategories,
) map[string]coredata.CookieBannerVersionSnapshotTranslation {
	if len(translations) == 0 {
		return nil
	}

	result := make(map[string]coredata.CookieBannerVersionSnapshotTranslation, len(translations))

	for _, t := range translations {
		var raw struct {
			Categories map[string]struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"categories"`
		}
		_ = json.Unmarshal(t.Translations, &raw)

		ui := make(map[string]string)
		var flat map[string]json.RawMessage
		_ = json.Unmarshal(t.Translations, &flat)
		for k, v := range flat {
			if k == "categories" || k == "cookies" {
				continue
			}
			var s string
			if json.Unmarshal(v, &s) == nil {
				ui[k] = s
			}
		}

		catTranslations := make([]coredata.CookieBannerVersionSnapshotCategoryTranslation, len(categories))
		for i, c := range categories {
			if raw.Categories != nil {
				if ct, ok := raw.Categories[c.ID.String()]; ok {
					catTranslations[i] = coredata.CookieBannerVersionSnapshotCategoryTranslation{
						Name:        ct.Name,
						Description: ct.Description,
					}
					continue
				}
			}
			catTranslations[i] = coredata.CookieBannerVersionSnapshotCategoryTranslation{
				Name:        c.Name,
				Description: c.Description,
			}
		}

		result[t.Language] = coredata.CookieBannerVersionSnapshotTranslation{
			UI:         ui,
			Categories: catTranslations,
		}
	}

	return result
}

func (s *Service) ensureDraftVersion(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	banner *coredata.CookieBanner,
	categories coredata.CookieCategories,
	allCookies coredata.Cookies,
	translations coredata.CookieBannerTranslations,
) (*coredata.CookieBannerVersion, error) {
	snapshot := buildSnapshot(banner, categories, allCookies, translations)

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

func (s *Service) ensureDraftVersionForBanner(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	bannerID gid.GID,
) (*coredata.CookieBannerVersion, error) {
	var banner coredata.CookieBanner
	if err := banner.LoadByID(ctx, tx, scope, bannerID); err != nil {
		return nil, fmt.Errorf("cannot load cookie banner: %w", err)
	}

	var categories coredata.CookieCategories
	if err := categories.LoadAllByCookieBannerID(ctx, tx, scope, bannerID); err != nil {
		return nil, fmt.Errorf("cannot load cookie categories: %w", err)
	}

	var allCookies coredata.Cookies
	if err := allCookies.LoadAllByCookieBannerID(ctx, tx, scope, bannerID); err != nil {
		return nil, fmt.Errorf("cannot load cookies: %w", err)
	}

	var translations coredata.CookieBannerTranslations
	if err := translations.LoadAllByCookieBannerID(ctx, tx, scope, bannerID); err != nil {
		return nil, fmt.Errorf("cannot load cookie banner translations: %w", err)
	}

	return s.ensureDraftVersion(ctx, tx, scope, &banner, categories, allCookies, translations)
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
				Origin:            CanonicalizeOrigin(req.Origin),
				State:             coredata.CookieBannerStateActive,
				PrivacyPolicyURL:  req.PrivacyPolicyURL,
				CookiePolicyURL:   req.CookiePolicyURL,
				ConsentExpiryDays: req.ConsentExpiryDays,
				ConsentMode:       req.ConsentMode,
				ShowBranding:      s.showBranding,
				DefaultLanguage:   "en",
				CreatedAt:         now,
				UpdatedAt:         now,
			}

			if err := banner.Insert(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrOriginAlreadyInUse
				}
				return fmt.Errorf("cannot insert cookie banner: %w", err)
			}

			slugToGID := make(map[string]gid.GID, len(defaultCategories))
			for _, dc := range defaultCategories {
				gcmConsentTypes := dc.GCMConsentTypes
				if gcmConsentTypes == nil {
					gcmConsentTypes = []string{}
				}
				category := &coredata.CookieCategory{
					ID:              gid.New(scope.GetTenantID(), coredata.CookieCategoryEntityType),
					OrganizationID:  banner.OrganizationID,
					CookieBannerID:  banner.ID,
					Name:            dc.Name,
					Slug:            dc.Slug,
					Description:     dc.Description,
					Kind:            dc.Kind,
					Rank:            dc.Rank,
					GCMConsentTypes: gcmConsentTypes,
					PostHogConsent:  dc.PostHogConsent,
					CreatedAt:       now,
					UpdatedAt:       now,
				}

				if err := category.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert default cookie category %q: %w", dc.Name, err)
				}

				slugToGID[dc.Slug] = category.ID

				if dc.Kind == coredata.CookieCategoryKindNecessary {
					consentCookie := &coredata.Cookie{
						ID:               gid.New(scope.GetTenantID(), coredata.CookieEntityType),
						OrganizationID:   banner.OrganizationID,
						CookieBannerID:   banner.ID,
						CookieCategoryID: category.ID,
						Name:             "probo_consent",
						Duration:         fmt.Sprintf("%d days", req.ConsentExpiryDays),
						Description:      "Stores your cookie consent preferences for this website.",
						CreatedAt:        now,
						UpdatedAt:        now,
					}
					if err := consentCookie.Insert(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot insert probo_consent cookie: %w", err)
					}
				}
			}

			for lang, uiStrings := range defaultUIStringsByLanguage {
				blob := make(map[string]any, len(uiStrings)+1)
				for k, v := range uiStrings {
					blob[k] = v
				}

				if catDefaults, ok := defaultCategoryTranslationsByLanguage[lang]; ok {
					catMap := make(map[string]map[string]string, len(catDefaults))
					for slug, ct := range catDefaults {
						if id, exists := slugToGID[slug]; exists {
							catMap[id.String()] = map[string]string{
								"name":        ct.Name,
								"description": ct.Description,
							}
						}
					}
					if len(catMap) > 0 {
						blob["categories"] = catMap
					}
				}

				translationsJSON, err := json.Marshal(blob)
				if err != nil {
					return fmt.Errorf("cannot marshal default translations for %s: %w", lang, err)
				}

				translation := &coredata.CookieBannerTranslation{
					ID:             gid.New(scope.GetTenantID(), coredata.CookieBannerTranslationEntityType),
					OrganizationID: banner.OrganizationID,
					CookieBannerID: banner.ID,
					Language:       lang,
					Translations:   translationsJSON,
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				if err := translation.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert default translation for %s: %w", lang, err)
				}
			}

			if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, banner.ID); err != nil {
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

func (s *Service) GetActiveCookieBanner(
	ctx context.Context,
	bannerID gid.GID,
) (*coredata.CookieBanner, error) {
	var banner coredata.CookieBanner

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := banner.LoadActiveByID(ctx, conn, bannerID); err != nil {
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

			consentChanged := req.PrivacyPolicyURL != nil ||
				req.CookiePolicyURL != nil ||
				req.ConsentExpiryDays != nil ||
				req.ConsentMode != nil ||
				req.DefaultLanguage != nil

			if req.Name != nil {
				banner.Name = *req.Name
			}
			if req.PrivacyPolicyURL != nil {
				banner.PrivacyPolicyURL = req.PrivacyPolicyURL
			}
			if req.CookiePolicyURL != nil {
				banner.CookiePolicyURL = *req.CookiePolicyURL
			}
			if req.ConsentExpiryDays != nil {
				banner.ConsentExpiryDays = *req.ConsentExpiryDays
			}
			if req.ConsentMode != nil {
				banner.ConsentMode = *req.ConsentMode
			}
			if req.DefaultLanguage != nil {
				banner.DefaultLanguage = *req.DefaultLanguage
			}

			banner.UpdatedAt = time.Now()

			if err := banner.Update(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrOriginAlreadyInUse
				}
				return fmt.Errorf("cannot update cookie banner: %w", err)
			}

			if consentChanged {
				if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, banner.ID); err != nil {
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

			category = &coredata.CookieCategory{
				ID:              gid.New(scope.GetTenantID(), coredata.CookieCategoryEntityType),
				OrganizationID:  banner.OrganizationID,
				CookieBannerID:  req.CookieBannerID,
				Name:            req.Name,
				Slug:            req.Slug,
				Description:     req.Description,
				Kind:            coredata.CookieCategoryKindNormal,
				Rank:            req.Rank,
				GCMConsentTypes: []string{},
				CreatedAt:       now,
				UpdatedAt:       now,
			}

			if err := category.Insert(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrCategorySlugAlreadyExists
				}
				return fmt.Errorf("cannot insert cookie category: %w", err)
			}

			if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, req.CookieBannerID); err != nil {
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

func (s *Service) CreateCookie(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateCookieRequest,
) (*coredata.Cookie, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var cookie *coredata.Cookie

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var category coredata.CookieCategory
			if err := category.LoadByID(ctx, tx, scope, req.CookieCategoryID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrCategoryNotFound
				}
				return fmt.Errorf("cannot load cookie category: %w", err)
			}

			now := time.Now()

			cookie = &coredata.Cookie{
				ID:               gid.New(scope.GetTenantID(), coredata.CookieEntityType),
				OrganizationID:   category.OrganizationID,
				CookieBannerID:   category.CookieBannerID,
				CookieCategoryID: category.ID,
				Name:             req.Name,
				Duration:         req.Duration,
				Description:      req.Description,
				CreatedAt:        now,
				UpdatedAt:        now,
			}

			if err := cookie.Insert(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrCookieNameAlreadyExists
				}
				return fmt.Errorf("cannot insert cookie: %w", err)
			}

			if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, category.CookieBannerID); err != nil {
				return fmt.Errorf("cannot ensure draft version: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return cookie, nil
}

func (s *Service) GetCookie(
	ctx context.Context,
	scope coredata.Scoper,
	cookieID gid.GID,
) (*coredata.Cookie, error) {
	var cookie coredata.Cookie

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := cookie.LoadByID(ctx, conn, scope, cookieID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrCookieNotFound
				}
				return fmt.Errorf("cannot load cookie: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &cookie, nil
}

func (s *Service) UpdateCookie(
	ctx context.Context,
	scope coredata.Scoper,
	req UpdateCookieRequest,
) (*coredata.Cookie, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var cookie coredata.Cookie

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := cookie.LoadByID(ctx, tx, scope, req.CookieID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrCookieNotFound
				}
				return fmt.Errorf("cannot load cookie: %w", err)
			}

			if req.Name != nil {
				cookie.Name = *req.Name
			}
			if req.Duration != nil {
				cookie.Duration = *req.Duration
			}
			if req.Description != nil {
				cookie.Description = *req.Description
			}

			cookie.UpdatedAt = time.Now()

			if err := cookie.Update(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrCookieNameAlreadyExists
				}
				return fmt.Errorf("cannot update cookie: %w", err)
			}

			if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, cookie.CookieBannerID); err != nil {
				return fmt.Errorf("cannot ensure draft version: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &cookie, nil
}

func (s *Service) DeleteCookie(
	ctx context.Context,
	scope coredata.Scoper,
	cookieID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var cookie coredata.Cookie
			if err := cookie.LoadByID(ctx, tx, scope, cookieID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrCookieNotFound
				}
				return fmt.Errorf("cannot load cookie: %w", err)
			}

			if err := cookie.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete cookie: %w", err)
			}

			if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, cookie.CookieBannerID); err != nil {
				return fmt.Errorf("cannot ensure draft version: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) ListCookiesForCategory(
	ctx context.Context,
	scope coredata.Scoper,
	categoryID gid.GID,
	cursor *page.Cursor[coredata.CookieOrderField],
) (coredata.Cookies, error) {
	var cookies coredata.Cookies

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := cookies.LoadByCookieCategoryID(ctx, conn, scope, categoryID, cursor); err != nil {
				return fmt.Errorf("cannot list cookies: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return cookies, nil
}

func (s *Service) CountCookiesForCategory(
	ctx context.Context,
	scope coredata.Scoper,
	categoryID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var cookies coredata.Cookies
			var err error

			count, err = cookies.CountByCookieCategoryID(ctx, conn, scope, categoryID)
			if err != nil {
				return fmt.Errorf("cannot count cookies: %w", err)
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
			if req.Slug != nil {
				category.Slug = *req.Slug
			}
			if req.Description != nil {
				category.Description = *req.Description
			}
			if req.GCMConsentTypes != nil {
				category.GCMConsentTypes = *req.GCMConsentTypes
			}
			if req.PostHogConsent != nil {
				if *req.PostHogConsent && category.Kind != coredata.CookieCategoryKindNormal {
					return ErrPostHogConsentKindInvalid
				}
				if *req.PostHogConsent {
					var categories coredata.CookieCategories
					if err := categories.ClearPostHogConsentByBannerID(ctx, tx, scope, category.CookieBannerID); err != nil {
						return fmt.Errorf("cannot clear posthog consent: %w", err)
					}
				}
				category.PostHogConsent = *req.PostHogConsent
			}

			category.UpdatedAt = time.Now()

			if err := category.Update(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrCategorySlugAlreadyExists
				}
				return fmt.Errorf("cannot update cookie category: %w", err)
			}

			if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, category.CookieBannerID); err != nil {
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

type MoveCookieToCategoryResult struct {
	Cookie *coredata.Cookie
	Banner *coredata.CookieBanner
}

func (s *Service) MoveCookieToCategory(
	ctx context.Context,
	scope coredata.Scoper,
	req MoveCookieToCategoryRequest,
) (*MoveCookieToCategoryResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var result MoveCookieToCategoryResult

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var cookie coredata.Cookie
			if err := cookie.LoadByID(ctx, tx, scope, req.CookieID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrCookieNotFound
				}
				return fmt.Errorf("cannot load cookie: %w", err)
			}

			var target coredata.CookieCategory
			if err := target.LoadByID(ctx, tx, scope, req.TargetCookieCategoryID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrCategoryNotFound
				}
				return fmt.Errorf("cannot load target cookie category: %w", err)
			}

			if cookie.CookieCategoryID == target.ID {
				return ErrSameCategoryMove
			}

			if cookie.CookieBannerID != target.CookieBannerID {
				return ErrCategoriesBannerMismatch
			}

			cookie.CookieCategoryID = target.ID
			cookie.UpdatedAt = time.Now()

			if err := cookie.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update cookie: %w", err)
			}

			var banner coredata.CookieBanner
			if err := banner.LoadByID(ctx, tx, scope, cookie.CookieBannerID); err != nil {
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, cookie.CookieBannerID); err != nil {
				return fmt.Errorf("cannot ensure draft version: %w", err)
			}

			result.Cookie = &cookie
			result.Banner = &banner

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *Service) ReorderCookieCategory(
	ctx context.Context,
	scope coredata.Scoper,
	req ReorderCookieCategoryRequest,
) (*coredata.CookieBanner, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var banner coredata.CookieBanner

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var category coredata.CookieCategory
			if err := category.LoadByID(ctx, tx, scope, req.CookieCategoryID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrCategoryNotFound
				}
				return fmt.Errorf("cannot load cookie category: %w", err)
			}

			category.Rank = req.Rank
			category.UpdatedAt = time.Now()

			if err := category.UpdateRank(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot reorder cookie category: %w", err)
			}

			if err := banner.LoadByID(ctx, tx, scope, category.CookieBannerID); err != nil {
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, category.CookieBannerID); err != nil {
				return fmt.Errorf("cannot ensure draft version: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &banner, nil
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

			if category.Kind != coredata.CookieCategoryKindNormal {
				return ErrCannotDeleteSystemCategory
			}

			bannerID := category.CookieBannerID

			var uncategorised coredata.CookieCategory
			if err := uncategorised.LoadUncategorisedByCookieBannerID(ctx, tx, scope, bannerID); err != nil {
				return fmt.Errorf("cannot load uncategorised cookie category: %w", err)
			}

			var cookies coredata.Cookies
			if err := cookies.MoveToCategoryByCookieCategoryID(ctx, tx, scope, category.ID, uncategorised.ID); err != nil {
				return fmt.Errorf("cannot move cookies to uncategorised: %w", err)
			}

			if err := category.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete cookie category: %w", err)
			}

			if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, bannerID); err != nil {
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
				SdkVersion:            req.SdkVersion,
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

func (s *Service) GetActiveBannerConfig(
	ctx context.Context,
	bannerID gid.GID,
	lang string,
) (*BannerConfig, error) {
	var config *BannerConfig

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var banner coredata.CookieBanner
			if err := banner.LoadActiveByID(ctx, conn, bannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot load active cookie banner: %w", err)
			}

			scope := coredata.NewScopeFromObjectID(banner.ID)

			var version coredata.CookieBannerVersion
			if err := version.LoadLatestPublishedByCookieBannerID(ctx, conn, scope, banner.ID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrNoPublishedVersion
				}
				return fmt.Errorf("cannot load latest published version: %w", err)
			}

			snapshot, err := version.GetSnapshot()
			if err != nil {
				return fmt.Errorf("cannot get version snapshot: %w", err)
			}

			config = buildBannerConfig(&banner, &version, &snapshot, lang)

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func buildBannerConfig(
	banner *coredata.CookieBanner,
	version *coredata.CookieBannerVersion,
	snapshot *coredata.CookieBannerVersionSnapshot,
	lang string,
) *BannerConfig {
	defaultLang := snapshot.DefaultLanguage
	if defaultLang == "" {
		defaultLang = "en"
	}

	resolvedLang := defaultLang
	if lang != "" {
		if _, ok := snapshot.Translations[lang]; ok {
			resolvedLang = lang
		}
	}

	categories := snapshot.Categories
	texts := make(map[string]string)

	if t, ok := snapshot.Translations[resolvedLang]; ok {
		maps.Copy(texts, t.UI)

		if len(t.Categories) == len(categories) {
			translated := make([]coredata.CookieBannerVersionSnapshotCategory, len(categories))
			copy(translated, categories)
			for i, ct := range t.Categories {
				if ct.Name != "" {
					translated[i].Name = ct.Name
				}
				if ct.Description != "" {
					translated[i].Description = ct.Description
				}
			}
			categories = translated
		}
	}

	var privacyPolicyURL string
	if snapshot.PrivacyPolicyURL != nil {
		privacyPolicyURL = *snapshot.PrivacyPolicyURL
	}

	return &BannerConfig{
		BannerID:          banner.ID,
		Version:           version.Version,
		Language:          resolvedLang,
		DefaultLanguage:   defaultLang,
		PrivacyPolicyURL:  privacyPolicyURL,
		CookiePolicyURL:   snapshot.CookiePolicyURL,
		ConsentExpiryDays: snapshot.ConsentExpiryDays,
		ConsentMode:       snapshot.ConsentMode,
		ShowBranding:      banner.ShowBranding,
		Categories:        categories,
		Texts:             texts,
	}
}

func (s *Service) SetShowBranding(
	ctx context.Context,
	bannerID gid.GID,
	show bool,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var banner coredata.CookieBanner
			banner.ID = bannerID
			if err := banner.UpdateShowBranding(ctx, tx, coredata.NewNoScope(), show); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot update show_branding: %w", err)
			}
			return nil
		},
	)
}

func (s *Service) UpsertCookieBannerTranslation(
	ctx context.Context,
	scope coredata.Scoper,
	req UpsertCookieBannerTranslationRequest,
) (*coredata.CookieBannerTranslation, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var result *coredata.CookieBannerTranslation

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

			var existing coredata.CookieBannerTranslation
			err := existing.LoadByCookieBannerIDAndLanguage(ctx, tx, scope, req.CookieBannerID, req.Language)

			if err == nil {
				existing.Translations = req.Translations
				existing.UpdatedAt = now
				if err := existing.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update cookie banner translation: %w", err)
				}
				result = &existing
			} else if errors.Is(err, coredata.ErrResourceNotFound) {
				t := &coredata.CookieBannerTranslation{
					ID:             gid.New(scope.GetTenantID(), coredata.CookieBannerTranslationEntityType),
					OrganizationID: banner.OrganizationID,
					CookieBannerID: req.CookieBannerID,
					Language:       req.Language,
					Translations:   req.Translations,
					CreatedAt:      now,
					UpdatedAt:      now,
				}
				if err := t.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert cookie banner translation: %w", err)
				}
				result = t
			} else {
				return fmt.Errorf("cannot load cookie banner translation: %w", err)
			}

			if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, req.CookieBannerID); err != nil {
				return fmt.Errorf("cannot ensure draft version: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) ListCookieBannerTranslations(
	ctx context.Context,
	scope coredata.Scoper,
	cookieBannerID gid.GID,
) (coredata.CookieBannerTranslations, error) {
	var translations coredata.CookieBannerTranslations

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return translations.LoadAllByCookieBannerID(ctx, conn, scope, cookieBannerID)
		},
	)
	if err != nil {
		return nil, err
	}

	return translations, nil
}

func (s *Service) GetVisitorConsent(
	ctx context.Context,
	bannerID gid.GID,
	visitorID string,
) (*VisitorConsent, error) {
	var consent *VisitorConsent

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var banner coredata.CookieBanner
			if err := banner.LoadActiveByID(ctx, conn, bannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot load active cookie banner: %w", err)
			}

			scope := coredata.NewScopeFromObjectID(banner.ID)

			var record coredata.CookieConsentRecord
			if err := record.LoadLatestByVisitorAndBannerID(ctx, conn, scope, banner.ID, visitorID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrConsentNotFound
				}
				return fmt.Errorf("cannot load consent record: %w", err)
			}

			var version coredata.CookieBannerVersion
			if err := version.LoadByID(ctx, conn, scope, record.CookieBannerVersionID); err != nil {
				return fmt.Errorf("cannot load cookie banner version: %w", err)
			}

			consent = &VisitorConsent{
				VisitorID:   record.VisitorID,
				Version:     version.Version,
				Action:      record.Action,
				ConsentData: record.ConsentData,
				CreatedAt:   record.CreatedAt,
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return consent, nil
}

func (s *Service) RecordConsent(
	ctx context.Context,
	bannerID gid.GID,
	req RecordConsentRequest,
) (*coredata.CookieConsentRecord, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if req.IPAddress != nil {
		anonymized := AnonymizeIP(*req.IPAddress)
		req.IPAddress = &anonymized
	}

	var record *coredata.CookieConsentRecord

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var banner coredata.CookieBanner
			if err := banner.LoadActiveByID(ctx, tx, bannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot load active cookie banner: %w", err)
			}

			scope := coredata.NewScopeFromObjectID(banner.ID)

			var publishedVersion coredata.CookieBannerVersion
			if err := publishedVersion.LoadByCookieBannerIDAndVersion(ctx, tx, scope, banner.ID, req.Version); err != nil {
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
				OrganizationID:        banner.OrganizationID,
				CookieBannerID:        banner.ID,
				CookieBannerVersionID: publishedVersion.ID,
				VisitorID:             req.VisitorID,
				IPAddress:             req.IPAddress,
				UserAgent:             req.UserAgent,
				ConsentData:           req.ConsentData,
				Action:                req.Action,
				SdkVersion:            req.SdkVersion,
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

func (s *Service) ReportDetectedCookies(
	ctx context.Context,
	bannerID gid.GID,
	req ReportDetectedCookiesRequest,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var banner coredata.CookieBanner
			if err := banner.LoadActiveByID(ctx, tx, bannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot load active cookie banner: %w", err)
			}

			scope := coredata.NewScopeFromObjectID(banner.ID)

			var uncategorised coredata.CookieCategory
			if err := uncategorised.LoadUncategorisedByCookieBannerID(ctx, tx, scope, banner.ID); err != nil {
				return fmt.Errorf("cannot load uncategorised category: %w", err)
			}

			inserted := 0
			now := time.Now()

			for _, dc := range req.Cookies {
				cookie := &coredata.Cookie{
					ID:               gid.New(scope.GetTenantID(), coredata.CookieEntityType),
					OrganizationID:   banner.OrganizationID,
					CookieBannerID:   banner.ID,
					CookieCategoryID: uncategorised.ID,
					Name:             dc.Name,
					Duration:         dc.Duration,
					Description:      "",
					CreatedAt:        now,
					UpdatedAt:        now,
				}

				ok, err := cookie.InsertIfNotExists(ctx, tx, scope)
				if err != nil {
					return fmt.Errorf("cannot insert detected cookie: %w", err)
				}
				if ok {
					inserted++
				}
			}

			if inserted > 0 {
				if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, banner.ID); err != nil {
					return fmt.Errorf("cannot ensure draft version: %w", err)
				}
			}

			return nil
		},
	)
}
