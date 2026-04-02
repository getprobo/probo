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

package security

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtocolName(t *testing.T) {
	t.Parallel()

	t.Run(
		"known protocols",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "TLS 1.0", protocolName(tls.VersionTLS10))
			assert.Equal(t, "TLS 1.1", protocolName(tls.VersionTLS11))
			assert.Equal(t, "TLS 1.2", protocolName(tls.VersionTLS12))
			assert.Equal(t, "TLS 1.3", protocolName(tls.VersionTLS13))
		},
	)

	t.Run(
		"unknown protocol",
		func(t *testing.T) {
			t.Parallel()

			result := protocolName(0x9999)
			assert.Contains(t, result, "unknown")
		},
	)
}
