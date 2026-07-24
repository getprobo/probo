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

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/dnsclient"
	"go.probo.inc/probo/pkg/gid"
	"golang.org/x/crypto/acme"
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
	persistFailureTimeout  = 15 * time.Second

	// provisioningPollLease is how long a claimed PROVISIONING row with an open
	// order stays ineligible for re-claim after each attempt. It must exceed
	// processTickTimeout: the claim's row lock is released before Process runs,
	// so a lease shorter than the maximum processing window would let another
	// worker claim and process the same row concurrently.
	provisioningPollLease = processTickTimeout + 30*time.Second

	tracerName = "go.probo.inc/probo/pkg/certmanager"
)

type (
	provisionCore struct {
		pg          *pg.Client
		acmeService *ACMEService
		logger      *log.Logger
		tracer      trace.Tracer
	}

	beginChallengeHandler struct {
		provisionCore

		cnameTarget       string
		caaIssuerDomain   string
		dnsClient         *dnsclient.Client
		managedBaseDomain string
	}

	pollOrderHandler struct {
		provisionCore

		encryptionKey cipher.EncryptionKey
	}

	provisioningOutcome struct {
		status         coredata.CertificateStatus
		retryCount     int
		clearACMEState bool
		errorCode      string
	}
)

var (
	_ worker.Handler[coredata.Certificate] = (*beginChallengeHandler)(nil)
	_ worker.Handler[coredata.Certificate] = (*pollOrderHandler)(nil)
	_ worker.StaleRecoverer                = (*pollOrderHandler)(nil)
)

func NewBeginChallengeWorker(
	pgClient *pg.Client,
	acmeService *ACMEService,
	cnameTarget string,
	caaIssuerDomain string,
	resolverAddr string,
	managedBaseDomain string,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.Certificate] {
	h := &beginChallengeHandler{
		provisionCore: provisionCore{
			pg:          pgClient,
			acmeService: acmeService,
			logger:      logger,
			tracer:      otel.Tracer(tracerName),
		},
		cnameTarget:       cnameTarget,
		caaIssuerDomain:   caaIssuerDomain,
		dnsClient:         dnsclient.NewClient(resolverAddr),
		managedBaseDomain: managedBaseDomain,
	}

	opts = append(opts, worker.WithMaxConcurrency(1))

	return worker.New(
		"certificate-begin-challenge-worker",
		h,
		logger,
		opts...,
	)
}

func NewPollOrderWorker(
	pgClient *pg.Client,
	acmeService *ACMEService,
	encryptionKey cipher.EncryptionKey,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.Certificate] {
	h := &pollOrderHandler{
		provisionCore: provisionCore{
			pg:          pgClient,
			acmeService: acmeService,
			logger:      logger,
			tracer:      otel.Tracer(tracerName),
		},
		encryptionKey: encryptionKey,
	}

	opts = append(opts, worker.WithMaxConcurrency(1))

	return worker.New(
		"certificate-poll-worker",
		h,
		logger,
		opts...,
	)
}

