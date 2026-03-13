// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package agent

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"go.probo.inc/probo/pkg/llm"
)

type (
	HandoffInputData struct {
		InputHistory []llm.Message
		NewItems     []llm.Message
	}

	HandoffInputFilter func(data HandoffInputData) []llm.Message

	HandoffOption func(*Handoff)

	Handoff struct {
		Agent           *Agent
		ToolName        string
		ToolDescription string
		InputFilter     HandoffInputFilter
		OnHandoff       func(ctx context.Context) error
	}

	handoffParams struct{}
)

var (
	handoffParamsSchema = jsonSchemaFor[handoffParams]()
)

func HandoffTo(agent *Agent, opts ...HandoffOption) *Handoff {
	h := &Handoff{Agent: agent}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func WithHandoffToolName(name string) HandoffOption {
	return func(h *Handoff) {
		h.ToolName = name
	}
}

func WithHandoffToolDescription(desc string) HandoffOption {
	return func(h *Handoff) {
		h.ToolDescription = desc
	}
}

func WithHandoffInputFilter(fn HandoffInputFilter) HandoffOption {
	return func(h *Handoff) {
		h.InputFilter = fn
	}
}

func WithOnHandoff(fn func(ctx context.Context) error) HandoffOption {
	return func(h *Handoff) {
		h.OnHandoff = fn
	}
}

func (h *Handoff) toolName() string {
	if h.ToolName != "" {
		return h.ToolName
	}
	return "transfer_to_" + sanitizeToolName(h.Agent.name)
}

func sanitizeToolName(name string) string {
	var b strings.Builder
	b.Grow(len(name))

	for _, r := range name {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}

	return strings.ToLower(b.String())
}

func (h *Handoff) toolDescription() string {
	if h.ToolDescription != "" {
		return h.ToolDescription
	}

	desc := fmt.Sprintf(
		"Transfer the conversation to %s.",
		h.Agent.name,
	)

	if h.Agent.handoffDescription != "" {
		desc += " " + h.Agent.handoffDescription
	}

	return desc
}

func (h *Handoff) tool() ToolDescriptor {
	return &handoffToolAdapter{handoff: h}
}

type handoffToolAdapter struct {
	handoff *Handoff
}

func (t *handoffToolAdapter) Name() string {
	return t.handoff.toolName()
}

func (t *handoffToolAdapter) Definition() llm.Tool {
	return llm.Tool{
		Name:        t.handoff.toolName(),
		Description: t.handoff.toolDescription(),
		Parameters:  handoffParamsSchema,
	}
}
