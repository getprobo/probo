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

const (
	yousignAPIHost   = "api.yousign.app"
	yousignUsersPath = "/v3/users"
	yousignPageSize  = 100
)

// YousignDriver lists the members of a single Yousign organization. A Yousign
// API key is bound to exactly one organization, so GET /v3/users returns every
// member with no tenant selector (Pattern 3). The Bearer credential is applied
// by the connection transport. The connector targets Yousign production; the
// sandbox runs on a separate host and is not a reviewed environment.
type YousignDriver struct {
	httpClient *http.Client
}

var _ Driver = (*YousignDriver)(nil)

type yousignUsersResponse struct {
	Meta struct {
		NextCursor *string `json:"next_cursor"`
	} `json:"meta"`
	Data []yousignUser `json:"data"`
}

type yousignUser struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	JobTitle  string `json:"job_title"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	Role      string `json:"role"`
}

func NewYousignDriver(httpClient *http.Client) *YousignDriver {
	return &YousignDriver{httpClient: httpClient}
}

func (d *YousignDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	after := ""

	for range maxPaginationPages {
		resp, err := d.fetchPage(ctx, after)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Data {
			email := strings.TrimSpace(u.Email)
			if email == "" {
				continue
			}

			active := u.IsActive

			records = append(records, AccountRecord{
				Email:       email,
				FullName:    yousignFullName(u, email),
				Roles:       yousignRoles(u.Role),
				JobTitle:    strings.TrimSpace(u.JobTitle),
				Active:      &active,
				IsAdmin:     yousignIsAdmin(u.Role),
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				CreatedAt:   parseRFC3339Ptr(u.CreatedAt),
				ExternalID:  strings.TrimSpace(u.ID),
			})
		}

		if resp.Meta.NextCursor == nil || strings.TrimSpace(*resp.Meta.NextCursor) == "" {
			return records, nil
		}

		after = *resp.Meta.NextCursor
	}

	return nil, fmt.Errorf("cannot list all yousign accounts: %w", ErrPaginationLimitReached)
}

func (d *YousignDriver) fetchPage(ctx context.Context, after string) (*yousignUsersResponse, error) {
	q := url.Values{}
	q.Set("limit", strconv.Itoa(yousignPageSize))

	if after != "" {
		q.Set("after", after)
	}

	endpoint := url.URL{
		Scheme:   "https",
		Host:     yousignAPIHost,
		Path:     yousignUsersPath,
		RawQuery: q.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create yousign users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute yousign users request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch yousign users: unexpected status %d", httpResp.StatusCode)
	}

	var resp yousignUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode yousign users response: %w", err)
	}

	return &resp, nil
}

func yousignFullName(u yousignUser, fallback string) string {
	if name := strings.TrimSpace(u.FirstName + " " + u.LastName); name != "" {
		return name
	}

	return fallback
}

// yousignRoles maps Yousign's single role string to a display label. Documented
// roles are admin/member, with owner reserved for the org owner; an unknown
// future value is passed through verbatim and no role yields an empty slice.
func yousignRoles(role string) []string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin":
		return []string{"Admin"}
	case "owner":
		return []string{"Owner"}
	case "member":
		return []string{"Member"}
	default:
		if r := strings.TrimSpace(role); r != "" {
			return []string{r}
		}

		return []string{}
	}
}

// yousignIsAdmin reports whether a Yousign role grants administration. Owner is
// strictly more privileged than admin, so both qualify; the match is exact, not
// a substring.
func yousignIsAdmin(role string) bool {
	return strings.EqualFold(strings.TrimSpace(role), "admin") ||
		strings.EqualFold(strings.TrimSpace(role), "owner")
}
