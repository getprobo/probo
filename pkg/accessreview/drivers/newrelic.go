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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	// The three data regions New Relic runs NerdGraph in. A user key belongs to
	// exactly one of them and the others answer it with 403 "not authorized for
	// account region", so the region cannot be discovered from the credential
	// — the customer names it.
	newRelicRegionUS = "us"
	newRelicRegionEU = "eu"
	newRelicRegionJP = "jp"

	newRelicUSEndpoint = "https://api.newrelic.com/graphql"
	newRelicEUEndpoint = "https://api.eu.newrelic.com/graphql"
	newRelicJPEndpoint = "https://api.jp.newrelic.com/graphql"

	// newRelicAdminGroup is the display name of New Relic's built-in
	// administrator group. It is matched exactly: a custom group can grant the
	// same authority under any name, so anything else reports no admin signal
	// rather than a false one.
	newRelicAdminGroup = "Admin"
)

// NewRelicEndpoint maps a region to its NerdGraph endpoint. It is the single
// place the two hosts are written down, so the driver, the name resolver and
// the connection probe cannot drift onto different regions for one connector.
func NewRelicEndpoint(region string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(region)) {
	case newRelicRegionUS:
		return newRelicUSEndpoint, nil
	case newRelicRegionEU:
		return newRelicEUEndpoint, nil
	case newRelicRegionJP:
		return newRelicJPEndpoint, nil
	default:
		return "", fmt.Errorf("unsupported new relic region")
	}
}

// NewRelicDriver lists the users of one New Relic organization, together with
// the groups each user belongs to.
//
// New Relic splits a user's identity from their authority: the user record
// carries a `type` that is the BILLING tier (Basic, Core, Full platform), not
// a role, while what a user may actually do comes from the groups they are in.
// The driver therefore reads both and reports group membership as the roles.
type NewRelicDriver struct {
	httpClient *http.Client
	endpoint   string
}

var _ Driver = (*NewRelicDriver)(nil)

// NewNewRelicDriver builds a driver against endpoint, the region's NerdGraph
// endpoint. NerdGraph is a single GraphQL endpoint, so there is no path to
// join onto it.
func NewNewRelicDriver(httpClient *http.Client, endpoint string) *NewRelicDriver {
	client := *httpClient
	client.Transport = &retryRoundTripper{
		next:       httpClient.Transport,
		maxRetries: 3,
	}

	return &NewRelicDriver{httpClient: &client, endpoint: endpoint}
}

type newRelicUser struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	LastActive *string `json:"lastActive"`
	Type       struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
	} `json:"type"`
}

type newRelicUserCollection struct {
	NextCursor *string        `json:"nextCursor"`
	Users      []newRelicUser `json:"users"`
}

type newRelicGroupMember struct {
	ID string `json:"id"`
}

type newRelicGroupMemberCollection struct {
	NextCursor *string               `json:"nextCursor"`
	Users      []newRelicGroupMember `json:"users"`
}

type newRelicGroup struct {
	ID          string                        `json:"id"`
	DisplayName string                        `json:"displayName"`
	Users       newRelicGroupMemberCollection `json:"users"`
}

type newRelicGroupCollection struct {
	NextCursor *string         `json:"nextCursor"`
	Groups     []newRelicGroup `json:"groups"`
}

type newRelicDomain struct {
	ID     string                  `json:"id"`
	Name   string                  `json:"name"`
	Users  newRelicUserCollection  `json:"users"`
	Groups newRelicGroupCollection `json:"groups"`
}

type newRelicDomainCollection struct {
	NextCursor            *string          `json:"nextCursor"`
	AuthenticationDomains []newRelicDomain `json:"authenticationDomains"`
}

