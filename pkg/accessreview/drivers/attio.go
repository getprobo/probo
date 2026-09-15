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

	"go.probo.inc/probo/pkg/coredata"
)

const (
	attioWorkspaceMembersPath = "/workspace_members"
	attioSelfPath             = "/self"
)

// access_level carries both privilege and account status: a suspended member
// reports "suspended" in place of the role they held.
const (
	attioAccessLevelAdmin     = "admin"
	attioAccessLevelMember    = "member"
	attioAccessLevelSuspended = "suspended"
)

type AttioDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*AttioDriver)(nil)

type attioWorkspaceMembersResponse struct {
	Data []attioWorkspaceMember `json:"data"`
}

type attioWorkspaceMember struct {
	ID struct {
		WorkspaceID       string `json:"workspace_id"`
		WorkspaceMemberID string `json:"workspace_member_id"`
	} `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	EmailAddress string `json:"email_address"`
	CreatedAt    string `json:"created_at"`
	AccessLevel  string `json:"access_level"`
}

func NewAttioDriver(httpClient *http.Client, baseURL string) *AttioDriver {
	return &AttioDriver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// ListAccounts fetches the workspace roster in a single request: Attio's
// GET /v2/workspace_members declares no parameters at all, so there is no
// pagination to follow and no next-page URL that could point off-host.
//
// Unaccepted invitations are not included. Attio exposes them only through its
// SCIM surface, which answers 402 "contact sales" on a workspace without the
// Enterprise entitlement, so there is no generally available way to list them.
func (d *AttioDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	endpoint, err := url.JoinPath(d.baseURL, attioWorkspaceMembersPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build attio workspace members URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create attio workspace members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute attio workspace members request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch attio workspace members: unexpected status %d", httpResp.StatusCode)
	}

	var resp attioWorkspaceMembersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode attio workspace members response: %w", err)
	}

	records := make([]AccountRecord, 0, len(resp.Data))

	for _, m := range resp.Data {
		active, isAdmin := attioAccountStatus(m.AccessLevel)

		record := AccountRecord{
			Email:       m.EmailAddress,
			FullName:    attioFullName(m.FirstName, m.LastName),
			Roles:       attioRoles(m.AccessLevel),
			Active:      active,
			IsAdmin:     isAdmin,
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  m.ID.WorkspaceMemberID,
			CreatedAt:   parseRFC3339Ptr(m.CreatedAt),
		}

		records = append(records, record)
	}

	return records, nil
}

// attioRoles maps the access level onto the role list. Suspension is a status,
// not a privilege, so a suspended member carries no role: Attio has already
// overwritten whatever level they held, and Active reports the suspension.
func attioRoles(accessLevel string) []string {
	switch attioNormalizeAccessLevel(accessLevel) {
	case attioAccessLevelAdmin:
		return []string{"Admin"}
	case attioAccessLevelMember:
		return []string{"Member"}
	case attioAccessLevelSuspended, "":
		return []string{}
	default:
		return []string{accessLevel}
	}
}

func attioNormalizeAccessLevel(accessLevel string) string {
	return strings.ToLower(strings.TrimSpace(accessLevel))
}

// attioAccountStatus derives Active and IsAdmin from the access level. An
// unrecognised level leaves both nil rather than guessing.
func attioAccountStatus(accessLevel string) (active *bool, isAdmin *bool) {
	switch attioNormalizeAccessLevel(accessLevel) {
	case attioAccessLevelAdmin:
		return new(true), new(true)
	case attioAccessLevelMember:
		return new(true), new(false)
	case attioAccessLevelSuspended:
		return new(false), new(false)
	default:
		return nil, nil
	}
}

func attioFullName(firstName, lastName string) string {
	return strings.TrimSpace(strings.TrimSpace(firstName) + " " + strings.TrimSpace(lastName))
}

// attioNameResolver resolves the Attio workspace name via /v2/self.
type attioNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

var _ NameResolver = (*attioNameResolver)(nil)

func NewAttioNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &attioNameResolver{httpClient: httpClient, baseURL: baseURL}
}

func (r *attioNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	endpoint, err := url.JoinPath(r.baseURL, attioSelfPath)
	if err != nil {
		return "", fmt.Errorf("cannot build attio self URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create attio self request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute attio self request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nameStatusError("attio self", httpResp.StatusCode)
	}

	var resp struct {
		Active        bool   `json:"active"`
		WorkspaceName string `json:"workspace_name"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode attio self response: %w", err)
	}

	// Attio answers 200 with {"active": false} for a revoked, deleted or
	// unknown token instead of a 4xx, so the status alone never retires this
	// resolver and the terminal marker has to come from the body.
	if !resp.Active {
		return "", fmt.Errorf("cannot fetch attio self: token is not active: %w", ErrTerminalNameResolution)
	}

	// An active token whose response omits the name leaves the source on its
	// generic title; retrying cannot help, since the token already answered.
	return strings.TrimSpace(resp.WorkspaceName), nil
}
