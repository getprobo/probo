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
	"strings"
	"time"
)

type PostHogDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*PostHogDriver)(nil)

const (
	posthogMembersPath     = "/api/organizations/@current/members/"
	posthogMembersPageSize = 100

	posthogMembershipLevelMember = 1
	posthogMembershipLevelAdmin  = 8
	posthogMembershipLevelOwner  = 15
)

type (
	posthogMembersResponse struct {
		Next    string          `json:"next"`
		Results []posthogMember `json:"results"`
	}

	posthogMember struct {
		ID           string            `json:"id"`
		User         posthogMemberUser `json:"user"`
		Level        int               `json:"level"`
		Is2FAEnabled *bool             `json:"is_2fa_enabled"`
		JoinedAt     string            `json:"joined_at"`
		LastLogin    string            `json:"last_login"`
	}

	posthogMemberUser struct {
		UUID               string `json:"uuid"`
		FirstName          string `json:"first_name"`
		LastName           string `json:"last_name"`
		Email              string `json:"email"`
		RoleAtOrganization string `json:"role_at_organization"`
	}
)

// NewPostHogDriver builds a driver against baseURL (e.g. https://us.posthog.com
// or a self-hosted instance URL). An empty baseURL marks a cloud OAuth
// connection whose region is discovered lazily on first use (see resolveBaseURL).
func NewPostHogDriver(httpClient *http.Client, baseURL string) *PostHogDriver {
	return &PostHogDriver{httpClient: httpClient, baseURL: baseURL}
}

// resolveBaseURL ensures the driver has a concrete data host. Explicit hosts
// (API-key region / self-hosted) are used as-is; an empty baseURL (cloud
// OAuth) is resolved by probing the PostHog Cloud regions with the
// connection's token, since the oauth.posthog.com gateway does not serve /api.
// The result is cached on the driver for subsequent pages.
func (d *PostHogDriver) resolveBaseURL(ctx context.Context) error {
	if d.baseURL != "" {
		return nil
	}

	host, err := provider.ResolvePostHogRegion(ctx, d.httpClient)
	if err != nil {
		return err
	}

	d.baseURL = host

	return nil
}

func (d *PostHogDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	if err := d.resolveBaseURL(ctx); err != nil {
		return nil, err
	}

	nextURL, err := d.membersURL()
	if err != nil {
		return nil, err
	}

	var records []AccountRecord

	for range maxPaginationPages {
		resp, err := d.fetchMembers(ctx, nextURL)
		if err != nil {
			return nil, err
		}

		for _, member := range resp.Results {
			record := posthogAccountRecord(member)
			if record.Email == "" {
				continue
			}

			records = append(records, record)
		}

		if resp.Next == "" {
			return records, nil
		}

		nextURL, err = d.resolveNextURL(resp.Next)
		if err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("cannot list all posthog accounts: %w", ErrPaginationLimitReached)
}

func (d *PostHogDriver) membersURL() (string, error) {
	endpoint, err := url.JoinPath(d.baseURL, posthogMembersPath)
	if err != nil {
		return "", fmt.Errorf("cannot build posthog members URL: %w", err)
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("cannot parse posthog members URL: %w", err)
	}

	q := u.Query()
	q.Set("limit", strconv.Itoa(posthogMembersPageSize))
	q.Set("order", "-joined_at")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (d *PostHogDriver) resolveNextURL(next string) (string, error) {
	// Resolve against the members endpoint so a relative `next` keeps its
	// path context, and let sameHostNextPageURL pin the result to the data
	// host. Both share one host, so checking against the members endpoint is
	// equivalent to checking against d.baseURL.
	endpoint, err := url.JoinPath(d.baseURL, posthogMembersPath)
	if err != nil {
		return "", fmt.Errorf("cannot build posthog members base URL: %w", err)
	}

	return sameHostNextPageURL("posthog", endpoint, next)
}

// PostHogRegionBaseURL maps a PostHog Cloud region ("US"/"EU",
// case-insensitive) to its data-API host. It is the single source of truth for
// the regional hosts, shared with the connector-settings resolver so the two
// never drift. Self-hosted instances use a full instance URL instead.
func PostHogRegionBaseURL(region string) (string, bool) {
	hosts := provider.PostHogRegionBaseURLs()

	switch strings.ToLower(region) {
	case "us":
		return hosts[0], true
	case "eu":
		return hosts[1], true
	default:
		return "", false
	}
}

func (d *PostHogDriver) fetchMembers(
	ctx context.Context,
	nextURL string,
) (*posthogMembersResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, nextURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create posthog members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute posthog members request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("cannot fetch posthog members: unexpected status %d", httpResp.StatusCode)
	}

	var resp posthogMembersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode posthog members response: %w", err)
	}

	return &resp, nil
}

