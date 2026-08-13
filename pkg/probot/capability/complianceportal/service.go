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
	"errors"
	"fmt"
	"net/url"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/bot"
	portal "go.probo.inc/probo/pkg/complianceportal"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/probot"
)

const (
	deliveryTargetNamespace = "compliance_portal"
	welcomeMessageType      = "COMPLIANCE_PORTAL_WELCOME"
)

var (
	errCompliancePortalMetadataAmbiguous = errors.New("compliance portal id is missing from message metadata and organization has multiple portals")
)

type (
	messageDocument struct {
		ID     string
		Title  string
		Status string
	}

	messageReport struct {
		ID      string
		Title   string
		AuditID string
		Status  string
	}

	messageFile struct {
		ID       string
		Name     string
		Category string
		Status   string
	}

	messageMetadata struct {
		CompliancePortalID       gid.GID
		CompliancePortalAccessID gid.GID
		Documents                []messageDocument
		Reports                  []messageReport
		Files                    []messageFile
	}

	accessRequestState struct {
		Identity  *coredata.Identity
		Portal    *coredata.CompliancePortal
		Access    *coredata.CompliancePortalAccess
		Documents []messageDocument
		Reports   []messageReport
		Files     []messageFile
	}

	messageDelivery interface {
		GetInitialMessage(
			ctx context.Context,
			organizationID gid.GID,
			anchor probot.MessageAnchor,
		) (*probot.DeliveredMessage, error)
		GetMessage(
			ctx context.Context,
			scope coredata.Scoper,
			messageID gid.GID,
		) (*probot.DeliveredMessage, error)
		UpdateMessage(
			ctx context.Context,
			messageID gid.GID,
			message probot.Message,
			intent probot.MessageIntent,
		) error
		DeliverVerification(
			ctx context.Context,
			organizationID gid.GID,
			target probot.DeliveryTarget,
			message probot.Message,
			intent probot.MessageIntent,
		) error
	}

	Service struct {
		pg       *pg.Client
		renderer bot.MessageRenderer
		delivery messageDelivery
		baseURL  string
	}
)

func NewService(
	pgClient *pg.Client,
	renderer bot.MessageRenderer,
	delivery messageDelivery,
	baseURL string,
) *Service {
	return &Service{
		pg:       pgClient,
		renderer: renderer,
		delivery: delivery,
		baseURL:  baseURL,
	}
}

func (s *Service) RenderMessage(
	ctx context.Context,
	message probot.Message,
) (probot.MessageIntent, error) {
	return s.renderer.RenderMessage(ctx, message)
}

func (s *Service) BuildOutboundMessage(
	ctx context.Context,
	organizationID gid.GID,
	messageType string,
	attributes map[string]any,
) (probot.OutboundMessage, error) {
	if organizationID == gid.Nil || messageType != portal.AccessMessageType {
		return probot.OutboundMessage{}, fmt.Errorf(
			"%w: invalid trusted compliance access request context",
			probot.ErrCapabilityInvalidInput,
		)
	}

	accessID, err := portal.AccessIDFromAttributes(attributes)
	if err != nil {
		return probot.OutboundMessage{}, fmt.Errorf(
			"%w: %v",
			probot.ErrCapabilityInvalidInput,
			err,
		)
	}

	state, err := s.loadAccessRequestState(ctx, organizationID, accessID)
	if err != nil {
		return probot.OutboundMessage{}, fmt.Errorf("cannot load access request state: %w", err)
	}

	message := s.accessRequestMessage(state)
	intent, err := s.renderer.RenderMessage(ctx, message)
	if err != nil {
		return probot.OutboundMessage{}, fmt.Errorf("cannot render access request intent: %w", err)
	}

	return probot.OutboundMessage{
		Message: message,
		Intent:  intent,
		DeliveryTarget: probot.DeliveryTarget{
			Namespace: deliveryTargetNamespace,
			Key:       state.Portal.ID.String(),
		},
	}, nil
}

func (s *Service) GetInitialMessage(
	ctx context.Context,
	organizationID gid.GID,
	anchor probot.MessageAnchor,
) (*probot.DeliveredMessage, error) {
	return s.delivery.GetInitialMessage(
		ctx,
		organizationID,
		anchor,
	)
}

func (s *Service) GetMessageResourceIDs(
	ctx context.Context,
	scope coredata.Scoper,
	messageID gid.GID,
) ([]gid.GID, []gid.GID, []gid.GID, error) {
	message, err := s.delivery.GetMessage(ctx, scope, messageID)
	if err != nil {
		return nil, nil, nil, err
	}

	return extractIDsFromMetadata(message.Message.Attributes, "documents"),
		extractIDsFromMetadata(message.Message.Attributes, "reports"),
		extractIDsFromMetadata(message.Message.Attributes, "files"), nil
}

