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

package aws_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/cloud"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

const (
	// organizationA and organizationB are two real organization GIDs. The
	// isolation test needs them to differ in more than spelling, since a
	// per-organization issuer is what makes tenant isolation structural.
	organizationA = "e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhg"
	organizationB = "e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhh"

	roleARNA = "arn:aws:iam::123456789012:role/ProboAudit"
	roleARNB = "arn:aws:iam::210987654321:role/ProboAudit"
)

// rsaTestKey is generated once: a 2048-bit key is the smallest jose accepts and
// generating it dominates this package's runtime.
var rsaTestKey = sync.OnceValue(
	func() *rsa.PrivateKey {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}

		return key
	},
)

// issuerHost serves the two public documents probod serves in production, built
// from the same *identityfederation.Issuer that mints the tokens. Only the
// GID-parse and organization-existence checks of the real mux are absent, and
// neither affects what a cloud provider reads.
type issuerHost struct {
	issuer *identityfederation.Issuer
	// publishedJWKS overrides what the key set endpoint serves, to reproduce
	// signing with a kid that was never published.
	publishedJWKS *jose.JWKS
}

// newIssuerHost starts the issuer host and returns the issuer advertising it.
//
// The listener is bound before the issuer is built, because the issuer must
// advertise the URL it is reached through and there is no way to learn that URL
// after the fact without a data race on the handler.
func newIssuerHost(t *testing.T, keyRing *jose.KeyRing) (*issuerHost, *identityfederation.Issuer) {
	t.Helper()

	host := &issuerHost{}
	srv := httptest.NewUnstartedServer(host)

	base, err := baseurl.Parse("http://" + srv.Listener.Addr().String() + identityfederation.PathPrefix)
	require.NoError(t, err)

	issuer, err := identityfederation.NewIssuer(base, keyRing, identityfederation.DefaultTokenTTL)
	require.NoError(t, err)

	host.issuer = issuer

	srv.Start()
	t.Cleanup(srv.Close)

	return host, issuer
}

func (h *issuerHost) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, identityfederation.PathPrefix+"/")

	organizationID, rest, ok := strings.Cut(path, "/")
	if !ok {
		http.NotFound(w, r)

		return
	}

	parsed, err := gid.ParseGID(organizationID)
	if err != nil || parsed.EntityType() != coredata.OrganizationEntityType {
		http.NotFound(w, r)

		return
	}

	switch rest {
	case ".well-known/openid-configuration":
		metadata, err := h.issuer.Metadata(parsed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		writeJSON(w, metadata)
	case "jwks":
		if h.publishedJWKS != nil {
			writeJSON(w, h.publishedJWKS)

			return
		}

		writeJSON(w, h.issuer.JWKS())
	default:
		http.NotFound(w, r)
	}
}

