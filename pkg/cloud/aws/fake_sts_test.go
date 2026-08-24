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

// This file implements a deliberately faithful fake of AWS STS
// AssumeRoleWithWebIdentity: it really fetches the discovery document named by
// the token's iss, really fetches the jwks_uri from it, really verifies the
// RS256 signature, and really evaluates the aud and sub conditions of a trust
// policy fixture.
//
// A mock that simply agreed with us would prove nothing, because the whole
// design rests on our claim shape being accepted by a real reading of
// trust-policy semantics. LocalStack does not help either: it verifies no JWT
// signature and evaluates no trust policy.

package aws_test

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/crypto/jose"
)

const (
	fakeAccessKeyID     = "ASIAFAKEACCESSKEY"
	fakeSecretAccessKey = "fake-secret-access-key"
	fakeSessionToken    = "fake-session-token"
)

type (
	// trustPolicy is the customer-side half of the exchange: which role may be
	// assumed, and the StringEquals conditions its trust statement pins. Keys
	// are the AWS condition-key form, "<issuer without scheme>:aud" and
	// "<issuer without scheme>:sub".
	trustPolicy struct {
		roleARN    string
		conditions map[string]string
	}

	// fakeSTS answers AssumeRoleWithWebIdentity for the roles it was given a
	// trust policy for, and refuses everything else the way STS does.
	fakeSTS struct {
		t        *testing.T
		policies []trustPolicy
		// jwksClient fetches the issuer's public documents. Loopback is
		// allowed because the issuer under test is an httptest server.
		jwksClient *http.Client

		mu    sync.Mutex
		calls []fakeSTSCall
	}

	// fakeSTSCall records what the exchange saw, so a test can assert on the
	// claims the SDK actually presented rather than on the ones it meant to.
	fakeSTSCall struct {
		roleARN         string
		roleSessionName string
		issuer          string
		subject         string
		audience        string
	}

	stsErrorResponse struct {
		XMLName xml.Name `xml:"ErrorResponse"`
		Type    string   `xml:"Error>Type"`
		Code    string   `xml:"Error>Code"`
		Message string   `xml:"Error>Message"`
	}

	stsCredentials struct {
		AccessKeyID     string    `xml:"AccessKeyId"`
		SecretAccessKey string    `xml:"SecretAccessKey"`
		SessionToken    string    `xml:"SessionToken"`
		Expiration      time.Time `xml:"Expiration"`
	}

	stsAssumeRoleResponse struct {
		XMLName     xml.Name       `xml:"AssumeRoleWithWebIdentityResponse"`
		Credentials stsCredentials `xml:"AssumeRoleWithWebIdentityResult>Credentials"`
		Subject     string         `xml:"AssumeRoleWithWebIdentityResult>SubjectFromWebIdentityToken"`
		Audience    string         `xml:"AssumeRoleWithWebIdentityResult>Audience"`
	}
)

// newFakeSTS starts a fake STS honouring the given trust policies and returns
// it with its endpoint URL.
func newFakeSTS(t *testing.T, policies ...trustPolicy) (*fakeSTS, string) {
	t.Helper()

	sts := &fakeSTS{
		t:          t,
		policies:   policies,
		jwksClient: loopbackClient(),
	}

	srv := httptest.NewServer(sts)
	t.Cleanup(srv.Close)

	return sts, srv.URL
}

// callsMade returns a copy of the exchanges the fake saw.
func (f *fakeSTS) callsMade() []fakeSTSCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]fakeSTSCall(nil), f.calls...)
}

func (f *fakeSTS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeSTSError(w, http.StatusBadRequest, "MalformedInput", err.Error())

		return
	}

	if action := r.PostForm.Get("Action"); action != "AssumeRoleWithWebIdentity" {
		writeSTSError(w, http.StatusBadRequest, "InvalidAction", "unexpected action "+action)

		return
	}

	call, claims, err := f.verifyToken(r.PostForm.Get("WebIdentityToken"))
	if err != nil {
		// STS reports an unverifiable token as InvalidIdentityToken, distinct
		// from a token that verifies but fails the trust policy.
		writeSTSError(w, http.StatusBadRequest, "InvalidIdentityToken", err.Error())

		return
	}

	call.roleARN = r.PostForm.Get("RoleArn")
	call.roleSessionName = r.PostForm.Get("RoleSessionName")

	f.mu.Lock()
	f.calls = append(f.calls, *call)
	f.mu.Unlock()

	if err := f.evaluateTrustPolicy(*call, claims); err != nil {
		writeSTSError(w, http.StatusForbidden, "AccessDenied", err.Error())

		return
	}

	writeSTSXML(
		w,
		http.StatusOK,
		&stsAssumeRoleResponse{
			Credentials: stsCredentials{
				AccessKeyID:     fakeAccessKeyID,
				SecretAccessKey: fakeSecretAccessKey,
				SessionToken:    fakeSessionToken,
				Expiration:      time.Now().Add(time.Hour).UTC().Truncate(time.Second),
			},
			Subject:  call.subject,
			Audience: call.audience,
		},
	)
}

