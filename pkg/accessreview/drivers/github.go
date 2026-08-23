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
	"fmt"
	"net/http"

	"encoding/json"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/rfc5988"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// GitHub REST path segments joined onto the driver's base URL. They are
// single segments rather than whole paths because every GitHub endpoint the
// driver calls interleaves them with an escaped org or login.
const (
	githubOrgsSegment        = "orgs"
	githubUsersSegment       = "users"
	githubMembersSegment     = "members"
	githubMembershipsSegment = "memberships"

	// Whole path, unlike the segments above: the picker's endpoint
	// interleaves no org or login.
	githubUserOrgsPath = "user/orgs"
)

// githubDefaultBaseURL is the GitHub REST API root. It backs only the
// exported ListGitHubOrganizations, and only when its caller resolves no
// APIBase for the provider (unregistered, or registered without one).
// Every other path goes through the injected baseURL instead.
const githubDefaultBaseURL = "https://api.github.com"

// GitHubDriver fetches organization members from the GitHub REST API
// using a pre-authenticated HTTP client (Bearer token).
type GitHubDriver struct {
	httpClient *http.Client
	org        string
	logger     *log.Logger
	baseURL    string
}

var _ Driver = (*GitHubDriver)(nil)

type githubMember struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Type  string `json:"type"`
}

type githubMembership struct {
	Role  string `json:"role"`
	State string `json:"state"`
}

type githubUserProfile struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	Type      string `json:"type"`
}

// NewGitHubDriver builds a driver against baseURL, the GitHub REST API
// origin (e.g. https://api.github.com).
func NewGitHubDriver(httpClient *http.Client, org string, logger *log.Logger, baseURL string) *GitHubDriver {
	return &GitHubDriver{
		httpClient: httpClient,
		org:        org,
		logger:     logger,
		baseURL:    baseURL,
	}
}

func (d *GitHubDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	members, err := d.fetchAllMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch github org members: %w", err)
	}

	no2FASet, err := d.fetchAll2FADisabledLogins(ctx)
	if err != nil {
		// If the 2FA list fetch fails (e.g. insufficient permissions),
		// we still proceed but mark MFA as Unknown for all members.
		no2FASet = nil
	}

	var records []AccountRecord

	for _, m := range members {
		membership, err := d.fetchMembership(ctx, m.Login)
		if err != nil {
			d.logger.WarnCtx(
				ctx,
				"cannot fetch github membership, skipping member",
				log.Error(err),
			)

			continue
		}

		profile, err := d.fetchUserProfile(ctx, m.Login)
		if err != nil {
			d.logger.WarnCtx(
				ctx,
				"cannot fetch github user profile, skipping member",
				log.Error(err),
			)

			continue
		}

		fullName := profile.Name
		if fullName == "" {
			fullName = m.Login
		}

		accountType := coredata.AccessReviewEntryAccountTypeUser
		if m.Type == "Bot" {
			accountType = coredata.AccessReviewEntryAccountTypeServiceAccount
		}

		mfaStatus := coredata.MFAStatusUnknown

		if no2FASet != nil {
			if no2FASet[m.Login] {
				mfaStatus = coredata.MFAStatusDisabled
			} else {
				mfaStatus = coredata.MFAStatusEnabled
			}
		}

		role := strings.TrimSpace(membership.Role)

		roles := []string{}
		if role != "" {
			roles = []string{role}
		}

		record := AccountRecord{
			Email:       profile.Email,
			FullName:    fullName,
			Roles:       roles,
			Active:      new(membership.State == "active"),
			IsAdmin:     new(membership.Role == "admin"),
			MFAStatus:   mfaStatus,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: accountType,
			ExternalID:  strconv.FormatInt(m.ID, 10),
		}

		if profile.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, profile.CreatedAt); err == nil {
				record.CreatedAt = &t
			}
		}

		records = append(records, record)
	}

	return records, nil
}

