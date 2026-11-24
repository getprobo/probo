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

package coredata_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestAccessInventoryGrantFilter_WithHideAfterMissingSyncs(t *testing.T) {
	t.Parallel()

	nilFilter := (*coredata.AccessInventoryGrantFilter)(nil).WithHideAfterMissingSyncs(5)
	require.NotNil(t, nilFilter)
	assert.Equal(t, 5, nilFilter.SQLArguments()["filter_hide_after_missing_syncs"])

	include := true
	sourceID := gid.New(gid.NewTenantID(), coredata.AccessReviewSourceEntityType)
	base := &coredata.AccessInventoryGrantFilter{
		AccessReviewSourceID: &sourceID,
		IncludeDisappeared:   &include,
	}
	filter := base.WithHideAfterMissingSyncs(7)

	args := filter.SQLArguments()
	assert.Equal(t, sourceID, args["filter_access_review_source_id"])
	assert.Equal(t, true, args["filter_include_disappeared"])
	assert.Equal(t, 7, args["filter_hide_after_missing_syncs"])
}

func TestAccessInventoryGrantFilter_NilSQLFragmentHidesPastThreshold(t *testing.T) {
	t.Parallel()

	var filter *coredata.AccessInventoryGrantFilter
	fragment := compactSQL(filter.SQLFragment())

	assert.Contains(t, fragment, "WHEN @filter_include_disappeared::boolean IS TRUE THEN TRUE")
	assert.Contains(
		t,
		fragment,
		"ELSE presence = @filter_presence_present::access_inventory_grant_presence OR consecutive_missing_syncs < @filter_hide_after_missing_syncs END",
	)
	assert.Equal(t, coredata.DefaultHideAfterMissingSyncs, filter.SQLArguments()["filter_hide_after_missing_syncs"])
	assert.Equal(
		t,
		string(coredata.AccessInventoryGrantPresencePresent),
		filter.SQLArguments()["filter_presence_present"],
	)
}

func compactSQL(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
