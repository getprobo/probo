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

func (f *AuditFilter) WithTrustCenterVisibilities(visibilities ...TrustCenterVisibility) *AuditFilter {
	f.trustCenterVisibilities = visibilities
	return f
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
