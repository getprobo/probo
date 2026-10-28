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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	ElasticCloudKeyNotOwner = "elastic_cloud_key_not_owner"

	elasticCloudOrganizationsPath = "organizations"
	elasticCloudMembersPath       = "members"
	elasticCloudAPIKeysPath       = "users/auth/keys"

	// The only role administering the whole organization.
	elasticCloudOrganizationOwnerRole = "organization-admin"

	elasticCloudScopeOrganization = "organization"
	elasticCloudScopeDeployment   = "deployment"
	elasticCloudScopeProject      = "project"
	elasticCloudScopePlatform     = "platform"
)

var (
	// elasticCloudRoleLabels carries the names the Cloud console shows.
	// Anything absent is humanized from its identifier.
	elasticCloudRoleLabels = map[string]string{
		"organization-admin": "Organization owner",
		"billing-admin":      "Billing admin",
		"deployment-admin":   "Deployment admin",
		"deployment-editor":  "Deployment editor",
		"deployment-viewer":  "Deployment viewer",
	}

	// elasticCloudScopeOrder lists role labels by how far a scope reaches. A
	// scope Elastic adds later sorts after these, by name.
	elasticCloudScopeOrder = []string{
		elasticCloudScopeOrganization,
		elasticCloudScopeDeployment,
		elasticCloudScopeProject,
		elasticCloudScopePlatform,
	}

	// errElasticCloudNoSoleOrg marks a key that does not resolve to exactly
	// one organization. Retrying cannot change what a key belongs to.
	errElasticCloudNoSoleOrg = errors.New("elastic cloud api key does not resolve to exactly one organization")
)

type (
	// ElasticCloudDriver lists everyone holding a role in the organization an
	// Elastic Cloud API key belongs to: its members and its API keys.
	ElasticCloudDriver struct {
		httpClient *http.Client
		baseURL    string
		now        func() time.Time
	}

	// elasticCloudStatusError carries the status of a non-2xx response so the
	// name resolver can classify it while the driver reports it plainly.
	elasticCloudStatusError struct {
		what string
		code int
	}

	elasticCloudOrganization struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	elasticCloudOrganizationsResponse struct {
		Organizations []elasticCloudOrganization `json:"organizations"`
		NextPage      string                     `json:"next_page"`
	}

	elasticCloudMember struct {
		UserID          string                     `json:"user_id"`
		Name            string                     `json:"name"`
		Email           string                     `json:"email"`
		MemberSince     string                     `json:"member_since"`
		RoleAssignments map[string]json.RawMessage `json:"role_assignments"`
	}

	elasticCloudMembersResponse struct {
		Members []elasticCloudMember `json:"members"`
	}

	elasticCloudAPIKey struct {
		ID              string                     `json:"id"`
		Description     string                     `json:"description"`
		CreationDate    string                     `json:"creation_date"`
		ExpirationDate  string                     `json:"expiration_date"`
		RoleAssignments map[string]json.RawMessage `json:"role_assignments"`
	}

	elasticCloudAPIKeysResponse struct {
		Keys     []elasticCloudAPIKey `json:"keys"`
		NextPage string               `json:"next_page"`
	}

	// elasticCloudRoleAssignment is the union of every scope's assignment
	// shape: organization, deployment and the project types share role_id and
	// application_roles, and the resource scopes add all and their own ids.
	elasticCloudRoleAssignment struct {
		RoleID           string   `json:"role_id"`
		All              bool     `json:"all"`
		DeploymentIDs    []string `json:"deployment_ids"`
		ProjectIDs       []string `json:"project_ids"`
		ApplicationRoles []string `json:"application_roles"`
	}

	// elasticCloudGrant is one role with one set of application roles in one
	// scope, merged across every assignment of it. scope is "organization",
	// "deployment", "project.elasticsearch", and so on.
	elasticCloudGrant struct {
		scope            string
		roleID           string
		all              bool
		resourceIDs      []string
		applicationRoles []string
	}

	// elasticCloudNameResolver titles the source with the organization the
	// key belongs to.
	elasticCloudNameResolver struct {
		httpClient *http.Client
		baseURL    string
	}
)

