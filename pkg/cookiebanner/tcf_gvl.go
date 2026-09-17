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
	"encoding/json"
	"strconv"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	// Placeholder until IAB Europe issues Probo a CMP ID. The official IAB
	// libraries reject ids below 2, so 0 cannot be encoded or passed to CmpApi.
	tcfCmpID                = 2
	tcfCmpVersion           = 1
	tcfPublisherCC          = "AA"
	tcfGVLSpecVersion       = 3
	tcfDefaultPolicyVersion = 5
)

type gvlPayload struct {
	GVLSpecificationVersion int                        `json:"gvlSpecificationVersion"`
	VendorListVersion       int                        `json:"vendorListVersion"`
	TCFPolicyVersion        int                        `json:"tcfPolicyVersion"`
	LastUpdated             string                     `json:"lastUpdated"`
	Purposes                json.RawMessage            `json:"purposes"`
	SpecialPurposes         json.RawMessage            `json:"specialPurposes"`
	Features                json.RawMessage            `json:"features"`
	SpecialFeatures         json.RawMessage            `json:"specialFeatures"`
	Stacks                  json.RawMessage            `json:"stacks"`
	DataCategories          json.RawMessage            `json:"dataCategories"`
	Vendors                 map[string]json.RawMessage `json:"vendors"`
}

type gvlVendorJSON struct {
	ID                  int     `json:"id"`
	Name                string  `json:"name"`
	Purposes            []int32 `json:"purposes"`
	LegIntPurposes      []int32 `json:"legIntPurposes"`
	FlexiblePurposes    []int32 `json:"flexiblePurposes"`
	SpecialPurposes     []int32 `json:"specialPurposes"`
	Features            []int32 `json:"features"`
	SpecialFeatures     []int32 `json:"specialFeatures"`
	PolicyURL           *string `json:"policyUrl,omitempty"`
	UsesCookies         *bool   `json:"usesCookies,omitempty"`
	CookieRefresh       *bool   `json:"cookieRefresh,omitempty"`
	UsesNonCookieAccess *bool   `json:"usesNonCookieAccess,omitempty"`
	CookieMaxAgeSeconds *int    `json:"cookieMaxAgeSeconds,omitempty"`
}

func buildTCFGVL(
	snapshot *coredata.CommonGVLSnapshot,
	vendors coredata.CommonGVLVendors,
	iabVendorIDs []int,
) *BannerTCFGVL {
	selected := make(map[string]struct{}, len(iabVendorIDs))
	for _, id := range iabVendorIDs {
		selected[strconv.Itoa(id)] = struct{}{}
	}

	emptyObject := json.RawMessage(`{}`)
	gvl := &BannerTCFGVL{
		GVLSpecificationVersion: tcfGVLSpecVersion,
		TCFPolicyVersion:        tcfDefaultPolicyVersion,
		Purposes:                emptyObject,
		SpecialPurposes:         emptyObject,
		Features:                emptyObject,
		SpecialFeatures:         emptyObject,
		Stacks:                  emptyObject,
		DataCategories:          emptyObject,
		Vendors:                 make(map[string]json.RawMessage, len(iabVendorIDs)),
	}

	if snapshot != nil {
		gvl.GVLSpecificationVersion = snapshot.GVLSpecificationVersion
		gvl.VendorListVersion = snapshot.VendorListVersion
		gvl.TCFPolicyVersion = snapshot.TCFPolicyVersion
		gvl.LastUpdated = snapshot.LastUpdated.UTC().Format("2006-01-02T15:04:05Z")

		var payload gvlPayload
		if len(snapshot.Payload) > 0 && string(snapshot.Payload) != "{}" && string(snapshot.Payload) != "null" {
			if err := json.Unmarshal(snapshot.Payload, &payload); err == nil {
				if raw := omitEmptyJSON(payload.Purposes); raw != nil {
					gvl.Purposes = raw
				}
				if raw := omitEmptyJSON(payload.SpecialPurposes); raw != nil {
					gvl.SpecialPurposes = raw
				}
				if raw := omitEmptyJSON(payload.Features); raw != nil {
					gvl.Features = raw
				}
				if raw := omitEmptyJSON(payload.SpecialFeatures); raw != nil {
					gvl.SpecialFeatures = raw
				}
				if raw := omitEmptyJSON(payload.Stacks); raw != nil {
					gvl.Stacks = raw
				}
				if raw := omitEmptyJSON(payload.DataCategories); raw != nil {
					gvl.DataCategories = raw
				}
				if payload.LastUpdated != "" {
					gvl.LastUpdated = payload.LastUpdated
				}
				if payload.GVLSpecificationVersion != 0 {
					gvl.GVLSpecificationVersion = payload.GVLSpecificationVersion
				}
				if payload.VendorListVersion != 0 {
					gvl.VendorListVersion = payload.VendorListVersion
				}
				if payload.TCFPolicyVersion != 0 {
					gvl.TCFPolicyVersion = payload.TCFPolicyVersion
				}

				for id, raw := range payload.Vendors {
					if _, ok := selected[id]; ok {
						gvl.Vendors[id] = raw
					}
				}
			}
		}
	}

	byID := make(map[string]*coredata.CommonGVLVendor, len(vendors))
	for _, vendor := range vendors {
		byID[strconv.Itoa(vendor.IABVendorID)] = vendor
	}

	for _, id := range iabVendorIDs {
		key := strconv.Itoa(id)
		if _, ok := gvl.Vendors[key]; ok {
			continue
		}

		vendor, ok := byID[key]
		if !ok {
			continue
		}

		raw, err := json.Marshal(synthesizeGVLVendor(vendor))
		if err != nil {
			continue
		}

		gvl.Vendors[key] = raw
	}

	return gvl
}

func synthesizeGVLVendor(vendor *coredata.CommonGVLVendor) gvlVendorJSON {
	return gvlVendorJSON{
		ID:                  vendor.IABVendorID,
		Name:                vendor.Name,
		Purposes:            emptyInt32s(vendor.Purposes),
		LegIntPurposes:      emptyInt32s(vendor.LegIntPurposes),
		FlexiblePurposes:    emptyInt32s(vendor.FlexiblePurposes),
		SpecialPurposes:     emptyInt32s(vendor.SpecialPurposes),
		Features:            emptyInt32s(vendor.Features),
		SpecialFeatures:     emptyInt32s(vendor.SpecialFeatures),
		PolicyURL:           vendor.PolicyURL,
		UsesCookies:         vendor.UsesCookies,
		CookieRefresh:       vendor.CookieRefresh,
		UsesNonCookieAccess: vendor.UsesNonCookieAccess,
		CookieMaxAgeSeconds: vendor.CookieMaxAgeSeconds,
	}
}

func emptyInt32s(v []int32) []int32 {
	if v == nil {
		return []int32{}
	}

	return v
}

func omitEmptyJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}

	return raw
}
