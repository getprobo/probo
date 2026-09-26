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
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestOVHcloudRoles(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name string
		user ovhcloudUser
		want []string
	}{
		{
			name: "groups wins over the single group field",
			user: ovhcloudUser{Group: "ADMIN", Groups: []string{"DEFAULT", "billing"}},
			want: []string{"DEFAULT", "billing"},
		},
		{
			// auth.User carries both; `group` is only the main one, so it is
			// the fallback rather than the source of truth.
			name: "falls back to group when groups is empty",
			user: ovhcloudUser{Group: "ADMIN"},
			want: []string{"ADMIN"},
		},
		{
			name: "deduplicates and sorts",
			user: ovhcloudUser{Groups: []string{"ops", "ADMIN", "ops"}},
			want: []string{"ADMIN", "ops"},
		},
		{
			name: "no group at all",
			user: ovhcloudUser{},
			want: []string{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, ovhcloudRoles(tt.user))
		})
	}
}

func TestOVHcloudIsAdmin(t *testing.T) {
	t.Parallel()

	roles := map[string]string{
		"ADMIN":         "ADMIN",
		"Billing Admin": "REGULAR",
		"ops":           "admin",
		"DEFAULT":       "REGULAR",
		"UNPRIVILEGED":  "UNPRIVILEGED",
	}

	for _, tt := range []struct {
		name string
		user ovhcloudUser
		want *bool
	}{
		{"admin group", ovhcloudUser{Groups: []string{"ADMIN"}}, new(true)},
		{"admin role under an unrelated name", ovhcloudUser{Groups: []string{"ops"}}, new(true)},
		{"name contains admin but role does not", ovhcloudUser{Groups: []string{"Billing Admin"}}, new(false)},
		{"admin via one of several groups", ovhcloudUser{Groups: []string{"DEFAULT", "ADMIN"}}, new(true)},
		{"every group resolved and none is admin", ovhcloudUser{Groups: []string{"DEFAULT", "UNPRIVILEGED"}}, new(false)},
		{"no groups at all is a confirmed non-admin", ovhcloudUser{}, new(false)},
		{"unknown group leaves it unknown", ovhcloudUser{Groups: []string{"ghost"}}, nil},
		{"admin wins over an unresolvable sibling", ovhcloudUser{Groups: []string{"ghost", "ADMIN"}}, new(true)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, ovhcloudIsAdmin(tt.user, roles))
		})
	}
}

func TestOVHcloudAuthMethod(t *testing.T) {
	t.Parallel()

	for kind, want := range map[string]coredata.AccessReviewEntryAuthMethod{
		"PROVIDER": coredata.AccessReviewEntryAuthMethodSSO,
		"ACCOUNT":  coredata.AccessReviewEntryAuthMethodPassword,
		"USER":     coredata.AccessReviewEntryAuthMethodPassword,
		"":         coredata.AccessReviewEntryAuthMethodUnknown,
	} {
		t.Run("kind="+kind, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, want, ovhcloudAuthMethod(ovhcloudSignIn{kind: kind}))
		})
	}
}

func TestOVHcloudServiceAccountRecord(t *testing.T) {
	t.Parallel()

	urn := "urn:v1:eu:identity:credential:ab1234-ovh/oauth2-EU.1111"
	got := ovhcloudServiceAccountRecord(ovhcloudOAuth2Client{
		ClientID: "EU.1111", Name: "ci-deploy", Flow: "CLIENT_CREDENTIALS", Identity: &urn,
	})

	assert.Equal(t, urn, got.ExternalID)
	assert.Equal(t, "ci-deploy", got.FullName)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, got.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodServiceAccount, got.AuthMethod)
	assert.Nil(t, got.IsAdmin)
	assert.Empty(t, got.Email)

	// With no identity URN the client id is the only stable handle.
	got = ovhcloudServiceAccountRecord(ovhcloudOAuth2Client{ClientID: "EU.2222", Flow: "CLIENT_CREDENTIALS"})
	assert.Equal(t, "EU.2222", got.ExternalID)
}

