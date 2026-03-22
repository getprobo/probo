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

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestCookieBanner_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("with required fields", func(t *testing.T) {
		t.Parallel()

		const query = `
			mutation($input: CreateCookieBannerInput!) {
				createCookieBanner(input: $input) {
					cookieBanner {
						id
						name
						domain
						state
						version
					}
				}
			}
		`

		var result struct {
			CreateCookieBanner struct {
				CookieBanner struct {
					ID      string `json:"id"`
					Name    string `json:"name"`
					Domain  string `json:"domain"`
					State   string `json:"state"`
					Version int    `json:"version"`
				} `json:"cookieBanner"`
			} `json:"createCookieBanner"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           "Main Site Banner",
				"domain":         "example.com",
			},
		}, &result)
		require.NoError(t, err)

		b := result.CreateCookieBanner.CookieBanner
		assert.NotEmpty(t, b.ID)
		assert.Equal(t, "Main Site Banner", b.Name)
		assert.Equal(t, "example.com", b.Domain)
		assert.Equal(t, "draft", b.State)
		assert.Equal(t, 1, b.Version)
	})

	t.Run("creates default categories", func(t *testing.T) {
		t.Parallel()

		bannerID := factory.NewCookieBanner(owner).WithName("Banner With Categories").Create()

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on CookieBanner {
						categories(first: 10) {
							totalCount
							edges {
								node {
									id
									name
									required
								}
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				Categories struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Node struct {
							ID       string `json:"id"`
							Name     string `json:"name"`
							Required bool   `json:"required"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"categories"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": bannerID}, &result)
		require.NoError(t, err)
		assert.Equal(t, 4, result.Node.Categories.TotalCount)

		hasRequired := false
		for _, edge := range result.Node.Categories.Edges {
			if edge.Node.Required {
				hasRequired = true
				break
			}
		}
		assert.True(t, hasRequired, "should have at least one required category")
	})
}

func TestCookieBanner_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bannerID := factory.NewCookieBanner(owner).WithName("Banner to Update").Create()

	const query = `
		mutation($input: UpdateCookieBannerInput!) {
			updateCookieBanner(input: $input) {
				cookieBanner {
					id
					name
					title
					description
				}
			}
		}
	`

	var result struct {
		UpdateCookieBanner struct {
			CookieBanner struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"cookieBanner"`
		} `json:"updateCookieBanner"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":          bannerID,
			"name":        "Updated Banner",
			"title":       "We use cookies",
			"description": "This site uses cookies to improve your experience.",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, bannerID, result.UpdateCookieBanner.CookieBanner.ID)
	assert.Equal(t, "Updated Banner", result.UpdateCookieBanner.CookieBanner.Name)
	assert.Equal(t, "We use cookies", result.UpdateCookieBanner.CookieBanner.Title)
}

