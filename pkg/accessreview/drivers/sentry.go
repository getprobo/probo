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
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/rfc5988"
)

// errSentryOrgNotAccessible signals a 404 scoped under an organization
// slug; Sentry uses 404 (not 403) so this also covers revoked memberships.
var errSentryOrgNotAccessible = errors.New("sentry organization is not accessible by this connector's token")

// Sentry path elements joined onto the driver's base URL. Both carry a
// trailing slash because Sentry's API only routes slashed paths and answers
// 404 (without redirecting) otherwise; url.JoinPath keeps the slash only
// when the LAST element carries it, so every join must end on one of these.
const (
	sentryOrganizationsPath = "organizations/"
	sentryMembersPath       = "members/"
)

// sentryDefaultBaseURL is the Sentry SaaS API root. It backs only the
// exported ListSentryOrganizations, and only when its caller resolves no
// APIBase for the provider (unregistered, or registered without one).
// Every driver path goes through the injected baseURL instead.
const sentryDefaultBaseURL = "https://sentry.io/api/0"

// SentryDriver fetches organization members from Sentry via Bearer
// token-authenticated REST API requests.
type SentryDriver struct {
	httpClient *http.Client
	orgSlug    string
	baseURL    string
}

var _ Driver = (*SentryDriver)(nil)

type sentryMember struct {
	ID          string          `json:"id"`
	Email       string          `json:"email"`
	Name        string          `json:"name"`
	Pending     bool            `json:"pending"`
	OrgRole     string          `json:"orgRole"`
	DateCreated string          `json:"dateCreated"`
	Flags       map[string]bool `json:"flags"`
	User        *sentryUser     `json:"user"`
}

type sentryUser struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	IsActive        bool   `json:"isActive"`
	Has2FA          bool   `json:"has2fa"`
	LastLogin       string `json:"lastLogin"`
	HasPasswordAuth bool   `json:"hasPasswordAuth"`
}

// NewSentryDriver builds a driver against baseURL, the versioned Sentry API
// origin (e.g. https://sentry.io/api/0).
func NewSentryDriver(httpClient *http.Client, orgSlug, baseURL string) *SentryDriver {
	return &SentryDriver{
		httpClient: httpClient,
		orgSlug:    orgSlug,
		baseURL:    baseURL,
	}
}

func (d *SentryDriver) resolveOrgSlug(ctx context.Context) (string, error) {
	orgs, err := listSentryOrganizations(ctx, d.httpClient, d.baseURL)
	if err != nil {
		return "", fmt.Errorf("cannot resolve sentry organization slug: %w", err)
	}

	if len(orgs) == 0 {
		return "", fmt.Errorf("no sentry organizations found for this token")
	}

	return orgs[0].Slug, nil
}

func (d *SentryDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	orgSlug := d.orgSlug
	if orgSlug == "" {
		slug, err := d.resolveOrgSlug(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve sentry organization slug: %w", err)
		}

		orgSlug = slug
	}

	var records []AccountRecord

	nextURL, err := url.JoinPath(d.baseURL, sentryOrganizationsPath, url.PathEscape(orgSlug), sentryMembersPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build sentry members URL: %w", err)
	}

	for range maxPaginationPages {
		members, linkHeader, err := d.queryMembers(ctx, nextURL)
		if err != nil {
			if errors.Is(err, errSentryOrgNotAccessible) {
				return nil, fmt.Errorf("sentry organization %q is not accessible; reconnect the connector with the correct organization: %w", orgSlug, err)
			}

			return nil, err
		}

		for _, m := range members {
			fullName := m.Name
			if fullName == "" && m.User != nil {
				fullName = m.User.Name
			}

			active := !m.Pending
			if m.User != nil {
				active = active && m.User.IsActive
			}

			role := strings.TrimSpace(m.OrgRole)
			isAdmin := role == "admin" || role == "owner"

			mfaStatus := coredata.MFAStatusUnknown

			if m.User != nil {
				if m.User.Has2FA {
					mfaStatus = coredata.MFAStatusEnabled
				} else {
					mfaStatus = coredata.MFAStatusDisabled
				}
			}

			authMethod := sentryAuthMethod(m.Flags, m.User)

			roles := []string{}
			if role != "" {
				roles = []string{role}
			}

			record := AccountRecord{
				Email:       m.Email,
				FullName:    fullName,
				Roles:       roles,
				Active:      new(active),
				IsAdmin:     new(isAdmin),
				ExternalID:  m.ID,
				MFAStatus:   mfaStatus,
				AuthMethod:  authMethod,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
			}

			if m.User != nil && m.User.LastLogin != "" {
				if t, err := time.Parse(time.RFC3339, m.User.LastLogin); err == nil {
					record.LastLogin = &t
				}
			}

			if m.DateCreated != "" {
				if t, err := time.Parse(time.RFC3339, m.DateCreated); err == nil {
					record.CreatedAt = &t
				}
			}

			if record.Email != "" {
				records = append(records, record)
			}
		}

		rawNext := sentryNextLink(linkHeader)
		if rawNext == "" {
			return records, nil
		}

		nextURL, err = sameHostNextPageURL("sentry", d.baseURL, rawNext)
		if err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("cannot list all sentry accounts: %w", ErrPaginationLimitReached)
}

func (d *SentryDriver) queryMembers(ctx context.Context, url string) ([]sentryMember, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("cannot create sentry members request: %w", err)
	}

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("cannot execute sentry members request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode == http.StatusNotFound {
		return nil, "", errSentryOrgNotAccessible
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("cannot fetch sentry members: unexpected status %d", httpResp.StatusCode)
	}

	var members []sentryMember
	if err := json.NewDecoder(httpResp.Body).Decode(&members); err != nil {
		return nil, "", fmt.Errorf("cannot decode sentry members response: %w", err)
	}

	return members, httpResp.Header.Get("Link"), nil
}

