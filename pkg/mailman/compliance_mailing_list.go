// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package mailman

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/baseurl"
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

func (s *Service) UpdateEmailConfig(
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
		mailingList       = &coredata.MailingList{}
		compliancePage    = &coredata.CompliancePortal{}
		organization      = &coredata.Organization{}
		compliancePageURL string
		logoFile          = &coredata.File{}
		defaultCfg        = emails.DefaultPresenterConfig(s.apiBaseURL.String())
	)

	scope := coredata.NewScopeFromObjectID(mailingListID)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
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

			publicURL, err := s.compliancePortal.PublicURLForCompliancePortal(
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
		return defaultCfg, "", "", nil, err
	}

	cfg, err := s.presenterConfigFromCompliancePortal(compliancePage, organization, compliancePageURL, logoFile)
	if err != nil {
		return defaultCfg, "", "", nil, err
	}

	compliancePageBase, err := baseurl.Parse(compliancePageURL)
	if err != nil {
		return defaultCfg, "", "", nil, fmt.Errorf("cannot parse compliance page URL: %w", err)
	}

	updatesPageURL, err := compliancePageBase.AppendPath("/updates").String()
	if err != nil {
		return defaultCfg, "", "", nil, fmt.Errorf("cannot build updates page URL: %w", err)
	}

	return cfg, organization.Name, updatesPageURL, mailingList.ReplyTo, nil
}

func (s *Service) presenterConfigFromCompliancePortal(
	compliancePage *coredata.CompliancePortal,
	organization *coredata.Organization,
	compliancePageURL string,
	logoFile *coredata.File,
) (emails.PresenterConfig, error) {
	cfg := emails.DefaultPresenterConfig(s.apiBaseURL.String())
	cfg.BaseURL = compliancePageURL

	if compliancePage.LogoFileID != nil && logoFile != nil && logoFile.FileKey != "" {
		cfg.SenderCompanyLogoPath = filepath.Join("/api/files/v1/public/", logoFile.ID.String())

		cfg.SenderCompanyName = organization.Name
		if compliancePage.WebsiteURL != nil {
			cfg.SenderCompanyWebsiteURL = *compliancePage.WebsiteURL
		}

		if compliancePage.HeadquarterAddress != nil {
			cfg.SenderCompanyHeadquarterAddress = *compliancePage.HeadquarterAddress
		}
	}

	return cfg, nil
}
