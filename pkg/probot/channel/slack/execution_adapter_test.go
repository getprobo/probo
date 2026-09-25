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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

type fakeExecutionBindings struct {
	subject identitybinding.Subject
	binding *identitybinding.Binding
	lookups int
}

func (f *fakeExecutionBindings) Lookup(
	_ context.Context,
	subject identitybinding.Subject,
) (*identitybinding.Binding, error) {
	f.subject = subject
	f.lookups++

	return f.binding, nil
}

func (*fakeExecutionBindings) BindURL(
	context.Context,
	identitybinding.Subject,
	gid.GID,
) (string, error) {
	return "", nil
}

func TestExecutionAdapterPreparesTrustedInboundSlackRun(t *testing.T) {
	t.Parallel()

	pgClient, organizationID, teamID := executionAdapterDatabase(t)
	identityID := gid.New(organizationID.TenantID(), coredata.IdentityEntityType)
	executionID := gid.New(organizationID.TenantID(), coredata.AgentExecutionEntityType)
	coordinates, err := json.Marshal(
		ExecutionSourceCoordinates{
			TeamID:         teamID,
			ChannelID:      "C123",
			ThreadTS:       "123.000",
			MessageTS:      "123.456",
			ExternalUserID: "U123",
		},
	)
	require.NoError(t, err)
	insertThreadSubject(
		t,
		pgClient,
		organizationID,
		"C123",
		"123.000",
		map[string]any{"trusted": "value"},
	)

	bindings := &fakeExecutionBindings{
		binding: &identitybinding.Binding{IdentityID: identityID},
	}
	adapter := newExecutionAdapter(t, pgClient, "", bindings)

	preparedCtx, _, err := adapter.Prepare(
		t.Context(),
		&coredata.AgentExecution{
			ID:                executionID,
			OrganizationID:    organizationID,
			StartAgentName:    "probot",
			SourceCoordinates: coordinates,
		},
		nil,
		nil,
	)
	require.NoError(t, err)

	runContext := agent.RunContextFrom[*probot.RunContext](preparedCtx)
	assert.Equal(t, executionID, runContext.ExecutionID)
	assert.Equal(t, identityID, runContext.IdentityID)
	assert.Equal(t, "C123", runContext.MessageAnchor.ConversationID)
	assert.Equal(t, "123.000", runContext.MessageAnchor.MessageID)
	assert.Equal(t, "123.456", runContext.CurrentMessageID)
	assert.Equal(t, "test", runContext.Capability)
	assert.Equal(t, "TEST_MESSAGE", runContext.MessageType)
	assert.Equal(t, "value", runContext.Attributes["trusted"])
	assert.Equal(t, "U123", bindings.subject.ExternalUserID)

	_, _, err = adapter.Prepare(
		t.Context(),
		&coredata.AgentExecution{
			ID:                executionID,
			OrganizationID:    organizationID,
			StartAgentName:    "probot",
			SourceCoordinates: coordinates,
		},
		nil,
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, 2, bindings.lookups)
}

func TestExecutionAdapterPreparesUnboundDirectSlackRun(t *testing.T) {
	t.Parallel()

	pgClient, organizationID, teamID := executionAdapterDatabase(t)
	coordinates, err := json.Marshal(
		ExecutionSourceCoordinates{
			TeamID:         teamID,
			ChannelID:      "D123",
			MessageTS:      "123.456",
			ExternalUserID: "U123",
		},
	)
	require.NoError(t, err)

	adapter := newExecutionAdapter(
		t,
		pgClient,
		"",
		&fakeExecutionBindings{
			binding: &identitybinding.Binding{
				IdentityID: gid.New(organizationID.TenantID(), coredata.IdentityEntityType),
			},
		},
	)

	preparedCtx, _, err := adapter.Prepare(
		t.Context(),
		&coredata.AgentExecution{
			ID:                gid.New(organizationID.TenantID(), coredata.AgentExecutionEntityType),
			OrganizationID:    organizationID,
			StartAgentName:    "probot",
			SourceCoordinates: coordinates,
		},
		nil,
		nil,
	)
	require.NoError(t, err)

	runContext := agent.RunContextFrom[*probot.RunContext](preparedCtx)
	assert.Empty(t, runContext.Capability)
	assert.Empty(t, runContext.MessageType)
	assert.Equal(t, "D123", runContext.MessageAnchor.ConversationID)
	assert.Equal(t, "123.456", runContext.CurrentMessageID)
}

func TestExecutionAdapterRejectsMalformedSourceCoordinates(t *testing.T) {
	t.Parallel()

	adapter := NewExecutionAdapter(nil, nil, nil, nil, nil, nil, log.NewLogger())
	_, _, err := adapter.Prepare(
		t.Context(),
		&coredata.AgentExecution{SourceCoordinates: json.RawMessage("{")},
		nil,
		nil,
	)
	assert.ErrorContains(t, err, "source coordinates")
}

