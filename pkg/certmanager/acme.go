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
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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
		client        *acme.Client
		email         string
		keyType       keys.Type
		logger        *log.Logger
		metrics       *metrics
		cooldownMu    sync.RWMutex
		cooldownUntil time.Time
	}

	HTTPChallenge struct {
		Domain   string
		Token    string
		KeyAuth  string
		URL      string
		OrderURL string
	}

	OrderPollStatus string
)

const (
	OrderPollStatusNotReady OrderPollStatus = "not_ready"
	OrderPollStatusReady    OrderPollStatus = "ready"
	OrderPollStatusValid    OrderPollStatus = "valid"
	OrderPollStatusInvalid  OrderPollStatus = "invalid"
)

type OrderPollResult struct {
	Status OrderPollStatus
	Order  *acme.Order
}

func NewACMEService(
	email string,
	keyType keys.Type,
	directoryURL string,
	accountKey crypto.Signer,
	rootCAs *x509.CertPool,
	logger *log.Logger,
	registerer prometheus.Registerer,
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
		metrics: newMetrics(registerer),
	}

	ctx := context.Background()
	if err := service.registerAccount(ctx); err != nil {
		return nil, fmt.Errorf("cannot register ACME account: %w", err)
	}

	return service, nil
}

func (s *ACMEService) InCooldown() bool {
	s.cooldownMu.RLock()
	defer s.cooldownMu.RUnlock()

	active := time.Now().Before(s.cooldownUntil)

	s.metrics.setCooldown(active)

	return active
}

func (s *ACMEService) CooldownUntil() time.Time {
	s.cooldownMu.RLock()
	defer s.cooldownMu.RUnlock()

	return s.cooldownUntil
}

