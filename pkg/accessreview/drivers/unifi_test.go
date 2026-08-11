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
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
)

// unifiTestConsoleID is the synthetic console the cassette was authored
// against. Its colon is the real shape Site Manager uses.
const unifiTestConsoleID = "ABCDEF0123456789:1234567890"

// unifiTestAPIBase mirrors the provider registration's Endpoints.APIBase.
const unifiTestAPIBase = "https://api.ui.com/v1/connector/consoles"

func unifiTestLogger() *log.Logger {
	return log.NewLogger(log.WithName("test"))
}

func TestUniFiDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/unifi", "UNIFI_API_KEY")
	// UniFi authenticates with the key in X-API-KEY. The matcher ignores it
	// and the BeforeSave hook strips it, so replay needs no credential.
	client := newVCRClientWithHeader(rec, "X-API-KEY", os.Getenv("UNIFI_API_KEY"))

	baseURL, err := UniFiConsoleBaseURL(unifiTestAPIBase, unifiTestConsoleID)
	require.NoError(t, err)

	records, err := NewUniFiDriver(client, unifiTestLogger(), baseURL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 8)

	// UniFi models none of these grants with an email or an MFA signal, so
	// every record keys on its ExternalID alone.
	for _, record := range records {
		assert.Emptyf(t, record.Email, "record %q must carry no email", record.ExternalID)
		assert.Equalf(t, coredata.MFAStatusUnknown, record.MFAStatus, "record %q", record.ExternalID)
		assert.NotEmpty(t, record.ExternalID)
		assert.NotEmpty(t, record.FullName)
		assert.False(t, record.IsAdmin, "neither API exposes an administrative grant")
		// Every record names somebody who holds access. A resource — an SSID, a
		// VPN server, a RADIUS directory — is the thing being accessed, and a
		// campaign cannot ask a reviewer to approve or revoke one.
		assert.Equalf(
			t,
			coredata.AccessReviewEntryAccountTypeUser,
			record.AccountType,
			"record %q must not be a resource row", record.ExternalID,
		)
	}

	byID := make(map[string]AccountRecord, len(records))
	for _, record := range records {
		_, duplicate := byID[record.ExternalID]
		require.Falsef(t, duplicate, "duplicate external id %q", record.ExternalID)
		byID[record.ExternalID] = record
	}

	// Every ExternalID is scoped to its site, so two sites' grants can never
	// collapse into one entry.
	const (
		site   = "site/11111111-1111-4111-8111-111111111111/"
		office = site + "wifi/22222222-2222-4222-8222-222222222001"
		guest  = site + "wifi/22222222-2222-4222-8222-222222222002"
		iot    = site + "wifi/22222222-2222-4222-8222-222222222003"
		corp   = site + "wifi/22222222-2222-4222-8222-222222222004"
	)

	// ---- RADIUS users: the accounts that authenticate to Enterprise SSIDs and
	// RADIUS-backed VPNs, and the closest thing UniFi models to a person.
	//
	// Both Enterprise SSIDs delegate to this site's sole directory, so each user
	// row names both networks — sorted, so the role list does not reshuffle when
	// the API reorders its broadcast pages. That join is the whole answer to
	// "who can access which WiFi network", and it lives on the person's row
	// rather than on a separate row for the directory.

	alice := byID[site+"radius-user/ra1"]
	assert.Equal(t, "alice", alice.FullName)
	assert.Equal(t, []string{
		"RADIUS user", "Site: HQ", "RADIUS profile: Built-in RADIUS",
		"WiFi: Corp-WiFi", "WiFi: Office", "VLAN: 20",
	}, alice.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodPassword, alice.AuthMethod)
	// The legacy entry carries no enabled/disabled state — an account exists or
	// it does not — so fabricating one would be worse than reporting none.
	assert.Nil(t, alice.Active)

	// Older firmware spells vlan as a number rather than a string; both must
	// decode, or a review would silently lose the VLAN on some consoles.
	bob := byID[site+"radius-user/ra2"]
	assert.Equal(t, "bob", bob.FullName)
	assert.Equal(t, []string{
		"RADIUS user", "Site: HQ", "RADIUS profile: Built-in RADIUS",
		"WiFi: Corp-WiFi", "WiFi: Office", "VLAN: 30",
	}, bob.Roles)

	carol := byID[site+"radius-user/ra3"]
	assert.Equal(t, []string{
		"RADIUS user", "Site: HQ", "RADIUS profile: Built-in RADIUS",
		"WiFi: Corp-WiFi", "WiFi: Office",
	}, carol.Roles)

	// The blank fourth entry has nothing stable to key on and is dropped.
	assert.Len(t, unifiRecordsWithPrefix(records, site+"radius-user/"), 3)

	// No record may carry a RADIUS password: the driver does not decode the
	// field at all, and this pins that it stays that way.
	for _, record := range records {
		assert.NotContains(t, record.FullName, "REDACTED")
		assert.NotContains(t, strings.Join(record.Roles, " "), "REDACTED")
	}

	// ---- Devices on an SSID's client-filtering list, the API's only per-SSID
	// statement of who may join.

	// An ALLOW filter is exhaustive: these MACs, and only these, may join. Both
	// are keyed with separators stripped so the same device keys identically
	// however UniFi formats the address, while the displayed name keeps the
	// API's formatting. The SSID's security mode rides along on the role list —
	// it used to live on the SSID's own row.
	allowedColon := byID[office+"/mac/942a6f26c6ca"]
	assert.Equal(t, "94:2a:6f:26:c6:ca", allowedColon.FullName)
	assert.Equal(t, []string{"WiFi: Office", "Site: HQ", "Security: WPA2_WPA3_ENTERPRISE"}, allowedColon.Roles)
	require.NotNil(t, allowedColon.Active)
	assert.True(t, *allowedColon.Active)

	allowedDash := byID[office+"/mac/aabbccddeeff"]
	assert.Equal(t, "AA-BB-CC-DD-EE-FF", allowedDash.FullName)
	require.NotNil(t, allowedDash.Active)
	assert.True(t, *allowedDash.Active)

	// A BLOCK filter lists denials, and the SSID is disabled besides — both
	// settle Active false. That the SSID's on-air state still reaches this row
	// is what stops a listed device reading as live access to a network that is
	// off the air, now that the SSID has no row of its own to report it.
	blocked := byID[iot+"/mac/001122334455"]
	assert.Equal(t, "00:11:22:33:44:55", blocked.FullName)
	assert.Equal(t, []string{"WiFi: IoT Legacy", "Site: HQ", "Security: WPA2_PERSONAL"}, blocked.Roles)
	require.NotNil(t, blocked.Active)
	assert.False(t, *blocked.Active)

	// ---- SSIDs and the directory contribute no rows of their own.

	// Neither the open Guest SSID nor either Enterprise SSID is an account, and
	// with no filtering policy Guest and Corp-WiFi name nobody at all. The MAC
	// records live UNDER these IDs, so absence has to be checked on the exact
	// key rather than the prefix.
	for _, ssid := range []string{office, guest, iot, corp} {
		_, present := byID[ssid]
		assert.Falsef(t, present, "the SSID at %q must not have a record of its own", ssid)
	}

	assert.Empty(t, unifiRecordsWithPrefix(records, guest+"/mac/"))
	assert.Empty(t, unifiRecordsWithPrefix(records, corp+"/mac/"))
	assert.Empty(t, unifiRecordsWithPrefix(records, site+"radius-profile/"))

	// The API exposes no VPN membership at all — no link from a server to a
	// RADIUS profile, and "VPN client access" describes live sessions rather
	// than a roster — so the driver does not request VPN servers and claims
	// nothing about them.
	assert.Empty(t, unifiRecordsWithPrefix(records, site+"vpn/"))

	// ---- Hotspot vouchers.

	liveVoucher := byID[site+"voucher/44444444-4444-4444-8444-444444444001"]
	assert.Equal(t, "hotel-guest", liveVoucher.FullName)
	assert.Equal(t, []string{"Hotspot voucher", "Site: HQ", "Guests authorized: 1"}, liveVoucher.Roles)
	require.NotNil(t, liveVoucher.Active)
	assert.True(t, *liveVoucher.Active)
	require.NotNil(t, liveVoucher.LastLogin)
	assert.Equal(t, "2026-07-02T14:20:00Z", liveVoucher.LastLogin.Format("2006-01-02T15:04:05Z"))

	expiredVoucher := byID[site+"voucher/44444444-4444-4444-8444-444444444002"]
	assert.Nil(t, expiredVoucher.LastLogin)
	require.NotNil(t, expiredVoucher.Active)
	assert.False(t, *expiredVoucher.Active)

	// Connected clients are deliberately NOT a record shape: they are a
	// point-in-time snapshot, so a device that happened to be offline during a
	// campaign would read as access removed.
	assert.Empty(t, unifiRecordsWithPrefix(records, site+"client/"))
}

