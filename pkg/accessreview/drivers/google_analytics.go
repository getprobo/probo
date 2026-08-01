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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	googleAnalyticsAPIHost  = "analyticsadmin.googleapis.com"
	googleAnalyticsPageSize = 200
	// googleAnalyticsAdminRole is the only GA4 predefined role that grants
	// administrative access; viewer/analyst/editor do not, and no-cost-data /
	// no-revenue-data are data restrictions rather than access levels.
	googleAnalyticsAdminRole  = "predefinedRoles/admin"
	googleAnalyticsRolePrefix = "predefinedRoles/"
)

// GoogleAnalyticsDriver lists the users who have access to a single GA4 account
// and every property beneath it, using the Analytics Admin API v1alpha (the
// only version that exposes accessBindings). Access is granted at two levels —
// account and property — so a user's effective roles are the union of their
// account-level binding and each of their property-level bindings, deduplicated
// by email.
type GoogleAnalyticsDriver struct {
	httpClient *http.Client
	accountID  string
}

var _ Driver = (*GoogleAnalyticsDriver)(nil)

type googleAnalyticsAccessBinding struct {
	// User is the email address the binding grants roles to.
	User  string   `json:"user"`
	Roles []string `json:"roles"`
}

type googleAnalyticsBindingsResponse struct {
	AccessBindings []googleAnalyticsAccessBinding `json:"accessBindings"`
	NextPageToken  string                         `json:"nextPageToken"`
}

type googleAnalyticsProperty struct {
	// Name is the resource name, e.g. "properties/67890".
	Name string `json:"name"`
}

type googleAnalyticsPropertiesResponse struct {
	Properties    []googleAnalyticsProperty `json:"properties"`
	NextPageToken string                    `json:"nextPageToken"`
}

// googleAnalyticsMember accumulates a user's roles and admin flag across their
// account-level and property-level bindings.
type googleAnalyticsMember struct {
	roles   map[string]struct{}
	isAdmin bool
}

func NewGoogleAnalyticsDriver(httpClient *http.Client, accountID string) *GoogleAnalyticsDriver {
	return &GoogleAnalyticsDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		accountID: accountID,
	}
}

func (d *GoogleAnalyticsDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	members := make(map[string]*googleAnalyticsMember)

	// Account-level bindings.
	if err := d.collectBindings(ctx, members, "v1alpha", "accounts", url.PathEscape(d.accountID), "accessBindings"); err != nil {
		return nil, fmt.Errorf("cannot list google analytics access bindings for account %q: %w", d.accountID, err)
	}

	// Property-level bindings, one loop per property beneath the account.
	propertyIDs, err := d.listProperties(ctx)
	if err != nil {
		return nil, err
	}

	for _, propertyID := range propertyIDs {
		err := d.collectBindings(ctx, members, "v1alpha", "properties", url.PathEscape(propertyID), "accessBindings")
		if err == nil {
			continue
		}

		// A property the token cannot read, or one deleted between the list
		// and the read, must not discard the bindings already collected: an
		// account whose properties are 49/50 readable is still worth
		// reviewing. Anything else invalidates the whole fetch.
		if e, ok := errors.AsType[*googleAnalyticsStatusError](err); ok &&
			(e.status == http.StatusForbidden || e.status == http.StatusNotFound) {
			continue
		}

		return nil, fmt.Errorf("cannot list google analytics access bindings for property %q: %w", propertyID, err)
	}

	return googleAnalyticsRecords(members), nil
}

// collectBindings paginates the accessBindings collection under the given
// resource path and folds each binding into members.
func (d *GoogleAnalyticsDriver) collectBindings(ctx context.Context, members map[string]*googleAnalyticsMember, segments ...string) error {
	pageToken := ""

	for range maxPaginationPages {
		endpoint, err := googleAnalyticsURL(pageToken, nil, segments...)
		if err != nil {
			return err
		}

		var resp googleAnalyticsBindingsResponse
		if err := d.getJSON(ctx, endpoint, &resp); err != nil {
			return err
		}

		for _, b := range resp.AccessBindings {
			addGoogleAnalyticsBinding(members, b.User, b.Roles)
		}

		if resp.NextPageToken == "" {
			return nil
		}

		pageToken = resp.NextPageToken
	}

	return fmt.Errorf("cannot list all google analytics access bindings: %w", ErrPaginationLimitReached)
}

