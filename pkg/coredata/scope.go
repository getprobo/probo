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
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.probo.inc/probo/pkg/gid"
)

type (
	Scoper interface {
		SQLArguments() pgx.StrictNamedArgs
		SQLFragment() string
		GetTenantID() gid.TenantID
	}

	NoScope struct{}

	Scope struct {
		tenantID gid.TenantID
	}
)

var (
	_ Scoper = (*NoScope)(nil)
	_ Scoper = (*Scope)(nil)
)

func NewNoScope() *NoScope {
	return &NoScope{}
}

func (*NoScope) SQLArguments() pgx.StrictNamedArgs {
	return pgx.StrictNamedArgs{}
}

func (*NoScope) SQLFragment() string {
	return "TRUE"
}

func (*NoScope) GetTenantID() gid.TenantID {
	panic(fmt.Errorf("cannot get tenant id from no scope"))
}

func NewScope(tenantID gid.TenantID) *Scope {
	return &Scope{
		tenantID: tenantID,
	}
}

func NewScopeFromObjectID(objectID gid.GID) *Scope {
	return NewScope(objectID.TenantID())
}

func (s *Scope) SQLArguments() pgx.StrictNamedArgs {
	return pgx.StrictNamedArgs{
		"tenant_id": s.tenantID,
	}
}

func (*Scope) SQLFragment() string {
	return "tenant_id = @tenant_id"
}

func (s *Scope) GetTenantID() gid.TenantID {
	return s.tenantID
}
