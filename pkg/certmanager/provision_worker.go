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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

const maxProvisioningRetries = 3

type provisionHandler struct {
	pg                *pg.Client
	acmeService       *ACMEService
	encryptionKey     cipher.EncryptionKey
	cnameTarget       string
	caaIssuerDomain   string
	resolverAddr      string
	managedBaseDomain string
	logger            *log.Logger
}

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
	}

	return worker.New(
		"certificate-provision-worker",
		h,
		logger,
		opts...,
	)
}

func (h *provisionHandler) Claim(ctx context.Context) (coredata.Certificate, error) {
	var certificate coredata.Certificate

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := certificate.LoadNextForProvisioningForUpdateSkipLocked(ctx, tx); err != nil {
				return err
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
	challengeInitiated, err := h.runProvisionCertificate(ctx, certificate.ID)
	if err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot provision certificate",
			log.String("hostname", certificate.Hostname),
			log.Error(err),
		)

		return err
	}

	if !challengeInitiated {
		return nil
	}

	if _, err := h.runProvisionCertificate(ctx, certificate.ID); err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot complete certificate challenge",
			log.String("hostname", certificate.Hostname),
			log.Error(err),
		)

		return err
	}

	return nil
}

func (h *provisionHandler) runProvisionCertificate(
	ctx context.Context,
	certificateID gid.GID,
) (bool, error) {
	var challengeInitiated bool

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var err error

			challengeInitiated, err = h.provisionCertificate(ctx, tx, certificateID)

			return err
		},
	)
	if err != nil {
		return false, err
	}

	return challengeInitiated, nil
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

