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
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewCookieBannerCapabilities(c coredata.CookieBannerCapabilities) *CookieBannerCapabilities {
	return &CookieBannerCapabilities{
		ResourceReporting: c.ResourceReporting,
		Tcf:               c.TCF,
		Corsless:          c.Corsless,
	}
}

func NewCookieBanner(b *coredata.CookieBanner, tcfCmpID int) *CookieBanner {
	return &CookieBanner{
		ID:                   b.ID,
		OrganizationID:       b.OrganizationID,
		Name:                 b.Name,
		Origin:               b.Origin,
		State:                CookieBannerState(b.State),
		PrivacyPolicyURL:     b.PrivacyPolicyURL,
		CookiePolicyURL:      b.CookiePolicyURL,
		ConsentExpiryDays:    b.ConsentExpiryDays,
		ShowBranding:         b.ShowBranding,
		ForceOptIn:           b.ForceOptIn,
		Capabilities:         NewCookieBannerCapabilities(b.Capabilities),
		DefaultLanguage:      b.DefaultLanguage,
		PublisherCountryCode: b.PublisherCountryCode,
		TcfCmpID:             tcfCmpID,
		CreatedAt:            b.CreatedAt,
		UpdatedAt:            b.UpdatedAt,
	}
}

func NewCookieBannerSample(s *coredata.CookieBannerSample) *CookieBannerSample {
	return &CookieBannerSample{
		ID:             s.ID,
		CookieBannerID: s.CookieBannerID,
		PageLoads:      []*SampleFamilyCount{},
		FrozenAt:       s.FrozenAt,
		CreatedAt:      s.CreatedAt,
	}
}

func (s *CookieBannerSample) SetPageLoads(loads []cookiebanner.SampleFamilyCount) {
	s.PageLoads = NewSampleFamilyCounts(loads)
}

func NewSampleFamilyCounts(counts []cookiebanner.SampleFamilyCount) []*SampleFamilyCount {
	out := make([]*SampleFamilyCount, 0, len(counts))
	for _, count := range counts {
		out = append(
			out,
			&SampleFamilyCount{
				Family: count.Family,
				Count:  count.Count,
			},
		)
	}

	return out
}

func NewListCookieBannerSamplesOutput(
	p *page.Page[*coredata.CookieBannerSample, coredata.CookieBannerSampleOrderField],
) ListCookieBannerSamplesOutput {
	samples := make([]*CookieBannerSample, 0, len(p.Data))
	for _, sample := range p.Data {
		samples = append(samples, NewCookieBannerSample(sample))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCookieBannerSamplesOutput{
		NextCursor:          nextCursor,
		CookieBannerSamples: samples,
	}
}

func NewListCookieBannersOutput(
	p *page.Page[*coredata.CookieBanner, coredata.CookieBannerOrderField],
	tcfCmpID int,
) ListCookieBannersOutput {
	banners := make([]*CookieBanner, 0, len(p.Data))
	for _, b := range p.Data {
		banners = append(banners, NewCookieBanner(b, tcfCmpID))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCookieBannersOutput{
		NextCursor:    nextCursor,
		CookieBanners: banners,
	}
}

func NewCookieBannerTranslation(t *coredata.CookieBannerTranslation) *CookieBannerTranslation {
	return &CookieBannerTranslation{
		ID:             t.ID,
		CookieBannerID: t.CookieBannerID,
		Language:       t.Language,
		Translations:   string(t.Translations),
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}
