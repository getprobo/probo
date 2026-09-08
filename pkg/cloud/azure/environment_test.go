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

package azure

import (
	"net/url"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		environment   Environment
		configuration cloud.Configuration
		graphBaseURL  string
	}{
		{
			name:          "public",
			environment:   EnvironmentPublic,
			configuration: cloud.AzurePublic,
			graphBaseURL:  "https://graph.microsoft.com",
		},
		{
			name:          "government",
			environment:   EnvironmentGovernment,
			configuration: cloud.AzureGovernment,
			graphBaseURL:  "https://graph.microsoft.us",
		},
		{
			name:          "government dod",
			environment:   EnvironmentGovernmentDoD,
			configuration: cloud.AzureGovernment,
			graphBaseURL:  "https://dod-graph.microsoft.us",
		},
		{
			name:          "china",
			environment:   EnvironmentChina,
			configuration: cloud.AzureChina,
			graphBaseURL:  "https://microsoftgraph.chinacloudapi.cn",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				ep, err := tt.environment.endpoints()
				require.NoError(t, err)
				assert.Equal(t, tt.configuration, ep.configuration)
				assert.Equal(t, tt.graphBaseURL, ep.graphBaseURL)

				wantGraphScope, err := url.JoinPath(tt.graphBaseURL, ".default")
				require.NoError(t, err)
				assert.Equal(t, wantGraphScope, ep.graphScope)

				wantARMScope, err := url.JoinPath(
					tt.configuration.Services[cloud.ResourceManager].Audience,
					".default",
				)
				require.NoError(t, err)
				assert.Equal(t, wantARMScope, ep.armScope)
			},
		)
	}
}

func TestEnvironmentTable_DoDSharesAuthorityWithGovernment(t *testing.T) {
	t.Parallel()

	gov, err := EnvironmentGovernment.endpoints()
	require.NoError(t, err)

	dod, err := EnvironmentGovernmentDoD.endpoints()
	require.NoError(t, err)

	assert.Equal(
		t,
		gov.configuration.ActiveDirectoryAuthorityHost,
		dod.configuration.ActiveDirectoryAuthorityHost,
	)
	assert.Equal(t, gov.armScope, dod.armScope)
	assert.NotEqual(t, gov.graphBaseURL, dod.graphBaseURL)
}

func TestEnvironment_UnmarshalText(t *testing.T) {
	t.Parallel()

	t.Run(
		"empty parses to public",
		func(t *testing.T) {
			t.Parallel()

			var env Environment

			err := env.UnmarshalText([]byte(""))
			require.NoError(t, err)
			assert.Equal(t, EnvironmentPublic, env)
		},
	)

	t.Run(
		"unknown value is rejected without echoing it",
		func(t *testing.T) {
			t.Parallel()

			const raw = "AZURE_GERMAN"

			var env Environment

			err := env.UnmarshalText([]byte(raw))
			require.Error(t, err)
			assert.NotContains(t, err.Error(), raw)
		},
	)
}

func TestEnvironments(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		[]Environment{
			EnvironmentPublic,
			EnvironmentGovernment,
			EnvironmentGovernmentDoD,
			EnvironmentChina,
		},
		Environments(),
	)
}
