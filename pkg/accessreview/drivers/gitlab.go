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
	"strconv"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/rfc5988"
)

// GitLabDriver fetches all-members of a GitLab group via REST API
// using a pre-authenticated HTTP client (Bearer token).
//
// Notes on data quality on gitlab.com SaaS:
//   - `email` is often null on Free; we leave it blank when the API
//     returns null. `username` is used as the FullName fallback.
//   - Per-user MFA status is admin-only on gitlab.com SaaS, so MFAStatus
//     is left Unknown.
//   - `last_login_at` is paid-plan only via the separate /billable_members
//     endpoint, so LastLogin is left nil for v1.
type GitLabDriver struct {
	httpClient *http.Client
	groupID    string
}

var _ Driver = (*GitLabDriver)(nil)

func NewGitLabDriver(httpClient *http.Client, groupID string) *GitLabDriver {
	return &GitLabDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		groupID: groupID,
	}
}

type gitlabMember struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	State       string `json:"state"`
	AccessLevel int    `json:"access_level"`
}

func (d *GitLabDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	u, err := url.JoinPath("https://gitlab.com", "api", "v4", "groups", url.PathEscape(d.groupID), "members", "all")
	if err != nil {
		return nil, fmt.Errorf("cannot build gitlab members URL: %w", err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("cannot parse gitlab members URL: %w", err)
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
			fullName := m.Name
			if fullName == "" {
				fullName = m.Username
			}

			active := m.State == "active"

			roles := gitlabRoles(m.AccessLevel)

			record := AccountRecord{
				Email:       m.Email,
				FullName:    fullName,
				Roles:       roles,
				Active:      &active,
				IsAdmin:     m.AccessLevel >= 50, // 50 = Owner
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  strconv.FormatInt(m.ID, 10),
			}

			records = append(records, record)
		}

		next = rfc5988.FindByRel(linkHeader, "next")
		if next == "" {
			return records, nil
		}
	}

	return nil, fmt.Errorf("cannot list all gitlab accounts: %w", ErrPaginationLimitReached)
}

func (d *GitLabDriver) queryMembers(ctx context.Context, endpoint string) ([]gitlabMember, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("cannot create gitlab members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("cannot execute gitlab members request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("cannot fetch gitlab members: unexpected status %d", httpResp.StatusCode)
	}

	var members []gitlabMember
	if err := json.NewDecoder(httpResp.Body).Decode(&members); err != nil {
		return nil, "", fmt.Errorf("cannot decode gitlab members response: %w", err)
	}

	return members, httpResp.Header.Get("Link"), nil
}

// gitlabRoles maps GitLab numeric access levels to human
// labels. Source: https://docs.gitlab.com/api/members/#roles
func gitlabRoles(level int) []string {
	switch level {
	case 5:
		return []string{"Minimal Access"}
	case 10:
		return []string{"Guest"}
	case 15:
		return []string{"Planner"}
	case 20:
		return []string{"Reporter"}
	case 30:
		return []string{"Developer"}
	case 40:
		return []string{"Maintainer"}
	case 50:
		return []string{"Owner"}
	default:
		return []string{}
	}
}
