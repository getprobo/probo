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

import "github.com/jackc/pgx/v5"

type CMPDomainSurfaceFilter struct {
	states []CMPDomainSurfaceState
}

func NewCMPDomainSurfaceFilter(states []CMPDomainSurfaceState) *CMPDomainSurfaceFilter {
	return &CMPDomainSurfaceFilter{states: states}
}

func (f *CMPDomainSurfaceFilter) SQLFragment() string {
	// COVERED wins over the stored status while any banner in the domain
	// uses the same origin. The comparison is case-insensitive because
	// banner origins are not stored in a single case.
	return `
(
	CASE
		WHEN @filter_states::text[] IS NULL THEN TRUE
		ELSE (
			CASE
				WHEN EXISTS (
					SELECT 1
					FROM cookie_banners
					WHERE cookie_banners.cmp_domain_id = cmp_domain_surfaces.cmp_domain_id
						AND lower(cookie_banners.origin) = lower(cmp_domain_surfaces.origin)
				) THEN @covered_state::text
				ELSE cmp_domain_surfaces.status
			END
		) = ANY(@filter_states::text[])
	END
)`
}

func (f *CMPDomainSurfaceFilter) SQLArguments() pgx.StrictNamedArgs {
	var states []CMPDomainSurfaceState
	if f != nil {
		states = f.states
	}

	return pgx.StrictNamedArgs{
		"filter_states": filterStrings(states),
		"covered_state": CMPDomainSurfaceStateCovered,
	}
}
