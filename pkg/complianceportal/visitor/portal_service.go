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

package visitor

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func (s *Service) GetPortal(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
) (*coredata.TrustCenter, error) {
	var compliancePage *coredata.TrustCenter

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			compliancePage = &coredata.TrustCenter{}
			if err := compliancePage.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load compliance page: %w", err)
	}

	return compliancePage, nil
}

func (s *Service) GetPortalNDAFile(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
) (*coredata.File, error) {
	var file *coredata.File

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			compliancePage := &coredata.TrustCenter{}
			if err := compliancePage.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			if compliancePage.NonDisclosureAgreementFileID == nil {
				return nil
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, scope, *compliancePage.NonDisclosureAgreementFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *Service) GeneratePortalNDAFileURL(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	expiresIn time.Duration,
) (string, error) {
	var file *coredata.File

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			compliancePage := &coredata.TrustCenter{}
			if err := compliancePage.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			if compliancePage.NonDisclosureAgreementFileID == nil {
				return fmt.Errorf("no NDA file found")
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, scope, *compliancePage.NonDisclosureAgreementFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	presignedURL, err := s.fileManager.GeneratePresignedURL(ctx, file, expiresIn)
	if err != nil {
		return "", fmt.Errorf("cannot generate file URL: %w", err)
	}

	return presignedURL, nil
}

func (s *Service) GetPortalEmailPresenterConfig(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
) (emails.PresenterConfig, error) {
	var (
		compliancePage    = &coredata.TrustCenter{}
		organization      = &coredata.Organization{}
		logoFile          = &coredata.File{}
		compliancePageURL string
		emailPresenterCfg = emails.DefaultPresenterConfig(s.baseURL)
	)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
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

			publicURL, err := s.management.PublicURLForCompliancePage(
				ctx,
				conn,
				scope,
				compliancePage,
			)
			if err != nil {
				return fmt.Errorf("cannot resolve compliance page URL: %w", err)
			}

			compliancePageURL = publicURL

			return nil
		},
	)
	if err != nil {
		return emailPresenterCfg, err
	}

	emailPresenterCfg.BaseURL = compliancePageURL

	if compliancePage.LogoFileID != nil {
		if logoFile.FileKey == "" {
			return emailPresenterCfg, nil
		}

		// If logo exists, then we will brand the emails with the org as a sender
		emailPresenterCfg.SenderCompanyLogoPath = filepath.Join("/api/files/v1/public/", logoFile.ID.String())
		emailPresenterCfg.SenderCompanyName = organization.Name

		if compliancePage.WebsiteURL != nil {
			emailPresenterCfg.SenderCompanyWebsiteURL = *compliancePage.WebsiteURL
		}

		if compliancePage.HeadquarterAddress != nil {
			emailPresenterCfg.SenderCompanyHeadquarterAddress = *compliancePage.HeadquarterAddress
		}
	}

	return emailPresenterCfg, nil
}
