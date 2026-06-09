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

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/rfc5988"
)

// NetlifyDriver fetches account members from the Netlify REST API
// using a pre-authenticated HTTP client (Bearer token). Pagination is
// driven by the standard RFC 5988 `Link` header with `rel="next"`.
//
// The Netlify member object exposes id / full_name / email / role only.
// There is no Active / MFA / last-login signal, so those fields are
// left at their zero defaults / nil / Unknown.
type NetlifyDriver struct {
	httpClient  *http.Client
	accountSlug string
}

var _ Driver = (*NetlifyDriver)(nil)

func NewNetlifyDriver(httpClient *http.Client, accountSlug string) *NetlifyDriver {
	return &NetlifyDriver{
		httpClient:  httpClient,
		accountSlug: accountSlug,
	}
}

type netlifyMember struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

func (d *NetlifyDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	u, err := url.JoinPath("https://api.netlify.com", "api", "v1", url.PathEscape(d.accountSlug), "members")
	if err != nil {
		return nil, fmt.Errorf("cannot build netlify members URL: %w", err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("cannot parse netlify members URL: %w", err)
	}

	q := parsed.Query()
	q.Set("per_page", "100")
	parsed.RawQuery = q.Encode()
	next := parsed.String()

	for range maxPaginationPages {
		members, linkHeader, err := d.queryMembers(ctx, next)
		if err != nil {
			return nil, err
		}

		for _, m := range members {
			record := AccountRecord{
				Email:       m.Email,
				FullName:    m.FullName,
				Role:        m.Role,
				ExternalID:  m.ID,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
			}
			records = append(records, record)
		}

		next = rfc5988.FindByRel(linkHeader, "next")
		if next == "" {
			return records, nil
		}
	}

	return nil, fmt.Errorf("cannot list all netlify accounts: %w", ErrPaginationLimitReached)
}

func (d *NetlifyDriver) queryMembers(ctx context.Context, endpoint string) ([]netlifyMember, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("cannot create netlify members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("cannot execute netlify members request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("cannot fetch netlify members: unexpected status %d", httpResp.StatusCode)
	}

	var members []netlifyMember
	if err := json.NewDecoder(httpResp.Body).Decode(&members); err != nil {
		return nil, "", fmt.Errorf("cannot decode netlify members response: %w", err)
	}

	return members, httpResp.Header.Get("Link"), nil
}
