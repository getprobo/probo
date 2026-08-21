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
	attioWorkspaceMembersPath = "/v2/workspace_members"
	attioIdentityPath         = "/v2/self"
)

type (
	AttioDriver struct {
		httpClient *http.Client
		baseURL    string
	}

	attioWorkspaceMember struct {
		ID struct {
			WorkspaceMemberID string `json:"workspace_member_id"`
		} `json:"id"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		EmailAddress string `json:"email_address"`
		CreatedAt    string `json:"created_at"`
		AccessLevel  string `json:"access_level"`
	}

	attioWorkspaceMembersResponse struct {
		Data []attioWorkspaceMember `json:"data"`
	}

	attioNameResolver struct {
		httpClient *http.Client
		baseURL    string
	}

	attioIdentityResponse struct {
		Active        bool   `json:"active"`
		WorkspaceName string `json:"workspace_name"`
	}
)

var (
	_ Driver       = (*AttioDriver)(nil)
	_ NameResolver = (*attioNameResolver)(nil)
)

func NewAttioDriver(httpClient *http.Client, baseURL string) *AttioDriver {
	return &AttioDriver{httpClient: httpClient, baseURL: baseURL}
}

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
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch attio workspace members: unexpected status %d", httpResp.StatusCode)
	}

	var resp attioWorkspaceMembersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode attio workspace members response: %w", err)
	}

	records := make([]AccountRecord, 0, len(resp.Data))
	for _, member := range resp.Data {
		email := strings.TrimSpace(member.EmailAddress)
		if email == "" {
			continue
		}

		accessLevel := strings.ToLower(strings.TrimSpace(member.AccessLevel))
		records = append(records, AccountRecord{
			Email:       email,
			FullName:    strings.TrimSpace(strings.Join([]string{member.FirstName, member.LastName}, " ")),
			Roles:       attioRoles(accessLevel),
			Active:      new(accessLevel != "suspended"),
			IsAdmin:     new(accessLevel == "admin"),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  strings.TrimSpace(member.ID.WorkspaceMemberID),
			CreatedAt:   parseRFC3339Ptr(member.CreatedAt),
		})
	}

	return records, nil
}

func NewAttioNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &attioNameResolver{httpClient: httpClient, baseURL: baseURL}
}

func (r *attioNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	endpoint, err := url.JoinPath(r.baseURL, attioIdentityPath)
	if err != nil {
		return "", fmt.Errorf("cannot build attio identity URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create attio identity request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute attio identity request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nameStatusError("attio identity", httpResp.StatusCode)
	}

	var resp attioIdentityResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode attio identity response: %w", err)
	}

	if !resp.Active {
		return "", fmt.Errorf("cannot resolve attio workspace name: inactive token: %w", ErrTerminalNameResolution)
	}

	return strings.TrimSpace(resp.WorkspaceName), nil
}

func attioRoles(accessLevel string) []string {
	switch accessLevel {
	case "admin":
		return []string{"Admin"}
	case "member":
		return []string{"Member"}
	case "suspended":
		return []string{"Suspended"}
	case "":
		return []string{}
	default:
		return []string{accessLevel}
	}
}
