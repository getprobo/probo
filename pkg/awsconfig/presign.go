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

package awsconfig

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// UnsignedChecksumMode keeps x-amz-checksum-mode out of a presigned GetObject URL.
//
// The default ResponseChecksumValidation resolved by config.LoadDefaultConfig is
// WHEN_SUPPORTED, which makes the SDK set "x-amz-checksum-mode: ENABLED" on every
// GetObject request so it can validate the object's checksum on the way in. On a
// presign stack that header is signed like any other, so it lands in the URL's
// X-Amz-SignedHeaders. Whoever follows the URL — a browser, an email client, curl
// — sends no such header, S3 recomputes the signature over an empty value and
// rejects the download with 403 SignatureDoesNotMatch.
//
// Nothing validates a checksum on that side of the URL anyway, so presigning asks
// for WHEN_REQUIRED and the header is never added. The SDK's own GetObject calls
// keep validating, because this only touches the presign client's options.
func UnsignedChecksumMode(opts *s3.PresignOptions) {
	opts.ClientOptions = append(
		opts.ClientOptions,
		func(o *s3.Options) {
			o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
		},
	)
}
