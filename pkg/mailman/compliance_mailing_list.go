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

package mailman

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

func (s *Service) SubscriptionConfirmationEmailConfig(
	ctx context.Context,
	mailingListID gid.GID,
) (emails.PresenterConfig, string, *mail.Addr, error) {
	cfg, orgName, _, replyTo, err := s.mailingListEmailConfig(ctx, mailingListID)
	return cfg, orgName, replyTo, err
}

func (s *Service) UnsubscribeEmailConfig(
	ctx context.Context,
	mailingListID gid.GID,
) (emails.PresenterConfig, string, *mail.Addr, error) {
	cfg, orgName, _, replyTo, err := s.mailingListEmailConfig(ctx, mailingListID)
	return cfg, orgName, replyTo, err
}

func (s *Service) NewsEmailConfig(
	ctx context.Context,
	mailingListID gid.GID,
) (emails.PresenterConfig, string, string, *mail.Addr, error) {
	return s.mailingListEmailConfig(ctx, mailingListID)
}

func (s *Service) mailingListEmailConfig(
	ctx context.Context,
	mailingListID gid.GID,
) (emails.PresenterConfig, string, string, *mail.Addr, error) {
	var (
		mailingList    = &coredata.MailingList{}
		compliancePage = &coredata.TrustCenter{}
		organization   = &coredata.Organization{}
		customDomain   *coredata.CustomDomain
		logoFile       = &coredata.File{}
		defaultCfg     = emails.DefaultPresenterConfig(s.bucket, s.apiBaseURL)
	)

	scope := coredata.NewScopeFromObjectID(mailingListID)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := mailingList.LoadByID(ctx, conn, scope, mailingListID); err != nil {
				return fmt.Errorf("cannot load mailing list: %w", err)
			}
			if err := compliancePage.LoadByMailingListID(ctx, conn, scope, mailingListID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return err
				}
				return fmt.Errorf("cannot load compliance page: %w", err)
			}
			if compliancePage.LogoFileID != nil {
				if err := logoFile.LoadByID(ctx, conn, scope, *compliancePage.LogoFileID); err != nil {
					return fmt.Errorf("cannot load logo file: %w", err)
				}
			}
			if err := organization.LoadByID(ctx, conn, scope, compliancePage.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}
			customDomain = &coredata.CustomDomain{}
			if err := customDomain.LoadByOrganizationID(ctx, conn, scope, s.encryptionKey, organization.ID); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load custom domain: %w", err)
				}
			}
			return nil
		},
	)
	if err != nil {
		return defaultCfg, "", "", nil, err
	}

	cfg, baseURL, err := s.presenterConfigFromTrustCenter(compliancePage, organization, customDomain, logoFile)
	if err != nil {
		return defaultCfg, "", "", nil, err
	}

	return cfg, organization.Name, baseURL, mailingList.ReplyTo, nil
}

func (s *Service) presenterConfigFromTrustCenter(
	compliancePage *coredata.TrustCenter,
	organization *coredata.Organization,
	customDomain *coredata.CustomDomain,
	logoFile *coredata.File,
) (emails.PresenterConfig, string, error) {
	cfg := emails.DefaultPresenterConfig(s.bucket, s.apiBaseURL)

	parsedBaseURL, err := url.Parse(s.apiBaseURL)
	if err != nil {
		return cfg, "", fmt.Errorf("cannot parse base URL: %w", err)
	}

	baseURL := url.URL{
		Scheme: parsedBaseURL.Scheme,
		Host:   parsedBaseURL.Host,
		Path:   "/trust/" + compliancePage.Slug,
	}

	if customDomain != nil && customDomain.SSLStatus == coredata.CustomDomainSSLStatusActive {
		baseURL.Host = customDomain.Domain
		baseURL.Scheme = "https"
		baseURL.Path = ""
	}

	cfg.BaseURL = baseURL.String()

	if compliancePage.LogoFileID != nil && logoFile != nil && logoFile.FileKey != "" {
		cfg.SenderCompanyLogo = emails.Asset{
			Name:       logoFile.FileName,
			ObjectKey:  logoFile.FileKey,
			BucketName: logoFile.BucketName,
			MimeType:   logoFile.MimeType,
		}
		cfg.SenderCompanyName = organization.Name
		if organization.WebsiteURL != nil {
			cfg.SenderCompanyWebsiteURL = *organization.WebsiteURL
		}
		if organization.HeadquarterAddress != nil {
			cfg.SenderCompanyHeadquarterAddress = *organization.HeadquarterAddress
		}
	}

	return cfg, baseURL.String(), nil
}
