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

type datumVendorCopy struct {
	datumVendors DatumVendors
	scope        Scoper
	position     int
}

func (c *datumVendorCopy) Next() bool {
	return c.position < len(c.datumVendors)
}

func (c *datumVendorCopy) Values() ([]interface{}, error) {
	if c.position >= len(c.datumVendors) {
		return nil, nil
	}

	datumVendor := c.datumVendors[c.position]
	c.position++

	return []any{
		c.scope.GetTenantID(),
		datumVendor.DatumID,
		datumVendor.VendorID,
		datumVendor.SnapshotID,
		datumVendor.CreatedAt,
	}, nil
}

func (c *datumVendorCopy) Err() error {
	return nil
}
