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
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	UniFiConsoleUnreachable = "unifi_console_unreachable"
	UniFiConsoleUnsupported = "unifi_console_unsupported"

	unifiLegacyOK    = "ok"
	unifiLegacyError = "error"
)

// unifiConsoleIDPattern is a Site Manager host ID, a hex hardware part and a
// number, as it appears in unifi.ui.com/consoles/<id>/. It becomes a path
// segment on api.ui.com.
var unifiConsoleIDPattern = regexp.MustCompile(`^[0-9A-F]+:[0-9]+$`)

type (
	// UniFiDriver lists the Network application administrators of one UniFi
	// console and the users of its built-in RADIUS server, through the Site
	// Manager cloud connector and the Network application's legacy API.
	UniFiDriver struct {
		httpClient *http.Client
		networkURL string
	}

	unifiNameResolver struct {
		httpClient *http.Client
		hostURL    string
	}

	// unifiStatusError carries the status of a non-2xx response so the name
	// resolver can classify it while the driver reports it plainly.
	unifiStatusError struct {
		what string
		code int
	}

	// unifiConnectorError is the envelope Site Manager answers with when the
	// connector itself refuses a request, before it reaches the console.
	unifiConnectorError struct {
		HTTPStatusCode int `json:"httpStatusCode"`
	}

	// unifiLegacyResponse is the legacy API envelope, which can carry an error
	// flag on a 200.
	unifiLegacyResponse[T any] struct {
		Meta struct {
			RC string `json:"rc"`
		} `json:"meta"`
		Data []T `json:"data"`
	}

	unifiAdmin struct {
		ID          string           `json:"_id"`
		Email       string           `json:"email"`
		Name        string           `json:"name"`
		FirstName   string           `json:"first_name"`
		LastName    string           `json:"last_name"`
		IsOwner     bool             `json:"is_owner"`
		IsSuper     bool             `json:"is_super"`
		Roles       []unifiAdminRole `json:"roles"`
		TimeCreated int64            `json:"time_created"`
	}

	unifiAdminRole struct {
		Role string `json:"role"`
	}

	// unifiSite is a site as the legacy API names it in its paths.
	unifiSite struct {
		Name string `json:"name"`
	}

	// unifiRadiusAccount leaves x_password undecoded: the endpoint returns the
	// account's password in plaintext.
	unifiRadiusAccount struct {
		ID   string          `json:"_id"`
		Name string          `json:"name"`
		VLAN json.RawMessage `json:"vlan"`
	}

	unifiHostResponse struct {
		Data struct {
			ReportedState struct {
				Name string `json:"name"`
			} `json:"reportedState"`
		} `json:"data"`
	}
)

var (
	_ Driver       = (*UniFiDriver)(nil)
	_ NameResolver = (*unifiNameResolver)(nil)
)

func (e *unifiStatusError) Error() string {
	return fmt.Sprintf("cannot fetch unifi %s: unexpected status %d", e.what, e.code)
}

// UniFiConsoleID canonicalises a customer-supplied console ID, so one console
// cannot be stored under two spellings and read as two sources.
func UniFiConsoleID(id string) (string, error) {
	id = strings.ToUpper(strings.TrimSpace(id))
	if !unifiConsoleIDPattern.MatchString(id) {
		return "", fmt.Errorf("invalid unifi console id")
	}

	return id, nil
}

// UniFiNetworkURL is a console's Network application root behind the Site
// Manager connector.
func UniFiNetworkURL(apiBase, consoleID string) (string, error) {
	id, err := UniFiConsoleID(consoleID)
	if err != nil {
		return "", err
	}

	endpoint, err := url.JoinPath(apiBase, "v1", "connector", "consoles", url.PathEscape(id), "proxy", "network")
	if err != nil {
		return "", fmt.Errorf("cannot build unifi network URL: %w", err)
	}

	return endpoint, nil
}

// UniFiAdminsURL lists the console's Network administrators.
func UniFiAdminsURL(networkURL string) (string, error) {
	endpoint, err := url.JoinPath(networkURL, "api", "stat", "admin")
	if err != nil {
		return "", fmt.Errorf("cannot build unifi admins URL: %w", err)
	}

	return endpoint, nil
}

