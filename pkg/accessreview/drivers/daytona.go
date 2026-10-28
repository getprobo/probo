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

	"go.probo.inc/probo/pkg/coredata"
)

const (
	daytonaCurrentAPIKeyPath = "/api-keys/current"
	daytonaOrganizationsPath = "/organizations"
	daytonaUsersSegment      = "users"
)

type (
	// DaytonaDriver lists the members of the Daytona organization the API
	// key belongs to. Daytona exposes no member status, sign-in time or MFA,
	// and an API key cannot read pending invitations.
	DaytonaDriver struct {
		httpClient *http.Client
		baseURL    string
	}

	daytonaAPIKey struct {
		OrganizationID string `json:"organizationId"`
	}

	daytonaMember struct {
		UserID        string `json:"userId"`
		Name          string `json:"name"`
		Email         string `json:"email"`
		Role          string `json:"role"`
		AssignedRoles []struct {
			Name string `json:"name"`
		} `json:"assignedRoles"`
		CreatedAt string `json:"createdAt"`
	}
)

var _ Driver = (*DaytonaDriver)(nil)

// NewDaytonaDriver builds a driver against baseURL (e.g.
// https://app.daytona.io/api).
func NewDaytonaDriver(httpClient *http.Client, baseURL string) *DaytonaDriver {
	client := *httpClient
	client.Transport = &retryRoundTripper{
		next:       httpClient.Transport,
		maxRetries: 3,
	}

	return &DaytonaDriver{httpClient: &client, baseURL: baseURL}
}

func (d *DaytonaDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	organizationID, err := d.fetchOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve daytona organization: %w", err)
	}

	endpoint, err := url.JoinPath(
		d.baseURL,
		daytonaOrganizationsPath,
		url.PathEscape(organizationID),
		daytonaUsersSegment,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot build daytona members URL: %w", err)
	}

	var members []daytonaMember
	if err := daytonaGet(ctx, d.httpClient, endpoint, "members", &members); err != nil {
		return nil, fmt.Errorf("cannot list daytona members: %w", err)
	}

	records := make([]AccountRecord, 0, len(members))
	for _, m := range members {
		if m.UserID == "" {
			return nil, fmt.Errorf("cannot list daytona members: member has no user id")
		}

		records = append(records, daytonaRecord(m))
	}

	return records, nil
}

// fetchOrganizationID reads the organization the API key belongs to. A key
// is bound to one organization, and the members endpoint takes its id.
func (d *DaytonaDriver) fetchOrganizationID(ctx context.Context) (string, error) {
	endpoint, err := url.JoinPath(d.baseURL, daytonaCurrentAPIKeyPath)
	if err != nil {
		return "", fmt.Errorf("cannot build daytona api key URL: %w", err)
	}

	var key daytonaAPIKey
	if err := daytonaGet(ctx, d.httpClient, endpoint, "api key", &key); err != nil {
		return "", fmt.Errorf("cannot read daytona api key: %w", err)
	}

	if key.OrganizationID == "" {
		return "", fmt.Errorf("cannot read daytona api key: no organization id")
	}

	return key.OrganizationID, nil
}

func daytonaRecord(m daytonaMember) AccountRecord {
	return AccountRecord{
		Email:       m.Email,
		FullName:    m.Name,
		Roles:       daytonaRoles(m),
		IsAdmin:     new(isOwnerRole(m.Role)),
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:  m.UserID,
		CreatedAt:   parseRFC3339Ptr(m.CreatedAt),
	}
}

// daytonaRoles lists the organization role, then the assignments that grant
// a member its permissions.
func daytonaRoles(m daytonaMember) []string {
	roles := ownerMemberRoles(m.Role)
	for _, r := range m.AssignedRoles {
		roles = append(roles, r.Name)
	}

	return roles
}

func daytonaGet(ctx context.Context, httpClient *http.Client, endpoint, what string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("cannot create daytona %s request: %w", what, err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot execute daytona %s request: %w", what, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))

		return fmt.Errorf("cannot fetch daytona %s: unexpected status %d", what, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("cannot decode daytona %s response: %w", what, err)
	}

	return nil
}
