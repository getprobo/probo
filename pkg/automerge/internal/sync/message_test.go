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

package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMessageEncodeDecodeEmptyV2 reproduces the upstream
// encode_decode_empty_message test (rust/automerge/src/sync.rs): an empty V2
// sync message must encode and parse back without error. It additionally
// asserts the exact wire bytes so the native V2 codec stays byte-compatible
// with the reference, whose Message::encode writes the type byte followed by
// four zero ULEB collection counts (heads, need, have, changes) and no flags.
func TestMessageEncodeDecodeEmptyV2(t *testing.T) {
	t.Parallel()

	message := Message{Version: MessageVersion2}

	encoded, err := message.Encode()
	require.NoError(t, err)
	assert.Equal(t, []byte{byte(MessageVersion2), 0x00, 0x00, 0x00, 0x00}, encoded)

	parsed, err := ParseMessage(encoded)
	require.NoError(t, err)
	assert.Equal(t, MessageVersion2, parsed.Version)
	assert.Empty(t, parsed.Heads)
	assert.Empty(t, parsed.Need)
	assert.Empty(t, parsed.Have)
	assert.Empty(t, parsed.Changes)
	assert.Nil(t, parsed.Flags)

	reencoded, err := parsed.Encode()
	require.NoError(t, err)
	assert.Equal(t, encoded, reencoded)
}

func TestMessageEncodeRejectsOversizedByteFields(t *testing.T) {
	t.Parallel()

	message := Message{
		Version: MessageVersion2,
		Have: []Have{
			{Bloom: make([]byte, maxSyncBloomBytes+1)},
		},
	}
	_, err := message.Encode()
	assert.Error(t, err)

	message = Message{
		Version: MessageVersion2,
		Changes: [][]byte{make([]byte, maxSyncChunkBytes+1)},
	}
	_, err = message.Encode()
	assert.Error(t, err)

	message = Message{
		Version: MessageVersion2,
		Flags:   make([]byte, maxSyncFlagsBytes+1),
	}
	_, err = message.Encode()
	assert.Error(t, err)
}
