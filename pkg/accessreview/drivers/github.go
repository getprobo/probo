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
	"bytes"
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

// GitHub REST path segments joined onto the driver's base URL. They are
// single segments rather than whole paths because every GitHub endpoint the
// driver calls interleaves them with an escaped org or login.
const (
	githubOrgsSegment                     = "orgs"
	githubUsersSegment                    = "users"
	githubMembersSegment                  = "members"
	githubMembershipsSegment              = "memberships"
	githubAuditLogSegment                 = "audit-log"
	githubInstallationsSegment            = "installations"
	githubPersonalAccessTokensSegment     = "personal-access-tokens"
	githubCredentialAuthorizationsSegment = "credential-authorizations"

	// Whole paths: the picker's endpoint and GraphQL interleave no
	// org or login.
	githubUserOrgsPath = "user/orgs"
	githubGraphqlPath  = "graphql"

	// githubAuditLogMaxPages bounds how far we walk the org audit log
	// looking for each member's latest event. The log can span years;
	// scanning it to completion starves the member and service-account
	// fetches inside the campaign timeout. Ten pages is the 1_000 most
	// recent events; anyone without activity in that window keeps
	// LastLogin unset.
	githubAuditLogMaxPages = 10

	githubAuditLogTimeout = 10 * time.Second
)

// githubDefaultBaseURL is the GitHub REST API root. It backs only the
// exported ListGitHubOrganizations, and only when its caller resolves no
// APIBase for the provider (unregistered, or registered without one).
// Every other path goes through the injected baseURL instead.
const githubDefaultBaseURL = "https://api.github.com"

// GitHubDriver fetches organization members from the GitHub REST API
// using a pre-authenticated HTTP client (Bearer token).
//
// LastLogin is the latest organization audit-log event for the member
// (pull requests and other audited actions, not a GitHub login). The
// audit-log API requires GitHub Enterprise Cloud and `read:audit_log`
// (OAuth) or Organization Administration: read (GitHub App). Missing
// permission, a non-Enterprise org, or a log too long to scan inside
// githubAuditLogTimeout leaves LastLogin nil rather than failing the
// fetch. Only the most recent githubAuditLogMaxPages of events are
// read; the driver does not walk the full history to find every member.
//
// Installed GitHub Apps, fine-grained PATs, SAML-authorized tokens, and
// deploy keys are appended as service accounts when the token can read
// them. Those calls need more than `read:org`: Organization
// Administration read (apps), Personal access tokens read (fine-grained
// PATs; GitHub App only), an Enterprise Cloud SAML org
// (credential-authorizations), and Repository Administration read
// (deploy keys). Missing permission omits those records instead of
// failing the member list.
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

type githubAuditEvent struct {
	Timestamp int64  `json:"@timestamp"`
	CreatedAt int64  `json:"created_at"`
	Actor     string `json:"actor"`
	ActorID   int64  `json:"actor_id"`
}

