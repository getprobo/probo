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
	"errors"
	"fmt"
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
	"go.probo.inc/probo/pkg/notification"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/server/web/employeeportal"
	"go.probo.inc/probo/pkg/statelesstoken"
)

type (
	notificationEmailTaskKind string

	notificationEmailTask struct {
		kind               notificationEmailTaskKind
		notifications      coredata.Notifications
		organizationID     gid.GID
		recipientProfileID gid.GID
	}

	NotificationEmailWorkerConfig struct {
		DebounceDelay    time.Duration
		ReminderInterval time.Duration
		WeeklyWeekday    time.Weekday
		WeeklyHour       int
	}

	notificationEmailHandler struct {
		service          *Service
		mailer           *notification.Mailer
		logger           *log.Logger
		debounceDelay    time.Duration
		reminderInterval time.Duration
		weeklyWeekday    time.Weekday
		weeklyHour       int
	}
)

const (
	notificationEmailTaskReminder notificationEmailTaskKind = "REMINDER"
	notificationEmailTaskWeekly   notificationEmailTaskKind = "WEEKLY"

	notificationDigestItemLimit = 50
)

func NewNotificationEmailWorker(
	service *Service,
	mailer *notification.Mailer,
	logger *log.Logger,
	cfg NotificationEmailWorkerConfig,
	opts ...worker.Option,
) *worker.Worker[notificationEmailTask] {
	h := &notificationEmailHandler{
		service:          service,
		mailer:           mailer,
		logger:           logger,
		debounceDelay:    cfg.DebounceDelay,
		reminderInterval: cfg.ReminderInterval,
		weeklyWeekday:    cfg.WeeklyWeekday,
		weeklyHour:       cfg.WeeklyHour,
	}

	return worker.New(
		"notification-email-worker",
		h,
		logger,
		opts...,
	)
}

func (h *notificationEmailHandler) Claim(ctx context.Context) (notificationEmailTask, error) {
	now := time.Now()
	debounceBefore := now.Add(-h.debounceDelay)

	var notifications coredata.Notifications
	err := h.service.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var (
				each      coredata.NotificationEmailGroup
				reminders coredata.NotificationEmailGroup
			)

			eachErr := each.LoadNextEachDue(ctx, tx, now, debounceBefore)
			if eachErr != nil && !errors.Is(eachErr, coredata.ErrResourceNotFound) {
				return eachErr
			}

			remindersErr := reminders.LoadNextRemindersDue(ctx, tx, now, debounceBefore, h.reminderInterval)
			if remindersErr != nil && !errors.Is(remindersErr, coredata.ErrResourceNotFound) {
				return remindersErr
			}

			useEach := eachErr == nil && (remindersErr != nil || !reminders.CreatedAt.Before(each.CreatedAt))
			switch {
			case useEach:
				if err := notifications.UpdateDueEach(ctx, tx, each, now, debounceBefore); err != nil {
					return err
				}
			case remindersErr == nil:
				if err := notifications.UpdateDueReminders(ctx, tx, reminders, now, debounceBefore, h.reminderInterval); err != nil {
					return err
				}
				if len(notifications) == 0 {
					return nil
				}

				if err := notifications.UpdateOpenReminders(ctx, tx, reminders, now); err != nil {
					return err
				}
			default:
				return nil
			}

			if len(notifications) == 0 {
				return nil
			}

			// The schedule advance commits with the email row. A failed send
			// rolls both back, so Each and Reminders stay due.
			return h.processReminder(ctx, tx, notifications)
		},
	)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot send notification email", log.Error(err))

		return notificationEmailTask{}, fmt.Errorf("cannot claim notification email: %w", err)
	}

	if len(notifications) > 0 {
		return notificationEmailTask{
			kind:          notificationEmailTaskReminder,
			notifications: notifications,
		}, nil
	}

	if !h.weeklySendDue(now) {
		return notificationEmailTask{}, worker.ErrNoTask
	}

	var recipient coredata.WeeklyNotificationRecipient
	err = h.service.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var claimErr error
			recipient, claimErr = coredata.ClaimNextWeeklyNotificationRecipient(ctx, tx, weekStartUTC(now), now)
			if claimErr != nil {
				return claimErr
			}

			// A failed digest rolls the week row back so the same week can try again.
			return h.processWeekly(ctx, tx, recipient.OrganizationID, recipient.RecipientProfileID)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return notificationEmailTask{}, worker.ErrNoTask
		}

		h.logger.ErrorCtx(ctx, "cannot send notification email", log.Error(err))

		return notificationEmailTask{}, fmt.Errorf("cannot claim weekly notification: %w", err)
	}

	return notificationEmailTask{
		kind:               notificationEmailTaskWeekly,
		organizationID:     recipient.OrganizationID,
		recipientProfileID: recipient.RecipientProfileID,
	}, nil
}

