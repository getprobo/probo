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
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	nukiAPIHost = "api.nuki.io"
	// nukiAccountUsersPath lists the account users of the Nuki Web account the
	// token belongs to — the people (and companies) the account holder has
	// invited and can grant smart lock authorizations to. It needs both the
	// `account` and `smartlock.auth` scopes.
	nukiAccountUsersPath = "/account/user"
	// nukiAccountUsersPageSize is the requested page size. Nuki documents an
	// undocumented server-side maximum ("If the value exceeds the maximum, then
	// the maximum value will be used"), so the pagination loop must not assume
	// a page of this size, only that a full page was not the last one.
	nukiAccountUsersPageSize = 100
	// nukiAccountUserTypeCompany is the AccountUser.type value marking an
	// account user that stands for a company rather than a person (0 .. user,
	// 1 .. company). Only caretaker accounts can create the company form.
	nukiAccountUserTypeCompany = 1
)

// NukiDriver lists the account users of a single Nuki Web account. A Nuki Web
// API token is bound to one account, so GET /account/user returns every account
// user of that account with no tenant selector. Pagination is offset/limit
// based.
//
// Nuki's access model has two layers: account users (this endpoint) are the
// identities that hold access, while the individual door grants live in
// per-device smart lock authorizations (/smartlock/auth). This driver reviews
// the identity layer, which is what an access review asks for — who can open
// the doors at all.
type NukiDriver struct {
	httpClient *http.Client
}

var _ Driver = (*NukiDriver)(nil)

