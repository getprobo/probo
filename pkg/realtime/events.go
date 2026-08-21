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

package realtime

import "sync"

const DocumentCollaborationChannel = "document_collaboration_changed"

type (
	Handler func(payload string)

	Events struct {
		mu          sync.RWMutex
		handlers    map[uint64]Handler
		nextHandler uint64
	}
)

func NewEvents() *Events {
	return &Events{handlers: make(map[uint64]Handler)}
}

func (e *Events) Subscribe(handler Handler) func() {
	e.mu.Lock()
	id := e.nextHandler
	e.nextHandler++
	e.handlers[id] = handler
	e.mu.Unlock()

	return func() {
		e.mu.Lock()
		delete(e.handlers, id)
		e.mu.Unlock()
	}
}

func (e *Events) Publish(payload string) {
	e.mu.RLock()

	handlers := make([]Handler, 0, len(e.handlers))
	for _, handler := range e.handlers {
		handlers = append(handlers, handler)
	}

	e.mu.RUnlock()

	for _, handler := range handlers {
		handler(payload)
	}
}
