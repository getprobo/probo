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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	squareTeamMembersSearchURL = "https://connect.squareup.com/v2/team-members/search"
	// squareAPIVersion pins the request version so behavior is deterministic
	// rather than following the application's console default.
	squareAPIVersion  = "2026-05-20"
	squareSearchLimit = 200
)

// SquareDriver lists the team members of a single Square merchant. A Square
// token — OAuth Bearer or Personal Access Token — is always scoped to one
// merchant, so POST /v2/team-members/search returns every team member of that
// merchant with no tenant selector. The search returns is_owner directly, so
// there is no role resolution.
type SquareDriver struct {
	httpClient *http.Client
}

var _ Driver = (*SquareDriver)(nil)

type squareTeamMember struct {
	ID           string `json:"id"`
	IsOwner      bool   `json:"is_owner"`
	Status       string `json:"status"`
	GivenName    string `json:"given_name"`
	FamilyName   string `json:"family_name"`
	EmailAddress string `json:"email_address"`
	CreatedAt    string `json:"created_at"`
}

type squareSearchResponse struct {
	TeamMembers []squareTeamMember `json:"team_members"`
	Cursor      string             `json:"cursor"`
}

func NewSquareDriver(httpClient *http.Client) *SquareDriver {
	return &SquareDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
	}
}

func (d *SquareDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	cursor := ""

	for range maxPaginationPages {
		page, err := d.searchTeamMembers(ctx, cursor)
		if err != nil {
			return nil, err
		}

		for _, m := range page.TeamMembers {
			email := strings.TrimSpace(m.EmailAddress)
			if email == "" {
				continue
			}

			role := "member"
			if m.IsOwner {
				role = "owner"
			}

			records = append(records, AccountRecord{
				Email:       email,
				FullName:    squareFullName(m, email),
				Roles:       ownerMemberRoles(role),
				Active:      activeFromStatus(m.Status),
				IsAdmin:     m.IsOwner,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				CreatedAt:   parseRFC3339Ptr(m.CreatedAt),
				ExternalID:  m.ID,
			})
		}

		if page.Cursor == "" {
			return records, nil
		}

		cursor = page.Cursor
	}

	return nil, fmt.Errorf("cannot list all square accounts: %w", ErrPaginationLimitReached)
}

func (d *SquareDriver) searchTeamMembers(ctx context.Context, cursor string) (*squareSearchResponse, error) {
	// No query filter: return team members of every status (ACTIVE and
	// INACTIVE) so deactivated members are reviewed too.
	reqBody := struct {
		Limit  int    `json:"limit"`
		Cursor string `json:"cursor,omitempty"`
	}{
		Limit:  squareSearchLimit,
		Cursor: cursor,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal square search request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, squareTeamMembersSearchURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cannot create square team members request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Square-Version", squareAPIVersion)

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute square team members request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch square team members: unexpected status %d", httpResp.StatusCode)
	}

	var resp squareSearchResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode square team members response: %w", err)
	}

	return &resp, nil
}

// squareFullName joins the team member's given and family names, falling back
// to the email when Square exposes neither.
func squareFullName(m squareTeamMember, email string) string {
	name := strings.TrimSpace(strings.TrimSpace(m.GivenName) + " " + strings.TrimSpace(m.FamilyName))
	if name == "" {
		return email
	}

	return name
}
