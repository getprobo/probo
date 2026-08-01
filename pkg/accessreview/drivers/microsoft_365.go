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
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
)

// Microsoft365Driver fetches user accounts from a Microsoft 365 / Microsoft
// Entra ID tenant via the Microsoft Graph API using a pre-authenticated
// HTTP client (Bearer token).
type Microsoft365Driver struct {
	httpClient *http.Client
	logger     *log.Logger
}

var _ Driver = (*Microsoft365Driver)(nil)

const (
	microsoft365GraphBaseURL         = "https://graph.microsoft.com/v1.0"
	microsoft365UsersSelect          = "id,userPrincipalName,mail,displayName,givenName,surname,accountEnabled,jobTitle,department,createdDateTime"
	microsoft365UserTypeMemberFilter = "userType eq 'Member'"
	microsoft365UsersPageSize        = 999
	microsoft365MaxPaginationOK      = maxPaginationPages
)

// adminRoleDisplayNames lists the directory role display names that the
// driver treats as administrative. Microsoft splits administration across
// many roles; matching by display name keeps the driver readable.
var adminRoleDisplayNames = map[string]bool{
	"Global Administrator":                    true,
	"Company Administrator":                   true,
	"Privileged Role Administrator":           true,
	"Privileged Authentication Administrator": true,
	"Security Administrator":                  true,
	"User Administrator":                      true,
	"Conditional Access Administrator":        true,
	"Compliance Administrator":                true,
	"Application Administrator":               true,
	"Cloud Application Administrator":         true,
	"Authentication Administrator":            true,
}

func NewMicrosoft365Driver(httpClient *http.Client, logger *log.Logger) *Microsoft365Driver {
	return &Microsoft365Driver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		logger: logger,
	}
}

type microsoft365User struct {
	ID                string `json:"id"`
	UserPrincipalName string `json:"userPrincipalName"`
	Mail              string `json:"mail"`
	DisplayName       string `json:"displayName"`
	GivenName         string `json:"givenName"`
	Surname           string `json:"surname"`
	AccountEnabled    bool   `json:"accountEnabled"`
	JobTitle          string `json:"jobTitle"`
	Department        string `json:"department"`
	CreatedDateTime   string `json:"createdDateTime"`
}

type microsoft365UsersPage struct {
	Value    []microsoft365User `json:"value"`
	NextLink string             `json:"@odata.nextLink"`
}

type microsoft365DirectoryRole struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type microsoft365RolesPage struct {
	Value    []microsoft365DirectoryRole `json:"value"`
	NextLink string                      `json:"@odata.nextLink"`
}

type microsoft365RoleMember struct {
	ID          string `json:"id"`
	ODataType   string `json:"@odata.type"`
	DisplayName string `json:"displayName"`
}

type microsoft365MembersPage struct {
	Value    []microsoft365RoleMember `json:"value"`
	NextLink string                   `json:"@odata.nextLink"`
}

type microsoft365UserRegistrationDetails struct {
	ID                string `json:"id"`
	UserPrincipalName string `json:"userPrincipalName"`
	IsMFARegistered   *bool  `json:"isMfaRegistered"`
}

type microsoft365UserRegistrationDetailsPage struct {
	Value    []microsoft365UserRegistrationDetails `json:"value"`
	NextLink string                                `json:"@odata.nextLink"`
}

