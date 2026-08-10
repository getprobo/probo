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

package slackbot_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/slackbot"
)

func TestTools_OmitChannelFromLLMParams(t *testing.T) {
	t.Parallel()

	for _, tool := range slackbot.Tools(nil) {
		t.Run(
			tool.Name(),
			func(t *testing.T) {
				t.Parallel()

				var schema struct {
					Properties map[string]json.RawMessage `json:"properties"`
				}
				err := json.Unmarshal(tool.Definition().Parameters, &schema)
				require.NoError(t, err)

				_, hasChannel := schema.Properties["channel"]
				assert.False(t, hasChannel, "channel must come from RunContext, not LLM params")

				_, hasThreadTS := schema.Properties["thread_ts"]
				assert.False(t, hasThreadTS, "thread_ts must come from RunContext, not LLM params")
			},
		)
	}
}