// sentryNextLink extracts the next page URL from a Sentry Link header.
// It returns the URL for the entry with rel="next" and results="true", or
// an empty string if no such entry exists.
func sentryNextLink(header string) string {
	for _, link := range rfc5988.Parse(header) {
		if link.Params["rel"] == "next" && link.Params["results"] == "true" {
			return link.URL
		}
	}

	return ""
}

func sentryAuthMethod(flags map[string]bool, user *sentryUser) coredata.AccessReviewEntryAuthMethod {
	if flags["sso:linked"] {
		return coredata.AccessReviewEntryAuthMethodSSO
	}

	if user != nil && user.HasPasswordAuth {
		return coredata.AccessReviewEntryAuthMethodPassword
	}

	return coredata.AccessReviewEntryAuthMethodUnknown
}

// sentryNameResolver resolves the Sentry organization name.
type sentryNameResolver struct {
	httpClient *http.Client
	orgSlug    string
	baseURL    string
}

// NewSentryNameResolver resolves the org name against baseURL, the
// versioned Sentry API origin (e.g. https://sentry.io/api/0).
func NewSentryNameResolver(httpClient *http.Client, orgSlug, baseURL string) NameResolver {
	return &sentryNameResolver{httpClient: httpClient, orgSlug: orgSlug, baseURL: baseURL}
}

func (r *sentryNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.orgSlug == "" {
		return "", nil
	}

	// Trailing slash required; see the sentryOrganizationsPath const.
	endpoint, err := url.JoinPath(r.baseURL, sentryOrganizationsPath, url.PathEscape(r.orgSlug)+"/")
	if err != nil {
		return "", fmt.Errorf("cannot build sentry organization URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create sentry organization request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute sentry organization request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	// 404 means the stored slug is no longer visible to this token.
	// Treat as terminal so the worker stops looping; other non-2xx
	// stay retryable for token refresh / transient outages.
	if httpResp.StatusCode == http.StatusNotFound {
		return "", nil
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nameStatusError("sentry organization", httpResp.StatusCode)
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode sentry organization response: %w", err)
	}

	return resp.Name, nil
}

// ListSentryOrganizations fetches the organizations the authenticated
// Sentry user belongs to, from baseURL ("" for Sentry SaaS).
func ListSentryOrganizations(ctx context.Context, httpClient *http.Client, baseURL string) ([]Organization, error) {
	if baseURL == "" {
		baseURL = sentryDefaultBaseURL
	}

	return listSentryOrganizations(ctx, httpClient, baseURL)
}

func listSentryOrganizations(ctx context.Context, httpClient *http.Client, baseURL string) ([]Organization, error) {
	endpoint, err := url.JoinPath(baseURL, sentryOrganizationsPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build sentry organizations URL: %w", err)
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("cannot parse sentry organizations URL: %w", err)
	}

	q := parsed.Query()
	q.Set("member", "true")
	parsed.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create sentry organizations request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch sentry organizations: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch sentry organizations: unexpected status %d", resp.StatusCode)
	}

	var orgs []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return nil, fmt.Errorf("cannot decode sentry organizations response: %w", err)
	}

	result := make([]Organization, len(orgs))
	for i, org := range orgs {
		displayName := org.Name
		if displayName == "" {
			displayName = org.Slug
		}

		result[i] = Organization{Slug: org.Slug, DisplayName: displayName}
	}

	return result, nil
}
