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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestCookieBannerGVLVendor(t *testing.T) {
	t.Parallel()

	t.Run("lists catalog and add/remove selected vendors when tcf is on", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		bannerID := factory.CreateCookieBanner(owner)
		factory.EnableCookieBannerTCF(t, bannerID)

		iabVendorID, _ := factory.SeedCommonGVLVendor(t, "Example Ad Vendor", false)

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

	t.Run("filters catalog by banner membership", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		bannerID := factory.CreateCookieBanner(owner)
		factory.EnableCookieBannerTCF(t, bannerID)

		onBannerID, _ := factory.SeedCommonGVLVendor(t, "On Banner Vendor", false)
		offBannerID, _ := factory.SeedCommonGVLVendor(t, "Off Banner Vendor", false)

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
				"iabVendorId":    onBannerID,
			},
		}, new(map[string]any))
		require.NoError(t, err)

		const catalogQuery = `
			query($filter: CommonGVLVendorFilter) {
				commonGVLVendors(first: 50, filter: $filter) {
					edges {
						node { iabVendorId }
					}
				}
			}
		`

		var onBanner struct {
			CommonGVLVendors struct {
				Edges []struct {
					Node struct {
						IabVendorID int `json:"iabVendorId"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"commonGVLVendors"`
		}

		err = owner.Execute(catalogQuery, map[string]any{
			"filter": map[string]any{
				"query":          "On Banner Vendor",
				"cookieBannerId": bannerID,
				"membership":     "ON_BANNER",
			},
		}, &onBanner)
		require.NoError(t, err)
		require.Len(t, onBanner.CommonGVLVendors.Edges, 1)
		assert.Equal(t, onBannerID, onBanner.CommonGVLVendors.Edges[0].Node.IabVendorID)

		var notOnBanner struct {
			CommonGVLVendors struct {
				Edges []struct {
					Node struct {
						IabVendorID int `json:"iabVendorId"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"commonGVLVendors"`
		}

		err = owner.Execute(catalogQuery, map[string]any{
			"filter": map[string]any{
				"query":          "Off Banner Vendor",
				"cookieBannerId": bannerID,
				"membership":     "NOT_ON_BANNER",
			},
		}, &notOnBanner)
		require.NoError(t, err)
		require.Len(t, notOnBanner.CommonGVLVendors.Edges, 1)
		assert.Equal(t, offBannerID, notOnBanner.CommonGVLVendors.Edges[0].Node.IabVendorID)

		err = owner.Execute(catalogQuery, map[string]any{
			"filter": map[string]any{"membership": "ON_BANNER"},
		}, new(map[string]any))
		testutil.RequireErrorCode(t, err, "INVALID")
	})

	t.Run("returns catalog versions", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		_, version := factory.SeedCommonGVLVendor(t, "Catalog Version Vendor", false)
		factory.SeedCommonGVLCatalogState(t, version)

		const query = `
			query {
				commonGVLCatalog {
					vendorListVersion
					tcfPolicyVersion
				}
			}
		`

		var result struct {
			CommonGVLCatalog struct {
				VendorListVersion *int `json:"vendorListVersion"`
				TcfPolicyVersion  *int `json:"tcfPolicyVersion"`
			} `json:"commonGVLCatalog"`
		}

		err := owner.Execute(query, nil, &result)
		require.NoError(t, err)
		require.NotNil(t, result.CommonGVLCatalog.VendorListVersion)
		require.NotNil(t, result.CommonGVLCatalog.TcfPolicyVersion)
		assert.Equal(t, version, *result.CommonGVLCatalog.VendorListVersion)
		assert.Equal(t, 5, *result.CommonGVLCatalog.TcfPolicyVersion)
	})

	t.Run("counts draft and published gvl vendors", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		bannerID := factory.CreateCookieBanner(owner)
		factory.EnableCookieBannerTCF(t, bannerID)

		firstID, _ := factory.SeedCommonGVLVendor(t, "Draft Count Vendor", false)
		secondID, _ := factory.SeedCommonGVLVendor(t, "Published Count Vendor", false)

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
				"iabVendorId":    firstID,
			},
		}, new(map[string]any))
		require.NoError(t, err)

		const statsQuery = `
			query($id: ID!) {
				node(id: $id) {
					... on CookieBanner {
						gvlVendors(first: 1) { totalCount }
						publishedVersion { gvlVendorCount }
					}
				}
			}
		`

		var beforePublish struct {
			Node struct {
				GVLVendors struct {
					TotalCount int `json:"totalCount"`
				} `json:"gvlVendors"`
				PublishedVersion *struct {
					GvlVendorCount int `json:"gvlVendorCount"`
				} `json:"publishedVersion"`
			} `json:"node"`
		}

		err = owner.Execute(statsQuery, map[string]any{"id": bannerID}, &beforePublish)
		require.NoError(t, err)
		assert.Equal(t, 1, beforePublish.Node.GVLVendors.TotalCount)
		assert.Nil(t, beforePublish.Node.PublishedVersion)

		published := publishBanner(t, owner, bannerID)
		assert.Equal(t, "PUBLISHED", published.State)

		err = owner.Execute(addMutation, map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"iabVendorId":    secondID,
			},
		}, new(map[string]any))
		require.NoError(t, err)

		var afterDraft struct {
			Node struct {
				GVLVendors struct {
					TotalCount int `json:"totalCount"`
				} `json:"gvlVendors"`
				PublishedVersion *struct {
					GvlVendorCount int `json:"gvlVendorCount"`
				} `json:"publishedVersion"`
			} `json:"node"`
		}

		err = owner.Execute(statsQuery, map[string]any{"id": bannerID}, &afterDraft)
		require.NoError(t, err)
		assert.Equal(t, 2, afterDraft.Node.GVLVendors.TotalCount)
		require.NotNil(t, afterDraft.Node.PublishedVersion)
		assert.Equal(t, 1, afterDraft.Node.PublishedVersion.GvlVendorCount)
	})

	t.Run("rejects add when tcf is off but still allows remove", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		bannerID := factory.CreateCookieBanner(owner)
		iabVendorID, _ := factory.SeedCommonGVLVendor(t, "Disabled TCF Vendor", false)

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

		// Removal is intentionally ungated so vendors linked while TCF was on
		// can still be cleaned up after it is turned off.
		err = owner.Execute(removeMutation, map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"iabVendorId":    iabVendorID,
			},
		}, new(map[string]any))
		require.NoError(t, err)
	})

	t.Run("removes a vendor linked before tcf was turned off", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		bannerID := factory.CreateCookieBanner(owner)
		factory.EnableCookieBannerTCF(t, bannerID)

		iabVendorID, _ := factory.SeedCommonGVLVendor(t, "Legacy Linked Vendor", false)

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
		require.NoError(t, err)

		factory.DisableCookieBannerTCF(t, bannerID)

		const removeMutation = `
			mutation($input: RemoveCookieBannerGVLVendorInput!) {
				removeCookieBannerGVLVendor(input: $input) {
					cookieBanner {
						gvlVendors { totalCount }
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

	t.Run("rejects a deleted catalog vendor", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		bannerID := factory.CreateCookieBanner(owner)
		factory.EnableCookieBannerTCF(t, bannerID)

		iabVendorID, _ := factory.SeedCommonGVLVendor(t, "Deleted Vendor", true)

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
