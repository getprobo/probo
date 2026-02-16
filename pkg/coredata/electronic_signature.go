// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package coredata

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/crypto/hash"
	"go.probo.inc/probo/pkg/gid"
)

type ElectronicSignature struct {
	ID                             gid.GID                         `db:"id"`
	TenantID                       gid.TenantID                    `db:"tenant_id"`
	OrganizationID                 gid.GID                         `db:"organization_id"`
	Status                         ElectronicSignatureStatus       `db:"status"`
	DocumentType                   ElectronicSignatureDocumentType `db:"document_type"`
	FileID                         gid.GID                         `db:"file_id"`
	SignerEmail                    string                          `db:"signer_email"`
	ConsentText                    string                          `db:"consent_text"`
	SignerFullName                 *string                         `db:"signer_full_name"`
	SignerIPAddress                *string                         `db:"signer_ip_address"`
	SignerUserAgent                *string                         `db:"signer_user_agent"`
	FileHash                       *string                         `db:"file_hash"`
	Seal                           *string                         `db:"seal"`
	SealVersion                    int                             `db:"seal_version"`
	TSAToken                       []byte                          `db:"tsa_token"`
	SignedAt                       *time.Time                      `db:"signed_at"`
	CertificateFileID              *gid.GID                        `db:"certificate_file_id"`
	CertificateProcessingStartedAt *time.Time                      `db:"certificate_processing_started_at"`
	AttemptCount                   int                             `db:"attempt_count"`
	MaxAttempts                    int                             `db:"max_attempts"`
	LastAttemptedAt                *time.Time                      `db:"last_attempted_at"`
	LastError                      *string                         `db:"last_error"`
	ProcessingStartedAt            *time.Time                      `db:"processing_started_at"`
	CreatedAt                      time.Time                       `db:"created_at"`
	UpdatedAt                      time.Time                       `db:"updated_at"`
}

func (es *ElectronicSignature) NewEvent(
	eventType ElectronicSignatureEventType,
	eventSource ElectronicSignatureEventSource,
) ElectronicSignatureEvent {
	now := time.Now()

	return ElectronicSignatureEvent{
		ID:                    gid.New(es.ID.TenantID(), ElectronicSignatureEventEntityType),
		ElectronicSignatureID: es.ID,
		EventType:             eventType,
		EventSource:           eventSource,
		ActorEmail:            es.SignerEmail,
		ActorIPAddress:        ref.UnrefOrZero(es.SignerIPAddress),
		ActorUserAgent:        ref.UnrefOrZero(es.SignerUserAgent),
		OccurredAt:            now,
		CreatedAt:             now,
	}
}

func (es *ElectronicSignature) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO electronic_signatures (
	id, tenant_id, organization_id, status, document_type, file_id,
	signer_email, consent_text, seal_version, attempt_count, max_attempts,
	created_at, updated_at
) VALUES (
	@id, @tenant_id, @organization_id, @status, @document_type, @file_id,
	@signer_email, @consent_text, @seal_version, @attempt_count, @max_attempts,
	@created_at, @updated_at
)
`
	args := pgx.StrictNamedArgs{
		"id":              es.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": es.OrganizationID,
		"status":          es.Status,
		"document_type":   es.DocumentType,
		"file_id":         es.FileID,
		"signer_email":    es.SignerEmail,
		"consent_text":    es.ConsentText,
		"seal_version":    es.SealVersion,
		"attempt_count":   es.AttemptCount,
		"max_attempts":    es.MaxAttempts,
		"created_at":      es.CreatedAt,
		"updated_at":      es.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert electronic signature: %w", err)
	}

	return nil
}

func (es *ElectronicSignature) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE electronic_signatures SET
	status = @status,
	signer_full_name = @signer_full_name,
	signer_ip_address = @signer_ip_address,
	signer_user_agent = @signer_user_agent,
	file_hash = @file_hash,
	seal = @seal,
	seal_version = @seal_version,
	tsa_token = @tsa_token,
	signed_at = @signed_at,
	certificate_file_id = @certificate_file_id,
	certificate_processing_started_at = @certificate_processing_started_at,
	attempt_count = @attempt_count,
	max_attempts = @max_attempts,
	last_attempted_at = @last_attempted_at,
	last_error = @last_error,
	processing_started_at = @processing_started_at,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                                es.ID,
		"status":                            es.Status,
		"signer_full_name":                  es.SignerFullName,
		"signer_ip_address":                 es.SignerIPAddress,
		"signer_user_agent":                 es.SignerUserAgent,
		"file_hash":                         es.FileHash,
		"seal":                              es.Seal,
		"seal_version":                      es.SealVersion,
		"tsa_token":                         es.TSAToken,
		"signed_at":                         es.SignedAt,
		"certificate_file_id":               es.CertificateFileID,
		"certificate_processing_started_at": es.CertificateProcessingStartedAt,
		"attempt_count":                     es.AttemptCount,
		"max_attempts":                      es.MaxAttempts,
		"last_attempted_at":                 es.LastAttemptedAt,
		"last_error":                        es.LastError,
		"processing_started_at":             es.ProcessingStartedAt,
		"updated_at":                        es.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update electronic signature: %w", err)
	}

	return nil
}

func (es *ElectronicSignature) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id, tenant_id, organization_id, status, document_type, file_id,
	signer_email, consent_text, signer_full_name, signer_ip_address,
	signer_user_agent, file_hash, seal, seal_version, tsa_token, signed_at,
	certificate_file_id, certificate_processing_started_at,
	attempt_count, max_attempts, last_attempted_at, last_error,
	processing_started_at, created_at, updated_at
FROM electronic_signatures
WHERE %s AND id = @id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query electronic signature: %w", err)
	}

	sig, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ElectronicSignature])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect electronic signature: %w", err)
	}

	*es = sig
	return nil
}

func (es *ElectronicSignature) LoadByOrgEmailAndDocType(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	orgID gid.GID,
	email string,
	docType ElectronicSignatureDocumentType,
	fileID gid.GID,
) error {
	q := `
