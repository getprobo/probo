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
	orgs, err := ListSentryOrganizations(ctx, d.httpClient)
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

	nextURL, err := url.JoinPath("https://sentry.io", "api", "0", "organizations", url.PathEscape(orgSlug), "members")
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
				IsAdmin:     isAdmin,
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
