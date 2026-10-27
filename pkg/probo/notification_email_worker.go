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
	"sync"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/notification"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/server/web/employeeportal"
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
		mu               sync.Mutex
		deferredGroups   map[notificationEmailGroupKey]time.Time
	}

	notificationEmailGroupKey struct {
		organizationID     gid.GID
		recipientProfileID gid.GID
		notificationType   coredata.NotificationType
	}
)

const (
	notificationEmailTaskReminder notificationEmailTaskKind = "REMINDER"
	notificationEmailTaskWeekly   notificationEmailTaskKind = "WEEKLY"

	notificationDigestItemLimit = 50

	// A failed Each or Reminders send stays due. Skip that group for this
	// long so it does not stay at the head of every poll.
	notificationEmailGroupBackoff = 15 * time.Minute
)

var errNotificationEmailDeferred = errors.New("notification email deferred")

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
		deferredGroups:   map[notificationEmailGroupKey]time.Time{},
	}

	return worker.New(
		"notification-email-worker",
		h,
		logger,
		opts...,
	)
}

func (h *notificationEmailHandler) Claim(ctx context.Context) (notificationEmailTask, error) {
	attempted := map[notificationEmailGroupKey]struct{}{}

	for {
		now := time.Now()
		debounceBefore := now.Add(-h.debounceDelay)

		notifications, key, err := h.claimReminder(ctx, now, debounceBefore)
		if errors.Is(err, errNotificationEmailDeferred) {
			if _, seen := attempted[key]; seen {
				return notificationEmailTask{}, fmt.Errorf("cannot defer notification email group: %w", err)
			}

			attempted[key] = struct{}{}

			continue
		}

		if err != nil {
			h.logger.ErrorCtx(ctx, "cannot send notification email", log.Error(err))

			return notificationEmailTask{}, fmt.Errorf("cannot claim notification email: %w", err)
		}

		if len(notifications) == 0 {
			if !h.weeklySendDue(now) {
				return notificationEmailTask{}, worker.ErrNoTask
			}

			return h.claimWeekly(ctx, now)
		}

		return notificationEmailTask{
			kind:          notificationEmailTaskReminder,
			notifications: notifications,
		}, nil
	}
}

func (h *notificationEmailHandler) claimReminder(
	ctx context.Context,
	now time.Time,
	debounceBefore time.Time,
) (coredata.Notifications, notificationEmailGroupKey, error) {
	skip := h.notificationEmailSkip(now)

	var (
		notifications coredata.Notifications
		group         coredata.NotificationEmailGroup
	)

	err := h.service.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var (
				each      coredata.NotificationEmailGroup
				reminders coredata.NotificationEmailGroup
			)

			eachErr := each.LoadNextEachDue(ctx, tx, now, skip)
			if eachErr != nil && !errors.Is(eachErr, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load next each notification: %w", eachErr)
			}

			remindersErr := reminders.LoadNextRemindersDue(
				ctx,
				tx,
				now,
				debounceBefore,
				h.reminderInterval,
				skip,
			)
			if remindersErr != nil && !errors.Is(remindersErr, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load next reminder notification: %w", remindersErr)
			}

			useEach := eachErr == nil && (remindersErr != nil || !reminders.CreatedAt.Before(each.CreatedAt))
			switch {
			case useEach:
				group = each

				if err := notifications.UpdateDueEach(ctx, tx, each, now); err != nil {
					return fmt.Errorf("cannot claim each notification: %w", err)
				}
			case remindersErr == nil:
				group = reminders

				if err := notifications.UpdateDueReminders(ctx, tx, reminders, now, debounceBefore, h.reminderInterval); err != nil {
					return fmt.Errorf("cannot claim reminder notification: %w", err)
				}

				if len(notifications) == 0 {
					return nil
				}

				var open coredata.Notifications

				if err := open.UpdateOpenReminders(ctx, tx, reminders, now); err != nil {
					return fmt.Errorf("cannot claim open reminder notifications: %w", err)
				}

				notifications = append(notifications, open...)
			default:
				return nil
			}

			if len(notifications) == 0 {
				return nil
			}

			// The schedule advance commits with the email row. A failed send
			// rolls both back, so Each and Reminders stay due. That group is
			// skipped until the backoff passes, and other groups still send.
			if err := h.processReminder(ctx, tx, notifications); err != nil {
				if ctx.Err() != nil {
					return err
				}

				return fmt.Errorf("%w: %w", errNotificationEmailDeferred, err)
			}

			return nil
		},
	)
	if errors.Is(err, errNotificationEmailDeferred) {
		key := notificationEmailGroupKey{
			organizationID:     group.OrganizationID,
			recipientProfileID: group.RecipientProfileID,
			notificationType:   group.Type,
		}

		h.deferNotificationEmailGroup(key, time.Now())

		h.logger.ErrorCtx(
			ctx,
			"cannot send notification email",
			log.Error(err),
			log.String("organization_id", group.OrganizationID.String()),
			log.String("recipient_profile_id", group.RecipientProfileID.String()),
			log.String("notification_type", group.Type.String()),
		)

		return nil, key, err
	}

	if err != nil {
		return nil, notificationEmailGroupKey{}, err
	}

	return notifications, notificationEmailGroupKey{}, nil
}

