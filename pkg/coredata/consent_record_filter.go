// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import "github.com/jackc/pgx/v5"

type ConsentRecordFilter struct {
	action *ConsentAction
}

func NewConsentRecordFilter(action *ConsentAction) *ConsentRecordFilter {
	return &ConsentRecordFilter{
		action: action,
	}
}

func (f *ConsentRecordFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @filter_action::text IS NOT NULL THEN
			action = @filter_action::consent_action
		ELSE TRUE
	END
)`
}

func (f *ConsentRecordFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"filter_action": nil,
	}

	if f.action != nil {
		args["filter_action"] = string(*f.action)
	}

	return args
}
