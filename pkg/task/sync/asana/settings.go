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
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	client "go.probo.inc/probo/pkg/task/sync/asana/client"
)

type (
	AsanaSettings struct {
		DefaultProject *client.Named
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
		return AsanaSettings{}, fmt.Errorf("cannot load Asana connectors: %w", err)
	}

	for _, connector := range connectors {
		settings, err := coredata.ConnectorSettings[coredata.AsanaConnectorSettings](connector)
		if err != nil {
			return AsanaSettings{}, fmt.Errorf("cannot read Asana settings: %w", err)
		}

		if settings.DefaultProjectGID == "" {
			continue
		}

		return AsanaSettings{
			DefaultProject: &client.Named{
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
		return fmt.Errorf("cannot load Asana connectors: %w", err)
	}

	if len(connectors) == 0 {
		return ErrAsanaNotConnected
	}

	accounts, err := s.asanaAccounts(ctx, scope, organizationID)
	if err != nil {
		return fmt.Errorf("cannot load Asana accounts: %w", err)
	}

	if update.DefaultProjectID != nil {
		if err := s.updateAsanaDefaultProject(ctx, scope, connectors, accounts, *update.DefaultProjectID, update.ImportNewTasks); err != nil {
			return fmt.Errorf("cannot update Asana default project: %w", err)
		}

		return nil
	}

	if err := s.updateAsanaImport(ctx, scope, accounts, *update.ImportNewTasks); err != nil {
		return fmt.Errorf("cannot update Asana task import: %w", err)
	}

	return nil
}

func (s *Service) updateAsanaDefaultProject(
	ctx context.Context,
	scope coredata.Scoper,
	connectors []*coredata.Connector,
	accounts []asanaAccount,
	projectID *string,
	importNewTasks *bool,
) error {
	clear := projectID == nil || strings.TrimSpace(*projectID) == ""
	if clear {
		if importNewTasks != nil && *importNewTasks {
			return ErrAsanaDefaultProjectRequired
		}

		if err := s.patchAsanaConnectorSettings(
			ctx,
			scope,
			asanaConnectorIDsFromConnectors(connectors),
			func(_ gid.GID, settings *coredata.AsanaConnectorSettings) error {
				settings.DefaultProjectGID = ""
				settings.DefaultProjectName = ""
				settings.ImportNewTasks = false

				return nil
			},
		); err != nil {
			return fmt.Errorf("cannot clear Asana default project: %w", err)
		}

		return nil
	}

	id := strings.TrimSpace(*projectID)

	account, setup, err := s.asanaAccountForProject(ctx, accounts, id)
	if err != nil {
		return fmt.Errorf("cannot resolve Asana project: %w", err)
	}

	chosenID := account.connector.ID

	err = s.patchAsanaConnectorSettings(
		ctx,
		scope,
		asanaConnectorIDs(accounts),
		func(connectorID gid.GID, settings *coredata.AsanaConnectorSettings) error {
			if connectorID != chosenID {
				settings.DefaultProjectGID = ""
				settings.DefaultProjectName = ""
				settings.ImportNewTasks = false

				return nil
			}

			settings.DefaultProjectGID = setup.Project.GID
			settings.DefaultProjectName = setup.Project.Name

			if importNewTasks != nil {
				settings.ImportNewTasks = *importNewTasks
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot store Asana default project: %w", err)
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

	err := s.patchAsanaConnectorSettings(
		ctx,
		scope,
		[]gid.GID{chosen.connector.ID},
		func(_ gid.GID, settings *coredata.AsanaConnectorSettings) error {
			if settings.DefaultProjectGID == "" {
				if active {
					return ErrAsanaDefaultProjectRequired
				}

				return nil
			}

			settings.ImportNewTasks = active

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot store Asana task import: %w", err)
	}

	if active {
		s.ensureAsanaWebhook(ctx, scope, chosen.connector.ID, chosen.settings.DefaultProjectGID)
	}

	return nil
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
		return fmt.Errorf("cannot load Asana accounts: %w", err)
	}

	account, setup, err := s.asanaAccountForProject(ctx, accounts, projectGID)
	if err != nil {
		return fmt.Errorf("cannot resolve Asana project: %w", err)
	}

	if update.Clear {
		if err := s.patchAsanaConnectorSettings(
			ctx,
			scope,
			[]gid.GID{account.connector.ID},
			func(_ gid.GID, settings *coredata.AsanaConnectorSettings) error {
				delete(settings.Mappings, setup.Project.GID)

				return nil
			},
		); err != nil {
			return fmt.Errorf("cannot clear Asana project mapping: %w", err)
		}

		return nil
	}

	if err := s.patchAsanaConnectorSettings(
		ctx,
		scope,
		[]gid.GID{account.connector.ID},
		func(_ gid.GID, settings *coredata.AsanaConnectorSettings) error {
			if settings.Mappings == nil {
				settings.Mappings = map[string]coredata.AsanaProjectMappingSettings{}
			}

			current := settings.Mappings[setup.Project.GID]

			states := update.States
			if !update.StatesSet {
				states = stateInputs(current)
			}

			field := strings.TrimSpace(current.PriorityFieldGID)

			if update.PriorityFieldSet {
				field = ""

				if update.PriorityFieldGID != nil {
					field = strings.TrimSpace(*update.PriorityFieldGID)
				}
			}

			priorities := update.Priorities
			if !update.PrioritiesSet && field == strings.TrimSpace(current.PriorityFieldGID) {
				priorities = priorityInputs(current)
			}

			var options []client.EnumOption

			for _, enumField := range setup.EnumFields {
				if enumField.GID == strings.TrimSpace(field) {
					options = enumField.Options
				}
			}

			mapped, err := BuildAsanaMapping(states, field, priorities, setup.Sections, options)
			if err != nil {
				return fmt.Errorf("cannot build Asana project mapping: %w", err)
			}

			settings.Mappings[setup.Project.GID] = mapped

			return nil
		},
	); err != nil {
		return fmt.Errorf("cannot store Asana project mapping: %w", err)
	}

	s.ensureAsanaWebhook(ctx, scope, account.connector.ID, setup.Project.GID)

	return nil
}

func (s *Service) AsanaProjectSetup(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	projectGID string,
) (client.ProjectSetup, *AsanaProjectMapping, error) {
	accounts, err := s.asanaAccounts(ctx, scope, organizationID)
	if err != nil {
		return client.ProjectSetup{}, nil, fmt.Errorf("cannot load Asana accounts: %w", err)
	}

	account, setup, err := s.asanaAccountForProject(ctx, accounts, projectGID)
	if err != nil {
		return client.ProjectSetup{}, nil, fmt.Errorf("cannot resolve Asana project: %w", err)
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

func asanaConnectorIDsFromConnectors(connectors []*coredata.Connector) []gid.GID {
	ids := make([]gid.GID, 0, len(connectors))
	for _, connector := range connectors {
		ids = append(ids, connector.ID)
	}

	return ids
}

func asanaConnectorIDs(accounts []asanaAccount) []gid.GID {
	ids := make([]gid.GID, 0, len(accounts))
	for i := range accounts {
		ids = append(ids, accounts[i].connector.ID)
	}

	return ids
}

// patchAsanaConnectorSettings reloads each connector under a row lock and
// applies patch to that fresh document. Callers must not write a settings
// snapshot they loaded earlier: a webhook registration and a mapping save
// would otherwise drop each other's fields.
func (s *Service) patchAsanaConnectorSettings(
	ctx context.Context,
	scope coredata.Scoper,
	connectorIDs []gid.GID,
	patch func(gid.GID, *coredata.AsanaConnectorSettings) error,
) error {
	ids := uniqueSortedGIDs(connectorIDs)

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			for _, connectorID := range ids {
				if err := s.saveAsanaConnectorSettings(ctx, tx, scope, connectorID, func(settings *coredata.AsanaConnectorSettings) error {
					return patch(connectorID, settings)
				}); err != nil {
					return fmt.Errorf("cannot store Asana settings: %w", err)
				}
			}

			return nil
		},
	)
}

func (s *Service) saveAsanaConnectorSettings(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	connectorID gid.GID,
	patch func(*coredata.AsanaConnectorSettings) error,
) error {
	connector := &coredata.Connector{ID: connectorID}
	if err := connector.LockByID(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot lock Asana connector: %w", err)
	}

	if err := connector.LoadMetadataByID(ctx, tx, scope, connectorID); err != nil {
		return fmt.Errorf("cannot load Asana connector: %w", err)
	}

	settings, err := coredata.ConnectorSettings[coredata.AsanaConnectorSettings](connector)
	if err != nil {
		return fmt.Errorf("cannot read Asana settings: %w", err)
	}

	if err := patch(&settings); err != nil {
		return fmt.Errorf("cannot apply Asana settings: %w", err)
	}

	if err := connector.SetSettings(&settings); err != nil {
		return fmt.Errorf("cannot store Asana settings: %w", err)
	}

	connector.UpdatedAt = time.Now()
	if err := connector.UpdateSettings(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot store Asana settings: %w", err)
	}

	return nil
}

func uniqueSortedGIDs(ids []gid.GID) []gid.GID {
	seen := map[gid.GID]struct{}{}
	unique := make([]gid.GID, 0, len(ids))

	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	slices.SortFunc(unique, func(a, b gid.GID) int {
		return bytes.Compare(a[:], b[:])
	})

	return unique
}