func (s *Service) ResolveCompliancePortalAccessID(
	ctx context.Context,
	scope coredata.Scoper,
	message probot.Message,
) (gid.GID, error) {
	if accessID, ok := gidFromMetadata(message.Attributes, "compliance_portal_access_id"); ok {
		return accessID, nil
	}

	requesterEmail, ok := requesterEmailFromMessage(message)
	if !ok {
		return gid.Nil, fmt.Errorf("message has no requester email")
	}

	portalID, err := s.ResolveCompliancePortalID(
		ctx,
		scope,
		message.OrganizationID,
		message.Attributes,
	)
	if err != nil {
		return gid.Nil, err
	}

	var accessID gid.GID
	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var identity coredata.Identity
			if err := identity.LoadByEmail(ctx, conn, requesterEmail); err != nil {
				return fmt.Errorf("cannot load requester identity: %w", err)
			}

			var access coredata.CompliancePortalAccess
			if err := access.LoadByCompliancePortalIDAndIdentityID(
				ctx,
				conn,
				scope,
				portalID,
				identity.ID,
			); err != nil {
				return fmt.Errorf("cannot load compliance portal access: %w", err)
			}

			accessID = access.ID
			return nil
		},
	)
	if err != nil {
		return gid.Nil, err
	}

	return accessID, nil
}

func (s *Service) ResolveCompliancePortalID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	metadata map[string]any,
) (gid.GID, error) {
	if portalID, ok := gidFromMetadata(metadata, "compliance_portal_id"); ok {
		var compliancePortal coredata.CompliancePortal
		err := s.pg.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return compliancePortal.LoadByID(ctx, conn, scope, portalID)
			},
		)
		if err != nil {
			return gid.Nil, fmt.Errorf("cannot load compliance portal from metadata: %w", err)
		}

		return portalID, nil
	}

	var (
		portalCount int
		portals     coredata.CompliancePortals
	)
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error
			portalCount, err = portals.CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count compliance portals: %w", err)
			}
			if portalCount != 1 {
				return nil
			}

			cursor := page.NewCursor(
				1,
				nil,
				page.Head,
				page.OrderBy[coredata.CompliancePortalOrderField]{
					Field:     coredata.CompliancePortalOrderFieldCreatedAt,
					Direction: page.OrderDirectionAsc,
				},
			)
			return portals.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor)
		},
	)
	if err != nil {
		return gid.Nil, err
	}
	if portalCount != 1 {
		return gid.Nil, errCompliancePortalMetadataAmbiguous
	}
	if len(portals) != 1 {
		return gid.Nil, fmt.Errorf("cannot resolve compliance portal: expected one portal, found %d", len(portals))
	}

	return portals[0].ID, nil
}

func (s *Service) UpdateAccessRequest(
	ctx context.Context,
	scope coredata.Scoper,
	messageID gid.GID,
	requesterEmail mail.Addr,
) error {
	delivered, err := s.delivery.GetMessage(ctx, scope, messageID)
	if err != nil {
		return fmt.Errorf("cannot load delivered access request: %w", err)
	}

	portalID, err := s.ResolveCompliancePortalID(
		ctx,
		scope,
		delivered.Message.OrganizationID,
		delivered.Message.Attributes,
	)
	if err != nil {
		return fmt.Errorf("cannot resolve compliance portal: %w", err)
	}

	var access coredata.CompliancePortalAccess
	var identity coredata.Identity
	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := identity.LoadByEmail(ctx, conn, requesterEmail); err != nil {
				return fmt.Errorf("cannot load requester identity: %w", err)
			}
			if err := access.LoadByCompliancePortalIDAndIdentityID(
				ctx,
				conn,
				scope,
				portalID,
				identity.ID,
			); err != nil {
				return fmt.Errorf("cannot load compliance portal access: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot load access request identity: %w", err)
	}

	state, err := s.loadAccessRequestState(
		ctx,
		delivered.Message.OrganizationID,
		access.ID,
	)
	if err != nil {
		return fmt.Errorf("cannot load access request for update: %w", err)
	}

	message := s.accessRequestMessage(state)
	message.ID = delivered.Message.ID
	intent, err := s.renderer.RenderMessage(ctx, message)
	if err != nil {
		return fmt.Errorf("cannot render updated access request: %w", err)
	}

	return s.delivery.UpdateMessage(ctx, delivered.Message.ID, message, intent)
}

