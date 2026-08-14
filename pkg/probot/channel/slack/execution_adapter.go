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
	"fmt"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/bot"
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

	ExecutionAdapter struct {
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
	installations *InstallationService,
	bindings identitybinding.Gate,
	profiles *probot.AgentProfileRegistry,
	capabilities *probot.CapabilityRegistry,
	deliveries *DeliveryService,
	logger *log.Logger,
) *ExecutionAdapter {
	return &ExecutionAdapter{
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

func (a *ExecutionAdapter) Prepare(
	ctx context.Context,
	execution *coredata.AgentExecution,
	_ agent.AgentRegistry,
	input *coredata.AgentInput,
) (context.Context, agent.AgentRegistry, error) {
	var trusted bot.ConversationTrustedContext
	if len(execution.TrustedContext) > 0 {
		if err := json.Unmarshal(execution.TrustedContext, &trusted); err != nil {
			return nil, nil, fmt.Errorf("cannot decode Slack execution trusted context: %w", err)
		}

		if (trusted.Capability == "") != (trusted.MessageType == "") {
			return nil, nil, fmt.Errorf("slack execution trusted capability context is incomplete")
		}
	}

	var coordinates ExecutionSourceCoordinates
	if len(execution.SourceCoordinates) > 0 {
		if err := json.Unmarshal(execution.SourceCoordinates, &coordinates); err != nil {
			return nil, nil, fmt.Errorf("cannot decode Slack execution source coordinates: %w", err)
		}
	}

	installation, identityID, err := a.resolveTrust(ctx, execution, coordinates, input)
	if err != nil {
		return nil, nil, err
	}

	if installation.OrganizationID != execution.OrganizationID {
		return nil, nil, fmt.Errorf("slack installation does not belong to execution organization")
	}

	base, err := a.profiles.Agent(execution.StartAgentName)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot resolve Slack execution agent profile: %w", err)
	}

	tools := []agent.Tool{}
	if trusted.MessageType != "" {
		tools = append(tools, a.capabilities.ToolsForMessageType(trusted.MessageType)...)
	}

	if coordinates.ChannelID != "" {
		tools = append(tools, Tools(a.deliveries)...)
	}

	prepared := base.Clone(agent.WithTools(tools...))
	runContext := &probot.RunContext{
		OrganizationID: execution.OrganizationID,
		ExecutionID:    execution.ID,
		MessageAnchor: probot.MessageAnchor{
			ConversationID: coordinates.ChannelID,
			MessageID:      coordinates.ThreadTS,
		},
		CurrentMessageID: coordinates.MessageTS,
		IdentityID:       identityID,
		Capability:       trusted.Capability,
		MessageType:      trusted.MessageType,
		Attributes:       cloneNotificationData(trusted.Attributes),
	}

	a.refreshAssistantStatus(ctx, coordinates, execution.OrganizationID)

	return agent.WithRunContext(ctx, runContext),
		probot.NewStaticAgentRegistry(execution.StartAgentName, prepared),
		nil
}

func (a *ExecutionAdapter) refreshAssistantStatus(
	ctx context.Context,
	coordinates ExecutionSourceCoordinates,
	organizationID gid.GID,
) {
	threadTS := assistantStatusThreadTS(coordinates.ThreadTS, coordinates.MessageTS)
	if coordinates.ChannelID == "" || threadTS == "" || a.installations == nil {
		return
	}

	var (
		client *Client
		err    error
	)
	if coordinates.TeamID != "" {
		client, _, err = a.installations.ClientByTeamID(ctx, coordinates.TeamID)
	} else {
		client, _, err = a.installations.ClientByOrganizationID(
			ctx,
			coredata.NewScopeFromObjectID(organizationID),
			organizationID,
		)
	}

	if err != nil {
		logAssistantStatusError(ctx, a.logger, err)

		return
	}

	if client == nil {
		return
	}

	setAssistantWorkingStatus(ctx, a.logger, client, coordinates.ChannelID, threadTS)
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
