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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestBuildTCFGVL_FiltersPayloadVendors(t *testing.T) {
	t.Parallel()

	payload := []byte(`{
		"gvlSpecificationVersion": 3,
		"vendorListVersion": 42,
		"tcfPolicyVersion": 5,
		"lastUpdated": "2026-01-15T17:00:00Z",
		"purposes": {"1": {"id": 1, "name": "Store and/or access information on a device"}},
		"vendors": {
			"52": {"id": 52, "name": "On banner"},
			"755": {"id": 755, "name": "Not selected"}
		}
	}`)

	snapshot := &coredata.CommonGVLSnapshot{
		VendorListVersion:       42,
		GVLSpecificationVersion: 3,
		TCFPolicyVersion:        5,
		LastUpdated:             time.Date(2026, 1, 15, 17, 0, 0, 0, time.UTC),
		Payload:                 payload,
	}

	gvl := buildTCFGVL(snapshot, nil, []int{52})
	require.NotNil(t, gvl)
	assert.Equal(t, 42, gvl.VendorListVersion)
	assert.Equal(t, 5, gvl.TCFPolicyVersion)
	assert.Equal(t, "2026-01-15T17:00:00Z", gvl.LastUpdated)
	require.Contains(t, gvl.Vendors, "52")
	assert.NotContains(t, gvl.Vendors, "755")
	assert.JSONEq(t, `{"id":1,"name":"Store and/or access information on a device"}`, string(mustRawMapValue(t, gvl.Purposes, "1")))
}

func TestBuildTCFGVL_SynthesizesFromVendorRowsWhenPayloadEmpty(t *testing.T) {
	t.Parallel()

	policyURL := "https://example.com/privacy"
	tenant := gid.NewTenantID()
	vendor := &coredata.CommonGVLVendor{
		ID:              gid.New(tenant, coredata.CommonGVLVendorEntityType),
		IABVendorID:     52,
		Name:            "Synthesized Vendor",
		Purposes:        []int32{1, 2},
		LegIntPurposes:  []int32{7},
		SpecialFeatures: []int32{1},
		PolicyURL:       &policyURL,
	}

	snapshot := &coredata.CommonGVLSnapshot{
		VendorListVersion:       9,
		GVLSpecificationVersion: 3,
		TCFPolicyVersion:        5,
		LastUpdated:             time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		Payload:                 json.RawMessage(`{}`),
	}

	gvl := buildTCFGVL(snapshot, coredata.CommonGVLVendors{vendor}, []int{52})
	require.NotNil(t, gvl)
	assert.Equal(t, 9, gvl.VendorListVersion)
	require.Contains(t, gvl.Vendors, "52")

	var got gvlVendorJSON
	require.NoError(t, json.Unmarshal(gvl.Vendors["52"], &got))
	assert.Equal(t, 52, got.ID)
	assert.Equal(t, "Synthesized Vendor", got.Name)
	assert.Equal(t, []int32{1, 2}, got.Purposes)
	assert.Equal(t, []int32{7}, got.LegIntPurposes)
	assert.Equal(t, []int32{1}, got.SpecialFeatures)
	require.NotNil(t, got.PolicyURL)
	assert.Equal(t, policyURL, *got.PolicyURL)
}

func TestBuildTCFGVL_EmptySelection(t *testing.T) {
	t.Parallel()

	gvl := buildTCFGVL(nil, nil, nil)
	require.NotNil(t, gvl)
	assert.Empty(t, gvl.Vendors)
	assert.Equal(t, tcfDefaultPolicyVersion, gvl.TCFPolicyVersion)
	assert.Equal(t, tcfGVLSpecVersion, gvl.GVLSpecificationVersion)
	assert.JSONEq(t, `{}`, string(gvl.Purposes))
	assert.JSONEq(t, `{}`, string(gvl.SpecialPurposes))
}

func mustRawMapValue(t *testing.T, raw json.RawMessage, key string) json.RawMessage {
	t.Helper()

	var obj map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &obj))
	value, ok := obj[key]
	require.True(t, ok)

	return value
}
