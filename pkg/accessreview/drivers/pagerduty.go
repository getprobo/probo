// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"strconv"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// PagerDutyDriver fetches users from the PagerDuty REST API using a
// pre-authenticated HTTP client (Bearer token from the Scoped OAuth
// PKCE flow). Pagination is offset / limit based.
type PagerDutyDriver struct {
	httpClient *http.Client
}

var _ Driver = (*PagerDutyDriver)(nil)

func NewPagerDutyDriver(httpClient *http.Client) *PagerDutyDriver {
	return &PagerDutyDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
	}
}

type pagerdutyUser struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	InvitationSent bool   `json:"invitation_sent"`
	CreatedAt      string `json:"created_at"`
}

type pagerdutyUsersPage struct {
	Users []pagerdutyUser `json:"users"`
	More  bool            `json:"more"`
	Limit int             `json:"limit"`
}

func (d *PagerDutyDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	const limit = 100

	offset := 0

	for range maxPaginationPages {
		page, err := d.queryUsers(ctx, offset, limit)
		if err != nil {
			return nil, err
		}

		for _, u := range page.Users {
			isAdmin := u.Role == "admin" || u.Role == "owner"

			record := AccountRecord{
				Email:       u.Email,
				FullName:    u.Name,
				Role:        u.Role,
				IsAdmin:     isAdmin,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  u.ID,
			}

			// invitation_sent=true means the invitation is still pending,
			// so the account is not yet active. Once accepted the field
			// flips to false; we cannot tell active-vs-deactivated apart
			// in that case and leave Active nil.
			if u.InvitationSent {
				active := false
				record.Active = &active
			}

			if u.CreatedAt != "" {
				if t, err := time.Parse(time.RFC3339, u.CreatedAt); err == nil {
					record.CreatedAt = &t
				}
			}

			records = append(records, record)
		}

		if !page.More {
			return records, nil
		}

		pageSize := page.Limit
		if pageSize <= 0 {
			pageSize = limit
		}

		offset += pageSize
	}

	return nil, fmt.Errorf("cannot list all pagerduty accounts: %w", ErrPaginationLimitReached)
}

func (d *PagerDutyDriver) queryUsers(ctx context.Context, offset, limit int) (*pagerdutyUsersPage, error) {
	q := url.Values{}
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	u := url.URL{Scheme: "https", Host: "api.pagerduty.com", Path: "/users", RawQuery: q.Encode()}
	endpoint := u.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create pagerduty users request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute pagerduty users request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch pagerduty users: unexpected status %d", httpResp.StatusCode)
	}

	var page pagerdutyUsersPage
	if err := json.NewDecoder(httpResp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("cannot decode pagerduty users response: %w", err)
	}

	return &page, nil
}
