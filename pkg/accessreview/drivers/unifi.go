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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	// Path elements below a console's UniFi Network application root. The
	// modern Integration API hangs off `integration`; the legacy application
	// API, which still owns the RADIUS user list, hangs off `api/s/<site>`.
	unifiIntegrationSegment = "integration"
	unifiVersionSegment     = "v1"
	unifiSitesSegment       = "sites"
	unifiWifiSegment        = "wifi"
	unifiBroadcastSegment   = "broadcasts"
	unifiHotspotSegment     = "hotspot"
	unifiVouchersSegment    = "vouchers"
	unifiRadiusSegment      = "radius"
	unifiProfilesSegment    = "profiles"
	unifiProxySegment       = "proxy"
	unifiNetworkSegment     = "network"

	// Legacy application API path elements for the RADIUS user list.
	unifiLegacyAPISegment     = "api"
	unifiLegacySiteSegment    = "s"
	unifiLegacyListSegment    = "list"
	unifiLegacyAccountSegment = "account"

	// unifiPageLimit is the documented maximum `limit` for every collection
	// except vouchers, which allows a larger page.
	unifiPageLimit = 200
	// unifiVoucherPageLimit is the documented maximum for
	// /hotspot/vouchers. Voucher counts dwarf every other collection (a
	// campaign can mint 1000 at a time), so the larger page keeps the run
	// inside the per-source deadline.
	unifiVoucherPageLimit = 1000
)

// UniFiConsoleBaseURL builds a console's UniFi Network application root under
// apiBase, the Site Manager console-proxy origin (e.g.
// https://api.ui.com/v1/connector/consoles). Both APIs the driver reads hang off
// this root: the Integration API under `integration`, and the legacy
// application API — still the only source of the RADIUS user list — under
// `api/s/<site>`.
//
// The console ID contains a colon (`<hardware id>:<numeric id>`), which is a
// legal path character, but it is escaped anyway so a malformed value cannot add
// a path segment.
//
// It lives here rather than in the provider registration so the driver, the
// name resolver and the connection probe all derive the same root from the same
// apiBase: an APIBase override must move all three together.
func UniFiConsoleBaseURL(apiBase, consoleID string) (string, error) {
	consoleID = strings.TrimSpace(consoleID)
	if consoleID == "" {
		return "", fmt.Errorf("cannot build unifi console URL: console_id is required")
	}

	baseURL, err := url.JoinPath(
		apiBase,
		url.PathEscape(consoleID),
		unifiProxySegment,
		unifiNetworkSegment,
	)
	if err != nil {
		return "", fmt.Errorf("cannot build unifi console URL: %w", err)
	}

	return baseURL, nil
}

// UniFiSitesURL builds the Integration API site-listing URL under baseURL, a
// console's Network application root produced by UniFiConsoleBaseURL. The
// connection probe uses it: listing sites is the cheapest call that proves both
// that the API key is valid and that it reaches the configured console.
func UniFiSitesURL(baseURL string) (string, error) {
	endpoint, err := url.JoinPath(baseURL, unifiIntegrationSegment, unifiVersionSegment, unifiSitesSegment)
	if err != nil {
		return "", fmt.Errorf("cannot build unifi sites URL: %w", err)
	}

	return endpoint, nil
}