// TestUniFiRadiusUsersBestEffort pins the containment around the LEGACY RADIUS
// endpoint. It is undocumented, so an auth or not-found answer must cost only
// the RADIUS users — not the whole review — while a transient failure must still
// abort, since a silently short answer would mark real accounts removed on the
// next campaign.
func TestUniFiRadiusUsersBestEffort(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		status    int
		body      string
		wantUsers int
		wantError bool
	}{
		{
			name:      "key does not reach the legacy api",
			status:    http.StatusUnauthorized,
			body:      `{"meta":{"rc":"error"}}`,
			wantUsers: 0,
		},
		{
			name:      "legacy route absent on this console",
			status:    http.StatusNotFound,
			body:      `{"meta":{"rc":"error","msg":"api.err.NoSiteContext"}}`,
			wantUsers: 0,
		},
		{
			// The legacy API reports application errors in the body with a 200.
			name:      "application level error on a 200",
			status:    http.StatusOK,
			body:      `{"meta":{"rc":"error","msg":"api.err.NoSiteContext"},"data":[]}`,
			wantUsers: 0,
		},
		{
			name:      "transient failure aborts",
			status:    http.StatusInternalServerError,
			body:      `{"meta":{"rc":"error"}}`,
			wantError: true,
		},
		{
			name:      "throttled aborts",
			status:    http.StatusTooManyRequests,
			body:      `{"meta":{"rc":"error"}}`,
			wantError: true,
		},
		{
			name:      "users returned",
			status:    http.StatusOK,
			body:      `{"meta":{"rc":"ok"},"data":[{"_id":"ra1","name":"alice"}]}`,
			wantUsers: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := unifiStubClient(t, func(req *http.Request) (int, string) {
				if strings.HasSuffix(req.URL.Path, "/list/account") {
					return tc.status, tc.body
				}

				if strings.HasSuffix(req.URL.Path, "/v1/sites") {
					return http.StatusOK, `{"count":1,"limit":200,"offset":0,"totalCount":1,` +
						`"data":[{"id":"site-a","name":"HQ","internalReference":"default"}]}`
				}

				return http.StatusOK, `{"count":0,"limit":200,"offset":0,"totalCount":0,"data":[]}`
			})

			records, err := NewUniFiDriver(client, unifiTestLogger(), "https://api.ui.com/x").
				ListAccounts(context.Background())

			if tc.wantError {
				require.Error(t, err)
				assert.Nil(t, records)

				return
			}

			require.NoError(t, err)
			assert.Len(t, unifiRecordsWithPrefix(records, "site/site-a/radius-user/"), tc.wantUsers)
		})
	}
}

