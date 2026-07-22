// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"errors"
	"fmt"
	"strings"
	"time"

	"codeberg.org/miekg/dns"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

const (
	// maxProvisioningRetries is the failure budget for ordinary transient ACME
	// errors. Rate limits do not consume this budget; they use the in-process
	// ACME cooldown instead. Claim SQL still caps the exponential backoff
	// exponent at 5 (LEAST(ssl_retry_count, 5)), so normal retries only reach
	// exponents 0–2 before FAILED.
	maxProvisioningRetries = 3
	dnsExchangeTimeout     = 10 * time.Second
	processTickTimeout     = 90 * time.Second

	tracerName = "go.probo.inc/probo/pkg/certmanager"
)

type (
	provisionHandler struct {
		pg                *pg.Client
		acmeService       *ACMEService
		encryptionKey     cipher.EncryptionKey
		cnameTarget       string
		caaIssuerDomain   string
		resolverAddr      string
		managedBaseDomain string
		logger            *log.Logger
		tracer            trace.Tracer
	}

	provisioningOutcome struct {
		status         coredata.CertificateStatus
		retryCount     int
		clearACMEState bool
		errorCode      string
	}
)

var (
	_ worker.Handler[coredata.Certificate] = (*provisionHandler)(nil)
	_ worker.StaleRecoverer                = (*provisionHandler)(nil)
)

func NewProvisionWorker(
	pgClient *pg.Client,
	acmeService *ACMEService,
	encryptionKey cipher.EncryptionKey,
	cnameTarget string,
	caaIssuerDomain string,
	resolverAddr string,
	managedBaseDomain string,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.Certificate] {
	h := &provisionHandler{
		pg:                pgClient,
		acmeService:       acmeService,
		encryptionKey:     encryptionKey,
		cnameTarget:       cnameTarget,
		caaIssuerDomain:   caaIssuerDomain,
		resolverAddr:      resolverAddr,
		managedBaseDomain: managedBaseDomain,
		logger:            logger,
		tracer:            otel.Tracer(tracerName),
	}

	opts = append(opts, worker.WithMaxConcurrency(1))

	return worker.New(
		"certificate-provision-worker",
		h,
		logger,
		opts...,
	)
}

func (h *provisionHandler) Claim(ctx context.Context) (coredata.Certificate, error) {
	if h.acmeService.InCooldown() {
		return coredata.Certificate{}, worker.ErrNoTask
	}

	var certificate coredata.Certificate

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := certificate.LoadNextForProvisioningForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			now := time.Now()
			certificate.SSLLastAttemptAt = &now

			if err := certificate.Update(ctx, tx, coredata.NewNoScope()); err != nil {
				return fmt.Errorf("cannot stamp certificate claim: %w", err)
			}

			return nil
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.Certificate{}, worker.ErrNoTask
		}

		return coredata.Certificate{}, err
	}

	return certificate, nil
}

func (h *provisionHandler) Process(ctx context.Context, certificate coredata.Certificate) error {
	ctx, cancel := context.WithTimeout(ctx, processTickTimeout)
	defer cancel()

	switch certificate.Status {
	case coredata.CertificateStatusPending, coredata.CertificateStatusRenewing:
		return h.processBeginChallenge(ctx, certificate)
	case coredata.CertificateStatusProvisioning:
		return h.processPollOrder(ctx, certificate)
	default:
		return nil
	}
}

