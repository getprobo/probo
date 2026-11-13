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

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const createDatadogAPIKeyConnectorMutation = `
	mutation($input: CreateAPIKeyConnectorInput!) {
		createAPIKeyConnector(input: $input) {
			connector { id provider selectedOrganization }
		}
	}
`

func TestCreateAPIKeyConnectorDatadog(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	var result struct {
		CreateAPIKeyConnector struct {
			Connector struct {
				ID                   string `json:"id"`
				Provider             string `json:"provider"`
				SelectedOrganization string `json:"selectedOrganization"`
			} `json:"connector"`
		} `json:"createAPIKeyConnector"`
	}

	err := owner.Execute(createDatadogAPIKeyConnectorMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Datadog",
			"provider":       "DATADOG",
			"apiKey":         "ddsat_e2eAlias_0123456789abcdef0123456789abcdef",
			"datadogSite":    "EU1",
		},
	}, &result)
	require.NoError(t, err)
	assert.NotEmpty(t, result.CreateAPIKeyConnector.Connector.ID)
	assert.Equal(t, "DATADOG", result.CreateAPIKeyConnector.Connector.Provider)
	assert.Equal(t, "datadoghq.eu", result.CreateAPIKeyConnector.Connector.SelectedOrganization)
}
