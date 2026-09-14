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
	"maps"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	mongoDBAtlasOrgsPath = "orgs"

	// Atlas clamps anything larger to 500 without saying so.
	mongoDBAtlasPageSize = 500

	// Atlas negotiates its resource version through the Accept media type and
	// answers 406 without one. 2025-02-19 is the first version reporting
	// orgMembershipStatus.
	mongoDBAtlasAPIVersion = "2025-02-19"

	// MongoDBAtlasAcceptHeader is exported for the connection probe, which must
	// send the same version this driver reads. Two copies could be re-versioned
	// apart silently.
	MongoDBAtlasAcceptHeader = "application/vnd.atlas." + mongoDBAtlasAPIVersion + "+json"

	mongoDBAtlasStatusActive  = "ACTIVE"
	mongoDBAtlasStatusPending = "PENDING"

	// The only organization role conferring full administrative control.
	mongoDBAtlasOrgOwnerRole = "ORG_OWNER"
)

// mongoDBAtlasReviewedStatuses is sent on the wire rather than left to the
// endpoint's default, which is MongoDB's to change. The statuses left out are
// invitations that never became access.
var mongoDBAtlasReviewedStatuses = []string{
	mongoDBAtlasStatusActive,
	mongoDBAtlasStatusPending,
}

// mongoDBAtlasRoleLabels carries the labels the Atlas console shows. It is a
// map rather than a transformation because two cannot be derived from the
// identifier: ORG_GROUP_CREATOR displays as "Project Creator" and
// ORG_BILLING_READ_ONLY as "Billing Viewer". Anything absent falls back to
// mongoDBAtlasHumanizeRole, since Atlas ships roles faster than this map.
var mongoDBAtlasRoleLabels = map[string]string{
	"ORG_OWNER":                   "Organization Owner",
	"ORG_GROUP_CREATOR":           "Organization Project Creator",
	"ORG_BILLING_ADMIN":           "Organization Billing Admin",
	"ORG_BILLING_READ_ONLY":       "Organization Billing Viewer",
	"ORG_STREAM_PROCESSING_ADMIN": "Organization Stream Processing Admin",
	"ORG_READ_ONLY":               "Organization Read Only",
	"ORG_MEMBER":                  "Organization Member",
}

type MongoDBAtlasDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*MongoDBAtlasDriver)(nil)

// errMongoDBAtlasNoSoleOrg marks a credential that does not resolve to exactly
// one organization. Retrying cannot change what a credential is scoped to.
var errMongoDBAtlasNoSoleOrg = errors.New("mongodb atlas credential does not resolve to exactly one organization")

// mongoDBAtlasStatusError carries the status of a non-2xx response so the name
// resolver can classify it with nameStatusError while the driver reports it
// plainly.
type mongoDBAtlasStatusError struct {
	what string
	code int
}

func (e *mongoDBAtlasStatusError) Error() string {
	return fmt.Sprintf("cannot fetch mongodb atlas %s: unexpected status %d", e.what, e.code)
}

type mongoDBAtlasOrg struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type mongoDBAtlasOrgsResponse struct {
	Results []mongoDBAtlasOrg `json:"results"`
}

type mongoDBAtlasGroupRoleAssignment struct {
	GroupID    string   `json:"groupId"`
	GroupRoles []string `json:"groupRoles"`
}

type mongoDBAtlasRoleAssignments struct {
	OrgRoles             []string                          `json:"orgRoles"`
	GroupRoleAssignments []mongoDBAtlasGroupRoleAssignment `json:"groupRoleAssignments"`
}

// mongoDBAtlasUser is the union of the ACTIVE and pending variants of
// OrgUserResponse. firstName, lastName and createdAt belong to the ACTIVE
// variant alone; invitationCreatedAt to the pending one.
type mongoDBAtlasUser struct {
	ID                  string                      `json:"id"`
	Username            string                      `json:"username"`
	FirstName           string                      `json:"firstName"`
	LastName            string                      `json:"lastName"`
	OrgMembershipStatus string                      `json:"orgMembershipStatus"`
	CreatedAt           string                      `json:"createdAt"`
	InvitationCreatedAt string                      `json:"invitationCreatedAt"`
	LastAuth            string                      `json:"lastAuth"`
	Roles               mongoDBAtlasRoleAssignments `json:"roles"`
}

