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
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// SigNozDriver fetches organization members from the SigNoz API. The API key
// is injected by the connector's API-key HTTP client via the SIGNOZ-API-KEY
// header. The same base URL serves SigNoz Cloud (region/tenant host) and
// self-hosted instances.
type SigNozDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*SigNozDriver)(nil)

// sigNozEnvelope is the standard SigNoz REST response wrapper:
// {"status":"success","data": <payload>}.
type sigNozEnvelope struct {
	Data json.RawMessage `json:"data"`
}

// sigNozUser models a user from GET /api/v1/user. That ("v1") list endpoint
// returns role inline; the v2 endpoint omits role entirely, which would
// silently disable admin detection.
type sigNozUser struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	IsRoot      bool   `json:"isRoot"`
	CreatedAt   string `json:"createdAt"`
}

func NewSigNozDriver(httpClient *http.Client, baseURL string) *SigNozDriver {
	return &SigNozDriver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// The list endpoint omits roles, so admin detection needs a second call.
type sigNozServiceAccount struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type sigNozServiceAccountRole struct {
	Role struct {
		Name string `json:"name"`
	} `json:"role"`
}

func (d *SigNozDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	users, err := d.queryUsers(ctx)
	if err != nil {
		return nil, err
	}

	serviceAccounts, err := d.queryServiceAccounts(ctx)
	if err != nil {
		return nil, err
	}

	records := make([]AccountRecord, 0, len(users)+len(serviceAccounts))

	for _, u := range users {
		email := strings.TrimSpace(u.Email)
		if email == "" {
			continue
		}

		roles := sigNozRoles(u.Role)

		record := AccountRecord{
			Email:       email,
			FullName:    strings.TrimSpace(u.DisplayName),
			Roles:       roles,
			Active:      sigNozActiveStatus(u.Status),
			IsAdmin:     new(u.IsRoot || slices.Contains(roles, "Admin")),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  strings.TrimSpace(u.ID),
		}

		if t, ok := parseSigNozTimestamp(u.CreatedAt); ok {
			record.CreatedAt = &t
		}

		records = append(records, record)
	}

	for _, sa := range serviceAccounts {
		email := strings.TrimSpace(sa.Email)
		if email == "" {
			continue
		}

		roles, err := d.serviceAccountRoles(ctx, sa.ID)
		if err != nil {
			return nil, err
		}

		record := AccountRecord{
			Email:       email,
			FullName:    strings.TrimSpace(sa.Name),
			Roles:       roles,
			Active:      sigNozServiceAccountActiveStatus(sa.Status),
			IsAdmin:     sigNozIsAdmin(roles),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodAPIKey,
			AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
			ExternalID:  strings.TrimSpace(sa.ID),
		}

		if t, ok := parseSigNozTimestamp(sa.CreatedAt); ok {
			record.CreatedAt = &t
		}

		records = append(records, record)
	}

	return records, nil
}

// Roles are best-effort: a nil slice means unknown, which keeps the account
// reviewable without reporting it as non-admin. A cancelled context is a real
// failure and propagates.
func (d *SigNozDriver) serviceAccountRoles(ctx context.Context, id string) ([]string, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}

	assigned, err := d.queryServiceAccountRoles(ctx, id)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("cannot fetch signoz service account roles: %w", err)
		}

		return nil, nil
	}

	roles := make([]string, 0, len(assigned))

	for _, a := range assigned {
		for _, role := range sigNozRoles(a.Role.Name) {
			if !slices.Contains(roles, role) {
				roles = append(roles, role)
			}
		}
	}

	return roles, nil
}

func sigNozIsAdmin(roles []string) *bool {
	if roles == nil {
		return nil
	}

	return new(slices.Contains(roles, "Admin"))
}

func (d *SigNozDriver) queryUsers(ctx context.Context) ([]sigNozUser, error) {
	baseURL, err := url.Parse(d.baseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse signoz base URL: %w", err)
	}

	endpoint := baseURL.JoinPath("api", "v1", "user")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create signoz users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute signoz users request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch signoz users: unexpected status %d", httpResp.StatusCode)
	}

	var envelope sigNozEnvelope
	if err := json.NewDecoder(httpResp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("cannot decode signoz users response: %w", err)
	}

	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return []sigNozUser{}, nil
	}

	var users []sigNozUser
	if err := json.Unmarshal(envelope.Data, &users); err != nil {
		return nil, fmt.Errorf("cannot decode signoz users data: %w", err)
	}

	return users, nil
}