// TestUniFiRadiusUsersNeedLegacySiteName pins that the legacy path is addressed
// by the site's legacy short name, never its UUID. A site without one is skipped
// rather than requested with an identifier the endpoint would reject.
func TestUniFiRadiusUsersNeedLegacySiteName(t *testing.T) {
	t.Parallel()

	var legacyPaths []string

	client := unifiStubClient(t, func(req *http.Request) (int, string) {
		if strings.Contains(req.URL.Path, "/list/account") {
			legacyPaths = append(legacyPaths, req.URL.Path)

			return http.StatusOK, `{"meta":{"rc":"ok"},"data":[{"_id":"ra1","name":"alice"}]}`
		}

		if strings.HasSuffix(req.URL.Path, "/v1/sites") {
			return http.StatusOK, `{"count":2,"limit":200,"offset":0,"totalCount":2,"data":[` +
				`{"id":"site-a","name":"HQ","internalReference":"default"},` +
				`{"id":"site-b","name":"Branch"}]}`
		}

		return http.StatusOK, `{"count":0,"limit":200,"offset":0,"totalCount":0,"data":[]}`
	})

	records, err := NewUniFiDriver(client, unifiTestLogger(), "https://api.ui.com/x").
		ListAccounts(context.Background())
	require.NoError(t, err)

	// Only the site carrying a legacy short name was asked, and it was asked
	// with that name rather than its UUID.
	require.Len(t, legacyPaths, 1)
	assert.Contains(t, legacyPaths[0], "/api/s/default/list/account")
	assert.NotContains(t, legacyPaths[0], "site-a")

	require.Len(t, records, 1)
	assert.Equal(t, "site/site-a/radius-user/ra1", records[0].ExternalID)
}

