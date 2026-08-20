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
	calendlyCurrentUserPath = "/users/me"
	calendlyMembershipsPath = "/organization_memberships"
	calendlyPageSize        = 100
)

type (
	CalendlyDriver struct {
		httpClient *http.Client
		baseURL    string
	}

	calendlyCurrentUserResponse struct {
		Resource struct {
			CurrentOrganization string `json:"current_organization"`
		} `json:"resource"`
	}

	calendlyMembership struct {
		Role string `json:"role"`
		User struct {
			URI       string `json:"uri"`
			Name      string `json:"name"`
			Email     string `json:"email"`
			CreatedAt string `json:"created_at"`
		} `json:"user"`
	}

	calendlyMembershipsResponse struct {
		Collection []calendlyMembership `json:"collection"`
		Pagination struct {
			NextPageToken string `json:"next_page_token"`
		} `json:"pagination"`
	}
)

var _ Driver = (*CalendlyDriver)(nil)

func NewCalendlyDriver(httpClient *http.Client, baseURL string) *CalendlyDriver {
	return &CalendlyDriver{httpClient: httpClient, baseURL: baseURL}
}

func (d *CalendlyDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	currentUser, err := d.fetchCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot discover calendly organization: %w", err)
	}

	organization := strings.TrimSpace(currentUser.Resource.CurrentOrganization)
	if organization == "" {
		return nil, fmt.Errorf("cannot list calendly accounts: authenticated user has no organization")
	}

	var (
		records   []AccountRecord
		pageToken string
	)

	for range maxPaginationPages {
		memberships, err := d.fetchMembershipsPage(ctx, organization, pageToken)
		if err != nil {
			return nil, fmt.Errorf("cannot fetch calendly memberships page: %w", err)
		}

		for _, membership := range memberships.Collection {
			email := strings.TrimSpace(membership.User.Email)
			if email == "" {
				continue
			}

			role := strings.ToLower(strings.TrimSpace(membership.Role))
			records = append(records, AccountRecord{
				Email:       email,
				FullName:    strings.TrimSpace(membership.User.Name),
				Roles:       calendlyRoles(role),
				IsAdmin:     new(role == "owner" || role == "admin"),
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				ExternalID:  strings.TrimSpace(membership.User.URI),
				CreatedAt:   parseRFC3339Ptr(membership.User.CreatedAt),
			})
		}

		if memberships.Pagination.NextPageToken == "" {
			return records, nil
		}

		pageToken = memberships.Pagination.NextPageToken
	}

	return nil, fmt.Errorf("cannot list all calendly accounts: %w", ErrPaginationLimitReached)
}

func (d *CalendlyDriver) fetchCurrentUser(ctx context.Context) (*calendlyCurrentUserResponse, error) {
	endpoint, err := url.JoinPath(d.baseURL, calendlyCurrentUserPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build calendly current user URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create calendly current user request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute calendly current user request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch calendly current user: unexpected status %d", httpResp.StatusCode)
	}

	var resp calendlyCurrentUserResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode calendly current user response: %w", err)
	}

	return &resp, nil
}

func (d *CalendlyDriver) fetchMembershipsPage(
	ctx context.Context,
	organization string,
	pageToken string,
) (*calendlyMembershipsResponse, error) {
	endpoint, err := url.JoinPath(d.baseURL, calendlyMembershipsPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build calendly memberships URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create calendly memberships request: %w", err)
	}

	query := req.URL.Query()
	query.Set("organization", organization)
	query.Set("count", strconv.Itoa(calendlyPageSize))

	if pageToken != "" {
		query.Set("page_token", pageToken)
	}

	req.URL.RawQuery = query.Encode()
	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute calendly memberships request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch calendly memberships: unexpected status %d", httpResp.StatusCode)
	}

	var resp calendlyMembershipsResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode calendly memberships response: %w", err)
	}

	return &resp, nil
}

func calendlyRoles(role string) []string {
	switch role {
	case "owner":
		return []string{"Owner"}
	case "admin":
		return []string{"Admin"}
	case "user":
		return []string{"User"}
	case "":
		return []string{}
	default:
		return []string{role}
	}
}
