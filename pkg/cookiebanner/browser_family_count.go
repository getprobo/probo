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
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

type BrowserFamilyCount struct {
	Family coredata.BrowserFamily
	Count  int
}

// DailyPageLoads is one UTC day of page loads. PageLoads is empty when the
// day has none, so a series can stay dense without the caller filling gaps.
type DailyPageLoads struct {
	Day       time.Time
	PageLoads []BrowserFamilyCount
}

func NewBrowserFamilyCountPtrs(counts []BrowserFamilyCount) []*BrowserFamilyCount {
	out := make([]*BrowserFamilyCount, len(counts))
	for i := range counts {
		out[i] = &counts[i]
	}

	return out
}

func NewDailyPageLoadsPtrs(series []DailyPageLoads) []*DailyPageLoads {
	out := make([]*DailyPageLoads, len(series))
	for i := range series {
		out[i] = &series[i]
	}

	return out
}

// dailyPageLoadsSeries returns exactly days entries, oldest first, starting
// at since. since must be a UTC midnight. Rows on other days are dropped,
// and days with no rows keep an empty family list.
func dailyPageLoadsSeries(rows []coredata.DailyPageLoad, since time.Time, days int) []DailyPageLoads {
	rowsByDay := map[time.Time][]coredata.DailyPageLoad{}
	for _, row := range rows {
		day := utcDay(row.Day)
		rowsByDay[day] = append(rowsByDay[day], row)
	}

	series := make([]DailyPageLoads, days)
	for i := range days {
		day := since.AddDate(0, 0, i)
		series[i] = DailyPageLoads{
			Day:       day,
			PageLoads: PageLoadsByFamily(rowsByDay[day]),
		}
	}

	return series
}

func PageLoadsByFamily(rows []coredata.DailyPageLoad) []BrowserFamilyCount {
	totals := map[coredata.BrowserFamily]int{}

	for _, row := range rows {
		totals[row.BrowserFamily] += row.Count
	}

	return countsByFamily(totals)
}

func SplitPatternHits(rows []coredata.DailyPatternHit) (total, allowed, denied []BrowserFamilyCount) {
	totals := map[coredata.BrowserFamily]int{}
	alloweds := map[coredata.BrowserFamily]int{}
	denieds := map[coredata.BrowserFamily]int{}

	for _, row := range rows {
		totals[row.BrowserFamily] += row.Count
		if row.CategoryGranted {
			alloweds[row.BrowserFamily] += row.Count
		} else {
			denieds[row.BrowserFamily] += row.Count
		}
	}

	return countsByFamily(totals), countsByFamily(alloweds), countsByFamily(denieds)
}

func countsByFamily(sums map[coredata.BrowserFamily]int) []BrowserFamilyCount {
	out := make([]BrowserFamilyCount, 0, len(sums))
	for _, family := range coredata.BrowserFamilies() {
		count := sums[family]
		if count == 0 {
			continue
		}

		out = append(out, BrowserFamilyCount{Family: family, Count: count})
	}

	return out
}