// TestUniFiRadiusUserRecord pins identity and labelling for one directory entry.
func TestUniFiRadiusUserRecord(t *testing.T) {
	t.Parallel()

	t.Run("keys on the opaque id, displays the username", func(t *testing.T) {
		t.Parallel()

		record, ok := unifiRadiusUserRecord(
			unifiRadiusUser{ID: "ra1", Name: " alice ", VLAN: "20"},
			"Built-in RADIUS", []string{"Office", "Corp-WiFi", "Lab"},
			"site/s/", "Site: HQ",
		)
		require.True(t, ok)
		assert.Equal(t, "site/s/radius-user/ra1", record.ExternalID)
		assert.Equal(t, "alice", record.FullName)
		// The SSID names are sorted so the role list does not reshuffle between
		// campaigns purely because the API paged its broadcasts differently.
		assert.Equal(t, []string{
			"RADIUS user", "Site: HQ", "RADIUS profile: Built-in RADIUS",
			"WiFi: Corp-WiFi", "WiFi: Lab", "WiFi: Office", "VLAN: 20",
		}, record.Roles)
	})

	// Several profiles means a user cannot be attributed to one, so neither the
	// directory nor the networks it unlocks are named — a guess would read as
	// fact on the reviewer's screen.
	t.Run("omits the profile when the site has several", func(t *testing.T) {
		t.Parallel()

		record, ok := unifiRadiusUserRecord(
			unifiRadiusUser{ID: "ra1", Name: "alice"}, "", nil, "site/s/", "Site: HQ",
		)
		require.True(t, ok)
		assert.Equal(t, []string{"RADIUS user", "Site: HQ"}, record.Roles)
	})

	// A directory no SSID delegates to still holds accounts; they simply reach
	// no WiFi network through it.
	t.Run("names the profile with no networks behind it", func(t *testing.T) {
		t.Parallel()

		record, ok := unifiRadiusUserRecord(
			unifiRadiusUser{ID: "ra1", Name: "alice"}, "Built-in RADIUS", nil, "site/s/", "Site: HQ",
		)
		require.True(t, ok)
		assert.Equal(t, []string{"RADIUS user", "Site: HQ", "RADIUS profile: Built-in RADIUS"}, record.Roles)
	})

	// An entry with no _id still keys stably off its username, which is the
	// directory's own key.
	t.Run("falls back to the username as the key", func(t *testing.T) {
		t.Parallel()

		record, ok := unifiRadiusUserRecord(
			unifiRadiusUser{Name: "alice"}, "", nil, "site/s/", "Site: HQ",
		)
		require.True(t, ok)
		assert.Equal(t, "site/s/radius-user/alice", record.ExternalID)
		assert.Equal(t, "alice", record.FullName)
	})

	// Nothing stable to key on: emitting it would produce a record that lands
	// under a different key on every run.
	t.Run("drops an entry with neither id nor username", func(t *testing.T) {
		t.Parallel()

		_, ok := unifiRadiusUserRecord(unifiRadiusUser{Name: "  "}, "", nil, "site/s/", "Site: HQ")
		assert.False(t, ok)
	})
}

// TestUniFiLenientString pins the tolerance the legacy shape needs: the same
// field has shipped as a string and as a number, and an unexpected shape must
// not fail a whole review over a descriptive field.
func TestUniFiLenientString(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`"20"`:     "20",
		`30`:       "30",
		`null`:     "",
		`""`:       "",
		`{"a":1}`:  "",
		`[1,2]`:    "",
		`true`:     "",
		`"  40  "`: "  40  ",
	}

	for raw, want := range cases {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			var got unifiLenientString
			require.NoError(t, got.UnmarshalJSON([]byte(raw)))
			assert.Equal(t, want, string(got))
		})
	}
}

// TestUniFiProfileLabel pins that a role always names something: a profile with
// no display name falls back to its ID rather than producing "RADIUS profile: ".
func TestUniFiProfileLabel(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Built-in RADIUS", unifiProfileLabel(unifiRadiusProfile{ID: "p1", Name: " Built-in RADIUS "}))
	assert.Equal(t, "p2", unifiProfileLabel(unifiRadiusProfile{ID: "p2"}))
}

