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

package certmanager_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/certmanager"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

type selectorFixture struct {
	client      *pg.Client
	scope       *coredata.Scope
	certificate *coredata.Certificate
	domain      *coredata.CustomDomain
	key         cipher.EncryptionKey
}

func setSelectorCertificate(t *testing.T, f *selectorFixture, expiresAt time.Time, serial int64) {
	t.Helper()

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(serial),
		DNSNames:     []string{f.domain.Domain},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     expiresAt,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	require.NoError(t, err)
	keyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	f.certificate.SSLCertificatePEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	f.certificate.SSLExpiresAt = new(expiresAt)
	f.certificate.Status = coredata.CertificateStatusActive
	require.NoError(
		t,
		f.certificate.EncryptPrivateKey(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), f.key),
	)
	require.NoError(t, f.certificate.ParseCertificate(f.key))
}

func newSelectorFixture(t *testing.T, expiresAt time.Time) *selectorFixture {
	t.Helper()

	client := test.PGClient(t)
	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	domain := coredata.NewCustomDomain(tenantID, organizationID, "cert-"+organizationID.String()+".example.com", false)
	certificate := coredata.NewCertificate(tenantID, domain.Domain)
	domain.CertificateID = new(certificate.ID)
	f := &selectorFixture{
		client:      client,
		scope:       coredata.NewScope(tenantID, gid.Nil),
		certificate: certificate,
		domain:      domain,
	}
	setSelectorCertificate(t, f, expiresAt, 1)

	t.Cleanup(
		func() {
			assert.NoError(
				t,
				client.WithTx(
					context.Background(),
					func(ctx context.Context, tx pg.Tx) error {
						if err := domain.Delete(ctx, tx, f.scope); err != nil {
							return err
						}

						if _, err := tx.Exec(ctx, "DELETE FROM organizations WHERE id = $1", organizationID); err != nil {
							return err
						}

						_, err := tx.Exec(ctx, "DELETE FROM certificates WHERE id = $1", certificate.ID)

						return err
					},
				),
			)
		},
	)
	require.NoError(
		t,
		client.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				now := time.Now()
				if _, err := tx.Exec(
					ctx,
					`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $4)`,
					organizationID,
					tenantID,
					"selector test",
					now,
				); err != nil {
					return err
				}

				if err := certificate.Insert(ctx, tx, f.scope); err != nil {
					return err
				}

				if err := domain.Insert(ctx, tx, f.scope); err != nil {
					return err
				}

				var cache coredata.CachedCertificate

				return cache.RefreshFromCertificate(ctx, tx, certificate, f.key)
			},
		),
	)

	return f
}

func TestSelectorGetCertificate_Renewal(t *testing.T) {
	t.Parallel()

	for _, status := range []coredata.CertificateStatus{
		coredata.CertificateStatusActive,
		coredata.CertificateStatusRenewing,
		coredata.CertificateStatusProvisioning,
	} {
		t.Run(
			strings.ToLower(string(status)),
			func(t *testing.T) {
				t.Parallel()

				f := newSelectorFixture(t, time.Now().Add(certmanager.RenewalWindow-24*time.Hour))
				hello := &tls.ClientHelloInfo{ServerName: f.domain.Domain}
				selectors := []*certmanager.Selector{
					certmanager.NewSelector(f.client, f.key),
					certmanager.NewSelector(f.client, f.key),
				}
				first, err := selectors[0].GetCertificate(hello)
				require.NoError(t, err)

				f.certificate.Status = status
				require.NoError(
					t,
					f.client.WithTx(
						t.Context(),
						func(ctx context.Context, tx pg.Tx) error {
							return f.certificate.Update(ctx, tx, f.scope)
						},
					),
				)

				for _, selector := range selectors {
					cert, err := selector.GetCertificate(hello)
					require.NoError(t, err)
					assert.Equal(t, first.Certificate, cert.Certificate, "serve the valid cache during renewal")
				}

				setSelectorCertificate(t, f, time.Now().Add(90*24*time.Hour), 2)
				require.NoError(
					t,
					f.client.WithTx(
						t.Context(),
						func(ctx context.Context, tx pg.Tx) error {
							if err := f.certificate.Update(ctx, tx, f.scope); err != nil {
								return err
							}

							var cache coredata.CachedCertificate

							return cache.RefreshFromCertificate(ctx, tx, f.certificate, f.key)
						},
					),
				)

				for _, selector := range selectors {
					cert, err := selector.GetCertificate(hello)
					require.NoError(t, err)
					assert.Equal(t, f.certificate.SSLCertificate.Certificate, cert.Certificate, "serve the renewed certificate")
				}
			},
		)
	}
}

