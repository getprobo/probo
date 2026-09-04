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

package gcp_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
)

const setupOrganizationID = "e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhg"

func TestBuildConnectorSetup(t *testing.T) {
	t.Parallel()

	issuer := "https://proboidentity.com/" + setupOrganizationID
	subject := setupOrganizationID

	t.Run("fills every Probo-derived value and preserves issuer casing", func(t *testing.T) {
		t.Parallel()

		setup, err := cloudgcp.BuildConnectorSetup(
			cloudgcp.ConnectorSetupInput{
				IssuerURL:             issuer,
				Subject:               subject,
				TerraformModuleSource: cloudgcp.DefaultTerraformModuleSource,
			},
		)
		require.NoError(t, err)

		assert.Equal(t, issuer, setup.Issuer)
		assert.Equal(t, cloudgcp.AudienceTemplate, setup.Audience)
		assert.Equal(t, subject, setup.Subject)
		assert.Equal(t, cloudgcp.DefaultServiceAccountName, setup.SuggestedServiceAccountName)
		assert.Contains(t, setup.TerraformSnippet, strconv.Quote(issuer))
		assert.Contains(t, setup.TerraformSnippet, strconv.Quote(subject))
		assert.Contains(t, setup.TerraformSnippet, "probo_issuer_url")
		assert.Contains(t, setup.TerraformSnippet, "probo_subject")
		assert.Contains(t, setup.TerraformSnippet, "service_account_name")
		assert.Contains(t, setup.TerraformSnippet, strconv.Quote(cloudgcp.DefaultServiceAccountName))
		assert.Contains(t, setup.TerraformSnippet, cloudgcp.DefaultTerraformModuleSource)
	})

	t.Run("omits the snippet when the module source is empty", func(t *testing.T) {
		t.Parallel()

		setup, err := cloudgcp.BuildConnectorSetup(
			cloudgcp.ConnectorSetupInput{
				IssuerURL: issuer,
				Subject:   subject,
			},
		)
		require.NoError(t, err)

		assert.Empty(t, setup.TerraformSnippet)
	})

	t.Run("refuses a missing issuer", func(t *testing.T) {
		t.Parallel()

		_, err := cloudgcp.BuildConnectorSetup(cloudgcp.ConnectorSetupInput{Subject: subject})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issuer is required")
	})

	t.Run("refuses a missing subject", func(t *testing.T) {
		t.Parallel()

		_, err := cloudgcp.BuildConnectorSetup(cloudgcp.ConnectorSetupInput{IssuerURL: issuer})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "subject is required")
	})
}

func TestConnectorSetupFor(t *testing.T) {
	t.Parallel()

	organizationID := testOrganizationID()

	setup, err := cloudgcp.ConnectorSetupFor(
		testIssuer(t),
		organizationID,
		cloudgcp.ConnectorInstallConfig{TerraformModuleSource: cloudgcp.DefaultTerraformModuleSource},
	)
	require.NoError(t, err)

	assert.Contains(t, setup.Issuer, organizationID.String())
	assert.Equal(t, organizationID.String(), setup.Subject)
	assert.Equal(t, cloudgcp.AudienceTemplate, setup.Audience)
	assert.Contains(t, setup.TerraformSnippet, strconv.Quote(organizationID.String()))
}

func TestConnectorSetupFor_RequiresIssuer(t *testing.T) {
	t.Parallel()

	_, err := cloudgcp.ConnectorSetupFor(nil, testOrganizationID(), cloudgcp.ConnectorInstallConfig{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "identity federation is not configured")
}
