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
	baseURL      string
}

var _ Driver = (*AsanaDriver)(nil)

const (
	asanaWorkspacesSegment = "workspaces"
	asanaUsersSegment      = "users"
)

// asanaDefaultBaseURL is the Asana API root. It backs only the exported
// ListAsanaOrganizations, and only when its caller resolves no APIBase for
// the provider (unregistered, or registered without one). Every other path
// goes through the injected baseURL instead.
const asanaDefaultBaseURL = "https://app.asana.com/api/1.0"

// NewAsanaDriver builds a driver against baseURL, the versioned Asana API
// origin (e.g. https://app.asana.com/api/1.0).
func NewAsanaDriver(httpClient *http.Client, workspaceGID, baseURL string) *AsanaDriver {
	return &AsanaDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		workspaceGID: workspaceGID,
		baseURL:      baseURL,
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

	u, err := url.JoinPath(d.baseURL, asanaWorkspacesSegment, url.PathEscape(d.workspaceGID), asanaUsersSegment)
	if err != nil {
		return nil, fmt.Errorf("cannot build asana users URL: %w", err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("cannot parse asana users URL: %w", err)
	}

	q := parsed.Query()
	q.Set("opt_fields", "email,name")
	q.Set("limit", "100")
	parsed.RawQuery = q.Encode()
	next := parsed.String()

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
			records = append(
				records,
				AccountRecord{
					Email:       u.Email,
					FullName:    u.Name,
					ExternalID:  u.GID,
					MFAStatus:   coredata.MFAStatusUnknown,
					AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
					AccountType: coredata.AccessReviewEntryAccountTypeUser,
				},
			)
		}

		if page.NextPage == nil || page.NextPage.URI == "" {
			return records, nil
		}

		next, err = sameHostNextPageURL("asana", d.baseURL, page.NextPage.URI)
		if err != nil {
			return nil, err
		}
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

// asanaNameResolver resolves the Asana workspace name.
type asanaNameResolver struct {
	httpClient   *http.Client
	workspaceGID string
	baseURL      string
}

// NewAsanaNameResolver resolves the workspace name against baseURL, the
// versioned Asana API origin (e.g. https://app.asana.com/api/1.0).
func NewAsanaNameResolver(httpClient *http.Client, workspaceGID, baseURL string) NameResolver {
	return &asanaNameResolver{httpClient: httpClient, workspaceGID: workspaceGID, baseURL: baseURL}
}

func (r *asanaNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.workspaceGID == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath(r.baseURL, asanaWorkspacesSegment, url.PathEscape(r.workspaceGID))
	if err != nil {
		return "", fmt.Errorf("cannot build asana workspace URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create asana workspace request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute asana workspace request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nameStatusError("asana workspace", httpResp.StatusCode)
	}

	var resp struct {
		Data struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode asana workspace response: %w", err)
	}

	return resp.Data.Name, nil
}

// ListAsanaOrganizations fetches the workspaces the authenticated Asana
// user belongs to, from baseURL ("" for the Asana SaaS API).
func ListAsanaOrganizations(ctx context.Context, httpClient *http.Client, baseURL string) ([]Organization, error) {
	if baseURL == "" {
		baseURL = asanaDefaultBaseURL
	}

	endpoint, err := url.JoinPath(baseURL, asanaWorkspacesSegment)
	if err != nil {
		return nil, fmt.Errorf("cannot build asana organizations URL: %w", err)
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("cannot parse asana organizations URL: %w", err)
	}

	q := parsed.Query()
	q.Set("limit", "100")
	parsed.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create asana organizations request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch asana organizations: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch asana organizations: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Data []struct {
			GID  string `json:"gid"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("cannot decode asana organizations response: %w", err)
	}

	result := make([]Organization, len(body.Data))
	for i, w := range body.Data {
		displayName := w.Name
		if displayName == "" {
			displayName = w.GID
		}

		result[i] = Organization{Slug: w.GID, DisplayName: displayName}
	}

	return result, nil
}
