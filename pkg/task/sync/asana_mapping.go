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

package tasksync

import (
	"strings"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/task/sync/asana"
	"go.probo.inc/probo/pkg/validator"
)

type (
	AsanaProjectMapping struct {
		States           map[coredata.TaskState]string
		SectionInbound   map[string]coredata.TaskState
		PriorityFieldGID string
		Priorities       map[coredata.TaskPriority]string
		OptionInbound    map[string]coredata.TaskPriority
	}

	AsanaStateInput struct {
		State      coredata.TaskState
		SectionGID string
		Inbound    bool
	}

	AsanaPriorityInput struct {
		Priority  coredata.TaskPriority
		OptionGID string
		Inbound   bool
	}
)

func CompletedForState(state coredata.TaskState) bool {
	switch state {
	case coredata.TaskStateDone, coredata.TaskStateCanceled, coredata.TaskStateDuplicate:
		return true
	default:
		return false
	}
}

func InboundState(
	sectionGID string,
	completed bool,
	current coredata.TaskState,
	mapping *AsanaProjectMapping,
) coredata.TaskState {
	if mapping != nil && sectionGID != "" {
		if state, ok := mapping.SectionInbound[sectionGID]; ok && state.IsValid() {
			return state
		}
	}

	if completed {
		switch current {
		case coredata.TaskStateDone, coredata.TaskStateCanceled, coredata.TaskStateDuplicate:
			return current
		default:
			return coredata.TaskStateDone
		}
	}

	switch current {
	case coredata.TaskStateDone, coredata.TaskStateCanceled, coredata.TaskStateDuplicate:
		return coredata.TaskStateTodo
	default:
		return current
	}
}

func OutboundSection(state coredata.TaskState, mapping *AsanaProjectMapping) string {
	if mapping == nil {
		return ""
	}

	return strings.TrimSpace(mapping.States[state])
}

func InboundPriority(
	optionGID string,
	current coredata.TaskPriority,
	mapping *AsanaProjectMapping,
) (coredata.TaskPriority, bool) {
	if mapping == nil || mapping.PriorityFieldGID == "" || optionGID == "" {
		return current, false
	}

	priority, ok := mapping.OptionInbound[optionGID]
	if !ok || !priority.IsValid() {
		return current, false
	}

	return priority, true
}

func OutboundPriorityOption(priority coredata.TaskPriority, mapping *AsanaProjectMapping) (string, bool) {
	if mapping == nil || mapping.PriorityFieldGID == "" {
		return "", false
	}

	option := strings.TrimSpace(mapping.Priorities[priority])
	if option == "" {
		return "", false
	}

	return option, true
}

func mappingFromSettings(raw coredata.AsanaProjectMappingSettings) *AsanaProjectMapping {
	mapping := &AsanaProjectMapping{
		States:           map[coredata.TaskState]string{},
		SectionInbound:   map[string]coredata.TaskState{},
		PriorityFieldGID: raw.PriorityFieldGID,
		Priorities:       map[coredata.TaskPriority]string{},
		OptionInbound:    map[string]coredata.TaskPriority{},
	}

	for state, section := range raw.States {
		parsed := coredata.TaskState(state)
		if parsed.IsValid() {
			mapping.States[parsed] = section
		}
	}

	for section, state := range raw.SectionInbound {
		parsed := coredata.TaskState(state)
		if section != "" && parsed.IsValid() {
			mapping.SectionInbound[section] = parsed
		}
	}

	for priority, option := range raw.Priorities {
		parsed := coredata.TaskPriority(priority)
		if parsed.IsValid() {
			mapping.Priorities[parsed] = option
		}
	}

	for option, priority := range raw.OptionInbound {
		parsed := coredata.TaskPriority(priority)
		if option != "" && parsed.IsValid() {
			mapping.OptionInbound[option] = parsed
		}
	}

	if len(mapping.States) == 0 && mapping.PriorityFieldGID == "" {
		return nil
	}

	return mapping
}

