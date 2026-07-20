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
	CompliancePortalFileFilter struct {
		compliancePortalVisibilities []CompliancePortalVisibility
	}
)

type CompliancePortalFileFilterOption func(f *CompliancePortalFileFilter)

func NewCompliancePortalFileFilter(opts ...CompliancePortalFileFilterOption) *CompliancePortalFileFilter {
	f := &CompliancePortalFileFilter{}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

func WithCompliancePortalFileVisibilities(visibilities ...CompliancePortalVisibility) CompliancePortalFileFilterOption {
	return func(f *CompliancePortalFileFilter) {
		f.compliancePortalVisibilities = visibilities
	}
}

func (f *CompliancePortalFileFilter) SQLArguments() pgx.NamedArgs {
	var visibilities []string
	if f.compliancePortalVisibilities != nil {
		visibilities = make([]string, len(f.compliancePortalVisibilities))
		for i, v := range f.compliancePortalVisibilities {
			visibilities[i] = v.String()
		}
	}

	return pgx.NamedArgs{
		"trust_center_visibilities": visibilities,
	}
}

func (f *CompliancePortalFileFilter) SQLFragment() string {
	return `CASE
  WHEN @trust_center_visibilities::trust_center_visibility[] IS NOT NULL THEN
    trust_center_visibility = ANY(@trust_center_visibilities::trust_center_visibility[])
  ELSE TRUE
END
`
}
