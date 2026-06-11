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
	"strconv"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

const grafanaUsersPageSize = 100

// GrafanaDriver fetches organization users from the Grafana HTTP API using
// Bearer-token authenticated REST requests against a configured Grafana base
// URL (Grafana Cloud stack URL or self-hosted Grafana URL).
type GrafanaDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*GrafanaDriver)(nil)

type grafanaOrgUser struct {
	UserID     int    `json:"userId"`
	Email      string `json:"email"`
	Login      string `json:"login"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	LastSeenAt string `json:"lastSeenAt"`
	IsDisabled *bool  `json:"isDisabled"`
}

func NewGrafanaDriver(httpClient *http.Client, baseURL string) *GrafanaDriver {
	return &GrafanaDriver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (d *GrafanaDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	records := make([]AccountRecord, 0)

	for page := 1; page <= maxPaginationPages; page++ {
		users, err := d.queryOrgUsers(ctx, page)
		if err != nil {
			return nil, err
		}

		for _, u := range users {
			email := strings.TrimSpace(u.Email)
			if email == "" {
				email = strings.TrimSpace(u.Login)
			}

			if email == "" {
				continue
			}

			record := AccountRecord{
				Email:       email,
				FullName:    strings.TrimSpace(u.Name),
				Role:        strings.TrimSpace(u.Role),
				IsAdmin:     strings.EqualFold(strings.TrimSpace(u.Role), "Admin"),
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  strconv.Itoa(u.UserID),
			}

			if u.IsDisabled != nil {
				active := !*u.IsDisabled
				record.Active = &active
			}

			if u.LastSeenAt != "" {
				if t, err := time.Parse(time.RFC3339, u.LastSeenAt); err == nil {
					record.LastLogin = &t
				} else if t, err := time.Parse(time.RFC3339Nano, u.LastSeenAt); err == nil {
					record.LastLogin = &t
				}
			}

			records = append(records, record)
		}

		if len(users) < grafanaUsersPageSize {
			return records, nil
		}
	}

	return nil, fmt.Errorf("cannot list all grafana accounts: %w", ErrPaginationLimitReached)
}

func (d *GrafanaDriver) queryOrgUsers(ctx context.Context, page int) ([]grafanaOrgUser, error) {
	u, err := url.Parse(d.baseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse grafana base URL: %w", err)
	}

	u = u.JoinPath("api", "org", "users")
	q := u.Query()
	q.Set("perpage", strconv.Itoa(grafanaUsersPageSize))
	q.Set("page", strconv.Itoa(page))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create grafana users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute grafana users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch grafana users: unexpected status %d", httpResp.StatusCode)
	}

	var users []grafanaOrgUser
	if err := json.NewDecoder(httpResp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("cannot decode grafana users response: %w", err)
	}

	return users, nil
}

// grafanaNameResolver resolves the Grafana organization display name by
// querying /api/org on the configured Grafana instance.
type grafanaNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

var _ NameResolver = (*grafanaNameResolver)(nil)

type grafanaOrg struct {
	Name string `json:"name"`
}

func NewGrafanaNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &grafanaNameResolver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (r *grafanaNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	u, err := url.Parse(r.baseURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse grafana base URL: %w", err)
	}

	u = u.JoinPath("api", "org")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("cannot create grafana organization request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute grafana organization request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch grafana organization: unexpected status %d", httpResp.StatusCode)
	}

	var org grafanaOrg
	if err := json.NewDecoder(httpResp.Body).Decode(&org); err != nil {
		return "", fmt.Errorf("cannot decode grafana organization response: %w", err)
	}

	return strings.TrimSpace(org.Name), nil
}
