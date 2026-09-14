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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/tasksync/linear"
)

type (
	Service struct {
		pg                *pg.Client
		encryptionKey     cipher.EncryptionKey
		connectorRegistry *connector.Registry
		baseURL           string
		// linearAPIBaseURL is the LINEAR provider registration's
		// Endpoints.APIBase, threaded in by probod so a deployment that
		// repoints the Linear connector moves these calls too. See
		// linear.NewClient.
		linearAPIBaseURL string
		logger           *log.Logger
	}

	LinearTeam struct {
		ID   string
		Name string
		Key  string
	}
)

func NewService(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	connectorRegistry *connector.Registry,
	baseURL string,
	linearAPIBaseURL string,
	logger *log.Logger,
) *Service {
	return &Service{
		pg:                pgClient,
		encryptionKey:     encryptionKey,
		connectorRegistry: connectorRegistry,
		baseURL:           baseURL,
		linearAPIBaseURL:  linearAPIBaseURL,
		logger:            logger,
	}
}

func (s *Service) GetLinkByTaskID(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
) (*coredata.TaskExternalLink, error) {
	link := &coredata.TaskExternalLink{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := link.LoadByTaskID(ctx, conn, scope, taskID); err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (s *Service) ListLinearTeams(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) ([]LinearTeam, error) {
	var client *linear.Client

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var err error

			client, _, err = s.linearClientForOrganization(ctx, tx, scope, organizationID)

			return err
		},
	)
	if err != nil {
		return nil, err
	}

	remoteTeams, err := client.ListTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot list Linear teams: %w", err)
	}

	teams := make([]LinearTeam, 0, len(remoteTeams))
	for _, team := range remoteTeams {
		teams = append(teams, LinearTeam{
			ID:   team.ID,
			Name: team.Name,
			Key:  team.Key,
		})
	}

	return teams, nil
}

