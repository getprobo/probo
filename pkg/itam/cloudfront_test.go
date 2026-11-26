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

func TestCollectPages_StopsAtLimit(t *testing.T) {
	t.Parallel()

	calls, err := itam.CollectPageLimitForTest(3)

	require.Error(t, err)
	assert.ErrorContains(t, err, "page limit")
	assert.Equal(t, 3, calls)
}

func TestCollectPages_ReturnsCompletedList(t *testing.T) {
	t.Parallel()

	items, err := itam.CollectSinglePageForTest()

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2}, items)
}

func TestMapCloudFrontDistribution_UsesListSummary(t *testing.T) {
	t.Parallel()

	observed, err := itam.MapCloudFrontForTest()
	require.NoError(t, err)

	assert.Equal(t, "arn:aws:cloudfront::123456789012:distribution/EDFDVBD632BHDS5", observed.ExternalID)
	assert.Equal(t, "d111111abcdef8.cloudfront.net", observed.Name)
	assert.Nil(t, observed.Region)
	assert.Equal(t, itam.CloudFrontDistributionSchemaVersion, observed.SchemaVersion)

	raw, err := json.Marshal(observed.Configuration)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"domain_name": "d111111abcdef8.cloudfront.net",
		"enabled": true,
		"status": "Deployed",
		"price_class": "PriceClass_100",
		"default_certificate": true
	}`, string(raw))
	assert.NotContains(t, string(raw), "Aliases")
}

func TestMapCloudFrontDistribution_RejectsMissingARN(t *testing.T) {
	t.Parallel()

	err := itam.MapCloudFrontMissingARNForTest()

	require.Error(t, err)
	assert.ErrorContains(t, err, "missing ARN")
}

func TestCloudFrontConfiguration_IgnoresUnknownFields(t *testing.T) {
	t.Parallel()

	var configuration itam.CloudFrontDistributionConfiguration

	err := json.Unmarshal([]byte(`{"domain_name":"d.example.net","unknown":1}`), &configuration)

	require.NoError(t, err)
	assert.Equal(t, "d.example.net", configuration.DomainName)
	assert.False(t, configuration.Enabled)
	assert.Empty(t, configuration.Status)
}