// TestUniFiConsoleBaseURL pins the composition every caller shares. The console
// ID is a path segment on api.ui.com, so a value that could add segments must
// not survive into a request URL.
func TestUniFiConsoleBaseURL(t *testing.T) {
	t.Parallel()

	baseURL, err := UniFiConsoleBaseURL(unifiTestAPIBase, unifiTestConsoleID)
	require.NoError(t, err)
	assert.Equal(
		t,
		"https://api.ui.com/v1/connector/consoles/ABCDEF0123456789:1234567890/proxy/network",
		baseURL,
	)

	// The probe URL must stay byte-identical to the one verified against the
	// live API, even though the driver now also reaches the legacy sibling path.
	sitesURL, err := UniFiSitesURL(baseURL)
	require.NoError(t, err)
	assert.Equal(t, baseURL+"/integration/v1/sites", sitesURL)

	// A blank console ID would resolve to the console collection itself,
	// silently reviewing something other than a console.
	_, err = UniFiConsoleBaseURL(unifiTestAPIBase, "   ")
	require.Error(t, err)

	// A traversal attempt escapes into an encoded segment rather than
	// climbing out of the console.
	escaped, err := UniFiConsoleBaseURL(unifiTestAPIBase, "../../../v1/hosts")
	require.NoError(t, err)
	assert.NotContains(t, escaped, "/../")
	assert.Contains(t, escaped, "%2F")
}

// TestUniFiDriverPagination pins that the driver walks a collection to
// completion off totalCount rather than stopping at the first page.
func TestUniFiDriverPagination(t *testing.T) {
	t.Parallel()

	client := unifiStubClient(t, func(req *http.Request) (int, string) {
		path := req.URL.Path

		switch {
		case strings.HasSuffix(path, "/v1/sites"):
			switch req.URL.Query().Get("offset") {
			case "0":
				return http.StatusOK, `{"count":1,"limit":200,"offset":0,"totalCount":2,` +
					`"data":[{"id":"site-a","name":"HQ"}]}`
			default:
				return http.StatusOK, `{"count":1,"limit":200,"offset":1,"totalCount":2,` +
					`"data":[{"id":"site-b","name":"Branch"}]}`
			}
		case strings.HasSuffix(path, "/hotspot/vouchers"):
			return http.StatusOK, `{"count":1,"limit":1000,"offset":0,"totalCount":1,` +
				`"data":[{"id":"v-1","name":"lobby","expired":false}]}`
		default:
			return http.StatusOK, `{"count":0,"limit":200,"offset":0,"totalCount":0,"data":[]}`
		}
	})

	records, err := NewUniFiDriver(client, unifiTestLogger(), "https://api.ui.com/x").
		ListAccounts(context.Background())
	require.NoError(t, err)

	// Both pages of sites were walked, so both sites contributed their voucher;
	// the site role distinguishes them.
	require.Len(t, records, 2)
	assert.Equal(t, []string{"Hotspot voucher", "Site: HQ"}, records[0].Roles)
	assert.Equal(t, []string{"Hotspot voucher", "Site: Branch"}, records[1].Roles)

	// The two sites deliberately answer with the SAME resource ID. Scoping each
	// ExternalID to its site is what stops them collapsing into one entry that
	// a reviewer would decide once for both.
	assert.Equal(t, "site/site-a/voucher/v-1", records[0].ExternalID)
	assert.Equal(t, "site/site-b/voucher/v-1", records[1].ExternalID)
}

// TestUniFiDriverAbortsOnFailure pins the identity guarantee: entries key on
// ExternalID, so a run that silently dropped a collection would mark every
// grant in it removed on the next campaign. Only a 404 — a stable "this console
// does not have that feature" — is allowed to yield nothing.
func TestUniFiDriverAbortsOnFailure(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		suffix    string
		status    int
		wantError bool
	}{
		{name: "wifi broadcasts unreadable", suffix: "/wifi/broadcasts", status: http.StatusInternalServerError, wantError: true},
		{name: "radius profiles throttled", suffix: "/radius/profiles", status: http.StatusTooManyRequests, wantError: true},
		{name: "vouchers forbidden", suffix: "/hotspot/vouchers", status: http.StatusForbidden, wantError: true},
		{name: "wifi absent on this console", suffix: "/wifi/broadcasts", status: http.StatusNotFound},
		{name: "radius absent on this console", suffix: "/radius/profiles", status: http.StatusNotFound},
		{name: "hotspot absent on this console", suffix: "/hotspot/vouchers", status: http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := unifiStubClient(t, func(req *http.Request) (int, string) {
				if strings.HasSuffix(req.URL.Path, tc.suffix) {
					return tc.status, `{"statusCode":0}`
				}

				if strings.HasSuffix(req.URL.Path, "/v1/sites") {
					return http.StatusOK, `{"count":1,"limit":200,"offset":0,"totalCount":1,` +
						`"data":[{"id":"site-a","name":"HQ"}]}`
				}

				if strings.HasSuffix(req.URL.Path, "/list/account") {
					return http.StatusOK, `{"meta":{"rc":"ok"},"data":[]}`
				}

				return http.StatusOK, `{"count":0,"limit":200,"offset":0,"totalCount":0,"data":[]}`
			})

			records, err := NewUniFiDriver(client, unifiTestLogger(), "https://api.ui.com/x").
				ListAccounts(context.Background())

			if tc.wantError {
				require.Error(t, err)
				assert.Nil(t, records)

				return
			}

			require.NoError(t, err)
			assert.Empty(t, records)
		})
	}
}