func (h *notificationEmailHandler) notificationEmailSkip(now time.Time) coredata.NotificationEmailSkip {
	h.mu.Lock()
	defer h.mu.Unlock()

	skip := coredata.NotificationEmailSkip{
		OrganizationIDs:     []string{},
		RecipientProfileIDs: []string{},
		Types:               []string{},
	}

	for key, until := range h.deferredGroups {
		if !until.After(now) {
			delete(h.deferredGroups, key)

			continue
		}

		organizationID := key.organizationID.String()
		skip.OrganizationIDs = append(skip.OrganizationIDs, organizationID)

		recipientProfileID := key.recipientProfileID.String()
		skip.RecipientProfileIDs = append(skip.RecipientProfileIDs, recipientProfileID)

		notificationType := key.notificationType.String()
		skip.Types = append(skip.Types, notificationType)
	}

	return skip
}

func (h *notificationEmailHandler) deferNotificationEmailGroup(key notificationEmailGroupKey, now time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.deferredGroups[key] = now.Add(notificationEmailGroupBackoff)
}

// claimWeekly commits one recipient before sending. A failed send keeps that
// claim and moves on, so a broken profile does not sit at the front of the
// queue for the rest of the week.
func (h *notificationEmailHandler) claimWeekly(ctx context.Context, now time.Time) (notificationEmailTask, error) {
	weekStart := weekStartUTC(now)

	for {
		var recipient coredata.WeeklyNotificationRecipient

		err := h.service.pg.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				var claimErr error

				recipient, claimErr = coredata.ClaimNextWeeklyNotificationRecipient(ctx, tx, weekStart, now)

				return claimErr
			},
		)
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return notificationEmailTask{}, worker.ErrNoTask
		}

		if err != nil {
			return notificationEmailTask{}, fmt.Errorf("cannot claim weekly notification: %w", err)
		}

		err = h.service.pg.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return h.processWeekly(ctx, tx, recipient.OrganizationID, recipient.RecipientProfileID)
			},
		)
		if err != nil {
			h.logger.ErrorCtx(ctx, "cannot send weekly notification", log.Error(err))

			if ctx.Err() != nil {
				return notificationEmailTask{}, err
			}

			continue
		}

		return notificationEmailTask{
			kind:               notificationEmailTaskWeekly,
			organizationID:     recipient.OrganizationID,
			recipientProfileID: recipient.RecipientProfileID,
		}, nil
	}
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
			h.mailer,
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
			h.mailer,
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
			return fmt.Errorf("cannot build task notification URL: %w", err)
		}

		if mainURL == "" {
			mainURL = taskURL
		}

		items = append(items, notification.TaskEmailItem{
			Title: task.Name,
			URL:   taskURL,
		})
	}

	if err := h.mailer.SendTasks(ctx, tx, profile, organization.Name, reason, mainURL, items); err != nil {
		return fmt.Errorf("cannot send task notification email: %w", err)
	}

	return nil
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
	if err := preferences.LoadByProfileID(ctx, tx, scope, profile.ID); err != nil {
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
		return fmt.Errorf("cannot build weekly notification digest: %w", err)
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
		return fmt.Errorf("cannot build weekly notification settings URL: %w", err)
	}

	settingsURL, err = h.mailer.LinkForRecipient(ctx, tx, profile, settingsURL)
	if err != nil {
		return fmt.Errorf("cannot build weekly notification settings link: %w", err)
	}

	presenter := emails.NewPresenter(h.service.baseURL, profile.FullName)

	subject, textBody, htmlBody, err := presenter.RenderNotificationDigest(
		ctx,
		*organization,
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
			return nil, fmt.Errorf("cannot load weekly signatures: %w", err)
		}

		organization.Signatures = items
	}

	if _, ok := weeklyTypes[coredata.NotificationTypeDocumentApprovalRequested]; ok {
		items, err := h.weeklyApprovals(ctx, tx, scope, profile)
		if err != nil {
			return nil, fmt.Errorf("cannot load weekly approvals: %w", err)
		}

		organization.Approvals = items
	}

	var taskTypes []coredata.NotificationType
	if _, ok := weeklyTypes[coredata.NotificationTypeTaskAssigned]; ok {
		taskTypes = append(taskTypes, coredata.NotificationTypeTaskAssigned)
	}

	if _, ok := weeklyTypes[coredata.NotificationTypeTaskMentioned]; ok {
		taskTypes = append(taskTypes, coredata.NotificationTypeTaskMentioned)
	}

	if len(taskTypes) > 0 {
		items, err := h.weeklyTasks(ctx, tx, scope, profile, taskTypes)
		if err != nil {
			return nil, fmt.Errorf("cannot load weekly tasks: %w", err)
		}

		organization.Tasks = items
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
	return h.weeklyDocuments(
		ctx,
		tx,
		scope,
		profile,
		coredata.NotificationTypeDocumentSignatureRequested,
		notificationKindSigning,
	)
}

