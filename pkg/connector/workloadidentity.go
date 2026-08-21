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
	"encoding/json"
)

// WorkloadIdentityConnection federates into a customer's cloud with OIDC rather
// than holding a credential. It cannot leak one because it holds none: the
// encrypted blob carries only the discriminator below, and the token is minted
// in-process per use.
type WorkloadIdentityConnection struct {
	// Cloud is a pkg/cloud discriminator (cloud.AWS, cloud.GCP). Not typed as
	// such because nothing here acts on the value.
	Cloud string `json:"cloud"`
}

var _ Connection = (*WorkloadIdentityConnection)(nil)

func (c *WorkloadIdentityConnection) Type() ProtocolType {
	return ProtocolWorkloadIdentity
}

// Scopes returns nil: access is granted by the customer's own cloud policy (an
// AWS IAM trust policy, a GCP attribute condition), not by an OAuth grant Probo
// holds. Nothing here to compare against a provider's required scopes.
func (c *WorkloadIdentityConnection) Scopes() []string {
	return nil
}

func (c WorkloadIdentityConnection) MarshalJSON() ([]byte, error) {
	type Alias WorkloadIdentityConnection

	return json.Marshal(&struct {
		Type string `json:"type"`
		Alias
	}{
		Type:  string(ProtocolWorkloadIdentity),
		Alias: Alias(c),
	})
}

func (c *WorkloadIdentityConnection) UnmarshalJSON(data []byte) error {
	type Alias WorkloadIdentityConnection

	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	return json.Unmarshal(data, &aux)
}
