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
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	dotfileAPIHost   = "api.dotfile.com"
	dotfileUsersPath = "/v1/users"
	dotfilePageSize  = 100
)

// DotfileDriver lists the users of a single Dotfile workspace. The API key
// (sent in the X-DOTFILE-API-KEY header by the connection transport) is bound
// to one workspace, so GET /v1/users returns every user of that workspace with
// no tenant selector. Pagination is page/limit based.
type DotfileDriver struct {
	httpClient *http.Client
}

var _ Driver = (*DotfileDriver)(nil)

type dotfileUser struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	// SuspendedAt is the suspension timestamp; null (decoded as empty) means
	// the user is active. The endpoint returns only active users unless
	// include_suspended=true is requested, so both states are enumerated.
	SuspendedAt string `json:"suspended_at"`
}

type dotfileUsersResponse struct {
	Data       []dotfileUser `json:"data"`
	Pagination struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
		Count int `json:"count"`
	} `json:"pagination"`
}

func NewDotfileDriver(httpClient *http.Client) *DotfileDriver {
	return &DotfileDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
	}
}

func (d *DotfileDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	collected := 0

	for page := 1; page <= maxPaginationPages; page++ {
		resp, err := d.fetchPage(ctx, page)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Data {
			email := strings.TrimSpace(u.Email)
			if email == "" {
				continue
			}

			active := u.SuspendedAt == ""

			records = append(records, AccountRecord{
				Email:       email,
				FullName:    dotfileFullName(u, email),
				Roles:       dotfileRoles(u.Role),
				Active:      &active,
				IsAdmin:     dotfileIsAdmin(u.Role),
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				CreatedAt:   parseRFC3339Ptr(u.CreatedAt),
				ExternalID:  u.ID,
			})
		}

		collected += len(resp.Data)
		if len(resp.Data) == 0 || collected >= resp.Pagination.Count {
			return records, nil
		}
	}

	return nil, fmt.Errorf("cannot list all dotfile accounts: %w", ErrPaginationLimitReached)
}

func (d *DotfileDriver) fetchPage(ctx context.Context, page int) (*dotfileUsersResponse, error) {
	q := url.Values{}
	q.Set("include_suspended", "true")
	q.Set("limit", strconv.Itoa(dotfilePageSize))
	q.Set("page", strconv.Itoa(page))

	endpoint := url.URL{
		Scheme:   "https",
		Host:     dotfileAPIHost,
		Path:     dotfileUsersPath,
		RawQuery: q.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create dotfile users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute dotfile users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch dotfile users: unexpected status %d", httpResp.StatusCode)
	}

	var resp dotfileUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode dotfile users response: %w", err)
	}

	return &resp, nil
}

// dotfileFullName joins the user's first and last names, falling back to the
// email when Dotfile exposes neither.
func dotfileFullName(u dotfileUser, email string) string {
	name := strings.TrimSpace(strings.TrimSpace(u.FirstName) + " " + strings.TrimSpace(u.LastName))
	if name == "" {
		return email
	}

	return name
}

// dotfileRoles returns the user's single Dotfile role as a one-element slice
// (owner / admin / member / a custom role name), or an empty slice when none
// is set.
func dotfileRoles(role string) []string {
	if r := strings.TrimSpace(role); r != "" {
		return []string{r}
	}

	return []string{}
}

// dotfileIsAdmin reports whether the role grants administrative access. Dotfile
// has two system admin roles — owner (can delete the workspace) and admin (can
// change settings / invite) — while member and custom roles do not.
func dotfileIsAdmin(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner", "admin":
		return true
	default:
		return false
	}
}
