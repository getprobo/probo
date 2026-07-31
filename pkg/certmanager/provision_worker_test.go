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

package certmanager

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel/trace/noop"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/dnsclient"
)

// stubDNSChecker records the checks it was asked to run and returns canned
// errors for them.
type stubDNSChecker struct {
	cnameErr error
	caaErr   error

	cnameChecked bool
	caaChecked   bool
}

func (c *stubDNSChecker) CheckCNAME(ctx context.Context, hostname, expectedTarget string) error {
	c.cnameChecked = true

	return c.cnameErr
}

func (c *stubDNSChecker) CheckCAA(ctx context.Context, hostname, permittedIssuer string) error {
	c.caaChecked = true

	return c.caaErr
}

func newBeginChallengeHandler(dnsClient dnsChecker, skipCNAMECheck bool) *beginChallengeHandler {
	return &beginChallengeHandler{
		provisionCore: provisionCore{
			logger: log.NewLogger(log.WithName("test")),
			tracer: noop.NewTracerProvider().Tracer("test"),
		},
		cnameTarget:     "custom.getprobo.com",
		caaIssuerDomain: "letsencrypt.org",
		dnsClient:       dnsClient,
		skipCNAMECheck:  skipCNAMECheck,
	}
}

func TestClassifyProvisioningError(t *testing.T) {
	t.Parallel()

	assert.Equal(t, ProvisioningErrorACMERateLimited, classifyProvisioningError(ErrACMERateLimited))
	assert.Equal(t, ProvisioningErrorACMEInvalidOrder, classifyProvisioningError(ErrOrderInvalid))
	assert.Equal(t, ProvisioningErrorDNSCNAME, classifyProvisioningError(errors.New("cname target mismatch")))
	assert.Equal(t, ProvisioningErrorDNSCAA, classifyProvisioningError(fmt.Errorf("%w: domain %q", ErrCAANotPermitted, "example.com")))
	assert.Equal(t, ProvisioningErrorACMETemporary, classifyProvisioningError(errors.New("network timeout")))
	// A CAA resolver/transport failure shares the "caa records" wording with a
	// real CAA misconfiguration but must consume the normal retry budget.
	assert.Equal(t, ProvisioningErrorACMETemporary, classifyProvisioningError(errors.New("cannot exchange dns message for caa records: i/o timeout")))
}

func TestDecideProvisioningOutcome_RateLimitKeepsRetryCountAndOrder(t *testing.T) {
	t.Parallel()

	orderURL := "https://acme.example/order/1"
	certificate := &coredata.Certificate{
		Status:        coredata.CertificateStatusProvisioning,
		SSLRetryCount: 2,
		HTTPOrderURL:  &orderURL,
	}

	outcome := decideProvisioningOutcome(certificate, ProvisioningErrorACMERateLimited)

	assert.Equal(t, coredata.CertificateStatusProvisioning, outcome.status)
	assert.Equal(t, 2, outcome.retryCount)
	assert.False(t, outcome.clearACMEState)
	assert.Equal(t, ProvisioningErrorACMERateLimited, outcome.errorCode)
}

func TestDecideProvisioningOutcome_RateLimitThenTransientDoesNotFail(t *testing.T) {
	t.Parallel()

	orderURL := "https://acme.example/order/1"
	certificate := &coredata.Certificate{
		Status:        coredata.CertificateStatusProvisioning,
		SSLRetryCount: 0,
		HTTPOrderURL:  &orderURL,
	}

	rateLimited := decideProvisioningOutcome(certificate, ProvisioningErrorACMERateLimited)
	certificate.SSLRetryCount = rateLimited.retryCount
	certificate.Status = rateLimited.status

	// Previously rate-limit floored ssl_retry_count to 5, so the next transient
	// immediately crossed maxProvisioningRetries and marked FAILED.
	transient := decideProvisioningOutcome(certificate, ProvisioningErrorACMETemporary)

	assert.Equal(t, coredata.CertificateStatusProvisioning, transient.status)
	assert.Equal(t, 1, transient.retryCount)
	assert.False(t, transient.clearACMEState)
	assert.NotEqual(t, coredata.CertificateStatusFailed, transient.status)
}

func TestDecideProvisioningOutcome_MarksFailedAfterMaxRetries(t *testing.T) {
	t.Parallel()

	certificate := &coredata.Certificate{
		Status:        coredata.CertificateStatusProvisioning,
		SSLRetryCount: maxProvisioningRetries - 1,
	}

	outcome := decideProvisioningOutcome(certificate, ProvisioningErrorACMEInvalidOrder)

	assert.Equal(t, coredata.CertificateStatusFailed, outcome.status)
	assert.Equal(t, maxProvisioningRetries, outcome.retryCount)
	assert.True(t, outcome.clearACMEState)
	assert.Equal(t, ProvisioningErrorACMEFailed, outcome.errorCode)
}