// TestUniFiDriverBlankSiteIDAborts pins the malformed-list guarantee: the site
// ID is the path segment every grant below it is fetched with, so a blank row
// would silently hide a whole site rather than fail.
func TestUniFiDriverBlankSiteIDAborts(t *testing.T) {
	t.Parallel()

	client := unifiStubClient(t, func(req *http.Request) (int, string) {
		if !strings.HasSuffix(req.URL.Path, "/v1/sites") {
			t.Errorf("driver must not fetch a site's grants after a malformed list row")
		}

		return http.StatusOK, `{"count":1,"limit":200,"offset":0,"totalCount":1,` +
			`"data":[{"id":"  ","name":"HQ"}]}`
	})

	records, err := NewUniFiDriver(client, unifiTestLogger(), "https://api.ui.com/x").
		ListAccounts(context.Background())
	require.Error(t, err)
	assert.Nil(t, records)
}

// TestUniFiDriverSkipsVanishedBroadcast pins that a broadcast deleted between
// the list and its detail fetch is dropped rather than failing the run: the
// 404 is a stable answer, and re-listing would race the same way.
func TestUniFiDriverSkipsVanishedBroadcast(t *testing.T) {
	t.Parallel()

	client := unifiStubClient(t, func(req *http.Request) (int, string) {
		path := req.URL.Path

		switch {
		case strings.HasSuffix(path, "/v1/sites"):
			return http.StatusOK, `{"count":1,"limit":200,"offset":0,"totalCount":1,` +
				`"data":[{"id":"site-a","name":"HQ"}]}`
		case strings.HasSuffix(path, "/wifi/broadcasts"):
			return http.StatusOK, `{"count":2,"limit":200,"offset":0,"totalCount":2,"data":[` +
				`{"id":"w-gone","name":"Old","enabled":true,"securityConfiguration":{"type":"OPEN"}},` +
				`{"id":"w-live","name":"Office","enabled":true,"securityConfiguration":{"type":"OPEN"}}]}`
		case strings.HasSuffix(path, "/wifi/broadcasts/w-gone"):
			return http.StatusNotFound, `{"statusCode":404}`
		case strings.HasSuffix(path, "/wifi/broadcasts/w-live"):
			// Carries a filtered device, so the surviving broadcast has
			// something to contribute and the run is not vacuously empty.
			return http.StatusOK, `{"id":"w-live","name":"Office","enabled":true,` +
				`"securityConfiguration":{"type":"OPEN"},` +
				`"clientFilteringPolicy":{"action":"ALLOW",` +
				`"macAddressFilter":["de:ad:be:ef:00:01"]}}`
		default:
			return http.StatusOK, `{"count":0,"limit":200,"offset":0,"totalCount":0,"data":[]}`
		}
	})

	records, err := NewUniFiDriver(client, unifiTestLogger(), "https://api.ui.com/x").
		ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "site/site-a/wifi/w-live/mac/deadbeef0001", records[0].ExternalID)
}

// TestUniFiWifiRecordsDedupesMACs pins that one device listed twice — plainly
// or once per format — yields one record. Two would inflate the fetched-account
// count and collapse into a single entry on upsert anyway.
func TestUniFiWifiRecordsDedupesMACs(t *testing.T) {
	t.Parallel()

	broadcast := unifiWifiBroadcast{ID: "w-1", Name: "Office", Enabled: true}
	broadcast.SecurityConfiguration.Type = "WPA2_PERSONAL"
	broadcast.ClientFilteringPolicy = &struct {
		Action           string   `json:"action"`
		MACAddressFilter []string `json:"macAddressFilter"`
	}{
		Action: "ALLOW",
		MACAddressFilter: []string{
			"94:2a:6f:26:c6:ca",
			"94-2A-6F-26-C6-CA",
			"  ",
			"de:ad:be:ef:00:01",
		},
	}

	records := unifiWifiRecords(broadcast, "Office", "site/s/", "Site: HQ")

	// Two distinct devices, and no row for the SSID itself.
	require.Len(t, records, 2)
	assert.Equal(t, "site/s/wifi/w-1/mac/942a6f26c6ca", records[0].ExternalID)
	assert.Equal(t, "site/s/wifi/w-1/mac/deadbeef0001", records[1].ExternalID)
	// The first spelling wins the display name; the key is format-independent.
	assert.Equal(t, "94:2a:6f:26:c6:ca", records[0].FullName)
}

