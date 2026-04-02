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
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/rfc5988"
)

// SentryDriver fetches organization members from Sentry via Bearer
// token-authenticated REST API requests.
type SentryDriver struct {
	httpClient *http.Client
	orgSlug    string
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

func NewSentryDriver(httpClient *http.Client, orgSlug string) *SentryDriver {
	return &SentryDriver{
		httpClient: httpClient,
		orgSlug:    orgSlug,
	}
}

func (d *SentryDriver) resolveOrgSlug(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://sentry.io/api/0/organizations/?member=true", nil)
	if err != nil {
		return "", fmt.Errorf("cannot create sentry organizations request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot fetch sentry organizations: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cannot fetch sentry organizations: status %d", resp.StatusCode)
	}

	var orgs []struct {
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return "", fmt.Errorf("cannot decode sentry organizations response: %w", err)
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

	nextURL := fmt.Sprintf(
		"https://sentry.io/api/0/organizations/%s/members/",
		orgSlug,
	)

	for range maxPaginationPages {
		members, linkHeader, err := d.queryMembers(ctx, nextURL)
		if err != nil {
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

			isAdmin := m.OrgRole == "admin" || m.OrgRole == "owner"

			mfaStatus := coredata.MFAStatusUnknown
			if m.User != nil {
				if m.User.Has2FA {
					mfaStatus = coredata.MFAStatusEnabled
				} else {
					mfaStatus = coredata.MFAStatusDisabled
				}
			}

			authMethod := sentryAuthMethod(m.Flags, m.User)

			record := AccountRecord{
				Email:       m.Email,
				FullName:    fullName,
				Role:        m.OrgRole,
				Active:      active,
				IsAdmin:     isAdmin,
				ExternalID:  m.ID,
				MFAStatus:   mfaStatus,
				AuthMethod:  authMethod,
				AccountType: coredata.AccessEntryAccountTypeUser,
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

		nextURL = sentryNextLink(linkHeader)
		if nextURL == "" {
			return records, nil
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

func sentryAuthMethod(flags map[string]bool, user *sentryUser) coredata.AccessEntryAuthMethod {
	if flags["sso:linked"] {
		return coredata.AccessEntryAuthMethodSSO
	}
	if user != nil && user.HasPasswordAuth {
		return coredata.AccessEntryAuthMethodPassword
	}
	return coredata.AccessEntryAuthMethodUnknown
}
