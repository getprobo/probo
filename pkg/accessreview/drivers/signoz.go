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

// SigNozDriver fetches organization members and service accounts from the
// SigNoz API. The API key is injected by the connector's API-key HTTP client
// via the SIGNOZ-API-KEY header. The same base URL serves SigNoz Cloud
// (region/tenant host) and self-hosted instances.
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

type (
	// sigNozUser models a user from GET /api/v2/users.
	sigNozUser struct {
		ID          string `json:"id"`
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
		Status      string `json:"status"`
		IsRoot      bool   `json:"isRoot"`
		CreatedAt   string `json:"createdAt"`
	}

	// sigNozServiceAccount models a service account from
	// GET /api/v1/service_accounts.
	sigNozServiceAccount struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		Status    string `json:"status"`
		CreatedAt string `json:"createdAt"`
	}

	sigNozRole struct {
		Name string `json:"name"`
	}

	sigNozAPIKey struct {
		CreatedAt      string `json:"createdAt"`
		LastObservedAt string `json:"lastObservedAt"`
	}

	sigNozStatusError struct {
		StatusCode int
	}
)

func NewSigNozDriver(httpClient *http.Client, baseURL string) *SigNozDriver {
	client := *httpClient
	client.Transport = &retryRoundTripper{
		next:       httpClient.Transport,
		maxRetries: 3,
	}

	return &SigNozDriver{
		httpClient: &client,
		baseURL:    baseURL,
	}
}

func (d *SigNozDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	baseURL, err := url.Parse(d.baseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse signoz base URL: %w", err)
	}

	users, err := d.listUsers(ctx, baseURL)
	if err != nil {
		return nil, err
	}

	serviceAccounts, err := d.listServiceAccounts(ctx, baseURL)
	if err != nil {
		return nil, err
	}

	return append(users, serviceAccounts...), nil
}

func (d *SigNozDriver) listUsers(ctx context.Context, baseURL *url.URL) ([]AccountRecord, error) {
	var users []sigNozUser
	if err := d.getData(ctx, baseURL.JoinPath("api", "v2", "users"), &users); err != nil {
		return nil, fmt.Errorf("cannot fetch signoz users: %w", err)
	}

	records := make([]AccountRecord, 0, len(users))

	for _, u := range users {
		email := strings.TrimSpace(u.Email)

		id := strings.TrimSpace(u.ID)
		if email == "" || id == "" {
			continue
		}

		roles, isAdmin, err := d.getRoles(ctx, baseURL.JoinPath("api", "v2", "users", url.PathEscape(id), "roles"))
		if err != nil {
			return nil, fmt.Errorf("cannot fetch signoz roles for user %q: %w", id, err)
		}

		record := AccountRecord{
			Email:       email,
			FullName:    strings.TrimSpace(u.DisplayName),
			Roles:       roles,
			Active:      sigNozActiveStatus(u.Status),
			IsAdmin:     new(u.IsRoot || isAdmin),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  id,
		}

		if t, ok := parseSigNozTimestamp(u.CreatedAt); ok {
			record.CreatedAt = &t
		}

		records = append(records, record)
	}

	return records, nil
}

