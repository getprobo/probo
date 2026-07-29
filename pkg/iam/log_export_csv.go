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

package iam

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type (
	scimProfileExportInfo struct {
		email    string
		fullName string
	}

	auditLogActorExportInfo struct {
		email string
		name  string
	}
)

var (
	auditLogExportCSVHeader = []string{
		"organization_name",
		"id",
		"created_at",
		"actor_type",
		"actor_id",
		"actor_email",
		"actor_name",
		"action",
		"resource_type",
		"resource_id",
	}

	scimEventExportCSVHeader = []string{
		"organization_name",
		"id",
		"created_at",
		"method",
		"path",
		"user_name",
		"email",
		"full_name",
		"status_code",
		"error_message",
		"ip_address",
	}

	auditLogExportOrderBy = page.OrderBy[coredata.AuditLogEntryOrderField]{
		Field:     coredata.AuditLogEntryOrderFieldCreatedAt,
		Direction: page.OrderDirectionAsc,
	}

	scimEventExportOrderBy = page.OrderBy[coredata.SCIMEventOrderField]{
		Field:     coredata.SCIMEventOrderFieldCreatedAt,
		Direction: page.OrderDirectionAsc,
	}
)

func (s *LogExportService) streamAuditLogCSV(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	organizationID gid.GID,
	organizationName string,
	filter *coredata.AuditLogEntryFilter,
	w *csv.Writer,
) error {
	if err := w.Write(auditLogExportCSVHeader); err != nil {
		return fmt.Errorf("cannot write audit log CSV header: %w", err)
	}

	return page.WalkAll(
		ctx,
		auditLogExportOrderBy,
		func(ctx context.Context, cursor *page.Cursor[coredata.AuditLogEntryOrderField]) (coredata.AuditLogEntries, error) {
			var logs coredata.AuditLogEntries
			if err := logs.LoadByOrganizationID(
				ctx,
				conn,
				scope,
				organizationID,
				cursor,
				filter,
			); err != nil {
				return nil, fmt.Errorf("cannot load audit log entries: %w", err)
			}

			return logs, nil
		},
		func(entries coredata.AuditLogEntries) error {
			actorsByID, err := loadAuditLogActorExportInfo(ctx, conn, entries)
			if err != nil {
				return err
			}

			for _, entry := range entries {
				row := auditLogEntryCSVRow(organizationName, entry, actorsByID[entry.ActorID])
				if err := w.Write(row); err != nil {
					return fmt.Errorf("cannot write audit log CSV row: %w", err)
				}
			}

			w.Flush()

			if err := w.Error(); err != nil {
				return fmt.Errorf("cannot flush audit log CSV writer: %w", err)
			}

			return nil
		},
	)
}

func (s *LogExportService) streamSCIMEventCSV(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	organizationID gid.GID,
	organizationName string,
	filter *coredata.SCIMEventFilter,
	w *csv.Writer,
) error {
	if err := w.Write(scimEventExportCSVHeader); err != nil {
		return fmt.Errorf("cannot write SCIM event CSV header: %w", err)
	}

	return page.WalkAll(
		ctx,
		scimEventExportOrderBy,
		func(ctx context.Context, cursor *page.Cursor[coredata.SCIMEventOrderField]) (coredata.SCIMEvents, error) {
			var events coredata.SCIMEvents
			if err := events.LoadByOrganizationID(
				ctx,
				conn,
				scope,
				organizationID,
				cursor,
				filter,
			); err != nil {
				return nil, fmt.Errorf("cannot load SCIM events: %w", err)
			}

			return events, nil
		},
		func(events coredata.SCIMEvents) error {
			profilesByUserName, err := loadSCIMProfileExportInfo(
				ctx,
				conn,
				scope,
				organizationID,
				events,
			)
			if err != nil {
				return err
			}

			for _, event := range events {
				row := scimEventCSVRow(organizationName, event, profilesByUserName)
				if err := w.Write(row); err != nil {
					return fmt.Errorf("cannot write SCIM event CSV row: %w", err)
				}
			}

			w.Flush()

			if err := w.Error(); err != nil {
				return fmt.Errorf("cannot flush SCIM event CSV writer: %w", err)
			}

			return nil
		},
	)
}

