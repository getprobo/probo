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

package coredata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	TaskExternalLink struct {
		OrganizationID     gid.GID                `db:"organization_id"`
		TaskID             gid.GID                `db:"task_id"`
		ConnectorID        gid.GID                `db:"connector_id"`
		Provider           ConnectorProvider      `db:"provider"`
		ExternalID         string                 `db:"external_id"`
		ExternalIdentifier string                 `db:"external_identifier"`
		ExternalURL        string                 `db:"external_url"`
		Destination        json.RawMessage        `db:"destination"`
		Origin             TaskExternalLinkOrigin `db:"origin"`
		RemoteUpdatedAt    *time.Time             `db:"remote_updated_at"`
		ContentHash        *string                `db:"content_hash"`
		Metadata           json.RawMessage        `db:"metadata"`
		CreatedAt          time.Time              `db:"created_at"`
		UpdatedAt          time.Time              `db:"updated_at"`
	}

	TaskExternalLinks []*TaskExternalLink

	TaskExternalLinkDestination struct {
		TeamID               string  `json:"team_id"`
		ProjectID            *string `json:"project_id,omitempty"`
		LinearOrganizationID string  `json:"linear_organization_id,omitempty"`
		PreserveLinearNone   bool    `json:"preserve_linear_none,omitempty"`
		LinearNoneSnapshot   string  `json:"linear_none_snapshot,omitempty"`
	}
)

func (l *TaskExternalLink) DestinationTeamID() (string, error) {
	var dest TaskExternalLinkDestination
	if err := json.Unmarshal(l.Destination, &dest); err != nil {
		return "", fmt.Errorf("cannot unmarshal destination: %w", err)
	}

	return dest.TeamID, nil
}

func (l *TaskExternalLink) DestinationLinearOrganizationID() string {
	var dest TaskExternalLinkDestination
	if err := json.Unmarshal(l.Destination, &dest); err != nil {
		return ""
	}

	return dest.LinearOrganizationID
}

func (l *TaskExternalLink) LoadByTaskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskID gid.GID,
) error {
	q := `
SELECT
    organization_id,
    task_id,
    connector_id,
    provider,
    external_id,
    external_identifier,
    external_url,
    destination,
    origin,
    remote_updated_at,
    content_hash,
    metadata,
    created_at,
    updated_at
FROM
    task_external_links
WHERE
    %s
    AND task_id = @task_id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"task_id": taskID}
	maps.Copy(args, scope.SQLArguments())

	return l.loadExactlyOne(ctx, conn, q, args)
}

// LoadByTaskIDForUpdate is LoadByTaskID under FOR UPDATE so concurrent
// inbound webhooks cannot apply an older Linear event after a newer one.
func (l *TaskExternalLink) LoadByTaskIDForUpdate(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	taskID gid.GID,
) error {
	q := `
SELECT
    organization_id,
    task_id,
    connector_id,
    provider,
    external_id,
    external_identifier,
    external_url,
    destination,
    origin,
    remote_updated_at,
    content_hash,
    metadata,
    created_at,
    updated_at
FROM
    task_external_links
WHERE
    %s
    AND task_id = @task_id
LIMIT 1
FOR UPDATE;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"task_id": taskID}
	maps.Copy(args, scope.SQLArguments())

	return l.loadExactlyOne(ctx, conn, q, args)
}

func (l *TaskExternalLinks) LoadByExternalID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	provider ConnectorProvider,
	externalID string,
) error {
	q := `
SELECT
    organization_id,
    task_id,
    connector_id,
    provider,
    external_id,
    external_identifier,
    external_url,
    destination,
    origin,
    remote_updated_at,
    content_hash,
    metadata,
    created_at,
    updated_at
FROM
    task_external_links
WHERE
    %s
    AND provider = @provider
    AND external_id = @external_id
ORDER BY
    created_at ASC,
    task_id ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"provider":    provider,
		"external_id": externalID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query task external links: %w", err)
	}

	links, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TaskExternalLink])
	if err != nil {
		return fmt.Errorf("cannot collect task external links: %w", err)
	}

	*l = links

	return nil
}

func (l *TaskExternalLinks) LoadByTaskIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskIDs []gid.GID,
) error {
	if len(taskIDs) == 0 {
		*l = nil
		return nil
	}

	q := `
SELECT
    organization_id,
    task_id,
    connector_id,
    provider,
    external_id,
    external_identifier,
    external_url,
    destination,
    origin,
    remote_updated_at,
    content_hash,
    metadata,
    created_at,
    updated_at
FROM
    task_external_links
WHERE
    %s
    AND task_id = ANY(@task_ids)
ORDER BY
    created_at ASC,
    task_id ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"task_ids": taskIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query task external links: %w", err)
	}

	links, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TaskExternalLink])
	if err != nil {
		return fmt.Errorf("cannot collect task external links: %w", err)
	}

	*l = links

	return nil
}

func (l *TaskExternalLink) Insert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO task_external_links (
    task_id,
    tenant_id,
    organization_id,
    connector_id,
    provider,
    external_id,
    external_identifier,
    external_url,
    destination,
    origin,
    remote_updated_at,
    content_hash,
    metadata,
    created_at,
    updated_at
) VALUES (
    @task_id,
    @tenant_id,
    @organization_id,
    @connector_id,
    @provider,
    @external_id,
    @external_identifier,
    @external_url,
    @destination,
    @origin,
    @remote_updated_at,
    @content_hash,
    @metadata,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"task_id":             l.TaskID,
		"tenant_id":           scope.GetTenantID(),
		"organization_id":     l.OrganizationID,
		"connector_id":        l.ConnectorID,
		"provider":            l.Provider,
		"external_id":         l.ExternalID,
		"external_identifier": l.ExternalIdentifier,
		"external_url":        l.ExternalURL,
		"destination":         l.Destination,
		"origin":              l.Origin,
		"remote_updated_at":   l.RemoteUpdatedAt,
		"content_hash":        l.ContentHash,
		"metadata":            l.Metadata,
		"created_at":          l.CreatedAt,
		"updated_at":          l.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "task_external_links_pkey",
					"task_external_links_organization_id_provider_external_id_key":
					return ErrResourceAlreadyExists
				}
			}
		}

		return fmt.Errorf("cannot insert task external link: %w", err)
	}

	return nil
}

func (l *TaskExternalLink) Update(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
UPDATE task_external_links
SET
    external_id = @external_id,
    external_identifier = @external_identifier,
    external_url = @external_url,
    destination = @destination,
    remote_updated_at = @remote_updated_at,
    content_hash = @content_hash,
    metadata = @metadata,
    updated_at = @updated_at
WHERE
    %s
    AND task_id = @task_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"task_id":             l.TaskID,
		"external_id":         l.ExternalID,
		"external_identifier": l.ExternalIdentifier,
		"external_url":        l.ExternalURL,
		"destination":         l.Destination,
		"remote_updated_at":   l.RemoteUpdatedAt,
		"content_hash":        l.ContentHash,
		"metadata":            l.Metadata,
		"updated_at":          l.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot update task external link: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (l *TaskExternalLink) Delete(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
DELETE FROM task_external_links
WHERE
    %s
    AND task_id = @task_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"task_id": l.TaskID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete task external link: %w", err)
	}

	return nil
}

func (l *TaskExternalLink) loadExactlyOne(
	ctx context.Context,
	conn pg.Querier,
	q string,
	args pgx.StrictNamedArgs,
) error {
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query task external link: %w", err)
	}

	link, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TaskExternalLink])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect task external link: %w", err)
	}

	*l = link

	return nil
}
