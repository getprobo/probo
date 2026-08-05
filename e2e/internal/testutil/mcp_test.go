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

package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeMCPResponseBody_JSON(t *testing.T) {
	t.Parallel()

	body := []byte(`{"jsonrpc":"2.0","id":1,"result":{}}`)

	decoded, err := decodeMCPResponseBody("application/json", body)
	require.NoError(t, err)
	assert.Equal(t, body, decoded)
}

func TestDecodeMCPResponseBody_EventStream(t *testing.T) {
	t.Parallel()

	body := []byte(
		"event: message\n" +
			"id: event-id\n" +
			`data: {"jsonrpc":"2.0","id":1,"result":{"ok":true}}` +
			"\n\n",
	)

	decoded, err := decodeMCPResponseBody("text/event-stream; charset=utf-8", body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"jsonrpc":"2.0","id":1,"result":{"ok":true}}`, string(decoded))
}