// UniFiDriver reports who can reach a UniFi console's network resources, using
// a pre-authenticated HTTP client (the API key rides in X-API-KEY, attached by
// the connection transport). It walks every site the key can see and emits one
// record per access grant it finds.
//
// Every record is a USER: an identifiable holder of access. There are three
// kinds — RADIUS directory users, MAC addresses on an SSID's client-filtering
// list, and Hotspot vouchers. A RADIUS user's row names the SSIDs its directory
// unlocks, so one entry carries both the person and what they can reach.
//
// # What is deliberately NOT reported
//
// The driver emits no row for a resource — no SSID, no VPN server, no RADIUS
// profile. Those are the things being accessed, not parties holding access, and
// a campaign asks its reviewer to approve or revoke each entry, which is not a
// question an SSID can answer. Their useful signal is folded into the rows for
// the people instead: an SSID's security mode and on-air state qualify the
// device grants on it, and a RADIUS profile's SSID list appears on its users.
//
// This has a cost worth stating plainly. Access granted by a credential that
// UniFi shares among everyone who holds it — a WPA-Personal passphrase, an L2TP
// or PPTP VPN secret — has no identifiable holder anywhere in the API, so it
// produces no records at all. A site whose networks are all PSK-based, or whose
// only remote access is a shared-secret VPN, will therefore yield an empty
// review rather than a row per shared credential.
//
// # Where the data comes from
//
// Everything except the RADIUS user list comes from the documented Network
// Integration API. That API has exactly one RADIUS path,
// /v1/sites/{siteId}/radius/profiles, and it returns only a profile's id, name
// and origin — never its users. The user list is therefore read from the LEGACY
// application API (GET api/s/{site}/list/account), which is what the Network
// application's own Settings > Profiles > RADIUS > Users screen reads.
//
// That endpoint is undocumented, so it is treated as best-effort: an auth or
// not-found answer skips the RADIUS users with a warning rather than failing
// the whole review (see fetchRadiusUsers). A review that silently omits
// accounts is a real hazard, so the warning names what was dropped — but
// failing every other grant because one undocumented endpoint moved would be
// worse.
//
// # What this API cannot answer
//
//   - Console administrators — the people who can sign in to the UniFi
//     configuration portal — are exposed by neither API path the driver reads.
//     A review of portal access has to come from the Ubiquiti SSO account
//     inventory instead.
//   - VPN membership. `VPN server overview` carries only {id, name, type,
//     enabled} with no link to a RADIUS profile, and the "VPN client access"
//     schemas describe live sessions rather than a roster. Nothing in the API
//     joins a person to a VPN, so no record claims to.
//   - An externally-hosted RADIUS profile keeps its users in that directory,
//     not on the console, so only profiles backed by the console's built-in
//     RADIUS server contribute user records.
//
// Data quality: UniFi models none of these grants with an email address or an
// MFA signal, so Email is always empty (records key on ExternalID alone) and
// MFAStatus is always Unknown.
type UniFiDriver struct {
	httpClient *http.Client
	logger     *log.Logger
	baseURL    string
}

var _ Driver = (*UniFiDriver)(nil)

// NewUniFiDriver builds a driver against baseURL, a console's Network
// application root as returned by UniFiConsoleBaseURL.
//
// Retries are layered on by COPYING the caller's client and swapping only its
// transport, rather than wrapping the transport in a fresh &http.Client. The
// connection's client carries SSRF protection in two places — the dial check on
// the transport and a CheckRedirect that refuses a redirect to another
// host/scheme/port — and a fresh client would keep the first while silently
// dropping the second, so a 302 off api.ui.com would carry the API key with it.
func NewUniFiDriver(httpClient *http.Client, logger *log.Logger, baseURL string) *UniFiDriver {
	retryClient := *httpClient
	retryClient.Transport = &retryRoundTripper{
		next:       httpClient.Transport,
		maxRetries: 3,
	}

	return &UniFiDriver{
		httpClient: &retryClient,
		logger:     logger,
		baseURL:    baseURL,
	}
}

// unifiPage is the envelope every Integration API collection returns.
// TotalCount is the full result size and drives pagination; Count is the size
// of this page.
type unifiPage[T any] struct {
	Count      int   `json:"count"`
	Data       []T   `json:"data"`
	Limit      int   `json:"limit"`
	Offset     int64 `json:"offset"`
	TotalCount int64 `json:"totalCount"`
}

// unifiLegacyEnvelope is the envelope the legacy application API returns. Its
// `meta.rc` carries the application-level verdict, which can be "error" on a
// 200 response, so it is checked rather than assumed.
type unifiLegacyEnvelope[T any] struct {
	Meta struct {
		RC  string `json:"rc"`
		Msg string `json:"msg"`
	} `json:"meta"`
	Data []T `json:"data"`
}

