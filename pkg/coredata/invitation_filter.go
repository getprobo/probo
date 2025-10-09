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
	InvitationFilter struct {
		onlyPending *bool
	}
)

func NewInvitationFilter(onlyPending *bool) *InvitationFilter {
	return &InvitationFilter{
		onlyPending: onlyPending,
	}
}

func (f *InvitationFilter) SQLArguments() pgx.NamedArgs {
	return pgx.NamedArgs{
		"only_pending": f.onlyPending,
	}
}

func (f *InvitationFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @only_pending::boolean IS NOT NULL AND @only_pending::boolean = true THEN
			(accepted_at IS NULL AND expires_at > NOW())
		ELSE TRUE
	END
)`
}

