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
	"errors"
	"fmt"
	"net/url"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	CompliancePageService struct {
		*Service
	}
)

func NewCompliancePageService(svc *Service) *CompliancePageService {
	return &CompliancePageService{Service: svc}
}

func (s *CompliancePageService) GenerateLogoURL(
	ctx context.Context,
	compliancePageID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	file := &coredata.File{}
	compliancePage := &coredata.TrustCenter{}

	scope := coredata.NewScopeFromObjectID(compliancePageID)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := compliancePage.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			if compliancePage.LogoFileID == nil {
				return nil
			}

			if err := file.LoadByID(ctx, conn, scope, *compliancePage.LogoFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if compliancePage.LogoFileID == nil {
		return nil, nil
	}

	if file.FileKey == "" {
		return nil, nil
	}

	presignedURL, err := s.fm.GenerateFileUrl(ctx, file, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return &presignedURL, nil
}

func (s *CompliancePageService) EmailPresenterConfig(ctx context.Context, compliancePageID gid.GID) (emails.PresenterConfig, error) {
	var (
		compliancePage    = &coredata.TrustCenter{}
		organization      = &coredata.Organization{}
		customDomain      *coredata.CustomDomain
		logoFile          = &coredata.File{}
		emailPresenterCfg = emails.DefaultPresenterConfig(s.bucket, s.baseURL)
	)

	scope := coredata.NewScopeFromObjectID(compliancePageID)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := compliancePage.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			if compliancePage.LogoFileID != nil {
				if err := logoFile.LoadByID(ctx, conn, scope, *compliancePage.LogoFileID); err != nil {
					return fmt.Errorf("cannot load logoFile: %w", err)
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
		return emailPresenterCfg, err
	}

	parsedBaseURL, err := url.Parse(s.baseURL)
	if err != nil {
		return emailPresenterCfg, fmt.Errorf("cannot parse base URL: %w", err)
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

	emailPresenterCfg.BaseURL = baseURL.String()

	if compliancePage.LogoFileID != nil {
		if logoFile.FileKey == "" {
			return emailPresenterCfg, nil
		}

		// If logo exists, then we will brand the emails with the org as a sender

		emailPresenterCfg.SenderCompanyLogo = emails.Asset{
			Name:       logoFile.FileName,
			ObjectKey:  logoFile.FileKey,
			BucketName: logoFile.BucketName,
			MimeType:   logoFile.MimeType,
		}

		emailPresenterCfg.SenderCompanyName = organization.Name

		if organization.WebsiteURL != nil {
			emailPresenterCfg.SenderCompanyWebsiteURL = *organization.WebsiteURL
		}

		if organization.HeadquarterAddress != nil {
			emailPresenterCfg.SenderCompanyHeadquarterAddress = *organization.HeadquarterAddress
		}
	}

	return emailPresenterCfg, nil
}
