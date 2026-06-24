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
	handoffParamsSchema = mustJSONSchemaFor[handoffParams]()
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
