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
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	// Invitation represents an invitation to join an organization
	Invitation struct {
		ID             gid.GID    `db:"id"`
		OrganizationID gid.GID    `db:"organization_id"`
		Email          string     `db:"email"`
		FullName       string     `db:"full_name"`
		Role           string     `db:"role"`
		ExpiresAt      time.Time  `db:"expires_at"`
		AcceptedAt     *time.Time `db:"accepted_at"`
		CreatedAt      time.Time  `db:"created_at"`
	}

	// Invitations is a slice of Invitation
	Invitations []*Invitation

	// InvitationData is the data stored in the invitation token
	InvitationData struct {
		InvitationID   gid.GID `json:"invitation_id"`
		OrganizationID gid.GID `json:"organization_id"`
		Email          string  `json:"email"`
		FullName       string  `json:"full_name"`
		Role           string  `json:"role"`
	}

	// ErrInvitationNotFound is returned when an invitation is not found
	ErrInvitationNotFound struct {
		Token string
	}
)

func (e ErrInvitationNotFound) Error() string {
	return fmt.Sprintf("invitation not found: %s", e.Token)
}

// CursorKey returns the cursor key for pagination
func (i Invitation) CursorKey(orderBy InvitationOrderField) page.CursorKey {
	switch orderBy {
	case InvitationOrderFieldCreatedAt:
		return page.NewCursorKey(i.ID, i.CreatedAt)
	case InvitationOrderFieldExpiresAt:
		return page.NewCursorKey(i.ID, i.ExpiresAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

// Create creates a new invitation
func (i *Invitation) Create(ctx context.Context, conn pg.Conn) error {
	query := `
		INSERT INTO authz_invitations (
			id, organization_id, email, full_name, role, expires_at, created_at
		) VALUES (
			@id, @organization_id, @email, @full_name, @role, @expires_at, @created_at
		)
	`

	args := pgx.StrictNamedArgs{
		"id":              i.ID,
		"organization_id": i.OrganizationID,
		"email":           i.Email,
		"full_name":       i.FullName,
		"role":            i.Role,
		"expires_at":      i.ExpiresAt,
		"created_at":      i.CreatedAt,
	}

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to create invitation: %w", err)
	}

	return nil
}

// LoadByID loads an invitation by ID
func (i *Invitation) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	id gid.GID,
) error {
	query := `
		SELECT id, organization_id, email, full_name, role, expires_at, accepted_at, created_at
		FROM authz_invitations
		WHERE id = @id
	`

	args := pgx.StrictNamedArgs{
		"id": id,
	}

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query invitation: %w", err)
	}

	invitation, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Invitation])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvitationNotFound{Token: id.String()}
		}
		return fmt.Errorf("cannot collect invitation: %w", err)
	}

	*i = invitation
	return nil
}

// Update updates an existing invitation
func (i *Invitation) Update(ctx context.Context, conn pg.Conn) error {
	query := `
		UPDATE authz_invitations
		SET accepted_at = @accepted_at
		WHERE id = @id
	`

	args := pgx.StrictNamedArgs{
		"id":          i.ID,
		"accepted_at": i.AcceptedAt,
	}

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrInvitationNotFound{Token: i.ID.String()}
	}

	return nil
}

// Delete deletes an invitation
func (i *Invitation) Delete(ctx context.Context, conn pg.Conn) error {
	query := `
		DELETE FROM authz_invitations
		WHERE id = @id
	`

	args := pgx.StrictNamedArgs{
		"id": i.ID,
	}

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to delete invitation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrInvitationNotFound{Token: i.ID.String()}
	}

	return nil
}

// LoadByEmail loads all invitations for a specific email
func (i *Invitations) LoadByEmail(
	ctx context.Context,
	conn pg.Conn,
	email string,
) error {
	query := `
		SELECT id, organization_id, email, full_name, role, expires_at, accepted_at, created_at
		FROM authz_invitations
		WHERE email = @email AND accepted_at IS NULL
		ORDER BY created_at DESC
	`

	args := pgx.StrictNamedArgs{
		"email": email,
	}

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query invitations: %w", err)
	}

	invitations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Invitation])
	if err != nil {
		return fmt.Errorf("cannot collect invitations: %w", err)
	}

	*i = invitations
	return nil
}

// LoadByOrganizationID loads all invitations for a specific organization with pagination
func (i *Invitations) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	orgID gid.GID,
	cursor *page.Cursor[InvitationOrderField],
) error {
	query := `
		SELECT id, organization_id, email, full_name, role, expires_at, accepted_at, created_at
		FROM authz_invitations
		WHERE organization_id = @organization_id
		AND %s
	`

	query = fmt.Sprintf(query, cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": orgID}
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query invitations: %w", err)
	}

	invitations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Invitation])
	if err != nil {
		return fmt.Errorf("cannot collect invitations: %w", err)
	}

	*i = invitations
	return nil
}
