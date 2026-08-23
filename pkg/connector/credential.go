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

package connector

import (
	"net/http"

	"go.probo.inc/probo/pkg/cloud"
)

type (
	// Credential is a stored Connection turned into something a caller can
	// actually talk to the service with. Neither member exposes the secret:
	// HTTPCredential's transport attaches it, CloudCredential's session mints
	// one per call.
	//
	// The set is closed to exactly the two members below — sealed by the
	// unexported method — because each one needs a different kind of consumer
	// code. A third member would be a new protocol family, not a variation.
	Credential interface {
		credential()
	}

	// HTTPCredential carries an authenticated *http.Client, which is what the
	// OAUTH2 and API_KEY protocols reduce to.
	HTTPCredential struct {
		Client *http.Client
	}

	// CloudCredential carries a cloud session, which is what WORKLOAD_IDENTITY
	// reduces to: there is no credential to attach to a request, so no client
	// can stand in for it.
	CloudCredential struct {
		Session cloud.Session
	}
)

func (HTTPCredential) credential()  {}
func (CloudCredential) credential() {}
