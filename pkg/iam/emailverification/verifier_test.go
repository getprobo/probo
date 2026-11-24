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

package emailverification_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/dnsclient"
	"go.probo.inc/probo/pkg/iam/emailverification"
)

type (
	signer struct {
		alg  string
		sign func([]byte) []byte
		jwk  map[string]any
	}

	fixture struct {
		t              *testing.T
		now            time.Time
		issuer         signer
		browser        signer
		kid            string
		metadataIssuer string
		metadataAlgs   []string
		jwksURI        string
		txt            []string
	}

	tokenSpec struct {
		evtHeader      map[string]any
		evtClaims      map[string]any
		kbHeader       map[string]any
		kbClaims       map[string]any
		issuerSigner   signer
		kbSigner       signer
		sdHashOverride string
		rawToken       string
	}
)

const (
	testEmail    = "user@email-domain.example"
	testIssuer   = "https://issuer.example"
	testAudience = "https://console.example"
	testNonce    = "relying-party-nonce"
	testKid      = "2026-11-24"
)

var testNow = time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)

func TestVerify_AcceptsValidToken(t *testing.T) {
	t.Parallel()

	t.Run("ed25519 issuer", func(t *testing.T) {
		t.Parallel()

		f := newFixture(t, newEd25519Signer(t))
		spec := f.defaultSpec()

		require.NoError(t, f.verify(spec))
	})

	t.Run("es256 issuer", func(t *testing.T) {
		t.Parallel()

		f := newFixture(t, newES256Signer(t))
		f.metadataAlgs = []string{"ES256"}
		spec := f.defaultSpec()

		require.NoError(t, f.verify(spec))
	})
}

func TestVerify_RejectsInvalidToken(t *testing.T) {
	t.Parallel()

	otherKey := newEd25519Signer(t)

	cases := []struct {
		name   string
		setup  func(f *fixture)
		mutate func(s *tokenSpec)
	}{
		{
			name: "audience mismatch",
			mutate: func(s *tokenSpec) {
				s.kbClaims["aud"] = "https://evil.example"
			},
		},
		{
			name: "nonce not issued by the relying party",
			mutate: func(s *tokenSpec) {
				s.kbClaims["nonce"] = "forged-nonce"
			},
		},
		{
			name: "sd_hash does not bind the email verification token",
			mutate: func(s *tokenSpec) {
				s.sdHashOverride = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
			},
		},
		{
			name: "key binding signed by a key other than the confirmation key",
			mutate: func(s *tokenSpec) {
				s.kbSigner = otherKey
			},
		},
		{
			name: "email differs from the token email claim",
			mutate: func(s *tokenSpec) {
				s.evtClaims["email"] = "other@email-domain.example"
			},
		},
		{
			name: "email_verified is false",
			mutate: func(s *tokenSpec) {
				s.evtClaims["email_verified"] = false
			},
		},
		{
			name: "issuer does not match the DNS delegation",
			setup: func(f *fixture) {
				f.txt = []string{"iss=evil.example"}
			},
		},
		{
			name: "DNS publishes a host with a port",
			setup: func(f *fixture) {
				f.txt = []string{"iss=issuer.example:443"}
			},
		},
		{
			name: "DNS record does not start with iss=",
			setup: func(f *fixture) {
				f.txt = []string{"v=spf1 include:issuer.example"}
			},
		},
		{
			name: "more than one DNS record",
			setup: func(f *fixture) {
				f.txt = []string{"iss=issuer.example", "iss=evil.example"}
			},
		},
		{
			name: "no DNS record",
			setup: func(f *fixture) {
				f.txt = nil
			},
		},
		{
			name: "metadata issuer does not match the fetched issuer",
			setup: func(f *fixture) {
				f.metadataIssuer = "https://evil.example"
			},
		},
		{
			name: "jwks_uri is not https",
			setup: func(f *fixture) {
				f.jwksURI = "http://issuer.example/jwks"
			},
		},
		{
			name: "issuer metadata does not allow the EVT alg",
			setup: func(f *fixture) {
				f.metadataAlgs = []string{"ES256"}
			},
		},
		{
			name: "EVT signed by a key the issuer did not publish",
			mutate: func(s *tokenSpec) {
				s.issuerSigner = otherKey
			},
		},
		{
			name: "kid not published by the issuer",
			mutate: func(s *tokenSpec) {
				s.evtHeader["kid"] = "unknown-kid"
			},
		},
		{
			name: "alg none",
			mutate: func(s *tokenSpec) {
				s.evtHeader["alg"] = "none"
			},
		},
		{
			name: "polymorphic EdDSA alg",
			mutate: func(s *tokenSpec) {
				s.evtHeader["alg"] = "EdDSA"
			},
		},
		{
			name: "wrong EVT typ",
			mutate: func(s *tokenSpec) {
				s.evtHeader["typ"] = "JWT"
			},
		},
		{
			name: "wrong KB typ",
			mutate: func(s *tokenSpec) {
				s.kbHeader["typ"] = "JWT"
			},
		},
		{
			name: "KB alg differs from confirmation key alg",
			mutate: func(s *tokenSpec) {
				s.kbHeader["alg"] = "ES256"
			},
		},
		{
			name: "EVT older than the freshness window",
			mutate: func(s *tokenSpec) {
				s.evtClaims["iat"] = testNow.Add(-time.Hour).Unix()
			},
		},
		{
			name: "EVT issued in the future",
			mutate: func(s *tokenSpec) {
				s.evtClaims["iat"] = testNow.Add(time.Hour).Unix()
			},
		},
		{
			name: "token without key binding",
			mutate: func(s *tokenSpec) {
				s.rawToken = "only-one-part"
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t, newEd25519Signer(t))
			if tc.setup != nil {
				tc.setup(f)
			}

			spec := f.defaultSpec()
			if tc.mutate != nil {
				tc.mutate(&spec)
			}

			require.Error(t, f.verify(spec))
		})
	}
}