// verifyToken walks the same path STS does: read iss from the unverified
// payload, fetch the discovery document beneath it, fetch the jwks_uri it
// advertises, and only then verify the signature.
func (f *fakeSTS) verifyToken(token string) (*fakeSTSCall, map[string]any, error) {
	if token == "" {
		return nil, nil, fmt.Errorf("missing web identity token")
	}

	unverified, err := decodeUnverifiedClaims(token)
	if err != nil {
		return nil, nil, err
	}

	issuer, _ := unverified["iss"].(string)
	if issuer == "" {
		return nil, nil, fmt.Errorf("token carries no iss claim")
	}

	jwks, err := f.fetchJWKS(issuer)
	if err != nil {
		return nil, nil, err
	}

	payload, err := jose.VerifyJWTWithJWKS(token, jwks)
	if err != nil {
		return nil, nil, fmt.Errorf("signature verification failed: %w", err)
	}

	claims := map[string]any{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, nil, fmt.Errorf("cannot decode verified claims: %w", err)
	}

	subject, _ := claims["sub"].(string)
	audience, _ := claims["aud"].(string)

	return &fakeSTSCall{
		issuer:   issuer,
		subject:  subject,
		audience: audience,
	}, claims, nil
}

func (f *fakeSTS) fetchJWKS(issuer string) (*jose.JWKS, error) {
	var metadata struct {
		Issuer  string `json:"issuer"`
		JWKSURI string `json:"jwks_uri"`
	}

	if err := f.getJSON(issuer+"/.well-known/openid-configuration", &metadata); err != nil {
		return nil, fmt.Errorf("cannot fetch discovery document: %w", err)
	}

	// STS trusts the document only if it claims the issuer it was reached
	// through; a mismatch is how a copied document would be caught.
	if metadata.Issuer != issuer {
		return nil, fmt.Errorf(
			"discovery document issuer %q does not match %q",
			metadata.Issuer,
			issuer,
		)
	}

	jwks := &jose.JWKS{}
	if err := f.getJSON(metadata.JWKSURI, jwks); err != nil {
		return nil, fmt.Errorf("cannot fetch jwks: %w", err)
	}

	return jwks, nil
}

func (f *fakeSTS) getJSON(endpoint string, out any) error {
	resp, err := f.jwksClient.Get(endpoint)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: status %d", endpoint, resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

// evaluateTrustPolicy applies the StringEquals conditions of the statement
// matching the requested role, or denies when no statement matches — which is
// what makes a per-organization issuer structurally isolating.
func (f *fakeSTS) evaluateTrustPolicy(call fakeSTSCall, claims map[string]any) error {
	for _, policy := range f.policies {
		if policy.roleARN != call.roleARN {
			continue
		}

		for key, want := range policy.conditions {
			claim, ok := conditionKeyClaim(key, call.issuer)
			if !ok {
				return fmt.Errorf("condition key %q does not name this issuer", key)
			}

			got, _ := claims[claim].(string)
			if got != want {
				return fmt.Errorf("condition %q requires %q", key, want)
			}
		}

		return nil
	}

	return fmt.Errorf("not authorized to perform sts:AssumeRoleWithWebIdentity on %s", call.roleARN)
}

// conditionKeyClaim splits an AWS condition key of the form
// "<issuer without scheme>:<claim>" and reports whether its issuer half names
// the issuer that signed the token.
// Split on the LAST colon, not the first: AWS forbids a port in an issuer URL
// so the two readings agree in production, but the loopback issuer these tests
// run against carries one.
func conditionKeyClaim(key, issuer string) (string, bool) {
	separator := strings.LastIndex(key, ":")
	if separator < 0 {
		return "", false
	}

	// A condition naming another issuer must never match: that is the
	// cross-tenant boundary the whole design rests on.
	return key[separator+1:], key[:separator] == issuerWithoutScheme(issuer)
}

// issuerWithoutScheme is the form AWS uses in a condition key and in an OIDC
// provider ARN.
func issuerWithoutScheme(issuer string) string {
	trimmed := strings.TrimPrefix(issuer, "https://")

	return strings.TrimPrefix(trimmed, "http://")
}

// conditionKeysFor builds the trust statement a customer's CloudFormation
// stack writes for one organization.
func conditionKeysFor(issuer, subject, audience string) map[string]string {
	host := issuerWithoutScheme(issuer)

	return map[string]string{
		host + ":aud": audience,
		host + ":sub": subject,
	}
}

func decodeUnverifiedClaims(token string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("token is not a three-part JWT")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("cannot decode token payload: %w", err)
	}

	claims := map[string]any{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("cannot unmarshal token payload: %w", err)
	}

	return claims, nil
}

func writeSTSError(w http.ResponseWriter, status int, code, message string) {
	writeSTSXML(
		w,
		status,
		&stsErrorResponse{
			Type:    "Sender",
			Code:    code,
			Message: message,
		},
	)
}

func writeSTSXML(w http.ResponseWriter, status int, body any) {
	encoded, err := xml.Marshal(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(status)
	_, _ = w.Write(encoded)
}

// requireNoSTSCall fails the test when the exchange was attempted, for cases
// that must be refused before any network call.
func requireNoSTSCall(t *testing.T, sts *fakeSTS) {
	t.Helper()

	require.Empty(t, sts.callsMade(), "expected no STS exchange")
}
