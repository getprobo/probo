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
	"sync"

	"go.probo.inc/probo/pkg/llm"
)

var _ Session = (*memorySession)(nil)

type memorySession struct {
	mu       sync.RWMutex
	sessions map[string][]llm.Message
}

func NewMemorySession() *memorySession {
	return &memorySession{
		sessions: make(map[string][]llm.Message),
	}
}

func (s *memorySession) Load(_ context.Context, sessionID string) ([]llm.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msgs, ok := s.sessions[sessionID]
	if !ok {
		return nil, nil
	}

	cp := make([]llm.Message, len(msgs))
	for i, m := range msgs {
		cp[i] = copyMessage(m)
	}

	return cp, nil
}

func (s *memorySession) Save(_ context.Context, sessionID string, messages []llm.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := make([]llm.Message, len(messages))
	for i, m := range messages {
		cp[i] = copyMessage(m)
	}

	s.sessions[sessionID] = cp
	return nil
}

func copyMessage(m llm.Message) llm.Message {
	cp := llm.Message{
		Role:       m.Role,
		ToolCallID: m.ToolCallID,
	}

	if len(m.Parts) > 0 {
		cp.Parts = make([]llm.Part, len(m.Parts))
		copy(cp.Parts, m.Parts)
	}

	if len(m.ToolCalls) > 0 {
		cp.ToolCalls = make([]llm.ToolCall, len(m.ToolCalls))
		copy(cp.ToolCalls, m.ToolCalls)
	}

	return cp
}