func (h *provisionHandler) processBeginChallenge(
	ctx context.Context,
	certificate coredata.Certificate,
) error {
	ctx, span := h.tracer.Start(ctx, "certmanager.create_order")
	defer span.End()

	h.setCertificateSpanAttributes(span, certificate)

	skipDNSChecks, err := h.loadSkipDNSChecks(ctx, certificate.Hostname)
	if err != nil {
		h.recordSpanError(span, err, "")
		return err
	}

	if !skipDNSChecks {
		dnsCtx, dnsSpan := h.tracer.Start(ctx, "certmanager.dns_check")
		dnsStarted := time.Now()

		if err := h.checkDNSConfiguration(dnsCtx, certificate.Hostname); err != nil {
			h.acmeService.metrics.observeStep(provisionPhaseDNSCheck, provisionResultDNSError, dnsStarted)
			h.recordSpanError(dnsSpan, err, classifyProvisioningError(err))
			dnsSpan.End()
			h.recordSpanError(span, err, classifyProvisioningError(err))
			// DNS/CAA misconfig is intentionally non-terminal: retry forever so a
			// customer DNS fix auto-recovers without marking the domain FAILED.
			return h.persistFailure(ctx, certificate.ID, err)
		}

		if err := h.checkCAARecords(dnsCtx, certificate.Hostname); err != nil {
			h.acmeService.metrics.observeStep(provisionPhaseDNSCheck, provisionResultDNSError, dnsStarted)
			h.recordSpanError(dnsSpan, err, classifyProvisioningError(err))
			dnsSpan.End()
			h.recordSpanError(span, err, classifyProvisioningError(err))
			// DNS/CAA misconfig is intentionally non-terminal (see above).
			return h.persistFailure(ctx, certificate.ID, err)
		}

		h.acmeService.metrics.observeStep(provisionPhaseDNSCheck, provisionResultOK, dnsStarted)
		dnsSpan.End()
	}

	challenge, err := h.acmeService.StartHTTPChallenge(ctx, certificate.Hostname)
	if err != nil {
		errorCode := classifyProvisioningError(err)
		h.logACMEOutcome(ctx, certificate, provisionPhaseCreateOrder, err, errorCode)
		h.recordSpanError(span, err, errorCode)

		return h.persistFailure(ctx, certificate.ID, err)
	}

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			row := &coredata.Certificate{}
			if err := row.LoadByIDForUpdateSkipLocked(ctx, tx, coredata.NewNoScope(), certificate.ID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load certificate %q: %w", certificate.ID, err)
			}

			if row.Status != coredata.CertificateStatusPending &&
				row.Status != coredata.CertificateStatusRenewing {
				return nil
			}

			row.HTTPChallengeToken = &challenge.Token
			row.HTTPChallengeKeyAuth = &challenge.KeyAuth
			row.HTTPChallengeURL = &challenge.URL
			row.HTTPOrderURL = &challenge.OrderURL
			row.Status = coredata.CertificateStatusProvisioning
			row.ProvisioningError = nil

			if err := row.Update(ctx, tx, coredata.NewNoScope()); err != nil {
				return fmt.Errorf("cannot update certificate with challenge: %w", err)
			}

			h.logger.InfoCtx(
				ctx,
				"HTTP challenge accepted, waiting for order validation",
				log.String("hostname", row.Hostname),
				log.String("certificate_id", row.ID.String()),
			)

			return nil
		},
	)
}

func (h *provisionHandler) processPollOrder(
	ctx context.Context,
	certificate coredata.Certificate,
) error {
	ctx, span := h.tracer.Start(ctx, "certmanager.poll_order")
	defer span.End()

	h.setCertificateSpanAttributes(span, certificate)

	if certificate.HTTPOrderURL == nil {
		return h.persistFailure(
			ctx,
			certificate.ID,
			fmt.Errorf("provisioning certificate missing order URL"),
		)
	}

	challenge := &HTTPChallenge{
		Domain:   certificate.Hostname,
		Token:    stringValue(certificate.HTTPChallengeToken),
		KeyAuth:  stringValue(certificate.HTTPChallengeKeyAuth),
		URL:      stringValue(certificate.HTTPChallengeURL),
		OrderURL: *certificate.HTTPOrderURL,
	}

	poll, err := h.acmeService.PollOrder(ctx, *certificate.HTTPOrderURL)
	if err != nil {
		errorCode := classifyProvisioningError(err)
		h.logACMEOutcome(ctx, certificate, provisionPhasePollOrder, err, errorCode)
		h.recordSpanError(span, err, errorCode)

		return h.persistFailure(ctx, certificate.ID, err)
	}

	switch poll.Status {
	case OrderPollStatusNotReady:
		h.logger.InfoCtx(
			ctx,
			"ACME order not ready yet, will poll again",
			log.String("hostname", certificate.Hostname),
			log.String("certificate_id", certificate.ID.String()),
			log.String("order_status", poll.Order.Status),
		)

		return nil
	case OrderPollStatusInvalid:
		err := ErrOrderInvalid
		errorCode := classifyProvisioningError(err)
		h.logACMEOutcome(ctx, certificate, provisionPhasePollOrder, err, errorCode)
		h.recordSpanError(span, err, errorCode)

		return h.persistFailure(ctx, certificate.ID, err)
	case OrderPollStatusReady, OrderPollStatusValid:
		return h.issueCertificate(ctx, certificate, challenge, poll)
	default:
		return nil
	}
}

