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
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"

	"codeberg.org/miekg/dns"
)

type (
	Provisioner struct {
		pg            *pg.Client
		acmeService   *ACMEService
		encryptionKey cipher.EncryptionKey
		cnameTarget   string
		interval      time.Duration
		resolverAddr  string
		logger        *log.Logger
	}
)

const (
	maxRetries = 3
)

func NewProvisioner(
	pg *pg.Client,
	acmeService *ACMEService,
	encryptionKey cipher.EncryptionKey,
	cnameTarget string,
	interval time.Duration,
	resolverAddr string,
	logger *log.Logger,
) *Provisioner {
	return &Provisioner{
		pg:            pg,
		acmeService:   acmeService,
		encryptionKey: encryptionKey,
		cnameTarget:   cnameTarget,
		interval:      interval,
		resolverAddr:  resolverAddr,
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
	customerFQDN := domain
	if !strings.HasSuffix(customerFQDN, ".") {
		customerFQDN = customerFQDN + "."
	}

	expectedFQDN := p.cnameTarget
	if !strings.HasSuffix(expectedFQDN, ".") {
		expectedFQDN = expectedFQDN + "."
	}

	msg := &dns.Msg{MsgHeader: dns.MsgHeader{ID: dns.ID(), RecursionDesired: true}}
	msg.Question = []dns.RR{&dns.CNAME{Hdr: dns.Header{Name: customerFQDN, Class: dns.ClassINET}}}

	client := dns.NewClient()

	resp, _, err := client.Exchange(context.Background(), msg, "udp", p.resolverAddr)
	if err != nil {
		return fmt.Errorf("cannot exchange dns message: %w", err)
	}

	if len(resp.Answer) == 0 {
		return fmt.Errorf("no cname records found for domain %q", domain)
	}

	if len(resp.Answer) > 1 {
		return fmt.Errorf("multiple cname records found for domain %q", domain)
	}

	resolvedRecord, ok := resp.Answer[0].(*dns.CNAME)
	if !ok {
		return fmt.Errorf("first answer is not a cname record for domain %q", domain)
	}

	if !strings.EqualFold(expectedFQDN, resolvedRecord.Target) {
		return fmt.Errorf(
			"cname target mismatch: domain %q resolves to %q, expected %q",
			domain,
			resolvedRecord.Target,
			expectedFQDN,
		)
	}

	return nil
}

func (p *Provisioner) checkPendingDomains(ctx context.Context) error {
	err := p.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := p.handleStaleProvisioningAttempts(ctx, tx); err != nil {
				return fmt.Errorf("cannot handle stale provisioning attempts: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot handle stale provisioning attempts: %w", err)
	}

	err = p.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			var domains coredata.CustomDomains
			if err := domains.ListDomainsWithPendingHTTPChallenges(ctx, tx, coredata.NewNoScope()); err != nil {
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

				if err := p.provisionDomainCertificate(ctx, tx, domain.ID); err != nil {
					p.logger.ErrorCtx(
						ctx,
						"cannot provision certificate",
						log.String("domain", domain.Domain),
						log.Error(err),
					)
				}
			}

			return nil
		},
	)

	if err != nil {
		return fmt.Errorf("cannot provision domains: %w", err)
	}

	return nil
}

