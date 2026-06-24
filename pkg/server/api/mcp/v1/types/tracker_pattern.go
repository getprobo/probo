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
	"go.probo.inc/probo/pkg/page"
)

func NewTrackerPattern(p *coredata.TrackerPattern) *TrackerPattern {
	var source TrackerPatternSource
	if p.Source != nil {
		source = TrackerPatternSource(*p.Source)
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
		CreatedAt:              p.CreatedAt,
		UpdatedAt:              p.UpdatedAt,
	}
}

func NewListTrackerPatternsOutput(pg *page.Page[*coredata.TrackerPattern, coredata.TrackerPatternOrderField]) ListTrackerPatternsOutput {
	patterns := make([]*TrackerPattern, 0, len(pg.Data))
	for _, p := range pg.Data {
		patterns = append(patterns, NewTrackerPattern(p))
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
