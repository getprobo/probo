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

package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

func TestNewAccessReviewEntry_SourceIdentity(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	connectorID := gid.New(tenantID, coredata.ConnectorEntityType)
	entry := &coredata.AccessReviewEntry{
		AccessReviewCampaignSourceID: gid.New(
			tenantID,
			coredata.AccessReviewCampaignSourceEntityType,
		),
	}

	tests := map[string]struct {
		source            *coredata.AccessReviewCampaignSource
		expectedName      string
		expectedConnector *gid.GID
	}{
		"connector source": {
			source: &coredata.AccessReviewCampaignSource{
				Name:        "Cursor",
				ConnectorID: &connectorID,
			},
			expectedName:      "Cursor",
			expectedConnector: &connectorID,
		},
		"manual source": {
			source: &coredata.AccessReviewCampaignSource{
				Name: "Uploaded CSV",
			},
			expectedName: "Uploaded CSV",
		},
	}

	for name, test := range tests {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				actual := types.NewAccessReviewEntry(entry, test.source)

				assert.Equal(t, test.expectedName, actual.SourceName)
				assert.Equal(t, test.expectedConnector, actual.ConnectorID)
			},
		)
	}
}

func TestNewAccessReviewEntry_ManualSourceIncludesNullConnectorID(t *testing.T) {
	t.Parallel()

	actual := types.NewAccessReviewEntry(
		&coredata.AccessReviewEntry{},
		&coredata.AccessReviewCampaignSource{Name: "Uploaded CSV"},
	)

	data, err := json.Marshal(actual)
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(data, &payload))
	assert.Contains(t, payload, "connector_id")
	assert.Nil(t, payload["connector_id"])
}
