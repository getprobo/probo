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
	"time"

	"github.com/jackc/pgx/v5"
	"go.probo.inc/probo/pkg/gid"
)

type AuditLogEntryFilter struct {
	action       *string
	actionPrefix *string
	actorID      *gid.GID
	actorType    *AuditLogActorType
	resourceType *string
	resourceID   *gid.GID
	outcome      *AuditLogOutcome
	requestID    *string
	createdAtGte *time.Time
	createdAtLt  *time.Time
}

func NewAuditLogEntryFilter() *AuditLogEntryFilter {
	return &AuditLogEntryFilter{}
}

func (f *AuditLogEntryFilter) WithAction(action string) *AuditLogEntryFilter {
	f.action = &action
	return f
}

func (f *AuditLogEntryFilter) WithActionPrefix(prefix string) *AuditLogEntryFilter {
	f.actionPrefix = &prefix
	return f
}

func (f *AuditLogEntryFilter) WithActorID(actorID gid.GID) *AuditLogEntryFilter {
	f.actorID = &actorID
	return f
}

func (f *AuditLogEntryFilter) WithActorType(actorType AuditLogActorType) *AuditLogEntryFilter {
	f.actorType = &actorType
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

func (f *AuditLogEntryFilter) WithOutcome(outcome AuditLogOutcome) *AuditLogEntryFilter {
	f.outcome = &outcome
	return f
}

func (f *AuditLogEntryFilter) WithRequestID(requestID string) *AuditLogEntryFilter {
	f.requestID = &requestID
	return f
}

func (f *AuditLogEntryFilter) WithCreatedAtGte(t time.Time) *AuditLogEntryFilter {
	f.createdAtGte = &t
	return f
}

func (f *AuditLogEntryFilter) WithCreatedAtLt(t time.Time) *AuditLogEntryFilter {
	f.createdAtLt = &t
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
        WHEN @filter_action_prefix::text IS NOT NULL THEN
            starts_with(action, @filter_action_prefix::text)
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
        WHEN @filter_actor_type::text IS NOT NULL THEN
            actor_type = @filter_actor_type::text
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
    AND
    CASE
        WHEN @filter_outcome::text IS NOT NULL THEN
            outcome = @filter_outcome::text
        ELSE TRUE
    END
    AND
    CASE
        WHEN @filter_request_id::text IS NOT NULL THEN
            request_id = @filter_request_id::text
        ELSE TRUE
    END
    AND
    CASE
        WHEN @filter_created_at_gte::timestamptz IS NOT NULL THEN
            created_at >= @filter_created_at_gte::timestamptz
        ELSE TRUE
    END
    AND
    CASE
        WHEN @filter_created_at_lt::timestamptz IS NOT NULL THEN
            created_at < @filter_created_at_lt::timestamptz
        ELSE TRUE
    END
)`
}

func (f *AuditLogEntryFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"filter_action":         nil,
		"filter_action_prefix":  nil,
		"filter_actor_id":       nil,
		"filter_actor_type":     nil,
		"filter_resource_type":  nil,
		"filter_resource_id":    nil,
		"filter_outcome":        nil,
		"filter_request_id":     nil,
		"filter_created_at_gte": nil,
		"filter_created_at_lt":  nil,
	}

	if f.action != nil {
		args["filter_action"] = *f.action
	}

	if f.actionPrefix != nil {
		args["filter_action_prefix"] = *f.actionPrefix
	}

	if f.actorID != nil {
		args["filter_actor_id"] = *f.actorID
	}

	if f.actorType != nil {
		args["filter_actor_type"] = *f.actorType
	}

	if f.resourceType != nil {
		args["filter_resource_type"] = *f.resourceType
	}

	if f.resourceID != nil {
		args["filter_resource_id"] = *f.resourceID
	}

	if f.outcome != nil {
		args["filter_outcome"] = *f.outcome
	}

	if f.requestID != nil {
		args["filter_request_id"] = *f.requestID
	}

	if f.createdAtGte != nil {
		args["filter_created_at_gte"] = *f.createdAtGte
	}

	if f.createdAtLt != nil {
		args["filter_created_at_lt"] = *f.createdAtLt
	}

	return args
}