func TestExecutionAdapterUsesInputIdentityWhenTurnActorMatches(t *testing.T) {
	t.Parallel()

	pgClient, organizationID, teamID := executionAdapterDatabase(t)
	inputIdentityID := gid.New(organizationID.TenantID(), coredata.IdentityEntityType)
	coordinates, err := json.Marshal(
		ExecutionSourceCoordinates{
			TeamID:         teamID,
			ChannelID:      "C123",
			ThreadTS:       "123.000",
			MessageTS:      "123.456",
			ExternalUserID: "U123",
		},
	)
	require.NoError(t, err)

	bindings := &fakeExecutionBindings{
		binding: &identitybinding.Binding{IdentityID: inputIdentityID},
	}
	adapter := newExecutionAdapter(t, pgClient, "", bindings)

	preparedCtx, _, err := adapter.Prepare(
		t.Context(),
		&coredata.AgentExecution{
			ID:                gid.New(organizationID.TenantID(), coredata.AgentExecutionEntityType),
			OrganizationID:    organizationID,
			StartAgentName:    "probot",
			SourceCoordinates: coordinates,
		},
		nil,
		&coredata.AgentInput{IdentityID: &inputIdentityID},
	)
	require.NoError(t, err)

	runContext := agent.RunContextFrom[*probot.RunContext](preparedCtx)
	assert.Equal(t, inputIdentityID, runContext.IdentityID)
	assert.Equal(t, 1, bindings.lookups)
	assert.Equal(t, "U123", bindings.subject.ExternalUserID)
}

func TestExecutionAdapterResolvesTrustFromInputActor(t *testing.T) {
	t.Parallel()

	pgClient, organizationID, teamID := executionAdapterDatabase(t)
	identityID := gid.New(organizationID.TenantID(), coredata.IdentityEntityType)
	executionCoords, err := json.Marshal(
		ExecutionSourceCoordinates{
			TeamID:         teamID,
			ChannelID:      "C123",
			ThreadTS:       "123.000",
			MessageTS:      "123.456",
			ExternalUserID: "U123",
		},
	)
	require.NoError(t, err)
	inputCoords, err := json.Marshal(
		ExecutionSourceCoordinates{
			ActorTeamID:    "TOTHER",
			ChannelID:      "C123",
			ThreadTS:       "123.000",
			MessageTS:      "124.000",
			ExternalUserID: "U999",
		},
	)
	require.NoError(t, err)

	bindings := &fakeExecutionBindings{
		binding: &identitybinding.Binding{IdentityID: identityID},
	}
	adapter := newExecutionAdapter(t, pgClient, "", bindings)

	preparedCtx, _, err := adapter.Prepare(
		t.Context(),
		&coredata.AgentExecution{
			ID:                gid.New(organizationID.TenantID(), coredata.AgentExecutionEntityType),
			OrganizationID:    organizationID,
			StartAgentName:    "probot",
			SourceCoordinates: executionCoords,
		},
		nil,
		&coredata.AgentInput{SourceCoordinates: inputCoords},
	)
	require.NoError(t, err)

	runContext := agent.RunContextFrom[*probot.RunContext](preparedCtx)
	assert.Equal(t, identityID, runContext.IdentityID)
	assert.Equal(t, 1, bindings.lookups)
	assert.Equal(t, "U999", bindings.subject.ExternalUserID)
	assert.Equal(t, "TOTHER", bindings.subject.ExternalTenantID)
	assert.Equal(t, "124.000", runContext.CurrentMessageID)
}

func TestExecutionAdapterRejectsMismatchedInputIdentityAndTurnActor(t *testing.T) {
	t.Parallel()

	pgClient, organizationID, teamID := executionAdapterDatabase(t)
	inputIdentityID := gid.New(organizationID.TenantID(), coredata.IdentityEntityType)
	executionCoords, err := json.Marshal(
		ExecutionSourceCoordinates{
			TeamID:         teamID,
			ChannelID:      "C123",
			ThreadTS:       "123.000",
			MessageTS:      "123.456",
			ExternalUserID: "U123",
		},
	)
	require.NoError(t, err)
	inputCoords, err := json.Marshal(
		ExecutionSourceCoordinates{
			ChannelID:      "C123",
			ThreadTS:       "123.000",
			MessageTS:      "124.000",
			ExternalUserID: "U999",
		},
	)
	require.NoError(t, err)

	adapter := newExecutionAdapter(
		t,
		pgClient,
		"",
		&fakeExecutionBindings{
			binding: &identitybinding.Binding{
				IdentityID: gid.New(organizationID.TenantID(), coredata.IdentityEntityType),
			},
		},
	)

	_, _, err = adapter.Prepare(
		t.Context(),
		&coredata.AgentExecution{
			ID:                gid.New(organizationID.TenantID(), coredata.AgentExecutionEntityType),
			OrganizationID:    organizationID,
			StartAgentName:    "probot",
			SourceCoordinates: executionCoords,
		},
		nil,
		&coredata.AgentInput{
			IdentityID:        &inputIdentityID,
			SourceCoordinates: inputCoords,
		},
	)
	assert.ErrorContains(t, err, "does not match turn actor")
}

