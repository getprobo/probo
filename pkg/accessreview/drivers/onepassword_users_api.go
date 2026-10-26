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

package drivers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

// 1Password serves each region from its own host, and the OAuth token
// endpoint sits on the same host as the Users API: a token one region issues
// is refused by the others.
const (
	onePasswordRegionUS  = "US"
	onePasswordRegionCA  = "CA"
	onePasswordRegionEU  = "EU"
	onePasswordRegionENT = "ENT"

	onePasswordUSHost  = "api.1password.com"
	onePasswordCAHost  = "api.1password.ca"
	onePasswordEUHost  = "api.1password.eu"
	onePasswordENTHost = "api.ent.1password.com"

	onePasswordTokenPath = "/v1/oauth/token"

	// onePasswordUsersPageSize is the largest page the Users API serves; every
	// page counts against its limit of 100 requests a minute per account.
	onePasswordUsersPageSize = 1000
)

// OnePasswordUsersAPIDriver fetches user accounts from the 1Password
// Users API (v1). This is distinct from the SCIM-based
// OnePasswordDriver and uses the native 1Password API with
// token-based pagination.
type OnePasswordUsersAPIDriver struct {
	httpClient *http.Client
	baseURL    string
	accountID  string
}

var _ Driver = (*OnePasswordUsersAPIDriver)(nil)

type onePasswordUsersAPIResponse struct {
	Users         []onePasswordUsersAPIUser `json:"results"`
	NextPageToken string                    `json:"next_page_token"`
}

type onePasswordUsersAPIUser struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	State       string `json:"state"`
	CreateTime  string `json:"create_time"`
}

// NewOnePasswordUsersAPIDriver builds a driver against baseURL, the Users API
// root of the connector's region (OnePasswordUsersAPIBaseURL).
func NewOnePasswordUsersAPIDriver(httpClient *http.Client, baseURL string, accountID string) *OnePasswordUsersAPIDriver {
	client := *httpClient
	client.Transport = &retryRoundTripper{
		next:       httpClient.Transport,
		maxRetries: 3,
	}

	return &OnePasswordUsersAPIDriver{
		httpClient: &client,
		baseURL:    baseURL,
		accountID:  accountID,
	}
}

// IsOnePasswordRegion reports whether region names one of 1Password's hosts,
// in the form the connector stores it.
func IsOnePasswordRegion(region string) bool {
	_, ok := onePasswordRegionHost(region)

	return ok
}

// OnePasswordUsersAPIBaseURL returns the Users API root of region.
func OnePasswordUsersAPIBaseURL(region string) string {
	u := url.URL{Scheme: "https", Host: onePasswordHost(region)}

	return u.String()
}

// OnePasswordTokenURL returns the OAuth token endpoint of region.
func OnePasswordTokenURL(region string) string {
	u := url.URL{Scheme: "https", Host: onePasswordHost(region), Path: onePasswordTokenPath}

	return u.String()
}

// OnePasswordUsersURL returns the list-users endpoint of accountID under
// baseURL: the request both the driver and the connection check make.
func OnePasswordUsersURL(baseURL string, accountID string) (*url.URL, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse 1password users api base url: %w", err)
	}

	return base.JoinPath("v1", "accounts", url.PathEscape(accountID), "users"), nil
}

// onePasswordHost maps the connector's Region setting to its host. The region
// was free text before it became a fixed choice, and those rows have always
// reached the US host, so anything unrecognised still does.
func onePasswordHost(region string) string {
	if host, ok := onePasswordRegionHost(strings.ToUpper(strings.TrimSpace(region))); ok {
		return host
	}

	return onePasswordUSHost
}

// onePasswordRegionHost returns the host of a region in its stored form.
func onePasswordRegionHost(region string) (string, bool) {
	switch region {
	case onePasswordRegionUS:
		return onePasswordUSHost, true
	case onePasswordRegionCA:
		return onePasswordCAHost, true
	case onePasswordRegionEU:
		return onePasswordEUHost, true
	case onePasswordRegionENT:
		return onePasswordENTHost, true
	default:
		return "", false
	}
}

// onePasswordUserActive maps a user's state. 1Password lists only ACTIVE and
// SUSPENDED users but may add states, so anything else is unknown rather than
// inactive.
func onePasswordUserActive(state string) *bool {
	switch state {
	case "ACTIVE":
		return new(true)
	case "SUSPENDED":
		return new(false)
	default:
		return nil
	}
}

func (d *OnePasswordUsersAPIDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records   []AccountRecord
		pageToken string
	)

	for range maxPaginationPages {
		resp, err := d.queryUsers(ctx, pageToken)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Users {
			if u.Email == "" {
				continue
			}

			records = append(
				records,
				AccountRecord{
					Email:       u.Email,
					FullName:    u.DisplayName,
					Active:      onePasswordUserActive(u.State),
					ExternalID:  u.ID,
					MFAStatus:   coredata.MFAStatusUnknown,
					AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
					AccountType: coredata.AccessReviewEntryAccountTypeUser,
					CreatedAt:   parseRFC3339Ptr(u.CreateTime),
				},
			)
		}

		if resp.NextPageToken == "" {
			return records, nil
		}

		pageToken = resp.NextPageToken
	}

	return nil, fmt.Errorf("cannot list all 1password users api accounts: %w", ErrPaginationLimitReached)
}

func (d *OnePasswordUsersAPIDriver) queryUsers(ctx context.Context, pageToken string) (*onePasswordUsersAPIResponse, error) {
	u, err := OnePasswordUsersURL(d.baseURL, d.accountID)
	if err != nil {
		return nil, fmt.Errorf("cannot build 1password users api request url: %w", err)
	}

	q := url.Values{}
	q.Set("max_page_size", strconv.Itoa(onePasswordUsersPageSize))

	if pageToken != "" {
		q.Set("page_token", pageToken)
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create 1password users api request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute 1password users api request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch 1password users api: unexpected status %d", httpResp.StatusCode)
	}

	var resp onePasswordUsersAPIResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode 1password users api response: %w", err)
	}

	return &resp, nil
}
