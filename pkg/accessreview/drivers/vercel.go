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

	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
)

// Vercel path elements joined onto the driver's base URL. The members
// prefix/suffix pair is concatenated rather than joined because the team ID
// between them must reach url.URL.String() unescaped; see queryMembers.
const (
	vercelTeamMembersPathPrefix = "/v3/teams/"
	vercelTeamMembersPathSuffix = "/members"

	vercelV2Segment    = "v2"
	vercelTeamsSegment = "teams"
)

// VercelDriver fetches team members from the Vercel REST API using a
// pre-authenticated HTTP client (Bearer token). The TeamID is captured
// during the OAuth callback (Pattern 2-auto). Pagination is via the
// `pagination.next` cursor on the response body, replayed as the
// `?until=<cursor>` query parameter on the next request.
//
// Notes on data quality:
//   - When `isEnterpriseManaged` is true on a member, the IdP is the
//     source of truth for MFA — the v3 members endpoint does not surface
//     MFA status, so MFAStatus is always Unknown.
//   - The driver does not wrap the transport with retryRoundTripper:
//     Vercel's documented rate-limit contract is loose enough that the
//     extra retry layer is not warranted in v1.
type VercelDriver struct {
	httpClient *http.Client
	teamID     string
	baseURL    string
}

var _ Driver = (*VercelDriver)(nil)

// NewVercelDriver builds a driver against baseURL, the Vercel REST API
// origin (e.g. https://api.vercel.com). The version segment is per-call:
// members come from v3, the team lookup from v2.
func NewVercelDriver(httpClient *http.Client, teamID, baseURL string) *VercelDriver {
	return &VercelDriver{
		httpClient: httpClient,
		teamID:     teamID,
		baseURL:    baseURL,
	}
}

type vercelMember struct {
	UID                 string `json:"uid"`
	Email               string `json:"email"`
	Username            string `json:"username"`
	Name                string `json:"name"`
	Role                string `json:"role"`
	Confirmed           bool   `json:"confirmed"`
	IsEnterpriseManaged bool   `json:"isEnterpriseManaged"`
	JoinedFrom          struct {
		Origin string `json:"origin"`
	} `json:"joinedFrom"`
}

// Vercel's documented pagination shape returns `next` as a Unix-millis
// cursor (number) or null on the last page; modelling it as `*int64`
// matches both. Decoding as a string would fail in production.
type vercelMembersPage struct {
	Members    []vercelMember `json:"members"`
	Pagination struct {
		Next *int64 `json:"next"`
	} `json:"pagination"`
}

func (d *VercelDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	cursor := ""

	for range maxPaginationPages {
		page, err := d.queryMembers(ctx, cursor)
		if err != nil {
			return nil, err
		}

		for _, m := range page.Members {
			fullName := m.Name
			if fullName == "" {
				fullName = m.Username
			}

			role := strings.TrimSpace(m.Role)

			roles := []string{}
			if role != "" {
				roles = []string{role}
			}

			confirmed := m.Confirmed
			record := AccountRecord{
				Email:       m.Email,
				FullName:    fullName,
				Roles:       roles,
				Active:      &confirmed,
				IsAdmin:     m.Role == "OWNER" || m.Role == "owner",
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  m.UID,
			}

			records = append(records, record)
		}

		if page.Pagination.Next == nil {
			return records, nil
		}

		cursor = strconv.FormatInt(*page.Pagination.Next, 10)
	}

	return nil, fmt.Errorf("cannot list all vercel accounts: %w", ErrPaginationLimitReached)
}

func (d *VercelDriver) queryMembers(ctx context.Context, cursor string) (*vercelMembersPage, error) {
	q := url.Values{}
	q.Set("limit", "100")

	if cursor != "" {
		q.Set("until", cursor)
	}

	base, err := url.Parse(d.baseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse vercel base URL: %w", err)
	}

	// The team ID is concatenated, not url.PathEscape'd: url.URL.String()
	// escapes the assembled Path on its own, and escaping here first would
	// change the emitted URL.
	u := url.URL{
		Scheme:   base.Scheme,
		Host:     base.Host,
		Path:     base.Path + vercelTeamMembersPathPrefix + d.teamID + vercelTeamMembersPathSuffix,
		RawQuery: q.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create vercel members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute vercel members request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch vercel members: unexpected status %d", httpResp.StatusCode)
	}

	var page vercelMembersPage
	if err := json.NewDecoder(httpResp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("cannot decode vercel members response: %w", err)
	}

	return &page, nil
}

// vercelNameResolver resolves the Vercel team name. When the captured
// TeamID is a personal-account UID, the v2 teams endpoint returns 404;
// the resolver falls back to /v2/user and uses `username` (or `name`)
// as the display name.
type vercelNameResolver struct {
	httpClient *http.Client
	teamID     string
	baseURL    string
}

// NewVercelNameResolver resolves the team name against baseURL, the Vercel
// REST API origin (e.g. https://api.vercel.com).
func NewVercelNameResolver(httpClient *http.Client, teamID, baseURL string) NameResolver {
	return &vercelNameResolver{httpClient: httpClient, teamID: teamID, baseURL: baseURL}
}

func (r *vercelNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.teamID == "" {
		return "", nil
	}

	teamURL, err := url.JoinPath(r.baseURL, vercelV2Segment, vercelTeamsSegment, url.PathEscape(r.teamID))
	if err != nil {
		return "", fmt.Errorf("cannot build vercel team URL: %w", err)
	}

	teamReq, err := http.NewRequestWithContext(ctx, http.MethodGet, teamURL, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create vercel team request: %w", err)
	}

	teamReq.Header.Set("Accept", "application/json")

	teamResp, err := r.httpClient.Do(teamReq)
	if err != nil {
		return "", fmt.Errorf("cannot execute vercel team request: %w", err)
	}

	defer func() { _ = teamResp.Body.Close() }()

	if teamResp.StatusCode == http.StatusOK {
		var body struct {
			Name string `json:"name"`
			Slug string `json:"slug"`
		}
		if err := json.NewDecoder(teamResp.Body).Decode(&body); err != nil {
			return "", fmt.Errorf("cannot decode vercel team response: %w", err)
		}

		if body.Name != "" {
			return body.Name, nil
		}

		return body.Slug, nil
	}

	if teamResp.StatusCode != http.StatusNotFound {
		return "", nameStatusError("vercel team", teamResp.StatusCode)
	}

	// Personal-account fallback: /v2/teams/<uid> returns 404, but
	// /v2/user works with the same Bearer token.
	user, err := connector.FetchVercelUser(ctx, r.httpClient, r.baseURL)
	if err != nil {
		return "", err
	}

	if user.Username != "" {
		return user.Username, nil
	}

	return user.Name, nil
}
