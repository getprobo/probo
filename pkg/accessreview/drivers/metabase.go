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

type MetabaseDriver struct {
	httpClient  *http.Client
	instanceURL string
}

var _ Driver = (*MetabaseDriver)(nil)

type metabaseUser struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	CommonName  string `json:"common_name"`
	IsActive    bool   `json:"is_active"`
	IsSuperuser bool   `json:"is_superuser"`
	LastLogin   string `json:"last_login"`
	DateJoined  string `json:"date_joined"`
}

func NewMetabaseDriver(httpClient *http.Client, instanceURL string) *MetabaseDriver {
	return &MetabaseDriver{
		httpClient:  httpClient,
		instanceURL: instanceURL,
	}
}

func (d *MetabaseDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	users, err := d.queryUsers(ctx)
	if err != nil {
		return nil, err
	}

	records := make([]AccountRecord, 0, len(users))

	for _, u := range users {
		if u.Email == "" {
			continue
		}

		record := AccountRecord{
			Email:       u.Email,
			FullName:    metabaseFullName(u),
			Role:        metabaseRole(u.IsSuperuser),
			Active:      new(u.IsActive),
			IsAdmin:     u.IsSuperuser,
			ExternalID:  strconv.Itoa(u.ID),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
			AccountType: coredata.AccessEntryAccountTypeUser,
		}

		if t, ok := parseMetabaseTimestamp(u.LastLogin); ok {
			record.LastLogin = &t
		}

		if t, ok := parseMetabaseTimestamp(u.DateJoined); ok {
			record.CreatedAt = &t
		}

		records = append(records, record)
	}

	return records, nil
}

func (d *MetabaseDriver) queryUsers(ctx context.Context) ([]metabaseUser, error) {
	baseURL, err := url.Parse(d.instanceURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse metabase instance url: %w", err)
	}

	endpoint := baseURL.JoinPath("api", "user")
	q := endpoint.Query()
	q.Set("status", "all")
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create metabase users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute metabase users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch metabase users: unexpected status %d", httpResp.StatusCode)
	}

	var users []metabaseUser
	if err := json.NewDecoder(httpResp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("cannot decode metabase users response: %w", err)
	}

	return users, nil
}

func metabaseFullName(u metabaseUser) string {
	if u.CommonName != "" {
		return u.CommonName
	}

	return strings.TrimSpace(strings.Join([]string{u.FirstName, u.LastName}, " "))
}

func metabaseRole(isSuperuser bool) string {
	if isSuperuser {
		return "Admin"
	}

	return "User"
}

func parseMetabaseTimestamp(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}

	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
	} {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}
