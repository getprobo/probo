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
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"go.probo.inc/probo/pkg/agent"
	portal "go.probo.inc/probo/pkg/complianceportal"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	messaging "go.probo.inc/probo/pkg/probot"
)

type (
	manageAccessParams struct {
		Decision   string `json:"decision" jsonschema:"Decision to apply; must be approve or deny"`
		ResourceID string `json:"resource_id" jsonschema:"Opaque resource ID from get_compliance_access_request, or an empty string to apply to the whole request"`
	}

	emptyParams struct{}
)

func (c *Capability) Tools() []agent.Tool {
	return []agent.Tool{
		agent.FunctionTool(
			"get_compliance_access_request",
			"Load the compliance access request attached to the current conversation, including resource IDs and statuses.",
			func(ctx context.Context, _ emptyParams) (agent.ToolResult, error) {
				message, err := c.messageFromRunContext(ctx)
				if err != nil {
					return agent.ResultErrorf("cannot load compliance access request: %s", err), nil
				}

				content, err := json.Marshal(message.Attributes)
				if err != nil {
					return agent.ResultErrorf("cannot encode compliance access request: %s", err), nil
				}

				return agent.ToolResult{Content: string(content)}, nil
			},
		),
		agent.FunctionTool(
			"manage_compliance_access_request",
			"Approve or deny all or one resource in the compliance access request attached to the current conversation. Call get_compliance_access_request first when targeting one resource.",
			func(ctx context.Context, params manageAccessParams) (agent.ToolResult, error) {
				if params.Decision != "approve" && params.Decision != "deny" {
					return agent.ResultErrorf("decision must be approve or deny"), nil
				}

				message, err := c.messageFromRunContext(ctx)
				if err != nil {
					return agent.ResultErrorf("cannot load compliance access request: %s", err), nil
				}

				rc := agent.RunContextFrom[*messaging.RunContext](ctx)
				actionID := capabilityName + "." + params.Decision + "_all"
				value := message.ID.String()

				if params.ResourceID != "" {
					actionID = capabilityName + "." + params.Decision + "_item"
					value = params.ResourceID
				}

				deduplicationKey, ok := manageAccessOperationKey(ctx, rc, params.Decision, params.ResourceID)
				if !ok {
					return agent.ResultError("cannot manage compliance access request without a stable tool call ID"), nil
				}

				result, err := c.execute(
					ctx,
					messaging.Action{
						ID:               actionID,
						Value:            value,
						DeduplicationKey: deduplicationKey,
						ActorIdentityID:  rc.IdentityID,
						Message:          message,
					},
				)
				if err != nil {
					return agent.ResultErrorf("cannot manage compliance access request: %s", err), nil
				}

				return agent.ToolResult{Content: result.Message}, nil
			},
		),
		agent.FunctionTool(
			"refresh_compliance_access_request_card",
			"Rebuild the actionable compliance access request card in the current conversation from current Probo data.",
			func(ctx context.Context, _ emptyParams) (agent.ToolResult, error) {
				message, err := c.messageFromRunContext(ctx)
				if err != nil {
					return agent.ResultErrorf("cannot load compliance access request: %s", err), nil
				}

				requesterEmail, ok := portal.RequesterEmailFromMessage(message)
				if !ok {
					return agent.ResultErrorf("compliance access request has no requester"), nil
				}

				rc := agent.RunContextFrom[*messaging.RunContext](ctx)

				trustedAccessID, err := portal.AccessIDFromAttributes(rc.Attributes)
				if err != nil {
					return agent.ResultErrorf("cannot resolve trusted compliance access request: %s", err), nil
				}

				messageAccessID, err := c.notifications.ResolveCompliancePortalAccessID(
					ctx,
					coredata.NewScopeFromObjectID(rc.OrganizationID),
					message,
				)
				if err != nil {
					return agent.ResultErrorf("cannot resolve delivered compliance access request: %s", err), nil
				}

				if messageAccessID != trustedAccessID {
					return agent.ResultError("delivered message does not match trusted compliance access request"), nil
				}

				if err := c.notifications.UpdateAccessRequest(
					ctx,
					coredata.NewScopeFromObjectID(trustedAccessID),
					message.ID,
					requesterEmail,
				); err != nil {
					return agent.ResultErrorf("cannot refresh compliance access request: %s", err), nil
				}

				return agent.ToolResult{Content: "access request card refreshed"}, nil
			},
		),
	}
}

func (c *Capability) messageFromRunContext(ctx context.Context) (messaging.Message, error) {
	rc := agent.RunContextFrom[*messaging.RunContext](ctx)
	if rc == nil ||
		rc.OrganizationID == gid.Nil ||
		rc.MessageAnchor.ConversationID == "" ||
		rc.MessageAnchor.MessageID == "" {
		return messaging.Message{}, fmt.Errorf("current conversation is not an access request thread")
	}

	message, err := c.resolveInitialMessage(ctx, rc.OrganizationID, rc.MessageAnchor)
	if err != nil {
		return messaging.Message{}, fmt.Errorf("cannot load notification message: %w", err)
	}

	if message.Message.Type != portal.AccessMessageType {
		return messaging.Message{}, fmt.Errorf("%w", messaging.ErrCapabilityNotFound)
	}

	return message.Message, nil
}

func (c *Capability) resolveInitialMessage(
	ctx context.Context,
	organizationID gid.GID,
	anchor messaging.MessageAnchor,
) (*messaging.DeliveredMessage, error) {
	return c.notifications.GetInitialMessage(ctx, organizationID, anchor)
}

func manageAccessOperationKey(
	ctx context.Context,
	rc *messaging.RunContext,
	decision string,
	resourceID string,
) (string, bool) {
	toolCallID, ok := agent.ToolCallIDFrom(ctx)
	if !ok {
		return "", false
	}

	sum := sha256.Sum256(
		fmt.Appendf(
			nil,
			"%s\x00%s\x00%s\x00%s\x00%s\x00%s",
			rc.OrganizationID.String(),
			rc.MessageAnchor.ConversationID,
			rc.MessageAnchor.MessageID,
			decision,
			resourceID,
			toolCallID,
		),
	)

	return fmt.Sprintf("agent-tool:%x", sum[:]), true
}