// UniFiHostURL is the Site Manager record of a console.
func UniFiHostURL(apiBase, consoleID string) (string, error) {
	id, err := UniFiConsoleID(consoleID)
	if err != nil {
		return "", err
	}

	endpoint, err := url.JoinPath(apiBase, "v1", "hosts", url.PathEscape(id))
	if err != nil {
		return "", fmt.Errorf("cannot build unifi host URL: %w", err)
	}

	return endpoint, nil
}

// UniFiConnectorRejection returns the refusal the Site Manager connector gives
// for a console it cannot reach, or nil. The console's own errors pass through
// the connector in a different envelope and are not settings problems.
func UniFiConnectorRejection(resp *http.Response) *SettingRejectedError {
	if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusBadRequest {
		return nil
	}

	if !respondsWithJSON(resp) {
		return nil
	}

	var body unifiConnectorError
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<16)).Decode(&body); err != nil ||
		body.HTTPStatusCode != resp.StatusCode {
		return nil
	}

	if resp.StatusCode == http.StatusForbidden {
		return &SettingRejectedError{
			Code:       UniFiConsoleUnreachable,
			StatusCode: resp.StatusCode,
			Message:    "UniFi Site Manager finds no console with this ID for this API key. Check the Console ID, and create the key from the UniFi account that owns the console, or as an organization key.",
		}
	}

	return &SettingRejectedError{
		Code:       UniFiConsoleUnsupported,
		StatusCode: resp.StatusCode,
		Message:    "UniFi Site Manager cannot reach this console: it must run UniFi OS 5.0.3 or later.",
	}
}

// UniFiLegacyRefused reports whether a legacy API body carries its error flag.
func UniFiLegacyRefused(body []byte) bool {
	var resp unifiLegacyResponse[json.RawMessage]

	return json.Unmarshal(body, &resp) == nil && resp.Meta.RC == unifiLegacyError
}

// NewUniFiDriver reads the console whose Network application root
// UniFiNetworkURL returned.
func NewUniFiDriver(httpClient *http.Client, networkURL string) *UniFiDriver {
	retryClient := *httpClient
	retryClient.Transport = &retryRoundTripper{
		next:       httpClient.Transport,
		maxRetries: 3,
	}

	return &UniFiDriver{
		httpClient: &retryClient,
		networkURL: networkURL,
	}
}

