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
	"context"
	"fmt"

	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

const acceptEncodingHeader = "Accept-Encoding"

// signingMiddlewareID is the ID of the SigV4 middleware in the Finalize step
// (aws/signer/v4.SignHTTPRequestMiddleware). The strip/restore pair below is
// positioned relative to it by name so the ordering is explicit rather than a
// consequence of registration order.
const signingMiddlewareID = "Signing"

// acceptEncodingValueKey carries the stripped header value from the middleware
// that runs before signing to the one that runs after it.
type acceptEncodingValueKey struct{}

// unsignedAcceptEncoding keeps Accept-Encoding out of the SigV4 SignedHeaders list
// while still sending it on the wire.
//
// aws-sdk-go-v2 sets "Accept-Encoding: identity" from a Finalize middleware
// (service/internal/accept-encoding.DisableGzip) so responses are never gzipped.
// That runs before signing, so SigV4 covers the header. Google Cloud Storage's
// S3-interoperability endpoint rewrites or strips Accept-Encoding in transit, so
// the signature it recomputes can never match the one we sent, and every request
// that signs the header is rejected with 403 SignatureDoesNotMatch. The error names
// the signature, which reads as a credentials problem and is not one.
//
// botocore does not sign Accept-Encoding. That is why the aws CLI and boto3 upload
// to the same bucket with the same HMAC credentials while this SDK cannot.
//
// The header is stripped immediately before signing and restored immediately after,
// so it still reaches the wire — preserving the SDK's no-gzip behaviour, and with it
// the Content-Length and Content-Range that ranged downloads depend on — but never
// enters SignedHeaders. SigV4 does not require Accept-Encoding to be signed and
// endpoints ignore headers they were not asked to sign, so this is equally safe
// against real S3 and is applied unconditionally rather than only for custom
// endpoints.
func unsignedAcceptEncoding(stack *middleware.Stack) error {
	err := stack.Finalize.Insert(
		middleware.FinalizeMiddlewareFunc(
			"proboStripAcceptEncodingBeforeSigning",
			func(
				ctx context.Context,
				in middleware.FinalizeInput,
				next middleware.FinalizeHandler,
			) (middleware.FinalizeOutput, middleware.Metadata, error) {
				req, ok := in.Request.(*smithyhttp.Request)
				if !ok {
					return next.HandleFinalize(ctx, in)
				}

				if value := req.Header.Get(acceptEncodingHeader); value != "" {
					req.Header.Del(acceptEncodingHeader)
					ctx = context.WithValue(ctx, acceptEncodingValueKey{}, value)
				}

				return next.HandleFinalize(ctx, in)
			},
		),
		signingMiddlewareID,
		middleware.Before,
	)
	if err != nil {
		return fmt.Errorf("cannot insert accept-encoding strip middleware: %w", err)
	}

	err = stack.Finalize.Insert(
		middleware.FinalizeMiddlewareFunc(
			"proboRestoreAcceptEncodingAfterSigning",
			func(
				ctx context.Context,
				in middleware.FinalizeInput,
				next middleware.FinalizeHandler,
			) (middleware.FinalizeOutput, middleware.Metadata, error) {
				req, ok := in.Request.(*smithyhttp.Request)
				if !ok {
					return next.HandleFinalize(ctx, in)
				}

				if value, ok := ctx.Value(acceptEncodingValueKey{}).(string); ok && value != "" {
					req.Header.Set(acceptEncodingHeader, value)
				}

				return next.HandleFinalize(ctx, in)
			},
		),
		signingMiddlewareID,
		middleware.After,
	)
	if err != nil {
		return fmt.Errorf("cannot insert accept-encoding restore middleware: %w", err)
	}

	return nil
}
