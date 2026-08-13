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

package probot

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/gid"
)

var (
	ErrCapabilityNotFound     = errors.New("probot capability not found")
	ErrCapabilityForbidden    = errors.New("probot capability action forbidden")
	ErrCapabilityInvalidInput = errors.New("invalid probot capability input")
)

type (
	Action struct {
		ID               string
		Value            string
		SelectedValue    string
		DeduplicationKey string
		ResponseToken    string
		ActorIdentityID  gid.GID
		Message          Message
	}

	ActionResult struct {
		Message string
	}

	// MessageAnchor identifies a provider message within its conversation.
	// Values are opaque provider coordinates; channel adapters define their
	// concrete representation.
	MessageAnchor struct {
		ConversationID string
		MessageID      string
	}

	// DeliveryTarget identifies a provider-neutral delivery destination.
	// Namespace is owned by the capability and Key identifies the target
	// within that namespace.
	DeliveryTarget struct {
		Namespace string
		Key       string
	}

	RunContext struct {
		OrganizationID   gid.GID
		ExecutionID      gid.GID
		MessageAnchor    MessageAnchor
		CurrentMessageID string
		IdentityID       gid.GID
		Capability       string
		MessageType      string
		Attributes       map[string]any
	}

	Capability interface {
		Name() string
	}

	MessageTypeContributor interface {
		MessageTypes() []string
	}

	ActionPrefixContributor interface {
		ActionPrefixes() []string
	}

	ToolContributor interface {
		Tools() []agent.Tool
	}

	OutboundMessage struct {
		Message        Message
		Intent         MessageIntent
		DeliveryTarget DeliveryTarget
	}

	OutboundMessageCapability interface {
		BuildOutboundMessage(
			ctx context.Context,
			organizationID gid.GID,
			messageType string,
			attributes map[string]any,
		) (OutboundMessage, error)
	}

	MessageCapability interface {
		MessageTypeContributor
		RenderMessage(ctx context.Context, message Message) (MessageIntent, error)
	}

	ActionCapability interface {
		ActionPrefixContributor
		HandleAction(ctx context.Context, action Action) (ActionResult, error)
	}

	ActionAliasContributor interface {
		NormalizeActionAlias(action Action) (Action, error)
	}

	CapabilityRegistry struct {
		mu             sync.RWMutex
		byName         map[string]Capability
		byMessageType  map[string]Capability
		byActionPrefix map[string]Capability
		byToolName     map[string]Capability
	}
)

func NewCapabilityRegistry() *CapabilityRegistry {
	return &CapabilityRegistry{
		byName:         make(map[string]Capability),
		byMessageType:  make(map[string]Capability),
		byActionPrefix: make(map[string]Capability),
		byToolName:     make(map[string]Capability),
	}
}

func (r *CapabilityRegistry) Register(capability Capability) error {
	if capability == nil {
		return fmt.Errorf("cannot register nil Probot capability")
	}

	name := strings.TrimSpace(capability.Name())
	if name == "" {
		return fmt.Errorf("cannot register Probot capability without a name")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byName[name]; ok {
		return fmt.Errorf("cannot register duplicate Probot capability %q", name)
	}

	messageCapability, contributesMessages := capability.(MessageCapability)
	if contributesMessages {
		for _, messageType := range messageCapability.MessageTypes() {
			if owner, ok := r.byMessageType[messageType]; ok {
				return fmt.Errorf(
					"cannot register message type %q: already owned by %q",
					messageType,
					owner.Name(),
				)
			}
		}
	}

	actionCapability, contributesActions := capability.(ActionCapability)
	if contributesActions {
		for _, prefix := range actionCapability.ActionPrefixes() {
			if prefix == "" {
				return fmt.Errorf("cannot register empty action prefix for Probot capability %q", name)
			}

			if owner, ok := r.byActionPrefix[prefix]; ok {
				return fmt.Errorf(
					"cannot register action prefix %q: already owned by %q",
					prefix,
					owner.Name(),
				)
			}
		}
	}

	toolContributor, contributesTools := capability.(ToolContributor)
	if contributesTools {
		for _, tool := range toolContributor.Tools() {
			if owner, ok := r.byToolName[tool.Name()]; ok {
				return fmt.Errorf(
					"cannot register agent tool %q: already owned by %q",
					tool.Name(),
					owner.Name(),
				)
			}
		}
	}

	r.byName[name] = capability

	if contributesMessages {
		for _, messageType := range messageCapability.MessageTypes() {
			r.byMessageType[messageType] = capability
		}
	}

	if contributesActions {
		for _, prefix := range actionCapability.ActionPrefixes() {
			r.byActionPrefix[prefix] = capability
		}
	}

	if contributesTools {
		for _, tool := range toolContributor.Tools() {
			r.byToolName[tool.Name()] = capability
		}
	}

	return nil
}

func (r *CapabilityRegistry) Tools() []agent.Tool {
	if r == nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.byName))
	for name := range r.byName {
		names = append(names, name)
	}

	sort.Strings(names)

	tools := make([]agent.Tool, 0)

	for _, name := range names {
		contributor, ok := r.byName[name].(ToolContributor)
		if ok {
			tools = append(tools, contributor.Tools()...)
		}
	}

	return tools
}

