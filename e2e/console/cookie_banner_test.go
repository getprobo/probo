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

	t.Run("with required fields", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		const query = `
			mutation CreateCookieBanner($input: CreateCookieBannerInput!) {
				createCookieBanner(input: $input) {
					cookieBannerEdge {
						node {
							id
							name
							origin
							state
							cookiePolicyUrl
							consentExpiryDays
							consentMode
							showBranding
							defaultLanguage
							createdAt
							updatedAt
						}
					}
				}
			}
		`

		name := factory.SafeName("Banner")
		origin := factory.SafeOrigin()

		var result struct {
			CreateCookieBanner struct {
				CookieBannerEdge struct {
					Node struct {
						ID                string `json:"id"`
						Name              string `json:"name"`
						Origin            string `json:"origin"`
						State             string `json:"state"`
						CookiePolicyUrl   string `json:"cookiePolicyUrl"`
						ConsentExpiryDays int    `json:"consentExpiryDays"`
						ConsentMode       string `json:"consentMode"`
						ShowBranding      bool   `json:"showBranding"`
						DefaultLanguage   string `json:"defaultLanguage"`
						CreatedAt         string `json:"createdAt"`
						UpdatedAt         string `json:"updatedAt"`
					} `json:"node"`
				} `json:"cookieBannerEdge"`
			} `json:"createCookieBanner"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId":    owner.GetOrganizationID().String(),
				"name":              name,
				"origin":            origin,
				"cookiePolicyUrl":   "https://example.com/cookies",
				"consentExpiryDays": 365,
				"consentMode":       "OPT_IN",
			},
		}, &result)

		require.NoError(t, err)
		node := result.CreateCookieBanner.CookieBannerEdge.Node
		assert.NotEmpty(t, node.ID)
		assert.Equal(t, name, node.Name)
		assert.Equal(t, "ACTIVE", node.State)
		assert.Equal(t, "https://example.com/cookies", node.CookiePolicyUrl)
		assert.Equal(t, 365, node.ConsentExpiryDays)
		assert.Equal(t, "OPT_IN", node.ConsentMode)
		assert.Equal(t, "en", node.DefaultLanguage)
		assert.NotEmpty(t, node.CreatedAt)
		assert.NotEmpty(t, node.UpdatedAt)
	})

	t.Run("with all fields including privacyPolicyUrl", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		const query = `
			mutation CreateCookieBanner($input: CreateCookieBannerInput!) {
				createCookieBanner(input: $input) {
					cookieBannerEdge {
						node {
							id
							privacyPolicyUrl
							consentMode
						}
					}
				}
			}
		`

		var result struct {
			CreateCookieBanner struct {
				CookieBannerEdge struct {
					Node struct {
						ID               string  `json:"id"`
						PrivacyPolicyUrl *string `json:"privacyPolicyUrl"`
						ConsentMode      string  `json:"consentMode"`
					} `json:"node"`
				} `json:"cookieBannerEdge"`
			} `json:"createCookieBanner"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId":    owner.GetOrganizationID().String(),
				"name":              factory.SafeName("Banner"),
				"origin":            factory.SafeOrigin(),
				"cookiePolicyUrl":   "https://example.com/cookies",
				"privacyPolicyUrl":  "https://example.com/privacy",
				"consentExpiryDays": 180,
				"consentMode":       "OPT_OUT",
			},
		}, &result)

		require.NoError(t, err)
		node := result.CreateCookieBanner.CookieBannerEdge.Node
		assert.NotEmpty(t, node.ID)
		require.NotNil(t, node.PrivacyPolicyUrl)
		assert.Equal(t, "https://example.com/privacy", *node.PrivacyPolicyUrl)
		assert.Equal(t, "OPT_OUT", node.ConsentMode)
	})

	t.Run("creates default categories", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

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
									kind
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
							ID   string `json:"id"`
							Name string `json:"name"`
							Kind string `json:"kind"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"categories"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": bannerID}, &result)
		require.NoError(t, err)
		assert.Greater(t, result.Node.Categories.TotalCount, 0)

		kinds := make(map[string]bool)
		for _, e := range result.Node.Categories.Edges {
			kinds[e.Node.Kind] = true
		}
		assert.True(t, kinds["NECESSARY"], "should have a NECESSARY category")
		assert.True(t, kinds["UNCATEGORISED"], "should have an UNCATEGORISED category")
	})

	t.Run("duplicate origin conflict", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		origin := factory.SafeOrigin()
		factory.NewCookieBanner(owner).WithOrigin(origin).Create()

		_, err := owner.Do(`
			mutation CreateCookieBanner($input: CreateCookieBannerInput!) {
				createCookieBanner(input: $input) {
					cookieBannerEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId":    owner.GetOrganizationID().String(),
				"name":              factory.SafeName("Banner"),
				"origin":            origin,
				"cookiePolicyUrl":   "https://example.com/cookies",
				"consentExpiryDays": 365,
				"consentMode":       "OPT_IN",
			},
		})
		require.Error(t, err)
	})

	t.Run("validation error on missing name", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		_, err := owner.Do(`
			mutation CreateCookieBanner($input: CreateCookieBannerInput!) {
				createCookieBanner(input: $input) {
					cookieBannerEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId":    owner.GetOrganizationID().String(),
				"name":              "",
				"origin":            factory.SafeOrigin(),
				"cookiePolicyUrl":   "https://example.com/cookies",
				"consentExpiryDays": 365,
				"consentMode":       "OPT_IN",
			},
		})
		require.Error(t, err)
	})
}

func TestCookieBanner_Update(t *testing.T) {
	t.Parallel()

	t.Run("partial update name only", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		const query = `
			mutation UpdateCookieBanner($input: UpdateCookieBannerInput!) {
				updateCookieBanner(input: $input) {
					cookieBanner {
						id
						name
						consentExpiryDays
					}
				}
			}
		`

		newName := factory.SafeName("Updated")
		var result struct {
			UpdateCookieBanner struct {
				CookieBanner struct {
					ID                string `json:"id"`
					Name              string `json:"name"`
					ConsentExpiryDays int    `json:"consentExpiryDays"`
				} `json:"cookieBanner"`
			} `json:"updateCookieBanner"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"name":           newName,
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, bannerID, result.UpdateCookieBanner.CookieBanner.ID)
		assert.Equal(t, newName, result.UpdateCookieBanner.CookieBanner.Name)
		assert.Equal(t, 365, result.UpdateCookieBanner.CookieBanner.ConsentExpiryDays)
	})

	t.Run("update consent settings", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		const query = `
			mutation UpdateCookieBanner($input: UpdateCookieBannerInput!) {
				updateCookieBanner(input: $input) {
					cookieBanner {
						consentExpiryDays
						consentMode
						defaultLanguage
					}
				}
			}
		`

		var result struct {
			UpdateCookieBanner struct {
				CookieBanner struct {
					ConsentExpiryDays int    `json:"consentExpiryDays"`
					ConsentMode       string `json:"consentMode"`
					DefaultLanguage   string `json:"defaultLanguage"`
				} `json:"cookieBanner"`
			} `json:"updateCookieBanner"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieBannerId":    bannerID,
				"consentExpiryDays": 90,
				"consentMode":       "OPT_OUT",
				"defaultLanguage":   "fr",
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, 90, result.UpdateCookieBanner.CookieBanner.ConsentExpiryDays)
		assert.Equal(t, "OPT_OUT", result.UpdateCookieBanner.CookieBanner.ConsentMode)
		assert.Equal(t, "fr", result.UpdateCookieBanner.CookieBanner.DefaultLanguage)
	})
}

func TestCookieBanner_Delete(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		const query = `
			mutation DeleteCookieBanner($input: DeleteCookieBannerInput!) {
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
				"cookieBannerId": bannerID,
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, bannerID, result.DeleteCookieBanner.DeletedCookieBannerID)
	})
}

