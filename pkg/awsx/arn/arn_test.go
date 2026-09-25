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
	"github.com/stretchr/testify/require"
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

func TestParseRole(t *testing.T) {
	t.Parallel()

	t.Run("accepts commercial govcloud and china role arns", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			raw  string
			want arn.Role
		}{
			{
				name: "commercial",
				raw:  "arn:aws:iam::123456789012:role/ProboAudit",
				want: arn.Role{Partition: arn.Partition, AccountID: "123456789012", Name: "ProboAudit"},
			},
			{
				name: "govcloud",
				raw:  "arn:aws-us-gov:iam::123456789012:role/ProboAudit",
				want: arn.Role{Partition: arn.PartitionGov, AccountID: "123456789012", Name: "ProboAudit"},
			},
			{
				name: "china",
				raw:  "arn:aws-cn:iam::123456789012:role/ProboAudit",
				want: arn.Role{Partition: arn.PartitionChina, AccountID: "123456789012", Name: "ProboAudit"},
			},
			{
				name: "path",
				raw:  "arn:aws:iam::123456789012:role/team/CustomAudit",
				want: arn.Role{Partition: arn.Partition, AccountID: "123456789012", Name: "CustomAudit"},
			},
			{
				name: "trims space",
				raw:  "  arn:aws:iam::123456789012:role/ProboAudit  ",
				want: arn.Role{Partition: arn.Partition, AccountID: "123456789012", Name: "ProboAudit"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := arn.ParseRole(tt.raw)
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			})
		}
	})

	t.Run("refuses shapes that are not an iam role", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			raw  string
		}{
			{name: "empty", raw: ""},
			{name: "not an arn", raw: "not-an-arn"},
			{name: "user", raw: "arn:aws:iam::123456789012:user/alice"},
			{name: "s3", raw: "arn:aws:s3:::bucket/key"},
			{name: "region set", raw: "arn:aws:iam:us-east-1:123456789012:role/ProboAudit"},
			{name: "short account", raw: "arn:aws:iam::123:role/ProboAudit"},
			{name: "empty name", raw: "arn:aws:iam::123456789012:role/"},
			{name: "invalid name", raw: "arn:aws:iam::123456789012:role/not a role"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, err := arn.ParseRole(tt.raw)
				require.ErrorIs(t, err, arn.ErrNotRole)

				if tt.raw != "" {
					assert.NotContains(t, err.Error(), tt.raw)
				}
			})
		}
	})

	t.Run("refuses an unsupported partition without echoing it", func(t *testing.T) {
		t.Parallel()

		raw := "arn:aws-iso:iam::123456789012:role/ProboAudit"

		_, err := arn.ParseRole(raw)
		require.ErrorIs(t, err, arn.ErrUnsupportedPartition)
		assert.NotContains(t, err.Error(), raw)
		assert.NotContains(t, err.Error(), "aws-iso")
	})
}