func (d *Microsoft365Driver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	roles, err := d.listDirectoryRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot list directory roles: %w", err)
	}

	rolesByUser := make(map[string][]string)

	for _, role := range roles {
		members, err := d.listRoleMembers(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("cannot list members of role %q: %w", role.DisplayName, err)
		}

		for _, m := range members {
			if m.ODataType != "" && m.ODataType != "#microsoft.graph.user" {
				continue
			}

			rolesByUser[m.ID] = append(rolesByUser[m.ID], role.DisplayName)
		}
	}

	users, err := d.listUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot list users: %w", err)
	}

	// MFA registration details require AuditLog.Read.All. A failure here
	// must not abort the account fetch — leave MFA unknown so the campaign
	// can still proceed. Context cancel/deadline still fail the fetch so a
	// timed-out source does not commit an incomplete result as success.
	mfaStatuses, err := d.listMFAStatuses(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("cannot list MFA statuses: %w", err)
		}

		d.logger.ErrorCtx(ctx, "cannot list microsoft 365 MFA statuses, reporting MFA unknown", log.Error(err))

		mfaStatuses = map[string]coredata.MFAStatus{}
	}

	records := make([]AccountRecord, 0, len(users))
	for _, u := range users {
		email := u.Mail
		if email == "" {
			email = u.UserPrincipalName
		}

		// Skip accounts without any usable identifier so the access
		// review only lists real members, matching the SCIM bridge.
		if email == "" {
			continue
		}

		userRoles := rolesByUser[u.ID]
		isAdmin := false

		for _, r := range userRoles {
			if adminRoleDisplayNames[r] {
				isAdmin = true
				break
			}
		}

		roles := userRoles
		if len(roles) == 0 {
			roles = []string{"User"}
		}

		active := u.AccountEnabled
		rec := AccountRecord{
			Email:       email,
			FullName:    u.DisplayName,
			Roles:       roles,
			JobTitle:    u.JobTitle,
			Active:      &active,
			IsAdmin:     isAdmin,
			MFAStatus:   microsoft365MFAStatus(u, mfaStatuses),
			AuthMethod:  coredata.AccessReviewEntryAuthMethodSSO,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  u.ID,
		}

		if u.CreatedDateTime != "" {
			if t, err := time.Parse(time.RFC3339, u.CreatedDateTime); err == nil {
				rec.CreatedAt = &t
			}
		}

		records = append(records, rec)
	}

	return records, nil
}

func microsoft365MFAStatus(u microsoft365User, statuses map[string]coredata.MFAStatus) coredata.MFAStatus {
	if status, ok := statuses[u.ID]; ok {
		return status
	}

	if status, ok := statuses[strings.ToLower(u.UserPrincipalName)]; ok {
		return status
	}

	return coredata.MFAStatusUnknown
}

func microsoft365RegistrationMFAStatus(
	details microsoft365UserRegistrationDetails,
) coredata.MFAStatus {
	if details.IsMFARegistered == nil {
		return coredata.MFAStatusUnknown
	}

	if *details.IsMFARegistered {
		return coredata.MFAStatusEnabled
	}

	return coredata.MFAStatusDisabled
}

func (d *Microsoft365Driver) listUsers(ctx context.Context) ([]microsoft365User, error) {
	pageURL, err := buildMicrosoft365UsersURL()
	if err != nil {
		return nil, err
	}

	var all []microsoft365User

	for range microsoft365MaxPaginationOK {
		var page microsoft365UsersPage
		if err := d.fetchJSON(ctx, pageURL, &page); err != nil {
			return nil, err
		}

		all = append(all, page.Value...)
		if page.NextLink == "" {
			return all, nil
		}

		pageURL = page.NextLink
	}

	return nil, fmt.Errorf("cannot list all microsoft 365 users: %w", ErrPaginationLimitReached)
}