func writeJSON(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

func testKeyRing(t *testing.T, kid string) *jose.KeyRing {
	t.Helper()

	keyRing, err := jose.NewKeyRing(
		[]jose.SigningKey{
			{PrivateKey: rsaTestKey(), KID: kid, Active: true},
		},
	)
	require.NoError(t, err)

	return keyRing
}

func loopbackClient() *http.Client {
	// The issuer and STS under test are httptest servers on loopback, which
	// production SSRF protection correctly refuses. Never use this exemption
	// outside a test.
	return httpclient.DefaultPooledClient(
		httpclient.WithSSRFProtection(),
		httpclient.WithSSRFAllowLoopback(),
	)
}

func parseOrganizationGID(t *testing.T, raw string) gid.GID {
	t.Helper()

	organizationID, err := gid.ParseGID(raw)
	require.NoError(t, err)
	require.Equal(t, coredata.OrganizationEntityType, organizationID.EntityType())

	return organizationID
}

// TestNewSession_AssumeRoleWithWebIdentity is the end-to-end credential
// exchange: a real issuer mints the assertion, the fake STS fetches the
// discovery document and key set it names, verifies the RS256 signature, and
// evaluates the trust policy the customer's stack would have written.
func TestNewSession_AssumeRoleWithWebIdentity(t *testing.T) {
	t.Parallel()

	_, issuer := newIssuerHost(t, testKeyRing(t, "a"))
	organizationID := parseOrganizationGID(t, organizationA)

	issuerURL, err := issuer.IssuerURL(organizationID)
	require.NoError(t, err)

	sts, endpoint := newFakeSTS(
		t,
		trustPolicy{
			roleARN: roleARNA,
			conditions: conditionKeysFor(
				issuerURL,
				organizationID.String(),
				identityfederation.AudienceAWS,
			),
		},
	)

	session, err := cloudaws.NewSession(
		issuer,
		organizationID,
		roleARNA,
		cloudaws.WithSTSEndpoint(endpoint),
		cloudaws.WithHTTPClient(loopbackClient()),
	)
	require.NoError(t, err)

	assert.Equal(t, cloud.AWS, session.Cloud())
	assert.Equal(t, "123456789012", session.AccountID(), "account comes from the role ARN")

	credentials, err := session.Config().Credentials.Retrieve(context.Background())
	require.NoError(t, err)
	assert.Equal(t, fakeAccessKeyID, credentials.AccessKeyID)
	assert.Equal(t, fakeSecretAccessKey, credentials.SecretAccessKey)
	assert.Equal(t, fakeSessionToken, credentials.SessionToken)

	calls := sts.callsMade()
	require.Len(t, calls, 1)
	assert.Equal(t, issuerURL, calls[0].issuer)
	assert.Equal(
		t,
		organizationID.String(),
		calls[0].subject,
		"sub is the bare organization GID",
	)
	assert.Equal(t, identityfederation.AudienceAWS, calls[0].audience)
	assert.Equal(
		t,
		"probo-"+organizationID.String(),
		calls[0].roleSessionName,
		"the session name attributes the call in the customer's CloudTrail",
	)
}

// TestNewSession_RejectsCrossTenantToken is the test releases gate on: one
// organization's identity must never reach another organization's role.
//
// It is checked twice over, because the design claims two independent barriers.
// Organization B minting against A's role fails on the sub condition even
// though both share an issuer host; B reaching A's role through B's own issuer
// fails because the condition keys name a different issuer entirely, which is
// what survives a customer wildcarding sub in their own Terraform.
func TestNewSession_RejectsCrossTenantToken(t *testing.T) {
	t.Parallel()

	t.Run("same issuer, different subject", func(t *testing.T) {
		t.Parallel()

		_, issuer := newIssuerHost(t, testKeyRing(t, "a"))

		organizationIDA := parseOrganizationGID(t, organizationA)
		organizationIDB := parseOrganizationGID(t, organizationB)

		issuerURLA, err := issuer.IssuerURL(organizationIDA)
		require.NoError(t, err)

		sts, endpoint := newFakeSTS(
			t,
			trustPolicy{
				roleARN: roleARNA,
				conditions: conditionKeysFor(
					issuerURLA,
					organizationIDA.String(),
					identityfederation.AudienceAWS,
				),
			},
		)

		// Organization B mints for itself, then aims at A's role.
		session, err := cloudaws.NewSession(
			issuer,
			organizationIDB,
			roleARNA,
			cloudaws.WithSTSEndpoint(endpoint),
			cloudaws.WithHTTPClient(loopbackClient()),
		)
		require.NoError(t, err)

		_, err = session.Config().Credentials.Retrieve(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "AccessDenied")

		calls := sts.callsMade()
		require.Len(t, calls, 1)
		assert.Equal(
			t,
			organizationIDB.String(),
			calls[0].subject,
			"B's token carries B's subject, which A's policy does not admit",
		)
	})

	// A trust policy pinned to organization A's issuer cannot be satisfied by a
	// token from B's issuer at all: STS finds no matching condition key before
	// it ever compares a subject. This is the barrier that holds even when a
	// customer wildcards sub.
	t.Run("different issuer", func(t *testing.T) {
		t.Parallel()

		_, issuer := newIssuerHost(t, testKeyRing(t, "a"))

		organizationIDA := parseOrganizationGID(t, organizationA)
		organizationIDB := parseOrganizationGID(t, organizationB)

		issuerURLA, err := issuer.IssuerURL(organizationIDA)
		require.NoError(t, err)

		// A's policy, deliberately wildcarded on sub the way a careless
		// Terraform edit would leave it.
		conditions := conditionKeysFor(issuerURLA, "*", identityfederation.AudienceAWS)

		_, endpoint := newFakeSTS(
			t,
			trustPolicy{roleARN: roleARNB, conditions: conditions},
		)

		session, err := cloudaws.NewSession(
			issuer,
			organizationIDB,
			roleARNB,
			cloudaws.WithSTSEndpoint(endpoint),
			cloudaws.WithHTTPClient(loopbackClient()),
		)
		require.NoError(t, err)

		_, err = session.Config().Credentials.Retrieve(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "AccessDenied")
		assert.Contains(t, err.Error(), "does not name this issuer")
	})
}

// TestNewSession_RejectsUnpublishedSigningKey covers the mistake CONTEXT.md
// singles out: signing with a kid absent from the published key set, which
// production surfaces as an opaque AccessDenied on every customer at once.
func TestNewSession_RejectsUnpublishedSigningKey(t *testing.T) {
	t.Parallel()

	host, issuer := newIssuerHost(t, testKeyRing(t, "rotated"))
	organizationID := parseOrganizationGID(t, organizationA)

	issuerURL, err := issuer.IssuerURL(organizationID)
	require.NoError(t, err)

	// Publish a key set that names a different kid, as if rotation had signed
	// with a key the JWKS did not carry yet.
	host.publishedJWKS = &jose.JWKS{
		Keys: []jose.JWK{
			jose.RSAPublicKeyToJWK(&rsaTestKey().PublicKey, "stale"),
		},
	}

	_, endpoint := newFakeSTS(
		t,
		trustPolicy{
			roleARN: roleARNA,
			conditions: conditionKeysFor(
				issuerURL,
				organizationID.String(),
				identityfederation.AudienceAWS,
			),
		},
	)

	session, err := cloudaws.NewSession(
		issuer,
		organizationID,
		roleARNA,
		cloudaws.WithSTSEndpoint(endpoint),
		cloudaws.WithHTTPClient(loopbackClient()),
	)
	require.NoError(t, err)

	_, err = session.Config().Credentials.Retrieve(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "InvalidIdentityToken")
}

func TestNewSession_Validation(t *testing.T) {
	t.Parallel()

	_, issuer := newIssuerHost(t, testKeyRing(t, "a"))
	organizationID := parseOrganizationGID(t, organizationA)

	tests := []struct {
		name        string
		roleARN     string
		wantMessage string
	}{
		{
			name:        "malformed role ARN",
			roleARN:     "ProboAudit",
			wantMessage: "cannot parse role ARN",
		},
		{
			name:        "role ARN without an account",
			roleARN:     "arn:aws:iam:::role/ProboAudit",
			wantMessage: "carries no account ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sts, endpoint := newFakeSTS(t)

			_, err := cloudaws.NewSession(
				issuer,
				organizationID,
				tt.roleARN,
				cloudaws.WithSTSEndpoint(endpoint),
				cloudaws.WithHTTPClient(loopbackClient()),
			)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantMessage)
			requireNoSTSCall(t, sts)
		})
	}
}
