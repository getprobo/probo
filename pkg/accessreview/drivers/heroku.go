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
	"net/url"
	"strings"
	"time"
)

const (
	// herokuPersonalAccountSlug is the reserved org-picker slug for a
	// personal Heroku account (one with no Team). Heroku Teams are an
	// opt-in paid construct, so a solo account has nothing in GET /teams;
	// selecting this entry runs the driver in personal mode (app owner +
	// collaborators) instead of team-member mode.
	herokuPersonalAccountSlug = "@personal"

	// herokuPersonalAccountDisplayName is the picker label and source name
	// shown for a personal Heroku account.
	herokuPersonalAccountDisplayName = "Personal account"

	// Platform API path segments joined onto the driver's base URL.
	herokuTeamsSegment         = "teams"
	herokuAppsSegment          = "apps"
	herokuMembersSegment       = "members"
	herokuCollaboratorsSegment = "collaborators"
)

// herokuDefaultBaseURL is the Heroku Platform API root. It backs only the
// exported ListHerokuOrganizations, and only when its caller resolves no
// APIBase for the provider (unregistered, or registered without one). Every
// driver path goes through the injected baseURL instead.
const herokuDefaultBaseURL = "https://api.heroku.com"

// HerokuDriver fetches members from the Heroku Platform API using a
// pre-authenticated HTTP client (Bearer token). Pagination is via Heroku's
// Range / Next-Range header pair (RFC 7233 style).
//
// The driver runs in one of two modes:
//   - team mode (teamID set): list the members of GET /teams/{id}/members.
//   - personal mode (teamID empty or the personal-account slug): a solo
//     Heroku account has no Team, so access is granted per-app; enumerate
//     the personal apps' owners and collaborators instead.
//
// Notes on data quality:
//   - The team-members endpoint does not expose suspension state, so
//     Active is left nil for v1.
//   - For federated teams the IdP is the source of truth for MFA, but
//     the API still reports `two_factor_authentication`. The driver
//     populates MFAStatus from that field and lets the access-review
//     UI surface federation context separately.
//   - The collaborators endpoint does not expose MFA, so personal-mode
//     records leave MFAStatus unknown.
type HerokuDriver struct {
	httpClient *http.Client
	teamID     string
	baseURL    string
}

var _ Driver = (*HerokuDriver)(nil)

// NewHerokuDriver builds a driver against baseURL, the Heroku Platform API
// origin (e.g. https://api.heroku.com).
func NewHerokuDriver(httpClient *http.Client, teamID, baseURL string) *HerokuDriver {
	return &HerokuDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		teamID:  teamID,
		baseURL: baseURL,
	}
}

type herokuTeamMember struct {
	ID                      string `json:"id"`
	Email                   string `json:"email"`
	Role                    string `json:"role"`
	TwoFactorAuthentication bool   `json:"two_factor_authentication"`
	Federated               bool   `json:"federated"`
	CreatedAt               string `json:"created_at"`
	User                    struct {
		Email string `json:"email"`
		ID    string `json:"id"`
		Name  string `json:"name"`
	} `json:"user"`
}