func (p *Provisioner) handleStaleProvisioningAttempts(ctx context.Context, tx pg.Conn) error {
	var domains coredata.CustomDomains
	if err := domains.ListStaleProvisioningDomains(ctx, tx, coredata.NewNoScope()); err != nil {
		return fmt.Errorf("cannot load stale provisioning domains: %w", err)
	}

	if len(domains) == 0 {
		return nil
	}

	p.logger.InfoCtx(ctx, "found stale provisioning attempts to reset", log.Int("count", len(domains)))

	for _, domain := range domains {
		if err := p.resetStaleDomain(ctx, tx, domain); err != nil {
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
	tx pg.Conn,
	domain *coredata.CustomDomain,
) error {
	fullDomain := &coredata.CustomDomain{}
	if err := fullDomain.LoadByIDForUpdateSkipLocked(ctx, tx, coredata.NewNoScope(), p.encryptionKey, domain.ID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil
		}

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

	if err := fullDomain.Update(ctx, tx, coredata.NewNoScope(), p.encryptionKey); err != nil {
		return fmt.Errorf("cannot update stale domain: %w", err)
	}

	return nil
}

func (p *Provisioner) provisionDomainCertificate(
	ctx context.Context,
	tx pg.Conn,
	domainID gid.GID,
) error {
	domain := &coredata.CustomDomain{}
	if err := domain.LoadByIDForUpdateSkipLocked(ctx, tx, coredata.NewNoScope(), p.encryptionKey, domainID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil
		}

		return fmt.Errorf("cannot load by id for update %q custom domain: %w", domainID, err)
	}

	if domain.SSLStatus == coredata.CustomDomainSSLStatusPending || domain.SSLStatus == coredata.CustomDomainSSLStatusRenewing {
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

		domain.HTTPChallengeToken = &challenge.Token
		domain.HTTPChallengeKeyAuth = &challenge.KeyAuth
		domain.HTTPChallengeURL = &challenge.URL
		domain.HTTPOrderURL = &challenge.OrderURL
		domain.SSLStatus = coredata.CustomDomainSSLStatusProvisioning

		if err := domain.Update(ctx, tx, coredata.NewNoScope(), p.encryptionKey); err != nil {
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

		domain.SSLRetryCount = domain.SSLRetryCount + 1
		domain.SSLLastAttemptAt = ref.Ref(time.Now())

		if domain.SSLRetryCount >= maxRetries {
			p.logger.ErrorCtx(
				ctx,
				"domain has exceeded max retry attempts, marking as failed",
				log.String("domain", domain.Domain),
				log.Int("retry_count", domain.SSLRetryCount),
			)

			domain.SSLStatus = coredata.CustomDomainSSLStatusFailed
			domain.HTTPChallengeToken = nil
			domain.HTTPChallengeKeyAuth = nil
			domain.HTTPChallengeURL = nil
			domain.HTTPOrderURL = nil
		}

		if err := domain.Update(ctx, tx, coredata.NewNoScope(), p.encryptionKey); err != nil {
			return fmt.Errorf("cannot update domain: %w", err)
		}

		return nil
	}

	p.logger.InfoCtx(
		ctx,
		"certificate obtained successfully",
		log.String("domain", domain.Domain),
		log.Time("expires_at", cert.ExpiresAt),
	)

	domain.SSLCertificatePEM = cert.CertPEM
	if err := domain.EncryptPrivateKey(cert.KeyPEM, p.encryptionKey); err != nil {
		return fmt.Errorf("cannot encrypt private key: %w", err)
	}
	chainStr := string(cert.ChainPEM)
	domain.SSLCertificateChain = &chainStr
	domain.SSLExpiresAt = &cert.ExpiresAt
	domain.SSLStatus = coredata.CustomDomainSSLStatusActive

	domain.SSLRetryCount = 0
	domain.SSLLastAttemptAt = nil

	domain.HTTPChallengeToken = nil
	domain.HTTPChallengeKeyAuth = nil
	domain.HTTPChallengeURL = nil
	domain.HTTPOrderURL = nil

	if err := domain.Update(ctx, tx, coredata.NewNoScope(), p.encryptionKey); err != nil {
		return fmt.Errorf("cannot update domain: %w", err)
	}

	cache := &coredata.CachedCertificate{
		Domain:           domain.Domain,
		CertificatePEM:   string(cert.CertPEM),
		PrivateKeyPEM:    string(cert.KeyPEM),
		CertificateChain: &chainStr,
		ExpiresAt:        cert.ExpiresAt,
		CachedAt:         time.Now(),
		CustomDomainID:   domain.ID,
	}

	if err := cache.Upsert(ctx, tx); err != nil {
		p.logger.ErrorCtx(
			ctx,
			"cannot update certificate cache",
			log.String("domain", domain.Domain),
			log.Error(err),
		)
	}

	return nil
}
