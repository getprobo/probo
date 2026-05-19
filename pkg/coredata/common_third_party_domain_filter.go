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

type CommonThirdPartyDomainFilter struct {
	domains []string
}

func NewCommonThirdPartyDomainFilter(domains []string) *CommonThirdPartyDomainFilter {
	return &CommonThirdPartyDomainFilter{domains: domains}
}

func (f *CommonThirdPartyDomainFilter) SQLFragment() string {
	return `(
	CASE
		WHEN @filter_domains::text[] IS NOT NULL THEN
			domain = ANY(@filter_domains::text[])
		ELSE TRUE
	END
)`
}

func (f *CommonThirdPartyDomainFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{"filter_domains": nil}
	if len(f.domains) > 0 {
		args["filter_domains"] = f.domains
	}
	return args
}
