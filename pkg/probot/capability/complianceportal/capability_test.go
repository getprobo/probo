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

package complianceportal

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/bot"
	portal "go.probo.inc/probo/pkg/complianceportal"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mail"
	messaging "go.probo.inc/probo/pkg/probot"
)

type (
	fakeAccessService struct {
		message              bot.Message
		renderer             *portal.Renderer
		documentIDs          []gid.GID
		reportIDs            []gid.GID
		fileIDs              []gid.GID
		resolvedAccessID     gid.GID
		lookupOrganizationID gid.GID
		lookupAnchor         messaging.MessageAnchor
		claimedKeys          map[string]struct{}
		eventIntent          bot.MessageIntent
		claimCount           int
		updateCount          int
		updateErr            error
	}

	fakeVisitor struct {
		grantCount  int
		rejectCount int
		operations  map[string]struct{}
		err         error
	}

	fakeAuthorizer struct {
		scope *coredata.Scope
		err   error
	}
)

func (f *fakeAccessService) RenderMessage(
	ctx context.Context,
	message bot.Message,
) (bot.MessageIntent, error) {
	return f.renderer.RenderMessage(ctx, message)
}

func (f *fakeAccessService) BuildOutboundMessage(
	context.Context,
	gid.GID,
	string,
	map[string]any,
) (messaging.OutboundMessage, error) {
	return messaging.OutboundMessage{Intent: f.eventIntent}, nil
}

func (f *fakeAccessService) GetInitialMessage(
	_ context.Context,
	organizationID gid.GID,
	anchor messaging.MessageAnchor,
) (*bot.DeliveredMessage, error) {
	f.lookupOrganizationID = organizationID
	f.lookupAnchor = anchor

	return &bot.DeliveredMessage{Message: f.message}, nil
}

func (f *fakeAccessService) GetMessageResourceIDs(
	context.Context,
	coredata.Scoper,
	gid.GID,
) ([]gid.GID, []gid.GID, []gid.GID, error) {
	return f.documentIDs, f.reportIDs, f.fileIDs, nil
}

func (f *fakeAccessService) ResolveCompliancePortalAccessID(
	context.Context,
	coredata.Scoper,
	bot.Message,
) (gid.GID, error) {
	if f.resolvedAccessID != gid.Nil {
		return f.resolvedAccessID, nil
	}

	return gid.New(f.message.ID.TenantID(), coredata.CompliancePortalAccessEntityType), nil
}

func (f *fakeAccessService) ResolveCompliancePortalID(
	context.Context,
	coredata.Scoper,
	gid.GID,
	map[string]any,
) (gid.GID, error) {
	return gid.New(f.message.ID.TenantID(), coredata.CompliancePortalEntityType), nil
}

func (f *fakeAccessService) ClaimInteractiveAction(
	_ context.Context,
	_ coredata.Scoper,
	_ gid.GID,
	interactionKey string,
) (string, bool, error) {
	f.claimCount++
	if f.claimedKeys == nil {
		f.claimedKeys = make(map[string]struct{})
	}

	if _, ok := f.claimedKeys[interactionKey]; ok {
		return "", false, nil
	}

	f.claimedKeys[interactionKey] = struct{}{}

	return "processing-token", true, nil
}

func (f *fakeAccessService) CompleteInteractiveAction(
	context.Context,
	coredata.Scoper,
	gid.GID,
	string,
	string,
) error {
	return nil
}

func (f *fakeAccessService) ReleaseInteractiveAction(
	_ context.Context,
	_ coredata.Scoper,
	_ gid.GID,
	interactionKey string,
	_ string,
) error {
	delete(f.claimedKeys, interactionKey)

	return nil
}

func (f *fakeAccessService) UpdateAccessRequest(
	context.Context,
	coredata.Scoper,
	gid.GID,
	mail.Addr,
) error {
	f.updateCount++
	return f.updateErr
}

func (f *fakeVisitor) GrantPortalAccessByIDsIdempotently(
	_ context.Context,
	_ coredata.Scoper,
	_ gid.GID,
	_ mail.Addr,
	_ []gid.GID,
	_ []gid.GID,
	_ []gid.GID,
	operationKey string,
) error {
	if f.err != nil {
		return f.err
	}

	if operationKey != "" {
		if f.operations == nil {
			f.operations = make(map[string]struct{})
		}

		if _, ok := f.operations[operationKey]; ok {
			return nil
		}

		f.operations[operationKey] = struct{}{}
	}

	f.grantCount++

	return nil
}

