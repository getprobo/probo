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

type HitSplitQuery struct {
	PatternID gid.GID
	Days      int
}

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

func validateAnalyticsDays(days int) error {
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
	if err := validateAnalyticsDays(days); err != nil {
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

func (s *Service) GetHitSplits(
	ctx context.Context,
	scope coredata.Scoper,
	trackerPatternID gid.GID,
	days int,
) (HitSplits, error) {
	query := HitSplitQuery{
		PatternID: trackerPatternID,
		Days:      days,
	}

	splits, err := s.GetHitSplitsForPatterns(ctx, scope, []HitSplitQuery{query})
	if err != nil {
		return HitSplits{}, err
	}

	return splits[query], nil
}

func (s *Service) GetHitSplitsForPatterns(
	ctx context.Context,
	scope coredata.Scoper,
	queries []HitSplitQuery,
) (map[HitSplitQuery]HitSplits, error) {
	result := make(map[HitSplitQuery]HitSplits, len(queries))
	patternIDsByDays := make(map[int][]gid.GID)

	for _, query := range queries {
		if err := validateAnalyticsDays(query.Days); err != nil {
			return nil, fmt.Errorf("invalid request: %w", err)
		}

		if _, seen := result[query]; seen {
			continue
		}

		result[query] = HitSplits{}
		patternIDsByDays[query.Days] = append(patternIDsByDays[query.Days], query.PatternID)
	}

	if len(patternIDsByDays) == 0 {
		return result, nil
	}

	now := s.now()

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			for days, patternIDs := range patternIDsByDays {
				hits, err := coredata.LoadDailyPatternHitsByTrackerPatternIDs(
					ctx,
					conn,
					scope,
					patternIDs,
					analyticsSince(now, days),
				)
				if err != nil {
					return err
				}

				hitsByPattern := make(map[gid.GID][]coredata.DailyPatternHit, len(patternIDs))
				for _, hit := range hits {
					hitsByPattern[hit.TrackerPatternID] = append(hitsByPattern[hit.TrackerPatternID], hit)
				}

				for patternID, patternHits := range hitsByPattern {
					total, allowed, denied := SplitPatternHits(patternHits)
					result[HitSplitQuery{PatternID: patternID, Days: days}] = HitSplits{
						Total:   total,
						Allowed: allowed,
						Denied:  denied,
					}
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}
