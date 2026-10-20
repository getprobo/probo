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

type DiscoveryFamilyCount struct {
	Family coredata.DiscoveryBrowserFamily
	Count  int
}

func discoveryPageLoads(stats coredata.CookieBannerDiscoveryStats) []DiscoveryFamilyCount {
	return discoveryFamilyCounts(
		stats,
		stats.ChromePageLoads,
		stats.EdgePageLoads,
		stats.FirefoxPageLoads,
		stats.SafariPageLoads,
		stats.OtherPageLoads,
	)
}

func discoveryHits(
	stats coredata.CookieBannerDiscoveryStats,
	hits *coredata.TrackerPatternDiscoveryHits,
) []DiscoveryFamilyCount {
	if hits == nil {
		return discoveryFamilyCounts(stats, 0, 0, 0, 0, 0)
	}

	return discoveryFamilyCounts(
		stats,
		hits.ChromeHits,
		hits.EdgeHits,
		hits.FirefoxHits,
		hits.SafariHits,
		hits.OtherHits,
	)
}

func discoveryFamilyCounts(
	stats coredata.CookieBannerDiscoveryStats,
	chrome int,
	edge int,
	firefox int,
	safari int,
	other int,
) []DiscoveryFamilyCount {
	counts := [5]int{chrome, edge, firefox, safari, other}
	loads := [5]int{
		stats.ChromePageLoads,
		stats.EdgePageLoads,
		stats.FirefoxPageLoads,
		stats.SafariPageLoads,
		stats.OtherPageLoads,
	}

	out := make([]DiscoveryFamilyCount, 0, len(coredata.DiscoveryBrowserFamilies()))
	for i, family := range coredata.DiscoveryBrowserFamilies() {
		if loads[i] == 0 {
			continue
		}

		out = append(out, DiscoveryFamilyCount{Family: family, Count: counts[i]})
	}

	return out
}
