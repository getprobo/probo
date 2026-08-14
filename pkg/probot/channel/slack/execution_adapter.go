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

package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

type (
	ExecutionSourceCoordinates struct {
		TeamID         string `json:"team_id,omitempty"`
		ActorTeamID    string `json:"actor_team_id,omitempty"`
		ChannelID      string `json:"channel_id,omitempty"`
		ThreadTS       string `json:"thread_ts,omitempty"`
		MessageTS      string `json:"message_ts,omitempty"`
		ExternalUserID string `json:"external_user_id,omitempty"`
	}

	threadSubject struct {
		Capability  string
		MessageType string
		Attributes  map[string]any
	}

	ExecutionAdapter struct {
		pg            *pg.Client
		installations *InstallationService
		bindings      identitybinding.Gate
		profiles      *probot.AgentProfileRegistry
		capabilities  *probot.CapabilityRegistry
		deliveries    *DeliveryService
		logger        *log.Logger
	}
)

var (
	_ probot.ExecutionAdapter = (*ExecutionAdapter)(nil)
)

func NewExecutionAdapter(
	pgClient *pg.Client,
	installations *InstallationService,
	bindings identitybinding.Gate,
	profiles *probot.AgentProfileRegistry,
	capabilities *probot.CapabilityRegistry,
	deliveries *DeliveryService,
	logger *log.Logger,
) *ExecutionAdapter {
	return &ExecutionAdapter{
		pg:            pgClient,
		installations: installations,
		bindings:      bindings,
		profiles:      profiles,
		capabilities:  capabilities,
		deliveries:    deliveries,
		logger:        logger,
	}
}

func (a *ExecutionAdapter) Provider() string {
	return ProviderName
}

// Prepare hydrates run context from the live BotThreadSubject and this
// input's turn coordinates. It runs on every Process batch, including
// checkpoint resume, so late subject binding can enable capability tools.
func (a *ExecutionAdapter) Prepare(
	ctx context.Context,
	execution *coredata.AgentExecution,
	_ agent.AgentRegistry,
	input *coredata.AgentInput,
) (context.Context, agent.AgentRegistry, error) {
	coordinates, err := decodeExecutionCoordinates(execution.SourceCoordinates)
	if err != nil {
		return nil, nil, err
	}

	turn, err := turnCoordinates(coordinates, input)
	if err != nil {
		return nil, nil, err
	}

	installation, identityID, err := a.resolveTrust(ctx, execution, coordinates, input)
	if err != nil {
		return nil, nil, err
	}

	if installation.OrganizationID != execution.OrganizationID {
		return nil, nil, fmt.Errorf("slack installation does not belong to execution organization")
	}

	subject, err := a.loadThreadSubject(ctx, execution.OrganizationID, turn)
	if err != nil {
		return nil, nil, err
	}

	runContext := hydrateRunContext(execution, turn, identityID, subject)
	registry, err := a.registryForRun(execution.OrganizationID, execution.StartAgentName, turn, subject)
	if err != nil {
		return nil, nil, err
	}

	return agent.WithRunContext(ctx, runContext), registry, nil
}

func hydrateRunContext(
	execution *coredata.AgentExecution,
	turn ExecutionSourceCoordinates,
	identityID gid.GID,
	subject threadSubject,
) *probot.RunContext {
	return &probot.RunContext{
		OrganizationID: execution.OrganizationID,
		ExecutionID:    execution.ID,
		MessageAnchor: probot.MessageAnchor{
			ConversationID: turn.ChannelID,
			MessageID:      turn.ThreadTS,
		},
		CurrentMessageID: turn.MessageTS,
		IdentityID:       identityID,
		Capability:       subject.Capability,
		MessageType:      subject.MessageType,
		Attributes:       cloneNotificationData(subject.Attributes),
	}
}

func (a *ExecutionAdapter) registryForRun(
	organizationID gid.GID,
	startAgentName string,
	turn ExecutionSourceCoordinates,
	subject threadSubject,
) (agent.AgentRegistry, error) {
	base, err := a.profiles.Agent(startAgentName)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve Slack execution agent profile: %w", err)
	}

	tools := []agent.Tool{}
	if subject.MessageType != "" {
		tools = append(tools, a.capabilities.ToolsForMessageType(subject.MessageType)...)
	}

	if turn.ChannelID != "" {
		tools = append(
			tools,
			Tools(
				a.deliveries,
				TurnBinding{
					OrganizationID: organizationID,
					ChannelID:      turn.ChannelID,
					ThreadTS:       turn.ThreadTS,
					MessageTS:      turn.MessageTS,
				},
			)...,
		)
	}

	return probot.NewStaticAgentRegistry(startAgentName, base.Clone(agent.WithTools(tools...))), nil
}

