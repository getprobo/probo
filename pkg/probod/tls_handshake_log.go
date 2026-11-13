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

package probod

import (
	"strings"

	"go.gearno.de/kit/log"
)

// noisyTLSHandshakeFragments are client handshake failures. Scanners emit
// them constantly, and they do not describe a fault in this server.
var noisyTLSHandshakeFragments = []string{
	"tls: no certificates configured",
	"client sent an HTTP request to an HTTPS server",
	"tls: client offered only unsupported versions",
	"EOF",
	" i/o timeout",
	"tls: first record does not look like a TLS handshake",
	"tls: client requested unsupported application protocols",
	"read: connection reset by peer",
	"tls: unsupported SSLv2 handshake received",
	"tls: no cipher suite supported by both client and server",
	"tls: no key exchanges supported by both client and server",
	"tls: received record with version",
}

func skipNoisyTLSHandshake(_ log.Level, msg string, _ []log.Attr) bool {
	for _, fragment := range noisyTLSHandshakeFragments {
		if strings.Contains(msg, fragment) {
			return true
		}
	}

	return false
}
