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

package identityfederation_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

const (
	// fixtureOrganizationID is the organization GID used throughout CONTEXT.md
	// and in the discovery golden file.
	fixtureOrganizationID = "e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhg"

	saasIssuerBase       = "https://proboidentity.com"
	selfHostedAppBase    = "http://localhost:8080"
	selfHostedIssuerBase = "http://localhost:8080/federation"
)

// rsaTestKeys generates the RSA keys shared by every test in this file. Key
// generation dominates the runtime of this package, and jose rejects a modulus
// below 2048 bits, so the keys are generated once and reused.
var rsaTestKeys = sync.OnceValue(
	func() []*rsa.PrivateKey {
		keys := make([]*rsa.PrivateKey, 3)

		for i := range keys {
			key, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				panic(err)
			}

			keys[i] = key
		}

		return keys
	},
)

func testKeyRing(t testing.TB, active ...bool) *jose.KeyRing {
	t.Helper()

	keys := rsaTestKeys()
	require.LessOrEqual(t, len(active), len(keys))

	signingKeys := make([]jose.SigningKey, 0, len(active))

	for i, isActive := range active {
		signingKeys = append(
			signingKeys,
			jose.SigningKey{
				PrivateKey: keys[i],
				KID:        kidForIndex(i),
				Active:     isActive,
			},
		)
	}

	keyRing, err := jose.NewKeyRing(signingKeys)
	require.NoError(t, err)

	return keyRing
}

func kidForIndex(i int) string {
	return string(rune('a' + i))
}

func testIssuer(t testing.TB, base string, active ...bool) *identityfederation.Issuer {
	t.Helper()

	if len(active) == 0 {
		active = []bool{true}
	}

	issuer, err := identityfederation.NewIssuer(
		baseurl.MustParse(base),
		testKeyRing(t, active...),
		identityfederation.DefaultTokenTTL,
	)
	require.NoError(t, err)

	return issuer
}

func fixtureOrganizationGID(t testing.TB) gid.GID {
	t.Helper()

	organizationID, err := gid.ParseGID(fixtureOrganizationID)
	require.NoError(t, err)
	require.Equal(t, coredata.OrganizationEntityType, organizationID.EntityType())

	return organizationID
}

func decodeTokenPayload(t testing.TB, token string, issuer *identityfederation.Issuer) map[string]any {
	t.Helper()

	payload, err := jose.VerifyJWTWithJWKS(token, issuer.JWKS())
	require.NoError(t, err)

	claims := map[string]any{}
	require.NoError(t, json.Unmarshal(payload, &claims))

	return claims
}

func decodeTokenHeader(t testing.TB, token string) jose.JWTHeader {
	t.Helper()

	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)

	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	require.NoError(t, err)

	header := jose.JWTHeader{}
	require.NoError(t, json.Unmarshal(raw, &header))

	return header
}

// Key material is validated by jose.NewKeyRing, which has its own tests; what
// remains here is what the issuer itself requires.
func TestNewIssuer_Errors(t *testing.T) {
	t.Parallel()

	t.Run(
		"nil base URL",
		func(t *testing.T) {
			t.Parallel()

			_, err := identityfederation.NewIssuer(nil, testKeyRing(t, true), identityfederation.DefaultTokenTTL)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "base URL is required")
		},
	)

	t.Run(
		"nil key ring",
		func(t *testing.T) {
			t.Parallel()

			_, err := identityfederation.NewIssuer(
				baseurl.MustParse(saasIssuerBase),
				nil,
				identityfederation.DefaultTokenTTL,
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "key ring is required")
		},
	)

	t.Run(
		"non-positive token TTL",
		func(t *testing.T) {
			t.Parallel()

			_, err := identityfederation.NewIssuer(
				baseurl.MustParse(saasIssuerBase),
				testKeyRing(t, true),
				0,
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "token TTL must be positive")
		},
	)
}

