// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package certmanager

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/crypto/keys"
	"go.probo.inc/probo/pkg/crypto/pem"
	"go.probo.inc/probo/pkg/version"
	"golang.org/x/crypto/acme"
)

type (
	Certificate struct {
		CertPEM   []byte
		KeyPEM    []byte
		ChainPEM  []byte
		ExpiresAt time.Time
	}

	ACMEService struct {
		client  *acme.Client
		email   string
		keyType keys.Type
		logger  *log.Logger
	}

	HTTPChallenge struct {
		Domain   string
		Token    string
		KeyAuth  string
		URL      string
		OrderURL string
	}
)

func NewACMEService(
	email string,
	keyType keys.Type,
	directoryURL string,
	accountKey crypto.Signer,
	rootCAs *x509.CertPool,
	logger *log.Logger,
) (*ACMEService, error) {
	if accountKey == nil {
		var err error

		accountKey, err = keys.Generate(keyType)
		if err != nil {
			return nil, fmt.Errorf("cannot generate account key: %w", err)
		}

		logger.Warn("no account key provided, generating new ACME account - this will create a new account on each restart")
	}

	httpClientOpts := []httpclient.Option{
		httpclient.WithLogger(logger),
	}

	if rootCAs != nil {
		httpClientOpts = append(
			httpClientOpts,
			httpclient.WithTLSConfig(&tls.Config{RootCAs: rootCAs}),
		)

		logger.Info("ACME service configured with custom root CA")
	}

	httpClient := httpclient.DefaultPooledClient(httpClientOpts...)

	client := &acme.Client{
		Key:          accountKey,
		DirectoryURL: directoryURL,
		UserAgent:    version.UserAgent("acme"),
		HTTPClient:   httpClient,
	}

	service := &ACMEService{
		client:  client,
		email:   email,
		keyType: keyType,
		logger:  logger.Named("acme"),
	}

	ctx := context.Background()
	if err := service.registerAccount(ctx); err != nil {
		return nil, fmt.Errorf("cannot register ACME account: %w", err)
	}

	return service, nil
}

func (s *ACMEService) registerAccount(ctx context.Context) error {
	account := &acme.Account{Contact: []string{"mailto:" + s.email}}

	if _, err := s.client.Register(ctx, account, acme.AcceptTOS); err != nil {
		if err != acme.ErrAccountAlreadyExists {
			return fmt.Errorf("cannot register account: %w", err)
		}
	}

	return nil
}

func (s *ACMEService) GetHTTPChallenge(ctx context.Context, domain string) (*HTTPChallenge, error) {
	order, err := s.client.AuthorizeOrder(ctx, acme.DomainIDs(domain))
	if err != nil {
		return nil, fmt.Errorf("cannot create order: %w", err)
	}

	var challenge *acme.Challenge

	for _, auth := range order.AuthzURLs {
		authz, err := s.client.GetAuthorization(ctx, auth)
		if err != nil {
			return nil, fmt.Errorf("cannot get authorization: %w", err)
		}

		for _, ch := range authz.Challenges {
			if ch.Type == "http-01" {
				challenge = ch
				break
			}
		}

		if challenge != nil {
			break
		}
	}

	if challenge == nil {
		return nil, fmt.Errorf("no HTTP-01 challenge found")
	}

	keyAuth, err := s.client.HTTP01ChallengeResponse(challenge.Token)
	if err != nil {
		return nil, fmt.Errorf("cannot get challenge response: %w", err)
	}

	return &HTTPChallenge{
		Domain:   domain,
		Token:    challenge.Token,
		KeyAuth:  keyAuth,
		URL:      challenge.URI,
		OrderURL: order.URI,
	}, nil
}

func (s *ACMEService) CompleteHTTPChallenge(
	ctx context.Context,
	challenge0 *HTTPChallenge,
) (*Certificate, error) {
	challenge1 := &acme.Challenge{
		URI:   challenge0.URL,
		Token: challenge0.Token,
	}

	if _, err := s.client.Accept(ctx, challenge1); err != nil {
		return nil, fmt.Errorf("cannot accept challenge: %w", err)
	}

	order, err := s.client.WaitOrder(ctx, challenge0.OrderURL)
	if err != nil {
		return nil, fmt.Errorf("cannot wait for order: %w", err)
	}

	certKey, err := keys.Generate(s.keyType)
	if err != nil {
		return nil, fmt.Errorf("cannot generate certificate key: %w", err)
	}

	csr, err := createCSR(challenge0.Domain, certKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create CSR: %w", err)
	}

	der, _, err := s.client.CreateOrderCert(ctx, order.FinalizeURL, csr, true)
	if err != nil {
		return nil, fmt.Errorf("cannot create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(der[0])
	if err != nil {
		return nil, fmt.Errorf("cannot parse certificate: %w", err)
	}

	certPEM := pem.EncodeCertificate(der[0])

	keyPEM, err := pem.EncodePrivateKey(certKey)
	if err != nil {
		return nil, fmt.Errorf("cannot encode key: %w", err)
	}

	var chainDER [][]byte
	if len(der) > 1 {
		chainDER = der[1:]
	}

	chainPEM := pem.EncodeCertificateChain(chainDER)

	return &Certificate{
		CertPEM:   certPEM,
		KeyPEM:    keyPEM,
		ChainPEM:  chainPEM,
		ExpiresAt: cert.NotAfter,
	}, nil
}

func (s *ACMEService) CheckRenewalNeeded(expiresAt time.Time, threshold time.Duration) bool {
	return time.Until(expiresAt) <= threshold
}

func createCSR(domain string, key crypto.Signer) ([]byte, error) {
	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: domain,
		},
		DNSNames: []string{domain},
	}

	return x509.CreateCertificateRequest(rand.Reader, template, key)
}