var (
	_ Driver       = (*ElasticCloudDriver)(nil)
	_ NameResolver = (*elasticCloudNameResolver)(nil)
)

func (e *elasticCloudStatusError) Error() string {
	return fmt.Sprintf("cannot fetch elastic cloud %s: unexpected status %d", e.what, e.code)
}

func NewElasticCloudDriver(httpClient *http.Client, baseURL string) *ElasticCloudDriver {
	return &ElasticCloudDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		baseURL: baseURL,
		now:     time.Now,
	}
}

// ListAccounts returns the organization's members, then its API keys, which
// can hold Organization owner too. A failure on either fails the whole sync.
func (d *ElasticCloudDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	org, err := elasticCloudFetchOrganization(ctx, d.httpClient, d.baseURL)
	if err != nil {
		return nil, err
	}

	members, err := d.listMembers(ctx, org.ID)
	if err != nil {
		return nil, err
	}

	records := make([]AccountRecord, 0, len(members))
	ownerSeen := false

	for _, m := range members {
		record, err := elasticCloudMemberRecord(m)
		if err != nil {
			return nil, fmt.Errorf("cannot map elastic cloud member: %w", err)
		}

		ownerSeen = ownerSeen || *record.IsAdmin
		records = append(records, record)
	}

	// An organization always keeps an Organization owner, so a roster without
	// one comes from a key that cannot see role assignments.
	if !ownerSeen {
		return nil, &SettingRejectedError{
			Code:       ElasticCloudKeyNotOwner,
			StatusCode: http.StatusOK,
			Message:    "This Elastic Cloud API key cannot see your organization's role assignments. Create a key with the Organization owner role and update the connector with it.",
		}
	}

	apiKeys, err := d.listAPIKeys(ctx)
	if err != nil {
		return nil, err
	}

	for _, k := range apiKeys {
		record, err := elasticCloudAPIKeyRecord(k, d.now())
		if err != nil {
			return nil, fmt.Errorf("cannot map elastic cloud api key: %w", err)
		}

		records = append(records, record)
	}

	return records, nil
}

// elasticCloudFetchOrganization returns the organization the key belongs to.
// Anything other than a single organization is refused rather than guessed.
func elasticCloudFetchOrganization(
	ctx context.Context,
	httpClient *http.Client,
	baseURL string,
) (elasticCloudOrganization, error) {
	endpoint, err := url.JoinPath(baseURL, elasticCloudOrganizationsPath)
	if err != nil {
		return elasticCloudOrganization{}, fmt.Errorf("cannot build elastic cloud organizations URL: %w", err)
	}

	// The first page carries a cursor even when it holds the only
	// organization; the page it leads to is empty.
	var orgs []elasticCloudOrganization

	err = elasticCloudFollowCursor(
		endpoint,
		"organizations",
		func(pageURL string) (string, error) {
			var resp elasticCloudOrganizationsResponse
			if err := elasticCloudGetJSON(ctx, httpClient, pageURL, "organizations", &resp); err != nil {
				return "", err
			}

			orgs = append(orgs, resp.Organizations...)

			return resp.NextPage, nil
		},
	)
	if err != nil {
		return elasticCloudOrganization{}, err
	}

	if len(orgs) != 1 {
		return elasticCloudOrganization{}, fmt.Errorf(
			"cannot resolve elastic cloud organization: expected exactly 1 organization, got %d: %w",
			len(orgs),
			errElasticCloudNoSoleOrg,
		)
	}

	org := orgs[0]
	if org.ID == "" {
		return elasticCloudOrganization{}, fmt.Errorf(
			"cannot resolve elastic cloud organization: organization has no id: %w",
			errElasticCloudNoSoleOrg,
		)
	}

	org.Name = strings.TrimSpace(org.Name)

	return org, nil
}

