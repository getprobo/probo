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

import "github.com/getprobo/probo/pkg/page"

type IncidentRegistryOrderField int

const (
	IncidentRegistryOrderFieldCreatedAt IncidentRegistryOrderField = iota
	IncidentRegistryOrderFieldIncidentDate
	IncidentRegistryOrderFieldStatus
	IncidentRegistryOrderFieldPriority
	IncidentRegistryOrderFieldSeverity
	IncidentRegistryOrderFieldReferenceId
)

func (f IncidentRegistryOrderField) String() string {
	switch f {
	case IncidentRegistryOrderFieldCreatedAt:
		return "created_at"
	case IncidentRegistryOrderFieldIncidentDate:
		return "incident_date"
	case IncidentRegistryOrderFieldStatus:
		return "status"
	case IncidentRegistryOrderFieldPriority:
		return "priority"
	case IncidentRegistryOrderFieldSeverity:
		return "severity"
	case IncidentRegistryOrderFieldReferenceId:
		return "reference_id"
	}

	panic("unknown incident registry order field")
}

func (f IncidentRegistryOrderField) SQL() string {
	switch f {
	case IncidentRegistryOrderFieldCreatedAt:
		return "created_at"
	case IncidentRegistryOrderFieldIncidentDate:
		return "incident_date"
	case IncidentRegistryOrderFieldStatus:
		return "status"
	case IncidentRegistryOrderFieldPriority:
		return "priority"
	case IncidentRegistryOrderFieldSeverity:
		return "severity"
	case IncidentRegistryOrderFieldReferenceId:
		return "reference_id"
	}

	panic("unknown incident registry order field")
}

func (f IncidentRegistryOrderField) GoType() string {
	switch f {
	case IncidentRegistryOrderFieldCreatedAt:
		return "time.Time"
	case IncidentRegistryOrderFieldIncidentDate:
		return "*time.Time"
	case IncidentRegistryOrderFieldStatus:
		return "IncidentRegistryStatus"
	case IncidentRegistryOrderFieldPriority:
		return "IncidentRegistryPriority"
	case IncidentRegistryOrderFieldSeverity:
		return "IncidentRegistrySeverity"
	case IncidentRegistryOrderFieldReferenceId:
		return "string"
	}

	panic("unknown incident registry order field")
}

func (f IncidentRegistryOrderField) NullsLast() bool {
	switch f {
	case IncidentRegistryOrderFieldCreatedAt:
		return false
	case IncidentRegistryOrderFieldIncidentDate:
		return true
	case IncidentRegistryOrderFieldStatus:
		return false
	case IncidentRegistryOrderFieldPriority:
		return false
	case IncidentRegistryOrderFieldSeverity:
		return false
	case IncidentRegistryOrderFieldReferenceId:
		return false
	}

	panic("unknown incident registry order field")
}

func (f IncidentRegistryOrderField) Column() string {
	return f.String()
}

// Implement page.OrderByField interface
func (f IncidentRegistryOrderField) OrderBySQL(direction page.OrderDirection) string {
	sql := f.SQL()
	if f.NullsLast() {
		if direction == page.OrderDirectionAsc {
			return sql + " ASC NULLS LAST"
		} else {
			return sql + " DESC NULLS FIRST"
		}
	}
	if direction == page.OrderDirectionAsc {
		return sql + " ASC"
	}
	return sql + " DESC"
}