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
	incidentIOUsersEndpoint = "https://api.incident.io/v2/users"
	// incidentIOPageSize is the page size requested from GET /v2/users (the
	// API defaults to 25 and accepts up to 10000).
	incidentIOPageSize = 100
)

// IncidentIODriver lists the users of a single incident.io organization. The
// API key (Bearer) is bound to one organization, so GET /v2/users returns
// every user of that organization with no tenant selector.
type IncidentIODriver struct {
	httpClient *http.Client
}

var _ Driver = (*IncidentIODriver)(nil)

type incidentIORole struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type incidentIOUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	// Role is the deprecated coarse role enum: owner / administrator /
	// responder / viewer / unset. base_role / custom_roles are the live
	// RBAC roles and take precedence when present.
	Role        string           `json:"role"`
	BaseRole    *incidentIORole  `json:"base_role"`
	CustomRoles []incidentIORole `json:"custom_roles"`
}

type incidentIOUsersResponse struct {
	Users          []incidentIOUser `json:"users"`
	PaginationMeta struct {
		After string `json:"after"`
	} `json:"pagination_meta"`
}

func NewIncidentIODriver(httpClient *http.Client) *IncidentIODriver {
	return &IncidentIODriver{httpClient: httpClient}
}

func (d *IncidentIODriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	after := ""

	for range maxPaginationPages {
		resp, err := d.fetchUsersPage(ctx, after)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Users {
			email := strings.TrimSpace(u.Email)
			if email == "" {
				continue
			}

			records = append(records, AccountRecord{
				Email:       email,
				FullName:    incidentIOFullName(u, email),
				Roles:       incidentIORoles(u),
				IsAdmin:     incidentIOIsAdmin(u),
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  strings.TrimSpace(u.ID),
			})
		}

		// The `after` cursor is the authoritative end-of-results signal: stop
		// when it is empty. A short page is NOT treated as the end (incident.io
		// may return fewer than page_size rows while more pages remain). The
		// empty-page guard is only a backstop against an API that never clears
		// the cursor, so the loop cannot spin past the data.
		if resp.PaginationMeta.After == "" || len(resp.Users) == 0 {
			return records, nil
		}

		after = resp.PaginationMeta.After
	}

	return nil, fmt.Errorf("cannot list all incident.io users: %w", ErrPaginationLimitReached)
}

func (d *IncidentIODriver) fetchUsersPage(ctx context.Context, after string) (*incidentIOUsersResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, incidentIOUsersEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create incident.io users request: %w", err)
	}

	q := req.URL.Query()
	q.Set("page_size", strconv.Itoa(incidentIOPageSize))

	if after != "" {
		q.Set("after", after)
	}

	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute incident.io users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch incident.io users: unexpected status %d", httpResp.StatusCode)
	}

	var resp incidentIOUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode incident.io users response: %w", err)
	}

	return &resp, nil
}

func incidentIOFullName(u incidentIOUser, fallback string) string {
	if name := strings.TrimSpace(u.Name); name != "" {
		return name
	}

	return fallback
}

// incidentIORoles returns the user's roles, preferring the live RBAC roles
// (base_role + custom_roles) and falling back to the deprecated `role` enum
// only when no RBAC role is present.
func incidentIORoles(u incidentIOUser) []string {
	roles := []string{}

	if u.BaseRole != nil {
		if name := strings.TrimSpace(u.BaseRole.Name); name != "" {
			roles = append(roles, name)
		}
	}

	for _, r := range u.CustomRoles {
		if name := strings.TrimSpace(r.Name); name != "" {
			roles = append(roles, name)
		}
	}

	if len(roles) > 0 {
		return roles
	}

	if name := incidentIODeprecatedRoleName(u.Role); name != "" {
		return []string{name}
	}

	return []string{}
}

// incidentIODeprecatedRoleName maps the deprecated coarse role enum to a
// display label, returning "" for "unset" or an absent value.
func incidentIODeprecatedRoleName(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner":
		return "Owner"
	case "administrator":
		return "Administrator"
	case "responder":
		return "Responder"
	case "viewer":
		return "Viewer"
	default:
		return ""
	}
}

// incidentIOIsAdmin reports whether the user holds an administrative role. It
// prefers the live base_role slug and falls back to the deprecated role enum;
// both "owner" and "administrator" are administrative.
func incidentIOIsAdmin(u incidentIOUser) bool {
	if u.BaseRole != nil && strings.TrimSpace(u.BaseRole.Slug) != "" {
		return incidentIOAdminSlug(u.BaseRole.Slug)
	}

	return incidentIOAdminSlug(u.Role)
}

func incidentIOAdminSlug(slug string) bool {
	switch strings.ToLower(strings.TrimSpace(slug)) {
	case "owner", "administrator":
		return true
	default:
		return false
	}
}
