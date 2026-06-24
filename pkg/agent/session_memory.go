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