func (s *Service) PublishToLinear(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
	teamID string,
) (*coredata.TaskExternalLink, error) {
	if teamID == "" {
		return nil, fmt.Errorf("cannot publish task: team id is required")
	}

	var (
		task        *coredata.Task
		client      *linear.Client
		dbConnector *coredata.Connector
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			task = &coredata.Task{}
			if err := task.LoadByID(ctx, tx, scope, taskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", taskID, err)
			}

			if err := ensureTaskNotLinked(ctx, tx, scope, taskID); err != nil {
				return err
			}

			var err error

			client, dbConnector, err = s.linearClientForOrganization(ctx, tx, scope, task.OrganizationID)

			return err
		},
	)
	if err != nil {
		return nil, err
	}

	states, err := client.ListWorkflowStates(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("cannot list Linear workflow states: %w", err)
	}

	stateID, err := PickWorkflowStateID(states, task.State)
	if err != nil {
		return nil, err
	}

	markdown, err := ContentToMarkdown(task.Content)
	if err != nil {
		return nil, err
	}

	viewerID, err := client.ViewerID(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load Linear viewer: %w", err)
	}

	linearOrganizationID, err := client.OrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load Linear organization: %w", err)
	}

	issue, err := client.CreateIssue(
		ctx,
		linear.IssueInput{
			TeamID:      teamID,
			Title:       task.Name,
			Description: markdown,
			StateID:     stateID,
			Priority:    TaskPriorityToLinear(task.Priority),
			DueDate:     DeadlineToLinearDate(task.Deadline),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create Linear issue: %w", err)
	}

	taskURL, err := s.taskURL(task.OrganizationID, task.ID)
	if err != nil {
		return nil, err
	}

	attachmentID, err := client.LinkAttachment(ctx, issue.ID, taskURL, "Probo task")
	if err != nil {
		return nil, fmt.Errorf("cannot link Linear attachment: %w", err)
	}

	destination, err := json.Marshal(
		coredata.TaskExternalLinkDestination{
			TeamID:               teamID,
			LinearOrganizationID: linearOrganizationID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal destination: %w", err)
	}

	metadata, err := json.Marshal(
		map[string]string{
			"attachment_id": attachmentID,
			"app_actor_id":  viewerID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal metadata: %w", err)
	}

	hash := ContentHash(task.Name, markdown, task.State, task.Priority, task.Deadline)
	now := time.Now()
	link := &coredata.TaskExternalLink{
		OrganizationID:     task.OrganizationID,
		TaskID:             task.ID,
		ConnectorID:        dbConnector.ID,
		Provider:           coredata.ConnectorProviderLinear,
		ExternalID:         issue.ID,
		ExternalIdentifier: issue.Identifier,
		ExternalURL:        issue.URL,
		Destination:        destination,
		Origin:             coredata.TaskExternalLinkOriginProbo,
		RemoteUpdatedAt:    &issue.UpdatedAt,
		ContentHash:        &hash,
		Metadata:           metadata,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := ensureTaskNotLinked(ctx, tx, scope, taskID); err != nil {
				return err
			}

			if err := link.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert task external link: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (s *Service) Unlink(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			link := &coredata.TaskExternalLink{}
			if err := link.LoadByTaskID(ctx, tx, scope, taskID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrTaskNotLinked
				}

				return fmt.Errorf("cannot load task external link: %w", err)
			}

			if err := link.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete task external link: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) EnqueueOutbound(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	taskID gid.GID,
	action SyncAction,
) error {
	link := &coredata.TaskExternalLink{}
	if err := link.LoadByTaskID(ctx, tx, scope, taskID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil
		}

		return fmt.Errorf("cannot load task external link: %w", err)
	}

	return s.enqueueOutboundTx(ctx, tx, scope, link, action)
}

func (s *Service) enqueueOutboundTx(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	link *coredata.TaskExternalLink,
	action SyncAction,
) error {
	teamID, err := link.DestinationTeamID()
	if err != nil {
		return err
	}

	payload, err := json.Marshal(
		JobPayload{
			Action:             action,
			TaskID:             link.TaskID,
			ExternalID:         link.ExternalID,
			ExternalIdentifier: link.ExternalIdentifier,
			TeamID:             teamID,
			ConnectorID:        link.ConnectorID,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot marshal task sync payload: %w", err)
	}

	now := time.Now()
	job := &coredata.TaskSyncJob{
		ID:             gid.New(scope.GetTenantID(), coredata.TaskSyncJobEntityType),
		OrganizationID: link.OrganizationID,
		Direction:      coredata.TaskSyncJobDirectionOutbound,
		Status:         coredata.TaskSyncJobStatusPending,
		Payload:        payload,
		CreatedAt:      now,
		UpdatedAt:      now,
		AttemptCount:   0,
	}

	if err := job.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert task sync job: %w", err)
	}

	return nil
}

func ensureTaskNotLinked(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	taskID gid.GID,
) error {
	existing := &coredata.TaskExternalLink{}

	err := existing.LoadByTaskID(ctx, conn, scope, taskID)
	if err == nil {
		return ErrTaskAlreadyLinked
	}

	if !errors.Is(err, coredata.ErrResourceNotFound) {
		return fmt.Errorf("cannot load task external link: %w", err)
	}

	return nil
}

func (s *Service) linearClientForOrganization(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*linear.Client, *coredata.Connector, error) {
	dbConnector := &coredata.Connector{}
	if err := dbConnector.LoadByOrganizationIDAndProvider(
		ctx,
		tx,
		scope,
		organizationID,
		coredata.ConnectorProviderLinear,
		s.encryptionKey,
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil, ErrLinearNotConnected
		}

		return nil, nil, fmt.Errorf("cannot load Linear connector: %w", err)
	}

	return s.linearClientForConnector(ctx, tx, scope, dbConnector)
}

func (s *Service) linearClientForConnector(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	dbConnector *coredata.Connector,
) (*linear.Client, *coredata.Connector, error) {
	oauthConn, ok := dbConnector.Connection.(*connector.OAuth2Connection)
	if !ok {
		return nil, nil, fmt.Errorf("cannot use Linear connector: unsupported credential")
	}

	if missing := missingTaskSyncScopes(oauthConn.Scopes()); len(missing) > 0 {
		return nil, nil, NewLinearReconnectRequiredError(missing)
	}

	if err := s.connectorRegistry.ConfigureConnection(string(dbConnector.Provider), oauthConn); err != nil {
		return nil, nil, fmt.Errorf("cannot configure Linear connector: %w", err)
	}

	tokenBefore := oauthConn.AccessToken

	httpClient, err := s.oauthHTTPClient(ctx, oauthConn, dbConnector.Provider)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create Linear HTTP client: %w", err)
	}

	if oauthConn.AccessToken != tokenBefore {
		dbConnector.UpdatedAt = time.Now()
		if err := dbConnector.Update(ctx, tx, scope, s.encryptionKey); err != nil {
			return nil, nil, fmt.Errorf("cannot persist refreshed Linear token: %w", err)
		}
	}

	return linear.NewClient(httpClient, s.linearAPIBaseURL), dbConnector, nil
}

func (s *Service) oauthHTTPClient(
	ctx context.Context,
	conn *connector.OAuth2Connection,
	provider coredata.ConnectorProvider,
) (*http.Client, error) {
	refreshCfg := s.connectorRegistry.GetOAuth2RefreshConfig(string(provider))
	if refreshCfg != nil {
		return conn.RefreshableClient(ctx, *refreshCfg)
	}

	return conn.Client(ctx)
}

func (s *Service) taskURL(organizationID, taskID gid.GID) (string, error) {
	u, err := url.JoinPath(
		s.baseURL,
		"organizations",
		url.PathEscape(organizationID.String()),
		"governance",
		"tasks",
		url.PathEscape(taskID.String()),
	)
	if err != nil {
		return "", fmt.Errorf("cannot build task URL: %w", err)
	}

	return u, nil
}
