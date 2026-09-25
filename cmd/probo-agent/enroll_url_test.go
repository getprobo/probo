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

package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/deviceagent"
)

func TestEnrollURLPreflight(t *testing.T) {
	t.Parallel()

	root := newRootCmd()
	root.SetArgs([]string{
		"enroll-url",
		"--preflight",
		"--dir", t.TempDir(),
		"probo://enroll?server=http%3A%2F%2Flocalhost%3A3000&token=abc123",
	})

	var stdout bytes.Buffer

	root.SetOut(&stdout)
	root.SetErr(&stdout)

	err := root.Execute()
	require.NoError(t, err)

	var payload enrollPreflightResponse

	err = json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload)
	require.NoError(t, err)
	require.Equal(t, "http://localhost:3000", payload.Server)
	require.Equal(t, "abc123", payload.Token)
	require.False(t, payload.AlreadyEnrolled)
	require.Contains(t, payload.ConfigDir, "TestEnrollURLPreflight")
	require.Equal(t, string(deviceagent.TrustUnknown), payload.Trust)
	require.Equal(t, deviceagent.EnrollmentConfirmTitle, payload.ConfirmTitle)
	require.Equal(
		t,
		deviceagent.EnrollmentConfirmMessage(payload.Server, deviceagent.TrustUnknown),
		payload.ConfirmMessage,
	)
}

func TestWriteEnrollPreflight(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	err := writeEnrollPreflight(
		&stdout,
		"https://example.com",
		"token-value",
		"/var/lib/probo-agent",
		true,
		deviceagent.TrustUnverified,
	)
	require.NoError(t, err)

	var payload enrollPreflightResponse

	err = json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload)
	require.NoError(t, err)
	require.True(t, payload.AlreadyEnrolled)
	require.Equal(t, "token-value", payload.Token)
	require.Equal(t, string(deviceagent.TrustUnverified), payload.Trust)
	require.Equal(t, deviceagent.EnrollmentConfirmTitle, payload.ConfirmTitle)
	require.Equal(
		t,
		deviceagent.EnrollmentConfirmMessage("https://example.com", deviceagent.TrustUnverified),
		payload.ConfirmMessage,
	)
}
