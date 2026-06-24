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
	"strconv"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	pylonUsersEndpoint     = "https://api.usepylon.com/users"
	pylonUserRolesEndpoint = "https://api.usepylon.com/user-roles"
	// pylonPageSize is the page size requested from the cursor-paginated
	// list endpoints. Pylon caps `limit` at 999 (it must be > 0 and < 1000);
	// 100 returns every member of typical organizations in one page.
	pylonPageSize = 100
)

// PylonDriver lists the users (agents) of a single Pylon organization. The
// API token (Bearer) is bound to one organization, so GET /users returns
// every member of that organization with no tenant selector. Each user
// carries an opaque role_id, which the driver resolves to a human-readable
// role name via GET /user-roles.
type PylonDriver struct {
	httpClient *http.Client
}

var _ Driver = (*PylonDriver)(nil)

type pylonUser struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	RoleID string `json:"role_id"`
	Status string `json:"status"`
}

type pylonRole struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type pylonPagination struct {
	Cursor      string `json:"cursor"`
	HasNextPage bool   `json:"has_next_page"`
}

type pylonUsersResponse struct {
	Data       []pylonUser     `json:"data"`
	Pagination pylonPagination `json:"pagination"`
}

type pylonRolesResponse struct {
	Data       []pylonRole     `json:"data"`
	Pagination pylonPagination `json:"pagination"`
}

func NewPylonDriver(httpClient *http.Client) *PylonDriver {
	return &PylonDriver{httpClient: httpClient}
}

func (d *PylonDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	roles, err := d.fetchRoles(ctx)
	if err != nil {
		return nil, err
	}

	var (
		records []AccountRecord
		cursor  string
	)

	for range maxPaginationPages {
		resp, err := d.fetchUsersPage(ctx, cursor)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Data {
			email := strings.TrimSpace(u.Email)
			if email == "" {
				continue
			}

			role := roles[u.RoleID]

			records = append(records, AccountRecord{
				Email:       email,
				FullName:    pylonFullName(u, email),
				Roles:       pylonRoles(role),
				Active:      activeFromStatus(u.Status),
				IsAdmin:     pylonIsAdmin(role),
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  strings.TrimSpace(u.ID),
			})
		}

		if !resp.Pagination.HasNextPage || resp.Pagination.Cursor == "" {
			return records, nil
		}

		cursor = resp.Pagination.Cursor
	}

	return nil, fmt.Errorf("cannot list all pylon users: %w", ErrPaginationLimitReached)
}

// fetchRoles loads the organization's role catalogue once, keyed by role ID,
// so each user's opaque role_id can be resolved to a role name and admin
// classification.
func (d *PylonDriver) fetchRoles(ctx context.Context) (map[string]pylonRole, error) {
	roles := make(map[string]pylonRole)

	cursor := ""

	for range maxPaginationPages {
		resp, err := d.fetchRolesPage(ctx, cursor)
		if err != nil {
			return nil, err
		}

		for _, r := range resp.Data {
			roles[r.ID] = r
		}

		if !resp.Pagination.HasNextPage || resp.Pagination.Cursor == "" {
			return roles, nil
		}

		cursor = resp.Pagination.Cursor
	}

	return nil, fmt.Errorf("cannot list all pylon user-roles: %w", ErrPaginationLimitReached)
}

func (d *PylonDriver) fetchUsersPage(ctx context.Context, cursor string) (*pylonUsersResponse, error) {
	httpResp, err := d.fetchPage(ctx, pylonUsersEndpoint, cursor)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch pylon users: unexpected status %d", httpResp.StatusCode)
	}

	var resp pylonUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode pylon users response: %w", err)
	}

	return &resp, nil
}

func (d *PylonDriver) fetchRolesPage(ctx context.Context, cursor string) (*pylonRolesResponse, error) {
	httpResp, err := d.fetchPage(ctx, pylonUserRolesEndpoint, cursor)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch pylon user-roles: unexpected status %d", httpResp.StatusCode)
	}

	var resp pylonRolesResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode pylon user-roles response: %w", err)
	}

	return &resp, nil
}

func (d *PylonDriver) fetchPage(ctx context.Context, endpoint, cursor string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create pylon request: %w", err)
	}

	q := req.URL.Query()
	q.Set("limit", strconv.Itoa(pylonPageSize))

	if cursor != "" {
		q.Set("cursor", cursor)
	}

	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute pylon request: %w", err)
	}

	return httpResp, nil
}

func pylonFullName(u pylonUser, fallback string) string {
	if name := strings.TrimSpace(u.Name); name != "" {
		return name
	}

	return fallback
}

// pylonRoles returns the user's role as a single-element slice using the
// resolved role name, or an empty slice when the role_id did not resolve.
func pylonRoles(role pylonRole) []string {
	if name := strings.TrimSpace(role.Name); name != "" {
		return []string{name}
	}

	return []string{}
}

// pylonIsAdmin reports whether the resolved role is Pylon's built-in Admin
// role. Pylon ships two default roles (Member and Admin); the match is on
// the stable slug, falling back to an exact (case-insensitive) name match,
// so a custom role merely containing "admin" is not auto-classified.
func pylonIsAdmin(role pylonRole) bool {
	if slug := strings.TrimSpace(role.Slug); slug != "" {
		return strings.EqualFold(slug, "admin")
	}

	return strings.EqualFold(strings.TrimSpace(role.Name), "Admin")
}