func (r *CapabilityRegistry) ToolsForMessageType(messageType string) []agent.Tool {
	if r == nil || messageType == "" {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	contributor, ok := r.byMessageType[messageType].(ToolContributor)
	if !ok {
		return nil
	}

	return contributor.Tools()
}

func (r *CapabilityRegistry) HandleAction(
	ctx context.Context,
	action Action,
) (ActionResult, error) {
	if r == nil || action.ID == "" {
		return ActionResult{}, ErrCapabilityNotFound
	}

	r.mu.RLock()

	var (
		capability    Capability
		matchedPrefix string
	)
	for prefix, candidate := range r.byActionPrefix {
		if strings.HasPrefix(action.ID, prefix) && len(prefix) > len(matchedPrefix) {
			capability = candidate
			matchedPrefix = prefix
		}
	}

	messageOwner := r.byMessageType[action.Message.Type]
	r.mu.RUnlock()

	if capability == nil {
		return ActionResult{}, fmt.Errorf("%w for action %q", ErrCapabilityNotFound, action.ID)
	}

	if messageOwner != nil && messageOwner != capability {
		return ActionResult{}, fmt.Errorf(
			"%w: action %q is not owned by message capability %q",
			ErrCapabilityInvalidInput,
			action.ID,
			messageOwner.Name(),
		)
	}

	handler := capability.(ActionCapability)

	return handler.HandleAction(ctx, action)
}

func (r *CapabilityRegistry) NormalizeActionAlias(action Action) (Action, error) {
	if r == nil || action.Message.Type == "" {
		return action, nil
	}

	r.mu.RLock()
	contributor, ok := r.byMessageType[action.Message.Type].(ActionAliasContributor)
	r.mu.RUnlock()
	if !ok {
		return action, nil
	}

	return contributor.NormalizeActionAlias(action)
}

func (r *CapabilityRegistry) BuildOutboundMessage(
	ctx context.Context,
	capabilityName string,
	organizationID gid.GID,
	messageType string,
	attributes map[string]any,
) (OutboundMessage, error) {
	if r == nil || capabilityName == "" || messageType == "" {
		return OutboundMessage{}, ErrCapabilityInvalidInput
	}

	r.mu.RLock()
	capability, ok := r.byName[capabilityName]
	messageOwner := r.byMessageType[messageType]
	r.mu.RUnlock()

	if !ok {
		return OutboundMessage{}, fmt.Errorf(
			"%w for capability %q",
			ErrCapabilityNotFound,
			capabilityName,
		)
	}

	if messageOwner != capability {
		return OutboundMessage{}, fmt.Errorf(
			"%w: message type %q is not owned by capability %q",
			ErrCapabilityInvalidInput,
			messageType,
			capabilityName,
		)
	}

	handler, ok := capability.(OutboundMessageCapability)
	if !ok {
		return OutboundMessage{}, fmt.Errorf(
			"%w: capability %q does not build outbound messages",
			ErrCapabilityInvalidInput,
			capabilityName,
		)
	}

	return handler.BuildOutboundMessage(ctx, organizationID, messageType, attributes)
}

func (r *CapabilityRegistry) RenderMessage(
	ctx context.Context,
	message Message,
) (MessageIntent, error) {
	if r == nil || message.Type == "" {
		return MessageIntent{}, ErrCapabilityNotFound
	}

	r.mu.RLock()
	capability, ok := r.byMessageType[message.Type]
	r.mu.RUnlock()

	if !ok {
		return MessageIntent{}, fmt.Errorf(
			"%w for message type %q",
			ErrCapabilityNotFound,
			message.Type,
		)
	}

	renderer := capability.(MessageCapability)

	return renderer.RenderMessage(ctx, message)
}
