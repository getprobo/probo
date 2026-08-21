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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/crypto/jose"
)

// This fake stands in for the half of the flow Probo does not own, and does so
// faithfully rather than agreeably: it fetches the published JWKS over HTTP,
// verifies the RS256 signature with it, and evaluates the trust policy against
// the token's own claims. A mock returning canned credentials would agree with
// whatever claim shape we happened to emit — the mistake that reaches production
// as an opaque AccessDenied. LocalStack is no alternative: it verifies no JWT
// signature and evaluates no trust policy.

const (
	// AWS distinguishes a refused condition from a token that does not verify,
	// and so does this fake: they mean different things when debugging.
	stsErrorAccessDenied         = "AccessDenied"
	stsErrorInvalidIdentityToken = "InvalidIdentityToken"

	fakeAccessKeyID     = "ASIAFAKEACCESSKEYID"
	fakeSecretAccessKey = "fake-secret-access-key"
	fakeSessionToken    = "fake-session-token"
)

type (
	// trustPolicy is the part of an IAM role trust policy this fake evaluates:
	// the registered Principal plus the two StringEquals conditions a Probo
	// stack pins. A "*" in Audience or Subject is a literal, not a wildcard.
	trustPolicy struct {
		// Issuer is the registered OIDC provider URL. Real STS resolves the
		// provider by the token's iss and finds nothing when it does not match,
		// which is what makes one issuer per organization structural isolation
		// rather than a condition to get right.
		Issuer    string
		Audience  string
		Subject   string
		AccountID string
	}

	// fakeSTS answers AssumeRoleWithWebIdentity for the roles it was given. A
	// role absent from policies models one that does not exist.
	fakeSTS struct {
		t        *testing.T
		jwksURL  string
		endpoint string
		policies map[string]trustPolicy

		mu        sync.Mutex
		exchanges int
		lastToken string
	}
)

// newFakeSTS starts an STS endpoint answering for the given roles, plus a second
// endpoint publishing jwks as the issuer's key set.
func newFakeSTS(t *testing.T, jwks *jose.JWKS, policies map[string]trustPolicy) *fakeSTS {
	t.Helper()

	document, err := json.Marshal(jwks)
	require.NoError(t, err)

	jwksServer := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Cache-Control", "public, max-age=3600")
				_, _ = w.Write(document)
			},
		),
	)
	t.Cleanup(jwksServer.Close)

	fake := &fakeSTS{
		t:        t,
		jwksURL:  jwksServer.URL,
		policies: policies,
	}

	stsServer := httptest.NewServer(fake)
	t.Cleanup(stsServer.Close)

	fake.endpoint = stsServer.URL

	return fake
}

func (f *fakeSTS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "cannot parse form", http.StatusBadRequest)

		return
	}

	if action := r.PostForm.Get("Action"); action != "AssumeRoleWithWebIdentity" {
		// The credential exchange is the only call reachable from a session.
		f.t.Errorf("unexpected STS action %q", action)
		http.Error(w, "unsupported action", http.StatusBadRequest)

		return
	}

	token := r.PostForm.Get("WebIdentityToken")
	roleARN := r.PostForm.Get("RoleArn")
	sessionName := r.PostForm.Get("RoleSessionName")

	f.mu.Lock()
	f.exchanges++
	f.lastToken = token
	f.mu.Unlock()

	claims, err := f.verifyToken(token)
	if err != nil {
		writeSTSError(w, stsErrorInvalidIdentityToken, err.Error())

		return
	}

	policy, ok := f.policies[roleARN]
	if !ok {
		writeSTSError(w, stsErrorAccessDenied, fmt.Sprintf("no role %q", roleARN))

		return
	}

	if err := policy.evaluate(claims); err != nil {
		writeSTSError(w, stsErrorAccessDenied, err.Error())

		return
	}

	writeAssumeRoleWithWebIdentityResponse(w, policy, roleARN, sessionName, claims)
}

