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
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

const (
	unifiTestAPIBase      = "https://api.ui.com"
	unifiFixtureConsoleID = "FFFFFFFFFFFF0000000000000000000000000000000000000000000000000000:1000000000"
	unifiFixtureConsole   = "Example Console"
)

// unifiTestConsoleID is the console a recording reads, and the fixture the
// sanitized cassette replays.
func unifiTestConsoleID() string {
	if id := os.Getenv("UNIFI_CONSOLE_ID"); id != "" {
		return id
	}

	return unifiFixtureConsoleID
}

// TestUniFiDriver replays a recording of a live UDM SE on Network 10.6.109
// whose Network application had an owner, a View Only admin and one RADIUS
// user.
func TestUniFiDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/unifi", "UNIFI_API_KEY", sanitizeUniFi)
	client := newVCRClientWithHeader(rec, "X-API-KEY", os.Getenv("UNIFI_API_KEY"))

	networkURL, err := UniFiNetworkURL(unifiTestAPIBase, unifiTestConsoleID())
	require.NoError(t, err)

	records, err := NewUniFiDriver(client, networkURL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	// Record mode asserts against the live console; only the sanitized
	// cassette may reach a failure message.
	if os.Getenv("UNIFI_API_KEY") != "" {
		return
	}

	owner := records[0]
	assert.Equal(t, "000000000000000000000001", owner.ExternalID)
	assert.Equal(t, "admin1@example.com", owner.Email)
	assert.Equal(t, "Admin 1", owner.FullName)
	assert.Equal(t, []string{"Owner", "Admin"}, owner.Roles)
	assert.Equal(t, new(true), owner.IsAdmin)
	assert.Nil(t, owner.Active)
	assert.Equal(t, coredata.MFAStatusUnknown, owner.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, owner.AuthMethod)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, owner.AccountType)
	assert.Nil(t, owner.LastLogin)
	require.NotNil(t, owner.CreatedAt)

	viewer := records[1]
	assert.Equal(t, "000000000000000000000002", viewer.ExternalID)
	assert.Equal(t, "admin2@example.com", viewer.Email)
	assert.Equal(t, []string{"View Only"}, viewer.Roles)
	assert.Equal(t, new(false), viewer.IsAdmin)

	radius := records[2]
	assert.Equal(t, "000000000000000000000101", radius.ExternalID)
	assert.Equal(t, "radius-user-1", radius.FullName)
	assert.Empty(t, radius.Email)
	assert.Equal(t, []string{"RADIUS user", "VLAN 3"}, radius.Roles)
	assert.Equal(t, new(false), radius.IsAdmin)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodPassword, radius.AuthMethod)
}

func TestUniFiNameResolver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/unifi_host", "UNIFI_API_KEY", sanitizeUniFi)
	client := newVCRClientWithHeader(rec, "X-API-KEY", os.Getenv("UNIFI_API_KEY"))

	hostURL, err := UniFiHostURL(unifiTestAPIBase, unifiTestConsoleID())
	require.NoError(t, err)

	name, err := NewUniFiNameResolver(client, hostURL).ResolveInstanceName(context.Background())
	require.NoError(t, err)

	if os.Getenv("UNIFI_API_KEY") != "" {
		return
	}

	assert.Equal(t, unifiFixtureConsole, name)
}

func TestUniFiConsoleID(t *testing.T) {
	t.Parallel()

	id, err := UniFiConsoleID("  70a741be825d00000000:1946460215 ")
	require.NoError(t, err)
	assert.Equal(t, "70A741BE825D00000000:1946460215", id)

	for _, invalid := range []string{"", "..", ".", "70A741BE825D", ":1946460215", "70A741BE825D:", "70A7/41BE:1", "70A741BE825D:19?x=1", "GG:1"} {
		_, err := UniFiConsoleID(invalid)
		assert.Error(t, err, invalid)
	}
}