func (s *Service) QueueWelcome(
	ctx context.Context,
	organizationID gid.GID,
	compliancePortalID gid.GID,
) error {
	scope := coredata.NewScopeFromObjectID(compliancePortalID)
	var compliancePortal coredata.CompliancePortal
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return compliancePortal.LoadByID(ctx, conn, scope, compliancePortalID)
		},
	)
	if err != nil {
		return fmt.Errorf("cannot load compliance portal for welcome message: %w", err)
	}

	accessURL, err := url.JoinPath(
		s.baseURL,
		"organizations",
		url.PathEscape(organizationID.String()),
		"compliance-portals",
		url.PathEscape(compliancePortalID.String()),
		"access",
	)
	if err != nil {
		return fmt.Errorf("cannot build compliance portal access URL: %w", err)
	}

	text := fmt.Sprintf(
		"Access requests for %s will be delivered here. Manage requests at %s.",
		compliancePortal.EntityName,
		accessURL,
	)
	message := probot.Message{
		OrganizationID: organizationID,
		Type:           welcomeMessageType,
		Attributes: map[string]any{
			"compliance_portal_id": compliancePortalID.String(),
		},
	}

	return s.delivery.DeliverVerification(
		ctx,
		organizationID,
		probot.DeliveryTarget{
			Namespace: deliveryTargetNamespace,
			Key:       compliancePortalID.String(),
		},
		message,
		probot.MessageIntent{FallbackText: text},
	)
}

func (s *Service) loadAccessRequestState(
	ctx context.Context,
	organizationID gid.GID,
	accessID gid.GID,
) (*accessRequestState, error) {
	scope := coredata.NewScopeFromObjectID(organizationID)
	state := &accessRequestState{
		Identity: &coredata.Identity{},
		Portal:   &coredata.CompliancePortal{},
		Access:   &coredata.CompliancePortalAccess{},
	}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := state.Access.LoadByID(ctx, conn, scope, accessID); err != nil {
				return fmt.Errorf("cannot load compliance portal access: %w", err)
			}
			if err := state.Identity.LoadByID(ctx, conn, state.Access.IdentityID); err != nil {
				return fmt.Errorf("cannot load requester identity: %w", err)
			}
			if err := state.Portal.LoadByID(
				ctx,
				conn,
				scope,
				state.Access.CompliancePortalID,
			); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}
			if state.Portal.OrganizationID != organizationID {
				return fmt.Errorf("compliance portal access does not belong to trusted organization")
			}

			var err error
			state.Documents, state.Reports, state.Files, err =
				loadResources(ctx, conn, scope, state.Portal.ID, state.Access.ID)
			if err != nil {
				return fmt.Errorf("cannot load access request resources: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load access request state: %w", err)
	}

	return state, nil
}

func (s *Service) accessRequestMessage(state *accessRequestState) probot.Message {
	attributes := messageMetadata{
		CompliancePortalID:       state.Portal.ID,
		CompliancePortalAccessID: state.Access.ID,
		Documents:                state.Documents,
		Reports:                  state.Reports,
		Files:                    state.Files,
	}.toMap()
	attributes["requester_name"] = state.Identity.FullName
	attributes["requester_email"] = state.Identity.EmailAddress.String()

	return probot.Message{
		OrganizationID: state.Portal.OrganizationID,
		Type:           portal.AccessMessageType,
		Attributes:     attributes,
	}
}

func (m messageMetadata) toMap() map[string]any {
	return map[string]any{
		"compliance_portal_id":        m.CompliancePortalID.String(),
		"compliance_portal_access_id": m.CompliancePortalAccessID.String(),
		"documents":                   m.Documents,
		"reports":                     m.Reports,
		"files":                       m.Files,
	}
}

func requesterEmailFromMessage(message probot.Message) (mail.Addr, bool) {
	raw, ok := message.Attributes["requester_email"].(string)
	return mail.Addr(raw), ok && raw != ""
}

func gidFromMetadata(metadata map[string]any, key string) (gid.GID, bool) {
	value, ok := metadata[key]
	if !ok {
		return gid.Nil, false
	}
	raw, ok := value.(string)
	if !ok {
		return gid.Nil, false
	}
	id, err := gid.ParseGID(raw)
	return id, err == nil
}

func extractIDsFromMetadata(metadata map[string]any, fieldName string) []gid.GID {
	ids := []gid.GID{}
	items, ok := metadata[fieldName].([]any)
	if !ok {
		return ids
	}

	for _, rawItem := range items {
		item, ok := rawItem.(map[string]any)
		if !ok {
			continue
		}
		rawID, ok := item["ID"].(string)
		if !ok {
			continue
		}
		id, err := gid.ParseGID(rawID)
		if err == nil {
			ids = append(ids, id)
		}
	}

	return ids
}
