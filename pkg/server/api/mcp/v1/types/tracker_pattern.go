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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func NewTrackerPatternWithAttribution(
	p *coredata.TrackerPattern,
	attribution *coredata.CommonTrackerPatternAttribution,
) *TrackerPattern {
	var source TrackerPatternSource
	if p.Source != nil {
		source = TrackerPatternSource(*p.Source)
	}

	var mapped *TrackerPatternAttribution

	if attribution != nil {
		value := TrackerPatternAttribution(*attribution)
		mapped = &value
	}

	return &TrackerPattern{
		ID:                     p.ID,
		OrganizationID:         p.OrganizationID,
		CookieBannerID:         p.CookieBannerID,
		CookieCategoryID:       p.CookieCategoryID,
		TrackerType:            TrackerPatternTrackerType(p.TrackerType),
		Pattern:                p.Pattern,
		MatchType:              TrackerPatternMatchType(p.MatchType),
		DisplayName:            p.DisplayName,
		MaxAgeSeconds:          p.MaxAgeSeconds,
		Description:            p.Description,
		Source:                 source,
		Excluded:               p.Excluded,
		LastMatchedAt:          p.LastMatchedAt,
		CommonTrackerPatternID: p.CommonTrackerPatternID,
		Attribution:            mapped,
		CreatedAt:              p.CreatedAt,
		UpdatedAt:              p.UpdatedAt,
	}
}

// AttributionByOrgPatternID maps each org pattern to the catalog verdict
// on its linked common row. Patterns with no catalog link are omitted.
func AttributionByOrgPatternID(
	patterns []*coredata.TrackerPattern,
	commons coredata.CommonTrackerPatterns,
) map[gid.GID]coredata.CommonTrackerPatternAttribution {
	byCommon := make(map[gid.GID]coredata.CommonTrackerPatternAttribution, len(commons))
	for _, common := range commons {
		byCommon[common.ID] = common.Attribution
	}

	out := make(map[gid.GID]coredata.CommonTrackerPatternAttribution)

	for _, pattern := range patterns {
		if pattern.CommonTrackerPatternID == nil {
			continue
		}

		attribution, ok := byCommon[*pattern.CommonTrackerPatternID]
		if !ok {
			continue
		}

		out[pattern.ID] = attribution
	}

	return out
}

func NewListTrackerPatternsOutput(
	pg *page.Page[*coredata.TrackerPattern, coredata.TrackerPatternOrderField],
	attributions map[gid.GID]coredata.CommonTrackerPatternAttribution,
) ListTrackerPatternsOutput {
	patterns := make([]*TrackerPattern, 0, len(pg.Data))
	for _, p := range pg.Data {
		var attribution *coredata.CommonTrackerPatternAttribution
		if a, ok := attributions[p.ID]; ok {
			attribution = &a
		}

		patterns = append(patterns, NewTrackerPatternWithAttribution(p, attribution))
	}

	var nextCursor *page.CursorKey

	if len(pg.Data) > 0 {
		cursorKey := pg.Data[len(pg.Data)-1].CursorKey(pg.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListTrackerPatternsOutput{
		NextCursor:      nextCursor,
		TrackerPatterns: patterns,
	}
}