// listProperties returns the numeric IDs of every property under the account,
// including subproperties and roll-up properties. The ancestor filter walks the
// whole account hierarchy (parent: would return only properties whose direct
// parent is the account, silently dropping subproperties parented to another
// property, and with them any subproperty-only members).
func (d *GoogleAnalyticsDriver) listProperties(ctx context.Context) ([]string, error) {
	var propertyIDs []string

	pageToken := ""
	filter := url.Values{"filter": {"ancestor:accounts/" + d.accountID}}

	for range maxPaginationPages {
		endpoint, err := googleAnalyticsURL(pageToken, filter, "v1alpha", "properties")
		if err != nil {
			return nil, err
		}

		var resp googleAnalyticsPropertiesResponse
		if err := d.getJSON(ctx, endpoint, &resp); err != nil {
			return nil, err
		}

		for _, p := range resp.Properties {
			if id := strings.TrimPrefix(p.Name, "properties/"); id != "" {
				propertyIDs = append(propertyIDs, id)
			}
		}

		if resp.NextPageToken == "" {
			return propertyIDs, nil
		}

		pageToken = resp.NextPageToken
	}

	return nil, fmt.Errorf("cannot list all google analytics properties: %w", ErrPaginationLimitReached)
}

// googleAnalyticsStatusError carries the HTTP status of a failed Admin API call
// so callers can tell a per-resource permission problem apart from a failure
// that invalidates the whole fetch.
type googleAnalyticsStatusError struct {
	status int
}

func (e *googleAnalyticsStatusError) Error() string {
	return fmt.Sprintf("unexpected status %d", e.status)
}

func (d *GoogleAnalyticsDriver) getJSON(ctx context.Context, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("cannot create google analytics request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot execute google analytics request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return &googleAnalyticsStatusError{status: httpResp.StatusCode}
	}

	if err := json.NewDecoder(httpResp.Body).Decode(out); err != nil {
		return fmt.Errorf("cannot decode google analytics response: %w", err)
	}

	return nil
}

// googleAnalyticsURL builds a v1alpha Admin API URL from path segments, adding
// the shared pageSize, an optional page token, and any extra query values.
// Keys present in extra replace the default rather than adding to it.
func googleAnalyticsURL(pageToken string, extra url.Values, segments ...string) (string, error) {
	joined, err := url.JoinPath("https://"+googleAnalyticsAPIHost, segments...)
	if err != nil {
		return "", fmt.Errorf("cannot build google analytics URL: %w", err)
	}

	parsed, err := url.Parse(joined)
	if err != nil {
		return "", fmt.Errorf("cannot parse google analytics URL: %w", err)
	}

	q := parsed.Query()
	q.Set("pageSize", strconv.Itoa(googleAnalyticsPageSize))

	for k, vs := range extra {
		q.Del(k)

		for _, v := range vs {
			q.Add(k, v)
		}
	}

	if pageToken != "" {
		q.Set("pageToken", pageToken)
	}

	parsed.RawQuery = q.Encode()

	return parsed.String(), nil
}

// GoogleAnalyticsAccountBindingsProbeURL builds a single-item account-level
// accessBindings request for accountID. The connection probe uses it so the
// check exercises the permission the driver actually needs — Administrator on
// the account, granted through analytics.manage.users.readonly — instead of the
// accounts list, which any analytics.readonly grant can call.
func GoogleAnalyticsAccountBindingsProbeURL(accountID string) (string, error) {
	return googleAnalyticsURL(
		"",
		url.Values{"pageSize": {"1"}},
		"v1alpha", "accounts", url.PathEscape(accountID), "accessBindings",
	)
}

