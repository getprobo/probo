// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package console_v1

import (
	"context"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

func commonGVLVendorFilter(
	ctx context.Context,
	r *queryResolver,
	filter *types.CommonGVLVendorFilter,
) (*coredata.CommonGVLVendorFilter, error) {
	cdFilter := coredata.NewCommonGVLVendorFilter(nil)
	if filter == nil {
		return cdFilter, nil
	}

	cdFilter = coredata.NewCommonGVLVendorFilter(filter.Query)
	if filter.Membership == nil {
		return cdFilter, nil
	}

	if filter.CookieBannerID == nil {
		return nil, gqlutils.Invalidf(ctx, "cookieBannerId is required when filtering by membership")
	}

	if _, err := r.authorize(ctx, *filter.CookieBannerID, probo.ActionCookieBannerGet); err != nil {
		return nil, err
	}

	return cdFilter.WithMembership(filter.CookieBannerID, filter.Membership), nil
}