func (h *provisionHandler) issueCertificate(
	ctx context.Context,
	certificate coredata.Certificate,
	challenge *HTTPChallenge,
	poll *OrderPollResult,
) error {
	ctx, span := h.tracer.Start(ctx, "certmanager.issue_cert")
	defer span.End()

	h.setCertificateSpanAttributes(span, certificate)

	cert, err := h.acmeService.IssueCertificate(ctx, challenge, poll)
	if err != nil {
		if errors.Is(err, ErrOrderNotReady) {
			h.logger.InfoCtx(
				ctx,
				"ACME order not ready to issue yet, will poll again",
				log.String("hostname", certificate.Hostname),
				log.String("certificate_id", certificate.ID.String()),
			)

			return nil
		}

		errorCode := classifyProvisioningError(err)
		h.logACMEOutcome(ctx, certificate, provisionPhaseIssueCert, err, errorCode)
		h.recordSpanError(span, err, errorCode)

		return h.persistFailure(ctx, certificate.ID, err)
	}

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			row := &coredata.Certificate{}
			if err := row.LoadByIDForUpdate(ctx, tx, coredata.NewNoScope(), certificate.ID); err != nil {
				return fmt.Errorf("cannot load certificate %q: %w", certificate.ID, err)
			}

			h.logger.InfoCtx(
				ctx,
				"certificate obtained successfully",
				log.String("hostname", row.Hostname),
				log.String("certificate_id", row.ID.String()),
				log.Time("expires_at", cert.ExpiresAt),
			)

			row.ProvisioningError = nil
			row.SSLCertificatePEM = cert.CertPEM

			if err := row.EncryptPrivateKey(cert.KeyPEM, h.encryptionKey); err != nil {
				return fmt.Errorf("cannot encrypt private key: %w", err)
			}

			chainStr := string(cert.ChainPEM)
			row.SSLCertificateChain = &chainStr
			row.SSLExpiresAt = &cert.ExpiresAt
			row.Status = coredata.CertificateStatusActive
			row.SSLRetryCount = 0
			row.SSLLastAttemptAt = nil
			row.HTTPChallengeToken = nil
			row.HTTPChallengeKeyAuth = nil
			row.HTTPChallengeURL = nil
			row.HTTPOrderURL = nil

			if err := row.Update(ctx, tx, coredata.NewNoScope()); err != nil {
				return fmt.Errorf("cannot update certificate: %w", err)
			}

			cache := &coredata.CachedCertificate{
				Domain:           row.Hostname,
				CertificatePEM:   string(cert.CertPEM),
				PrivateKeyPEM:    string(cert.KeyPEM),
				CertificateChain: &chainStr,
				ExpiresAt:        cert.ExpiresAt,
				CachedAt:         time.Now(),
				CertificateID:    row.ID,
			}

			if err := cache.Upsert(ctx, tx); err != nil {
				h.logger.ErrorCtx(
					ctx,
					"cannot update certificate cache",
					log.String("hostname", row.Hostname),
					log.Error(err),
				)
			}

			return nil
		},
	)
}

