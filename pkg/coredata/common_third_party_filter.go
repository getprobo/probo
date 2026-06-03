// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

type CommonThirdPartyFilter struct {
	name *string
}

func NewCommonThirdPartyFilter(name *string) *CommonThirdPartyFilter {
	return &CommonThirdPartyFilter{name: name}
}

func (f *CommonThirdPartyFilter) SQLFragment() string {
	return `(
	CASE
		WHEN @filter_name::text IS NOT NULL THEN
			name ILIKE '%' || @filter_name || '%'
		ELSE TRUE
	END
)`
}

func (f *CommonThirdPartyFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{"filter_name": nil}
	if f.name != nil {
		args["filter_name"] = *f.name
	}

	return args
}
