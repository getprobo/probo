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

// HubSpotDriver fetches account users from HubSpot via OAuth2-authenticated
// REST requests.
//
// The Settings Users API (`/settings/v3/users`) exposes roles and super-admin
// flags but has no active/archived signal. Deactivated HubSpot users are
// owners with `archived=true` on `/crm/v3/owners`; this driver merges both
// APIs so inactive accounts are returned with Active=false instead of being
// misclassified as active.
type HubSpotDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*HubSpotDriver)(nil)

type (
	hubspotRolesResponse struct {
		Results []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"results"`
	}

	hubspotUser struct {
		ID            string   `json:"id"`
		Email         string   `json:"email"`
		FirstName     string   `json:"firstName"`
		LastName      string   `json:"lastName"`
		RoleID        string   `json:"roleId"`
		RoleIDs       []string `json:"roleIds"`
		PrimaryTeamID string   `json:"primaryTeamId"`
		SuperAdmin    bool     `json:"superAdmin"`
	}

	hubspotUsersResponse struct {
		Results []hubspotUser `json:"results"`
		Paging  *struct {
			Next *struct {
				After string `json:"after"`
			} `json:"next"`
		} `json:"paging"`
	}

	hubspotOwner struct {
		ID                      string `json:"id"`
		Email                   string `json:"email"`
		FirstName               string `json:"firstName"`
		LastName                string `json:"lastName"`
		Type                    string `json:"type"`
		Archived                bool   `json:"archived"`
		UserID                  *int64 `json:"userId"`
		UserIDIncludingInactive *int64 `json:"userIdIncludingInactive"`
	}

	hubspotOwnersResponse struct {
		Results []hubspotOwner `json:"results"`
		Paging  *struct {
			Next *struct {
				After string `json:"after"`
			} `json:"next"`
		} `json:"paging"`
	}

	hubspotSeenKeys struct {
		userIDs map[string]struct{}
		emails  map[string]struct{}
	}
)

const (
	hubspotUsersPath       = "/settings/v3/users"
	hubspotRolesPath       = "/settings/v3/users/roles"
	hubspotAccountInfoPath = "/account-info/v3/details"
	hubspotOwnersPath      = "/crm/v3/owners"
)

func NewHubSpotDriver(httpClient *http.Client, baseURL string) *HubSpotDriver {
	return &HubSpotDriver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (d *HubSpotDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	roleMap, _ := d.fetchRoles(ctx)

	users, err := d.fetchAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	// Owners archived=true is HubSpot's authoritative deactivated-user list.
	// Settings Users has no status field; without this call every account
	// would be stored as active via COALESCE(@active, TRUE) on upsert.
	archivedOwners, err := d.fetchAllOwners(ctx, true)
	if err != nil {
		return nil, err
	}

	archivedByUserID, archivedByEmail := hubspotIndexArchivedOwners(archivedOwners)

	records := make([]AccountRecord, 0, len(users)+len(archivedOwners))
	seen := newHubspotSeenKeys(len(users) + len(archivedOwners))

	for _, u := range users {
		owner, inactive := hubspotLookupArchived(u.ID, u.Email, archivedByUserID, archivedByEmail)

		fullName := hubspotDisplayName(u.FirstName, u.LastName, u.Email)

		record := AccountRecord{
			Email:       u.Email,
			FullName:    fullName,
			Roles:       hubspotRoles(u, roleMap),
			Active:      new(!inactive),
			IsAdmin:     u.SuperAdmin,
			ExternalID:  u.ID,
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
		}

		if record.Email == "" && record.ExternalID == "" {
			continue
		}

		records = append(records, record)
		seen.mark(u.ID, u.Email)
		// If we matched an archived owner by email only, also mark its
		// settings user ID so the archived-only pass cannot emit a duplicate.
		if inactive {
			seen.mark(hubspotOwnerUserID(owner), owner.Email)
		}
	}

	// Settings Users does not always list deactivated accounts. Surface
	// archived owners that never appeared above so reviews still see them.
	for _, owner := range archivedOwners {
		if !hubspotOwnerIsPerson(owner) {
			continue
		}

		userID := hubspotOwnerUserID(owner)
		if seen.has(userID, owner.Email) {
			continue
		}

		fullName := hubspotDisplayName(owner.FirstName, owner.LastName, owner.Email)
		externalID := userID
		if externalID == "" {
			externalID = owner.ID
		}

		record := AccountRecord{
			Email:       owner.Email,
			FullName:    fullName,
			Roles:       []string{"User"},
			Active:      new(false),
			IsAdmin:     false,
			ExternalID:  externalID,
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
		}

		if record.Email == "" && record.ExternalID == "" {
			continue
		}

		records = append(records, record)
		seen.mark(userID, owner.Email)
	}

	return records, nil
}

func (d *HubSpotDriver) fetchAllUsers(ctx context.Context) ([]hubspotUser, error) {
	var (
		users []hubspotUser
		after string
	)

	for range maxPaginationPages {
		resp, err := d.fetchUsers(ctx, after)
		if err != nil {
			return nil, err
		}

		users = append(users, resp.Results...)

		if resp.Paging == nil || resp.Paging.Next == nil || resp.Paging.Next.After == "" {
			return users, nil
		}

		after = resp.Paging.Next.After
	}

	return nil, fmt.Errorf("cannot list all hubspot users: %w", ErrPaginationLimitReached)
}

func (d *HubSpotDriver) fetchAllOwners(ctx context.Context, archived bool) ([]hubspotOwner, error) {
	var (
		owners []hubspotOwner
		after  string
	)

	for range maxPaginationPages {
		resp, err := d.fetchOwners(ctx, archived, after)
		if err != nil {
			return nil, err
		}

		owners = append(owners, resp.Results...)

		if resp.Paging == nil || resp.Paging.Next == nil || resp.Paging.Next.After == "" {
			return owners, nil
		}

		after = resp.Paging.Next.After
	}

	return nil, fmt.Errorf("cannot list all hubspot owners: %w", ErrPaginationLimitReached)
}

func (d *HubSpotDriver) fetchUsers(ctx context.Context, after string) (*hubspotUsersResponse, error) {
	endpoint, err := url.JoinPath(d.baseURL, hubspotUsersPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build hubspot users URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create hubspot users request: %w", err)
	}

	q := req.URL.Query()
	q.Set("limit", "100")

	if after != "" {
		q.Set("after", after)
	}

	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute hubspot users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch hubspot users: unexpected status %d", httpResp.StatusCode)
	}

	var resp hubspotUsersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode hubspot users response: %w", err)
	}

	return &resp, nil
}

func (d *HubSpotDriver) fetchOwners(
	ctx context.Context,
	archived bool,
	after string,
) (*hubspotOwnersResponse, error) {
	endpoint, err := url.JoinPath(d.baseURL, hubspotOwnersPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build hubspot owners URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create hubspot owners request: %w", err)
	}

	q := req.URL.Query()
	q.Set("limit", "100")
	q.Set("archived", strconv.FormatBool(archived))

	if after != "" {
		q.Set("after", after)
	}

	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute hubspot owners request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch hubspot owners: unexpected status %d", httpResp.StatusCode)
	}

	var resp hubspotOwnersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode hubspot owners response: %w", err)
	}

	return &resp, nil
}

func (d *HubSpotDriver) fetchRoles(ctx context.Context) (map[string]string, error) {
	endpoint, err := url.JoinPath(d.baseURL, hubspotRolesPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build hubspot roles URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create hubspot roles request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute hubspot roles request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch hubspot roles: unexpected status %d", httpResp.StatusCode)
	}

	var resp hubspotRolesResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode hubspot roles response: %w", err)
	}

	roleMap := make(map[string]string, len(resp.Results))
	for _, r := range resp.Results {
		roleMap[r.ID] = r.Name
	}

	return roleMap, nil
}

func hubspotRoles(user hubspotUser, roleMap map[string]string) []string {
	roleIDs := hubspotRoleIDs(user)

	seen := make(map[string]struct{}, len(roleIDs)+1)
	roles := make([]string, 0, len(roleIDs)+1)

	if roleMap != nil {
		for _, id := range roleIDs {
			name, ok := roleMap[id]
			if !ok {
				continue
			}

			if _, dup := seen[name]; dup {
				continue
			}

			seen[name] = struct{}{}
			roles = append(roles, name)
		}
	}

	if user.SuperAdmin {
		if _, dup := seen["Super Admin"]; !dup {
			roles = append(roles, "Super Admin")
		}
	}

	if len(roles) > 0 {
		return roles
	}

	return []string{"User"}
}

func hubspotRoleIDs(user hubspotUser) []string {
	seen := make(map[string]struct{}, 1+len(user.RoleIDs))
	ids := make([]string, 0, 1+len(user.RoleIDs))

	add := func(id string) {
		if id == "" {
			return
		}

		if _, ok := seen[id]; ok {
			return
		}

		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	add(user.RoleID)

	for _, id := range user.RoleIDs {
		add(id)
	}

	return ids
}

func hubspotIndexArchivedOwners(
	owners []hubspotOwner,
) (byUserID map[string]hubspotOwner, byEmail map[string]hubspotOwner) {
	byUserID = make(map[string]hubspotOwner, len(owners))
	byEmail = make(map[string]hubspotOwner, len(owners))

	for _, owner := range owners {
		if !hubspotOwnerIsPerson(owner) {
			continue
		}

		if id := hubspotOwnerUserID(owner); id != "" {
			byUserID[id] = owner
		}

		if email := hubspotNormalizeEmail(owner.Email); email != "" {
			byEmail[email] = owner
		}
	}

	return byUserID, byEmail
}

func hubspotLookupArchived(
	userID string,
	email string,
	byUserID map[string]hubspotOwner,
	byEmail map[string]hubspotOwner,
) (hubspotOwner, bool) {
	if userID != "" {
		if owner, ok := byUserID[userID]; ok {
			return owner, true
		}
	}

	if normalized := hubspotNormalizeEmail(email); normalized != "" {
		if owner, ok := byEmail[normalized]; ok {
			return owner, true
		}
	}

	return hubspotOwner{}, false
}

func hubspotOwnerIsPerson(owner hubspotOwner) bool {
	return owner.Type == "" || strings.EqualFold(owner.Type, "PERSON")
}

// hubspotOwnerUserID returns the Settings Users ID for an owner. Archived
// owners leave userId null and expose the ID only on userIdIncludingInactive.
func hubspotOwnerUserID(owner hubspotOwner) string {
	if owner.UserIDIncludingInactive != nil && *owner.UserIDIncludingInactive != 0 {
		return strconv.FormatInt(*owner.UserIDIncludingInactive, 10)
	}

	if owner.UserID != nil && *owner.UserID != 0 {
		return strconv.FormatInt(*owner.UserID, 10)
	}

	return ""
}

func hubspotDisplayName(firstName, lastName, email string) string {
	fullName := strings.TrimSpace(firstName + " " + lastName)
	if fullName != "" {
		return fullName
	}

	return strings.TrimSpace(email)
}

func hubspotNormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func newHubspotSeenKeys(capacity int) *hubspotSeenKeys {
	return &hubspotSeenKeys{
		userIDs: make(map[string]struct{}, capacity),
		emails:  make(map[string]struct{}, capacity),
	}
}

func (s *hubspotSeenKeys) mark(userID, email string) {
	if userID != "" {
		s.userIDs[userID] = struct{}{}
	}

	if normalized := hubspotNormalizeEmail(email); normalized != "" {
		s.emails[normalized] = struct{}{}
	}
}

func (s *hubspotSeenKeys) has(userID, email string) bool {
	if userID != "" {
		if _, ok := s.userIDs[userID]; ok {
			return true
		}
	}

	if normalized := hubspotNormalizeEmail(email); normalized != "" {
		if _, ok := s.emails[normalized]; ok {
			return true
		}
	}

	return false
}

// hubspotNameResolver resolves the HubSpot account name.
type hubspotNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

func NewHubSpotNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &hubspotNameResolver{httpClient: httpClient, baseURL: baseURL}
}

func (r *hubspotNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	endpoint, err := url.JoinPath(r.baseURL, hubspotAccountInfoPath)
	if err != nil {
		return "", fmt.Errorf("cannot build hubspot account-info URL: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("cannot create hubspot account-info request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute hubspot account-info request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nameStatusError("hubspot account info", httpResp.StatusCode)
	}

	var resp struct {
		PortalID    int    `json:"portalId"`
		AccountName string `json:"accountName"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode hubspot account-info response: %w", err)
	}

	return resp.AccountName, nil
}