func (h *provisionHandler) persistFailure(
	ctx context.Context,
	certificateID gid.GID,
	provisionErr error,
) error {
	errorCode := classifyProvisioningError(provisionErr)

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			row := &coredata.Certificate{}
			if err := row.LoadByIDForUpdateSkipLocked(ctx, tx, coredata.NewNoScope(), certificateID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load certificate %q: %w", certificateID, err)
			}

			outcome := decideProvisioningOutcome(row, errorCode)

			row.ProvisioningError = provisioningErrorCodePtr(outcome.errorCode)
			now := time.Now()
			row.SSLLastAttemptAt = &now
			row.Status = outcome.status
			row.SSLRetryCount = outcome.retryCount

			if outcome.clearACMEState {
				row.HTTPChallengeToken = nil
				row.HTTPChallengeKeyAuth = nil
				row.HTTPChallengeURL = nil
				row.HTTPOrderURL = nil
			}

			if err := row.Update(ctx, tx, coredata.NewNoScope()); err != nil {
				return fmt.Errorf("cannot update certificate with provisioning error: %w", err)
			}

			return nil
		},
	)
}

// decideProvisioningOutcome computes status/retry/clear policy for a failed step.
//
// Two backoff regimes:
//   - Normal transient errors: ssl_retry_count increments; at maxProvisioningRetries
//     the domain becomes FAILED.
//   - Rate limits: never increment ssl_retry_count and never mark FAILED; the
//     in-process ACME cooldown gates Claim. When an order URL is present
//     it is preserved so the next tick resumes polling instead of minting a new
//     order.
//
// DNS/CAA misconfig is intentionally non-terminal so customer DNS fixes
// auto-recover on the claim backoff schedule.
func decideProvisioningOutcome(
	certificate *coredata.Certificate,
	errorCode string,
) provisioningOutcome {
	retryCount := certificate.SSLRetryCount
	hasResumableOrder := certificate.HTTPOrderURL != nil && *certificate.HTTPOrderURL != ""

	switch errorCode {
	case ProvisioningErrorACMERateLimited:
		status := coredata.CertificateStatusPending
		if hasResumableOrder {
			status = coredata.CertificateStatusProvisioning
		}

		return provisioningOutcome{
			status:         status,
			retryCount:     retryCount,
			clearACMEState: false,
			errorCode:      errorCode,
		}

	case ProvisioningErrorDNSCNAME, ProvisioningErrorDNSCAA:
		return provisioningOutcome{
			status:         coredata.CertificateStatusPending,
			retryCount:     retryCount,
			clearACMEState: true,
			errorCode:      errorCode,
		}
	}

	retryCount++
	if retryCount >= maxProvisioningRetries {
		return provisioningOutcome{
			status:         coredata.CertificateStatusFailed,
			retryCount:     retryCount,
			clearACMEState: true,
			errorCode:      ProvisioningErrorACMEFailed,
		}
	}

	if errorCode == ProvisioningErrorACMEInvalidOrder || !hasResumableOrder {
		return provisioningOutcome{
			status:         coredata.CertificateStatusPending,
			retryCount:     retryCount,
			clearACMEState: true,
			errorCode:      errorCode,
		}
	}

	return provisioningOutcome{
		status:         coredata.CertificateStatusProvisioning,
		retryCount:     retryCount,
		clearACMEState: false,
		errorCode:      errorCode,
	}
}

func (h *provisionHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var certificates coredata.Certificates
			if err := certificates.ListStaleProvisioning(ctx, tx, coredata.NewNoScope()); err != nil {
				return fmt.Errorf("cannot load stale provisioning certificates: %w", err)
			}

			if len(certificates) == 0 {
				return nil
			}

			h.logger.InfoCtx(ctx, "found stale provisioning attempts to reset", log.Int("count", len(certificates)))

			for _, certificate := range certificates {
				if err := h.resetStaleCertificate(ctx, tx, certificate); err != nil {
					h.logger.ErrorCtx(
						ctx,
						"cannot reset stale certificate",
						log.String("hostname", certificate.Hostname),
						log.Error(err),
					)
				}
			}

			return nil
		},
	)
}