func (h *provisionHandler) checkDNSConfiguration(hostname string) error {
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

	client := dns.NewClient()

	resp, _, err := client.Exchange(context.Background(), msg, "udp", h.resolverAddr)
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

func (h *provisionHandler) checkCAARecords(hostname string) error {
	fqdn := hostname
	if !strings.HasSuffix(fqdn, ".") {
		fqdn = fqdn + "."
	}

	msg := &dns.Msg{MsgHeader: dns.MsgHeader{ID: dns.ID(), RecursionDesired: true}}
	msg.Question = []dns.RR{&dns.CAA{Hdr: dns.Header{Name: fqdn, Class: dns.ClassINET}}}

	client := dns.NewClient()

	resp, _, err := client.Exchange(
		context.Background(),
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

func (h *provisionHandler) provisionCertificate(
	ctx context.Context,
	tx pg.Tx,
	certificateID gid.GID,
) (bool, error) {
	certificate := &coredata.Certificate{}
	if err := certificate.LoadByIDForUpdateSkipLocked(ctx, tx, coredata.NewNoScope(), certificateID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("cannot load by id for update %q certificate: %w", certificateID, err)
	}

	if certificate.Status == coredata.CertificateStatusPending || certificate.Status == coredata.CertificateStatusRenewing {
		skipDNSChecks, err := h.skipsDNSChecks(ctx, tx, certificate.Hostname)
		if err != nil {
			return false, fmt.Errorf("cannot check managed domain: %w", err)
		}

		if !skipDNSChecks {
			if err := h.checkDNSConfiguration(certificate.Hostname); err != nil {
				h.logger.WarnCtx(
					ctx,
					"dns configuration check failed",
					log.String("hostname", certificate.Hostname),
					log.Error(err),
				)

				errMsg := err.Error()

				certificate.ProvisioningError = &errMsg
				if err := certificate.Update(ctx, tx, coredata.NewNoScope()); err != nil {
					return false, fmt.Errorf("cannot update certificate with provisioning error: %w", err)
				}

				return false, nil
			}

			if err := h.checkCAARecords(certificate.Hostname); err != nil {
				h.logger.WarnCtx(
					ctx,
					"caa record check failed",
					log.String("hostname", certificate.Hostname),
					log.Error(err),
				)

				errMsg := err.Error()

				certificate.ProvisioningError = &errMsg
				if err := certificate.Update(ctx, tx, coredata.NewNoScope()); err != nil {
					return false, fmt.Errorf("cannot update certificate with provisioning error: %w", err)
				}

				return false, nil
			}
		}

		certificate.ProvisioningError = nil
		if err := certificate.Update(ctx, tx, coredata.NewNoScope()); err != nil {
			return false, fmt.Errorf("cannot clear provisioning error: %w", err)
		}

		h.logger.InfoCtx(ctx, "DNS configuration verified, initiating HTTP challenge for hostname", log.String("hostname", certificate.Hostname))

		challenge, err := h.acmeService.GetHTTPChallenge(ctx, certificate.Hostname)
		if err != nil {
			h.logger.ErrorCtx(
				ctx,
				"cannot get HTTP challenge",
				log.String("hostname", certificate.Hostname),
				log.Error(err),
			)

			return false, err
		}

		certificate.HTTPChallengeToken = &challenge.Token
		certificate.HTTPChallengeKeyAuth = &challenge.KeyAuth
		certificate.HTTPChallengeURL = &challenge.URL
		certificate.HTTPOrderURL = &challenge.OrderURL
		certificate.Status = coredata.CertificateStatusProvisioning

		if err := certificate.Update(ctx, tx, coredata.NewNoScope()); err != nil {
			return false, fmt.Errorf("cannot update certificate with challenge: %w", err)
		}

		h.logger.InfoCtx(
			ctx,
			"HTTP challenge initiated, completing in same cycle",
			log.String("hostname", certificate.Hostname),
			log.String("token", challenge.Token),
		)

		return true, nil
	}

	if certificate.HTTPChallengeToken == nil ||
		certificate.HTTPChallengeKeyAuth == nil ||
		certificate.HTTPChallengeURL == nil ||
		certificate.HTTPOrderURL == nil {
		certificate.Status = coredata.CertificateStatusPending

		if err := certificate.Update(ctx, tx, coredata.NewNoScope()); err != nil {
			return false, fmt.Errorf("cannot reset certificate without challenge data: %w", err)
		}

		return false, nil
	}

	challenge := &HTTPChallenge{
		Domain:   certificate.Hostname,
		Token:    *certificate.HTTPChallengeToken,
		KeyAuth:  *certificate.HTTPChallengeKeyAuth,
		URL:      *certificate.HTTPChallengeURL,
		OrderURL: *certificate.HTTPOrderURL,
	}

	cert, err := h.acmeService.CompleteHTTPChallenge(ctx, challenge)
	if err != nil {
		h.logger.WarnCtx(
			ctx,
			"cannot complete HTTP challenge",
			log.String("hostname", certificate.Hostname),
			log.Int("retry_count", certificate.SSLRetryCount),
			log.Error(err),
		)

		errMsg := err.Error()
		certificate.ProvisioningError = &errMsg
		certificate.SSLRetryCount = certificate.SSLRetryCount + 1
		now := time.Now()
		certificate.SSLLastAttemptAt = &now

		certificate.HTTPChallengeToken = nil
		certificate.HTTPChallengeKeyAuth = nil
		certificate.HTTPChallengeURL = nil
		certificate.HTTPOrderURL = nil

		if certificate.SSLRetryCount >= maxProvisioningRetries {
			h.logger.ErrorCtx(
				ctx,
				"certificate has exceeded max retry attempts, marking as failed",
				log.String("hostname", certificate.Hostname),
				log.Int("retry_count", certificate.SSLRetryCount),
			)

			certificate.Status = coredata.CertificateStatusFailed
		} else {
			certificate.Status = coredata.CertificateStatusPending
		}

		if err := certificate.Update(ctx, tx, coredata.NewNoScope()); err != nil {
			return false, fmt.Errorf("cannot update certificate: %w", err)
		}

		return false, nil
	}

	h.logger.InfoCtx(
		ctx,
		"certificate obtained successfully",
		log.String("hostname", certificate.Hostname),
		log.Time("expires_at", cert.ExpiresAt),
	)

	certificate.ProvisioningError = nil

	certificate.SSLCertificatePEM = cert.CertPEM
	if err := certificate.EncryptPrivateKey(cert.KeyPEM, h.encryptionKey); err != nil {
		return false, fmt.Errorf("cannot encrypt private key: %w", err)
	}

	chainStr := string(cert.ChainPEM)
	certificate.SSLCertificateChain = &chainStr
	certificate.SSLExpiresAt = &cert.ExpiresAt
	certificate.Status = coredata.CertificateStatusActive

	certificate.SSLRetryCount = 0
	certificate.SSLLastAttemptAt = nil

	certificate.HTTPChallengeToken = nil
	certificate.HTTPChallengeKeyAuth = nil
	certificate.HTTPChallengeURL = nil
	certificate.HTTPOrderURL = nil

	if err := certificate.Update(ctx, tx, coredata.NewNoScope()); err != nil {
		return false, fmt.Errorf("cannot update certificate: %w", err)
	}

	cache := &coredata.CachedCertificate{
		Domain:           certificate.Hostname,
		CertificatePEM:   string(cert.CertPEM),
		PrivateKeyPEM:    string(cert.KeyPEM),
		CertificateChain: &chainStr,
		ExpiresAt:        cert.ExpiresAt,
		CachedAt:         time.Now(),
		CertificateID:    certificate.ID,
	}

	if err := cache.Upsert(ctx, tx); err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot update certificate cache",
			log.String("hostname", certificate.Hostname),
			log.Error(err),
		)
	}

	return false, nil
}
