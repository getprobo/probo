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

	portal "go.probo.inc/probo/pkg/complianceportal"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mail"
	messaging "go.probo.inc/probo/pkg/probot"
)

const capabilityName = portal.AccessCapability

type (
	accessService interface {
		RenderMessage(ctx context.Context, message messaging.Message) (messaging.MessageIntent, error)
		BuildOutboundMessage(ctx context.Context, organizationID gid.GID, messageType string, attributes map[string]any) (messaging.OutboundMessage, error)
		GetInitialMessage(ctx context.Context, organizationID gid.GID, anchor messaging.MessageAnchor) (*messaging.DeliveredMessage, error)
		GetMessageResourceIDs(ctx context.Context, scope coredata.Scoper, messageID gid.GID) ([]gid.GID, []gid.GID, []gid.GID, error)
		ResolveCompliancePortalAccessID(ctx context.Context, scope coredata.Scoper, message messaging.Message) (gid.GID, error)
		ResolveCompliancePortalID(ctx context.Context, scope coredata.Scoper, organizationID gid.GID, metadata map[string]any) (gid.GID, error)
		UpdateAccessRequest(ctx context.Context, scope coredata.Scoper, messageID gid.GID, requesterEmail mail.Addr) error
	}

	accessVisitor interface {
		GrantPortalAccessByIDsIdempotently(ctx context.Context, scope coredata.Scoper, compliancePortalID gid.GID, requesterEmail mail.Addr, documentIDs, reportIDs, fileIDs []gid.GID, operationKey string) error
		RejectOrRevokePortalAccessByIDsIdempotently(ctx context.Context, scope coredata.Scoper, compliancePortalID gid.GID, requesterEmail mail.Addr, documentIDs, reportIDs, fileIDs []gid.GID, operationKey string) error
	}

	actionAuthorizer interface {
		Authorize(ctx context.Context, params iam.AuthorizeParams) (*coredata.Scope, error)
	}

	Capability struct {
		notifications accessService
		visitor       accessVisitor
		authorizer    actionAuthorizer
	}
)

var (
	_ messaging.Capability                = (*Capability)(nil)
	_ messaging.MessageCapability         = (*Capability)(nil)
	_ messaging.ActionCapability          = (*Capability)(nil)
	_ messaging.ActionAliasContributor    = (*Capability)(nil)
	_ messaging.ToolContributor           = (*Capability)(nil)
	_ messaging.OutboundMessageCapability = (*Capability)(nil)
)

func NewCapability(
	notificationService accessService,
	visitorService accessVisitor,
	authorizer actionAuthorizer,
) *Capability {
	return &Capability{
		notifications: notificationService,
		visitor:       visitorService,
		authorizer:    authorizer,
	}
}

func (c *Capability) Name() string {
	return capabilityName
}

func (c *Capability) RenderMessage(
	ctx context.Context,
	message messaging.Message,
) (messaging.MessageIntent, error) {
	return c.notifications.RenderMessage(ctx, message)
}

func (c *Capability) BuildOutboundMessage(
	ctx context.Context,
	organizationID gid.GID,
	messageType string,
	attributes map[string]any,
) (messaging.OutboundMessage, error) {
	return c.notifications.BuildOutboundMessage(ctx, organizationID, messageType, attributes)
}

func (c *Capability) MessageTypes() []string {
	return []string{
		portal.AccessMessageType,
	}
}

func (c *Capability) ActionPrefixes() []string {
	return []string{
		capabilityName + ".",
	}
}
