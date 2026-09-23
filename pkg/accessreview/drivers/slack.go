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
	"fmt"
	"net/http"
	"net/url"

	"go.probo.inc/probo/pkg/coredata"
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
	Has2FA            *bool        `json:"has_2fa"`
	Updated           int          `json:"updated"`
	Profile           slackProfile `json:"profile"`
}

type slackProfile struct {
	Email string `json:"email"`
	Title string `json:"title"`
}

const (
	slackUsersListPath = "/users.list"
	slackUsersInfoPath = "/users.info"
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
		query := url.Values{"limit": {"200"}}
		if cursor != "" {
			query.Set("cursor", cursor)
		}

		var resp slackUsersListResponse
		if err := slackGet(ctx, d.httpClient, d.baseURL, slackUsersListPath, query, &resp); err != nil {
			return nil, fmt.Errorf("cannot list slack users: %w", err)
		}

		if !resp.OK {
			return nil, &SlackAPIError{Method: slackUsersListPath, Code: resp.Error}
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

// slackMFAStatus maps has_2fa, which Slack only returns to an admin or owner
// caller: a bot token or a non-admin user token never sees it.
func slackMFAStatus(has2FA *bool) coredata.MFAStatus {
	switch {
	case has2FA == nil:
		return coredata.MFAStatusUnknown
	case *has2FA:
		return coredata.MFAStatusEnabled
	default:
		return coredata.MFAStatusDisabled
	}
}

// SlackAPIError is a Slack Web API answer of ok=false. Slack reports a
// revoked or invalid token this way under HTTP 200.
type SlackAPIError struct {
	Method string
	Code   string
}

func (e *SlackAPIError) Error() string {
	return fmt.Sprintf("slack %s request failed: %s", e.Method, e.Code)
}

// CheckSlackInstallerIsAdmin rejects a token whose user is not a workspace
// admin or owner: Slack hides has_2fa from anyone else, so such an install
// could never report MFA.
func CheckSlackInstallerIsAdmin(ctx context.Context, httpClient *http.Client, baseURL string) error {
	var identity struct {
		OK     bool   `json:"ok"`
		Error  string `json:"error"`
		UserID string `json:"user_id"`
	}
	if err := slackGet(ctx, httpClient, baseURL, slackAuthTestPath, nil, &identity); err != nil {
		return fmt.Errorf("cannot identify slack installer: %w", err)
	}

	if !identity.OK {
		return &SlackAPIError{Method: slackAuthTestPath, Code: identity.Error}
	}

	var info struct {
		OK    bool        `json:"ok"`
		Error string      `json:"error"`
		User  slackMember `json:"user"`
	}
	if err := slackGet(ctx, httpClient, baseURL, slackUsersInfoPath, url.Values{"user": {identity.UserID}}, &info); err != nil {
		return fmt.Errorf("cannot load slack installer: %w", err)
	}

	if !info.OK {
		return &SlackAPIError{Method: slackUsersInfoPath, Code: info.Error}
	}

	if !info.User.IsAdmin && !info.User.IsOwner && !info.User.IsPrimaryOwner {
		return &InstallRejectedError{
			Message: "Slack must be connected by a workspace admin or owner, the only role Slack shares MFA status with.",
		}
	}

	return nil
}

func slackGet(ctx context.Context, httpClient *http.Client, baseURL, path string, query url.Values, out any) error {
	endpoint, err := url.JoinPath(baseURL, path)
	if err != nil {
		return fmt.Errorf("cannot build slack %s URL: %w", path, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("cannot create slack %s request: %w", path, err)
	}

	req.URL.RawQuery = query.Encode()

	httpResp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot execute slack %s request: %w", path, err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return fmt.Errorf("cannot execute slack %s request: unexpected status %d", path, httpResp.StatusCode)
	}

	if err := json.NewDecoder(httpResp.Body).Decode(out); err != nil {
		return fmt.Errorf("cannot decode slack %s response: %w", path, err)
	}

	return nil
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
	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
		Team  string `json:"team"`
	}
	if err := slackGet(ctx, r.httpClient, r.baseURL, slackAuthTestPath, nil, &resp); err != nil {
		return "", fmt.Errorf("cannot resolve slack workspace name: %w", err)
	}

	if !resp.OK {
		return "", &SlackAPIError{Method: slackAuthTestPath, Code: resp.Error}
	}

	return resp.Team, nil
}
