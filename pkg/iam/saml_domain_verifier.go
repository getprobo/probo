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

package iam

import (
	"context"
	"fmt"
	"strings"
	"time"

	"codeberg.org/miekg/dns"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	SAMLDomainVerifier struct {
		pg           *pg.Client
		interval     time.Duration
		resolverAddr string
		logger       *log.Logger
		tracer       trace.Tracer
	}
)

const (
	txtRecordValuePrefix = "probo-verification="
)

func NewSAMLDomainVerifier(
	pgClient *pg.Client,
	logger *log.Logger,
	tp trace.TracerProvider,
	interval time.Duration,
	resolverAddr string,
) *SAMLDomainVerifier {
	return &SAMLDomainVerifier{
		pg:           pgClient,
		interval:     interval,
		resolverAddr: resolverAddr,
		logger:       logger.Named("saml-domain-verifier"),
		tracer:       tp.Tracer("go.probo.inc/probo/pkg/iam/saml_domain_verifier"),
	}
}

func (v *SAMLDomainVerifier) Run(ctx context.Context) error {
	v.logger.InfoCtx(ctx, "starting", log.Duration("interval", v.interval))

	for {
		v.runOnce(ctx)

		select {
		case <-ctx.Done():
			v.logger.InfoCtx(ctx, "shutting down")
			return ctx.Err()
		case <-time.After(v.interval):
		}
	}
}

func (v *SAMLDomainVerifier) runOnce(ctx context.Context) {
	ctx, span := v.tracer.Start(ctx, "SAMLDomainVerifier.runOnce")
	defer span.End()

	if err := v.checkUnverifiedDomains(ctx); err != nil {
		v.logger.ErrorCtx(ctx, "cannot check unverified domains", log.Error(err))
	}
}

func (v *SAMLDomainVerifier) checkUnverifiedDomains(ctx context.Context) error {
	var configs coredata.SAMLConfigurations

	err := v.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := configs.LoadUnverified(ctx, conn)
			if err != nil {
				return fmt.Errorf("cannot load unverified SAML configurations: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	for _, config := range configs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := v.tryVerifyDomain(ctx, config.ID); err != nil {
			v.logger.ErrorCtx(ctx, "cannot verify domain",
				log.String("config_id", config.ID.String()),
				log.Error(err),
			)

			continue
		}
	}

	return nil
}

func (v *SAMLDomainVerifier) tryVerifyDomain(ctx context.Context, configID gid.GID) error {
	return v.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			config := &coredata.SAMLConfiguration{}
			if err := config.LoadByIDForUpdateSkipLocked(ctx, tx, configID); err != nil {
				if err == coredata.ErrResourceNotFound {
					return nil
				}

				return fmt.Errorf("cannot load SAML configuration: %w", err)
			}

			if config.DomainVerifiedAt != nil {
				return nil
			}

			if config.DomainVerificationToken == nil {
				return fmt.Errorf("cannot verify domain %q: no verification token", config.EmailDomain)
			}

			expectedValue := txtRecordValuePrefix + *config.DomainVerificationToken

			if err := v.checkDNSTXTRecord(config.EmailDomain, expectedValue); err != nil {
				return err
			}

			v.logger.InfoCtx(ctx, "domain verified",
				log.String("config_id", config.ID.String()),
			)

			now := time.Now()
			config.DomainVerificationToken = nil
			config.DomainVerifiedAt = &now
			config.EnforcementPolicy = coredata.SAMLEnforcementPolicyOptional
			config.UpdatedAt = now

			scope := coredata.NewScopeFromObjectID(config.OrganizationID)
			if err := config.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update SAML configuration: %w", err)
			}

			return nil
		},
	)
}

func (v *SAMLDomainVerifier) checkDNSTXTRecord(emailDomain string, expectedValue string) error {
	fqdn := emailDomain
	if !strings.HasSuffix(fqdn, ".") {
		fqdn = fqdn + "."
	}

	msg := &dns.Msg{MsgHeader: dns.MsgHeader{ID: dns.ID(), RecursionDesired: true}}
	msg.Question = []dns.RR{&dns.TXT{Hdr: dns.Header{Name: fqdn, Class: dns.ClassINET}}}

	client := dns.NewClient()
	resp, _, err := client.Exchange(context.Background(), msg, "udp", v.resolverAddr)
	if err != nil {
		return fmt.Errorf("cannot query TXT record for %q: %w", emailDomain, err)
	}

	if resp.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("cannot query TXT record for %q: %s", emailDomain, dns.RcodeToString[resp.Rcode])
	}

	if len(resp.Answer) == 0 {
		return fmt.Errorf("cannot find TXT record for %q", emailDomain)
	}

	for _, answer := range resp.Answer {
		txt, ok := answer.(*dns.TXT)
		if !ok {
			continue
		}

		value := strings.Join(txt.Txt, "")

		if value == expectedValue {
			return nil
		}
	}

	return fmt.Errorf("cannot find matching TXT record for %q", emailDomain)
}