func buildMicrosoft365UsersURL() (string, error) {
	u, err := url.Parse(microsoft365GraphBaseURL + "/users")
	if err != nil {
		return "", fmt.Errorf("cannot parse graph users URL: %w", err)
	}

	q := u.Query()
	q.Set("$select", microsoft365UsersSelect)
	q.Set("$top", strconv.Itoa(microsoft365UsersPageSize))
	q.Set("$filter", microsoft365UserTypeMemberFilter)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (d *Microsoft365Driver) listMFAStatuses(ctx context.Context) (map[string]coredata.MFAStatus, error) {
	pageURL, err := buildMicrosoft365UserRegistrationDetailsURL()
	if err != nil {
		return nil, err
	}

	statuses := make(map[string]coredata.MFAStatus)

	for range microsoft365MaxPaginationOK {
		var page microsoft365UserRegistrationDetailsPage
		if err := d.fetchJSON(ctx, pageURL, &page); err != nil {
			return nil, err
		}

		for _, details := range page.Value {
			status := microsoft365RegistrationMFAStatus(details)
			if details.ID != "" {
				statuses[details.ID] = status
			}

			if details.UserPrincipalName != "" {
				statuses[strings.ToLower(details.UserPrincipalName)] = status
			}
		}

		if page.NextLink == "" {
			return statuses, nil
		}

		pageURL = page.NextLink
	}

	return nil, fmt.Errorf("cannot list all microsoft 365 MFA statuses: %w", ErrPaginationLimitReached)
}

func buildMicrosoft365UserRegistrationDetailsURL() (string, error) {
	endpoint, err := url.JoinPath(
		microsoft365GraphBaseURL,
		"reports",
		"authenticationMethods",
		"userRegistrationDetails",
	)
	if err != nil {
		return "", fmt.Errorf("cannot build graph user registration details URL: %w", err)
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("cannot parse graph user registration details URL: %w", err)
	}

	return u.String(), nil
}

func (d *Microsoft365Driver) listDirectoryRoles(ctx context.Context) ([]microsoft365DirectoryRole, error) {
	endpoint, err := url.JoinPath(microsoft365GraphBaseURL, "directoryRoles")
	if err != nil {
		return nil, fmt.Errorf("cannot build graph directory roles URL: %w", err)
	}

	var all []microsoft365DirectoryRole

	for range microsoft365MaxPaginationOK {
		var page microsoft365RolesPage
		if err := d.fetchJSON(ctx, endpoint, &page); err != nil {
			return nil, err
		}

		all = append(all, page.Value...)
		if page.NextLink == "" {
			return all, nil
		}

		endpoint = page.NextLink
	}

	return nil, fmt.Errorf("cannot list all microsoft 365 directory roles: %w", ErrPaginationLimitReached)
}

func (d *Microsoft365Driver) listRoleMembers(ctx context.Context, roleID string) ([]microsoft365RoleMember, error) {
	endpoint, err := url.JoinPath(microsoft365GraphBaseURL, "directoryRoles", url.PathEscape(roleID), "members")
	if err != nil {
		return nil, fmt.Errorf("cannot build graph role members URL: %w", err)
	}

	var all []microsoft365RoleMember

	for range microsoft365MaxPaginationOK {
		var page microsoft365MembersPage
		if err := d.fetchJSON(ctx, endpoint, &page); err != nil {
			return nil, err
		}

		all = append(all, page.Value...)
		if page.NextLink == "" {
			return all, nil
		}

		endpoint = page.NextLink
	}

	return nil, fmt.Errorf("cannot list all members of role %q: %w", roleID, ErrPaginationLimitReached)
}

func (d *Microsoft365Driver) fetchJSON(ctx context.Context, url string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("cannot create graph request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot execute graph request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("microsoft graph error: status %d, body: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("cannot decode graph response: %w", err)
	}

	return nil
}

// microsoft365NameResolver resolves the Microsoft 365 tenant display name
// via the Microsoft Graph organization endpoint.
type microsoft365NameResolver struct {
	httpClient *http.Client
}

func NewMicrosoft365NameResolver(httpClient *http.Client) NameResolver {
	return &microsoft365NameResolver{httpClient: httpClient}
}

func (r *microsoft365NameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	msURL, err := url.Parse("https://graph.microsoft.com/v1.0/organization")
	if err != nil {
		return "", fmt.Errorf("cannot parse microsoft 365 organization URL: %w", err)
	}

	q := msURL.Query()
	q.Set("$select", "displayName,verifiedDomains")
	msURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, msURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("cannot create microsoft 365 organization request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute microsoft 365 organization request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nameStatusError("microsoft 365 organization", httpResp.StatusCode)
	}

	var resp struct {
		Value []struct {
			DisplayName     string `json:"displayName"`
			VerifiedDomains []struct {
				Name      string `json:"name"`
				IsDefault bool   `json:"isDefault"`
			} `json:"verifiedDomains"`
		} `json:"value"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode microsoft 365 organization response: %w", err)
	}

	if len(resp.Value) == 0 {
		return "", nil
	}

	org := resp.Value[0]
	if org.DisplayName != "" {
		return org.DisplayName, nil
	}

	for _, d := range org.VerifiedDomains {
		if d.IsDefault {
			return d.Name, nil
		}
	}

	if len(org.VerifiedDomains) > 0 {
		return org.VerifiedDomains[0].Name, nil
	}

	return "", nil
}