func (d *GitHubDriver) fetchAllMembers(ctx context.Context) ([]githubMember, error) {
	var members []githubMember

	u, err := url.JoinPath(d.baseURL, githubOrgsSegment, url.PathEscape(d.org), githubMembersSegment)
	if err != nil {
		return nil, fmt.Errorf("cannot build github members URL: %w", err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("cannot parse github members URL: %w", err)
	}

	q := parsed.Query()
	q.Set("per_page", "100")
	parsed.RawQuery = q.Encode()
	endpoint := parsed.String()

	for range maxPaginationPages {
		page, nextURL, err := d.fetchMembersPage(ctx, endpoint)
		if err != nil {
			return nil, err
		}

		members = append(members, page...)

		if nextURL == "" {
			return members, nil
		}

		endpoint = nextURL
	}

	return nil, fmt.Errorf("cannot list all github members: %w", ErrPaginationLimitReached)
}

func (d *GitHubDriver) fetchMembersPage(ctx context.Context, url string) ([]githubMember, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("cannot create github members request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("cannot execute github members request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("cannot fetch github members: unexpected status %d", httpResp.StatusCode)
	}

	var members []githubMember
	if err := json.NewDecoder(httpResp.Body).Decode(&members); err != nil {
		return nil, "", fmt.Errorf("cannot decode github members response: %w", err)
	}

	nextURL := rfc5988.FindByRel(httpResp.Header.Get("Link"), "next")
	if nextURL != "" {
		nextURL, err = sameHostNextPageURL("github", d.baseURL, nextURL)
		if err != nil {
			return nil, "", err
		}
	}

	return members, nextURL, nil
}

func (d *GitHubDriver) fetchAll2FADisabledLogins(ctx context.Context) (map[string]bool, error) {
	set := make(map[string]bool)

	u, err := url.JoinPath(d.baseURL, githubOrgsSegment, url.PathEscape(d.org), githubMembersSegment)
	if err != nil {
		return nil, fmt.Errorf("cannot build github 2fa-disabled URL: %w", err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("cannot parse github 2fa-disabled URL: %w", err)
	}

	q := parsed.Query()
	q.Set("filter", "2fa_disabled")
	q.Set("per_page", "100")
	parsed.RawQuery = q.Encode()
	endpoint := parsed.String()

	for range maxPaginationPages {
		page, nextURL, err := d.fetchMembersPage(ctx, endpoint)
		if err != nil {
			return nil, err
		}

		for _, m := range page {
			set[m.Login] = true
		}

		if nextURL == "" {
			return set, nil
		}

		endpoint = nextURL
	}

	return nil, fmt.Errorf("cannot list all github 2fa-disabled members: %w", ErrPaginationLimitReached)
}

func (d *GitHubDriver) fetchMembership(ctx context.Context, login string) (*githubMembership, error) {
	endpoint, err := url.JoinPath(d.baseURL, githubOrgsSegment, url.PathEscape(d.org), githubMembershipsSegment, url.PathEscape(login))
	if err != nil {
		return nil, fmt.Errorf("cannot build github membership URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create github membership request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute github membership request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch github membership for %s: unexpected status %d", login, httpResp.StatusCode)
	}

	var membership githubMembership
	if err := json.NewDecoder(httpResp.Body).Decode(&membership); err != nil {
		return nil, fmt.Errorf("cannot decode github membership response: %w", err)
	}

	return &membership, nil
}

func (d *GitHubDriver) fetchUserProfile(ctx context.Context, login string) (*githubUserProfile, error) {
	endpoint, err := url.JoinPath(d.baseURL, githubUsersSegment, url.PathEscape(login))
	if err != nil {
		return nil, fmt.Errorf("cannot build github user profile URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create github user profile request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute github user profile request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch github user profile for %s: unexpected status %d", login, httpResp.StatusCode)
	}

	var profile githubUserProfile
	if err := json.NewDecoder(httpResp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("cannot decode github user profile response: %w", err)
	}

	return &profile, nil
}

// githubNameResolver resolves the GitHub organization name.
type githubNameResolver struct {
	httpClient *http.Client
	org        string
	baseURL    string
}

// NewGitHubNameResolver resolves the org name against baseURL, the GitHub
// REST API origin (e.g. https://api.github.com).
func NewGitHubNameResolver(httpClient *http.Client, org, baseURL string) NameResolver {
	return &githubNameResolver{httpClient: httpClient, org: org, baseURL: baseURL}
}

func (r *githubNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.org == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath(r.baseURL, githubOrgsSegment, url.PathEscape(r.org))
	if err != nil {
		return "", fmt.Errorf("cannot build github organization URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create github organization request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute github organization request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nameStatusError("github organization", httpResp.StatusCode)
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode github organization response: %w", err)
	}

	if resp.Name == "" {
		return r.org, nil
	}

	return resp.Name, nil
}

// ListGitHubOrganizations fetches the organizations the authenticated
// GitHub user belongs to, from baseURL ("" for GitHub.com).
func ListGitHubOrganizations(ctx context.Context, httpClient *http.Client, baseURL string) ([]Organization, error) {
	if baseURL == "" {
		baseURL = githubDefaultBaseURL
	}

	endpoint, err := url.JoinPath(baseURL, githubUserOrgsPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build github organizations URL: %w", err)
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("cannot parse github organizations URL: %w", err)
	}

	q := parsed.Query()
	q.Set("per_page", "100")
	parsed.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create github organizations request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch github organizations: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch github organizations: unexpected status %d", resp.StatusCode)
	}

	var orgs []struct {
		Login string `json:"login"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return nil, fmt.Errorf("cannot decode github organizations response: %w", err)
	}

	result := make([]Organization, len(orgs))
	for i, org := range orgs {
		displayName := org.Name
		if displayName == "" {
			displayName = org.Login
		}

		result[i] = Organization{Slug: org.Login, DisplayName: displayName}
	}

	return result, nil
}

func githubSource() Factory {
	return provider.Over(func(
		ctx context.Context,
		credential connector.HTTPCredential,
		opened *provider.Handle,
		logger *log.Logger,
	) (Driver, error) {
		driver, err := githubSourceDriver(ctx, credential.Client, opened.Connector, logger, opened.Endpoints)
		if err != nil {
			return nil, err
		}

		return capable(
			driver,
			githubSourceNameResolver(ctx, credential.Client, opened.Connector, logger, opened.Endpoints),
			organizationListerFunc(func(ctx context.Context) ([]Organization, error) {
				return ListGitHubOrganizations(ctx, credential.Client, organizationsBase(opened.Endpoints))
			}),
		), nil
	})
}

func githubSourceDriver(
	_ context.Context,
	c *http.Client,
	conn *coredata.Connector,
	logger *log.Logger,
	ep provider.Endpoints,
) (Driver, error) {
	s, err := coredata.ConnectorSettings[coredata.GitHubConnectorSettings](conn)
	if err != nil {
		return nil, fmt.Errorf("cannot read github connector settings: %w", err)
	}

	if s.Organization == "" {
		return nil, fmt.Errorf("cannot create github driver: organization is required")
	}

	return NewGitHubDriver(c, s.Organization, logger.Named("github"), ep.APIBase), nil
}

func githubSourceNameResolver(
	ctx context.Context,
	c *http.Client,
	conn *coredata.Connector,
	logger *log.Logger,
	ep provider.Endpoints,
) NameResolver {
	s, err := coredata.ConnectorSettings[coredata.GitHubConnectorSettings](conn)
	if err != nil {
		logger.ErrorCtx(ctx, "cannot read github connector settings", log.Error(err))
		return nil
	}

	return NewGitHubNameResolver(c, s.Organization, ep.APIBase)
}
