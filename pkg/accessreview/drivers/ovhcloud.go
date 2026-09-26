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
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"go.probo.inc/probo/pkg/coredata"
)

// maxOVHcloudCollection bounds every per-item fan-out: OVHcloud lists
// collections as bare identifiers, so each entry costs one request.
const maxOVHcloudCollection = 2000

const ovhcloudFanout = 8

type (
	// ovhcloudUser is auth.User.
	ovhcloudUser struct {
		Login       string     `json:"login"`
		Email       string     `json:"email"`
		Description string     `json:"description"`
		Group       string     `json:"group"`
		Groups      []string   `json:"groups"`
		Status      string     `json:"status"`
		Type        string     `json:"type"`
		URN         string     `json:"urn"`
		Creation    *time.Time `json:"creation"`
	}

	// ovhcloudGroup is auth.Group. Role is the privilege the group confers.
	ovhcloudGroup struct {
		Name string `json:"name"`
		Role string `json:"role"`
	}

	// ovhcloudAccount is the subset of nichandle.Nichandle describing the
	// account owner, who is not a local user and so never appears in
	// /me/identity/user.
	ovhcloudAccount struct {
		Nichandle string `json:"nichandle"`
		Email     string `json:"email"`
		Firstname string `json:"firstname"`
		Name      string `json:"name"`
	}

	// ovhcloudOAuth2Client is oauth2.client. Identity is the IAM URN a
	// client-credentials client is bound to, and is null for an
	// authorization-code client, which acts as whoever authorises it.
	ovhcloudOAuth2Client struct {
		ClientID string  `json:"clientId"`
		Name     string  `json:"name"`
		Flow     string  `json:"flow"`
		Identity *string `json:"identity"`
	}

	// ovhcloudAPIApplication is api.Application, the registration a classic
	// API credential belongs to. It carries the only human-readable name a
	// credential has.
	ovhcloudAPIApplication struct {
		ApplicationID int64  `json:"applicationId"`
		Name          string `json:"name"`
		Description   string `json:"description"`
		Status        string `json:"status"`
	}

	// ovhcloudAPICredential is api.Credential: one issued consumer key, which
	// is standing API access to the account independent of any IAM identity.
	ovhcloudAPICredential struct {
		CredentialID  int64      `json:"credentialId"`
		ApplicationID int64      `json:"applicationId"`
		Status        string     `json:"status"`
		Creation      *time.Time `json:"creation"`
		Expiration    *time.Time `json:"expiration"`
		LastUse       *time.Time `json:"lastUse"`
		// OVHSupport marks a credential OVHcloud's own support team created
		// rather than the customer, which is third-party access a review has
		// to surface.
		OVHSupport bool `json:"ovhSupport"`
		Rules      []struct {
			Method string `json:"method"`
			Path   string `json:"path"`
		} `json:"rules"`
	}

	// ovhcloudAuditLog is audit.Log, narrowed to the LOGIN_SUCCESS fields.
	ovhcloudAuditLog struct {
		CreatedAt   time.Time `json:"createdAt"`
		Type        string    `json:"type"`
		AuthDetails *struct {
			UserDetails *struct {
				Type string  `json:"type"`
				User *string `json:"user"`
			} `json:"userDetails"`
		} `json:"authDetails"`
		LoginSuccessDetails *struct {
			MFAType string `json:"mfaType"`
		} `json:"loginSuccessDetails"`
	}

	// ovhcloudSignInKey identifies one authenticated identity. The kind is
	// part of the key because a federated identity and a local user can carry
	// the same login, and merging them would attribute one's MFA evidence to
	// the other.
	ovhcloudSignInKey struct {
		kind  string
		login string
	}

	// ovhcloudSignIn is what the audit log reveals about one identity.
	ovhcloudSignIn struct {
		lastLogin time.Time
		mfa       coredata.MFAStatus
		// kind is audit.LogAuthUserTypeEnum: ACCOUNT, USER or PROVIDER. It is
		// the only evidence of HOW an identity authenticated.
		kind string
	}

	OVHcloudDriver struct {
		httpClient *http.Client
		baseURL    string
	}
)

