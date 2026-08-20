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
	calComMePath        = "/v2/me"
	calComTeamsPath     = "/v2/teams"
	calComPageSize      = 250
	calComOrganizations = "organizations"
	calComTeams         = "teams"
	calComMemberships   = "memberships"
)

type (
	CalComDriver struct {
		httpClient *http.Client
		baseURL    string
	}

	calComMeResponse struct {
		Data struct {
			ID             int64  `json:"id"`
			Name           string `json:"name"`
			Email          string `json:"email"`
			OrganizationID *int64 `json:"organizationId"`
		} `json:"data"`
	}

	calComTeam struct {
		ID int64 `json:"id"`
	}

	calComTeamsResponse struct {
		Data []calComTeam `json:"data"`
	}

	calComMembership struct {
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

	calComAccountAggregate struct {
		record  AccountRecord
		roles   map[string]struct{}
		active  bool
		isAdmin bool
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
		teams, err := d.fetchTeams(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot discover cal.com teams: %w", err)
		}

		if len(teams.Data) == 0 {
			return calComSoloRecord(me), nil
		}

		var memberships []calComMembership

		for _, team := range teams.Data {
			teamMemberships, err := d.fetchAllMemberships(ctx, calComTeams, team.ID)
			if err != nil {
				return nil, fmt.Errorf("cannot fetch cal.com team memberships: %w", err)
			}

			memberships = append(memberships, teamMemberships...)
		}

		return calComMembershipRecords(memberships), nil
	}

	memberships, err := d.fetchAllMemberships(ctx, calComOrganizations, *me.Data.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch cal.com organization memberships: %w", err)
	}

	return calComMembershipRecords(memberships), nil
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

func (d *CalComDriver) fetchTeams(ctx context.Context) (*calComTeamsResponse, error) {
	endpoint, err := url.JoinPath(d.baseURL, calComTeamsPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build cal.com teams URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create cal.com teams request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute cal.com teams request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch cal.com teams: unexpected status %d", httpResp.StatusCode)
	}

	var resp calComTeamsResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode cal.com teams response: %w", err)
	}

	return &resp, nil
}

func (d *CalComDriver) fetchAllMemberships(
	ctx context.Context,
	resource string,
	resourceID int64,
) ([]calComMembership, error) {
	var memberships []calComMembership

	for page := range maxPaginationPages {
		resp, err := d.fetchMembershipsPage(ctx, resource, resourceID, page*calComPageSize)
		if err != nil {
			return nil, fmt.Errorf("cannot fetch cal.com memberships page: %w", err)
		}

		memberships = append(memberships, resp.Data...)
		if len(resp.Data) < calComPageSize {
			return memberships, nil
		}
	}

	return nil, fmt.Errorf("cannot list all cal.com memberships: %w", ErrPaginationLimitReached)
}

func (d *CalComDriver) fetchMembershipsPage(
	ctx context.Context,
	resource string,
	resourceID int64,
	skip int,
) (*calComMembershipsResponse, error) {
	endpoint, err := url.JoinPath(
		d.baseURL,
		"v2",
		resource,
		strconv.FormatInt(resourceID, 10),
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

func calComSoloRecord(me *calComMeResponse) []AccountRecord {
	email := strings.TrimSpace(me.Data.Email)
	if email == "" {
		return []AccountRecord{}
	}

	return []AccountRecord{
		{
			Email:       email,
			FullName:    strings.TrimSpace(me.Data.Name),
			Roles:       []string{},
			Active:      new(true),
			IsAdmin:     new(false),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  strconv.FormatInt(me.Data.ID, 10),
		},
	}
}

func calComMembershipRecords(memberships []calComMembership) []AccountRecord {
	order := make([]string, 0)
	byUser := make(map[string]*calComAccountAggregate)

	for _, membership := range memberships {
		email := strings.TrimSpace(membership.User.Email)
		if email == "" {
			continue
		}

		externalID := strconv.FormatInt(membership.UserID, 10)

		key := externalID
		if membership.UserID == 0 {
			externalID = ""
			key = strings.ToLower(email)
		}

		aggregate, ok := byUser[key]
		if !ok {
			aggregate = &calComAccountAggregate{
				record: AccountRecord{
					Email:       email,
					FullName:    strings.TrimSpace(membership.User.Name),
					MFAStatus:   coredata.MFAStatusUnknown,
					AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
					AccountType: coredata.AccessReviewEntryAccountTypeUser,
					ExternalID:  externalID,
				},
				roles: make(map[string]struct{}),
			}
			byUser[key] = aggregate
			order = append(order, key)
		}

		role := strings.ToUpper(strings.TrimSpace(membership.Role))
		for _, name := range calComRoles(role) {
			aggregate.roles[name] = struct{}{}
		}

		aggregate.active = aggregate.active || membership.Accepted
		aggregate.isAdmin = aggregate.isAdmin || role == "OWNER" || role == "ADMIN"
	}

	records := make([]AccountRecord, 0, len(order))
	for _, key := range order {
		aggregate := byUser[key]

		roles := make([]string, 0, len(aggregate.roles))
		for role := range aggregate.roles {
			roles = append(roles, role)
		}

		sort.Strings(roles)

		aggregate.record.Roles = roles
		aggregate.record.Active = new(aggregate.active)
		aggregate.record.IsAdmin = new(aggregate.isAdmin)
		records = append(records, aggregate.record)
	}

	return records
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
