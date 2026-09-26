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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

const elasticCloudCassetteOrganizationID = "org-example"

func sanitizeElasticCloudMembers(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize elastic cloud response with status %d", i.Response.Code)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal([]byte(i.Response.Body), &body); err != nil {
		return fmt.Errorf("cannot decode recorded elastic cloud response: %w", err)
	}

	raw, ok := body["members"]
	if !ok {
		return fmt.Errorf("recorded elastic cloud response has no members field")
	}

	var members []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &members); err != nil {
		return fmt.Errorf("cannot decode recorded elastic cloud members: %w", err)
	}

	if len(members) == 0 {
		return fmt.Errorf("recorded elastic cloud response lists no members")
	}

	var recorded []string

	for idx, member := range members {
		for _, field := range []string{"user_id", "email", "organization_id"} {
			value, err := elasticCloudRequiredString(member, field, idx)
			if err != nil {
				return err
			}

			recorded = append(recorded, value)
		}

		if value, present, err := elasticCloudOptionalString(member, "name", idx); err != nil {
			return err
		} else if present {
			recorded = append(recorded, value)
		}

		if value, present, err := elasticCloudOptionalString(member, "member_since", idx); err != nil {
			return err
		} else if present {
			recorded = append(recorded, value)
		}

		member["user_id"] = json.RawMessage(fmt.Sprintf(`"ec-user-%d"`, idx+1))
		member["email"] = json.RawMessage(fmt.Sprintf(`"member%d@example.com"`, idx+1))
		member["name"] = json.RawMessage(fmt.Sprintf(`"Member %d"`, idx+1))
		member["organization_id"] = json.RawMessage(`"` + elasticCloudCassetteOrganizationID + `"`)
		member["member_since"] = json.RawMessage(`"2024-01-02T03:04:05Z"`)

		if raw, ok := member["role_assignments"]; ok {
			rewritten, err := rewriteElasticCloudOrganizationIDs(raw, &recorded)
			if err != nil {
				return fmt.Errorf("cannot sanitize elastic cloud member %d role assignments: %w", idx, err)
			}

			member["role_assignments"] = rewritten
		}
	}

	sanitizedMembers, err := json.Marshal(members)
	if err != nil {
		return fmt.Errorf("cannot re-encode elastic cloud members: %w", err)
	}

	body["members"] = sanitizedMembers

	sanitized, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cannot re-encode elastic cloud response: %w", err)
	}

	for _, value := range recorded {
		if strings.Contains(string(sanitized), value) {
			return fmt.Errorf("sanitized elastic cloud response still contains a recorded identity value")
		}
	}

	if err := rewriteElasticCloudRequestURL(i); err != nil {
		return err
	}

	replaceCassetteBody(i, string(sanitized))

	return nil
}

func elasticCloudRequiredString(member map[string]json.RawMessage, field string, idx int) (string, error) {
	raw, ok := member[field]
	if !ok {
		return "", fmt.Errorf("recorded elastic cloud member %d has no %s field", idx, field)
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("recorded elastic cloud member %d has a non-string %s: %w", idx, field, err)
	}

	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("recorded elastic cloud member %d has an empty %s", idx, field)
	}

	return value, nil
}

func elasticCloudOptionalString(member map[string]json.RawMessage, field string, idx int) (string, bool, error) {
	raw, ok := member[field]
	if !ok {
		return "", false, nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", false, fmt.Errorf("recorded elastic cloud member %d has a non-string %s: %w", idx, field, err)
	}

	return value, strings.TrimSpace(value) != "", nil
}

func rewriteElasticCloudOrganizationIDs(raw json.RawMessage, recorded *[]string) (json.RawMessage, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}

	rewriteElasticCloudOrganizationIDsValue(value, recorded)

	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	return encoded, nil
}

func rewriteElasticCloudOrganizationIDsValue(value any, recorded *[]string) {
	switch typed := value.(type) {
	case map[string]any:
		if orgID, ok := typed["organization_id"].(string); ok && strings.TrimSpace(orgID) != "" {
			*recorded = append(*recorded, orgID)
			typed["organization_id"] = elasticCloudCassetteOrganizationID
		}

		for _, nested := range typed {
			rewriteElasticCloudOrganizationIDsValue(nested, recorded)
		}
	case []any:
		for _, nested := range typed {
			rewriteElasticCloudOrganizationIDsValue(nested, recorded)
		}
	}
}

func rewriteElasticCloudRequestURL(i *cassette.Interaction) error {
	parsed, err := url.Parse(i.Request.URL)
	if err != nil {
		return fmt.Errorf("cannot parse recorded elastic cloud request URL: %w", err)
	}

	segments := strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/")
	for idx, segment := range segments {
		if segment == "organizations" && idx+1 < len(segments) {
			segments[idx+1] = url.PathEscape(elasticCloudCassetteOrganizationID)
			break
		}
	}

	parsed.Path = "/" + strings.Join(segments, "/")
	parsed.RawPath = ""
	i.Request.URL = parsed.String()

	return nil
}

func TestElasticCloudDriver_ListAccounts(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/elastic_cloud", "ELASTIC_CLOUD_API_KEY", sanitizeElasticCloudMembers)
	client := newVCRClientWithHeader(
		rec,
		"Authorization",
		"ApiKey "+os.Getenv("ELASTIC_CLOUD_API_KEY"),
	)

	organizationID := os.Getenv("ELASTIC_CLOUD_ORGANIZATION_ID")
	if organizationID == "" {
		organizationID = elasticCloudCassetteOrganizationID
	}

	driver := NewElasticCloudDriver(
		client,
		organizationID,
		"https://api.elastic-cloud.com/api/v1",
	)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	assert.Equal(t, "member1@example.com", records[0].Email)
	assert.Equal(t, "Member 1", records[0].FullName)
	assert.Equal(t, "ec-user-1", records[0].ExternalID)
	assert.Equal(t, []string{"organization:organization-admin"}, records[0].Roles)
	require.NotNil(t, records[0].IsAdmin)
	assert.True(t, *records[0].IsAdmin)
	assert.Equal(t, coredata.MFAStatusUnknown, records[0].MFAStatus)
	assert.Nil(t, records[0].Active)
	assert.Nil(t, records[0].CreatedAt)
	assert.Nil(t, records[0].LastLogin)
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