// listMembers fetches the whole roster in one request: the endpoint takes no
// pagination parameters.
func (d *ElasticCloudDriver) listMembers(ctx context.Context, orgID string) ([]elasticCloudMember, error) {
	endpoint, err := url.JoinPath(
		d.baseURL,
		elasticCloudOrganizationsPath,
		url.PathEscape(orgID),
		elasticCloudMembersPath,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot build elastic cloud members URL: %w", err)
	}

	var resp elasticCloudMembersResponse
	if err := elasticCloudGetJSON(ctx, d.httpClient, endpoint, "organization members", &resp); err != nil {
		return nil, err
	}

	return resp.Members, nil
}

func (d *ElasticCloudDriver) listAPIKeys(ctx context.Context) ([]elasticCloudAPIKey, error) {
	endpoint, err := url.JoinPath(d.baseURL, elasticCloudAPIKeysPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build elastic cloud api keys URL: %w", err)
	}

	var keys []elasticCloudAPIKey

	err = elasticCloudFollowCursor(
		endpoint,
		"api keys",
		func(pageURL string) (string, error) {
			var resp elasticCloudAPIKeysResponse
			if err := elasticCloudGetJSON(ctx, d.httpClient, pageURL, "api keys", &resp); err != nil {
				return "", err
			}

			keys = append(keys, resp.Keys...)

			return resp.NextPage, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

// elasticCloudFollowCursor fetches endpoint, then each next_page cursor the
// previous page returned, until a page returns none. The cursor is an opaque
// token passed back as a query parameter rather than a URL, so it can never
// move the request off the configured host.
func elasticCloudFollowCursor(endpoint, what string, fetch func(pageURL string) (string, error)) error {
	base, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("cannot parse elastic cloud %s URL: %w", what, err)
	}

	cursor := ""

	for range maxPaginationPages {
		page := *base
		if cursor != "" {
			page.RawQuery = url.Values{"next_page": {cursor}}.Encode()
		}

		next, err := fetch(page.String())
		if err != nil {
			return err
		}

		if next == "" {
			return nil
		}

		if next == cursor {
			return fmt.Errorf("cannot list all elastic cloud %s: next_page cursor did not advance", what)
		}

		cursor = next
	}

	return fmt.Errorf("cannot list all elastic cloud %s: %w", what, ErrPaginationLimitReached)
}

func elasticCloudGetJSON(ctx context.Context, httpClient *http.Client, endpoint, what string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("cannot create elastic cloud %s request: %w", what, err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot execute elastic cloud %s request: %w", what, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &elasticCloudStatusError{what: what, code: resp.StatusCode}
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("cannot decode elastic cloud %s response: %w", what, err)
	}

	return nil
}

func elasticCloudMemberRecord(m elasticCloudMember) (AccountRecord, error) {
	grants, err := elasticCloudGrants(m.RoleAssignments)
	if err != nil {
		return AccountRecord{}, err
	}

	isAdmin := elasticCloudIsOwner(grants)

	return AccountRecord{
		Email:    strings.TrimSpace(m.Email),
		FullName: strings.TrimSpace(m.Name),
		Roles:    elasticCloudGrantLabels(grants),
		// No status, last login or MFA on the members API.
		Active:      nil,
		IsAdmin:     &isAdmin,
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:  m.UserID,
		CreatedAt:   parseRFC3339Ptr(m.MemberSince),
	}, nil
}

// elasticCloudAPIKeyRecord maps an organization API key. A revoked key leaves
// the list and the console drops an expired one, so a key still listed past
// its expiration date is what Active false records.
func elasticCloudAPIKeyRecord(k elasticCloudAPIKey, now time.Time) (AccountRecord, error) {
	grants, err := elasticCloudGrants(k.RoleAssignments)
	if err != nil {
		return AccountRecord{}, err
	}

	isAdmin := elasticCloudIsOwner(grants)

	active := true
	if expires := parseRFC3339Ptr(k.ExpirationDate); expires != nil && !expires.After(now) {
		active = false
	}

	return AccountRecord{
		FullName:    strings.TrimSpace(k.Description),
		Roles:       elasticCloudGrantLabels(grants),
		Active:      &active,
		IsAdmin:     &isAdmin,
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodAPIKey,
		AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
		ExternalID:  k.ID,
		CreatedAt:   parseRFC3339Ptr(k.CreationDate),
	}, nil
}

// elasticCloudGrants flattens role_assignments into grants. Every scope is
// decoded, including ones the API does not document, so none drops out.
func elasticCloudGrants(raw map[string]json.RawMessage) ([]elasticCloudGrant, error) {
	scopes := make([]string, 0, len(raw))
	for scope := range raw {
		scopes = append(scopes, scope)
	}

	slices.SortFunc(scopes, elasticCloudCompareScopes)

	var grants []elasticCloudGrant

	for _, scope := range scopes {
		value := raw[scope]

		var assignments []elasticCloudRoleAssignment
		if err := json.Unmarshal(value, &assignments); err == nil {
			grants = elasticCloudMergeGrants(grants, scope, assignments)

			continue
		}

		// project nests one list per project type.
		var nested map[string][]elasticCloudRoleAssignment
		if err := json.Unmarshal(value, &nested); err != nil {
			return nil, fmt.Errorf("cannot decode elastic cloud %s role assignments: %w", scope, err)
		}

		types := make([]string, 0, len(nested))
		for t := range nested {
			types = append(types, t)
		}

		slices.Sort(types)

		for _, t := range types {
			grants = elasticCloudMergeGrants(grants, scope+"."+t, nested[t])
		}
	}

	return grants, nil
}

func elasticCloudCompareScopes(a, b string) int {
	ia, ib := slices.Index(elasticCloudScopeOrder, a), slices.Index(elasticCloudScopeOrder, b)

	switch {
	case ia >= 0 && ib >= 0:
		return ia - ib
	case ia >= 0:
		return -1
	case ib >= 0:
		return 1
	default:
		return strings.Compare(a, b)
	}
}

// elasticCloudMergeGrants folds assignments into grants. The same role
// assigned twice in one scope is shown once, with the union of what it
// reaches, but only when both grant the same application roles: those are
// bound to the assignment's own resources.
func elasticCloudMergeGrants(
	grants []elasticCloudGrant,
	scope string,
	assignments []elasticCloudRoleAssignment,
) []elasticCloudGrant {
	for _, a := range assignments {
		roleID := strings.TrimSpace(a.RoleID)
		if roleID == "" {
			continue
		}

		applicationRoles := elasticCloudAppendUnique(nil, a.ApplicationRoles...)
		slices.Sort(applicationRoles)

		i := slices.IndexFunc(
			grants,
			func(g elasticCloudGrant) bool {
				return g.scope == scope && g.roleID == roleID && slices.Equal(g.applicationRoles, applicationRoles)
			},
		)
		if i < 0 {
			grants = append(
				grants,
				elasticCloudGrant{scope: scope, roleID: roleID, applicationRoles: applicationRoles},
			)
			i = len(grants) - 1
		}

		g := &grants[i]
		g.all = g.all || a.All
		g.resourceIDs = elasticCloudAppendUnique(g.resourceIDs, a.DeploymentIDs...)
		g.resourceIDs = elasticCloudAppendUnique(g.resourceIDs, a.ProjectIDs...)
	}

	return grants
}

// elasticCloudAppendUnique appends the trimmed, non-empty values dst does not
// already hold.
func elasticCloudAppendUnique(dst []string, values ...string) []string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" && !slices.Contains(dst, v) {
			dst = append(dst, v)
		}
	}

	return dst
}