func TestOVHcloudFederatedRecord(t *testing.T) {
	t.Parallel()

	last := time.Date(2026, 9, 21, 9, 0, 0, 0, time.UTC)
	got := ovhcloudFederatedRecord("alice@corp.example", ovhcloudSignIn{
		lastLogin: last, mfa: coredata.MFAStatusEnabled, kind: "PROVIDER",
	})

	assert.Equal(t, "alice@corp.example", got.ExternalID)
	assert.Equal(t, "alice@corp.example", got.Email)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSO, got.AuthMethod)
	assert.Equal(t, coredata.MFAStatusEnabled, got.MFAStatus)
	require.NotNil(t, got.LastLogin)
	assert.Nil(t, got.Active)
	assert.Nil(t, got.IsAdmin)

	// A non-email subject is still a usable identifier, just not an email.
	got = ovhcloudFederatedRecord("CORP\\alice", ovhcloudSignIn{kind: "PROVIDER"})
	assert.Empty(t, got.Email)
	assert.Equal(t, "CORP\\alice", got.ExternalID)
}

func TestOVHcloudMFAStatus(t *testing.T) {
	t.Parallel()

	for mfaType, want := range map[string]coredata.MFAStatus{
		"TOTP":        coredata.MFAStatusEnabled,
		"U2F":         coredata.MFAStatusEnabled,
		"SMS":         coredata.MFAStatusEnabled,
		"BACKUP_CODE": coredata.MFAStatusEnabled,
		"MAIL":        coredata.MFAStatusEnabled,
		"NONE":        coredata.MFAStatusDisabled,
		"UNKNOWN":     coredata.MFAStatusUnknown,
		"":            coredata.MFAStatusUnknown,
		"FUTURE_KIND": coredata.MFAStatusUnknown,
	} {
		t.Run(mfaType, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, want, ovhcloudMFAStatus(mfaType))
		})
	}
}

func TestOVHcloudURL(t *testing.T) {
	t.Parallel()

	base := "https://eu.api.ovh.com/1.0"

	endpoint, err := ovhcloudURL(base, "me", "identity", "user")
	require.NoError(t, err)
	assert.Equal(t, "https://eu.api.ovh.com/1.0/me/identity/user", endpoint)

	// A login is operator-controlled and reaches the path verbatim, so it must
	// not be able to climb out of the collection.
	endpoint, err = ovhcloudURL(base, "me", "identity", "user", "../../../me")
	require.NoError(t, err)
	assert.Equal(t, "https://eu.api.ovh.com/1.0/me/identity/user/..%2F..%2F..%2Fme", endpoint)

	endpoint, err = ovhcloudURL(base, "me", "identity", "user", "a b/c")
	require.NoError(t, err)
	assert.Equal(t, "https://eu.api.ovh.com/1.0/me/identity/user/a%20b%2Fc", endpoint)
}

