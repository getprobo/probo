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

package mcp_v1

import (
	"context"
	"errors"
	"fmt"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/complianceportal/management"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/probot"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

func (r *Resolver) loadCompliancePortalDetails(
	ctx context.Context,
	scope *coredata.Scope,
	portal *coredata.CompliancePortal,
	out *types.CompliancePortal,
) error {
	publicURL, err := r.management.PublicURL(ctx, scope, portal.ID)
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot resolve compliance portal public url", log.Error(err))

		return fmt.Errorf("internal error")
	}

	out.PublicURL = &publicURL

	customDomain, err := r.loadCompliancePortalCustomDomain(ctx, portal)
	if err != nil {
		return err
	}

	out.CustomDomain = customDomain

	slack, err := r.loadCompliancePortalSlack(ctx, portal)
	if err != nil {
		return err
	}

	out.Slack = slack

	return nil
}

func (r *Resolver) loadCompliancePortalCustomDomain(
	ctx context.Context,
	portal *coredata.CompliancePortal,
) (*types.CustomDomain, error) {
	if portal.CustomDomainID == nil {
		return nil, nil
	}

	scope, err := r.Authorize(ctx, portal.ID, management.ActionCustomDomainGet)
	if err != nil {
		if authorizationDenied(err) {
			return nil, nil
		}

		return nil, err
	}

	domain, err := r.management.GetDomain(ctx, scope, *portal.CustomDomainID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		r.logger.ErrorCtx(ctx, "cannot load custom domain", log.Error(err))

		return nil, fmt.Errorf("internal error")
	}

	var cert *coredata.Certificate
	if domain.CertificateID != nil {
		cert, err = r.certManager.Get(ctx, scope, *domain.CertificateID)
		if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
			r.logger.ErrorCtx(ctx, "cannot load custom domain certificate", log.Error(err))

			return nil, fmt.Errorf("internal error")
		}
	}

	return types.NewCustomDomain(domain, cert, r.customDomainCname), nil
}

func (r *Resolver) loadCompliancePortalSlack(
	ctx context.Context,
	portal *coredata.CompliancePortal,
) (*types.CompliancePortalSlack, error) {
	scope, err := r.Authorize(ctx, portal.ID, probo.ActionConnectorInitiate)
	if err != nil {
		if authorizationDenied(err) {
			return nil, nil
		}

		return nil, err
	}

	installed, err := r.slackAppInstalled(ctx, scope, portal)
	if err != nil {
		return nil, err
	}

	channel, err := r.slackNotificationChannel(ctx, scope, portal)
	if err != nil {
		return nil, err
	}

	return &types.CompliancePortalSlack{
		Installed: installed,
		Channel:   channel,
	}, nil
}

func (r *Resolver) slackAppInstalled(
	ctx context.Context,
	scope *coredata.Scope,
	portal *coredata.CompliancePortal,
) (bool, error) {
	if r.slackbotInstallations == nil {
		return false, nil
	}

	installation, err := r.slackbotInstallations.GetByOrganizationID(
		ctx,
		scope,
		portal.OrganizationID,
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return false, nil
		}

		r.logger.ErrorCtx(ctx, "cannot load Slack installation", log.Error(err))

		return false, fmt.Errorf("internal error")
	}

	return installation.Status == coredata.SlackbotInstallationStatusActive, nil
}

func (r *Resolver) slackNotificationChannel(
	ctx context.Context,
	scope *coredata.Scope,
	portal *coredata.CompliancePortal,
) (*types.CompliancePortalSlackChannel, error) {
	if r.slackDestinations == nil {
		return nil, nil
	}

	destination, err := r.slackDestinations.GetDestination(
		ctx,
		scope,
		portal.OrganizationID,
		probot.DeliveryTarget{
			Namespace: "compliance_portal",
			Key:       portal.ID.String(),
		},
	)
	if err != nil {
		if errors.Is(err, slackchannel.ErrSlackbotChannelNotFound) {
			return nil, nil
		}

		r.logger.ErrorCtx(ctx, "cannot load Slack notification channel", log.Error(err))

		return nil, fmt.Errorf("internal error")
	}

	return &types.CompliancePortalSlackChannel{
		ChannelID:   destination.ExternalDestinationID,
		ChannelName: destination.ExternalName,
	}, nil
}