func TestCookieBanner_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bannerID := factory.NewCookieBanner(owner).WithName("Banner to Delete").Create()

	const query = `
		mutation($input: DeleteCookieBannerInput!) {
			deleteCookieBanner(input: $input) {
				deletedCookieBannerId
			}
		}
	`

	var result struct {
		DeleteCookieBanner struct {
			DeletedCookieBannerID string `json:"deletedCookieBannerId"`
		} `json:"deleteCookieBanner"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id": bannerID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, bannerID, result.DeleteCookieBanner.DeletedCookieBannerID)
}

func TestCookieBanner_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	for _, name := range []string{"Banner A", "Banner B", "Banner C"} {
		factory.NewCookieBanner(owner).WithName(name).Create()
	}

	const query = `
		query($orgId: ID!) {
			node(id: $orgId) {
				... on Organization {
					cookieBanners(first: 10) {
						edges {
							node {
								id
								name
							}
						}
						totalCount
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			CookieBanners struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"cookieBanners"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"orgId": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.CookieBanners.TotalCount, 3)
}

func TestCookieBanner_Publish(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bannerID := factory.NewCookieBanner(owner).WithName("Banner to Publish").Create()

	const query = `
		mutation($input: PublishCookieBannerInput!) {
			publishCookieBanner(input: $input) {
				cookieBanner {
					id
					state
				}
			}
		}
	`

	var result struct {
		PublishCookieBanner struct {
			CookieBanner struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"cookieBanner"`
		} `json:"publishCookieBanner"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{"id": bannerID},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, "published", result.PublishCookieBanner.CookieBanner.State)
}

func TestCookieBanner_Disable(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bannerID := factory.NewCookieBanner(owner).WithName("Banner to Disable").Create()

	// Publish first so we can disable.
	const publishQuery = `
		mutation($input: PublishCookieBannerInput!) {
			publishCookieBanner(input: $input) {
				cookieBanner { id }
			}
		}
	`

	var publishResult struct {
		PublishCookieBanner struct {
			CookieBanner struct {
				ID string `json:"id"`
			} `json:"cookieBanner"`
		} `json:"publishCookieBanner"`
	}

	err := owner.Execute(publishQuery, map[string]any{
		"input": map[string]any{"id": bannerID},
	}, &publishResult)
	require.NoError(t, err)

	const query = `
		mutation($input: DisableCookieBannerInput!) {
			disableCookieBanner(input: $input) {
				cookieBanner {
					id
					state
				}
			}
		}
	`

	var result struct {
		DisableCookieBanner struct {
			CookieBanner struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"cookieBanner"`
		} `json:"disableCookieBanner"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{"id": bannerID},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, "disabled", result.DisableCookieBanner.CookieBanner.State)
}

func TestCookieCategory_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bannerID := factory.NewCookieBanner(owner).WithName("Banner for Categories").Create()

	const query = `
		mutation($input: CreateCookieCategoryInput!) {
			createCookieCategory(input: $input) {
				cookieCategory {
					id
					name
					description
					required
					rank
				}
			}
		}
	`

	var result struct {
		CreateCookieCategory struct {
			CookieCategory struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				Required    bool   `json:"required"`
				Rank        int    `json:"rank"`
			} `json:"cookieCategory"`
		} `json:"createCookieCategory"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"cookieBannerId": bannerID,
			"name":           "Marketing",
			"description":    "Marketing cookies",
			"rank":           5,
		},
	}, &result)
	require.NoError(t, err)

	c := result.CreateCookieCategory.CookieCategory
	assert.NotEmpty(t, c.ID)
	assert.Equal(t, "Marketing", c.Name)
	assert.Equal(t, "Marketing cookies", c.Description)
	assert.False(t, c.Required)
	assert.Equal(t, 5, c.Rank)
}

func TestCookieCategory_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bannerID := factory.NewCookieBanner(owner).WithName("Banner for Category Update").Create()
	categoryID := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{
		"name": "Old Name",
	})

	const query = `
		mutation($input: UpdateCookieCategoryInput!) {
			updateCookieCategory(input: $input) {
				cookieCategory {
					id
					name
				}
			}
		}
	`

	var result struct {
		UpdateCookieCategory struct {
			CookieCategory struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"cookieCategory"`
		} `json:"updateCookieCategory"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":   categoryID,
			"name": "New Name",
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, categoryID, result.UpdateCookieCategory.CookieCategory.ID)
	assert.Equal(t, "New Name", result.UpdateCookieCategory.CookieCategory.Name)
}

func TestCookieCategory_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bannerID := factory.NewCookieBanner(owner).WithName("Banner for Category Delete").Create()
	categoryID := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{
		"name": "Category to Delete",
	})

	const query = `
		mutation($input: DeleteCookieCategoryInput!) {
			deleteCookieCategory(input: $input) {
				deletedCookieCategoryId
			}
		}
	`

	var result struct {
		DeleteCookieCategory struct {
			DeletedCookieCategoryID string `json:"deletedCookieCategoryId"`
		} `json:"deleteCookieCategory"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id": categoryID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, categoryID, result.DeleteCookieCategory.DeletedCookieCategoryID)
}

func TestCookieBanner_RBAC(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	bannerID := factory.NewCookieBanner(owner).WithName("RBAC Banner").Create()

	t.Run("viewer can read cookie banner", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on CookieBanner {
						id
						name
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		}

		err := viewer.Execute(query, map[string]any{"id": bannerID}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.Node)
		assert.Equal(t, bannerID, result.Node.ID)
	})

	t.Run("viewer cannot create cookie banner", func(t *testing.T) {
		t.Parallel()

		const query = `
			mutation($input: CreateCookieBannerInput!) {
				createCookieBanner(input: $input) {
					cookieBanner { id }
				}
			}
		`

		_, err := viewer.Do(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           factory.SafeName("Viewer Banner"),
			},
		})
		testutil.RequireForbiddenError(t, err)
	})

	t.Run("viewer cannot update cookie banner", func(t *testing.T) {
		t.Parallel()

		const query = `
			mutation($input: UpdateCookieBannerInput!) {
				updateCookieBanner(input: $input) {
					cookieBanner { id }
				}
			}
		`

		_, err := viewer.Do(query, map[string]any{
			"input": map[string]any{
				"id":   bannerID,
				"name": "Viewer Updated",
			},
		})
		testutil.RequireForbiddenError(t, err)
	})

	t.Run("viewer cannot delete cookie banner", func(t *testing.T) {
		t.Parallel()

		const query = `
			mutation($input: DeleteCookieBannerInput!) {
				deleteCookieBanner(input: $input) {
					deletedCookieBannerId
				}
			}
		`

		_, err := viewer.Do(query, map[string]any{
			"input": map[string]any{
				"id": bannerID,
			},
		})
		testutil.RequireForbiddenError(t, err)
	})
}

func TestCookieBanner_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	bannerID := factory.NewCookieBanner(org1Owner).WithName("Org1 Banner").Create()

	t.Run("cannot read banner from another organization", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on CookieBanner {
						id
						name
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": bannerID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "cookie banner")
	})

	t.Run("cannot update banner from another organization", func(t *testing.T) {
		t.Parallel()

		const query = `
			mutation($input: UpdateCookieBannerInput!) {
				updateCookieBanner(input: $input) {
					cookieBanner { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   bannerID,
				"name": "Hijacked Banner",
			},
		})
		require.Error(t, err, "Should not be able to update banner from another org")
	})

	t.Run("cannot delete banner from another organization", func(t *testing.T) {
		t.Parallel()

		const query = `
			mutation($input: DeleteCookieBannerInput!) {
				deleteCookieBanner(input: $input) {
					deletedCookieBannerId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id": bannerID,
			},
		})
		require.Error(t, err, "Should not be able to delete banner from another org")
	})
}
