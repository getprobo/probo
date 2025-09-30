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

package cert

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"time"

	"github.com/getprobo/probo/pkg/crypto/keys"
	cryptopem "github.com/getprobo/probo/pkg/crypto/pem"
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
	}

	DNSChallenge struct {
		Domain      string
		RecordName  string
		RecordValue string
		Token       string
		URL         string
		OrderURL    string
	}
)

func NewACMEService(email string, keyType keys.Type, directoryURL string) (*ACMEService, error) {
	accountKey, err := keys.Generate(keyType)
	if err != nil {
		return nil, fmt.Errorf("cannot generate account key: %w", err)
	}

	client := &acme.Client{Key: accountKey, DirectoryURL: directoryURL}

	service := &ACMEService{
		client:  client,
		email:   email,
		keyType: keyType,
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

func (s *ACMEService) GetDNSChallenge(ctx context.Context, domain string) (*DNSChallenge, error) {
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
			if ch.Type == "dns-01" {
				challenge = ch
				break
			}

			if ch.Type == "http-01" {
				return nil, fmt.Errorf("http-01 challenges are not supported")
			}
		}

		if challenge != nil {
			break
		}
	}

	if challenge == nil {
		return nil, fmt.Errorf("no DNS-01 challenge found")
	}

	recordValue, err := s.client.DNS01ChallengeRecord(challenge.Token)
	if err != nil {
		return nil, fmt.Errorf("cannot get DNS record value: %w", err)
	}

	return &DNSChallenge{
		Domain:      domain,
		RecordName:  fmt.Sprintf("_acme-challenge.%s", domain),
		RecordValue: recordValue,
		Token:       challenge.Token,
		URL:         challenge.URI,
		OrderURL:    order.URI,
	}, nil
}

func (s *ACMEService) CompleteDNSChallenge(
	ctx context.Context,
	challenge0 *DNSChallenge,
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

	certPEM := cryptopem.EncodeCertificate(der[0])
	keyPEM, err := cryptopem.EncodePrivateKey(certKey)
	if err != nil {
		return nil, fmt.Errorf("cannot encode key: %w", err)
	}

	var buf bytes.Buffer
	for i := 1; i < len(der); i++ {
		buf.Write(cryptopem.EncodeCertificate(der[i]))
	}

	return &Certificate{
		CertPEM:   certPEM,
		KeyPEM:    keyPEM,
		ChainPEM:  buf.Bytes(),
		ExpiresAt: cert.NotAfter,
	}, nil
}

func (s *ACMEService) InitiateRenewal(
	ctx context.Context,
	domain string,
) (*DNSChallenge, error) {
	challenge, err := s.GetDNSChallenge(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("cannot get DNS challenge for renewal: %w", err)
	}

	return challenge, nil
}

func (s *ACMEService) CompleteRenewal(
	ctx context.Context,
	challenge *DNSChallenge,
) (*Certificate, error) {
	cert, err := s.CompleteDNSChallenge(ctx, challenge)
	if err != nil {
		return nil, fmt.Errorf("cannot complete renewal challenge: %w", err)
	}

	return cert, nil
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
