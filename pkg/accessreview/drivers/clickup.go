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
	"strconv"
	"time"
)

// ClickUpDriver fetches workspace ("team") members from the ClickUp
// REST API using a pre-authenticated HTTP client (Bearer token). The
// team endpoint returns the full member list inline in a single
// response — no pagination is performed.
//
// ClickUp does not issue refresh tokens; the existing RefreshableClient
// falls back to a non-refreshing client when RefreshToken == "" and the
// access source resolver re-prompts for re-authorization on 401.
type ClickUpDriver struct {
	httpClient *http.Client
	teamID     string
	baseURL    string
}

var _ Driver = (*ClickUpDriver)(nil)

const clickupTeamSegment = "team"

// clickupDefaultBaseURL is the ClickUp API root. It backs only the exported
// ListClickUpOrganizations, and only when its caller resolves no APIBase for
// the provider (unregistered, or registered without one). Every other path
// goes through the injected baseURL instead.
const clickupDefaultBaseURL = "https://api.clickup.com/api/v2"

// NewClickUpDriver builds a driver against baseURL, the versioned ClickUp
// API origin (e.g. https://api.clickup.com/api/v2).
func NewClickUpDriver(httpClient *http.Client, teamID, baseURL string) *ClickUpDriver {
	return &ClickUpDriver{
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

type clickupMember struct {
	User struct {
		ID         json.Number `json:"id"`
		Email      string      `json:"email"`
		Username   string      `json:"username"`
		Role       int         `json:"role"`
		LastActive string      `json:"last_active"`
	} `json:"user"`
	InvitePending *bool `json:"invite_pending"`
}

type clickupTeamResponse struct {
	Team struct {
		Members []clickupMember `json:"members"`
	} `json:"team"`
}

func (d *ClickUpDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	endpoint, err := url.JoinPath(d.baseURL, clickupTeamSegment, url.PathEscape(d.teamID))
	if err != nil {
		return nil, fmt.Errorf("cannot build clickup team URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create clickup team request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute clickup team request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch clickup team: unexpected status %d", httpResp.StatusCode)
	}

	var resp clickupTeamResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode clickup team response: %w", err)
	}

	records := make([]AccountRecord, 0, len(resp.Team.Members))
	for _, m := range resp.Team.Members {
		roles := clickupRoles(m.User.Role)
		isAdmin := m.User.Role == 1 || m.User.Role == 2

		record := AccountRecord{
			Email:       m.User.Email,
			FullName:    m.User.Username,
			Roles:       roles,
			IsAdmin:     new(isAdmin),
			ExternalID:  m.User.ID.String(),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
		}

		if m.InvitePending != nil {
			active := !*m.InvitePending
			record.Active = &active
		}

		if m.User.LastActive != "" {
			// ClickUp emits last_active as a Unix-millis string; fall
			// back to RFC3339 if a future API change switches format.
			if t, err := parseClickUpTime(m.User.LastActive); err == nil {
				record.LastLogin = &t
			}
		}

		records = append(records, record)
	}

	return records, nil
}

// clickupRoles maps ClickUp numeric role codes to human-readable
// labels. Source: https://clickup.com/api (Team Members endpoint).
func clickupRoles(role int) []string {
	switch role {
	case 1:
		return []string{"owner"}
	case 2:
		return []string{"admin"}
	case 3:
		return []string{"member"}
	case 4:
		return []string{"guest"}
	default:
		return []string{}
	}
}

// parseClickUpTime accepts both ClickUp's Unix-millis-as-string format
// and RFC3339 timestamps so the driver remains forward-compatible.
func parseClickUpTime(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}

	// strconv.ParseInt rejects trailing non-digit garbage that fmt.Sscanf
	// would silently truncate (e.g. "123abc" → 123).
	ms, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse clickup time %q: %w", raw, err)
	}

	return time.UnixMilli(ms).UTC(), nil
}

// clickupNameResolver resolves the ClickUp team name.
type clickupNameResolver struct {
	httpClient *http.Client
	teamID     string
	baseURL    string
}

// NewClickUpNameResolver resolves the team name against baseURL, the
// versioned ClickUp API origin (e.g. https://api.clickup.com/api/v2).
func NewClickUpNameResolver(httpClient *http.Client, teamID, baseURL string) NameResolver {
	return &clickupNameResolver{httpClient: httpClient, teamID: teamID, baseURL: baseURL}
}

func (r *clickupNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.teamID == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath(r.baseURL, clickupTeamSegment, url.PathEscape(r.teamID))
	if err != nil {
		return "", fmt.Errorf("cannot build clickup team URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create clickup team request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute clickup team request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nameStatusError("clickup team", httpResp.StatusCode)
	}

	var resp struct {
		Team struct {
			Name string `json:"name"`
		} `json:"team"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode clickup team response: %w", err)
	}

	return resp.Team.Name, nil
}

// ListClickUpOrganizations fetches the ClickUp teams (workspaces) the
// authenticated user belongs to, from baseURL ("" for the ClickUp SaaS API).
func ListClickUpOrganizations(ctx context.Context, httpClient *http.Client, baseURL string) ([]Organization, error) {
	if baseURL == "" {
		baseURL = clickupDefaultBaseURL
	}

	endpoint, err := url.JoinPath(baseURL, clickupTeamSegment)
	if err != nil {
		return nil, fmt.Errorf("cannot build clickup organizations URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create clickup organizations request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch clickup organizations: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch clickup organizations: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Teams []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"teams"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("cannot decode clickup organizations response: %w", err)
	}

	result := make([]Organization, len(body.Teams))
	for i, t := range body.Teams {
		displayName := t.Name
		if displayName == "" {
			displayName = t.ID
		}

		result[i] = Organization{Slug: t.ID, DisplayName: displayName}
	}

	return result, nil
}

func clickupSource() Factory {
	return provider.Over(func(
		ctx context.Context,
		credential connector.HTTPCredential,
		opened *provider.Handle,
		logger *log.Logger,
	) (Driver, error) {
		driver, err := clickupSourceDriver(ctx, credential.Client, opened.Connector, logger, opened.Endpoints)
		if err != nil {
			return nil, err
		}

		return capable(
			driver,
			clickupSourceNameResolver(ctx, credential.Client, opened.Connector, logger, opened.Endpoints),
			organizationListerFunc(func(ctx context.Context) ([]Organization, error) {
				return ListClickUpOrganizations(ctx, credential.Client, organizationsBase(opened.Endpoints))
			}),
		), nil
	})
}

func clickupSourceDriver(
	_ context.Context,
	c *http.Client,
	conn *coredata.Connector,
	_ *log.Logger,
	ep provider.Endpoints,
) (Driver, error) {
	s, err := coredata.ConnectorSettings[coredata.ClickUpConnectorSettings](conn)
	if err != nil {
		return nil, fmt.Errorf("cannot read clickup connector settings: %w", err)
	}

	if s.TeamID == "" {
		return nil, fmt.Errorf("cannot create clickup driver: team_id is required")
	}

	return NewClickUpDriver(c, s.TeamID, ep.APIBase), nil
}

func clickupSourceNameResolver(
	ctx context.Context,
	c *http.Client,
	conn *coredata.Connector,
	logger *log.Logger,
	ep provider.Endpoints,
) NameResolver {
	s, err := coredata.ConnectorSettings[coredata.ClickUpConnectorSettings](conn)
	if err != nil {
		logger.ErrorCtx(ctx, "cannot read clickup connector settings", log.Error(err))
		return nil
	}

	return NewClickUpNameResolver(c, s.TeamID, ep.APIBase)
}
