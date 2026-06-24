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
	"time"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/rfc5988"
)

// GitHubDriver fetches organization members from the GitHub REST API
// using a pre-authenticated HTTP client (Bearer token).
type GitHubDriver struct {
	httpClient *http.Client
	org        string
	logger     *log.Logger
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

func NewGitHubDriver(httpClient *http.Client, org string, logger *log.Logger) *GitHubDriver {
	return &GitHubDriver{
		httpClient: httpClient,
		org:        org,
		logger:     logger,
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
			IsAdmin:     membership.Role == "admin",
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

	u, err := url.JoinPath("https://api.github.com", "orgs", url.PathEscape(d.org), "members")
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

	return members, nextURL, nil
}

func (d *GitHubDriver) fetchAll2FADisabledLogins(ctx context.Context) (map[string]bool, error) {
	set := make(map[string]bool)

	u, err := url.JoinPath("https://api.github.com", "orgs", url.PathEscape(d.org), "members")
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
	endpoint, err := url.JoinPath("https://api.github.com", "orgs", url.PathEscape(d.org), "memberships", url.PathEscape(login))
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
	endpoint, err := url.JoinPath("https://api.github.com", "users", url.PathEscape(login))
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
