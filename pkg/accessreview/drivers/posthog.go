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
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

type PostHogDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*PostHogDriver)(nil)

const (
	posthogMembersPath      = "/api/organizations/@current/members/"
	posthogOrganizationPath = "/api/organizations/@current/"
	posthogMembersPageSize  = 100

	// PostHog Cloud regional data hosts. OAuth connections carry no region
	// (empty baseURL): the region-agnostic oauth.posthog.com gateway used
	// for the OAuth handshake does NOT serve the data API, so the driver
	// discovers the region by probing these hosts with the connection's
	// token. API-key (us/eu) and self-hosted connections always carry an
	// explicit host instead.
	posthogUSBaseURL = "https://us.posthog.com"
	posthogEUBaseURL = "https://eu.posthog.com"

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

	host, err := resolvePostHogRegion(ctx, d.httpClient)
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
	nextURL, err := url.Parse(next)
	if err != nil {
		return "", fmt.Errorf("cannot parse posthog next page URL: %w", err)
	}

	base, err := url.Parse(d.baseURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse posthog base URL: %w", err)
	}

	if nextURL.IsAbs() {
		// Pin pagination to the resolved data host. The connection's bearer
		// token is attached to every request, so an absolute `next` pointing
		// at a different host (a compromised or spoofed API response) would
		// forward the token off-host. Refuse cross-host pagination; the
		// error is static so it never echoes an attacker-controlled host.
		if !strings.EqualFold(nextURL.Host, base.Host) {
			return "", fmt.Errorf("cannot follow posthog next page URL: cross-host pagination is not allowed")
		}

		return nextURL.String(), nil
	}

	endpoint, err := url.JoinPath(d.baseURL, posthogMembersPath)
	if err != nil {
		return "", fmt.Errorf("cannot build posthog members base URL: %w", err)
	}

	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("cannot parse posthog members base URL: %w", err)
	}

	return baseURL.ResolveReference(nextURL).String(), nil
}

// resolvePostHogRegion probes the PostHog Cloud region hosts with the given
// token-bearing client and returns the first that answers 2xx on the @current
// organization endpoint. OAuth connections authenticate via the region-agnostic
// oauth.posthog.com gateway, which does not serve /api, so the actual data
// region (us/eu) must be discovered against the regional hosts directly.
func resolvePostHogRegion(ctx context.Context, client *http.Client) (string, error) {
	for _, host := range []string{posthogUSBaseURL, posthogEUBaseURL} {
		endpoint, err := url.JoinPath(host, posthogOrganizationPath)
		if err != nil {
			continue
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			continue
		}

		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			// Surface a cancelled/expired context as the real cause rather
			// than masking it behind "no region accepted the connection".
			if ctx.Err() != nil {
				return "", fmt.Errorf("cannot resolve posthog region: %w", ctx.Err())
			}

			continue
		}

		status := resp.StatusCode
		_ = resp.Body.Close()

		if status >= http.StatusOK && status < http.StatusMultipleChoices {
			return host, nil
		}
	}

	return "", fmt.Errorf("cannot resolve posthog region: no region accepted the connection")
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
		Role:        posthogRole(member.Level, member.User.RoleAtOrganization),
		IsAdmin:     posthogIsAdmin(member.Level),
		ExternalID:  member.User.UUID,
		MFAStatus:   posthogMFAStatus(member.Is2FAEnabled),
		AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
		AccountType: coredata.AccessEntryAccountTypeUser,
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

func posthogRole(level int, fallback string) string {
	switch {
	case level >= posthogMembershipLevelOwner:
		return "Owner"
	case level >= posthogMembershipLevelAdmin:
		return "Admin"
	case level >= posthogMembershipLevelMember:
		return "Member"
	case fallback != "":
		return fallback
	default:
		return "Member"
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
		host, err := resolvePostHogRegion(ctx, r.httpClient)
		if err != nil {
			// Terminal: cannot determine the region (e.g. revoked token).
			// Keep the generic source name rather than making the
			// source-name worker retry forever.
			return "", nil
		}

		baseURL = host
	}

	endpoint, err := url.JoinPath(baseURL, posthogOrganizationPath)
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
