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
	"strconv"
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
			d.logger.WarnCtx(ctx, "cannot fetch github membership, skipping member",
				log.Error(err),
			)
			continue
		}

		profile, err := d.fetchUserProfile(ctx, m.Login)
		if err != nil {
			d.logger.WarnCtx(ctx, "cannot fetch github user profile, skipping member",
				log.Error(err),
			)
			continue
		}

		fullName := profile.Name
		if fullName == "" {
			fullName = m.Login
		}

		accountType := coredata.AccessEntryAccountTypeUser
		if m.Type == "Bot" {
			accountType = coredata.AccessEntryAccountTypeServiceAccount
		}

		mfaStatus := coredata.MFAStatusUnknown
		if no2FASet != nil {
			if no2FASet[m.Login] {
				mfaStatus = coredata.MFAStatusDisabled
			} else {
				mfaStatus = coredata.MFAStatusEnabled
			}
		}

		record := AccountRecord{
			Email:       profile.Email,
			FullName:    fullName,
			Role:        membership.Role,
			Active:      new(membership.State == "active"),
			IsAdmin:     membership.Role == "admin",
			MFAStatus:   mfaStatus,
			AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
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

	url := fmt.Sprintf(
		"https://api.github.com/orgs/%s/members?per_page=100",
		d.org,
	)

	for range maxPaginationPages {
		page, nextURL, err := d.fetchMembersPage(ctx, url)
		if err != nil {
			return nil, err
		}

		members = append(members, page...)

		if nextURL == "" {
			return members, nil
		}
		url = nextURL
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

	url := fmt.Sprintf(
		"https://api.github.com/orgs/%s/members?filter=2fa_disabled&per_page=100",
		d.org,
	)

	for range maxPaginationPages {
		page, nextURL, err := d.fetchMembersPage(ctx, url)
		if err != nil {
			return nil, err
		}

		for _, m := range page {
			set[m.Login] = true
		}

		if nextURL == "" {
			return set, nil
		}
		url = nextURL
	}

	return nil, fmt.Errorf("cannot list all github 2fa-disabled members: %w", ErrPaginationLimitReached)
}

func (d *GitHubDriver) fetchMembership(ctx context.Context, login string) (*githubMembership, error) {
	url := fmt.Sprintf(
		"https://api.github.com/orgs/%s/memberships/%s",
		d.org,
		login,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
	url := fmt.Sprintf("https://api.github.com/users/%s", login)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
