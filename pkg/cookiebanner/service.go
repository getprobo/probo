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
	"slices"
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

	ReorderCookieCategoryRequest struct {
		CookieCategoryID gid.GID
		Rank             int
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
		Regulation  *Regulation
		CountryCode *coredata.CountryCode
	}

	DetectedCookie struct {
		Name          string
		MaxAgeSeconds *int
		Source        coredata.CookieSource
	}

	ReportDetectedCookiesRequest struct {
		Cookies []DetectedCookie
	}

	DetectedStorageItem struct {
		Key         string
		StorageType coredata.TrackerType
		ValueSize   *int
	}

	DetectedResourceItem struct {
		Origin       string
		ResourceType coredata.TrackerType
	}

	ReportDetectedTrackersRequest struct {
		Cookies   []DetectedCookie
		Storage   []DetectedStorageItem
		Resources []DetectedResourceItem
	}

	CreateTrackerPatternRequest struct {
		CookieCategoryID gid.GID
		TrackerType      coredata.TrackerType
		Pattern          string
		MatchType        coredata.TrackerPatternMatchType
		DisplayName      string
		MaxAgeSeconds    *int
		Description      string
		Source           *coredata.CookieSource
	}

	UpdateTrackerPatternRequest struct {
		TrackerPatternID gid.GID
		DisplayName      *string
		MaxAgeSeconds    **int
		Description      *string
		Excluded         *bool
	}

	MoveTrackerPatternToCategoryRequest struct {
		TrackerPatternID       gid.GID
		TargetCookieCategoryID gid.GID
	}

	MoveTrackerPatternToCategoryResult struct {
		TrackerPattern *coredata.TrackerPattern
		Banner         *coredata.CookieBanner
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
		Regulation        Regulation                                     `json:"regulation"`
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

func (r *ReorderCookieCategoryRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieCategoryID, "cookie_category_id", validator.Required(), validator.GID(coredata.CookieCategoryEntityType))
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

		validators := []validator.ValidatorFunc{validator.NoHTML(), validator.MaxLen(2000)}
		if key == "banner_description" {
			validators = append(validators, validator.ContainsSubstring("{{cookie_policy_link}}"))
		}
		v.Check(s, "translations."+key, validators...)
	}

	return v.Error()
}

func (r *CreateTrackerPatternRequest) Validate() error {
	v := validator.New()

	v.Check(r.CookieCategoryID, "cookie_category_id", validator.Required(), validator.GID(coredata.CookieCategoryEntityType))
	v.Check(string(r.TrackerType), "tracker_type", validator.Required(), validator.OneOfSlice(
		func() []string {
			types := coredata.TrackerTypes()
			s := make([]string, len(types))
			for i, t := range types {
				s[i] = string(t)
			}
			return s
		}(),
	))
	v.Check(r.Pattern, "pattern", validator.Required(), validator.SafeTextNoNewLine(255))
	v.Check(string(r.MatchType), "match_type", validator.Required(), validator.OneOfSlice(
		func() []string {
			types := coredata.TrackerPatternMatchTypes()
			s := make([]string, len(types))
			for i, t := range types {
				s[i] = string(t)
			}
			return s
		}(),
	))
	v.Check(r.DisplayName, "display_name", validator.Required(), validator.SafeTextNoNewLine(255))
	v.Check(r.Description, "description", validator.SafeText(1000))

	return v.Error()
}