func TestIssuer_IssuerURL(t *testing.T) {
	t.Parallel()

	organizationID := fixtureOrganizationGID(t)

	tests := []struct {
		name        string
		base        string
		wantIssuer  string
		wantJWKSURI string
	}{
		{
			name:        "saas dedicated apex",
			base:        saasIssuerBase,
			wantIssuer:  "https://proboidentity.com/" + fixtureOrganizationID,
			wantJWKSURI: "https://proboidentity.com/" + fixtureOrganizationID + "/jwks",
		},
		{
			name:        "self hosted under the application base URL",
			base:        selfHostedIssuerBase,
			wantIssuer:  "http://localhost:8080/federation/" + fixtureOrganizationID,
			wantJWKSURI: "http://localhost:8080/federation/" + fixtureOrganizationID + "/jwks",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				t.Parallel()

				issuer := testIssuer(t, test.base)

				issuerURL, err := issuer.IssuerURL(organizationID)
				require.NoError(t, err)
				assert.Equal(t, test.wantIssuer, issuerURL)

				jwksURI, err := issuer.JWKSURI(organizationID)
				require.NoError(t, err)
				assert.Equal(t, test.wantJWKSURI, jwksURI)
			},
		)
	}
}

func TestIssuer_Token_Claims(t *testing.T) {
	t.Parallel()

	organizationID := fixtureOrganizationGID(t)
	issuer := testIssuer(t, saasIssuerBase)

	before := time.Now().Unix()

	token, err := issuer.Token(context.Background(), organizationID, identityfederation.AudienceAWS)
	require.NoError(t, err)

	after := time.Now().Unix()

	claims := decodeTokenPayload(t, token, issuer)

	// The claim set is closed: a future contributor must not be able to add a
	// claim without this test failing.
	assert.Len(t, claims, 7)

	for _, name := range []string{"iss", "sub", "aud", "iat", "nbf", "exp", "jti"} {
		assert.Contains(t, claims, name)
	}

	assert.Equal(t, "https://proboidentity.com/"+fixtureOrganizationID, claims["iss"])
	assert.Equal(t, fixtureOrganizationID, claims["sub"])
	assert.Equal(t, "sts.amazonaws.com", claims["aud"])
	assert.NotEmpty(t, claims["jti"])

	issuedAt := int64(claims["iat"].(float64))
	notBefore := int64(claims["nbf"].(float64))
	expiresAt := int64(claims["exp"].(float64))

	assert.GreaterOrEqual(t, issuedAt, before)
	assert.LessOrEqual(t, issuedAt, after)
	assert.Equal(t, issuedAt, notBefore)
	assert.Equal(t, int64(300), expiresAt-issuedAt)

	assert.Equal(t, "RS256", decodeTokenHeader(t, token).Algorithm)
}

func TestIssuer_Token_SelfVerification(t *testing.T) {
	t.Parallel()

	organizationID := fixtureOrganizationGID(t)
	issuer := testIssuer(t, saasIssuerBase)

	token, err := issuer.Token(context.Background(), organizationID, identityfederation.AudienceAWS)
	require.NoError(t, err)

	// This is the round trip AWS performs: fetch the published key set, find the
	// key named by the header, verify the signature. A kid or JWKS wiring bug
	// surfaces in production as an opaque AccessDenied, so it is checked here.
	payload, err := jose.VerifyJWTWithJWKS(token, issuer.JWKS())
	require.NoError(t, err)

	claims := identityfederation.Claims{}
	require.NoError(t, json.Unmarshal(payload, &claims))

	assert.Equal(t, organizationID.String(), claims.Subject)
}

func TestIssuer_Token_RejectsInvalidInput(t *testing.T) {
	t.Parallel()

	issuer := testIssuer(t, saasIssuerBase)

	t.Run(
		"zero identifier",
		func(t *testing.T) {
			t.Parallel()

			_, err := issuer.Token(context.Background(), gid.Nil, identityfederation.AudienceAWS)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "organization is required")
		},
	)

	t.Run(
		"identifier is not an organization",
		func(t *testing.T) {
			t.Parallel()

			connectorID := gid.New(gid.TenantID{}, coredata.ConnectorEntityType)

			_, err := issuer.Token(context.Background(), connectorID, identityfederation.AudienceAWS)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "not an organization")
		},
	)

	t.Run(
		"empty audience",
		func(t *testing.T) {
			t.Parallel()

			_, err := issuer.Token(context.Background(), fixtureOrganizationGID(t), "")

			require.Error(t, err)
			assert.Contains(t, err.Error(), "audience is required")
		},
	)
}

