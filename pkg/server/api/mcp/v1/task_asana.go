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

package mcp_v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
	asanasvc "go.probo.inc/probo/pkg/task/sync/asana"
	asana "go.probo.inc/probo/pkg/task/sync/asana/client"
	"go.probo.inc/probo/pkg/task/sync/shared"
	"go.probo.inc/probo/pkg/validator"
)

func asanaPublishToolError(r *Resolver, ctx context.Context, err error) error {
	switch {
	case errors.Is(err, coredata.ErrResourceNotFound):
		return fmt.Errorf("task not found")
	case errors.Is(err, asanasvc.ErrAsanaNotConnected):
		return fmt.Errorf("asana connector is not connected")
	case errors.Is(err, shared.ErrTaskAlreadyLinked),
		errors.Is(err, coredata.ErrResourceAlreadyExists):
		return fmt.Errorf("task is already linked to an external issue")
	case errors.Is(err, asanasvc.ErrAsanaProjectIDRequired):
		return fmt.Errorf("asana project id is required")
	case errors.Is(err, asanasvc.ErrAsanaProjectNotFound):
		return fmt.Errorf("asana project was not found")
	case errors.Is(err, asanasvc.ErrAsanaTaskIDRequired):
		return fmt.Errorf("asana task id is required")
	case errors.Is(err, asanasvc.ErrAsanaTaskNotFound):
		return fmt.Errorf("asana task was not found")
	default:
		r.logger.ErrorCtx(ctx, "cannot publish task to Asana", log.Error(err))

		return fmt.Errorf("internal error")
	}
}

func asanaSettingsToolError(r *Resolver, ctx context.Context, err error) error {
	if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
		return fmt.Errorf("%s", validationErrors.Error())
	}

	switch {
	case errors.Is(err, coredata.ErrResourceNotFound):
		return fmt.Errorf("organization not found")
	case errors.Is(err, asanasvc.ErrAsanaNotConnected):
		return fmt.Errorf("asana connector is not connected")
	case errors.Is(err, asanasvc.ErrAsanaProjectNotFound):
		return fmt.Errorf("asana project was not found")
	case errors.Is(err, asanasvc.ErrAsanaProjectIDRequired):
		return fmt.Errorf("asana project id is required")
	case errors.Is(err, asanasvc.ErrAsanaDefaultProjectRequired):
		return fmt.Errorf("a default Asana project is required")
	case errors.Is(err, asanasvc.ErrAsanaSettingsRequired):
		return fmt.Errorf("an Asana project or a task import status is required")
	default:
		r.logger.ErrorCtx(ctx, "cannot update Asana settings", log.Error(err))

		return fmt.Errorf("internal error")
	}
}

func asanaSettingsUpdate(raw json.RawMessage, taskImport *types.AsanaTaskImport) (asanasvc.AsanaSettingsUpdate, error) {
	projectSet, projectID, err := rawOptionalString(raw, "default_project_id")
	if err != nil {
		return asanasvc.AsanaSettingsUpdate{}, fmt.Errorf("cannot read default project id: %w", err)
	}

	update := asanasvc.AsanaSettingsUpdate{}
	if projectSet {
		update.DefaultProjectID = &projectID
	}

	if taskImport != nil {
		if !taskImport.IsValid() {
			return asanasvc.AsanaSettingsUpdate{}, fmt.Errorf("task_import must be ACTIVE or INACTIVE")
		}

		active := *taskImport == types.AsanaTaskImportACTIVE
		update.ImportNewTasks = &active
	}

	return update, nil
}

func asanaMappingUpdate(raw json.RawMessage, input *types.UpdateAsanaProjectMappingInput) (asanasvc.AsanaMappingUpdate, error) {
	fieldSet, fieldID, err := rawOptionalString(raw, "priority_field_id")
	if err != nil {
		return asanasvc.AsanaMappingUpdate{}, err
	}

	update := asanasvc.AsanaMappingUpdate{
		ProjectGID:       input.ProjectID,
		Clear:            input.Clear != nil && *input.Clear,
		StatesSet:        input.States != nil,
		PriorityFieldSet: fieldSet,
		PrioritiesSet:    input.Priorities != nil,
	}
	if fieldSet {
		update.PriorityFieldGID = fieldID
	}

	update.States = asanaStateInputs(input.States)
	update.Priorities = asanaPriorityInputs(input.Priorities)

	return update, nil
}

func asanaStateInputs(rows []*types.AsanaStateMapping) []asanasvc.AsanaStateInput {
	inputs := make([]asanasvc.AsanaStateInput, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}

		inputs = append(inputs, asanasvc.AsanaStateInput{
			State:      row.State,
			SectionGID: row.SectionID,
			Inbound:    row.Inbound,
		})
	}

	return inputs
}

func asanaPriorityInputs(rows []*types.AsanaPriorityMapping) []asanasvc.AsanaPriorityInput {
	inputs := make([]asanasvc.AsanaPriorityInput, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}

		inputs = append(inputs, asanasvc.AsanaPriorityInput{
			Priority:  row.Priority,
			OptionGID: row.OptionID,
			Inbound:   row.Inbound,
		})
	}

	return inputs
}

func mcpAsanaSettings(settings asanasvc.AsanaSettings) *types.AsanaSettings {
	status := types.AsanaTaskImportINACTIVE
	if settings.ImportNewTasks {
		status = types.AsanaTaskImportACTIVE
	}

	node := &types.AsanaSettings{TaskImport: &status}
	if settings.DefaultProject != nil {
		node.DefaultProject = &types.AsanaNamed{
			ID:   settings.DefaultProject.GID,
			Name: settings.DefaultProject.Name,
		}
	}

	return node
}

func mcpAsanaSetup(setup asana.ProjectSetup, mapping *asanasvc.AsanaProjectMapping) *types.AsanaProjectSetup {
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

		fields = append(fields, &types.AsanaEnumField{
			ID:      field.GID,
			Name:    field.Name,
			Options: options,
		})
	}

	return &types.AsanaProjectSetup{
		Project:               &types.AsanaNamed{ID: setup.Project.GID, Name: setup.Project.Name},
		Sections:              sections,
		EnumFields:            fields,
		CustomFieldsAvailable: setup.CustomFieldsAvailable,
		Mapping:               mcpAsanaMapping(mapping),
	}
}

func mcpAsanaMapping(mapping *asanasvc.AsanaProjectMapping) *types.AsanaProjectMapping {
	if mapping == nil {
		return nil
	}

	stateKeys := make([]coredata.TaskState, 0, len(mapping.States))
	for state := range mapping.States {
		stateKeys = append(stateKeys, state)
	}

	slices.Sort(stateKeys)

	states := make([]*types.AsanaStateMapping, 0, len(stateKeys))
	for _, state := range stateKeys {
		section := mapping.States[state]
		states = append(states, &types.AsanaStateMapping{
			State:     state,
			SectionID: section,
			Inbound:   section != "" && mapping.SectionInbound[section] == state,
		})
	}

	priorityKeys := make([]coredata.TaskPriority, 0, len(mapping.Priorities))
	for priority := range mapping.Priorities {
		priorityKeys = append(priorityKeys, priority)
	}

	slices.Sort(priorityKeys)

	priorities := make([]*types.AsanaPriorityMapping, 0, len(priorityKeys))
	for _, priority := range priorityKeys {
		option := mapping.Priorities[priority]
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
