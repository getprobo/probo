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
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/awsconfig"
)

var signedHeadersRe = regexp.MustCompile(`SignedHeaders=([^,]+)`)

// signedHeaders returns the SigV4 SignedHeaders list from an Authorization header.
func signedHeaders(t *testing.T, h http.Header) string {
	t.Helper()

	match := signedHeadersRe.FindStringSubmatch(h.Get("Authorization"))
	require.NotNil(t, match, "request carried no SigV4 Authorization header")

	return match[1]
}

// TestUnsignedAcceptEncoding pins the fix for S3-compatible endpoints that rewrite
// Accept-Encoding in transit. Google Cloud Storage's S3-interop endpoint does, so any
// request signing that header fails with 403 SignatureDoesNotMatch. Both directions
// are asserted so the guard cannot pass vacuously: without the middleware the header
// must be signed, with it the header must be absent from SignedHeaders but still
// present on the wire.
func TestUnsignedAcceptEncoding(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		middleware bool
		wantSigned bool
	}{
		{name: "without middleware the header is signed", middleware: false, wantSigned: true},
		{name: "with middleware it is not", middleware: true, wantSigned: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotHeader http.Header

			srv := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						gotHeader = r.Header.Clone()

						_, _ = io.Copy(io.Discard, r.Body)

						w.WriteHeader(http.StatusOK)
					},
				),
			)
			t.Cleanup(srv.Close)

			cfg, err := awsconfig.NewConfig(
				log.NewLogger(log.WithOutput(io.Discard)),
				nil,
				awsconfig.Options{
					Region:          "auto",
					Endpoint:        srv.URL,
					AccessKeyID:     "access-key",
					SecretAccessKey: "secret-key",
				},
			)
			require.NoError(t, err)

			if !tc.middleware {
				cfg.APIOptions = nil
			}

			client := s3.NewFromConfig(
				cfg,
				func(o *s3.Options) {
					o.UsePathStyle = true
				},
			)

			_, err = client.PutObject(
				context.Background(),
				&s3.PutObjectInput{
					Bucket: aws.String("bucket"),
					Key:    aws.String("key"),
					Body:   strings.NewReader("hello"),
				},
			)
			require.NoError(t, err)

			signed := signedHeaders(t, gotHeader)

			if tc.wantSigned {
				assert.Contains(
					t,
					signed,
					"accept-encoding",
					"expected the SDK to sign Accept-Encoding, otherwise this test cannot detect a regression",
				)

				return
			}

			assert.NotContains(
				t,
				signed,
				"accept-encoding",
				"signing Accept-Encoding makes GCS reject the request with SignatureDoesNotMatch",
			)

			// The header must still reach the wire: the SDK sets it to disable gzip so
			// that Content-Length and Content-Range survive for ranged downloads.
			assert.Equal(
				t,
				"identity",
				gotHeader.Get("Accept-Encoding"),
				"Accept-Encoding must be restored after signing, only unsigned",
			)
		})
	}
}
