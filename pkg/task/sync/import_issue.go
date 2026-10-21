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

package tasksync

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

const importedTaskNameMaxLen = 1000

func isLinearApplicationActor(actor linear.WebhookActor) bool {
	return strings.EqualFold(actor.Type, "application")
}

type preparedImportedIssue struct {
	connector *coredata.Connector
	content   string
	markdown  string
	files     map[string]string
	created   []createdInboundFile
}

func (s *Service) importCreatedLinearIssue(
	ctx context.Context,
	tx pg.Tx,
	envelope *linear.WebhookEnvelope,
	data *linear.IssueWebhookData,
	prepared []preparedImportedIssue,
) error {
	if isLinearApplicationActor(envelope.Actor) {
		return nil
	}

	if data.Team.ID == "" || envelope.OrganizationID == "" {
		return nil
	}

	if prepared != nil {
		for i := range prepared {
			if err := s.importLinearIssueForConnector(
				ctx,
				tx,
				envelope,
				data,
				prepared[i].connector,
				&prepared[i],
			); err != nil {
				return err
			}
		}

		return nil
	}

	connectors := coredata.Connectors{}
	if err := connectors.LoadByProviderAndSettings(
		ctx,
		tx,
		coredata.ConnectorProviderLinearSync,
		map[string]any{
			"import_new_issues":      true,
			"default_team_id":        data.Team.ID,
			"linear_organization_id": envelope.OrganizationID,
		},
	); err != nil {
		return fmt.Errorf("cannot load Linear issue import connectors: %w", err)
	}

	for _, connector := range connectors {
		if err := s.importLinearIssueForConnector(ctx, tx, envelope, data, connector, nil); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) importLinearIssueForConnector(
	ctx context.Context,
	tx pg.Tx,
	envelope *linear.WebhookEnvelope,
	data *linear.IssueWebhookData,
	connector *coredata.Connector,
	ready *preparedImportedIssue,
) error {
	settings, err := coredata.ConnectorSettings[coredata.LinearSyncConnectorSettings](connector)
	if err != nil {
		return fmt.Errorf("cannot read Linear settings: %w", err)
	}

	name := importedTaskName(data)
	if name == "" {
		return nil
	}

	scope := coredata.NewScope(connector.OrganizationID.TenantID())

	var (
		content  string
		markdown string
		files    map[string]string
	)

	if ready != nil {
		content = ready.content
		markdown = ready.markdown
		files = ready.files
	} else {
		var importErr error

		content, markdown, files, _, importErr = s.importedTaskContent(ctx, scope, connector, data)
		if importErr != nil {
			return importErr
		}
	}

	now := time.Now()

	referenceID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("cannot generate reference id: %w", err)
	}

	state := coredata.TaskStateTodo
	if data.Has("state") {
		state = LinearTypeToTaskState(data.State.Type)
	}

	priority := coredata.TaskPriorityMedium

	if data.Has("priority") {
		if mapped, ok := LinearPriorityToTask(data.Priority); ok {
			priority = mapped
		}
	}

	var deadline *time.Time
	if data.Has("dueDate") {
		deadline = LinearDateToDeadline(data.DueDate)
	}

	task := &coredata.Task{
		ID:             gid.New(scope.GetTenantID(), coredata.TaskEntityType),
		OrganizationID: connector.OrganizationID,
		Name:           name,
		Content:        content,
		State:          state,
		Priority:       priority,
		Deadline:       deadline,
		ReferenceID:    "custom-task-" + referenceID.String(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if _, err := applyInboundAssignee(ctx, tx, scope, connector.OrganizationID, task, data); err != nil {
		return fmt.Errorf("cannot apply Linear assignee: %w", err)
	}

	if err := task.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert imported task: %w", err)
	}

	destination, err := json.Marshal(coredata.TaskExternalLinkDestination{
		TeamID:               data.Team.ID,
		LinearOrganizationID: envelope.OrganizationID,
	})
	if err != nil {
		return fmt.Errorf("cannot marshal destination: %w", err)
	}

	metadata := json.RawMessage(`{}`)

	if settings.LinearAppActorID != "" {
		encoded, err := json.Marshal(map[string]string{
			"app_actor_id": settings.LinearAppActorID,
		})
		if err != nil {
			return fmt.Errorf("cannot marshal Linear link metadata: %w", err)
		}

		metadata = encoded
	}

	if len(files) > 0 {
		metadata, err = mergeAttachments(metadata, files)
		if err != nil {
			return fmt.Errorf("cannot merge Linear attachments: %w", err)
		}
	}

	hash := ContentHash(
		task.Name,
		markdown,
		task.State,
		task.Priority,
		task.Deadline,
		task.AssignedToID,
	)
	remoteUpdatedAt := linearDateTime(data.UpdatedAt)
	link := &coredata.TaskExternalLink{
		OrganizationID:     connector.OrganizationID,
		TaskID:             task.ID,
		ConnectorID:        connector.ID,
		Provider:           coredata.ConnectorProviderLinearSync,
		ExternalID:         data.ID,
		ExternalIdentifier: data.Identifier,
		ExternalURL:        data.URL,
		Destination:        destination,
		Origin:             coredata.TaskExternalLinkOriginLinear,
		RemoteUpdatedAt:    remoteUpdatedAt,
		ContentHash:        &hash,
		Metadata:           metadata,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := link.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert imported task link: %w", err)
	}

	if s.recordImportedTask == nil {
		return nil
	}

	actorID, err := inboundActivityActorID(ctx, tx, scope, connector.OrganizationID, envelope.Actor)
	if err != nil {
		return fmt.Errorf("cannot resolve Linear activity actor: %w", err)
	}

	actorIdentityID, err := linearActorIdentityID(ctx, tx, scope, connector.OrganizationID, envelope.Actor.Email)
	if err != nil {
		return fmt.Errorf("cannot match Linear actor: %w", err)
	}

	if err := s.recordImportedTask(ctx, tx, scope, task, actorID, actorIdentityID, now); err != nil {
		return fmt.Errorf("cannot record imported task: %w", err)
	}

	return nil
}

func (s *Service) prepareImportedIssues(
	ctx context.Context,
	envelope *linear.WebhookEnvelope,
	data *linear.IssueWebhookData,
) ([]preparedImportedIssue, error) {
	if isLinearApplicationActor(envelope.Actor) || data.Team.ID == "" || envelope.OrganizationID == "" {
		return []preparedImportedIssue{}, nil
	}

	if s.pg == nil {
		return nil, fmt.Errorf("cannot load Linear issue import connectors: database is not configured")
	}

	connectors := coredata.Connectors{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := connectors.LoadByProviderAndSettings(
				ctx,
				conn,
				coredata.ConnectorProviderLinearSync,
				map[string]any{
					"import_new_issues":      true,
					"default_team_id":        data.Team.ID,
					"linear_organization_id": envelope.OrganizationID,
				},
			); err != nil {
				return fmt.Errorf("cannot load Linear issue import connectors: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	prepared := make([]preparedImportedIssue, 0, len(connectors))

	for _, connector := range connectors {
		scope := coredata.NewScope(connector.OrganizationID.TenantID())

		content, markdown, files, created, err := s.importedTaskContent(ctx, scope, connector, data)
		if err != nil {
			return nil, err
		}

		prepared = append(prepared, preparedImportedIssue{
			connector: connector,
			content:   content,
			markdown:  markdown,
			files:     files,
			created:   created,
		})
	}

	return prepared, nil
}

func importedTaskName(data *linear.IssueWebhookData) string {
	name := strings.Join(strings.Fields(data.Title), " ")
	if name == "" {
		name = strings.TrimSpace(data.Identifier)
	}

	if utf8.RuneCountInString(name) <= importedTaskNameMaxLen {
		return name
	}

	runes := []rune(name)

	return string(runes[:importedTaskNameMaxLen])
}

func (s *Service) importedTaskContent(
	ctx context.Context,
	scope coredata.Scoper,
	connector *coredata.Connector,
	data *linear.IssueWebhookData,
) (string, string, map[string]string, []createdInboundFile, error) {
	description := ""

	var (
		files   map[string]string
		metas   map[string]importedFileMeta
		created []createdInboundFile
	)

	if data.Has("description") {
		rewritten, imported, stored, storedFiles, err := s.importLinearFiles(
			ctx,
			nil,
			scope,
			connector.ID,
			connector.OrganizationID,
			data.Description,
			nil,
			true,
		)
		if err != nil {
			return "", "", nil, nil, fmt.Errorf("cannot import Linear description files: %w", err)
		}

		description = rewritten
		files = imported
		metas = stored
		created = storedFiles
	}

	content, err := MarkdownToContent(description)
	if err != nil {
		return "", "", nil, nil, err
	}

	content, markdown, err := applyImportedFiles(content, metas)
	if err != nil {
		return "", "", nil, nil, err
	}

	return content, markdown, files, created, nil
}