type mongoDBAtlasUsersResponse struct {
	Results []mongoDBAtlasUser `json:"results"`
}

type mongoDBAtlasServiceAccount struct {
	ClientID  string   `json:"clientId"`
	Name      string   `json:"name"`
	CreatedAt string   `json:"createdAt"`
	Roles     []string `json:"roles"`
	Secrets   []struct {
		LastUsedAt string `json:"lastUsedAt"`
	} `json:"secrets"`
}

type mongoDBAtlasServiceAccountsResponse struct {
	Results []mongoDBAtlasServiceAccount `json:"results"`
}

type mongoDBAtlasAPIKeyRole struct {
	OrgID    string `json:"orgId"`
	GroupID  string `json:"groupId"`
	RoleName string `json:"roleName"`
}

// mongoDBAtlasAPIKey is a programmatic API key, the older non-human principal.
// It exposes no creation or last-use timestamp and no enabled state.
type mongoDBAtlasAPIKey struct {
	ID        string                   `json:"id"`
	Desc      string                   `json:"desc"`
	PublicKey string                   `json:"publicKey"`
	Roles     []mongoDBAtlasAPIKeyRole `json:"roles"`
}

type mongoDBAtlasAPIKeysResponse struct {
	Results []mongoDBAtlasAPIKey `json:"results"`
}

func NewMongoDBAtlasDriver(httpClient *http.Client, baseURL string) *MongoDBAtlasDriver {
	return &MongoDBAtlasDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		baseURL: baseURL,
	}
}

// ListAccounts returns every principal holding organization-level access: the
// human members, the service accounts and the programmatic API keys. All three
// carry organization roles, and all three can hold ORG_OWNER.
//
// A failure on any of them fails the whole sync: a roster silently missing one
// population reads as a complete one.
func (d *MongoDBAtlasDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	org, err := d.fetchOrganization(ctx)
	if err != nil {
		return nil, err
	}

	users, err := d.listUsers(ctx, org.ID)
	if err != nil {
		return nil, err
	}

	serviceAccounts, err := d.listServiceAccounts(ctx, org.ID)
	if err != nil {
		return nil, err
	}

	apiKeys, err := d.listAPIKeys(ctx, org.ID)
	if err != nil {
		return nil, err
	}

	records := make([]AccountRecord, 0, len(users)+len(serviceAccounts)+len(apiKeys))

	for _, u := range users {
		records = append(records, mongoDBAtlasUserRecord(u))
	}

	for _, sa := range serviceAccounts {
		records = append(records, mongoDBAtlasServiceAccountRecord(sa))
	}

	for _, k := range apiKeys {
		records = append(records, mongoDBAtlasAPIKeyRecord(k))
	}

	return records, nil
}

// fetchOrganization returns the organization the credential belongs to. An
// Atlas service account is scoped to exactly one, so anything other than a
// single result means the credential is not the one this connector is built
// for, and guessing which organization to review would be worse than refusing.
func (d *MongoDBAtlasDriver) fetchOrganization(ctx context.Context) (mongoDBAtlasOrg, error) {
	return mongoDBAtlasFetchOrganization(ctx, d.httpClient, d.baseURL)
}

func mongoDBAtlasFetchOrganization(
	ctx context.Context,
	httpClient *http.Client,
	baseURL string,
) (mongoDBAtlasOrg, error) {
	endpoint, err := url.JoinPath(baseURL, mongoDBAtlasOrgsPath)
	if err != nil {
		return mongoDBAtlasOrg{}, fmt.Errorf("cannot build mongodb atlas organizations URL: %w", err)
	}

	var resp mongoDBAtlasOrgsResponse
	if err := mongoDBAtlasGetJSON(ctx, httpClient, endpoint, "organizations", &resp); err != nil {
		return mongoDBAtlasOrg{}, err
	}

	if len(resp.Results) != 1 {
		return mongoDBAtlasOrg{}, fmt.Errorf(
			"cannot resolve mongodb atlas organization: expected exactly 1 organization, got %d: %w",
			len(resp.Results),
			errMongoDBAtlasNoSoleOrg,
		)
	}

	org := resp.Results[0]
	if org.ID == "" {
		return mongoDBAtlasOrg{}, fmt.Errorf(
			"cannot resolve mongodb atlas organization: organization has no id: %w",
			errMongoDBAtlasNoSoleOrg,
		)
	}

	org.Name = strings.TrimSpace(org.Name)

	return org, nil
}

