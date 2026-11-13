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

package cookiebanner

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"golang.org/x/net/publicsuffix"

	"go.probo.inc/probo/pkg/cookiebanner/scanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/uri"
)

func domainNameForOrigin(origin string) string {
	host := hostOfOrigin(origin)
	if host == "" {
		host = strings.ToLower(strings.TrimSpace(origin))
	}

	domain := uri.ICANNRegistrableDomain(host)
	if domain == "" {
		return host
	}

	return domain
}

func hostOfOrigin(origin string) string {
	parsed, err := url.Parse(strings.TrimSpace(origin))
	if err != nil {
		return ""
	}

	return strings.ToLower(parsed.Hostname())
}

// cookieDomainCoversHost reports whether a cookie set for cookieDomain is
// sent to host. A public suffix never covers a host.
func cookieDomainCoversHost(cookieDomain, host string) bool {
	cookieDomain = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(cookieDomain), "."))

	host = strings.ToLower(strings.TrimSpace(host))
	if cookieDomain == "" || host == "" {
		return false
	}

	suffix, _ := publicsuffix.PublicSuffix(host)
	if suffix == cookieDomain {
		return false
	}

	if cookieDomain == host {
		return true
	}

	return strings.HasSuffix(host, "."+cookieDomain)
}

func canonicalSurfaceOrigin(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" {
		return "", false
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", false
	}

	host := strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")

	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return "", false
	}

	port := parsed.Port()
	if (parsed.Scheme == "https" && port == "443") || (parsed.Scheme == "http" && port == "80") {
		port = ""
	}

	origin := parsed.Scheme + "://" + host
	if port != "" {
		origin += ":" + port
	}

	return origin, true
}

func sourcePromotionRelevant(info detectedTrackerInfo) bool {
	if info.TrackerType != coredata.TrackerTypeCookie || info.Source == nil {
		return false
	}

	switch *info.Source {
	case coredata.CookieSourcePreExisting:
		return true
	case coredata.CookieSourceScript, coredata.CookieSourceHTTP:
		return info.HostOnly != nil && !*info.HostOnly
	default:
		return false
	}
}

func (s *Service) ensureCMPDomain(
	ctx context.Context,
	tx pg.Tx,
	organizationID gid.GID,
	tenantID gid.TenantID,
	origin string,
	now time.Time,
) (*coredata.CMPDomain, error) {
	domain := coredata.CMPDomain{
		ID: gid.New(tenantID, coredata.CMPDomainEntityType),
	}
	if err := domain.UpsertByOrganizationAndDomain(
		ctx,
		tx,
		tenantID,
		organizationID,
		domainNameForOrigin(origin),
		now,
	); err != nil {
		return nil, fmt.Errorf("cannot attach cmp domain: %w", err)
	}

	return &domain, nil
}

func (s *Service) requestCMPSourcePromotion(
	ctx context.Context,
	tx pg.Tx,
	banner *coredata.CookieBanner,
	now time.Time,
) error {
	if banner.Capabilities.Corsless || banner.CMPDomainID == nil {
		return nil
	}

	if err := banner.SetSourcePromotionRequested(ctx, tx, now); err != nil {
		return fmt.Errorf("cannot request source promotion: %w", err)
	}

	return nil
}

