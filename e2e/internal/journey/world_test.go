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

package journey

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorld_StepFailureMessagesAreRedacted(t *testing.T) {
	t.Parallel()

	raw := fmt.Errorf("request failed: Authorization: Bearer secret-frag-token").Error()
	redacted := redactSensitiveText(raw)

	assert.Contains(t, redacted, "[REDACTED]")
	assert.NotContains(t, redacted, "secret-frag-token")
	assert.Contains(t, redacted, "request failed:")
}

func TestRedactSensitiveTextPreservesNonSecretContext(t *testing.T) {
	t.Parallel()

	input := "step 02 [Alice] cannot load document doc-123: token=leak-value"
	actual := redactSensitiveText(input)

	assert.Contains(t, actual, "step 02 [Alice]")
	assert.Contains(t, actual, "doc-123")
	assert.NotContains(t, actual, "leak-value")
}
