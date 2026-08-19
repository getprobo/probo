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
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/cloud"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

const (
	// Two organizations on the same Probo deployment: they share signing keys, so
	// only the issuer and subject claims separate them.
	organizationAID = "e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhg"
	organizationBID = "e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhh"

	issuerBase = "https://proboidentity.com"

	accountID = "123456789012"
	roleARN   = "arn:aws:iam::123456789012:role/ProboAudit"
)

// Generated once: key generation dominates this package's runtime and jose
// refuses a modulus below 2048 bits.
var rsaTestKey = sync.OnceValue(
	func() *rsa.PrivateKey {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}

		return key
	},
)

func testKeyRing(t *testing.T, privateKey *rsa.PrivateKey) *jose.KeyRing {
	t.Helper()

	keyRing, err := jose.NewKeyRing(
		[]jose.SigningKey{{PrivateKey: privateKey, KID: "a", Active: true}},
	)
	require.NoError(t, err)

	return keyRing
}

func testIssuer(t *testing.T) *identityfederation.Issuer {
	t.Helper()

	issuer, err := identityfederation.NewIssuer(
		baseurl.MustParse(issuerBase),
		testKeyRing(t, rsaTestKey()),
		identityfederation.DefaultTokenTTL,
	)
	require.NoError(t, err)

	return issuer
}

func organizationGID(t *testing.T, raw string) gid.GID {
	t.Helper()

	organizationID, err := gid.ParseGID(raw)
	require.NoError(t, err)
	require.Equal(t, coredata.OrganizationEntityType, organizationID.EntityType())

	return organizationID
}

// policyFor builds the trust policy a Probo stack installs for one organization:
// provider and both conditions pin that organization.
func policyFor(t *testing.T, issuer *identityfederation.Issuer, organizationID gid.GID) trustPolicy {
	t.Helper()

	issuerURL, err := issuer.IssuerURL(organizationID)
	require.NoError(t, err)

	return trustPolicy{
		Issuer:    issuerURL,
		Audience:  identityfederation.AudienceAWS,
		Subject:   organizationID.String(),
		AccountID: accountID,
	}
}

func TestNewSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	issuer := testIssuer(t)
	organizationA := organizationGID(t, organizationAID)

	fake := newFakeSTS(
		t,
		issuer.JWKS(),
		map[string]trustPolicy{roleARN: policyFor(t, issuer, organizationA)},
	)

	session, err := cloudaws.NewSession(
		ctx,
		issuer,
		organizationA,
		roleARN,
		cloudaws.WithSTSEndpoint(fake.endpoint),
	)
	require.NoError(t, err)

	assert.Equal(t, cloud.AWS, session.Cloud())
	assert.Equal(t, accountID, session.AccountID())
	assert.Equal(t, 1, fake.exchangeCount())

	// The credentials a driver would sign with are the federated ones, served
	// from the SDK's cache rather than exchanged again.
	credentials, err := session.Config().Credentials.Retrieve(ctx)
	require.NoError(t, err)
	assert.Equal(t, "ASIAFAKEACCESSKEYID", credentials.AccessKeyID)
	assert.Equal(t, 1, fake.exchangeCount(), "cached credentials must not trigger a second exchange")

	// The token STS received was minted in this process for this organization. No
	// file is written or read on this path — the SDK's WebIdentityTokenFile
	// mechanism is never involved — so nothing else on the host can pick it up.
	claims := map[string]any{}
	payload, err := jose.VerifyJWTWithJWKS(fake.receivedToken(), issuer.JWKS())
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(payload, &claims))

	issuerURL, err := issuer.IssuerURL(organizationA)
	require.NoError(t, err)

	assert.Equal(t, issuerURL, claims["iss"])
	assert.Equal(t, organizationA.String(), claims["sub"])
	assert.Equal(t, identityfederation.AudienceAWS, claims["aud"])
}

