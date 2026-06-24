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
	"strconv"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// DocuSignDriver fetches account users from DocuSign via OAuth2-authenticated
// REST API requests. It resolves the data-center base URI for the configured
// account from the OAuth2 userinfo endpoint, then paginates through the
// eSignature Users API. The account is the one the user picked after OAuth
// (a DocuSign user may have access to several).
type DocuSignDriver struct {
	httpClient *http.Client
	accountID  string
}

var _ Driver = (*DocuSignDriver)(nil)

// errDocuSignUserInfoStatus marks a non-2xx userinfo response (typically a
// revoked or expired token). Callers that treat a dead token as terminal —
// the name resolver — branch on it via errors.Is; the driver and picker
// surface it as an ordinary failure.
var errDocuSignUserInfoStatus = errors.New("docusign userinfo returned a non-success status")

// docusignAccount is one entry of the /oauth/userinfo accounts list, carrying
// every field the driver, name resolver and picker key off.
type docusignAccount struct {
	AccountID   string `json:"account_id"`
	AccountName string `json:"account_name"`
	BaseURI     string `json:"base_uri"`
}

// fetchDocuSignAccounts returns the DocuSign accounts the access token can
// reach, from the OAuth2 userinfo endpoint. A non-2xx response yields
// errDocuSignUserInfoStatus; request and decode failures surface as ordinary
// errors so transient conditions stay distinguishable from a dead token.
func fetchDocuSignAccounts(ctx context.Context, httpClient *http.Client) ([]docusignAccount, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, docusignUserInfoEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create docusign userinfo request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute docusign userinfo request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: status %d", errDocuSignUserInfoStatus, httpResp.StatusCode)
	}

	var resp struct {
		Accounts []docusignAccount `json:"accounts"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode docusign userinfo response: %w", err)
	}

	return resp.Accounts, nil
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
		JobTitle              string `json:"jobTitle"`
	} `json:"users"`
	ResultSetSize string `json:"resultSetSize"`
	TotalSetSize  string `json:"totalSetSize"`
	StartPosition string `json:"startPosition"`
	EndPosition   string `json:"endPosition"`
}

const (
	docusignUserInfoEndpoint = "https://account.docusign.com/oauth/userinfo"
	docusignUsersPageSize    = 100
)

func NewDocuSignDriver(httpClient *http.Client, accountID string) *DocuSignDriver {
	return &DocuSignDriver{
		httpClient: httpClient,
		accountID:  accountID,
	}
}

func (d *DocuSignDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	baseURI, err := d.discoverBaseURI(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot discover docusign account: %w", err)
	}

	var records []AccountRecord

	startPosition := 0

	for range maxPaginationPages {
		resp, err := d.queryUsers(ctx, baseURI, d.accountID, startPosition)
		if err != nil {
			return nil, err
		}

		for _, u := range resp.Users {
			role := strings.TrimSpace(u.PermissionProfileName)

			roles := []string{}
			if role != "" {
				roles = []string{role}
			}

			record := AccountRecord{
				Email:       u.Email,
				FullName:    u.UserName,
				Roles:       roles,
				JobTitle:    u.JobTitle,
				Active:      new(strings.EqualFold(u.UserStatus, "active")),
				IsAdmin:     strings.EqualFold(u.IsAdmin, "True"),
				ExternalID:  u.UserID,
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
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

		totalSetSize, err := strconv.Atoi(resp.TotalSetSize)
		if err != nil {
			return nil, fmt.Errorf("cannot parse docusign total set size %q: %w", resp.TotalSetSize, err)
		}

		endPosition, err := strconv.Atoi(resp.EndPosition)
		if err != nil {
			return nil, fmt.Errorf("cannot parse docusign end position %q: %w", resp.EndPosition, err)
		}

		if totalSetSize == 0 || endPosition >= totalSetSize-1 {
			return records, nil
		}

		startPosition = endPosition + 1
	}

	return nil, fmt.Errorf("cannot list all docusign accounts: %w", ErrPaginationLimitReached)
}

// discoverBaseURI resolves the data-center base URI (e.g.
// https://na3.docusign.net) for the configured account from the OAuth2
// userinfo endpoint. DocuSign issues account-specific base URIs, so the
// eSignature REST host cannot be hardcoded.
func (d *DocuSignDriver) discoverBaseURI(ctx context.Context) (string, error) {
	accounts, err := fetchDocuSignAccounts(ctx, d.httpClient)
	if err != nil {
		return "", err
	}

	for _, account := range accounts {
		if account.AccountID == d.accountID {
			return account.BaseURI, nil
		}
	}

	return "", fmt.Errorf("docusign account not found in userinfo")
}

func (d *DocuSignDriver) queryUsers(ctx context.Context, baseURI string, accountID string, startPosition int) (*docusignUsersResponse, error) {
	u, err := url.JoinPath(baseURI, "restapi", "v2.1", "accounts", url.PathEscape(accountID), "users")
	if err != nil {
		return nil, fmt.Errorf("cannot build docusign users URL: %w", err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("cannot parse docusign users URL: %w", err)
	}

	q := parsed.Query()
	q.Set("additional_info", "true")
	q.Set("count", strconv.Itoa(docusignUsersPageSize))
	q.Set("start_position", strconv.Itoa(startPosition))
	parsed.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
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
