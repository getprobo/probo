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
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/connector"
)

// Scopes is nil because the customer's own cloud policy grants the access, so
// there is nothing for the missing-scope check to compare against.
func TestWorkloadIdentityConnection_Type(t *testing.T) {
	t.Parallel()

	conn := &connector.WorkloadIdentityConnection{Cloud: cloud.AWS}

	assert.Equal(t, connector.ProtocolWorkloadIdentity, conn.Type())
	assert.Nil(t, conn.Scopes())
}

// The stored blob carries the protocol discriminator and the cloud, nothing else.
func TestWorkloadIdentityConnection_MarshalJSON(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(&connector.WorkloadIdentityConnection{Cloud: cloud.AWS})
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.Equal(
		t,
		map[string]any{
			"type":  "WORKLOAD_IDENTITY",
			"cloud": cloud.AWS,
		},
		raw,
	)
}

func TestUnmarshalConnection_WorkloadIdentity(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(&connector.WorkloadIdentityConnection{Cloud: cloud.AWS})
	require.NoError(t, err)

	conn, err := connector.UnmarshalConnection(
		string(connector.ProtocolWorkloadIdentity),
		"AWS",
		data,
	)
	require.NoError(t, err)

	workloadIdentityConn, ok := conn.(*connector.WorkloadIdentityConnection)
	require.Truef(t, ok, "expected *connector.WorkloadIdentityConnection, got %T", conn)
	assert.Equal(t, cloud.AWS, workloadIdentityConn.Cloud)
	assert.Equal(t, connector.ProtocolWorkloadIdentity, workloadIdentityConn.Type())
}
