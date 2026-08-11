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
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestCookieConsent_PublicAPIAndConsole(t *testing.T) {
	t.Parallel()

	fixture := setupPublishedCookieBanner(t)
	owner := fixture.Owner
	visitorA := uniqueCookieBannerVisitorID()
	visitorB := uniqueCookieBannerVisitorID()
	consentDataAccept := json.RawMessage(`{"necessary":true,"analytics":true}`)
	consentDataReject := json.RawMessage(`{"necessary":true,"analytics":false}`)

	configResp := doCookieBannerHTTP(
		t,
		owner,
		cookieBannerHTTPOptions{
			Method:     http.MethodGet,
			BannerID:   fixture.BannerID,
			Path:       []string{"config"},
			Origin:     fixture.Origin,
			SDKVersion: cookieBannerE2ESDKVersion,
			Query:      url.Values{"lang": []string{"en"}},
		},
	)
	require.Equal(t, http.StatusOK, configResp.StatusCode, "config body: %s", string(configResp.Body))
	assert.Equal(t, fixture.Origin, configResp.Header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, configResp.Header.Get("Vary"), "Origin")

	var config struct {
		BannerID                 string `json:"banner_id"`
		Version                  int    `json:"version"`
		Language                 string `json:"language"`
		DefaultLanguage          string `json:"default_language"`
		CookiePolicyURL          string `json:"cookie_policy_url"`
		ConsentExpiryDays        int    `json:"consent_expiry_days"`
		ConsentMode              string `json:"consent_mode"`
		Regulation               string `json:"regulation"`
		ShowBranding             bool   `json:"show_branding"`
		ResourceReportingEnabled bool   `json:"resource_reporting_enabled"`
		Layout                   struct {
			Presentation string `json:"presentation"`
			InitialState string `json:"initial_state"`
		} `json:"layout"`
		Categories []struct {
			Slug string `json:"slug"`
			Kind string `json:"kind"`
			Name string `json:"name"`
		} `json:"categories"`
		Texts map[string]string `json:"texts"`
	}
	require.NoError(t, json.Unmarshal(configResp.Body, &config))
	assert.Equal(t, fixture.BannerID, config.BannerID)
	assert.Equal(t, fixture.Version, config.Version)
	assert.Equal(t, "en", config.Language)
	assert.NotEmpty(t, config.DefaultLanguage)
	assert.Equal(t, "https://example.com/cookies", config.CookiePolicyURL)
	assert.Equal(t, 365, config.ConsentExpiryDays)
	assert.NotEmpty(t, config.ConsentMode)
	assert.NotEmpty(t, config.Regulation)
	assert.True(t, config.ResourceReportingEnabled)
	assert.NotEmpty(t, config.Layout.Presentation)
	assert.NotEmpty(t, config.Layout.InitialState)
	assert.NotEmpty(t, config.Categories)
	assert.NotEmpty(t, config.Texts["banner_title"])

	created := postCookieConsent(
		t,
		owner,
		fixture,
		visitorA,
		"ACCEPT_ALL",
		consentDataAccept,
	)
	assert.Equal(t, visitorA, created.VisitorID)
	assert.Equal(t, "ACCEPT_ALL", created.Action)
	assert.NotEmpty(t, created.ID)
	assert.NotEmpty(t, created.CreatedAt)

	getResp := doCookieBannerHTTP(
		t,
		owner,
		cookieBannerHTTPOptions{
			Method:     http.MethodGet,
			BannerID:   fixture.BannerID,
			Path:       []string{"consents", url.PathEscape(visitorA)},
			Origin:     fixture.Origin,
			SDKVersion: cookieBannerE2ESDKVersion,
		},
	)
	require.Equal(t, http.StatusOK, getResp.StatusCode, "body: %s", string(getResp.Body))

	var visitorConsent struct {
		VisitorID   string          `json:"visitor_id"`
		Version     int             `json:"version"`
		Action      string          `json:"action"`
		ConsentData json.RawMessage `json:"consent_data"`
		CreatedAt   string          `json:"created_at"`
	}
	require.NoError(t, json.Unmarshal(getResp.Body, &visitorConsent))
	assert.Equal(t, visitorA, visitorConsent.VisitorID)
	assert.Equal(t, fixture.Version, visitorConsent.Version)
	assert.Equal(t, "ACCEPT_ALL", visitorConsent.Action)
	assert.JSONEq(t, string(consentDataAccept), string(visitorConsent.ConsentData))

	missingResp := doCookieBannerHTTP(
		t,
		owner,
		cookieBannerHTTPOptions{
			Method:     http.MethodGet,
			BannerID:   fixture.BannerID,
			Path:       []string{"consents", url.PathEscape(uniqueCookieBannerVisitorID())},
			Origin:     fixture.Origin,
			SDKVersion: cookieBannerE2ESDKVersion,
		},
	)
	assert.Equal(t, http.StatusNotFound, missingResp.StatusCode)

	rejectRecord := postCookieConsent(
		t,
		owner,
		fixture,
		visitorA,
		"REJECT_ALL",
		consentDataReject,
	)
	customizeRecord := postCookieConsent(
		t,
		owner,
		fixture,
		visitorB,
		"CUSTOMIZE",
		consentDataReject,
	)

	latestResp := doCookieBannerHTTP(
		t,
		owner,
		cookieBannerHTTPOptions{
			Method:     http.MethodGet,
			BannerID:   fixture.BannerID,
			Path:       []string{"consents", url.PathEscape(visitorA)},
			Origin:     fixture.Origin,
			SDKVersion: cookieBannerE2ESDKVersion,
		},
	)
	require.Equal(t, http.StatusOK, latestResp.StatusCode)
	require.NoError(t, json.Unmarshal(latestResp.Body, &visitorConsent))
	assert.Equal(t, "REJECT_ALL", visitorConsent.Action)

	const consentRecordsQuery = `
		query CookieBannerConsentRecords($id: ID!, $filter: CookieConsentRecordFilter) {
			node(id: $id) {
				... on CookieBanner {
					canListConsentRecords: permission(action: "core:cookie-consent-record:list")
					consentRecords(
						first: 20
						orderBy: { field: CREATED_AT, direction: DESC }
						filter: $filter
					) {
						totalCount
						edges {
							cursor
							node {
								id
								visitorId
								action
								consentData
								sdkVersion
								userAgent
								regulation
								regulationSource
								cookieBannerVersion { version }
								createdAt
							}
						}
						pageInfo {
							hasNextPage
							hasPreviousPage
							startCursor
							endCursor
						}
					}
				}
			}
		}
	`

	var allRecords struct {
		Node struct {
			CanListConsentRecords bool `json:"canListConsentRecords"`
			ConsentRecords        struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Cursor string `json:"cursor"`
					Node   struct {
						ID                  string  `json:"id"`
						VisitorID           string  `json:"visitorId"`
						Action              string  `json:"action"`
						ConsentData         string  `json:"consentData"`
						SdkVersion          string  `json:"sdkVersion"`
						UserAgent           *string `json:"userAgent"`
						Regulation          *string `json:"regulation"`
						RegulationSource    *string `json:"regulationSource"`
						CookieBannerVersion struct {
							Version int `json:"version"`
						} `json:"cookieBannerVersion"`
						CreatedAt string `json:"createdAt"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					HasNextPage     bool    `json:"hasNextPage"`
					HasPreviousPage bool    `json:"hasPreviousPage"`
					StartCursor     *string `json:"startCursor"`
					EndCursor       *string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"consentRecords"`
		} `json:"node"`
	}

	require.NoError(t, owner.Execute(consentRecordsQuery, map[string]any{
		"id":     fixture.BannerID,
		"filter": nil,
	}, &allRecords))
	assert.True(t, allRecords.Node.CanListConsentRecords)
	consentConn := allRecords.Node.ConsentRecords
	assert.GreaterOrEqual(t, consentConn.TotalCount, 3)
	require.NotEmpty(t, consentConn.Edges)
	assert.LessOrEqual(t, len(consentConn.Edges), 20)
	assert.NotEmpty(t, consentConn.Edges[0].Cursor)
	require.NotNil(t, consentConn.PageInfo.StartCursor)
	require.NotNil(t, consentConn.PageInfo.EndCursor)

	listedIDs := make(map[string]struct{}, len(consentConn.Edges))
	for _, edge := range consentConn.Edges {
		listedIDs[edge.Node.ID] = struct{}{}
	}

	require.Contains(t, listedIDs, created.ID)
	require.Contains(t, listedIDs, rejectRecord.ID)
	require.Contains(t, listedIDs, customizeRecord.ID)

	var foundCreated bool

	for _, edge := range consentConn.Edges {
		node := edge.Node
		if node.ID != created.ID {
			continue
		}

		foundCreated = true

		assert.Equal(t, visitorA, node.VisitorID)
		assert.Equal(t, cookieBannerE2ESDKVersion, node.SdkVersion)
		require.NotNil(t, node.UserAgent)
		assert.Contains(t, *node.UserAgent, "Probo-CookieBanner-E2E")
		assert.Equal(t, fixture.Version, node.CookieBannerVersion.Version)
		assert.NotEmpty(t, node.CreatedAt)
	}

	assert.True(t, foundCreated, "expected consent record %s in list", created.ID)

	var filtered struct {
		Node struct {
			ConsentRecords struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						VisitorID string `json:"visitorId"`
						Action    string `json:"action"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"consentRecords"`
		} `json:"node"`
	}

	require.NoError(t, owner.Execute(consentRecordsQuery, map[string]any{
		"id": fixture.BannerID,
		"filter": map[string]any{
			"visitorId": visitorB,
			"action":    "CUSTOMIZE",
		},
	}, &filtered))
	assert.Equal(t, 1, filtered.Node.ConsentRecords.TotalCount)
	require.Len(t, filtered.Node.ConsentRecords.Edges, 1)
	assert.Equal(t, visitorB, filtered.Node.ConsentRecords.Edges[0].Node.VisitorID)
	assert.Equal(t, "CUSTOMIZE", filtered.Node.ConsentRecords.Edges[0].Node.Action)

	const recordNodeQuery = `
		query CookieConsentRecordNode($id: ID!) {
			node(id: $id) {
				... on CookieConsentRecord {
					id
					visitorId
					action
					consentData
					sdkVersion
					cookieBanner { id }
					cookieBannerVersion { version }
				}
			}
		}
	`

	var recordNode struct {
		Node *struct {
			ID           string `json:"id"`
			VisitorID    string `json:"visitorId"`
			Action       string `json:"action"`
			SdkVersion   string `json:"sdkVersion"`
			CookieBanner struct {
				ID string `json:"id"`
			} `json:"cookieBanner"`
			CookieBannerVersion struct {
				Version int `json:"version"`
			} `json:"cookieBannerVersion"`
		} `json:"node"`
	}

	require.NoError(t, owner.Execute(recordNodeQuery, map[string]any{"id": created.ID}, &recordNode))
	require.NotNil(t, recordNode.Node)
	assert.Equal(t, visitorA, recordNode.Node.VisitorID)
	assert.Equal(t, fixture.BannerID, recordNode.Node.CookieBanner.ID)
	assert.Equal(t, fixture.Version, recordNode.Node.CookieBannerVersion.Version)

	t.Run(
		"read-only config and consent roundtrip",
		func(t *testing.T) {
			t.Parallel()

			resp := doCookieBannerHTTP(
				t,
				owner,
				cookieBannerHTTPOptions{
					Method:     http.MethodGet,
					BannerID:   fixture.BannerID,
					Path:       []string{"consents", url.PathEscape(visitorB)},
					Origin:     fixture.Origin,
					SDKVersion: cookieBannerE2ESDKVersion,
				},
			)
			require.Equal(t, http.StatusOK, resp.StatusCode)
		},
	)
}

func TestCookieConsent_PublicValidationAndSecurity(t *testing.T) {
	t.Parallel()

	t.Run(
		"malformed banner id returns 400",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)

			resp := doCookieBannerHTTP(
				t,
				owner,
				cookieBannerHTTPOptions{
					Method:   http.MethodGet,
					BannerID: "not-a-valid-gid",
					Path:     []string{"config"},
				},
			)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		},
	)

	t.Run(
		"inactive banner returns 404",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)
			deactivateCookieBanner(t, fixture.Owner, fixture.BannerID)

			resp := doCookieBannerHTTP(
				t,
				fixture.Owner,
				cookieBannerHTTPOptions{
					Method:   http.MethodGet,
					BannerID: fixture.BannerID,
					Path:     []string{"config"},
				},
			)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		},
	)

	t.Run(
		"disallowed origin returns 403",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)

			resp := doCookieBannerHTTP(
				t,
				fixture.Owner,
				cookieBannerHTTPOptions{
					Method:   http.MethodGet,
					BannerID: fixture.BannerID,
					Path:     []string{"config"},
					Origin:   factory.SafeOrigin(),
				},
			)
			assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		},
	)

	t.Run(
		"invalid JSON body returns 400",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)

			resp := doCookieBannerHTTP(
				t,
				fixture.Owner,
				cookieBannerHTTPOptions{
					Method:   http.MethodPost,
					BannerID: fixture.BannerID,
					Path:     []string{"consents"},
					Origin:   fixture.Origin,
					Body:     []byte(`{`),
				},
			)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		},
	)

	t.Run(
		"unpublished version returns 400",
		func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			origin := factory.SafeOrigin()
			bannerID := factory.CreateCookieBanner(owner, factory.Attrs{"origin": origin})
			// Banner ships with a draft only; never published.

			body, err := json.Marshal(postConsentRequest{
				VisitorID:   uniqueCookieBannerVisitorID(),
				Version:     1,
				Action:      "ACCEPT_ALL",
				ConsentData: json.RawMessage(`{"necessary":true}`),
			})
			require.NoError(t, err)

			resp := doCookieBannerHTTP(
				t,
				owner,
				cookieBannerHTTPOptions{
					Method:   http.MethodPost,
					BannerID: bannerID,
					Path:     []string{"consents"},
					Origin:   origin,
					Body:     body,
				},
			)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		},
	)

	t.Run(
		"unknown published version returns 400",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)

			body, err := json.Marshal(postConsentRequest{
				VisitorID:   uniqueCookieBannerVisitorID(),
				Version:     fixture.Version + 1000,
				Action:      "ACCEPT_ALL",
				ConsentData: json.RawMessage(`{"necessary":true}`),
			})
			require.NoError(t, err)

			resp := doCookieBannerHTTP(
				t,
				fixture.Owner,
				cookieBannerHTTPOptions{
					Method:   http.MethodPost,
					BannerID: fixture.BannerID,
					Path:     []string{"consents"},
					Origin:   fixture.Origin,
					Body:     body,
				},
			)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		},
	)

	t.Run(
		"invalid consent action returns 400",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)

			body, err := json.Marshal(postConsentRequest{
				VisitorID:   uniqueCookieBannerVisitorID(),
				Version:     fixture.Version,
				Action:      "NOT_A_REAL_ACTION",
				ConsentData: json.RawMessage(`{"necessary":true}`),
			})
			require.NoError(t, err)

			resp := doCookieBannerHTTP(
				t,
				fixture.Owner,
				cookieBannerHTTPOptions{
					Method:   http.MethodPost,
					BannerID: fixture.BannerID,
					Path:     []string{"consents"},
					Origin:   fixture.Origin,
					Body:     body,
				},
			)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		},
	)

	t.Run(
		"empty visitor id returns 400",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)

			body, err := json.Marshal(postConsentRequest{
				VisitorID:   "",
				Version:     fixture.Version,
				Action:      "ACCEPT_ALL",
				ConsentData: json.RawMessage(`{"necessary":true}`),
			})
			require.NoError(t, err)

			resp := doCookieBannerHTTP(
				t,
				fixture.Owner,
				cookieBannerHTTPOptions{
					Method:   http.MethodPost,
					BannerID: fixture.BannerID,
					Path:     []string{"consents"},
					Origin:   fixture.Origin,
					Body:     body,
				},
			)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		},
	)

	t.Run(
		"CORS preflight succeeds for allowed origin",
		func(t *testing.T) {
			t.Parallel()

			fixture := setupPublishedCookieBanner(t)

			resp := doCookieBannerHTTP(
				t,
				fixture.Owner,
				cookieBannerHTTPOptions{
					Method:                      http.MethodOptions,
					BannerID:                    fixture.BannerID,
					Path:                        []string{"consents"},
					Origin:                      fixture.Origin,
					AccessControlRequestMethod:  http.MethodPost,
					AccessControlRequestHeaders: "Content-Type, X-SDK-Version",
				},
			)
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
			assert.Equal(t, fixture.Origin, resp.Header.Get("Access-Control-Allow-Origin"))
			assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), http.MethodPost)
			assert.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "Content-Type")
			assert.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "X-SDK-Version")
		},
	)
}