// ListAccounts returns the Network administrators, then the RADIUS users of
// every site. A failure on either fails the whole sync.
func (d *UniFiDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	adminsURL, err := UniFiAdminsURL(d.networkURL)
	if err != nil {
		return nil, err
	}

	admins, err := unifiGetLegacy[unifiAdmin](ctx, d.httpClient, adminsURL, "admins")
	if err != nil {
		return nil, err
	}

	sitesURL, err := url.JoinPath(d.networkURL, "api", "self", "sites")
	if err != nil {
		return nil, fmt.Errorf("cannot build unifi sites URL: %w", err)
	}

	sites, err := unifiGetLegacy[unifiSite](ctx, d.httpClient, sitesURL, "sites")
	if err != nil {
		return nil, err
	}

	records := make([]AccountRecord, 0, len(admins))

	for _, a := range admins {
		record, err := unifiAdminRecord(a)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	for _, site := range sites {
		if site.Name == "" {
			return nil, fmt.Errorf("cannot list unifi radius users: site has no name")
		}

		accountsURL, err := url.JoinPath(d.networkURL, "api", "s", url.PathEscape(site.Name), "rest", "account")
		if err != nil {
			return nil, fmt.Errorf("cannot build unifi radius users URL: %w", err)
		}

		accounts, err := unifiGetLegacy[unifiRadiusAccount](ctx, d.httpClient, accountsURL, "radius users")
		if err != nil {
			return nil, err
		}

		for _, a := range accounts {
			record, err := unifiRadiusRecord(a)
			if err != nil {
				return nil, err
			}

			records = append(records, record)
		}
	}

	return records, nil
}

// unifiGetLegacy reads a legacy API list, which is never paginated.
func unifiGetLegacy[T any](ctx context.Context, httpClient *http.Client, endpoint, what string) ([]T, error) {
	var resp unifiLegacyResponse[T]
	if err := unifiGetJSON(ctx, httpClient, endpoint, what, &resp); err != nil {
		return nil, err
	}

	if resp.Meta.RC != unifiLegacyOK {
		return nil, fmt.Errorf("cannot list unifi %s: the network application refused the request", what)
	}

	// A missing or null list is not an empty one: reading it as empty would
	// mark every account removed.
	if resp.Data == nil {
		return nil, fmt.Errorf("cannot list unifi %s: the network application returned no list", what)
	}

	return resp.Data, nil
}

func unifiGetJSON(ctx context.Context, httpClient *http.Client, endpoint, what string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("cannot create unifi %s request: %w", what, err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot execute unifi %s request: %w", what, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if rejected := UniFiConnectorRejection(resp); rejected != nil {
		return rejected
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &unifiStatusError{what: what, code: resp.StatusCode}
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("cannot decode unifi %s response: %w", what, err)
	}

	return nil
}

func unifiAdminRecord(a unifiAdmin) (AccountRecord, error) {
	if a.ID == "" {
		return AccountRecord{}, fmt.Errorf("cannot map unifi admin: admin has no id")
	}

	name := strings.TrimSpace(a.Name)
	if name == "" {
		name = strings.TrimSpace(a.FirstName + " " + a.LastName)
	}

	isAdmin := a.IsOwner || a.IsSuper || slices.ContainsFunc(
		a.Roles,
		func(r unifiAdminRole) bool { return r.Role == "admin" },
	)

	record := AccountRecord{
		Email:       strings.TrimSpace(a.Email),
		FullName:    name,
		Roles:       unifiAdminRoles(a),
		IsAdmin:     &isAdmin,
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:  a.ID,
	}

	if a.TimeCreated > 0 {
		createdAt := time.Unix(a.TimeCreated, 0).UTC()
		record.CreatedAt = &createdAt
	}

	return record, nil
}

// unifiAdminRoles names the console-wide role first, then each site role once.
func unifiAdminRoles(a unifiAdmin) []string {
	roles := []string{}

	switch {
	case a.IsOwner:
		roles = append(roles, "Owner")
	case a.IsSuper:
		roles = append(roles, "Super Admin")
	}

	for _, r := range a.Roles {
		label := r.Role

		switch r.Role {
		case "admin":
			label = "Admin"
		case "readonly":
			label = "View Only"
		}

		if label != "" && !slices.Contains(roles, label) {
			roles = append(roles, label)
		}
	}

	return roles
}

func unifiRadiusRecord(a unifiRadiusAccount) (AccountRecord, error) {
	if a.ID == "" {
		return AccountRecord{}, fmt.Errorf("cannot map unifi radius user: user has no id")
	}

	roles := []string{"RADIUS user"}
	if vlan := strings.Trim(string(a.VLAN), `"`); vlan != "" && vlan != "null" {
		roles = append(roles, "VLAN "+vlan)
	}

	isAdmin := false

	return AccountRecord{
		FullName:    strings.TrimSpace(a.Name),
		Roles:       roles,
		IsAdmin:     &isAdmin,
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodPassword,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:  a.ID,
	}, nil
}

// NewUniFiNameResolver titles the source with the console's name in Site
// Manager.
func NewUniFiNameResolver(httpClient *http.Client, hostURL string) NameResolver {
	return &unifiNameResolver{httpClient: httpClient, hostURL: hostURL}
}

func (r *unifiNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	var resp unifiHostResponse

	err := unifiGetJSON(ctx, r.httpClient, r.hostURL, "host", &resp)
	if rejected, ok := errors.AsType[*SettingRejectedError](err); ok && rejected != nil {
		return "", nameStatusError("unifi host", rejected.StatusCode)
	}

	if statusErr, ok := errors.AsType[*unifiStatusError](err); ok && statusErr != nil {
		return "", nameStatusError("unifi "+statusErr.what, statusErr.code)
	}

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Data.ReportedState.Name), nil
}
