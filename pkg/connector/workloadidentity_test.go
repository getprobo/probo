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

package connector_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connector"
)

func TestWorkloadIdentityConnection_RoundTrip(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(&connector.WorkloadIdentityConnection{})
	require.NoError(t, err)

	got, err := connector.UnmarshalConnection(
		string(connector.ProtocolWorkloadIdentity),
		"AWS",
		raw,
	)
	require.NoError(t, err)

	conn, ok := got.(*connector.WorkloadIdentityConnection)
	require.True(t, ok, "expected a *WorkloadIdentityConnection")
	assert.Equal(t, connector.ProtocolWorkloadIdentity, conn.Type())
	assert.Empty(t, conn.Scopes())
}

// TestWorkloadIdentityConnection_StoresNothing pins the property the whole
// protocol rests on: the encrypted blob carries the protocol discriminator and
// nothing else, so this connector cannot leak a credential because it stores
// none. A field added here without thought would show up as an extra key.
func TestWorkloadIdentityConnection_StoresNothing(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(&connector.WorkloadIdentityConnection{})
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))

	assert.Equal(t, map[string]any{"type": "WORKLOAD_IDENTITY"}, decoded)
}

// TestWorkloadIdentityConnection_IsNotHTTP pins that the connection offers no
// HTTP credential at all, rather than a Client that fails at call time. Callers
// branch on this assertion, so a Client method appearing here would silently
// route cloud connectors down the HTTP path.
func TestWorkloadIdentityConnection_IsNotHTTP(t *testing.T) {
	t.Parallel()

	var conn connector.Connection = &connector.WorkloadIdentityConnection{}

	_, isHTTP := conn.(connector.HTTPConnection)
	assert.False(t, isHTTP)
}

// TestWorkloadIdentityConnection_RejectsNonObject pins that an unusable blob
// fails at unmarshal rather than yielding a connection that looks fine.
func TestWorkloadIdentityConnection_RejectsNonObject(t *testing.T) {
	t.Parallel()

	_, err := connector.UnmarshalConnection(
		string(connector.ProtocolWorkloadIdentity),
		"AWS",
		[]byte(`["not","an","object"]`),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot unmarshal workload identity connection")
}
