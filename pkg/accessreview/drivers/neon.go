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

	"go.probo.inc/probo/pkg/coredata"
)

const (
	neonAPIBaseURL = "https://console.neon.tech/api/v2"

	// neonMembersPageLimit is the largest page size the Neon
	// list-members endpoint documents (limit: 1..500).
	neonMembersPageLimit = "500"
)

type NeonDriver struct {
	httpClient     *http.Client
	organizationID string
}

var _ Driver = (*NeonDriver)(nil)

type neonMembersResponse struct {
	Members    []neonOrgMember `json:"members"`
	Pagination struct {
		Next string `json:"next"`
	} `json:"pagination"`
}

type neonOrgMember struct {
	Member struct {
		ID       string `json:"id"`
		UserID   string `json:"user_id"`
		Role     string `json:"role"`
		JoinedAt string `json:"joined_at"`
	} `json:"member"`
	User struct {
		Email         string `json:"email"`
		HasMFA        *bool  `json:"has_mfa"`
		DeactivatedAt string `json:"deactivated_at"`
	} `json:"user"`
}

func NewNeonDriver(httpClient *http.Client, organizationID string) *NeonDriver {
	return &NeonDriver{
		httpClient:     httpClient,
		organizationID: organizationID,
	}
}

func (d *NeonDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records []AccountRecord
		cursor  string
	)

	for range maxPaginationPages {
		resp, err := d.queryMembers(ctx, cursor)
		if err != nil {
			return nil, err
		}

		for _, m := range resp.Members {
			if m.User.Email == "" {
				continue
			}

			records = append(records, AccountRecord{
				Email: m.User.Email,
				// The members endpoint exposes no display name;
				// fall back to the email.
				FullName: m.User.Email,
				Roles:    neonRoles(m.Member.Role),
				// deactivated_at is absent for active accounts.
				Active:      new(m.User.DeactivatedAt == ""),
				IsAdmin:     neonIsAdmin(m.Member.Role),
				MFAStatus:   neonMFAStatus(m.User.HasMFA),
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  neonExternalID(m),
				CreatedAt:   parseRFC3339Ptr(m.Member.JoinedAt),
			})
		}

		if resp.Pagination.Next == "" {
			return records, nil
		}

		cursor = resp.Pagination.Next
	}

	return nil, fmt.Errorf("cannot list all neon accounts: %w", ErrPaginationLimitReached)
}

func (d *NeonDriver) queryMembers(ctx context.Context, cursor string) (*neonMembersResponse, error) {
	endpoint, err := url.JoinPath(
		neonAPIBaseURL,
		"organizations",
		url.PathEscape(d.organizationID),
		"members",
	)
	if err != nil {
		return nil, fmt.Errorf("cannot build neon members URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create neon members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	q := req.URL.Query()
	q.Set("limit", neonMembersPageLimit)

	if cursor != "" {
		q.Set("cursor", cursor)
	}

	req.URL.RawQuery = q.Encode()

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute neon members request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch neon members: unexpected status %d", httpResp.StatusCode)
	}

	var resp neonMembersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode neon members response: %w", err)
	}

	return &resp, nil
}

// neonRoles maps Neon's lowercase member roles to their display form,
// passing unknown values through unchanged.
func neonRoles(role string) []string {
	if role == "" {
		return []string{}
	}

	switch strings.ToLower(role) {
	case "admin":
		return []string{"Admin"}
	case "member":
		return []string{"Member"}
	case "editor":
		return []string{"Editor"}
	case "viewer":
		return []string{"Viewer"}
	case "collaborator":
		return []string{"Collaborator"}
	default:
		return []string{role}
	}
}

func neonIsAdmin(role string) bool {
	return strings.EqualFold(role, "admin")
}

func neonMFAStatus(hasMFA *bool) coredata.MFAStatus {
	switch {
	case hasMFA == nil:
		return coredata.MFAStatusUnknown
	case *hasMFA:
		return coredata.MFAStatusEnabled
	default:
		return coredata.MFAStatusDisabled
	}
}

// neonExternalID prefers the stable Neon account UUID over the
// membership ID.
func neonExternalID(m neonOrgMember) string {
	if m.Member.UserID != "" {
		return m.Member.UserID
	}

	return m.Member.ID
}