// addGoogleAnalyticsBinding folds one access binding into the per-email member
// map, deduplicating roles and setting the admin flag when the admin role is
// present.
func addGoogleAnalyticsBinding(members map[string]*googleAnalyticsMember, user string, roles []string) {
	email := strings.ToLower(strings.TrimSpace(user))
	if email == "" {
		return
	}

	member, ok := members[email]
	if !ok {
		member = &googleAnalyticsMember{roles: make(map[string]struct{})}
		members[email] = member
	}

	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}

		member.roles[role] = struct{}{}

		if role == googleAnalyticsAdminRole {
			member.isAdmin = true
		}
	}
}

// googleAnalyticsRecords turns the merged member map into a deterministically
// ordered slice of AccountRecords. GA4 access bindings identify a user only by
// email — there is no stable per-user ID and no display name exposed — so the
// email is used as both ExternalID and FullName. Active is left nil: bindings
// carry no account-status signal.
func googleAnalyticsRecords(members map[string]*googleAnalyticsMember) []AccountRecord {
	emails := make([]string, 0, len(members))
	for email := range members {
		emails = append(emails, email)
	}

	sort.Strings(emails)

	records := make([]AccountRecord, 0, len(members))

	for _, email := range emails {
		member := members[email]

		roles := make([]string, 0, len(member.roles))
		for role := range member.roles {
			roles = append(roles, strings.TrimPrefix(role, googleAnalyticsRolePrefix))
		}

		sort.Strings(roles)

		records = append(records, AccountRecord{
			Email:       email,
			FullName:    email,
			Roles:       roles,
			IsAdmin:     member.isAdmin,
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  email,
		})
	}

	return records
}

// googleAnalyticsNameResolver resolves a GA4 account's display name.
type googleAnalyticsNameResolver struct {
	httpClient *http.Client
	accountID  string
}

func NewGoogleAnalyticsNameResolver(httpClient *http.Client, accountID string) NameResolver {
	return &googleAnalyticsNameResolver{httpClient: httpClient, accountID: accountID}
}

func (r *googleAnalyticsNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.accountID == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath("https://"+googleAnalyticsAPIHost, "v1alpha", "accounts", url.PathEscape(r.accountID))
	if err != nil {
		return "", fmt.Errorf("cannot build google analytics account URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create google analytics account request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute google analytics account request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	// A non-2xx (revoked token, renamed/deleted account) is terminal: keep the
	// generic source name rather than retry forever.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		DisplayName string `json:"displayName"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode google analytics account response: %w", err)
	}

	return resp.DisplayName, nil
}

// ListGoogleAnalyticsOrganizations fetches the GA4 accounts the authenticated
// Google user can access, surfacing each account's numeric ID as the picker
// slug. Listing accounts requires the analytics.readonly scope.
func ListGoogleAnalyticsOrganizations(ctx context.Context, httpClient *http.Client) ([]Organization, error) {
	var orgs []Organization

	pageToken := ""

	for range maxPaginationPages {
		endpoint, err := googleAnalyticsURL(pageToken, nil, "v1alpha", "accounts")
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot create google analytics accounts request: %w", err)
		}

		req.Header.Set("Accept", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("cannot fetch google analytics accounts: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()

			return nil, fmt.Errorf("cannot fetch google analytics accounts: unexpected status %d", resp.StatusCode)
		}

		var out struct {
			Accounts []struct {
				Name        string `json:"name"`
				DisplayName string `json:"displayName"`
			} `json:"accounts"`
			NextPageToken string `json:"nextPageToken"`
		}

		decodeErr := json.NewDecoder(resp.Body).Decode(&out)
		_ = resp.Body.Close()

		if decodeErr != nil {
			return nil, fmt.Errorf("cannot decode google analytics accounts response: %w", decodeErr)
		}

		for _, a := range out.Accounts {
			id := strings.TrimPrefix(a.Name, "accounts/")
			if id == "" {
				continue
			}

			displayName := a.DisplayName
			if displayName == "" {
				displayName = id
			}

			orgs = append(orgs, Organization{Slug: id, DisplayName: displayName})
		}

		if out.NextPageToken == "" {
			return orgs, nil
		}

		pageToken = out.NextPageToken
	}

	return nil, fmt.Errorf("cannot list all google analytics accounts: %w", ErrPaginationLimitReached)
}
