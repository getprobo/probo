// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	Session struct {
		ID              gid.GID       `db:"id"`
		IdentityID      gid.GID       `db:"identity_id"`
		TenantID        *gid.TenantID `db:"tenant_id"`
		MembershipID    *gid.GID      `db:"membership_id"`
		ParentSessionID *gid.GID      `db:"parent_session_id"`
		Data            SessionData   `db:"data"`
		AuthMethod      AuthMethod    `db:"auth_method"`
		AuthenticatedAt time.Time     `db:"authenticated_at"`
		UserAgent       string        `db:"user_agent"`
		IPAddress       net.IP        `db:"ip_address"`
		ExpireReason    *ExpireReason `db:"expire_reason"`
		ExpiredAt       time.Time     `db:"expired_at"`
		CreatedAt       time.Time     `db:"created_at"`
		UpdatedAt       time.Time     `db:"updated_at"`
	}

	Sessions []*Session

	SessionData struct {
		BoundHost string `json:"bound_host,omitempty"`
	}

	AuthMethod string
)

func NormalizeBoundHost(host string) string {
	return strings.ToLower(strings.TrimSpace(host))
}

func SessionDataForHost(host string) SessionData {
	return SessionData{
		BoundHost: NormalizeBoundHost(host),
	}
}

func (d SessionData) IsDomainBound() bool {
	return d.BoundHost != ""
}

func (d SessionData) MatchesBoundHost(requestHost string) bool {
	if !d.IsDomainBound() {
		return false
	}

	return NormalizeBoundHost(d.BoundHost) == NormalizeBoundHost(requestHost)
}

func (d SessionData) Value() (driver.Value, error) {
	data, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal session data: %w", err)
	}

	return data, nil
}

func (d *SessionData) Scan(value any) error {
	if value == nil {
		*d = SessionData{}

		return nil
	}

	var data []byte

	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("cannot scan session data: unsupported type %T", value)
	}

	if len(data) == 0 {
		*d = SessionData{}

		return nil
	}

	if err := json.Unmarshal(data, d); err != nil {
		return fmt.Errorf("cannot unmarshal session data: %w", err)
	}

	return nil
}

const (
	AuthMethodMagicLink AuthMethod = "MAGIC_LINK"
	AuthMethodPassword  AuthMethod = "PASSWORD"
	AuthMethodSAML      AuthMethod = "SAML"
	AuthMethodOIDC      AuthMethod = "OIDC"
)

var (
	_ fmt.Stringer             = AuthMethod("")
	_ encoding.TextMarshaler   = AuthMethod("")
	_ encoding.TextUnmarshaler = (*AuthMethod)(nil)
)

func AuthMethods() []AuthMethod {
	return []AuthMethod{
		AuthMethodMagicLink,
		AuthMethodPassword,
		AuthMethodSAML,
		AuthMethodOIDC,
	}
}

func (v AuthMethod) IsValid() bool {
	switch v {
	case
		AuthMethodMagicLink,
		AuthMethodPassword,
		AuthMethodSAML,
		AuthMethodOIDC:
		return true
	}

	return false
}

func (v AuthMethod) String() string {
	return string(v)
}

func (v AuthMethod) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AuthMethod) UnmarshalText(text []byte) error {
	val := AuthMethod(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AuthMethod value: %q", string(text))
	}

	*v = val

	return nil
}

