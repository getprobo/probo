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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
)

func TestNewConnectorSettings(t *testing.T) {
	t.Parallel()

	t.Run(
		"stores canonical lowercase identifiers",
		func(t *testing.T) {
			t.Parallel()

			settings, err := cloudazure.NewConnectorSettings(
				"  "+testTenantIDUpper+"  ",
				"  "+testClientIDUpper+"  ",
				"  "+testSubscriptionIDUpper+"  ",
				"",
			)
			require.NoError(t, err)
			assert.Equal(t, testTenantID, settings.TenantID)
			assert.Equal(t, testClientID, settings.ClientID)
			assert.Equal(t, testSubscriptionID, settings.SubscriptionID)
			assert.Equal(t, cloudazure.EnvironmentPublic, settings.Environment)
		},
	)

	t.Run(
		"accepts each environment",
		func(t *testing.T) {
			t.Parallel()

			for _, env := range cloudazure.Environments() {
				t.Run(
					env.String(),
					func(t *testing.T) {
						t.Parallel()

						settings, err := cloudazure.NewConnectorSettings(
							testTenantID,
							testClientID,
							testSubscriptionID,
							env.String(),
						)
						require.NoError(t, err)
						assert.Equal(t, env, settings.Environment)
					},
				)
			}
		},
	)
}

func TestNewConnectorSettings_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		tenantID       string
		clientID       string
		subscriptionID string
		environment    string
		wantMessage    string
		forbidden      string
	}{
		{
			name:           "malformed tenant id",
			tenantID:       "not-a-guid",
			clientID:       testClientID,
			subscriptionID: testSubscriptionID,
			wantMessage:    "tenantId is not a GUID",
			forbidden:      "not-a-guid",
		},
		{
			name:           "malformed client id",
			tenantID:       testTenantID,
			clientID:       "also-not-a-guid",
			subscriptionID: testSubscriptionID,
			wantMessage:    "clientId is not a GUID",
			forbidden:      "also-not-a-guid",
		},
		{
			name:           "malformed subscription id",
			tenantID:       testTenantID,
			clientID:       testClientID,
			subscriptionID: "still-not-a-guid",
			wantMessage:    "subscriptionId is not a GUID",
			forbidden:      "still-not-a-guid",
		},
		{
			name:           "nil tenant id",
			tenantID:       nilGUID,
			clientID:       testClientID,
			subscriptionID: testSubscriptionID,
			wantMessage:    "tenantId is not a GUID",
			forbidden:      nilGUID,
		},
		{
			name:           "nil client id",
			tenantID:       testTenantID,
			clientID:       nilGUID,
			subscriptionID: testSubscriptionID,
			wantMessage:    "clientId is not a GUID",
			forbidden:      nilGUID,
		},
		{
			name:           "nil subscription id",
			tenantID:       testTenantID,
			clientID:       testClientID,
			subscriptionID: nilGUID,
			wantMessage:    "subscriptionId is not a GUID",
			forbidden:      nilGUID,
		},
		{
			name:           "unknown environment",
			tenantID:       testTenantID,
			clientID:       testClientID,
			subscriptionID: testSubscriptionID,
			environment:    "AZURE_GERMAN",
			wantMessage:    "environment is not a supported Azure environment",
			forbidden:      "AZURE_GERMAN",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				_, err := cloudazure.NewConnectorSettings(
					tt.tenantID,
					tt.clientID,
					tt.subscriptionID,
					tt.environment,
				)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantMessage)
				assert.NotContains(t, err.Error(), tt.forbidden)
			},
		)
	}
}
