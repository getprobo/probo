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

package probo

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/statelesstoken"
)

type (
	notificationKind string

	claimStatus string

	// documentNotificationTask is one consolidated email to send: every document
	// awaiting a given recipient's signature or approval in one organization. The
	// schedule has already been advanced (the requests are claimed) by the time a
	// task is handed to Process.
	documentNotificationTask struct {
		kind           notificationKind
		organizationID gid.GID
		recipientID    gid.GID
		versionIDs     []gid.GID
	}

	DocumentNotificationWorkerConfig struct {
		DebounceDelay    time.Duration
		ReminderInterval time.Duration
	}

	documentNotificationHandler struct {
		service          *Service
		logger           *log.Logger
		debounceDelay    time.Duration
		reminderInterval time.Duration
	}
)

const (
	notificationKindSigning  notificationKind = "signing"
	notificationKindApproval notificationKind = "approval"

	claimStatusNone    claimStatus = "none"
	claimStatusClaimed claimStatus = "claimed"
	claimStatusRaced   claimStatus = "raced"
)

// NewDocumentNotificationWorker builds the worker that emails recipients, one
// consolidated message per organization, about the documents awaiting their
// signature or approval. Each claim advances the request's reminder schedule, so
// the conditional update doubles as the claim and concurrent workers never email
// the same group twice.
func NewDocumentNotificationWorker(
	service *Service,
	logger *log.Logger,
	cfg DocumentNotificationWorkerConfig,
	opts ...worker.Option,
) *worker.Worker[documentNotificationTask] {
	h := &documentNotificationHandler{
		service:          service,
		logger:           logger,
		debounceDelay:    cfg.DebounceDelay,
		reminderInterval: cfg.ReminderInterval,
	}

	return worker.New(
		"document-notification-worker",
		h,
		logger,
		opts...,
	)
}

func (h *documentNotificationHandler) Claim(ctx context.Context) (documentNotificationTask, error) {
	now := time.Now()
	debounceBefore := now.Add(-h.debounceDelay)

	task, status, err := h.claimNextSigningGroup(ctx, now, debounceBefore)
	if err != nil {
		return documentNotificationTask{}, err
	}

	for status == claimStatusRaced {
		task, status, err = h.claimNextSigningGroup(ctx, now, debounceBefore)
		if err != nil {
			return documentNotificationTask{}, err
		}
	}

	if status == claimStatusClaimed {
		return task, nil
	}

	task, status, err = h.claimNextApprovalGroup(ctx, now, debounceBefore)
	if err != nil {
		return documentNotificationTask{}, err
	}

	for status == claimStatusRaced {
		task, status, err = h.claimNextApprovalGroup(ctx, now, debounceBefore)
		if err != nil {
			return documentNotificationTask{}, err
		}
	}

	if status == claimStatusClaimed {
		return task, nil
	}

	return documentNotificationTask{}, worker.ErrNoTask
}

func (h *documentNotificationHandler) Process(ctx context.Context, task documentNotificationTask) error {
	if len(task.versionIDs) == 0 {
		return nil
	}

	scope := coredata.NewScopeFromObjectID(task.organizationID)

	if err := h.service.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return h.service.Documents.sendNotification(ctx, tx, scope, task.kind, task.organizationID, task.recipientID, task.versionIDs)
		},
	); err != nil {
		h.logger.ErrorCtx(
			ctx,
			"document notification worker failure",
			log.Error(err),
			log.String("organization_id", task.organizationID.String()),
		)

		return err
	}

	return nil
}

// claimNextSigningGroup claims the next (organization, signatory) group whose
// signatures are due and returns the documents to list. claimStatusRaced means
// another worker claimed the candidate group first and the caller should retry.
func (h *documentNotificationHandler) claimNextSigningGroup(
	ctx context.Context,
	now time.Time,
	debounceBefore time.Time,
) (documentNotificationTask, claimStatus, error) {
	var (
		signatures coredata.DocumentVersionSignatures
		claimed    []gid.GID
	)

	if err := h.service.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := signatures.LoadNextDueGroupForNotification(ctx, tx, now, debounceBefore, h.reminderInterval); err != nil {
				return fmt.Errorf("cannot load next due signature group: %w", err)
			}

			if len(signatures) == 0 {
				return nil
			}

			dueClaimed, err := signatures.ClaimForNotification(ctx, tx, now, debounceBefore, h.reminderInterval)
			if err != nil {
				return fmt.Errorf("cannot claim signatures for notification: %w", err)
			}

			if len(dueClaimed) == 0 {
				return nil
			}

			bumped, err := signatures.BumpRemainingForNotification(ctx, tx, dueClaimed, now)
			if err != nil {
				return fmt.Errorf("cannot bump remaining signatures for notification: %w", err)
			}

			claimed = append(dueClaimed, bumped...)

			return nil
		},
	); err != nil {
		return documentNotificationTask{}, claimStatusNone, err
	}

	if len(signatures) == 0 {
		return documentNotificationTask{}, claimStatusNone, nil
	}

	if len(claimed) == 0 {
		return documentNotificationTask{}, claimStatusRaced, nil
	}

	claimedSet := make(map[gid.GID]struct{}, len(claimed))
	for _, id := range claimed {
		claimedSet[id] = struct{}{}
	}

	versionIDs := make([]gid.GID, 0, len(signatures))
	for _, signature := range signatures {
		if _, ok := claimedSet[signature.ID]; ok {
			versionIDs = append(versionIDs, signature.DocumentVersionID)
		}
	}

	group := signatures[0]

	return documentNotificationTask{
		kind:           notificationKindSigning,
		organizationID: group.OrganizationID,
		recipientID:    group.SignedBy,
		versionIDs:     versionIDs,
	}, claimStatusClaimed, nil
}