type (
	gitHubVerifiedDomainEmailsRequest struct {
		Query     string                              `json:"query"`
		Variables gitHubVerifiedDomainEmailsVariables `json:"variables"`
	}

	gitHubVerifiedDomainEmailsVariables struct {
		Org   string  `json:"org"`
		After *string `json:"after"`
	}

	gitHubVerifiedDomainEmailsResponse struct {
		Data struct {
			Organization *struct {
				MembersWithRole struct {
					Nodes []struct {
						Login                            string   `json:"login"`
						OrganizationVerifiedDomainEmails []string `json:"organizationVerifiedDomainEmails"`
					} `json:"nodes"`
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
				} `json:"membersWithRole"`
			} `json:"organization"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
)

// githubVerifiedDomainEmailsQuery is the same field GitHub's org-admin
// members page uses. REST GET /users/{login} only returns a public
// profile email, which most people hide.
const githubVerifiedDomainEmailsQuery = `query AccessReviewGitHubVerifiedDomainEmails($org: String!, $after: String) { organization(login: $org) { membersWithRole(first: 100, after: $after) { nodes { login organizationVerifiedDomainEmails(login: $org) } pageInfo { hasNextPage endCursor } } } }`

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

	// Verified-domain emails are what org admins see in GitHub's
	// members UI. The query needs a verified domain (and typically
	// Enterprise Cloud); if it fails we still list members with
	// whatever public profile email exists.
	verifiedEmails, err := d.fetchVerifiedDomainEmails(ctx)
	if err != nil {
		d.logger.WarnCtx(
			ctx,
			"cannot fetch github verified domain emails",
			log.Error(err),
		)

		verifiedEmails = nil
	}

	var records []AccountRecord

	for _, m := range members {
		records = append(records, d.memberRecord(ctx, m, no2FASet, verifiedEmails))
	}

	// Latest audit-log event per member is the closest GitHub gives us
	// to last activity. Cap time and pages so a long log cannot eat the
	// campaign fetch budget and drop members. Team orgs and tokens
	// without audit-log access 403 here; keep listing members with
	// LastLogin unset.
	auditCtx, cancelAudit := context.WithTimeout(ctx, githubAuditLogTimeout)
	lastByLogin, lastByID, err := d.fetchLastActivity(auditCtx, members)
	cancelAudit()

	if err != nil {
		d.logger.WarnCtx(
			ctx,
			"cannot fetch github audit log last activity",
			log.Error(err),
		)
	}

	for i, m := range members {
		githubApplyLastActivity(&records[i], m, lastByLogin, lastByID)
	}

	return d.appendServiceAccounts(ctx, records), nil
}

func (d *GitHubDriver) memberRecord(
	ctx context.Context,
	m githubMember,
	no2FASet map[string]bool,
	verifiedEmails map[string]string,
) AccountRecord {
	membership, err := d.fetchMembership(ctx, m.Login)
	if err != nil {
		d.logger.WarnCtx(
			ctx,
			"cannot fetch github membership",
			log.Error(err),
		)
	}

	profile, err := d.fetchUserProfile(ctx, m.Login)
	if err != nil {
		d.logger.WarnCtx(
			ctx,
			"cannot fetch github user profile",
			log.Error(err),
		)

		profile = nil
	}

	fullName := m.Login
	if profile != nil && strings.TrimSpace(profile.Name) != "" {
		fullName = profile.Name
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

	roles := []string{}
	var active *bool
	var isAdmin *bool
	if membership != nil {
		role := strings.TrimSpace(membership.Role)
		if role != "" {
			roles = []string{role}
		}

		if membership.State != "" {
			active = new(membership.State == "active")
		}

		if membership.Role != "" {
			isAdmin = new(membership.Role == "admin")
		}
	}

	email := verifiedEmails[m.Login]
	if email == "" && profile != nil {
		email = profile.Email
	}

	record := AccountRecord{
		Email:       email,
		FullName:    fullName,
		Roles:       roles,
		Active:      active,
		IsAdmin:     isAdmin,
		MFAStatus:   mfaStatus,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType: accountType,
		ExternalID:  strconv.FormatInt(m.ID, 10),
	}

	if profile != nil && profile.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, profile.CreatedAt); err == nil {
			record.CreatedAt = &t
		}
	}

	return record
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

func (d *GitHubDriver) fetchLastActivity(
	ctx context.Context,
	members []githubMember,
) (map[string]time.Time, map[int64]time.Time, error) {
	byLogin := make(map[string]time.Time)
	byID := make(map[int64]time.Time)

	if len(members) == 0 {
		return byLogin, byID, nil
	}

	wantedLogin := make(map[string]struct{}, len(members))
	wantedID := make(map[int64]struct{}, len(members))
	for _, m := range members {
		if login := strings.ToLower(strings.TrimSpace(m.Login)); login != "" {
			wantedLogin[login] = struct{}{}
		}

		if m.ID != 0 {
			wantedID[m.ID] = struct{}{}
		}
	}

	u, err := url.JoinPath(d.baseURL, githubOrgsSegment, url.PathEscape(d.org), githubAuditLogSegment)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot build github audit log URL: %w", err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse github audit log URL: %w", err)
	}

	q := parsed.Query()
	q.Set("per_page", "100")
	q.Set("order", "desc")
	parsed.RawQuery = q.Encode()
	endpoint := parsed.String()

	for range githubAuditLogMaxPages {
		page, nextURL, err := d.fetchAuditLogPage(ctx, endpoint)
		if err != nil {
			return byLogin, byID, err
		}

		for _, ev := range page {
			if !githubWantedAuditActor(ev, wantedLogin, wantedID) {
				continue
			}

			githubRecordLastActivity(byLogin, byID, ev)
		}

		if githubLastActivityComplete(members, byLogin, byID) || nextURL == "" {
			return byLogin, byID, nil
		}

		endpoint = nextURL
	}

	return byLogin, byID, nil
}

func (d *GitHubDriver) fetchAuditLogPage(ctx context.Context, pageURL string) ([]githubAuditEvent, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("cannot create github audit log request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("cannot execute github audit log request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("cannot fetch github audit log: unexpected status %d", httpResp.StatusCode)
	}

	var events []githubAuditEvent
	if err := json.NewDecoder(httpResp.Body).Decode(&events); err != nil {
		return nil, "", fmt.Errorf("cannot decode github audit log response: %w", err)
	}

	nextURL := rfc5988.FindByRel(httpResp.Header.Get("Link"), "next")
	if nextURL != "" {
		nextURL, err = sameHostNextPageURL("github", d.baseURL, nextURL)
		if err != nil {
			return nil, "", err
		}
	}

	return events, nextURL, nil
}

func githubAuditEventTime(ev githubAuditEvent) (time.Time, bool) {
	ms := ev.Timestamp
	if ms == 0 {
		ms = ev.CreatedAt
	}

	if ms == 0 {
		return time.Time{}, false
	}

	if ms < 1e12 {
		return time.Unix(ms, 0).UTC(), true
	}

	return time.UnixMilli(ms).UTC(), true
}

func githubWantedAuditActor(
	ev githubAuditEvent,
	wantedLogin map[string]struct{},
	wantedID map[int64]struct{},
) bool {
	if ev.ActorID != 0 {
		if _, ok := wantedID[ev.ActorID]; ok {
			return true
		}
	}

	login := strings.ToLower(strings.TrimSpace(ev.Actor))
	if login == "" {
		return false
	}

	_, ok := wantedLogin[login]

	return ok
}

func githubRecordLastActivity(
	byLogin map[string]time.Time,
	byID map[int64]time.Time,
	ev githubAuditEvent,
) {
	t, ok := githubAuditEventTime(ev)
	if !ok {
		return
	}

	if ev.ActorID != 0 {
		if _, exists := byID[ev.ActorID]; !exists {
			byID[ev.ActorID] = t
		}
	}

	login := strings.ToLower(strings.TrimSpace(ev.Actor))
	if login == "" {
		return
	}

	if _, exists := byLogin[login]; !exists {
		byLogin[login] = t
	}
}

func githubLastActivityComplete(
	members []githubMember,
	byLogin map[string]time.Time,
	byID map[int64]time.Time,
) bool {
	if len(members) == 0 {
		return false
	}

	for _, m := range members {
		if _, ok := byID[m.ID]; ok {
			continue
		}

		if _, ok := byLogin[strings.ToLower(m.Login)]; ok {
			continue
		}

		return false
	}

	return true
}

func githubApplyLastActivity(
	record *AccountRecord,
	member githubMember,
	byLogin map[string]time.Time,
	byID map[int64]time.Time,
) {
	if t, ok := byID[member.ID]; ok {
		record.LastLogin = new(t)

		return
	}

	if member.Login == "" {
		return
	}

	if t, ok := byLogin[strings.ToLower(member.Login)]; ok {
		record.LastLogin = new(t)
	}
}

func (d *GitHubDriver) fetchVerifiedDomainEmails(ctx context.Context) (map[string]string, error) {
	emails := make(map[string]string)

	var after *string

	for range maxPaginationPages {
		page, next, err := d.fetchVerifiedDomainEmailsPage(ctx, after)
		if err != nil {
			return nil, err
		}

		for login, email := range page {
			emails[login] = email
		}

		if next == nil {
			return emails, nil
		}

		after = next
	}

	return nil, fmt.Errorf("cannot list all github verified domain emails: %w", ErrPaginationLimitReached)
}

func (d *GitHubDriver) fetchVerifiedDomainEmailsPage(
	ctx context.Context,
	after *string,
) (map[string]string, *string, error) {
	endpoint, err := url.JoinPath(d.baseURL, githubGraphqlPath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot build github graphql URL: %w", err)
	}

	payload, err := json.Marshal(
		gitHubVerifiedDomainEmailsRequest{
			Query: githubVerifiedDomainEmailsQuery,
			Variables: gitHubVerifiedDomainEmailsVariables{
				Org:   d.org,
				After: after,
			},
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot marshal github verified domain emails query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create github verified domain emails request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot execute github verified domain emails request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, nil, fmt.Errorf(
			"cannot fetch github verified domain emails: unexpected status %d",
			httpResp.StatusCode,
		)
	}

	var resp gitHubVerifiedDomainEmailsResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, nil, fmt.Errorf("cannot decode github verified domain emails response: %w", err)
	}

	if resp.Data.Organization == nil {
		if len(resp.Errors) > 0 {
			return nil, nil, fmt.Errorf("cannot fetch github verified domain emails: graphql error")
		}

		return map[string]string{}, nil, nil
	}

	emails := make(map[string]string, len(resp.Data.Organization.MembersWithRole.Nodes))
	for _, node := range resp.Data.Organization.MembersWithRole.Nodes {
		if node.Login == "" {
			continue
		}

		if email := firstGitHubVerifiedDomainEmail(node.OrganizationVerifiedDomainEmails); email != "" {
			emails[node.Login] = email
		}
	}

	pageInfo := resp.Data.Organization.MembersWithRole.PageInfo
	if !pageInfo.HasNextPage || pageInfo.EndCursor == "" {
		return emails, nil, nil
	}

	next := pageInfo.EndCursor

	return emails, &next, nil
}

func firstGitHubVerifiedDomainEmail(emails []string) string {
	for _, email := range emails {
		email = strings.TrimSpace(email)
		if email != "" {
			return email
		}
	}

	return ""
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
