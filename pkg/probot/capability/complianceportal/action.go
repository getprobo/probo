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
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/bot"
	portal "go.probo.inc/probo/pkg/complianceportal"
	"go.probo.inc/probo/pkg/complianceportal/management"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	messaging "go.probo.inc/probo/pkg/probot"
)

type actionSelection struct {
	decision    string
	documentIDs []gid.GID
	reportIDs   []gid.GID
	fileIDs     []gid.GID
}

func (c *Capability) NormalizeActionAlias(action messaging.Action) (messaging.Action, error) {
	switch {
	case action.ID == "accept_all":
		action.ID = capabilityName + ".approve_all"
	case action.ID == "reject_all":
		action.ID = capabilityName + ".deny_all"
	case strings.HasPrefix(action.ID, "accept_"):
		action.ID = capabilityName + ".approve_item"
	case strings.HasPrefix(action.ID, "reject_"),
		strings.HasPrefix(action.ID, "revoke_"):
		action.ID = capabilityName + ".deny_item"
	case strings.HasPrefix(action.ID, "handle_"):
		decision, resourceID, ok := strings.Cut(action.SelectedValue, "/")
		if !ok || resourceID == "" {
			return messaging.Action{}, fmt.Errorf("invalid legacy selected option")
		}

		switch decision {
		case "accept":
			action.ID = capabilityName + ".approve_item"
		case "reject":
			action.ID = capabilityName + ".deny_item"
		default:
			return messaging.Action{}, fmt.Errorf("unsupported legacy decision %q", decision)
		}

		action.Value = resourceID
		action.SelectedValue = ""
	}

	return action, nil
}

func (c *Capability) HandleAction(
	ctx context.Context,
	action messaging.Action,
) (messaging.ActionResult, error) {
	return c.execute(ctx, action)
}

func (c *Capability) execute(
	ctx context.Context,
	action messaging.Action,
) (messaging.ActionResult, error) {
	requesterEmail, ok := portal.RequesterEmailFromMessage(action.Message)
	if !ok {
		return messaging.ActionResult{}, fmt.Errorf(
			"%w: compliance access request is missing requester information",
			messaging.ErrCapabilityInvalidInput,
		)
	}

	scope, err := c.authorize(ctx, action.ActorIdentityID, action.Message)
	if err != nil {
		return messaging.ActionResult{}, err
	}

	selection, err := c.selectResources(ctx, scope, action)
	if err != nil {
		return messaging.ActionResult{}, err
	}

	portalID, err := c.notifications.ResolveCompliancePortalID(
		ctx,
		scope,
		action.Message.OrganizationID,
		action.Message.Attributes,
	)
	if err != nil {
		return messaging.ActionResult{}, fmt.Errorf("cannot resolve compliance portal: %w", err)
	}

	switch selection.decision {
	case "approve":
		err = c.visitor.GrantPortalAccessByIDsIdempotently(
			ctx,
			scope,
			portalID,
			requesterEmail,
			selection.documentIDs,
			selection.reportIDs,
			selection.fileIDs,
			action.DeduplicationKey,
		)
	case "deny":
		err = c.visitor.RejectOrRevokePortalAccessByIDsIdempotently(
			ctx,
			scope,
			portalID,
			requesterEmail,
			selection.documentIDs,
			selection.reportIDs,
			selection.fileIDs,
			action.DeduplicationKey,
		)
	default:
		return messaging.ActionResult{}, fmt.Errorf(
			"%w: unknown decision %q",
			messaging.ErrCapabilityInvalidInput,
			selection.decision,
		)
	}

	if err != nil {
		return messaging.ActionResult{}, fmt.Errorf(
			"cannot %s compliance access request: %w",
			selection.decision,
			err,
		)
	}

	// Re-render the originating notification from current domain state so Slack
	// (and any other channel) shows the new statuses and controls in place.
	if err := c.notifications.UpdateAccessRequest(
		ctx,
		scope,
		action.Message.ID,
		requesterEmail,
	); err != nil {
		return messaging.ActionResult{}, fmt.Errorf(
			"cannot refresh compliance access request message: %w",
			err,
		)
	}

	message := "Access request denied"
	if selection.decision == "approve" {
		message = "Access request approved"
	}

	return messaging.ActionResult{Message: message}, nil
}