func TestVerify_RejectsInvalidRequest(t *testing.T) {
	t.Parallel()

	t.Run("email with display name", func(t *testing.T) {
		t.Parallel()

		f := newFixture(t, newEd25519Signer(t))
		spec := f.defaultSpec()

		err := f.verifyRequest(spec, "Name <"+testEmail+">", true)

		require.Error(t, err)
	})

	t.Run("missing nonce validator", func(t *testing.T) {
		t.Parallel()

		f := newFixture(t, newEd25519Signer(t))
		spec := f.defaultSpec()

		err := f.verifyRequest(spec, testEmail, false)

		require.Error(t, err)
	})
}

func newEd25519Signer(t *testing.T) signer {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	return signer{
		alg: "Ed25519",
		sign: func(input []byte) []byte {
			return ed25519.Sign(priv, input)
		},
		jwk: map[string]any{
			"kty": "OKP",
			"crv": "Ed25519",
			"x":   b64(pub),
		},
	}
}

func newES256Signer(t *testing.T) signer {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	ecdhKey, err := key.ECDH()
	require.NoError(t, err)

	raw := ecdhKey.PublicKey().Bytes()

	return signer{
		alg: "ES256",
		sign: func(input []byte) []byte {
			digest := sha256.Sum256(input)

			r, s, err := ecdsa.Sign(rand.Reader, key, digest[:])
			require.NoError(t, err)

			sig := make([]byte, 64)
			r.FillBytes(sig[:32])
			s.FillBytes(sig[32:])

			return sig
		},
		jwk: map[string]any{
			"kty": "EC",
			"crv": "P-256",
			"x":   b64(raw[1:33]),
			"y":   b64(raw[33:]),
		},
	}
}

func newFixture(t *testing.T, issuer signer) *fixture {
	return &fixture{
		t:              t,
		now:            testNow,
		issuer:         issuer,
		browser:        newEd25519Signer(t),
		kid:            testKid,
		metadataIssuer: testIssuer,
		jwksURI:        testIssuer + "/jwks",
		txt:            []string{"iss=issuer.example"},
	}
}

