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

package deviceagent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventMessageFromXML_ReturnsTheInsertionString(t *testing.T) {
	t.Parallel()

	raw := `<Event xmlns="http://schemas.microsoft.com/win/2004/08/events/event">
  <System><Provider Name="ProboAgent" /></System>
  <EventData><Data>{"level":"INFO","msg":"started"}</Data></EventData>
</Event>`

	message, err := eventMessageFromXML(raw)
	require.NoError(t, err)
	assert.Equal(t, `{"level":"INFO","msg":"started"}`, message)
}

func TestWevtutilQueryArgs_SelectsNewestProboAgentEvents(t *testing.T) {
	t.Parallel()

	args := wevtutilQueryArgs(64 << 10)

	assert.Equal(t, "qe", args[0])
	assert.Equal(t, "Application", args[1])
	assert.Contains(t, args, "/q:*[System[Provider[@Name='ProboAgent']]]")
	assert.Contains(t, args, "/rd:true")
	assert.Contains(t, args, "/f:xml")
	assert.Contains(t, args, "/c:2048")
}

func TestDecodeCommandOutput_ReadsUTF16LE(t *testing.T) {
	t.Parallel()

	out := []byte{0xFF, 0xFE, 'A', 0x00, 'B', 0x00}

	assert.Equal(t, "AB", decodeCommandOutput(out))
	assert.Equal(t, "AB", decodeCommandOutput([]byte("AB")))
}

func TestEventMessagesFromXML_ReadsEveryEvent(t *testing.T) {
	t.Parallel()

	raw := `<?xml version="1.0" encoding="UTF-16"?>
<Event xmlns="http://schemas.microsoft.com/win/2004/08/events/event"><EventData><Data>newest</Data></EventData></Event>
<Event xmlns="http://schemas.microsoft.com/win/2004/08/events/event"><EventData><Data>older</Data></EventData></Event>`

	messages, err := eventMessagesFromXML(raw)
	require.NoError(t, err)
	assert.Equal(t, []string{"newest", "older"}, messages)
}

func TestJoinNewestEventLines_KeepsTheNewestWithinTheLimit(t *testing.T) {
	t.Parallel()

	text := joinNewestEventLines([]string{"newest", "middle", "oldest"}, len("newest\nmiddle"))

	assert.Equal(t, "middle\nnewest", text)
}
