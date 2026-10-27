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

package notification

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/server/web/employeeportal"
	"go.probo.inc/probo/pkg/statelesstoken"
)

type (
	Mailer struct {
		baseURL            *baseurl.BaseURL
		tokenSecret        string
		invitationValidity time.Duration
	}

	TaskEmailItem struct {
		Title string
		URL   string
	}

	TaskEmailReason string
)

const (
	TaskEmailAssigned  TaskEmailReason = "assigned"
	TaskEmailMentioned TaskEmailReason = "mentioned"
)

func NewMailer(rawBaseURL string, tokenSecret string, invitationValidity time.Duration) (*Mailer, error) {
	parsed, err := baseurl.Parse(rawBaseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse notification mailer base URL: %w", err)
	}

	return &Mailer{
		baseURL:            parsed,
		tokenSecret:        tokenSecret,
		invitationValidity: invitationValidity,
	}, nil
}

func SettingsPath(ctx context.Context, conn pg.Querier, organizationID gid.GID) (string, error) {
	segment := organizationID.String()

	portal := &coredata.EmployeePortal{}

	err := portal.LoadOldestByOrganizationID(
		ctx,
		conn,
		coredata.NewScopeFromObjectID(organizationID),
		organizationID,
	)
	if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return "", fmt.Errorf("cannot load employee portal for notification settings: %w", err)
	}

	if err == nil {
		segment = portal.ID.String()
	}

	path, err := employeeportal.OrganizationPath(segment, "notification-settings")
	if err != nil {
		return "", fmt.Errorf("cannot build notification settings path: %w", err)
	}

	return path, nil
}

func (m *Mailer) SettingsURL(ctx context.Context, conn pg.Querier, organizationID gid.GID) (string, error) {
	path, err := SettingsPath(ctx, conn, organizationID)
	if err != nil {
		return "", fmt.Errorf("cannot load notification settings path: %w", err)
	}

	settingsURL, err := m.baseURL.AppendPath(path).String()
	if err != nil {
		return "", fmt.Errorf("cannot build notification settings URL: %w", err)
	}

	return settingsURL, nil
}

func (m *Mailer) TaskURL(organizationID gid.GID, taskID gid.GID) (string, error) {
	path, err := employeeportal.OrganizationPath(organizationID.String(), "tasks", taskID.String())
	if err != nil {
		return "", fmt.Errorf("cannot build task path: %w", err)
	}

	taskURL, err := m.baseURL.AppendPath(path).String()
	if err != nil {
		return "", fmt.Errorf("cannot build task notification URL: %w", err)
	}

	return taskURL, nil
}

func (m *Mailer) SendTasks(
	ctx context.Context,
	tx pg.Tx,
	recipient *coredata.MembershipProfile,
	organizationName string,
	reason TaskEmailReason,
	mainURL string,
	tasks []TaskEmailItem,
) error {
	if len(tasks) == 0 || recipient == nil {
		return nil
	}

	settingsURL, err := m.SettingsURL(ctx, tx, recipient.OrganizationID)
	if err != nil {
		return fmt.Errorf("cannot build notification settings URL: %w", err)
	}

	targets := make([]string, 0, len(tasks)+2)
	targets = append(targets, mainURL, settingsURL)

	for _, task := range tasks {
		targets = append(targets, task.URL)
	}

	links, err := m.LinkTargets(ctx, tx, recipient, targets)
	if err != nil {
		return fmt.Errorf("cannot build task notification links: %w", err)
	}

	linkedMainURL := links[0]
	settingsURL = links[1]
	linkedTasks := make([]emails.TaskSummary, 0, len(tasks))

	for i, task := range tasks {
		linkedTasks = append(linkedTasks, emails.TaskSummary{
			Title: task.Title,
			URL:   links[i+2],
		})
	}

	presenter := emails.NewPresenter(m.baseURL.String(), recipient.FullName)

	var (
		subject  string
		textBody string
		htmlBody *string
	)

	switch reason {
	case TaskEmailAssigned:
		subject, textBody, htmlBody, err = presenter.RenderTaskAssigned(
			ctx,
			linkedMainURL,
			organizationName,
			linkedTasks,
			settingsURL,
		)
	case TaskEmailMentioned:
		subject, textBody, htmlBody, err = presenter.RenderTaskMentioned(
			ctx,
			linkedMainURL,
			organizationName,
			linkedTasks,
			settingsURL,
		)
	default:
		return fmt.Errorf("cannot send task notification email: unknown reason %q", reason)
	}

	if err != nil {
		return fmt.Errorf("cannot render task notification email: %w", err)
	}

	email := coredata.NewEmail(
		recipient.FullName,
		recipient.EmailAddress,
		subject,
		textBody,
		htmlBody,
		&coredata.EmailOptions{SenderName: new(organizationName)},
	)
	if err := email.Insert(ctx, tx); err != nil {
		return fmt.Errorf("cannot insert task notification email: %w", err)
	}

	return nil
}

