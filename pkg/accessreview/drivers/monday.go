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
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// mondayGraphQLEndpoint is Monday.com's REST endpoint that accepts
// GraphQL queries via POST.
const mondayGraphQLEndpoint = "https://api.monday.com/v2"

// mondayUsersListQuery paginates Monday.com users by `page` (1-indexed).
// MFA is exposed only via SCIM Enterprise — leave MFAStatus=Unknown.
const mondayUsersListQuery = `query($p: Int!) { users(limit: 200, page: $p) { id email name enabled is_admin is_guest is_pending last_activity created_at title } }`

// MondayDriver fetches users from the Monday.com GraphQL API using a
// pre-authenticated HTTP client. The token flows in the Authorization
// header as a Bearer credential, which Monday.com accepts alongside the
// legacy bare-token form.
type MondayDriver struct {
	httpClient *http.Client
}

var _ Driver = (*MondayDriver)(nil)

func NewMondayDriver(httpClient *http.Client) *MondayDriver {
	return &MondayDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
	}
}

type mondayUser struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	IsAdmin      bool   `json:"is_admin"`
	IsGuest      bool   `json:"is_guest"`
	IsPending    bool   `json:"is_pending"`
	LastActivity string `json:"last_activity"`
	CreatedAt    string `json:"created_at"`
	Title        string `json:"title"`
}

type mondayUsersResponse struct {
	Data struct {
		Users []mondayUser `json:"users"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (d *MondayDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	page := 1
	for range maxPaginationPages {
		users, err := d.queryUsers(ctx, page)
		if err != nil {
			return nil, err
		}

		if len(users) == 0 {
			return records, nil
		}

		for _, u := range users {
			active := u.Enabled && !u.IsPending

			record := AccountRecord{
				Email:       u.Email,
				FullName:    u.Name,
				JobTitle:    u.Title,
				Active:      &active,
				IsAdmin:     u.IsAdmin,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  u.ID,
			}

			if u.LastActivity != "" {
				if t, err := time.Parse(time.RFC3339, u.LastActivity); err == nil {
					record.LastLogin = &t
				}
			}

			if u.CreatedAt != "" {
				if t, err := time.Parse(time.RFC3339, u.CreatedAt); err == nil {
					record.CreatedAt = &t
				}
			}

			records = append(records, record)
		}

		page++
	}

	return nil, fmt.Errorf("cannot list all monday accounts: %w", ErrPaginationLimitReached)
}

func (d *MondayDriver) queryUsers(ctx context.Context, page int) ([]mondayUser, error) {
	body := struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}{
		Query:     mondayUsersListQuery,
		Variables: map[string]any{"p": page},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal monday users query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, mondayGraphQLEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("cannot create monday users request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute monday users request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch monday users: unexpected status %d", httpResp.StatusCode)
	}

	var resp mondayUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode monday users response: %w", err)
	}

	if len(resp.Errors) > 0 {
		// Provider-supplied error messages may carry tenant identifiers
		// or query fragments — never embed them in the returned error.
		return nil, fmt.Errorf("cannot fetch monday users: graphql error")
	}

	return resp.Data.Users, nil
}
