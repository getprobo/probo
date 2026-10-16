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
	"strconv"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	planetScaleOrganizationsPath = "/organizations"
	planetScaleMembersSegment    = "members"

	// The members endpoint caps per_page at 100.
	planetScalePageSize = 100

	planetScaleRoleAdmin = "admin"
)

const planetScaleDefaultBaseURL = "https://api.planetscale.com/v1"

// PlanetScaleDriver lists the members of one PlanetScale organization.
// PlanetScale exposes no member status or sign-in time.
type PlanetScaleDriver struct {
	httpClient   *http.Client
	organization string
	baseURL      string
}

var _ Driver = (*PlanetScaleDriver)(nil)

type planetScaleOrganization struct {
	Name string `json:"name"`
	SSO  bool   `json:"sso"`
}

type planetScaleOrganizationsResponse struct {
	NextPage *int                      `json:"next_page"`
	Data     []planetScaleOrganization `json:"data"`
}

type planetScaleMembersResponse struct {
	NextPage *int                    `json:"next_page"`
	Data     []planetScaleMembership `json:"data"`
}

type planetScaleMembership struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	User      struct {
		ID                      string `json:"id"`
		Email                   string `json:"email"`
		Name                    string `json:"name"`
		DisplayName             string `json:"display_name"`
		TwoFactorAuthConfigured *bool  `json:"two_factor_auth_configured"`
	} `json:"user"`
}

// NewPlanetScaleDriver builds a driver scoped to organization against baseURL
// (e.g. https://api.planetscale.com/v1).
func NewPlanetScaleDriver(httpClient *http.Client, organization, baseURL string) *PlanetScaleDriver {
	client := *httpClient
	client.Transport = &retryRoundTripper{
		next:       httpClient.Transport,
		maxRetries: 3,
	}

	return &PlanetScaleDriver{httpClient: &client, organization: organization, baseURL: baseURL}
}

func (d *PlanetScaleDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	org, err := d.fetchOrganization(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load planetscale organization: %w", err)
	}

	records := []AccountRecord{}
	page := 1

	for range maxPaginationPages {
		resp, err := d.queryMembers(ctx, page)
		if err != nil {
			return nil, fmt.Errorf("cannot list planetscale members page %d: %w", page, err)
		}

		for _, m := range resp.Data {
			if m.User.ID == "" {
				return nil, fmt.Errorf("cannot list planetscale members: membership %q has no user id", m.ID)
			}

			records = append(records, planetScaleRecord(m, org.SSO))
		}

		if resp.NextPage == nil {
			return records, nil
		}

		if *resp.NextPage <= page {
			return nil, fmt.Errorf("cannot list all planetscale members: next page %d does not follow page %d", *resp.NextPage, page)
		}

		page = *resp.NextPage
	}

	return nil, fmt.Errorf("cannot list all planetscale members: %w", ErrPaginationLimitReached)
}

func (d *PlanetScaleDriver) fetchOrganization(ctx context.Context) (*planetScaleOrganization, error) {
	endpoint, err := url.JoinPath(d.baseURL, planetScaleOrganizationsPath, url.PathEscape(d.organization))
	if err != nil {
		return nil, fmt.Errorf("cannot build planetscale organization URL: %w", err)
	}

	var org planetScaleOrganization

	status, err := planetScaleGet(ctx, d.httpClient, endpoint, "organization", &org)
	if err != nil {
		return nil, fmt.Errorf("cannot request planetscale organization: %w", err)
	}

	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("cannot fetch planetscale organization: unexpected status %d", status)
	}

	return &org, nil
}

func (d *PlanetScaleDriver) queryMembers(ctx context.Context, page int) (*planetScaleMembersResponse, error) {
	endpoint, err := planetScalePageURL(
		d.baseURL,
		page,
		planetScalePageSize,
		planetScaleOrganizationsPath,
		url.PathEscape(d.organization),
		planetScaleMembersSegment,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot build planetscale members URL: %w", err)
	}

	var resp planetScaleMembersResponse

	status, err := planetScaleGet(ctx, d.httpClient, endpoint, "members", &resp)
	if err != nil {
		return nil, fmt.Errorf("cannot request planetscale members: %w", err)
	}

	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("cannot fetch planetscale members: unexpected status %d", status)
	}

	return &resp, nil
}

func planetScaleRecord(m planetScaleMembership, orgSSO bool) AccountRecord {
	return AccountRecord{
		Email:       m.User.Email,
		FullName:    planetScaleFullName(m.User.Name, m.User.DisplayName, m.User.Email),
		Roles:       planetScaleRoles(m.Role),
		IsAdmin:     planetScaleIsAdmin(m.Role),
		MFAStatus:   planetScaleMFAStatus(m.User.TwoFactorAuthConfigured, orgSSO),
		AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:  m.User.ID,
		CreatedAt:   parseRFC3339Ptr(m.CreatedAt),
	}
}

// planetScaleFullName ignores a display_name that only repeats the email.
func planetScaleFullName(name, displayName, email string) string {
	if name = strings.TrimSpace(name); name != "" {
		return name
	}

	displayName = strings.TrimSpace(displayName)
	if strings.EqualFold(displayName, strings.TrimSpace(email)) {
		return ""
	}

	return displayName
}