func TestCookieBanner_ActivateDeactivate(t *testing.T) {
	t.Parallel()

	t.Run("deactivate active banner", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		const query = `
			mutation DeactivateCookieBanner($input: DeactivateCookieBannerInput!) {
				deactivateCookieBanner(input: $input) {
					cookieBanner {
						id
						state
					}
				}
			}
		`

		var result struct {
			DeactivateCookieBanner struct {
				CookieBanner struct {
					ID    string `json:"id"`
					State string `json:"state"`
				} `json:"cookieBanner"`
			} `json:"deactivateCookieBanner"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{"cookieBannerId": bannerID},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, "INACTIVE", result.DeactivateCookieBanner.CookieBanner.State)
	})

	t.Run("activate inactive banner", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		// Deactivate first
		var deactivateResult struct {
			DeactivateCookieBanner struct {
				CookieBanner struct {
					State string `json:"state"`
				} `json:"cookieBanner"`
			} `json:"deactivateCookieBanner"`
		}
		err := owner.Execute(`
			mutation($input: DeactivateCookieBannerInput!) {
				deactivateCookieBanner(input: $input) {
					cookieBanner { state }
				}
			}
		`, map[string]any{"input": map[string]any{"cookieBannerId": bannerID}}, &deactivateResult)
		require.NoError(t, err)
		require.Equal(t, "INACTIVE", deactivateResult.DeactivateCookieBanner.CookieBanner.State)

		// Activate
		const query = `
			mutation ActivateCookieBanner($input: ActivateCookieBannerInput!) {
				activateCookieBanner(input: $input) {
					cookieBanner {
						id
						state
					}
				}
			}
		`

		var result struct {
			ActivateCookieBanner struct {
				CookieBanner struct {
					ID    string `json:"id"`
					State string `json:"state"`
				} `json:"cookieBanner"`
			} `json:"activateCookieBanner"`
		}

		err = owner.Execute(query, map[string]any{
			"input": map[string]any{"cookieBannerId": bannerID},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, "ACTIVE", result.ActivateCookieBanner.CookieBanner.State)
	})

	t.Run("deactivate already inactive returns error", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		deactivateQuery := `
			mutation($input: DeactivateCookieBannerInput!) {
				deactivateCookieBanner(input: $input) {
					cookieBanner { state }
				}
			}
		`

		var result struct {
			DeactivateCookieBanner struct {
				CookieBanner struct {
					State string `json:"state"`
				} `json:"cookieBanner"`
			} `json:"deactivateCookieBanner"`
		}

		err := owner.Execute(deactivateQuery, map[string]any{
			"input": map[string]any{"cookieBannerId": bannerID},
		}, &result)
		require.NoError(t, err)

		_, err = owner.Do(deactivateQuery, map[string]any{
			"input": map[string]any{"cookieBannerId": bannerID},
		})
		require.Error(t, err)
	})
}

func TestCookieBanner_List(t *testing.T) {
	t.Parallel()

	t.Run("lists banners via organization", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		factory.CreateCookieBanner(owner)
		factory.CreateCookieBanner(owner)

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						cookieBanners(first: 10) {
							totalCount
							edges {
								node {
									id
									name
								}
							}
							pageInfo {
								hasNextPage
								hasPreviousPage
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				CookieBanners struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						HasNextPage     bool `json:"hasNextPage"`
						HasPreviousPage bool `json:"hasPreviousPage"`
					} `json:"pageInfo"`
				} `json:"cookieBanners"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
		}, &result)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Node.CookieBanners.TotalCount, 2)
		assert.GreaterOrEqual(t, len(result.Node.CookieBanners.Edges), 2)
	})
}