var _ Driver = (*OVHcloudDriver)(nil)

// NewOVHcloudDriver builds a driver against baseURL, the OVHcloud API root for
// the account's region (e.g. https://eu.api.ovh.com/1.0).
func NewOVHcloudDriver(httpClient *http.Client, baseURL string) *OVHcloudDriver {
	return &OVHcloudDriver{
		httpClient: &http.Client{
			Transport: &retryRoundTripper{
				next:       httpClient.Transport,
				maxRetries: 3,
			},
		},
		baseURL: baseURL,
	}
}

// ListAccounts returns every identity that can reach the OVHcloud account.
// Five endpoints are merged because no single one holds the roster:
// /me/identity/user (local users), /me (the owner, absent from the user list),
// /me/api/oauth2/client, /me/api/credential (classic keys, invisible to IAM)
// and /me/logs/audit.
//
// The roster is a floor, not a census: a federated (SSO) user appears only in
// the audit log, so one who has not signed in within its retention is reachable
// by no endpoint.
func (d *OVHcloudDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	roles, err := d.fetchGroupRoles(ctx)
	if err != nil {
		return nil, err
	}

	owner, err := d.fetchOwner(ctx)
	if err != nil {
		return nil, err
	}

	signIns, err := d.fetchSignIns(ctx, owner.Nichandle)
	if err != nil {
		return nil, err
	}

	users, err := d.fetchUsers(ctx)
	if err != nil {
		return nil, err
	}

	serviceAccounts, err := d.fetchServiceAccounts(ctx)
	if err != nil {
		return nil, err
	}

	apiCredentials, err := d.fetchAPICredentials(ctx)
	if err != nil {
		return nil, err
	}

	records := make([]AccountRecord, 0, len(users)+len(serviceAccounts)+len(apiCredentials)+1)
	records = append(records, ovhcloudOwnerRecord(owner, signIns[ovhcloudOwnerSignInKey]))

	for _, u := range users {
		records = append(records, ovhcloudUserRecord(u, roles, signIns[ovhcloudSignInKey{kind: "USER", login: u.Login}]))
	}

	for _, sa := range serviceAccounts {
		records = append(records, ovhcloudServiceAccountRecord(sa))
	}

	records = append(records, apiCredentials...)

	// Federated identities exist only as audit evidence. Sorted, because ranging
	// a map would reorder the roster on every sync.
	federated := make([]string, 0, len(signIns))

	for key := range signIns {
		if key.kind == "PROVIDER" {
			federated = append(federated, key.login)
		}
	}

	slices.Sort(federated)

	for _, login := range federated {
		records = append(records, ovhcloudFederatedRecord(login, signIns[ovhcloudSignInKey{kind: "PROVIDER", login: login}]))
	}

	return records, nil
}

// ovhcloudOwnerSignInKey is the audit-log bucket for the account owner. Owner
// sign-ins carry userDetails.type ACCOUNT with a null user, so they have no
// login to key on.
var ovhcloudOwnerSignInKey = ovhcloudSignInKey{kind: "ACCOUNT"}

func (d *OVHcloudDriver) fetchOwner(ctx context.Context) (ovhcloudAccount, error) {
	var account ovhcloudAccount
	if err := d.get(ctx, &account, "me"); err != nil {
		return ovhcloudAccount{}, fmt.Errorf("cannot fetch ovhcloud account: %w", err)
	}

	return account, nil
}

