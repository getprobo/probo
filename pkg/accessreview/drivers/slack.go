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

	"go.probo.inc/probo/pkg/coredata"
)

type SlackDriver struct {
	httpClient *http.Client
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

const slackUsersListEndpoint = "https://slack.com/api/users.list"

func NewSlackDriver(httpClient *http.Client) *SlackDriver {
	return &SlackDriver{
		httpClient: httpClient,
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

			accountType := coredata.AccessEntryAccountTypeUser
			if m.IsBot || m.IsAppUser {
				accountType = coredata.AccessEntryAccountTypeServiceAccount
			}

			record := AccountRecord{
				Email:       m.Profile.Email,
				FullName:    m.RealName,
				JobTitle:    m.Profile.Title,
				Role:        slackRole(m),
				Active:      new(!m.Deleted),
				IsAdmin:     m.IsAdmin || m.IsOwner || m.IsPrimaryOwner,
				ExternalID:  m.ID,
				MFAStatus:   slackMFAStatus(m.Has2FA),
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, slackUsersListEndpoint, nil)
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

func slackRole(m slackMember) string {
	switch {
	case m.IsPrimaryOwner:
		return "Primary Owner"
	case m.IsOwner:
		return "Owner"
	case m.IsAdmin:
		return "Admin"
	case m.IsUltraRestricted:
		return "Ultra Restricted"
	case m.IsRestricted:
		return "Restricted"
	default:
		return "Member"
	}
}

func slackMFAStatus(has2FA bool) coredata.MFAStatus {
	if has2FA {
		return coredata.MFAStatusEnabled
	}
	return coredata.MFAStatusDisabled
}