func elasticCloudIsOwner(grants []elasticCloudGrant) bool {
	return slices.ContainsFunc(
		grants,
		func(g elasticCloudGrant) bool {
			return g.scope == elasticCloudScopeOrganization && g.roleID == elasticCloudOrganizationOwnerRole
		},
	)
}

func elasticCloudGrantLabels(grants []elasticCloudGrant) []string {
	labels := make([]string, 0, len(grants))
	for _, g := range grants {
		labels = append(labels, elasticCloudGrantLabel(g))
	}

	return labels
}

// elasticCloudGrantLabel names a role and how far it reaches:
// "Deployment admin (all deployments)", "Elasticsearch admin (2 projects)".
// Resource ids are counted rather than listed; they are opaque hashes.
func elasticCloudGrantLabel(g elasticCloudGrant) string {
	scope, projectType, _ := strings.Cut(g.scope, ".")

	var (
		name     string
		resource string
		details  []string
	)

	switch scope {
	case elasticCloudScopeOrganization:
		name = elasticCloudRoleName(g.roleID)
	case elasticCloudScopeDeployment:
		name, resource = elasticCloudRoleName(g.roleID), "deployment"
	case elasticCloudScopeProject:
		name, resource = elasticCloudRoleName(elasticCloudQualifiedRoleID(projectType, g.roleID)), "project"
	default:
		name, resource = elasticCloudRoleName(elasticCloudQualifiedRoleID(g.scope, g.roleID)), "resource"
	}

	switch {
	case resource == "":
	case g.all:
		details = append(details, "all "+resource+"s")
	case len(g.resourceIDs) == 1:
		details = append(details, "1 "+resource)
	case len(g.resourceIDs) > 1:
		details = append(details, strconv.Itoa(len(g.resourceIDs))+" "+resource+"s")
	}

	if len(g.applicationRoles) > 0 {
		details = append(details, "application roles: "+strings.Join(g.applicationRoles, ", "))
	}

	if len(details) == 0 {
		return name
	}

	return name + " (" + strings.Join(details, "; ") + ")"
}

