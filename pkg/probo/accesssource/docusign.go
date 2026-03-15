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

package accesssource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	docusignUserInfoEndpoint = "https://account-d.docusign.com/oauth/userinfo"
	docusignUsersPageSize    = 100
)

// DocuSignDriver fetches account users from DocuSign via OAuth2-authenticated
// REST API requests. It auto-discovers the account ID and base URI from the
// OAuth2 userinfo endpoint, then paginates through the eSignature Users API.
type DocuSignDriver struct {
	httpClient *http.Client
}

func NewDocuSignDriver(httpClient *http.Client) *DocuSignDriver {
	return &DocuSignDriver{
		httpClient: httpClient,
	}
}

func (d *DocuSignDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	accountID, baseURI, err := d.discoverAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot discover docusign account: %w", err)
	}

	var records []AccountRecord
	startPosition := 0

	for {
		resp, err := d.queryUsers(ctx, baseURI, accountID, startPosition)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Users {
			record := AccountRecord{
				Email:      u.Email,
				FullName:   u.UserName,
				Role:       u.PermissionProfileName,
				Active:     strings.EqualFold(u.UserStatus, "active"),
				IsAdmin:    strings.EqualFold(u.IsAdmin, "True"),
				ExternalID: u.UserID,
				MFAStatus:  coredata.MFAStatusUnknown,
				AuthMethod: coredata.AccessEntryAuthMethodUnknown,
			}

			if u.LastLogin != "" {
				if t, err := time.Parse(time.RFC3339, u.LastLogin); err == nil {
					record.LastLogin = &t
				}
			}

			if u.CreatedDateTime != "" {
				if t, err := time.Parse(time.RFC3339, u.CreatedDateTime); err == nil {
					record.CreatedAt = &t
				}
			}

			if record.Email != "" {
				records = append(records, record)
			}
		}

		totalSetSize, _ := strconv.Atoi(resp.TotalSetSize)
		endPosition, _ := strconv.Atoi(resp.EndPosition)

		if totalSetSize == 0 || endPosition >= totalSetSize-1 {
			break
		}

		startPosition = endPosition + 1
	}

	return records, nil
}

func (d *DocuSignDriver) discoverAccount(ctx context.Context) (accountID string, baseURI string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, docusignUserInfoEndpoint, nil)
	if err != nil {
		return "", "", fmt.Errorf("cannot create docusign userinfo request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("cannot execute docusign userinfo request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", "", fmt.Errorf("cannot fetch docusign userinfo: unexpected status %d", httpResp.StatusCode)
	}

	var resp docusignUserInfoResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", "", fmt.Errorf("cannot decode docusign userinfo response: %w", err)
	}

	for _, account := range resp.Accounts {
		if account.IsDefault {
			return account.AccountID, account.BaseURI, nil
		}
	}

	if len(resp.Accounts) > 0 {
		return resp.Accounts[0].AccountID, resp.Accounts[0].BaseURI, nil
	}

	return "", "", fmt.Errorf("no docusign accounts found in userinfo response")
}

func (d *DocuSignDriver) queryUsers(ctx context.Context, baseURI string, accountID string, startPosition int) (*docusignUsersResponse, error) {
	url := fmt.Sprintf("%s/restapi/v2.1/accounts/%s/users?count=%d&start_position=%d",
		baseURI, accountID, docusignUsersPageSize, startPosition)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create docusign users request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute docusign users request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch docusign users: unexpected status %d", httpResp.StatusCode)
	}

	var resp docusignUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode docusign users response: %w", err)
	}

	return &resp, nil
}

type docusignUserInfoResponse struct {
	Accounts []struct {
		AccountID string `json:"account_id"`
		IsDefault bool   `json:"is_default"`
		BaseURI   string `json:"base_uri"`
	} `json:"accounts"`
}

type docusignUsersResponse struct {
	Users []struct {
		UserID                string `json:"userId"`
		UserName              string `json:"userName"`
		Email                 string `json:"email"`
		UserStatus            string `json:"userStatus"`
		IsAdmin               string `json:"isAdmin"`
		CreatedDateTime       string `json:"createdDateTime"`
		LastLogin             string `json:"lastLogin"`
		PermissionProfileName string `json:"permissionProfileName"`
	} `json:"users"`
	ResultSetSize string `json:"resultSetSize"`
	TotalSetSize  string `json:"totalSetSize"`
	StartPosition string `json:"startPosition"`
	EndPosition   string `json:"endPosition"`
}
