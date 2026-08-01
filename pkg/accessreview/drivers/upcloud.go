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
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	upcloudAccountPath = "account"
	upcloudListPath    = "list"
	upcloudDetailsPath = "details"
)

// UpCloudDriver lists the main account and its sub-accounts via UpCloud's
// account/list endpoint, then enriches each with account/details/{username}
// (email, first/last name), using a pre-authenticated HTTP client (Bearer API
// token) attached by the connection transport.
//
// Notes on data quality:
//   - Neither endpoint exposes an account-status or MFA field, so Active
//     stays nil and MFAStatus Unknown.
//   - An account the token cannot read (403/404) keeps its list-only
//     fields, with a blank email and the username as its name. Any other
//     details failure aborts the run rather than emit a half-identified
//     record.
type UpCloudDriver struct {
	httpClient *http.Client
	logger     *log.Logger
	baseURL    string
}

var _ Driver = (*UpCloudDriver)(nil)

// NewUpCloudDriver builds a driver against baseURL, the versioned UpCloud API
// origin (e.g. https://api.upcloud.com/1.3).
func NewUpCloudDriver(httpClient *http.Client, logger *log.Logger, baseURL string) *UpCloudDriver {
	return &UpCloudDriver{
		httpClient: httpClient,
		logger:     logger,
		baseURL:    baseURL,
	}
}

type upcloudAccountListResponse struct {
	Accounts struct {
		Account []upcloudAccount `json:"account"`
	} `json:"accounts"`
}

type upcloudAccount struct {
	Username string `json:"username"`
	Type     string `json:"type"`
	Labels   []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"labels"`
	Roles struct {
		Role []string `json:"role"`
	} `json:"roles"`
}

type upcloudAccountDetailsResponse struct {
	Account upcloudAccountDetails `json:"account"`
}

type upcloudAccountDetails struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

func (d *UpCloudDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	endpoint, err := url.JoinPath(d.baseURL, upcloudAccountPath, upcloudListPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build upcloud account list URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create upcloud account list request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute upcloud account list request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch upcloud accounts: unexpected status %d", httpResp.StatusCode)
	}

	var resp upcloudAccountListResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode upcloud account list response: %w", err)
	}

	records := make([]AccountRecord, 0, len(resp.Accounts.Account))
	unreadable := 0

	for _, a := range resp.Accounts.Account {
		// The username is the account's only stable identifier. Dropping a
		// blank row would hide an account the source still exposes and mark
		// it removed on the next review, so a malformed list fails the sync.
		username := strings.TrimSpace(a.Username)
		if username == "" {
			return nil, fmt.Errorf("cannot list upcloud accounts: account with an empty username")
		}

		// The review keys accounts on email plus external ID, so a detail
		// fetch that fails for a transient reason must abort rather than
		// yield a record with a blank email: that record would land under a
		// different key and read as one account removed and another added.
		details, err := d.fetchAccountDetails(ctx, username)
		if err != nil {
			return nil, fmt.Errorf("cannot list upcloud accounts: %w", err)
		}

		record := AccountRecord{
			FullName:    username,
			Roles:       upcloudRoles(a.Roles.Role),
			IsAdmin:     upcloudIsMainAccount(a.Type),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  username,
		}

		if details == nil {
			unreadable++
		} else {
			record.Email = strings.TrimSpace(details.Email)

			if name := strings.TrimSpace(details.FirstName + " " + details.LastName); name != "" {
				record.FullName = name
			}
		}

		records = append(records, record)
	}

	if unreadable > 0 {
		d.logger.WarnCtx(ctx, "upcloud accounts listed without readable details", log.Int("count", unreadable))
	}

	return records, nil
}

func (d *UpCloudDriver) fetchAccountDetails(ctx context.Context, username string) (*upcloudAccountDetails, error) {
	endpoint, err := url.JoinPath(d.baseURL, upcloudAccountPath, upcloudDetailsPath, url.PathEscape(username))
	if err != nil {
		return nil, fmt.Errorf("cannot build upcloud account details URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create upcloud account details request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute upcloud account details request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	// 403 ACCOUNT_FORBIDDEN (out of the token's reach) and 404 (gone) are
	// stable answers, not failures: the list call already proved the
	// credential, and both give the same blank email on every run, so the
	// account keeps a consistent key across campaigns.
	if httpResp.StatusCode == http.StatusForbidden || httpResp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch upcloud account details: unexpected status %d", httpResp.StatusCode)
	}

	var resp upcloudAccountDetailsResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode upcloud account details response: %w", err)
	}

	return &resp.Account, nil
}

// upcloudRoles copies the account's role list, mapping a missing/empty role
// list to an empty (non-nil) slice rather than nil.
func upcloudRoles(roles []string) []string {
	out := make([]string, 0, len(roles))
	out = append(out, roles...)

	return out
}

// upcloudIsMainAccount reports whether the account is the contract's primary
// account, which holds full administrative access. The docs call it "mymain"
// and the live API "main", so classify off "sub" — the only other kind, and
// the one both agree on.
func upcloudIsMainAccount(accountType string) bool {
	return !strings.EqualFold(strings.TrimSpace(accountType), "sub")
}

// upcloudNameResolver names the source after the username the API token
// belongs to, via GET /1.3/account. UpCloud exposes no organisation or
// workspace name, and account/list carries no marker for which of its rows
// the token authenticated as.
type upcloudNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

func NewUpCloudNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &upcloudNameResolver{httpClient: httpClient, baseURL: baseURL}
}

func (r *upcloudNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	endpoint, err := url.JoinPath(r.baseURL, upcloudAccountPath)
	if err != nil {
		return "", fmt.Errorf("cannot build upcloud account URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create upcloud account request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute upcloud account request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nameStatusError("upcloud account", httpResp.StatusCode)
	}

	var resp struct {
		Account struct {
			Username string `json:"username"`
		} `json:"account"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode upcloud account response: %w", err)
	}

	return resp.Account.Username, nil
}
