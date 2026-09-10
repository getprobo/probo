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

package azure_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	"go.probo.inc/probo/pkg/identityfederation"
)

func TestBuildConnectorSetup(t *testing.T) {
	t.Parallel()

	issuer := "https://proboidentity.com/" + setupOrganizationID
	subject := setupOrganizationID

	t.Run(
		"fills every Probo-derived value and preserves issuer casing",
		func(t *testing.T) {
			t.Parallel()

			setup, err := cloudazure.BuildConnectorSetup(
				cloudazure.ConnectorSetupInput{
					IssuerURL:             issuer,
					Subject:               subject,
					TerraformModuleSource: cloudazure.DefaultTerraformModuleSource,
				},
			)
			require.NoError(t, err)

			assert.Equal(t, issuer, setup.Issuer)
			assert.Equal(t, identityfederation.AudienceAzure, setup.Audience)
			assert.Equal(t, subject, setup.Subject)
			assert.Equal(t, cloudazure.DefaultApplicationName, setup.SuggestedApplicationName)
			assert.Contains(t, setup.TerraformSnippet, strconv.Quote(issuer))
			assert.Contains(t, setup.TerraformSnippet, strconv.Quote(subject))
			assert.Contains(t, setup.TerraformSnippet, "probo_issuer_url")
			assert.Contains(t, setup.TerraformSnippet, "probo_subject")
			assert.Contains(t, setup.TerraformSnippet, "application_display_name")
			assert.Contains(t, setup.TerraformSnippet, "subscription_id")
			assert.Contains(t, setup.TerraformSnippet, strconv.Quote(cloudazure.DefaultApplicationName))
			assert.Contains(t, setup.TerraformSnippet, cloudazure.DefaultTerraformModuleSource)
			assert.Contains(t, setup.TerraformSnippet, "azurerm")
			assert.Contains(t, setup.TerraformSnippet, "azuread")
		},
	)

	t.Run(
		"omits the snippet when the module source is empty",
		func(t *testing.T) {
			t.Parallel()

			setup, err := cloudazure.BuildConnectorSetup(
				cloudazure.ConnectorSetupInput{
					IssuerURL: issuer,
					Subject:   subject,
				},
			)
			require.NoError(t, err)

			assert.Empty(t, setup.TerraformSnippet)
		},
	)

	t.Run(
		"refuses a missing issuer",
		func(t *testing.T) {
			t.Parallel()

			_, err := cloudazure.BuildConnectorSetup(cloudazure.ConnectorSetupInput{Subject: subject})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "issuer is required")
		},
	)

	t.Run(
		"refuses a missing subject",
		func(t *testing.T) {
			t.Parallel()

			_, err := cloudazure.BuildConnectorSetup(cloudazure.ConnectorSetupInput{IssuerURL: issuer})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "subject is required")
		},
	)
}

func TestConnectorSetupFor(t *testing.T) {
	t.Parallel()

	organizationID := testOrganizationID()

	setup, err := cloudazure.ConnectorSetupFor(
		testIssuer(t),
		organizationID,
		cloudazure.ConnectorInstallConfig{TerraformModuleSource: cloudazure.DefaultTerraformModuleSource},
	)
	require.NoError(t, err)

	assert.Contains(t, setup.Issuer, organizationID.String())
	assert.Equal(t, organizationID.String(), setup.Subject)
	assert.Equal(t, identityfederation.AudienceAzure, setup.Audience)
	assert.Contains(t, setup.TerraformSnippet, strconv.Quote(organizationID.String()))
}

func TestConnectorSetupFor_RequiresIssuer(t *testing.T) {
	t.Parallel()

	_, err := cloudazure.ConnectorSetupFor(nil, testOrganizationID(), cloudazure.ConnectorInstallConfig{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "identity federation is not configured")
}
