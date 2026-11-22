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
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

const (
	DefaultAnalyticsDays = 7
	MaxAnalyticsDays     = 90
)

// utcDay returns the UTC midnight that starts the day of t. Analytics are
// bucketed by UTC day because organizations have no timezone.
func utcDay(t time.Time) time.Time {
	year, month, day := t.UTC().Date()

	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// analyticsSince returns the first day of the window of days ending today,
// today included.
func analyticsSince(now time.Time, days int) time.Time {
	return utcDay(now).AddDate(0, 0, 1-days)
}

func ValidateAnalyticsDays(days int) error {
	v := validator.New()

	v.Check(days, "days", validator.Min(1), validator.Max(MaxAnalyticsDays))

	return v.Error()
}

func (s *Service) GetPageLoads(
	ctx context.Context,
	scope coredata.Scoper,
	cookieBannerIDs []gid.GID,
	days int,
) (map[gid.GID][]BrowserFamilyCount, error) {
	if err := ValidateAnalyticsDays(days); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	result := make(map[gid.GID][]BrowserFamilyCount, len(cookieBannerIDs))
	for _, cookieBannerID := range cookieBannerIDs {
		result[cookieBannerID] = []BrowserFamilyCount{}
	}

	if len(cookieBannerIDs) == 0 {
		return result, nil
	}

	var rows []coredata.DailyPageLoad

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			loaded, err := coredata.LoadDailyPageLoadsByCookieBannerIDs(
				ctx,
				conn,
				scope,
				cookieBannerIDs,
				analyticsSince(s.now(), days),
			)
			if err != nil {
				return err
			}

			rows = loaded

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	rowsByBanner := make(map[gid.GID][]coredata.DailyPageLoad, len(cookieBannerIDs))
	for _, row := range rows {
		rowsByBanner[row.CookieBannerID] = append(rowsByBanner[row.CookieBannerID], row)
	}

	for cookieBannerID, bannerRows := range rowsByBanner {
		result[cookieBannerID] = PageLoadsByFamily(bannerRows)
	}

	return result, nil
}

func (s *Service) GetDailyPageLoads(
	ctx context.Context,
	scope coredata.Scoper,
	cookieBannerIDs []gid.GID,
	days int,
) (map[gid.GID][]DailyPageLoads, error) {
	if err := ValidateAnalyticsDays(days); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	since := analyticsSince(s.now(), days)

	result := make(map[gid.GID][]DailyPageLoads, len(cookieBannerIDs))
	for _, cookieBannerID := range cookieBannerIDs {
		result[cookieBannerID] = dailyPageLoadsSeries(nil, since, days)
	}

	if len(cookieBannerIDs) == 0 {
		return result, nil
	}

	var rows []coredata.DailyPageLoad

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			loaded, err := coredata.LoadDailyPageLoadsByCookieBannerIDs(
				ctx,
				conn,
				scope,
				cookieBannerIDs,
				since,
			)
			if err != nil {
				return err
			}

			rows = loaded

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	rowsByBanner := make(map[gid.GID][]coredata.DailyPageLoad, len(cookieBannerIDs))
	for _, row := range rows {
		rowsByBanner[row.CookieBannerID] = append(rowsByBanner[row.CookieBannerID], row)
	}

	for cookieBannerID, bannerRows := range rowsByBanner {
		result[cookieBannerID] = dailyPageLoadsSeries(bannerRows, since, days)
	}

	return result, nil
}

func (s *Service) GetDailyHits(
	ctx context.Context,
	scope coredata.Scoper,
	trackerPatternIDs []gid.GID,
	days int,
) (map[gid.GID][]DailyHits, error) {
	if err := ValidateAnalyticsDays(days); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	since := analyticsSince(s.now(), days)

	result := make(map[gid.GID][]DailyHits, len(trackerPatternIDs))
	for _, trackerPatternID := range trackerPatternIDs {
		result[trackerPatternID] = dailyHitsSeries(nil, since, days)
	}

	if len(trackerPatternIDs) == 0 {
		return result, nil
	}

	var rows []coredata.DailyPatternHit

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			loaded, err := coredata.LoadDailyPatternHitsByTrackerPatternIDs(
				ctx,
				conn,
				scope,
				trackerPatternIDs,
				since,
			)
			if err != nil {
				return err
			}

			rows = loaded

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	rowsByPattern := make(map[gid.GID][]coredata.DailyPatternHit, len(trackerPatternIDs))
	for _, row := range rows {
		rowsByPattern[row.TrackerPatternID] = append(rowsByPattern[row.TrackerPatternID], row)
	}

	for trackerPatternID, patternRows := range rowsByPattern {
		result[trackerPatternID] = dailyHitsSeries(patternRows, since, days)
	}

	return result, nil
}
