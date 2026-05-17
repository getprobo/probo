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

// LeverDriver fetches users from the Lever REST API using a
// pre-authenticated HTTP client (Bearer token from the Auth0-backed
// flow). Pagination is body-cursor based: response carries `data[]`,
// `hasNext` (bool), and `next` (cursor). The next request appends
// `?offset=<cursor>`.
//
// Notes on data quality:
//   - `lastLoggedInAt` and `createdAt` are epoch milliseconds —
//     defensive parse, may be null/undocumented.
//   - MFA is exposed only via SCIM/SSO, not the REST API.
//   - Active is derived from `deactivatedAt`: nil/missing = active.
type LeverDriver struct {
	httpClient *http.Client
}

var _ Driver = (*LeverDriver)(nil)

func NewLeverDriver(httpClient *http.Client) *LeverDriver {
	return &LeverDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
	}
}

type leverUser struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	AccessRole     string `json:"accessRole"`
	DeactivatedAt  *int64 `json:"deactivatedAt"`
	LastLoggedInAt *int64 `json:"lastLoggedInAt"`
	CreatedAt      *int64 `json:"createdAt"`
}

type leverUsersPage struct {
	Data    []leverUser `json:"data"`
	HasNext bool        `json:"hasNext"`
	Next    string      `json:"next"`
}

func (d *LeverDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	cursor := ""
	for range maxPaginationPages {
		page, err := d.queryUsers(ctx, cursor)
		if err != nil {
			return nil, err
		}

		for _, u := range page.Data {
			active := u.DeactivatedAt == nil

			record := AccountRecord{
				Email:       u.Email,
				FullName:    u.Name,
				Role:        u.AccessRole,
				Active:      &active,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
				ExternalID:  u.ID,
			}

			if u.LastLoggedInAt != nil {
				t := time.UnixMilli(*u.LastLoggedInAt)
				record.LastLogin = &t
			}

			if u.CreatedAt != nil {
				t := time.UnixMilli(*u.CreatedAt)
				record.CreatedAt = &t
			}

			records = append(records, record)
		}

		if !page.HasNext || page.Next == "" {
			return records, nil
		}
		cursor = page.Next
	}

	return nil, fmt.Errorf("cannot list all lever accounts: %w", ErrPaginationLimitReached)
}

func (d *LeverDriver) queryUsers(ctx context.Context, cursor string) (*leverUsersPage, error) {
	q := url.Values{}
	q.Set("limit", "100")
	if cursor != "" {
		q.Set("offset", cursor)
	}
	endpoint := "https://api.lever.co/v1/users?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create lever users request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute lever users request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch lever users: unexpected status %d", httpResp.StatusCode)
	}

	var page leverUsersPage
	if err := json.NewDecoder(httpResp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("cannot decode lever users response: %w", err)
	}

	return &page, nil
}