// The test the whole design rests on. Organization B is a tenant of the same
// deployment, so its token is signed by the same key and verifies perfectly —
// only the claims stand between it and organization A's AWS account.
func TestNewSession_CrossTenantRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	issuer := testIssuer(t)
	organizationA := organizationGID(t, organizationAID)
	organizationB := organizationGID(t, organizationBID)

	// Isolation is structural: B's token carries a different iss, so the account
	// has no provider matching it and refuses before evaluating any condition. A
	// customer who mangles the conditions in their own IaC still cannot be
	// reached by another tenant.
	t.Run("issuer does not match the registered provider", func(t *testing.T) {
		t.Parallel()

		fake := newFakeSTS(
			t,
			issuer.JWKS(),
			map[string]trustPolicy{roleARN: policyFor(t, issuer, organizationA)},
		)

		session, err := cloudaws.NewSession(
			ctx,
			issuer,
			organizationB,
			roleARN,
			cloudaws.WithSTSEndpoint(fake.endpoint),
		)
		require.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "AccessDenied")
		assert.Contains(t, err.Error(), "no OIDC provider registered for issuer")
		assert.Equal(t, 1, fake.exchangeCount(), "the exchange must be attempted and refused, not skipped")
	})

	// The subject condition is the backstop behind that: a customer who registered
	// B's issuer as a second provider — by mistake, or because they onboarded two
	// Probo organizations — still cannot use it against a role pinning A.
	t.Run("subject does not match the pinned organization", func(t *testing.T) {
		t.Parallel()

		policy := policyFor(t, issuer, organizationA)

		issuerB, err := issuer.IssuerURL(organizationB)
		require.NoError(t, err)

		policy.Issuer = issuerB

		fake := newFakeSTS(t, issuer.JWKS(), map[string]trustPolicy{roleARN: policy})

		session, err := cloudaws.NewSession(
			ctx,
			issuer,
			organizationB,
			roleARN,
			cloudaws.WithSTSEndpoint(fake.endpoint),
		)
		require.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "AccessDenied")
		assert.Contains(t, err.Error(), "sub condition not satisfied")
	})
}

// A subject condition wildcarded while editing IaC still admits nobody: the
// template pins it with StringEquals, where "*" is a literal character.
func TestNewSession_WildcardedSubjectRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	issuer := testIssuer(t)
	organizationA := organizationGID(t, organizationAID)

	policy := policyFor(t, issuer, organizationA)
	policy.Subject = "*"

	fake := newFakeSTS(t, issuer.JWKS(), map[string]trustPolicy{roleARN: policy})

	session, err := cloudaws.NewSession(
		ctx,
		issuer,
		organizationA,
		roleARN,
		cloudaws.WithSTSEndpoint(fake.endpoint),
	)
	require.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "AccessDenied")
}

// The key/kid/JWKS wiring bug class, which reaches production as an opaque
// AccessDenied: a key absent from the published set obtains no credentials, and
// the failure names the token rather than the trust policy.
func TestNewSession_UnpublishedSigningKeyRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	issuer := testIssuer(t)
	organizationA := organizationGID(t, organizationAID)

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	strangerIssuer, err := identityfederation.NewIssuer(
		baseurl.MustParse(issuerBase),
		testKeyRing(t, otherKey),
		identityfederation.DefaultTokenTTL,
	)
	require.NoError(t, err)

	// The endpoint publishes the real issuer's key set; the session signs with a
	// key absent from it.
	fake := newFakeSTS(
		t,
		issuer.JWKS(),
		map[string]trustPolicy{roleARN: policyFor(t, issuer, organizationA)},
	)

	session, err := cloudaws.NewSession(
		ctx,
		strangerIssuer,
		organizationA,
		roleARN,
		cloudaws.WithSTSEndpoint(fake.endpoint),
	)
	require.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "InvalidIdentityToken")
}

