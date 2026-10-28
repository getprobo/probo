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

package asana

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	client "go.probo.inc/probo/pkg/task/sync/asana/client"
)

func TestAsanaUserLookupMessageOmitsTheRequestURL(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf(
		"cannot call asana: Get \"https://app.client.com/api/1.0/workspaces/1/typeahead?query=ada@example.com\": refused",
	)
	message := asanaUserLookupMessage(err)
	assert.Equal(t, "asana user lookup failed", message)
	assert.NotContains(t, message, "ada@example.com")

	message = asanaUserLookupMessage(
		&client.APIError{
			Status:  429,
			Message: "query=ada@example.com",
		},
	)
	assert.Equal(t, "asana user lookup failed: status 429", message)
	assert.NotContains(t, message, "ada@example.com")
}
