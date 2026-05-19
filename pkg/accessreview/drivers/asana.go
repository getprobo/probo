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

	"go.probo.inc/probo/pkg/coredata"
)

// AsanaDriver fetches users for a given Asana workspace via REST API
// using a pre-authenticated HTTP client (Bearer token). Pagination is
// driven by the body field `next_page.uri`.
//
// The user object exposes very little: gid, name, email. There is no
// role / MFA / last-login signal. Active is derived defensively from
// the presence of an email — Asana hides the email field for
// deactivated or privacy-protected users.
type AsanaDriver struct {
	httpClient   *http.Client
	workspaceGID string
}

var _ Driver = (*AsanaDriver)(nil)

func NewAsanaDriver(httpClient *http.Client, workspaceGID string) *AsanaDriver {
	return &AsanaDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		workspaceGID: workspaceGID,
	}
}

type asanaUser struct {
	GID   string `json:"gid"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type asanaUsersPage struct {
	Data     []asanaUser `json:"data"`
	NextPage *struct {
		URI string `json:"uri"`
	} `json:"next_page"`
}

func (d *AsanaDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	next := fmt.Sprintf(
		"https://app.asana.com/api/1.0/workspaces/%s/users?opt_fields=email,name&limit=100",
		url.PathEscape(d.workspaceGID),
	)

	for range maxPaginationPages {
		page, err := d.queryUsers(ctx, next)
		if err != nil {
			return nil, err
		}

		for _, u := range page.Data {
			// Asana's workspace-users endpoint exposes no active flag,
			// and a missing email can mean deactivated, privacy-protected,
			// limited-access, or an external collaborator. Inferring
			// Active=false from any of those would fabricate state, so
			// leave Active nil (unknown) and let downstream review surface
			// the gap honestly.
			records = append(records, AccountRecord{
				Email:       u.Email,
				FullName:    u.Name,
				ExternalID:  u.GID,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
			})
		}

		if page.NextPage == nil || page.NextPage.URI == "" {
			return records, nil
		}

		next = page.NextPage.URI
	}

	return nil, fmt.Errorf("cannot list all asana accounts: %w", ErrPaginationLimitReached)
}

func (d *AsanaDriver) queryUsers(ctx context.Context, endpoint string) (*asanaUsersPage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create asana users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute asana users request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch asana users: unexpected status %d", httpResp.StatusCode)
	}

	var page asanaUsersPage
	if err := json.NewDecoder(httpResp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("cannot decode asana users response: %w", err)
	}

	return &page, nil
}
