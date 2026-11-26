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

func TestListByRegion_SecondErrorReturnsNoItems(t *testing.T) {
	t.Parallel()

	calls, items, err := itam.ListByRegionSecondErrorForTest()

	require.Error(t, err)
	assert.ErrorContains(t, err, "denied")
	assert.Equal(t, 2, calls)
	assert.Nil(t, items)
}

func TestListByRegion_AccessErrorKeepsOtherRegions(t *testing.T) {
	t.Parallel()

	items, err := itam.ListByRegionAccessErrorKeepsEarlierForTest()

	require.NoError(t, err)
	assert.Equal(t, []int{1, 3}, items)
}

func TestEC2Instance_TerminatedIsSkipped(t *testing.T) {
	t.Parallel()

	assert.True(t, itam.EC2InstanceTerminatedForTest("terminated"))
	assert.False(t, itam.EC2InstanceTerminatedForTest("running"))
	assert.False(t, itam.EC2InstanceTerminatedForTest("stopped"))
}

func TestMapEC2Instance_UsesInstanceFields(t *testing.T) {
	t.Parallel()

	observed, err := itam.MapEC2ForTest()
	require.NoError(t, err)

	assert.Equal(t, "arn:aws:ec2:us-east-1:123456789012:instance/i-abc", observed.ExternalID)
	assert.Equal(t, "web", observed.Name)
	require.NotNil(t, observed.Region)
	assert.Equal(t, "us-east-1", *observed.Region)
	assert.Equal(t, itam.EC2InstanceSchemaVersion, observed.SchemaVersion)

	raw, err := json.Marshal(observed.Configuration)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"instance_type": "t3.micro",
		"state": "running",
		"vpc_id": "vpc-1",
		"subnet_id": "subnet-1",
		"public_ip": "203.0.113.10",
		"image_id": "ami-1",
		"iam_instance_profile_arn": "arn:aws:iam::123456789012:instance-profile/web",
		"security_group_ids": ["sg-1"]
	}`, string(raw))
}

func TestMapEC2Instance_FillsAbsentFields(t *testing.T) {
	t.Parallel()

	observed, err := itam.MapEC2MissingOptionalFieldsForTest()
	require.NoError(t, err)

	assert.Equal(t, "arn:aws:ec2:eu-west-1:123456789012:instance/i-bare", observed.ExternalID)
	assert.Equal(t, "i-bare", observed.Name)

	raw, err := json.Marshal(observed.Configuration)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"instance_type": "",
		"state": "",
		"vpc_id": "",
		"subnet_id": "",
		"public_ip": "",
		"image_id": "",
		"iam_instance_profile_arn": "",
		"security_group_ids": []
	}`, string(raw))
}

func TestMapEC2Instance_RejectsMissingInstanceID(t *testing.T) {
	t.Parallel()

	err := itam.MapEC2MissingInstanceIDForTest()

	require.Error(t, err)
	assert.ErrorContains(t, err, "missing instance id")
}

func TestEC2Configuration_IgnoresUnknownFields(t *testing.T) {
	t.Parallel()

	var configuration itam.EC2InstanceConfiguration
	err := json.Unmarshal([]byte(`{"instance_type":"t3.micro","unknown":1}`), &configuration)

	require.NoError(t, err)
	assert.Equal(t, "t3.micro", configuration.InstanceType)
	assert.Empty(t, configuration.State)
	assert.Empty(t, configuration.SecurityGroupIDs)
}
