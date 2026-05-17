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
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// RampDriver fetches users from the Ramp Developer API using a
// pre-authenticated HTTP client (Bearer token). Ramp grants are scoped
// to a single business — there is no per-business picker, so this is a
// Pattern 1 driver. Pagination is via the absolute URL exposed in
// `page.next` on the response body.
type RampDriver struct {
	httpClient *http.Client
}

var _ Driver = (*RampDriver)(nil)

func NewRampDriver(httpClient *http.Client) *RampDriver {
	return &RampDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
	}
}

type rampUser struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	LastLoginAt string `json:"last_login_at"`
	IsManager   bool   `json:"is_manager"`
}

type rampUsersPage struct {
	Data []rampUser `json:"data"`
	Page struct {
		Next string `json:"next"`
	} `json:"page"`
}

func (d *RampDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	next := "https://api.ramp.com/developer/v1/users?page_size=100"

	for range maxPaginationPages {
		page, err := d.queryUsers(ctx, next)
		if err != nil {
			return nil, err
		}

		for _, u := range page.Data {
			fullName := strings.TrimSpace(u.FirstName + " " + u.LastName)

			active := u.Status == "USER_ACTIVE"

			record := AccountRecord{
				Email:       u.Email,
				FullName:    fullName,
				Role:        u.Role,
				Active:      &active,
				IsAdmin:     u.IsManager,
				ExternalID:  u.ID,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
			}

			if u.LastLoginAt != "" {
				if t, err := time.Parse(time.RFC3339, u.LastLoginAt); err == nil {
					record.LastLogin = &t
				}
			}

			records = append(records, record)
		}

		if page.Page.Next == "" {
			return records, nil
		}
		next = page.Page.Next
	}

	return nil, fmt.Errorf("cannot list all ramp accounts: %w", ErrPaginationLimitReached)
}

func (d *RampDriver) queryUsers(ctx context.Context, endpoint string) (*rampUsersPage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create ramp users request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute ramp users request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch ramp users: unexpected status %d", httpResp.StatusCode)
	}

	var page rampUsersPage
	if err := json.NewDecoder(httpResp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("cannot decode ramp users response: %w", err)
	}

	return &page, nil
}
