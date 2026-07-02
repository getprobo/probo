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
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

// HubSpotDriver fetches account users from HubSpot via OAuth2-authenticated
// REST requests.
type HubSpotDriver struct {
	httpClient *http.Client
}

var _ Driver = (*HubSpotDriver)(nil)

type (
	hubspotRolesResponse struct {
		Results []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"results"`
	}

	hubspotUser struct {
		ID            string   `json:"id"`
		Email         string   `json:"email"`
		FirstName     string   `json:"firstName"`
		LastName      string   `json:"lastName"`
		RoleID        string   `json:"roleId"`
		RoleIDs       []string `json:"roleIds"`
		PrimaryTeamID string   `json:"primaryTeamId"`
		SuperAdmin    bool     `json:"superAdmin"`
		Archived      *bool    `json:"archived"`
		Deactivated   *bool    `json:"deactivated"`
		IsActive      *bool    `json:"isActive"`
		Active        *bool    `json:"active"`
		HSDeactivated *bool    `json:"hs_deactivated"`
	}

	hubspotUsersResponse struct {
		Results []hubspotUser `json:"results"`
		Paging  *struct {
			Next *struct {
				After string `json:"after"`
			} `json:"next"`
		} `json:"paging"`
	}
)

const (
	hubspotUsersEndpoint = "https://api.hubapi.com/settings/v3/users"
	hubspotRolesEndpoint = "https://api.hubapi.com/settings/v3/users/roles"
)

func NewHubSpotDriver(httpClient *http.Client) *HubSpotDriver {
	return &HubSpotDriver{
		httpClient: httpClient,
	}
}

func (d *HubSpotDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	roleMap, _ := d.fetchRoles(ctx)

	var (
		records []AccountRecord
		after   string
	)

	for range maxPaginationPages {
		resp, err := d.fetchUsers(ctx, after)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Results {
			fullName := strings.TrimSpace(u.FirstName + " " + u.LastName)

			record := AccountRecord{
				Email:       u.Email,
				FullName:    fullName,
				Roles:       hubspotRoles(u, roleMap),
				Active:      hubspotUserActive(u),
				IsAdmin:     u.SuperAdmin,
				ExternalID:  u.ID,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
			}

			if record.Email != "" || record.ExternalID != "" {
				records = append(records, record)
			}
		}

		if resp.Paging == nil || resp.Paging.Next == nil || resp.Paging.Next.After == "" {
			return records, nil
		}

		after = resp.Paging.Next.After
	}

	return nil, fmt.Errorf("cannot list all hubspot accounts: %w", ErrPaginationLimitReached)
}

func (d *HubSpotDriver) fetchUsers(ctx context.Context, after string) (*hubspotUsersResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hubspotUsersEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create hubspot users request: %w", err)
	}

	q := req.URL.Query()
	q.Set("limit", "100")

	if after != "" {
		q.Set("after", after)
	}

	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute hubspot users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch hubspot users: unexpected status %d", httpResp.StatusCode)
	}

	var resp hubspotUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode hubspot users response: %w", err)
	}

	return &resp, nil
}

func (d *HubSpotDriver) fetchRoles(ctx context.Context) (map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hubspotRolesEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create hubspot roles request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute hubspot roles request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch hubspot roles: unexpected status %d", httpResp.StatusCode)
	}

	var resp hubspotRolesResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode hubspot roles response: %w", err)
	}

	roleMap := make(map[string]string, len(resp.Results))
	for _, r := range resp.Results {
		roleMap[r.ID] = r.Name
	}

	return roleMap, nil
}

func hubspotRoles(user hubspotUser, roleMap map[string]string) []string {
	roleIDs := hubspotRoleIDs(user)

	seen := make(map[string]struct{}, len(roleIDs)+1)
	roles := make([]string, 0, len(roleIDs)+1)

	if roleMap != nil {
		for _, id := range roleIDs {
			name, ok := roleMap[id]
			if !ok {
				continue
			}

			if _, dup := seen[name]; dup {
				continue
			}

			seen[name] = struct{}{}
			roles = append(roles, name)
		}
	}

	if user.SuperAdmin {
		if _, dup := seen["Super Admin"]; !dup {
			roles = append(roles, "Super Admin")
		}
	}

	if len(roles) > 0 {
		return roles
	}

	return []string{"User"}
}

func hubspotRoleIDs(user hubspotUser) []string {
	seen := make(map[string]struct{}, 1+len(user.RoleIDs))
	ids := make([]string, 0, 1+len(user.RoleIDs))

	add := func(id string) {
		if id == "" {
			return
		}

		if _, ok := seen[id]; ok {
			return
		}

		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	add(user.RoleID)

	for _, id := range user.RoleIDs {
		add(id)
	}

	return ids
}

func hubspotUserActive(user hubspotUser) *bool {
	if user.IsActive != nil {
		return new(*user.IsActive)
	}

	if user.Active != nil {
		return new(*user.Active)
	}

	if user.Deactivated != nil {
		return new(!*user.Deactivated)
	}

	if user.HSDeactivated != nil {
		return new(!*user.HSDeactivated)
	}

	if user.Archived != nil {
		return new(!*user.Archived)
	}

	return nil
}