type unifiSite struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// InternalReference is the site's legacy short name (e.g. "default"). The
	// Integration API documents it as "internal unique name of the site used in
	// older APIs", and it is exactly what the legacy RADIUS path needs as its
	// {site} segment — the UUID does not work there.
	InternalReference string `json:"internalReference"`
}

type unifiWifiBroadcast struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Enabled               bool   `json:"enabled"`
	SecurityConfiguration struct {
		Type string `json:"type"`
		// RadiusConfiguration is present on the Enterprise variants (and on
		// non-Enterprise SSIDs using RADIUS MAC authentication). Its profile ID
		// is what links an SSID to the directory that authenticates its users.
		RadiusConfiguration *struct {
			ProfileID string `json:"profileId"`
		} `json:"radiusConfiguration"`
	} `json:"securityConfiguration"`
	// ClientFilteringPolicy is present only on the detail representation. A
	// nil pointer means the SSID admits any client that satisfies its
	// security configuration.
	ClientFilteringPolicy *struct {
		Action           string   `json:"action"`
		MACAddressFilter []string `json:"macAddressFilter"`
	} `json:"clientFilteringPolicy"`
}

type unifiRadiusProfile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// unifiRadiusUser is one entry of the console's built-in RADIUS directory, as
// the legacy api/s/{site}/list/account endpoint returns it.
//
// It deliberately does NOT decode the response's `x_password` field. That is
// the user's live RADIUS credential, and a review record is not the place for
// one — leaving it undecoded means it cannot reach an entry, a log line or an
// error message by accident.
type unifiRadiusUser struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
	// VLAN is the VLAN the directory assigns on successful authentication.
	// Legacy responses have spelled it as both a string and a number across
	// versions, so it is decoded leniently.
	VLAN unifiLenientString `json:"vlan"`
}

// unifiLenientString decodes a JSON value that a legacy endpoint may send as
// either a string or a number into its string form. An absent, null or
// otherwise unusable value decodes to "".
type unifiLenientString string

func (s *unifiLenientString) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*s = ""

		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = unifiLenientString(str)

		return nil
	}

	var number json.Number
	if err := json.Unmarshal(data, &number); err == nil {
		*s = unifiLenientString(number.String())

		return nil
	}

	// An unexpected shape (object, array, bool) is not worth failing a whole
	// review over: the field is descriptive, not identifying.
	*s = ""

	return nil
}

type unifiVoucher struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	CreatedAt            string `json:"createdAt"`
	ActivatedAt          string `json:"activatedAt"`
	Expired              bool   `json:"expired"`
	AuthorizedGuestCount int64  `json:"authorizedGuestCount"`
}

func (d *UniFiDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	sites, err := d.fetchSites(ctx)
	if err != nil {
		return nil, err
	}

	var records []AccountRecord

	for _, site := range sites {
		// The ID is the only identifier every site is guaranteed to carry and
		// the path segment for every Integration API call below. Dropping a
		// blank row would hide a whole site's grants and mark them removed on
		// the next review.
		siteID := strings.TrimSpace(site.ID)
		if siteID == "" {
			return nil, fmt.Errorf("cannot list unifi accounts: site with an empty id")
		}

		siteRecords, err := d.listSiteAccounts(ctx, site, siteID)
		if err != nil {
			return nil, err
		}

		records = append(records, siteRecords...)
	}

	return records, nil
}

