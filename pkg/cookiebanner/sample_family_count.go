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

import "go.probo.inc/probo/pkg/coredata"

type SampleFamilyCount struct {
	Family coredata.BrowserFamily
	Count  int
}

func NewSampleFamilyCountPtrs(counts []SampleFamilyCount) []*SampleFamilyCount {
	out := make([]*SampleFamilyCount, len(counts))
	for i := range counts {
		out[i] = &counts[i]
	}

	return out
}

func PageLoadsByFamily(rows []coredata.SamplePageLoad) []SampleFamilyCount {
	totals := map[coredata.BrowserFamily]int{}

	for _, row := range rows {
		totals[row.BrowserFamily] += row.Count
	}

	return countsByFamily(totals)
}

func SplitPatternHits(rows []coredata.SamplePatternHit) (total, allowed, denied []SampleFamilyCount) {
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

func consentActionPtr(action string) *coredata.CookieConsentAction {
	if action == "" {
		return nil
	}

	value := coredata.CookieConsentAction(action)

	return &value
}

func countsByFamily(sums map[coredata.BrowserFamily]int) []SampleFamilyCount {
	out := make([]SampleFamilyCount, 0, len(sums))
	for _, family := range coredata.BrowserFamilies() {
		count := sums[family]
		if count == 0 {
			continue
		}

		out = append(out, SampleFamilyCount{Family: family, Count: count})
	}

	return out
}
