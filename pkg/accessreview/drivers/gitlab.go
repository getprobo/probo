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
	baseURL    string
}

var _ Driver = (*GitLabDriver)(nil)

const (
	gitlabGroupsSegment  = "groups"
	gitlabAllMembersPath = "members/all"
)

// gitlabDefaultBaseURL is the GitLab SaaS API root. It backs only the
// exported ListGitLabOrganizations, and only when its caller resolves no
// APIBase for the provider (unregistered, or registered without one).
// Every other path goes through the injected baseURL instead.
const gitlabDefaultBaseURL = "https://gitlab.com/api/v4"

// NewGitLabDriver builds a driver against baseURL, the versioned GitLab API
// origin (e.g. https://gitlab.com/api/v4).
func NewGitLabDriver(httpClient *http.Client, groupID, baseURL string) *GitLabDriver {
	return &GitLabDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		groupID: groupID,
		baseURL: baseURL,
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

	u, err := url.JoinPath(d.baseURL, gitlabGroupsSegment, url.PathEscape(d.groupID), gitlabAllMembersPath)
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
				IsAdmin:     new(m.AccessLevel >= 50), // 50 = Owner
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  strconv.FormatInt(m.ID, 10),
			}

			records = append(records, record)
		}

		rawNext := rfc5988.FindByRel(linkHeader, "next")
		if rawNext == "" {
			return records, nil
		}

		next, err = sameHostNextPageURL("gitlab", d.baseURL, rawNext)
		if err != nil {
			return nil, err
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

// gitlabNameResolver resolves the GitLab group name.
type gitlabNameResolver struct {
	httpClient *http.Client
	groupID    string
	baseURL    string
}

// NewGitLabNameResolver resolves the group name against baseURL, the
// versioned GitLab API origin (e.g. https://gitlab.com/api/v4).
func NewGitLabNameResolver(httpClient *http.Client, groupID, baseURL string) NameResolver {
	return &gitlabNameResolver{httpClient: httpClient, groupID: groupID, baseURL: baseURL}
}

func (r *gitlabNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.groupID == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath(r.baseURL, gitlabGroupsSegment, url.PathEscape(r.groupID))
	if err != nil {
		return "", fmt.Errorf("cannot build gitlab group URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create gitlab group request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute gitlab group request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nameStatusError("gitlab group", httpResp.StatusCode)
	}

	var resp struct {
		Name     string `json:"name"`
		FullPath string `json:"full_path"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode gitlab group response: %w", err)
	}

	if resp.Name != "" {
		return resp.Name, nil
	}

	return resp.FullPath, nil
}

// ListGitLabOrganizations fetches the GitLab groups the authenticated
// user owns, from baseURL ("" for GitLab SaaS). Group IDs are int64; we
// surface them as strings so they fit the Organization.Slug shape.
func ListGitLabOrganizations(ctx context.Context, httpClient *http.Client, baseURL string) ([]Organization, error) {
	if baseURL == "" {
		baseURL = gitlabDefaultBaseURL
	}

	endpoint, err := url.JoinPath(baseURL, gitlabGroupsSegment)
	if err != nil {
		return nil, fmt.Errorf("cannot build gitlab organizations URL: %w", err)
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("cannot parse gitlab organizations URL: %w", err)
	}

	q := parsed.Query()
	q.Set("min_access_level", "50")
	q.Set("per_page", "100")
	parsed.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create gitlab organizations request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch gitlab organizations: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch gitlab organizations: unexpected status %d", resp.StatusCode)
	}

	var groups []struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		FullPath string `json:"full_path"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, fmt.Errorf("cannot decode gitlab organizations response: %w", err)
	}

	result := make([]Organization, len(groups))
	for i, g := range groups {
		displayName := g.Name
		if displayName == "" {
			displayName = g.FullPath
		}

		result[i] = Organization{
			Slug:        strconv.FormatInt(g.ID, 10),
			DisplayName: displayName,
		}
	}

	return result, nil
}