// ovhcloudFetchEach resolves one detail document per identifier, bounded and
// cancelling its siblings on the first failure: a partial roster must never be
// returned as a whole one.
func ovhcloudFetchEach[T any](ctx context.Context, d *OVHcloudDriver, noun string, ids []string, segments ...string) ([]T, error) {
	if len(ids) > maxOVHcloudCollection {
		return nil, fmt.Errorf("cannot list ovhcloud %s: %d exceeds the supported maximum of %d", noun, len(ids), maxOVHcloudCollection)
	}

	out := make([]T, len(ids))

	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(ovhcloudFanout)

	for i, id := range ids {
		group.Go(func() error {
			path := append(slices.Clone(segments), id)

			var v T
			if err := d.get(groupCtx, &v, path...); err != nil {
				return fmt.Errorf("cannot fetch ovhcloud %s: %w", noun, err)
			}

			out[i] = v

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return out, nil
}

// ovhcloudFetchEachSub is ovhcloudFetchEach for a sub-resource hanging off each
// identifier, e.g. /me/api/credential/{id}/application.
func ovhcloudFetchEachSub[T any](ctx context.Context, d *OVHcloudDriver, noun, sub string, ids []string, segments ...string) ([]T, error) {
	if len(ids) > maxOVHcloudCollection {
		return nil, fmt.Errorf("cannot list ovhcloud %s: %d exceeds the supported maximum of %d", noun, len(ids), maxOVHcloudCollection)
	}

	out := make([]T, len(ids))

	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(ovhcloudFanout)

	for i, id := range ids {
		group.Go(func() error {
			path := append(slices.Clone(segments), id, sub)

			var v T
			if err := d.get(groupCtx, &v, path...); err != nil {
				return fmt.Errorf("cannot fetch ovhcloud %s: %w", noun, err)
			}

			out[i] = v

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return out, nil
}

func (d *OVHcloudDriver) fetchUsers(ctx context.Context) ([]ovhcloudUser, error) {
	var logins []string
	if err := d.get(ctx, &logins, "me", "identity", "user"); err != nil {
		return nil, fmt.Errorf("cannot list ovhcloud identity users: %w", err)
	}

	users, err := ovhcloudFetchEach[ovhcloudUser](ctx, d, "identity users", logins, "me", "identity", "user")
	if err != nil {
		return nil, err
	}

	// The detail payload omits the login it was addressed by.
	for i := range users {
		users[i].Login = logins[i]
	}

	return users, nil
}

// fetchServiceAccounts lists the account's OAuth2 clients, keeping only
// CLIENT_CREDENTIALS: an authorization-code client binds to no identity and
// acts as whoever authorises it.
func (d *OVHcloudDriver) fetchServiceAccounts(ctx context.Context) ([]ovhcloudOAuth2Client, error) {
	var ids []string
	if err := d.get(ctx, &ids, "me", "api", "oauth2", "client"); err != nil {
		return nil, fmt.Errorf("cannot list ovhcloud oauth2 clients: %w", err)
	}

	clients, err := ovhcloudFetchEach[ovhcloudOAuth2Client](ctx, d, "oauth2 clients", ids, "me", "api", "oauth2", "client")
	if err != nil {
		return nil, err
	}

	return slices.DeleteFunc(clients, func(c ovhcloudOAuth2Client) bool {
		return c.Flow != "CLIENT_CREDENTIALS"
	}), nil
}

// fetchAPICredentials lists the classic API credentials issued on the account.
// A credential carries no name, so the application it was granted to supplies
// one. The application is resolved per credential rather than from
// /me/api/application, because a key is commonly granted to an application
// another account registered, which that list does not contain.
func (d *OVHcloudDriver) fetchAPICredentials(ctx context.Context) ([]AccountRecord, error) {
	var ids []int64
	if err := d.get(ctx, &ids, "me", "api", "credential"); err != nil {
		return nil, fmt.Errorf("cannot list ovhcloud api credentials: %w", err)
	}

	if len(ids) == 0 {
		return nil, nil
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = strconv.FormatInt(id, 10)
	}

	credentials, err := ovhcloudFetchEach[ovhcloudAPICredential](ctx, d, "api credentials", keys, "me", "api", "credential")
	if err != nil {
		return nil, err
	}

	applications, err := ovhcloudFetchEachSub[ovhcloudAPIApplication](ctx, d, "api credential applications", "application", keys, "me", "api", "credential")
	if err != nil {
		return nil, err
	}

	records := make([]AccountRecord, 0, len(credentials))
	for i, c := range credentials {
		records = append(records, ovhcloudAPICredentialRecord(c, applications[i], time.Now()))
	}

	return records, nil
}

// fetchGroupRoles maps each group name to the role it confers.
func (d *OVHcloudDriver) fetchGroupRoles(ctx context.Context) (map[string]string, error) {
	var names []string
	if err := d.get(ctx, &names, "me", "identity", "group"); err != nil {
		return nil, fmt.Errorf("cannot list ovhcloud identity groups: %w", err)
	}

	groups, err := ovhcloudFetchEach[ovhcloudGroup](ctx, d, "identity groups", names, "me", "identity", "group")
	if err != nil {
		return nil, err
	}

	roles := make(map[string]string, len(groups))
	for i, g := range groups {
		roles[names[i]] = g.Role
	}

	return roles, nil
}

// fetchSignIns reduces the audit log to the most recent successful sign-in per
// identity. The MFA reading is what the login actually used, not what the
// identity has enrolled — OVHcloud exposes no per-user enrolment state.
func (d *OVHcloudDriver) fetchSignIns(ctx context.Context, nichandle string) (map[ovhcloudSignInKey]ovhcloudSignIn, error) {
	var logs []ovhcloudAuditLog
	if err := d.get(ctx, &logs, "me", "logs", "audit"); err != nil {
		return nil, fmt.Errorf("cannot fetch ovhcloud audit log: %w", err)
	}

	signIns := make(map[ovhcloudSignInKey]ovhcloudSignIn)

	for _, entry := range logs {
		if entry.Type != "LOGIN_SUCCESS" || entry.AuthDetails == nil || entry.AuthDetails.UserDetails == nil {
			continue
		}

		details := entry.AuthDetails.UserDetails

		var key ovhcloudSignInKey

		switch details.Type {
		case "ACCOUNT":
			key = ovhcloudOwnerSignInKey
		case "USER", "PROVIDER":
			if details.User == nil || *details.User == "" {
				continue
			}

			key = ovhcloudSignInKey{kind: details.Type, login: ovhcloudAuditLogin(*details.User, nichandle)}
		default:
			continue
		}

		if existing, ok := signIns[key]; ok && !entry.CreatedAt.After(existing.lastLogin) {
			continue
		}

		mfa := coredata.MFAStatusUnknown
		if entry.LoginSuccessDetails != nil {
			mfa = ovhcloudMFAStatus(entry.LoginSuccessDetails.MFAType)
		}

		signIns[key] = ovhcloudSignIn{lastLogin: entry.CreatedAt, mfa: mfa, kind: details.Type}
	}

	return signIns, nil
}

// ovhcloudAuditLogin maps an audit identity onto the login /me/identity/user is
// keyed by: auth.User.login is the SUFFIX, but a local user signs in as
// "<nichandle>/<suffix>", so strip the handle when present.
func ovhcloudAuditLogin(login, nichandle string) string {
	if nichandle == "" {
		return login
	}

	if suffix, ok := strings.CutPrefix(login, nichandle+"/"); ok {
		return suffix
	}

	return login
}

// ovhcloudMFAStatus reads audit.LogAuthMFATypeEnum. NONE means the sign-in used
// no second factor, which is as close as OVHcloud comes to reporting unprotected.
func ovhcloudMFAStatus(mfaType string) coredata.MFAStatus {
	switch mfaType {
	case "TOTP", "U2F", "SMS", "BACKUP_CODE", "MAIL":
		return coredata.MFAStatusEnabled
	case "NONE":
		return coredata.MFAStatusDisabled
	default:
		return coredata.MFAStatusUnknown
	}
}

func ovhcloudOwnerRecord(account ovhcloudAccount, signIn ovhcloudSignIn) AccountRecord {
	active, isAdmin := true, true

	record := AccountRecord{
		Email:       account.Email,
		FullName:    strings.TrimSpace(account.Firstname + " " + account.Name),
		Roles:       []string{"Account owner"},
		Active:      &active,
		IsAdmin:     &isAdmin,
		MFAStatus:   signIn.mfa,
		AuthMethod:  ovhcloudAuthMethod(signIn),
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:  account.Nichandle,
	}

	if record.MFAStatus == "" {
		record.MFAStatus = coredata.MFAStatusUnknown
	}

	if !signIn.lastLogin.IsZero() {
		record.LastLogin = &signIn.lastLogin
	}

	return record
}

func ovhcloudUserRecord(u ovhcloudUser, roles map[string]string, signIn ovhcloudSignIn) AccountRecord {
	// DISABLED is the only status that denies access; PASSWORD_CHANGE_REQUIRED
	// still authenticates once the password is rotated.
	active := u.Status != "DISABLED"

	record := AccountRecord{
		Email:       u.Email,
		JobTitle:    u.Description,
		Roles:       ovhcloudRoles(u),
		Active:      &active,
		IsAdmin:     ovhcloudIsAdmin(u, roles),
		MFAStatus:   signIn.mfa,
		AuthMethod:  ovhcloudAuthMethod(signIn),
		AccountType: ovhcloudAccountType(u.Type),
		CreatedAt:   u.Creation,
		ExternalID:  u.URN,
	}

	if record.MFAStatus == "" {
		record.MFAStatus = coredata.MFAStatusUnknown
	}

	if record.ExternalID == "" {
		record.ExternalID = u.Login
	}

	if !signIn.lastLogin.IsZero() {
		record.LastLogin = &signIn.lastLogin
	}

	return record
}

// ovhcloudServiceAccountRecord describes a client-credentials OAuth2 client.
// Its privilege lives in IAM policy the API does not expose per client, so
// IsAdmin stays unknown rather than false.
func ovhcloudServiceAccountRecord(c ovhcloudOAuth2Client) AccountRecord {
	active := true

	record := AccountRecord{
		FullName:    c.Name,
		Roles:       []string{},
		Active:      &active,
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodServiceAccount,
		AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
		ExternalID:  c.ClientID,
	}

	if c.Identity != nil && *c.Identity != "" {
		record.ExternalID = *c.Identity
	}

	return record
}

// ovhcloudFederatedRecord describes an SSO identity known only from the audit
// log. OVHcloud does not manage federated users, so nothing beyond the sign-in
// evidence is knowable.
func ovhcloudFederatedRecord(login string, signIn ovhcloudSignIn) AccountRecord {
	record := AccountRecord{
		Roles:       []string{},
		MFAStatus:   signIn.mfa,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodSSO,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:  login,
	}

	// A federated login is the subject the IdP asserted, usually an email.
	if strings.Contains(login, "@") {
		record.Email = login
	}

	if record.MFAStatus == "" {
		record.MFAStatus = coredata.MFAStatusUnknown
	}

	if !signIn.lastLogin.IsZero() {
		record.LastLogin = &signIn.lastLogin
	}

	return record
}

// ovhcloudAPICredentialRecord describes one classic API credential. Its access
// rules stand in for a role: they are all OVHcloud says about its reach.
func ovhcloudAPICredentialRecord(c ovhcloudAPICredential, app ovhcloudAPIApplication, now time.Time) AccountRecord {
	active := c.Status == "validated" &&
		(c.Expiration == nil || c.Expiration.After(now)) &&
		(app.Status == "" || app.Status == "active" || app.Status == "trusted")

	roles := make([]string, 0, len(c.Rules)+1)
	if c.OVHSupport {
		roles = append(roles, "Created by OVHcloud support")
	}

	for _, rule := range c.Rules {
		roles = append(roles, strings.TrimSpace(rule.Method+" "+rule.Path))
	}

	slices.Sort(roles)

	name := app.Name
	if name == "" {
		// A blank name renders as "N/A" in the console, so name the id instead.
		name = fmt.Sprintf("Application %d", c.ApplicationID)
	}

	record := AccountRecord{
		FullName:    name,
		JobTitle:    app.Description,
		Roles:       slices.Compact(roles),
		Active:      &active,
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodAPIKey,
		AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
		CreatedAt:   c.Creation,
		LastLogin:   c.LastUse,
		ExternalID:  strconv.FormatInt(c.CredentialID, 10),
	}

	return record
}

func ovhcloudAuthMethod(signIn ovhcloudSignIn) coredata.AccessReviewEntryAuthMethod {
	switch signIn.kind {
	case "PROVIDER":
		return coredata.AccessReviewEntryAuthMethodSSO
	case "ACCOUNT", "USER":
		return coredata.AccessReviewEntryAuthMethodPassword
	default:
		return coredata.AccessReviewEntryAuthMethodUnknown
	}
}

// ovhcloudRoles reports group names rather than the three roles they collapse
// into: the group is what an operator grants and revokes.
func ovhcloudRoles(u ovhcloudUser) []string {
	groups := slices.Clone(u.Groups)
	if len(groups) == 0 && u.Group != "" {
		groups = []string{u.Group}
	}

	if groups == nil {
		// Empty slice, not nil, so callers need not distinguish the two.
		return []string{}
	}

	slices.Sort(groups)

	return slices.Compact(groups)
}

// ovhcloudIsAdmin is true when ANY group confers ADMIN: auth.User carries both
// `group` (the main one) and `groups`, and reading only the former under-reports
// privilege. An unresolvable group returns nil, not false — "unknown" must not
// read as a cleared row.
func ovhcloudIsAdmin(u ovhcloudUser, roles map[string]string) *bool {
	groups := ovhcloudRoles(u)

	unresolved := false

	for _, g := range groups {
		role, ok := roles[g]
		if !ok {
			unresolved = true

			continue
		}

		if strings.EqualFold(role, "ADMIN") {
			return new(true)
		}
	}

	if unresolved {
		return nil
	}

	return new(false)
}

func ovhcloudAccountType(userType string) coredata.AccessReviewEntryAccountType {
	if userType == "SERVICE" {
		return coredata.AccessReviewEntryAccountTypeServiceAccount
	}

	return coredata.AccessReviewEntryAccountTypeUser
}

func (d *OVHcloudDriver) get(ctx context.Context, out any, segments ...string) error {
	endpoint, err := ovhcloudURL(d.baseURL, segments...)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot perform request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		// Never embed the body: it carries provider wording and IAM detail.
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("cannot decode response: %w", err)
	}

	return nil
}

func ovhcloudURL(baseURL string, segments ...string) (string, error) {
	escaped := make([]string, len(segments))
	for i, s := range segments {
		escaped[i] = url.PathEscape(s)
	}

	endpoint, err := url.JoinPath(baseURL, escaped...)
	if err != nil {
		return "", fmt.Errorf("cannot build ovhcloud url: %w", err)
	}

	return endpoint, nil
}

type ovhcloudNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

var _ NameResolver = (*ovhcloudNameResolver)(nil)

func NewOVHcloudNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &ovhcloudNameResolver{httpClient: httpClient, baseURL: baseURL}
}

// ResolveInstanceName names the source after the NIC handle: organisation is
// null on individual accounts.
func (r *ovhcloudNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	endpoint, err := ovhcloudURL(r.baseURL, "me")
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot perform request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", nameStatusError("ovhcloud account", resp.StatusCode)
	}

	var account ovhcloudAccount
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return "", fmt.Errorf("cannot decode response: %w", err)
	}

	return account.Nichandle, nil
}
