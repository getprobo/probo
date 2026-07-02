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
	mercuryUsersEndpoint = "https://api.mercury.com/api/v1/users"
	// mercuryUsersPageSize is the page size requested from GET /api/v1/users.
	// The API caps `limit` at 1000; 500 keeps responses small while still
	// returning every member of typical organizations in one page.
	mercuryUsersPageSize = 500
)

// MercuryDriver lists the users of a single Mercury organization. The
// access token (Bearer) is bound to one organization, so GET /api/v1/users
// returns every member of that organization with no tenant selector.
type MercuryDriver struct {
	httpClient *http.Client
}

var _ Driver = (*MercuryDriver)(nil)

type mercuryUser struct {
	UserID           string `json:"userId"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	Email            string `json:"email"`
	OrganizationRole string `json:"organizationRole"`
}

type mercuryUsersResponse struct {
	Users []mercuryUser `json:"users"`
	Page  struct {
		NextPage *string `json:"nextPage"`
	} `json:"page"`
}

func NewMercuryDriver(httpClient *http.Client) *MercuryDriver {
	return &MercuryDriver{httpClient: httpClient}
}

func (d *MercuryDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records    []AccountRecord
		startAfter string
	)

	for range maxPaginationPages {
		resp, err := d.fetchUsersPage(ctx, startAfter)
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
				FullName:    mercuryFullName(u, email),
				Roles:       mercuryRoles(u.OrganizationRole),
				IsAdmin:     u.OrganizationRole == "administrator",
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  strings.TrimSpace(u.UserID),
			})
		}

		if resp.Page.NextPage == nil || *resp.Page.NextPage == "" {
			return records, nil
		}

		startAfter = *resp.Page.NextPage
	}

	return nil, fmt.Errorf("cannot list all mercury users: %w", ErrPaginationLimitReached)
}

func (d *MercuryDriver) fetchUsersPage(ctx context.Context, startAfter string) (*mercuryUsersResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mercuryUsersEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create mercury users request: %w", err)
	}

	q := req.URL.Query()
	q.Set("limit", strconv.Itoa(mercuryUsersPageSize))

	if startAfter != "" {
		q.Set("start_after", startAfter)
	}

	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute mercury users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch mercury users: unexpected status %d", httpResp.StatusCode)
	}

	var resp mercuryUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode mercury users response: %w", err)
	}

	return &resp, nil
}

func mercuryFullName(u mercuryUser, fallback string) string {
	fullName := strings.TrimSpace(strings.TrimSpace(u.FirstName) + " " + strings.TrimSpace(u.LastName))
	if fullName != "" {
		return fullName
	}

	return fallback
}

// mercuryRoles maps Mercury's organizationRole enum
// (administrator/bookkeeper/customUser/cardOnlyUser/employee) to a stable
// display label, preserving unknown future roles verbatim.
func mercuryRoles(role string) []string {
	switch role {
	case "administrator":
		return []string{"Administrator"}
	case "bookkeeper":
		return []string{"Bookkeeper"}
	case "customUser":
		return []string{"Custom User"}
	case "cardOnlyUser":
		return []string{"Card Only User"}
	case "employee":
		return []string{"Employee"}
	default:
		if role != "" {
			return []string{role}
		}

		return []string{}
	}
}