// TestUniFiWifiRecordsNoFilterNoRecords pins that an SSID whose membership the
// API does not state contributes nothing. A WPA-Personal or open SSID with no
// filter is reached with a credential UniFi shares among everyone holding it,
// and there is no identifiable holder to review.
func TestUniFiWifiRecordsNoFilterNoRecords(t *testing.T) {
	t.Parallel()

	broadcast := unifiWifiBroadcast{ID: "w-1", Name: "Guest", Enabled: true}
	broadcast.SecurityConfiguration.Type = "OPEN"

	assert.Empty(t, unifiWifiRecords(broadcast, "Guest", "site/s/", "Site: HQ"))
}

// TestUniFiWifiRecordsDisabledSSID pins that an SSID's on-air state still
// reaches its device rows. The broadcast used to carry a row of its own to
// report it; without one, a MAC on a disabled SSID would otherwise read as live
// access to a network nobody can join.
func TestUniFiWifiRecordsDisabledSSID(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		enabled    bool
		action     string
		wantActive bool
	}{
		{name: "allowed on a live ssid", enabled: true, action: "ALLOW", wantActive: true},
		{name: "allowed on a disabled ssid", enabled: false, action: "ALLOW"},
		{name: "blocked on a live ssid", enabled: true, action: "BLOCK"},
		{name: "blocked on a disabled ssid", enabled: false, action: "BLOCK"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			broadcast := unifiWifiBroadcast{ID: "w-1", Name: "Office", Enabled: tc.enabled}
			broadcast.SecurityConfiguration.Type = "WPA2_PERSONAL"
			broadcast.ClientFilteringPolicy = &struct {
				Action           string   `json:"action"`
				MACAddressFilter []string `json:"macAddressFilter"`
			}{Action: tc.action, MACAddressFilter: []string{"de:ad:be:ef:00:01"}}

			records := unifiWifiRecords(broadcast, "Office", "site/s/", "Site: HQ")
			require.Len(t, records, 1)
			require.NotNil(t, records[0].Active)
			assert.Equal(t, tc.wantActive, *records[0].Active)
		})
	}
}

// TestUniFiSoleProfileAttribution pins the ambiguity rule at the driver level:
// a user is attributed to a RADIUS profile only when the site has exactly one.
func TestUniFiSoleProfileAttribution(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		profiles  string
		wantRoles []string
	}{
		{
			name: "one profile is unambiguous",
			profiles: `{"count":1,"limit":200,"offset":0,"totalCount":1,` +
				`"data":[{"id":"p1","name":"Built-in"}]}`,
			wantRoles: []string{"RADIUS user", "Site: HQ", "RADIUS profile: Built-in"},
		},
		{
			name: "two profiles leave it unattributed",
			profiles: `{"count":2,"limit":200,"offset":0,"totalCount":2,"data":[` +
				`{"id":"p1","name":"Built-in"},{"id":"p2","name":"Corporate AD"}]}`,
			wantRoles: []string{"RADIUS user", "Site: HQ"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := unifiStubClient(t, func(req *http.Request) (int, string) {
				switch {
				case strings.HasSuffix(req.URL.Path, "/v1/sites"):
					return http.StatusOK, `{"count":1,"limit":200,"offset":0,"totalCount":1,` +
						`"data":[{"id":"site-a","name":"HQ","internalReference":"default"}]}`
				case strings.HasSuffix(req.URL.Path, "/radius/profiles"):
					return http.StatusOK, tc.profiles
				case strings.HasSuffix(req.URL.Path, "/list/account"):
					return http.StatusOK, `{"meta":{"rc":"ok"},"data":[{"_id":"ra1","name":"alice"}]}`
				default:
					return http.StatusOK, `{"count":0,"limit":200,"offset":0,"totalCount":0,"data":[]}`
				}
			})

			records, err := NewUniFiDriver(client, unifiTestLogger(), "https://api.ui.com/x").
				ListAccounts(context.Background())
			require.NoError(t, err)

			users := unifiRecordsWithPrefix(records, "site/site-a/radius-user/")
			require.Len(t, users, 1)
			assert.Equal(t, tc.wantRoles, users[0].Roles)
		})
	}
}