func (s *Service) recordDiscoveredSurfaces(
	ctx context.Context,
	tx pg.Tx,
	banner *coredata.CookieBanner,
	scanID gid.GID,
	origins []string,
) error {
	if banner.Capabilities.Corsless || banner.CMPDomainID == nil || len(origins) == 0 {
		return nil
	}

	scope := coredata.NewScopeFromObjectID(banner.ID)
	now := time.Now()
	seen := map[string]struct{}{}

	for _, raw := range origins {
		if len(seen) >= scanner.MaxSiblingOrigins {
			break
		}

		origin, ok := canonicalSurfaceOrigin(raw)
		if !ok || strings.EqualFold(origin, banner.Origin) {
			continue
		}

		if _, dup := seen[origin]; dup {
			continue
		}

		seen[origin] = struct{}{}

		surface := &coredata.CMPDomainSurface{
			ID:                         gid.New(banner.ID.TenantID(), coredata.CMPDomainSurfaceEntityType),
			OrganizationID:             banner.OrganizationID,
			CMPDomainID:                *banner.CMPDomainID,
			Origin:                     origin,
			Status:                     coredata.CMPDomainSurfaceStatusUnconfirmed,
			DiscoveredByCookieBannerID: &banner.ID,
			LastCookieBannerScanID:     &scanID,
		}
		if err := surface.UpsertSeen(ctx, tx, scope, now); err != nil {
			return fmt.Errorf("cannot record discovered surface: %w", err)
		}
	}

	return nil
}

type siblingSourceCandidate struct {
	patternID gid.GID
	bannerID  gid.GID
	source    coredata.CookieSource
}

// cookieDomainSet is the pre-existing detections of one pattern. A nil
// cookie domain matches every sibling domain.
type cookieDomainSet struct {
	anyDomain bool
	domains   map[string]struct{}
}

func (s *cookieDomainSet) add(domain *string) {
	if domain == nil {
		s.anyDomain = true
		return
	}

	if s.domains == nil {
		s.domains = map[string]struct{}{}
	}

	s.domains[strings.ToLower(*domain)] = struct{}{}
}

func (s *cookieDomainSet) covers(domain string) bool {
	if s == nil {
		return false
	}

	if s.anyDomain {
		return true
	}

	_, ok := s.domains[strings.ToLower(domain)]
	return ok
}