func (s *ACMEService) enterCooldown(until time.Time) {
	s.cooldownMu.Lock()
	defer s.cooldownMu.Unlock()

	if until.After(s.cooldownUntil) {
		s.cooldownUntil = until
	}

	s.metrics.setCooldown(time.Now().Before(s.cooldownUntil))
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

func (s *ACMEService) StartHTTPChallenge(ctx context.Context, domain string) (*HTTPChallenge, error) {
	started := time.Now()

	order, err := s.client.AuthorizeOrder(ctx, acme.DomainIDs(domain))
	if err != nil {
		return nil, s.handleError(provisionPhaseCreateOrder, started, "cannot create order", err)
	}

	var challenge *acme.Challenge

	for _, auth := range order.AuthzURLs {
		authz, err := s.client.GetAuthorization(ctx, auth)
		if err != nil {
			return nil, s.handleError(provisionPhaseCreateOrder, started, "cannot get authorization", err)
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
		s.metrics.observeStep(provisionPhaseCreateOrder, provisionResultError, started)
		return nil, fmt.Errorf("no HTTP-01 challenge found")
	}

	keyAuth, err := s.client.HTTP01ChallengeResponse(challenge.Token)
	if err != nil {
		s.metrics.observeStep(provisionPhaseCreateOrder, provisionResultError, started)
		return nil, fmt.Errorf("cannot get challenge response: %w", err)
	}

	s.metrics.observeStep(provisionPhaseCreateOrder, provisionResultOK, started)

	return &HTTPChallenge{
		Domain:   domain,
		Token:    challenge.Token,
		KeyAuth:  keyAuth,
		URL:      challenge.URI,
		OrderURL: order.URI,
	}, nil
}

func (s *ACMEService) AcceptHTTPChallenge(ctx context.Context, challenge *HTTPChallenge) error {
	started := time.Now()

	acceptChallenge := &acme.Challenge{
		URI:   challenge.URL,
		Token: challenge.Token,
	}

	if _, err := s.client.Accept(ctx, acceptChallenge); err != nil && !isChallengeAlreadyValid(err) {
		return s.handleError(provisionPhaseCreateOrder, started, "cannot accept challenge", err)
	}

	s.metrics.observeStep(provisionPhaseCreateOrder, provisionResultOK, started)

	return nil
}

func (s *ACMEService) PollOrder(ctx context.Context, orderURL string) (*OrderPollResult, error) {
	started := time.Now()

	order, err := s.client.GetOrder(ctx, orderURL)
	if err != nil {
		return nil, s.handleError(provisionPhasePollOrder, started, "cannot get order", err)
	}

	result := &OrderPollResult{Order: order}

	switch order.Status {
	case acme.StatusPending, acme.StatusProcessing:
		result.Status = OrderPollStatusNotReady

		s.metrics.observeStep(provisionPhasePollOrder, provisionResultNotReady, started)
	case acme.StatusReady:
		result.Status = OrderPollStatusReady

		s.metrics.observeStep(provisionPhasePollOrder, provisionResultOK, started)
	case acme.StatusValid:
		result.Status = OrderPollStatusValid

		s.metrics.observeStep(provisionPhasePollOrder, provisionResultOK, started)
	case acme.StatusInvalid:
		result.Status = OrderPollStatusInvalid

		s.metrics.observeStep(provisionPhasePollOrder, provisionResultError, started)
	default:
		s.metrics.observeStep(provisionPhasePollOrder, provisionResultError, started)
		return nil, fmt.Errorf("order is in unexpected status %q", order.Status)
	}

	return result, nil
}

func (s *ACMEService) IssueCertificate(
	ctx context.Context,
	challenge *HTTPChallenge,
	poll *OrderPollResult,
) (*Certificate, error) {
	started := time.Now()

	if poll == nil || poll.Order == nil {
		s.metrics.observeStep(provisionPhaseIssueCert, provisionResultError, started)
		return nil, fmt.Errorf("missing order to issue certificate")
	}

	if poll.Status == OrderPollStatusInvalid {
		s.metrics.observeStep(provisionPhaseIssueCert, provisionResultError, started)
		return nil, ErrOrderInvalid
	}

	if poll.Status != OrderPollStatusReady && poll.Status != OrderPollStatusValid {
		s.metrics.observeStep(provisionPhaseIssueCert, provisionResultNotReady, started)
		return nil, ErrOrderNotReady
	}

	certKey, err := keys.Generate(s.keyType)
	if err != nil {
		s.metrics.observeStep(provisionPhaseIssueCert, provisionResultError, started)
		return nil, fmt.Errorf("cannot generate certificate key: %w", err)
	}

	csr, err := createCSR(challenge.Domain, certKey)
	if err != nil {
		s.metrics.observeStep(provisionPhaseIssueCert, provisionResultError, started)
		return nil, fmt.Errorf("cannot create CSR: %w", err)
	}

	der, err := s.issueOrderCertificate(ctx, poll.Order, challenge.OrderURL, csr)
	if err != nil {
		return nil, s.handleError(provisionPhaseIssueCert, started, "cannot create certificate", err)
	}

	cert, err := x509.ParseCertificate(der[0])
	if err != nil {
		s.metrics.observeStep(provisionPhaseIssueCert, provisionResultError, started)
		return nil, fmt.Errorf("cannot parse certificate: %w", err)
	}

	certPEM := pem.EncodeCertificate(der[0])

	keyPEM, err := pem.EncodePrivateKey(certKey)
	if err != nil {
		s.metrics.observeStep(provisionPhaseIssueCert, provisionResultError, started)
		return nil, fmt.Errorf("cannot encode key: %w", err)
	}

	var chainDER [][]byte
	if len(der) > 1 {
		chainDER = der[1:]
	}

	chainPEM := pem.EncodeCertificateChain(chainDER)

	s.metrics.observeStep(provisionPhaseIssueCert, provisionResultOK, started)

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

func (s *ACMEService) handleError(
	phase provisionPhase,
	started time.Time,
	op string,
	err error,
) error {
	acmeErr := newACMEError(op, err)
	s.metrics.recordACMEError(acmeErr.problemType)

	result := provisionResultError
	if acmeErr.rateLimited {
		result = provisionResultRateLimited

		s.enterCooldown(time.Now().Add(acmeErr.RetryAfter()))
	}

	s.metrics.observeStep(phase, result, started)

	return acmeErr
}

func isChallengeAlreadyValid(err error) bool {
	acmeErr, ok := errors.AsType[*acme.Error](err)
	if !ok {
		return false
	}

	return acmeErr.ProblemType == "urn:ietf:params:acme:error:malformed" &&
		strings.Contains(acmeErr.Detail, "status valid")
}

func (s *ACMEService) issueOrderCertificate(
	ctx context.Context,
	order *acme.Order,
	orderURL string,
	csr []byte,
) ([][]byte, error) {
	pollURL := acmeOrderURL(order, orderURL)

	switch order.Status {
	case acme.StatusValid:
		return s.fetchOrderCertificate(ctx, pollURL, order)
	case acme.StatusReady:
		if order.FinalizeURL == "" {
			return nil, fmt.Errorf("order is ready but finalize URL is missing")
		}

		der, _, err := s.client.CreateOrderCert(ctx, order.FinalizeURL, csr, true)
		if err == nil {
			return der, nil
		}

		return s.fetchOrderCertificateAfterFinalize(ctx, pollURL, err)
	default:
		return nil, fmt.Errorf("order is in unexpected status %q", order.Status)
	}
}

func acmeOrderURL(order *acme.Order, orderURL string) string {
	if order.URI != "" {
		return order.URI
	}

	return orderURL
}

func (s *ACMEService) fetchOrderCertificateAfterFinalize(
	ctx context.Context,
	orderURL string,
	finalizeErr error,
) ([][]byte, error) {
	refreshed, err := s.client.GetOrder(ctx, orderURL)
	if err != nil {
		return nil, fmt.Errorf("cannot refresh order after finalize: %w", err)
	}

	if refreshed.Status != acme.StatusValid {
		return nil, fmt.Errorf("%w: order status %q after finalize", ErrOrderNotReady, refreshed.Status)
	}

	der, err := s.fetchOrderCertificate(ctx, orderURL, refreshed)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch certificate after finalize: %w: %w", finalizeErr, err)
	}

	return der, nil
}

func (s *ACMEService) fetchOrderCertificate(
	ctx context.Context,
	orderURL string,
	order *acme.Order,
) ([][]byte, error) {
	orderWithCertURL, err := s.refreshOrderCertificateURL(ctx, orderURL, order)
	if err != nil {
		return nil, err
	}

	der, err := s.client.FetchCert(ctx, orderWithCertURL.CertURL, true)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch certificate: %w", err)
	}

	return der, nil
}

func (s *ACMEService) refreshOrderCertificateURL(
	ctx context.Context,
	orderURL string,
	order *acme.Order,
) (*acme.Order, error) {
	if order.CertURL != "" {
		return order, nil
	}

	refreshed, err := s.client.GetOrder(ctx, orderURL)
	if err != nil {
		return nil, fmt.Errorf("cannot refresh order: %w", err)
	}

	if refreshed.CertURL != "" {
		return refreshed, nil
	}

	if refreshed.Status != acme.StatusValid {
		return nil, fmt.Errorf("%w: order left valid state while waiting for certificate URL", ErrOrderNotReady)
	}

	return nil, fmt.Errorf("%w: certificate URL not yet available", ErrOrderNotReady)
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
