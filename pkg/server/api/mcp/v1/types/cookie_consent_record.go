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

func NewCookieConsentRecord(r *coredata.CookieConsentRecord) *CookieConsentRecord {
	var consentData string
	if r.ConsentData != nil {
		consentData = string(r.ConsentData)
	}

	return &CookieConsentRecord{
		ID:                    r.ID,
		CookieBannerID:        r.CookieBannerID,
		CookieBannerVersionID: r.CookieBannerVersionID,
		VisitorID:             r.VisitorID,
		IPAddress:             r.IPAddress,
		UserAgent:             r.UserAgent,
		ConsentData:           consentData,
		Action:                CookieConsentRecordAction(r.Action),
		SdkVersion:            r.SdkVersion,
		Regulation:            r.Regulation,
		RegulationSource:      r.RegulationSource,
		CountryCode:           r.CountryCode,
		CreatedAt:             r.CreatedAt,
	}
}

func NewListCookieConsentRecordsOutput(p *page.Page[*coredata.CookieConsentRecord, coredata.CookieConsentRecordOrderField]) ListCookieConsentRecordsOutput {
	records := make([]*CookieConsentRecord, 0, len(p.Data))
	for _, r := range p.Data {
		records = append(records, NewCookieConsentRecord(r))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCookieConsentRecordsOutput{
		NextCursor:           nextCursor,
		CookieConsentRecords: records,
	}
}