func (f *fakeVisitor) RejectOrRevokePortalAccessByIDsIdempotently(
	_ context.Context,
	_ coredata.Scoper,
	_ gid.GID,
	_ mail.Addr,
	_ []gid.GID,
	_ []gid.GID,
	_ []gid.GID,
	operationKey string,
) error {
	if f.err != nil {
		return f.err
	}

	if operationKey != "" {
		if f.operations == nil {
			f.operations = make(map[string]struct{})
		}

		if _, ok := f.operations[operationKey]; ok {
			return nil
		}

		f.operations[operationKey] = struct{}{}
	}

	f.rejectCount++

	return nil
}

func (f *fakeAuthorizer) Authorize(
	context.Context,
	iam.AuthorizeParams,
) (*coredata.Scope, error) {
	return f.scope, f.err
}

func TestCapability_BuildsOutboundMessage(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	notifications := &fakeAccessService{
		eventIntent: bot.MessageIntent{FallbackText: "New access request"},
	}
	capability := NewCapability(notifications, &fakeVisitor{}, &fakeAuthorizer{})

	result, err := capability.BuildOutboundMessage(
		t.Context(),
		gid.New(tenantID, coredata.OrganizationEntityType),
		portal.AccessMessageType,
		map[string]any{
			portal.AccessIDAttribute: gid.New(
				tenantID,
				coredata.CompliancePortalAccessEntityType,
			).String(),
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "New access request", result.Intent.FallbackText)
}

func TestCapability_ButtonAndAgentUseSameCommand(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	requesterEmail := mail.Addr("requester@example.com")
	message := bot.Message{
		ID:             gid.New(tenantID, coredata.CompliancePortalAccessEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Type:           portal.AccessMessageType,
		Attributes:     map[string]any{"requester_email": requesterEmail.String()},
	}
	notificationService := &fakeAccessService{
		message: message,
		documentIDs: []gid.GID{
			gid.New(tenantID, coredata.DocumentEntityType),
		},
	}
	visitor := &fakeVisitor{}
	capability := NewCapability(
		notificationService,
		visitor,
		&fakeAuthorizer{scope: coredata.NewScope(tenantID)},
	)
	identityID := gid.New(gid.NilTenant, coredata.IdentityEntityType)

	result, err := capability.HandleAction(
		context.Background(),
		messaging.Action{
			ID:               "compliance_access.approve_all",
			DeduplicationKey: "event-123",
			Value:            message.ID.String(),
			ActorIdentityID:  identityID,
			Message:          message,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "Access request approved", result.Message)

	var manageTool agent.Tool

	for _, tool := range capability.Tools() {
		if tool.Name() == "manage_compliance_access_request" {
			manageTool = tool
			break
		}
	}

	require.NotNil(t, manageTool)

	ctx := agent.WithToolCallID(
		agent.WithRunContext(
			context.Background(),
			&messaging.RunContext{
				OrganizationID: message.OrganizationID,
				MessageAnchor: messaging.MessageAnchor{
					ConversationID: "C123",
					MessageID:      "123.456",
				},
				IdentityID: identityID,
			},
		),
		"call-approve-all",
	)
	toolResult, err := manageTool.Execute(
		ctx,
		`{"decision":"approve","resource_id":""}`,
	)
	require.NoError(t, err)
	assert.False(t, toolResult.IsError)
	assert.Equal(t, "Access request approved", toolResult.Content)
	assert.Equal(t, 2, visitor.grantCount)
	assert.Zero(t, visitor.rejectCount)
	assert.Equal(t, 2, notificationService.updateCount)
	assert.Equal(t, message.OrganizationID, notificationService.lookupOrganizationID)
	assert.Equal(
		t,
		messaging.MessageAnchor{ConversationID: "C123", MessageID: "123.456"},
		notificationService.lookupAnchor,
	)

	duplicateResult, err := manageTool.Execute(
		ctx,
		`{"decision":"approve","resource_id":""}`,
	)
	require.NoError(t, err)
	assert.Equal(t, "Access request approved", duplicateResult.Content)
	assert.Equal(t, 2, visitor.grantCount)
	assert.Equal(t, 3, notificationService.updateCount)
}

func TestCapability_FailedActionCanRetry(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	message := bot.Message{
		ID:             gid.New(tenantID, coredata.CompliancePortalAccessEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Type:           portal.AccessMessageType,
		Attributes:     map[string]any{"requester_email": "requester@example.com"},
	}
	notifications := &fakeAccessService{
		message: message,
		documentIDs: []gid.GID{
			gid.New(tenantID, coredata.DocumentEntityType),
		},
	}
	visitor := &fakeVisitor{err: errors.New("temporary action failure")}
	capability := NewCapability(
		notifications,
		visitor,
		&fakeAuthorizer{scope: coredata.NewScope(tenantID)},
	)
	action := messaging.Action{
		ID:               "compliance_access.approve_all",
		DeduplicationKey: "retryable-action",
		Value:            message.ID.String(),
		ActorIdentityID:  gid.New(gid.NilTenant, coredata.IdentityEntityType),
		Message:          message,
	}

	_, err := capability.HandleAction(context.Background(), action)
	require.ErrorContains(t, err, "temporary action failure")

	visitor.err = nil
	result, err := capability.HandleAction(context.Background(), action)
	require.NoError(t, err)
	assert.Equal(t, "Access request approved", result.Message)
	assert.Equal(t, 1, visitor.grantCount)
	assert.Equal(t, 1, notifications.updateCount)
}

func TestCapability_RejectsEmptyApproveAll(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	message := bot.Message{
		ID:             gid.New(tenantID, coredata.CompliancePortalAccessEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Type:           portal.AccessMessageType,
		Attributes:     map[string]any{"requester_email": "requester@example.com"},
	}
	visitor := &fakeVisitor{}
	capability := NewCapability(
		&fakeAccessService{message: message},
		visitor,
		&fakeAuthorizer{scope: coredata.NewScope(tenantID)},
	)

	_, err := capability.HandleAction(
		context.Background(),
		messaging.Action{
			ID:               "compliance_access.approve_all",
			DeduplicationKey: "empty-all",
			Value:            message.ID.String(),
			ActorIdentityID:  gid.New(gid.NilTenant, coredata.IdentityEntityType),
			Message:          message,
		},
	)
	require.ErrorIs(t, err, messaging.ErrCapabilityInvalidInput)
	assert.Zero(t, visitor.grantCount)
}

func TestCapability_ManageToolRequiresToolCallID(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	message := bot.Message{
		ID:             gid.New(tenantID, coredata.CompliancePortalAccessEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Type:           portal.AccessMessageType,
		Attributes:     map[string]any{"requester_email": "requester@example.com"},
	}
	visitor := &fakeVisitor{}
	capability := NewCapability(
		&fakeAccessService{
			message: message,
			documentIDs: []gid.GID{
				gid.New(tenantID, coredata.DocumentEntityType),
			},
		},
		visitor,
		&fakeAuthorizer{scope: coredata.NewScope(tenantID)},
	)

	var manageTool agent.Tool

	for _, tool := range capability.Tools() {
		if tool.Name() == "manage_compliance_access_request" {
			manageTool = tool
			break
		}
	}

	require.NotNil(t, manageTool)

	ctx := agent.WithRunContext(
		context.Background(),
		&messaging.RunContext{
			OrganizationID: message.OrganizationID,
			MessageAnchor: messaging.MessageAnchor{
				ConversationID: "C123",
				MessageID:      "123.456",
			},
			IdentityID: gid.New(gid.NilTenant, coredata.IdentityEntityType),
		},
	)
	result, err := manageTool.Execute(ctx, `{"decision":"approve","resource_id":""}`)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "stable tool call ID")
	assert.Zero(t, visitor.grantCount)
}

func TestCapability_HandlesReviewMenuSelection(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)
	message := bot.Message{
		ID:             gid.New(tenantID, coredata.CompliancePortalAccessEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Type:           portal.AccessMessageType,
		Attributes:     map[string]any{"requester_email": "requester@example.com"},
	}

	for _, tt := range []struct {
		name          string
		selectedValue string
		wantMessage   string
		wantGrants    int
		wantRejects   int
	}{
		{
			name:          "grant",
			selectedValue: "approve/" + documentID.String(),
			wantMessage:   "Access request approved",
			wantGrants:    1,
		},
		{
			name:          "reject",
			selectedValue: "reject/" + documentID.String(),
			wantMessage:   "Access request denied",
			wantRejects:   1,
		},
	} {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				visitor := &fakeVisitor{}
				notifications := &fakeAccessService{
					message:     message,
					documentIDs: []gid.GID{documentID},
				}
				capability := NewCapability(
					notifications,
					visitor,
					&fakeAuthorizer{scope: coredata.NewScope(tenantID)},
				)

				result, err := capability.HandleAction(
					context.Background(),
					messaging.Action{
						ID:               "compliance_access.review_item",
						SelectedValue:    tt.selectedValue,
						DeduplicationKey: "digest-review",
						ActorIdentityID:  gid.New(gid.NilTenant, coredata.IdentityEntityType),
						Message:          message,
					},
				)
				require.NoError(t, err)
				assert.Equal(t, tt.wantMessage, result.Message)
				assert.Equal(t, tt.wantGrants, visitor.grantCount)
				assert.Equal(t, tt.wantRejects, visitor.rejectCount)
				assert.Equal(t, 1, notifications.updateCount)
			},
		)
	}
}

func TestCapability_SkipsRefreshWhenActionFails(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	message := bot.Message{
		ID:             gid.New(tenantID, coredata.CompliancePortalAccessEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Type:           portal.AccessMessageType,
		Attributes:     map[string]any{"requester_email": "requester@example.com"},
	}
	notifications := &fakeAccessService{
		message: message,
		documentIDs: []gid.GID{
			gid.New(tenantID, coredata.DocumentEntityType),
		},
	}
	capability := NewCapability(
		notifications,
		&fakeVisitor{err: errors.New("temporary grant failure")},
		&fakeAuthorizer{scope: coredata.NewScope(tenantID)},
	)

	_, err := capability.HandleAction(
		context.Background(),
		messaging.Action{
			ID:               "compliance_access.approve_all",
			DeduplicationKey: "digest-fail",
			ActorIdentityID:  gid.New(gid.NilTenant, coredata.IdentityEntityType),
			Message:          message,
		},
	)
	require.ErrorContains(t, err, "temporary grant failure")
	assert.Zero(t, notifications.updateCount)
}

func TestCapability_SurfacesRefreshFailureAfterAction(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	message := bot.Message{
		ID:             gid.New(tenantID, coredata.CompliancePortalAccessEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Type:           portal.AccessMessageType,
		Attributes:     map[string]any{"requester_email": "requester@example.com"},
	}
	notifications := &fakeAccessService{
		message: message,
		documentIDs: []gid.GID{
			gid.New(tenantID, coredata.DocumentEntityType),
		},
		updateErr: errors.New("cannot queue revision"),
	}
	capability := NewCapability(
		notifications,
		&fakeVisitor{},
		&fakeAuthorizer{scope: coredata.NewScope(tenantID)},
	)

	_, err := capability.HandleAction(
		context.Background(),
		messaging.Action{
			ID:               "compliance_access.approve_all",
			DeduplicationKey: "digest-refresh-fail",
			ActorIdentityID:  gid.New(gid.NilTenant, coredata.IdentityEntityType),
			Message:          message,
		},
	)
	require.ErrorContains(t, err, "cannot refresh compliance access request message")
	assert.Equal(t, 1, notifications.updateCount)
}

func TestCapability_RejectsMalformedReviewMenuSelection(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	message := bot.Message{
		ID:             gid.New(tenantID, coredata.CompliancePortalAccessEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Type:           portal.AccessMessageType,
		Attributes:     map[string]any{"requester_email": "requester@example.com"},
	}
	visitor := &fakeVisitor{}
	capability := NewCapability(
		&fakeAccessService{message: message},
		visitor,
		&fakeAuthorizer{scope: coredata.NewScope(tenantID)},
	)

	_, err := capability.HandleAction(
		context.Background(),
		messaging.Action{
			ID:               "compliance_access.review_item",
			SelectedValue:    "escalate/" + gid.New(tenantID, coredata.DocumentEntityType).String(),
			DeduplicationKey: "digest-review",
			ActorIdentityID:  gid.New(gid.NilTenant, coredata.IdentityEntityType),
			Message:          message,
		},
	)
	require.ErrorIs(t, err, messaging.ErrCapabilityInvalidInput)
	assert.Zero(t, visitor.grantCount)
	assert.Zero(t, visitor.rejectCount)
}

func TestCapability_RejectsResourceIDNotOnMessage(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	attachedID := gid.New(tenantID, coredata.DocumentEntityType)
	foreignID := gid.New(tenantID, coredata.DocumentEntityType)
	message := bot.Message{
		ID:             gid.New(tenantID, coredata.CompliancePortalAccessEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Type:           portal.AccessMessageType,
		Attributes:     map[string]any{"requester_email": "requester@example.com"},
	}
	visitor := &fakeVisitor{}
	capability := NewCapability(
		&fakeAccessService{
			message:     message,
			documentIDs: []gid.GID{attachedID},
		},
		visitor,
		&fakeAuthorizer{scope: coredata.NewScope(tenantID)},
	)

	_, err := capability.HandleAction(
		context.Background(),
		messaging.Action{
			ID:               "compliance_access.approve_item",
			Value:            foreignID.String(),
			DeduplicationKey: "foreign-resource",
			ActorIdentityID:  gid.New(gid.NilTenant, coredata.IdentityEntityType),
			Message:          message,
		},
	)
	require.ErrorIs(t, err, messaging.ErrCapabilityInvalidInput)
	assert.Zero(t, visitor.grantCount)
}

func TestCapability_ManageToolRejectsResourceIDNotOnMessage(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	attachedID := gid.New(tenantID, coredata.DocumentEntityType)
	foreignID := gid.New(tenantID, coredata.DocumentEntityType)
	message := bot.Message{
		ID:             gid.New(tenantID, coredata.CompliancePortalAccessEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Type:           portal.AccessMessageType,
		Attributes:     map[string]any{"requester_email": "requester@example.com"},
	}
	visitor := &fakeVisitor{}
	capability := NewCapability(
		&fakeAccessService{
			message:     message,
			documentIDs: []gid.GID{attachedID},
		},
		visitor,
		&fakeAuthorizer{scope: coredata.NewScope(tenantID)},
	)

	var manageTool agent.Tool

	for _, tool := range capability.Tools() {
		if tool.Name() == "manage_compliance_access_request" {
			manageTool = tool
			break
		}
	}

	require.NotNil(t, manageTool)

	ctx := agent.WithToolCallID(
		agent.WithRunContext(
			context.Background(),
			&messaging.RunContext{
				OrganizationID: message.OrganizationID,
				MessageAnchor: messaging.MessageAnchor{
					ConversationID: "C123",
					MessageID:      "123.456",
				},
				IdentityID: gid.New(gid.NilTenant, coredata.IdentityEntityType),
			},
		),
		"call-foreign-resource",
	)
	result, err := manageTool.Execute(
		ctx,
		`{"decision":"approve","resource_id":"`+foreignID.String()+`"}`,
	)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not attached")
	assert.Zero(t, visitor.grantCount)
}

func TestCapability_RenderMessageProducesChannelNeutralIntent(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	messageID := gid.New(tenantID, coredata.CompliancePortalAccessEntityType)
	documentID := gid.New(tenantID, coredata.DocumentEntityType)
	compliancePortalID := gid.New(tenantID, coredata.CompliancePortalEntityType)
	capability := NewCapability(
		&fakeAccessService{renderer: portal.NewRenderer("https://app.example.com")},
		nil,
		nil,
	)

	intent, err := capability.RenderMessage(
		context.Background(),
		bot.Message{
			ID:             messageID,
			OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
			Type:           portal.AccessMessageType,
			Attributes: map[string]any{
				"requester_name":       "Jane Requester",
				"requester_email":      "jane@example.com",
				"compliance_portal_id": compliancePortalID.String(),
				"documents": []any{
					map[string]any{
						"ID":     documentID.String(),
						"Title":  "Security policy",
						"Status": "REQUESTED",
					},
				},
				"reports": []any{},
				"files":   []any{},
			},
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "New compliance portal access request", intent.FallbackText)
	assert.Equal(t, "Requested by Jane Requester <jane@example.com>", intent.Context)

	require.Len(t, intent.Actions, 3)
	assert.Equal(t, "compliance_access.approve_all", intent.Actions[0].ID)
	assert.Empty(t, intent.Actions[0].Value)
	assert.Equal(t, "compliance_access.deny_all", intent.Actions[1].ID)
	assert.Contains(t, intent.Actions[2].URL, "/compliance-portals/"+compliancePortalID.String()+"/permissions")

	require.Len(t, intent.Groups, 1)
	assert.Equal(t, "Documents (1)", intent.Groups[0].Title)

	require.Len(t, intent.Groups[0].Items, 1)
	item := intent.Groups[0].Items[0]
	assert.Equal(t, "Security policy", item.Label)
	assert.Contains(t, item.URL, "/documents/"+documentID.String())

	require.NotNil(t, item.Action)
	assert.Equal(t, "compliance_access.review_item", item.Action.ID)
	assert.Equal(t, "approve/"+documentID.String(), item.Action.Options[0].Value)
	assert.Equal(t, "reject/"+documentID.String(), item.Action.Options[1].Value)
}

func TestCapability_ToolsUseTrustedRunContext(t *testing.T) {
	t.Parallel()

	capability := NewCapability(
		nil,
		nil,
		nil,
	)
	for _, tool := range capability.Tools() {
		var schema struct {
			Properties map[string]json.RawMessage `json:"properties"`
		}

		err := json.Unmarshal(tool.Definition().Parameters, &schema)
		require.NoError(t, err)

		for _, field := range []string{
			"transport",
			"conversation_id",
			"message_id",
			"identity_id",
		} {
			_, exists := schema.Properties[field]
			assert.False(t, exists, "%s must come from messaging.RunContext", field)
		}
	}
}

func TestCapability_RefreshesFromRunContextWithoutActor(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	accessID := gid.New(tenantID, coredata.CompliancePortalAccessEntityType)
	notifications := &fakeAccessService{
		message: bot.Message{
			ID:             gid.New(tenantID, coredata.CompliancePortalAccessEntityType),
			OrganizationID: organizationID,
			Type:           portal.AccessMessageType,
			Attributes: map[string]any{
				portal.RequesterEmailAttribute: "requester@example.com",
			},
		},
		resolvedAccessID: accessID,
	}
	capability := NewCapability(
		notifications,
		nil,
		&fakeAuthorizer{err: errors.New("domain refresh must not require an actor")},
	)

	var refreshTool agent.Tool

	for _, tool := range capability.Tools() {
		if tool.Name() == "refresh_compliance_access_request_card" {
			refreshTool = tool
			break
		}
	}

	require.NotNil(t, refreshTool)

	ctx := agent.WithRunContext(
		t.Context(),
		&messaging.RunContext{
			OrganizationID: organizationID,
			MessageAnchor: messaging.MessageAnchor{
				ConversationID: "C123",
				MessageID:      "123.456",
			},
			Attributes: map[string]any{
				portal.AccessIDAttribute: accessID.String(),
			},
		},
	)
	result, err := refreshTool.Execute(ctx, `{}`)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "access request card refreshed", result.Content)
	assert.Equal(t, 1, notifications.updateCount)
}

func TestCapability_RejectsUnauthorizedAgentAction(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	requesterEmail := mail.Addr("requester@example.com")
	message := bot.Message{
		ID:             gid.New(tenantID, coredata.CompliancePortalAccessEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Type:           portal.AccessMessageType,
		Attributes:     map[string]any{"requester_email": requesterEmail.String()},
	}
	capability := NewCapability(
		&fakeAccessService{message: message},
		&fakeVisitor{},
		&fakeAuthorizer{err: errors.New("forbidden")},
	)

	_, err := capability.HandleAction(
		context.Background(),
		messaging.Action{
			ID:              "compliance_access.approve_all",
			Value:           message.ID.String(),
			ActorIdentityID: gid.New(gid.NilTenant, coredata.IdentityEntityType),
			Message:         message,
		},
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, messaging.ErrCapabilityForbidden)
}