func (d *MongoDBAtlasDriver) listUsers(ctx context.Context, orgID string) ([]mongoDBAtlasUser, error) {
	var users []mongoDBAtlasUser

	for page := 1; page <= maxPaginationPages; page++ {
		endpoint, err := mongoDBAtlasCollectionURL(
			d.baseURL,
			orgID,
			"users",
			page,
			url.Values{"orgMembershipStatuses": mongoDBAtlasReviewedStatuses},
		)
		if err != nil {
			return nil, err
		}

		var resp mongoDBAtlasUsersResponse
		if err := mongoDBAtlasGetJSON(ctx, d.httpClient, endpoint, "organization users", &resp); err != nil {
			return nil, err
		}

		// Terminate on an empty page, not on a short one and not on
		// totalCount: itemsPerPage clamps silently, and the schema documents
		// totalCount as an estimate. Either would truncate the roster.
		if len(resp.Results) == 0 {
			return users, nil
		}

		users = append(users, resp.Results...)
	}

	return nil, fmt.Errorf("cannot list all mongodb atlas organization users: %w", ErrPaginationLimitReached)
}

func (d *MongoDBAtlasDriver) listServiceAccounts(ctx context.Context, orgID string) ([]mongoDBAtlasServiceAccount, error) {
	var serviceAccounts []mongoDBAtlasServiceAccount

	for page := 1; page <= maxPaginationPages; page++ {
		endpoint, err := mongoDBAtlasCollectionURL(d.baseURL, orgID, "serviceAccounts", page, nil)
		if err != nil {
			return nil, err
		}

		var resp mongoDBAtlasServiceAccountsResponse
		if err := mongoDBAtlasGetJSON(ctx, d.httpClient, endpoint, "organization service accounts", &resp); err != nil {
			return nil, err
		}

		if len(resp.Results) == 0 {
			return serviceAccounts, nil
		}

		serviceAccounts = append(serviceAccounts, resp.Results...)
	}

	return nil, fmt.Errorf("cannot list all mongodb atlas organization service accounts: %w", ErrPaginationLimitReached)
}

func (d *MongoDBAtlasDriver) listAPIKeys(ctx context.Context, orgID string) ([]mongoDBAtlasAPIKey, error) {
	var apiKeys []mongoDBAtlasAPIKey

	for page := 1; page <= maxPaginationPages; page++ {
		endpoint, err := mongoDBAtlasCollectionURL(d.baseURL, orgID, "apiKeys", page, nil)
		if err != nil {
			return nil, err
		}

		var resp mongoDBAtlasAPIKeysResponse
		if err := mongoDBAtlasGetJSON(ctx, d.httpClient, endpoint, "organization api keys", &resp); err != nil {
			return nil, err
		}

		if len(resp.Results) == 0 {
			return apiKeys, nil
		}

		apiKeys = append(apiKeys, resp.Results...)
	}

	return nil, fmt.Errorf("cannot list all mongodb atlas organization api keys: %w", ErrPaginationLimitReached)
}

func mongoDBAtlasCollectionURL(baseURL, orgID, collection string, page int, extra url.Values) (string, error) {
	endpoint, err := url.JoinPath(baseURL, mongoDBAtlasOrgsPath, url.PathEscape(orgID), collection)
	if err != nil {
		return "", fmt.Errorf("cannot build mongodb atlas %s URL: %w", collection, err)
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("cannot parse mongodb atlas %s URL: %w", collection, err)
	}

	query := url.Values{
		"pageNum":      {strconv.Itoa(page)},
		"itemsPerPage": {strconv.Itoa(mongoDBAtlasPageSize)},
	}

	maps.Copy(query, extra)

	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

// mongoDBAtlasGetJSON issues one authenticated GET and decodes the body. The
// connection transport supplies the bearer token; what this adds is the
// versioned Accept header Atlas requires on every request.
func mongoDBAtlasGetJSON(ctx context.Context, httpClient *http.Client, endpoint, what string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("cannot create mongodb atlas %s request: %w", what, err)
	}

	req.Header.Set("Accept", MongoDBAtlasAcceptHeader)

	httpResp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot execute mongodb atlas %s request: %w", what, err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return &mongoDBAtlasStatusError{what: what, code: httpResp.StatusCode}
	}

	if err := json.NewDecoder(httpResp.Body).Decode(out); err != nil {
		return fmt.Errorf("cannot decode mongodb atlas %s response: %w", what, err)
	}

	return nil
}

