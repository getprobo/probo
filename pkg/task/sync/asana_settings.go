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
	"context"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/task/sync/asana"
)

type (
	AsanaSettings struct {
		DefaultProject *asana.Named
		ImportNewTasks bool
	}

	AsanaSettingsUpdate struct {
		DefaultProjectID **string
		ImportNewTasks   *bool
	}

	AsanaMappingUpdate struct {
		ProjectGID       string
		Clear            bool
		StatesSet        bool
		States           []AsanaStateInput
		PriorityFieldSet bool
		PriorityFieldGID *string
		PrioritiesSet    bool
		Priorities       []AsanaPriorityInput
	}
)

func (s *Service) AsanaSettings(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (AsanaSettings, error) {
	connectors, err := s.loadAsanaConnectors(ctx, scope, organizationID)
	if err != nil {
		return AsanaSettings{}, err
	}

	for _, connector := range connectors {
		settings, err := coredata.ConnectorSettings[coredata.AsanaConnectorSettings](connector)
		if err != nil {
			return AsanaSettings{}, err
		}

		if settings.DefaultProjectGID == "" {
			continue
		}

		return AsanaSettings{
			DefaultProject: &asana.Named{
				GID:  settings.DefaultProjectGID,
				Name: settings.DefaultProjectName,
			},
			ImportNewTasks: settings.ImportNewTasks,
		}, nil
	}

	return AsanaSettings{}, nil
}

func (s *Service) UpdateAsanaSettings(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	update AsanaSettingsUpdate,
) error {
	if update.DefaultProjectID == nil && update.ImportNewTasks == nil {
		return ErrAsanaSettingsRequired
	}

	connectors, err := s.loadAsanaConnectors(ctx, scope, organizationID)
	if err != nil {
		return err
	}

	if len(connectors) == 0 {
		return ErrAsanaNotConnected
	}

	accounts, err := s.asanaAccounts(ctx, scope, organizationID)
	if err != nil {
		return err
	}

	if update.DefaultProjectID != nil {
		if err := s.updateAsanaDefaultProject(ctx, scope, accounts, *update.DefaultProjectID, update.ImportNewTasks); err != nil {
			return err
		}

		return nil
	}

	return s.updateAsanaImport(ctx, scope, accounts, *update.ImportNewTasks)
}

func (s *Service) updateAsanaDefaultProject(
	ctx context.Context,
	scope coredata.Scoper,
	accounts []asanaAccount,
	projectID *string,
	importNewTasks *bool,
) error {
	clear := projectID == nil || strings.TrimSpace(*projectID) == ""
	if clear {
		if importNewTasks != nil && *importNewTasks {
			return ErrAsanaDefaultProjectRequired
		}

		for i := range accounts {
			accounts[i].settings.DefaultProjectGID = ""
			accounts[i].settings.DefaultProjectName = ""

			accounts[i].settings.ImportNewTasks = false
			if err := accounts[i].connector.SetSettings(&accounts[i].settings); err != nil {
				return err
			}
		}

		return s.persistAsanaAccounts(ctx, scope, accounts)
	}

	id := strings.TrimSpace(*projectID)

	account, setup, err := s.asanaAccountForProject(ctx, accounts, id)
	if err != nil {
		return err
	}

	for i := range accounts {
		if accounts[i].connector.ID == account.connector.ID {
			accounts[i].settings.DefaultProjectGID = setup.Project.GID

			accounts[i].settings.DefaultProjectName = setup.Project.Name
			if importNewTasks != nil {
				accounts[i].settings.ImportNewTasks = *importNewTasks
			}

			continue
		}

		accounts[i].settings.DefaultProjectGID = ""
		accounts[i].settings.DefaultProjectName = ""
		accounts[i].settings.ImportNewTasks = false
	}

	if err := s.storeAsanaAccounts(accounts); err != nil {
		return err
	}

	if err := s.persistAsanaAccounts(ctx, scope, accounts); err != nil {
		return err
	}

	s.ensureAsanaWebhook(ctx, scope, account.connector.ID, setup.Project.GID)

	return nil
}

func (s *Service) updateAsanaImport(
	ctx context.Context,
	scope coredata.Scoper,
	accounts []asanaAccount,
	active bool,
) error {
	var chosen *asanaAccount

	for i := range accounts {
		if accounts[i].settings.DefaultProjectGID != "" {
			chosen = &accounts[i]

			break
		}
	}

	if chosen == nil {
		if active {
			return ErrAsanaDefaultProjectRequired
		}

		return nil
	}

	if active && chosen.settings.DefaultProjectGID == "" {
		return ErrAsanaDefaultProjectRequired
	}

	chosen.settings.ImportNewTasks = active
	if err := chosen.connector.SetSettings(&chosen.settings); err != nil {
		return err
	}

	return s.persistAsanaAccounts(ctx, scope, []asanaAccount{*chosen})
}

func (s *Service) UpdateAsanaProjectMapping(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	update AsanaMappingUpdate,
) error {
	projectGID := strings.TrimSpace(update.ProjectGID)
	if projectGID == "" {
		return ErrAsanaProjectIDRequired
	}

	accounts, err := s.asanaAccounts(ctx, scope, organizationID)
	if err != nil {
		return err
	}

	account, setup, err := s.asanaAccountForProject(ctx, accounts, projectGID)
	if err != nil {
		return err
	}

	if account.settings.Mappings == nil {
		account.settings.Mappings = map[string]coredata.AsanaProjectMappingSettings{}
	}

	if update.Clear {
		delete(account.settings.Mappings, setup.Project.GID)

		if err := account.connector.SetSettings(&account.settings); err != nil {
			return err
		}

		return s.persistAsanaAccounts(ctx, scope, []asanaAccount{account})
	}

	current := account.settings.Mappings[setup.Project.GID]

	states := update.States
	if !update.StatesSet {
		states = stateInputs(current)
	}

	field := current.PriorityFieldGID
	if update.PriorityFieldSet {
		if update.PriorityFieldGID == nil {
			field = ""
		} else {
			field = *update.PriorityFieldGID
		}
	}

	priorities := update.Priorities
	if !update.PrioritiesSet {
		priorities = priorityInputs(current)
	}

	var options []asana.EnumOption

	for _, enumField := range setup.EnumFields {
		if enumField.GID == strings.TrimSpace(field) {
			options = enumField.Options
		}
	}

	mapped, err := BuildAsanaMapping(states, field, priorities, setup.Sections, options)
	if err != nil {
		return err
	}

	account.settings.Mappings[setup.Project.GID] = mapped
	if err := account.connector.SetSettings(&account.settings); err != nil {
		return err
	}

	if err := s.persistAsanaAccounts(ctx, scope, []asanaAccount{account}); err != nil {
		return err
	}

	s.ensureAsanaWebhook(ctx, scope, account.connector.ID, setup.Project.GID)

	return nil
}

func (s *Service) AsanaProjectSetup(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	projectGID string,
) (asana.ProjectSetup, *AsanaProjectMapping, error) {
	accounts, err := s.asanaAccounts(ctx, scope, organizationID)
	if err != nil {
		return asana.ProjectSetup{}, nil, err
	}

	account, setup, err := s.asanaAccountForProject(ctx, accounts, projectGID)
	if err != nil {
		return asana.ProjectSetup{}, nil, err
	}

	return setup, projectMapping(account.settings, setup.Project.GID), nil
}

func stateInputs(raw coredata.AsanaProjectMappingSettings) []AsanaStateInput {
	inboundBySection := map[string]coredata.TaskState{}
	for section, state := range raw.SectionInbound {
		inboundBySection[section] = coredata.TaskState(state)
	}

	inputs := make([]AsanaStateInput, 0, len(raw.States))
	for state, section := range raw.States {
		parsed := coredata.TaskState(state)
		inputs = append(inputs, AsanaStateInput{
			State:      parsed,
			SectionGID: section,
			Inbound:    section != "" && inboundBySection[section] == parsed,
		})
	}

	return inputs
}

func priorityInputs(raw coredata.AsanaProjectMappingSettings) []AsanaPriorityInput {
	inboundByOption := map[string]coredata.TaskPriority{}
	for option, priority := range raw.OptionInbound {
		inboundByOption[option] = coredata.TaskPriority(priority)
	}

	inputs := make([]AsanaPriorityInput, 0, len(raw.Priorities))
	for priority, option := range raw.Priorities {
		parsed := coredata.TaskPriority(priority)
		inputs = append(inputs, AsanaPriorityInput{
			Priority:  parsed,
			OptionGID: option,
			Inbound:   option != "" && inboundByOption[option] == parsed,
		})
	}

	return inputs
}

func (s *Service) storeAsanaAccounts(accounts []asanaAccount) error {
	for i := range accounts {
		if err := accounts[i].connector.SetSettings(&accounts[i].settings); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) persistAsanaAccounts(
	ctx context.Context,
	scope coredata.Scoper,
	accounts []asanaAccount,
) error {
	connectors := make([]*coredata.Connector, 0, len(accounts))
	for i := range accounts {
		connectors = append(connectors, accounts[i].connector)
	}

	return s.persistConnectorSettings(ctx, scope, connectors)
}
