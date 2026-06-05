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
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/rfc5988"
)

// OktaDriver lists the users of a single Okta org. The org is identified by
// its bare domain host (e.g. "acme.okta.com") supplied with the API token —
// Okta has no central API gateway, so every request targets that org's own
// host. The connection's transport attaches the `Authorization: SSWS <token>`
// header (see connector.APIKeyConnection); the driver only sets Accept.
type OktaDriver struct {
	httpClient *http.Client
	domain     string
}

var _ Driver = (*OktaDriver)(nil)

type oktaUser struct {
	ID        string      `json:"id"`
	Status    string      `json:"status"`
	Created   string      `json:"created"`
	LastLogin string      `json:"lastLogin"`
	Profile   oktaProfile `json:"profile"`
}

type oktaProfile struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Login       string `json:"login"`
	Title       string `json:"title"`
}

// oktaUsersPageLimit is Okta's maximum page size for GET /api/v1/users.
const oktaUsersPageLimit = "200"

func NewOktaDriver(httpClient *http.Client, domain string) *OktaDriver {
	return &OktaDriver{
		httpClient: httpClient,
		domain:     domain,
	}
}

func (d *OktaDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	endpoint := url.URL{
		Scheme:   "https",
		Host:     d.domain,
		Path:     "/api/v1/users",
		RawQuery: url.Values{"limit": {oktaUsersPageLimit}}.Encode(),
	}

	next := endpoint.String()

	var records []AccountRecord

	for range maxPaginationPages {
		users, linkNext, err := d.fetchUsersPage(ctx, next)
		if err != nil {
			return nil, err
		}

		for _, u := range users {
			email := u.Profile.Email
			if email == "" {
				email = u.Profile.Login
			}

			if email == "" {
				continue
			}

			record := AccountRecord{
				Email:       email,
				FullName:    oktaFullName(u.Profile),
				JobTitle:    u.Profile.Title,
				Active:      oktaActive(u.Status),
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
				ExternalID:  u.ID,
			}

			if t, ok := parseOktaTimestamp(u.Created); ok {
				record.CreatedAt = &t
			}

			if t, ok := parseOktaTimestamp(u.LastLogin); ok {
				record.LastLogin = &t
			}

			records = append(records, record)
		}

		if linkNext == "" {
			return records, nil
		}

		next = linkNext
	}

	return nil, fmt.Errorf("cannot list all okta accounts: %w", ErrPaginationLimitReached)
}

func (d *OktaDriver) fetchUsersPage(ctx context.Context, endpoint string) ([]oktaUser, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("cannot create okta users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("cannot execute okta users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("cannot fetch okta users: unexpected status %d", httpResp.StatusCode)
	}

	var users []oktaUser
	if err := json.NewDecoder(httpResp.Body).Decode(&users); err != nil {
		return nil, "", fmt.Errorf("cannot decode okta users response: %w", err)
	}

	// Okta emits one Link header per relation (self, next), so Header.Get
	// would return only the first. Join all of them before scanning for
	// rel="next".
	next, err := d.nextPageURL(strings.Join(httpResp.Header.Values("Link"), ", "))
	if err != nil {
		return nil, "", err
	}

	return users, next, nil
}

// nextPageURL returns the rel="next" pagination URL from an Okta Link header,
// or "" when the list is exhausted. It pins pagination to the configured org
// host: a `next` link pointing at any other host is rejected rather than
// followed, so the response cannot redirect the crawl off-tenant.
func (d *OktaDriver) nextPageURL(linkHeader string) (string, error) {
	raw := rfc5988.FindByRel(linkHeader, "next")
	if raw == "" {
		return "", nil
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("cannot parse okta next-page link: %w", err)
	}

	// Pin the next page to the same https origin: reject a scheme downgrade
	// (http), an explicit port, or a different host so a provider response
	// cannot redirect the crawl off-TLS, to another port, or off-tenant.
	if !strings.EqualFold(u.Scheme, "https") || u.Port() != "" || !strings.EqualFold(u.Hostname(), d.domain) {
		return "", fmt.Errorf("cannot follow okta next-page link: invalid target")
	}

	return u.String(), nil
}

func oktaFullName(p oktaProfile) string {
	if p.DisplayName != "" {
		return p.DisplayName
	}

	return strings.TrimSpace(strings.Join([]string{p.FirstName, p.LastName}, " "))
}

// oktaActive maps an Okta user status to the three-valued Active flag.
// SUSPENDED and DEPROVISIONED are the explicitly disabled/deactivated states;
// every other status (ACTIVE, PROVISIONED, STAGED, RECOVERY, PASSWORD_EXPIRED,
// LOCKED_OUT) is a usable account. An empty status leaves Active nil.
func oktaActive(status string) *bool {
	if status == "" {
		return nil
	}

	active := status != "SUSPENDED" && status != "DEPROVISIONED"

	return &active
}

func parseOktaTimestamp(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}

	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, value); err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}
