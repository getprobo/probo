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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
)

func TestPageLoadsByFamily_SumsBuckets(t *testing.T) {
	t.Parallel()

	loads := PageLoadsByFamily([]coredata.DailyPageLoad{
		{BrowserFamily: coredata.BrowserFamilyChrome, Presentation: "", Count: 4},
		{
			BrowserFamily: coredata.BrowserFamilyChrome,
			Presentation:  coredata.CookieBannerPresentationOptIn,
			ConsentAction: coredata.CookieConsentActionRejectAll,
			Count:         1,
		},
		{
			BrowserFamily: coredata.BrowserFamilyFirefox,
			Presentation:  coredata.CookieBannerPresentationOptIn,
			ConsentAction: coredata.CookieConsentActionAcceptAll,
			Count:         2,
		},
	})

	assert.Equal(t, []BrowserFamilyCount{
		{Family: coredata.BrowserFamilyChrome, Count: 5},
		{Family: coredata.BrowserFamilyFirefox, Count: 2},
	}, loads)
}

func TestDailyPageLoadsSeries_FillsTheWindowOldestFirst(t *testing.T) {
	t.Parallel()

	since := time.Date(2026, time.March, 8, 0, 0, 0, 0, time.UTC)
	outside := since.AddDate(0, 0, -1)
	firstDay := time.Date(2026, time.March, 8, 23, 30, 0, 0, time.UTC)
	lastDay := since.AddDate(0, 0, 2)

	series := dailyPageLoadsSeries(
		[]coredata.DailyPageLoad{
			{Day: outside, BrowserFamily: coredata.BrowserFamilySafari, Count: 9},
			{Day: firstDay, BrowserFamily: coredata.BrowserFamilyChrome, Count: 2},
			{Day: firstDay, BrowserFamily: coredata.BrowserFamilyFirefox, Count: 1},
			{
				Day:           lastDay,
				BrowserFamily: coredata.BrowserFamilyChrome,
				Presentation:  coredata.CookieBannerPresentationOptIn,
				Count:         4,
			},
			{
				Day:           lastDay,
				BrowserFamily: coredata.BrowserFamilyChrome,
				Presentation:  coredata.CookieBannerPresentationOptOut,
				Count:         1,
			},
		},
		since,
		3,
	)

	assert.Equal(t, []DailyPageLoads{
		{
			Day: since,
			PageLoads: []BrowserFamilyCount{
				{Family: coredata.BrowserFamilyChrome, Count: 2},
				{Family: coredata.BrowserFamilyFirefox, Count: 1},
			},
		},
		{Day: since.AddDate(0, 0, 1), PageLoads: []BrowserFamilyCount{}},
		{
			Day:       lastDay,
			PageLoads: []BrowserFamilyCount{{Family: coredata.BrowserFamilyChrome, Count: 5}},
		},
	}, series)
}

func TestSplitPatternHits_SplitsGranted(t *testing.T) {
	t.Parallel()

	total, allowed, denied := SplitPatternHits([]coredata.DailyPatternHit{
		{BrowserFamily: coredata.BrowserFamilyChrome, CategoryGranted: true, Count: 3},
		{BrowserFamily: coredata.BrowserFamilyChrome, CategoryGranted: false, Count: 1},
		{BrowserFamily: coredata.BrowserFamilySafari, CategoryGranted: false, Count: 2},
	})

	assert.Equal(t, []BrowserFamilyCount{
		{Family: coredata.BrowserFamilyChrome, Count: 4},
		{Family: coredata.BrowserFamilySafari, Count: 2},
	}, total)
	assert.Equal(t, []BrowserFamilyCount{
		{Family: coredata.BrowserFamilyChrome, Count: 3},
	}, allowed)
	assert.Equal(t, []BrowserFamilyCount{
		{Family: coredata.BrowserFamilyChrome, Count: 1},
		{Family: coredata.BrowserFamilySafari, Count: 2},
	}, denied)
}