func (h *notificationEmailHandler) weeklyApprovals(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	profile *coredata.MembershipProfile,
) ([]emails.DocumentSummary, error) {
	return h.weeklyDocuments(
		ctx,
		tx,
		scope,
		profile,
		coredata.NotificationTypeDocumentApprovalRequested,
		notificationKindApproval,
	)
}

func (h *notificationEmailHandler) weeklyDocuments(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	profile *coredata.MembershipProfile,
	notificationType coredata.NotificationType,
	kind notificationKind,
) ([]emails.DocumentSummary, error) {
	subjectIDs, err := h.pendingNotificationSubjects(ctx, tx, scope, profile, notificationType)
	if err != nil {
		return nil, err
	}

	seen := make(map[gid.GID]struct{}, len(subjectIDs))
	documentIDs := make([]gid.GID, 0, len(subjectIDs))

	for _, id := range subjectIDs {
		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}

		documentIDs = append(documentIDs, id)
	}

	if len(documentIDs) == 0 {
		return nil, nil
	}

	var documents coredata.Documents

	if err := documents.LoadByIDs(ctx, tx, scope, documentIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load weekly notification documents: %w", err)
	}

	byID := make(map[gid.GID]*coredata.Document, len(documents))

	for _, document := range documents {
		byID[document.ID] = document
	}

	items := make([]emails.DocumentSummary, 0, len(documentIDs))
	paths := make([]string, 0, len(documentIDs))

	for _, id := range documentIDs {
		document, ok := byID[id]
		if !ok || document.ArchivedAt != nil {
			continue
		}

		path, err := documentDestinationPath(kind, profile.OrganizationID, document.ID)
		if err != nil {
			return nil, err
		}

		paths = append(paths, path)

		items = append(items, emails.DocumentSummary{
			Title: document.Title,
			Type:  document.DocumentType.Label(),
		})
	}

	return h.linkDocumentSummaries(ctx, tx, profile, items, paths)
}