func TestCookieBanner_Node(t *testing.T) {
	t.Parallel()

	t.Run("fetch by ID", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on CookieBanner {
						id
						name
						origin
						state
						consentMode
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Origin      string `json:"origin"`
				State       string `json:"state"`
				ConsentMode string `json:"consentMode"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": bannerID}, &result)
		require.NoError(t, err)
		assert.Equal(t, bannerID, result.Node.ID)
		assert.NotEmpty(t, result.Node.Name)
		assert.Equal(t, "ACTIVE", result.Node.State)
	})
}

func TestCookieBanner_PublishVersion(t *testing.T) {
	t.Parallel()

	t.Run("publish draft version", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		const publishQuery = `
			mutation PublishCookieBannerVersion($input: PublishCookieBannerVersionInput!) {
				publishCookieBannerVersion(input: $input) {
					cookieBannerVersion {
						id
						version
						state
					}
					cookieBanner {
						id
					}
				}
			}
		`

		var publishResult struct {
			PublishCookieBannerVersion struct {
				CookieBannerVersion struct {
					ID      string `json:"id"`
					Version int    `json:"version"`
					State   string `json:"state"`
				} `json:"cookieBannerVersion"`
				CookieBanner struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"publishCookieBannerVersion"`
		}

		err := owner.Execute(publishQuery, map[string]any{
			"input": map[string]any{"cookieBannerId": bannerID},
		}, &publishResult)

		require.NoError(t, err)
		assert.Equal(t, 1, publishResult.PublishCookieBannerVersion.CookieBannerVersion.Version)
		assert.Equal(t, "published", publishResult.PublishCookieBannerVersion.CookieBannerVersion.State)
		assert.Equal(t, bannerID, publishResult.PublishCookieBannerVersion.CookieBanner.ID)
	})

	t.Run("latestVersion resolver", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on CookieBanner {
						latestVersion {
							id
							version
							state
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				LatestVersion *struct {
					ID      string `json:"id"`
					Version int    `json:"version"`
					State   string `json:"state"`
				} `json:"latestVersion"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": bannerID}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.Node.LatestVersion)
		assert.Equal(t, "draft", result.Node.LatestVersion.State)
		assert.Equal(t, 1, result.Node.LatestVersion.Version)
	})
}

func TestCookieBanner_UpsertTranslation(t *testing.T) {
	t.Parallel()

	t.Run("insert new language", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		const query = `
			mutation UpsertCookieBannerTranslation($input: UpsertCookieBannerTranslationInput!) {
				upsertCookieBannerTranslation(input: $input) {
					cookieBannerTranslation {
						id
						language
						translations
					}
					cookieBanner {
						id
					}
				}
			}
		`

		var result struct {
			UpsertCookieBannerTranslation struct {
				CookieBannerTranslation struct {
					ID           string `json:"id"`
					Language     string `json:"language"`
					Translations string `json:"translations"`
				} `json:"cookieBannerTranslation"`
				CookieBanner struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"upsertCookieBannerTranslation"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"language":       "de",
				"translations":   `{"title":"Cookie Einstellungen","description":"Wir verwenden Cookies"}`,
			},
		}, &result)

		require.NoError(t, err)
		assert.NotEmpty(t, result.UpsertCookieBannerTranslation.CookieBannerTranslation.ID)
		assert.Equal(t, "de", result.UpsertCookieBannerTranslation.CookieBannerTranslation.Language)
		assert.Equal(t, bannerID, result.UpsertCookieBannerTranslation.CookieBanner.ID)
	})

	t.Run("update existing language", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		const query = `
			mutation UpsertCookieBannerTranslation($input: UpsertCookieBannerTranslationInput!) {
				upsertCookieBannerTranslation(input: $input) {
					cookieBannerTranslation {
						id
						language
						translations
					}
				}
			}
		`

		input := map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"language":       "es",
				"translations":   `{"title":"Configuracion de cookies"}`,
			},
		}

		var result1 struct {
			UpsertCookieBannerTranslation struct {
				CookieBannerTranslation struct {
					ID string `json:"id"`
				} `json:"cookieBannerTranslation"`
			} `json:"upsertCookieBannerTranslation"`
		}
		err := owner.Execute(query, input, &result1)
		require.NoError(t, err)
		firstID := result1.UpsertCookieBannerTranslation.CookieBannerTranslation.ID

		input["input"].(map[string]any)["translations"] = `{"title":"Ajustes de cookies"}`

		var result2 struct {
			UpsertCookieBannerTranslation struct {
				CookieBannerTranslation struct {
					ID           string `json:"id"`
					Translations string `json:"translations"`
				} `json:"cookieBannerTranslation"`
			} `json:"upsertCookieBannerTranslation"`
		}
		err = owner.Execute(query, input, &result2)
		require.NoError(t, err)
		assert.Equal(t, firstID, result2.UpsertCookieBannerTranslation.CookieBannerTranslation.ID)
		assert.Contains(t, result2.UpsertCookieBannerTranslation.CookieBannerTranslation.Translations, "Ajustes de cookies")
	})

	t.Run("translations resolver on banner", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on CookieBanner {
						translations {
							id
							language
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				Translations []struct {
					ID       string `json:"id"`
					Language string `json:"language"`
				} `json:"translations"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": bannerID}, &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Node.Translations)
	})
}

