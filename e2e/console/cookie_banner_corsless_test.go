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
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestCookieBanner_Corsless(t *testing.T) {
	t.Parallel()

	t.Run(
		"rejects a second origin when corsless is off",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)
			foreignOrigin := factory.SafeOrigin()

			preflight := doCookieBannerHTTP(
				t,
				fixture.Owner,
				cookieBannerHTTPOptions{
					Method:                      http.MethodOptions,
					BannerID:                    fixture.BannerID,
					Path:                        []string{"consents"},
					Origin:                      foreignOrigin,
					AccessControlRequestMethod:  http.MethodPost,
					AccessControlRequestHeaders: "Content-Type, X-SDK-Version",
				},
			)
			assert.Equal(t, http.StatusForbidden, preflight.StatusCode)

			body, err := json.Marshal(
				postConsentRequest{
					VisitorID:   uniqueCookieBannerVisitorID(),
					Version:     fixture.Version,
					Action:      "ACCEPT_ALL",
					ConsentData: json.RawMessage(`{"necessary":true}`),
				},
			)
			require.NoError(t, err)

			resp := doCookieBannerHTTP(
				t,
				fixture.Owner,
				cookieBannerHTTPOptions{
					Method:     http.MethodPost,
					BannerID:   fixture.BannerID,
					Path:       []string{"consents"},
					Origin:     foreignOrigin,
					SDKVersion: cookieBannerE2ESDKVersion,
					Body:       body,
				},
			)
			assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		},
	)

	t.Run(
		"accepts a second origin and stores it when corsless is on",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)
			factory.EnableCookieBannerCorsless(t, fixture.BannerID)

			foreignOrigin := factory.SafeOrigin()
			visitorID := uniqueCookieBannerVisitorID()

			preflight := doCookieBannerHTTP(
				t,
				fixture.Owner,
				cookieBannerHTTPOptions{
					Method:                      http.MethodOptions,
					BannerID:                    fixture.BannerID,
					Path:                        []string{"consents"},
					Origin:                      foreignOrigin,
					AccessControlRequestMethod:  http.MethodPost,
					AccessControlRequestHeaders: "Content-Type, X-SDK-Version",
				},
			)
			require.Equal(t, http.StatusNoContent, preflight.StatusCode)
			assert.Equal(t, foreignOrigin, preflight.Header.Get("Access-Control-Allow-Origin"))

			body, err := json.Marshal(
				postConsentRequest{
					VisitorID:   visitorID,
					Version:     fixture.Version,
					Action:      "ACCEPT_ALL",
					ConsentData: json.RawMessage(`{"necessary":true}`),
				},
			)
			require.NoError(t, err)

			resp := doCookieBannerHTTP(
				t,
				fixture.Owner,
				cookieBannerHTTPOptions{
					Method:     http.MethodPost,
					BannerID:   fixture.BannerID,
					Path:       []string{"consents"},
					Origin:     foreignOrigin,
					SDKVersion: cookieBannerE2ESDKVersion,
					Body:       body,
				},
			)
			require.Equal(t, http.StatusCreated, resp.StatusCode, "body: %s", string(resp.Body))
			assert.Equal(t, foreignOrigin, resp.Header.Get("Access-Control-Allow-Origin"))

			var created postConsentResponseBody
			require.NoError(t, json.Unmarshal(resp.Body, &created))

			record := loadConsentRecord(t, fixture.Owner, created.ID)
			require.NotNil(t, record.Origin)
			assert.Equal(t, foreignOrigin, *record.Origin)
		},
	)

	t.Run(
		"stores the raw www origin on corsless consent records",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)
			factory.EnableCookieBannerCorsless(t, fixture.BannerID)

			rawOrigin := "https://www.customer.example"
			body, err := json.Marshal(
				postConsentRequest{
					VisitorID:   uniqueCookieBannerVisitorID(),
					Version:     fixture.Version,
					Action:      "ACCEPT_ALL",
					ConsentData: json.RawMessage(`{"necessary":true}`),
				},
			)
			require.NoError(t, err)

			resp := doCookieBannerHTTP(
				t,
				fixture.Owner,
				cookieBannerHTTPOptions{
					Method:     http.MethodPost,
					BannerID:   fixture.BannerID,
					Path:       []string{"consents"},
					Origin:     rawOrigin,
					SDKVersion: cookieBannerE2ESDKVersion,
					Body:       body,
				},
			)
			require.Equal(t, http.StatusCreated, resp.StatusCode, "body: %s", string(resp.Body))
			assert.Equal(t, rawOrigin, resp.Header.Get("Access-Control-Allow-Origin"))

			var created postConsentResponseBody
			require.NoError(t, json.Unmarshal(resp.Body, &created))

			record := loadConsentRecord(t, fixture.Owner, created.ID)
			require.NotNil(t, record.Origin)
			assert.Equal(t, rawOrigin, *record.Origin)
		},
	)

	t.Run(
		"rejects a null origin even when corsless is on",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)
			factory.EnableCookieBannerCorsless(t, fixture.BannerID)

			resp := doCookieBannerHTTP(
				t,
				fixture.Owner,
				cookieBannerHTTPOptions{
					Method:                      http.MethodOptions,
					BannerID:                    fixture.BannerID,
					Path:                        []string{"consents"},
					Origin:                      "null",
					AccessControlRequestMethod:  http.MethodPost,
					AccessControlRequestHeaders: "Content-Type, X-SDK-Version",
				},
			)
			assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		},
	)

	t.Run(
		"stores the request origin on a normal banner consent",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)
			created := postCookieConsent(
				t,
				fixture.Owner,
				fixture,
				uniqueCookieBannerVisitorID(),
				"ACCEPT_ALL",
				json.RawMessage(`{"necessary":true}`),
			)

			record := loadConsentRecord(t, fixture.Owner, created.ID)
			require.NotNil(t, record.Origin)
			assert.Equal(t, fixture.Origin, *record.Origin)
		},
	)

	t.Run(
		"reads corsless and leaves it unchanged when updating resource reporting",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)
			factory.EnableCookieBannerCorsless(t, fixture.BannerID)
			factory.EnableCookieBannerTCF(t, fixture.BannerID)

			const query = `
				mutation UpdateCookieBanner($input: UpdateCookieBannerInput!) {
					updateCookieBanner(input: $input) {
						cookieBanner {
							capabilities {
								resourceReporting
								tcf
								corsless
							}
						}
					}
				}
			`

			var result struct {
				UpdateCookieBanner struct {
					CookieBanner struct {
						Capabilities struct {
							ResourceReporting bool `json:"resourceReporting"`
							TCF               bool `json:"tcf"`
							Corsless          bool `json:"corsless"`
						} `json:"capabilities"`
					} `json:"cookieBanner"`
				} `json:"updateCookieBanner"`
			}

			err := fixture.Owner.Execute(
				query,
				map[string]any{
					"input": map[string]any{
						"cookieBannerId": fixture.BannerID,
						"capabilities":   map[string]any{"resourceReporting": false},
					},
				},
				&result,
			)
			require.NoError(t, err)
			assert.False(t, result.UpdateCookieBanner.CookieBanner.Capabilities.ResourceReporting)
			assert.True(t, result.UpdateCookieBanner.CookieBanner.Capabilities.TCF)
			assert.True(t, result.UpdateCookieBanner.CookieBanner.Capabilities.Corsless)
		},
	)

	t.Run(
		"uses the banner name in the generated policy when corsless",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)
			factory.EnableCookieBannerCorsless(t, fixture.BannerID)

			var banner struct {
				Node struct {
					Name   string `json:"name"`
					Origin string `json:"origin"`
				} `json:"node"`
			}
			require.NoError(
				t,
				fixture.Owner.Execute(
					`query($id: ID!) { node(id: $id) { ... on CookieBanner { name origin } } }`,
					map[string]any{"id": fixture.BannerID},
					&banner,
				),
			)

			var regen struct {
				RegenerateCookieBannerTrackerPolicy struct {
					CookieBanner struct {
						ID string `json:"id"`
					} `json:"cookieBanner"`
				} `json:"regenerateCookieBannerTrackerPolicy"`
			}
			require.NoError(
				t,
				fixture.Owner.Execute(
					regeneratePolicyMutation,
					map[string]any{
						"input": map[string]any{"cookieBannerId": fixture.BannerID},
					},
					&regen,
				),
			)

			type policyVersion struct {
				Title   string `json:"title"`
				Content string `json:"content"`
			}

			var version policyVersion

			require.Eventually(
				t,
				func() bool {
					var policy struct {
						Node struct {
							PolicyDocument *struct {
								Versions struct {
									Edges []struct {
										Node policyVersion `json:"node"`
									} `json:"edges"`
								} `json:"versions"`
							} `json:"policyDocument"`
						} `json:"node"`
					}
					if err := fixture.Owner.Execute(
						`
						query($id: ID!) {
							node(id: $id) {
								... on CookieBanner {
									policyDocument {
										versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
											edges { node { title content } }
										}
									}
								}
							}
						}
						`,
						map[string]any{"id": fixture.BannerID},
						&policy,
					); err != nil {
						return false
					}

					if policy.Node.PolicyDocument == nil || len(policy.Node.PolicyDocument.Versions.Edges) == 0 {
						return false
					}

					version = policy.Node.PolicyDocument.Versions.Edges[0].Node

					return strings.Contains(version.Title, banner.Node.Name)
				},
				15*time.Second,
				200*time.Millisecond,
			)
			assert.Contains(t, version.Title, banner.Node.Name)
			assert.NotContains(t, version.Title, banner.Node.Origin)
			assert.Contains(t, version.Content, banner.Node.Name)
			assert.NotContains(t, version.Content, banner.Node.Origin)
		},
	)
}

type consentRecordView struct {
	Origin *string `json:"origin"`
}

func loadConsentRecord(t *testing.T, owner *testutil.Client, recordID string) consentRecordView {
	t.Helper()

	var result struct {
		Node consentRecordView `json:"node"`
	}

	err := owner.Execute(
		`
		query($id: ID!) {
			node(id: $id) {
				... on CookieConsentRecord {
					origin
				}
			}
		}
		`,
		map[string]any{"id": recordID},
		&result,
	)
	require.NoError(t, err)

	return result.Node
}
