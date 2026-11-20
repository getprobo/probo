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

package itam_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/itam"
)

func TestMapRDSInstance_UsesInstanceFields(t *testing.T) {
	t.Parallel()

	observed, err := itam.MapRDSForTest()
	require.NoError(t, err)

	assert.Equal(t, "arn:aws:rds:us-east-1:123456789012:db:app", observed.ExternalID)
	assert.Equal(t, "app", observed.Name)
	require.NotNil(t, observed.Region)
	assert.Equal(t, "us-east-1", *observed.Region)
	assert.Equal(t, itam.RDSDBInstanceSchemaVersion, observed.SchemaVersion)

	raw, err := json.Marshal(observed.Configuration)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"engine": "postgres",
		"engine_version": "16.3",
		"class": "db.t3.micro",
		"status": "available",
		"publicly_accessible": true,
		"storage_encrypted": true,
		"multi_az": true,
		"endpoint_host": "app.example.us-east-1.rds.amazonaws.com"
	}`, string(raw))
}

func TestMapRDSInstance_FillsAbsentFields(t *testing.T) {
	t.Parallel()

	observed, err := itam.MapRDSMissingOptionalFieldsForTest()
	require.NoError(t, err)

	assert.Equal(t, "bare", observed.Name)
	require.NotNil(t, observed.Region)
	assert.Equal(t, "eu-west-1", *observed.Region)

	raw, err := json.Marshal(observed.Configuration)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"engine": "",
		"engine_version": "",
		"class": "",
		"status": "",
		"publicly_accessible": false,
		"storage_encrypted": false,
		"multi_az": false,
		"endpoint_host": ""
	}`, string(raw))
}

func TestMapRDSInstance_RejectsMissingARN(t *testing.T) {
	t.Parallel()

	err := itam.MapRDSMissingARNForTest()

	require.Error(t, err)
	assert.ErrorContains(t, err, "missing ARN")
}

func TestMapRDSInstance_RejectsMissingIdentifier(t *testing.T) {
	t.Parallel()

	err := itam.MapRDSMissingIdentifierForTest()

	require.Error(t, err)
	assert.ErrorContains(t, err, "missing identifier")
}

func TestRDSConfiguration_IgnoresUnknownFields(t *testing.T) {
	t.Parallel()

	var configuration itam.RDSDBInstanceConfiguration
	err := json.Unmarshal([]byte(`{"engine":"postgres","unknown":1}`), &configuration)

	require.NoError(t, err)
	assert.Equal(t, "postgres", configuration.Engine)
	assert.False(t, configuration.MultiAZ)
	assert.Empty(t, configuration.EndpointHost)
}
