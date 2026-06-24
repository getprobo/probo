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
