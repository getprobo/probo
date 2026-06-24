// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
