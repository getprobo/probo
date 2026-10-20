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

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
)

func TestSamplePageLoads_OmitsZeroFamilies(t *testing.T) {
	t.Parallel()

	sample := coredata.CookieBannerSample{
		ChromePageLoads:  10,
		FirefoxPageLoads: 3,
	}

	assert.Equal(
		t,
		[]SampleFamilyCount{
			{Family: coredata.BrowserFamilyChrome, Count: 10},
			{Family: coredata.BrowserFamilyFirefox, Count: 3},
		},
		samplePageLoads(sample),
	)
}

func TestSampleExpectedPageLoads_OmitsZeroFamilies(t *testing.T) {
	t.Parallel()

	sample := coredata.CookieBannerSample{
		ChromeConsentingPageLoads: 4,
		SafariConsentingPageLoads: 2,
		FirefoxPageLoads:          9,
	}

	assert.Equal(
		t,
		[]SampleFamilyCount{
			{Family: coredata.BrowserFamilyChrome, Count: 4},
			{Family: coredata.BrowserFamilySafari, Count: 2},
		},
		sampleExpectedPageLoads(sample),
	)
}

func TestSampleHits_UsesPageLoadFamilies(t *testing.T) {
	t.Parallel()

	sample := coredata.CookieBannerSample{
		ChromePageLoads:  10,
		FirefoxPageLoads: 3,
	}

	assert.Equal(
		t,
		[]SampleFamilyCount{
			{Family: coredata.BrowserFamilyChrome, Count: 0},
			{Family: coredata.BrowserFamilyFirefox, Count: 0},
		},
		sampleHits(sample, nil),
	)

	hits := &coredata.TrackerPatternSampleHits{
		ChromeHits:  4,
		SafariHits:  9,
		FirefoxHits: 1,
	}

	assert.Equal(
		t,
		[]SampleFamilyCount{
			{Family: coredata.BrowserFamilyChrome, Count: 4},
			{Family: coredata.BrowserFamilyFirefox, Count: 1},
		},
		sampleHits(sample, hits),
	)
}