// Listing service accounts needs `serviceaccount:list`, which a key below
// signoz-admin lacks. Degrade to empty rather than failing the whole source:
// the users already fetched remain a valid review population.
func (d *SigNozDriver) queryServiceAccounts(ctx context.Context) ([]sigNozServiceAccount, error) {
	var accounts []sigNozServiceAccount

	err := d.getJSON(ctx, "service accounts", &accounts, "api", "v1", "service_accounts")
	if err != nil {
		if errors.Is(err, errSigNozUnavailable) {
			return nil, nil
		}

		return nil, err
	}

	return accounts, nil
}

func (d *SigNozDriver) queryServiceAccountRoles(ctx context.Context, id string) ([]sigNozServiceAccountRole, error) {
	var account struct {
		ServiceAccountRoles []sigNozServiceAccountRole `json:"serviceAccountRoles"`
	}

	err := d.getJSON(ctx, "service account roles", &account, "api", "v1", "service_accounts", url.PathEscape(id))
	if err != nil {
		return nil, err
	}

	return account.ServiceAccountRoles, nil
}

var (
	errSigNozUnavailable = errors.New("signoz endpoint unavailable to this key")
)

func (d *SigNozDriver) getJSON(ctx context.Context, what string, out any, segments ...string) error {
	baseURL, err := url.Parse(d.baseURL)
	if err != nil {
		return fmt.Errorf("cannot parse signoz base URL: %w", err)
	}

	endpoint := baseURL.JoinPath(segments...)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("cannot create signoz %s request: %w", what, err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot execute signoz %s request: %w", what, err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	switch {
	case httpResp.StatusCode == http.StatusUnauthorized,
		httpResp.StatusCode == http.StatusForbidden,
		httpResp.StatusCode == http.StatusNotFound:
		return errSigNozUnavailable
	case httpResp.StatusCode < 200 || httpResp.StatusCode >= 300:
		return fmt.Errorf("cannot fetch signoz %s: unexpected status %d", what, httpResp.StatusCode)
	}

	var envelope sigNozEnvelope
	if err := json.NewDecoder(httpResp.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("cannot decode signoz %s response: %w", what, err)
	}

	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil
	}

	if err := json.Unmarshal(envelope.Data, out); err != nil {
		return fmt.Errorf("cannot decode signoz %s data: %w", what, err)
	}

	return nil
}

func sigNozServiceAccountActiveStatus(status string) *bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active":
		return new(true)
	case "revoked", "deleted", "inactive":
		return new(false)
	default:
		return nil
	}
}

// sigNozRoles normalizes a SigNoz role string (ADMIN / EDITOR / VIEWER, or the
// managed-role display names signoz-admin / signoz-editor / signoz-viewer)
// into a stable label, preserving unknown custom roles verbatim. Matching is
// exact (not substring) so a custom role merely containing "admin" is not
// silently promoted to Admin.
func sigNozRoles(raw string) []string {
	role := strings.TrimSpace(raw)
	if role == "" {
		return []string{}
	}

	switch strings.ToLower(role) {
	case "admin", "signoz-admin":
		return []string{"Admin"}
	case "editor", "signoz-editor":
		return []string{"Editor"}
	case "viewer", "signoz-viewer":
		return []string{"Viewer"}
	default:
		return []string{role}
	}
}

// sigNozActiveStatus maps the SigNoz user status. SigNoz emits exactly
// "active", "pending_invite" and "deleted"; anything else is treated as an
// unknown signal (nil) rather than fabricated.
func sigNozActiveStatus(status string) *bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active":
		return new(true)
	case "pending_invite", "deleted":
		return new(false)
	default:
		return nil
	}
}

func parseSigNozTimestamp(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}

	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
	} {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

// signozNameResolver resolves the SigNoz organization display name via
// GET /api/v2/orgs/me on the configured instance. The organization is derived
// from the API key's claims, so no identifier is needed in the path.
type signozNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

var _ NameResolver = (*signozNameResolver)(nil)

func NewSigNozNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &signozNameResolver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (r *signozNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	baseURL, err := url.Parse(r.baseURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse signoz base URL: %w", err)
	}

	endpoint := baseURL.JoinPath("api", "v2", "orgs", "me")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return "", fmt.Errorf("cannot create signoz organization request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute signoz organization request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	// Best-effort: a non-2xx (revoked key, or an older SigNoz without this
	// route) must not make the source-name worker retry forever. Keep the
	// generic source name; a dead key surfaces on the next ListAccounts.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var envelope struct {
		Data struct {
			DisplayName string `json:"displayName"`
			Name        string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&envelope); err != nil {
		return "", fmt.Errorf("cannot decode signoz organization response: %w", err)
	}

	if name := strings.TrimSpace(envelope.Data.DisplayName); name != "" {
		return name, nil
	}

	return strings.TrimSpace(envelope.Data.Name), nil
}
