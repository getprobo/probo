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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const cookieBannerE2ESDKVersion = "e2e-cookie-banner-sdk/1.0.0"

type (
	cookieBannerHTTPOptions struct {
		Method      string
		Origin      string
		SDKVersion  string
		UserAgent   string
		Body        []byte
		Query       url.Values
		BannerID    string
		Path        []string
		ContentType string
	}

	cookieBannerHTTPResult struct {
		StatusCode int
		Header     http.Header
		Body       []byte
	}

	publishedCookieBannerFixture struct {
		Owner    *testutil.Client
		BannerID string
		Origin   string
		Version  int
	}
)

func uniqueCookieBannerVisitorID() string {
	return "visitor-" + strings.ToLower(factory.SafeName("v"))
}

func cookieBannerPublicURL(baseURL, bannerID string, path ...string) (string, error) {
	segments := append(
		[]string{"api", "cookie-banner", "v1", url.PathEscape(bannerID)},
		path...,
	)

	return url.JoinPath(baseURL, segments...)
}

func doCookieBannerHTTP(t *testing.T, c *testutil.Client, opts cookieBannerHTTPOptions) cookieBannerHTTPResult {
	t.Helper()

	if opts.Method == "" {
		opts.Method = http.MethodGet
	}

	if opts.BannerID == "" {
		t.Fatal("banner id is required")
	}

	endpoint, err := cookieBannerPublicURL(c.BaseURL(), opts.BannerID, opts.Path...)
	require.NoError(t, err)

	if len(opts.Query) > 0 {
		parsed, err := url.Parse(endpoint)
		require.NoError(t, err)

		parsed.RawQuery = opts.Query.Encode()
		endpoint = parsed.String()
	}

	var body io.Reader
	if len(opts.Body) > 0 {
		body = bytes.NewReader(opts.Body)
	}

	req, err := http.NewRequest(opts.Method, endpoint, body)
	require.NoError(t, err)

	if opts.Origin != "" {
		req.Header.Set("Origin", opts.Origin)
	}

	if opts.SDKVersion != "" {
		req.Header.Set("X-SDK-Version", opts.SDKVersion)
	}

	if opts.UserAgent != "" {
		req.Header.Set("User-Agent", opts.UserAgent)
	}

	if len(opts.Body) > 0 {
		contentType := opts.ContentType
		if contentType == "" {
			contentType = "application/json"
		}

		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.HTTPClient().Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return cookieBannerHTTPResult{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       respBody,
	}
}

func setupPublishedCookieBanner(t *testing.T) publishedCookieBannerFixture {
	t.Helper()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	origin := factory.SafeOrigin()

	bannerID := factory.CreateCookieBanner(owner, factory.Attrs{
		"origin":            origin,
		"cookiePolicyUrl":   "https://example.com/cookies",
		"consentExpiryDays": 365,
	})

	upsertTranslation(
		t,
		owner,
		bannerID,
		"en",
		`{"banner_title":"Cookies","banner_description":"We use cookies. See our {{cookie_policy_link}}."}`,
	)
	published := publishBanner(
		t,
		owner,
		bannerID,
	)
	require.Equal(t, "PUBLISHED", published.State)

	return publishedCookieBannerFixture{
		Owner:    owner,
		BannerID: bannerID,
		Origin:   origin,
		Version:  published.Version,
	}
}

func deactivateCookieBanner(t *testing.T, c *testutil.Client, bannerID string) {
	t.Helper()

	const query = `
		mutation DeactivateCookieBanner($input: DeactivateCookieBannerInput!) {
			deactivateCookieBanner(input: $input) {
				cookieBanner { id state }
			}
		}
	`

	var result struct{}
	require.NoError(t, c.Execute(query, map[string]any{
		"input": map[string]any{"cookieBannerId": bannerID},
	}, &result))
}

type postConsentRequest struct {
	VisitorID   string          `json:"visitor_id"`
	Version     int             `json:"version"`
	Action      string          `json:"action"`
	ConsentData json.RawMessage `json:"consent_data"`
}

type postConsentResponseBody struct {
	ID        string `json:"id"`
	VisitorID string `json:"visitor_id"`
	Action    string `json:"action"`
	CreatedAt string `json:"created_at"`
}

func postCookieConsent(
	t *testing.T,
	c *testutil.Client,
	fixture publishedCookieBannerFixture,
	visitorID string,
	action string,
	consentData json.RawMessage,
) postConsentResponseBody {
	t.Helper()

	body, err := json.Marshal(postConsentRequest{
		VisitorID:   visitorID,
		Version:     fixture.Version,
		Action:      action,
		ConsentData: consentData,
	})
	require.NoError(t, err)

	const ua = "Probo-CookieBanner-E2E/1.0"

	resp := doCookieBannerHTTP(
		t,
		c,
		cookieBannerHTTPOptions{
			Method:     http.MethodPost,
			BannerID:   fixture.BannerID,
			Path:       []string{"consents"},
			Origin:     fixture.Origin,
			SDKVersion: cookieBannerE2ESDKVersion,
			UserAgent:  ua,
			Body:       body,
		},
	)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "body: %s", string(resp.Body))

	var parsed postConsentResponseBody
	require.NoError(t, json.Unmarshal(resp.Body, &parsed))

	return parsed
}
