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

package agent_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

func TestResult_FinalMessage(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns zero value when messages is empty",
		func(t *testing.T) {
			t.Parallel()

			r := &agent.Result{}
			msg := r.FinalMessage()

			assert.Equal(t, llm.Message{}, msg)
			assert.Equal(t, "", msg.Text())
		},
	)

	t.Run(
		"returns the last message with single message",
		func(t *testing.T) {
			t.Parallel()

			r := &agent.Result{
				Messages: []llm.Message{
					assistantMessage("Hello!"),
				},
			}

			assert.Equal(t, "Hello!", r.FinalMessage().Text())
		},
	)

	t.Run(
		"returns the last message with multiple messages",
		func(t *testing.T) {
			t.Parallel()

			r := &agent.Result{
				Messages: []llm.Message{
					userMessage("Hi"),
					assistantMessage("Hello!"),
					userMessage("How are you?"),
					assistantMessage("I'm fine."),
				},
			}

			assert.Equal(t, "I'm fine.", r.FinalMessage().Text())
		},
	)

	t.Run(
		"returns message regardless of role",
		func(t *testing.T) {
			t.Parallel()

			r := &agent.Result{
				Messages: []llm.Message{
					assistantMessage("first"),
					userMessage("last user message"),
				},
			}

			msg := r.FinalMessage()
			assert.Equal(t, llm.RoleUser, msg.Role)
			assert.Equal(t, "last user message", msg.Text())
		},
	)
}
