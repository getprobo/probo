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

package coredata

import (
	"github.com/jackc/pgx/v5"
	"go.probo.inc/probo/pkg/gid"
)

type AccessInventoryGrantFilter struct {
	AccessReviewSourceID *gid.GID
	PrincipalID          *gid.GID
	ResourceID           *gid.GID
	Presence             *AccessInventoryGrantPresence
	IncludeDisappeared   *bool

	// hideAfterMissingSyncs is set by the service from org settings (or the
	// product default). It is not part of the GraphQL filter input.
	hideAfterMissingSyncs int
}

// WithHideAfterMissingSyncs returns a copy with the hide-after-X threshold used
// when IncludeDisappeared is absent or false.
func (f *AccessInventoryGrantFilter) WithHideAfterMissingSyncs(hideAfter int) *AccessInventoryGrantFilter {
	if f == nil {
		return &AccessInventoryGrantFilter{hideAfterMissingSyncs: hideAfter}
	}

	clone := *f
	clone.hideAfterMissingSyncs = hideAfter

	return &clone
}

func (f *AccessInventoryGrantFilter) SQLFragment() string {
	if f == nil {
		return `(
	presence = @filter_presence_present::access_inventory_grant_presence
	OR consecutive_missing_syncs < @filter_hide_after_missing_syncs
)`
	}

	return `
(
	CASE
		WHEN @filter_access_review_source_id::text IS NOT NULL THEN
			access_review_source_id = @filter_access_review_source_id
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_principal_id::text IS NOT NULL THEN
			principal_id = @filter_principal_id
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_resource_id::text IS NOT NULL THEN
			resource_id = @filter_resource_id
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_presence::access_inventory_grant_presence IS NOT NULL THEN
			presence = @filter_presence::access_inventory_grant_presence
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_include_disappeared::boolean IS TRUE THEN
			TRUE
		ELSE
			presence = @filter_presence_present::access_inventory_grant_presence
			OR consecutive_missing_syncs < @filter_hide_after_missing_syncs
	END
)`
}

func (f *AccessInventoryGrantFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"filter_access_review_source_id":  nil,
		"filter_principal_id":             nil,
		"filter_resource_id":              nil,
		"filter_presence":                 nil,
		"filter_include_disappeared":      nil,
		"filter_presence_present":         string(AccessInventoryGrantPresencePresent),
		"filter_hide_after_missing_syncs": DefaultHideAfterMissingSyncs,
	}

	if f == nil {
		return args
	}

	if f.AccessReviewSourceID != nil {
		args["filter_access_review_source_id"] = *f.AccessReviewSourceID
	}

	if f.PrincipalID != nil {
		args["filter_principal_id"] = *f.PrincipalID
	}

	if f.ResourceID != nil {
		args["filter_resource_id"] = *f.ResourceID
	}

	if f.Presence != nil {
		args["filter_presence"] = string(*f.Presence)
	}

	if f.IncludeDisappeared != nil {
		args["filter_include_disappeared"] = *f.IncludeDisappeared
	}

	if f.hideAfterMissingSyncs > 0 {
		args["filter_hide_after_missing_syncs"] = f.hideAfterMissingSyncs
	}

	return args
}
