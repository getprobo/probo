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

package trust

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type TrustCenterService struct {
	svc *Service
}

func (s TrustCenterService) Get(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
) (*coredata.TrustCenter, error) {
	var trustCenter *coredata.TrustCenter

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, scope, trustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load trust center: %w", err)
	}

	return trustCenter, nil
}

func (s TrustCenterService) GetByOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*coredata.TrustCenter, error) {
	trustCenter := &coredata.TrustCenter{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := trustCenter.LoadByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return trustCenter, nil
}

func (s TrustCenterService) GetNDAFile(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
) (*coredata.File, error) {
	var file *coredata.File

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, scope, trustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if trustCenter.NonDisclosureAgreementFileID == nil {
				return nil
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, scope, *trustCenter.NonDisclosureAgreementFileID); err != nil {
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

func (s TrustCenterService) GenerateNDAFileURL(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
	expiresIn time.Duration,
) (string, error) {
	var file *coredata.File

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, scope, trustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if trustCenter.NonDisclosureAgreementFileID == nil {
				return fmt.Errorf("no NDA file found")
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, scope, *trustCenter.NonDisclosureAgreementFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	presignedURL, err := s.svc.fileManager.GeneratePresignedURL(ctx, file, expiresIn)
	if err != nil {
		return "", fmt.Errorf("cannot generate file URL: %w", err)
	}

	return presignedURL, nil
}

func (s TrustCenterService) GenerateLogoURL(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	file := &coredata.File{}
	compliancePage := &coredata.TrustCenter{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
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

	presignedURL, err := s.svc.fileManager.GeneratePresignedURL(ctx, file, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return &presignedURL, nil
}

func (s TrustCenterService) GenerateDarkLogoURL(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	file := &coredata.File{}
	compliancePage := &coredata.TrustCenter{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := compliancePage.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			if compliancePage.DarkLogoFileID == nil {
				return nil
			}

			if err := file.LoadByID(ctx, conn, scope, *compliancePage.DarkLogoFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if compliancePage.DarkLogoFileID == nil {
		return nil, nil
	}

	if file.FileKey == "" {
		return nil, nil
	}

	presignedURL, err := s.svc.fileManager.GeneratePresignedURL(ctx, file, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return &presignedURL, nil
}

func (s *TrustCenterService) EmailPresenterConfig(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
) (emails.PresenterConfig, error) {
	var (
		compliancePage    = &coredata.TrustCenter{}
		organization      = &coredata.Organization{}
		customDomain      *coredata.CustomDomain
		logoFile          = &coredata.File{}
		emailPresenterCfg = emails.DefaultPresenterConfig(s.svc.baseURL)
	)

	err := s.svc.pg.WithConn(
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

			customDomain = &coredata.CustomDomain{}
			if err := customDomain.LoadByOrganizationID(ctx, conn, scope, organization.ID); err != nil {
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

	parsedBaseURL, err := url.Parse(s.svc.baseURL)
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
		emailPresenterCfg.SenderCompanyLogoPath = filepath.Join("/api/files/v1/public/", logoFile.ID.String())
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

func (s *TrustCenterService) GetMailingList(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
) (*coredata.MailingList, error) {
	var mailingList *coredata.MailingList

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, scope, trustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if trustCenter.MailingListID == nil {
				return nil
			}

			mailingList = &coredata.MailingList{}
			if err := mailingList.LoadByID(ctx, conn, scope, *trustCenter.MailingListID); err != nil {
				return fmt.Errorf("cannot load mailing list: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return mailingList, nil
}