func TestDecideProvisioningOutcome_TransientPreservesOrder(t *testing.T) {
	t.Parallel()

	orderURL := "https://acme.example/order/1"
	certificate := &coredata.Certificate{
		Status:        coredata.CertificateStatusProvisioning,
		SSLRetryCount: 0,
		HTTPOrderURL:  &orderURL,
	}

	outcome := decideProvisioningOutcome(certificate, ProvisioningErrorACMETemporary)

	assert.Equal(t, coredata.CertificateStatusProvisioning, outcome.status)
	assert.Equal(t, 1, outcome.retryCount)
	assert.False(t, outcome.clearACMEState)
	assert.Equal(t, ProvisioningErrorACMETemporary, outcome.errorCode)
}

func TestDecideProvisioningOutcome_InvalidOrderClearsState(t *testing.T) {
	t.Parallel()

	orderURL := "https://acme.example/order/1"
	certificate := &coredata.Certificate{
		Status:        coredata.CertificateStatusProvisioning,
		SSLRetryCount: 0,
		HTTPOrderURL:  &orderURL,
	}

	outcome := decideProvisioningOutcome(certificate, ProvisioningErrorACMEInvalidOrder)

	assert.Equal(t, coredata.CertificateStatusPending, outcome.status)
	assert.Equal(t, 1, outcome.retryCount)
	assert.True(t, outcome.clearACMEState)
	assert.Equal(t, ProvisioningErrorACMEInvalidOrder, outcome.errorCode)
}

func TestDecideProvisioningOutcome_DNSIsNonTerminal(t *testing.T) {
	t.Parallel()

	certificate := &coredata.Certificate{
		Status:        coredata.CertificateStatusPending,
		SSLRetryCount: 0,
	}

	outcome := decideProvisioningOutcome(certificate, ProvisioningErrorDNSCNAME)

	assert.Equal(t, coredata.CertificateStatusPending, outcome.status)
	assert.Equal(t, 0, outcome.retryCount)
	assert.True(t, outcome.clearACMEState)
	assert.Equal(t, ProvisioningErrorDNSCNAME, outcome.errorCode)
}

func TestDecideProvisioningOutcome_RateLimitWithoutOrderStaysPending(t *testing.T) {
	t.Parallel()

	certificate := &coredata.Certificate{
		Status:        coredata.CertificateStatusPending,
		SSLRetryCount: 1,
	}

	outcome := decideProvisioningOutcome(certificate, ProvisioningErrorACMERateLimited)

	require.Equal(t, coredata.CertificateStatusPending, outcome.status)
	assert.Equal(t, 1, outcome.retryCount)
	assert.False(t, outcome.clearACMEState)
}

func TestCheckDNSConfiguration_CNAMEMismatchBlocksByDefault(t *testing.T) {
	t.Parallel()

	cnameErr := errors.New("no cname records found for domain \"trust.example.com\"")
	dnsClient := &stubDNSChecker{cnameErr: cnameErr}

	err := newBeginChallengeHandler(dnsClient, false).
		checkDNSConfiguration(context.Background(), "trust.example.com")

	require.ErrorIs(t, err, cnameErr)
	assert.False(t, dnsClient.caaChecked)
}

func TestCheckDNSConfiguration_CNAMEMismatchIsAHintWhenSkipped(t *testing.T) {
	t.Parallel()

	// A hostname behind a proxying CDN never exposes its origin CNAME in public
	// DNS, so the pre-check cannot pass even though HTTP-01 would.
	dnsClient := &stubDNSChecker{cnameErr: errors.New("no cname records found for domain \"trust.example.com\"")}

	err := newBeginChallengeHandler(dnsClient, true).
		checkDNSConfiguration(context.Background(), "trust.example.com")

	require.NoError(t, err)
	assert.True(t, dnsClient.caaChecked)
}

func TestCheckDNSConfiguration_CAAStillBlocksWhenCNAMECheckIsSkipped(t *testing.T) {
	t.Parallel()

	dnsClient := &stubDNSChecker{
		cnameErr: errors.New("no cname records found for domain \"trust.example.com\""),
		caaErr:   dnsclient.ErrCAADenied,
	}

	err := newBeginChallengeHandler(dnsClient, true).
		checkDNSConfiguration(context.Background(), "trust.example.com")

	require.ErrorIs(t, err, ErrCAANotPermitted)
}

func TestCheckDNSConfiguration_RunsBothChecks(t *testing.T) {
	t.Parallel()

	dnsClient := &stubDNSChecker{}

	err := newBeginChallengeHandler(dnsClient, false).
		checkDNSConfiguration(context.Background(), "trust.example.com")

	require.NoError(t, err)
	assert.True(t, dnsClient.cnameChecked)
	assert.True(t, dnsClient.caaChecked)
}