func (c *Capability) authorize(
	ctx context.Context,
	identityID gid.GID,
	message bot.Message,
) (*coredata.Scope, error) {
	lookupScope := coredata.NewScopeFromObjectID(message.OrganizationID)

	accessID, err := c.notifications.ResolveCompliancePortalAccessID(
		ctx,
		lookupScope,
		message,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve compliance portal access: %w", err)
	}

	scope, err := c.authorizer.Authorize(
		ctx,
		iam.AuthorizeParams{
			Principal: identityID,
			Resource:  accessID,
			Action:    management.ActionCompliancePortalAccessUpdate,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", messaging.ErrCapabilityForbidden, err)
	}

	return scope, nil
}

func (c *Capability) selectResources(
	ctx context.Context,
	scope coredata.Scoper,
	action messaging.Action,
) (actionSelection, error) {
	decision := ""
	resourceValue := action.Value

	switch {
	case action.ID == capabilityName+".review_item":
		selected, resourceID, ok := strings.Cut(action.SelectedValue, "/")
		if !ok || resourceID == "" {
			return actionSelection{}, fmt.Errorf(
				"%w: invalid selected review option",
				messaging.ErrCapabilityInvalidInput,
			)
		}

		switch selected {
		case "approve":
			decision = "approve"
		case "reject":
			decision = "deny"
		default:
			return actionSelection{}, fmt.Errorf(
				"%w: unsupported review decision %q",
				messaging.ErrCapabilityInvalidInput,
				selected,
			)
		}

		resourceValue = resourceID
	case strings.HasPrefix(action.ID, capabilityName+".approve_"):
		decision = "approve"
	case strings.HasPrefix(action.ID, capabilityName+".deny_"):
		decision = "deny"
	}

	if decision == "" {
		return actionSelection{}, fmt.Errorf(
			"%w: unsupported action %q",
			messaging.ErrCapabilityInvalidInput,
			action.ID,
		)
	}

	if strings.HasSuffix(action.ID, "_all") {
		documentIDs, reportIDs, fileIDs, err := c.notifications.GetMessageResourceIDs(
			ctx,
			scope,
			action.Message.ID,
		)
		if err != nil {
			return actionSelection{}, fmt.Errorf("cannot load requested resources: %w", err)
		}

		if len(documentIDs) == 0 && len(reportIDs) == 0 && len(fileIDs) == 0 {
			return actionSelection{}, fmt.Errorf(
				"%w: access request has no resources to %s",
				messaging.ErrCapabilityInvalidInput,
				decision,
			)
		}

		return actionSelection{
			decision:    decision,
			documentIDs: documentIDs,
			reportIDs:   reportIDs,
			fileIDs:     fileIDs,
		}, nil
	}

	resourceID, err := gid.ParseGID(resourceValue)
	if err != nil {
		return actionSelection{}, fmt.Errorf(
			"%w: invalid resource ID",
			messaging.ErrCapabilityInvalidInput,
		)
	}

	documentIDs, reportIDs, fileIDs, err := c.notifications.GetMessageResourceIDs(
		ctx,
		scope,
		action.Message.ID,
	)
	if err != nil {
		return actionSelection{}, fmt.Errorf("cannot load requested resources: %w", err)
	}

	if !resourceIDOnMessage(resourceID, documentIDs, reportIDs, fileIDs) {
		return actionSelection{}, fmt.Errorf(
			"%w: resource is not attached to this access request",
			messaging.ErrCapabilityInvalidInput,
		)
	}

	selection := actionSelection{decision: decision}

	switch resourceID.EntityType() {
	case coredata.DocumentEntityType:
		selection.documentIDs = []gid.GID{resourceID}
	case coredata.FileEntityType:
		selection.reportIDs = []gid.GID{resourceID}
	case coredata.CompliancePortalFileEntityType:
		selection.fileIDs = []gid.GID{resourceID}
	default:
		return actionSelection{}, fmt.Errorf(
			"%w: unsupported resource type %d",
			messaging.ErrCapabilityInvalidInput,
			resourceID.EntityType(),
		)
	}

	return selection, nil
}

func resourceIDOnMessage(
	resourceID gid.GID,
	documentIDs []gid.GID,
	reportIDs []gid.GID,
	fileIDs []gid.GID,
) bool {
	for _, ids := range [][]gid.GID{documentIDs, reportIDs, fileIDs} {
		for _, id := range ids {
			if id == resourceID {
				return true
			}
		}
	}

	return false
}
