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

package console_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func uniqueIABVendorID() int {
	return int(time.Now().UnixNano()%90_000_000) + 10_000_000
}

func TestCookieBannerGVLVendor(t *testing.T) {
	t.Parallel()

	t.Run("lists catalog and add/remove selected vendors when tcf is on", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		bannerID := factory.CreateCookieBanner(owner)
		factory.EnableCookieBannerTCF(t, bannerID)

		iabVendorID := uniqueIABVendorID()
		factory.SeedCommonGVLVendor(t, iabVendorID, "Example Ad Vendor", false)

		const catalogQuery = `
			query($query: String) {
				commonGVLVendors(first: 50, filter: { query: $query }) {
					totalCount
					edges {
						node {
							iabVendorId
							name
						}
					}
				}
			}
		`

		var catalog struct {
			CommonGVLVendors struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						IabVendorID int    `json:"iabVendorId"`
						Name        string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"commonGVLVendors"`
		}

		err := owner.Execute(catalogQuery, map[string]any{"query": "Example Ad Vendor"}, &catalog)
		require.NoError(t, err)
		require.GreaterOrEqual(t, catalog.CommonGVLVendors.TotalCount, 1)

		found := false
		for _, edge := range catalog.CommonGVLVendors.Edges {
			if edge.Node.IabVendorID == iabVendorID {
				found = true
				assert.Equal(t, "Example Ad Vendor", edge.Node.Name)
			}
		}
		require.True(t, found, "seeded catalog vendor must appear in commonGVLVendors")

		const addMutation = `
			mutation($input: AddCookieBannerGVLVendorInput!) {
				addCookieBannerGVLVendor(input: $input) {
					commonGVLVendor {
						iabVendorId
						name
					}
					cookieBanner {
						id
						gvlVendors(first: 20) {
							totalCount
							edges {
								node { iabVendorId }
							}
						}
					}
				}
			}
		`

		var added struct {
			AddCookieBannerGVLVendor struct {
				CommonGVLVendor struct {
					IabVendorID int    `json:"iabVendorId"`
					Name        string `json:"name"`
				} `json:"commonGVLVendor"`
				CookieBanner struct {
					ID         string `json:"id"`
					GVLVendors struct {
						TotalCount int `json:"totalCount"`
						Edges      []struct {
							Node struct {
								IabVendorID int `json:"iabVendorId"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"gvlVendors"`
				} `json:"cookieBanner"`
			} `json:"addCookieBannerGVLVendor"`
		}

		err = owner.Execute(addMutation, map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"iabVendorId":    iabVendorID,
			},
		}, &added)
		require.NoError(t, err)
		assert.Equal(t, iabVendorID, added.AddCookieBannerGVLVendor.CommonGVLVendor.IabVendorID)
		assert.Equal(t, 1, added.AddCookieBannerGVLVendor.CookieBanner.GVLVendors.TotalCount)

		const removeMutation = `
			mutation($input: RemoveCookieBannerGVLVendorInput!) {
				removeCookieBannerGVLVendor(input: $input) {
					cookieBanner {
						gvlVendors(first: 20) {
							totalCount
						}
					}
				}
			}
		`

		var removed struct {
			RemoveCookieBannerGVLVendor struct {
				CookieBanner struct {
					GVLVendors struct {
						TotalCount int `json:"totalCount"`
					} `json:"gvlVendors"`
				} `json:"cookieBanner"`
			} `json:"removeCookieBannerGVLVendor"`
		}

		err = owner.Execute(removeMutation, map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"iabVendorId":    iabVendorID,
			},
		}, &removed)
		require.NoError(t, err)
		assert.Equal(t, 0, removed.RemoveCookieBannerGVLVendor.CookieBanner.GVLVendors.TotalCount)
	})

	t.Run("rejects add and remove when tcf is off", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		bannerID := factory.CreateCookieBanner(owner)
		iabVendorID := uniqueIABVendorID()
		factory.SeedCommonGVLVendor(t, iabVendorID, "Disabled TCF Vendor", false)

		const addMutation = `
			mutation($input: AddCookieBannerGVLVendorInput!) {
				addCookieBannerGVLVendor(input: $input) {
					cookieBanner { id }
				}
			}
		`

		err := owner.Execute(addMutation, map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"iabVendorId":    iabVendorID,
			},
		}, new(map[string]any))
		testutil.RequireErrorCode(t, err, "INVALID")

		const removeMutation = `
			mutation($input: RemoveCookieBannerGVLVendorInput!) {
				removeCookieBannerGVLVendor(input: $input) {
					cookieBanner { id }
				}
			}
		`

		err = owner.Execute(removeMutation, map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"iabVendorId":    iabVendorID,
			},
		}, new(map[string]any))
		testutil.RequireErrorCode(t, err, "INVALID")
	})

	t.Run("rejects a deleted catalog vendor", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		bannerID := factory.CreateCookieBanner(owner)
		factory.EnableCookieBannerTCF(t, bannerID)

		iabVendorID := uniqueIABVendorID()
		factory.SeedCommonGVLVendor(t, iabVendorID, "Deleted Vendor", true)

		const addMutation = `
			mutation($input: AddCookieBannerGVLVendorInput!) {
				addCookieBannerGVLVendor(input: $input) {
					cookieBanner { id }
				}
			}
		`

		err := owner.Execute(addMutation, map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"iabVendorId":    iabVendorID,
			},
		}, new(map[string]any))
		testutil.RequireErrorCode(t, err, "INVALID")
	})
}
