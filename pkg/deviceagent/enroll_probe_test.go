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
)

func TestProbeEnrollmentTrust_NoNetwork(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  EnrollmentTrust
	}{
		{name: "non-loopback http", input: "http://evil.example", want: TrustInsecure},
		{name: "http Probo host", input: "http://us.probo.com", want: TrustInsecure},
		{name: "loopback http", input: "http://localhost:3000", want: TrustUnknown},
		{name: "loopback ipv4 http", input: "http://127.0.0.1:3000", want: TrustUnknown},
		{name: "https self-hosted", input: "https://probo.example.com", want: TrustUnverified},
		{name: "empty", input: "", want: TrustUnknown},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got := ProbeEnrollmentTrust(t.Context(), tt.input)
				assert.Equal(t, tt.want, got)
			},
		)
	}
}
