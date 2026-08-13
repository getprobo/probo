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

type BotThreadSubject struct {
	ID                     gid.GID         `db:"id"`
	OrganizationID         gid.GID         `db:"organization_id"`
	Provider               string          `db:"provider"`
	ExternalConversationID string          `db:"external_conversation_id"`
	ExternalMessageID      string          `db:"external_message_id"`
	Capability             string          `db:"capability"`
	MessageType            string          `db:"message_type"`
	Attributes             json.RawMessage `db:"attributes"`
	SubjectNamespace       string          `db:"subject_namespace"`
	SubjectKey             string          `db:"subject_key"`
	CreatedAt              time.Time       `db:"created_at"`
	UpdatedAt              time.Time       `db:"updated_at"`
}

func (s *BotThreadSubject) Upsert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) (bool, error) {
	q := `
INSERT INTO bot_thread_subjects (
	id,
	tenant_id,
	organization_id,
	provider,
	external_conversation_id,
	external_message_id,
	capability,
	message_type,
	attributes,
	subject_namespace,
	subject_key,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@provider,
	@external_conversation_id,
	@external_message_id,
	@capability,
	@message_type,
	@attributes,
	@subject_namespace,
	@subject_key,
	@created_at,
	@updated_at
)
ON CONFLICT (tenant_id, organization_id, provider, external_conversation_id, external_message_id)
DO UPDATE SET
	capability = EXCLUDED.capability,
	message_type = EXCLUDED.message_type,
	attributes = EXCLUDED.attributes,
	subject_namespace = EXCLUDED.subject_namespace,
	subject_key = EXCLUDED.subject_key,
	updated_at = EXCLUDED.updated_at
RETURNING
	id,
	organization_id,
	provider,
	external_conversation_id,
	external_message_id,
	capability,
	message_type,
	attributes,
	subject_namespace,
	subject_key,
	created_at,
	updated_at
`

	if s.Attributes == nil {
		s.Attributes = json.RawMessage("{}")
	}

	originalID := s.ID
	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"id":                       s.ID,
			"tenant_id":                scope.GetTenantID(),
			"organization_id":          s.OrganizationID,
			"provider":                 s.Provider,
			"external_conversation_id": s.ExternalConversationID,
			"external_message_id":      s.ExternalMessageID,
			"capability":               s.Capability,
			"message_type":             s.MessageType,
			"attributes":               s.Attributes,
			"subject_namespace":        s.SubjectNamespace,
			"subject_key":              s.SubjectKey,
			"created_at":               s.CreatedAt,
			"updated_at":               s.UpdatedAt,
		},
	)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "bot_thread_subjects_subject" {
				return false, ErrResourceAlreadyExists
			}
		}

		return false, fmt.Errorf("cannot upsert bot thread subject: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[BotThreadSubject])
	if err != nil {
		return false, fmt.Errorf("cannot collect bot thread subject: %w", err)
	}

	*s = row

	return originalID == s.ID, nil
}

func (s *BotThreadSubject) LoadByProviderCoordinates(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	provider string,
	externalConversationID string,
	externalMessageID string,
) error {
	q := `
SELECT
	id,
	organization_id,
	provider,
	external_conversation_id,
	external_message_id,
	capability,
	message_type,
	attributes,
	subject_namespace,
	subject_key,
	created_at,
	updated_at
FROM bot_thread_subjects
WHERE %s
	AND organization_id = @organization_id
	AND provider = @provider
	AND external_conversation_id = @external_conversation_id
	AND external_message_id = @external_message_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id":          organizationID,
		"provider":                 provider,
		"external_conversation_id": externalConversationID,
		"external_message_id":      externalMessageID,
	}
	maps.Copy(args, scope.SQLArguments())

	return s.loadExactlyOne(
		ctx,
		conn,
		q,
		args,
	)
}

func (s *BotThreadSubject) LoadBySubject(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	subjectNamespace string,
	subjectKey string,
) error {
	q := `
SELECT
	id,
	organization_id,
	provider,
	external_conversation_id,
	external_message_id,
	capability,
	message_type,
	attributes,
	subject_namespace,
	subject_key,
	created_at,
	updated_at
FROM bot_thread_subjects
WHERE %s
	AND organization_id = @organization_id
	AND subject_namespace = @subject_namespace
	AND subject_key = @subject_key
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id":   organizationID,
		"subject_namespace": subjectNamespace,
		"subject_key":       subjectKey,
	}
	maps.Copy(args, scope.SQLArguments())

	return s.loadExactlyOne(
		ctx,
		conn,
		q,
		args,
	)
}

func (s *BotThreadSubject) loadExactlyOne(
	ctx context.Context,
	conn pg.Querier,
	q string,
	args pgx.StrictNamedArgs,
) error {
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query bot thread subject: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[BotThreadSubject])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect bot thread subject: %w", err)
	}

	*s = row

	return nil
}
