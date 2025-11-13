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
	"fmt"
	"net"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
)

type (
	Provisioner struct {
		pg            *pg.Client
		acmeService   *ACMEService
		encryptionKey cipher.EncryptionKey
		cnameTarget   string
		interval      time.Duration
		logger        *log.Logger
	}
)

func NewProvisioner(
	pg *pg.Client,
	acmeService *ACMEService,
	encryptionKey cipher.EncryptionKey,
	cnameTarget string,
	interval time.Duration,
	logger *log.Logger,
) *Provisioner {
	return &Provisioner{
		pg:            pg,
		acmeService:   acmeService,
		encryptionKey: encryptionKey,
		cnameTarget:   cnameTarget,
		interval:      interval,
		logger:        logger.Named("certmanager.provisioner"),
	}
}

func (p *Provisioner) Run(ctx context.Context) error {
	p.logger.InfoCtx(ctx, "certificate provisioner starting", log.Duration("interval", p.interval))

	if err := p.checkPendingDomains(ctx); err != nil {
		p.logger.ErrorCtx(ctx, "initial check failed", log.Error(err))
	}

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.InfoCtx(ctx, "certificate provisioner shutting down")
			return ctx.Err()
		case <-ticker.C:
			if err := p.checkPendingDomains(ctx); err != nil {
				p.logger.ErrorCtx(ctx, "periodic check failed", log.Error(err))
			}
		}
	}
}

func (p *Provisioner) checkDNSConfiguration(domain string) error {
	cnameRecords, err := net.LookupCNAME(domain)
	if err != nil {
		return fmt.Errorf("cannot lookup cname for domain %q: %w", domain, err)
	}
	expectedTarget := strings.TrimSuffix(p.cnameTarget, ".")
	actualTarget := strings.TrimSuffix(cnameRecords, ".")

	if !strings.EqualFold(actualTarget, expectedTarget) {
		return fmt.Errorf(
			"cname target mismatch: domain %q resolves to %q, expected %q",
			domain,
			actualTarget,
			p.cnameTarget,
		)
	}

	return nil
}

func (p *Provisioner) checkPendingDomains(ctx context.Context) error {
	return p.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := p.handleStaleProvisioningAttempts(ctx, conn); err != nil {
				p.logger.ErrorCtx(ctx, "cannot handle stale provisioning attempts", log.Error(err))
			}

			var domains coredata.CustomDomains
			if err := domains.ListDomainsWithPendingHTTPChallenges(ctx, conn, coredata.NewNoScope()); err != nil {
				return fmt.Errorf("cannot load domains with pending challenges: %w", err)
			}

			if len(domains) == 0 {
				return nil
			}

			p.logger.InfoCtx(ctx, "found domains needing SSL provisioning", log.Int("count", len(domains)))

			for _, domain := range domains {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				if err := p.provisionDomainCertificate(ctx, conn, domain); err != nil {
					p.logger.ErrorCtx(
						ctx,
						"cannot provision certificate for domain",
						log.String("domain", domain.Domain),
						log.Error(err),
					)
				}
			}

			return nil
		},
	)
}

func isFatalChallengeError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	return (strings.Contains(errStr, "invalid") &&
		(strings.Contains(errStr, "challenge") ||
			strings.Contains(errStr, "authorization") ||
			strings.Contains(errStr, "order"))) ||
		strings.Contains(errStr, "authorization must be pending") ||
		strings.Contains(errStr, "expired")
}

func (p *Provisioner) handleStaleProvisioningAttempts(ctx context.Context, conn pg.Conn) error {
	var domains coredata.CustomDomains
	if err := domains.ListStaleProvisioningDomains(ctx, conn, coredata.NewNoScope()); err != nil {
		return fmt.Errorf("cannot load stale provisioning domains: %w", err)
	}

	if len(domains) == 0 {
		return nil
	}

	p.logger.InfoCtx(ctx, "found stale provisioning attempts to reset", log.Int("count", len(domains)))

	for _, domain := range domains {
		if err := p.resetStaleDomain(ctx, conn, domain); err != nil {
			p.logger.ErrorCtx(
				ctx,
				"cannot reset stale domain",
				log.String("domain", domain.Domain),
				log.Error(err),
			)
		}
	}

	return nil
}

