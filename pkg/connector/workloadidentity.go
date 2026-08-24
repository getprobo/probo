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

type (
	// WorkloadIdentityConnection connects to a customer's cloud account by
	// federated workload identity: Probo mints a short-lived OIDC assertion
	// (pkg/identityfederation) and the cloud provider's STS exchanges it for
	// temporary credentials.
	//
	// It is empty on purpose, and must stay empty. Everything needed to obtain
	// credentials is the public issuer plus the non-secret connector settings,
	// so this connector cannot leak a credential because it stores nothing at
	// all. Which cloud to reach comes from the connector's provider.
	//
	// It is deliberately not an HTTPConnection: cloud SDK credentials sign
	// requests the SDK builds itself, so there is no transport to carry them.
	WorkloadIdentityConnection struct{}
)

const (
	ProtocolWorkloadIdentity ProtocolType = "WORKLOAD_IDENTITY"
)

var (
	_ Connection           = (*WorkloadIdentityConnection)(nil)
	_ OrganizationSelector = (*WorkloadIdentityConnection)(nil)
	_ ScopeGranter         = (*WorkloadIdentityConnection)(nil)
	_ Reconnector          = (*WorkloadIdentityConnection)(nil)
)

func (c *WorkloadIdentityConnection) Type() ProtocolType {
	return ProtocolWorkloadIdentity
}

func (c *WorkloadIdentityConnection) Scopes() []string {
	return []string{}
}

func (c *WorkloadIdentityConnection) SupportsOrganizationPicker() bool { return false }
func (c *WorkloadIdentityConnection) SupportsScopeGrantCheck() bool    { return false }

// SupportsReconnect is false because there is nothing to refresh: the
// connection stores no credential, and the trust the customer granted lives
// in their own cloud account. Re-running setup means redeploying their stack,
// not walking a Probo redirect flow.
func (c *WorkloadIdentityConnection) SupportsReconnect() bool { return false }

func (c WorkloadIdentityConnection) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		&struct {
			Type string `json:"type"`
		}{
			Type: string(ProtocolWorkloadIdentity),
		},
	)
}

// UnmarshalJSON reads nothing, because the stored blob carries nothing but the
// protocol discriminator. It still decodes, so a blob that is not a JSON object
// fails here rather than somewhere later.
func (c *WorkloadIdentityConnection) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &struct{}{})
}