func NewRootSession(identityID gid.GID, method AuthMethod, duration time.Duration) *Session {
	now := time.Now()

	return &Session{
		ID:              gid.New(gid.NilTenant, SessionEntityType),
		IdentityID:      identityID,
		ExpiredAt:       now.Add(duration),
		AuthMethod:      method,
		AuthenticatedAt: now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (s Session) CursorKey(orderBy SessionOrderField) page.CursorKey {
	switch orderBy {
	case SessionOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	case SessionOrderFieldExpiredAt:
		return page.NewCursorKey(s.ID, s.ExpiredAt)
	case SessionOrderFieldUpdatedAt:
		return page.NewCursorKey(s.ID, s.UpdatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *Session) IsRootSession() bool {
	return s.ParentSessionID == nil
}

func (s *Session) IsChildSession() bool {
	return s.ParentSessionID != nil
}

func (s *Session) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	sessionID gid.GID,
) error {
	q := `
SELECT
    id,
    identity_id,
    tenant_id,
    membership_id,
    data,
    parent_session_id,
    auth_method,
    authenticated_at,
    expire_reason,
    user_agent,
    ip_address,
    expired_at,
    created_at,
    updated_at
FROM
    iam_sessions
WHERE
    id = @session_id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"session_id": sessionID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query session: %w", err)
	}

	session, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Session])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect session: %w", err)
	}

	*s = session

	return nil
}

// AuthorizationAttributes loads the minimal authorization attributes for policy condition evaluation.
// It is intentionally lightweight and does not populate the Session struct.
func (s *Session) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT
    id,
    identity_id
FROM
    iam_sessions
WHERE
    id = ANY(@resource_ids::text[])
`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query session authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID, len(resourceIDs))

	for rows.Next() {
		var (
			id         gid.GID
			identityID gid.GID
		)

		err = rows.Scan(&id, &identityID)
		if err != nil {
			return nil, fmt.Errorf("cannot scan session authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{"identity_id": identityID.String()}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate session authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (s *Session) Insert(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
INSERT INTO
    iam_sessions (id, identity_id, tenant_id, membership_id, data, parent_session_id, auth_method, authenticated_at, expire_reason, user_agent, ip_address, expired_at, created_at, updated_at)
VALUES (
    @session_id,
    @identity_id,
    @tenant_id,
    @membership_id,
    @data,
    @parent_session_id,
    @auth_method,
    @authenticated_at,
    @expire_reason,
    @user_agent,
    @ip_address,
    @expired_at,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"session_id":        s.ID,
		"identity_id":       s.IdentityID,
		"tenant_id":         s.TenantID,
		"membership_id":     s.MembershipID,
		"data":              s.Data,
		"parent_session_id": s.ParentSessionID,
		"auth_method":       s.AuthMethod,
		"authenticated_at":  s.AuthenticatedAt,
		"expire_reason":     s.ExpireReason,
		"user_agent":        s.UserAgent,
		"ip_address":        s.IPAddress,
		"expired_at":        s.ExpiredAt,
		"created_at":        s.CreatedAt,
		"updated_at":        s.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)

	return err
}

func (s *Session) Update(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
UPDATE iam_sessions
SET
    expired_at = @expired_at,
    updated_at = @updated_at,
    user_agent = @user_agent,
    ip_address = @ip_address,
    expire_reason = @expire_reason,
    data = @data
WHERE
    id = @session_id
`

	args := pgx.StrictNamedArgs{
		"session_id":    s.ID,
		"user_agent":    s.UserAgent,
		"ip_address":    s.IPAddress,
		"expire_reason": s.ExpireReason,
		"data":          s.Data,
		"expired_at":    s.ExpiredAt,
		"updated_at":    s.UpdatedAt,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *Sessions) LoadByIdentityID(ctx context.Context, conn pg.Querier, identityID gid.GID, cursor *page.Cursor[SessionOrderField]) error {
	q := `
SELECT
    id,
    identity_id,
    tenant_id,
    membership_id,
    data,
    parent_session_id,
    auth_method,
    authenticated_at,
    expire_reason,
    user_agent,
    ip_address,
    expired_at,
    created_at,
    updated_at
FROM
    iam_sessions
WHERE
    identity_id = @identity_id
    AND %s
`

	q = fmt.Sprintf(q, cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"identity_id": identityID}
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query sessions: %w", err)
	}

	sessions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Session])
	if err != nil {
		return fmt.Errorf("cannot collect sessions: %w", err)
	}

	*s = sessions

	return nil
}

func (s *Sessions) CountByIdentityID(ctx context.Context, conn pg.Querier, identityID gid.GID) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    iam_sessions
WHERE
    identity_id = @identity_id
`

	args := pgx.StrictNamedArgs{"identity_id": identityID}

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (s *Sessions) ExpireAllForIdentityExceptOneSession(ctx context.Context, conn pg.Querier, identityID gid.GID, sessionID gid.GID) (int64, error) {
	q := `
UPDATE iam_sessions
SET
    expired_at = NOW(),
    updated_at = NOW(),
    expire_reason = 'revoked'
WHERE
    id != @session_id
    AND identity_id = @identity_id
    AND expire_reason IS NULL
`

	args := pgx.StrictNamedArgs{
		"session_id":  sessionID,
		"identity_id": identityID,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot query sessions: %w", err)
	}

	return result.RowsAffected(), nil
}

func (s *Sessions) ExpireAllForIdentity(ctx context.Context, conn pg.Querier, identityID gid.GID) (int64, error) {
	q := `
UPDATE iam_sessions
SET
    expired_at = NOW(),
    updated_at = NOW(),
    expire_reason = 'revoked'
WHERE
    identity_id = @identity_id
    AND expire_reason IS NULL
`

	args := pgx.StrictNamedArgs{
		"identity_id": identityID,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot query sessions: %w", err)
	}

	return result.RowsAffected(), nil
}

func (s *Session) LoadByRootSessionIDAndMembershipID(
	ctx context.Context,
	conn pg.Querier,
	rootSessionID gid.GID,
	membershipID gid.GID,
) error {
	q := `
SELECT
    id,
    identity_id,
    tenant_id,
    membership_id,
    data,
    parent_session_id,
    auth_method,
    authenticated_at,
    expire_reason,
    user_agent,
    ip_address,
    expired_at,
    created_at,
    updated_at
FROM
    iam_sessions
WHERE
    parent_session_id = @root_session_id
    AND membership_id = @membership_id
ORDER BY created_at DESC
LIMIT 1
`

	args := pgx.StrictNamedArgs{
		"root_session_id": rootSessionID,
		"membership_id":   membershipID,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query session: %w", err)
	}

	session, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Session])
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect session: %w", err)
	}

	*s = session

	return nil
}
