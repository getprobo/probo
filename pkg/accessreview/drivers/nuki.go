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
	"sort"
	"strconv"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	nukiAccountSegment   = "account"
	nukiUserSegment      = "user"
	nukiSmartlockSegment = "smartlock"
	nukiAuthSegment      = "auth"
	// Nuki clamps limit to an undocumented maximum, so pagination must not
	// treat a short page as the end of the collection.
	nukiAccountUsersPageSize   = 100
	nukiAuthsPageSize          = 100
	nukiAccountUserTypeCompany = 1
	nukiAuthTypeApp            = 0
	nukiAuthTypeBridge         = 1
	nukiAuthTypeFob            = 2
	nukiAuthTypeKeypad         = 3
	nukiAuthTypeKeypadCode     = 13
	nukiAuthTypeZKey           = 14
	nukiAuthTypeVirtual        = 15
	nukiRemoteAccessRole       = "Remote access"
)

// NukiDriver lists account users and smartlock authorizations for one Nuki Web
// account. Authorizations without an accountUserId (keypad codes, fobs, …) are
// emitted as service accounts.
type NukiDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*NukiDriver)(nil)

type (
	nukiAccountUser struct {
		AccountUserID int    `json:"accountUserId"`
		AccountID     int    `json:"accountId"`
		Type          int    `json:"type"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		Language      string `json:"language"`
		CreationDate  string `json:"creationDate"`
		UpdateDate    string `json:"updateDate"`
	}

	// Enabled and RemoteAllowed are pointers: Nuki omits defaults, and a bare
	// bool would treat "missing" as false.
	nukiSmartlockAuth struct {
		ID               string `json:"id"`
		SmartlockID      int64  `json:"smartlockId"`
		AccountUserID    int    `json:"accountUserId"`
		AuthID           int    `json:"authId"`
		Type             int    `json:"type"`
		Name             string `json:"name"`
		Enabled          *bool  `json:"enabled"`
		RemoteAllowed    *bool  `json:"remoteAllowed"`
		AllowedFromDate  string `json:"allowedFromDate"`
		AllowedUntilDate string `json:"allowedUntilDate"`
		LastActiveDate   string `json:"lastActiveDate"`
		CreationDate     string `json:"creationDate"`
	}

	nukiSmartlock struct {
		SmartlockID int64  `json:"smartlockId"`
		Name        string `json:"name"`
	}

	nukiAuthPage struct {
		Results    []nukiSmartlockAuth `json:"results"`
		Pagination struct {
			TotalPages  int `json:"totalPages"`
			CurrentPage int `json:"currentPage"`
		} `json:"pagination"`
	}
)

func NewNukiDriver(httpClient *http.Client, baseURL string) *NukiDriver {
	return &NukiDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		baseURL: baseURL,
	}
}

func (d *NukiDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	// Best-effort: needs smartlock.readOnly, which existing tokens may lack.
	lockNames, _ := d.fetchSmartlockNames(ctx)

	auths, err := d.fetchAllAuths(ctx)
	if err != nil {
		return nil, err
	}

	authsByUser := make(map[int][]nukiSmartlockAuth)
	var orphanAuths []nukiSmartlockAuth

	for _, auth := range auths {
		if auth.AccountUserID > 0 {
			authsByUser[auth.AccountUserID] = append(authsByUser[auth.AccountUserID], auth)

			continue
		}

		orphanAuths = append(orphanAuths, auth)
	}

	var records []AccountRecord

	offset := 0

	for range maxPaginationPages {
		users, err := d.fetchAccountUsers(ctx, offset)
		if err != nil {
			return nil, err
		}

		if len(users) == 0 {
			for _, auth := range orphanAuths {
				record, err := nukiServiceAccountRecord(auth, lockNames)
				if err != nil {
					return nil, err
				}

				records = append(records, record)
			}

			return records, nil
		}

		for _, u := range users {
			record, err := nukiAccountRecord(u, authsByUser[u.AccountUserID], lockNames)
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
	endpoint, err := url.JoinPath(d.baseURL, nukiAccountSegment, nukiUserSegment)
	if err != nil {
		return nil, fmt.Errorf("cannot build nuki account users URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create nuki account users request: %w", err)
	}

	q := req.URL.Query()
	q.Set("limit", strconv.Itoa(nukiAccountUsersPageSize))
	q.Set("offset", strconv.Itoa(offset))
	req.URL.RawQuery = q.Encode()

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

func (d *NukiDriver) fetchAllAuths(ctx context.Context) ([]nukiSmartlockAuth, error) {
	var (
		auths []nukiSmartlockAuth
		page  int
	)

	for range maxPaginationPages {
		resp, err := d.fetchAuthPage(ctx, page)
		if err != nil {
			return nil, err
		}

		if len(resp.Results) == 0 {
			return auths, nil
		}

		auths = append(auths, resp.Results...)

		if resp.Pagination.TotalPages > 0 && page+1 >= resp.Pagination.TotalPages {
			return auths, nil
		}

		if resp.Pagination.TotalPages <= 0 && len(resp.Results) < nukiAuthsPageSize {
			return auths, nil
		}

		page++
	}

	return nil, fmt.Errorf("cannot list all nuki smartlock authorizations: %w", ErrPaginationLimitReached)
}

func (d *NukiDriver) fetchAuthPage(ctx context.Context, page int) (*nukiAuthPage, error) {
	endpoint, err := url.JoinPath(d.baseURL, nukiSmartlockSegment, nukiAuthSegment, "paged")
	if err != nil {
		return nil, fmt.Errorf("cannot build nuki smartlock auth URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create nuki smartlock auth request: %w", err)
	}

	q := req.URL.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("size", strconv.Itoa(nukiAuthsPageSize))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute nuki smartlock auth request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch nuki smartlock authorizations: unexpected status %d", httpResp.StatusCode)
	}

	var resp nukiAuthPage
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode nuki smartlock auth response: %w", err)
	}

	return &resp, nil
}

func (d *NukiDriver) fetchSmartlockNames(ctx context.Context) (map[int64]string, error) {
	endpoint, err := url.JoinPath(d.baseURL, nukiSmartlockSegment)
	if err != nil {
		return nil, fmt.Errorf("cannot build nuki smartlocks URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create nuki smartlocks request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute nuki smartlocks request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch nuki smartlocks: unexpected status %d", httpResp.StatusCode)
	}

	var locks []nukiSmartlock
	if err := json.NewDecoder(httpResp.Body).Decode(&locks); err != nil {
		return nil, fmt.Errorf("cannot decode nuki smartlocks response: %w", err)
	}

	names := make(map[int64]string, len(locks))
	for _, lock := range locks {
		if name := strings.TrimSpace(lock.Name); name != "" {
			names[lock.SmartlockID] = name
		}
	}

	return names, nil
}

func nukiAccountRecord(
	u nukiAccountUser,
	auths []nukiSmartlockAuth,
	lockNames map[int64]string,
) (AccountRecord, error) {
	email := strings.TrimSpace(u.Email)

	externalID := ""
	if u.AccountUserID > 0 {
		externalID = strconv.Itoa(u.AccountUserID)
	}

	if email == "" && externalID == "" {
		return AccountRecord{}, fmt.Errorf("cannot map nuki account user: row has neither an email nor an account user id")
	}

	return AccountRecord{
		Email:       email,
		FullName:    nukiFullName(u, email, externalID),
		Roles:       nukiRoles(auths, lockNames),
		IsAdmin:     false,
		AccountType: nukiAccountType(u.Type),
		Active:      nukiActiveFromAuths(auths, time.Now()),
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		CreatedAt:   parseRFC3339Ptr(u.CreationDate),
		LastLogin:   nukiLastActive(auths),
		ExternalID:  externalID,
	}, nil
}

func nukiServiceAccountRecord(
	auth nukiSmartlockAuth,
	lockNames map[int64]string,
) (AccountRecord, error) {
	externalID := strings.TrimSpace(auth.ID)
	if externalID == "" && auth.AuthID > 0 {
		externalID = strconv.Itoa(auth.AuthID)
	}

	if externalID == "" {
		return AccountRecord{}, fmt.Errorf(
			"cannot map nuki smartlock authorization: row has neither an id nor an auth id (smartlock %d)",
			auth.SmartlockID,
		)
	}

	fullName := strings.TrimSpace(auth.Name)
	if fullName == "" {
		fullName = nukiAuthTypeLabel(auth.Type)
	}

	return AccountRecord{
		FullName:    fullName,
		Roles:       nukiRoles([]nukiSmartlockAuth{auth}, lockNames),
		IsAdmin:     false,
		AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
		Active:      nukiAuthActive(auth, time.Now()),
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodServiceAccount,
		CreatedAt:   parseRFC3339Ptr(auth.CreationDate),
		LastLogin:   parseRFC3339Ptr(auth.LastActiveDate),
		ExternalID:  externalID,
	}, nil
}

func nukiFullName(u nukiAccountUser, email, externalID string) string {
	if name := strings.TrimSpace(u.Name); name != "" {
		return name
	}

	if email != "" {
		return email
	}

	return externalID
}

func nukiAccountType(accountUserType int) coredata.AccessReviewEntryAccountType {
	if accountUserType == nukiAccountUserTypeCompany {
		return coredata.AccessReviewEntryAccountTypeServiceAccount
	}

	return coredata.AccessReviewEntryAccountTypeUser
}

func nukiRoles(auths []nukiSmartlockAuth, lockNames map[int64]string) []string {
	if len(auths) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(auths)+1)
	remote := false

	for _, auth := range auths {
		role := nukiLockRole(auth.SmartlockID, lockNames)
		if role == "" {
			role = nukiAuthTypeLabel(auth.Type)
		}

		if role != "" {
			seen[role] = struct{}{}
		}

		if auth.RemoteAllowed != nil && *auth.RemoteAllowed {
			remote = true
		}
	}

	if remote {
		seen[nukiRemoteAccessRole] = struct{}{}
	}

	roles := make([]string, 0, len(seen))
	for role := range seen {
		roles = append(roles, role)
	}

	sort.Strings(roles)

	return roles
}

func nukiLockRole(smartlockID int64, lockNames map[int64]string) string {
	if smartlockID <= 0 {
		return ""
	}

	if name := strings.TrimSpace(lockNames[smartlockID]); name != "" {
		return name
	}

	return "Smartlock " + strconv.FormatInt(smartlockID, 10)
}

func nukiAuthTypeLabel(authType int) string {
	switch authType {
	case nukiAuthTypeApp:
		return "App"
	case nukiAuthTypeBridge:
		return "Bridge"
	case nukiAuthTypeFob:
		return "Fob"
	case nukiAuthTypeKeypad:
		return "Keypad"
	case nukiAuthTypeKeypadCode:
		return "Keypad code"
	case nukiAuthTypeZKey:
		return "Z-Key"
	case nukiAuthTypeVirtual:
		return "Virtual"
	default:
		return "Authorization"
	}
}

// nukiActiveFromAuths returns nil when there are no grants (no signal on the
// account-user row alone); otherwise whether any grant is currently usable.
func nukiActiveFromAuths(auths []nukiSmartlockAuth, now time.Time) *bool {
	if len(auths) == 0 {
		return nil
	}

	for _, auth := range auths {
		if active := nukiAuthActive(auth, now); active != nil && *active {
			return new(true)
		}
	}

	return new(false)
}

func nukiAuthActive(auth nukiSmartlockAuth, now time.Time) *bool {
	if auth.Enabled != nil && !*auth.Enabled {
		return new(false)
	}

	if until := parseRFC3339Ptr(auth.AllowedUntilDate); until != nil && now.After(*until) {
		return new(false)
	}

	return new(true)
}

func nukiLastActive(auths []nukiSmartlockAuth) *time.Time {
	var latest *time.Time

	for _, auth := range auths {
		t := parseRFC3339Ptr(auth.LastActiveDate)
		if t == nil {
			continue
		}

		if latest == nil || t.After(*latest) {
			latest = t
		}
	}

	return latest
}

type nukiNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

func NewNukiNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &nukiNameResolver{httpClient: httpClient, baseURL: baseURL}
}

func (r *nukiNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	endpoint, err := url.JoinPath(r.baseURL, nukiAccountSegment)
	if err != nil {
		return "", fmt.Errorf("cannot build nuki account URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create nuki account request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute nuki account request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	// Soft-fail: keep the generic source name rather than retrying forever.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode nuki account response: %w", err)
	}

	// Do not fall back to the account email; it is shown to every reviewer.
	return resp.Name, nil
}
