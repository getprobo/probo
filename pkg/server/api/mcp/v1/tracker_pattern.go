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

package mcp_v1

import (
	"context"
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

func (r *Resolver) trackerPatternAttributions(
	ctx context.Context,
	patterns ...*coredata.TrackerPattern,
) (map[gid.GID]coredata.CommonTrackerPatternAttribution, error) {
	ids := make([]gid.GID, 0, len(patterns))
	seen := make(map[gid.GID]struct{}, len(patterns))
	for _, pattern := range patterns {
		if pattern == nil || pattern.CommonTrackerPatternID == nil {
			continue
		}

		id := *pattern.CommonTrackerPatternID
		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil, nil
	}

	commons, err := r.cookieBanner.GetCommonTrackerPatternsByIDs(ctx, ids...)
	if err != nil {
		return nil, fmt.Errorf("cannot load common tracker patterns: %w", err)
	}

	return types.AttributionByOrgPatternID(patterns, commons), nil
}

func (r *Resolver) newTrackerPattern(
	ctx context.Context,
	pattern *coredata.TrackerPattern,
) (*types.TrackerPattern, error) {
	attributions, err := r.trackerPatternAttributions(ctx, pattern)
	if err != nil {
		return nil, err
	}

	var attribution *coredata.CommonTrackerPatternAttribution
	if a, ok := attributions[pattern.ID]; ok {
		attribution = &a
	}

	return types.NewTrackerPatternWithAttribution(pattern, attribution), nil
}
