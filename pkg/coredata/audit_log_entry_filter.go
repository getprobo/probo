// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"github.com/jackc/pgx/v5"
	"go.probo.inc/probo/pkg/gid"
)

type AuditLogEntryFilter struct {
	action       *string
	actorID      *gid.GID
	resourceType *string
	resourceID   *gid.GID
}

func NewAuditLogEntryFilter() *AuditLogEntryFilter {
	return &AuditLogEntryFilter{}
}

func (f *AuditLogEntryFilter) WithAction(action string) *AuditLogEntryFilter {
	f.action = &action
	return f
}

func (f *AuditLogEntryFilter) WithActorID(actorID gid.GID) *AuditLogEntryFilter {
	f.actorID = &actorID
	return f
}

func (f *AuditLogEntryFilter) WithResourceType(resourceType string) *AuditLogEntryFilter {
	f.resourceType = &resourceType
	return f
}

func (f *AuditLogEntryFilter) WithResourceID(resourceID gid.GID) *AuditLogEntryFilter {
	f.resourceID = &resourceID
	return f
}

func (f *AuditLogEntryFilter) SQLFragment() string {
	return `
(
    CASE
        WHEN @filter_action::text IS NOT NULL THEN
            action = @filter_action::text
        ELSE TRUE
    END
    AND
    CASE
        WHEN @filter_actor_id::text IS NOT NULL THEN
            actor_id = @filter_actor_id::text
        ELSE TRUE
    END
    AND
    CASE
        WHEN @filter_resource_type::text IS NOT NULL THEN
            resource_type = @filter_resource_type::text
        ELSE TRUE
    END
    AND
    CASE
        WHEN @filter_resource_id::text IS NOT NULL THEN
            resource_id = @filter_resource_id::text
        ELSE TRUE
    END
)`
}

func (f *AuditLogEntryFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"filter_action":        nil,
		"filter_actor_id":      nil,
		"filter_resource_type": nil,
		"filter_resource_id":   nil,
	}

	if f.action != nil {
		args["filter_action"] = *f.action
	}

	if f.actorID != nil {
		args["filter_actor_id"] = *f.actorID
	}

	if f.resourceType != nil {
		args["filter_resource_type"] = *f.resourceType
	}

	if f.resourceID != nil {
		args["filter_resource_id"] = *f.resourceID
	}

	return args
}