// claimNextApprovalGroup claims the next (organization, approver) group whose
// decisions are due and resolves the documents to list via their quorums.
// claimStatusRaced means another worker won the candidate group first.
func (h *documentNotificationHandler) claimNextApprovalGroup(
	ctx context.Context,
	now time.Time,
	debounceBefore time.Time,
) (documentNotificationTask, claimStatus, error) {
	var (
		decisions      coredata.DocumentVersionApprovalDecisions
		claimed        []gid.GID
		versionIDs     []gid.GID
		organizationID gid.GID
		recipientID    gid.GID
	)

	if err := h.service.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := decisions.LoadNextDueGroupForNotification(ctx, tx, now, debounceBefore, h.reminderInterval); err != nil {
				return fmt.Errorf("cannot load next due approval group: %w", err)
			}

			if len(decisions) == 0 {
				return nil
			}

			dueClaimed, err := decisions.ClaimForNotification(ctx, tx, now, debounceBefore, h.reminderInterval)
			if err != nil {
				return fmt.Errorf("cannot claim approval decisions for notification: %w", err)
			}

			if len(dueClaimed) == 0 {
				return nil
			}

			bumped, err := decisions.BumpRemainingForNotification(ctx, tx, dueClaimed, now)
			if err != nil {
				return fmt.Errorf("cannot bump remaining approval decisions for notification: %w", err)
			}

			claimed = append(dueClaimed, bumped...)

			claimedSet := make(map[gid.GID]struct{}, len(claimed))
			for _, id := range claimed {
				claimedSet[id] = struct{}{}
			}

			quorumIDs := make([]gid.GID, 0, len(claimed))

			for _, decision := range decisions {
				if _, ok := claimedSet[decision.ID]; ok {
					quorumIDs = append(quorumIDs, decision.QuorumID)
				}
			}

			scope := coredata.NewScopeFromObjectID(decisions[0].OrganizationID)

			var quorums coredata.DocumentVersionApprovalQuorums
			if err := quorums.LoadByIDs(ctx, tx, scope, quorumIDs); err != nil {
				return fmt.Errorf("cannot load approval quorums: %w", err)
			}

			for _, quorum := range quorums {
				versionIDs = append(versionIDs, quorum.VersionID)
			}

			organizationID = decisions[0].OrganizationID
			recipientID = decisions[0].ApproverID

			return nil
		},
	); err != nil {
		return documentNotificationTask{}, claimStatusNone, err
	}

	if len(decisions) == 0 {
		return documentNotificationTask{}, claimStatusNone, nil
	}

	if len(claimed) == 0 {
		return documentNotificationTask{}, claimStatusRaced, nil
	}

	return documentNotificationTask{
		kind:           notificationKindApproval,
		organizationID: organizationID,
		recipientID:    recipientID,
		versionIDs:     versionIDs,
	}, claimStatusClaimed, nil
}

func (s *DocumentService) sendNotification(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	kind notificationKind,
	organizationID gid.GID,
	recipientID gid.GID,
	versionIDs []gid.GID,
) error {
	if len(versionIDs) == 0 {
		return nil
	}

	var versions coredata.DocumentVersions
	if err := versions.LoadByIDs(ctx, tx, scope, versionIDs); err != nil {
		return fmt.Errorf("cannot load document versions for notification: %w", err)
	}

	if len(versions) == 0 {
		return nil
	}

	var profiles coredata.MembershipProfiles
	if err := profiles.LoadByIDs(ctx, tx, scope, []gid.GID{recipientID}); err != nil {
		return fmt.Errorf("cannot load notification recipient: %w", err)
	}

	if len(profiles) == 0 {
		return nil
	}

	recipient := profiles[0]

	organization := &coredata.Organization{}
	if err := organization.LoadByID(ctx, tx, scope, organizationID); err != nil {
		return fmt.Errorf("cannot load notification organization: %w", err)
	}

	token, err := s.buildInvitationToken(ctx, tx, recipient)
	if err != nil {
		return err
	}

	documents := make([]emails.DocumentSummary, 0, len(versions))
	for _, version := range versions {
		documentPath, err := documentDestinationPath(kind, organizationID, version.DocumentID)
		if err != nil {
			return fmt.Errorf("cannot build document destination path: %w", err)
		}

		documentURL, err := s.recipientURL(documentPath, token)
		if err != nil {
			return fmt.Errorf("cannot build document notification URL: %w", err)
		}

		documents = append(documents, emails.DocumentSummary{
			Title: version.Title,
			Type:  version.DocumentType.Label(),
			URL:   documentURL,
		})
	}

	sort.Slice(documents, func(i, j int) bool {
		return documents[i].Title < documents[j].Title
	})

	mainPath, err := mainDestinationPath(kind, organizationID)
	if err != nil {
		return fmt.Errorf("cannot build main destination path: %w", err)
	}

	mainURL, err := s.recipientURL(mainPath, token)
	if err != nil {
		return fmt.Errorf("cannot build main notification URL: %w", err)
	}

	email, err := s.renderNotificationEmail(ctx, kind, recipient, organization.Name, mainURL, documents)
	if err != nil {
		return fmt.Errorf("cannot render notification email: %w", err)
	}

	if err := email.Insert(ctx, tx); err != nil {
		return fmt.Errorf("cannot insert notification email: %w", err)
	}

	return nil
}

