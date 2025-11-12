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

package types

import "go.probo.inc/probo/pkg/coredata"

type Organization struct {
	Name     string `json:"name" jsonschema:"the organization name"`
	ID       string `json:"id" jsonschema:"the organization ID"`
	TenantID string `json:"tenantID" jsonschema:"the tenant ID this organization belongs to"`
}

func NewOrganization(o *coredata.Organization) Organization {
	return Organization{
		Name:     o.Name,
		ID:       o.ID.String(),
		TenantID: o.ID.TenantID().String(),
	}
}
