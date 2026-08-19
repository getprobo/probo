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

// Package cloud holds the one abstraction the AWS and GCP audit connectors
// share: a Session, meaning "authenticated access to one cloud account".
//
// The boundary stays narrow deliberately. The two SDKs diverge too much below it
// to share driver code, so a consumer needing more type-asserts the concrete
// session (e.g. *aws.Session) and works against that cloud's SDK directly.
package cloud

const (
	AWS = "AWS"
	GCP = "GCP"
)

// Session is what the WORKLOAD_IDENTITY protocol replaces an *http.Client with.
// Its credential is minted per use and never stored.
type Session interface {
	Cloud() string
	// AccountID is the cloud's own identifier: an AWS account ID, a GCP project
	// number.
	AccountID() string
}
