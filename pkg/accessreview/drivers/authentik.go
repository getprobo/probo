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
	"errors"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

// AuthentikDriver lists the users of a self-hosted authentik instance.
// authentik accepts the bearer token only when its intent is `api`; the app
// password POST /core/users/service_account/ returns reads as invalid.
type AuthentikDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*AuthentikDriver)(nil)

type (
	// Next is a page NUMBER, 0 when exhausted -- not a URL to follow.
	authentikPagination struct {
		Next       int `json:"next"`
		TotalPages int `json:"total_pages"`
	}

	authentikUser struct {
		PK          int64                  `json:"pk"`
		UUID        string                 `json:"uuid"`
		Username    string                 `json:"username"`
		Name        string                 `json:"name"`
		Email       string                 `json:"email"`
		Type        string                 `json:"type"`
		IsActive    bool                   `json:"is_active"`
		IsSuperuser bool                   `json:"is_superuser"`
		LastLogin   string                 `json:"last_login"`
		DateJoined  string                 `json:"date_joined"`
		GroupsObj   []authentikPartialItem `json:"groups_obj"`
		RolesObj    []authentikPartialItem `json:"roles_obj"`
	}

	authentikPartialItem struct {
		Name string `json:"name"`
	}

	authentikDevice struct {
		User struct {
			PK int64 `json:"pk"`
		} `json:"user"`
	}

	authentikBrand struct {
		BrandingTitle string `json:"branding_title"`
		Default       bool   `json:"default"`
	}
)

const authentikPageSize = "100"

// `static` is excluded: those are recovery codes, not a login factor.
var authentikFactorPaths = []string{"totp", "webauthn", "duo", "sms"}

func NewAuthentikDriver(httpClient *http.Client, baseURL string) *AuthentikDriver {
	return &AuthentikDriver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (d *AuthentikDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	usersWithFactor, factorsComplete, err := d.listUsersWithFactor(ctx)
	if err != nil {
		return nil, err
	}

	users, err := fetchAuthentikPages[authentikUser](ctx, d.httpClient, d.baseURL, "core", "users")
	if err != nil {
		return nil, err
	}

	records := make([]AccountRecord, 0, len(users))

	for _, u := range users {
		email := strings.TrimSpace(u.Email)
		if email == "" {
			email = strings.TrimSpace(u.Username)
		}

		if email == "" {
			continue
		}

		serviceAccount := authentikIsServiceAccount(u.Type)

		record := AccountRecord{
			Email:       email,
			FullName:    strings.TrimSpace(u.Name),
			Roles:       authentikRoles(u),
			Active:      new(u.IsActive),
			IsAdmin:     new(u.IsSuperuser),
			MFAStatus:   authentikMFAStatus(u, usersWithFactor, factorsComplete),
			AuthMethod:  authentikAuthMethod(serviceAccount),
			AccountType: authentikAccountType(serviceAccount),
			ExternalID:  strings.TrimSpace(u.UUID),
			CreatedAt:   parseRFC3339Ptr(u.DateJoined),
			LastLogin:   parseRFC3339Ptr(u.LastLogin),
		}

		records = append(records, record)
	}

	return records, nil
}

// listUsersWithFactor returns the users holding a second factor, and whether
// every factor kind was readable -- a missing stage answers 404 and a token
// lacking the device permission 403, either of which hides a factor. The typed
// endpoints are used because /authenticators/admin/all/ serializes no user.
func (d *AuthentikDriver) listUsersWithFactor(ctx context.Context) (map[int64]bool, bool, error) {
	usersWithFactor := make(map[int64]bool)
	complete := true

	for _, factor := range authentikFactorPaths {
		devices, err := fetchAuthentikPages[authentikDevice](ctx, d.httpClient, d.baseURL, "authenticators", "admin", factor)

		for _, device := range devices {
			if device.User.PK != 0 {
				usersWithFactor[device.User.PK] = true
			}
		}

		if err != nil {
			if e, ok := errors.AsType[*authentikStatusError](err); ok && (e.code == http.StatusForbidden || e.code == http.StatusNotFound) {
				complete = false

				continue
			}

			return nil, false, err
		}
	}

	return usersWithFactor, complete, nil
}

func fetchAuthentikPages[T any](
	ctx context.Context,
	httpClient *http.Client,
	baseURL string,
	pathSegments ...string,
) ([]T, error) {
	collection := strings.Join(pathSegments, "/")

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse authentik base URL: %w", err)
	}

	// A slash-less collection path costs a 301 on every page.
	segments := append([]string{"api", "v3"}, pathSegments...)
	segments[len(segments)-1] += "/"

	endpoint := base.JoinPath(segments...)

	var items []T

	page := 1

	for range maxPaginationPages {
		endpoint.RawQuery = url.Values{
			"page":      {strconv.Itoa(page)},
			"page_size": {authentikPageSize},
		}.Encode()

		result, err := fetchAuthentikPage[T](ctx, httpClient, endpoint.String(), collection)
		if err != nil {
			return items, err
		}

		items = append(items, result.Results...)

		if result.Pagination.Next <= page {
			return items, nil
		}

		page = result.Pagination.Next
	}

	return nil, fmt.Errorf("cannot list all authentik %s: %w", collection, ErrPaginationLimitReached)
}

