// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

// Package microsoft365 provides a Microsoft 365 (Microsoft Entra ID)
// identity provider for SCIM synchronization using OAuth2 against the
// Microsoft Graph API.
package microsoft365

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	scimclient "go.probo.inc/probo/pkg/iam/scim/bridge/client"
	"go.probo.inc/probo/pkg/iam/scim/bridge/provider"
)

// graphBaseURL is the Microsoft Graph v1.0 endpoint. The bridge uses raw
// HTTP rather than the official SDK because the SDK pulls in a large
// dependency tree for what is a small set of read-only calls.
const graphBaseURL = "https://graph.microsoft.com/v1.0"

// graphUserSelect is the projection used when listing users. Limiting the
// response to only the fields we map keeps payloads small and avoids the
// extra permission scopes some properties would otherwise require.
const graphUserSelect = "id,userPrincipalName,mail,displayName,givenName,surname,accountEnabled,jobTitle,department,companyName,employeeId,preferredLanguage,usageLocation"

// graphPageSize is the maximum page size for /users. Microsoft Graph
// caps /users at 999.
const graphPageSize = 999

// graphMaxPages bounds pagination to prevent unbounded loops if the
// API misbehaves.
const graphMaxPages = 1000

// graphMemberFilter restricts /users to home-tenant members and
// excludes B2B guest users invited from external tenants — those guests
// should not be provisioned into the connected organization.
const graphMemberFilter = "userType eq 'Member'"

var _ provider.Provider = (*Provider)(nil)

type Provider struct {
	httpClient        *http.Client
	excludedUserNames []string
}

func New(httpClient *http.Client, excludedUserNames []string) *Provider {
	return &Provider{
		httpClient:        httpClient,
		excludedUserNames: excludedUserNames,
	}
}

func (p *Provider) Name() string {
	return "microsoft-365"
}

func (p *Provider) isExcluded(email string) bool {
	emailLower := strings.ToLower(email)
	for _, excluded := range p.excludedUserNames {
		if strings.ToLower(excluded) == emailLower {
			return true
		}
	}
	return false
}

// graphUser is the subset of the Microsoft Graph user resource the
// bridge consumes. Managers are intentionally not fetched here: the
// /users endpoint does not return them and per-user lookups would
// multiply the call volume by the directory size. Manager data can be
// added later via /users/{id}/manager when needed.
type graphUser struct {
	ID                string `json:"id"`
	UserPrincipalName string `json:"userPrincipalName"`
	Mail              string `json:"mail"`
	DisplayName       string `json:"displayName"`
	GivenName         string `json:"givenName"`
	Surname           string `json:"surname"`
	AccountEnabled    bool   `json:"accountEnabled"`
	JobTitle          string `json:"jobTitle"`
	Department        string `json:"department"`
	CompanyName       string `json:"companyName"`
	EmployeeID        string `json:"employeeId"`
	PreferredLanguage string `json:"preferredLanguage"`
	UsageLocation     string `json:"usageLocation"`
}

type graphUsersResponse struct {
	Value    []graphUser `json:"value"`
	NextLink string      `json:"@odata.nextLink"`
}

func (p *Provider) ListUsers(ctx context.Context) (scimclient.Users, error) {
	endpoint, err := buildListUsersURL()
	if err != nil {
		return nil, fmt.Errorf("cannot build list users URL: %w", err)
	}

	var allUsers scimclient.Users
	for range graphMaxPages {
		users, next, err := p.fetchPage(ctx, endpoint)
		if err != nil {
			return nil, err
		}

		for _, u := range users {
			email := u.Mail
			if email == "" {
				email = u.UserPrincipalName
			}
			if email == "" {
				continue
			}
			if p.isExcluded(email) {
				continue
			}

			allUsers = append(allUsers, scimclient.User{
				UserName:               email,
				DisplayName:            u.DisplayName,
				GivenName:              u.GivenName,
				FamilyName:             u.Surname,
				Active:                 u.AccountEnabled,
				ExternalID:             u.ID,
				Title:                  u.JobTitle,
				Department:             u.Department,
				EnterpriseOrganization: u.CompanyName,
				EmployeeNumber:         u.EmployeeID,
				PreferredLanguage:      u.PreferredLanguage,
			})
		}

		if next == "" {
			return allUsers, nil
		}
		endpoint = next
	}

	return nil, fmt.Errorf("microsoft 365: pagination limit of %d pages reached", graphMaxPages)
}

func buildListUsersURL() (string, error) {
	u, err := url.Parse(graphBaseURL + "/users")
	if err != nil {
		return "", fmt.Errorf("cannot parse graph base URL: %w", err)
	}

	q := u.Query()
	q.Set("$select", graphUserSelect)
	q.Set("$top", strconv.Itoa(graphPageSize))
	q.Set("$filter", graphMemberFilter)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (p *Provider) fetchPage(ctx context.Context, endpoint string) ([]graphUser, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("cannot create graph users request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("cannot list graph users: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("microsoft graph error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var page graphUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, "", fmt.Errorf("cannot decode graph users response: %w", err)
	}

	return page.Value, page.NextLink, nil
}