func (a *ExecutionAdapter) loadThreadSubject(
	ctx context.Context,
	organizationID gid.GID,
	turn ExecutionSourceCoordinates,
) (threadSubject, error) {
	if a.pg == nil || turn.ChannelID == "" || turn.ThreadTS == "" {
		return threadSubject{}, nil
	}

	var subject coredata.BotThreadSubject

	err := a.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return subject.LoadByProviderCoordinates(
				ctx,
				conn,
				coredata.NewScopeFromObjectID(organizationID),
				organizationID,
				ProviderName,
				turn.ChannelID,
				turn.ThreadTS,
			)
		},
	)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		return threadSubject{}, nil
	}

	if err != nil {
		return threadSubject{}, fmt.Errorf("cannot load Slack thread subject: %w", err)
	}

	if (subject.Capability == "") != (subject.MessageType == "") {
		return threadSubject{}, fmt.Errorf("slack thread subject capability context is incomplete")
	}

	var attributes map[string]any
	if len(subject.Attributes) > 0 {
		if err := json.Unmarshal(subject.Attributes, &attributes); err != nil {
			return threadSubject{}, fmt.Errorf("cannot decode Slack thread subject attributes: %w", err)
		}
	}

	return threadSubject{
		Capability:  subject.Capability,
		MessageType: subject.MessageType,
		Attributes:  attributes,
	}, nil
}

func (a *ExecutionAdapter) resolveTrust(
	ctx context.Context,
	execution *coredata.AgentExecution,
	coordinates ExecutionSourceCoordinates,
	input *coredata.AgentInput,
) (*coredata.SlackbotInstallation, gid.GID, error) {
	if coordinates.TeamID == "" && coordinates.ExternalUserID == "" {
		installation, err := a.installations.GetByOrganizationID(
			ctx,
			coredata.NewScopeFromObjectID(execution.ID),
			execution.OrganizationID,
		)
		if err != nil {
			return nil, gid.Nil, fmt.Errorf("cannot revalidate Slack installation: %w", err)
		}

		if installation.Status != coredata.SlackbotInstallationStatusActive {
			return nil, gid.Nil, ErrSlackbotNotInstalled
		}

		return installation, identityIDFromInput(input), nil
	}

	if coordinates.TeamID == "" || coordinates.ChannelID == "" || coordinates.MessageTS == "" {
		return nil, gid.Nil, fmt.Errorf("inbound Slack execution coordinates are incomplete")
	}

	installation, err := a.installations.GetByTeamID(ctx, coordinates.TeamID)
	if err != nil {
		return nil, gid.Nil, fmt.Errorf("cannot revalidate inbound Slack installation: %w", err)
	}

	if identityID := identityIDFromInput(input); identityID != gid.Nil {
		return installation, identityID, nil
	}

	if coordinates.ExternalUserID == "" {
		return nil, gid.Nil, fmt.Errorf("inbound Slack execution coordinates are incomplete")
	}

	binding, err := a.bindings.Lookup(
		ctx,
		identitybinding.Subject{
			Provider:         ProviderName,
			ExternalTenantID: externalActorTeamID(coordinates),
			ExternalUserID:   coordinates.ExternalUserID,
		},
	)
	if err != nil {
		return nil, gid.Nil, fmt.Errorf("cannot revalidate inbound Slack identity binding: %w", err)
	}

	return installation, binding.IdentityID, nil
}

func decodeExecutionCoordinates(raw json.RawMessage) (ExecutionSourceCoordinates, error) {
	if len(raw) == 0 {
		return ExecutionSourceCoordinates{}, nil
	}

	var coordinates ExecutionSourceCoordinates
	if err := json.Unmarshal(raw, &coordinates); err != nil {
		return ExecutionSourceCoordinates{}, fmt.Errorf("cannot decode Slack source coordinates: %w", err)
	}

	return coordinates, nil
}

func turnCoordinates(
	executionCoords ExecutionSourceCoordinates,
	input *coredata.AgentInput,
) (ExecutionSourceCoordinates, error) {
	turn := executionCoords
	if input == nil || len(input.SourceCoordinates) == 0 {
		return turn, nil
	}

	inputCoords, err := decodeExecutionCoordinates(input.SourceCoordinates)
	if err != nil {
		return ExecutionSourceCoordinates{}, fmt.Errorf("cannot decode Slack input source coordinates: %w", err)
	}

	if inputCoords.ChannelID != "" {
		turn.ChannelID = inputCoords.ChannelID
	}

	if inputCoords.ThreadTS != "" {
		turn.ThreadTS = inputCoords.ThreadTS
	}

	if inputCoords.MessageTS != "" {
		turn.MessageTS = inputCoords.MessageTS
	}

	return turn, nil
}

func identityIDFromInput(input *coredata.AgentInput) gid.GID {
	if input == nil || input.IdentityID == nil {
		return gid.Nil
	}

	return *input.IdentityID
}

func externalActorTeamID(coordinates ExecutionSourceCoordinates) string {
	if coordinates.ActorTeamID != "" {
		return coordinates.ActorTeamID
	}

	return coordinates.TeamID
}
