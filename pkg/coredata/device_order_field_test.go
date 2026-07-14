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

package coredata_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
)

func TestDeviceOrderField_Column(t *testing.T) {
	t.Parallel()

	tests := []struct {
		field coredata.DeviceOrderField
		want  string
	}{
		{field: coredata.DeviceOrderFieldCreatedAt, want: "created_at"},
		{field: coredata.DeviceOrderFieldUpdatedAt, want: "updated_at"},
		{field: coredata.DeviceOrderFieldHostname, want: "COALESCE(hostname, '')"},
		{
			field: coredata.DeviceOrderFieldLastSeenAt,
			want:  "COALESCE(last_seen_at, '0001-01-01T00:00:00Z'::timestamptz)",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.field), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.field.Column())
		})
	}
}