func (p *Provisioner) resetStaleDomain(
	ctx context.Context,
	conn pg.Conn,
	domain *coredata.CustomDomain,
) error {
	fullDomain := &coredata.CustomDomain{}
	if err := fullDomain.LoadByIDForUpdate(ctx, conn, coredata.NewNoScope(), p.encryptionKey, domain.ID); err != nil {
		return fmt.Errorf("cannot load stale domain for update: %w", err)
	}

	staleDuration := time.Since(fullDomain.UpdatedAt)

	p.logger.InfoCtx(
		ctx,
		"resetting stale domain",
		log.String("domain", fullDomain.Domain),
		log.String("status", string(fullDomain.SSLStatus)),
		log.Duration("stale_duration", staleDuration),
		log.Int("retry_count", fullDomain.SSLRetryCount),
	)

	fullDomain.HTTPChallengeToken = nil
	fullDomain.HTTPChallengeKeyAuth = nil
	fullDomain.HTTPChallengeURL = nil
	fullDomain.HTTPOrderURL = nil
	fullDomain.SSLStatus = coredata.CustomDomainSSLStatusPending

	if fullDomain.SSLLastAttemptAt != nil && time.Since(*fullDomain.SSLLastAttemptAt) > 24*time.Hour {
		p.logger.InfoCtx(
			ctx,
			"resetting retry count due to old last attempt",
			log.String("domain", fullDomain.Domain),
			log.Time("last_attempt", *fullDomain.SSLLastAttemptAt),
		)
		fullDomain.SSLRetryCount = 0
		fullDomain.SSLLastAttemptAt = nil
	}

	if err := fullDomain.Update(ctx, conn, coredata.NewNoScope(), p.encryptionKey); err != nil {
		return fmt.Errorf("cannot update stale domain: %w", err)
	}

	return nil
}

