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

package azure_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

const (
	testIssuerBase          = "https://proboidentity.com"
	testTenantID            = "a1111111-1111-4111-8111-111111111111"
	testClientID            = "b2222222-2222-4222-8222-222222222222"
	testSubscriptionID      = "c3333333-3333-4333-8333-333333333333"
	testTenantIDUpper       = "A1111111-1111-4111-8111-111111111111"
	testClientIDUpper       = "B2222222-2222-4222-8222-222222222222"
	testSubscriptionIDUpper = "C3333333-3333-4333-8333-333333333333"
	nilGUID                 = "00000000-0000-0000-0000-000000000000"
	setupOrganizationID     = "e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhg"
)

func testIssuer(t *testing.T) *identityfederation.Issuer {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyRing, err := jose.NewKeyRing(
		[]jose.SigningKey{{PrivateKey: key, KID: "test", Active: true}},
	)
	require.NoError(t, err)

	base, err := baseurl.Parse(testIssuerBase)
	require.NoError(t, err)

	issuer, err := identityfederation.NewIssuer(base, keyRing, identityfederation.DefaultTokenTTL)
	require.NoError(t, err)

	return issuer
}

func testOrganizationID() gid.GID {
	return gid.New(gid.NewTenantID(), coredata.OrganizationEntityType)
}
