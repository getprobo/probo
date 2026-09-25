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

package identityfederation

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/gid"
	idfed "go.probo.inc/probo/pkg/identityfederation"
)

const (
	knownOrganizationID = "e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhg"

	issuerBase          = "https://proboidentity.com"
	discoveryPathSuffix = "/.well-known/openid-configuration"
)

var rsaTestKey = sync.OnceValue(
	func() *rsa.PrivateKey {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}

		return key
	},
)

func testIssuer(t testing.TB) *idfed.Issuer {
	t.Helper()

	keyRing, err := jose.NewKeyRing(
		[]jose.SigningKey{{PrivateKey: rsaTestKey(), KID: "test", Active: true}},
	)
	require.NoError(t, err)

	issuer, err := idfed.NewIssuer(
		baseurl.MustParse(issuerBase),
		keyRing,
		idfed.DefaultTokenTTL,
	)
	require.NoError(t, err)

	return issuer
}

func testMux(t testing.TB) http.Handler {
	t.Helper()

	return NewMux(
		log.NewLogger(log.WithOutput(io.Discard)),
		testIssuer(t),
		nil,
	)
}

func doGet(t testing.TB, handler http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, path, nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	return recorder
}

func discoveryPath(organizationID string) string {
	return "/" + organizationID + discoveryPathSuffix
}

func TestJWKS_IgnoresRequestHost(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/"+knownOrganizationID+"/jwks", nil)
	request.Host = "evil.example.com"
	request.Header.Set("X-Forwarded-Host", "evil.example.com")

	recorder := httptest.NewRecorder()
	testMux(t).ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	// Echoing the request Host would let any caller obtain a document claiming
	// to be an issuer they control.
	assert.NotContains(t, recorder.Body.String(), "evil.example.com")
}

func TestDiscovery_NotFound(t *testing.T) {
	t.Parallel()

	connectorID := gid.New(gid.TenantID{}, coredata.ConnectorEntityType)

	tests := []struct {
		name           string
		organizationID string
	}{
		{
			name:           "nil identifier",
			organizationID: gid.Nil.String(),
		},
		{
			name:           "identifier is not an organization",
			organizationID: connectorID.String(),
		},
		{
			name:           "mis-cased identifier",
			organizationID: strings.ToLower(knownOrganizationID),
		},
		{
			name:           "malformed identifier",
			organizationID: "not-a-gid",
		},
		{
			name:           "empty identifier",
			organizationID: "%20",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				t.Parallel()

				response := doGet(t, testMux(t), discoveryPath(test.organizationID))

				assert.Equal(t, http.StatusNotFound, response.Code)
				assert.Empty(t, response.Header().Get("Cache-Control"))
			},
		)
	}
}

func TestJWKS(t *testing.T) {
	t.Parallel()

	response := doGet(t, testMux(t), "/"+knownOrganizationID+"/jwks")

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "public, max-age=3600", response.Header().Get("Cache-Control"))

	jwks := jose.JWKS{}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &jwks))

	require.Len(t, jwks.Keys, 1)
	assert.Equal(t, "test", jwks.Keys[0].KeyID)
	assert.Equal(t, "RSA", jwks.Keys[0].KeyType)
	assert.Equal(t, "RS256", jwks.Keys[0].Algorithm)
}

func TestJWKS_UnknownOrganizationStillServesKeys(t *testing.T) {
	t.Parallel()

	unknownOrganizationID := gid.New(gid.TenantID{}, coredata.OrganizationEntityType)

	response := doGet(t, testMux(t), "/"+unknownOrganizationID.String()+"/jwks")

	assert.Equal(t, http.StatusOK, response.Code)
}

func TestJWKS_RejectsNonOrganizationIdentifier(t *testing.T) {
	t.Parallel()

	connectorID := gid.New(gid.TenantID{}, coredata.ConnectorEntityType)

	response := doGet(t, testMux(t), "/"+connectorID.String()+"/jwks")

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestJWKS_RejectsNilIdentifier(t *testing.T) {
	t.Parallel()

	response := doGet(t, testMux(t), "/"+gid.Nil.String()+"/jwks")

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestMux_NoWriteOrTokenSurface(t *testing.T) {
	t.Parallel()

	handler := testMux(t)

	t.Run(
		"no token endpoint",
		func(t *testing.T) {
			t.Parallel()

			for _, path := range []string{
				"/" + knownOrganizationID + "/token",
				"/token",
				"/" + knownOrganizationID,
			} {
				response := doGet(t, handler, path)
				assert.Equal(t, http.StatusNotFound, response.Code, path)
			}
		},
	)

	t.Run(
		"no RFC 8414 discovery variant",
		func(t *testing.T) {
			t.Parallel()

			response := doGet(
				t,
				handler,
				"/.well-known/oauth-authorization-server/"+knownOrganizationID,
			)

			assert.Equal(t, http.StatusNotFound, response.Code)
		},
	)

	t.Run(
		"documents reject writes",
		func(t *testing.T) {
			t.Parallel()

			for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
				request := httptest.NewRequest(method, discoveryPath(knownOrganizationID), nil)
				recorder := httptest.NewRecorder()

				handler.ServeHTTP(recorder, request)

				assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code, method)
			}
		},
	)
}