// newRelicResponse is the one envelope every query in this file decodes into.
// The targeted queries fill a subset of it, which is why each collection is a
// value rather than a pointer: an absent collection reads as empty.
type newRelicResponse struct {
	Data struct {
		Actor struct {
			Organization struct {
				Name           string `json:"name"`
				UserManagement struct {
					AuthenticationDomains newRelicDomainCollection `json:"authenticationDomains"`
				} `json:"userManagement"`
			} `json:"organization"`
		} `json:"actor"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

const (
	newRelicUserFields = `id name email lastActive type { id displayName }`

	newRelicDomainsQuery = `query($domainCursor: String) {
  actor { organization { userManagement { authenticationDomains(cursor: $domainCursor) {
    nextCursor
    authenticationDomains {
      id
      name
      users { nextCursor users { ` + newRelicUserFields + ` } }
      groups { nextCursor groups { id displayName users { nextCursor users { id } } } }
    }
  } } } }
}`

	newRelicDomainUsersQuery = `query($domainId: [ID!], $userCursor: String) {
  actor { organization { userManagement { authenticationDomains(id: $domainId) {
    authenticationDomains { id users(cursor: $userCursor) { nextCursor users { ` + newRelicUserFields + ` } } }
  } } } }
}`

	newRelicDomainGroupsQuery = `query($domainId: [ID!], $groupCursor: String) {
  actor { organization { userManagement { authenticationDomains(id: $domainId) {
    authenticationDomains { id groups(cursor: $groupCursor) { nextCursor groups { id displayName users { nextCursor users { id } } } } }
  } } } }
}`

	newRelicGroupUsersQuery = `query($domainId: [ID!], $groupId: [ID!], $groupUserCursor: String) {
  actor { organization { userManagement { authenticationDomains(id: $domainId) {
    authenticationDomains { id groups(id: $groupId) { groups { id users(cursor: $groupUserCursor) { nextCursor users { id } } } } }
  } } } }
}`
)

func (d *NewRelicDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		users        []newRelicUser
		groupsByUser = map[string][]string{}
		domainCursor *string
	)

	for range maxPaginationPages {
		resp, err := d.query(ctx, newRelicDomainsQuery, map[string]any{
			"domainCursor": domainCursor,
		})
		if err != nil {
			return nil, err
		}

		collection := resp.Data.Actor.Organization.UserManagement.AuthenticationDomains

		for _, domain := range collection.AuthenticationDomains {
			domainUsers, err := d.drainDomainUsers(ctx, domain)
			if err != nil {
				return nil, err
			}

			users = append(users, domainUsers...)

			if err := d.drainDomainGroups(ctx, domain, groupsByUser); err != nil {
				return nil, err
			}
		}

		if collection.NextCursor == nil || *collection.NextCursor == "" {
			return newRelicRecords(users, groupsByUser), nil
		}

		domainCursor = collection.NextCursor
	}

	return nil, fmt.Errorf("cannot list all new relic authentication domains: %w", ErrPaginationLimitReached)
}

// drainDomainUsers returns every user of one authentication domain, starting
// from the page the domains query already returned.
func (d *NewRelicDriver) drainDomainUsers(ctx context.Context, domain newRelicDomain) ([]newRelicUser, error) {
	users := slices.Clone(domain.Users.Users)
	cursor := domain.Users.NextCursor

	for range maxPaginationPages {
		if cursor == nil || *cursor == "" {
			return users, nil
		}

		resp, err := d.query(ctx, newRelicDomainUsersQuery, map[string]any{
			"domainId":   []string{domain.ID},
			"userCursor": cursor,
		})
		if err != nil {
			return nil, err
		}

		page := newRelicFindDomain(resp, domain.ID)
		if page == nil {
			return nil, fmt.Errorf("cannot list new relic users: domain missing from its own targeted page")
		}

		users = append(users, page.Users.Users...)
		cursor = page.Users.NextCursor
	}

	return nil, fmt.Errorf("cannot list all new relic users of a domain: %w", ErrPaginationLimitReached)
}

// drainDomainGroups records every group membership of one authentication
// domain into groupsByUser, following both the group cursor and each group's
// own member cursor.
func (d *NewRelicDriver) drainDomainGroups(
	ctx context.Context,
	domain newRelicDomain,
	groupsByUser map[string][]string,
) error {
	groups := domain.Groups.Groups
	cursor := domain.Groups.NextCursor

	for range maxPaginationPages {
		for _, group := range groups {
			if err := d.drainGroupMembers(ctx, domain.ID, group, groupsByUser); err != nil {
				return err
			}
		}

		if cursor == nil || *cursor == "" {
			return nil
		}

		resp, err := d.query(ctx, newRelicDomainGroupsQuery, map[string]any{
			"domainId":    []string{domain.ID},
			"groupCursor": cursor,
		})
		if err != nil {
			return err
		}

		page := newRelicFindDomain(resp, domain.ID)
		if page == nil {
			return fmt.Errorf("cannot list new relic groups: domain missing from its own targeted page")
		}

		groups = page.Groups.Groups
		cursor = page.Groups.NextCursor
	}

	return fmt.Errorf("cannot list all new relic groups of a domain: %w", ErrPaginationLimitReached)
}

func (d *NewRelicDriver) drainGroupMembers(
	ctx context.Context,
	domainID string,
	group newRelicGroup,
	groupsByUser map[string][]string,
) error {
	name := strings.TrimSpace(group.DisplayName)
	members := group.Users.Users
	cursor := group.Users.NextCursor

	for range maxPaginationPages {
		for _, member := range members {
			if name == "" || member.ID == "" {
				continue
			}

			if !slices.Contains(groupsByUser[member.ID], name) {
				groupsByUser[member.ID] = append(groupsByUser[member.ID], name)
			}
		}

		if cursor == nil || *cursor == "" {
			return nil
		}

		resp, err := d.query(ctx, newRelicGroupUsersQuery, map[string]any{
			"domainId":        []string{domainID},
			"groupId":         []string{group.ID},
			"groupUserCursor": cursor,
		})
		if err != nil {
			return err
		}

		page := newRelicFindGroup(resp, domainID, group.ID)
		if page == nil {
			return fmt.Errorf("cannot list new relic group members: group missing from its own targeted page")
		}

		members = page.Users.Users
		cursor = page.Users.NextCursor
	}

	return fmt.Errorf("cannot list all new relic members of a group: %w", ErrPaginationLimitReached)
}

// newRelicFindDomain picks the requested domain out of a targeted response
// rather than trusting its position, so a provider that ignores the id filter
// cannot silently graft another domain's page onto this one.
//
// Its callers treat a miss as an error rather than stopping early. Both are
// failures, but a short roster is the worse one: a member who is simply absent
// from the campaign gets reviewed by nobody, while a failed sync is visible.
func newRelicFindDomain(resp *newRelicResponse, domainID string) *newRelicDomain {
	domains := resp.Data.Actor.Organization.UserManagement.AuthenticationDomains.AuthenticationDomains
	for i := range domains {
		if domains[i].ID == domainID {
			return &domains[i]
		}
	}

	return nil
}

func newRelicFindGroup(resp *newRelicResponse, domainID, groupID string) *newRelicGroup {
	domain := newRelicFindDomain(resp, domainID)
	if domain == nil {
		return nil
	}

	for i := range domain.Groups.Groups {
		if domain.Groups.Groups[i].ID == groupID {
			return &domain.Groups.Groups[i]
		}
	}

	return nil
}

func newRelicRecords(users []newRelicUser, groupsByUser map[string][]string) []AccountRecord {
	records := make([]AccountRecord, 0, len(users))

	for _, u := range users {
		email := strings.TrimSpace(u.Email)
		if email == "" {
			continue
		}

		groups := groupsByUser[u.ID]

		record := AccountRecord{
			Email:       email,
			FullName:    newRelicFullName(u, email),
			Roles:       newRelicRoles(groups),
			IsAdmin:     new(slices.ContainsFunc(groups, newRelicIsAdminGroup)),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  strings.TrimSpace(u.ID),
		}

		if u.LastActive != nil {
			record.LastLogin = parseRFC3339Ptr(*u.LastActive)
		}

		records = append(records, record)
	}

	return records
}

// newRelicIsAdminGroup matches New Relic's built-in Admin group by its exact
// name. A customer is free to create a group called "admin" that grants
// nothing, so case is a real distinction here rather than noise.
func newRelicIsAdminGroup(group string) bool {
	return group == newRelicAdminGroup
}

func newRelicFullName(u newRelicUser, fallback string) string {
	if name := strings.TrimSpace(u.Name); name != "" {
		return name
	}

	return fallback
}

// newRelicRoles reports the groups a user belongs to, which is where New Relic
// keeps what a user may do.
//
// The user type is deliberately left out. New Relic documents it as a billing
// factor and says it is "not meant to be a way to set permissions", and Roles
// on a review record means permissions — anything else here is read as granted
// authority by a reviewer and by whatever filters these records.
func newRelicRoles(groups []string) []string {
	roles := slices.Clone(groups)
	if roles == nil {
		roles = []string{}
	}

	return roles
}

func (d *NewRelicDriver) query(
	ctx context.Context,
	query string,
	variables map[string]any,
) (*newRelicResponse, error) {
	payload, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal new relic request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("cannot create new relic request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute new relic request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, &newRelicStatusError{statusCode: httpResp.StatusCode}
	}

	var resp newRelicResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode new relic response: %w", err)
	}

	// NerdGraph answers a rejected query with 200 and an errors array, so the
	// status alone never reports failure. Provider-supplied messages may carry
	// account identifiers or query fragments — never embed them.
	if len(resp.Errors) > 0 {
		return nil, errNewRelicGraphQL
	}

	return &resp, nil
}

// errNewRelicGraphQL reports a query NerdGraph answered with 200 and an errors
// array. It carries no provider text: those messages can name accounts and
// echo the query back.
var errNewRelicGraphQL = errors.New("cannot query new relic: graphql error")

// newRelicStatusError carries the status of a refused NerdGraph call so the
// name resolver can hand it to nameStatusError, the package's shared
// permanent-versus-transient table. The driver only propagates it.
type newRelicStatusError struct {
	statusCode int
}

func (e *newRelicStatusError) Error() string {
	return fmt.Sprintf("cannot query new relic: unexpected status %d", e.statusCode)
}

const newRelicOrganizationQuery = `{ actor { organization { name } } }`

// newRelicNameResolver names the source after the New Relic organization the
// user key belongs to.
type newRelicNameResolver struct {
	driver *NewRelicDriver
}

// NewNewRelicNameResolver resolves the organization name against endpoint, the
// region's NerdGraph endpoint. It reuses the driver so the query path, the
// error handling and the retry transport cannot drift from the roster's.
func NewNewRelicNameResolver(httpClient *http.Client, endpoint string) NameResolver {
	return &newRelicNameResolver{driver: NewNewRelicDriver(httpClient, endpoint)}
}

func (r *newRelicNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	resp, err := r.driver.query(ctx, newRelicOrganizationQuery, nil)
	if err != nil {
		// Retryable is the default, because most of what can go wrong here is
		// a blip: a reset, a decode failure, the worker's own deadline. Only a
		// refusal that answers the same way every time is worth retiring on
		// the first attempt.
		if statusErr, ok := errors.AsType[*newRelicStatusError](err); ok {
			return "", nameStatusError("new relic organization", statusErr.statusCode)
		}

		// A GraphQL errors array stays retryable. NerdGraph reports execution
		// failures that way as well as refusals, and the two are not
		// distinguishable without reading provider text this driver refuses to
		// read. The worker's attempt budget bounds a genuine refusal anyway,
		// where calling it terminal would name the source generically for good
		// on one bad minute.
		return "", err
	}

	return strings.TrimSpace(resp.Data.Actor.Organization.Name), nil
}
