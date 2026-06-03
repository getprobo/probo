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
	"strconv"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// ClickUpDriver fetches workspace ("team") members from the ClickUp
// REST API using a pre-authenticated HTTP client (Bearer token). The
// team endpoint returns the full member list inline in a single
// response — no pagination is performed.
//
// ClickUp does not issue refresh tokens; the existing RefreshableClient
// falls back to a non-refreshing client when RefreshToken == "" and the
// access source resolver re-prompts for re-authorization on 401.
type ClickUpDriver struct {
	httpClient *http.Client
	teamID     string
}

var _ Driver = (*ClickUpDriver)(nil)

func NewClickUpDriver(httpClient *http.Client, teamID string) *ClickUpDriver {
	return &ClickUpDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		teamID: teamID,
	}
}

type clickupMember struct {
	User struct {
		ID         json.Number `json:"id"`
		Email      string      `json:"email"`
		Username   string      `json:"username"`
		Role       int         `json:"role"`
		LastActive string      `json:"last_active"`
	} `json:"user"`
	InvitePending *bool `json:"invite_pending"`
}

type clickupTeamResponse struct {
	Team struct {
		Members []clickupMember `json:"members"`
	} `json:"team"`
}

func (d *ClickUpDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	endpoint, err := url.JoinPath("https://api.clickup.com", "api", "v2", "team", url.PathEscape(d.teamID))
	if err != nil {
		return nil, fmt.Errorf("cannot build clickup team URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create clickup team request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute clickup team request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch clickup team: unexpected status %d", httpResp.StatusCode)
	}

	var resp clickupTeamResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode clickup team response: %w", err)
	}

	records := make([]AccountRecord, 0, len(resp.Team.Members))
	for _, m := range resp.Team.Members {
		role := clickupRoleLabel(m.User.Role)
		isAdmin := m.User.Role == 1 || m.User.Role == 2

		record := AccountRecord{
			Email:       m.User.Email,
			FullName:    m.User.Username,
			Role:        role,
			IsAdmin:     isAdmin,
			ExternalID:  m.User.ID.String(),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
			AccountType: coredata.AccessEntryAccountTypeUser,
		}

		if m.InvitePending != nil {
			active := !*m.InvitePending
			record.Active = &active
		}

		if m.User.LastActive != "" {
			// ClickUp emits last_active as a Unix-millis string; fall
			// back to RFC3339 if a future API change switches format.
			if t, err := parseClickUpTime(m.User.LastActive); err == nil {
				record.LastLogin = &t
			}
		}

		records = append(records, record)
	}

	return records, nil
}

// clickupRoleLabel maps ClickUp numeric role codes to human-readable
// labels. Source: https://clickup.com/api (Team Members endpoint).
func clickupRoleLabel(role int) string {
	switch role {
	case 1:
		return "owner"
	case 2:
		return "admin"
	case 3:
		return "member"
	case 4:
		return "guest"
	default:
		return ""
	}
}

// parseClickUpTime accepts both ClickUp's Unix-millis-as-string format
// and RFC3339 timestamps so the driver remains forward-compatible.
func parseClickUpTime(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}

	// strconv.ParseInt rejects trailing non-digit garbage that fmt.Sscanf
	// would silently truncate (e.g. "123abc" → 123).
	ms, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse clickup time %q: %w", raw, err)
	}

	return time.UnixMilli(ms).UTC(), nil
}
