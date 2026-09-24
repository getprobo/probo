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
	"io"
	"net/http"
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	daytonaOrganizationsSegment = "organizations"
	daytonaUsersSegment         = "users"
	daytonaInvitationsSegment   = "invitations"

	daytonaRoleOwner  = "owner"
	daytonaRoleMember = "member"

	daytonaInvitationPending = "pending"
)

type (
	DaytonaDriver struct {
		httpClient     *http.Client
		organizationID string
		baseURL        string
	}

	daytonaAssignedRole struct {
		Name string `json:"name"`
	}

	daytonaUser struct {
		UserID         string                `json:"userId"`
		OrganizationID string                `json:"organizationId"`
		Name           string                `json:"name"`
		Email          string                `json:"email"`
		Role           string                `json:"role"`
		AssignedRoles  []daytonaAssignedRole `json:"assignedRoles"`
		CreatedAt      string                `json:"createdAt"`
	}

	daytonaInvitation struct {
		ID     string `json:"id"`
		Email  string `json:"email"`
		Role   string `json:"role"`
		Status string `json:"status"`
	}

	daytonaNameResolver struct {
		httpClient     *http.Client
		organizationID string
		baseURL        string
	}
)

var _ Driver = (*DaytonaDriver)(nil)

// DaytonaUsersURL builds GET /organizations/{organizationId}/users from an API
// origin. It is exported because the connection probe checks this same
// endpoint from another package: were it to re-derive the path, moving the
// driver's would leave the probe reporting a healthy connection against the
// old one.
func DaytonaUsersURL(baseURL, organizationID string) (string, error) {
	endpoint, err := url.JoinPath(
		baseURL,
		daytonaOrganizationsSegment,
		url.PathEscape(organizationID),
		daytonaUsersSegment,
	)
	if err != nil {
		return "", fmt.Errorf("cannot build daytona users URL: %w", err)
	}

	return endpoint, nil
}

func daytonaInvitationsURL(baseURL, organizationID string) (string, error) {
	endpoint, err := url.JoinPath(
		baseURL,
		daytonaOrganizationsSegment,
		url.PathEscape(organizationID),
		daytonaInvitationsSegment,
	)
	if err != nil {
		return "", fmt.Errorf("cannot build daytona invitations URL: %w", err)
	}

	return endpoint, nil
}

func daytonaOrganizationURL(baseURL, organizationID string) (string, error) {
	endpoint, err := url.JoinPath(
		baseURL,
		daytonaOrganizationsSegment,
		url.PathEscape(organizationID),
	)
	if err != nil {
		return "", fmt.Errorf("cannot build daytona organization URL: %w", err)
	}

	return endpoint, nil
}

func NewDaytonaDriver(httpClient *http.Client, organizationID, baseURL string) *DaytonaDriver {
	return &DaytonaDriver{
		httpClient:     httpClient,
		organizationID: organizationID,
		baseURL:        baseURL,
	}
}

func (d *DaytonaDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	members, err := d.fetchUsers(ctx)
	if err != nil {
		return nil, err
	}

	invitations, err := d.fetchInvitations(ctx)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(members))
	records := make([]AccountRecord, 0, len(members)+len(invitations))

	for _, m := range members {
		email := strings.TrimSpace(m.Email)
		if email == "" {
			continue
		}

		seen[strings.ToLower(email)] = struct{}{}
		records = append(records, daytonaMemberRecord(m, email))
	}

	for _, inv := range invitations {
		email := strings.TrimSpace(inv.Email)
		if email == "" {
			continue
		}

		if !daytonaInvitationIsPending(inv.Status) {
			continue
		}

		if _, ok := seen[strings.ToLower(email)]; ok {
			continue
		}

		records = append(records, daytonaInvitationRecord(inv, email))
	}

	return records, nil
}

