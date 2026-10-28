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

package console_v1

import (
	"context"
	"errors"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlx"
	tasksync "go.probo.inc/probo/pkg/task/sync"
	"go.probo.inc/probo/pkg/task/sync/asana"
	"go.probo.inc/probo/pkg/validator"
)

func (r *mutationResolver) asanaPublishError(ctx context.Context, err error) error {
	if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
		return gqlx.InvalidValidationErrors(ctx, validationErrors)
	}

	switch {
	case errors.Is(err, coredata.ErrResourceNotFound):
		return gqlx.NotFound(ctx, err)
	case errors.Is(err, tasksync.ErrAsanaNotConnected),
		errors.Is(err, tasksync.ErrAsanaProjectIDRequired),
		errors.Is(err, tasksync.ErrAsanaProjectNotFound),
		errors.Is(err, tasksync.ErrAsanaTaskIDRequired),
		errors.Is(err, tasksync.ErrAsanaTaskNotFound),
		errors.Is(err, tasksync.ErrAsanaMappingInvalid):
		return gqlx.Invalid(ctx, err)
	case errors.Is(err, tasksync.ErrTaskAlreadyLinked),
		errors.Is(err, coredata.ErrResourceAlreadyExists):
		return gqlx.Conflict(ctx, err)
	default:
		r.logger.ErrorCtx(ctx, "cannot publish task to Asana", log.Error(err))

		return gqlx.Internal(ctx)
	}
}

func (r *mutationResolver) asanaSettingsError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, coredata.ErrResourceNotFound):
		return gqlx.NotFound(ctx, err)
	case errors.Is(err, tasksync.ErrAsanaNotConnected),
		errors.Is(err, tasksync.ErrAsanaProjectNotFound),
		errors.Is(err, tasksync.ErrAsanaProjectIDRequired),
		errors.Is(err, tasksync.ErrAsanaDefaultProjectRequired),
		errors.Is(err, tasksync.ErrAsanaSettingsRequired),
		errors.Is(err, tasksync.ErrAsanaMappingInvalid):
		return gqlx.Invalid(ctx, err)
	default:
		r.logger.ErrorCtx(ctx, "cannot update Asana settings", log.Error(err))

		return gqlx.Internal(ctx)
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func pageSize(first *int) int {
	if first == nil {
		return 0
	}

	return *first
}

func asanaPageInfo(edges []*types.AsanaTeamEdge, endCursor string, hasNextPage bool) *types.AsanaPageInfo {
	start := ""
	if len(edges) > 0 {
		start = edges[0].Cursor
	}

	return asanaPageInfoFrom(start, endCursor, hasNextPage)
}

func asanaProjectPageInfo(edges []*types.AsanaProjectEdge, endCursor string, hasNextPage bool) *types.AsanaPageInfo {
	start := ""
	if len(edges) > 0 {
		start = edges[0].Cursor
	}

	return asanaPageInfoFrom(start, endCursor, hasNextPage)
}

func asanaPageInfoFrom(startCursor, endCursor string, hasNextPage bool) *types.AsanaPageInfo {
	info := &types.AsanaPageInfo{
		HasNextPage:     hasNextPage && endCursor != "",
		HasPreviousPage: false,
	}
	if startCursor != "" {
		info.StartCursor = &startCursor
	}

	if endCursor != "" {
		info.EndCursor = &endCursor
	}

	return info
}

func emptyAsanaTeamConnection() *types.AsanaTeamConnection {
	return &types.AsanaTeamConnection{
		Edges:    []*types.AsanaTeamEdge{},
		PageInfo: asanaPageInfoFrom("", "", false),
	}
}

func emptyAsanaProjectConnection() *types.AsanaProjectConnection {
	return &types.AsanaProjectConnection{
		Edges:    []*types.AsanaProjectEdge{},
		PageInfo: asanaPageInfoFrom("", "", false),
	}
}

func emptyAsanaTaskConnection() *types.AsanaTaskConnection {
	return &types.AsanaTaskConnection{
		Edges:    []*types.AsanaTaskEdge{},
		PageInfo: asanaPageInfoFrom("", "", false),
	}
}

func asanaSettingsNode(settings tasksync.AsanaSettings) *types.AsanaSettings {
	taskImport := types.AsanaTaskImportInactive
	if settings.ImportNewTasks {
		taskImport = types.AsanaTaskImportActive
	}

	node := &types.AsanaSettings{TaskImport: taskImport}
	if settings.DefaultProject != nil {
		node.DefaultProject = &types.AsanaProject{
			ID:   settings.DefaultProject.GID,
			Name: settings.DefaultProject.Name,
		}
	}

	return node
}

func asanaSetupNode(setup asana.ProjectSetup, mapping *tasksync.AsanaProjectMapping) *types.AsanaProjectSetup {
	sections := make([]*types.AsanaSection, 0, len(setup.Sections))
	for _, section := range setup.Sections {
		sections = append(sections, &types.AsanaSection{ID: section.GID, Name: section.Name})
	}

	fields := make([]*types.AsanaEnumField, 0, len(setup.EnumFields))
	for _, field := range setup.EnumFields {
		options := make([]*types.AsanaEnumOption, 0, len(field.Options))
		for _, option := range field.Options {
			options = append(options, &types.AsanaEnumOption{ID: option.GID, Name: option.Name})
		}

		fields = append(fields, &types.AsanaEnumField{ID: field.GID, Name: field.Name, Options: options})
	}

	return &types.AsanaProjectSetup{
		Project:    &types.AsanaProject{ID: setup.Project.GID, Name: setup.Project.Name},
		Sections:   sections,
		EnumFields: fields,
		Mapping:    asanaMappingNode(mapping),
	}
}

func asanaMappingNode(mapping *tasksync.AsanaProjectMapping) *types.AsanaProjectMapping {
	if mapping == nil {
		return nil
	}

	states := make([]*types.AsanaStateMapping, 0, len(mapping.States))
	for state, section := range mapping.States {
		states = append(states, &types.AsanaStateMapping{
			State:     state,
			SectionID: section,
			Inbound:   section != "" && mapping.SectionInbound[section] == state,
		})
	}

	priorities := make([]*types.AsanaPriorityMapping, 0, len(mapping.Priorities))
	for priority, option := range mapping.Priorities {
		priorities = append(priorities, &types.AsanaPriorityMapping{
			Priority: priority,
			OptionID: option,
			Inbound:  option != "" && mapping.OptionInbound[option] == priority,
		})
	}

	node := &types.AsanaProjectMapping{States: states, Priorities: priorities}

	if mapping.PriorityFieldGID != "" {
		field := mapping.PriorityFieldGID
		node.PriorityFieldID = &field
	}

	return node
}

func asanaStateInputs(rows []*types.AsanaStateMappingInput) []tasksync.AsanaStateInput {
	inputs := make([]tasksync.AsanaStateInput, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}

		inputs = append(inputs, tasksync.AsanaStateInput{
			State:      row.State,
			SectionGID: row.SectionID,
			Inbound:    row.Inbound,
		})
	}

	return inputs
}

func asanaPriorityInputs(rows []*types.AsanaPriorityMappingInput) []tasksync.AsanaPriorityInput {
	inputs := make([]tasksync.AsanaPriorityInput, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}

		inputs = append(inputs, tasksync.AsanaPriorityInput{
			Priority:  row.Priority,
			OptionGID: row.OptionID,
			Inbound:   row.Inbound,
		})
	}

	return inputs
}
