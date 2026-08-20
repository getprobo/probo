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
	calComMePath        = "/v2/me"
	calComPageSize      = 250
	calComOrganizations = "organizations"
	calComMemberships   = "memberships"
)

type (
	CalComDriver struct {
		httpClient *http.Client
		baseURL    string
	}

	calComMeResponse struct {
		Data struct {
			OrganizationID *int64 `json:"organizationId"`
		} `json:"data"`
	}

	calComMembership struct {
		ID       int64  `json:"id"`
		UserID   int64  `json:"userId"`
		Accepted bool   `json:"accepted"`
		Role     string `json:"role"`
		User     struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"user"`
	}

	calComMembershipsResponse struct {
		Data []calComMembership `json:"data"`
	}
)

var _ Driver = (*CalComDriver)(nil)

func NewCalComDriver(httpClient *http.Client, baseURL string) *CalComDriver {
	return &CalComDriver{httpClient: httpClient, baseURL: baseURL}
}

func (d *CalComDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	me, err := d.fetchMe(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot discover cal.com organization: %w", err)
	}

	if me.Data.OrganizationID == nil {
		return nil, fmt.Errorf("cannot list cal.com accounts: authenticated user has no organization")
	}

	var records []AccountRecord

	for page := range maxPaginationPages {
		memberships, err := d.fetchMembershipsPage(ctx, *me.Data.OrganizationID, page*calComPageSize)
		if err != nil {
			return nil, fmt.Errorf("cannot fetch cal.com memberships page: %w", err)
		}

		for _, membership := range memberships.Data {
			email := strings.TrimSpace(membership.User.Email)
			if email == "" {
				continue
			}

			role := strings.ToUpper(strings.TrimSpace(membership.Role))
			records = append(records, AccountRecord{
				Email:       email,
				FullName:    strings.TrimSpace(membership.User.Name),
				Roles:       calComRoles(role),
				Active:      new(membership.Accepted),
				IsAdmin:     new(role == "OWNER" || role == "ADMIN"),
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  strconv.FormatInt(membership.UserID, 10),
			})
		}

		if len(memberships.Data) < calComPageSize {
			return records, nil
		}
	}

	return nil, fmt.Errorf("cannot list all cal.com accounts: %w", ErrPaginationLimitReached)
}

func (d *CalComDriver) fetchMe(ctx context.Context) (*calComMeResponse, error) {
	endpoint, err := url.JoinPath(d.baseURL, calComMePath)
	if err != nil {
		return nil, fmt.Errorf("cannot build cal.com profile URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create cal.com profile request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute cal.com profile request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch cal.com profile: unexpected status %d", httpResp.StatusCode)
	}

	var resp calComMeResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode cal.com profile response: %w", err)
	}

	return &resp, nil
}

func (d *CalComDriver) fetchMembershipsPage(
	ctx context.Context,
	organizationID int64,
	skip int,
) (*calComMembershipsResponse, error) {
	endpoint, err := url.JoinPath(
		d.baseURL,
		"v2",
		calComOrganizations,
		strconv.FormatInt(organizationID, 10),
		calComMemberships,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot build cal.com memberships URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create cal.com memberships request: %w", err)
	}

	query := req.URL.Query()
	query.Set("take", strconv.Itoa(calComPageSize))
	query.Set("skip", strconv.Itoa(skip))
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute cal.com memberships request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch cal.com memberships: unexpected status %d", httpResp.StatusCode)
	}

	var resp calComMembershipsResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode cal.com memberships response: %w", err)
	}

	return &resp, nil
}

func calComRoles(role string) []string {
	switch role {
	case "OWNER":
		return []string{"Owner"}
	case "ADMIN":
		return []string{"Admin"}
	case "MEMBER":
		return []string{"Member"}
	case "":
		return []string{}
	default:
		return []string{role}
	}
}