func (h *notificationEmailHandler) Process(context.Context, notificationEmailTask) error {
	return nil
}

func (h *notificationEmailHandler) processReminder(
	ctx context.Context,
	tx pg.Tx,
	notifications coredata.Notifications,
) error {
	if len(notifications) == 0 {
		return nil
	}

	first := notifications[0]
	scope := coredata.NewScopeFromObjectID(first.OrganizationID)
	seen := make(map[gid.GID]struct{}, len(notifications))
	subjectIDs := make([]gid.GID, 0, len(notifications))
	for _, item := range notifications {
		if _, ok := seen[item.SubjectID]; ok {
			continue
		}

		seen[item.SubjectID] = struct{}{}
		subjectIDs = append(subjectIDs, item.SubjectID)
	}

	switch first.Type {
	case coredata.NotificationTypeDocumentSignatureRequested:
		return h.service.Documents.sendNotification(
			ctx,
			tx,
			scope,
			notificationKindSigning,
			first.OrganizationID,
			first.RecipientProfileID,
			subjectIDs,
		)
	case coredata.NotificationTypeDocumentApprovalRequested:
		return h.service.Documents.sendNotification(
			ctx,
			tx,
			scope,
			notificationKindApproval,
			first.OrganizationID,
			first.RecipientProfileID,
			subjectIDs,
		)
	case coredata.NotificationTypeTaskAssigned:
		return h.sendTaskReminder(ctx, tx, scope, first.OrganizationID, first.RecipientProfileID, notification.TaskEmailAssigned, subjectIDs)
	case coredata.NotificationTypeTaskMentioned:
		return h.sendTaskReminder(ctx, tx, scope, first.OrganizationID, first.RecipientProfileID, notification.TaskEmailMentioned, subjectIDs)
	default:
		return fmt.Errorf("unknown notification email type %q", first.Type)
	}
}

