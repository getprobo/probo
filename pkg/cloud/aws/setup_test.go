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

package aws_test

import (
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/identityfederation"
)

const setupOrganizationID = "e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhg"

func TestBuildConnectorSetup(t *testing.T) {
	t.Parallel()

	issuer := "https://proboidentity.com/" + setupOrganizationID
	subject := setupOrganizationID

	t.Run("fills every Probo-derived value and preserves issuer casing", func(t *testing.T) {
		t.Parallel()

		setup, err := cloudaws.BuildConnectorSetup(
			cloudaws.ConnectorSetupInput{
				IssuerURL:                 issuer,
				Subject:                   subject,
				CloudFormationTemplateURL: cloudaws.DefaultCloudFormationTemplateURL,
				TerraformModuleSource:     cloudaws.DefaultTerraformModuleSource,
			},
		)
		require.NoError(t, err)

		assert.Equal(t, issuer, setup.Issuer)
		assert.Equal(t, identityfederation.AudienceAWS, setup.Audience)
		assert.Equal(t, subject, setup.Subject)
		assert.Equal(t, coredata.DefaultAWSRoleName, setup.SuggestedRoleName)
		assert.Contains(t, setup.TerraformSnippet, issuer)
		assert.Contains(t, setup.TerraformSnippet, subject)
		assert.Contains(t, setup.TerraformSnippet, "probo_issuer_url")
		assert.Contains(t, setup.TerraformSnippet, "role_name")
		assert.Contains(t, setup.TerraformSnippet, strconv.Quote(coredata.DefaultAWSRoleName))
		assert.Contains(t, setup.TerraformSnippet, cloudaws.DefaultTerraformModuleSource)
		assert.Contains(t, setup.CloudFormationQuickCreateURL, "quickcreate")
		assert.Contains(t, setup.CloudFormationQuickCreateURL, url.QueryEscape(cloudaws.DefaultCloudFormationTemplateURL))
		assert.Contains(t, setup.CloudFormationQuickCreateURL, url.QueryEscape(issuer))
		assert.Contains(t, setup.CloudFormationQuickCreateURL, url.QueryEscape(subject))
		assert.Contains(t, setup.CloudFormationQuickCreateURL, url.QueryEscape(coredata.DefaultAWSRoleName))
	})

	t.Run("omits install artifacts when their sources are empty", func(t *testing.T) {
		t.Parallel()

		setup, err := cloudaws.BuildConnectorSetup(
			cloudaws.ConnectorSetupInput{
				IssuerURL: issuer,
				Subject:   subject,
			},
		)
		require.NoError(t, err)

		assert.Empty(t, setup.TerraformSnippet)
		assert.Empty(t, setup.CloudFormationQuickCreateURL)
	})

	t.Run("refuses a missing issuer", func(t *testing.T) {
		t.Parallel()

		_, err := cloudaws.BuildConnectorSetup(cloudaws.ConnectorSetupInput{Subject: subject})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issuer is required")
	})
}

func TestNewConnectorSettings(t *testing.T) {
	t.Parallel()

	t.Run("stores commercial govcloud and china role arns unchanged", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			roleARN string
		}{
			{name: "commercial", roleARN: "arn:aws:iam::123456789012:role/ProboAudit"},
			{name: "govcloud", roleARN: "arn:aws-us-gov:iam::123456789012:role/ProboAudit"},
			{name: "china", roleARN: "arn:aws-cn:iam::123456789012:role/ProboAudit"},
			{name: "path", roleARN: "arn:aws:iam::123456789012:role/team/CustomAudit"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				settings, err := cloudaws.NewConnectorSettings(tt.roleARN)
				require.NoError(t, err)
				assert.Equal(t, tt.roleARN, settings.RoleARN)
			})
		}
	})

	t.Run("trims space and keeps the given arn", func(t *testing.T) {
		t.Parallel()

		roleARN := "arn:aws:iam::123456789012:role/ProboAudit"

		settings, err := cloudaws.NewConnectorSettings("  " + roleARN + "  ")
		require.NoError(t, err)
		assert.Equal(t, roleARN, settings.RoleARN)
	})

	t.Run("refuses a non-role arn without echoing it", func(t *testing.T) {
		t.Parallel()

		raw := "arn:aws:iam::123456789012:user/alice"

		_, err := cloudaws.NewConnectorSettings(raw)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "awsRoleArn is not an IAM role ARN")
		assert.NotContains(t, err.Error(), raw)
		assert.NotContains(t, err.Error(), "alice")
	})

	t.Run("refuses an unsupported partition without echoing it", func(t *testing.T) {
		t.Parallel()

		raw := "arn:aws-iso:iam::123456789012:role/ProboAudit"

		_, err := cloudaws.NewConnectorSettings(raw)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "supported AWS partition")
		assert.NotContains(t, err.Error(), raw)
		assert.NotContains(t, err.Error(), "aws-iso")
	})
}