func TestNewSession_UnknownRoleRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	issuer := testIssuer(t)
	organizationA := organizationGID(t, organizationAID)

	fake := newFakeSTS(
		t,
		issuer.JWKS(),
		map[string]trustPolicy{roleARN: policyFor(t, issuer, organizationA)},
	)

	session, err := cloudaws.NewSession(
		ctx,
		issuer,
		organizationA,
		"arn:aws:iam::123456789012:role/Deleted",
		cloudaws.WithSTSEndpoint(fake.endpoint),
	)
	require.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "AccessDenied")
}

func TestNewSession_Errors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	organizationA := organizationGID(t, organizationAID)

	t.Run("missing issuer", func(t *testing.T) {
		t.Parallel()

		session, err := cloudaws.NewSession(ctx, nil, organizationA, roleARN)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "identity federation issuer is required")
		assert.Nil(t, session)
	})

	t.Run("missing role ARN", func(t *testing.T) {
		t.Parallel()

		session, err := cloudaws.NewSession(ctx, testIssuer(t), organizationA, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "role ARN is required")
		assert.Nil(t, session)
	})

	// A GID naming anything but an organization must not reach STS at all.
	t.Run("identifier is not an organization", func(t *testing.T) {
		t.Parallel()

		issuer := testIssuer(t)
		organizationA := organizationGID(t, organizationAID)

		fake := newFakeSTS(
			t,
			issuer.JWKS(),
			map[string]trustPolicy{roleARN: policyFor(t, issuer, organizationA)},
		)

		session, err := cloudaws.NewSession(
			ctx,
			issuer,
			gid.New(organizationA.TenantID(), coredata.ConnectorEntityType),
			roleARN,
			cloudaws.WithSTSEndpoint(fake.endpoint),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "identifier is not an organization")
		assert.Nil(t, session)
		assert.Zero(t, fake.exchangeCount())
	})
}

// The region must reach the config drivers are built from, so a driver signs for
// the region the connector named.
func TestNewSession_Region(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	issuer := testIssuer(t)
	organizationA := organizationGID(t, organizationAID)

	fake := newFakeSTS(
		t,
		issuer.JWKS(),
		map[string]trustPolicy{roleARN: policyFor(t, issuer, organizationA)},
	)

	session, err := cloudaws.NewSession(
		ctx,
		issuer,
		organizationA,
		roleARN,
		cloudaws.WithSTSEndpoint(fake.endpoint),
		cloudaws.WithRegion("eu-west-3"),
	)
	require.NoError(t, err)
	assert.Equal(t, "eu-west-3", session.Config().Region)

	// An empty region keeps the default rather than producing a config the SDK
	// would reject at call time.
	defaulted, err := cloudaws.NewSession(
		ctx,
		issuer,
		organizationA,
		roleARN,
		cloudaws.WithSTSEndpoint(fake.endpoint),
		cloudaws.WithRegion(""),
	)
	require.NoError(t, err)
	assert.Equal(t, cloudaws.DefaultRegion, defaulted.Config().Region)
}

// The STS override must not follow a driver's clients: keeping the audited APIs
// pointed at AWS is what stops a test from faking the data it asserts on.
func TestNewSession_STSEndpointDoesNotLeakIntoConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	issuer := testIssuer(t)
	organizationA := organizationGID(t, organizationAID)

	fake := newFakeSTS(
		t,
		issuer.JWKS(),
		map[string]trustPolicy{roleARN: policyFor(t, issuer, organizationA)},
	)

	session, err := cloudaws.NewSession(
		ctx,
		issuer,
		organizationA,
		roleARN,
		cloudaws.WithSTSEndpoint(fake.endpoint),
	)
	require.NoError(t, err)

	assert.Nil(t, session.Config().BaseEndpoint)

	// Check the fake was reached, so this is not passing on nothing configured.
	endpoint, err := url.Parse(fake.endpoint)
	require.NoError(t, err)
	assert.NotEmpty(t, endpoint.Host)
	assert.Equal(t, 1, fake.exchangeCount())
}
