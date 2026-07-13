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

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

type orgScanner struct {
	httpClient *http.Client
	org        string
	logger     *log.Logger
}

func newOrgScanner(httpClient *http.Client, org string, logger *log.Logger) *orgScanner {
	return &orgScanner{
		httpClient: httpClient,
		org:        org,
		logger:     logger,
	}
}

type githubOrganization struct {
	TwoFactorRequirementEnabled        *bool  `json:"two_factor_requirement_enabled"`
	DefaultRepositoryPermission        string `json:"default_repository_permission"`
	MembersCanCreatePublicRepositories *bool  `json:"members_can_create_public_repositories"`
	MembersCanChangeRepoVisibility     *bool  `json:"members_can_change_repo_visibility"`
}

func (s *orgScanner) scan(ctx context.Context) (*FactSheet, error) {
	sheet := &FactSheet{
		GitHubOrg: s.org,
	}

	org, err := s.fetchOrganization(ctx)
	if err != nil {
		return nil, err
	}

	sheet.Facts = append(sheet.Facts, orgFacts(org)...)

	disabled2FA, err := s.count2FADisabled(ctx)
	if err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot list 2FA-disabled members")
	} else {
		sheet.Facts = append(sheet.Facts, Fact{
			FactID:  "f-members-without-2fa",
			FactKey: "org_no_2fa_members",
			Scope:   "org",
			Value:   disabled2FA,
			APIRef:  "GET /orgs/{org}/members?filter=2fa_disabled",
		})
	}

	adminCount, memberCount, err := s.countAdmins(ctx)
	if err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot count org admins")
	} else {
		sheet.Facts = append(sheet.Facts, Fact{
			FactID:  "f-admin-ratio",
			FactKey: "org_admin_minimization",
			Scope:   "org",
			Value: map[string]int{
				"admins":  adminCount,
				"members": memberCount,
			},
			APIRef: "GET /orgs/{org}/members + memberships",
		})
	}

	publicRepos, totalRepos, err := s.countPublicRepos(ctx)
	if err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot list organization repositories")
	} else {
		sheet.ReposScanned = totalRepos
		sheet.Facts = append(sheet.Facts, Fact{
			FactID:  "f-public-repos",
			FactKey: "org_public_repos",
			Scope:   "org",
			Value: map[string]int{
				"public": publicRepos,
				"total":  totalRepos,
			},
			APIRef: "GET /orgs/{org}/repos",
		})
	}

	return sheet, nil
}

func orgFacts(org *githubOrganization) []Fact {
	facts := make([]Fact, 0, 4)

	if org.TwoFactorRequirementEnabled != nil {
		facts = append(facts, Fact{
			FactID:  "f-org-mfa-required",
			FactKey: "org_mfa_required",
			Scope:   "org",
			Value:   *org.TwoFactorRequirementEnabled,
			APIRef:  "GET /orgs/{org}",
		})
	}

	if org.DefaultRepositoryPermission != "" {
		facts = append(facts, Fact{
			FactID:  "f-base-permissions",
			FactKey: "org_base_permissions",
			Scope:   "org",
			Value:   org.DefaultRepositoryPermission,
			APIRef:  "GET /orgs/{org}",
		})
	}

	if org.MembersCanCreatePublicRepositories != nil {
		facts = append(facts, Fact{
			FactID:  "f-no-public-repo-creation",
			FactKey: "org_no_public_repo_creation",
			Scope:   "org",
			Value:   !*org.MembersCanCreatePublicRepositories,
			APIRef:  "GET /orgs/{org}",
		})
	}

	if org.MembersCanChangeRepoVisibility != nil {
		facts = append(facts, Fact{
			FactID:  "f-no-visibility-change",
			FactKey: "org_no_visibility_change",
			Scope:   "org",
			Value:   !*org.MembersCanChangeRepoVisibility,
			APIRef:  "GET /orgs/{org}",
		})
	}

	return facts
}