func TestOVHcloudUserRecord(t *testing.T) {
	t.Parallel()

	roles := map[string]string{"ADMIN": "ADMIN", "DEFAULT": "REGULAR"}
	created := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	lastLogin := time.Date(2026, 9, 20, 8, 30, 0, 0, time.UTC)

	t.Run("active admin with a sign-in", func(t *testing.T) {
		t.Parallel()

		got := ovhcloudUserRecord(ovhcloudUser{
			Login:       "alice",
			Email:       "alice@example.com",
			Description: "Platform lead",
			Groups:      []string{"ADMIN"},
			Status:      "OK",
			Type:        "USER",
			URN:         "urn:v1:eu:identity:user:ab1234-ovh/alice",
			Creation:    &created,
		}, roles, ovhcloudSignIn{lastLogin: lastLogin, mfa: coredata.MFAStatusEnabled, kind: "USER"})

		assert.Equal(t, "urn:v1:eu:identity:user:ab1234-ovh/alice", got.ExternalID)
		assert.Equal(t, "alice@example.com", got.Email)
		assert.Equal(t, "Platform lead", got.JobTitle)
		assert.Equal(t, []string{"ADMIN"}, got.Roles)
		require.NotNil(t, got.Active)
		assert.True(t, *got.Active)
		require.NotNil(t, got.IsAdmin)
		assert.True(t, *got.IsAdmin)
		assert.Equal(t, coredata.MFAStatusEnabled, got.MFAStatus)
		assert.Equal(t, coredata.AccessReviewEntryAuthMethodPassword, got.AuthMethod)
		assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, got.AccountType)
		require.NotNil(t, got.LastLogin)
		assert.Equal(t, lastLogin, *got.LastLogin)
		assert.Equal(t, &created, got.CreatedAt)
	})

	t.Run("disabled user is inactive", func(t *testing.T) {
		t.Parallel()

		got := ovhcloudUserRecord(ovhcloudUser{Login: "bob", Status: "DISABLED", Groups: []string{"DEFAULT"}}, roles, ovhcloudSignIn{})

		require.NotNil(t, got.Active)
		assert.False(t, *got.Active)
		require.NotNil(t, got.IsAdmin)
		assert.False(t, *got.IsAdmin)
	})

	t.Run("password change required still authenticates", func(t *testing.T) {
		t.Parallel()

		got := ovhcloudUserRecord(ovhcloudUser{Login: "carol", Status: "PASSWORD_CHANGE_REQUIRED"}, roles, ovhcloudSignIn{})

		require.NotNil(t, got.Active)
		assert.True(t, *got.Active)
	})

	t.Run("service account type and login fallback for a missing urn", func(t *testing.T) {
		t.Parallel()

		got := ovhcloudUserRecord(ovhcloudUser{Login: "ci-runner", Type: "SERVICE"}, roles, ovhcloudSignIn{})

		assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, got.AccountType)
		assert.Equal(t, "ci-runner", got.ExternalID)
	})

	t.Run("no sign-in leaves MFA unknown and last login nil", func(t *testing.T) {
		t.Parallel()

		got := ovhcloudUserRecord(ovhcloudUser{Login: "dave"}, roles, ovhcloudSignIn{})

		assert.Equal(t, coredata.MFAStatusUnknown, got.MFAStatus)
		assert.Nil(t, got.LastLogin)
	})
}

func TestOVHcloudOwnerRecord(t *testing.T) {
	t.Parallel()

	// The owner is a root identity that /me/identity/user never lists, so the
	// record is synthesised from /me and is always an active admin.
	got := ovhcloudOwnerRecord(ovhcloudAccount{
		Nichandle: "ab1234-ovh",
		Email:     "owner@example.com",
		Firstname: "Ada",
		Name:      "Lovelace",
	}, ovhcloudSignIn{})

	assert.Equal(t, "ab1234-ovh", got.ExternalID)
	assert.Equal(t, "Ada Lovelace", got.FullName)
	assert.Equal(t, []string{"Account owner"}, got.Roles)
	require.NotNil(t, got.IsAdmin)
	assert.True(t, *got.IsAdmin)
	require.NotNil(t, got.Active)
	assert.True(t, *got.Active)
	assert.Equal(t, coredata.MFAStatusUnknown, got.MFAStatus)
	assert.Nil(t, got.LastLogin)
}

// TestOVHcloudDriver exercises the merge against a recorded account. It has no
// local users, so only the owner-synthesis path is covered here.
func TestOVHcloudDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/ovhcloud", "OVHCLOUD_TOKEN", sanitizeOVHcloud)
	client := newVCRClient(rec, bearerAuth(os.Getenv("OVHCLOUD_TOKEN")))

	driver := NewOVHcloudDriver(client, "https://eu.api.ovh.com/1.0")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)

	// The fixture account holds the owner, one client-credentials service
	// account and one classic API credential. Its two authorization-code
	// clients are deliberately absent: they bind to no identity and act as
	// whoever authorises them.
	require.Len(t, records, 3)

	owner := records[0]
	assert.Equal(t, "ab1234-ovh", owner.ExternalID)
	assert.Equal(t, "owner@example.com", owner.Email)
	assert.Equal(t, "Ada Lovelace", owner.FullName)
	assert.Equal(t, []string{"Account owner"}, owner.Roles)
	require.NotNil(t, owner.Active)
	assert.True(t, *owner.Active)
	require.NotNil(t, owner.IsAdmin)
	assert.True(t, *owner.IsAdmin)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, owner.AccountType)
	// The recorded audit log holds a real owner sign-in, so the reduction and
	// the MFA mapping are exercised against wire data rather than a fixture.
	require.NotNil(t, owner.LastLogin)
	assert.Equal(t, coredata.MFAStatusEnabled, owner.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodPassword, owner.AuthMethod)

	service := records[1]
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, service.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodServiceAccount, service.AuthMethod)
	assert.Empty(t, service.Email)
	// Its privilege lives in IAM policy the API does not expose per client.
	assert.Nil(t, service.IsAdmin)
	assert.NotEmpty(t, service.ExternalID)

	// A classic API credential: standing access that appears in no identity
	// endpoint, named after the application it belongs to and described by the
	// access rules it was granted.
	credential := records[2]
	assert.Equal(t, "635061199", credential.ExternalID)
	assert.Equal(t, "test", credential.FullName)
	assert.Equal(t, []string{"GET", "PUT"}, credential.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodAPIKey, credential.AuthMethod)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, credential.AccountType)
	require.NotNil(t, credential.Active)
	assert.True(t, *credential.Active, "a validated credential can still call the API")
	assert.NotNil(t, credential.CreatedAt)
	// It has never been used, so there is no last-use timestamp to report.
	assert.Nil(t, credential.LastLogin)
}

