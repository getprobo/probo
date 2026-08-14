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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/bot"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
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
	trusted, err := json.Marshal(
		bot.ConversationTrustedContext{
			Capability:  "test",
			MessageType: "TEST_MESSAGE",
			Attributes:  map[string]any{"trusted": "value"},
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
			ID:                executionID,
			OrganizationID:    organizationID,
			StartAgentName:    "probot",
			SourceCoordinates: coordinates,
			TrustedContext:    trusted,
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
	assert.Equal(t, "value", runContext.Attributes["trusted"])
	assert.Equal(t, "U123", bindings.subject.ExternalUserID)

	_, _, err = adapter.Prepare(
		t.Context(),
		&coredata.AgentExecution{
			ID:                executionID,
			OrganizationID:    organizationID,
			StartAgentName:    "probot",
			SourceCoordinates: coordinates,
			TrustedContext:    trusted,
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

func TestExecutionAdapterRejectsMalformedTrustedContext(t *testing.T) {
	t.Parallel()

	adapter := NewExecutionAdapter(nil, nil, nil, nil, nil, log.NewLogger())
	_, _, err := adapter.Prepare(
		t.Context(),
		&coredata.AgentExecution{TrustedContext: json.RawMessage("{")},
		nil,
		nil,
	)
	assert.ErrorContains(t, err, "trusted context")
}

func TestExecutionAdapterUsesInputIdentityWithoutBindingLookup(t *testing.T) {
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
		binding: &identitybinding.Binding{
			IdentityID: gid.New(organizationID.TenantID(), coredata.IdentityEntityType),
		},
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
	assert.Equal(t, 0, bindings.lookups)
}

func TestExecutionAdapterPrepare_RefreshesAssistantStatus(t *testing.T) {
	t.Parallel()

	var got []map[string]any
	server := capturingAssistantStatusServer(t, &got)
	pgClient, organizationID, teamID := executionAdapterDatabase(t)
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

	adapter := newExecutionAdapter(
		t,
		pgClient,
		server.URL+"/api",
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
			SourceCoordinates: coordinates,
		},
		nil,
		nil,
	)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(
		t,
		map[string]any{
			"channel_id": "C123",
			"thread_ts":  "123.000",
			"status":     assistantWorkingStatus,
			"loading_messages": []any{
				"Looking that up…",
				"Checking your workspace…",
				"Drafting a reply…",
			},
		},
		got[0],
	)
}

func TestExecutionAdapterPrepare_UsesMessageTSWhenThreadTSEmpty(t *testing.T) {
	t.Parallel()

	var got []map[string]any
	server := capturingAssistantStatusServer(t, &got)
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
		server.URL+"/api",
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
			SourceCoordinates: coordinates,
		},
		nil,
		nil,
	)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "D123", got[0]["channel_id"])
	assert.Equal(t, "123.456", got[0]["thread_ts"])
}

func TestExecutionAdapterPrepare_SkipsAssistantStatusWithoutCoordinates(t *testing.T) {
	t.Parallel()

	var got []map[string]any
	server := capturingAssistantStatusServer(t, &got)
	pgClient, organizationID, _ := executionAdapterDatabase(t)

	adapter := newExecutionAdapter(
		t,
		pgClient,
		server.URL+"/api",
		&fakeExecutionBindings{},
	)

	_, _, err := adapter.Prepare(
		t.Context(),
		&coredata.AgentExecution{
			ID:             gid.New(organizationID.TenantID(), coredata.AgentExecutionEntityType),
			OrganizationID: organizationID,
			StartAgentName: "probot",
		},
		nil,
		nil,
	)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestExecutionAdapterPrepare_RefreshesAssistantStatusByOrganization(t *testing.T) {
	t.Parallel()

	var got []map[string]any
	server := capturingAssistantStatusServer(t, &got)
	pgClient, organizationID, _ := executionAdapterDatabase(t)
	coordinates, err := json.Marshal(
		ExecutionSourceCoordinates{
			ChannelID: "C123",
			MessageTS: "123.456",
		},
	)
	require.NoError(t, err)

	adapter := newExecutionAdapter(
		t,
		pgClient,
		server.URL+"/api",
		&fakeExecutionBindings{},
	)

	_, _, err = adapter.Prepare(
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
	require.Len(t, got, 1)
	assert.Equal(t, "C123", got[0]["channel_id"])
	assert.Equal(t, "123.456", got[0]["thread_ts"])
}

func TestExecutionAdapterPrepare_AssistantStatusErrorsDoNotFailPrepare(t *testing.T) {
	t.Parallel()

	t.Run(
		"client lookup error",
		func(t *testing.T) {
			t.Parallel()

			pgClient, _, organizationID := executionIngressDatabase(t)
			teamID := uniqueSlackTeamID(t)
			insertInstallationWithKey(
				t,
				pgClient,
				organizationID,
				teamID,
				"UBOT",
				cipher.EncryptionKey{9, 9, 9},
			)
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

			adapter := NewExecutionAdapter(
				newTestInstallationService(t, pgClient, ""),
				&fakeExecutionBindings{
					binding: &identitybinding.Binding{
						IdentityID: gid.New(organizationID.TenantID(), coredata.IdentityEntityType),
					},
				},
				newTestAgentProfiles(t),
				probot.NewCapabilityRegistry(),
				NewDeliveryService(pgClient),
				log.NewLogger(),
			)

			_, _, err = adapter.Prepare(
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
		},
	)

	t.Run(
		"set status api error",
		func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						assert.Equal(t, "/api/assistant.threads.setStatus", r.URL.Path)
						w.Header().Set("Content-Type", "application/json")
						_, err := w.Write([]byte(`{"ok":false,"error":"fatal_error"}`))
						require.NoError(t, err)
					},
				),
			)
			t.Cleanup(server.Close)

			pgClient, organizationID, teamID := executionAdapterDatabase(t)
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

			adapter := newExecutionAdapter(
				t,
				pgClient,
				server.URL+"/api",
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
					SourceCoordinates: coordinates,
				},
				nil,
				nil,
			)
			require.NoError(t, err)
		},
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
		newTestInstallationService(t, pgClient, apiBaseURL),
		bindings,
		newTestAgentProfiles(t),
		probot.NewCapabilityRegistry(),
		NewDeliveryService(pgClient),
		log.NewLogger(),
	)
}

func capturingAssistantStatusServer(t *testing.T, got *[]map[string]any) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/assistant.threads.setStatus", r.URL.Path)

				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				*got = append(*got, body)

				_, err := w.Write([]byte(`{"ok":true}`))
				require.NoError(t, err)
			},
		),
	)
	t.Cleanup(server.Close)

	return server
}
