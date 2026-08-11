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

package visitor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNdaConsentText(t *testing.T) {
	t.Parallel()

	t.Run(
		"uses the portal contact email when set",
		func(t *testing.T) {
			t.Parallel()

			text := ndaConsentText("compliance@acme.com")

			assert.Contains(t, text, "please contact compliance@acme.com.")
			assert.NotContains(t, text, DefaultNDAContactEmail)
		},
	)

	t.Run(
		"falls back to the default contact email when empty",
		func(t *testing.T) {
			t.Parallel()

			text := ndaConsentText("")

			assert.Contains(t, text, "please contact "+DefaultNDAContactEmail+".")
		},
	)

	t.Run(
		"fallback matches the previous hardcoded default text",
		func(t *testing.T) {
			t.Parallel()

			const previousDefault = "By clicking \"Review and sign\", I consent to sign this document electronically and agree that my electronic signature has the same legal validity as a handwritten signature. If you have questions about the NDA, please contact security@probo.com."

			assert.Equal(t, previousDefault, ndaConsentText(""))
		},
	)
}