func (h *beginChallengeHandler) Claim(ctx context.Context) (coredata.Certificate, error) {
	if h.acmeService.InCooldown() {
		return coredata.Certificate{}, worker.ErrNoTask
	}

	var certificate coredata.Certificate

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := certificate.LoadNextForBeginChallengeForUpdateSkipLocked(ctx, tx); err != nil {
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

func (h *beginChallengeHandler) Process(ctx context.Context, certificate coredata.Certificate) error {
	ctx, cancel := context.WithTimeout(ctx, processTickTimeout)
	defer cancel()

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

		cnameCtx, cnameCancel := context.WithTimeout(dnsCtx, dnsExchangeTimeout)
		err := h.dnsClient.CheckCNAME(cnameCtx, certificate.Hostname, h.cnameTarget)
		cnameCancel()
		if err != nil {
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

	persisted, err := h.persistChallenge(ctx, certificate, challenge)
	if err != nil {
		h.recordSpanError(span, err, classifyProvisioningError(err))
		return err
	}

	if !persisted {
		return nil
	}

	// The key authorization is now committed and served by the challenge
	// handler, so it is safe to ask the CA to begin validation. Accepting
	// earlier risks the CA hitting the token before this instance can serve it,
	// yielding a 404 and an invalid order.
	if err := h.acmeService.AcceptHTTPChallenge(ctx, challenge); err != nil {
		errorCode := classifyProvisioningError(err)
		h.logACMEOutcome(ctx, certificate, provisionPhaseCreateOrder, err, errorCode)
		h.recordSpanError(span, err, errorCode)

		return h.persistFailure(ctx, certificate.ID, err)
	}

	h.logger.InfoCtx(
		ctx,
		"HTTP challenge accepted, waiting for order validation",
		log.String("hostname", certificate.Hostname),
		log.String("certificate_id", certificate.ID.String()),
	)

	return nil
}

// persistChallenge stores the HTTP-01 challenge metadata and flips the row to
// PROVISIONING. It takes a blocking write-back lock so the metadata is persisted
// once any competing transaction commits; a genuinely deleted row is the only
// no-op. It reports whether the row was persisted (false when the row is gone or
// has moved on to another status).
func (h *beginChallengeHandler) persistChallenge(
	ctx context.Context,
	certificate coredata.Certificate,
	challenge *HTTPChallenge,
) (bool, error) {
	persisted := false

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			persisted = false

			row := &coredata.Certificate{}
			if err := row.LoadByIDForUpdate(ctx, tx, coredata.NewNoScope(), certificate.ID); err != nil {
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

			persisted = true

			return nil
		},
	)
	if err != nil {
		return false, err
	}

	return persisted, nil
}

func (h *beginChallengeHandler) loadSkipDNSChecks(ctx context.Context, hostname string) (bool, error) {
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

func (h *beginChallengeHandler) checkCAARecords(ctx context.Context, hostname string) error {
	err := h.dnsClient.CheckCAA(ctx, hostname, h.caaIssuerDomain)
	if err == nil {
		return nil
	}

	if errors.Is(err, dnsclient.ErrCAADenied) {
		return fmt.Errorf("%w: domain %q by %q", ErrCAANotPermitted, hostname, h.caaIssuerDomain)
	}

	return err
}

func (h *beginChallengeHandler) skipsDNSChecks(
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

func (h *pollOrderHandler) Claim(ctx context.Context) (coredata.Certificate, error) {
	// Deliberately NOT gated by acmeService.InCooldown(). A cooldown is entered
	// when minting a NEW order hits the CA's rate limit (see
	// beginChallengeHandler.Claim); advancing an order already in flight only
	// polls/finalizes that specific order and does not mint new ones. Gating
	// this claim too would stall every other tenant's in-flight provisioning
	// for up to the cooldown duration just because one unrelated hostname
	// tripped a limit.
	var certificate coredata.Certificate

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := certificate.LoadNextForPollOrderForUpdateSkipLocked(ctx, tx, provisioningPollLease); err != nil {
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

func (h *pollOrderHandler) Process(ctx context.Context, certificate coredata.Certificate) error {
	ctx, cancel := context.WithTimeout(ctx, processTickTimeout)
	defer cancel()

	return h.processPollOrder(ctx, certificate)
}

func (h *pollOrderHandler) processPollOrder(
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
		// Re-accept best-effort: if a prior tick committed the challenge but its
		// Accept never reached the CA, the order would otherwise stay pending
		// forever since this path only polls. Only do this while the order is
		// still PENDING: once it has moved to PROCESSING, the CA has already
		// registered the Accept and rejects a second one with
		// malformed/"Only pending challenges may be validated" (RFC 8555), so
		// re-accepting there is not a no-op and just produces noisy failures.
		if poll.Order.Status == acme.StatusPending {
			if err := h.acmeService.AcceptHTTPChallenge(ctx, challenge); err != nil {
				h.logger.WarnCtx(
					ctx,
					"re-accepting HTTP challenge for not-ready order failed, will poll again",
					log.String("hostname", certificate.Hostname),
					log.String("certificate_id", certificate.ID.String()),
					log.Error(err),
				)
			}
		}

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
	case OrderPollStatusReady:
		return h.issueCertificate(ctx, certificate, challenge, poll)
	case OrderPollStatusValid:
		// A VALID order observed while polling means a previous attempt already
		// finalized it at the CA but the certificate/private-key write never
		// landed (crash or a failed transaction). The private key generated for
		// that finalize is gone, so fetching the issued certificate now would
		// pair it with a freshly generated key and break TLS loading. Abandon
		// the unrecoverable order and start a new one.
		return h.abandonRecoveredValidOrder(ctx, span, certificate)
	default:
		return nil
	}
}

// abandonRecoveredValidOrder clears the ACME order state and returns the row to
// PENDING so the next tick mints a fresh order (and a matching private key).
func (h *pollOrderHandler) abandonRecoveredValidOrder(
	ctx context.Context,
	span trace.Span,
	certificate coredata.Certificate,
) error {
	h.logger.WarnCtx(
		ctx,
		"abandoning recovered valid ACME order without a matching private key, restarting provisioning",
		log.String("hostname", certificate.Hostname),
		log.String("certificate_id", certificate.ID.String()),
	)
	span.AddEvent("abandoned recovered valid order without matching private key")

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			row := &coredata.Certificate{}
			if err := row.LoadByIDForUpdate(ctx, tx, coredata.NewNoScope(), certificate.ID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load certificate %q: %w", certificate.ID, err)
			}

			if row.Status != coredata.CertificateStatusProvisioning {
				return nil
			}

			row.HTTPChallengeToken = nil
			row.HTTPChallengeKeyAuth = nil
			row.HTTPChallengeURL = nil
			row.HTTPOrderURL = nil
			row.ProvisioningError = nil
			row.Status = coredata.CertificateStatusPending

			if err := row.Update(ctx, tx, coredata.NewNoScope()); err != nil {
				return fmt.Errorf("cannot reset certificate after abandoning valid order: %w", err)
			}

			return nil
		},
	)
}

func (h *pollOrderHandler) issueCertificate(
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

func (h *provisionCore) persistFailure(
	ctx context.Context,
	certificateID gid.GID,
	provisionErr error,
) error {
	errorCode := classifyProvisioningError(provisionErr)

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), persistFailureTimeout)
	defer cancel()

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
			row.Status = outcome.status
			row.SSLRetryCount = outcome.retryCount

			if isImmediateRetryRateLimit(provisionErr) {
				// Retry-After: 0 permits an immediate retry; clearing the
				// timestamp exempts the row from the claim query's backoff gates
				// that would otherwise hold it despite the zero cooldown.
				row.SSLLastAttemptAt = nil
			} else {
				now := time.Now()
				row.SSLLastAttemptAt = &now
			}

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

// isImmediateRetryRateLimit reports whether err is an ACME rate limit with an
// explicit Retry-After of zero, i.e. the CA permits an immediate retry.
func isImmediateRetryRateLimit(err error) bool {
	acmeErr, ok := errors.AsType[*ACMEError](err)
	if !ok {
		return false
	}

	return errors.Is(acmeErr, ErrACMERateLimited) && acmeErr.RetryAfter() == 0
}

// decideProvisioningOutcome computes status/retry/clear policy for a failed step.
//
// Two backoff regimes:
//   - Normal transient errors: ssl_retry_count increments; at maxProvisioningRetries
//     the domain becomes FAILED.
//   - Rate limits: never increment ssl_retry_count and never mark FAILED; the
//     in-process ACME cooldown gates the begin-challenge worker's Claim. When an
//     order URL is present it is preserved so the next tick resumes polling
//     instead of minting a new order.
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

func (h *pollOrderHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var certificates coredata.Certificates
			if err := certificates.ListStaleProvisioning(ctx, tx, coredata.NewNoScope(), ProvisioningErrorACMERateLimited); err != nil {
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

func (h *provisionCore) logACMEOutcome(
	ctx context.Context,
	certificate coredata.Certificate,
	phase provisionPhase,
	err error,
	errorCode string,
) {
	var (
		statusCode  int
		problemType string
		detail      string
		instance    string
		link        string
		subproblems string
		retryAfter  time.Duration
	)
	if acmeErr, ok := errors.AsType[*ACMEError](err); ok {
		statusCode = acmeErr.StatusCode()
		problemType = acmeErr.ProblemType()
		detail = acmeErr.Detail()
		instance = acmeErr.Instance()
		link = acmeErr.Link()

		subproblems = acmeErr.Subproblems()
		if errors.Is(acmeErr, ErrACMERateLimited) {
			retryAfter = acmeErr.RetryAfter()
		}
	}

	level := h.logger.WarnCtx

	if errorCode == ProvisioningErrorACMETemporary {
		level = h.logger.ErrorCtx
	}

	level(
		ctx,
		"certificate provisioning step failed",
		log.String("acme_hostname", certificate.Hostname),
		log.String("acme_certificate_id", certificate.ID.String()),
		log.String("acme_phase", string(phase)),
		log.String("acme_error_code", errorCode),
		log.Int("acme_status_code", statusCode),
		log.String("acme_problem_type", problemType),
		log.String("acme_detail", detail),
		log.String("acme_instance", instance),
		log.Duration("acme_retry_after", retryAfter),
		log.String("acme_link", link),
		log.String("acme_subproblems", subproblems),
		log.Int("acme_retry_count", certificate.SSLRetryCount),
		log.Time("acme_cooldown_until", h.acmeService.CooldownUntil()),
		log.Error(err),
	)
}

func (h *provisionCore) setCertificateSpanAttributes(span trace.Span, certificate coredata.Certificate) {
	span.SetAttributes(
		attribute.String("certificate.id", certificate.ID.String()),
		attribute.String("certificate.hostname", certificate.Hostname),
		attribute.String("certificate.status", string(certificate.Status)),
	)
}

func (h *provisionCore) recordSpanError(span trace.Span, err error, errorCode string) {
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

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func (h *pollOrderHandler) resetStaleCertificate(
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
