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
	"crypto/rand"
	"crypto/rsa"
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
	testIssuerBase = "https://proboidentity.com"
	testRoleARN    = "arn:aws:iam::123456789012:role/ProboAudit"
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

func TestNewSession(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		roleARN string
		region  string
	}{
		{
			name:    "commercial",
			roleARN: testRoleARN,
			region:  cloudaws.DefaultCommercialRegion,
		},
		{
			name:    "govcloud",
			roleARN: "arn:aws-us-gov:iam::123456789012:role/ProboAudit",
			region:  cloudaws.DefaultGovRegion,
		},
		{
			name:    "china",
			roleARN: "arn:aws-cn:iam::123456789012:role/ProboAudit",
			region:  cloudaws.DefaultChinaRegion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			session, err := cloudaws.NewSession(testIssuer(t), testOrganizationID(), tt.roleARN)
			require.NoError(t, err)

			assert.Equal(t, cloud.AWS, session.Cloud())
			assert.Equal(t, "123456789012", session.AccountID(), "account comes from the role ARN")
			assert.Equal(t, tt.region, session.Config().Region)
		})
	}
}

func TestNewSession_Validation(t *testing.T) {
	t.Parallel()

	issuer := testIssuer(t)
	organizationID := testOrganizationID()

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

			_, err := cloudaws.NewSession(issuer, organizationID, tt.roleARN)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantMessage)
		})
	}
}
