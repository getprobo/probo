// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

type (
	TrustCenterFileFilter struct {
		trustCenterVisibilities []TrustCenterVisibility
	}
)

type TrustCenterFileFilterOption func(f *TrustCenterFileFilter)

func NewTrustCenterFileFilter(opts ...TrustCenterFileFilterOption) *TrustCenterFileFilter {
	f := &TrustCenterFileFilter{}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

func WithTrustCenterFileVisibilities(visibilities ...TrustCenterVisibility) TrustCenterFileFilterOption {
	return func(f *TrustCenterFileFilter) {
		f.trustCenterVisibilities = visibilities
	}
}

func (f *TrustCenterFileFilter) SQLArguments() pgx.NamedArgs {
	var visibilities []string
	if f.trustCenterVisibilities != nil {
		visibilities = make([]string, len(f.trustCenterVisibilities))
		for i, v := range f.trustCenterVisibilities {
			visibilities[i] = v.String()
		}
	}
	return pgx.NamedArgs{
		"trust_center_visibilities": visibilities,
	}
}

func (f *TrustCenterFileFilter) SQLFragment() string {
	return `CASE
  WHEN @trust_center_visibilities::trust_center_visibility[] IS NOT NULL THEN
    trust_center_visibility = ANY(@trust_center_visibilities::trust_center_visibility[])
  ELSE TRUE
END
`
}
