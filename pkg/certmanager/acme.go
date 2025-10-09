// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package certmanager

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/getprobo/probo/pkg/crypto/keys"
	"github.com/getprobo/probo/pkg/crypto/pem"
	"github.com/getprobo/probo/pkg/version"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
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

// ErrHTTPChallengeRequired indicates that an HTTP-01 challenge needs to be
// completed before the certificate can be issued or renewed.
var ErrHTTPChallengeRequired = errors.New("HTTP challenge required")

func NewACMEService(email string, keyType keys.Type, directoryURL string, insecureTLS bool, logger *log.Logger) (*ACMEService, error) {
	accountKey, err := keys.Generate(keyType)
	if err != nil {
		return nil, fmt.Errorf("cannot generate account key: %w", err)
	}

	var httpClient *http.Client

	if insecureTLS {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient = &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		}
		logger.Warn("ACME service configured with insecure TLS - use only for local testing")
	} else {
		httpClient = httpclient.DefaultPooledClient(
			httpclient.WithLogger(logger),
		)
	}

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

func (s *ACMEService) ObtainCertificate(
	ctx context.Context,
	domain string,
) (*Certificate, error) {
	// For HTTP-01, we always need to serve the challenge
	challenge, err := s.GetHTTPChallenge(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("cannot get HTTP challenge: %w", err)
	}

	// The challenge token and key auth will be stored and served via HTTP
	// The caller is responsible for ensuring the HTTP endpoint is ready
	// before calling CompleteHTTPChallenge
	return nil, fmt.Errorf("%w: token=%s", ErrHTTPChallengeRequired, challenge.Token)
}

func (s *ACMEService) RenewCertificate(
	ctx context.Context,
	domain string,
) (*Certificate, error) {
	cert, err := s.renewWithExistingAuth(ctx, domain)
	if err == nil {
		return cert, nil
	}

	s.logger.WarnCtx(ctx, "renewal with existing authorization failed, need new HTTP challenge",
		log.String("domain", domain),
		log.Error(err))

	return s.ObtainCertificate(ctx, domain)
}

func (s *ACMEService) renewWithExistingAuth(ctx context.Context, domain string) (*Certificate, error) {
	order, err := s.client.AuthorizeOrder(ctx, acme.DomainIDs(domain))
	if err != nil {
		return nil, fmt.Errorf("cannot create renewal order: %w", err)
	}

	if order.Status != acme.StatusReady {
		order, err = s.client.WaitOrder(ctx, order.URI)
		if err != nil {
			return nil, fmt.Errorf("authorization not valid or expired: %w", err)
		}
	}

	certKey, err := keys.Generate(s.keyType)
	if err != nil {
		return nil, fmt.Errorf("cannot generate certificate key: %w", err)
	}

	csr, err := createCSR(domain, certKey)
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
