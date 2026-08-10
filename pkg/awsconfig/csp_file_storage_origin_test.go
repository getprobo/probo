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

package awsconfig_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/awsconfig"
)

func TestCSPFileStorageOrigin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		endpoint     string
		region       string
		bucket       string
		usePathStyle bool
		want         string
		wantErr      bool
	}{
		{
			name:         "seaweedfs path style",
			endpoint:     "http://127.0.0.1:8333",
			region:       "us-east-1",
			bucket:       "probod",
			usePathStyle: true,
			want:         "http://127.0.0.1:8333",
		},
		{
			name:         "seaweedfs ip endpoint ignores virtual hosted",
			endpoint:     "http://127.0.0.1:8333",
			region:       "us-east-1",
			bucket:       "probod",
			usePathStyle: false,
			want:         "http://127.0.0.1:8333",
		},
		{
			name:         "custom endpoint virtual hosted",
			endpoint:     "https://nyc3.digitaloceanspaces.com",
			region:       "us-east-1",
			bucket:       "probod",
			usePathStyle: false,
			want:         "https://probod.nyc3.digitaloceanspaces.com",
		},
		{
			name:         "gcs path style",
			endpoint:     "https://storage.googleapis.com",
			region:       "auto",
			bucket:       "probod",
			usePathStyle: true,
			want:         "https://storage.googleapis.com",
		},
		{
			name:         "aws virtual hosted",
			endpoint:     "",
			region:       "eu-west-1",
			bucket:       "probod",
			usePathStyle: false,
			want:         "https://probod.s3.eu-west-1.amazonaws.com",
		},
		{
			name:         "aws path style",
			endpoint:     "",
			region:       "eu-west-1",
			bucket:       "probod",
			usePathStyle: true,
			want:         "https://s3.eu-west-1.amazonaws.com",
		},
		{
			name:         "defaults region when empty",
			endpoint:     "",
			region:       "",
			bucket:       "probod",
			usePathStyle: false,
			want:         "https://probod.s3." + awsconfig.DefaultRegion + ".amazonaws.com",
		},
		{
			name:         "strips path on custom endpoint",
			endpoint:     "http://127.0.0.1:8333/unused",
			region:       "us-east-1",
			bucket:       "probod",
			usePathStyle: true,
			want:         "http://127.0.0.1:8333",
		},
		{
			name:    "rejects empty bucket",
			bucket:  "",
			wantErr: true,
		},
		{
			name:    "rejects bucket with semicolon",
			bucket:  "evil;frame-ancestors",
			wantErr: true,
		},
		{
			name:     "rejects endpoint with userinfo",
			endpoint: "https://user:pass@storage.example.com", // trufflehog:ignore
			bucket:   "probod",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got, err := awsconfig.CSPFileStorageOrigin(
					tt.endpoint,
					tt.region,
					tt.bucket,
					tt.usePathStyle,
				)
				if tt.wantErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			},
		)
	}
}
