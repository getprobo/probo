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
	"fmt"
	"net/http"

	"encoding/json"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"net/url"
	"strconv"
	"strings"
)

const (
	dotfileUsersPath = "/users"
	dotfilePageSize  = 100
)

// DotfileDriver lists the users of a single Dotfile workspace. The API key
// (sent in the X-DOTFILE-API-KEY header by the connection transport) is bound
// to one workspace, so GET /v1/users returns every user of that workspace with
// no tenant selector. Pagination is page/limit based.
type DotfileDriver struct {
	httpClient *http.Client
	baseURL    string
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

func NewDotfileDriver(httpClient *http.Client, baseURL string) *DotfileDriver {
	return &DotfileDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		baseURL: baseURL,
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
				IsAdmin:     new(dotfileIsAdmin(u.Role)),
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
	endpoint, err := url.JoinPath(d.baseURL, dotfileUsersPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build dotfile users URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create dotfile users request: %w", err)
	}

	q := req.URL.Query()
	q.Set("include_suspended", "true")
	q.Set("limit", strconv.Itoa(dotfilePageSize))
	q.Set("page", strconv.Itoa(page))
	req.URL.RawQuery = q.Encode()

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

func dotfileSource() Factory {
	return provider.Over(func(
		ctx context.Context,
		credential connector.HTTPCredential,
		opened *provider.Handle,
		logger *log.Logger,
	) (Driver, error) {
		return capable(
			NewDotfileDriver(credential.Client, opened.Endpoints.APIBase),
			nil,
			nil,
		), nil
	})
}