func TestIssuer_Token_SignsWithActiveKeysOnly(t *testing.T) {
	t.Parallel()

	organizationID := fixtureOrganizationGID(t)
	issuer := testIssuer(t, saasIssuerBase, false, true, true)

	seen := map[string]bool{}

	for range 10 {
		token, err := issuer.Token(context.Background(), organizationID, identityfederation.AudienceAWS)
		require.NoError(t, err)

		seen[decodeTokenHeader(t, token).KeyID] = true
	}

	assert.False(t, seen["a"], "inactive key must never sign")
	assert.True(t, seen["b"])
	assert.True(t, seen["c"])
}

func TestIssuer_JWKS_KeyRotation(t *testing.T) {
	t.Parallel()

	organizationID := fixtureOrganizationGID(t)
	keys := rsaTestKeys()

	// Key "a" was retired but stays published, which is what lets a token
	// already minted with it keep verifying while AWS holds a cached key set.
	issuer := testIssuer(t, saasIssuerBase, false, true)

	jwks := issuer.JWKS()
	require.Len(t, jwks.Keys, 2)

	published := map[string]string{}
	for _, key := range jwks.Keys {
		published[key.KeyID] = key.Algorithm
	}

	assert.Equal(t, map[string]string{"a": "RS256", "b": "RS256"}, published)

	retiredToken, err := jose.SignJWT(
		keys[0],
		"a",
		identityfederation.Claims{Subject: organizationID.String()},
	)
	require.NoError(t, err)

	_, err = jose.VerifyJWTWithJWKS(retiredToken, issuer.JWKS())
	assert.NoError(t, err, "a retired but published key must still verify")

	// Once the key is dropped from configuration it stops verifying.
	remainingKeyRing, err := jose.NewKeyRing(
		[]jose.SigningKey{{PrivateKey: keys[1], KID: "b", Active: true}},
	)
	require.NoError(t, err)

	withoutRetiredKey, err := identityfederation.NewIssuer(
		baseurl.MustParse(saasIssuerBase),
		remainingKeyRing,
		identityfederation.DefaultTokenTTL,
	)
	require.NoError(t, err)

	_, err = jose.VerifyJWTWithJWKS(retiredToken, withoutRetiredKey.JWKS())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot find signing key")
}

func TestIssuer_Metadata_GoldenFile(t *testing.T) {
	t.Parallel()

	organizationID := fixtureOrganizationGID(t)
	issuer := testIssuer(t, saasIssuerBase)

	metadata, err := issuer.Metadata(organizationID)
	require.NoError(t, err)

	encoded, err := json.MarshalIndent(metadata, "", "  ")
	require.NoError(t, err)

	want, err := os.ReadFile("testdata/discovery.json")
	require.NoError(t, err)

	// AWS rejects an OIDC provider whose discovery document omits any of the
	// five fields, so the whole document is pinned.
	assert.Equal(t, strings.TrimSpace(string(want)), string(encoded))
}

func TestIssuer_Metadata_SelfHosted(t *testing.T) {
	t.Parallel()

	organizationID := fixtureOrganizationGID(t)
	issuer := testIssuer(t, selfHostedIssuerBase)

	metadata, err := issuer.Metadata(organizationID)
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:8080/federation/"+fixtureOrganizationID, metadata.Issuer)
	assert.Equal(t, "http://localhost:8080/federation/"+fixtureOrganizationID+"/jwks", metadata.JWKSURI)
	assert.Equal(t, []string{"id_token"}, metadata.ResponseTypesSupported)
	assert.Equal(t, []string{"public"}, metadata.SubjectTypesSupported)
	assert.Equal(t, []string{"RS256"}, metadata.IDTokenSigningAlgValuesSupported)
}

func TestResolveIssuerBaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configured string
		appBaseURL string
		want       string
		wantErr    string
	}{
		{
			name:       "derived from the application base URL",
			configured: "",
			appBaseURL: selfHostedAppBase,
			want:       selfHostedIssuerBase,
		},
		{
			name:       "derived base tolerates http and a port",
			configured: "",
			appBaseURL: "http://probo.internal:9000",
			want:       "http://probo.internal:9000/federation",
		},
		{
			name:       "explicit saas apex",
			configured: saasIssuerBase,
			appBaseURL: "https://app.probo.com",
			want:       saasIssuerBase,
		},
		{
			name:       "explicit loopback base skips the AWS rules",
			configured: "http://localhost:9999/federation",
			appBaseURL: selfHostedAppBase,
			want:       "http://localhost:9999/federation",
		},
		{
			name:       "explicit public base must be https",
			configured: "http://proboidentity.com",
			appBaseURL: "https://app.probo.com",
			wantErr:    "scheme must be https",
		},
		{
			name:       "explicit public base must not carry a port",
			configured: "https://proboidentity.com:8443",
			appBaseURL: "https://app.probo.com",
			wantErr:    "port is not allowed",
		},
		{
			// The colon leaves Host non-empty while Port stays empty, so the
			// AWS port rule alone would advertise an unreachable issuer.
			name:       "hostless base is refused",
			configured: "https://:/federation",
			appBaseURL: "https://app.probo.com",
			wantErr:    "cannot parse identity federation issuer base URL",
		},
		{
			name:       "empty port is refused",
			configured: "https://proboidentity.com:",
			appBaseURL: "https://app.probo.com",
			wantErr:    "cannot parse identity federation issuer base URL",
		},
		{
			name:       "query string is refused",
			configured: "https://proboidentity.com?tenant=1",
			appBaseURL: "https://app.probo.com",
			wantErr:    "query string is not allowed",
		},
		{
			// A bare "?" leaves RawQuery empty, so only ForceQuery reveals it.
			// It survives into the minted issuer if it is not refused here.
			name:       "empty query marker is refused",
			configured: "https://proboidentity.com?",
			appBaseURL: "https://app.probo.com",
			wantErr:    "query string is not allowed",
		},
		{
			name:       "empty query marker after a path is refused",
			configured: "https://proboidentity.com/federation?",
			appBaseURL: "https://app.probo.com",
			wantErr:    "query string is not allowed",
		},
		{
			// The issuer is a public identifier a customer registers with their
			// cloud provider, and probod logs it at startup, so credentials must
			// never survive into it.
			name:       "userinfo is refused",
			configured: "https://user:pass@proboidentity.com", // trufflehog:ignore
			appBaseURL: "https://app.probo.com",
			wantErr:    "userinfo is not allowed",
		},
		{
			// url.JoinPath keeps the userinfo of the application base URL, so
			// the derived issuer carries it too.
			name:       "userinfo on the derived base is refused",
			configured: "",
			appBaseURL: "https://user:pass@app.probo.com", // trufflehog:ignore
			wantErr:    "userinfo is not allowed",
		},
		{
			// A root path would be served the OAuth2 server's discovery document.
			name:       "root path on the application host is refused",
			configured: "https://app.probo.com",
			appBaseURL: "https://app.probo.com",
			wantErr:    `must use the "/federation" path`,
		},
		{
			name:       "bare slash path on the application host is refused",
			configured: "https://app.probo.com/",
			appBaseURL: "https://app.probo.com",
			wantErr:    `must use the "/federation" path`,
		},
		{
			// Any other path on that host serves documents nothing answers for.
			name:       "unserved path on the application host is refused",
			configured: "https://app.probo.com/oidc",
			appBaseURL: "https://app.probo.com",
			wantErr:    `must use the "/federation" path`,
		},
		{
			name:       "mis-cased application host is still constrained",
			configured: "https://APP.probo.com/oidc",
			appBaseURL: "https://app.probo.com",
			wantErr:    `must use the "/federation" path`,
		},
		{
			name:       "served path on the application host is allowed",
			configured: "https://app.probo.com/federation",
			appBaseURL: "https://app.probo.com",
			want:       "https://app.probo.com/federation",
		},
		{
			name:       "served path with a trailing slash is allowed",
			configured: "https://app.probo.com/federation/",
			appBaseURL: "https://app.probo.com",
			want:       "https://app.probo.com/federation/",
		},
		{
			// The route tree sits beneath the application base URL's own path.
			name:       "application base URL with a path derives beneath it",
			configured: "",
			appBaseURL: "https://app.probo.com/probo",
			want:       "https://app.probo.com/probo/federation",
		},
		{
			name:       "explicit issuer must match the path under a based application URL",
			configured: "https://app.probo.com/federation",
			appBaseURL: "https://app.probo.com/probo",
			wantErr:    `must use the "/probo/federation" path`,
		},
		{
			// Another host arrives through an edge rewrite, so its path is free.
			name:       "a different host may use any path",
			configured: "https://proboidentity.com/anything",
			appBaseURL: "https://app.probo.com",
			want:       "https://proboidentity.com/anything",
		},
		{
			name:       "per-organization issuer must fit in 255 characters",
			configured: "https://proboidentity.com/" + strings.Repeat("x", 230),
			appBaseURL: "https://app.probo.com",
			wantErr:    "maximum is 255",
		},
		{
			// Percent-encoding expands the path, so the raw length of the base
			// understates the issuer a customer registers: 50 U+00E9 characters
			// are 100 UTF-8 bytes, the configured URL is 126 bytes, and the
			// escaped per-organization issuer serializes to 363 characters.
			name:       "length is measured on the escaped issuer",
			configured: "https://proboidentity.com/" + strings.Repeat("é", 50),
			appBaseURL: "https://app.probo.com",
			wantErr:    "maximum is 255",
		},
		{
			// url.JoinPath collapses the trailing slash, so this base serializes
			// to exactly 255 characters and must be accepted. Summing the raw
			// parts would count the slash twice and reject it at 256.
			name:       "trailing slash does not consume the budget",
			configured: "https://proboidentity.com/" + strings.Repeat("x", 192) + "/",
			appBaseURL: "https://app.probo.com",
			want:       "https://proboidentity.com/" + strings.Repeat("x", 192) + "/",
		},
		{
			name:       "relative base is refused",
			configured: "proboidentity.com",
			appBaseURL: "https://app.probo.com",
			wantErr:    "cannot parse identity federation issuer base URL",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				t.Parallel()

				resolved, err := identityfederation.ResolveIssuerBaseURL(
					test.configured,
					baseurl.MustParse(test.appBaseURL),
				)

				if test.wantErr != "" {
					require.Error(t, err)
					assert.Contains(t, err.Error(), test.wantErr)

					return
				}

				require.NoError(t, err)
				assert.Equal(t, test.want, resolved.String())
			},
		)
	}
}

