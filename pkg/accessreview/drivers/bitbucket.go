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

// BitbucketDriver fetches workspace members from the Bitbucket Cloud
// REST API using a pre-authenticated HTTP client (Bearer token).
//
// The Bitbucket member object exposes very little: account_id, display
// name, optional email (often hidden by privacy), nickname. There is no
// role / MFA / last-login data available, so those fields are left at
// their zero defaults / nil / Unknown.
type BitbucketDriver struct {
	httpClient *http.Client
	workspace  string
}

var _ Driver = (*BitbucketDriver)(nil)

func NewBitbucketDriver(httpClient *http.Client, workspace string) *BitbucketDriver {
	return &BitbucketDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		workspace: workspace,
	}
}

type bitbucketMember struct {
	User struct {
		AccountID   string `json:"account_id"`
		DisplayName string `json:"display_name"`
		Nickname    string `json:"nickname"`
		Email       string `json:"email"`
	} `json:"user"`
}

type bitbucketMembersPage struct {
	Values []bitbucketMember `json:"values"`
	Next   string            `json:"next"`
}

func (d *BitbucketDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	next := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/workspaces/%s/members?fields=%%2Bvalues.user.email&pagelen=100",
		url.PathEscape(d.workspace),
	)

	for range maxPaginationPages {
		page, err := d.queryMembers(ctx, next)
		if err != nil {
			return nil, err
		}

		for _, m := range page.Values {
			fullName := m.User.DisplayName
			if fullName == "" {
				fullName = m.User.Nickname
			}

			record := AccountRecord{
				Email:       m.User.Email,
				FullName:    fullName,
				ExternalID:  m.User.AccountID,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
			}

			records = append(records, record)
		}

		next = page.Next
		if next == "" {
			return records, nil
		}
	}

	return nil, fmt.Errorf("cannot list all bitbucket accounts: %w", ErrPaginationLimitReached)
}

func (d *BitbucketDriver) queryMembers(ctx context.Context, endpoint string) (*bitbucketMembersPage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create bitbucket members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute bitbucket members request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch bitbucket members: unexpected status %d", httpResp.StatusCode)
	}

	var page bitbucketMembersPage
	if err := json.NewDecoder(httpResp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("cannot decode bitbucket members response: %w", err)
	}

	return &page, nil
}
