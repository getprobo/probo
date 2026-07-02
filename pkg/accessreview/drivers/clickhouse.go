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
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

const clickhouseAPIBaseURL = "https://api.clickhouse.cloud"

// ClickHouseDriver lists the members of a single ClickHouse Cloud
// organization. A key/secret pair (HTTP Basic) is scoped to exactly one
// organization, so the driver discovers that organization via
// GET /v1/organizations and then lists its members — no org ID needs to be
// configured. The Basic credential is applied by the connection transport.
type ClickHouseDriver struct {
	httpClient *http.Client
}

var _ Driver = (*ClickHouseDriver)(nil)

type clickhouseOrgsResponse struct {
	Result []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"result"`
}

type clickhouseMembersResponse struct {
	Result []clickhouseMember `json:"result"`
}

type clickhouseMember struct {
	UserID        string `json:"userId"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	JoinedAt      string `json:"joinedAt"`
	AssignedRoles []struct {
		RoleName string `json:"roleName"`
	} `json:"assignedRoles"`
}

func NewClickHouseDriver(httpClient *http.Client) *ClickHouseDriver {
	return &ClickHouseDriver{httpClient: httpClient}
}

func (d *ClickHouseDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	organizationID, err := d.resolveOrganizationID(ctx)
	if err != nil {
		return nil, err
	}

	members, err := d.fetchMembers(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	records := make([]AccountRecord, 0, len(members))

	for _, m := range members {
		email := strings.TrimSpace(m.Email)
		if email == "" {
			continue
		}

		record := AccountRecord{
			Email:       email,
			FullName:    clickhouseFullName(m, email),
			Roles:       clickhouseRoles(m),
			IsAdmin:     clickhouseIsAdmin(m),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  strings.TrimSpace(m.UserID),
		}

		if m.JoinedAt != "" {
			if t, err := time.Parse(time.RFC3339, m.JoinedAt); err == nil {
				record.CreatedAt = &t
			}
		}

		records = append(records, record)
	}

	return records, nil
}

func (d *ClickHouseDriver) resolveOrganizationID(ctx context.Context) (string, error) {
	endpoint, err := url.JoinPath(clickhouseAPIBaseURL, "v1", "organizations")
	if err != nil {
		return "", fmt.Errorf("cannot build clickhouse organizations URL: %w", err)
	}

	var resp clickhouseOrgsResponse
	if err := d.getJSON(ctx, endpoint, "clickhouse organizations", &resp); err != nil {
		return "", err
	}

	if len(resp.Result) == 0 || strings.TrimSpace(resp.Result[0].ID) == "" {
		return "", fmt.Errorf("cannot determine clickhouse organization: API key is not associated with any organization")
	}

	return strings.TrimSpace(resp.Result[0].ID), nil
}

func (d *ClickHouseDriver) fetchMembers(ctx context.Context, organizationID string) ([]clickhouseMember, error) {
	endpoint, err := url.JoinPath(clickhouseAPIBaseURL, "v1", "organizations", url.PathEscape(organizationID), "members")
	if err != nil {
		return nil, fmt.Errorf("cannot build clickhouse members URL: %w", err)
	}

	var resp clickhouseMembersResponse
	if err := d.getJSON(ctx, endpoint, "clickhouse members", &resp); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

func (d *ClickHouseDriver) getJSON(ctx context.Context, endpoint, what string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("cannot create %s request: %w", what, err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot execute %s request: %w", what, err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return fmt.Errorf("cannot fetch %s: unexpected status %d", what, httpResp.StatusCode)
	}

	if err := json.NewDecoder(httpResp.Body).Decode(out); err != nil {
		return fmt.Errorf("cannot decode %s response: %w", what, err)
	}

	return nil
}

func clickhouseFullName(m clickhouseMember, fallback string) string {
	if name := strings.TrimSpace(m.Name); name != "" {
		return name
	}

	return fallback
}

// clickhouseRoles prefers the custom/system roles in assignedRoles (the live
// source of truth) and falls back to the deprecated `role` field, which is
// frozen for organizations migrated to custom roles.
func clickhouseRoles(m clickhouseMember) []string {
	names := make([]string, 0, len(m.AssignedRoles))

	for _, r := range m.AssignedRoles {
		if name := strings.TrimSpace(r.RoleName); name != "" {
			names = append(names, name)
		}
	}

	if len(names) > 0 {
		return names
	}

	switch strings.ToLower(strings.TrimSpace(m.Role)) {
	case "admin":
		return []string{"Admin"}
	case "developer":
		return []string{"Developer"}
	default:
		if r := strings.TrimSpace(m.Role); r != "" {
			return []string{r}
		}

		return []string{}
	}
}

func clickhouseIsAdmin(m clickhouseMember) bool {
	// assignedRoles is the live source of truth. The deprecated role field is
	// frozen at its pre-migration value for organizations that moved to custom
	// roles, so consult it only when there are no assigned roles — otherwise a
	// stale "admin" could misclassify a since-demoted member.
	if len(m.AssignedRoles) > 0 {
		for _, r := range m.AssignedRoles {
			if strings.EqualFold(strings.TrimSpace(r.RoleName), "admin") {
				return true
			}
		}

		return false
	}

	return strings.EqualFold(strings.TrimSpace(m.Role), "admin")
}