func (s *orgScanner) fetchOrganization(ctx context.Context) (*githubOrganization, error) {
	endpoint, err := url.JoinPath("https://api.github.com", "orgs", url.PathEscape(s.org))
	if err != nil {
		return nil, fmt.Errorf("cannot build github org URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create github org request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute github org request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch github organization: unexpected status %d", resp.StatusCode)
	}

	var org githubOrganization
	if err := json.NewDecoder(resp.Body).Decode(&org); err != nil {
		return nil, fmt.Errorf("cannot decode github organization response: %w", err)
	}

	return &org, nil
}

func (s *orgScanner) count2FADisabled(ctx context.Context) (int, error) {
	driver := drivers.NewGitHubDriver(s.httpClient, s.org, s.logger)

	records, err := driver.ListAccounts(ctx)
	if err != nil {
		return 0, err
	}

	count := 0

	for _, record := range records {
		if record.MFAStatus == coredata.MFAStatusDisabled {
			count++
		}
	}

	return count, nil
}

func (s *orgScanner) countAdmins(ctx context.Context) (int, int, error) {
	driver := drivers.NewGitHubDriver(s.httpClient, s.org, s.logger)

	records, err := driver.ListAccounts(ctx)
	if err != nil {
		return 0, 0, err
	}

	admins := 0

	for _, record := range records {
		if record.IsAdmin {
			admins++
		}
	}

	return admins, len(records), nil
}

type repoPageItem struct {
	Private bool `json:"private"`
}

func (s *orgScanner) countPublicRepos(ctx context.Context) (int, int, error) {
	public := 0
	total := 0

	endpoint, err := url.JoinPath("https://api.github.com", "orgs", url.PathEscape(s.org), "repos")
	if err != nil {
		return 0, 0, fmt.Errorf("cannot build github repos URL: %w", err)
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot parse github repos URL: %w", err)
	}

	q := parsed.Query()
	q.Set("per_page", "100")
	parsed.RawQuery = q.Encode()
	next := parsed.String()

	for page := 0; page < 10 && next != ""; page++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, next, nil)
		if err != nil {
			return 0, 0, fmt.Errorf("cannot create github repos request: %w", err)
		}

		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			return 0, 0, fmt.Errorf("cannot execute github repos request: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()

			return 0, 0, fmt.Errorf("cannot fetch github repos: unexpected status %d", resp.StatusCode)
		}

		var repos []repoPageItem
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			_ = resp.Body.Close()

			return 0, 0, fmt.Errorf("cannot decode github repos response: %w", err)
		}

		_ = resp.Body.Close()

		for _, repo := range repos {
			total++

			if !repo.Private {
				public++
			}
		}

		next = parseLinkNext(resp.Header.Get("Link"))
	}

	return public, total, nil
}

func parseLinkNext(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}

	for part := range strings.SplitSeq(linkHeader, ",") {
		part = strings.TrimSpace(part)
		if !strings.Contains(part, `rel="next"`) {
			continue
		}

		start := strings.Index(part, "<")
		end := strings.Index(part, ">")

		if start >= 0 && end > start {
			return part[start+1 : end]
		}
	}

	return ""
}

// evaluateAdminMinimization returns measure state for admin ratio fact.
func evaluateAdminMinimization(value any) coredata.MeasureState {
	m, ok := value.(map[string]any)
	if !ok {
		return coredata.MeasureStateUnknown
	}

	admins, _ := toInt(m["admins"])
	members, _ := toInt(m["members"])

	if members == 0 {
		return coredata.MeasureStateUnknown
	}

	if admins <= 3 {
		return coredata.MeasureStateImplemented
	}

	if float64(admins)/float64(members) <= 0.15 {
		return coredata.MeasureStateImplemented
	}

	return coredata.MeasureStateNotImplemented
}

func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case json.Number:
		i, err := n.Int64()

		return int(i), err == nil
	case string:
		i, err := strconv.Atoi(n)

		return i, err == nil
	default:
		return 0, false
	}
}