// verifyToken fetches the published key set over HTTP and verifies against it,
// the way STS does before it looks at any condition.
func (f *fakeSTS) verifyToken(token string) (map[string]any, error) {
	resp, err := http.Get(f.jwksURL)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch jwks: %w", err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	jwks := &jose.JWKS{}
	if err := json.NewDecoder(resp.Body).Decode(jwks); err != nil {
		return nil, fmt.Errorf("cannot decode jwks: %w", err)
	}

	payload, err := jose.VerifyJWTWithJWKS(token, jwks)
	if err != nil {
		return nil, fmt.Errorf("cannot verify token signature: %w", err)
	}

	claims := map[string]any{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("cannot decode token claims: %w", err)
	}

	expiresAt, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("token has no exp claim")
	}

	if time.Now().After(time.Unix(int64(expiresAt), 0)) {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

// evaluate applies the trust policy to a verified token's claims. Every
// comparison is an exact string match, mirroring StringEquals.
func (p trustPolicy) evaluate(claims map[string]any) error {
	issuer, _ := claims["iss"].(string)
	if issuer != p.Issuer {
		// Real STS fails harder: it finds no provider registered for the token's
		// issuer and reaches no condition at all.
		return fmt.Errorf("no OIDC provider registered for issuer %q", issuer)
	}

	if audience, _ := claims["aud"].(string); audience != p.Audience {
		return fmt.Errorf("aud condition not satisfied")
	}

	if subject, _ := claims["sub"].(string); subject != p.Subject {
		return fmt.Errorf("sub condition not satisfied")
	}

	return nil
}

func (f *fakeSTS) exchangeCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.exchanges
}

func (f *fakeSTS) receivedToken() string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.lastToken
}

func writeSTSError(w http.ResponseWriter, code string, message string) {
	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(http.StatusForbidden)

	_, _ = fmt.Fprintf(
		w,
		`<ErrorResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <Error>
    <Type>Sender</Type>
    <Code>%s</Code>
    <Message>%s</Message>
  </Error>
  <RequestId>fake-request-id</RequestId>
</ErrorResponse>`,
		code,
		escapeXML(message),
	)
}

func writeAssumeRoleWithWebIdentityResponse(
	w http.ResponseWriter,
	policy trustPolicy,
	roleARN string,
	sessionName string,
	claims map[string]any,
) {
	subject, _ := claims["sub"].(string)
	audience, _ := claims["aud"].(string)
	issuer, _ := claims["iss"].(string)

	roleName := roleARN
	if _, after, found := strings.Cut(roleARN, ":role/"); found {
		roleName = after
	}

	w.Header().Set("Content-Type", "text/xml")

	_, _ = fmt.Fprintf(
		w,
		`<AssumeRoleWithWebIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <AssumeRoleWithWebIdentityResult>
    <Credentials>
      <AccessKeyId>%s</AccessKeyId>
      <SecretAccessKey>%s</SecretAccessKey>
      <SessionToken>%s</SessionToken>
      <Expiration>%s</Expiration>
    </Credentials>
    <AssumedRoleUser>
      <Arn>arn:aws:sts::%s:assumed-role/%s/%s</Arn>
      <AssumedRoleId>AROAFAKEROLEID:%s</AssumedRoleId>
    </AssumedRoleUser>
    <SubjectFromWebIdentityToken>%s</SubjectFromWebIdentityToken>
    <Audience>%s</Audience>
    <Provider>%s</Provider>
  </AssumeRoleWithWebIdentityResult>
  <ResponseMetadata>
    <RequestId>fake-request-id</RequestId>
  </ResponseMetadata>
</AssumeRoleWithWebIdentityResponse>`,
		fakeAccessKeyID,
		fakeSecretAccessKey,
		fakeSessionToken,
		time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
		policy.AccountID,
		escapeXML(roleName),
		escapeXML(sessionName),
		escapeXML(sessionName),
		escapeXML(subject),
		escapeXML(audience),
		escapeXML(strings.TrimPrefix(issuer, "https://")),
	)
}

func escapeXML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
	)

	return replacer.Replace(s)
}
