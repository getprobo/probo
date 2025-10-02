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
	AuditFilter struct {
		trustCenterVisibilities []TrustCenterVisibility
	}
)

func NewAuditFilter() *AuditFilter {
	return &AuditFilter{}
}

func NewAuditTrustCenterFilter() *AuditFilter {
	return &AuditFilter{
		trustCenterVisibilities: []TrustCenterVisibility{
			TrustCenterVisibilityPrivate,
			TrustCenterVisibilityPublic,
		},
	}
}

func (f *AuditFilter) SQLArguments() pgx.NamedArgs {
	args := pgx.NamedArgs{}

	if f.trustCenterVisibilities != nil {
		visibilities := make([]string, len(f.trustCenterVisibilities))
		for i, v := range f.trustCenterVisibilities {
			visibilities[i] = v.String()
		}
		args["trust_center_visibilities"] = visibilities
	}

	return args
}

func (f *AuditFilter) SQLFragment() string {
	if f.trustCenterVisibilities != nil {
		return "trust_center_visibility = ANY(@trust_center_visibilities::trust_center_visibility[])"
	}

	return "TRUE"
}