func mongoDBAtlasUserRecord(u mongoDBAtlasUser) AccountRecord {
	active := u.OrgMembershipStatus == mongoDBAtlasStatusActive
	isAdmin := mongoDBAtlasHasOrgOwner(u.Roles.OrgRoles)

	// A pending invitation carries invitationCreatedAt in place of createdAt
	// and no name at all. The name stays empty rather than borrowing the
	// email, which would read as a real one.
	createdAt := u.CreatedAt
	if createdAt == "" {
		createdAt = u.InvitationCreatedAt
	}

	return AccountRecord{
		// Atlas has no separate email field on this version: username is the
		// email, typed as format "email" in the schema.
		Email:       u.Username,
		FullName:    mongoDBAtlasFullName(u.FirstName, u.LastName),
		Roles:       mongoDBAtlasUserRoles(u),
		Active:      &active,
		IsAdmin:     &isAdmin,
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:  u.ID,
		LastLogin:   parseRFC3339Ptr(u.LastAuth),
		CreatedAt:   parseRFC3339Ptr(createdAt),
	}
}

func mongoDBAtlasServiceAccountRecord(sa mongoDBAtlasServiceAccount) AccountRecord {
	isAdmin := mongoDBAtlasHasOrgOwner(sa.Roles)

	return AccountRecord{
		Email:    "",
		FullName: strings.TrimSpace(sa.Name),
		Roles:    mongoDBAtlasRoles(sa.Roles),
		// Atlas states no enabled/disabled status for a service account, and
		// secret expiry would be an inference rather than a status: an expired
		// secret can be replaced without the principal ever being disabled.
		// Note the entry is still stored active — coredata COALESCEs a nil
		// Active to TRUE — so this records the absence of a signal, it does not
		// make the row read as inactive.
		Active:      nil,
		IsAdmin:     &isAdmin,
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodServiceAccount,
		AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
		ExternalID:  sa.ClientID,
		LastLogin:   mongoDBAtlasLastSecretUse(sa),
		CreatedAt:   parseRFC3339Ptr(sa.CreatedAt),
	}
}

func mongoDBAtlasAPIKeyRecord(k mongoDBAtlasAPIKey) AccountRecord {
	roleNames := make([]string, 0, len(k.Roles))
	for _, r := range k.Roles {
		roleNames = append(roleNames, r.RoleName)
	}

	isAdmin := mongoDBAtlasHasOrgOwner(roleNames)

	// desc is what the console labels the key with; the public half of the key
	// identifies it when the description is blank.
	name := strings.TrimSpace(k.Desc)
	if name == "" {
		name = strings.TrimSpace(k.PublicKey)
	}

	return AccountRecord{
		Email:    "",
		FullName: name,
		Roles:    mongoDBAtlasRoles(roleNames),
		// Atlas exposes no enabled state, no creation date and no last use for
		// a programmatic API key. Active nil is stored as TRUE, as above.
		Active:      nil,
		IsAdmin:     &isAdmin,
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodAPIKey,
		AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
		ExternalID:  k.ID,
	}
}

// mongoDBAtlasLastSecretUse reports the most recent use of any of the service
// account's secrets, the closest thing Atlas offers to a last login for a
// non-human principal. A secret never used carries an empty lastUsedAt and is
// skipped rather than counted as an ancient one.
func mongoDBAtlasLastSecretUse(sa mongoDBAtlasServiceAccount) *time.Time {
	var latest *time.Time

	for _, secret := range sa.Secrets {
		used := parseRFC3339Ptr(secret.LastUsedAt)
		if used == nil {
			continue
		}

		if latest == nil || used.After(*latest) {
			latest = used
		}
	}

	return latest
}

