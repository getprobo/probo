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
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/console/v1/dataloader"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

func (r *trackerPatternResolver) loadSampleHitSplits(
	ctx context.Context,
	obj *types.TrackerPattern,
	sampleID *gid.GID,
) (cookiebanner.SampleHitSplits, error) {
	if _, err := r.authorize(ctx, obj.ID, probo.ActionTrackerPatternGet); err != nil {
		return cookiebanner.SampleHitSplits{}, err
	}

	key := dataloader.SampleHitSplitKey{
		PatternID: obj.ID,
		Latest:    sampleID == nil,
	}
	if sampleID != nil {
		key.SampleID = *sampleID
	}

	splits, err := dataloader.FromContext(ctx).SampleHitSplits.Load(ctx, key)
	if err != nil {
		if errors.Is(err, cookiebanner.ErrSampleNotFound) || errors.Is(err, cookiebanner.ErrTrackerPatternNotFound) {
			return cookiebanner.SampleHitSplits{}, gqlutils.NotFound(ctx, err)
		}

		if errors.Is(err, dataloadgen.ErrNotFound) {
			return cookiebanner.SampleHitSplits{}, nil
		}

		r.logger.ErrorCtx(ctx, "cannot get tracker pattern sample hits", log.Error(err))

		return cookiebanner.SampleHitSplits{}, gqlutils.Internal(ctx)
	}

	return splits, nil
}
