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

import (
	"github.com/jackc/pgx/v5"
)

type CookieConsentRecordFilter struct {
	action    *CookieConsentAction
	visitorID *string
	version   *int
}

func NewCookieConsentRecordFilter(
	action *CookieConsentAction,
	visitorID *string,
	version *int,
) *CookieConsentRecordFilter {
	return &CookieConsentRecordFilter{
		action:    action,
		visitorID: visitorID,
		version:   version,
	}
}

func (f *CookieConsentRecordFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @filter_action::text IS NOT NULL THEN
			action = @filter_action::cookie_consent_action
		ELSE TRUE
	END
)
AND
(
	CASE
		WHEN @filter_visitor_id::text IS NOT NULL THEN
			visitor_id = @filter_visitor_id
		ELSE TRUE
	END
)
AND
(
	CASE
		WHEN @filter_version::int IS NOT NULL THEN
			cookie_banner_version_id = (
				SELECT id FROM cookie_banner_versions
				WHERE cookie_banner_id = cookie_consent_records.cookie_banner_id
				AND version = @filter_version
			)
		ELSE TRUE
	END
)`
}

func (f *CookieConsentRecordFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"filter_action":     nil,
		"filter_visitor_id": nil,
		"filter_version":    nil,
	}

	if f.action != nil {
		args["filter_action"] = string(*f.action)
	}

	if f.visitorID != nil {
		args["filter_visitor_id"] = *f.visitorID
	}

	if f.version != nil {
		args["filter_version"] = *f.version
	}

	return args
}