// listSiteAccounts collects every access grant one site exposes.
//
// Every ExternalID it produces is scoped to the site. Resource IDs look
// console-global today, but a review keys entries on ExternalID: were any of
// them ever site-scoped instead, two sites' grants would collapse into one
// entry and a reviewer would approve one while seeing the other. Scoping costs
// nothing and makes that unrepresentable.
func (d *UniFiDriver) listSiteAccounts(ctx context.Context, site unifiSite, siteID string) ([]AccountRecord, error) {
	siteRole := "Site: " + unifiSiteLabel(site)
	sitePrefix := "site/" + siteID + "/"

	overviews, err := d.fetchWifiBroadcasts(ctx, siteID)
	if err != nil {
		return nil, err
	}

	profiles, err := d.fetchRadiusProfiles(ctx, siteID)
	if err != nil {
		return nil, err
	}

	vouchers, err := d.fetchVouchers(ctx, siteID)
	if err != nil {
		return nil, err
	}

	var (
		records []AccountRecord
		// ssidsByProfile is what lets a RADIUS user's row name the WiFi
		// networks its directory unlocks.
		ssidsByProfile = make(map[string][]string, len(profiles))
	)

	for _, overview := range overviews {
		broadcastID := strings.TrimSpace(overview.ID)
		if broadcastID == "" {
			return nil, fmt.Errorf("cannot list unifi accounts: wifi broadcast with an empty id")
		}

		// Both the client-filtering policy — the API's only per-SSID list of
		// who may join — and the RADIUS profile link exist on the detail
		// representation alone, so the overview is never enough.
		broadcast, err := d.fetchWifiBroadcast(ctx, siteID, broadcastID)
		if err != nil {
			return nil, err
		}

		if broadcast == nil {
			// Deleted between the list and this fetch: it is genuinely gone,
			// not a failure, and re-listing would race the same way.
			d.logger.WarnCtx(
				ctx,
				"unifi wifi broadcast vanished between list and detail fetch",
				log.String("site_id", siteID),
			)

			continue
		}

		ssid := unifiBroadcastLabel(*broadcast)

		if radius := broadcast.SecurityConfiguration.RadiusConfiguration; radius != nil {
			if id := strings.TrimSpace(radius.ProfileID); id != "" {
				ssidsByProfile[id] = append(ssidsByProfile[id], ssid)
			}
		}

		records = append(records, unifiWifiRecords(*broadcast, ssid, sitePrefix, siteRole)...)
	}

	// The directory is per-site and reached through the legacy API, which the
	// Integration API's UUID does not address — it needs the site's legacy
	// short name.
	radiusUsers, err := d.fetchRadiusUsers(ctx, site)
	if err != nil {
		return nil, err
	}

	// A user authenticates against the directory, not against one profile, so
	// the profile — and the SSIDs delegating to it — are named only when the
	// site leaves exactly one candidate. With two directories in play, nothing
	// in either API says which one holds a given username, and a guess would
	// read as fact on the reviewer's screen.
	var (
		soleProfile string
		soleSSIDs   []string
	)

	if len(profiles) == 1 {
		soleProfile = unifiProfileLabel(profiles[0])
		soleSSIDs = ssidsByProfile[profiles[0].ID]
	}

	for _, user := range radiusUsers {
		record, ok := unifiRadiusUserRecord(user, soleProfile, soleSSIDs, sitePrefix, siteRole)
		if !ok {
			continue
		}

		records = append(records, record)
	}

	for _, voucher := range vouchers {
		if strings.TrimSpace(voucher.ID) == "" {
			return nil, fmt.Errorf("cannot list unifi accounts: hotspot voucher with an empty id")
		}

		records = append(records, unifiVoucherRecord(voucher, sitePrefix, siteRole))
	}

	return records, nil
}

