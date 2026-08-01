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
	"strings"

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
	baseURL     string
}

var _ Driver = (*NetlifyDriver)(nil)

const (
	netlifyAccountsSegment = "accounts"
	netlifyMembersSegment  = "members"
)

// NewNetlifyDriver builds a driver against baseURL, the versioned Netlify
// API origin (e.g. https://api.netlify.com/api/v1).
func NewNetlifyDriver(httpClient *http.Client, accountSlug, baseURL string) *NetlifyDriver {
	return &NetlifyDriver{
		httpClient:  httpClient,
		accountSlug: accountSlug,
		baseURL:     baseURL,
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

	// The members endpoint hangs the account slug straight off the version
	// root — there is no "accounts" segment, unlike the account endpoint.
	u, err := url.JoinPath(d.baseURL, url.PathEscape(d.accountSlug), netlifyMembersSegment)
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
			role := strings.TrimSpace(m.Role)

			roles := []string{}
			if role != "" {
				roles = []string{role}
			}

			record := AccountRecord{
				Email:       m.Email,
				FullName:    m.FullName,
				Roles:       roles,
				ExternalID:  m.ID,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
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

// netlifyNameResolver resolves the Netlify account name.
type netlifyNameResolver struct {
	httpClient  *http.Client
	accountSlug string
	baseURL     string
}

// NewNetlifyNameResolver resolves the account name against baseURL, the
// versioned Netlify API origin (e.g. https://api.netlify.com/api/v1).
func NewNetlifyNameResolver(httpClient *http.Client, accountSlug, baseURL string) NameResolver {
	return &netlifyNameResolver{httpClient: httpClient, accountSlug: accountSlug, baseURL: baseURL}
}

func (r *netlifyNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.accountSlug == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath(r.baseURL, netlifyAccountsSegment, url.PathEscape(r.accountSlug))
	if err != nil {
		return "", fmt.Errorf("cannot build netlify account URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create netlify account request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute netlify account request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nameStatusError("netlify account", httpResp.StatusCode)
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode netlify account response: %w", err)
	}

	return resp.Name, nil
}

// ListNetlifyOrganizations fetches the Netlify accounts the authenticated
// user belongs to.
func ListNetlifyOrganizations(ctx context.Context, httpClient *http.Client) ([]Organization, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.netlify.com/api/v1/accounts",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create netlify organizations request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch netlify organizations: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch netlify organizations: unexpected status %d", resp.StatusCode)
	}

	var accounts []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, fmt.Errorf("cannot decode netlify organizations response: %w", err)
	}

	result := make([]Organization, len(accounts))
	for i, a := range accounts {
		displayName := a.Name
		if displayName == "" {
			displayName = a.Slug
		}

		result[i] = Organization{Slug: a.Slug, DisplayName: displayName}
	}

	return result, nil
}
