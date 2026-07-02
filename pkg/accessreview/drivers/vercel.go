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

// Vercel's documented pagination shape returns `next` as a Unix-millis
// cursor (number) or null on the last page; modelling it as `*int64`
// matches both. Decoding as a string would fail in production.
type vercelMembersPage struct {
	Members    []vercelMember `json:"members"`
	Pagination struct {
		Next *int64 `json:"next"`
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

			role := strings.TrimSpace(m.Role)

			roles := []string{}
			if role != "" {
				roles = []string{role}
			}

			confirmed := m.Confirmed
			record := AccountRecord{
				Email:       m.Email,
				FullName:    fullName,
				Roles:       roles,
				Active:      &confirmed,
				IsAdmin:     m.Role == "OWNER" || m.Role == "owner",
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  m.UID,
			}

			records = append(records, record)
		}

		if page.Pagination.Next == nil {
			return records, nil
		}

		cursor = strconv.FormatInt(*page.Pagination.Next, 10)
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