func TestCookieBanner_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("viewer cannot create", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		_, err := viewer.Do(`
			mutation CreateCookieBanner($input: CreateCookieBannerInput!) {
				createCookieBanner(input: $input) {
					cookieBannerEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId":    viewer.GetOrganizationID().String(),
				"name":              factory.SafeName("Banner"),
				"origin":            factory.SafeOrigin(),
				"cookiePolicyUrl":   "https://example.com/cookies",
				"consentExpiryDays": 365,
				"consentMode":       "OPT_IN",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create cookie banner")
	})

	t.Run("viewer cannot update", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)

		_, err := viewer.Do(`
			mutation UpdateCookieBanner($input: UpdateCookieBannerInput!) {
				updateCookieBanner(input: $input) {
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"name":           "Updated",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to update cookie banner")
	})

	t.Run("viewer cannot delete", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)

		_, err := viewer.Do(`
			mutation DeleteCookieBanner($input: DeleteCookieBannerInput!) {
				deleteCookieBanner(input: $input) {
					deletedCookieBannerId
				}
			}
		`, map[string]any{
			"input": map[string]any{"cookieBannerId": bannerID},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to delete cookie banner")
	})
}

func TestCookieBanner_TenantIsolation(t *testing.T) {
	t.Parallel()

	t.Run("other org cannot access banner", func(t *testing.T) {
		t.Parallel()
		owner1 := testutil.NewClient(t, testutil.RoleOwner)
		owner2 := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner1)

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
				ID string `json:"id"`
			} `json:"node"`
		}

		err := owner2.Execute(query, map[string]any{"id": bannerID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "cookie banner")
	})
}