func TestOVHcloudNameResolver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/ovhcloud_name", "OVHCLOUD_TOKEN", sanitizeOVHcloud)
	client := newVCRClient(rec, bearerAuth(os.Getenv("OVHCLOUD_TOKEN")))

	name, err := NewOVHcloudNameResolver(client, "https://eu.api.ovh.com/1.0").ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "ab1234-ovh", name)
}

// sanitizeOVHcloud scrubs the account holder's real identity from a recorded
// cassette. GET /me answers with the owner's full billing record — name,
// address, customer code — and every response can embed the account handle or
// the recording credential's client id inside a URN. The structure is left
// exactly as OVHcloud returned it so the cassette still proves the decoding;
// only the values change.
func sanitizeOVHcloud(i *cassette.Interaction) error {
	// Client ids are addressed by URL, so scrubbing only the body would still
	// commit them.
	i.Request.URL = rewriteOVHcloudIdentifiers(i.Request.URL)

	// X-Iplb-Request-Id opens with the caller's public IP in hex
	// (53C7FE47 is 83.199.254.71), so the body sanitiser scrubbing the
	// decimal form is not enough. These headers carry nothing a replay needs.
	for _, header := range []string{"X-Iplb-Request-Id", "X-Iplb-Instance", "X-Ovh-Queryid"} {
		i.Response.Headers.Del(header)
	}

	if err := rewriteOVHcloudBody(i); err != nil {
		return err
	}

	return assertNoRealOVHcloudIdentifiers(i)
}

// rewriteOVHcloudBody substitutes identifiers inside a JSON body and scrubs the
// account holder's billing record. A non-JSON body is left alone, but it is
// still verified by the caller.
func rewriteOVHcloudBody(i *cassette.Interaction) error {
	var body any
	if err := json.Unmarshal([]byte(i.Response.Body), &body); err != nil {
		return nil //nolint:nilerr // not JSON: nothing to rewrite, still verified below
	}

	// Identifiers leak through free-form strings (URNs above all), so they are
	// rewritten wherever they appear rather than key by key.
	body = replaceOVHcloudIdentifiers(body)

	// The billing record is scrubbed only at the top level of /me. Doing it at
	// any depth would rewrite unrelated fields that happen to share a key —
	// an IAM policy's own "name", for instance.
	if account, ok := body.(map[string]any); ok {
		if _, isAccount := account["nichandle"]; isAccount {
			for key, replacement := range map[string]string{
				"email":        "owner@example.com",
				"spareEmail":   "spare@example.com",
				"firstname":    "Ada",
				"name":         "Lovelace",
				"address":      "1 Test Street",
				"city":         "Testville",
				"zip":          "00000",
				"phone":        "+33.000000000",
				"fax":          "+33.000000000",
				"customerCode": "0000-0000-00",
				"organisation": "Example Ltd",
			} {
				if _, present := account[key]; present {
					account[key] = replacement
				}
			}
		}
	}

	out, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cannot re-marshal sanitized ovhcloud body: %w", err)
	}

	i.Response.Body = string(out)

	return nil
}

