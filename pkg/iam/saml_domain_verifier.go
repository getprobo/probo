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

package iam

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/dnsclient"
	"go.probo.inc/probo/pkg/gid"
)

type (
	SAMLDomainVerifier struct {
		pg        *pg.Client
		interval  time.Duration
		dnsClient *dnsclient.Client
		logger    *log.Logger
		tracer    trace.Tracer
	}
)

const (
	txtRecordValuePrefix = "probo-verification="
)

var (
	errDomainTXTRecordNotFound = errors.New("domain TXT record not found")
	errDomainTXTRecordMismatch = errors.New("domain TXT record mismatch")
)

func NewSAMLDomainVerifier(
	pgClient *pg.Client,
	logger *log.Logger,
	tp trace.TracerProvider,
	interval time.Duration,
	resolverAddr string,
) *SAMLDomainVerifier {
	return &SAMLDomainVerifier{
		pg:        pgClient,
		interval:  interval,
		dnsClient: dnsclient.NewClient(resolverAddr),
		logger:    logger.Named("saml-domain-verifier"),
		tracer:    tp.Tracer("go.probo.inc/probo/pkg/iam/saml_domain_verifier"),
	}
}

func (v *SAMLDomainVerifier) Run(ctx context.Context) error {
	v.logger.InfoCtx(ctx, "starting", log.Duration("interval", v.interval))

	v.runOnce(ctx)

	if v.interval <= 0 {
		return fmt.Errorf("cannot run SAML domain verifier: interval must be greater than zero")
	}

	ticker := time.NewTicker(v.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			v.logger.InfoCtx(ctx, "shutting down")
			return ctx.Err()
		case <-ticker.C:
			v.runOnce(ctx)
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
		func(ctx context.Context, conn pg.Querier) error {
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
			if errors.Is(err, errDomainTXTRecordNotFound) || errors.Is(err, errDomainTXTRecordMismatch) {
				v.logger.InfoCtx(
					ctx,
					"domain verification pending",
					log.String("config_id", config.ID.String()),
					log.Error(err),
				)
			} else {
				v.logger.ErrorCtx(
					ctx,
					"cannot verify domain",
					log.String("config_id", config.ID.String()),
					log.Error(err),
				)
			}

			continue
		}
	}

	return nil
}

func (v *SAMLDomainVerifier) tryVerifyDomain(ctx context.Context, configID gid.GID) error {
	return v.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
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

			if err := v.checkDNSTXTRecord(ctx, config.EmailDomain, expectedValue); err != nil {
				return err
			}

			v.logger.InfoCtx(
				ctx,
				"domain verified",
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

func (v *SAMLDomainVerifier) checkDNSTXTRecord(ctx context.Context, emailDomain string, expectedValue string) error {
	err := v.dnsClient.CheckTXT(ctx, emailDomain, expectedValue)
	if err == nil {
		return nil
	}

	if errors.Is(err, dnsclient.ErrTXTNotFound) {
		return fmt.Errorf("%w for %q", errDomainTXTRecordNotFound, emailDomain)
	}

	if errors.Is(err, dnsclient.ErrTXTMismatch) {
		return fmt.Errorf("%w for %q", errDomainTXTRecordMismatch, emailDomain)
	}

	return err
}
