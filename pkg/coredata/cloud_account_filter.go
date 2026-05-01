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

type (
	CloudAccountFilter struct {
		provider  *CloudAccountProvider
		status    *CloudAccountStatus
		scopeKind *CloudAccountScopeKind
	}
)

func NewCloudAccountFilter() *CloudAccountFilter {
	return &CloudAccountFilter{}
}

func (f *CloudAccountFilter) WithProvider(provider CloudAccountProvider) *CloudAccountFilter {
	f.provider = &provider
	return f
}

func (f *CloudAccountFilter) WithStatus(status CloudAccountStatus) *CloudAccountFilter {
	f.status = &status
	return f
}

func (f *CloudAccountFilter) WithScopeKind(scopeKind CloudAccountScopeKind) *CloudAccountFilter {
	f.scopeKind = &scopeKind
	return f
}

func (f *CloudAccountFilter) SQLFragment() string {
	if f == nil {
		return "TRUE"
	}

	return `
(
    CASE
        WHEN @filter_cloud_account_provider::text IS NULL THEN TRUE
        ELSE provider = @filter_cloud_account_provider::text
    END
    AND
    CASE
        WHEN @filter_cloud_account_status::text IS NULL THEN TRUE
        ELSE status = @filter_cloud_account_status::text
    END
    AND
    CASE
        WHEN @filter_cloud_account_scope_kind::text IS NULL THEN TRUE
        ELSE scope_kind = @filter_cloud_account_scope_kind::text
    END
)`
}

func (f *CloudAccountFilter) SQLArguments() pgx.StrictNamedArgs {
	if f == nil {
		return pgx.StrictNamedArgs{}
	}

	args := pgx.StrictNamedArgs{
		"filter_cloud_account_provider":   nil,
		"filter_cloud_account_status":     nil,
		"filter_cloud_account_scope_kind": nil,
	}

	if f.provider != nil {
		args["filter_cloud_account_provider"] = f.provider.String()
	}
	if f.status != nil {
		args["filter_cloud_account_status"] = f.status.String()
	}
	if f.scopeKind != nil {
		args["filter_cloud_account_scope_kind"] = f.scopeKind.String()
	}

	return args
}
