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

func TestElasticCloudDriver_ListAccounts(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/elastic_cloud", "ELASTIC_CLOUD_API_KEY")
	client := newVCRClientWithHeader(
		rec,
		"Authorization",
		"ApiKey "+os.Getenv("ELASTIC_CLOUD_API_KEY"),
	)

	organizationID := os.Getenv("ELASTIC_CLOUD_ORGANIZATION_ID")
	if organizationID == "" {
		organizationID = "00000000000000000000000000000000"
	}

	driver := NewElasticCloudDriver(
		client,
		organizationID,
		"https://api.elastic-cloud.com/api/v1",
	)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	for _, record := range records {
		assert.NotEmpty(t, record.Email)
		assert.NotEmpty(t, record.ExternalID)
		assert.Nil(t, record.Active)
		assert.Nil(t, record.CreatedAt)
		assert.Nil(t, record.LastLogin)
	}
}

func TestElasticCloudRoles(t *testing.T) {
	t.Parallel()

	var assignments elasticCloudRoleAssignments
	assignments.Platform = []elasticCloudRoleAssignment{{RoleID: "platform-viewer"}}
	assignments.Organization = []elasticCloudRoleAssignment{{RoleID: "organization-admin"}}
	assignments.Deployment = []elasticCloudRoleAssignment{{RoleID: "deployment-editor"}}
	assignments.Project.Elasticsearch = []elasticCloudRoleAssignment{{RoleID: "elasticsearch-admin"}}
	assignments.Project.Observability = []elasticCloudRoleAssignment{{RoleID: "observability-viewer"}}

	roles := elasticCloudRoles(assignments)

	assert.Equal(
		t,
		[]string{
			"platform:platform-viewer",
			"organization:organization-admin",
			"deployment:deployment-editor",
			"project:elasticsearch:elasticsearch-admin",
			"project:observability:observability-viewer",
		},
		roles,
	)
	assert.True(t, elasticCloudRolesIncludeAdmin(roles))
	assert.False(t, elasticCloudRolesIncludeAdmin([]string{
		"organization:billing-viewer",
		"deployment:deployment-editor",
	}))
}
