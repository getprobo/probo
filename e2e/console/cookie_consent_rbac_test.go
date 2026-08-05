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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestCookieConsent_RBAC(t *testing.T) {
	t.Parallel()

	t.Run(
		"viewer can list consent records",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, fixture.Owner)

			visitorID := uniqueCookieBannerVisitorID()
			created := postCookieConsent(
				t,
				fixture.Owner,
				fixture,
				visitorID,
				"ACCEPT_ALL",
				json.RawMessage(`{"necessary":true}`),
			)

			const query = `
				query($id: ID!) {
					node(id: $id) {
						... on CookieBanner {
							canList: permission(action: "core:cookie-consent-record:list")
							consentRecords(first: 5) {
								totalCount
								edges { node { id visitorId } }
							}
						}
					}
				}
			`

			var result struct {
				Node struct {
					CanList        bool `json:"canList"`
					ConsentRecords struct {
						TotalCount int `json:"totalCount"`
						Edges      []struct {
							Node struct {
								ID        string `json:"id"`
								VisitorID string `json:"visitorId"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"consentRecords"`
				} `json:"node"`
			}

			require.NoError(t, viewer.Execute(query, map[string]any{"id": fixture.BannerID}, &result))
			assert.True(t, result.Node.CanList)
			assert.GreaterOrEqual(t, result.Node.ConsentRecords.TotalCount, 1)

			var found bool

			for _, edge := range result.Node.ConsentRecords.Edges {
				if edge.Node.ID == created.ID {
					found = true

					assert.Equal(t, visitorID, edge.Node.VisitorID)
				}
			}

			assert.True(t, found)
		},
	)

	t.Run(
		"viewer can load consent record node",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, fixture.Owner)

			created := postCookieConsent(
				t,
				fixture.Owner,
				fixture,
				uniqueCookieBannerVisitorID(),
				"REJECT_ALL",
				json.RawMessage(`{"necessary":true,"analytics":false}`),
			)

			const query = `
				query($id: ID!) {
					node(id: $id) {
						... on CookieConsentRecord {
							id
							action
							visitorId
						}
					}
				}
			`

			var result struct {
				Node *struct {
					ID        string `json:"id"`
					Action    string `json:"action"`
					VisitorID string `json:"visitorId"`
				} `json:"node"`
			}

			require.NoError(t, viewer.Execute(query, map[string]any{"id": created.ID}, &result))
			require.NotNil(t, result.Node)
			assert.Equal(t, created.ID, result.Node.ID)
			assert.Equal(t, "REJECT_ALL", result.Node.Action)
		},
	)

	t.Run(
		"other organization cannot list consent records",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)
			otherOwner := testutil.NewClient(t, testutil.RoleOwner)

			postCookieConsent(
				t,
				fixture.Owner,
				fixture,
				uniqueCookieBannerVisitorID(),
				"ACCEPT_ALL",
				json.RawMessage(`{"necessary":true}`),
			)

			const query = `
				query($id: ID!) {
					node(id: $id) {
						... on CookieBanner {
							consentRecords(first: 5) {
								totalCount
							}
						}
					}
				}
			`

			var result struct {
				Node *struct {
					ConsentRecords struct {
						TotalCount int `json:"totalCount"`
					} `json:"consentRecords"`
				} `json:"node"`
			}

			err := otherOwner.Execute(query, map[string]any{"id": fixture.BannerID}, &result)
			testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "cookie banner consent records")
		},
	)
}
