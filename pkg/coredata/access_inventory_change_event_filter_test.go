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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestAccessInventoryChangeEventFilter_SQLArguments(t *testing.T) {
	t.Parallel()

	sourceID := gid.New(gid.NewTenantID(), coredata.AccessReviewSourceEntityType)
	after := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	before := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)

	filter := &coredata.AccessInventoryChangeEventFilter{
		AccessReviewSourceID: &sourceID,
		ChangeTypes: []coredata.AccessInventoryChangeType{
			coredata.AccessInventoryChangeTypeAppeared,
			coredata.AccessInventoryChangeTypeDisappeared,
		},
		ObservedAtOnOrAfter:  &after,
		ObservedAtOnOrBefore: &before,
	}

	args := filter.SQLArguments()
	require.Equal(t, sourceID, args["filter_access_review_source_id"])
	require.Equal(t, after, args["filter_observed_at_on_or_after"])
	require.Equal(t, before, args["filter_observed_at_on_or_before"])

	changeTypes, ok := args["filter_change_types"].(coredata.AccessInventoryChangeTypeValues)
	require.True(t, ok)
	assert.Equal(
		t,
		coredata.AccessInventoryChangeTypeValues{
			coredata.AccessInventoryChangeTypeAppeared,
			coredata.AccessInventoryChangeTypeDisappeared,
		},
		changeTypes,
	)

	fragment := compactSQL(filter.SQLFragment())
	assert.Contains(
		t,
		fragment,
		"WHEN @filter_access_review_source_id::text IS NOT NULL THEN access_review_source_id = @filter_access_review_source_id ELSE TRUE END",
	)
	assert.Contains(
		t,
		fragment,
		"WHEN @filter_change_types::access_inventory_change_type[] IS NOT NULL THEN change_type = ANY(@filter_change_types::access_inventory_change_type[]) ELSE TRUE END",
	)
	assert.Contains(
		t,
		fragment,
		"WHEN @filter_observed_at_on_or_after::timestamptz IS NOT NULL THEN observed_at >= @filter_observed_at_on_or_after ELSE TRUE END",
	)
	assert.Contains(
		t,
		fragment,
		"WHEN @filter_observed_at_on_or_before::timestamptz IS NOT NULL THEN observed_at <= @filter_observed_at_on_or_before ELSE TRUE END",
	)
}

func TestAccessInventoryChangeEventFilter_NilDefaults(t *testing.T) {
	t.Parallel()

	args := (*coredata.AccessInventoryChangeEventFilter)(nil).SQLArguments()
	assert.Nil(t, args["filter_access_review_source_id"])
	assert.Nil(t, args["filter_change_types"])
	assert.Nil(t, args["filter_observed_at_on_or_after"])
	assert.Nil(t, args["filter_observed_at_on_or_before"])
}