func (r *UpdateTrackerPatternRequest) Validate() error {
	v := validator.New()

	v.Check(r.TrackerPatternID, "tracker_pattern_id", validator.Required(), validator.GID(coredata.TrackerPatternEntityType))
	if r.DisplayName != nil {
		v.Check(*r.DisplayName, "display_name", validator.Required(), validator.SafeTextNoNewLine(255))
	}
	if r.Description != nil {
		v.Check(*r.Description, "description", validator.SafeText(1000))
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

func (s *Service) ensureDraftVersion(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	banner *coredata.CookieBanner,
	categories coredata.CookieCategories,
	allPatterns coredata.TrackerPatterns,
) (*coredata.CookieBannerVersion, error) {
	snapshot := buildSnapshot(banner, categories, allPatterns)

	var latest coredata.CookieBannerVersion
	err := latest.LoadLatestByCookieBannerID(ctx, tx, scope, banner.ID)

	if err == nil {
		if latestSnapshot, snapErr := latest.GetSnapshot(); snapErr == nil && snapshotsEqual(snapshot, latestSnapshot) {
			return &latest, nil
		}

		if latest.State == coredata.CookieBannerVersionStateDraft {
			if err := latest.SetSnapshot(snapshot); err != nil {
				return nil, fmt.Errorf("cannot set snapshot: %w", err)
			}
			latest.UpdatedAt = time.Now()
			if err := latest.Update(ctx, tx, scope); err != nil {
				return nil, fmt.Errorf("cannot update draft version: %w", err)
			}
			return &latest, nil
		}
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
	if err := categories.LoadAllConsentCategoriesByCookieBannerID(ctx, tx, scope, bannerID); err != nil {
		return nil, fmt.Errorf("cannot load cookie categories: %w", err)
	}

	var allPatterns coredata.TrackerPatterns
	if err := allPatterns.LoadAllByCookieBannerID(
		ctx,
		tx,
		scope,
		bannerID,
		coredata.NewTrackerPatternFilter(nil, nil, new(false)),
		nil,
	); err != nil {
		return nil, fmt.Errorf("cannot load tracker patterns: %w", err)
	}

	return s.ensureDraftVersion(ctx, tx, scope, &banner, categories, allPatterns)
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
					consentMaxAge := req.ConsentExpiryDays * 86400
					consentPattern := &coredata.TrackerPattern{
						ID:               gid.New(scope.GetTenantID(), coredata.TrackerPatternEntityType),
						OrganizationID:   banner.OrganizationID,
						CookieBannerID:   banner.ID,
						CookieCategoryID: category.ID,
						TrackerType:      coredata.TrackerTypeCookie,
						Pattern:          "probo_consent",
						MatchType:        coredata.TrackerPatternMatchTypeExact,
						DisplayName:      "probo_consent",
						MaxAgeSeconds:    &consentMaxAge,
						Description:      "Stores your cookie consent preferences for this website.",
						Source:           new(coredata.CookieSourceScript),
						CreatedAt:        now,
						UpdatedAt:        now,
					}
					if err := consentPattern.Insert(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot insert probo_consent pattern: %w", err)
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

func (s *Service) GetCookieBannersByIDs(
	ctx context.Context,
	scope coredata.Scoper,
	bannerIDs ...gid.GID,
) (coredata.CookieBanners, error) {
	var banners coredata.CookieBanners

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := banners.LoadByIDs(ctx, conn, scope, bannerIDs); err != nil {
				return fmt.Errorf("cannot load cookie banners by ids: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return banners, nil
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

			nameChanged := req.Name != nil && *req.Name != banner.Name
			privacyChanged := req.PrivacyPolicyURL != nil && !ptrEqual(req.PrivacyPolicyURL, banner.PrivacyPolicyURL)
			cookiePolicyChanged := req.CookiePolicyURL != nil && *req.CookiePolicyURL != banner.CookiePolicyURL
			expiryChanged := req.ConsentExpiryDays != nil && *req.ConsentExpiryDays != banner.ConsentExpiryDays
			consentModeChanged := req.ConsentMode != nil && *req.ConsentMode != banner.ConsentMode
			defaultLangChanged := req.DefaultLanguage != nil && *req.DefaultLanguage != banner.DefaultLanguage

			snapshotChanged := privacyChanged || cookiePolicyChanged || expiryChanged || consentModeChanged || defaultLangChanged

			if !nameChanged && !snapshotChanged {
				return nil
			}

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

			if snapshotChanged {
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

func (s *Service) GetCookieCategoriesByIDs(
	ctx context.Context,
	scope coredata.Scoper,
	categoryIDs ...gid.GID,
) (coredata.CookieCategories, error) {
	var categories coredata.CookieCategories

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := categories.LoadByIDs(ctx, conn, scope, categoryIDs); err != nil {
				return fmt.Errorf("cannot load cookie categories by ids: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return categories, nil
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
			if err := categories.LoadConsentCategoriesByCookieBannerID(ctx, conn, scope, bannerID, cursor); err != nil {
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

			count, err = categories.CountConsentCategoriesByCookieBannerID(ctx, conn, scope, bannerID)
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

			nameChanged := req.Name != nil && *req.Name != category.Name
			slugChanged := req.Slug != nil && *req.Slug != category.Slug
			descChanged := req.Description != nil && *req.Description != category.Description
			gcmChanged := req.GCMConsentTypes != nil && !slices.Equal(*req.GCMConsentTypes, category.GCMConsentTypes)
			posthogChanged := req.PostHogConsent != nil && *req.PostHogConsent != category.PostHogConsent

			if !nameChanged && !slugChanged && !descChanged && !gcmChanged && !posthogChanged {
				return nil
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
			if posthogChanged {
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

			if err := banner.LoadByID(ctx, tx, scope, category.CookieBannerID); err != nil {
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			if category.Rank == req.Rank {
				return nil
			}

			category.Rank = req.Rank
			category.UpdatedAt = time.Now()

			if err := category.UpdateRank(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot reorder cookie category: %w", err)
			}

			// Rank is admin-only metadata; the snapshot is sorted by
			// (Kind weight, ID) in buildSnapshot, so reordering does not
			// affect visitor view and must not bump the version.

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

			var patterns coredata.TrackerPatterns
			if err := patterns.MoveToCategoryByCookieCategoryID(ctx, tx, scope, category.ID, uncategorised.ID); err != nil {
				return fmt.Errorf("cannot move tracker patterns to uncategorised: %w", err)
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

func (s *Service) GetCookieConsentRecord(
	ctx context.Context,
	scope coredata.Scoper,
	id gid.GID,
) (*coredata.CookieConsentRecord, error) {
	var record coredata.CookieConsentRecord

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := record.LoadByID(ctx, conn, scope, id); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrConsentNotFound
				}

				return fmt.Errorf("cannot load consent record: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &record, nil
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
	regulation Regulation,
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

			var categories coredata.CookieCategories
			if err := categories.LoadAllConsentCategoriesByCookieBannerID(ctx, conn, scope, banner.ID); err != nil {
				return fmt.Errorf("cannot load cookie categories: %w", err)
			}

			var translations coredata.CookieBannerTranslations
			if err := translations.LoadAllByCookieBannerID(ctx, conn, scope, banner.ID); err != nil {
				return fmt.Errorf("cannot load cookie banner translations: %w", err)
			}

			resolved := resolveTranslations(translations, categories)
			config = buildBannerConfig(&banner, &version, &snapshot, resolved, lang)

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	config.Regulation = regulation
	if cm := ConsentModeForRegulation(regulation); cm != "" {
		config.ConsentMode = cm
	}
	remapTextsForConsentMode(config.Texts, config.ConsentMode, regulation)

	return config, nil
}

func buildBannerConfig(
	banner *coredata.CookieBanner,
	version *coredata.CookieBannerVersion,
	snapshot *coredata.CookieBannerVersionSnapshot,
	translations map[string]coredata.CookieBannerVersionSnapshotTranslation,
	lang string,
) *BannerConfig {
	defaultLang := snapshot.DefaultLanguage
	if defaultLang == "" {
		defaultLang = "en"
	}

	resolvedLang := defaultLang
	if lang != "" {
		if _, ok := translations[lang]; ok {
			resolvedLang = lang
		}
	}

	categories := make([]coredata.CookieBannerVersionSnapshotCategory, 0, len(snapshot.Categories))
	for _, c := range snapshot.Categories {
		if c.Kind != coredata.CookieCategoryKindUncategorised {
			categories = append(categories, c)
		}
	}
	texts := make(map[string]string)

	if t, ok := translations[resolvedLang]; ok {
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

// remapTextsForConsentMode overrides the generic banner text keys with
// mode-specific variants so the client renders the appropriate copy
// without needing consent-mode awareness itself.
func remapTextsForConsentMode(texts map[string]string, consentMode string, regulation Regulation) {
	if texts == nil {
		return
	}

	switch {
	case consentMode == ConsentModeOptOut:
		remapTextKey(texts, "banner_title_opt_out", "banner_title")
		remapTextKey(texts, "banner_description_opt_out", "banner_description")
		remapTextKey(texts, "button_acknowledge", "button_accept_all")
		remapTextKey(texts, "button_opt_out", "button_customize")
		texts["button_reject_all"] = ""

	case regulation == RegulationNone:
		remapTextKey(texts, "banner_title_notice", "banner_title")
		remapTextKey(texts, "banner_description_notice", "banner_description")
		remapTextKey(texts, "button_dismiss", "button_accept_all")
		texts["button_reject_all"] = ""
		texts["button_customize"] = ""
	}
}

func remapTextKey(texts map[string]string, src, dst string) {
	if v, ok := texts[src]; ok && v != "" {
		texts[dst] = v
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
				same, eqErr := jsonEqual(existing.Translations, req.Translations)
				if eqErr == nil && same {
					result = &existing
					return nil
				}

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
				Regulation:            req.Regulation,
				CountryCode:           req.CountryCode,
				CreatedAt:             time.Now(),
			}

			if record.Regulation != nil && *record.Regulation == coredata.RegulationNone {
				record.Regulation = nil
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
	return s.ReportDetectedTrackers(ctx,
		bannerID,
		ReportDetectedTrackersRequest{
			Cookies: req.Cookies,
		},
	)
}

func (s *Service) ReportDetectedTrackers(
	ctx context.Context,
	bannerID gid.GID,
	req ReportDetectedTrackersRequest,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(bannerID)

			var banner coredata.CookieBanner
			if err := banner.LoadByID(ctx, tx, scope, bannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			var uncategorised coredata.CookieCategory
			if err := uncategorised.LoadUncategorisedByCookieBannerID(ctx, tx, scope, banner.ID); err != nil {
				return fmt.Errorf("cannot load uncategorised category: %w", err)
			}

			inserted := 0
			now := time.Now()

			for _, dc := range req.Cookies {
				if err := s.reportDetectedTracker(
					ctx,
					tx,
					scope,
					&banner,
					uncategorised.ID,
					now,
					detectedTrackerInfo{
						TrackerType:   coredata.TrackerTypeCookie,
						Identifier:    dc.Name,
						MaxAgeSeconds: dc.MaxAgeSeconds,
						Source:        &dc.Source,
					},
					&inserted,
				); err != nil {
					return err
				}
			}

			for _, ds := range req.Storage {
				if err := s.reportDetectedTracker(
					ctx,
					tx,
					scope,
					&banner,
					uncategorised.ID,
					now,
					detectedTrackerInfo{
						TrackerType: ds.StorageType,
						Identifier:  ds.Key,
						ValueSize:   ds.ValueSize,
					},
					&inserted,
				); err != nil {
					return err
				}
			}

			for _, dr := range req.Resources {
				if err := s.reportDetectedTracker(
					ctx,
					tx,
					scope,
					&banner,
					uncategorised.ID,
					now,
					detectedTrackerInfo{
						TrackerType: dr.ResourceType,
						Identifier:  dr.Origin,
					},
					&inserted,
				); err != nil {
					return err
				}
			}

			if inserted > 0 {
				if err := banner.SetPatternAnalysisRequested(ctx, tx); err != nil {
					return fmt.Errorf("cannot request pattern analysis: %w", err)
				}
			}

			return nil
		},
	)
}

type detectedTrackerInfo struct {
	TrackerType   coredata.TrackerType
	Identifier    string
	MaxAgeSeconds *int
	Source        *coredata.CookieSource
	ValueSize     *int
}

func (s *Service) reportDetectedTracker(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	banner *coredata.CookieBanner,
	uncategorisedID gid.GID,
	now time.Time,
	info detectedTrackerInfo,
	inserted *int,
) error {
	var matchedPattern coredata.TrackerPattern
	err := matchedPattern.FindMatchingPattern(ctx, tx, scope, banner.ID, info.TrackerType, info.Identifier)
	if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return fmt.Errorf("cannot find matching tracker pattern: %w", err)
	}

	if err == nil && matchedPattern.Excluded {
		return nil
	}

	var patternID *gid.GID
	if err == nil {
		patternID = &matchedPattern.ID
		matchedPattern.LastMatchedAt = &now
		matchedPattern.UpdatedAt = now
		if updateErr := matchedPattern.Update(ctx, tx, scope); updateErr != nil {
			return fmt.Errorf("cannot update tracker pattern last_matched_at: %w", updateErr)
		}
	} else {
		newPattern := &coredata.TrackerPattern{
			ID:               gid.New(scope.GetTenantID(), coredata.TrackerPatternEntityType),
			OrganizationID:   banner.OrganizationID,
			CookieBannerID:   banner.ID,
			CookieCategoryID: uncategorisedID,
			TrackerType:      info.TrackerType,
			Pattern:          info.Identifier,
			MatchType:        coredata.TrackerPatternMatchTypeExact,
			DisplayName:      info.Identifier,
			Description:      "",
			MaxAgeSeconds:    info.MaxAgeSeconds,
			Source:           info.Source,
			LastMatchedAt:    &now,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		wasInserted, err := newPattern.InsertIfNotExists(ctx, tx, scope)
		if err != nil {
			return fmt.Errorf("cannot insert tracker pattern: %w", err)
		}
		if wasInserted {
			patternID = &newPattern.ID
			*inserted++
		} else {
			var existingPattern coredata.TrackerPattern
			if err := existingPattern.FindMatchingPattern(ctx, tx, scope, banner.ID, info.TrackerType, info.Identifier); err != nil {
				return fmt.Errorf("cannot load existing tracker pattern: %w", err)
			}
			patternID = &existingPattern.ID
		}
	}

	tracker := &coredata.DetectedTracker{
		ID:               gid.New(scope.GetTenantID(), coredata.DetectedTrackerEntityType),
		CookieBannerID:   banner.ID,
		TrackerPatternID: patternID,
		TrackerType:      info.TrackerType,
		Identifier:       info.Identifier,
		MaxAgeSeconds:    info.MaxAgeSeconds,
		Source:           info.Source,
		ValueSize:        info.ValueSize,
		LastDetectedAt:   now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if _, err := tracker.InsertIfNotExists(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert detected tracker: %w", err)
	}

	return nil
}

func (s *Service) CreateTrackerPattern(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateTrackerPatternRequest,
) (*coredata.TrackerPattern, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var pattern *coredata.TrackerPattern

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

			pattern = &coredata.TrackerPattern{
				ID:               gid.New(scope.GetTenantID(), coredata.TrackerPatternEntityType),
				OrganizationID:   category.OrganizationID,
				CookieBannerID:   category.CookieBannerID,
				CookieCategoryID: category.ID,
				TrackerType:      req.TrackerType,
				Pattern:          req.Pattern,
				MatchType:        req.MatchType,
				DisplayName:      req.DisplayName,
				MaxAgeSeconds:    req.MaxAgeSeconds,
				Description:      req.Description,
				Source:           req.Source,
				CreatedAt:        now,
				UpdatedAt:        now,
			}

			if err := pattern.Insert(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrPatternAlreadyExists
				}
				return fmt.Errorf("cannot insert tracker pattern: %w", err)
			}

			if !pattern.Excluded && pattern.TrackerType == coredata.TrackerTypeCookie {
				if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, pattern.CookieBannerID); err != nil {
					return fmt.Errorf("cannot ensure draft version: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return pattern, nil
}

func (s *Service) ListTrackerPatternsForCategory(
	ctx context.Context,
	scope coredata.Scoper,
	categoryID gid.GID,
	cursor *page.Cursor[coredata.TrackerPatternOrderField],
) (coredata.TrackerPatterns, error) {
	var patterns coredata.TrackerPatterns

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return patterns.LoadByCookieCategoryID(ctx, conn, scope, categoryID, cursor)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list tracker patterns for category: %w", err)
	}

	return patterns, nil
}

func (s *Service) CountTrackerPatternsForCategory(
	ctx context.Context,
	scope coredata.Scoper,
	categoryID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var patterns coredata.TrackerPatterns
			var err error

			count, err = patterns.CountByCookieCategoryID(ctx, conn, scope, categoryID)
			return err
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count tracker patterns for category: %w", err)
	}

	return count, nil
}

func (s *Service) GetTrackerPattern(
	ctx context.Context,
	scope coredata.Scoper,
	trackerPatternID gid.GID,
) (*coredata.TrackerPattern, error) {
	var pattern coredata.TrackerPattern

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := pattern.LoadByID(ctx, conn, scope, trackerPatternID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrTrackerPatternNotFound
				}
				return fmt.Errorf("cannot load tracker pattern: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &pattern, nil
}

func (s *Service) UpdateTrackerPattern(
	ctx context.Context,
	scope coredata.Scoper,
	req UpdateTrackerPatternRequest,
) (*coredata.TrackerPattern, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var pattern coredata.TrackerPattern

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := pattern.LoadByID(ctx, tx, scope, req.TrackerPatternID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrTrackerPatternNotFound
				}
				return fmt.Errorf("cannot load tracker pattern: %w", err)
			}

			displayNameChanged := req.DisplayName != nil && *req.DisplayName != pattern.DisplayName
			maxAgeChanged := req.MaxAgeSeconds != nil && !ptrEqual(*req.MaxAgeSeconds, pattern.MaxAgeSeconds)
			descChanged := req.Description != nil && *req.Description != pattern.Description
			excludedChanged := req.Excluded != nil && *req.Excluded != pattern.Excluded

			if !displayNameChanged && !maxAgeChanged && !descChanged && !excludedChanged {
				return nil
			}

			staysExcluded := pattern.Excluded && (req.Excluded == nil || *req.Excluded)

			if req.DisplayName != nil {
				pattern.DisplayName = *req.DisplayName
			}
			if req.MaxAgeSeconds != nil {
				pattern.MaxAgeSeconds = *req.MaxAgeSeconds
			}
			if req.Description != nil {
				pattern.Description = *req.Description
			}
			if req.Excluded != nil {
				pattern.Excluded = *req.Excluded
			}

			pattern.UpdatedAt = time.Now()

			if err := pattern.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update tracker pattern: %w", err)
			}

			if !staysExcluded && pattern.TrackerType == coredata.TrackerTypeCookie {
				if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, pattern.CookieBannerID); err != nil {
					return fmt.Errorf("cannot ensure draft version: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &pattern, nil
}

func (s *Service) DeleteTrackerPattern(
	ctx context.Context,
	scope coredata.Scoper,
	trackerPatternID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var pattern coredata.TrackerPattern
			if err := pattern.LoadByID(ctx, tx, scope, trackerPatternID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrTrackerPatternNotFound
				}
				return fmt.Errorf("cannot load tracker pattern: %w", err)
			}

			wasExcluded := pattern.Excluded

			if err := pattern.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete tracker pattern: %w", err)
			}

			if !wasExcluded && pattern.TrackerType == coredata.TrackerTypeCookie {
				if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, pattern.CookieBannerID); err != nil {
					return fmt.Errorf("cannot ensure draft version: %w", err)
				}
			}

			return nil
		},
	)
}

func (s *Service) MoveTrackerPatternToCategory(
	ctx context.Context,
	scope coredata.Scoper,
	req MoveTrackerPatternToCategoryRequest,
) (*MoveTrackerPatternToCategoryResult, error) {
	var result MoveTrackerPatternToCategoryResult

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var pattern coredata.TrackerPattern
			if err := pattern.LoadByID(ctx, tx, scope, req.TrackerPatternID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrTrackerPatternNotFound
				}
				return fmt.Errorf("cannot load tracker pattern: %w", err)
			}

			var target coredata.CookieCategory
			if err := target.LoadByID(ctx, tx, scope, req.TargetCookieCategoryID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrCategoryNotFound
				}
				return fmt.Errorf("cannot load target cookie category: %w", err)
			}

			if pattern.CookieCategoryID == target.ID {
				return ErrSamePatternCategoryMove
			}

			if pattern.CookieBannerID != target.CookieBannerID {
				return ErrCategoriesBannerMismatch
			}

			wasExcluded := pattern.Excluded

			pattern.CookieCategoryID = target.ID
			pattern.UpdatedAt = time.Now()

			if err := pattern.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update tracker pattern: %w", err)
			}

			var banner coredata.CookieBanner
			if err := banner.LoadByID(ctx, tx, scope, pattern.CookieBannerID); err != nil {
				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			if !wasExcluded && pattern.TrackerType == coredata.TrackerTypeCookie {
				if _, err := s.ensureDraftVersionForBanner(ctx, tx, scope, pattern.CookieBannerID); err != nil {
					return fmt.Errorf("cannot ensure draft version: %w", err)
				}
			}

			result.TrackerPattern = &pattern
			result.Banner = &banner

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *Service) ListUncategorisedTrackerPatterns(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
	cursor *page.Cursor[coredata.TrackerPatternOrderField],
	filter *coredata.TrackerPatternFilter,
) (coredata.TrackerPatterns, error) {
	var patterns coredata.TrackerPatterns

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := patterns.LoadUncategorisedByCookieBannerID(ctx, conn, scope, bannerID, cursor, filter); err != nil {
				return fmt.Errorf("cannot list uncategorised tracker patterns: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return patterns, nil
}

func (s *Service) CountUncategorisedTrackerPatterns(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
	filter *coredata.TrackerPatternFilter,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var patterns coredata.TrackerPatterns
			var err error

			count, err = patterns.CountUncategorisedByCookieBannerID(ctx, conn, scope, bannerID, filter)
			if err != nil {
				return fmt.Errorf("cannot count uncategorised tracker patterns: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) CountDetectedTrackersByPatternID(
	ctx context.Context,
	scope coredata.Scoper,
	trackerPatternID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var trackers coredata.DetectedTrackers
			var err error

			count, err = trackers.CountByTrackerPatternID(ctx, conn, scope, trackerPatternID)
			if err != nil {
				return fmt.Errorf("cannot count detected trackers: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
