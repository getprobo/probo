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
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/cookiebanner/scanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func (s *Service) RequestScan(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
) (*coredata.CookieBannerScan, error) {
	var scan *coredata.CookieBannerScan

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var banner coredata.CookieBanner
			if err := banner.LoadByID(ctx, tx, scope, bannerID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrBannerNotFound
				}

				return fmt.Errorf("cannot load cookie banner: %w", err)
			}

			now := time.Now()
			scan = &coredata.CookieBannerScan{
				ID:             gid.New(scope.GetTenantID(), coredata.CookieBannerScanEntityType),
				OrganizationID: banner.OrganizationID,
				CookieBannerID: banner.ID,
				State:          coredata.CookieBannerScanStatePending,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := scan.Insert(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrScanAlreadyInProgress
				}

				return fmt.Errorf("cannot request cookie banner scan: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return scan, nil
}

func (s *Service) GetScan(
	ctx context.Context,
	scope coredata.Scoper,
	scanID gid.GID,
) (*coredata.CookieBannerScan, error) {
	var scan coredata.CookieBannerScan

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := scan.LoadByID(ctx, conn, scope, scanID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrScanNotFound
				}

				return fmt.Errorf("cannot load cookie banner scan: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &scan, nil
}

func (s *Service) ListScans(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
	cursor *page.Cursor[coredata.CookieBannerScanOrderField],
) (coredata.CookieBannerScans, error) {
	var scans coredata.CookieBannerScans

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := scans.LoadByCookieBannerID(ctx, conn, scope, bannerID, cursor); err != nil {
				return fmt.Errorf("cannot list cookie banner scans: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return scans, nil
}

func (s *Service) CountScans(
	ctx context.Context,
	scope coredata.Scoper,
	bannerID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var (
				scans coredata.CookieBannerScans
				err   error
			)

			count, err = scans.CountByCookieBannerID(ctx, conn, scope, bannerID)
			if err != nil {
				return fmt.Errorf("cannot count cookie banner scans: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ingestScanPage writes one finished page and asks for pattern analysis
// when something new was stored. It ignores ResourceReporting: that
// capability only gates SDK reports. It does not record sample hits.
func (s *Service) ingestScanPage(
	ctx context.Context,
	tx pg.Tx,
	banner *coredata.CookieBanner,
	scanID gid.GID,
	page scanner.Page,
) (int, error) {
	scope := coredata.NewScopeFromObjectID(banner.ID)

	var uncategorised coredata.CookieCategory
	if err := uncategorised.LoadUncategorisedByCookieBannerID(ctx, tx, scope, banner.ID); err != nil {
		return 0, fmt.Errorf("cannot load uncategorised category: %w", err)
	}

	inserted := 0
	now := time.Now()

	for i := range page.Cookies {
		cookie := page.Cookies[i]
		if _, err := s.persistDetectedTracker(
			ctx,
			tx,
			scope,
			banner,
			uncategorised.ID,
			now,
			detectedTrackerInfo{
				TrackerType:   coredata.TrackerTypeCookie,
				Identifier:    cookie.Name,
				MaxAgeSeconds: cookie.MaxAgeSeconds,
				Source:        new(cookie.Source),
				ObservedBy:    coredata.DetectedTrackerObserverScanner,
				ValueSize:     new(cookie.ValueSize),
				InitiatorURL:  cookie.InitiatorURL,
				CookieDomain:  cookie.Domain,
				HostOnly:      cookie.HostOnly,
				HTTPOnly:      cookie.HTTPOnly,
				Secure:        cookie.Secure,
				SameSite:      cookie.SameSite,
				Partitioned:   cookie.Partitioned,
				SetByURL:      cookie.SetByURL,
				ScanID:        &scanID,
			},
			&inserted,
		); err != nil {
			return 0, err
		}
	}

	for i := range page.Storage {
		item := page.Storage[i]
		if _, err := s.persistDetectedTracker(
			ctx,
			tx,
			scope,
			banner,
			uncategorised.ID,
			now,
			detectedTrackerInfo{
				TrackerType: item.Type,
				Identifier:  item.Key,
				ObservedBy:  coredata.DetectedTrackerObserverScanner,
				ValueSize:   new(item.ValueSize),
				ScanID:      &scanID,
			},
			&inserted,
		); err != nil {
			return 0, err
		}
	}

	for i := range page.Resources {
		item := page.Resources[i]

		wasInserted, err := s.upsertTrackerResource(
			ctx,
			tx,
			scope,
			banner,
			uncategorised.ID,
			now,
			item.Type,
			item.Origin,
			item.Path,
			&scanID,
		)
		if err != nil {
			return 0, err
		}

		if wasInserted {
			inserted++
		}
	}

	if inserted > 0 {
		if err := banner.SetPatternAnalysisRequested(ctx, tx); err != nil {
			return 0, fmt.Errorf("cannot request pattern analysis: %w", err)
		}
	}

	return inserted, nil
}