// unifiWifiRecords turns one WiFi broadcast into one USER record per MAC
// address on its client-filtering list — the API's only per-SSID statement of
// who may join.
//
// The broadcast itself gets no record. An SSID is a network, not somebody who
// holds access, and a row named after it asks a reviewer a question they cannot
// answer in approve/revoke terms. What that row did carry — the security mode,
// and whether the SSID is even on air — is folded into the device rows below
// instead, where it qualifies an actual grant.
func unifiWifiRecords(broadcast unifiWifiBroadcast, ssid, sitePrefix, siteRole string) []AccountRecord {
	if broadcast.ClientFilteringPolicy == nil {
		return nil
	}

	roles := []string{"WiFi: " + ssid, siteRole}
	if security := strings.TrimSpace(broadcast.SecurityConfiguration.Type); security != "" {
		roles = append(roles, "Security: "+security)
	}

	// ALLOW makes the list exhaustive — only these MACs may join. BLOCK
	// inverts it: the SSID admits everyone else, and the listed MACs are
	// explicitly denied, which is exactly what a false Active means.
	//
	// A disabled SSID grants nothing either way, so it settles Active for every
	// device on the list: the broadcast's own row used to be what reported that,
	// and dropping the signal with the row would have left a listed MAC reading
	// as live access to a network that is off the air.
	allowed := broadcast.Enabled &&
		strings.EqualFold(strings.TrimSpace(broadcast.ClientFilteringPolicy.Action), "ALLOW")

	// The same device listed twice — plainly, or once per format — normalizes
	// to one key. Emitting it twice would inflate the fetched-account count and
	// hand the campaign two rows that collapse into one entry on upsert.
	var (
		records []AccountRecord
		seen    = make(map[string]struct{}, len(broadcast.ClientFilteringPolicy.MACAddressFilter))
	)

	for _, mac := range broadcast.ClientFilteringPolicy.MACAddressFilter {
		mac = strings.TrimSpace(mac)
		if mac == "" {
			continue
		}

		key := unifiMACKey(mac)
		if _, duplicate := seen[key]; duplicate {
			continue
		}

		seen[key] = struct{}{}

		permitted := allowed

		records = append(records, AccountRecord{
			FullName: mac,
			// Cloned rather than shared: a slice handed to several records
			// with spare capacity turns any later append into a write the
			// siblings can see.
			Roles:     append([]string(nil), roles...),
			Active:    &permitted,
			MFAStatus: coredata.MFAStatusUnknown,
			// A MAC filter authenticates a device by its hardware address,
			// which is not a credential the holder proves.
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  unifiWifiExternalID(sitePrefix, broadcast.ID) + "/mac/" + key,
		})
	}

	return records
}

// unifiRadiusUserRecord turns one entry of the console's built-in RADIUS
// directory into a USER record. These are the accounts that authenticate to
// WPA-Enterprise SSIDs and to RADIUS-backed VPN servers, so they are the
// closest thing UniFi has to a person.
//
// ssids are the SSIDs that delegate authentication to the user's directory, and
// they carry the whole answer to "which WiFi networks can this person reach".
// They are named on the person's own row rather than on a row for the directory,
// so a reviewer reads the grant without cross-referencing a second entry.
//
// It reports false when the entry carries neither a username nor an ID, which
// would leave nothing stable to key on.
func unifiRadiusUserRecord(
	user unifiRadiusUser,
	soleProfile string,
	ssids []string,
	sitePrefix, siteRole string,
) (AccountRecord, bool) {
	name := strings.TrimSpace(user.Name)
	id := strings.TrimSpace(user.ID)

	if name == "" && id == "" {
		return AccountRecord{}, false
	}

	// The username is the directory's own key and is what an administrator
	// recognises; the opaque _id is the fallback.
	external := id
	if external == "" {
		external = unifiMACKey(name)
	}

	if name == "" {
		name = id
	}

	roles := []string{"RADIUS user", siteRole}
	if soleProfile != "" {
		roles = append(roles, "RADIUS profile: "+soleProfile)
	}

	// Sorted so the role list does not reshuffle between campaigns purely
	// because the API returned the broadcasts in a different order.
	networks := append([]string(nil), ssids...)
	sort.Strings(networks)

	for _, ssid := range networks {
		roles = append(roles, "WiFi: "+ssid)
	}

	if vlan := strings.TrimSpace(string(user.VLAN)); vlan != "" {
		roles = append(roles, "VLAN: "+vlan)
	}

	return AccountRecord{
		FullName:  name,
		Roles:     roles,
		MFAStatus: coredata.MFAStatusUnknown,
		// A RADIUS account authenticates with a password held by the console's
		// built-in directory. Active is left nil: the legacy entry carries no
		// enabled/disabled state — an account exists or it does not.
		AuthMethod:  coredata.AccessReviewEntryAuthMethodPassword,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:  sitePrefix + "radius-user/" + external,
	}, true
}