func TestUniFiConnectorRejection(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		status      int
		contentType string
		body        string
		want        string
	}{
		{
			name:        "unknown console",
			status:      http.StatusForbidden,
			contentType: "application/json; charset=utf-8",
			body:        `{"code":"forbidden","httpStatusCode":403,"message":"forbidden: host not found","traceId":"0"}`,
			want:        UniFiConsoleUnreachable,
		},
		{
			name:        "console firmware too old",
			status:      http.StatusBadRequest,
			contentType: "application/json; charset=utf-8",
			body:        `{"code":"bad_request","httpStatusCode":400,"message":"validation error","traceId":"0"}`,
			want:        UniFiConsoleUnsupported,
		},
		{
			name:        "network application error",
			status:      http.StatusBadRequest,
			contentType: "application/json;charset=UTF-8",
			body:        `{"meta":{"rc":"error","msg":"api.err.Invalid"},"data":[]}`,
		},
		{
			name:        "html page",
			status:      http.StatusForbidden,
			contentType: "text/html",
			body:        `<html></html>`,
		},
		{
			name:        "bad key",
			status:      http.StatusUnauthorized,
			contentType: "application/json; charset=utf-8",
			body:        `{"code":"unauthorized","httpStatusCode":401,"message":"unauthorized","traceId":"0"}`,
		},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				resp := &http.Response{
					StatusCode: tc.status,
					Header:     http.Header{"Content-Type": {tc.contentType}},
					Body:       io.NopCloser(strings.NewReader(tc.body)),
				}

				rejected := UniFiConnectorRejection(resp)
				if tc.want == "" {
					assert.Nil(t, rejected)
					return
				}

				require.NotNil(t, rejected)
				assert.Equal(t, tc.want, rejected.Code)
				assert.Equal(t, tc.status, rejected.StatusCode)
			},
		)
	}
}

func TestUniFiDriverRefusesPartialRoster(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		path   string
		status int
		body   string
	}{
		{name: "admins error flag", path: "/stat/admin", status: http.StatusOK, body: `{"meta":{"rc":"error"},"data":[]}`},
		{name: "admins without a list", path: "/stat/admin", status: http.StatusOK, body: `{"meta":{"rc":"ok"}}`},
		{name: "radius users with a null list", path: "/rest/account", status: http.StatusOK, body: `{"meta":{"rc":"ok"},"data":null}`},
		{name: "admin without id", path: "/stat/admin", status: http.StatusOK, body: `{"meta":{"rc":"ok"},"data":[{"email":"jane@example.com"}]}`},
		{name: "admins server error", path: "/stat/admin", status: http.StatusBadGateway, body: `{}`},
		{name: "sites error flag", path: "/self/sites", status: http.StatusOK, body: `{"meta":{"rc":"error"},"data":[]}`},
		{name: "site without name", path: "/self/sites", status: http.StatusOK, body: `{"meta":{"rc":"ok"},"data":[{}]}`},
		{name: "radius users error flag", path: "/rest/account", status: http.StatusOK, body: `{"meta":{"rc":"error"},"data":[]}`},
		{name: "radius user without id", path: "/rest/account", status: http.StatusOK, body: `{"meta":{"rc":"ok"},"data":[{"name":"jane"}]}`},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				records, err := unifiStubDriver(t, tc.path, tc.status, tc.body).ListAccounts(context.Background())
				require.Error(t, err)
				assert.Nil(t, records)
			},
		)
	}
}

// TestUniFiDriverSurfacesConnectorRefusal pins that a console the connector
// refuses mid-sync reaches the customer as a setting rejection, the one error
// whose message they see.
func TestUniFiDriverSurfacesConnectorRefusal(t *testing.T) {
	t.Parallel()

	_, err := unifiStubDriver(
		t,
		"/stat/admin",
		http.StatusForbidden,
		`{"code":"forbidden","httpStatusCode":403,"message":"forbidden: host not found"}`,
	).ListAccounts(context.Background())

	rejected, ok := errors.AsType[*SettingRejectedError](err)
	require.Truef(t, ok, "a connector refusal should be a setting rejection, got %v", err)
	assert.Equal(t, UniFiConsoleUnreachable, rejected.Code)
}

// unifiStubDriver answers path with status and body, and every other legacy
// read with an empty success: one default site, no admins, no RADIUS users.
func unifiStubDriver(t *testing.T, path string, status int, body string) *UniFiDriver {
	t.Helper()

	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		respStatus, respBody := http.StatusOK, `{"meta":{"rc":"ok"},"data":[]}`

		switch {
		case strings.HasSuffix(req.URL.Path, path):
			respStatus, respBody = status, body
		case strings.HasSuffix(req.URL.Path, "/self/sites"):
			respBody = `{"meta":{"rc":"ok"},"data":[{"name":"default"}]}`
		}

		return &http.Response{
			StatusCode: respStatus,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(strings.NewReader(respBody)),
		}, nil
	})}

	networkURL, err := UniFiNetworkURL(unifiTestAPIBase, unifiFixtureConsoleID)
	require.NoError(t, err)

	return NewUniFiDriver(client, networkURL)
}

