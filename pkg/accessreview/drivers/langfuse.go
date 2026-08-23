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
	"fmt"
	"net/http"

	"encoding/json"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"net/url"
	"strings"
)

// LangfuseDriver lists the members of a single Langfuse organization via the
// organization-scoped public API. The organization API key (HTTP Basic,
// publicKey:secretKey) is bound to one organization on the configured host,
// so GET /api/public/organizations/memberships returns every member with no
// tenant selector. The Basic credential is applied by the connection
// transport; the base URL spans the regional cloud hosts and self-hosting.
type LangfuseDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*LangfuseDriver)(nil)

type langfuseMembershipsResponse struct {
	Memberships []langfuseMembership `json:"memberships"`
}

type langfuseMembership struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

func NewLangfuseDriver(httpClient *http.Client, baseURL string) *LangfuseDriver {
	return &LangfuseDriver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (d *LangfuseDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	baseURL, err := url.Parse(d.baseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse langfuse base URL: %w", err)
	}

	endpoint := baseURL.JoinPath("api", "public", "organizations", "memberships")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create langfuse memberships request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute langfuse memberships request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch langfuse memberships: unexpected status %d", httpResp.StatusCode)
	}

	var resp langfuseMembershipsResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode langfuse memberships response: %w", err)
	}

	records := make([]AccountRecord, 0, len(resp.Memberships))

	for _, m := range resp.Memberships {
		email := strings.TrimSpace(m.Email)
		if email == "" {
			continue
		}

		records = append(records, AccountRecord{
			Email:       email,
			FullName:    langfuseFullName(m, email),
			Roles:       langfuseRoles(m.Role),
			IsAdmin:     new(langfuseIsAdmin(m.Role)),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  strings.TrimSpace(m.UserID),
		})
	}

	return records, nil
}

func langfuseFullName(m langfuseMembership, fallback string) string {
	if name := strings.TrimSpace(m.Name); name != "" {
		return name
	}

	return fallback
}

// langfuseRoles maps the Langfuse organization MembershipRole
// (OWNER/ADMIN/MEMBER/VIEWER, and the RBAC NONE) to a stable display label,
// preserving unknown future roles verbatim.
func langfuseRoles(role string) []string {
	switch strings.ToUpper(strings.TrimSpace(role)) {
	case "OWNER":
		return []string{"Owner"}
	case "ADMIN":
		return []string{"Admin"}
	case "MEMBER":
		return []string{"Member"}
	case "VIEWER":
		return []string{"Viewer"}
	case "NONE":
		return []string{"None"}
	default:
		if r := strings.TrimSpace(role); r != "" {
			return []string{r}
		}

		return []string{}
	}
}

func langfuseIsAdmin(role string) bool {
	switch strings.ToUpper(strings.TrimSpace(role)) {
	case "OWNER", "ADMIN":
		return true
	default:
		return false
	}
}

func langfuseSource() Factory {
	return provider.Over(func(
		ctx context.Context,
		credential connector.HTTPCredential,
		opened *provider.Handle,
		logger *log.Logger,
	) (Driver, error) {
		driver, err := langfuseSourceDriver(ctx, credential.Client, opened.Connector, logger, opened.Endpoints)
		if err != nil {
			return nil, err
		}

		return capable(
			driver,
			nil,
			nil,
		), nil
	})
}

func langfuseSourceDriver(
	_ context.Context,
	c *http.Client,
	conn *coredata.Connector,
	_ *log.Logger,
	_ provider.Endpoints,
) (Driver, error) {
	settings, err := coredata.ConnectorSettings[coredata.LangfuseConnectorSettings](conn)
	if err != nil {
		return nil, fmt.Errorf("cannot read langfuse connector settings: %w", err)
	}

	baseURL, err := provider.NormalizeSelfHostedBaseURL(settings.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot create langfuse driver: %w", err)
	}

	return NewLangfuseDriver(c, baseURL), nil
}