func (h *provisionHandler) logACMEOutcome(
	ctx context.Context,
	certificate coredata.Certificate,
	phase provisionPhase,
	err error,
	errorCode string,
) {
	var (
		problemType string
		detail      string
	)
	if acmeErr, ok := errors.AsType[*ACMEError](err); ok {
		problemType = acmeErr.ProblemType()
		detail = acmeErr.Detail()
	}

	level := h.logger.WarnCtx

	if errorCode == ProvisioningErrorACMETemporary {
		level = h.logger.ErrorCtx
	}

	level(
		ctx,
		"certificate provisioning step failed",
		log.String("hostname", certificate.Hostname),
		log.String("certificate_id", certificate.ID.String()),
		log.String("phase", string(phase)),
		log.String("error_code", errorCode),
		log.String("acme_problem_type", problemType),
		log.String("acme_detail", detail),
		log.Int("retry_count", certificate.SSLRetryCount),
		log.Time("cool_down_until", h.acmeService.CooldownUntil()),
		log.Error(err),
	)
}

func (h *provisionHandler) setCertificateSpanAttributes(span trace.Span, certificate coredata.Certificate) {
	span.SetAttributes(
		attribute.String("certificate.id", certificate.ID.String()),
		attribute.String("certificate.hostname", certificate.Hostname),
		attribute.String("certificate.status", string(certificate.Status)),
	)
}

func (h *provisionHandler) recordSpanError(span trace.Span, err error, errorCode string) {
	if err == nil {
		return
	}

	var (
		problemType string
		detail      string
	)
	if acmeErr, ok := errors.AsType[*ACMEError](err); ok {
		problemType = acmeErr.ProblemType()
		detail = acmeErr.Detail()
	}

	span.RecordError(err)
	span.SetStatus(codes.Error, errorCode)
	span.SetAttributes(
		attribute.String("provisioning.error_code", errorCode),
		attribute.String("acme.problem_type", problemType),
		attribute.String("acme.detail", detail),
	)
}

func (h *provisionHandler) loadSkipDNSChecks(ctx context.Context, hostname string) (bool, error) {
	var skip bool

	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error

			skip, err = h.skipsDNSChecks(ctx, conn, hostname)

			return err
		},
	)
	if err != nil {
		return false, err
	}

	return skip, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func (h *provisionHandler) checkDNSConfiguration(ctx context.Context, hostname string) error {
	customerFQDN := hostname
	if !strings.HasSuffix(customerFQDN, ".") {
		customerFQDN = customerFQDN + "."
	}

	expectedFQDN := h.cnameTarget
	if !strings.HasSuffix(expectedFQDN, ".") {
		expectedFQDN = expectedFQDN + "."
	}

	msg := &dns.Msg{MsgHeader: dns.MsgHeader{ID: dns.ID(), RecursionDesired: true}}
	msg.Question = []dns.RR{&dns.CNAME{Hdr: dns.Header{Name: customerFQDN, Class: dns.ClassINET}}}

	dnsCtx, cancel := context.WithTimeout(ctx, dnsExchangeTimeout)
	defer cancel()

	client := dns.NewClient()

	resp, _, err := client.Exchange(dnsCtx, msg, "udp", h.resolverAddr)
	if err != nil {
		return fmt.Errorf("cannot exchange dns message: %w", err)
	}

	if len(resp.Answer) == 0 {
		return fmt.Errorf("no cname records found for domain %q", hostname)
	}

	if len(resp.Answer) > 1 {
		return fmt.Errorf("multiple cname records found for domain %q", hostname)
	}

	resolvedRecord, ok := resp.Answer[0].(*dns.CNAME)
	if !ok {
		return fmt.Errorf("first answer is not a cname record for domain %q", hostname)
	}

	if !strings.EqualFold(expectedFQDN, resolvedRecord.Target) {
		return fmt.Errorf(
			"cname target mismatch: domain %q resolves to %q, expected %q",
			hostname,
			resolvedRecord.Target,
			expectedFQDN,
		)
	}

	return nil
}

