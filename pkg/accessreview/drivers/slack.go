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
	"fmt"
	"net/http"

	"encoding/json"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"net/url"
)

type SlackDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*SlackDriver)(nil)

type slackUsersListResponse struct {
	OK               bool                  `json:"ok"`
	Error            string                `json:"error,omitempty"`
	Members          []slackMember         `json:"members"`
	ResponseMetadata slackResponseMetadata `json:"response_metadata"`
}

type slackResponseMetadata struct {
	NextCursor string `json:"next_cursor"`
}

type slackMember struct {
	ID                string       `json:"id"`
	Name              string       `json:"name"`
	RealName          string       `json:"real_name"`
	Deleted           bool         `json:"deleted"`
	IsAdmin           bool         `json:"is_admin"`
	IsOwner           bool         `json:"is_owner"`
	IsPrimaryOwner    bool         `json:"is_primary_owner"`
	IsRestricted      bool         `json:"is_restricted"`
	IsUltraRestricted bool         `json:"is_ultra_restricted"`
	IsBot             bool         `json:"is_bot"`
	IsAppUser         bool         `json:"is_app_user"`
	Has2FA            bool         `json:"has_2fa"`
	Updated           int          `json:"updated"`
	Profile           slackProfile `json:"profile"`
}

type slackProfile struct {
	Email string `json:"email"`
	Title string `json:"title"`
}

const (
	slackUsersListPath = "/users.list"
	slackAuthTestPath  = "/auth.test"
)

func NewSlackDriver(httpClient *http.Client, baseURL string) *SlackDriver {
	return &SlackDriver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (d *SlackDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records []AccountRecord
		cursor  string
	)

	for range maxPaginationPages {
		resp, err := d.queryUsers(ctx, cursor)
		if err != nil {
			return nil, err
		}

		if !resp.OK {
			return nil, fmt.Errorf("slack users.list request failed: %s", resp.Error)
		}

		for _, m := range resp.Members {
			if m.ID == "USLACKBOT" {
				continue
			}

			accountType := coredata.AccessReviewEntryAccountTypeUser
			if m.IsBot || m.IsAppUser {
				accountType = coredata.AccessReviewEntryAccountTypeServiceAccount
			}

			record := AccountRecord{
				Email:       m.Profile.Email,
				FullName:    m.RealName,
				JobTitle:    m.Profile.Title,
				Roles:       slackRoles(m),
				Active:      new(!m.Deleted),
				IsAdmin:     new(m.IsAdmin || m.IsOwner || m.IsPrimaryOwner),
				ExternalID:  m.ID,
				MFAStatus:   slackMFAStatus(m.Has2FA),
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: accountType,
			}

			// Note: Slack's Updated field is the profile update time, not
			// the last login time, so we intentionally do not map it.

			if record.Email != "" {
				records = append(records, record)
			}
		}

		if resp.ResponseMetadata.NextCursor == "" {
			return records, nil
		}

		cursor = resp.ResponseMetadata.NextCursor
	}

	return nil, fmt.Errorf("cannot list all slack accounts: %w", ErrPaginationLimitReached)
}

func (d *SlackDriver) queryUsers(ctx context.Context, cursor string) (*slackUsersListResponse, error) {
	endpoint, err := url.JoinPath(d.baseURL, slackUsersListPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build slack users.list URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create slack users.list request: %w", err)
	}

	q := req.URL.Query()
	q.Set("limit", "200")

	if cursor != "" {
		q.Set("cursor", cursor)
	}

	req.URL.RawQuery = q.Encode()

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute slack users.list request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch slack users: unexpected status %d", httpResp.StatusCode)
	}

	var resp slackUsersListResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode slack users.list response: %w", err)
	}

	return &resp, nil
}

func slackRoles(m slackMember) []string {
	switch {
	case m.IsPrimaryOwner:
		return []string{"Primary Owner"}
	case m.IsOwner:
		return []string{"Owner"}
	case m.IsAdmin:
		return []string{"Admin"}
	case m.IsUltraRestricted:
		return []string{"Ultra Restricted"}
	case m.IsRestricted:
		return []string{"Restricted"}
	default:
		return []string{"Member"}
	}
}

func slackMFAStatus(has2FA bool) coredata.MFAStatus {
	if has2FA {
		return coredata.MFAStatusEnabled
	}

	return coredata.MFAStatusDisabled
}

// slackNameResolver resolves the Slack workspace name via auth.test.
type slackNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

func NewSlackNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &slackNameResolver{httpClient: httpClient, baseURL: baseURL}
}

func (r *slackNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	endpoint, err := url.JoinPath(r.baseURL, slackAuthTestPath)
	if err != nil {
		return "", fmt.Errorf("cannot build slack auth.test URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create slack auth.test request: %w", err)
	}

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute slack auth.test request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	var resp struct {
		OK   bool   `json:"ok"`
		Team string `json:"team"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode slack auth.test response: %w", err)
	}

	if !resp.OK {
		return "", fmt.Errorf("slack auth.test returned ok=false")
	}

	return resp.Team, nil
}

func slackSource() Factory {
	return provider.Over(func(
		ctx context.Context,
		credential connector.HTTPCredential,
		opened *provider.Handle,
		logger *log.Logger,
	) (Driver, error) {
		return capable(
			NewSlackDriver(credential.Client, opened.Endpoints.APIBase),
			NewSlackNameResolver(credential.Client, opened.Endpoints.APIBase),
			nil,
		), nil
	})
}
