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

package accessreview

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/identityfederation"
)

func testIssuer(t *testing.T) *identityfederation.Issuer {
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

// TestOpenSession covers the account argument, which is the whole point of
// the helper: an empty account is today's settings-implied one, a named
// account is that one instead. Every consumer routes through here, so this is
// where a forgotten account id would show.
func TestOpenSession(t *testing.T) {
	t.Parallel()

	awsConnector := func(t *testing.T) *coredata.Connector {
		t.Helper()

		c := &coredata.Connector{Provider: coredata.ConnectorProviderAWS}
		require.NoError(t, c.SetSettings(coredata.AWSConnectorSettings{
			RoleARN: "arn:aws:iam::111111111111:role/ProboAudit",
		}))

		return c
	}

	t.Run("an empty account opens the settings-implied one", func(t *testing.T) {
		t.Parallel()

		session, err := openSession(
			context.Background(), testIssuer(t), provider.NewBuiltinRegistry(), awsConnector(t), "",
		)
		require.NoError(t, err)

		assert.Equal(t, cloud.AWS, session.Cloud())
		assert.Equal(t, "111111111111", session.AccountID())
	})

	t.Run("a named account opens that one instead", func(t *testing.T) {
		t.Parallel()

		session, err := openSession(
			context.Background(), testIssuer(t), provider.NewBuiltinRegistry(), awsConnector(t), "222222222222",
		)
		require.NoError(t, err)

		assert.Equal(t, "222222222222", session.AccountID())
	})
}