func (d *DaytonaDriver) fetchUsers(ctx context.Context) ([]daytonaUser, error) {
	endpoint, err := DaytonaUsersURL(d.baseURL, d.organizationID)
	if err != nil {
		return nil, err
	}

	var users []daytonaUser
	if err := d.getJSON(ctx, endpoint, "users", &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (d *DaytonaDriver) fetchInvitations(ctx context.Context) ([]daytonaInvitation, error) {
	endpoint, err := daytonaInvitationsURL(d.baseURL, d.organizationID)
	if err != nil {
		return nil, err
	}

	var invitations []daytonaInvitation
	if err := d.getJSON(ctx, endpoint, "invitations", &invitations); err != nil {
		return nil, err
	}

	return invitations, nil
}

func (d *DaytonaDriver) getJSON(ctx context.Context, endpoint, what string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("cannot create daytona %s request: %w", what, err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot execute daytona %s request: %w", what, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, resp.Body)

		return fmt.Errorf("cannot fetch daytona %s: unexpected status %d", what, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("cannot decode daytona %s response: %w", what, err)
	}

	return nil
}

func daytonaMemberRecord(m daytonaUser, email string) AccountRecord {
	return AccountRecord{
		Email:       email,
		FullName:    daytonaFullName(m.Name, email),
		Roles:       daytonaRoles(m.Role, m.AssignedRoles),
		IsAdmin:     new(daytonaIsAdmin(m.Role)),
		Active:      new(true),
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:  strings.TrimSpace(m.UserID),
		CreatedAt:   parseRFC3339Ptr(m.CreatedAt),
	}
}

func daytonaInvitationRecord(inv daytonaInvitation, email string) AccountRecord {
	return AccountRecord{
		Email:       email,
		FullName:    email,
		Roles:       daytonaRoles(inv.Role, nil),
		IsAdmin:     new(daytonaIsAdmin(inv.Role)),
		Active:      new(false),
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:  strings.TrimSpace(inv.ID),
	}
}

func daytonaFullName(name, fallback string) string {
	if trimmed := strings.TrimSpace(name); trimmed != "" {
		return trimmed
	}

	return fallback
}

func daytonaNormalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func daytonaIsAdmin(role string) bool {
	return daytonaNormalizeRole(role) == daytonaRoleOwner
}

func daytonaInvitationIsPending(status string) bool {
	normalized := strings.ToLower(strings.TrimSpace(status))

	return normalized == "" || normalized == daytonaInvitationPending
}

func daytonaRoles(role string, assigned []daytonaAssignedRole) []string {
	roles := make([]string, 0, 1+len(assigned))
	seen := make(map[string]struct{})

	add := func(label string) {
		if label == "" {
			return
		}

		key := strings.ToLower(label)
		if _, ok := seen[key]; ok {
			return
		}

		seen[key] = struct{}{}
		roles = append(roles, label)
	}

	switch daytonaNormalizeRole(role) {
	case daytonaRoleOwner:
		add("Owner")
	case daytonaRoleMember:
		add("Member")
	case "":
	default:
		add(strings.TrimSpace(role))
	}

	for _, assignedRole := range assigned {
		add(strings.TrimSpace(assignedRole.Name))
	}

	return roles
}

func NewDaytonaNameResolver(httpClient *http.Client, organizationID, baseURL string) NameResolver {
	return &daytonaNameResolver{
		httpClient:     httpClient,
		organizationID: organizationID,
		baseURL:        baseURL,
	}
}

func (r *daytonaNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.organizationID == "" {
		return "", nil
	}

	endpoint, err := daytonaOrganizationURL(r.baseURL, r.organizationID)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create daytona organization request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute daytona organization request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	// Best-effort: a non-2xx (revoked key, deleted org, stale ID) must not
	// make the source-name worker retry forever. Give up gracefully and keep
	// the generic source name; a dead key surfaces on the next ListAccounts.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode daytona organization response: %w", err)
	}

	return strings.TrimSpace(resp.Name), nil
}
