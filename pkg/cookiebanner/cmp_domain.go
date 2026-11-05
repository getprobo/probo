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
) error {
	if banner.Capabilities.Corsless || banner.CMPDomainID == nil {
		return nil
	}

	domain := &coredata.CMPDomain{ID: *banner.CMPDomainID}
	if err := domain.SetSourcePromotionRequested(ctx, tx); err != nil {
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

func (s *Service) promoteSiblingCookieSources(
	ctx context.Context,
	tx pg.Tx,
	domain *coredata.CMPDomain,
) error {
	scope := coredata.NewScopeFromObjectID(domain.ID)

	banners, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.CookieBannerOrderField]{
			Field:     coredata.CookieBannerOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.CookieBannerOrderField]) ([]*coredata.CookieBanner, error) {
			var batch coredata.CookieBanners
			if err := batch.LoadByCMPDomainID(ctx, tx, scope, domain.ID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load cmp domain banners: %w", err)
			}

			return []*coredata.CookieBanner(batch), nil
		},
	)
	if err != nil {
		return err
	}

	type bannerTrackers struct {
		banner     *coredata.CookieBanner
		patterns   []*coredata.TrackerPattern
		detections []*coredata.DetectedTracker
	}

	loaded := make([]bannerTrackers, 0, len(banners))
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
				if err := batch.LoadByCookieBannerID(ctx, tx, scope, banner.ID, cursor, nil); err != nil {
					return nil, fmt.Errorf("cannot load tracker patterns: %w", err)
				}

				return batch, nil
			},
		)
		if err != nil {
			return err
		}

		detections, err := page.LoadAll(
			ctx,
			page.OrderBy[coredata.DetectedTrackerOrderField]{
				Field:     coredata.DetectedTrackerOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
			func(ctx context.Context, cursor *page.Cursor[coredata.DetectedTrackerOrderField]) ([]*coredata.DetectedTracker, error) {
				var batch coredata.DetectedTrackers

				filter := coredata.NewDetectedTrackerFilter([]coredata.TrackerType{coredata.TrackerTypeCookie})
				if err := batch.LoadByCookieBannerID(ctx, tx, scope, banner.ID, filter, cursor); err != nil {
					return nil, fmt.Errorf("cannot load detected trackers: %w", err)
				}

				return batch, nil
			},
		)
		if err != nil {
			return err
		}

		loaded = append(loaded, bannerTrackers{
			banner:     banner,
			patterns:   patterns,
			detections: detections,
		})
	}

	best := map[gid.GID]siblingSourceCandidate{}

	for _, banner := range loaded {
		hostOnly := map[string]struct{}{}

		for _, detection := range banner.detections {
			if detection.HostOnly != nil && *detection.HostOnly {
				hostOnly[detection.Identifier] = struct{}{}
			}
		}

		for _, other := range loaded {
			if other.banner.ID == banner.banner.ID {
				continue
			}

			patternsByID := make(map[gid.GID]*coredata.TrackerPattern, len(other.patterns))
			for _, pattern := range other.patterns {
				patternsByID[pattern.ID] = pattern
			}

			for _, pattern := range banner.patterns {
				if pattern.Excluded || pattern.TrackerType != coredata.TrackerTypeCookie || pattern.Source == nil || *pattern.Source != coredata.CookieSourcePreExisting {
					continue
				}

				if _, blocked := hostOnly[pattern.Pattern]; blocked {
					continue
				}

				for _, detection := range other.detections {
					candidate, ok := siblingCandidate(pattern, banner.banner, banner.detections, detection, patternsByID, other.banner.ID)
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

func siblingCandidate(
	pattern *coredata.TrackerPattern,
	banner *coredata.CookieBanner,
	bannerDetections []*coredata.DetectedTracker,
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

	matched := false

	for _, existing := range bannerDetections {
		if existing.TrackerPatternID == nil || *existing.TrackerPatternID != pattern.ID || existing.Source == nil || *existing.Source != coredata.CookieSourcePreExisting {
			continue
		}

		if existing.CookieDomain == nil || strings.EqualFold(*existing.CookieDomain, *detection.CookieDomain) {
			matched = true
			break
		}
	}

	if !matched {
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
