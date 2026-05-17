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

	"go.probo.inc/probo/pkg/coredata"
)

// VercelDriver fetches team members from the Vercel REST API using a
// pre-authenticated HTTP client (Bearer token). The TeamID is captured
// during the OAuth callback (Pattern 2-auto). Pagination is via the
// `pagination.next` cursor on the response body, replayed as the
// `?until=<cursor>` query parameter on the next request.
//
// Notes on data quality:
//   - When `isEnterpriseManaged` is true on a member, the IdP is the
//     source of truth for MFA — the v3 members endpoint does not surface
//     MFA status, so MFAStatus is always Unknown.
//   - The driver does not wrap the transport with retryRoundTripper:
//     Vercel's documented rate-limit contract is loose enough that the
//     extra retry layer is not warranted in v1.
type VercelDriver struct {
	httpClient *http.Client
	teamID     string
}

var _ Driver = (*VercelDriver)(nil)

func NewVercelDriver(httpClient *http.Client, teamID string) *VercelDriver {
	return &VercelDriver{
		httpClient: httpClient,
		teamID:     teamID,
	}
}

type vercelMember struct {
	UID                 string `json:"uid"`
	Email               string `json:"email"`
	Username            string `json:"username"`
	Name                string `json:"name"`
	Role                string `json:"role"`
	Confirmed           bool   `json:"confirmed"`
	IsEnterpriseManaged bool   `json:"isEnterpriseManaged"`
	JoinedFrom          struct {
		Origin string `json:"origin"`
	} `json:"joinedFrom"`
}

type vercelMembersPage struct {
	Members    []vercelMember `json:"members"`
	Pagination struct {
		Next string `json:"next"`
	} `json:"pagination"`
}

func (d *VercelDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	cursor := ""

	for range maxPaginationPages {
		page, err := d.queryMembers(ctx, cursor)
		if err != nil {
			return nil, err
		}

		for _, m := range page.Members {
			fullName := m.Name
			if fullName == "" {
				fullName = m.Username
			}

			confirmed := m.Confirmed
			record := AccountRecord{
				Email:       m.Email,
				FullName:    fullName,
				Role:        m.Role,
				Active:      &confirmed,
				IsAdmin:     m.Role == "OWNER" || m.Role == "owner",
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
				ExternalID:  m.UID,
			}

			records = append(records, record)
		}

		if page.Pagination.Next == "" {
			return records, nil
		}
		cursor = page.Pagination.Next
	}

	return nil, fmt.Errorf("cannot list all vercel accounts: %w", ErrPaginationLimitReached)
}

func (d *VercelDriver) queryMembers(ctx context.Context, cursor string) (*vercelMembersPage, error) {
	q := url.Values{}
	q.Set("limit", "100")
	if cursor != "" {
		q.Set("until", cursor)
	}
	u := url.URL{
		Scheme:   "https",
		Host:     "api.vercel.com",
		Path:     "/v3/teams/" + d.teamID + "/members",
		RawQuery: q.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create vercel members request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute vercel members request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch vercel members: unexpected status %d", httpResp.StatusCode)
	}

	var page vercelMembersPage
	if err := json.NewDecoder(httpResp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("cannot decode vercel members response: %w", err)
	}

	return &page, nil
}