// assertNoRealOVHcloudIdentifiers refuses to save an interaction still carrying
// a real identifier, on every path including a non-JSON body. The rewrite map
// alone is not enough — it is hand-maintained and an identifier minted later
// passes through — so shapes are matched structurally too.
func assertNoRealOVHcloudIdentifiers(i *cassette.Interaction) error {
	haystack := i.Response.Body + " " + i.Request.URL
	for _, values := range i.Response.Headers {
		haystack += " " + strings.Join(values, " ")
	}

	for real := range ovhcloudCassetteRewrites {
		if strings.Contains(haystack, real) {
			return fmt.Errorf("refusing to save ovhcloud cassette: a known real identifier survived sanitizing")
		}
	}

	for _, candidate := range ovhcloudIdentifierShapes.FindAllString(haystack, -1) {
		if !ovhcloudSyntheticIdentifier(candidate) {
			return fmt.Errorf("refusing to save ovhcloud cassette: an unrecognised real identifier survived sanitizing (add it to ovhcloudCassetteRewrites)")
		}
	}

	return nil
}

// ovhcloudCassetteRewrites maps every real identifier the fixture account
// exposes to its synthetic stand-in. The OAuth2 client ids matter most: two of
// them are Probo's own registered connector clients, and they appear in request
// URLs as well as in response bodies.
var ovhcloudCassetteRewrites = map[string]string{
	"sn609323-ovh":        "ab1234-ovh",
	"EU.28d43fdfc0f81aee": "EU.0000000000000000",
	"EU.605fa850d354b921": "EU.1111111111111111",
	"EU.d13f320e69d836ff": "EU.2222222222222222",
	"EU.a21f12032ce38dfa": "EU.3333333333333333",
	"83.199.254.71":       "203.0.113.1",
	// The classic API application's key. The driver never reads it, but it
	// rides along in the /me/api/application response.
	"3f1c18a4c1221636": "0000000000000000",
}

// ovhcloudIdentifierShapes matches the identifier formats OVHcloud mints that
// would identify a real account: an OAuth2 client id and a NIC handle.
// A bare 16-hex run matches an application key; an IPv4 pattern is deliberately
// NOT matched, because a browser version string ("Chrome/152.0.0.0") in a
// recorded user agent looks identical and would fail every save.
var ovhcloudIdentifierShapes = regexp.MustCompile(`EU\.[0-9a-f]{16}|[a-z]{2}[0-9]{1,8}-ovh|\b[0-9a-f]{16}\b`)

// ovhcloudSyntheticIdentifier reports whether a matched identifier is one of
// the stand-ins this file substitutes, rather than a real value.
func ovhcloudSyntheticIdentifier(v string) bool {
	for _, synthetic := range ovhcloudCassetteRewrites {
		if v == synthetic {
			return true
		}
	}

	return false
}

// rewriteOVHcloudIdentifiers applies the rewrites to a plain string.
func rewriteOVHcloudIdentifiers(v string) string {
	for real, synthetic := range ovhcloudCassetteRewrites {
		v = strings.ReplaceAll(v, real, synthetic)
	}

	return v
}

// replaceOVHcloudIdentifiers walks decoded JSON and rewrites identifiers inside
// every string, at any depth.
func replaceOVHcloudIdentifiers(node any) any {
	switch v := node.(type) {
	case map[string]any:
		for key, val := range v {
			v[key] = replaceOVHcloudIdentifiers(val)
		}

		return v
	case []any:
		for i, val := range v {
			v[i] = replaceOVHcloudIdentifiers(val)
		}

		return v
	case string:
		return rewriteOVHcloudIdentifiers(v)
	default:
		return node
	}
}