// elasticCloudQualifiedRoleID puts the scope in front of a role id that does
// not already start with it, so a bare "admin" on Security projects reads as
// "Security admin" and not as an administrator of everything.
func elasticCloudQualifiedRoleID(scope, roleID string) string {
	if scope == "" || strings.HasPrefix(roleID, scope+"-") {
		return roleID
	}

	return scope + "-" + roleID
}

func elasticCloudRoleName(roleID string) string {
	if label, ok := elasticCloudRoleLabels[roleID]; ok {
		return label
	}

	return elasticCloudHumanizeIdentifier(roleID)
}

// elasticCloudHumanizeIdentifier reads an unmapped identifier without
// pretending to know its console name: "security-t1-analyst" becomes
// "Security t1 analyst".
func elasticCloudHumanizeIdentifier(id string) string {
	words := strings.FieldsFunc(
		id,
		func(r rune) bool {
			return r == '-' || r == '_' || r == '.'
		},
	)
	if len(words) == 0 {
		return id
	}

	words[0] = strings.ToUpper(words[0][:1]) + words[0][1:]

	return strings.Join(words, " ")
}

func NewElasticCloudNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &elasticCloudNameResolver{httpClient: httpClient, baseURL: baseURL}
}

func (r *elasticCloudNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	org, err := elasticCloudFetchOrganization(ctx, r.httpClient, r.baseURL)
	if err != nil {
		if statusErr, ok := errors.AsType[*elasticCloudStatusError](err); ok && statusErr != nil {
			return "", nameStatusError("elastic cloud "+statusErr.what, statusErr.code)
		}

		// A key reaching any number of organizations other than one keeps the
		// generic title; retrying cannot change its scope.
		if errors.Is(err, errElasticCloudNoSoleOrg) {
			return "", nil
		}

		return "", err
	}

	return org.Name, nil
}