func TestUniFiAdminRoles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		admin unifiAdmin
		want  []string
	}{
		{name: "owner", admin: unifiAdmin{IsOwner: true, IsSuper: true, Roles: []unifiAdminRole{{Role: "admin"}}}, want: []string{"Owner", "Admin"}},
		{name: "super admin", admin: unifiAdmin{IsSuper: true, Roles: []unifiAdminRole{{Role: "admin"}}}, want: []string{"Super Admin", "Admin"}},
		{name: "view only on two sites", admin: unifiAdmin{Roles: []unifiAdminRole{{Role: "readonly"}, {Role: "readonly"}}}, want: []string{"View Only"}},
		{name: "unknown role", admin: unifiAdmin{Roles: []unifiAdminRole{{Role: "hotspot"}}}, want: []string{"hotspot"}},
		{name: "no role", admin: unifiAdmin{}, want: []string{}},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.want, unifiAdminRoles(tc.admin))
			},
		)
	}
}

func TestUniFiNameResolverTerminalOnClientError(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"code":"forbidden","httpStatusCode":403}`)),
		}, nil
	})}

	hostURL, err := UniFiHostURL(unifiTestAPIBase, unifiFixtureConsoleID)
	require.NoError(t, err)

	_, err = NewUniFiNameResolver(client, hostURL).ResolveInstanceName(context.Background())
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTerminalNameResolution))
}

// sanitizeUniFi rewrites every identity a UniFi recording carries: the console
// ID in request paths, admin ids, emails and names, site ids, RADIUS account
// ids and usernames, and the RADIUS password, which the API returns in
// plaintext. The real console ID comes from the record-time environment.
func sanitizeUniFi(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize unifi response with status %d", i.Response.Code)
	}

	consoleID, err := UniFiConsoleID(os.Getenv("UNIFI_CONSOLE_ID"))
	if err != nil {
		return fmt.Errorf("record with UNIFI_CONSOLE_ID set to the recorded console: %w", err)
	}

	recorded := []string{consoleID}
	i.Request.URL = strings.ReplaceAll(i.Request.URL, consoleID, unifiFixtureConsoleID)

	var body string

	switch path, _, _ := strings.Cut(i.Request.URL, "?"); {
	case strings.HasSuffix(path, "/api/stat/admin"):
		body, err = sanitizeUniFiAdmins(i.Response.Body, &recorded)
	case strings.HasSuffix(path, "/api/self/sites"):
		body, err = sanitizeUniFiSites(i.Response.Body, &recorded)
	case strings.HasSuffix(path, "/rest/account"):
		body, err = sanitizeUniFiRadiusAccounts(i.Response.Body, &recorded)
	case strings.HasSuffix(path, "/v1/hosts/"+unifiFixtureConsoleID):
		body, err = sanitizeUniFiHost(i.Response.Body, &recorded)
	default:
		return fmt.Errorf("refusing to sanitize an unrecognised unifi interaction")
	}

	if err != nil {
		return err
	}

	for _, value := range recorded {
		if value != "" && (strings.Contains(body, value) || strings.Contains(i.Request.URL, value)) {
			return fmt.Errorf("sanitized unifi interaction still contains a recorded identity value")
		}
	}

	replaceCassetteBody(i, body)

	return nil
}

func sanitizeUniFiAdmins(raw string, recorded *[]string) (string, error) {
	var resp struct {
		Meta map[string]any   `json:"meta"`
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "", fmt.Errorf("cannot decode recorded unifi admins: %w", err)
	}

	sites := map[string]string{}

	for n, admin := range resp.Data {
		if err := unifiRejectUnknownFields(
			admin,
			"admin",
			"_id", "avatar_url", "email", "email_alert_enabled", "first_name", "is_owner", "is_super",
			"is_verified", "last_name", "last_site_name", "name", "org_role", "push_alert_enabled",
			"requires_new_password", "roles", "super_roles", "time_created", "ubic_name", "ubic_uuid",
			"ui_settings", "uid",
		); err != nil {
			return "", err
		}

		k := n + 1
		for field, replacement := range map[string]string{
			"_id":        fmt.Sprintf("%024x", k),
			"uid":        fmt.Sprintf("00000000-0000-4000-8000-%012d", k),
			"ubic_uuid":  fmt.Sprintf("00000000-0000-4000-9000-%012d", k),
			"email":      fmt.Sprintf("admin%d@example.com", k),
			"ubic_name":  fmt.Sprintf("admin%d@example.com", k),
			"name":       fmt.Sprintf("Admin %d", k),
			"first_name": "Admin",
			"last_name":  fmt.Sprintf("%d", k),
		} {
			if value, _ := admin[field].(string); len(value) > 2 {
				*recorded = append(*recorded, value)
			}

			admin[field] = replacement
		}

		delete(admin, "ui_settings")
		admin["avatar_url"] = nil

		for _, field := range []string{"roles", "super_roles"} {
			grants, _ := admin[field].([]any)
			for _, g := range grants {
				grant, ok := g.(map[string]any)
				if !ok {
					return "", fmt.Errorf("recorded unifi admin %s entry is not an object", field)
				}

				if err := unifiRejectUnknownFields(grant, field, "permissions", "role", "site_desc", "site_id", "site_name"); err != nil {
					return "", err
				}

				siteID, _ := grant["site_id"].(string)
				if _, ok := sites[siteID]; !ok {
					sites[siteID] = fmt.Sprintf("%024x", 0x100+len(sites))
					*recorded = append(*recorded, siteID)
				}

				grant["site_id"] = sites[siteID]
			}
		}
	}

	return unifiEncode(resp)
}

func sanitizeUniFiSites(raw string, recorded *[]string) (string, error) {
	var resp struct {
		Meta map[string]any   `json:"meta"`
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "", fmt.Errorf("cannot decode recorded unifi sites: %w", err)
	}

	for n, site := range resp.Data {
		if err := unifiRejectUnknownFields(
			site,
			"site",
			"_id", "anonymous_id", "attr_hidden_id", "attr_no_delete", "desc", "device_count",
			"external_id", "name", "role", "role_hotspot",
		); err != nil {
			return "", err
		}

		// The RADIUS path is addressed by this name, so only the default
		// site's generic one can be kept in a recorded URL.
		if site["name"] != "default" || site["desc"] != "Default" {
			return "", fmt.Errorf("refusing to record a unifi site other than the default site")
		}

		k := 0x100 + n
		for field, replacement := range map[string]string{
			"_id":          fmt.Sprintf("%024x", k),
			"anonymous_id": fmt.Sprintf("00000000-0000-4000-8000-%012d", k),
			"external_id":  fmt.Sprintf("00000000-0000-4000-9000-%012d", k),
		} {
			if value, _ := site[field].(string); value != "" {
				*recorded = append(*recorded, value)
			}

			site[field] = replacement
		}
	}

	return unifiEncode(resp)
}

func sanitizeUniFiRadiusAccounts(raw string, recorded *[]string) (string, error) {
	var resp struct {
		Meta map[string]any   `json:"meta"`
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "", fmt.Errorf("cannot decode recorded unifi radius users: %w", err)
	}

	for n, account := range resp.Data {
		if err := unifiRejectUnknownFields(
			account,
			"radius user",
			"_id", "group_policy", "name", "site_id", "tunnel_medium_type", "tunnel_type", "vlan", "x_password",
		); err != nil {
			return "", err
		}

		k := n + 1
		for field, replacement := range map[string]string{
			"_id":        fmt.Sprintf("%024x", 0x100+k),
			"name":       fmt.Sprintf("radius-user-%d", k),
			"site_id":    fmt.Sprintf("%024x", 0x100),
			"x_password": "redacted",
		} {
			if value, _ := account[field].(string); value != "" {
				*recorded = append(*recorded, value)
			}

			account[field] = replacement
		}
	}

	return unifiEncode(resp)
}

// sanitizeUniFiHost keeps only what the name resolver reads: a host record
// carries the console's addresses, location and hardware.
func sanitizeUniFiHost(raw string, recorded *[]string) (string, error) {
	var resp unifiHostResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "", fmt.Errorf("cannot decode recorded unifi host: %w", err)
	}

	if resp.Data.ReportedState.Name == "" {
		return "", fmt.Errorf("recorded unifi host has no name")
	}

	*recorded = append(*recorded, resp.Data.ReportedState.Name)

	return unifiEncode(map[string]any{
		"data": map[string]any{
			"id":            unifiFixtureConsoleID,
			"type":          "console",
			"reportedState": map[string]any{"name": unifiFixtureConsole},
		},
		"httpStatusCode": http.StatusOK,
		"traceId":        "00000000000000000000000000000000",
	})
}

// unifiRejectUnknownFields refuses a field the sanitizer does not classify, so
// an identity UniFi adds later is never recorded verbatim.
func unifiRejectUnknownFields(object map[string]any, what string, known ...string) error {
	for field := range object {
		if !slices.Contains(known, field) {
			return fmt.Errorf("refusing to record unclassified field %q in unifi %s", field, what)
		}
	}

	return nil
}

func unifiEncode(v any) (string, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("cannot re-encode unifi response: %w", err)
	}

	return string(body), nil
}