// TestOVHcloudDriverPopulatedAccount covers the per-identity join the recorded
// cassette cannot reach: the fixture account has no local users and no
// federated sign-in. Its cassette is hand-authored from OVHcloud's published
// schema, so it pins the merge logic, not the payload shape.
func TestOVHcloudDriverPopulatedAccount(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/ovhcloud_populated", "")
	client := newVCRClient(rec, bearerAuth("test"))

	records, err := NewOVHcloudDriver(client, "https://eu.api.ovh.com/1.0").ListAccounts(context.Background())
	require.NoError(t, err)

	// owner + three local users + one client-credentials client + two classic
	// API credentials + two federated. The authorization-code client is
	// excluded: it binds to no identity.
	require.Len(t, records, 9)

	byID := make(map[string]AccountRecord, len(records))
	for _, r := range records {
		byID[r.ExternalID] = r
	}

	// The owner's sign-in is keyed by kind rather than login, because an
	// ACCOUNT audit entry carries a null user. The reduction must keep the
	// most recent of the two owner sign-ins, not the first one seen.
	owner := byID["ab1234-ovh"]
	require.NotNil(t, owner.LastLogin)
	assert.Equal(t, time.Date(2026, 9, 20, 11, 15, 0, 0, time.UTC), owner.LastLogin.UTC())
	assert.Equal(t, coredata.MFAStatusEnabled, owner.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodPassword, owner.AuthMethod)

	// A local user resolved through the fan-out, with a sign-in that used no
	// second factor.
	alice := byID["urn:v1:eu:identity:user:ab1234-ovh/alice"]
	assert.Equal(t, "alice@example.com", alice.Email)
	assert.Equal(t, []string{"ADMIN"}, alice.Roles)
	require.NotNil(t, alice.IsAdmin)
	assert.True(t, *alice.IsAdmin)
	assert.Equal(t, coredata.MFAStatusDisabled, alice.MFAStatus)
	require.NotNil(t, alice.LastLogin)

	// PASSWORD_CHANGE_REQUIRED still authenticates, and no sign-in was
	// recorded for bob, so there is no evidence either way about MFA.
	bob := byID["urn:v1:eu:identity:user:ab1234-ovh/bob"]
	require.NotNil(t, bob.Active)
	assert.True(t, *bob.Active)
	require.NotNil(t, bob.IsAdmin)
	assert.False(t, *bob.IsAdmin)
	assert.Equal(t, coredata.MFAStatusUnknown, bob.MFAStatus)
	assert.Nil(t, bob.LastLogin)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, bob.AuthMethod)

	// carol is in a group the group listing does not resolve, so her privilege
	// is unknown rather than denied.
	carol := byID["urn:v1:eu:identity:user:ab1234-ovh/carol"]
	require.NotNil(t, carol.Active)
	assert.False(t, *carol.Active)
	assert.Nil(t, carol.IsAdmin)

	// The client-credentials client is a machine identity keyed by its IAM URN.
	service := byID["urn:v1:eu:identity:credential:ab1234-ovh/oauth2-EU.1111111111111111"]
	assert.Equal(t, "ci-deploy", service.FullName)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, service.AccountType)
	assert.Nil(t, service.IsAdmin)
	assert.NotContains(t, byID, "EU.2222222222222222")

	// Federated identities exist only as audit evidence.
	dave := byID["dave@corp.example.com"]
	assert.Equal(t, "dave@corp.example.com", dave.Email)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSO, dave.AuthMethod)
	assert.Equal(t, coredata.MFAStatusEnabled, dave.MFAStatus)
	assert.Nil(t, dave.Active)
	assert.Nil(t, dave.IsAdmin)

	// A federated subject that is not an email is still an identity, but Probo
	// must not invent an email address for it.
	erin := byID["CORP\\erin"]
	assert.Empty(t, erin.Email)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSO, erin.AuthMethod)

	// An expired classic API credential stays in the roster so a reviewer sees
	// it existed, but it can no longer call the API.
	expired := byID["900001"]
	assert.Equal(t, "terraform", expired.FullName)
	assert.Equal(t, []string{"GET /*"}, expired.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodAPIKey, expired.AuthMethod)
	require.NotNil(t, expired.Active)
	assert.False(t, *expired.Active)
	require.NotNil(t, expired.LastLogin)

	// A live credential OVHcloud's own support team created. The reviewer has
	// to see that the customer did not create this one, so it leads the roles.
	support := byID["900002"]
	require.NotNil(t, support.Active)
	assert.True(t, *support.Active)
	assert.Equal(t, []string{"Created by OVHcloud support", "GET /*", "POST /*"}, support.Roles)

	// Federated records are sorted, so the roster does not reorder per sync.
	assert.Equal(t, "CORP\\erin", records[7].ExternalID)
	assert.Equal(t, "dave@corp.example.com", records[8].ExternalID)
}

