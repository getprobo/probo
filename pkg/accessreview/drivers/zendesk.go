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

	"go.probo.inc/probo/pkg/coredata"
)

// ZendeskDriver lists a Zendesk account's staff (agents and admins) via GET
// /api/v2/users.json. The API host is per-customer (<subdomain>.zendesk.com),
// captured at connect time and stored on the connector settings. End-users
// (ticket submitters) are excluded — they are customers, not access subjects.
type ZendeskDriver struct {
	httpClient *http.Client
	subdomain  string // e.g. "acme" for acme.zendesk.com
}

var _ Driver = (*ZendeskDriver)(nil)

// NewZendeskDriver wraps the connection's SSRF-protected transport with a
// retrying transport for transient 5xx, matching the canonical sibling
// drivers (datadog.go, heroku.go). The caller's *http.Client is not mutated.
func NewZendeskDriver(httpClient *http.Client, subdomain string) *ZendeskDriver {
	return &ZendeskDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		subdomain: subdomain,
	}
}

const zendeskPageSize = 100

type zendeskUser struct {
	ID                   int64   `json:"id"`
	Email                string  `json:"email"`
	Name                 string  `json:"name"`
	Role                 string  `json:"role"` // "end-user", "agent", "admin"
	Suspended            bool    `json:"suspended"`
	Active               bool    `json:"active"`
	TwoFactorAuthEnabled *bool   `json:"two_factor_auth_enabled"`
	LastLoginAt          *string `json:"last_login_at"`
	CreatedAt            string  `json:"created_at"`
}

// zendeskUsersResponse is the GET /api/v2/users.json payload. Zendesk uses
// cursor pagination: meta.has_more signals more pages and meta.after_cursor is
// the token for the next page.
type zendeskUsersResponse struct {
	Users []zendeskUser `json:"users"`
	Meta  struct {
		HasMore     bool   `json:"has_more"`
		AfterCursor string `json:"after_cursor"`
	} `json:"meta"`
}

func (d *ZendeskDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records     []AccountRecord
		afterCursor string
	)

	for range maxPaginationPages {
		resp, err := d.queryUsers(ctx, afterCursor)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Users {
			// The query already filters to agents + admins, but guard here
			// too: end-users are ticket submitters, not staff with access.
			if u.Role == "end-user" {
				continue
			}

			records = append(records, zendeskRecord(u))
		}

		if !resp.Meta.HasMore || resp.Meta.AfterCursor == "" {
			return records, nil
		}

		afterCursor = resp.Meta.AfterCursor
	}

	return nil, fmt.Errorf("cannot list all zendesk users: %w", ErrPaginationLimitReached)
}

func zendeskRecord(u zendeskUser) AccountRecord {
	active := u.Active && !u.Suspended
	isAdmin := u.Role == "admin"

	// Zendesk reports 2FA per user; map it to the MFA status. A null value
	// (absent on some plans) stays unknown rather than asserting "disabled".
	mfaStatus := coredata.MFAStatusUnknown

	if u.TwoFactorAuthEnabled != nil {
		if *u.TwoFactorAuthEnabled {
			mfaStatus = coredata.MFAStatusEnabled
		} else {
			mfaStatus = coredata.MFAStatusDisabled
		}
	}

	lastLogin := ""
	if u.LastLoginAt != nil {
		lastLogin = *u.LastLoginAt
	}

	return AccountRecord{
		Email:     u.Email,
		FullName:  u.Name,
		Role:      u.Role,
		Active:    &active,
		IsAdmin:   isAdmin,
		MFAStatus: mfaStatus,
		// Zendesk's users API does not expose the sign-in method
		// (password / SSO / social), so the auth method is unknown.
		AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
		AccountType: coredata.AccessEntryAccountTypeUser,
		ExternalID:  strconv.FormatInt(u.ID, 10),
		LastLogin:   parseRFC3339Ptr(lastLogin),
		CreatedAt:   parseRFC3339Ptr(u.CreatedAt),
	}
}

func (d *ZendeskDriver) queryUsers(ctx context.Context, afterCursor string) (*zendeskUsersResponse, error) {
	q := url.Values{}
	q.Set("page[size]", strconv.Itoa(zendeskPageSize))
	// Restrict to staff (agents + admins); end-users are ticket submitters.
	q.Add("role[]", "agent")
	q.Add("role[]", "admin")

	if afterCursor != "" {
		q.Set("page[after]", afterCursor)
	}

	endpoint := url.URL{
		Scheme:   "https",
		Host:     d.subdomain + ".zendesk.com",
		Path:     "/api/v2/users.json",
		RawQuery: q.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create zendesk users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot list zendesk users: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot list zendesk users: unexpected status %d", resp.StatusCode)
	}

	var out zendeskUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("cannot decode zendesk users response: %w", err)
	}

	return &out, nil
}