// unifiVoucherRecord turns one Hotspot voucher into a USER record. A voucher
// is a named, time-boxed grant of guest network access, and its note is the
// only identity UniFi records for the guest who holds it.
func unifiVoucherRecord(voucher unifiVoucher, sitePrefix, siteRole string) AccountRecord {
	name := strings.TrimSpace(voucher.Name)
	if name == "" {
		name = voucher.ID
	}

	roles := []string{"Hotspot voucher", siteRole}
	if voucher.AuthorizedGuestCount > 0 {
		roles = append(roles, "Guests authorized: "+strconv.FormatInt(voucher.AuthorizedGuestCount, 10))
	}

	// `expired` is the API's own verdict on whether the voucher still grants
	// access, which is precisely the Active signal.
	usable := !voucher.Expired

	return AccountRecord{
		FullName:  name,
		Roles:     roles,
		Active:    &usable,
		MFAStatus: coredata.MFAStatusUnknown,
		// The guest types the voucher code to get on the network; it is a
		// shared secret, not a delegated identity.
		AuthMethod:  coredata.AccessReviewEntryAuthMethodPassword,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		CreatedAt:   parseRFC3339Ptr(voucher.CreatedAt),
		// activatedAt is when the first guest redeemed the voucher — the only
		// use timestamp UniFi keeps, so it stands in for a last login.
		LastLogin:  parseRFC3339Ptr(voucher.ActivatedAt),
		ExternalID: sitePrefix + "voucher/" + voucher.ID,
	}
}

// unifiWifiExternalID is the identity prefix a WiFi broadcast and its
// filtered MAC addresses share, under the site they belong to.
func unifiWifiExternalID(sitePrefix, broadcastID string) string {
	return sitePrefix + "wifi/" + broadcastID
}

// unifiMACKey normalizes a MAC address for use inside an ExternalID: lower
// case with separators removed, so the same device keys identically across
// campaigns even if UniFi changes how it formats the address. The record's
// FullName keeps the address as the API returned it.
func unifiMACKey(mac string) string {
	var b strings.Builder

	for _, r := range strings.ToLower(mac) {
		if r == ':' || r == '-' || r == '.' || r == ' ' {
			continue
		}

		b.WriteRune(r)
	}

	return b.String()
}

// unifiSiteLabel is the human name for a site, falling back to the internal
// reference older UniFi APIs use and finally to the ID, so a role always names
// something.
func unifiSiteLabel(site unifiSite) string {
	if name := strings.TrimSpace(site.Name); name != "" {
		return name
	}

	if ref := strings.TrimSpace(site.InternalReference); ref != "" {
		return ref
	}

	return strings.TrimSpace(site.ID)
}

// unifiBroadcastLabel is the SSID name, falling back to the ID.
func unifiBroadcastLabel(broadcast unifiWifiBroadcast) string {
	if name := strings.TrimSpace(broadcast.Name); name != "" {
		return name
	}

	return broadcast.ID
}

// unifiProfileLabel is the RADIUS profile name, falling back to the ID.
func unifiProfileLabel(profile unifiRadiusProfile) string {
	if name := strings.TrimSpace(profile.Name); name != "" {
		return name
	}

	return profile.ID
}

func (d *UniFiDriver) fetchSites(ctx context.Context) ([]unifiSite, error) {
	return fetchUniFiCollection[unifiSite](
		ctx, d, "sites", unifiPageLimit,
		unifiVersionSegment, unifiSitesSegment,
	)
}

func (d *UniFiDriver) fetchWifiBroadcasts(ctx context.Context, siteID string) ([]unifiWifiBroadcast, error) {
	return fetchUniFiCollection[unifiWifiBroadcast](
		ctx, d, "wifi broadcasts", unifiPageLimit,
		unifiVersionSegment, unifiSitesSegment, url.PathEscape(siteID), unifiWifiSegment, unifiBroadcastSegment,
	)
}

func (d *UniFiDriver) fetchRadiusProfiles(ctx context.Context, siteID string) ([]unifiRadiusProfile, error) {
	return fetchUniFiCollection[unifiRadiusProfile](
		ctx, d, "radius profiles", unifiPageLimit,
		unifiVersionSegment, unifiSitesSegment, url.PathEscape(siteID), unifiRadiusSegment, unifiProfilesSegment,
	)
}

