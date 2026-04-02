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

package drivers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// OnePasswordUsersAPIDriver fetches user accounts from the 1Password
// Users API (v1beta1). This is distinct from the SCIM-based
// OnePasswordDriver and uses the native 1Password API with
// token-based pagination.
type OnePasswordUsersAPIDriver struct {
	httpClient *http.Client
	baseURL    string
	accountID  string
}

var _ Driver = (*OnePasswordUsersAPIDriver)(nil)

type onePasswordUsersAPIResponse struct {
	Users         []onePasswordUsersAPIUser `json:"users"`
	NextPageToken string                    `json:"next_page_token"`
}

type onePasswordUsersAPIUser struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	State       string `json:"state"`
	CreateTime  string `json:"create_time"`
	Path        string `json:"path"`
}

func NewOnePasswordUsersAPIDriver(httpClient *http.Client, accountID string, region string) *OnePasswordUsersAPIDriver {
	return &OnePasswordUsersAPIDriver{
		httpClient: httpClient,
		baseURL:    onePasswordBaseURL(region),
		accountID:  accountID,
	}
}

func onePasswordBaseURL(region string) string {
	switch region {
	case "CA", "ca":
		return "https://api.1password.ca"
	case "EU", "eu":
		return "https://api.1password.eu"
	default:
		return "https://api.1password.com"
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
			record := AccountRecord{
				Email:       u.Email,
				FullName:    u.DisplayName,
				Active:      u.State == "ACTIVE",
				ExternalID:  u.ID,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
			}

			if u.CreateTime != "" {
				if t, err := time.Parse(time.RFC3339, u.CreateTime); err == nil {
					record.CreatedAt = &t
				}
			}

			if record.Email != "" {
				records = append(records, record)
			}
		}

		if resp.NextPageToken == "" {
			return records, nil
		}
		pageToken = resp.NextPageToken
	}

	return nil, fmt.Errorf("cannot list all 1password users api accounts: %w", ErrPaginationLimitReached)
}

func (d *OnePasswordUsersAPIDriver) queryUsers(ctx context.Context, pageToken string) (*onePasswordUsersAPIResponse, error) {
	u, err := url.Parse(d.baseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse 1password users api base url: %w", err)
	}
	u = u.JoinPath("v1beta1", "accounts", d.accountID, "users")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create 1password users api request: %w", err)
	}

	q := req.URL.Query()
	q.Set("max_page_size", "100")
	if pageToken != "" {
		q.Set("page_token", pageToken)
	}
	req.URL.RawQuery = q.Encode()

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