type herokuApp struct {
	ID    string `json:"id"`
	Owner struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"owner"`
	// Team is nil for personal apps and set for team-owned apps; we use it
	// to keep personal mode scoped to the user's own apps.
	Team *struct {
		ID string `json:"id"`
	} `json:"team"`
}

type herokuCollaborator struct {
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	User      struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

func (d *HerokuDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	if d.teamID == "" || d.teamID == herokuPersonalAccountSlug {
		return d.listPersonalAccounts(ctx)
	}

	return d.listTeamMembers(ctx)
}

func (d *HerokuDriver) listTeamMembers(ctx context.Context) ([]AccountRecord, error) {
	endpoint, err := url.JoinPath(d.baseURL, herokuTeamsSegment, url.PathEscape(d.teamID), herokuMembersSegment)
	if err != nil {
		return nil, fmt.Errorf("cannot build heroku members URL: %w", err)
	}

	members, err := herokuListAll[herokuTeamMember](ctx, d.httpClient, endpoint, "members")
	if err != nil {
		return nil, fmt.Errorf("cannot list heroku team members: %w", err)
	}

	var records []AccountRecord

	for _, m := range members {
		email := m.Email
		if email == "" {
			email = m.User.Email
		}

		fullName := m.User.Name

		mfaStatus := coredata.MFAStatusDisabled
		if m.TwoFactorAuthentication {
			mfaStatus = coredata.MFAStatusEnabled
		}

		isAdmin := m.Role == "admin" || m.Role == "owner"

		externalID := m.User.ID
		if externalID == "" {
			externalID = m.ID
		}

		role := strings.TrimSpace(m.Role)

		roles := []string{}
		if role != "" {
			roles = []string{role}
		}

		record := AccountRecord{
			Email:       email,
			FullName:    fullName,
			Roles:       roles,
			IsAdmin:     new(isAdmin),
			MFAStatus:   mfaStatus,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  externalID,
		}

		if m.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
				record.CreatedAt = &t
			}
		}

		records = append(records, record)
	}

	return records, nil
}

// listPersonalAccounts enumerates the people with access to a personal
// (non-team) Heroku account: the owner of each personal app plus every
// collaborator on it, deduplicated by Heroku user ID. Personal accounts
// have no Team, so there are no team members to list; access is granted
// per-app via collaborators.
func (d *HerokuDriver) listPersonalAccounts(ctx context.Context) ([]AccountRecord, error) {
	// GET /apps returns every app the token can see, both owned and
	// collaborated-on. We skip team-owned apps below (those belong to a
	// team-scoped review); the remaining personal apps are the account
	// under review. Scoping strictly to apps the connector owns would need
	// the account ID from GET /account, which requires the identity OAuth
	// scope we deliberately do not request.
	appsURL, err := url.JoinPath(d.baseURL, herokuAppsSegment)
	if err != nil {
		return nil, fmt.Errorf("cannot build heroku apps URL: %w", err)
	}

	apps, err := herokuListAll[herokuApp](ctx, d.httpClient, appsURL, "apps")
	if err != nil {
		return nil, fmt.Errorf("cannot list heroku personal apps: %w", err)
	}

	var records []AccountRecord

	// Dedupe by Heroku user ID across all apps: a user collaborating on
	// several apps is one account in the review. An admin grant on any
	// app wins.
	seen := make(map[string]int)

	upsert := func(rec AccountRecord) {
		key := rec.ExternalID
		if key == "" {
			key = rec.Email
		}

		if key == "" {
			return
		}

		if i, ok := seen[key]; ok {
			if rec.IsAdmin != nil && *rec.IsAdmin {
				records[i].IsAdmin = new(true)
			}

			return
		}

		seen[key] = len(records)
		records = append(records, rec)
	}

	for _, app := range apps {
		// Skip team-owned apps; those belong to a team-scoped review.
		if app.Team != nil {
			continue
		}

		// The owner always has access, whether or not they also appear in
		// the collaborators list.
		upsert(herokuPersonalRecord(app.Owner.ID, app.Owner.Email, "owner", true))

		endpoint, err := url.JoinPath(d.baseURL, herokuAppsSegment, url.PathEscape(app.ID), herokuCollaboratorsSegment)
		if err != nil {
			return nil, fmt.Errorf("cannot build heroku collaborators URL: %w", err)
		}

		collaborators, err := herokuListAll[herokuCollaborator](ctx, d.httpClient, endpoint, "collaborators")
		if err != nil {
			return nil, fmt.Errorf("cannot list heroku collaborators for app %q: %w", app.ID, err)
		}

		for _, c := range collaborators {
			record := herokuPersonalRecord(c.User.ID, c.User.Email, c.Role, c.Role == "admin" || c.Role == "owner")

			if c.CreatedAt != "" {
				if t, err := time.Parse(time.RFC3339, c.CreatedAt); err == nil {
					record.CreatedAt = &t
				}
			}

			upsert(record)
		}
	}

	return records, nil
}

// herokuPersonalRecord builds an AccountRecord for a person with access to a
// personal Heroku app. These endpoints expose no display name or MFA signal,
// so the email doubles as the full name and MFA is left unknown.
func herokuPersonalRecord(externalID, email, role string, isAdmin bool) AccountRecord {
	role = strings.TrimSpace(role)

	roles := []string{}
	if role != "" {
		roles = []string{role}
	}

	return AccountRecord{
		Email:       email,
		FullName:    email,
		Roles:       roles,
		IsAdmin:     new(isAdmin),
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:  externalID,
	}
}

// herokuListAll fetches every page of a Heroku collection endpoint,
// following the Range / Next-Range pagination header pair, and decodes each
// page into T. label names the resource in error messages (e.g. "members").
func herokuListAll[T any](ctx context.Context, client *http.Client, endpoint, label string) ([]T, error) {
	var all []T

	rangeHeader := ""

	for range maxPaginationPages {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot create heroku %s request: %w", label, err)
		}

		req.Header.Set("Accept", "application/vnd.heroku+json; version=3")

		if rangeHeader != "" {
			req.Header.Set("Range", rangeHeader)
		}

		httpResp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("cannot execute heroku %s request: %w", label, err)
		}

		// Heroku returns 206 Partial Content for ranged responses with more
		// pages, and 200 OK for the final/only page.
		if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
			_ = httpResp.Body.Close()
			return nil, fmt.Errorf("cannot fetch heroku %s: unexpected status %d", label, httpResp.StatusCode)
		}

		var page []T
		if err := json.NewDecoder(httpResp.Body).Decode(&page); err != nil {
			_ = httpResp.Body.Close()
			return nil, fmt.Errorf("cannot decode heroku %s response: %w", label, err)
		}

		nextRange := httpResp.Header.Get("Next-Range")
		_ = httpResp.Body.Close()

		all = append(all, page...)

		if nextRange == "" {
			return all, nil
		}

		rangeHeader = nextRange
	}

	return nil, fmt.Errorf("cannot list all heroku %s: %w", label, ErrPaginationLimitReached)
}

// herokuNameResolver resolves the Heroku team name.
type herokuNameResolver struct {
	httpClient *http.Client
	teamID     string
	baseURL    string
}

// NewHerokuNameResolver resolves the team name against baseURL, the Heroku
// Platform API origin (e.g. https://api.heroku.com).
func NewHerokuNameResolver(httpClient *http.Client, teamID, baseURL string) NameResolver {
	return &herokuNameResolver{httpClient: httpClient, teamID: teamID, baseURL: baseURL}
}

func (r *herokuNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.teamID == "" {
		return "", nil
	}

	// A personal account has no Team to name; short-circuit before hitting
	// GET /teams/@personal, which 404s and would loop the source-name worker.
	if r.teamID == herokuPersonalAccountSlug {
		return herokuPersonalAccountDisplayName, nil
	}

	endpoint, err := url.JoinPath(r.baseURL, herokuTeamsSegment, url.PathEscape(r.teamID))
	if err != nil {
		return "", fmt.Errorf("cannot build heroku team URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create heroku team request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute heroku team request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nameStatusError("heroku team", httpResp.StatusCode)
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode heroku team response: %w", err)
	}

	return resp.Name, nil
}

// ListHerokuOrganizations fetches the teams the authenticated Heroku
// user belongs to, and always appends a synthetic "Personal account"
// entry. Heroku Teams are an opt-in paid construct, so a solo account has
// no team to discover; the personal entry lets the picker offer personal
// mode (app owner + collaborators) instead of dead-ending at a free-text
// slug the user cannot fill. Teams are fetched from baseURL ("" for the
// Heroku SaaS Platform API).
func ListHerokuOrganizations(ctx context.Context, httpClient *http.Client, baseURL string) ([]Organization, error) {
	if baseURL == "" {
		baseURL = herokuDefaultBaseURL
	}

	endpoint, err := url.JoinPath(baseURL, herokuTeamsSegment)
	if err != nil {
		return nil, fmt.Errorf("cannot build heroku organizations URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create heroku organizations request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch heroku organizations: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch heroku organizations: unexpected status %d", resp.StatusCode)
	}

	var teams []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&teams); err != nil {
		return nil, fmt.Errorf("cannot decode heroku organizations response: %w", err)
	}

	result := make([]Organization, 0, len(teams)+1)
	for _, t := range teams {
		displayName := t.Name
		if displayName == "" {
			displayName = t.ID
		}

		result = append(result, Organization{Slug: t.ID, DisplayName: displayName})
	}

	result = append(result, Organization{
		Slug:        herokuPersonalAccountSlug,
		DisplayName: herokuPersonalAccountDisplayName,
	})

	return result, nil
}

func herokuSource() Factory {
	return provider.Over(func(
		ctx context.Context,
		credential connector.HTTPCredential,
		opened *provider.Handle,
		logger *log.Logger,
	) (Driver, error) {
		driver, err := herokuSourceDriver(ctx, credential.Client, opened.Connector, logger, opened.Endpoints)
		if err != nil {
			return nil, err
		}

		return capable(
			driver,
			herokuSourceNameResolver(ctx, credential.Client, opened.Connector, logger, opened.Endpoints),
			organizationListerFunc(func(ctx context.Context) ([]Organization, error) {
				return ListHerokuOrganizations(ctx, credential.Client, organizationsBase(opened.Endpoints))
			}),
		), nil
	})
}

func herokuSourceDriver(
	_ context.Context,
	c *http.Client,
	conn *coredata.Connector,
	_ *log.Logger,
	ep provider.Endpoints,
) (Driver, error) {
	s, err := coredata.ConnectorSettings[coredata.HerokuConnectorSettings](conn)
	if err != nil {
		return nil, fmt.Errorf("cannot read heroku connector settings: %w", err)
	}

	// TeamID may be empty or the personal-account slug for a solo
	// Heroku account (no Team); the driver runs in personal mode
	// (app owner + collaborators) in that case.
	return NewHerokuDriver(c, s.TeamID, ep.APIBase), nil
}

func herokuSourceNameResolver(
	ctx context.Context,
	c *http.Client,
	conn *coredata.Connector,
	logger *log.Logger,
	ep provider.Endpoints,
) NameResolver {
	s, err := coredata.ConnectorSettings[coredata.HerokuConnectorSettings](conn)
	if err != nil {
		logger.ErrorCtx(ctx, "cannot read heroku connector settings", log.Error(err))
		return nil
	}

	return NewHerokuNameResolver(c, s.TeamID, ep.APIBase)
}