func (d *SigNozDriver) listServiceAccounts(ctx context.Context, baseURL *url.URL) ([]AccountRecord, error) {
	var serviceAccounts []sigNozServiceAccount
	if err := d.getData(ctx, baseURL.JoinPath("api", "v1", "service_accounts"), &serviceAccounts); err != nil {
		return nil, fmt.Errorf("cannot fetch signoz service accounts: %w", err)
	}

	records := make([]AccountRecord, 0, len(serviceAccounts))

	for _, sa := range serviceAccounts {
		id := strings.TrimSpace(sa.ID)
		if id == "" {
			continue
		}

		roles, isAdmin, err := d.getRoles(ctx, baseURL.JoinPath("api", "v1", "service_accounts", url.PathEscape(id), "roles"))
		if err != nil {
			return nil, fmt.Errorf("cannot fetch signoz roles for service account %q: %w", id, err)
		}

		lastUsed, err := d.getLastKeyUse(ctx, baseURL.JoinPath("api", "v1", "service_accounts", url.PathEscape(id), "keys"))
		if err != nil {
			return nil, fmt.Errorf("cannot fetch signoz keys for service account %q: %w", id, err)
		}

		record := AccountRecord{
			Email:       strings.TrimSpace(sa.Email),
			FullName:    strings.TrimSpace(sa.Name),
			Roles:       roles,
			Active:      sigNozActiveStatus(sa.Status),
			IsAdmin:     new(isAdmin),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodAPIKey,
			AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
			ExternalID:  id,
			LastLogin:   lastUsed,
		}

		if t, ok := parseSigNozTimestamp(sa.CreatedAt); ok {
			record.CreatedAt = &t
		}

		records = append(records, record)
	}

	return records, nil
}

// getRoles reads an account's roles; an account deleted since the listing
// answers 404 and has none.
func (d *SigNozDriver) getRoles(ctx context.Context, endpoint *url.URL) ([]string, bool, error) {
	var assigned []sigNozRole
	if err := d.getData(ctx, endpoint, &assigned); err != nil && !isSigNozNotFound(err) {
		return nil, false, err
	}

	roles := make([]string, 0, len(assigned))
	isAdmin := false

	for _, r := range assigned {
		name := strings.TrimSpace(r.Name)
		if name == "signoz-admin" {
			isAdmin = true
		}

		if role := normalizeSigNozRole(name); role != "" && !slices.Contains(roles, role) {
			roles = append(roles, role)
		}
	}

	return roles, isAdmin, nil
}

// getLastKeyUse returns the latest use of any of a service account's keys.
// SigNoz stamps lastObservedAt when it creates a key, so a key still carrying
// its creation time was never used.
func (d *SigNozDriver) getLastKeyUse(ctx context.Context, endpoint *url.URL) (*time.Time, error) {
	var keys []sigNozAPIKey
	if err := d.getData(ctx, endpoint, &keys); err != nil && !isSigNozNotFound(err) {
		return nil, err
	}

	var lastUsed *time.Time

	for _, k := range keys {
		observed, ok := parseSigNozTimestamp(k.LastObservedAt)
		if !ok {
			continue
		}

		if created, ok := parseSigNozTimestamp(k.CreatedAt); ok && observed.Sub(created) < 100*time.Millisecond {
			continue
		}

		if lastUsed == nil || observed.After(*lastUsed) {
			lastUsed = &observed
		}
	}

	return lastUsed, nil
}

func isSigNozNotFound(err error) bool {
	statusErr, ok := errors.AsType[*sigNozStatusError](err)

	return ok && statusErr.StatusCode == http.StatusNotFound
}

func (e *sigNozStatusError) Error() string {
	return fmt.Sprintf("unexpected status %d", e.StatusCode)
}

// getData decodes the envelope's data into out, leaving out untouched when
// data is null.
func (d *SigNozDriver) getData(ctx context.Context, endpoint *url.URL, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot execute request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return &sigNozStatusError{StatusCode: httpResp.StatusCode}
	}

	var envelope sigNozEnvelope
	if err := json.NewDecoder(httpResp.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("cannot decode response: %w", err)
	}

	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil
	}

	if err := json.Unmarshal(envelope.Data, out); err != nil {
		return fmt.Errorf("cannot decode response data: %w", err)
	}

	return nil
}

// normalizeSigNozRole labels the managed roles and keeps custom ones verbatim.
func normalizeSigNozRole(role string) string {
	switch role {
	case "signoz-admin":
		return "Admin"
	case "signoz-editor":
		return "Editor"
	case "signoz-viewer":
		return "Viewer"
	default:
		return role
	}
}

// sigNozActiveStatus maps the SigNoz user and service account status. SigNoz
// emits exactly "active", "pending_invite" and "deleted" for users, and
// "active" and "deleted" for service accounts; anything else is treated as an
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