func (h *provisionHandler) checkCAARecords(ctx context.Context, hostname string) error {
	fqdn := hostname
	if !strings.HasSuffix(fqdn, ".") {
		fqdn = fqdn + "."
	}

	msg := &dns.Msg{MsgHeader: dns.MsgHeader{ID: dns.ID(), RecursionDesired: true}}
	msg.Question = []dns.RR{&dns.CAA{Hdr: dns.Header{Name: fqdn, Class: dns.ClassINET}}}

	dnsCtx, cancel := context.WithTimeout(ctx, dnsExchangeTimeout)
	defer cancel()

	client := dns.NewClient()

	resp, _, err := client.Exchange(
		dnsCtx,
		msg,
		"udp",
		h.resolverAddr,
	)
	if err != nil {
		return fmt.Errorf("cannot exchange dns message for caa records: %w", err)
	}

	var caaRecords []*dns.CAA

	for _, rr := range resp.Answer {
		if caa, ok := rr.(*dns.CAA); ok {
			caaRecords = append(caaRecords, caa)
		}
	}

	if len(caaRecords) == 0 {
		return nil
	}

	for _, caa := range caaRecords {
		if caa.Tag == "issue" {
			issuer, _, _ := strings.Cut(caa.Value, ";")
			if strings.EqualFold(strings.TrimSpace(issuer), h.caaIssuerDomain) {
				return nil
			}
		}
	}

	return fmt.Errorf(
		"caa records for domain %q do not permit issuance by %q",
		hostname,
		h.caaIssuerDomain,
	)
}

func (h *provisionHandler) skipsDNSChecks(
	ctx context.Context,
	conn pg.Querier,
	hostname string,
) (bool, error) {
	if h.managedBaseDomain == "" {
		return false, nil
	}

	suffix := "." + h.managedBaseDomain
	if hostname != h.managedBaseDomain && !strings.HasSuffix(hostname, suffix) {
		return false, nil
	}

	domain := &coredata.CustomDomain{}

	err := domain.LoadByDomain(ctx, conn, coredata.NewNoScope(), hostname)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		return true, nil
	}

	if err != nil {
		return false, fmt.Errorf("cannot load custom domain: %w", err)
	}

	return domain.Managed, nil
}

func (h *provisionHandler) resetStaleCertificate(
	ctx context.Context,
	tx pg.Tx,
	certificate *coredata.Certificate,
) error {
	fullCertificate := &coredata.Certificate{}
	if err := fullCertificate.LoadByIDForUpdateSkipLocked(ctx, tx, coredata.NewNoScope(), certificate.ID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil
		}

		return fmt.Errorf("cannot load stale certificate for update: %w", err)
	}

	staleDuration := time.Since(fullCertificate.UpdatedAt)

	h.logger.InfoCtx(
		ctx,
		"resetting stale certificate",
		log.String("hostname", fullCertificate.Hostname),
		log.String("status", string(fullCertificate.Status)),
		log.Duration("stale_duration", staleDuration),
		log.Int("retry_count", fullCertificate.SSLRetryCount),
	)

	fullCertificate.HTTPChallengeToken = nil
	fullCertificate.HTTPChallengeKeyAuth = nil
	fullCertificate.HTTPChallengeURL = nil
	fullCertificate.HTTPOrderURL = nil
	fullCertificate.ProvisioningError = nil
	fullCertificate.Status = coredata.CertificateStatusPending

	if fullCertificate.SSLLastAttemptAt != nil && time.Since(*fullCertificate.SSLLastAttemptAt) > 24*time.Hour {
		h.logger.InfoCtx(
			ctx,
			"resetting retry count due to old last attempt",
			log.String("hostname", fullCertificate.Hostname),
			log.Time("last_attempt", *fullCertificate.SSLLastAttemptAt),
		)
		fullCertificate.SSLRetryCount = 0
		fullCertificate.SSLLastAttemptAt = nil
	}

	if err := fullCertificate.Update(ctx, tx, coredata.NewNoScope()); err != nil {
		return fmt.Errorf("cannot update stale certificate: %w", err)
	}

	return nil
}
