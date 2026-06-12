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
	apolloUsersEndpoint = "https://api.apollo.io/api/v1/users/search"
	apolloUsersPageSize = 100
)

// ApolloDriver lists the teammates (seats) of a single Apollo.io account.
// The master API key is bound to one account, so GET /api/v1/users/search
// returns every teammate of that account. The key is presented in the
// x-api-key header by the connection transport, not here.
type ApolloDriver struct {
	httpClient *http.Client
}

var _ Driver = (*ApolloDriver)(nil)

type apolloUser struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	// Role is Apollo's permission-profile name (e.g. "Admin", "Billing and
	// Seat Manager"). It is decoded as RawMessage and read via apolloRole so
	// a non-string shape (object/null on some plans) degrades to an empty
	// role instead of failing the decode of the whole page.
	Role json.RawMessage `json:"role"`
}

type apolloUsersResponse struct {
	Users      []apolloUser `json:"users"`
	Pagination struct {
		Page         int `json:"page"`
		PerPage      int `json:"per_page"`
		TotalPages   int `json:"total_pages"`
		TotalEntries int `json:"total_entries"`
	} `json:"pagination"`
}

func NewApolloDriver(httpClient *http.Client) *ApolloDriver {
	return &ApolloDriver{httpClient: httpClient}
}

func (d *ApolloDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	for page := 1; page <= maxPaginationPages; page++ {
		resp, err := d.fetchUsersPage(ctx, page)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Users {
			email := strings.TrimSpace(u.Email)
			if email == "" {
				continue
			}

			role := apolloRole(u.Role)

			records = append(records, AccountRecord{
				Email:       email,
				FullName:    apolloFullName(u, email),
				Roles:       apolloRoles(role),
				IsAdmin:     apolloIsAdmin(role),
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  strings.TrimSpace(u.ID),
			})
		}

		if resp.Pagination.TotalPages <= page || len(resp.Users) == 0 {
			return records, nil
		}
	}

	return nil, fmt.Errorf("cannot list all apollo users: %w", ErrPaginationLimitReached)
}

func (d *ApolloDriver) fetchUsersPage(ctx context.Context, page int) (*apolloUsersResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apolloUsersEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create apollo users request: %w", err)
	}

	q := req.URL.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("per_page", strconv.Itoa(apolloUsersPageSize))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute apollo users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch apollo users: unexpected status %d", httpResp.StatusCode)
	}

	var resp apolloUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode apollo users response: %w", err)
	}

	return &resp, nil
}

func apolloFullName(u apolloUser, fallback string) string {
	if name := strings.TrimSpace(u.Name); name != "" {
		return name
	}

	combined := strings.TrimSpace(strings.TrimSpace(u.FirstName) + " " + strings.TrimSpace(u.LastName))
	if combined != "" {
		return combined
	}

	return fallback
}

// apolloRole extracts a role label from Apollo's `role` field, which is
// normally a plain string. It also tolerates a `{"name": ...}` object and a
// null/absent value so an unexpected shape on some plans yields an empty
// role rather than failing the whole page decode.
func apolloRole(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}

	var obj struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil {
		return strings.TrimSpace(obj.Name)
	}

	return ""
}

// apolloRoles wraps a single Apollo permission-profile name as the account's
// role set, returning an empty slice when no profile is present.
func apolloRoles(role string) []string {
	if role == "" {
		return []string{}
	}

	return []string{role}
}

// apolloIsAdmin reports whether a permission-profile name is Apollo's
// built-in super-admin profile ("Admin"). Apollo exposes no boolean admin
// flag and lets customers name custom profiles freely, so the match is exact
// (case-insensitive), not a substring: a profile merely containing "admin"
// (e.g. "Billing Admin") is not auto-classified — its Role is still surfaced
// for the reviewer to judge.
func apolloIsAdmin(role string) bool {
	return strings.EqualFold(strings.TrimSpace(role), "Admin")
}
