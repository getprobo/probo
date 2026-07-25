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

	"go.probo.inc/probo/pkg/coredata"
)

const (
	segmentPageSize = 200
	// segmentWorkspaceOwnerRole is the only Segment role that grants
	// workspace-wide administrative access; the resource-scoped *Admin roles
	// (Source Admin, Warehouse Admin, …) administer a single resource, not the
	// workspace.
	segmentWorkspaceOwnerRole = "Workspace Owner"
)

// SegmentDriver lists the members of a single Twilio Segment workspace. The
// Public API token (sent as Authorization: Bearer by the connection transport)
// is bound to one workspace on one regional host. GET /users returns members
// without their roles, so each member's roles are fetched with a follow-up GET
// /users/{id} (an unavoidable N+1); GET /invites adds pending invitations as
// inactive members.
type SegmentDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*SegmentDriver)(nil)

type segmentUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	// Permissions is declared on the shared UserV1 schema that both /users
	// and /users/{id} return, but is optional and today only populated by
	// the single-user read. Decoding it here means the driver uses whatever
	// the list gives it rather than assuming: nil (field absent) triggers the
	// per-user fetch, non-nil — including an empty array for a user with no
	// roles — is taken as authoritative.
	Permissions []segmentPermission `json:"permissions"`
}

type segmentPermission struct {
	RoleName string `json:"roleName"`
}

type segmentPaginationOut struct {
	// Next is absent (nil) on the last page.
	Next *string `json:"next"`
}

type segmentUsersResponse struct {
	Data struct {
		Users      []segmentUser        `json:"users"`
		Pagination segmentPaginationOut `json:"pagination"`
	} `json:"data"`
}

type segmentUserResponse struct {
	Data struct {
		User struct {
			Permissions []segmentPermission `json:"permissions"`
		} `json:"user"`
	} `json:"data"`
}

type segmentInvitesResponse struct {
	Data struct {
		Invites    []string             `json:"invites"`
		Pagination segmentPaginationOut `json:"pagination"`
	} `json:"data"`
}

func NewSegmentDriver(httpClient *http.Client, baseURL string) *SegmentDriver {
	return &SegmentDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		baseURL: baseURL,
	}
}

func (d *SegmentDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	base, err := url.Parse(d.baseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse segment base URL: %w", err)
	}

	users, err := d.listUsers(ctx, base)
	if err != nil {
		return nil, err
	}

	records := make([]AccountRecord, 0, len(users))
	seen := make(map[string]struct{}, len(users))

	for _, u := range users {
		email := strings.TrimSpace(u.Email)
		if email == "" {
			continue
		}

		seen[strings.ToLower(email)] = struct{}{}

		perms := u.Permissions
		if perms == nil {
			perms, err = d.userPermissions(ctx, base, u.ID)
			if err != nil {
				return nil, fmt.Errorf("cannot list segment permissions for user %q: %w", u.ID, err)
			}
		}

		roles, isAdmin := segmentRolesAndAdmin(perms)

		// Segment's user API exposes no active/suspended status field, so
		// leave Active nil (unknown) rather than fabricate a value, per the
		// AccountRecord contract.
		records = append(records, AccountRecord{
			Email:       email,
			FullName:    segmentFullName(u.Name, email),
			Roles:       roles,
			Active:      nil,
			IsAdmin:     isAdmin,
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  u.ID,
		})
	}

	invites, err := d.listInvites(ctx, base)
	if err != nil {
		return nil, err
	}

	for _, email := range invites {
		email = strings.TrimSpace(email)
		if email == "" {
			continue
		}

		// An invite that has already been accepted can still be listed; the
		// member record above is the authoritative one, so skip the duplicate
		// rather than emit two rows for the same person.
		if _, ok := seen[strings.ToLower(email)]; ok {
			continue
		}

		// A pending invite carries no role and no stable id at the workspace
		// level, so it is surfaced as an inactive member keyed by email.
		active := false

		records = append(records, AccountRecord{
			Email:       email,
			FullName:    email,
			Roles:       []string{},
			Active:      &active,
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  email,
		})
	}

	return records, nil
}

