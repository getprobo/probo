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
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// betterStackTeamMembersEndpoint is the Better Stack Uptime API team-members
// resource. The team-members reference documents this apex host and returns
// its pagination links on the same host, so the driver pins every page
// request to it (instead of following the response's `next` URL) to avoid a
// cross-host redirect that would drop the Authorization header.
const betterStackTeamMembersEndpoint = "https://betterstack.com/api/v2/team-members"

// BetterStackDriver fetches team members and pending invitations from the
// Better Stack Uptime API via Bearer token-authenticated REST requests. The
// teamName scopes the listing; it is required when authenticating with a
// global API token and ignored for team-scoped tokens.
type BetterStackDriver struct {
	httpClient *http.Client
	teamName   string
}

var _ Driver = (*BetterStackDriver)(nil)

type betterStackTeamMembersResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Email     string `json:"email"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			CreatedAt string `json:"created_at"`
			InvitedAt string `json:"invited_at"`
			Role      string `json:"role"`
		} `json:"attributes"`
	} `json:"data"`
	Pagination struct {
		Next *string `json:"next"`
	} `json:"pagination"`
}

func NewBetterStackDriver(httpClient *http.Client, teamName string) *BetterStackDriver {
	return &BetterStackDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		teamName: teamName,
	}
}

func (d *BetterStackDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	page := 1

	for range maxPaginationPages {
		resp, err := d.fetchTeamMembersPage(ctx, page)
		if err != nil {
			return nil, err
		}

		for _, member := range resp.Data {
			if member.Attributes.Email == "" {
				continue
			}

			record := AccountRecord{
				Email:       member.Attributes.Email,
				FullName:    strings.TrimSpace(member.Attributes.FirstName + " " + member.Attributes.LastName),
				Roles:       betterStackRoles(member.Attributes.Role),
				Active:      betterStackActive(member.Type),
				IsAdmin:     betterStackIsAdmin(member.Attributes.Role),
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  member.ID,
			}

			if t, ok := parseBetterStackTimestamp(member.Attributes.CreatedAt); ok {
				record.CreatedAt = &t
			} else if t, ok := parseBetterStackTimestamp(member.Attributes.InvitedAt); ok {
				// Invitations carry invited_at but no created_at; keep the
				// first-seen timestamp in CreatedAt for review context.
				record.CreatedAt = &t
			}

			records = append(records, record)
		}

		if resp.Pagination.Next == nil || *resp.Pagination.Next == "" {
			return records, nil
		}

		page++
	}

	return nil, fmt.Errorf("cannot list all better stack team members: %w", ErrPaginationLimitReached)
}

func (d *BetterStackDriver) fetchTeamMembersPage(
	ctx context.Context,
	page int,
) (*betterStackTeamMembersResponse, error) {
	endpoint, err := url.Parse(betterStackTeamMembersEndpoint)
	if err != nil {
		return nil, fmt.Errorf("cannot parse better stack team members URL: %w", err)
	}

	q := endpoint.Query()
	q.Set("page", strconv.Itoa(page))

	if d.teamName != "" {
		q.Set("team_name", d.teamName)
	}

	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create better stack team members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute better stack team members request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch better stack team members: unexpected status %d", httpResp.StatusCode)
	}

	var resp betterStackTeamMembersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode better stack team members response: %w", err)
	}

	return &resp, nil
}

// betterStackRoles maps a Better Stack role token to a human-readable label.
// Unknown roles (including the Enterprise "custom" roles) are passed through
// unchanged so the reviewer still sees the source value.
func betterStackRoles(role string) []string {
	if role == "" {
		return []string{}
	}

	switch role {
	case "admin":
		return []string{"Admin"}
	case "billing_admin":
		return []string{"Billing admin"}
	case "team_lead":
		return []string{"Team lead"}
	case "responder":
		return []string{"Responder"}
	case "member":
		return []string{"Member"}
	default:
		return []string{role}
	}
}

// betterStackIsAdmin flags roles with administrative control over team access:
// admin (full control) and team_lead (can manage team members and roles).
// billing_admin is billing-only, so it is not flagged.
func betterStackIsAdmin(role string) bool {
	return role == "admin" || role == "team_lead"
}

// betterStackActive maps the member record type to an explicit active signal:
// confirmed members are active, pending invitations are not, and unknown
// types leave the signal nil rather than fabricate one.
func betterStackActive(memberType string) *bool {
	switch memberType {
	case "team_member":
		return new(true)
	case "team_member_invitation":
		return new(false)
	default:
		return nil
	}
}

func parseBetterStackTimestamp(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}

	// time.Parse accepts a fractional second even when the layout omits it,
	// so RFC3339 covers Better Stack's ".000Z" timestamps too.
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, false
	}

	return t, true
}
