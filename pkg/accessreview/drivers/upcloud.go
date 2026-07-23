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

const upcloudAccountListURL = "https://api.upcloud.com/1.3/account/list"

// UpCloudDriver lists the main account and its sub-accounts via UpCloud's
// account/list endpoint, then enriches each with account/details/{username}
// (email, first/last name), using a pre-authenticated HTTP client (Bearer API
// token) attached by the connection transport.
//
// Notes on data quality:
//   - account/details has no explicit account-status field, so Active is
//     left nil (no signal).
//   - Neither endpoint exposes per-account MFA status, so MFAStatus is left
//     Unknown.
//   - If the details fetch for an account fails, the account is still
//     returned (per Driver contract, no account may be dropped) with just
//     the list fields; Email stays blank and FullName falls back to the
//     username.
type UpCloudDriver struct {
	httpClient *http.Client
	logger     *log.Logger
}

var _ Driver = (*UpCloudDriver)(nil)

func NewUpCloudDriver(httpClient *http.Client, logger *log.Logger) *UpCloudDriver {
	return &UpCloudDriver{
		httpClient: httpClient,
		logger:     logger,
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, upcloudAccountListURL, nil)
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

		username := strings.TrimSpace(a.Username)
		if username == "" {
			return nil, fmt.Errorf("upcloud returned an account with an empty username")
		}

		fullName := username

		details, err := d.fetchAccountDetails(ctx, username)
		if err != nil {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("cannot list upcloud accounts: %w", ctx.Err())
			}

			d.logger.WarnCtx(ctx, "cannot fetch upcloud account details, using list fields only", log.Error(err))
		} else {
			if name := strings.TrimSpace(details.FirstName + " " + details.LastName); name != "" {
				fullName = name
			}
		}

		record := AccountRecord{
			FullName:    fullName,
			Roles:       upcloudRoles(a.Roles.Role),
			IsAdmin:     upcloudIsMainAccount(a.Type),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  username,
		}

		if details != nil {
			record.Email = strings.TrimSpace(details.Email)
		}

		records = append(records, record)
	}

	return records, nil
}

func (d *UpCloudDriver) fetchAccountDetails(ctx context.Context, username string) (*upcloudAccountDetails, error) {
	endpoint, err := url.JoinPath("https://api.upcloud.com", "1.3", "account", "details", url.PathEscape(username))
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

// upcloudIsMainAccount reports whether the account is the primary account on
// the contract ("mymain"), as opposed to a "sub" account. The main account
// holds full administrative access; sub-accounts are scoped by their roles.
func upcloudIsMainAccount(accountType string) bool {
	return strings.EqualFold(strings.TrimSpace(accountType), "mymain")
}
