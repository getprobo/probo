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
		status *InvitationStatus
	}
)

func NewInvitationFilter(status *InvitationStatus) *InvitationFilter {
	return &InvitationFilter{
		status: status,
	}
}

func (f *InvitationFilter) SQLArguments() pgx.NamedArgs {
	return pgx.NamedArgs{
		"status": f.status,
	}
}

func (f *InvitationFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @status::text IS NOT NULL THEN
			(CASE
				WHEN accepted_at IS NOT NULL THEN 'ACCEPTED'
				WHEN expires_at < NOW() THEN 'EXPIRED'
				ELSE 'PENDING'
			END) = @status::text
		ELSE TRUE
	END
)`
}