// mongoDBAtlasUserRoles lists the organization roles, then the distinct project
// roles held anywhere in the organization. Project roles are included because
// an Organization Member who owns the production project holds far more access
// than the organization role alone suggests. They carry no project name:
// naming every project would cost one request per project.
func mongoDBAtlasUserRoles(u mongoDBAtlasUser) []string {
	roles := mongoDBAtlasRoles(u.Roles.OrgRoles)
	seen := make(map[string]bool, len(roles))

	for _, r := range roles {
		seen[r] = true
	}

	for _, assignment := range u.Roles.GroupRoleAssignments {
		for _, raw := range assignment.GroupRoles {
			label := mongoDBAtlasRoleLabel(raw)
			if label == "" || seen[label] {
				continue
			}

			seen[label] = true

			roles = append(roles, label)
		}
	}

	return roles
}

func mongoDBAtlasRoles(raw []string) []string {
	roles := make([]string, 0, len(raw))
	seen := make(map[string]bool, len(raw))

	for _, r := range raw {
		label := mongoDBAtlasRoleLabel(r)
		if label == "" || seen[label] {
			continue
		}

		seen[label] = true

		roles = append(roles, label)
	}

	return roles
}

func mongoDBAtlasRoleLabel(raw string) string {
	role := strings.TrimSpace(raw)
	if role == "" {
		return ""
	}

	if label, ok := mongoDBAtlasRoleLabels[role]; ok {
		return label
	}

	return mongoDBAtlasHumanizeRole(role)
}

// mongoDBAtlasHumanizeRole reads an unmapped role identifier without pretending
// to know its console label: GROUP_DATA_ACCESS_ADMIN becomes "Project Data
// Access Admin". A "group" in the API is what the console calls a project.
func mongoDBAtlasHumanizeRole(role string) string {
	scope := ""

	switch {
	case strings.HasPrefix(role, "ORG_"):
		scope, role = "Organization", strings.TrimPrefix(role, "ORG_")
	case strings.HasPrefix(role, "GROUP_"):
		scope, role = "Project", strings.TrimPrefix(role, "GROUP_")
	}

	words := strings.Split(role, "_")
	for i, w := range words {
		if w == "" {
			continue
		}

		words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
	}

	return strings.TrimSpace(scope + " " + strings.Join(words, " "))
}

func mongoDBAtlasHasOrgOwner(roles []string) bool {
	for _, r := range roles {
		if strings.EqualFold(strings.TrimSpace(r), mongoDBAtlasOrgOwnerRole) {
			return true
		}
	}

	return false
}

func mongoDBAtlasFullName(firstName, lastName string) string {
	return strings.TrimSpace(strings.TrimSpace(firstName) + " " + strings.TrimSpace(lastName))
}

// mongoDBAtlasNameResolver titles the source with the organization the
// credential belongs to, reusing the same GET /orgs the driver resolves the
// organization id from.
type mongoDBAtlasNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

var _ NameResolver = (*mongoDBAtlasNameResolver)(nil)

func NewMongoDBAtlasNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &mongoDBAtlasNameResolver{httpClient: httpClient, baseURL: baseURL}
}

func (r *mongoDBAtlasNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	org, err := mongoDBAtlasFetchOrganization(ctx, r.httpClient, r.baseURL)
	if err != nil {
		// A 4xx is the credential's own answer; nameStatusError retires the
		// resolver on those and leaves everything else retryable.
		if statusErr, ok := errors.AsType[*mongoDBAtlasStatusError](err); ok && statusErr != nil {
			return "", nameStatusError("mongodb atlas "+statusErr.what, statusErr.code)
		}

		// A credential reaching any number of organizations other than one
		// keeps the generic title; retrying cannot change its scope.
		if errors.Is(err, errMongoDBAtlasNoSoleOrg) {
			return "", nil
		}

		// Anything else (transport, decode) may succeed on the next attempt.
		return "", err
	}

	return org.Name, nil
}
