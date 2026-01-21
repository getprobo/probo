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

type ControlObligationFilter struct {
	obligationType *ObligationType
}

func NewControlObligationFilter(obligationType *ObligationType) *ControlObligationFilter {
	return &ControlObligationFilter{
		obligationType: obligationType,
	}
}

func (f *ControlObligationFilter) SQLArguments() pgx.NamedArgs {
	args := pgx.NamedArgs{
		"filter_obligation_type": nil,
	}

	if f.obligationType != nil {
		args["filter_obligation_type"] = *f.obligationType
	}

	return args
}

func (f *ControlObligationFilter) SQLFragment() string {
	return "(@filter_obligation_type::obligation_type IS NULL OR type = @filter_obligation_type)"
}
