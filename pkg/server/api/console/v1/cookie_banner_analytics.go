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

package console_v1

import (
	"context"
	"errors"

	"github.com/vikstrous/dataloadgen"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/console/v1/dataloader"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlx"
	"go.probo.inc/probo/pkg/validator"
)

// analyticsDays is checked before the dataloader so an out-of-range value
// fails its own field instead of every key batched with it.
func analyticsDays(ctx context.Context, days *int) (int, error) {
	if days == nil {
		return cookiebanner.DefaultAnalyticsDays, nil
	}

	if err := cookiebanner.ValidateAnalyticsDays(*days); err != nil {
		if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
			return 0, gqlx.InvalidValidationErrors(ctx, validationErrors)
		}

		return 0, gqlx.Invalid(ctx, err)
	}

	return *days, nil
}

func (r *cookieBannerResolver) loadPageLoads(
	ctx context.Context,
	obj *types.CookieBanner,
	days *int,
) ([]*cookiebanner.BrowserFamilyCount, error) {
	if _, err := r.authorize(ctx, obj.ID, probo.ActionCookieBannerGet); err != nil {
		return nil, err
	}

	window, err := analyticsDays(ctx, days)
	if err != nil {
		return nil, err
	}

	key := dataloader.PageLoadsKey{
		CookieBannerID: obj.ID,
		Days:           window,
	}

	pageLoads, err := dataloader.FromContext(ctx).PageLoads.Load(ctx, key)
	if err != nil {
		if errors.Is(err, dataloadgen.ErrNotFound) {
			return []*cookiebanner.BrowserFamilyCount{}, nil
		}

		r.logger.ErrorCtx(ctx, "cannot get cookie banner page loads", log.Error(err))

		return nil, gqlx.Internal(ctx)
	}

	return cookiebanner.NewBrowserFamilyCountPtrs(pageLoads), nil
}

func (r *trackerPatternResolver) loadHitSplits(
	ctx context.Context,
	obj *types.TrackerPattern,
	days *int,
) (cookiebanner.HitSplits, error) {
	if _, err := r.authorize(ctx, obj.ID, probo.ActionTrackerPatternGet); err != nil {
		return cookiebanner.HitSplits{}, err
	}

	window, err := analyticsDays(ctx, days)
	if err != nil {
		return cookiebanner.HitSplits{}, err
	}

	key := dataloader.HitSplitsKey{
		PatternID: obj.ID,
		Days:      window,
	}

	splits, err := dataloader.FromContext(ctx).HitSplits.Load(ctx, key)
	if err != nil {
		if errors.Is(err, dataloadgen.ErrNotFound) {
			return cookiebanner.HitSplits{}, nil
		}

		r.logger.ErrorCtx(ctx, "cannot get tracker pattern hits", log.Error(err))

		return cookiebanner.HitSplits{}, gqlx.Internal(ctx)
	}

	return splits, nil
}
