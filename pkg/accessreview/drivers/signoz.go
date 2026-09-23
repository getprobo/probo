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

	sigNozRole struct {
		Name string `json:"name"`
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

		// A user deleted since the listing answers 404.
		var assigned []sigNozRole
		if err := d.getData(ctx, baseURL.JoinPath("api", "v2", "users", url.PathEscape(id), "roles"), &assigned); err != nil {
			if statusErr, ok := errors.AsType[*sigNozStatusError](err); !ok || statusErr.StatusCode != http.StatusNotFound {
				return nil, fmt.Errorf("cannot fetch signoz roles for user %q: %w", id, err)
			}
		}

		roles := make([]string, 0, len(assigned))
		isAdmin := u.IsRoot

		for _, r := range assigned {
			name := strings.TrimSpace(r.Name)
			if name == "signoz-admin" {
				isAdmin = true
			}

			if role := normalizeSigNozRole(name); role != "" && !slices.Contains(roles, role) {
				roles = append(roles, role)
			}
		}

		record := AccountRecord{
			Email:       email,
			FullName:    strings.TrimSpace(u.DisplayName),
			Roles:       roles,
			Active:      sigNozActiveStatus(u.Status),
			IsAdmin:     new(isAdmin),
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
