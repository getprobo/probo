// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

// Predicater is the authorized SQL WHERE fragment and bound arguments that
// constrain which rows a query may read or write. Constraints are composed in
// SQLFragment() and SQLArguments(); callers inject the fragment into static
// query templates via fmt.Sprintf.
//
// GetTenantID() exposes the tenant component for GID allocation and INSERT
// tenant_id columns. It is not the full predicate — additional accessors may be
// added as new constraints (e.g. file visibility) land on Predicate.
type Predicater interface {
	SQLArguments() pgx.StrictNamedArgs
	SQLFragment() string
	GetTenantID() gid.TenantID
}

// NoPredicate applies no row filter (SQL TRUE). Use only for intentional
// cross-tenant or system queries. GetTenantID() panics — do not call it.
type NoPredicate struct{}

// Predicate carries the authorized row-access constraints for a query. Today
// this is tenant isolation only; additional fields may be composed into
// SQLFragment() over time.
type Predicate struct {
	tenantID gid.TenantID
}

var (
	_ Predicater = (*NoPredicate)(nil)
	_ Predicater = (*Predicate)(nil)
)

// NewNoPredicate returns a predicate that applies no row filter.
func NewNoPredicate() *NoPredicate {
	return &NoPredicate{}
}

func (*NoPredicate) SQLArguments() pgx.StrictNamedArgs {
	return pgx.StrictNamedArgs{}
}

func (*NoPredicate) SQLFragment() string {
	return "TRUE"
}

func (*NoPredicate) GetTenantID() gid.TenantID {
	panic(fmt.Errorf("cannot get tenant id from no predicate"))
}

// NewPredicate builds a predicate with tenant isolation only. Prefer the
// predicate returned by authorize when authorization has already run.
func NewPredicate(tenantID gid.TenantID) *Predicate {
	return &Predicate{
		tenantID: tenantID,
	}
}

// NewPredicateFromObjectID derives tenant isolation from the tenant component
// of a GID. It does not carry resource-specific constraints; never use it in
// place of the predicate returned by authorize after a resource lookup.
func NewPredicateFromObjectID(objectID gid.GID) *Predicate {
	return NewPredicate(objectID.TenantID())
}

func (p *Predicate) SQLArguments() pgx.StrictNamedArgs {
	return pgx.StrictNamedArgs{
		"tenant_id": p.tenantID,
	}
}

func (*Predicate) SQLFragment() string {
	return "tenant_id = @tenant_id"
}

func (p *Predicate) GetTenantID() gid.TenantID {
	return p.tenantID
}
