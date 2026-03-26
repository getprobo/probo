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

// CloudflareDriver fetches account members from the Cloudflare API.
type CloudflareDriver struct {
	httpClient *http.Client
}

func NewCloudflareDriver(httpClient *http.Client) *CloudflareDriver {
	return &CloudflareDriver{
		httpClient: httpClient,
	}
}

func (d *CloudflareDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	accounts, err := d.queryAllAccounts(ctx)
	if err != nil {
		return nil, err
	}

	var records []AccountRecord

	for _, account := range accounts {
		members, err := d.queryAllMembers(ctx, account.ID)
		if err != nil {
			return nil, fmt.Errorf("cannot fetch members for cloudflare account %s: %w", account.ID, err)
		}

		records = append(records, members...)
	}

	return records, nil
}

func (d *CloudflareDriver) queryAllAccounts(ctx context.Context) ([]cloudflareAccount, error) {
	var accounts []cloudflareAccount

	for page := 1; ; page++ {
		resp, err := d.queryAccounts(ctx, page)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, resp.Result...)

		if page >= resp.ResultInfo.TotalPages {
			break
		}
	}

	return accounts, nil
}

func (d *CloudflareDriver) queryAccounts(ctx context.Context, page int) (*cloudflareListAccountsResponse, error) {
	url := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/accounts?page=%d&per_page=50",
		page,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create cloudflare accounts request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute cloudflare accounts request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch cloudflare accounts: unexpected status %d", httpResp.StatusCode)
	}

	var resp cloudflareListAccountsResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode cloudflare accounts response: %w", err)
	}

	return &resp, nil
}

func (d *CloudflareDriver) queryAllMembers(ctx context.Context, accountID string) ([]AccountRecord, error) {
	var records []AccountRecord

	for page := 1; ; page++ {
		resp, err := d.queryMembers(ctx, accountID, page)
		if err != nil {
			return nil, err
		}

		for _, m := range resp.Result {
			roles := make([]string, 0, len(m.Roles))
			for _, r := range m.Roles {
				roles = append(roles, r.Name)
			}

			role := "Member"
			if len(roles) > 0 {
				role = roles[0]
			}

			isAdmin := false
			for _, r := range m.Roles {
				if r.Name == "Super Administrator - All Privileges" || r.Name == "Administrator" {
					isAdmin = true
					break
				}
			}

			mfaStatus := coredata.MFAStatusUnknown
			if m.User.TwoFactorEnabled {
				mfaStatus = coredata.MFAStatusEnabled
			}

			record := AccountRecord{
				Email:       m.User.Email,
				FullName:    m.User.FirstName + " " + m.User.LastName,
				Role:        role,
				Active:      m.Status == "accepted",
				IsAdmin:     isAdmin,
				ExternalID:  m.ID,
				MFAStatus:   mfaStatus,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
			}

			if record.Email != "" {
				records = append(records, record)
			}
		}

		if page >= resp.ResultInfo.TotalPages {
			break
		}
	}

	return records, nil
}

func (d *CloudflareDriver) queryMembers(ctx context.Context, accountID string, page int) (*cloudflareListMembersResponse, error) {
	url := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/accounts/%s/members?page=%d&per_page=50",
		accountID,
		page,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create cloudflare members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute cloudflare members request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch cloudflare members: unexpected status %d", httpResp.StatusCode)
	}

	var resp cloudflareListMembersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode cloudflare members response: %w", err)
	}

	return &resp, nil
}

type cloudflareAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type cloudflareListAccountsResponse struct {
	Result     []cloudflareAccount  `json:"result"`
	ResultInfo cloudflareResultInfo `json:"result_info"`
}

type cloudflareResultInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
}

type cloudflareListMembersResponse struct {
	Result []struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		User   struct {
			ID               string `json:"id"`
			FirstName        string `json:"first_name"`
			LastName         string `json:"last_name"`
			Email            string `json:"email"`
			TwoFactorEnabled bool   `json:"two_factor_authentication_enabled"`
		} `json:"user"`
		Roles []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"roles"`
	} `json:"result"`
	ResultInfo cloudflareResultInfo `json:"result_info"`
}
