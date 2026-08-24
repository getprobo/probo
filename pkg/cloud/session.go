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

// Package cloud holds the narrow boundary the per-cloud packages share.
//
// The boundary is deliberately just "authenticated access to one account".
// AWS and GCP diverge too far below that for a shared abstraction to pay for
// itself, so pkg/cloud/aws and a future pkg/cloud/gcp implement Session and
// share nothing else.
package cloud

// AWS and GCP are the cloud discriminators, so that the per-cloud packages and
// the connector framework cannot drift apart on the spelling.
const (
	AWS = "AWS"
	GCP = "GCP"
)

// Session is authenticated access to one cloud account. Implementations hold
// whatever their SDK needs to sign requests; callers only ever need to know
// which cloud and which account they are talking to.
type Session interface {
	// Cloud is the discriminator of the cloud this session reaches.
	Cloud() string
	// AccountID identifies the account within that cloud — an AWS account ID,
	// a GCP project number.
	AccountID() string
}