func auditLogEntryCSVRow(
	organizationName string,
	entry *coredata.AuditLogEntry,
	actor auditLogActorExportInfo,
) []string {
	return []string{
		organizationName,
		entry.ID.String(),
		entry.CreatedAt.Format(time.RFC3339),
		string(entry.ActorType),
		entry.ActorID.String(),
		actor.email,
		actor.name,
		entry.Action,
		entry.ResourceType,
		entry.ResourceID.String(),
	}
}

func scimEventCSVRow(
	organizationName string,
	event *coredata.SCIMEvent,
	profilesByUserName map[string]scimProfileExportInfo,
) []string {
	profile := profilesByUserName[strings.ToLower(event.UserName)]
	email := profile.email
	fullName := profile.fullName

	if email == "" {
		email = scimEmailFromUserName(event.UserName)
	}

	return []string{
		organizationName,
		event.ID.String(),
		event.CreatedAt.Format(time.RFC3339),
		event.Method,
		event.Path,
		event.UserName,
		email,
		fullName,
		strconv.Itoa(event.StatusCode),
		stringPtrValue(event.ErrorMessage),
		event.IPAddress.String(),
	}
}

func loadAuditLogActorExportInfo(
	ctx context.Context,
	conn pg.Querier,
	entries coredata.AuditLogEntries,
) (map[gid.GID]auditLogActorExportInfo, error) {
	identityIDs := make([]gid.GID, 0)
	apiKeyIDs := make([]gid.GID, 0)

	for _, entry := range entries {
		switch entry.ActorType {
		case coredata.AuditLogActorTypeUser:
			identityIDs = append(identityIDs, entry.ActorID)
		case coredata.AuditLogActorTypeAPIKey:
			apiKeyIDs = append(apiKeyIDs, entry.ActorID)
		case coredata.AuditLogActorTypeSystem:
		default:
		}
	}

	result := make(map[gid.GID]auditLogActorExportInfo)

	if len(identityIDs) > 0 {
		var identities coredata.Identities
		if err := identities.LoadByIDs(ctx, conn, identityIDs); err != nil {
			return nil, fmt.Errorf("cannot load audit log actor identities: %w", err)
		}

		for _, identity := range identities {
			result[identity.ID] = auditLogActorExportInfo{
				email: identity.EmailAddress.String(),
				name:  identity.FullName,
			}
		}
	}

	if len(apiKeyIDs) > 0 {
		var apiKeys coredata.PersonalAPIKeys
		if err := apiKeys.LoadByIDs(ctx, conn, apiKeyIDs); err != nil {
			return nil, fmt.Errorf("cannot load audit log actor API keys: %w", err)
		}

		for _, apiKey := range apiKeys {
			result[apiKey.ID] = auditLogActorExportInfo{
				name: apiKey.Name,
			}
		}
	}

	return result, nil
}

func loadSCIMProfileExportInfo(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	organizationID gid.GID,
	events coredata.SCIMEvents,
) (map[string]scimProfileExportInfo, error) {
	userNames := uniqueNonEmptyStrings(scimEventUserNames(events))
	if len(userNames) == 0 {
		return map[string]scimProfileExportInfo{}, nil
	}

	var profiles coredata.MembershipProfiles
	if err := profiles.LoadByOrganizationIDAndUserNames(
		ctx,
		conn,
		scope,
		organizationID,
		userNames,
	); err != nil {
		return nil, fmt.Errorf("cannot load SCIM profile export info: %w", err)
	}

	result := make(map[string]scimProfileExportInfo, len(profiles))
	for _, profile := range profiles {
		if profile.UserName == nil {
			continue
		}

		key := strings.ToLower(*profile.UserName)
		result[key] = scimProfileExportInfo{
			email:    scimEmailFromUserName(*profile.UserName),
			fullName: profileFullName(profile),
		}
	}

	return result, nil
}

func scimEventUserNames(events coredata.SCIMEvents) []string {
	userNames := make([]string, 0, len(events))
	for _, event := range events {
		userNames = append(userNames, event.UserName)
	}

	return userNames
}

func uniqueNonEmptyStrings(values []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		out = append(out, value)
	}

	return out
}

func scimEmailFromUserName(userName string) string {
	userName = strings.TrimSpace(userName)
	if userName == "" {
		return ""
	}

	if _, err := mail.ParseAddr(userName); err == nil {
		return userName
	}

	return ""
}

func profileFullName(profile *coredata.MembershipProfile) string {
	if profile.FormattedName != nil && *profile.FormattedName != "" {
		return *profile.FormattedName
	}

	return profile.FullName
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
