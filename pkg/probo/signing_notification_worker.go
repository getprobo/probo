// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package probo

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/statelesstoken"
)

type (
	SigningNotificationWorker struct {
		pg                      *pg.Client
		fileManager             *filemanager.Service
		bucket                  string
		baseURL                 string
		tokenSecret             string
		invitationTokenValidity time.Duration
		logger                  *log.Logger
		interval                time.Duration
	}

	SigningNotificationWorkerOption func(*SigningNotificationWorker)
)

func WithSigningNotificationWorkerInterval(d time.Duration) SigningNotificationWorkerOption {
	return func(w *SigningNotificationWorker) {
		if d > 0 {
			w.interval = d
		}
	}
}

func NewSigningNotificationWorker(
	pgClient *pg.Client,
	fileManager *filemanager.Service,
	bucket string,
	baseURL string,
	tokenSecret string,
	invitationTokenValidity time.Duration,
	logger *log.Logger,
	opts ...SigningNotificationWorkerOption,
) *SigningNotificationWorker {
	w := &SigningNotificationWorker{
		pg:                      pgClient,
		fileManager:             fileManager,
		bucket:                  bucket,
		baseURL:                 baseURL,
		tokenSecret:             tokenSecret,
		invitationTokenValidity: invitationTokenValidity,
		logger:                  logger,
		interval:                10 * time.Minute,
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

func (w *SigningNotificationWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.processPendingNotifications(ctx); err != nil {
				w.logger.ErrorCtx(ctx, "cannot process signing notifications", log.Error(err))
			}
		}
	}
}

func (w *SigningNotificationWorker) processPendingNotifications(ctx context.Context) error {
	nonCancelableCtx := context.WithoutCancel(ctx)

	return w.pg.WithTx(
		nonCancelableCtx,
		func(tx pg.Conn) error {
			var signatures coredata.DocumentVersionSignatures
			if err := signatures.LoadPendingNotificationsForUpdate(nonCancelableCtx, tx, w.interval); err != nil {
				return fmt.Errorf("cannot load pending notifications: %w", err)
			}

			if len(signatures) == 0 {
				return nil
			}

			type orgProfile struct {
				organizationID gid.GID
				profileID      gid.GID
			}

			seen := make(map[orgProfile]bool)
			now := time.Now()

			for _, sig := range signatures {
				key := orgProfile{
					organizationID: sig.OrganizationID,
					profileID:      sig.SignedBy,
				}

				if !seen[key] {
					seen[key] = true
					if err := w.sendNotification(nonCancelableCtx, tx, sig); err != nil {
						return fmt.Errorf("cannot send signing notification: %w", err)
					}
				}

				sig.State = coredata.DocumentVersionSignatureStateNotified
				sig.UpdatedAt = now

				scope := coredata.NewScopeFromObjectID(sig.ID)
				if err := sig.Update(nonCancelableCtx, tx, scope); err != nil {
					return fmt.Errorf("cannot update signature state to notified: %w", err)
				}
			}

			return nil
		},
	)
}

func (w *SigningNotificationWorker) sendNotification(
	ctx context.Context,
	conn pg.Conn,
	sig *coredata.DocumentVersionSignature,
) error {
	scope := coredata.NewScopeFromObjectID(sig.ID)

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByID(ctx, conn, scope, sig.SignedBy); err != nil {
		return fmt.Errorf("cannot load profile: %w", err)
	}

	organization := &coredata.Organization{}
	if err := organization.LoadByID(ctx, conn, scope, sig.OrganizationID); err != nil {
		return fmt.Errorf("cannot load organization: %w", err)
	}

	emailPresenter := emails.NewPresenter(w.fileManager, w.bucket, w.baseURL, profile.FullName)

	var (
		employeeDocumentsURLPath = "/organizations/" + sig.OrganizationID.String() + "/employee"
		emailLinkURLPath         = employeeDocumentsURLPath
		query                    = make(url.Values)
	)

	if profile.State != coredata.ProfileStateActive {
		if profile.Source != coredata.ProfileSourceSCIM {
			now := time.Now()
			invitation := &coredata.Invitation{
				ID:             gid.New(sig.OrganizationID.TenantID(), coredata.InvitationEntityType),
				OrganizationID: sig.OrganizationID,
				UserID:         profile.ID,
				Status:         coredata.InvitationStatusPending,
				ExpiresAt:      now.Add(w.invitationTokenValidity),
				CreatedAt:      now,
			}
			if err := invitation.Insert(ctx, conn, coredata.NewScopeFromObjectID(sig.OrganizationID)); err != nil {
				return fmt.Errorf("cannot insert invitation: %w", err)
			}

			invitationToken, err := statelesstoken.NewToken(
				w.tokenSecret,
				iam.TokenTypeOrganizationInvitation,
				w.invitationTokenValidity,
				iam.InvitationTokenData{InvitationID: invitation.ID},
			)
			if err != nil {
				return fmt.Errorf("cannot generate invitation token: %w", err)
			}

			emailLinkURLPath = "/auth/activate-account"
			continueURL := baseurl.MustParse(w.baseURL).AppendPath(employeeDocumentsURLPath).MustString()
			query.Add("token", invitationToken)
			query.Add("continue", continueURL)
		}
	}

	subject, textBody, htmlBody, err := emailPresenter.RenderDocumentSigning(
		ctx,
		emailLinkURLPath,
		query,
		organization.Name,
	)
	if err != nil {
		return fmt.Errorf("cannot render signing request email: %w", err)
	}

	email := coredata.NewEmail(
		profile.FullName,
		profile.EmailAddress,
		subject,
		textBody,
		htmlBody,
		&coredata.EmailOptions{
			SenderName: new(organization.Name),
		},
	)

	if err := email.Insert(ctx, conn); err != nil {
		return fmt.Errorf("cannot insert email: %w", err)
	}

	return nil
}