func (d *UniFiDriver) fetchVouchers(ctx context.Context, siteID string) ([]unifiVoucher, error) {
	return fetchUniFiCollection[unifiVoucher](
		ctx, d, "hotspot vouchers", unifiVoucherPageLimit,
		unifiVersionSegment, unifiSitesSegment, url.PathEscape(siteID), unifiHotspotSegment, unifiVouchersSegment,
	)
}

// fetchWifiBroadcast reads one broadcast's detail representation, the only one
// carrying the client-filtering policy and the RADIUS profile link. A 404
// yields (nil, nil): the broadcast was deleted between the list and this call,
// which is a stable answer rather than a failure.
func (d *UniFiDriver) fetchWifiBroadcast(ctx context.Context, siteID, broadcastID string) (*unifiWifiBroadcast, error) {
	endpoint, err := url.JoinPath(
		d.baseURL,
		unifiIntegrationSegment,
		unifiVersionSegment, unifiSitesSegment, url.PathEscape(siteID),
		unifiWifiSegment, unifiBroadcastSegment, url.PathEscape(broadcastID),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot build unifi wifi broadcast URL: %w", err)
	}

	body, status, err := d.get(ctx, endpoint, "wifi broadcast")
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, nil
	}

	var broadcast unifiWifiBroadcast
	if err := json.Unmarshal(body, &broadcast); err != nil {
		return nil, fmt.Errorf("cannot decode unifi wifi broadcast response: %w", err)
	}

	return &broadcast, nil
}

// fetchRadiusUsers reads the console's built-in RADIUS directory for one site
// through the LEGACY application API, the only place UniFi exposes it — the
// Integration API's /radius/profiles returns a profile's id and name and
// nothing about its members.
//
// The legacy path is addressed by the site's legacy short name, not its UUID,
// so a site with no internalReference is skipped rather than requested with an
// identifier the endpoint would reject.
//
// Being undocumented, the endpoint is treated as best-effort: 401/403 (the API
// key does not reach the legacy API), 404 (the route moved or the console
// predates it) and an application-level `meta.rc` error all yield no users plus
// a warning naming what was dropped. Everything else — a 5xx, a decode failure
// — is still an error, because those are transient and a silently short answer
// would mark real accounts removed on the next campaign.
func (d *UniFiDriver) fetchRadiusUsers(ctx context.Context, site unifiSite) ([]unifiRadiusUser, error) {
	siteRef := strings.TrimSpace(site.InternalReference)
	if siteRef == "" {
		d.logger.WarnCtx(
			ctx,
			"skipping unifi radius users: site has no legacy short name to address the legacy API with",
			log.String("site_id", site.ID),
		)

		return nil, nil
	}

	endpoint, err := url.JoinPath(
		d.baseURL,
		unifiLegacyAPISegment, unifiLegacySiteSegment, url.PathEscape(siteRef),
		unifiLegacyListSegment, unifiLegacyAccountSegment,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot build unifi radius users URL: %w", err)
	}

	body, status, err := d.get(ctx, endpoint, "radius users")

	switch {
	case status == http.StatusUnauthorized,
		status == http.StatusForbidden,
		status == http.StatusNotFound:
		d.logger.WarnCtx(
			ctx,
			"unifi radius users unavailable, review will omit them",
			log.String("site", siteRef),
			log.Int("status", status),
		)

		return nil, nil
	case err != nil:
		return nil, err
	}

	var resp unifiLegacyEnvelope[unifiRadiusUser]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("cannot decode unifi radius users response: %w", err)
	}

	// The legacy API reports application errors in the body with a 200 status.
	if rc := strings.TrimSpace(resp.Meta.RC); rc != "" && !strings.EqualFold(rc, "ok") {
		d.logger.WarnCtx(
			ctx,
			"unifi radius users refused by the legacy API, review will omit them",
			log.String("site", siteRef),
			log.String("rc", rc),
		)

		return nil, nil
	}

	return resp.Data, nil
}