SELECT
	id, tenant_id, organization_id, status, document_type, file_id,
	signer_email, consent_text, signer_full_name, signer_ip_address,
	signer_user_agent, file_hash, seal, seal_version, tsa_token, signed_at,
	certificate_file_id, certificate_processing_started_at,
	attempt_count, max_attempts, last_attempted_at, last_error,
	processing_started_at, created_at, updated_at
FROM electronic_signatures
WHERE %s
	AND organization_id = @organization_id
	AND signer_email = @signer_email
	AND document_type = @document_type
	AND file_id = @file_id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": orgID,
		"signer_email":    email,
		"document_type":   docType,
		"file_id":         fileID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query electronic signature: %w", err)
	}

	sig, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ElectronicSignature])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect electronic signature: %w", err)
	}

	*es = sig
	return nil
}

func (es *ElectronicSignature) LoadNextAcceptedForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
SELECT
	id, tenant_id, organization_id, status, document_type, file_id,
	signer_email, consent_text, signer_full_name, signer_ip_address,
	signer_user_agent, file_hash, seal, seal_version, tsa_token, signed_at,
	certificate_file_id, certificate_processing_started_at,
	attempt_count, max_attempts, last_attempted_at, last_error,
	processing_started_at, created_at, updated_at
FROM electronic_signatures
WHERE status = 'ACCEPTED' AND attempt_count < max_attempts
ORDER BY updated_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query accepted signatures: %w", err)
	}

	sig, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ElectronicSignature])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect electronic signature: %w", err)
	}

	*es = sig
	return nil
}

func (es *ElectronicSignature) LoadNextCompletedWithoutCertificateForUpdate(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
SELECT
	id, tenant_id, organization_id, status, document_type, file_id,
	signer_email, consent_text, signer_full_name, signer_ip_address,
	signer_user_agent, file_hash, seal, seal_version, tsa_token, signed_at,
	certificate_file_id, certificate_processing_started_at,
	attempt_count, max_attempts, last_attempted_at, last_error,
	processing_started_at, created_at, updated_at
FROM electronic_signatures
WHERE status = 'COMPLETED'
	AND certificate_file_id IS NULL
	AND certificate_processing_started_at IS NULL
ORDER BY signed_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query completed signatures: %w", err)
	}

	sig, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ElectronicSignature])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect electronic signature: %w", err)
	}

	*es = sig
	return nil
}

func ResetStaleProcessingSignatures(
	ctx context.Context,
	conn pg.Conn,
	staleAfter time.Duration,
) error {
	q := `
UPDATE electronic_signatures
SET status = 'ACCEPTED', processing_started_at = NULL, updated_at = NOW()
WHERE status = 'PROCESSING'
	AND processing_started_at < NOW() - $1::interval
`
	_, err := conn.Exec(ctx, q, staleAfter.String())
	if err != nil {
		return fmt.Errorf("cannot reset stale processing signatures: %w", err)
	}

	return nil
}

func (es *ElectronicSignature) ComputeSeal(version int) (string, error) {
	switch version {
	case 1:
		return es.computeSealV1()
	default:
		return "", fmt.Errorf("unsupported seal version %d", version)
	}
}

func (es *ElectronicSignature) computeSealV1() (string, error) {
	if es.SignedAt == nil {
		return "", fmt.Errorf("signed_at must not be nil")
	}

	fields := []string{
		es.ID.String(),
		es.OrganizationID.String(),
		es.DocumentType.String(),
		es.FileID.String(),
		ref.UnrefOrZero(es.FileHash),
		ref.UnrefOrZero(es.SignerFullName),
		strings.ToLower(es.SignerEmail),
		ref.UnrefOrZero(es.SignerIPAddress),
		ref.UnrefOrZero(es.SignerUserAgent),
		es.ConsentText,
		es.SignedAt.UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano),
	}

	for i, f := range fields {
		if f == "" {
			return "", fmt.Errorf("seal field %d must not be empty", i)
		}
		if strings.Contains(f, "\n") {
			return "", fmt.Errorf("seal field %d must not contain newline", i)
		}
	}

	input := strings.Join(fields, "\n")
	return hash.SHA256Hex([]byte(input)), nil
}

func ResetStaleCertificateProcessing(
	ctx context.Context,
	conn pg.Conn,
	staleAfter time.Duration,
) error {
	q := `
UPDATE electronic_signatures
SET certificate_processing_started_at = NULL, updated_at = NOW()
WHERE status = 'COMPLETED'
	AND certificate_file_id IS NULL
	AND certificate_processing_started_at IS NOT NULL
	AND certificate_processing_started_at < NOW() - $1::interval
`
	_, err := conn.Exec(ctx, q, staleAfter.String())
	if err != nil {
		return fmt.Errorf("cannot reset stale certificate processing: %w", err)
	}

	return nil
}