// planetScaleRoles keeps the organization role as PlanetScale names it.
func planetScaleRoles(role string) []string {
	if role = strings.TrimSpace(role); role == "" {
		return []string{}
	}

	return []string{role}
}

func planetScaleIsAdmin(role string) *bool {
	if role = strings.TrimSpace(role); role == "" {
		return nil
	}

	return new(strings.EqualFold(role, planetScaleRoleAdmin))
}

// planetScaleMFAStatus reports unknown under SSO, where the flag says nothing about the IdP login.
func planetScaleMFAStatus(configured *bool, orgSSO bool) coredata.MFAStatus {
	switch {
	case configured == nil:
		return coredata.MFAStatusUnknown
	case orgSSO:
		return coredata.MFAStatusUnknown
	case *configured:
		return coredata.MFAStatusEnabled
	default:
		return coredata.MFAStatusDisabled
	}
}

// planetScalePageURL joins segments onto baseURL and asks for one page.
func planetScalePageURL(baseURL string, page, perPage int, segments ...string) (string, error) {
	endpoint, err := url.JoinPath(baseURL, segments...)
	if err != nil {
		return "", fmt.Errorf("cannot join planetscale URL: %w", err)
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("cannot parse planetscale URL: %w", err)
	}

	q := parsed.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("per_page", strconv.Itoa(perPage))
	parsed.RawQuery = q.Encode()

	return parsed.String(), nil
}

// planetScaleGet decodes a 2xx body into v and returns any other status
// without an error, for the caller to classify.
func planetScaleGet(ctx context.Context, httpClient *http.Client, endpoint, what string, v any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("cannot create planetscale %s request: %w", what, err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("cannot execute planetscale %s request: %w", what, err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(httpResp.Body, 1<<20))

		return httpResp.StatusCode, nil
	}

	if err := json.NewDecoder(httpResp.Body).Decode(v); err != nil {
		return 0, fmt.Errorf("cannot decode planetscale %s response: %w", what, err)
	}

	return httpResp.StatusCode, nil
}

// ListPlanetScaleOrganizations lists the organizations whose members the token
// can read, from baseURL ("" for the PlanetScale API).
func ListPlanetScaleOrganizations(ctx context.Context, httpClient *http.Client, baseURL string) ([]Organization, error) {
	if baseURL == "" {
		baseURL = planetScaleDefaultBaseURL
	}

	result := []Organization{}
	page := 1

	for range maxPaginationPages {
		resp, err := queryPlanetScaleOrganizations(ctx, httpClient, baseURL, page)
		if err != nil {
			return nil, fmt.Errorf("cannot list planetscale organizations page %d: %w", page, err)
		}

		for _, o := range resp.Data {
			granted, err := planetScaleMembersGranted(ctx, httpClient, baseURL, o.Name)
			if err != nil {
				return nil, fmt.Errorf("cannot check planetscale organization grant: %w", err)
			}

			if granted {
				result = append(result, Organization{Slug: o.Name, DisplayName: o.Name})
			}
		}

		if resp.NextPage == nil {
			return result, nil
		}

		if *resp.NextPage <= page {
			return nil, fmt.Errorf("cannot list all planetscale organizations: next page %d does not follow page %d", *resp.NextPage, page)
		}

		page = *resp.NextPage
	}

	return nil, fmt.Errorf("cannot list all planetscale organizations: %w", ErrPaginationLimitReached)
}

func queryPlanetScaleOrganizations(
	ctx context.Context,
	httpClient *http.Client,
	baseURL string,
	page int,
) (*planetScaleOrganizationsResponse, error) {
	endpoint, err := planetScalePageURL(baseURL, page, planetScalePageSize, planetScaleOrganizationsPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build planetscale organizations URL: %w", err)
	}

	var resp planetScaleOrganizationsResponse

	status, err := planetScaleGet(ctx, httpClient, endpoint, "organizations", &resp)
	if err != nil {
		return nil, fmt.Errorf("cannot request planetscale organizations: %w", err)
	}

	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("cannot fetch planetscale organizations: unexpected status %d", status)
	}

	return &resp, nil
}

// planetScaleMembersGranted is false on 403: the organization was not granted at consent.
func planetScaleMembersGranted(ctx context.Context, httpClient *http.Client, baseURL, organization string) (bool, error) {
	endpoint, err := planetScalePageURL(
		baseURL,
		1,
		1,
		planetScaleOrganizationsPath,
		url.PathEscape(organization),
		planetScaleMembersSegment,
	)
	if err != nil {
		return false, fmt.Errorf("cannot build planetscale members URL: %w", err)
	}

	var resp planetScaleMembersResponse

	status, err := planetScaleGet(ctx, httpClient, endpoint, "members", &resp)
	if err != nil {
		return false, fmt.Errorf("cannot request planetscale members: %w", err)
	}

	switch {
	case status >= 200 && status < 300:
		return true, nil
	case status == http.StatusForbidden:
		return false, nil
	default:
		return false, fmt.Errorf("cannot fetch planetscale members: unexpected status %d", status)
	}
}
