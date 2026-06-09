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

	"go.probo.inc/probo/pkg/coredata"
)

type CursorDriver struct {
	httpClient *http.Client
}

var _ Driver = (*CursorDriver)(nil)

// cursorMembersEndpoint lists every member of the team the admin API key
// belongs to. Cursor's Admin API authenticates with the key as the HTTP
// Basic auth username (handled by the connection transport) and exposes
// no pagination on this endpoint, so a single GET returns the full team.
const cursorMembersEndpoint = "https://api.cursor.com/teams/members"

type cursorMembersResponse struct {
	TeamMembers []struct {
		// ID is the stable Cursor member identifier. The Admin API
		// returns it as a JSON string (despite the docs labelling it a
		// number), so it is decoded as a string and used verbatim.
		ID        string `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		Role      string `json:"role"`
		IsRemoved bool   `json:"isRemoved"`
	} `json:"teamMembers"`
}

func NewCursorDriver(httpClient *http.Client) *CursorDriver {
	return &CursorDriver{
		httpClient: httpClient,
	}
}

func (d *CursorDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cursorMembersEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create cursor members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute cursor members request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch cursor members: unexpected status %d", httpResp.StatusCode)
	}

	var resp cursorMembersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode cursor members response: %w", err)
	}

	records := make([]AccountRecord, 0, len(resp.TeamMembers))
	for _, m := range resp.TeamMembers {
		if m.Email == "" {
			continue
		}

		// isRemoved is Cursor's only account-status signal, so Active is
		// always populated (never nil): a removed member is reported
		// inactive rather than dropped, per the AccountRecord contract.
		active := !m.IsRemoved

		records = append(records, AccountRecord{
			Email:       m.Email,
			FullName:    m.Name,
			Role:        cursorRole(m.Role),
			Active:      &active,
			IsAdmin:     cursorIsAdmin(m.Role),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
			AccountType: coredata.AccessEntryAccountTypeUser,
			ExternalID:  m.ID,
		})
	}

	return records, nil
}

// cursorIsAdmin reports whether a Cursor team role carries team
// administration rights. Both paid ("owner") and free-tier
// ("free-owner") owners administer the team.
func cursorIsAdmin(role string) bool {
	return role == "owner" || role == "free-owner"
}

func cursorRole(role string) string {
	switch role {
	case "owner", "free-owner":
		return "Owner"
	case "member":
		return "Member"
	case "removed":
		return "Removed"
	default:
		return role
	}
}