func (h *notificationEmailHandler) sendTaskReminder(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	recipientID gid.GID,
	reason notification.TaskEmailReason,
	taskIDs []gid.GID,
) error {
	if h.mailer == nil {
		return fmt.Errorf("cannot send task notification email: mailer is not configured")
	}

	var tasks coredata.Tasks
	if err := tasks.LoadByIDs(ctx, tx, scope, taskIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return fmt.Errorf("cannot load tasks for notification reminder: %w", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByID(ctx, tx, scope, recipientID); err != nil {
		return fmt.Errorf("cannot load task notification recipient: %w", err)
	}

	organization := &coredata.Organization{}
	if err := organization.LoadByID(ctx, tx, scope, organizationID); err != nil {
		return fmt.Errorf("cannot load task notification organization: %w", err)
	}

	items := make([]notification.TaskEmailItem, 0, len(tasks))
	var mainURL string
	for _, task := range tasks {
		taskURL, err := h.mailer.TaskURL(organizationID, task.ID)
		if err != nil {
			return err
		}

		if mainURL == "" {
			mainURL = taskURL
		}

		items = append(items, notification.TaskEmailItem{
			Title: task.Name,
			URL:   taskURL,
		})
	}

	return h.mailer.SendTasks(ctx, tx, profile, organization.Name, reason, mainURL, items)
}

func (h *notificationEmailHandler) processWeekly(
	ctx context.Context,
	tx pg.Tx,
	organizationID gid.GID,
	recipientProfileID gid.GID,
) error {
	if h.mailer == nil {
		return fmt.Errorf("cannot send weekly notification email: mailer is not configured")
	}

	scope := coredata.NewScopeFromObjectID(organizationID)
	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByID(ctx, tx, scope, recipientProfileID); err != nil {
		return fmt.Errorf("cannot load weekly notification recipient: %w", err)
	}

	if !profileCanReceiveNotification(profile, time.Now()) {
		return nil
	}

	var preferences coredata.NotificationPreferences
	if err := preferences.LoadByProfileID(ctx, tx, profile.ID); err != nil {
		return fmt.Errorf("cannot load weekly notification preferences: %w", err)
	}

	weeklyTypes := map[coredata.NotificationType]struct{}{}
	for _, preference := range preferences {
		if preference.EmailDelivery == coredata.NotificationEmailDeliveryWeekly {
			weeklyTypes[preference.Type] = struct{}{}
		}
	}

	if len(weeklyTypes) == 0 {
		return nil
	}

	organization, err := h.weeklyOrganization(ctx, tx, profile, weeklyTypes)
	if err != nil {
		return err
	}

	if organization == nil {
		return nil
	}

	record := &coredata.Organization{}
	if err := record.LoadByID(ctx, tx, scope, organizationID); err != nil {
		return fmt.Errorf("cannot load weekly notification organization: %w", err)
	}

	organization.Name = record.Name

	settingsURL, err := h.mailer.SettingsURL(ctx, tx, organizationID)
	if err != nil {
		return err
	}

	settingsURL, err = h.mailer.LinkForRecipient(ctx, tx, profile, settingsURL)
	if err != nil {
		return err
	}

	presenter := emails.NewPresenter(h.service.baseURL, profile.FullName)
	subject, textBody, htmlBody, err := presenter.RenderNotificationDigest(
		ctx,
		[]emails.NotificationDigestOrganization{*organization},
		settingsURL,
	)
	if err != nil {
		return fmt.Errorf("cannot render weekly notification email: %w", err)
	}

	email := coredata.NewEmail(
		profile.FullName,
		profile.EmailAddress,
		subject,
		textBody,
		htmlBody,
		&coredata.EmailOptions{SenderName: &record.Name},
	)
	if err := email.Insert(ctx, tx); err != nil {
		return fmt.Errorf("cannot insert weekly notification email: %w", err)
	}

	return nil
}

func (h *notificationEmailHandler) weeklyOrganization(
	ctx context.Context,
	tx pg.Tx,
	profile *coredata.MembershipProfile,
	weeklyTypes map[coredata.NotificationType]struct{},
) (*emails.NotificationDigestOrganization, error) {
	scope := coredata.NewScopeFromObjectID(profile.OrganizationID)
	organization := emails.NotificationDigestOrganization{Name: profile.OrganizationName}

	if _, ok := weeklyTypes[coredata.NotificationTypeDocumentSignatureRequested]; ok {
		items, err := h.weeklySignatures(ctx, tx, scope, profile)
		if err != nil {
			return nil, err
		}

		organization.Signatures = items
	}

	if _, ok := weeklyTypes[coredata.NotificationTypeDocumentApprovalRequested]; ok {
		items, err := h.weeklyApprovals(ctx, tx, scope, profile)
		if err != nil {
			return nil, err
		}

		organization.Approvals = items
	}

	if _, ok := weeklyTypes[coredata.NotificationTypeTaskAssigned]; ok {
		items, err := h.weeklyTasks(ctx, tx, scope, profile)
		if err != nil {
			return nil, err
		}

		organization.Tasks = append(organization.Tasks, items...)
	}

	if _, ok := weeklyTypes[coredata.NotificationTypeTaskMentioned]; ok {
		items, err := h.weeklyMentions(ctx, tx, scope, profile)
		if err != nil {
			return nil, err
		}

		organization.Tasks = append(organization.Tasks, items...)
	}

	if len(organization.Signatures) == 0 && len(organization.Approvals) == 0 && len(organization.Tasks) == 0 {
		return nil, nil
	}

	return &organization, nil
}

func (h *notificationEmailHandler) weeklySignatures(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	profile *coredata.MembershipProfile,
) ([]emails.DocumentSummary, error) {
	var signatures coredata.DocumentVersionSignatures
	if err := signatures.LoadByStateAndSignatory(
		ctx,
		tx,
		scope,
		profile.ID,
		[]coredata.DocumentVersionSignatureState{coredata.DocumentVersionSignatureStateRequested},
		page.NewCursor(
			notificationDigestItemLimit,
			nil,
			page.Head,
			page.OrderBy[coredata.DocumentVersionSignatureOrderField]{
				Field:     coredata.DocumentVersionSignatureOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		),
	); err != nil {
		return nil, fmt.Errorf("cannot load weekly signature requests: %w", err)
	}

	if len(signatures) == 0 {
		return nil, nil
	}

	token, err := h.service.Documents.buildInvitationToken(ctx, tx, profile)
	if err != nil {
		return nil, err
	}

	items := make([]emails.DocumentSummary, 0, len(signatures))
	for _, signature := range signatures {
		version := &coredata.DocumentVersion{}
		err := version.LoadByID(ctx, tx, scope, signature.DocumentVersionID)
		if errors.Is(err, coredata.ErrResourceNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("cannot load weekly signature document version: %w", err)
		}

		open, err := documentVersionOpen(ctx, tx, scope, version)
		if err != nil {
			return nil, err
		}
		if !open {
			continue
		}

		path, err := documentDestinationPath(notificationKindSigning, profile.OrganizationID, version.DocumentID)
		if err != nil {
			return nil, err
		}

		documentURL, err := h.service.Documents.recipientURL(path, token)
		if err != nil {
			return nil, err
		}

		items = append(items, emails.DocumentSummary{
			Title: version.Title,
			Type:  version.DocumentType.Label(),
			URL:   documentURL,
		})
	}

	return items, nil
}

func (h *notificationEmailHandler) weeklyApprovals(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	profile *coredata.MembershipProfile,
) ([]emails.DocumentSummary, error) {
	var decisions coredata.DocumentVersionApprovalDecisions
	if err := decisions.LoadByStateAndApprover(
		ctx,
		tx,
		scope,
		profile.ID,
		[]coredata.DocumentVersionApprovalDecisionState{coredata.DocumentVersionApprovalDecisionStatePending},
		page.NewCursor(
			notificationDigestItemLimit,
			nil,
			page.Head,
			page.OrderBy[coredata.DocumentVersionApprovalDecisionOrderField]{
				Field:     coredata.DocumentVersionApprovalDecisionOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		),
	); err != nil {
		return nil, fmt.Errorf("cannot load weekly approval requests: %w", err)
	}

	if len(decisions) == 0 {
		return nil, nil
	}

	token, err := h.service.Documents.buildInvitationToken(ctx, tx, profile)
	if err != nil {
		return nil, err
	}

	items := make([]emails.DocumentSummary, 0, len(decisions))
	for _, decision := range decisions {
		quorum := &coredata.DocumentVersionApprovalQuorum{}
		err := quorum.LoadByID(ctx, tx, scope, decision.QuorumID)
		if errors.Is(err, coredata.ErrResourceNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("cannot load weekly approval quorum: %w", err)
		}

		version := &coredata.DocumentVersion{}
		err = version.LoadByID(ctx, tx, scope, quorum.VersionID)
		if errors.Is(err, coredata.ErrResourceNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("cannot load weekly approval document version: %w", err)
		}

		open, err := documentVersionOpen(ctx, tx, scope, version)
		if err != nil {
			return nil, err
		}
		if !open {
			continue
		}

		path, err := documentDestinationPath(notificationKindApproval, profile.OrganizationID, version.DocumentID)
		if err != nil {
			return nil, err
		}

		documentURL, err := h.service.Documents.recipientURL(path, token)
		if err != nil {
			return nil, err
		}

		items = append(items, emails.DocumentSummary{
			Title: version.Title,
			Type:  version.DocumentType.Label(),
			URL:   documentURL,
		})
	}

	return items, nil
}

func (h *notificationEmailHandler) weeklyTasks(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	profile *coredata.MembershipProfile,
) ([]emails.TaskSummary, error) {
	var tasks coredata.Tasks
	if err := tasks.LoadByStateAndAssignee(
		ctx,
		tx,
		scope,
		profile.ID,
		[]coredata.TaskState{
			coredata.TaskStateBacklog,
			coredata.TaskStateTodo,
			coredata.TaskStateInProgress,
		},
		page.NewCursor(
			notificationDigestItemLimit,
			nil,
			page.Head,
			page.OrderBy[coredata.TaskOrderField]{
				Field:     coredata.TaskOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		),
	); err != nil {
		return nil, fmt.Errorf("cannot load weekly assigned tasks: %w", err)
	}

	items := make([]emails.TaskSummary, 0, len(tasks))
	for _, task := range tasks {
		taskURL, err := h.mailer.TaskURL(profile.OrganizationID, task.ID)
		if err != nil {
			return nil, err
		}

		taskURL, err = h.mailer.LinkForRecipient(ctx, tx, profile, taskURL)
		if err != nil {
			return nil, err
		}

		items = append(items, emails.TaskSummary{
			Title: task.Name,
			URL:   taskURL,
		})
	}

	return items, nil
}

func (h *notificationEmailHandler) weeklyMentions(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	profile *coredata.MembershipProfile,
) ([]emails.TaskSummary, error) {
	var notifications coredata.Notifications
	cursor := page.NewCursor(
		notificationDigestItemLimit,
		nil,
		page.Head,
		page.OrderBy[coredata.NotificationOrderField]{
			Field:     coredata.NotificationOrderFieldCreatedAt,
			Direction: page.OrderDirectionDesc,
		},
	)
	filter := coredata.NewNotificationFilter(
		[]coredata.NotificationType{coredata.NotificationTypeTaskMentioned},
		[]coredata.NotificationStatus{coredata.NotificationStatusPending},
	)
	if err := notifications.LoadByRecipientProfileID(
		ctx,
		tx,
		scope,
		profile.OrganizationID,
		profile.ID,
		cursor,
		filter,
	); err != nil {
		return nil, fmt.Errorf("cannot load weekly task mentions: %w", err)
	}

	if len(notifications) == 0 {
		return nil, nil
	}

	taskIDs := make([]gid.GID, 0, len(notifications))
	for _, item := range notifications {
		taskIDs = append(taskIDs, item.SubjectID)
	}

	var tasks coredata.Tasks
	if err := tasks.LoadByIDs(ctx, tx, scope, taskIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load weekly mentioned tasks: %w", err)
	}

	items := make([]emails.TaskSummary, 0, len(tasks))
	for _, task := range tasks {
		taskURL, err := h.mailer.TaskURL(profile.OrganizationID, task.ID)
		if err != nil {
			return nil, err
		}

		taskURL, err = h.mailer.LinkForRecipient(ctx, tx, profile, taskURL)
		if err != nil {
			return nil, err
		}

		items = append(items, emails.TaskSummary{
			Title: task.Name,
			URL:   taskURL,
		})
	}

	return items, nil
}

func documentVersionOpen(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	version *coredata.DocumentVersion,
) (bool, error) {
	document := &coredata.Document{}
	if err := document.LoadByID(ctx, tx, scope, version.DocumentID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("cannot load notification document: %w", err)
	}

	return document.ArchivedAt == nil, nil
}

func profileCanReceiveNotification(profile *coredata.MembershipProfile, now time.Time) bool {
	if profile.State != coredata.ProfileStateActive {
		return false
	}

	if profile.ContractEndDate == nil {
		return true
	}

	today := time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day(), 0, 0, 0, 0, time.UTC)

	return !profile.ContractEndDate.Before(today)
}

func (h *notificationEmailHandler) weeklySendDue(now time.Time) bool {
	utc := now.UTC()

	return utc.Weekday() == h.weeklyWeekday && utc.Hour() >= h.weeklyHour
}

func weekStartUTC(now time.Time) time.Time {
	utc := now.UTC()

	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

type notificationKind string

const (
	notificationKindSigning  notificationKind = "signing"
	notificationKindApproval notificationKind = "approval"
)

func (s *DocumentService) sendNotification(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	kind notificationKind,
	organizationID gid.GID,
	recipientID gid.GID,
	documentIDs []gid.GID,
) error {
	if len(documentIDs) == 0 {
		return nil
	}

	var loaded coredata.Documents
	if err := loaded.LoadByIDs(ctx, tx, scope, documentIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return fmt.Errorf("cannot load documents for notification: %w", err)
	}

	if len(loaded) == 0 {
		return nil
	}

	var profiles coredata.MembershipProfiles
	if err := profiles.LoadByIDs(ctx, tx, scope, []gid.GID{recipientID}); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
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

	documents := make([]emails.DocumentSummary, 0, len(loaded))
	for _, document := range loaded {
		documentPath, err := documentDestinationPath(kind, organizationID, document.ID)
		if err != nil {
			return fmt.Errorf("cannot build document destination path: %w", err)
		}

		documentURL, err := s.recipientURL(documentPath, token)
		if err != nil {
			return fmt.Errorf("cannot build document notification URL: %w", err)
		}

		documents = append(documents, emails.DocumentSummary{
			Title: document.Title,
			Type:  document.DocumentType.Label(),
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

	settingsPath, err := notification.SettingsPath(ctx, tx, organizationID)
	if err != nil {
		return fmt.Errorf("cannot build notification settings path: %w", err)
	}

	settingsURL, err := s.recipientURL(settingsPath, token)
	if err != nil {
		return fmt.Errorf("cannot build notification settings URL: %w", err)
	}

	email, err := s.renderNotificationEmail(ctx, kind, recipient, organization.Name, mainURL, documents, settingsURL)
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
	settingsURL string,
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
		subject, textBody, htmlBody, err = emailPresenter.RenderDocumentSigning(ctx, mainURL, organizationName, documents, settingsURL)
	case notificationKindApproval:
		subject, textBody, htmlBody, err = emailPresenter.RenderDocumentApproval(ctx, mainURL, organizationName, documents, settingsURL)
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
	path, err := employeeportal.OrganizationPath(organizationID.String(), notificationSection(kind))
	if err != nil {
		return "", fmt.Errorf("cannot build notification path: %w", err)
	}

	return path, nil
}

func documentDestinationPath(kind notificationKind, organizationID gid.GID, documentID gid.GID) (string, error) {
	path, err := employeeportal.OrganizationPath(
		organizationID.String(),
		notificationSection(kind),
		documentID.String(),
	)
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