func BuildAsanaMapping(
	states []AsanaStateInput,
	priorityFieldGID string,
	priorities []AsanaPriorityInput,
	sections []asana.Section,
	options []asana.EnumOption,
) (coredata.AsanaProjectMappingSettings, error) {
	sectionIDs := map[string]struct{}{}
	for _, section := range sections {
		sectionIDs[section.GID] = struct{}{}
	}

	optionIDs := map[string]struct{}{}
	for _, option := range options {
		optionIDs[option.GID] = struct{}{}
	}

	settings := coredata.AsanaProjectMappingSettings{
		States:           map[string]string{},
		SectionInbound:   map[string]string{},
		PriorityFieldGID: strings.TrimSpace(priorityFieldGID),
		Priorities:       map[string]string{},
		OptionInbound:    map[string]string{},
	}

	bySection := map[string][]AsanaStateInput{}

	for _, input := range states {
		if !input.State.IsValid() {
			return settings, mappingError("states", "state is invalid")
		}

		sectionGID := strings.TrimSpace(input.SectionGID)
		if sectionGID != "" {
			if _, ok := sectionIDs[sectionGID]; !ok {
				return settings, mappingError("states", "section was not found in the project")
			}
		}

		settings.States[input.State.String()] = sectionGID
		if sectionGID != "" {
			bySection[sectionGID] = append(bySection[sectionGID], input)
		}
	}

	for sectionGID, group := range bySection {
		inbound, err := inboundState(group)
		if err != nil {
			return settings, err
		}

		settings.SectionInbound[sectionGID] = inbound.String()
	}

	if settings.PriorityFieldGID == "" {
		if len(priorities) > 0 {
			return settings, mappingError("priority_field_id", "a priority field is required")
		}

		settings.Priorities = nil
		settings.OptionInbound = nil

		return settings, nil
	}

	if len(optionIDs) == 0 {
		return settings, mappingError("priority_field_id", "priority field was not found in the project")
	}

	byOption := map[string][]AsanaPriorityInput{}

	for _, input := range priorities {
		if !input.Priority.IsValid() {
			return settings, mappingError("priorities", "priority is invalid")
		}

		optionGID := strings.TrimSpace(input.OptionGID)
		if optionGID != "" {
			if _, ok := optionIDs[optionGID]; !ok {
				return settings, mappingError("priorities", "option was not found on the priority field")
			}
		}

		settings.Priorities[input.Priority.String()] = optionGID
		if optionGID != "" {
			byOption[optionGID] = append(byOption[optionGID], input)
		}
	}

	for optionGID, group := range byOption {
		inbound, err := inboundPriority(group)
		if err != nil {
			return settings, err
		}

		settings.OptionInbound[optionGID] = inbound.String()
	}

	return settings, nil
}

func inboundState(group []AsanaStateInput) (coredata.TaskState, error) {
	if len(group) == 1 {
		return group[0].State, nil
	}

	var chosen *coredata.TaskState

	for _, input := range group {
		if !input.Inbound {
			continue
		}

		if chosen != nil {
			return "", mappingError("states", "a shared section can return one state")
		}

		state := input.State
		chosen = &state
	}

	if chosen == nil {
		return "", mappingError("states", "a shared section needs the state incoming tasks should use")
	}

	return *chosen, nil
}

func inboundPriority(group []AsanaPriorityInput) (coredata.TaskPriority, error) {
	if len(group) == 1 {
		return group[0].Priority, nil
	}

	var chosen *coredata.TaskPriority

	for _, input := range group {
		if !input.Inbound {
			continue
		}

		if chosen != nil {
			return "", mappingError("priorities", "a shared option can return one priority")
		}

		priority := input.Priority
		chosen = &priority
	}

	if chosen == nil {
		return "", mappingError("priorities", "a shared option needs the priority incoming tasks should use")
	}

	return *chosen, nil
}

func mappingError(field, message string) error {
	return validator.ValidationErrors{&validator.ValidationError{
		Field:   field,
		Code:    validator.ErrorCodeCustom,
		Message: message,
	}}
}

func sectionForProject(task asana.Task, projectGID string) string {
	for _, membership := range task.Memberships {
		if membership.ProjectGID == projectGID {
			return membership.SectionGID
		}
	}

	return ""
}

func enumOption(task asana.Task, fieldGID string) string {
	if fieldGID == "" {
		return ""
	}

	for _, field := range task.CustomFields {
		if field.GID == fieldGID {
			return field.EnumOptionGID
		}
	}

	return ""
}

func projectMapping(settings coredata.AsanaConnectorSettings, projectGID string) *AsanaProjectMapping {
	if settings.Mappings == nil {
		return nil
	}

	raw, ok := settings.Mappings[projectGID]
	if !ok {
		return nil
	}

	return mappingFromSettings(raw)
}