func (m *Mailer) LinkForRecipient(
	ctx context.Context,
	tx pg.Tx,
	recipient *coredata.MembershipProfile,
	target string,
) (string, error) {
	links, err := m.LinkTargets(ctx, tx, recipient, []string{target})
	if err != nil {
		return "", fmt.Errorf("cannot build notification link: %w", err)
	}

	return links[0], nil
}

// LinkPaths builds an absolute URL for each path, then wraps them with one
// activation token when the recipient still needs to activate their account.
func (m *Mailer) LinkPaths(
	ctx context.Context,
	tx pg.Tx,
	recipient *coredata.MembershipProfile,
	paths []string,
) ([]string, error) {
	targets := make([]string, len(paths))
	for i, path := range paths {
		target, err := m.baseURL.AppendPath(path).String()
		if err != nil {
			return nil, fmt.Errorf("cannot build notification URL: %w", err)
		}

		targets[i] = target
	}

	links, err := m.LinkTargets(ctx, tx, recipient, targets)
	if err != nil {
		return nil, fmt.Errorf("cannot build notification links: %w", err)
	}

	return links, nil
}

// LinkTargets wraps every absolute URL with the same activation token.
func (m *Mailer) LinkTargets(
	ctx context.Context,
	tx pg.Tx,
	recipient *coredata.MembershipProfile,
	targets []string,
) ([]string, error) {
	if len(targets) == 0 {
		return nil, nil
	}

	token, err := m.invitationToken(ctx, tx, recipient)
	if err != nil {
		return nil, fmt.Errorf("cannot build invitation token: %w", err)
	}

	links := make([]string, len(targets))
	for i, target := range targets {
		link, err := m.urlWithToken(target, token)
		if err != nil {
			return nil, fmt.Errorf("cannot build notification link: %w", err)
		}

		links[i] = link
	}

	return links, nil
}

func (m *Mailer) urlWithToken(target string, token string) (string, error) {
	if token == "" {
		return target, nil
	}

	activationURL, err := m.baseURL.AppendPath("/auth/activate-account").
		WithQuery("token", token).
		WithQuery("continue", target).
		String()
	if err != nil {
		return "", fmt.Errorf("cannot build activation URL: %w", err)
	}

	return activationURL, nil
}

func (m *Mailer) invitationToken(
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
		ExpiresAt:      now.Add(m.invitationValidity),
		CreatedAt:      now,
	}
	if err := invitation.Insert(ctx, tx, coredata.NewScopeFromObjectID(recipient.OrganizationID)); err != nil {
		return "", fmt.Errorf("cannot insert invitation: %w", err)
	}

	token, err := statelesstoken.NewToken(
		m.tokenSecret,
		iam.TokenTypeOrganizationInvitation,
		m.invitationValidity,
		iam.InvitationTokenData{InvitationID: invitation.ID},
	)
	if err != nil {
		return "", fmt.Errorf("cannot generate invitation token: %w", err)
	}

	return token, nil
}