func (f *fixture) defaultSpec() tokenSpec {
	confirmationJWK := make(map[string]any, len(f.browser.jwk)+1)
	maps.Copy(confirmationJWK, f.browser.jwk)

	confirmationJWK["alg"] = f.browser.alg

	return tokenSpec{
		evtHeader: map[string]any{
			"alg": f.issuer.alg,
			"kid": f.kid,
			"typ": "evt+jwt",
		},
		evtClaims: map[string]any{
			"iss":            testIssuer,
			"iat":            f.now.Unix(),
			"cnf":            map[string]any{"jwk": confirmationJWK},
			"email":          testEmail,
			"email_verified": true,
		},
		kbHeader: map[string]any{
			"alg": f.browser.alg,
			"typ": "kb+jwt",
		},
		kbClaims: map[string]any{
			"aud":   testAudience,
			"nonce": testNonce,
			"iat":   f.now.Unix(),
		},
		issuerSigner: f.issuer,
		kbSigner:     f.browser,
	}
}

func (f *fixture) verify(spec tokenSpec) error {
	return f.verifyRequest(spec, testEmail, true)
}

func (f *fixture) verifyRequest(spec tokenSpec, email string, withNonceValidator bool) error {
	t := f.t

	token := spec.token(t)

	client, closeServer := f.serve()
	defer closeServer()

	txt := append([]string(nil), f.txt...)
	lookup := func(_ context.Context, name string) ([]string, error) {
		if len(txt) == 0 {
			return nil, dnsclient.ErrTXTNotFound
		}

		if name != "_email-verification.email-domain.example" {
			return nil, errors.New("unexpected lookup " + name)
		}

		return txt, nil
	}

	now := f.now
	opts := []emailverification.Option{
		emailverification.WithLookupTXT(lookup),
		emailverification.WithHTTPClient(client),
		emailverification.WithClock(func() time.Time { return now }),
	}

	v := emailverification.NewVerifier("", opts...)

	req := emailverification.Request{
		Token:    token,
		Email:    email,
		Audience: testAudience,
	}
	if withNonceValidator {
		req.ValidateNonce = func(nonce string) error {
			if nonce != testNonce {
				return errors.New("nonce not recognized")
			}

			return nil
		}
	}

	return v.Verify(context.Background(), req)
}

// serve starts an issuer over TLS that publishes the fixture's metadata and
// keys. The client routes every request to that server, so the issuer host
// name stays the one the DNS record and metadata refer to.
func (f *fixture) serve() (*http.Client, func()) {
	metadata := map[string]any{
		"issuer":   f.metadataIssuer,
		"jwks_uri": f.jwksURI,
	}
	if f.metadataAlgs != nil {
		metadata["signing_alg_values_supported"] = f.metadataAlgs
	}

	publicKey := f.issuer.jwk
	jwkEntry := make(map[string]any, len(publicKey)+2)
	maps.Copy(jwkEntry, publicKey)

	jwkEntry["kid"] = f.kid
	jwkEntry["alg"] = f.issuer.alg

	jwksDocument := map[string]any{"keys": []any{jwkEntry}}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/email-verification", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, metadata)
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, jwksDocument)
	})

	server := httptest.NewTLSServer(mux)

	serverHost := strings.TrimPrefix(server.URL, "https://")
	client := &http.Client{
		Transport: routeTo(server.Client().Transport, serverHost),
	}

	return client, server.Close
}

func (s tokenSpec) token(t *testing.T) string {
	if s.rawToken != "" {
		return s.rawToken
	}

	evt := signJWS(t, s.evtHeader, s.evtClaims, s.issuerSigner.sign)

	sdHash := s.sdHashOverride
	if sdHash == "" {
		digest := sha256.Sum256([]byte(evt + "~"))
		sdHash = b64(digest[:])
	}

	s.kbClaims["sd_hash"] = sdHash

	kb := signJWS(t, s.kbHeader, s.kbClaims, s.kbSigner.sign)

	return evt + "~" + kb
}

func signJWS(t *testing.T, header map[string]any, claims map[string]any, sign func([]byte) []byte) string {
	headerJSON, err := json.Marshal(header)
	require.NoError(t, err)

	claimsJSON, err := json.Marshal(claims)
	require.NoError(t, err)

	input := b64(headerJSON) + "." + b64(claimsJSON)

	return input + "." + b64(sign([]byte(input)))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

type routeTransport struct {
	base http.RoundTripper
	host string
}

func routeTo(base http.RoundTripper, host string) http.RoundTripper {
	return routeTransport{base: base, host: host}
}

func (rt routeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Host = rt.host

	return rt.base.RoundTrip(req)
}

func b64(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