func fetchAuthentikPage[T any](
	ctx context.Context,
	httpClient *http.Client,
	endpoint string,
	collection string,
) (*authentikPage[T], error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create authentik %s request: %w", collection, err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute authentik %s request: %w", collection, err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, &authentikStatusError{collection: collection, code: httpResp.StatusCode}
	}

	page := &authentikPage[T]{}
	if err := json.NewDecoder(httpResp.Body).Decode(page); err != nil {
		return nil, fmt.Errorf("cannot decode authentik %s response: %w", collection, err)
	}

	return page, nil
}

type authentikPage[T any] struct {
	Pagination authentikPagination `json:"pagination"`
	Results    []T                 `json:"results"`
}

type authentikStatusError struct {
	collection string
	code       int
}

func (e *authentikStatusError) Error() string {
	return fmt.Sprintf("cannot fetch authentik %s: unexpected status %d", e.collection, e.code)
}

func authentikIsServiceAccount(userType string) bool {
	switch strings.ToLower(strings.TrimSpace(userType)) {
	case "service_account", "internal_service_account":
		return true
	default:
		return false
	}
}

func authentikAccountType(serviceAccount bool) coredata.AccessReviewEntryAccountType {
	if serviceAccount {
		return coredata.AccessReviewEntryAccountTypeServiceAccount
	}

	return coredata.AccessReviewEntryAccountTypeUser
}

func authentikAuthMethod(serviceAccount bool) coredata.AccessReviewEntryAuthMethod {
	if serviceAccount {
		return coredata.AccessReviewEntryAuthMethodServiceAccount
	}

	return coredata.AccessReviewEntryAuthMethodUnknown
}

// A service account holds no factor by construction, so DISABLED there would
// be a finding no reviewer can action.
func authentikMFAStatus(u authentikUser, usersWithFactor map[int64]bool, factorsComplete bool) coredata.MFAStatus {
	if usersWithFactor[u.PK] {
		return coredata.MFAStatusEnabled
	}

	if !factorsComplete || authentikIsServiceAccount(u.Type) {
		return coredata.MFAStatusUnknown
	}

	return coredata.MFAStatusDisabled
}

func authentikRoles(u authentikUser) []string {
	roles := make([]string, 0, len(u.RolesObj)+len(u.GroupsObj))

	for _, item := range slices.Concat(u.RolesObj, u.GroupsObj) {
		if name := strings.TrimSpace(item.Name); name != "" && !slices.Contains(roles, name) {
			roles = append(roles, name)
		}
	}

	return roles
}

// A non-2xx yields no name rather than an error: a token may list users
// without being able to read brands.
type authentikNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

var _ NameResolver = (*authentikNameResolver)(nil)

func NewAuthentikNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &authentikNameResolver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (r *authentikNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	brands, err := fetchAuthentikPages[authentikBrand](ctx, r.httpClient, r.baseURL, "core", "brands")
	if err != nil {
		if _, ok := errors.AsType[*authentikStatusError](err); ok {
			return "", nil
		}

		return "", err
	}

	var fallback string

	for _, brand := range brands {
		title := strings.TrimSpace(brand.BrandingTitle)
		if title == "" {
			continue
		}

		if brand.Default {
			return title, nil
		}

		if fallback == "" {
			fallback = title
		}
	}

	return fallback, nil
}

func authentikSource() Factory {
	return provider.Over(func(
		ctx context.Context,
		credential connector.HTTPCredential,
		opened *provider.Handle,
		logger *log.Logger,
	) (Driver, error) {
		driver, err := authentikSourceDriver(ctx, credential.Client, opened.Connector, logger, opened.Endpoints)
		if err != nil {
			return nil, err
		}

		return capable(
			driver,
			authentikSourceNameResolver(ctx, credential.Client, opened.Connector, logger, opened.Endpoints),
			nil,
		), nil
	})
}

func authentikSourceDriver(
	_ context.Context,
	c *http.Client,
	conn *coredata.Connector,
	_ *log.Logger,
	_ provider.Endpoints,
) (Driver, error) {
	settings, err := coredata.ConnectorSettings[coredata.AuthentikConnectorSettings](conn)
	if err != nil {
		return nil, fmt.Errorf("cannot read authentik connector settings: %w", err)
	}

	baseURL, err := provider.NormalizeSelfHostedBaseURL(settings.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot create authentik driver: %w", err)
	}

	return NewAuthentikDriver(c, baseURL), nil
}

func authentikSourceNameResolver(
	ctx context.Context,
	c *http.Client,
	conn *coredata.Connector,
	logger *log.Logger,
	_ provider.Endpoints,
) NameResolver {
	settings, err := coredata.ConnectorSettings[coredata.AuthentikConnectorSettings](conn)
	if err != nil {
		logger.ErrorCtx(ctx, "cannot read authentik connector settings", log.Error(err))
		return nil
	}

	baseURL, err := provider.NormalizeSelfHostedBaseURL(settings.BaseURL)
	if err != nil {
		logger.ErrorCtx(ctx, "invalid authentik base url in connector settings", log.Error(err))
		return nil
	}

	return NewAuthentikNameResolver(c, baseURL)
}
