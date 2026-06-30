// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package coredata

import (
	"github.com/jackc/pgx/v5"
)

type AccessReviewEntryFilter struct {
	Decision       *AccessReviewEntryDecision
	Flag           *AccessReviewEntryFlag
	IncrementalTag *AccessReviewEntryIncrementalTag
	IsAdmin        *bool
	Active         *bool
	AuthMethod     *AccessReviewEntryAuthMethod
	AccountType    *AccessReviewEntryAccountType
}

func (f *AccessReviewEntryFilter) SQLFragment() string {
	if f == nil {
		return "TRUE"
	}

	return `
(
	CASE
		WHEN @filter_decision::access_review_entry_decision IS NOT NULL THEN
			decision = @filter_decision::access_review_entry_decision
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_flag::text IS NOT NULL THEN
			@filter_flag::text = ANY(flags)
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_incremental_tag::access_review_entry_incremental_tag IS NOT NULL THEN
			incremental_tag = @filter_incremental_tag::access_review_entry_incremental_tag
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_is_admin::boolean IS NOT NULL THEN
			is_admin = @filter_is_admin::boolean
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_active::boolean IS NOT NULL THEN
			active = @filter_active::boolean
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_auth_method::auth_method IS NOT NULL THEN
			auth_method = @filter_auth_method::auth_method
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_account_type::text IS NOT NULL THEN
			account_type = @filter_account_type::text
		ELSE TRUE
	END
)`
}

func (f *AccessReviewEntryFilter) SQLArguments() pgx.StrictNamedArgs {
	if f == nil {
		return pgx.StrictNamedArgs{}
	}

	args := pgx.StrictNamedArgs{
		"filter_decision":        nil,
		"filter_flag":            nil,
		"filter_incremental_tag": nil,
		"filter_is_admin":        nil,
		"filter_active":          nil,
		"filter_auth_method":     nil,
		"filter_account_type":    nil,
	}

	if f.Decision != nil {
		args["filter_decision"] = string(*f.Decision)
	}

	if f.Flag != nil {
		args["filter_flag"] = string(*f.Flag)
	}

	if f.IncrementalTag != nil {
		args["filter_incremental_tag"] = string(*f.IncrementalTag)
	}

	if f.IsAdmin != nil {
		args["filter_is_admin"] = *f.IsAdmin
	}

	if f.Active != nil {
		args["filter_active"] = *f.Active
	}

	if f.AuthMethod != nil {
		args["filter_auth_method"] = string(*f.AuthMethod)
	}

	if f.AccountType != nil {
		args["filter_account_type"] = string(*f.AccountType)
	}

	return args
}
