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

package gcp

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

const (
	exchangeProviderResource = "projects/123456789012/locations/global/workloadIdentityPools/probo/providers/probo"
	exchangeServiceAccount   = "probo-audit@my-project.iam.gserviceaccount.com"
	exchangeIAMPath          = "/v1/projects/-/serviceAccounts/" + exchangeServiceAccount + ":generateAccessToken"
	exchangeFederatedToken   = "federated-access-token"
	exchangeSAToken          = "sa-access-token"
)

func TestAudienceForms(t *testing.T) {
	t.Parallel()

	p := providerResource{
		projectNumber: "123456789012",
		poolID:        "probo",
		providerID:    "probo",
	}

	jwt, err := jwtAudience(p)
	require.NoError(t, err)
	assert.Equal(
		t,
		"https://iam.googleapis.com/projects/123456789012/locations/global/workloadIdentityPools/probo/providers/probo",
		jwt,
	)

	sts, err := stsAudience(p)
	require.NoError(t, err)
	assert.Equal(
		t,
		"//iam.googleapis.com/projects/123456789012/locations/global/workloadIdentityPools/probo/providers/probo",
		sts,
	)
}

func TestCheckAccess_ExchangesAndImpersonates(t *testing.T) {
	t.Parallel()

	probe := newExchangeTestSession(t)

	err := probe.session.CheckAccess(context.Background())
	require.NoError(t, err)

	assert.Equal(
		t,
		"https://iam.googleapis.com/projects/123456789012/locations/global/workloadIdentityPools/probo/providers/probo",
		probe.jwtAudience,
	)
	assert.Equal(
		t,
		"//iam.googleapis.com/projects/123456789012/locations/global/workloadIdentityPools/probo/providers/probo",
		probe.stsAudience,
	)
	assert.Equal(t, "Bearer "+exchangeFederatedToken, probe.iamAuth)
}

func TestCheckAccess_CachesTokenForFirstAPICall(t *testing.T) {
	t.Parallel()

	probe := newExchangeTestSession(t)

	err := probe.session.CheckAccess(context.Background())
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		probe.serverURL+"/probe",
		nil,
	)
	require.NoError(t, err)

	resp, err := probe.session.HTTPClient().Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "Bearer "+exchangeSAToken, probe.apiAuth)
	assert.Equal(t, 1, probe.stsHits)
	assert.Equal(t, 1, probe.iamHits)
}

func TestHTTPClient_HonorsRequestContext(t *testing.T) {
	t.Parallel()

	probe := newExchangeTestSession(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, probe.serverURL+"/probe", nil)
	require.NoError(t, err)

	_, err = probe.session.HTTPClient().Do(req)
	require.Error(t, err)
}

type exchangeProbe struct {
	session     *Session
	serverURL   string
	stsHits     int
	iamHits     int
	stsAudience string
	jwtAudience string
	iamAuth     string
	apiAuth     string
}

func newExchangeTestSession(t *testing.T) *exchangeProbe {
	t.Helper()

	probe := &exchangeProbe{}

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/v1/token":
				probe.stsHits++

				var req struct {
					Audience     string `json:"audience"`
					SubjectToken string `json:"subjectToken"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, "cannot decode sts request", http.StatusBadRequest)
					return
				}

				probe.stsAudience = req.Audience

				audience, err := jwtAudienceFromToken(req.SubjectToken)
				if err != nil {
					http.Error(w, "cannot decode subject token", http.StatusBadRequest)
					return
				}

				probe.jwtAudience = audience

				_ = json.NewEncoder(w).Encode(
					map[string]any{
						"access_token":      exchangeFederatedToken,
						"expires_in":        3600,
						"token_type":        "Bearer",
						"issued_token_type": accessTokenType,
					},
				)
			case exchangeIAMPath:
				var req struct {
					Scope    []string `json:"scope"`
					Lifetime string   `json:"lifetime"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, "cannot decode iam request", http.StatusBadRequest)
					return
				}

				if !slices.Equal(
					req.Scope,
					[]string{cloudPlatformScope, adminDirectoryUserReadonlyScope},
				) || req.Lifetime != serviceAccountLifetime {
					http.Error(w, "unexpected iam request", http.StatusBadRequest)
					return
				}

				probe.iamHits++
				probe.iamAuth = r.Header.Get("Authorization")

				_ = json.NewEncoder(w).Encode(
					map[string]any{
						"accessToken": exchangeSAToken,
						"expireTime":  time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
					},
				)
			case "/probe":
				probe.apiAuth = r.Header.Get("Authorization")

				w.WriteHeader(http.StatusNoContent)
			default:
				http.NotFound(w, r)
			}
		}),
	)
	t.Cleanup(server.Close)

	session, err := NewSession(
		exchangeTestIssuer(t),
		gid.New(gid.NewTenantID(), coredata.OrganizationEntityType),
		exchangeProviderResource,
		exchangeServiceAccount,
	)
	require.NoError(t, err)

	session.httpClient = httpclient.DefaultPooledClient(
		httpclient.WithSSRFProtection(),
		httpclient.WithSSRFAllowLoopback(),
	)
	session.stsEndpoint = server.URL + "/"
	session.iamCredentialsEndpoint = server.URL + "/"
	session.authorizedClient = authorizeSession(session.httpClient, session)
	probe.session = session
	probe.serverURL = server.URL

	return probe
}

func jwtAudienceFromToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return "", errors.New("subject token is not a jwt")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	var claims identityfederation.Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", err
	}

	return claims.Audience, nil
}

func exchangeTestIssuer(t *testing.T) *identityfederation.Issuer {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyRing, err := jose.NewKeyRing(
		[]jose.SigningKey{{PrivateKey: key, KID: "test", Active: true}},
	)
	require.NoError(t, err)

	base, err := baseurl.Parse("https://proboidentity.com")
	require.NoError(t, err)

	issuer, err := identityfederation.NewIssuer(base, keyRing, identityfederation.DefaultTokenTTL)
	require.NoError(t, err)

	return issuer
}