func TestExecutionAdapterBindsReactionToInputMessageTS(t *testing.T) {
	t.Parallel()

	pgClient, organizationID, teamID := executionAdapterDatabase(t)
	executionCoords, err := json.Marshal(
		ExecutionSourceCoordinates{
			TeamID:         teamID,
			ChannelID:      "C123",
			ThreadTS:       "900.000",
			MessageTS:      "902.000",
			ExternalUserID: "U123",
		},
	)
	require.NoError(t, err)
	inputCoords, err := json.Marshal(
		ExecutionSourceCoordinates{
			ChannelID: "C123",
			ThreadTS:  "900.000",
			MessageTS: "901.000",
		},
	)
	require.NoError(t, err)

	adapter := newExecutionAdapter(
		t,
		pgClient,
		"",
		&fakeExecutionBindings{
			binding: &identitybinding.Binding{
				IdentityID: gid.New(organizationID.TenantID(), coredata.IdentityEntityType),
			},
		},
	)

	preparedCtx, _, err := adapter.Prepare(
		t.Context(),
		&coredata.AgentExecution{
			ID:                gid.New(organizationID.TenantID(), coredata.AgentExecutionEntityType),
			OrganizationID:    organizationID,
			StartAgentName:    "probot",
			SourceCoordinates: executionCoords,
		},
		nil,
		&coredata.AgentInput{SourceCoordinates: inputCoords},
	)
	require.NoError(t, err)

	runContext := agent.RunContextFrom[*probot.RunContext](preparedCtx)
	assert.Equal(t, "901.000", runContext.CurrentMessageID)
	assert.Equal(t, "900.000", runContext.MessageAnchor.MessageID)
}

func TestExecutionAdapterClearsAssistantStatusForSilentTurn(t *testing.T) {
	t.Parallel()

	slackAPI := newRecordingSlackAPI(t, nil)
	pgClient, organizationID, _ := executionAdapterDatabase(t)
	adapter := newExecutionAdapter(t, pgClient, slackAPI.URL, &fakeExecutionBindings{})

	hook := adapter.assistantStatusHook(
		organizationID,
		ExecutionSourceCoordinates{
			ChannelID: "C123",
			ThreadTS:  "123.000",
			MessageTS: "123.456",
		},
	)
	require.NotNil(t, hook)

	hook.OnRunEnd(t.Context(), nil, nil, nil)

	statuses := slackAPI.statusesSnapshot()
	require.Len(t, statuses, 1)
	assert.Equal(t, "C123", statuses[0]["channel_id"])
	assert.Equal(t, "123.000", statuses[0]["thread_ts"])
	assert.Empty(t, statuses[0]["status"])
	assert.Nil(t, statuses[0]["loading_messages"])
}

func TestExecutionAdapterClearsAssistantStatusOnTriggeringMessage(t *testing.T) {
	t.Parallel()

	pgClient, organizationID, _ := executionAdapterDatabase(t)
	adapter := newExecutionAdapter(t, pgClient, "", &fakeExecutionBindings{})

	hook := adapter.assistantStatusHook(
		organizationID,
		ExecutionSourceCoordinates{ChannelID: "D123", MessageTS: "333.444"},
	)
	require.NotNil(t, hook)
	assert.Equal(t, "333.444", hook.threadTS)

	assert.Nil(
		t,
		adapter.assistantStatusHook(
			organizationID,
			ExecutionSourceCoordinates{MessageTS: "333.444"},
		),
	)
}

func executionAdapterDatabase(t *testing.T) (*pg.Client, gid.GID, string) {
	t.Helper()

	pgClient, _, organizationID := executionIngressDatabase(t)
	teamID := uniqueSlackTeamID(t)
	insertActiveInstallation(t, pgClient, organizationID, teamID, "UBOT")

	return pgClient, organizationID, teamID
}

func newExecutionAdapter(
	t *testing.T,
	pgClient *pg.Client,
	apiBaseURL string,
	bindings identitybinding.Gate,
) *ExecutionAdapter {
	t.Helper()

	return NewExecutionAdapter(
		pgClient,
		newTestInstallationService(t, pgClient, apiBaseURL),
		bindings,
		newTestAgentProfiles(t),
		probot.NewCapabilityRegistry(),
		NewDeliveryService(pgClient),
		log.NewLogger(),
	)
}

func insertThreadSubject(
	t *testing.T,
	pgClient *pg.Client,
	organizationID gid.GID,
	channelID string,
	threadTS string,
	attributes map[string]any,
) {
	t.Helper()

	raw, err := json.Marshal(attributes)
	require.NoError(t, err)

	now := time.Now()
	scope := coredata.NewScopeFromObjectID(organizationID)

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				subject := &coredata.BotThreadSubject{
					ID:                     gid.New(scope.GetTenantID(), coredata.BotThreadSubjectEntityType),
					OrganizationID:         organizationID,
					Provider:               ProviderName,
					ExternalConversationID: channelID,
					ExternalMessageID:      threadTS,
					Capability:             "test",
					MessageType:            "TEST_MESSAGE",
					Attributes:             raw,
					SubjectNamespace:       "subject",
					SubjectKey:             "stable",
					CreatedAt:              now,
					UpdatedAt:              now,
				}
				_, err := subject.Upsert(ctx, tx, scope)

				return err
			},
		),
	)
}