func (s *DocumentService) renderNotificationEmail(
	ctx context.Context,
	kind notificationKind,
	recipient *coredata.MembershipProfile,
	organizationName string,
	mainURL string,
	documents []emails.DocumentSummary,
) (*coredata.Email, error) {
	emailPresenter := emails.NewPresenter(s.svc.baseURL, recipient.FullName)

	var (
		subject  string
		textBody string
		htmlBody *string
		err      error
	)

	switch kind {
	case notificationKindSigning:
		subject, textBody, htmlBody, err = emailPresenter.RenderDocumentSigning(ctx, mainURL, organizationName, documents)
	case notificationKindApproval:
		subject, textBody, htmlBody, err = emailPresenter.RenderDocumentApproval(ctx, mainURL, organizationName, documents)
	default:
		return nil, fmt.Errorf("unknown notification kind %q", kind)
	}

	if err != nil {
		return nil, fmt.Errorf("cannot render notification email body: %w", err)
	}

	return coredata.NewEmail(
		recipient.FullName,
		recipient.EmailAddress,
		subject,
		textBody,
		htmlBody,
		&coredata.EmailOptions{
			SenderName: new(organizationName),
		},
	), nil
}

// buildInvitationToken returns an activation token for invited recipients that
// are not yet active and not managed by SCIM; others get an empty token and a
// direct link.
func (s *DocumentService) buildInvitationToken(
	ctx context.Context,
	tx pg.Tx,
	recipient *coredata.MembershipProfile,
) (string, error) {
	if recipient.State == coredata.ProfileStateActive || recipient.Source == coredata.ProfileSourceSCIM {
		return "", nil
	}

	now := time.Now()
	invitation := &coredata.Invitation{
		ID:             gid.New(recipient.OrganizationID.TenantID(), coredata.InvitationEntityType),
		OrganizationID: recipient.OrganizationID,
		UserID:         recipient.ID,
		Status:         coredata.InvitationStatusPending,
		ExpiresAt:      now.Add(s.invitationTokenValidity),
		CreatedAt:      now,
	}

	if err := invitation.Insert(ctx, tx, coredata.NewScopeFromObjectID(recipient.OrganizationID)); err != nil {
		return "", fmt.Errorf("cannot insert invitation: %w", err)
	}

	invitationToken, err := statelesstoken.NewToken(
		s.tokenSecret,
		iam.TokenTypeOrganizationInvitation,
		s.invitationTokenValidity,
		iam.InvitationTokenData{InvitationID: invitation.ID},
	)
	if err != nil {
		return "", fmt.Errorf("cannot generate invitation token: %w", err)
	}

	return invitationToken, nil
}

// recipientURL builds the absolute link for destinationPath, routing through the
// account activation flow when token is set.
func (s *DocumentService) recipientURL(destinationPath string, token string) (string, error) {
	target, err := baseurl.MustParse(s.svc.baseURL).AppendPath(destinationPath).String()
	if err != nil {
		return "", fmt.Errorf("cannot build destination URL: %w", err)
	}

	if token == "" {
		return target, nil
	}

	activationURL, err := baseurl.MustParse(s.svc.baseURL).
		AppendPath("/auth/activate-account").
		WithQuery("token", token).
		WithQuery("continue", target).
		String()
	if err != nil {
		return "", fmt.Errorf("cannot build activation URL: %w", err)
	}

	return activationURL, nil
}

func mainDestinationPath(kind notificationKind, organizationID gid.GID) (string, error) {
	path, err := url.JoinPath("/organizations", organizationID.String(), "employee", notificationSection(kind))
	if err != nil {
		return "", fmt.Errorf("cannot build notification path: %w", err)
	}

	return path, nil
}

func documentDestinationPath(kind notificationKind, organizationID gid.GID, documentID gid.GID) (string, error) {
	path, err := url.JoinPath("/organizations", organizationID.String(), "employee", notificationSection(kind), documentID.String())
	if err != nil {
		return "", fmt.Errorf("cannot build document notification path: %w", err)
	}

	return path, nil
}

func notificationSection(kind notificationKind) string {
	if kind == notificationKindApproval {
		return "approvals"
	}

	return "signatures"
}
