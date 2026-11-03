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
	"encoding/json"

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

func NewCookieBannerScan(scan *coredata.CookieBannerScan) *CookieBannerScan {
	return &CookieBannerScan{
		ID:             scan.ID,
		CookieBannerID: scan.CookieBannerID,
		State:          scan.State,
		FailureReason:  scan.FailureReason,
		Truncated:      scan.Truncated,
		PagesScanned:   scan.PagesScanned,
		CookiesFound:   scan.CookiesFound,
		StorageFound:   scan.StorageFound,
		ResourcesFound: scan.ResourcesFound,
		StartedAt:      scan.StartedAt,
		CompletedAt:    scan.CompletedAt,
		CreatedAt:      scan.CreatedAt,
		UpdatedAt:      scan.UpdatedAt,
		Checkpoint:     decodeCookieBannerScanCheckpoint(scan.Checkpoint),
	}
}

func NewListCookieBannerScansOutput(
	p *page.Page[*coredata.CookieBannerScan, coredata.CookieBannerScanOrderField],
) ListCookieBannerScansOutput {
	scans := make([]*CookieBannerScan, 0, len(p.Data))
	for _, scan := range p.Data {
		scans = append(scans, NewCookieBannerScan(scan))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCookieBannerScansOutput{
		NextCursor:        nextCursor,
		CookieBannerScans: scans,
	}
}

func decodeCookieBannerScanCheckpoint(raw *json.RawMessage) *CookieBannerScanCheckpoint {
	decoded, err := cookiebanner.UnmarshalScanCheckpoint(raw)
	if err != nil || decoded == nil {
		return nil
	}

	queue := make([]*CookieBannerScanCheckpointQueueItem, 0, len(decoded.Queue))
	for _, item := range decoded.Queue {
		if item.URL == "" {
			continue
		}

		queue = append(queue, &CookieBannerScanCheckpointQueueItem{URL: item.URL})
	}

	visited := decoded.VisitedTemplates
	if visited == nil {
		visited = []string{}
	}

	hosts := decoded.CoveredHosts
	if hosts == nil {
		hosts = []string{}
	}

	return &CookieBannerScanCheckpoint{
		VisitedTemplates: visited,
		Queue:            queue,
		CoveredHosts:     hosts,
		PagesScanned:     decoded.PagesScanned,
		ElapsedMs:        int(decoded.ElapsedMS),
		Truncated:        decoded.Truncated,
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