// TestUniFiMACKey pins that the same device keys identically however UniFi
// formats its address, so a formatting change cannot read as one account
// removed and another added.
func TestUniFiMACKey(t *testing.T) {
	t.Parallel()

	for _, mac := range []string{
		"94:2a:6f:26:c6:ca",
		"94-2A-6F-26-C6-CA",
		"942A6F26C6CA",
		"94 2a 6f 26 c6 ca",
		"942a.6f26.c6ca",
	} {
		t.Run(mac, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "942a6f26c6ca", unifiMACKey(mac))
		})
	}
}

// TestUniFiSiteLabel pins that a role always names something: the display name
// first, then the internal reference older UniFi APIs use, then the ID.
func TestUniFiSiteLabel(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "HQ", unifiSiteLabel(unifiSite{ID: "x", Name: " HQ ", InternalReference: "default"}))
	assert.Equal(t, "default", unifiSiteLabel(unifiSite{ID: "x", InternalReference: "default"}))
	assert.Equal(t, "x", unifiSiteLabel(unifiSite{ID: "x"}))
}

// TestUniFiNameResolver pins the single-site rule: a console exposes no name of
// its own, so it is named after its site when it has exactly one.
func TestUniFiNameResolver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/unifi", "UNIFI_API_KEY")
	client := newVCRClientWithHeader(rec, "X-API-KEY", os.Getenv("UNIFI_API_KEY"))

	baseURL, err := UniFiConsoleBaseURL(unifiTestAPIBase, unifiTestConsoleID)
	require.NoError(t, err)

	name, err := NewUniFiNameResolver(client, unifiTestLogger(), baseURL).ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "HQ", name)
}

// TestUniFiNameResolverMultiSite pins the other half of the rule: a multi-site
// console has no single right label, and an empty name leaves the source with
// the generic provider name rather than an arbitrary pick.
func TestUniFiNameResolverMultiSite(t *testing.T) {
	t.Parallel()

	body := `{"count":2,"limit":200,"offset":0,"totalCount":2,"data":[` +
		`{"id":"aaaa","name":"HQ","internalReference":"default"},` +
		`{"id":"bbbb","name":"Branch","internalReference":"site2"}]}`

	client := unifiStubClient(t, func(*http.Request) (int, string) {
		return http.StatusOK, body
	})

	name, err := NewUniFiNameResolver(client, unifiTestLogger(), "https://api.ui.com/x").
		ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Empty(t, name)
}

// TestUniFiNameResolverTerminalStatus pins that a revoked key or a console the
// key cannot reach is classified terminal, so the source-name worker keeps the
// generic name instead of re-claiming the source on every poll.
func TestUniFiNameResolverTerminalStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		status   int
		terminal bool
	}{
		{status: http.StatusUnauthorized, terminal: true},
		{status: http.StatusForbidden, terminal: true},
		{status: http.StatusBadRequest, terminal: true},
		{status: http.StatusInternalServerError, terminal: false},
	}

	for _, tc := range cases {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			t.Parallel()

			client := unifiStubClient(t, func(*http.Request) (int, string) {
				return tc.status, `{"statusCode":0}`
			})

			_, err := NewUniFiNameResolver(client, unifiTestLogger(), "https://api.ui.com/x").
				ResolveInstanceName(context.Background())
			require.Error(t, err)
			assert.Equal(t, tc.terminal, errors.Is(err, ErrTerminalNameResolution))
		})
	}
}

// TestUniFiDriverContextCancellation verifies that a context canceled mid-run
// aborts with the cancellation error rather than being mistaken for a complete,
// smaller sync.
func TestUniFiDriverContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if !strings.HasSuffix(req.URL.Path, "/v1/sites") {
				cancel()

				return nil, ctx.Err()
			}

			body := `{"count":1,"limit":200,"offset":0,"totalCount":1,` +
				`"data":[{"id":"site-a","name":"HQ"}]}`

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	}

	records, err := NewUniFiDriver(client, unifiTestLogger(), "https://api.ui.com/x").ListAccounts(ctx)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
	assert.Nil(t, records)
}

// unifiRecordsWithPrefix returns the records whose ExternalID starts with prefix.
func unifiRecordsWithPrefix(records []AccountRecord, prefix string) []AccountRecord {
	var out []AccountRecord

	for _, record := range records {
		if strings.HasPrefix(record.ExternalID, prefix) {
			out = append(out, record)
		}
	}

	return out
}

// unifiStubClient returns an *http.Client answering every request from respond,
// which maps a request to a status code and a JSON body.
func unifiStubClient(t *testing.T, respond func(*http.Request) (int, string)) *http.Client {
	t.Helper()

	return &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		status, body := respond(req)

		return &http.Response{
			StatusCode: status,
			Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})}
}
