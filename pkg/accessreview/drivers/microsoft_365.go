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
	"io"
	"net/http"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// Microsoft365Driver fetches user accounts from a Microsoft 365 / Microsoft
// Entra ID tenant via the Microsoft Graph API using a pre-authenticated
// HTTP client (Bearer token).
type Microsoft365Driver struct {
	httpClient *http.Client
}

var _ Driver = (*Microsoft365Driver)(nil)

const (
	microsoft365GraphBaseURL    = "https://graph.microsoft.com/v1.0"
	microsoft365UsersSelect     = "id,userPrincipalName,mail,displayName,givenName,surname,accountEnabled,jobTitle,department,createdDateTime"
	microsoft365UsersPageSize   = 999
	microsoft365MaxPaginationOK = maxPaginationPages
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

func NewMicrosoft365Driver(httpClient *http.Client) *Microsoft365Driver {
	return &Microsoft365Driver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
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

	records := make([]AccountRecord, 0, len(users))
	for _, u := range users {
		email := u.Mail
		if email == "" {
			email = u.UserPrincipalName
		}

		userRoles := rolesByUser[u.ID]
		isAdmin := false
		for _, r := range userRoles {
			if adminRoleDisplayNames[r] {
				isAdmin = true
				break
			}
		}

		role := pickHighestRole(userRoles)
		if role == "" {
			role = "User"
		}

		active := u.AccountEnabled
		rec := AccountRecord{
			Email:       email,
			FullName:    u.DisplayName,
			Role:        role,
			JobTitle:    u.JobTitle,
			Active:      &active,
			IsAdmin:     isAdmin,
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessEntryAuthMethodSSO,
			AccountType: coredata.AccessEntryAccountTypeUser,
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

// pickHighestRole returns the most privileged admin role from the list,
// falling back to the first non-admin role when no admin role is present.
// Privilege order is hard-coded to Microsoft's well-known directory roles.
func pickHighestRole(roles []string) string {
	priority := []string{
		"Global Administrator",
		"Company Administrator",
		"Privileged Role Administrator",
		"Privileged Authentication Administrator",
		"Security Administrator",
		"Application Administrator",
		"Cloud Application Administrator",
		"User Administrator",
		"Conditional Access Administrator",
		"Compliance Administrator",
		"Authentication Administrator",
	}

	for _, p := range priority {
		for _, r := range roles {
			if r == p {
				return r
			}
		}
	}

	if len(roles) > 0 {
		return roles[0]
	}
	return ""
}

func (d *Microsoft365Driver) listUsers(ctx context.Context) ([]microsoft365User, error) {
	url := fmt.Sprintf(
		"%s/users?$select=%s&$top=%d",
		microsoft365GraphBaseURL,
		microsoft365UsersSelect,
		microsoft365UsersPageSize,
	)

	var all []microsoft365User
	for range microsoft365MaxPaginationOK {
		var page microsoft365UsersPage
		if err := d.fetchJSON(ctx, url, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Value...)
		if page.NextLink == "" {
			return all, nil
		}
		url = page.NextLink
	}

	return nil, fmt.Errorf("cannot list all microsoft 365 users: %w", ErrPaginationLimitReached)
}

func (d *Microsoft365Driver) listDirectoryRoles(ctx context.Context) ([]microsoft365DirectoryRole, error) {
	url := fmt.Sprintf("%s/directoryRoles", microsoft365GraphBaseURL)

	var all []microsoft365DirectoryRole
	for range microsoft365MaxPaginationOK {
		var page microsoft365RolesPage
		if err := d.fetchJSON(ctx, url, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Value...)
		if page.NextLink == "" {
			return all, nil
		}
		url = page.NextLink
	}

	return nil, fmt.Errorf("cannot list all microsoft 365 directory roles: %w", ErrPaginationLimitReached)
}

func (d *Microsoft365Driver) listRoleMembers(ctx context.Context, roleID string) ([]microsoft365RoleMember, error) {
	url := fmt.Sprintf("%s/directoryRoles/%s/members", microsoft365GraphBaseURL, roleID)

	var all []microsoft365RoleMember
	for range microsoft365MaxPaginationOK {
		var page microsoft365MembersPage
		if err := d.fetchJSON(ctx, url, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Value...)
		if page.NextLink == "" {
			return all, nil
		}
		url = page.NextLink
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