func (d *SegmentDriver) listUsers(ctx context.Context, base *url.URL) ([]segmentUser, error) {
	var users []segmentUser

	cursor := ""

	for range maxPaginationPages {
		endpoint := *base
		endpoint.Path = "/users"

		q := url.Values{}
		q.Set("pagination.count", strconv.Itoa(segmentPageSize))

		if cursor != "" {
			q.Set("pagination.cursor", cursor)
		}

		endpoint.RawQuery = q.Encode()

		var resp segmentUsersResponse
		if err := d.getJSON(ctx, endpoint.String(), &resp); err != nil {
			return nil, err
		}

		users = append(users, resp.Data.Users...)

		if resp.Data.Pagination.Next == nil || *resp.Data.Pagination.Next == "" {
			return users, nil
		}

		cursor = *resp.Data.Pagination.Next
	}

	return nil, fmt.Errorf("cannot list all segment users: %w", ErrPaginationLimitReached)
}

func (d *SegmentDriver) userPermissions(ctx context.Context, base *url.URL, userID string) ([]segmentPermission, error) {
	endpoint, err := url.JoinPath(base.String(), "users", url.PathEscape(userID))
	if err != nil {
		return nil, fmt.Errorf("cannot build segment user URL: %w", err)
	}

	var resp segmentUserResponse
	if err := d.getJSON(ctx, endpoint, &resp); err != nil {
		return nil, err
	}

	return resp.Data.User.Permissions, nil
}

func (d *SegmentDriver) listInvites(ctx context.Context, base *url.URL) ([]string, error) {
	var invites []string

	cursor := ""

	for range maxPaginationPages {
		endpoint := *base
		endpoint.Path = "/invites"

		q := url.Values{}
		q.Set("pagination.count", strconv.Itoa(segmentPageSize))

		if cursor != "" {
			q.Set("pagination.cursor", cursor)
		}

		endpoint.RawQuery = q.Encode()

		var resp segmentInvitesResponse
		if err := d.getJSON(ctx, endpoint.String(), &resp); err != nil {
			return nil, err
		}

		invites = append(invites, resp.Data.Invites...)

		if resp.Data.Pagination.Next == nil || *resp.Data.Pagination.Next == "" {
			return invites, nil
		}

		cursor = *resp.Data.Pagination.Next
	}

	return nil, fmt.Errorf("cannot list all segment invites: %w", ErrPaginationLimitReached)
}

func (d *SegmentDriver) getJSON(ctx context.Context, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("cannot create segment request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot execute segment request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return fmt.Errorf("cannot fetch segment resource: unexpected status %d", httpResp.StatusCode)
	}

	if err := json.NewDecoder(httpResp.Body).Decode(out); err != nil {
		return fmt.Errorf("cannot decode segment response: %w", err)
	}

	return nil
}

// segmentFullName returns the member's name, falling back to the email when
// Segment stores an empty name.
func segmentFullName(name, email string) string {
	if n := strings.TrimSpace(name); n != "" {
		return n
	}

	return email
}

// segmentRolesAndAdmin collapses a member's permission entries into a
// de-duplicated, sorted set of role names and reports whether any of them is
// the workspace-level Workspace Owner role.
func segmentRolesAndAdmin(perms []segmentPermission) ([]string, bool) {
	seen := make(map[string]struct{})
	isAdmin := false

	for _, p := range perms {
		name := strings.TrimSpace(p.RoleName)
		if name == "" {
			continue
		}

		seen[name] = struct{}{}

		if strings.EqualFold(name, segmentWorkspaceOwnerRole) {
			isAdmin = true
		}
	}

	roles := make([]string, 0, len(seen))
	for name := range seen {
		roles = append(roles, name)
	}

	sort.Strings(roles)

	return roles, isAdmin
}
