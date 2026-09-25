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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
)

func TestCookieBannerReport_BeaconPayload(t *testing.T) {
	t.Parallel()

	fixture := setupPublishedCookieBanner(t)
	cookieName := strings.ReplaceAll(factory.SafeName("beacon-cookie"), " ", "_")

	body, err := json.Marshal(
		map[string]any{
			"cookies": []map[string]any{
				{
					"name":   cookieName,
					"source": "script",
				},
			},
		},
	)
	require.NoError(t, err)

	resp := doCookieBannerHTTP(
		t,
		fixture.Owner,
		cookieBannerHTTPOptions{
			Method:      http.MethodPost,
			BannerID:    fixture.BannerID,
			Path:        []string{"report"},
			Origin:      fixture.Origin,
			ContentType: "text/plain;charset=UTF-8",
			Body:        body,
		},
	)
	require.Equal(t, http.StatusNoContent, resp.StatusCode, "report body: %s", string(resp.Body))
	assert.Equal(t, fixture.Origin, resp.Header.Get("Access-Control-Allow-Origin"))

	const query = `
		query CookieBannerReportBeaconPayload($id: ID!, $filter: TrackerPatternFilter) {
			node(id: $id) {
				... on CookieBanner {
					trackerPatterns(first: 10, filter: $filter) {
						totalCount
						edges {
							node {
								pattern
								trackerType
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			TrackerPatterns struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						Pattern     string `json:"pattern"`
						TrackerType string `json:"trackerType"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"trackerPatterns"`
		} `json:"node"`
	}

	require.NoError(
		t,
		fixture.Owner.Execute(
			query,
			map[string]any{
				"id": fixture.BannerID,
				"filter": map[string]any{
					"query":       cookieName,
					"trackerType": "COOKIE",
				},
			},
			&result,
		),
	)
	assert.Equal(t, 1, result.Node.TrackerPatterns.TotalCount)
	require.Len(t, result.Node.TrackerPatterns.Edges, 1)
	assert.Equal(t, cookieName, result.Node.TrackerPatterns.Edges[0].Node.Pattern)
	assert.Equal(t, "COOKIE", result.Node.TrackerPatterns.Edges[0].Node.TrackerType)
}
