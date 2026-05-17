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
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

// SnykDriver fetches organization memberships from the Snyk REST API
// using a pre-authenticated HTTP client (Bearer token from the Snyk
// Apps OAuth PKCE flow). Pagination is via the `links.next` field on
// the response body (a relative URL fragment under api.snyk.io).
//
// Note: Snyk uses a single-use rotating refresh token (~180d TTL).
// Persistence of the rotated refresh token is handled by the existing
// callers — see pkg/accessreview/access_source_service.go:336-347 for
// the campaign-fetch path and pkg/accessreview/source_name_worker.go:121-128
// for the source-name path. Both run inside a transaction so concurrent
// runs serialise per-row.
type SnykDriver struct {
	httpClient *http.Client
	orgID      string
}

var _ Driver = (*SnykDriver)(nil)

func NewSnykDriver(httpClient *http.Client, orgID string) *SnykDriver {
	return &SnykDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		orgID: orgID,
	}
}

type snykMembership struct {
	ID         string `json:"id"`
	Attributes struct {
		User struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"user"`
		Role struct {
			Name string `json:"name"`
		} `json:"role"`
	} `json:"attributes"`
}

type snykMembershipsPage struct {
	Data  []snykMembership `json:"data"`
	Links struct {
		Next string `json:"next"`
	} `json:"links"`
}

func (d *SnykDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	next := fmt.Sprintf(
		"https://api.snyk.io/rest/orgs/%s/memberships?version=2024-10-15&limit=100",
		url.PathEscape(d.orgID),
	)

	for range maxPaginationPages {
		page, err := d.queryMemberships(ctx, next)
		if err != nil {
			return nil, err
		}

		for _, m := range page.Data {
			record := AccountRecord{
				Email:       m.Attributes.User.Email,
				FullName:    m.Attributes.User.Name,
				Role:        m.Attributes.Role.Name,
				ExternalID:  m.ID,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
			}
			records = append(records, record)
		}

		if page.Links.Next == "" {
			return records, nil
		}

		// Snyk surfaces `links.next` as either a path-only fragment
		// (e.g. "/rest/orgs/<id>/memberships?...&starting_after=...")
		// or an absolute URL. Normalise to absolute.
		if strings.HasPrefix(page.Links.Next, "http://") || strings.HasPrefix(page.Links.Next, "https://") {
			next = page.Links.Next
		} else {
			next = "https://api.snyk.io" + page.Links.Next
		}
	}

	return nil, fmt.Errorf("cannot list all snyk accounts: %w", ErrPaginationLimitReached)
}

func (d *SnykDriver) queryMemberships(ctx context.Context, endpoint string) (*snykMembershipsPage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create snyk memberships request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.api+json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute snyk memberships request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch snyk memberships: unexpected status %d", httpResp.StatusCode)
	}

	var page snykMembershipsPage
	if err := json.NewDecoder(httpResp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("cannot decode snyk memberships response: %w", err)
	}

	return &page, nil
}