func TestOVHcloudAuditLogin(t *testing.T) {
	t.Parallel()

	// auth.User.login is the login SUFFIX, and a local user signs in as
	// "<nichandle>/<suffix>", so the audit log may name either form. The join
	// has to land on the suffix whichever arrives.
	for _, tt := range []struct {
		name, login, nichandle, want string
	}{
		{"prefixed with the handle", "ab1234-ovh/alice", "ab1234-ovh", "alice"},
		{"already a bare suffix", "alice", "ab1234-ovh", "alice"},
		{"a different handle is not stripped", "zz9999-ovh/alice", "ab1234-ovh", "zz9999-ovh/alice"},
		{"no handle known", "ab1234-ovh/alice", "", "ab1234-ovh/alice"},
		{"a federated subject is left alone", "alice@corp.example.com", "ab1234-ovh", "alice@corp.example.com"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, ovhcloudAuditLogin(tt.login, tt.nichandle))
		})
	}
}

func TestOVHcloudAPICredentialRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)
	app := ovhcloudAPIApplication{ApplicationID: 42, Name: "terraform", Description: "IaC", Status: "active"}

	t.Run("validated and unexpired is active", func(t *testing.T) {
		t.Parallel()

		got := ovhcloudAPICredentialRecord(ovhcloudAPICredential{
			CredentialID: 1, ApplicationID: 42, Status: "validated", Expiration: &future,
		}, app, now)

		require.NotNil(t, got.Active)
		assert.True(t, *got.Active)
		assert.Equal(t, "terraform", got.FullName)
		assert.Equal(t, "1", got.ExternalID)
	})

	t.Run("a passed expiry is inactive even while validated", func(t *testing.T) {
		t.Parallel()

		got := ovhcloudAPICredentialRecord(ovhcloudAPICredential{
			CredentialID: 2, Status: "validated", Expiration: &past,
		}, app, now)

		require.NotNil(t, got.Active)
		assert.False(t, *got.Active)
	})

	t.Run("no expiry means it never expires", func(t *testing.T) {
		t.Parallel()

		got := ovhcloudAPICredentialRecord(ovhcloudAPICredential{CredentialID: 3, Status: "validated"}, app, now)

		require.NotNil(t, got.Active)
		assert.True(t, *got.Active)
	})

	t.Run("a blocked application deactivates its credential", func(t *testing.T) {
		t.Parallel()

		blocked := ovhcloudAPIApplication{ApplicationID: 42, Name: "terraform", Status: "blocked"}
		got := ovhcloudAPICredentialRecord(ovhcloudAPICredential{CredentialID: 4, Status: "validated"}, blocked, now)

		require.NotNil(t, got.Active)
		assert.False(t, *got.Active)
	})

	t.Run("an unresolved application still names the row", func(t *testing.T) {
		t.Parallel()

		// A consumer key granted to somebody else's application: the reviewer
		// must see which application rather than a blank the console renders
		// as "N/A".
		got := ovhcloudAPICredentialRecord(ovhcloudAPICredential{
			CredentialID: 5, ApplicationID: 987, Status: "validated",
		}, ovhcloudAPIApplication{}, now)

		assert.Equal(t, "Application 987", got.FullName)
		require.NotNil(t, got.Active)
		assert.True(t, *got.Active, "an unknown application must not silently deactivate the credential")
	})

	t.Run("a support credential leads its roles with that fact", func(t *testing.T) {
		t.Parallel()

		got := ovhcloudAPICredentialRecord(ovhcloudAPICredential{
			CredentialID: 6, Status: "validated", OVHSupport: true,
			Rules: []struct {
				Method string `json:"method"`
				Path   string `json:"path"`
			}{{Method: "GET", Path: "/*"}},
		}, app, now)

		assert.Equal(t, []string{"Created by OVHcloud support", "GET /*"}, got.Roles)
	})
}
