// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package console_v1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/coredata"
)

// TestIsFirstAttach covers the helper the credential-upload route
// uses to pick between ActionCloudAccountCreate (first attach) and
// ActionCloudAccountRotateCredentials (rotation). The contract is:
// PENDING_VERIFICATION + empty encrypted_credentials => first attach;
// every other shape => rotation.
func TestIsFirstAttach(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		account *coredata.CloudAccount
		want    bool
	}{
		{
			name: "fresh row pending with empty envelope is first attach",
			account: &coredata.CloudAccount{
				Status:               coredata.CloudAccountStatusPendingVerification,
				EncryptedCredentials: nil,
			},
			want: true,
		},
		{
			name: "pending with non-empty envelope is rotation",
			account: &coredata.CloudAccount{
				Status:               coredata.CloudAccountStatusPendingVerification,
				EncryptedCredentials: []byte("encrypted"),
			},
			want: false,
		},
		{
			name: "verified is rotation",
			account: &coredata.CloudAccount{
				Status:               coredata.CloudAccountStatusVerified,
				EncryptedCredentials: []byte("encrypted"),
			},
			want: false,
		},
		{
			name: "errored is rotation",
			account: &coredata.CloudAccount{
				Status:               coredata.CloudAccountStatusErrored,
				EncryptedCredentials: []byte("encrypted"),
			},
			want: false,
		},
		{
			name: "disconnected is rotation",
			account: &coredata.CloudAccount{
				Status:               coredata.CloudAccountStatusDisconnected,
				EncryptedCredentials: []byte("encrypted"),
			},
			want: false,
		},
		{
			name: "verified with empty envelope is still rotation -- never auto-promote",
			account: &coredata.CloudAccount{
				Status:               coredata.CloudAccountStatusVerified,
				EncryptedCredentials: nil,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, isFirstAttach(tt.account))
		})
	}
}

// TestCloudAccountCredentialUploadResponse_JSONShape pins the JSON
// envelope returned to the wizard: {"status": "...", "lastProbeError":
// "..."?}. The frontend keys on these field names; a rename here
// silently breaks the upload modal.
func TestCloudAccountCredentialUploadResponse_JSONShape(t *testing.T) {
	t.Parallel()

	t.Run("with last probe error", func(t *testing.T) {
		t.Parallel()

		errMsg := "AccessDenied"
		resp := cloudAccountCredentialUploadResponse{
			Status:         coredata.CloudAccountStatusErrored,
			LastProbeError: &errMsg,
		}

		out, err := json.Marshal(resp)
		require.NoError(t, err)

		var decoded map[string]any
		require.NoError(t, json.Unmarshal(out, &decoded))
		assert.Equal(t, "ERRORED", decoded["status"])
		assert.Equal(t, "AccessDenied", decoded["lastProbeError"])
	})

	t.Run("without last probe error", func(t *testing.T) {
		t.Parallel()

		resp := cloudAccountCredentialUploadResponse{
			Status:         coredata.CloudAccountStatusVerified,
			LastProbeError: nil,
		}

		out, err := json.Marshal(resp)
		require.NoError(t, err)

		var decoded map[string]any
		require.NoError(t, json.Unmarshal(out, &decoded))
		assert.Equal(t, "VERIFIED", decoded["status"])
		// omitempty keeps a nil pointer out of the wire payload.
		_, present := decoded["lastProbeError"]
		assert.False(t, present, "nil LastProbeError must be omitted via omitempty")
	})
}

// TestCloudAccountCredentialUploadConstants pins the two body-cap
// constants. The 16 KiB limit is deliberate: even an Azure JWT
// bundle fits comfortably under it. Any widening must be reviewed
// alongside an audit of the upload's memory model.
func TestCloudAccountCredentialUploadConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 16*1024, cloudAccountCredentialUploadMaxBodyBytes)
	assert.Equal(t, 16*1024, cloudAccountCredentialUploadMaxMemoryBytes)
	assert.LessOrEqual(
		t,
		cloudAccountCredentialUploadMaxMemoryBytes,
		cloudAccountCredentialUploadMaxBodyBytes,
		"in-memory cap must be <= total body cap so disk spill never kicks in for legitimate payloads",
	)
}
