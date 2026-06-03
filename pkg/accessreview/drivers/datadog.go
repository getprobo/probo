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

// DatadogDriver lists Datadog org members via GET /api/v2/users. The API
// host is per-customer (api.<domain>), captured during the OAuth callback
// and stored on the connector settings.
type DatadogDriver struct {
	httpClient *http.Client
	domain     string // e.g. "us3.datadoghq.com"
}

var _ Driver = (*DatadogDriver)(nil)

// NewDatadogDriver wraps the connection's SSRF-protected transport with a
// retrying transport for transient 5xx, matching the canonical sibling
// drivers (heroku.go, pagerduty.go). The caller's *http.Client is not
// mutated.
func NewDatadogDriver(httpClient *http.Client, domain string) *DatadogDriver {
	return &DatadogDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		domain: domain,
	}
}

const datadogPageSize = 100

type datadogUsersResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Email          string `json:"email"`
			Name           string `json:"name"`
			Handle         string `json:"handle"`
			Title          string `json:"title"`
			Disabled       bool   `json:"disabled"`
			Status         string `json:"status"`
			Verified       bool   `json:"verified"`
			ServiceAccount bool   `json:"service_account"`
			MFAEnabled     bool   `json:"mfa_enabled"`
			CreatedAt      string `json:"created_at"`
			ModifiedAt     string `json:"modified_at"`
		} `json:"attributes"`
		Relationships struct {
			Roles struct {
				Data []struct {
					ID string `json:"id"`
				} `json:"data"`
			} `json:"roles"`
		} `json:"relationships"`
	} `json:"data"`
	Included []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			Name string `json:"name"`
		} `json:"attributes"`
	} `json:"included"`
}

func (d *DatadogDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	for page := range maxPaginationPages {
		resp, err := d.queryUsers(ctx, page)
		if err != nil {
			return nil, err
		}

		roleNames := make(map[string]string, len(resp.Included))
		for _, inc := range resp.Included {
			if inc.Type == "roles" {
				roleNames[inc.ID] = inc.Attributes.Name
			}
		}

		for _, u := range resp.Data {
			active := !u.Attributes.Disabled

			var (
				role    string
				isAdmin bool
			)

			for _, r := range u.Relationships.Roles.Data {
				name := roleNames[r.ID]
				if role == "" {
					role = name
				}

				if strings.Contains(strings.ToLower(name), "admin") {
					isAdmin = true
					role = name
				}
			}

			accountType := coredata.AccessEntryAccountTypeUser
			if u.Attributes.ServiceAccount {
				accountType = coredata.AccessEntryAccountTypeServiceAccount
			}

			mfaStatus := coredata.MFAStatusDisabled
			if u.Attributes.MFAEnabled {
				mfaStatus = coredata.MFAStatusEnabled
			}

			records = append(records, AccountRecord{
				Email:     u.Attributes.Email,
				FullName:  u.Attributes.Name,
				Role:      role,
				JobTitle:  u.Attributes.Title,
				Active:    &active,
				IsAdmin:   isAdmin,
				MFAStatus: mfaStatus,
				// Datadog's /api/v2/users does not expose the login method
				// used (no allowed_login_methods in the schema), so the
				// auth method is unknown.
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: accountType,
				ExternalID:  u.ID,
				CreatedAt:   parseDatadogTime(u.Attributes.CreatedAt),
			})
		}

		if len(resp.Data) < datadogPageSize {
			return records, nil
		}
	}

	return nil, fmt.Errorf("cannot list all datadog users: %w", ErrPaginationLimitReached)
}

func (d *DatadogDriver) queryUsers(ctx context.Context, page int) (*datadogUsersResponse, error) {
	q := url.Values{}
	q.Set("page[size]", strconv.Itoa(datadogPageSize))
	q.Set("page[number]", strconv.Itoa(page))

	endpoint := url.URL{
		Scheme:   "https",
		Host:     "api." + d.domain,
		Path:     "/api/v2/users",
		RawQuery: q.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create datadog users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot list datadog users: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot list datadog users: unexpected status %d", resp.StatusCode)
	}

	var out datadogUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("cannot decode datadog users response: %w", err)
	}

	return &out, nil
}

func parseDatadogTime(s string) *time.Time {
	if s == "" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}

	return &t
}