func TestResolveIssuerBaseURL_RequiresApplicationBaseURL(t *testing.T) {
	t.Parallel()

	_, err := identityfederation.ResolveIssuerBaseURL(saasIssuerBase, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "application base URL is required")
}

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		issuerBaseURL string
		wantErr       string
	}{
		{
			name:          "public https apex is registrable",
			issuerBaseURL: saasIssuerBase,
		},
		{
			name:          "self-hosted https host with a path is registrable",
			issuerBaseURL: "https://compliance.acme.com/federation",
		},
		{
			// Never registered, so they stay quiet.
			name:          "loopback host is not reported",
			issuerBaseURL: selfHostedIssuerBase,
		},
		{
			name:          "loopback address is not reported",
			issuerBaseURL: "http://127.0.0.1:8080/federation",
		},
		{
			// The two shapes a self-hoster realistically trips over.
			name:          "plain http is reported",
			issuerBaseURL: "http://compliance.acme.com/federation",
			wantErr:       "scheme must be https",
		},
		{
			name:          "a port is reported",
			issuerBaseURL: "https://compliance.acme.com:8443/federation",
			wantErr:       "port is not allowed",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				t.Parallel()

				err := identityfederation.ValidateConfig(
					baseurl.MustParse(test.issuerBaseURL),
				)

				if test.wantErr == "" {
					assert.NoError(t, err)

					return
				}

				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
			},
		)
	}
}

func TestValidateConfig_NilIssuer(t *testing.T) {
	t.Parallel()

	assert.NoError(t, identityfederation.ValidateConfig(nil))
}

func TestResolveIssuerBaseURL_SaasIssuerHasHeadroom(t *testing.T) {
	t.Parallel()

	resolved, err := identityfederation.ResolveIssuerBaseURL(saasIssuerBase, baseurl.MustParse("https://app.probo.com"))
	require.NoError(t, err)

	issuer := testIssuer(t, resolved.String())

	issuerURL, err := issuer.IssuerURL(fixtureOrganizationGID(t))
	require.NoError(t, err)

	// AWS caps an OIDC provider URL at 255 characters. Assert the headroom
	// rather than an exact length, so that changing the apex cannot break a
	// test whose subject is the limit.
	const maxAWSIssuerURLLength = 255

	assert.Less(t, len(issuerURL), maxAWSIssuerURLLength/2)
}
