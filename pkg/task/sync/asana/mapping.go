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

package asana

import (
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
	client "go.probo.inc/probo/pkg/task/sync/asana/client"
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
	if OutboundSection(current, mapping) == "" {
		return asanaStateFromCompleted(completed, current)
	}

	return inboundStateFromSection(sectionGID, completed, current, mapping)
}

// Import passes a placeholder state, so the remote section still applies.
func inboundStateFromSection(
	sectionGID string,
	completed bool,
	current coredata.TaskState,
	mapping *AsanaProjectMapping,
) coredata.TaskState {
	sameSection := OutboundSection(current, mapping) == sectionGID
	if sameSection && CompletedForState(current) == completed {
		return current
	}

	if mapping != nil && sectionGID != "" {
		if state, ok := mapping.SectionInbound[sectionGID]; ok && state.IsValid() {
			return state
		}
	}

	return asanaStateFromCompleted(completed, current)
}

func asanaStateFromCompleted(completed bool, current coredata.TaskState) coredata.TaskState {
	if completed {
		if CompletedForState(current) {
			return current
		}

		return coredata.TaskStateDone
	}

	if CompletedForState(current) {
		return coredata.TaskStateTodo
	}

	return current
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
	if _, ok := OutboundPriorityOption(current, mapping); !ok {
		return current, false
	}

	return inboundPriorityFromOption(optionGID, current, mapping)
}

func inboundPriorityFromOption(
	optionGID string,
	current coredata.TaskPriority,
	mapping *AsanaProjectMapping,
) (coredata.TaskPriority, bool) {
	if option, ok := OutboundPriorityOption(current, mapping); ok && option == optionGID {
		return current, true
	}

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
	sections []client.Section,
	options []client.EnumOption,
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
			return settings, fmt.Errorf("cannot map Asana section: %w", err)
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
			return settings, fmt.Errorf("cannot map Asana priority: %w", err)
		}

		settings.OptionInbound[optionGID] = inbound.String()
	}

	return settings, nil
}

func inboundState(group []AsanaStateInput) (coredata.TaskState, error) {
	present := map[coredata.TaskState]struct{}{}

	for _, input := range group {
		present[input.State] = struct{}{}
	}

	for _, state := range coredata.TaskStates() {
		if _, ok := present[state]; ok {
			return state, nil
		}
	}

	return "", mappingError("states", "state is invalid")
}

func inboundPriority(group []AsanaPriorityInput) (coredata.TaskPriority, error) {
	present := map[coredata.TaskPriority]struct{}{}

	for _, input := range group {
		present[input.Priority] = struct{}{}
	}

	for _, priority := range coredata.TaskPriorities() {
		if _, ok := present[priority]; ok {
			return priority, nil
		}
	}

	return "", mappingError("priorities", "priority is invalid")
}

func mappingError(field, message string) error {
	return validator.ValidationErrors{&validator.ValidationError{
		Field:   field,
		Code:    validator.ErrorCodeCustom,
		Message: message,
	}}
}

func sectionForProject(task client.Task, projectGID string) string {
	for _, membership := range task.Memberships {
		if membership.ProjectGID == projectGID {
			return membership.SectionGID
		}
	}

	return ""
}

func enumOption(task client.Task, fieldGID string) string {
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