func TestSelectorGetCertificate_ReusesCertificateOutsideRenewalWindow(t *testing.T) {
	t.Parallel()

	f := newSelectorFixture(t, time.Now().Add(certmanager.RenewalWindow+24*time.Hour))
	selector := certmanager.NewSelector(f.client, f.key)
	hello := &tls.ClientHelloInfo{ServerName: f.domain.Domain}
	first, err := selector.GetCertificate(hello)
	require.NoError(t, err)
	second, err := selector.GetCertificate(hello)
	require.NoError(t, err)
	assert.Same(t, first, second)
}

func TestSelectorGetCertificate_DomainNoLongerRoutable(t *testing.T) {
	t.Parallel()

	for _, offset := range []time.Duration{
		certmanager.RenewalWindow - 24*time.Hour,
		certmanager.RenewalWindow + 24*time.Hour,
	} {
		for _, deleted := range []bool{false, true} {
			t.Run(
				fmt.Sprintf("%s/deleted=%t", offset, deleted),
				func(t *testing.T) {
					t.Parallel()

					f := newSelectorFixture(t, time.Now().Add(offset))
					selector := certmanager.NewSelector(f.client, f.key)
					hello := &tls.ClientHelloInfo{ServerName: f.domain.Domain}
					_, err := selector.GetCertificate(hello)
					require.NoError(t, err)
					require.NoError(
						t,
						f.client.WithTx(
							t.Context(),
							func(ctx context.Context, tx pg.Tx) error {
								if deleted {
									return f.domain.Delete(ctx, tx, f.scope)
								}

								f.domain.CertificateID = nil

								return f.domain.Update(ctx, tx, f.scope)
							},
						),
					)

					cert, err := selector.GetCertificate(hello)
					assert.ErrorIs(t, err, coredata.ErrResourceNotFound)
					assert.Nil(t, cert)
				},
			)
		}
	}
}

func TestSelectorGetCertificate_RebuildsExpiredCache(t *testing.T) {
	t.Parallel()

	f := newSelectorFixture(t, time.Now().Add(24*time.Hour))
	selector := certmanager.NewSelector(f.client, f.key)
	hello := &tls.ClientHelloInfo{ServerName: f.domain.Domain}
	_, err := selector.GetCertificate(hello)
	require.NoError(t, err)

	setSelectorCertificate(t, f, time.Now().Add(90*24*time.Hour), 2)
	require.NoError(
		t,
		f.client.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				if err := f.certificate.Update(ctx, tx, f.scope); err != nil {
					return err
				}

				_, err := tx.Exec(ctx, "UPDATE cached_certificates SET expires_at = NOW() - INTERVAL '1 hour' WHERE domain = $1", f.domain.Domain)

				return err
			},
		),
	)

	cert, err := selector.GetCertificate(hello)
	require.NoError(t, err)
	assert.Equal(t, f.certificate.SSLCertificate.Certificate, cert.Certificate)
}

func TestSelectorGetCertificate_NoSNI(t *testing.T) {
	t.Parallel()

	selector := certmanager.NewSelector(nil, cipher.EncryptionKey{})
	cert, err := selector.GetCertificate(&tls.ClientHelloInfo{})
	assert.IsType(t, &certmanager.NoSNIError{}, err)
	assert.Nil(t, cert)
}
