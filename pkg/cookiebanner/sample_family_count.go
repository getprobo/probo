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

func SamplePageLoads(sample coredata.CookieBannerSample) []SampleFamilyCount {
	return samplePageLoads(sample)
}

func SampleExpectedPageLoads(sample coredata.CookieBannerSample) []SampleFamilyCount {
	return sampleExpectedPageLoads(sample)
}

func NewSampleFamilyCountPtrs(counts []SampleFamilyCount) []*SampleFamilyCount {
	out := make([]*SampleFamilyCount, len(counts))
	for i := range counts {
		out[i] = &counts[i]
	}

	return out
}

func samplePageLoads(sample coredata.CookieBannerSample) []SampleFamilyCount {
	return sampleFamilyCounts(
		sample.ChromePageLoads,
		sample.EdgePageLoads,
		sample.FirefoxPageLoads,
		sample.SafariPageLoads,
		sample.OtherPageLoads,
		sample.ChromePageLoads,
		sample.EdgePageLoads,
		sample.FirefoxPageLoads,
		sample.SafariPageLoads,
		sample.OtherPageLoads,
	)
}

func sampleExpectedPageLoads(sample coredata.CookieBannerSample) []SampleFamilyCount {
	return sampleFamilyCounts(
		sample.ChromeConsentingPageLoads,
		sample.EdgeConsentingPageLoads,
		sample.FirefoxConsentingPageLoads,
		sample.SafariConsentingPageLoads,
		sample.OtherConsentingPageLoads,
		sample.ChromeConsentingPageLoads,
		sample.EdgeConsentingPageLoads,
		sample.FirefoxConsentingPageLoads,
		sample.SafariConsentingPageLoads,
		sample.OtherConsentingPageLoads,
	)
}

func sampleHits(
	sample coredata.CookieBannerSample,
	hits *coredata.TrackerPatternSampleHits,
) []SampleFamilyCount {
	if hits == nil {
		return sampleFamilyCounts(0, 0, 0, 0, 0,
			sample.ChromePageLoads,
			sample.EdgePageLoads,
			sample.FirefoxPageLoads,
			sample.SafariPageLoads,
			sample.OtherPageLoads,
		)
	}

	return sampleFamilyCounts(
		hits.ChromeHits,
		hits.EdgeHits,
		hits.FirefoxHits,
		hits.SafariHits,
		hits.OtherHits,
		sample.ChromePageLoads,
		sample.EdgePageLoads,
		sample.FirefoxPageLoads,
		sample.SafariPageLoads,
		sample.OtherPageLoads,
	)
}

func sampleFamilyCounts(
	chrome int,
	edge int,
	firefox int,
	safari int,
	other int,
	chromeDenom int,
	edgeDenom int,
	firefoxDenom int,
	safariDenom int,
	otherDenom int,
) []SampleFamilyCount {
	counts := [5]int{chrome, edge, firefox, safari, other}
	denoms := [5]int{chromeDenom, edgeDenom, firefoxDenom, safariDenom, otherDenom}

	out := make([]SampleFamilyCount, 0, len(coredata.BrowserFamilies()))
	for i, family := range coredata.BrowserFamilies() {
		if denoms[i] == 0 {
			continue
		}

		out = append(out, SampleFamilyCount{Family: family, Count: counts[i]})
	}

	return out
}
