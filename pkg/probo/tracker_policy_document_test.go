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

package probo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/docgen"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/prosemirror"
)

func TestBuildTrackerPolicyDocument_OmitsCoverMetadata(t *testing.T) {
	t.Parallel()

	content, err := probo.BuildTrackerPolicyDocument(
		docgen.TrackerPolicyData{
			OrganizationName:  "Acme",
			WebsiteOrigin:     "https://example.com",
			ConsentExpiryDays: 180,
		},
	)
	require.NoError(t, err)

	document, err := prosemirror.Parse(content)
	require.NoError(t, err)
	require.NotEmpty(t, document.Content)

	assert.Equal(t, prosemirror.NodeParagraph, document.Content[0].Type)
	assert.NotContains(t, content, "Policy for https://example.com")
	assert.NotContains(t, content, "Last updated")
}