// nukiAccountUser mirrors the documented AccountUser schema. There is no
// status, role, or last-login field: an account user exists or it does not, and
// what it may open is expressed by its smart lock authorizations.
type nukiAccountUser struct {
	AccountUserID int    `json:"accountUserId"`
	AccountID     int    `json:"accountId"`
	Type          int    `json:"type"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Language      string `json:"language"`
	CreationDate  string `json:"creationDate"`
	UpdateDate    string `json:"updateDate"`
}

func NewNukiDriver(httpClient *http.Client) *NukiDriver {
	return &NukiDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
	}
}

func (d *NukiDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	offset := 0

	for range maxPaginationPages {
		users, err := d.fetchAccountUsers(ctx, offset)
		if err != nil {
			return nil, err
		}

		// An empty page is the only reliable end-of-collection signal. Nuki
		// silently clamps `limit` to an undocumented maximum, so a page shorter
		// than nukiAccountUsersPageSize may well be a full page and stopping on
		// it would truncate the review. Advancing by the number of rows
		// actually returned keeps the walk correct under any clamp.
		if len(users) == 0 {
			return records, nil
		}

		for _, u := range users {
			record, err := nukiAccountRecord(u)
			if err != nil {
				return nil, err
			}

			records = append(records, record)
		}

		offset += len(users)
	}

	return nil, fmt.Errorf("cannot list all nuki account users: %w", ErrPaginationLimitReached)
}

func (d *NukiDriver) fetchAccountUsers(ctx context.Context, offset int) ([]nukiAccountUser, error) {
	q := url.Values{}
	q.Set("limit", strconv.Itoa(nukiAccountUsersPageSize))
	q.Set("offset", strconv.Itoa(offset))

	endpoint := url.URL{
		Scheme:   "https",
		Host:     nukiAPIHost,
		Path:     nukiAccountUsersPath,
		RawQuery: q.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create nuki account users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute nuki account users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch nuki account users: unexpected status %d", httpResp.StatusCode)
	}

	var users []nukiAccountUser
	if err := json.NewDecoder(httpResp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("cannot decode nuki account users response: %w", err)
	}

	return users, nil
}

// nukiAccountRecord maps one account user onto an AccountRecord.
//
// A row with neither an email nor a positive account user ID carries no stable
// identity, and the review keys entries on exactly that pair: emitting it would
// collide with every other identity-less row, so a malformed page fails the
// whole fetch rather than silently merging or dropping accounts.
func nukiAccountRecord(u nukiAccountUser) (AccountRecord, error) {
	email := strings.TrimSpace(u.Email)

	externalID := ""
	if u.AccountUserID > 0 {
		externalID = strconv.Itoa(u.AccountUserID)
	}

	if email == "" && externalID == "" {
		return AccountRecord{}, fmt.Errorf("cannot map nuki account user: row has neither an email nor an account user id")
	}

	return AccountRecord{
		Email:    email,
		FullName: nukiFullName(u, email, externalID),
		// GET /account/user exposes no role or permission field: an account
		// user's rights are the smart lock authorizations granted to it, not an
		// account-level role. Nuki's administrative rights bitmask lives on
		// accounts and sub-accounts (/account/sub), a different population.
		Roles: nil,
		// Same reason: no account user is an administrator of the Nuki Web
		// account. Administrators are the account and its sub-accounts.
		IsAdmin: false,
		// A company account user is an organisation holding a key rather than a
		// person, which is what SERVICE_ACCOUNT marks for reviewers.
		AccountType: nukiAccountType(u.Type),
		// Nuki has no deactivated state for an account user — it is deleted
		// instead — so there is no active/inactive signal to report. The
		// `operationId` field means "locked by an in-flight operation", which
		// is transient and says nothing about access.
		Active: nil,
		// One-time password enrolment is a property of the authenticated Nuki
		// account (Account.Config.otpEnabledDate), not of the account users it
		// invited, so MFA is unknown per account user.
		MFAStatus: coredata.MFAStatusUnknown,
		// An account user is reached by email invitation and may or may not
		// have redeemed it into a password-backed Nuki account; the endpoint
		// does not say which.
		AuthMethod: coredata.AccessReviewEntryAuthMethodUnknown,
		CreatedAt:  parseRFC3339Ptr(u.CreationDate),
		// No last-login signal: `updateDate` tracks edits to the account user
		// record, and door activity lives in the smart lock logs.
		LastLogin:  nil,
		ExternalID: externalID,
	}, nil
}

// nukiFullName returns the account user's display name, falling back to the
// email and then to the account user ID when Nuki carries no name.
func nukiFullName(u nukiAccountUser, email, externalID string) string {
	if name := strings.TrimSpace(u.Name); name != "" {
		return name
	}

	if email != "" {
		return email
	}

	return externalID
}

// nukiAccountType maps AccountUser.type onto the review's account taxonomy:
// a company account user is a non-human key holder, anything else a person.
func nukiAccountType(accountUserType int) coredata.AccessReviewEntryAccountType {
	if accountUserType == nukiAccountUserTypeCompany {
		return coredata.AccessReviewEntryAccountTypeServiceAccount
	}

	return coredata.AccessReviewEntryAccountTypeUser
}

// nukiNameResolver names the source after the Nuki Web account the API token
// belongs to, via GET /account. Nuki has no organisation or workspace above the
// account, and a token is bound to exactly one account, so "the account's own
// name" is the only instance label available.
type nukiNameResolver struct {
	httpClient *http.Client
}

func NewNukiNameResolver(httpClient *http.Client) NameResolver {
	return &nukiNameResolver{httpClient: httpClient}
}

func (r *nukiNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	endpoint := url.URL{Scheme: "https", Host: nukiAPIHost, Path: "/account"}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return "", fmt.Errorf("cannot create nuki account request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute nuki account request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	// Best-effort: a non-2xx (revoked token, or a token without the `account`
	// scope) must not make the source-name worker retry forever. Give up
	// gracefully and keep the generic source name; a dead token surfaces on the
	// next ListAccounts.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode nuki account response: %w", err)
	}

	// Never fall back to the account's email: the source name is displayed to
	// every reviewer, and an empty result simply keeps the generic name.
	return resp.Name, nil
}