func (s *Service) promoteSiblingCookieSources(
	ctx context.Context,
	tx pg.Tx,
	banner *coredata.CookieBanner,
) error {
	if banner.Capabilities.Corsless || banner.CMPDomainID == nil || banner.SourcePromotionSince == nil {
		return nil
	}

	scope := coredata.NewScopeFromObjectID(banner.ID)
	recentFilter := coredata.NewDetectedTrackerFilter([]coredata.TrackerType{coredata.TrackerTypeCookie}).
		WithUpdatedSince(banner.SourcePromotionSince)

	candidates := map[string]struct{}{}

	err := page.WalkAll(
		ctx,
		page.OrderBy[coredata.DetectedTrackerOrderField]{
			Field:     coredata.DetectedTrackerOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.DetectedTrackerOrderField]) ([]*coredata.DetectedTracker, error) {
			var batch coredata.DetectedTrackers
			if err := batch.LoadByCookieBannerID(ctx, tx, scope, banner.ID, recentFilter, cursor); err != nil {
				return nil, fmt.Errorf("cannot load detected trackers: %w", err)
			}

			return batch, nil
		},
		func(rows []*coredata.DetectedTracker) error {
			for _, detection := range rows {
				if !sourcePromotionRelevant(detectedTrackerInfo{
					TrackerType: detection.TrackerType,
					Source:      detection.Source,
					HostOnly:    detection.HostOnly,
				}) {
					continue
				}

				candidates[detection.Identifier] = struct{}{}
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	if len(candidates) == 0 {
		return nil
	}

	cursor := page.NewCursor(
		page.MaxCursorSize,
		nil,
		page.Head,
		page.OrderBy[coredata.CookieBannerOrderField]{
			Field:     coredata.CookieBannerOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
	)

	var domainBanners coredata.CookieBanners
	if err := domainBanners.LoadByCMPDomainID(ctx, tx, scope, *banner.CMPDomainID, cursor); err != nil {
		return fmt.Errorf("cannot load cmp domain banners: %w", err)
	}

	banners := page.NewPage([]*coredata.CookieBanner(domainBanners), cursor).Data

	names := make([]string, 0, len(candidates))
	for name := range candidates {
		names = append(names, name)
	}

	slices.Sort(names)

	for batch := range slices.Chunk(names, page.MaxCursorSize) {
		if err := s.promoteSiblingCookieNames(ctx, tx, scope, banners, batch); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) promoteSiblingCookieNames(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	banners []*coredata.CookieBanner,
	names []string,
) error {
	cookie := coredata.TrackerTypeCookie

	type bannerTrackers struct {
		banner      *coredata.CookieBanner
		patterns    []*coredata.TrackerPattern
		byName      map[string][]*coredata.TrackerPattern
		hostOnly    map[string]struct{}
		preExisting map[gid.GID]*cookieDomainSet
	}

	patternsForNames := coredata.NewTrackerPatternFilter(nil, nil, new(false)).
		WithTrackerType(&cookie).
		WithIdentifiers(names)
	detectionsForNames := coredata.NewDetectedTrackerFilter([]coredata.TrackerType{coredata.TrackerTypeCookie}).
		WithIdentifiers(names)

	loaded := make([]*bannerTrackers, 0, len(banners))
	for _, banner := range banners {
		if banner.Capabilities.Corsless {
			continue
		}

		patterns, err := page.LoadAll(
			ctx,
			page.OrderBy[coredata.TrackerPatternOrderField]{
				Field:     coredata.TrackerPatternOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
			func(ctx context.Context, cursor *page.Cursor[coredata.TrackerPatternOrderField]) ([]*coredata.TrackerPattern, error) {
				var batch coredata.TrackerPatterns
				if err := batch.LoadByCookieBannerID(ctx, tx, scope, banner.ID, cursor, patternsForNames); err != nil {
					return nil, fmt.Errorf("cannot load tracker patterns: %w", err)
				}

				return batch, nil
			},
		)
		if err != nil {
			return err
		}

		tracked := &bannerTrackers{
			banner:      banner,
			patterns:    patterns,
			byName:      map[string][]*coredata.TrackerPattern{},
			hostOnly:    map[string]struct{}{},
			preExisting: map[gid.GID]*cookieDomainSet{},
		}
		for _, pattern := range patterns {
			if pattern.Excluded || pattern.TrackerType != coredata.TrackerTypeCookie || pattern.Source == nil || *pattern.Source != coredata.CookieSourcePreExisting {
				continue
			}

			tracked.byName[pattern.Pattern] = append(tracked.byName[pattern.Pattern], pattern)
		}

		err = walkBannerCookieDetections(
			ctx,
			tx,
			scope,
			banner.ID,
			detectionsForNames,
			func(rows []*coredata.DetectedTracker) error {
				for _, detection := range rows {
					if detection.HostOnly != nil && *detection.HostOnly {
						tracked.hostOnly[detection.Identifier] = struct{}{}
					}

					if detection.TrackerPatternID == nil || detection.Source == nil || *detection.Source != coredata.CookieSourcePreExisting {
						continue
					}

					set := tracked.preExisting[*detection.TrackerPatternID]
					if set == nil {
						set = &cookieDomainSet{}
						tracked.preExisting[*detection.TrackerPatternID] = set
					}

					set.add(detection.CookieDomain)
				}

				return nil
			},
		)
		if err != nil {
			return err
		}

		loaded = append(loaded, tracked)
	}

	best := map[gid.GID]siblingSourceCandidate{}

	for _, sibling := range loaded {
		patternsByID := make(map[gid.GID]*coredata.TrackerPattern, len(sibling.patterns))
		for _, pattern := range sibling.patterns {
			patternsByID[pattern.ID] = pattern
		}

		err := walkBannerCookieDetections(
			ctx,
			tx,
			scope,
			sibling.banner.ID,
			detectionsForNames,
			func(rows []*coredata.DetectedTracker) error {
				for _, detection := range rows {
					for _, reader := range loaded {
						if reader.banner.ID == sibling.banner.ID {
							continue
						}

						if _, blocked := reader.hostOnly[detection.Identifier]; blocked {
							continue
						}

						for _, pattern := range reader.byName[detection.Identifier] {
							candidate, ok := siblingCandidate(
								pattern,
								reader.banner,
								reader.preExisting[pattern.ID],
								detection,
								patternsByID,
								sibling.banner.ID,
							)
							if !ok {
								continue
							}

							current, seen := best[pattern.ID]
							if !seen || preferSiblingEvidence(candidate, current) {
								best[pattern.ID] = candidate
							}
						}
					}
				}

				return nil
			},
		)
		if err != nil {
			return err
		}
	}

	now := time.Now()

	for patternID, candidate := range best {
		var pattern coredata.TrackerPattern
		if err := pattern.LoadByIDForUpdate(ctx, tx, scope, patternID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				continue
			}

			return err
		}

		if pattern.Excluded || pattern.Source == nil || *pattern.Source != coredata.CookieSourcePreExisting {
			continue
		}

		source := candidate.source
		sibling := candidate.bannerID
		pattern.Source = &source
		pattern.SourceCookieBannerID = &sibling
		pattern.UpdatedAt = now

		if err := pattern.Update(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot promote tracker pattern source: %w", err)
		}

		if err := pattern.SetMappingRequested(ctx, tx); err != nil {
			return fmt.Errorf("cannot request mapping after source promotion: %w", err)
		}
	}

	return nil
}

func walkBannerCookieDetections(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	bannerID gid.GID,
	filter *coredata.DetectedTrackerFilter,
	walk func([]*coredata.DetectedTracker) error,
) error {
	return page.WalkAll(
		ctx,
		page.OrderBy[coredata.DetectedTrackerOrderField]{
			Field:     coredata.DetectedTrackerOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.DetectedTrackerOrderField]) ([]*coredata.DetectedTracker, error) {
			var batch coredata.DetectedTrackers
			if err := batch.LoadByCookieBannerID(ctx, tx, scope, bannerID, filter, cursor); err != nil {
				return nil, fmt.Errorf("cannot load detected trackers: %w", err)
			}

			return batch, nil
		},
		walk,
	)
}

func siblingCandidate(
	pattern *coredata.TrackerPattern,
	banner *coredata.CookieBanner,
	preExisting *cookieDomainSet,
	detection *coredata.DetectedTracker,
	siblingPatterns map[gid.GID]*coredata.TrackerPattern,
	siblingBannerID gid.GID,
) (siblingSourceCandidate, bool) {
	if detection.Identifier != pattern.Pattern || detection.TrackerPatternID == nil || detection.Source == nil || detection.CookieDomain == nil || *detection.CookieDomain == "" {
		return siblingSourceCandidate{}, false
	}

	if detection.HostOnly == nil || *detection.HostOnly {
		return siblingSourceCandidate{}, false
	}

	if *detection.Source != coredata.CookieSourceScript && *detection.Source != coredata.CookieSourceHTTP {
		return siblingSourceCandidate{}, false
	}

	siblingPattern, ok := siblingPatterns[*detection.TrackerPatternID]
	if !ok || siblingPattern.Excluded {
		return siblingSourceCandidate{}, false
	}

	if !cookieDomainCoversHost(*detection.CookieDomain, hostOfOrigin(banner.Origin)) {
		return siblingSourceCandidate{}, false
	}

	if !preExisting.covers(*detection.CookieDomain) {
		return siblingSourceCandidate{}, false
	}

	return siblingSourceCandidate{
		patternID: pattern.ID,
		bannerID:  siblingBannerID,
		source:    *detection.Source,
	}, true
}

func preferSiblingEvidence(candidate, current siblingSourceCandidate) bool {
	if sourceRank(&candidate.source) != sourceRank(&current.source) {
		return sourceRank(&candidate.source) > sourceRank(&current.source)
	}

	return candidate.bannerID.String() < current.bannerID.String()
}
