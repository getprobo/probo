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

package arn_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/awsx/arn"
)

func TestIAM(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		partition string
		want      string
	}{
		{
			name:      "commercial",
			partition: arn.Partition,
			want:      "arn:aws:iam::123456789012:root",
		},
		{
			name:      "govcloud",
			partition: "aws-us-gov",
			want:      "arn:aws-us-gov:iam::123456789012:root",
		},
		{
			name:      "china",
			partition: "aws-cn",
			want:      "arn:aws-cn:iam::123456789012:root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, arn.IAM(tt.partition, "123456789012", "root"))
		})
	}

	assert.Equal(
		t,
		"arn:aws:iam::123456789012:role/ProboAudit",
		arn.IAM(arn.Partition, "123456789012", "role/ProboAudit"),
	)
}

func TestRoleARN(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		"arn:aws:iam::123456789012:role/ProboAudit",
		arn.RoleARN(arn.Partition, "123456789012", "ProboAudit"),
	)
	assert.Equal(
		t,
		"arn:aws-us-gov:iam::123456789012:role/ProboAudit",
		arn.RoleARN("aws-us-gov", "123456789012", "ProboAudit"),
	)
}