func (h *notificationEmailHandler) linkDocumentSummaries(
	ctx context.Context,
	tx pg.Tx,
	profile *coredata.MembershipProfile,
	items []emails.DocumentSummary,
	paths []string,
) ([]emails.DocumentSummary, error) {
	if len(items) == 0 {
		return nil, nil
	}

	if h.mailer == nil {
		return nil, fmt.Errorf("cannot build notification URL: mailer is not configured")
	}

	links, err := h.mailer.LinkPaths(ctx, tx, profile, paths)
	if err != nil {
		return nil, err
	}

	for i := range items {
		items[i].URL = links[i]
	}

	return items, nil
}

func (h *notificationEmailHandler) weeklyTasks(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	profile *coredata.MembershipProfile,
	notificationTypes []coredata.NotificationType,
) ([]emails.TaskSummary, error) {
	seen := make(map[gid.GID]struct{})
	taskIDs := make([]gid.GID, 0)

	for _, notificationType := range notificationTypes {
		ids, err := h.pendingNotificationSubjects(ctx, tx, scope, profile, notificationType)
		if err != nil {
			return nil, err
		}

		for _, id := range ids {
			if _, ok := seen[id]; ok {
				continue
			}

			seen[id] = struct{}{}
			taskIDs = append(taskIDs, id)
		}
	}

	if len(taskIDs) == 0 {
		return nil, nil
	}

	var tasks coredata.Tasks
	if err := tasks.LoadByIDs(ctx, tx, scope, taskIDs); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load weekly tasks: %w", err)
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	if h.mailer == nil {
		return nil, fmt.Errorf("cannot build task notification URL: mailer is not configured")
	}

	targets := make([]string, 0, len(tasks))

	for _, task := range tasks {
		taskURL, err := h.mailer.TaskURL(profile.OrganizationID, task.ID)
		if err != nil {
			return nil, err
		}

		targets = append(targets, taskURL)
	}

	links, err := h.mailer.LinkTargets(ctx, tx, profile, targets)
	if err != nil {
		return nil, err
	}

	items := make([]emails.TaskSummary, 0, len(tasks))

	for i, task := range tasks {
		items = append(items, emails.TaskSummary{
			Title: task.Name,
			URL:   links[i],
		})
	}

	return items, nil
}

func (h *notificationEmailHandler) pendingNotificationSubjects(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	profile *coredata.MembershipProfile,
	notificationType coredata.NotificationType,
) ([]gid.GID, error) {
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
		[]coredata.NotificationType{notificationType},
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
		return nil, fmt.Errorf("cannot load weekly notifications: %w", err)
	}

	ids := make([]gid.GID, 0, len(notifications))
	for _, item := range notifications {
		ids = append(ids, item.SubjectID)
	}

	return ids, nil
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
	mailer *notification.Mailer,
) error {
	if len(documentIDs) == 0 {
		return nil
	}

	if mailer == nil {
		return fmt.Errorf("cannot send notification email: mailer is not configured")
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

	type notificationDocument struct {
		summary emails.DocumentSummary
		path    string
	}

	rows := make([]notificationDocument, 0, len(loaded))
	for _, document := range loaded {
		documentPath, err := documentDestinationPath(kind, organizationID, document.ID)
		if err != nil {
			return fmt.Errorf("cannot build document destination path: %w", err)
		}

		rows = append(rows, notificationDocument{
			summary: emails.DocumentSummary{
				Title: document.Title,
				Type:  document.DocumentType.Label(),
			},
			path: documentPath,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].summary.Title < rows[j].summary.Title
	})

	paths := make([]string, len(rows))
	documents := make([]emails.DocumentSummary, len(rows))

	for i, row := range rows {
		paths[i] = row.path
		documents[i] = row.summary
	}

	mainPath, err := mainDestinationPath(kind, organizationID)
	if err != nil {
		return fmt.Errorf("cannot build main destination path: %w", err)
	}

	settingsPath, err := notification.SettingsPath(ctx, tx, organizationID)
	if err != nil {
		return fmt.Errorf("cannot build notification settings path: %w", err)
	}

	links, err := mailer.LinkPaths(ctx, tx, recipient, append(paths, mainPath, settingsPath))
	if err != nil {
		return fmt.Errorf("cannot build notification links: %w", err)
	}

	for i := range documents {
		documents[i].URL = links[i]
	}

	mainURL := links[len(documents)]
	settingsURL := links[len(documents)+1]

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
