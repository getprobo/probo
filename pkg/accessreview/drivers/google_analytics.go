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

package drivers

import (
	"context"
	"encoding/json"
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
		return nil, err
	}

	// Property-level bindings, one loop per property beneath the account.
	propertyIDs, err := d.listProperties(ctx)
	if err != nil {
		return nil, err
	}

	for _, propertyID := range propertyIDs {
		if err := d.collectBindings(ctx, members, "v1alpha", "properties", url.PathEscape(propertyID), "accessBindings"); err != nil {
			return nil, err
		}
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
		return fmt.Errorf("cannot fetch google analytics resource: unexpected status %d", httpResp.StatusCode)
	}

	if err := json.NewDecoder(httpResp.Body).Decode(out); err != nil {
		return fmt.Errorf("cannot decode google analytics response: %w", err)
	}

	return nil
}

// googleAnalyticsURL builds a v1alpha Admin API URL from path segments, adding
// the shared pageSize, an optional page token, and any extra query values.
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
