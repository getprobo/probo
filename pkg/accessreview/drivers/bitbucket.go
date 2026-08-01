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
	baseURL    string
}

var _ Driver = (*BitbucketDriver)(nil)

const (
	bitbucketWorkspacesSegment = "workspaces"
	bitbucketMembersSegment    = "members"
)

// NewBitbucketDriver builds a driver against baseURL, the versioned
// Bitbucket Cloud API origin (e.g. https://api.bitbucket.org/2.0).
func NewBitbucketDriver(httpClient *http.Client, workspace, baseURL string) *BitbucketDriver {
	return &BitbucketDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		workspace: workspace,
		baseURL:   baseURL,
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

	u, err := url.JoinPath(d.baseURL, bitbucketWorkspacesSegment, url.PathEscape(d.workspace), bitbucketMembersSegment)
	if err != nil {
		return nil, fmt.Errorf("cannot build bitbucket members URL: %w", err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("cannot parse bitbucket members URL: %w", err)
	}

	q := parsed.Query()
	q.Set("fields", "+values.user.email")
	q.Set("pagelen", "100")
	parsed.RawQuery = q.Encode()
	next := parsed.String()

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
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
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

// bitbucketNameResolver resolves the Bitbucket workspace name.
type bitbucketNameResolver struct {
	httpClient *http.Client
	workspace  string
	baseURL    string
}

// NewBitbucketNameResolver resolves the workspace name against baseURL, the
// versioned Bitbucket Cloud API origin (e.g. https://api.bitbucket.org/2.0).
func NewBitbucketNameResolver(httpClient *http.Client, workspace, baseURL string) NameResolver {
	return &bitbucketNameResolver{httpClient: httpClient, workspace: workspace, baseURL: baseURL}
}

func (r *bitbucketNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.workspace == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath(r.baseURL, bitbucketWorkspacesSegment, url.PathEscape(r.workspace))
	if err != nil {
		return "", fmt.Errorf("cannot build bitbucket workspace URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create bitbucket workspace request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute bitbucket workspace request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nameStatusError("bitbucket workspace", httpResp.StatusCode)
	}

	var resp struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode bitbucket workspace response: %w", err)
	}

	if resp.Name != "" {
		return resp.Name, nil
	}

	return resp.Slug, nil
}

// ListBitbucketOrganizations fetches the workspaces the authenticated
// Bitbucket user belongs to. The legacy /2.0/workspaces endpoint was
// sunset by CHANGE-2770 (April 2026); /2.0/user/workspaces is the
// supported cross-workspace replacement (CHANGE-3022). Bitbucket pages
// via an absolute `next` URL on each response; follow until exhausted.
func ListBitbucketOrganizations(ctx context.Context, httpClient *http.Client) ([]Organization, error) {
	pageURL := "https://api.bitbucket.org/2.0/user/workspaces?pagelen=100"
	result := make([]Organization, 0)

	for range maxPaginationPages {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot create bitbucket organizations request: %w", err)
		}

		req.Header.Set("Accept", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("cannot fetch bitbucket organizations: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("cannot fetch bitbucket organizations: unexpected status %d", resp.StatusCode)
		}

		// We tolerate both shapes (flat and nested under `workspace`) since
		// Atlassian has shipped variants of similar endpoints with both.
		var body struct {
			Values []struct {
				Slug      string `json:"slug"`
				Name      string `json:"name"`
				Workspace struct {
					Slug string `json:"slug"`
					Name string `json:"name"`
				} `json:"workspace"`
			} `json:"values"`
			Next string `json:"next"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("cannot decode bitbucket organizations response: %w", err)
		}

		_ = resp.Body.Close()

		for _, v := range body.Values {
			slug, name := v.Slug, v.Name
			if slug == "" {
				slug = v.Workspace.Slug
				name = v.Workspace.Name
			}

			displayName := name
			if displayName == "" {
				displayName = slug
			}

			result = append(result, Organization{Slug: slug, DisplayName: displayName})
		}

		if body.Next == "" {
			return result, nil
		}

		pageURL = body.Next
	}

	return nil, fmt.Errorf("cannot list all bitbucket organizations: %w", ErrPaginationLimitReached)
}
