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

package drivers

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocuSignDriver(t *testing.T) {
	t.Parallel()

	// Account UUID matches the userinfo cassette: the driver resolves the
	// selected account's data-center base URI before listing its users.
	const accountID = "a1a1a1a1-1111-4111-8111-111111111111"

	rec := newRecorder(t, "testdata/docusign", "DOCUSIGN_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("DOCUSIGN_TOKEN")))
	driver := NewDocuSignDriver(client, accountID)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	assert.Len(t, records, 2)

	r := records[0]
	assert.Equal(t, "jane.doe@example.com", r.Email)
	assert.Equal(t, "Jane Doe", r.FullName)
	assert.Equal(t, "11111111-1111-4111-8111-111111111111", r.ExternalID)
	assert.Equal(t, []string{"Account Administrator"}, r.Roles)
	assert.Equal(t, "CTO", r.JobTitle)
	assert.True(t, r.IsAdmin)
	require.NotNil(t, r.Active)
	assert.True(t, *r.Active)

	assert.False(t, records[1].IsAdmin)
}