// fetchUniFiCollection walks one offset/limit paginated Integration API
// collection to completion. UniFi returns the full result size in totalCount,
// so the walk stops on reaching it rather than trusting a page to come back
// short.
//
// A 404 on the collection means the site does not expose that feature (a
// console with no gateway serves no VPN servers, one with no Hotspot serves no
// vouchers). It answers the same way on every run, so an empty result is
// stable and cannot make one campaign's grants look removed in the next; any
// other non-2xx aborts, because a partial fetch WOULD.
func fetchUniFiCollection[T any](
	ctx context.Context,
	d *UniFiDriver,
	what string,
	pageLimit int,
	segments ...string,
) ([]T, error) {
	base, err := url.JoinPath(d.baseURL, append([]string{unifiIntegrationSegment}, segments...)...)
	if err != nil {
		return nil, fmt.Errorf("cannot build unifi %s URL: %w", what, err)
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("cannot parse unifi %s URL: %w", what, err)
	}

	var (
		items  []T
		offset int64
	)

	for range maxPaginationPages {
		q := url.Values{}
		q.Set("limit", strconv.Itoa(pageLimit))
		q.Set("offset", strconv.FormatInt(offset, 10))

		endpoint := *parsed
		endpoint.RawQuery = q.Encode()

		body, status, err := d.get(ctx, endpoint.String(), what)
		if err != nil {
			return nil, err
		}

		if status == http.StatusNotFound {
			d.logger.WarnCtx(
				ctx,
				"unifi collection not available on this console",
				log.String("collection", what),
			)

			return nil, nil
		}

		var page unifiPage[T]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("cannot decode unifi %s response: %w", what, err)
		}

		items = append(items, page.Data...)

		// An empty page terminates regardless of what totalCount claims: a
		// count that never comes into reach would otherwise spin until the
		// pagination guard trips.
		if len(page.Data) == 0 || int64(len(items)) >= page.TotalCount {
			return items, nil
		}

		offset += int64(len(page.Data))
	}

	return nil, fmt.Errorf("cannot list all unifi %s: %w", what, ErrPaginationLimitReached)
}

// get performs an authenticated GET and returns the body together with the
// status code. The status is returned rather than swallowed so callers can
// treat some statuses as answers; every other non-2xx becomes an error here,
// wrapping ErrTerminalNameResolution on a permanent client error so the name
// resolver (which shares these calls) stops re-claiming a source it can never
// read.
//
// The status is returned on the error path too, so a caller that tolerates a
// particular status can inspect it without having to unwrap the error.
func (d *UniFiDriver) get(ctx context.Context, endpoint, what string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot create unifi %s request: %w", what, err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot execute unifi %s request: %w", what, err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode == http.StatusNotFound {
		return nil, httpResp.StatusCode, nil
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, httpResp.StatusCode, nameStatusError("unifi "+what, httpResp.StatusCode)
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, httpResp.StatusCode, fmt.Errorf("cannot read unifi %s response: %w", what, err)
	}

	return body, httpResp.StatusCode, nil
}

// unifiNameResolver names the source after the console's site. A UniFi console
// exposes no name of its own through the Integration API — only its sites — so
// a single-site console (the overwhelmingly common shape) is named after that
// site. A multi-site console has no one right label, and an empty name leaves
// the source with the generic provider name rather than an arbitrary pick.
type unifiNameResolver struct {
	httpClient *http.Client
	logger     *log.Logger
	baseURL    string
}

var _ NameResolver = (*unifiNameResolver)(nil)

// NewUniFiNameResolver resolves the console's name against baseURL, a console's
// Network application root as returned by UniFiConsoleBaseURL.
func NewUniFiNameResolver(httpClient *http.Client, logger *log.Logger, baseURL string) NameResolver {
	return &unifiNameResolver{httpClient: httpClient, logger: logger, baseURL: baseURL}
}

func (r *unifiNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	// Constructed rather than composite-literalled so the driver always
	// carries the resolver's base URL.
	driver := NewUniFiDriver(r.httpClient, r.logger, r.baseURL)

	sites, err := driver.fetchSites(ctx)
	if err != nil {
		return "", err
	}

	if len(sites) != 1 {
		return "", nil
	}

	return unifiSiteLabel(sites[0]), nil
}