func posthogAccountRecord(member posthogMember) AccountRecord {
	record := AccountRecord{
		Email:       member.User.Email,
		FullName:    posthogFullName(member.User),
		Roles:       posthogRoles(member.Level, member.User.RoleAtOrganization),
		IsAdmin:     new(posthogIsAdmin(member.Level)),
		ExternalID:  member.User.UUID,
		MFAStatus:   posthogMFAStatus(member.Is2FAEnabled),
		AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
	}

	if record.ExternalID == "" {
		record.ExternalID = member.ID
	}

	if t, ok := parseRFC3339(member.JoinedAt); ok {
		record.CreatedAt = &t
	}

	if t, ok := parseRFC3339(member.LastLogin); ok {
		record.LastLogin = &t
	}

	return record
}

func posthogFullName(user posthogMemberUser) string {
	return strings.TrimSpace(strings.Join([]string{user.FirstName, user.LastName}, " "))
}

func posthogRoles(level int, fallback string) []string {
	switch {
	case level >= posthogMembershipLevelOwner:
		return []string{"Owner"}
	case level >= posthogMembershipLevelAdmin:
		return []string{"Admin"}
	case level >= posthogMembershipLevelMember:
		return []string{"Member"}
	case fallback != "":
		return []string{fallback}
	default:
		return []string{"Member"}
	}
}

func posthogIsAdmin(level int) bool {
	return level >= posthogMembershipLevelAdmin
}

func posthogMFAStatus(twoFAEnabled *bool) coredata.MFAStatus {
	if twoFAEnabled == nil {
		return coredata.MFAStatusUnknown
	}

	if *twoFAEnabled {
		return coredata.MFAStatusEnabled
	}

	return coredata.MFAStatusDisabled
}

// posthogNameResolver resolves the PostHog organization name from the
// current organization endpoint, which returns the org the connection
// belongs to.
type posthogNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

var _ NameResolver = (*posthogNameResolver)(nil)

// NewPostHogNameResolver resolves the org name against baseURL. An empty
// baseURL marks a cloud OAuth connection whose region is discovered lazily.
func NewPostHogNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &posthogNameResolver{httpClient: httpClient, baseURL: baseURL}
}

func (r *posthogNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	baseURL := r.baseURL
	if baseURL == "" {
		host, err := provider.ResolvePostHogRegion(ctx, r.httpClient)
		if err != nil {
			// Terminal: cannot determine the region (e.g. revoked token).
			// Keep the generic source name rather than making the
			// source-name worker retry forever.
			return "", nil
		}

		baseURL = host
	}

	endpoint, err := url.JoinPath(baseURL, provider.PostHogOrganizationPath)
	if err != nil {
		return "", fmt.Errorf("cannot build posthog organization URL: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("cannot create posthog organization request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute posthog organization request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	// Best-effort: a non-2xx (e.g. a revoked key) must not make the
	// source-name worker retry forever. Give up gracefully and keep the
	// generic source name; a dead key surfaces on the next ListAccounts.
	if httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		return "", nil
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode posthog organization response: %w", err)
	}

	return resp.Name, nil
}

func parseRFC3339(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}

	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, false
	}

	return t, true
}

func posthogSource() Factory {
	return provider.Over(func(
		ctx context.Context,
		credential connector.HTTPCredential,
		opened *provider.Handle,
		logger *log.Logger,
	) (Driver, error) {
		driver, err := posthogSourceDriver(ctx, credential.Client, opened.Connector, logger, opened.Endpoints)
		if err != nil {
			return nil, err
		}

		return capable(
			driver,
			posthogSourceNameResolver(ctx, credential.Client, opened.Connector, logger, opened.Endpoints),
			nil,
		), nil
	})
}

func posthogSourceDriver(
	_ context.Context,
	c *http.Client,
	conn *coredata.Connector,
	_ *log.Logger,
	_ provider.Endpoints,
) (Driver, error) {
	s, err := coredata.ConnectorSettings[coredata.PostHogConnectorSettings](conn)
	if err != nil {
		return nil, fmt.Errorf("cannot read posthog connector settings: %w", err)
	}

	// BaseURL is empty for cloud OAuth connections; the driver then
	// discovers the region (us/eu) lazily by probing, since the
	// oauth.posthog.com gateway does not serve the data API.
	return NewPostHogDriver(c, s.BaseURL), nil
}

func posthogSourceNameResolver(
	ctx context.Context,
	c *http.Client,
	conn *coredata.Connector,
	logger *log.Logger,
	_ provider.Endpoints,
) NameResolver {
	s, err := coredata.ConnectorSettings[coredata.PostHogConnectorSettings](conn)
	if err != nil {
		logger.ErrorCtx(ctx, "cannot read posthog connector settings", log.Error(err))
		return nil
	}

	return NewPostHogNameResolver(c, s.BaseURL)
}