func (p *Provisioner) provisionDomainCertificate(
	ctx context.Context,
	conn pg.Conn,
	domain *coredata.CustomDomain,
) error {
	if domain.SSLStatus == coredata.CustomDomainSSLStatusPending {
		if err := p.checkDNSConfiguration(domain.Domain); err != nil {
			p.logger.WarnCtx(
				ctx,
				"dns configuration check failed",
				log.String("domain", domain.Domain),
				log.Error(err),
			)

			return err
		}

		p.logger.InfoCtx(ctx, "DNS configuration verified, initiating HTTP challenge for domain", log.String("domain", domain.Domain))

		challenge, err := p.acmeService.GetHTTPChallenge(ctx, domain.Domain)
		if err != nil {
			p.logger.ErrorCtx(
				ctx,
				"cannot get HTTP challenge",
				log.String("domain", domain.Domain),
				log.Error(err),
			)
			return err
		}

		fullDomain := &coredata.CustomDomain{}
		if err := fullDomain.LoadByIDForUpdate(ctx, conn, coredata.NewNoScope(), p.encryptionKey, domain.ID); err != nil {
			return fmt.Errorf("cannot load domain for update: %w", err)
		}

		fullDomain.HTTPChallengeToken = &challenge.Token
		fullDomain.HTTPChallengeKeyAuth = &challenge.KeyAuth
		fullDomain.HTTPChallengeURL = &challenge.URL
		fullDomain.HTTPOrderURL = &challenge.OrderURL
		fullDomain.SSLStatus = coredata.CustomDomainSSLStatusProvisioning

		if err := fullDomain.Update(ctx, conn, coredata.NewNoScope(), p.encryptionKey); err != nil {
			return fmt.Errorf("cannot update domain with challenge: %w", err)
		}

		p.logger.InfoCtx(
			ctx,
			"HTTP challenge initiated, will complete in next cycle",
			log.String("domain", domain.Domain),
			log.String("token", challenge.Token),
		)

		return nil
	}

	challenge := &HTTPChallenge{
		Domain:   domain.Domain,
		Token:    *domain.HTTPChallengeToken,
		KeyAuth:  *domain.HTTPChallengeKeyAuth,
		URL:      *domain.HTTPChallengeURL,
		OrderURL: *domain.HTTPOrderURL,
	}

	cert, err := p.acmeService.CompleteHTTPChallenge(ctx, challenge)
	if err != nil {
		p.logger.WarnCtx(
			ctx,
			"cannot complete HTTP challenge",
			log.String("domain", domain.Domain),
			log.Int("retry_count", domain.SSLRetryCount),
			log.Error(err),
		)

		fullDomain := &coredata.CustomDomain{}
		if loadErr := fullDomain.LoadByIDForUpdate(ctx, conn, coredata.NewNoScope(), p.encryptionKey, domain.ID); loadErr != nil {
			p.logger.ErrorCtx(
				ctx,
				"cannot load domain for retry tracking",
				log.String("domain", domain.Domain),
				log.Error(loadErr),
			)
			return loadErr
		}

		fullDomain.SSLRetryCount++
		now := time.Now()
		fullDomain.SSLLastAttemptAt = &now

		const maxRetries = 3
		if fullDomain.SSLRetryCount >= maxRetries {
			p.logger.ErrorCtx(
				ctx,
				"domain has exceeded max retry attempts, marking as failed",
				log.String("domain", domain.Domain),
				log.Int("retry_count", fullDomain.SSLRetryCount),
			)

			fullDomain.SSLStatus = coredata.CustomDomainSSLStatusFailed
			fullDomain.HTTPChallengeToken = nil
			fullDomain.HTTPChallengeKeyAuth = nil
			fullDomain.HTTPChallengeURL = nil
			fullDomain.HTTPOrderURL = nil

			if updateErr := fullDomain.Update(ctx, conn, coredata.NewNoScope(), p.encryptionKey); updateErr != nil {
				p.logger.ErrorCtx(
					ctx,
					"cannot mark domain as failed",
					log.String("domain", domain.Domain),
					log.Error(updateErr),
				)
				return updateErr
			}

			return nil
		}

		if isFatalChallengeError(err) {
			p.logger.InfoCtx(
				ctx,
				"fatal challenge error, resetting domain to retry with fresh challenge",
				log.String("domain", domain.Domain),
				log.Int("retry_count", fullDomain.SSLRetryCount),
			)

			fullDomain.HTTPChallengeToken = nil
			fullDomain.HTTPChallengeKeyAuth = nil
			fullDomain.HTTPChallengeURL = nil
			fullDomain.HTTPOrderURL = nil
			fullDomain.SSLStatus = coredata.CustomDomainSSLStatusPending

			if updateErr := fullDomain.Update(ctx, conn, coredata.NewNoScope(), p.encryptionKey); updateErr != nil {
				p.logger.ErrorCtx(
					ctx,
					"cannot reset domain for retry",
					log.String("domain", domain.Domain),
					log.Error(updateErr),
				)
				return updateErr
			}

			return nil
		}

		p.logger.InfoCtx(
			ctx,
			"transient error, keeping existing challenge for retry",
			log.String("domain", domain.Domain),
			log.Int("retry_count", fullDomain.SSLRetryCount),
		)

		if updateErr := fullDomain.Update(ctx, conn, coredata.NewNoScope(), p.encryptionKey); updateErr != nil {
			p.logger.ErrorCtx(
				ctx,
				"cannot update domain retry tracking",
				log.String("domain", domain.Domain),
				log.Error(updateErr),
			)
			return updateErr
		}

		return nil
	}

	p.logger.InfoCtx(
		ctx,
		"certificate obtained successfully",
		log.String("domain", domain.Domain),
		log.Time("expires_at", cert.ExpiresAt),
	)

	fullDomain := &coredata.CustomDomain{}
	if err := fullDomain.LoadByID(ctx, conn, coredata.NewNoScope(), p.encryptionKey, domain.ID); err != nil {
		return fmt.Errorf("cannot load domain: %w", err)
	}

	fullDomain.SSLCertificatePEM = cert.CertPEM
	if err := fullDomain.EncryptPrivateKey(cert.KeyPEM, p.encryptionKey); err != nil {
		return fmt.Errorf("cannot encrypt private key: %w", err)
	}
	chainStr := string(cert.ChainPEM)
	fullDomain.SSLCertificateChain = &chainStr
	fullDomain.SSLExpiresAt = &cert.ExpiresAt
	fullDomain.SSLStatus = coredata.CustomDomainSSLStatusActive

	fullDomain.SSLRetryCount = 0
	fullDomain.SSLLastAttemptAt = nil

	fullDomain.HTTPChallengeToken = nil
	fullDomain.HTTPChallengeKeyAuth = nil
	fullDomain.HTTPChallengeURL = nil
	fullDomain.HTTPOrderURL = nil

	if err := fullDomain.Update(ctx, conn, coredata.NewNoScope(), p.encryptionKey); err != nil {
		return fmt.Errorf("cannot update domain: %w", err)
	}

	cache := &coredata.CachedCertificate{
		Domain:           fullDomain.Domain,
		CertificatePEM:   string(cert.CertPEM),
		PrivateKeyPEM:    string(cert.KeyPEM),
		CertificateChain: &chainStr,
		ExpiresAt:        cert.ExpiresAt,
		CachedAt:         time.Now(),
		CustomDomainID:   fullDomain.ID,
	}

	if err := cache.Upsert(ctx, conn); err != nil {
		p.logger.ErrorCtx(
			ctx,
			"cannot update certificate cache",
			log.String("domain", fullDomain.Domain),
			log.Error(err),
		)
	}

	return nil
}
