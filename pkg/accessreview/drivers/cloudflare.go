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
	"strconv"

	"go.probo.inc/probo/pkg/coredata"
)

// CloudflareDriver fetches account members from the Cloudflare API for a
// single selected account.
type CloudflareDriver struct {
	httpClient *http.Client
	accountID  string
	baseURL    string
}

var _ Driver = (*CloudflareDriver)(nil)

const (
	cloudflareAccountsPath   = "/accounts"
	cloudflareMembersSegment = "members"
)

// cloudflareDefaultBaseURL is the Cloudflare API root. It backs only the
// exported ListCloudflareOrganizations, and only when its caller resolves no
// APIBase for the provider (unregistered, or registered without one). Every
// other path goes through the injected baseURL instead.
const cloudflareDefaultBaseURL = "https://api.cloudflare.com/client/v4"

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

// NewCloudflareDriver builds a driver scoped to accountID against baseURL
// (e.g. https://api.cloudflare.com/client/v4).
func NewCloudflareDriver(httpClient *http.Client, accountID, baseURL string) *CloudflareDriver {
	return &CloudflareDriver{
		httpClient: httpClient,
		accountID:  accountID,
		baseURL:    baseURL,
	}
}

func (d *CloudflareDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	return d.queryAllMembers(ctx, d.accountID)
}

func (d *CloudflareDriver) queryAllMembers(ctx context.Context, accountID string) ([]AccountRecord, error) {
	var records []AccountRecord

	for page := range maxPaginationPages {
		resp, err := d.queryMembers(ctx, accountID, page+1)
		if err != nil {
			return nil, err
		}

		for _, m := range resp.Result {
			roles := make([]string, 0, len(m.Roles))
			for _, r := range m.Roles {
				roles = append(roles, r.Name)
			}

			if len(roles) == 0 {
				roles = []string{"Member"}
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
				Roles:       roles,
				Active:      new(m.Status == "accepted"),
				IsAdmin:     isAdmin,
				ExternalID:  m.ID,
				MFAStatus:   mfaStatus,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
			}

			if record.Email != "" {
				records = append(records, record)
			}
		}

		if page+1 >= resp.ResultInfo.TotalPages {
			return records, nil
		}
	}

	return nil, fmt.Errorf("cannot list all cloudflare members: %w", ErrPaginationLimitReached)
}

func (d *CloudflareDriver) queryMembers(ctx context.Context, accountID string, page int) (*cloudflareListMembersResponse, error) {
	u, err := url.JoinPath(d.baseURL, cloudflareAccountsPath, url.PathEscape(accountID), cloudflareMembersSegment)
	if err != nil {
		return nil, fmt.Errorf("cannot build cloudflare members URL: %w", err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("cannot parse cloudflare members URL: %w", err)
	}

	q := parsed.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("per_page", "50")
	parsed.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
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

// cloudflareNameResolver resolves the Cloudflare account name for the
// selected account.
type cloudflareNameResolver struct {
	httpClient *http.Client
	accountID  string
	baseURL    string
}

func NewCloudflareNameResolver(httpClient *http.Client, accountID, baseURL string) NameResolver {
	return &cloudflareNameResolver{httpClient: httpClient, accountID: accountID, baseURL: baseURL}
}

func (r *cloudflareNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.accountID == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath(r.baseURL, cloudflareAccountsPath, url.PathEscape(r.accountID))
	if err != nil {
		return "", fmt.Errorf("cannot build cloudflare account URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create cloudflare account request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute cloudflare account request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nameStatusError("cloudflare account", httpResp.StatusCode)
	}

	var resp struct {
		Result struct {
			Name string `json:"name"`
		} `json:"result"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode cloudflare account response: %w", err)
	}

	return resp.Result.Name, nil
}

// ListCloudflareOrganizations fetches the Cloudflare accounts the
// authenticated token can access, from baseURL ("" for the Cloudflare SaaS API).
func ListCloudflareOrganizations(ctx context.Context, httpClient *http.Client, baseURL string) ([]Organization, error) {
	if baseURL == "" {
		baseURL = cloudflareDefaultBaseURL
	}

	var result []Organization

	for page := range maxPaginationPages {
		resp, err := queryCloudflareAccounts(ctx, httpClient, baseURL, page+1)
		if err != nil {
			return nil, err
		}

		for _, a := range resp.Result {
			displayName := a.Name
			if displayName == "" {
				displayName = a.ID
			}

			result = append(result, Organization{Slug: a.ID, DisplayName: displayName})
		}

		if page+1 >= resp.ResultInfo.TotalPages {
			return result, nil
		}
	}

	return nil, fmt.Errorf("cannot list all cloudflare accounts: %w", ErrPaginationLimitReached)
}

func queryCloudflareAccounts(
	ctx context.Context,
	httpClient *http.Client,
	baseURL string,
	page int,
) (*cloudflareListAccountsResponse, error) {
	endpoint, err := url.JoinPath(baseURL, cloudflareAccountsPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build cloudflare accounts URL: %w", err)
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("cannot parse cloudflare accounts URL: %w", err)
	}

	q := parsed.Query()
	q.Set("page", strconv.Itoa(page))
	// Cloudflare requires per_page in the range 5..50; per_page=1 is rejected
	// with a 400 (which, before terminal classification, caused a 400 storm).
	// Do not "optimize" this back down to 1.
	q.Set("per_page", "50")
	parsed.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create cloudflare accounts request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := httpClient.Do(req)
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
