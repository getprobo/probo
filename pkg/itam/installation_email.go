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

package itam

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/server/web/employeeportal"
	"go.probo.inc/probo/pkg/statelesstoken"
)

type SendDeviceInstallationEmailsRequest struct {
	OrganizationID gid.GID
}

func (s *Service) SendDeviceInstallationEmails(
	ctx context.Context,
	scope coredata.Scoper,
	req SendDeviceInstallationEmailsRequest,
) (int, error) {
	var count int

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			ended := false
			kind := "EMPLOYEE"
			filter := coredata.NewMembershipProfileFilter(&ended).
				WithMembership().
				WithStates(coredata.ProfileStateActive, coredata.ProfileStatePending).
				WithKind(&kind).
				WithoutActiveDevice()

			profiles, err := page.LoadAll(
				ctx,
				page.OrderBy[coredata.MembershipProfileOrderField]{
					Field:     coredata.MembershipProfileOrderFieldCreatedAt,
					Direction: page.OrderDirectionAsc,
				},
				func(
					ctx context.Context,
					cursor *page.Cursor[coredata.MembershipProfileOrderField],
				) ([]*coredata.MembershipProfile, error) {
					var batch coredata.MembershipProfiles
					if err := batch.LoadByOrganizationID(
						ctx,
						conn,
						scope,
						req.OrganizationID,
						cursor,
						filter,
					); err != nil {
						return nil, fmt.Errorf("cannot load profiles: %w", err)
					}

					return batch, nil
				},
			)
			if err != nil {
				return fmt.Errorf("cannot load uncovered employees: %w", err)
			}

			for _, profile := range profiles {
				installURL, err := s.installationURL(ctx, conn, scope, profile)
				if err != nil {
					return fmt.Errorf("cannot build installation URL: %w", err)
				}

				presenter := emails.NewPresenter(s.baseURL, profile.FullName)

				subject, textBody, htmlBody, err := presenter.RenderDeviceInstallation(
					ctx,
					installURL,
					organization.Name,
				)
				if err != nil {
					return fmt.Errorf("cannot render device installation email: %w", err)
				}

				email := coredata.NewEmail(
					profile.FullName,
					profile.EmailAddress,
					subject,
					textBody,
					htmlBody,
					&coredata.EmailOptions{SenderName: new(organization.Name)},
				)
				if err := email.Insert(ctx, conn); err != nil {
					return fmt.Errorf("cannot insert device installation email: %w", err)
				}
			}

			count = len(profiles)

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	s.logger.InfoCtx(
		ctx,
		"queued device installation emails",
		log.String("organization_id", req.OrganizationID.String()),
		log.Int("recipient_count", count),
	)

	return count, nil
}

func (s *Service) installationURL(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	profile *coredata.MembershipProfile,
) (string, error) {
	destination, err := employeeportal.OrganizationPath(
		profile.OrganizationID.String(),
		"devices",
		"register",
	)
	if err != nil {
		return "", fmt.Errorf("cannot build employee portal path: %w", err)
	}

	target, err := baseurl.MustParse(s.baseURL).AppendPath(destination).String()
	if err != nil {
		return "", fmt.Errorf("cannot build installation URL: %w", err)
	}

	if profile.State != coredata.ProfileStatePending || profile.Source == coredata.ProfileSourceSCIM {
		return target, nil
	}

	now := time.Now()

	invitation := &coredata.Invitation{
		ID:             gid.New(profile.OrganizationID.TenantID(), coredata.InvitationEntityType),
		OrganizationID: profile.OrganizationID,
		UserID:         profile.ID,
		Status:         coredata.InvitationStatusPending,
		ExpiresAt:      now.Add(s.invitationTokenValidity),
		CreatedAt:      now,
	}
	if err := invitation.Insert(ctx, conn, scope); err != nil {
		return "", fmt.Errorf("cannot insert invitation: %w", err)
	}

	token, err := statelesstoken.NewToken(
		s.tokenSecret,
		iam.TokenTypeOrganizationInvitation,
		s.invitationTokenValidity,
		iam.InvitationTokenData{InvitationID: invitation.ID},
	)
	if err != nil {
		return "", fmt.Errorf("cannot generate invitation token: %w", err)
	}

	activationURL, err := baseurl.MustParse(s.baseURL).
		AppendPath("/auth/activate-account").
		WithQuery("token", token).
		WithQuery("continue", target).
		String()
	if err != nil {
		return "", fmt.Errorf("cannot build activation URL: %w", err)
	}

	return activationURL, nil
}
